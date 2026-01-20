/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
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

package org.yuanrong.executor;

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.errorcode.Pair;
import org.yuanrong.exception.YRException;
import org.yuanrong.exception.handler.filter.FilterFactory;
import org.yuanrong.exception.handler.traceback.StackTraceInfo;
import org.yuanrong.libruntime.generated.Libruntime;
import org.yuanrong.libruntime.generated.Libruntime.ApiType;
import org.yuanrong.libruntime.generated.Libruntime.FunctionMeta;
import org.yuanrong.runtime.config.RuntimeContext;
import org.yuanrong.runtime.util.Utils;
import org.yuanrong.runtime.util.ExtClasspathLoader;
import org.yuanrong.serialization.Serializer;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.exc.MismatchedInputException;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.io.PrintWriter;
import java.io.StringWriter;
import java.lang.reflect.Constructor;
import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;
import java.nio.ByteBuffer;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * FunctionHandler
 *
 * @since 2024/08/15
 */
public class FunctionHandler implements HandlerIntf {
    /**
     * String (clzName) => Object (instance)
     */
    public static final Map<String, Class<?>> classCache = new HashMap<String, Class<?>>();

    /**
     * String (clzName) => String (funcName) => Method (method id)
     */
    public static final Map<String, Map<String, Method>> methodCache = new HashMap<String, Map<String, Method>>();

    /**
     * CONSTRUCTOR_FUNC_NAME init value
     */
    public static final String CONSTRUCTOR_FUNC_NAME = "<init>";

    private static final Logger LOG = LoggerFactory.getLogger(FunctionHandler.class);

    private static final String USER_SHUTDOWN_FUNC_NAME = "yrShutdown";

    private static final String USER_SHUTDOWN_FUNC_SIGNATURE = "(I)V";

    private static final String USER_RECOVER_FUNC_NAME = "yrRecover";

    private static final String USER_RECOVER_FUNC_SIGNATURE = "()V";

    private static Object instance = null;

    private static String instanceClassName = "";

    private static Class<?> getClassThroughCache(String clzName) throws ClassNotFoundException {
        Class<?> res = classCache.get(clzName);
        if (res != null) {
            return res;
        }
        Class<?> clz = ExtClasspathLoader.getFunctionClassLoader().loadClass(clzName);
        classCache.put(clzName, clz);
        return clz;
    }

    /**
     * Execute function
     *
     * @param meta the meta
     * @param type the type
     * @param args the args
     * @return the return value, in ByteBuffer type, may need to release bytebuffer
     * `buffer.clear()`
     * @throws Exception the exception
     */
    @Override
    public ReturnType execute(FunctionMeta meta, Libruntime.InvokeType type, List<ByteBuffer> args) throws Exception {
        LOG.info("executing udf methods, current type: {}", type);

        switch (type) {
            case CreateInstance:
                // $FALL-THROUGH$
            case CreateInstanceStateless:
                LOG.debug("Invoking udf method matched, case create");
                return create(meta, args);
            case InvokeFunction:
                // $FALL-THROUGH$
            case InvokeFunctionStateless:
                LOG.debug("Invoking udf method matched, case invoke");
                return invoke(meta, args);
            default:
                LOG.debug("Invoking udf method matched, case dft");
                return new ReturnType(ErrorCode.ERR_INCORRECT_INVOKE_USAGE, "invalid invoke type");
        }
    }

