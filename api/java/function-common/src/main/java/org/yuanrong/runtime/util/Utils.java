/*
 *
 *  Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package org.yuanrong.runtime.util;

import org.yuanrong.api.InvokeArg;
import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.exception.YRException;
import org.yuanrong.exception.HandlerNotAvailableException;
import org.yuanrong.serialization.Serializer;

import com.google.gson.Gson;

import org.apache.commons.lang3.exception.ExceptionUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.objectweb.asm.Type;

import java.io.IOException;
import java.lang.reflect.Constructor;
import java.lang.reflect.Field;
import java.lang.reflect.Method;
import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.UUID;
import java.util.concurrent.TimeUnit;

/**
 * The type Utils.
 *
 * @since 2023 /08/30
 */
public class Utils {
    /**
     * The constant REFLECT_REGEX.
     */
    public static final String REFLECT_REGEX = ".reflect.";

    /**
     * The key of IAM authorization token in POSIX requests metadata.
     */
    public static final String HEADER_AUTHORIZATION = "Authorization";

    private static final Logger LOGGER = LoggerFactory.getLogger(Utils.class);

    private static final String TRACE_ID = "-trace-";

    private static final String TASK_CONNECTOR_SYMBOL = "-task-";

    private static final String SPLIT_SYMBOL_MINUS = "-";

    private static final int DESIGNATED_INSTANCE_ID_MAX_LENGTH = 128;

    private static final int MAX_STACK_TRACE_LENGTH = 10000;

    private static final char DOT_SEPARATOR = '.';

    private static final String UDF_METHOD_ILLEGAL = "udf method illegal";

    private static final String FAILED_TO_SPLIT_HANDLER_METHOD_NAME = "failed to split handler method name.";

    private static final String REGEX_BEFORE = "((\\w+\\.?)+\\w+)\\.(\\w+)";

    private static final String REGEX_AFTER = "(\\w+\\.java):(\\d+)";

    private static final String RUNTIME_CLASS_PATH = "org.yuanrong";

    private static final String SUPPRESSED_EXCEPTION_BEGIN_TAG = "Caused by:";

    private static final String RUNTIME_MORE_INFO_SIGNAL = "...";

    private static final String ACTOR_TASK_STR = "YRException";

    private static final String AT_STR = "at";

    private static final String AT_PATTERN = "^\\tat.*$";

    private static final String DOUBLE_COLON_SEPARATOR = "::";

    private static final String USER_CODE_CLASS_NOT_FOUND = "user code class not exist ";

    /**
     * The gson.
     */
    private static Gson gson = new Gson();

    /**
     * Is empty str boolean.
     *
     * @param str the str
     * @return the boolean
     */
    public static boolean isEmptyStr(String str) {
        return str == null || str.isEmpty();
    }

    /**
     * Generate uuid string.
     *
     * @return the string
     */
    public static String generateUUID() {
        return UUID.randomUUID().toString().replaceAll("-", "");
    }

    /**
     * Generate trace id string.
     *
     * @param jobId the job id
     * @return the string
     */
    public static String generateTraceId(String jobId) {
        return jobId + TRACE_ID + UUID.randomUUID().toString().replaceAll("-", "");
    }

    /**
     * Valid designated instance id boolean.
     *
     * @param designatedInstanceID the designated instance id
     * @return the boolean
     */
    public static boolean invalidDesignatedInstanceID(String designatedInstanceID) {
        return designatedInstanceID == null || designatedInstanceID.length() >= DESIGNATED_INSTANCE_ID_MAX_LENGTH;
    }

    /**
     * Generate job id string.
     *
     * @return the string
     */
    public static String generateCloudJobId() {
        // Generate job id when YR.init() executes
        return "job-" + UUID.randomUUID().toString().replace("-", "").substring(0, 8);
    }

