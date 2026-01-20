/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package org.yuanrong.runtime;

import org.yuanrong.Config;
import org.yuanrong.CreateParam;
import org.yuanrong.FunctionWrapper;
import org.yuanrong.GetParams;
import org.yuanrong.GroupOptions;
import org.yuanrong.InvokeOptions;
import org.yuanrong.MSetParam;
import org.yuanrong.SetParam;
import org.yuanrong.api.InvokeArg;
import org.yuanrong.api.Node;
import org.yuanrong.api.YR;
import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.errorcode.Pair;
import org.yuanrong.exception.YRException;
import org.yuanrong.exception.LibRuntimeException;
import org.yuanrong.exception.handler.traceback.StackTraceUtils;
import org.yuanrong.jni.LibRuntime;
import org.yuanrong.jni.LibRuntimeConfig;
import org.yuanrong.libruntime.generated.Libruntime;
import org.yuanrong.libruntime.generated.Libruntime.FunctionMeta;
import org.yuanrong.runtime.client.KVManager;
import org.yuanrong.runtime.client.ObjectRef;
import org.yuanrong.runtime.util.Constants;
import org.yuanrong.runtime.util.Utils;
import org.yuanrong.serialization.Serializer;
import org.yuanrong.serialization.strategy.Strategy;
import org.yuanrong.storage.InternalWaitResult;
import org.yuanrong.stream.Consumer;
import org.yuanrong.stream.ConsumerImpl;
import org.yuanrong.stream.Producer;
import org.yuanrong.stream.ProducerConfig;
import org.yuanrong.stream.ProducerImpl;
import org.yuanrong.stream.SubscriptionConfig;
import org.yuanrong.utils.SdkUtils;

import org.apache.commons.lang3.tuple.ImmutablePair;
import org.objectweb.asm.Type;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.lang.reflect.Constructor;
import java.lang.reflect.Executable;
import java.lang.reflect.Method;
import java.nio.ByteBuffer;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Iterator;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;
import java.util.concurrent.locks.Lock;
import java.util.concurrent.locks.ReentrantReadWriteLock;
import java.util.stream.Collectors;

/**
 * Cluster mode runtime implementation
 *
 * @since 2023 /10/16
 */
public class ClusterModeRuntime implements Runtime {
    private static final Logger LOGGER = LoggerFactory.getLogger(ClusterModeRuntime.class);

    /**
     * The Constructor name.
     */
    private static final String CONSTRUCTOR_NAME = "<init>";

    /**
     * The Functions.
     */
    ConcurrentMap<String, Map<org.apache.commons.lang3.tuple.Pair<String, String>,
        org.apache.commons.lang3.tuple.Pair<FunctionWrapper, Boolean>>>
        functions = new ConcurrentHashMap<>();

    private final ReentrantReadWriteLock rwLock = new ReentrantReadWriteLock();

    private final Lock rLock = rwLock.readLock();

    private final Lock wLock = rwLock.writeLock();

    /**
     * Runtime configuration
     */
    private Config config;

    /**
     * Instantiates a new Cluster mode runtime.
     */
    public ClusterModeRuntime() {}

    @Override
    public void init(LibRuntimeConfig config) throws YRException {
        LOGGER.debug("cluster mode runtime init get java functionIds: {}",
            config.getFunctionIds().get(Libruntime.LanguageType.Java));
        ErrorInfo errorInfo = LibRuntime.Init(config);
        Utils.checkErrorAndThrow(errorInfo, "failed to init libruntime");
        LOGGER.debug("succeed to init libruntime)");
    }