    /**
     * Shutdown the instance gracefully.
     *
     * @param gracePeriodSeconds the time to wait for the instance to shutdown gracefully.
     * @return ErrorInfo, the ErrorInfo of the execution of shutdown function.
     */
    @Override
    public ErrorInfo shutdown(int gracePeriodSeconds) {
        if (instance == null) {
            return new ErrorInfo(ErrorCode.ERR_INNER_SYSTEM_ERROR, ModuleCode.RUNTIME,
                "Failed to invoke instance function [" + USER_SHUTDOWN_FUNC_NAME
                    + "], instance has not been initialized");
        }
        Class<?> clz = null;
        Method method = null;
        try {
            LOG.debug("Start to get shutdown method {} from class {}", USER_SHUTDOWN_FUNC_NAME, instanceClassName);
            clz = getClassThroughCache(instanceClassName);
            Class<?>[] paramTypes = Utils.getParameterTypeFromSignature(clz, USER_SHUTDOWN_FUNC_NAME,
                USER_SHUTDOWN_FUNC_SIGNATURE);
            method = clz.getMethod(USER_SHUTDOWN_FUNC_NAME, paramTypes);
        } catch (ClassNotFoundException | NoSuchMethodException | IllegalArgumentException e) {
            LOG.debug("Cannot get {} from the class cache, exception: {}", USER_SHUTDOWN_FUNC_NAME,
                getCausedByString(e));
            return new ErrorInfo();
        }
        LOG.debug("Succeeded to get shutdown method {} from class {}", USER_SHUTDOWN_FUNC_NAME, instanceClassName);

        try {
            method.invoke(instance, gracePeriodSeconds);
        } catch (IllegalAccessException | IllegalArgumentException | InvocationTargetException e) {
            LOG.debug("Failed to invoke user {} function, exception: {}", USER_SHUTDOWN_FUNC_NAME,
                getCausedByString(e));
            return new ErrorInfo(ErrorCode.ERR_INNER_SYSTEM_ERROR, ModuleCode.RUNTIME,
                "Failed to invoke user " + USER_SHUTDOWN_FUNC_NAME + " function");
        }

        LOG.debug("Succeeded to invoke user {} function", USER_SHUTDOWN_FUNC_NAME);
        return new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.RUNTIME, "");
    }

    /**
     * Serializes the instance of the CodeExecutor class and returns a Pair
     * containing the serialized byte arrays of the instance and the class name.
     *
     * @param instanceID the ID of the instance to be dumped
     * @return a Pair containing the serialized byte arrays of the instance and the
     * class name
     * @throws JsonProcessingException if there is an error during serialization
     */
    @Override
    public Pair<byte[], byte[]> dumpInstance(String instanceID) throws JsonProcessingException {
        byte[] instanceBytes = Serializer.serialize(FunctionHandler.instance);
        byte[] clzNameBytes = Serializer.serialize(FunctionHandler.instanceClassName);
        LOG.debug("Succeeded to DumpInstance(clzName: {}), instanceID: {}", FunctionHandler.instanceClassName,
            instanceID);
        return new Pair<>(instanceBytes, clzNameBytes);
    }

    /**
     * Loads an instance of a class from serialized byte arrays.
     *
     * @param instanceBytes the serialized byte array representing the instance
     * @param clzNameBytes the serialized byte array representing the class name
     * @throws IOException if there is an error reading the byte arrays
     * @throws ClassNotFoundException if the class specified by the class name is not found
     */
    @Override
    public void loadInstance(byte[] instanceBytes, byte[] clzNameBytes) throws IOException, ClassNotFoundException {
        String clzName = (String) Serializer.deserialize(clzNameBytes, String.class);
        Class<?> clz = getClassThroughCache(clzName);
        Object inst;
        try {
            inst = Serializer.deserialize(instanceBytes, clz);
        } catch (MismatchedInputException e) {
            throw new IOException("Catch MismatchedInputException while running LoadInstance, please check "
                + "whether the getter/setter methods in user class comply with the Java Beans naming rules. "
                + "Exception:", e);
        }
        LOG.debug("Succeeded to LoadInstance, clzName: {}", clzName);
        instance = inst;
        instanceClassName = clzName;
    }

    /**
     * Recover the instance.
     *
     * @return ErrorInfo, the ErrorInfo of the execution of recover function.
     */
    @Override
    public ErrorInfo recover() {
        if (instance == null) {
            return new ErrorInfo(ErrorCode.ERR_INNER_SYSTEM_ERROR, ModuleCode.RUNTIME,
                "Failed to invoke instance function [" + USER_RECOVER_FUNC_NAME
                    + "], instance has not been initialized");
        }
        Class<?> clz = null;
        Method method = null;
        try {
            LOG.debug("Start to get recover method {} from class {}", USER_RECOVER_FUNC_NAME, instanceClassName);
            clz = getClassThroughCache(instanceClassName);
            Class<?>[] paramTypes = Utils.getParameterTypeFromSignature(clz, USER_RECOVER_FUNC_NAME,
                USER_RECOVER_FUNC_SIGNATURE);
            method = clz.getMethod(USER_RECOVER_FUNC_NAME, paramTypes);
        } catch (ClassNotFoundException | NoSuchMethodException | IllegalArgumentException e) {
            LOG.debug("Cannot get {} from the class cache, exception: {}", USER_RECOVER_FUNC_NAME,
                getCausedByString(e));
            return new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.RUNTIME, "");
        }
        LOG.debug("Succeeded to get recover method {} from class {}", USER_RECOVER_FUNC_NAME, instanceClassName);

        try {
            method.invoke(instance);
        } catch (IllegalAccessException | IllegalArgumentException | InvocationTargetException e) {
            LOG.debug("Failed to invoke user {} function, exception: {}", USER_RECOVER_FUNC_NAME,
                getCausedByString(e));
            return new ErrorInfo(ErrorCode.ERR_INNER_SYSTEM_ERROR, ModuleCode.RUNTIME,
                "Failed to invoke user " + USER_RECOVER_FUNC_NAME + " function");
        }

        LOG.debug("Succeeded to invoke user {} function", USER_RECOVER_FUNC_NAME);
        return new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.RUNTIME, "");
    }

    /**
     * Create ReturnType
     *
     * @param meta the meta
     * @param args the args
     * @return ReturnType
     */
    private static ReturnType create(FunctionMeta meta, List<ByteBuffer> args) {
        String clzName = meta.getClassName();
        String funcName = meta.getFunctionName();
        String signature = meta.getSignature();
        LOG.info("Creating for udf methods, clz({}), func({}), sig({})", clzName, funcName, signature);
        try {
            constructObject(clzName, funcName, signature, args);
            if (ApiType.Faas.equals(meta.getApiType()) && !CONSTRUCTOR_FUNC_NAME.equals(funcName)) {
                return invoke(meta, args);
            }
        } catch (ClassNotFoundException | InstantiationException | IllegalAccessException | IllegalArgumentException
                 | InvocationTargetException | IOException e) {
            LOG.error("failed to execute udf methods, clz({}), func({}), sig({}), reason: {}", clzName, funcName,
                signature, e);
            return new ReturnType(ErrorCode.ERR_USER_FUNCTION_EXCEPTION, ModuleCode.RUNTIME_CREATE,
                "failed to create instance due to the cause: \n" + getCausedByString(e));
        }
        LOG.info("successfully finish the create job with clz({}), func({}), sig({})", clzName, funcName, signature);
        return new ReturnType(ErrorCode.ERR_OK, "");
    }

    /**
     * construct an object, and cache it in the `instance`
     *
     * @param clzName the class name, required
     * @param funcName the function name, if is <init>, will use constructors
     * @param signature the constructor signature, optional, will use default
     * constructor if missing
     * @param args the constructor args, will be passed to the constructor
     * @return the new object
     * @throws ClassNotFoundException the ClassNotFoundException
     * @throws InstantiationException the InstantiationException
     * @throws IllegalAccessException the IllegalAccessException
     * @throws IllegalArgumentException the IllegalArgumentException
     * @throws InvocationTargetException the InvocationTargetException
     * @throws IOException the IOException
     */
    private static Object constructObject(String clzName, String funcName, String signature, List<ByteBuffer> args)
        throws ClassNotFoundException, InstantiationException, IllegalAccessException, IllegalArgumentException,
               InvocationTargetException, IOException {
        Class<?> clz = ExtClasspathLoader.getFunctionClassLoader().loadClass(clzName);
        LOG.debug("constructing clz({}), func({}), sig({}), get class ok: {}", clzName, funcName, signature, clz);
        // Match constructor
        Constructor<?> constructor = null;
        for (Constructor<?> cs : clz.getConstructors()) {
            // if funcName is not a constructor, will just use the default one
            // if signature missed, use the first one
            LOG.debug("Creating for udf methods, current cs: {}", cs);
            if (signature == null || signature.isEmpty() || !CONSTRUCTOR_FUNC_NAME.equals(funcName) || signature.equals(
                Utils.getMethodSignature(cs))) {
                constructor = cs;
                break;
            }
        }
        // failed to find anyone match the requirements
        if (constructor == null) {
            LOG.error("Failed to get constructor");
            throw new RuntimeException("failed to find constructor with specified signature: (" + signature + ")");
        }

        LOG.debug("successfully get constructor: {}", constructor);
        Object[] objArgs = Utils.deserializeObjects(Arrays.asList(constructor.getParameterTypes()), args).toArray();
        LOG.debug("successfully deserialize objects: {}", objArgs.length);

        // deserialize the params to objects, and new instance
        Object inst = constructor.newInstance(objArgs);
        LOG.debug("successfully construct obj: {}", inst);
        FunctionHandler.instance = inst;
        FunctionHandler.instanceClassName = clzName;
        return inst;
    }

    private static ReturnType invoke(FunctionMeta meta, List<ByteBuffer> args) {
        LOG.debug("Invoking udf methods, meta: {}", meta);
        String clzName = meta.getClassName();
        String funcName = meta.getFunctionName();
        String signature = meta.getSignature();
        LOG.debug("Invoking udf methods, clz({}), func({}), sig({})", clzName, funcName, signature);
        Class<?>[] paramTypes = null;
        Method method = null;

        try {
            // New such class in cache, just new hash map
            if (!methodCache.containsKey(clzName)) {
                methodCache.put(clzName, new HashMap<String, Method>());
            }
            // No such function in cache, new function
            if (!methodCache.get(clzName).containsKey(funcName)) {
                Class<?> clz = getClassThroughCache(clzName);
                paramTypes = Utils.getParameterTypeFromSignature(clz, funcName, signature);
                method = clz.getMethod(funcName, paramTypes);
                methodCache.get(clzName).put(funcName, method);
            }
            method = method == null ? methodCache.get(clzName).get(funcName) : method;
            paramTypes = paramTypes == null ? Utils.getParameterTypeFromSignature(getClassThroughCache(clzName),
                funcName, signature) : paramTypes;
            if (instance == null) {
                throw new RuntimeException(
                    "failed to find instance object when invoke method (" + meta.toString() + ")");
            }
            List<Object> argObjs = Utils.deserializeObjects(Arrays.asList(paramTypes), args);
            Object retValue = method.invoke(instance, argObjs.toArray());
            LOG.debug("Succeeded to invoke method, class({}), function name({}).", clzName, funcName);
            return new ReturnType(ErrorCode.ERR_OK, Serializer.serialize(retValue));
        } catch (ClassNotFoundException | IllegalAccessException | IllegalArgumentException
                 | InvocationTargetException | IOException | NoSuchMethodException e) {
            return processException(e, clzName, funcName);
        }
    }

    private static ReturnType processException(Exception e, String clzName, String funcName) {
        LOG.debug("Get exception when invoke for {}.{}, exception type:{}", clzName, funcName, e.getClass().getName());
        // if throwable is not InvocationTargetException, consider it as inner system exception
        if (!(e instanceof InvocationTargetException)) {
            LOG.error("Get not InvocationTargetException exception when invoke for {}.{}, exception:{}, message:{}",
                clzName, funcName, e.toString(), getCausedByString(e));
            return new ReturnType(ErrorCode.ERR_INNER_SYSTEM_ERROR, ModuleCode.RUNTIME_INVOKE, getCausedByString(e));
        }
        InvocationTargetException targetException = (InvocationTargetException) e;
        Throwable throwable = targetException.getTargetException();
        LOG.debug("Get throwable after peel off InvocationTargetException when invoke for {}.{}, {}", clzName, funcName,
            throwable.toString());

        // if throwable is not YRException, consider it as user code origin exception
        if (!(throwable instanceof YRException)) {
            FilterFactory filterFactory = new FilterFactory(throwable, true);
            StackTraceInfo stackTraceInfo = filterFactory.generateStackTraceInfo(clzName, funcName);
            LOG.error("Get origin exception when invoke for {}.{}, throwable:{}, stackTraceInfo:{}", clzName, funcName,
                throwable, stackTraceInfo.toString());
            List<StackTraceInfo> stackTraces = new ArrayList<StackTraceInfo>();
            stackTraces.add(stackTraceInfo);
            String errorMsg = "exception occurred when processing user code, runtimeId: " + RuntimeContext.runtimeId
                + ", throwable: " + throwable;
            return new ReturnType(ErrorCode.ERR_USER_FUNCTION_EXCEPTION, ModuleCode.RUNTIME_INVOKE, errorMsg,
                stackTraces);
        }

        // if throwable is YRException, and it is also an user function exception, then process it
        LOG.error("Get throwable from YRException when invoke for {}.{}, throwable:{}", clzName, funcName,
            throwable.toString());
        if (ErrorCode.ERR_USER_FUNCTION_EXCEPTION.equals(((YRException) throwable).getErrorCode())) {
            LOG.error("Get YRException exception when invoke for {}.{}, exception class:{}", clzName, funcName,
                e.getClass().getName());
            YRException yrException = (YRException) throwable;
            // if Exception is YRException and exception is caused by user code, then get exceptions in it and
            // join stacktrace of YRException and inner exception stacktraces
            LOG.debug("Get exceptions stackTraces:{} in YRException when invoke for {}.{}",
                Arrays.asList(yrException.getStackTrace()).toArray(), clzName, funcName);
            FilterFactory filterFactory = new FilterFactory(yrException.getCause(), false);
            StackTraceInfo stackTraceInfo = filterFactory.generateStackTraceInfo(clzName, funcName);
            LOG.debug("Get {}.{} result, YRException stackTraces:{}, innerExp.stackTraceInfo:{}, ", clzName,
                funcName, Arrays.asList(yrException.getStackTrace()).toArray(), stackTraceInfo.toString());
            List<StackTraceElement> finalElements = appendStackTraceArrayToList(stackTraceInfo.getStackTraceElements(),
                yrException.getStackTrace());

            LOG.debug("print final element:{} for {}.{}", finalElements, clzName, funcName);
            stackTraceInfo.setStackTraceElements(finalElements);
            List<StackTraceInfo> stackTraces = new ArrayList<StackTraceInfo>();
            stackTraces.add(stackTraceInfo);
            String errorMsg = "exception occurred when processing user code, runtimeId: "
                    + RuntimeContext.runtimeId + ", throwable: " + throwable;
            return new ReturnType(ErrorCode.ERR_USER_FUNCTION_EXCEPTION, ModuleCode.RUNTIME_INVOKE, errorMsg,
                stackTraces);
        }

        return new ReturnType(ErrorCode.ERR_USER_FUNCTION_EXCEPTION, ModuleCode.RUNTIME_INVOKE,
            getCausedByString(throwable));
    }

    private static List<StackTraceElement> appendStackTraceArrayToList(List<StackTraceElement> list,
        StackTraceElement[] stackTraceArray) {
        if (list.isEmpty()) {
            return new ArrayList<>(Arrays.asList(stackTraceArray));
        }
        if (stackTraceArray.length > 0) {
            List<StackTraceElement> stackTraceElementList = new ArrayList<>(Arrays.asList(stackTraceArray));
            list.addAll(stackTraceElementList);
            return list;
        }
        return new ArrayList<>();
    }

    private static String getCausedByString(Throwable throwable) {
        StringWriter stringWriter = new StringWriter();
        PrintWriter printWriter = new PrintWriter(stringWriter);
        throwable.printStackTrace(printWriter);
        return stringWriter.toString();
    }
}