    /**
     * Generate cloud runtime id string.
     *
     * @param jobId     the job id
     * @param runtimeId the runtime id
     * @param orderNum  the order num
     * @return the string
     */
    public static String generateRequestId(String jobId, String runtimeId, int orderNum) {
        return jobId + TASK_CONNECTOR_SYMBOL + UUID.randomUUID() + SPLIT_SYMBOL_MINUS
                + runtimeId + SPLIT_SYMBOL_MINUS + orderNum;
    }

    /**
     * Gets method signature.
     *
     * @param method the method
     * @return the method signature
     */
    public static String getMethodSignature(Method method) {
        return Type.getType(method).getDescriptor();
    }

    /**
     * Gets method signature.
     *
     * @param method the method
     * @return the method signature
     */
    public static String getMethodSignature(Constructor<?> method) {
        return Type.getType(method).getDescriptor();
    }

    /**
     * Get parameter type class [ ].
     *
     * @param uClass          the user class
     * @param methodName      the method name
     * @param methodSignature the method signature
     * @return the parameter type class[]
     * @throws IllegalArgumentException the illegal argument exception
     */
    public static Class<?>[] getParameterTypeFromSignature(Class<?> uClass, String methodName, String methodSignature)
            throws IllegalArgumentException {
        Method[] methods = uClass.getMethods();
        Class<?>[] parameterTypes = null;
        for (Method method : methods) {
            String name = method.getName();
            String signature = getMethodSignature(method);
            // When the cpp function invokes the java function,
            // the cpp function does not transfer the methodSignature.
            // The methodSignature would be an empty string.
            if (methodName.equals(name) && (methodSignature.equals(signature) || methodSignature.isEmpty())) {
                parameterTypes = method.getParameterTypes();
                break;
            }
        }
        if (Objects.isNull(parameterTypes)) {
            String errorMsg = new StringBuilder("Failed to find user-definded method: [")
                    .append(uClass.getName())
                    .append(".")
                    .append(methodName)
                    .append(" signature: ")
                    .append(methodSignature)
                    .append("] from class ")
                    .append(uClass.getName())
                    .append(" loaded in runtime.")
                    .toString();
            throw new IllegalArgumentException(errorMsg);
        }
        return parameterTypes;
    }

    /**
     * split class name and method.
     *
     * @param handler handler string
     * @return class name and method string[]
     * @throws HandlerNotAvailableException the handler not available exception
     */
    public static String[] splitMethodName(String handler) throws HandlerNotAvailableException {
        if (Objects.isNull(handler) || handler.isEmpty()) {
            throw new HandlerNotAvailableException(new IllegalArgumentException(FAILED_TO_SPLIT_HANDLER_METHOD_NAME));
        }
        int lastIndex = handler.lastIndexOf(DOT_SEPARATOR);
        if (lastIndex == -1) {
            throw new HandlerNotAvailableException(new IllegalArgumentException(FAILED_TO_SPLIT_HANDLER_METHOD_NAME));
        }
        return new String[] {handler.substring(0, lastIndex), handler.substring(lastIndex + 1)};
    }

    /**
     * Deserializ objects list.
     *
     * @param types   the types
     * @param rawArgs the raw args
     * @return the list
     * @throws IOException the io exception
     */
    // JNI layer uses direct byte buffer, which copys a C++ address to the Java
    // object, so if want, zero-copy is possible.
    // But right now, Java deserialization requires `byte[]` as raw data, so there
    // will be data copy in Java code for at least once.
    public static List<Object> deserializeObjects(List<Class<?>> types, List<ByteBuffer> rawArgs) throws IOException {
        int len = Math.min(types.size(), rawArgs.size());
        List<Object> objs = new ArrayList<Object>(len);
        for (int i = 0; i < len; i++) {
            byte[] bfBytes = new byte[rawArgs.get(i).limit()];
            rawArgs.get(i).get(bfBytes);
            objs.add(Serializer.deserialize(bfBytes, types.get(i)));
        }
        return objs;
    }