    @Override
    public List<Object> get(List<ObjectRef> refs, int timeoutSec, boolean allowPartial) throws YRException {
        if (timeoutSec < Constants.NO_TIMEOUT) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                    "get config timeout (" + timeoutSec + ") is invalid");
        }
        List<String> refIds = refs.stream().map(ObjectRef::getObjectID).collect(Collectors.toList());
        long timeoutMs = Constants.NO_TIMEOUT;
        if (timeoutSec != Constants.NO_TIMEOUT) {
            timeoutMs = (long) timeoutSec * Constants.SEC_TO_MS;
        }
        Pair<ErrorInfo, List<byte[]>> getRes;
        try {
            // This method may throw 'LibruntimeException' when timeout occurs.
            getRes = LibRuntime.Get(refIds, (int) timeoutMs, allowPartial);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        } catch (Exception e) {
            throw new YRException(ErrorCode.ERR_GET_OPERATION_FAILED, ModuleCode.RUNTIME, e);
        }
        StackTraceUtils.checkErrorAndThrowForInvokeException(getRes.getFirst(), getRes.getFirst().getErrorMessage());
        List<byte[]> res = getRes.getSecond();
        return Strategy.getObjects(res, refs);
    }

    @Override
    public Object get(ObjectRef ref, int timeoutSec) throws YRException {
        if (ref == null) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME, "ObjectRef cannot be null");
        }
        List<ObjectRef> refs = Collections.singletonList(ref);
        List<Object> objects;
        try {
            objects = get(refs, timeoutSec, false);
        } catch (YRException e) {
            SdkUtils.filterGetTraceInfo(e);
            throw e;
        }

        return objects.size() > 0 ? objects.get(0) : null;
    }

    @Override
    public ObjectRef put(Object obj, long interval, long retryTime) throws YRException {
        ByteBuffer byteBuffer;
        try {
            byteBuffer = Strategy.selectSerializationStrategy(obj).serialize(obj);
        } catch (IOException e) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME, e);
        }
        List<String> nestObjIds = new ArrayList<>(Serializer.CONTAINED_OBJECT_IDS.get());
        Pair<ErrorInfo, String> res;
        try {
            res = LibRuntime.Put(byteBuffer.array(), nestObjIds);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        }
        Utils.checkErrorAndThrow(res.getFirst(), "put object");
        ObjectRef ret = new ObjectRef(res.getSecond());
        ret.setByteBuffer((obj instanceof ByteBuffer));
        ret.setType(obj.getClass());
        return ret;
    }

    @Override
    public ObjectRef put(Object obj, long interval, long retryTime, CreateParam createParam)
        throws YRException {
        ByteBuffer byteBuffer;
        try {
            byteBuffer = Strategy.selectSerializationStrategy(obj).serialize(obj);
        } catch (IOException e) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME, e);
        }
        List<String> nestObjIds = new ArrayList<>(Serializer.CONTAINED_OBJECT_IDS.get());
        Pair<ErrorInfo, String> res;
        try {
            res = LibRuntime.PutWithParam(byteBuffer.array(), nestObjIds, createParam);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        }
        Utils.checkErrorAndThrow(res.getFirst(), "put object");
        ObjectRef ret = new ObjectRef(res.getSecond());
        ret.setByteBuffer((obj instanceof ByteBuffer));
        ret.setType(obj.getClass());
        return ret;
    }

    @Override
    public InternalWaitResult wait(List<String> objIds, int waitNum, int timeoutSec) throws YRException {
        if (timeoutSec <= 0 && timeoutSec != Constants.NO_TIMEOUT) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                    "timeout should be larger than 0 or be -1");
        }
        if (checkRepeated(objIds)) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                    "duplicate objectRef exists in objs list");
        }
        if (waitNum == 0) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME, "waitNum cannot be 0");
        }
        InternalWaitResult waitResult;
        try {
            waitResult = LibRuntime.Wait(objIds, waitNum, timeoutSec);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        }
        if (waitResult == null) {
            throw new YRException(ErrorCode.ERR_INNER_SYSTEM_ERROR, ModuleCode.RUNTIME,
                    "failed to get wait result");
        }
        if (waitResult.getExceptionIds().size() > 0) {
            LOGGER.warn("wait objIds, exception ids size is {}", waitResult.getExceptionIds().size());
            Iterator<Map.Entry<String, ErrorInfo>> it = waitResult.getExceptionIds().entrySet().iterator();
            Map.Entry<String, ErrorInfo> entry = it.next();
            StackTraceUtils.checkErrorAndThrowForInvokeException(entry.getValue(),
                "wait objIds(" + entry.getKey() + "...)");
        }
        return waitResult;
    }

    private boolean checkRepeated(List<String> objIds) {
        Set<String> sets = new HashSet<>();
        for (String id : objIds) {
            if (sets.contains(id)) {
                return true;
            }
            sets.add(id);
        }
        return false;
    }

    /**
     * Stop runtime.
     */
    static void stopRuntime() {
    }

    /**
     * Invoke function by function name as method name in Java.
     *
     * @param functionMeta the function meta
     * @param args         the args
     * @param opt          the opt
     * @return string of object id
     * @throws YRException the YR exception.
     */
    @Override
    public String invokeByName(
            FunctionMeta functionMeta, List<InvokeArg> args, InvokeOptions opt) throws YRException {
        String language = functionMeta.getLanguage().name();
        LOGGER.debug("start invoke function by name({}), language({})", functionMeta.getFunctionName(), language);
        Pair<ErrorInfo, String> res;
        try {
            res = LibRuntime.InvokeInstance(functionMeta, "", args, opt);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        }
        Utils.checkErrorAndThrow(res.getFirst(), "invoke function");
        String objId = res.getSecond();
        LOGGER.debug("succeed to invoke function by name({}), language({}), objId({})", functionMeta.getFunctionName(),
                language, objId);
        return objId;
    }

    @Override
    public ObjectRef invokeJavaByName(FunctionMeta functionMeta, InvokeOptions opt, Object... args)
        throws YRException {
        FunctionWrapper function = YR.getRuntime().getJavaFunction(functionMeta);
        SdkUtils.checkJavaParameterTypes(function, args);
        String objId = this.invokeByName(functionMeta, SdkUtils.packInvokeArgs(args), opt);
        Class<?> returnType = function.getReturnType().orElse(null);
        return new ObjectRef(objId, returnType);
    }

    /**
     * Invokes a function on a specific instance.
     *
     * @param functionMeta The metadata of the function to be invoked.
     * @param instanceId   The ID of the instance on which the function will be
     *                     invoked.
     * @param args         The list of arguments to be passed to the function.
     * @param opt          The invoke options to be used for the function
     *                     invocation.
     * @return The result of the function invocation.
     * @throws YRException if an error occurs during the invocation.
     */
    @Override
    public String invokeInstance(FunctionMeta functionMeta, String instanceId, List<InvokeArg> args,
            InvokeOptions opt) throws YRException {
        LOGGER.debug("start to invoke instance({}) functionMeta is {}", instanceId, functionMeta);
        Pair<ErrorInfo, String> res;
        try {
            res = LibRuntime.InvokeInstance(functionMeta, instanceId, args, opt);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        }
        Utils.checkErrorAndThrow(res.getFirst(), "invoke instance");
        String objId = res.getSecond();
        LOGGER.debug("succeed to invoke instance({}) objId({})", instanceId, objId);
        return objId;
    }

    /**
     * Create instance string.
     *
     * @param functionMeta the function meta
     * @param args         the args
     * @param opt          the opt
     * @return instanceId
     * @throws YRException the YR exception.
     */
    @Override
    public String createInstance(FunctionMeta functionMeta, List<InvokeArg> args, InvokeOptions opt)
        throws YRException {
        LOGGER.debug("start to create instance {}", functionMeta);
        Pair<ErrorInfo, String> res;
        try {
            res = LibRuntime.CreateInstance(functionMeta, args, opt);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        }
        Utils.checkErrorAndThrow(res.getFirst(), "create instance");
        String instanceId = res.getSecond();
        LOGGER.info("succeed to create instance({})", instanceId);
        return instanceId;
    }

    /**
     * Cancel.
     *
     * @param objectId the object id
     */
    @Override
    public void cancel(String objectId) {
    }

    /**
     * Terminate instance.
     *
     * @param instanceId the instance id
     * @throws YRException the YR exception.
     */
    @Override
    public void terminateInstance(String instanceId) throws YRException {
        ErrorInfo errorInfo;
        try {
            errorInfo = LibRuntime.Kill(instanceId);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        }
        Utils.checkErrorAndThrow(errorInfo, "kill instance(" + instanceId + ")");
        LOGGER.info("succeed to terminate instance({})", instanceId);
    }

    /**
     * Decrease reference.
     *
     * @param ids the ids.
     */
    @Override
    public void decreaseReference(List<String> ids) {
        LibRuntime.DecreaseReference(ids);
    }

    /**
     * Is on cloud boolean.
     *
     * @return the boolean
     */
    @Override
    public boolean isDriver() {
        return true;
    }

    /**
     * Get real instance id.
     *
     * @param objectId the object id
     * @return the string representing the real instance id
     * @throws YRException the actor task exception.
     */
    @Override
    public String getRealInstanceId(String objectId) throws YRException {
        try {
            return LibRuntime.GetRealInstanceId(objectId);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        }
    }

    /**
     * Save real instance id.
     *
     * @param objectId the object id
     * @param instanceId the instance id
     * @param opts the invoke options
     * @throws YRException the actor task exception.
     */
    @Override
    public void saveRealInstanceId(String objectId, String instanceId, InvokeOptions opts) throws YRException {
        try {
            LibRuntime.SaveRealInstanceId(objectId, instanceId, opts);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        }
    }

    /**
     * Finalizes all actors and tasks and release any resources associated with
     * them.
     * Shuts down the KVManager Thread executor service. Any currently executing
     * tasks will be interrupted.
     */
    @Override
    public void Finalize() {
        LibRuntime.Finalize();
    }

    /**
     * In the runtimeCtx, finalizes all actors and tasks and release any resources
     * associated with them.
     * Shuts down the KVManager Thread executor service. Any currently executing
     * tasks will be interrupted.
     *
     * @param runtimeCtx the runtimeCtx
     * @param leftRuntimeNum the leftRuntimeNum
     */
    @Override
    public void Finalize(String runtimeCtx, int leftRuntimeNum) {
        LibRuntime.FinalizeWithCtx(runtimeCtx);
    }

    /**
     * Exit the current instance. This function only supports in runtime.
     */
    @Override
    public void exit() {
        LibRuntime.Exit();
    }

    /**
     * Returns the key-value manager instance.
     *
     * @return The key-value manager instance
     */
    @Override
    public KVManager getKVManager() {
        return KVManager.getSingleton();
    }

    /**
     * Writes a key-value pair to the key-value store with the specified setParam
     * option.
     *
     * @param key       The key of the key-value pair
     * @param value     The value of the key-value pair
     * @param setParam  The setParam option for datasystem stateClient, contain writeMode,
     *                  ttlSecond, existence and cacheType.
     * @throws YRException if an error occurs while performing the write operation
     */
    @Override
    public void KVWrite(String key, byte[] value, SetParam setParam) throws YRException {
        if (value == null) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                    "Cannot set a null value to key: " + key);
        }
        ErrorInfo errorInfo;
        try {
            errorInfo = LibRuntime.KVWrite(key, value, setParam);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        }
        if (!ErrorCode.ERR_OK.equals(errorInfo.getErrorCode())) {
            LOGGER.error("KVWrite err: Code:{}, MCode:{}, Msg:{}", errorInfo.getErrorCode(), errorInfo.getModuleCode(),
                    errorInfo.getErrorMessage());
            throw new YRException(errorInfo);
        }
    }

    /**
     * Set Multiple key-value pairs.
     * Operations on all key-value pairs must either succeed or fail completely.
     *
     * @param keys      The keys of the key-value pair
     * @param values    The values of the key-value pair
     * @param mSetParam The mSetParam option for datasystem stateClient, contain writeMode,
     *                  ttlSecond, existence and cacheType.
     * @throws YRException if an error occurs while performing the write operation
     */
    @Override
    public void KVMSetTx(List<String> keys, List<byte[]> values, MSetParam mSetParam) throws YRException {
        for (int i = 0; i < keys.size(); ++i) {
            if (values.get(i) == null) {
                throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                        "Cannot set a null value to key: " + keys.get(i));
            }
        }
        ErrorInfo errorInfo;
        try {
            errorInfo = LibRuntime.KVMSetTx(keys, values, mSetParam);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        }
        if (!ErrorCode.ERR_OK.equals(errorInfo.getErrorCode())) {
            LOGGER.error("KVMSetTx err: Code:{}, MCode:{}, Msg:{}", errorInfo.getErrorCode(), errorInfo.getModuleCode(),
                    errorInfo.getErrorMessage());
            throw new YRException(errorInfo);
        }
    }

    /**
     * Gets java function.
     *
     * @param functionMeta the descriptor
     * @return the function
     * @throws YRException the YR exception
     */
    @Override
    public FunctionWrapper getJavaFunction(FunctionMeta functionMeta) throws YRException {
        String className = functionMeta.getClassName();
        Map<org.apache.commons.lang3.tuple.Pair<String, String>,
            org.apache.commons.lang3.tuple.Pair<FunctionWrapper, Boolean>>
            allClassFunctions = functions.get(className);
        if (allClassFunctions == null) {
            synchronized (this) {
                allClassFunctions = functions.get(className);
                if (allClassFunctions == null) {
                    allClassFunctions = loadFunctionsForClass(className);
                    functions.putIfAbsent(className, allClassFunctions);
                }
            }
        }
        final org.apache.commons.lang3.tuple.Pair<String, String> key = ImmutablePair.of(functionMeta.getFunctionName(),
                functionMeta.getSignature());
        FunctionWrapper func = allClassFunctions.get(key).getLeft();
        if (func == null) {
            if (allClassFunctions.containsKey(key)) {
                throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                        String.format("FunctionWrapper %s is overloaded, the signature can't be empty.",
                                functionMeta.getFunctionName()));
            } else {
                throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME, String.format(
                        "FunctionWrapper %s not found", functionMeta.getFunctionName()));
            }
        }
        return func;
    }

    /**
     * Load functions for class map.
     *
     * @param className the class name
     * @return the map
     * @throws YRException the YR exception.
     */
    private Map<org.apache.commons.lang3.tuple.Pair<String, String>,
        org.apache.commons.lang3.tuple.Pair<FunctionWrapper, Boolean>> loadFunctionsForClass(
        String className) throws YRException {
        ClassLoader classLoader = getClass().getClassLoader();
        Class<?> clazz;
        try {
            clazz = Class.forName(className, true, classLoader);
        } catch (ClassNotFoundException e) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                "class not found:" + className);
        }
        List<Executable> executables = new ArrayList<>();
        executables.addAll(Arrays.asList(clazz.getDeclaredMethods()));
        executables.addAll(Arrays.asList(clazz.getDeclaredConstructors()));
        Class<?> clz = clazz;
        clz = clz.getSuperclass();
        while (clz != null && clz != Object.class) {
            executables.addAll(Arrays.asList(clz.getDeclaredMethods()));
            clz = clz.getSuperclass();
        }
        for (Class<?> baseInterface : clazz.getInterfaces()) {
            for (Method method : baseInterface.getDeclaredMethods()) {
                if (method.isDefault()) {
                    executables.add(method);
                }
            }
        }
        Collections.reverse(executables);
        Map<org.apache.commons.lang3.tuple.Pair<String, String>,
            org.apache.commons.lang3.tuple.Pair<FunctionWrapper, Boolean>> map = new HashMap<>();
        // method name, signature ;yrfunction, isDefault function in interface
        for (Executable executable : executables) {
            executable.setAccessible(true);
            final Type type;
            if (executable instanceof Method) {
                type = Type.getType((Method) executable);
            } else if (executable instanceof Constructor) {
                type = Type.getType((Constructor) executable);
            } else {
                throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                    "executable is not an instance of Method or Constructor");
            }
            final String signature = type.getDescriptor();
            String declaringClassName = executable.getDeclaringClass().getName();
            final String methodName = executable instanceof Method ? executable.getName() : CONSTRUCTOR_NAME;
            FunctionMeta meta = FunctionMeta.newBuilder().setClassName(declaringClassName)
                    .setFunctionName(methodName)
                    .setSignature(signature)
                    .setLanguage(Libruntime.LanguageType.Java).build();
            FunctionWrapper yrFunction = new FunctionWrapper(executable, meta);
            final boolean isDefault = executable instanceof Method && ((Method) executable).isDefault();
            map.put(ImmutablePair.of(methodName, signature), ImmutablePair.of(yrFunction, isDefault));
        }
        return map;
    }

    /**
     * Reads the value associated with the specified key from the key-value store.
     *
     * @param key       The key to read value for
     * @param timeoutMS The maximum time to wait for the read operation to complete, in
     *                  milliseconds
     * @return A byte array representing the value associated with the given key
     * @throws YRException If an error occurs while performing the read operation
     */
    @Override
    public byte[] KVRead(String key, int timeoutMS) throws YRException {
        Pair<byte[], ErrorInfo> result;
        try {
            result = LibRuntime.KVRead(key, timeoutMS);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        }
        ErrorInfo errorInfo = result.getSecond();
        if (!ErrorCode.ERR_OK.equals(errorInfo.getErrorCode())) {
            LOGGER.error("KVRead err: Code:{}, MCode:{}, Msg:{}", errorInfo.getErrorCode(), errorInfo.getModuleCode(),
            errorInfo.getErrorMessage());
            throw new YRException(errorInfo);
        }
        return result.getFirst();
    }

    /**
     * Reads the values associated with the given keys from the key-value store.
     *
     * @param keys         The list of keys to read values for
     * @param timeoutMS    The maximum time to wait for the read operation to
     *                     complete, in milliseconds
     * @param allowPartial Specifies whether to allow partial reads or not
     * @return A list of byte arrays representing the values associated with the
     *         given keys
     * @throws YRException If an error occurs while performing the read operation
     */
    @Override
    public List<byte[]> KVRead(List<String> keys, int timeoutMS, boolean allowPartial) throws YRException {
        Pair<List<byte[]>, ErrorInfo> result;
        try {
            result = LibRuntime.KVRead(keys, timeoutMS, allowPartial);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        }
        ErrorInfo errorInfo = result.getSecond();
        if (!ErrorCode.ERR_OK.equals(errorInfo.getErrorCode())) {
            LOGGER.error("KVRead err: Code:{}, MCode:{}, Msg:{}", errorInfo.getErrorCode(), errorInfo.getModuleCode(),
            errorInfo.getErrorMessage());
            throw new YRException(errorInfo);
        }
        return result.getFirst();
    }

    /**
     * Reads the values associated with the given keys and get params from the key-value store.
     *
     * @param keys The keys to retrieve values for
     * @param params The get params for datasystem
     * @param timeoutMS    The maximum time to wait for the read operation to
     *                     complete, in milliseconds
     * @return A list of byte arrays representing the values associated with the
     *         keys
     * @throws YRException if an error occurs while retrieving the values
     */
    @Override
    public List<byte[]> KVGetWithParam(List<String> keys, GetParams params, int timeoutMS) throws YRException {
        Pair<List<byte[]>, ErrorInfo> result;
        try {
            result = LibRuntime.KVGetWithParam(keys, params, timeoutMS);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        }
        ErrorInfo errorInfo = result.getSecond();
        if (!ErrorCode.ERR_OK.equals(errorInfo.getErrorCode())) {
            LOGGER.error("KVGetWithParam err: Code:{}, MCode:{}, Msg:{}", errorInfo.getErrorCode(),
                    errorInfo.getModuleCode(), errorInfo.getErrorMessage());
            throw new YRException(errorInfo);
        }
        return result.getFirst();
    }

    /**
     * Deletes a key-value pair.
     *
     * @param key The key of the pair to be deleted.
     * @throws YRException If an error occurs while performing the delete operation
     */
    @Override
    public void KVDel(String key) throws YRException {
        ErrorInfo errorInfo;
        try {
            errorInfo = LibRuntime.KVDel(key);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        }
        if (!ErrorCode.ERR_OK.equals(errorInfo.getErrorCode())) {
            LOGGER.error("KVDel err: Code:{}, MCode:{}, Msg:{}", errorInfo.getErrorCode(), errorInfo.getModuleCode(),
            errorInfo.getErrorMessage());
            throw new YRException(errorInfo);
        }
    }

    /**
     * Deletes multiple key-value pairs from the key-value store.
     *
     * @param keys A list of keys of the pairs to be deleted
     * @return A list of keys that were failed to be deleted
     * @throws YRException If an error occurs while performing the delete operation
     */
    @Override
    public List<String> KVDel(List<String> keys) throws YRException {
        Pair<List<String>, ErrorInfo> result = null;
        try {
            result = LibRuntime.KVDel(keys);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        }
        ErrorInfo errorInfo = result.getSecond();
        if (!ErrorCode.ERR_OK.equals(errorInfo.getErrorCode())) {
            LOGGER.error("KVDel err: Code:{}, MCode:{}, Msg:{}", errorInfo.getErrorCode(), errorInfo.getModuleCode(),
            errorInfo.getErrorMessage());
        }
        return result.getFirst();
    }

    /**
     * Load state of the runtime with a timeout.
     *
     * @param timeoutSec The maximum time to wait for the load operation to complete,
     *                   in seconds.
     * @throws YRException If an error occurs while performing the load operation.
     */
    @Override
    public void loadState(int timeoutSec) throws YRException {
        int timeoutMs = Constants.NO_TIMEOUT;
        if (timeoutSec != Constants.NO_TIMEOUT) {
            timeoutMs = timeoutSec * Constants.SEC_TO_MS;
        }
        ErrorInfo errorInfo;
        try {
            errorInfo = LibRuntime.LoadState(timeoutMs);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        }
        Utils.checkErrorAndThrow(errorInfo, "Load state error");
    }

    /**
     * Saves state of the runtime with a timeout.
     *
     * @param timeoutSec The maximum time to wait for the save operation to complete,
     *                   in seconds.
     * @throws YRException If an error occurs while performing the save operation.
     */
    @Override
    public void saveState(int timeoutSec) throws YRException {
        int timeoutMs = Constants.NO_TIMEOUT;
        if (timeoutSec != Constants.NO_TIMEOUT) {
            timeoutMs = timeoutSec * Constants.SEC_TO_MS;
        }
        ErrorInfo errorInfo;
        try {
            errorInfo = LibRuntime.SaveState(timeoutMs);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        }
        Utils.checkErrorAndThrow(errorInfo, "Save state error");
    }

    @Override
    public void groupCreate(String groupName, GroupOptions opts) throws YRException {
        if (opts.getTimeout() != -1 && opts.getTimeout() < 0) {
            ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                "The value of timeout should be -1 or greater than 0");
            throw new YRException(errorInfo);
        }
        ErrorInfo errorInfo;
        try {
            errorInfo = LibRuntime.GroupCreate(groupName, opts);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        }
        if (!ErrorCode.ERR_OK.equals(errorInfo.getErrorCode())) {
            LOGGER.error("group create error: Code:{}, MCode:{}, Msg:{}", errorInfo.getErrorCode(),
                errorInfo.getModuleCode(), errorInfo.getErrorMessage());
            throw new YRException(errorInfo);
        }
    }
    @Override
    public void groupWait(String groupName) throws YRException {
        ErrorInfo errorInfo;
        try {
            errorInfo = LibRuntime.GroupWait(groupName);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        }
        if (!ErrorCode.ERR_OK.equals(errorInfo.getErrorCode())) {
            LOGGER.error("group wait error: Code:{}, MCode:{}, Msg:{}", errorInfo.getErrorCode(),
                errorInfo.getModuleCode(), errorInfo.getErrorMessage());
            throw new YRException(errorInfo);
        }
    }

    @Override
    public void groupTerminate(String groupName) {
        LibRuntime.GroupTerminate(groupName);
    }

    /**
     * createStreamProducer
     *
     * @param streamName the stream name
     * @param producerConf the producer conf
     * @return Producer the stream producer
     * @throws YRException if there is an exception during creating stream producer
     */
    @Override
    public Producer createStreamProducer(String streamName, ProducerConfig producerConf) throws YRException {
        if (producerConf.getMaxStreamSize() < 0) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                    "maxStreamSize (" + producerConf.getMaxStreamSize() + ") is invalid, expect >= 0");
        }
        if (producerConf.getRetainForNumConsumers() < 0) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                    "retainForNumConsumers (" + producerConf.getRetainForNumConsumers() + ") is invalid, expect >= 0");
        }
        if (producerConf.getReserveSize() < 0) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                    "reserveSize (" + producerConf.getReserveSize() + ") is invalid, expect >= 0");
        }
        rLock.lock();
        try {
            long producerPtr = LibRuntime.CreateStreamProducerWithConfig(streamName, producerConf);
            return new ProducerImpl(producerPtr);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        } finally {
            rLock.unlock();
        }
    }

    /**
     * createStreamConsumer
     *
     * @param streamName the stream name
     * @param config the subscription conf
     * @param autoAck if consumer auto ack
     * @return Consumer the stream consumer
     * @throws YRException if there is an exception during creating stream consumer
     */
    @Override
    public Consumer createStreamConsumer(String streamName, SubscriptionConfig config, boolean autoAck)
        throws YRException {
        rLock.lock();
        try {
            long consumerPtr = LibRuntime.CreateStreamConsumer(streamName, config.getSubscriptionName(),
                config.getSubscriptionType(), autoAck);
            return new ConsumerImpl(consumerPtr);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        } finally {
            rLock.unlock();
        }
    }

    /**
     * deleteStream
     *
     * @param streamName the stream name
     * @throws YRException if there is an exception when delete stream
     */
    @Override
    public void deleteStream(String streamName) throws YRException {
        rLock.lock();
        try {
            ErrorInfo err = LibRuntime.DeleteStream(streamName);
            StackTraceUtils.checkErrorAndThrowForInvokeException(err, err.getErrorMessage());
        } finally {
            rLock.unlock();
        }
    }

    /**
     * queryGlobalProducersNum
     *
     * @param streamName the stream name
     * @return long the producers num
     * @throws YRException if there is an exception during quering global producersNum
     */
    @Override
    public long queryGlobalProducersNum(String streamName) throws YRException {
        rLock.lock();
        try {
            return LibRuntime.QueryGlobalProducersNum(streamName);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        } finally {
            rLock.unlock();
        }
    }

    /**
     * queryGlobalConsumersNum
     *
     * @param streamName the stream name
     * @return long the consumers num
     * @throws YRException if there is an exception during quering global consumersNum
     */
    @Override
    public long queryGlobalConsumersNum(String streamName) throws YRException {
        rLock.lock();
        try {
            return LibRuntime.QueryGlobalConsumersNum(streamName);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        } finally {
            rLock.unlock();
        }
    }

    /**
     * Get instance route.
     *
     * @param objectId the object id
     * @return the string representing the instance route
     * @throws YRException the YR exception.
     */
    @Override
    public String getInstanceRoute(String objectId) throws YRException {
        try {
            return LibRuntime.GetInstanceRoute(objectId);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        }
    }

    /**
     * Save instance route.
     *
     * @param objectId the object id
     * @param instanceRoute the instance route
     * @throws YRException the YR exception.
     */
    @Override
    public void saveInstanceRoute(String objectId, String instanceRoute) throws YRException {
        try {
            LibRuntime.SaveInstanceRoute(objectId, instanceRoute);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        }
    }

    /**
     * Sync terminate instance.
     *
     * @param instanceId the instance id
     * @throws YRException the YR exception.
     */
    @Override
    public void terminateInstanceSync(String instanceId) throws YRException {
        ErrorInfo errorInfo = LibRuntime.KillSync(instanceId);
        Utils.checkErrorAndThrow(errorInfo, "kill instance sync(" + instanceId + ")");
        LOGGER.info("succeed to terminate instance sync({})", instanceId);
    }

    /**
     * Get node information in the cluster.
     *
     * @return List<Node>: node information
     * @throws YRException the actor task exception.
     */
    @Override
    public List<Node> nodes() throws YRException {
        Pair<ErrorInfo, List<Node>> res;
        try {
            res = LibRuntime.nodes();
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        }
        Utils.checkErrorAndThrow(res.getFirst(), "get node information");
        return res.getSecond();
    }
}
