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

package org.yuanrong.utils;

import org.yuanrong.Config;
import org.yuanrong.ConfigManager;
import org.yuanrong.ExistenceOpt;
import org.yuanrong.FunctionWrapper;
import org.yuanrong.MSetParam;
import org.yuanrong.api.InvokeArg;
import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.exception.YRException;
import org.yuanrong.jni.LibRuntimeConfig;
import org.yuanrong.libruntime.generated.Libruntime;
import org.yuanrong.runtime.client.ObjectRef;
import org.yuanrong.runtime.util.Utils;
import org.yuanrong.serialization.Serializer;

import com.fasterxml.jackson.core.JsonProcessingException;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import java.util.stream.Stream;

/**
 * The type Utils.
 *
 * @since 2023/12/04
 */
public class SdkUtils {
    /**
     * The Pattern.
     */
    static final Pattern IP_V4_PATTERN =
        Pattern.compile(
            "^((\\d|[1-9]\\d|1\\d\\d|2[0-4]\\d|25[0-5]"
                + "|[*])\\.){3}(\\d|[1-9]\\d|1\\d\\d|2[0-4]\\d|25[0-5]|[*])$");

    /**
     * The Urn item num.
     */
    static final int URN_ITEM_NUM = 7;

    /**
     * The Type map.
     */
    static HashMap<Class<?>, Class<?>> typeMap = new HashMap<>();

    private static final Logger LOGGER = LoggerFactory.getLogger(SdkUtils.class);

    private static final String YUANRONG_REGEX = ".yuanrong.";

    private static final Boolean IS_ENABLE_DIS_CONV_CALL_STACK = Boolean.parseBoolean(
        System.getenv("ENABLE_DIS_CONV_CALL_STACK"));

    private static final String OBJECT_REF_PREFIX = "yr-api-obj-";

    private static final int FUNCTION_URN_LEASER_INDEX = 3;

    private static final int FUNCTION_URN_FUNCTION_INDEX = 5;

    private static final int FUNCTION_URN_VERSION_INDEX = 6;

    private static final int FUNCTION_URN_LENGTH = 7;

    static {
        typeMap.put(int.class, Integer.class);
        typeMap.put(long.class, Long.class);
        typeMap.put(byte.class, Byte.class);
        typeMap.put(short.class, Short.class);
        typeMap.put(float.class, Float.class);
        typeMap.put(double.class, Double.class);
        typeMap.put(char.class, Character.class);
        typeMap.put(boolean.class, Boolean.class);
    }

    /**
     * Check ip boolean.
     *
     * @param ip the ip
     * @return the boolean
     */
    public static boolean checkIP(String ip) {
        Matcher matcher = IP_V4_PATTERN.matcher(ip);
        return matcher.matches();
    }

    /**
     * Check urn boolean.
     *
     * @param urn the urn
     * @return the boolean
     */
    public static boolean checkURN(String urn) {
        return urn.split(":").length == URN_ITEM_NUM;
    }

    /**
     * Generate job id string.
     *
     * @return the string
     */
    public static String generateJobId() {
        // Generate job id when YR.init() executes
        return "job-" + UUID.randomUUID().toString().replace("-", "").substring(0, 8);
    }

    /**
     * Generate runtime id string.
     *
     * @return the string
     */
    public static String generateRuntimeId() {
        // Generate job id when YR.init() executes
        return UUID.randomUUID().toString().replace("-", "").substring(0, 8);
    }

    /**
     * Gets object ref id.
     *
     * @return the object ref id
     */
    public static String getObjectRefId() {
        return OBJECT_REF_PREFIX + UUID.randomUUID().toString().replace("-", "");
    }