    /**
     * cast instance to string type.
     *
     * @param obj the object
     * @return string type data
     */
    public static String getObjectString(Object obj) {
        if (obj instanceof String) {
            return (String) obj;
        }
        return "";
    }

    /**
     * cast instance to map type.
     *
     * @param obj the object
     * @return map type data
     */
    public static Map<?, ?> getObjectMap(Object obj) {
        if (obj instanceof Map) {
            return (Map<?, ?>) obj;
        }
        return gson.fromJson(getObjectString(obj), Map.class);
    }

    /**
     * get exception stackTrace msg
     *
     * @param throwable exp
     * @return stack processed exception msg
     */
    public static String getProcessedExceptionMsg(Throwable throwable) {
        String expMsg = ExceptionUtils.getStackTrace(throwable);
        if (expMsg.length() > MAX_STACK_TRACE_LENGTH) {
            return expMsg.substring(0, MAX_STACK_TRACE_LENGTH);
        }
        return expMsg;
    }

    /**
     * The put method retry 3 times.
     *
     * @param objectId        the object ref ids
     * @param putData         the id of object ref
     * @param nestedObjectIds the nestedObjectIds that current ObjectRef depends on.
     * @param client          dataSystem client
     * @param interval        interval time to continuously put when spill may occur    in milliseconds
     * @param retryTime       retry time to continuously put when spill may occur
     * @return the boolean
     */
    public static boolean put(
            String objectId,
            ByteBuffer putData,
            List<String> nestedObjectIds,
            Object client,
            long interval,
            long retryTime) {
        return true;
    }

    /**
     * Get list retry 3 times.
     *
     * @param ids     the object ref ids
     * @param timeout the timeout
     * @param client  the client
     * @return Result : the list which contains all objects got from DataSystem If
     * some objectRef   failed to get its object, elements in the corresponding positions in
     * the Result is null
     * @throws YRException exception
     */
    public static List<?> get(List<String> ids, int timeout, Object client) throws YRException {
        throw new YRException(ErrorCode.ERR_GET_OPERATION_FAILED, ModuleCode.DATASYSTEM,
                "failed to retry get from dataSystem");
    }

    /**
     * Sleep seconds.
     *
     * @param timeout the timeout
     */
    public static void sleepSeconds(int timeout) {
        try {
            TimeUnit.SECONDS.sleep(timeout);
        } catch (InterruptedException e) {
            LOGGER.error("wait interrupted ", e);
        }
    }

    /**
     * Reads environment value regarding to a given key. DefaultValue would be
     * returned if environment value is null or an empty string.
     *
     * @param envKey       the given key for reading environment value.
     * @param defaultValue the defalut value returned if failed to read then                     environment value.
     * @param logPrefix    the prefix of logs printed by this method. It can be an                     empty String.
     * @return the environment value String.
     */
    public static String getEnvWithDefualtValue(String envKey, String defaultValue, String logPrefix) {
        String value = System.getenv(envKey);
        if (value == null || value.isEmpty()) {
            String valDesc = value == null ? "null" : "empty string";
            LOGGER.debug(
                    "{}Failed to read the value of environment key ({}), value is {}. Use default value: ({}: {}).",
                    logPrefix, envKey, valDesc, envKey, defaultValue);
            return defaultValue;
        }

        LOGGER.debug("{}Succeeded to read environment key-value pair({}: {}).", logPrefix, envKey, value);
        return value;
    }

    /**
     * checkErrorAndThrow for errorInfo
     *
     * @param errorInfo the errorInfo
     * @param msg the msg
     * @throws YRException the YR exception
     */
    public static void checkErrorAndThrow(ErrorInfo errorInfo, String msg) throws YRException {
        if (!ErrorCode.ERR_OK.equals(errorInfo.getErrorCode())) {
            LOGGER.error("{} occurs exception {}", msg, errorInfo);
            throw new YRException(errorInfo);
        }
    }

    /**
     *  getFromJavaEnv
     *
     * @param key the key
     * @return String
     */
    public static String getFromJavaEnv(String key) {
        String res = "";
        String filedName = "theUnmodifiableEnvironment";
        try {
            Class<?> cls = Class.forName("java.lang.ProcessEnvironment");
            // get field and access
            Field oldFiled = cls.getDeclaredField(filedName);
            oldFiled.setAccessible(true);
            // get Filed map
            Object map = oldFiled.get(null);
            Class<?> unmodifiableMap = Class.forName("java.util.Collections$UnmodifiableMap");
            Field field = unmodifiableMap.getDeclaredField("m");
            field.setAccessible(true);
            Object obj = field.get(map);
            res = ((Map<String, String>) obj).get(key);
        } catch (ReflectiveOperationException e) {
            LOGGER.error("get field: {} has an error: {}", filedName, e);
        }
        return res;
    }

    /**
     * Pack the arguments to be invoked into a list of InvokeArg objects.
     *
     * @param args The arguments to be packed.
     * @return A list of InvokeArg objects.
     */
    public static List<InvokeArg> packFaasInvokeArgs(Object... args) {
        List<InvokeArg> invokeArgs = new ArrayList<InvokeArg>();
        for (Object arg : args) {
            InvokeArg invokeArg;
            invokeArg = new InvokeArg(gson.toJson(arg).getBytes(StandardCharsets.UTF_8));
            invokeArg.setObjectRef(false);
            invokeArg.setNestedObjects(new HashSet<>());
            invokeArgs.add(invokeArg);
        }
        return invokeArgs;
    }

    /**
     * get method of class
     *
     * @param clazz Class<?>
     * @param methodName String
     * @return Method
     * @throws NoSuchMethodException Exception
     */
    public static Method getMethod(Class<?> clazz, String methodName) throws NoSuchMethodException {
        Method specifiedMethod = null;
        Method[] methods = clazz.getDeclaredMethods();
        for (Method method : methods) {
            if (methodName.equals(method.getName())) {
                specifiedMethod = method;
                break;
            }
        }
        if (specifiedMethod == null) {
            throw new NoSuchMethodException("cannot find such method: " + methodName);
        }
        return specifiedMethod;
    }

    /**
     * Get User Code Entry class and method
     *
     * @param userCodeEntry user code string
     * @param isInitialize isInitialize
     * @return [class, method]
     * @throws NoSuchMethodException user code not found class or entry method
     */
    public static String[] splitUserClassAndMethod(String userCodeEntry, boolean isInitialize)
            throws NoSuchMethodException {
        if (userCodeEntry == null || userCodeEntry.isEmpty()) {
            throw new NoSuchMethodException(USER_CODE_CLASS_NOT_FOUND);
        }
        return splitEntryWithSeparators(userCodeEntry,
                new String[] {DOUBLE_COLON_SEPARATOR, String.valueOf(DOT_SEPARATOR)});
    }

    /**
     * Get User Code Entry class and method with separator
     *
     * @param userCodeEntry user code string
     * @param separator separator in [., ::]
     * @return [class, method]
     * @throws NoSuchMethodException user code not found separator
     */
    public static String[] splitEntryWithSeparators(String userCodeEntry, String[] separator)
            throws NoSuchMethodException {
        for (String sep : separator) {
            int lastIndex = userCodeEntry.lastIndexOf(sep);
            if (lastIndex == -1) {
                continue;
            }
            if (String.valueOf(DOT_SEPARATOR).equals(sep)) {
                return new String[]{userCodeEntry.substring(0, lastIndex),
                        userCodeEntry.substring(lastIndex + 1)};
            } else if (DOUBLE_COLON_SEPARATOR.equals(sep)) {
                // classPath::Method  lastIndex must plus 2
                return new String[]{userCodeEntry.substring(0, lastIndex),
                        userCodeEntry.substring(lastIndex + 2)};
            } else {
                throw new NoSuchMethodException("user class and method separator invalid");
            }
        }
        throw new NoSuchMethodException("cannot separate user entry: " + userCodeEntry);
    }
}