    /**
     * Gets lib runtime config.
     *
     * @param configManager the config manager
     * @return the lib runtime config
     * @throws YRException the YR exception
     */
    public static LibRuntimeConfig getLibRuntimeConfig(ConfigManager configManager) throws YRException {
        LibRuntimeConfig libConfig = new LibRuntimeConfig();
        libConfig.setFunctionSystemIpAddr(configManager.getServerAddress());
        libConfig.setFunctionSystemRtServerIpAddr(configManager.getServerAddress());
        libConfig.setDataSystemIpAddr(configManager.getDataSystemAddress());
        libConfig.setFunctionSystemPort(configManager.getServerAddressPort());
        libConfig.setFunctionSystemRtServerPort(configManager.getServerAddressPort());
        libConfig.setDataSystemPort(configManager.getDataSystemAddressPort());
        libConfig.setDriver(true);
        libConfig.setJobId(configManager.getJobId());
        libConfig.setRuntimeId(configManager.getRuntimeId());
        libConfig.setLogLevel(configManager.getLogLevel());
        libConfig.setLogDir(configManager.getLogDir());
        libConfig.setLogFileSizeMax(configManager.getLogFileSizeMax());
        libConfig.setLogFileNumMax(configManager.getLogFileNumMax());
        libConfig.setLogMerge(configManager.isLogMerge());
        libConfig.setRecycleTime(configManager.getRecycleTime());
        libConfig.setMaxTaskInstanceNum(configManager.getMaxTaskInstanceNum());
        libConfig.setMaxConcurrencyCreateNum(configManager.getMaxConcurrencyCreateNum());
        libConfig.setEnableMetrics(configManager.isEnableMetrics());
        libConfig.setThreadPoolSize(configManager.getThreadPoolSize());
        libConfig.setLoadPaths(configManager.getLoadPaths());
        libConfig.setInCluster(configManager.isInCluster());
        libConfig.setHttpIocThreadsNum(configManager.getHttpIocThreadsNum());
        libConfig.setHttpIdleTime(configManager.getHttpIdleTime());
        libConfig.setRpcTimeout(configManager.getRpcTimeout());
        libConfig.setTenantId(configManager.getTenantId());
        libConfig.setCustomEnvs(configManager.getCustomEnvs());
        libConfig.setEnableMTLS(configManager.isEnableMTLS());
        libConfig.setCertificateFilePath(configManager.getCertificateFilePath());
        libConfig.setVerifyFilePath(configManager.getVerifyFilePath());
        libConfig.setPrivateKeyPath(configManager.getPrivateKeyPath());
        libConfig.setServerName(configManager.getServerName());
        libConfig.setPrivateKeyPaaswd(configManager.getPrivateKeyPaaswd());
        libConfig.setCodePath(configManager.getCodePath());
        Map<Libruntime.LanguageType, String> functionIds =
                new HashMap<Libruntime.LanguageType, String>() {};
        if (configManager.getFunctionURN() != null && !configManager.getFunctionURN().isEmpty()) {
            functionIds.put(Libruntime.LanguageType.Java, reformatFunctionUrn(configManager.getFunctionURN()));
        }
        if (configManager.getCppFunctionURN() != null && !configManager.getCppFunctionURN().isEmpty()) {
            functionIds.put(Libruntime.LanguageType.Cpp, reformatFunctionUrn(configManager.getCppFunctionURN()));
        }
        if (configManager.getGoFunctionURN() != null && !configManager.getGoFunctionURN().isEmpty()) {
            functionIds.put(Libruntime.LanguageType.Golang, reformatFunctionUrn(configManager.getGoFunctionURN()));
        }
        libConfig.setFunctionIds(functionIds);
        LOGGER.debug("java functionIds: {}", functionIds.get(Libruntime.LanguageType.Java));
        LOGGER.debug("cpp functionIds: {}", functionIds.get(Libruntime.LanguageType.Cpp));
        LOGGER.debug("Log file directory path: {}", libConfig.getLogDir());
        libConfig.setEnableDsEncrypt(configManager.isEnableDsEncrypt());
        libConfig.setDsPublicKeyContextPath(configManager.getDsPublicKeyContextPath());
        libConfig.setRuntimePublicKeyContextPath(configManager.getRuntimePublicKeyContextPath());
        libConfig.setRuntimePrivateKeyContextPath(configManager.getRuntimePrivateKeyContextPath());
        libConfig.setEnableFrontendTLS(configManager.isEnableFrontendTLS());
        return libConfig;
    }

    /**
     * reformat functionurn to functionid
     *
     * @param functionUrn string like this format "sn:cn:yrk:default:function:
     *                    0-perf-helloworld:$latest"
     * @return functionid in this format "default/0-perf-helloworld/$latest"
     * @throws YRException the YR exception
     */
    public static String reformatFunctionUrn(String functionUrn) throws YRException {
        String[] splited = functionUrn.split(":");
        if (splited.length != FUNCTION_URN_LENGTH) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                    "function urn format is wrong");
        }
        return splited[FUNCTION_URN_LEASER_INDEX] + "/" + splited[FUNCTION_URN_FUNCTION_INDEX] + "/"
                + splited[FUNCTION_URN_VERSION_INDEX];
    }

    /**
     * Pack the arguments to be invoked into a list of InvokeArg objects.
     *
     * @param args The arguments to be packed.
     * @return A list of InvokeArg objects.
     * @throws YRException If there is an error in packing the arguments.
     */
    public static List<InvokeArg> packInvokeArgs(Object... args) throws YRException {
        List<InvokeArg> invokeArgs = new ArrayList<InvokeArg>();
        for (Object arg : args) {
            InvokeArg invokeArg;
            try {
                invokeArg = new InvokeArg(Serializer.serialize(arg));
            } catch (JsonProcessingException e) {
                throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME_INVOKE, e);
            }
            invokeArg.setObjectRef(false);
            invokeArg.setNestedObjects(new HashSet<>());
            invokeArgs.add(invokeArg);
        }
        return invokeArgs;
    }

    /**
     * Auto boxing class [ ].
     *
     * @param parameterTypes the parameter types
     * @param actualParameterTypes the actual parameter types
     * @return the class [ ]
     */
    private static Class<?>[] autoBoxing(Class<?>[] parameterTypes, Class<?>[] actualParameterTypes) {
        Class<?>[] oriParameterTypes = new Class<?>[parameterTypes.length];
        for (int i = 0; i < parameterTypes.length; i++) {
            oriParameterTypes[i] = parameterTypes[i];
            if (typeMap.containsKey(parameterTypes[i])) {
                parameterTypes[i] = typeMap.get(parameterTypes[i]);
            }
        }
        for (int i = 0; i < actualParameterTypes.length; i++) {
            if (typeMap.containsKey(actualParameterTypes[i])) {
                actualParameterTypes[i] = typeMap.get(actualParameterTypes[i]);
            }
        }
        return oriParameterTypes;
    }

    /**
     * Check parameter types.
     *
     * @param yrFunction the yr function
     * @param args       the args
     * @throws YRException the YR exception
     */
    public static void checkJavaParameterTypes(FunctionWrapper yrFunction, Object... args) throws YRException {
        Class<?>[] parameterTypes = yrFunction.getMethod().getParameterTypes();
        if (parameterTypes.length != args.length) {
            StringBuilder errBuilder = new StringBuilder();
            errBuilder.append("method ");
            errBuilder.append(yrFunction.getFunctionMeta().getFunctionName());
            errBuilder.append(" get wrong number of parameter, you should input:");
            errBuilder.append(parameterTypes.length);
            errBuilder.append(" parameter(s), actually get:");
            errBuilder.append(args.length);
            errBuilder.append(" parameter(s)");
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME, errBuilder.toString());
        }
        Class<?>[] actualType = new Class<?>[parameterTypes.length];
        for (int i = 0; i < parameterTypes.length; i++) {
            actualType[i] = args[i].getClass();
        }
        Class<?>[] oriParameterTypes = SdkUtils.autoBoxing(parameterTypes, actualType);
        int idx = 0;
        while (idx < parameterTypes.length
                && (parameterTypes[idx].isAssignableFrom(actualType[idx]) || (actualType[idx] == ObjectRef.class))) {
            idx++;
        }
        if (idx != parameterTypes.length) {
            StringBuilder errBuilder = new StringBuilder("wrong parameter type, you should input: ");
            errBuilder.append(Arrays.toString(oriParameterTypes));
            errBuilder.append(", actually get: ");
            errBuilder.append(Arrays.toString(actualType));
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME, errBuilder.toString());
        }
    }

    /**
     * Filter yuanrong trace info stack trace element [ ] when get api is called.
     *
     * @param exception the exception
     */
    public static void filterGetTraceInfo(Exception exception) {
        java.lang.StackTraceElement[] stackTrace = exception.getStackTrace();
        Stream<java.lang.StackTraceElement> stream = Stream.of(stackTrace);
        stream =
                stream.filter(element -> !element.getClassName().contains(YUANRONG_REGEX))
                        .filter(element -> !element.getClassName().contains(Utils.REFLECT_REGEX));
        java.lang.StackTraceElement[] filteredStackTraceElements = stream.toArray(java.lang.StackTraceElement[]::new);
        exception.setStackTrace(filteredStackTraceElements);
    }

    /**
     * Check invoke accurate exception location boolean.
     *
     * @param elements the elements
     * @return the boolean
     */
    public static boolean checkInvokeAccurateExceptionLocation(StackTraceElement[] elements) {
        return (ConfigManager.getInstance().isEnableDisConvCallStack() || IS_ENABLE_DIS_CONV_CALL_STACK) && (
            elements.length > 2);
    }

    /**
     * Return the tenant context.
     *
     * @param config the config of yuanrong
     * @param threadID the threadID of the thread
     * @return the string of tenant context
     */
    public static String getTenantContext(Config config, String threadID) {
        long hash = 41L;
        hash = (53 * hash) + config.hashCode() & 0x7fffffff;
        hash = (37 * hash) + threadID.hashCode() & 0x7fffffff;
        long timestamp = System.currentTimeMillis();
        Long t = Long.valueOf(timestamp);
        hash = (19 * hash) + t.hashCode() & 0x7fffffff;
        return String.valueOf(hash);
    }

    /**
     * Return the value of key in the map, if not exist return def.
     *
     * @param input the input map
     * @param key the key will be found in input
     * @param def the default return value
     * @return the string of return value
     */
    public static String defaultIfNotFound(Map<String, String> input, String key, String def) {
        String ret = "";
        ret = input.getOrDefault(key, def);
        return ret;
    }

    /**
     * Check MSetParam.
     *
     * @param mSetParam the MSetParam class
     * @throws YRException the YR exception
     */
    public static void checkMSetParam(MSetParam mSetParam) throws YRException {
        if (mSetParam.getTtlSecond() < 0) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                    "Cannot set a negative value to mSetParam's ttlSecond: " + mSetParam.getTtlSecond());
        }
        if (mSetParam.getExistence() != ExistenceOpt.NX) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                    "MSetParam's existence should be NX.");
        }
    }

    /**
     * Validates that the given value is non-negative.
     *
     * @param value the value to validate
     * @param message the error message to throw if the value is negative
     */
    public static void validateNonNegative(int value, String message) {
        if (value < 0) {
            throw new IllegalArgumentException(message);
        }
    }
}
