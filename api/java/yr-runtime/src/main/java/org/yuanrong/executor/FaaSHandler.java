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

package org.yuanrong.executor;

import static org.yuanrong.services.enums.FaasErrorCode.ENTRY_NOT_FOUND;
import static org.yuanrong.services.enums.FaasErrorCode.FAAS_INIT_ERROR;
import static org.yuanrong.services.enums.FaasErrorCode.FUNCTION_RUN_ERROR;
import static org.yuanrong.services.enums.FaasErrorCode.INITIALIZE_FUNCTION_ERROR;
import static org.yuanrong.services.enums.FaasErrorCode.INIT_FUNCTION_FAIL;
import static org.yuanrong.services.enums.FaasErrorCode.NONE_ERROR;
import static org.yuanrong.services.enums.FaasErrorCode.RESPONSE_EXCEED_LIMIT;
import static org.yuanrong.runtime.util.Utils.getMethod;
import static org.yuanrong.runtime.util.Utils.splitUserClassAndMethod;

import org.yuanrong.services.UDFManager;
import org.yuanrong.services.enums.FaasErrorCode;
import org.yuanrong.services.model.CallResponse;
import org.yuanrong.services.model.CallResponseJsonObject;
import org.yuanrong.services.model.Response;
import org.yuanrong.services.runtime.Context;
import org.yuanrong.services.runtime.action.ContextImpl;
import org.yuanrong.services.runtime.action.ContextInvokeParams;
import org.yuanrong.services.runtime.action.DelegateDecrypt;
import org.yuanrong.services.runtime.action.LogTankService;
import org.yuanrong.services.runtime.utils.DataTypeAdapter;
import org.yuanrong.services.runtime.utils.Util;
import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.errorcode.Pair;
import org.yuanrong.exception.HandlerNotAvailableException;
import org.yuanrong.exception.LibRuntimeException;
import org.yuanrong.jni.LibRuntime;
import org.yuanrong.libruntime.generated.Libruntime;
import org.yuanrong.libruntime.generated.Libruntime.FunctionMeta;
import org.yuanrong.runtime.util.Constants;
import org.yuanrong.runtime.util.ExtClasspathLoader;
import org.yuanrong.utils.RuntimeUtils;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.google.gson.JsonElement;
import com.google.gson.JsonObject;
import com.google.gson.JsonSyntaxException;
import com.google.gson.reflect.TypeToken;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.io.PrintWriter;
import java.io.StringWriter;
import java.lang.reflect.Field;
import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;
import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;
import java.util.Base64;
import java.util.HashMap;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.Objects;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.TimeoutException;

/**
 * The type Faas handler.
 *
 * @since 2024-07-01
 */
public class FaaSHandler implements HandlerIntf {
    /**
     * The constant LOG.
     */
    private static final Logger LOG = LoggerFactory.getLogger(FaaSHandler.class);

    /**
     * The constant KEY_USER_INIT_ENTRY.
     */
    private static final String KEY_USER_INIT_ENTRY = "userInitEntry";

    /**
     * The constant KEY_USER_CALL_ENTRY.
     */
    private static final String KEY_USER_CALL_ENTRY = "userCallEntry";

    /**
     * The constant KEY_INSTANCE_LABEL
     */
    private static final String KEY_INSTANCE_LABEL = "instanceLabel";

    /**
     * The Index call context.
     */
    private static final int INDEX_CALL_CONTEXT = 0;

    /**
     * The Index user handler
     */
    private static final int INDEX_USER_HANDLER = 1;

    private static final int INDEX_USER_ENTRY_CLASS = 0;

    private static final int INDEX_USER_ENTRY_METHOD = 1;

    /**
     * The Index call user event.
     */
    private static final int INDEX_CALL_USER_EVENT = 1;

    /**
     * The first parameter type index.
     */
    private static final int FIRST_PARAMETER_TYPE = 0;

    private static final int ARGS_MINIMUM_LENGTH = 2;

    private static final int USER_EVENT_MAX_SIZE = 10 * 1024 * 1024;

    private static final int RESPONSE_MAX_SIZE = 10 * 1024 * 1024;

    private static final String INVALID_ARGS_EXCEPTION = "faas get args invalid";

    private static final String RUNTIME_ROOT = "/home/snuser/runtime";

    private static final String RUNTIME_CODE_ROOT = "/opt/function/code";

    private static final String RUNTIME_LOG_DIR = "/home/snuser/log";

    private static final String LD_LIBRARY_PATH = "LD_LIBRARY_PATH";

    private static final String HEADER_STR = "header";

    private static final String BODY_STR = "body";

    private static final String X_TRACE_ID = "X-Trace-Id";

    private static final String X_INSTANCE_SESSION = "X-Instance-Session";

    private static final String SESSION_ID = "sessionID";

    private static final String ENV_DELEGATE_DECRYPT = "ENV_DELEGATE_DECRYPT";

    private static final String EVENT_HEADER = "Accept";

    private static final String EVENT_HEADER_VALUE = "text/event-stream";

    private static final String EVENT_EOF = "yuanrong_event_EOF";

    /**
     * The Gson.
     */
    private static final Gson GSON = new GsonBuilder().serializeNulls().registerTypeAdapter(
            new TypeToken<Map<String, Object>>() {
            }.getType(), new DataTypeAdapter()
        ).disableHtmlEscaping().setPrettyPrinting().create();

    private Context context = new ContextImpl();
    private Method callMethod;
    private Class<?> userEventClazz;
    private int initializerTimeout;
    private String preStopHandler;
    private int preStopTimeout;

    /**
     * Execute function
     *
     * @param meta the meta
     * @param type the type
     * @param args the args
     * @return the return value, in ByteBuffer type, may need to release bytebuffer
     *         `buffer.clear()`
     * @throws Exception the Exception
     */
    @Override
    public ReturnType execute(FunctionMeta meta, Libruntime.InvokeType type, List<ByteBuffer> args)
            throws Exception {
        LOG.info("executing udf methods, current type: {}", type);
        List<String> argList = RuntimeUtils.convertArgListToStringList(args);
        for (ByteBuffer buffer : args) {
            releaseDirectBuffer(buffer);
        }
        args.clear();
        if (Objects.isNull(argList) || argList.isEmpty()) {
            return new ReturnType(ErrorCode.ERR_PARAM_INVALID, "call handler arg list is empty.");
        }
        switch (type) {
            case CreateInstance:
            case CreateInstanceStateless:
                LOG.debug("Invoking udf method matched, case create");
                ErrorInfo initErrorInfo = faasInitHandler(argList);
                if (ErrorCode.ERR_OK.equals(initErrorInfo.getErrorCode())) {
                    return new ReturnType(ErrorCode.ERR_OK, initErrorInfo.getErrorMessage());
                }
                return new ReturnType(initErrorInfo.getErrorCode(), initErrorInfo.getErrorMessage());
            case InvokeFunction:
            case InvokeFunctionStateless:
                LOG.debug("Invoking udf method matched, case invoke");
                ContextInvokeParams params = new ContextInvokeParams();
                ((ContextImpl) context).setContextInvokeParams(params);
                return new ReturnType(ErrorCode.ERR_OK, faasCallHandler(argList));
            default:
                LOG.debug("Invoking udf method matched, case dft");
                return new ReturnType(ErrorCode.ERR_INCORRECT_INVOKE_USAGE, "invalid invoke type");
        }
    }

    private void releaseDirectBuffer(ByteBuffer buffer) {
        if (buffer == null || !buffer.isDirect()) {
            return;
        }
        // For java9+
        try {
            Method cleanerMethod = buffer.getClass().getDeclaredMethod("cleaner");
            cleanerMethod.setAccessible(true);
            Object cleaner = cleanerMethod.invoke(buffer);
            if (cleaner != null) {
                Method cleanMethod = cleaner.getClass().getMethod("clean");
                cleanMethod.invoke(cleaner);
            }
            return;
        } catch (Exception e) {
            LOG.warn("Failed to release direct buffer, it may be java8", e);
        }
        // For java8
        try {
            Class<?> directBufferClass = Class.forName("sun.nio.ch.DirectBuffer");
            if (directBufferClass.isInstance(buffer)) {
                Method cleanerMethod = directBufferClass.getDeclaredMethod("cleaner");
                cleanerMethod.setAccessible(true);
                Object cleaner = cleanerMethod.invoke(buffer);
                if (cleaner != null) {
                    Method cleanMethod = cleaner.getClass().getMethod("clean");
                    cleanMethod.invoke(cleaner);
                }
            }
        } catch (Exception e) {
            LOG.error("Failed to release direct buffer", e);
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
        if (preStopTimeout == 0 || preStopHandler == null || preStopHandler.isEmpty()) {
            return new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.RUNTIME, "no need to run shut down handler.");
        }
        ExecutorService shutdownExecutorService = Executors.newSingleThreadExecutor();
        Future<ErrorInfo> shutdownFuture = shutdownExecutorService
                .submit(() -> faasShutDownHandler(gracePeriodSeconds));
        try {
            return shutdownFuture.get(preStopTimeout, TimeUnit.SECONDS);
        } catch (TimeoutException e) {
            String errorMassage = String.format(Locale.ROOT, "prestop timed out after %d s", preStopTimeout);
            LOG.error(errorMassage);
            return new ErrorInfo(new ErrorCode(FaasErrorCode.INVOKE_FUNCTION_TIMEOUT.getCode()), ModuleCode.RUNTIME,
                    errorMassage);
        } catch (InterruptedException | ExecutionException e) {
            String errorMassage = String.format(Locale.ROOT,
                    "faas failed to run user preStop code. err: %s , cause: %s", e.getMessage(), getCausedByString(e));
            LOG.error(errorMassage);
            return new ErrorInfo(ErrorCode.ERR_USER_FUNCTION_EXCEPTION, ModuleCode.RUNTIME, errorMassage);
        } finally {
            shutdownExecutorService.shutdownNow();
        }
    }

    /**
     * Loads an instance of a class from serialized byte arrays.
     *
     * @param  instanceBytes  the serialized byte array representing the instance
     * @param  clzNameBytes   the serialized byte array representing the class name
     * @throws IOException     if there is an error reading the byte arrays
     * @throws ClassNotFoundException if the class specified by the class name is not found
     */
    @Override
    public void loadInstance(byte[] instanceBytes, byte[] clzNameBytes) {}

    /**
     * Serializes the instance of the CodeExecutor class and returns a Pair
     * containing the serialized byte arrays of the instance and the class name.
     *
     * @param instanceID the ID of the instance to be dumped
     * @return a Pair containing the serialized byte arrays of the instance and the
     *         class name
     * @throws JsonProcessingException if there is an error during serialization
     */
    @Override
    public Pair<byte[], byte[]> dumpInstance(String instanceID) throws JsonProcessingException {
        return new Pair<>(new byte[0], new byte[0]);
    }

    /**
     * Recover the instance.
     *
     * @return ErrorInfo, the ErrorInfo of the execution of recover function.
     */
    @Override
    public ErrorInfo recover() {
        return new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.RUNTIME, "");
    }

    private Object getInputParameters(String userEvent, Class<?> userEventClazz) {
        if (String.class.equals(userEventClazz)) {
            return userEvent;
        }
        return GSON.fromJson(userEvent, userEventClazz);
    }

    /**
     * Faas init handler.
     *
     * @param args the args
     * @return InitErrorResponse
     */
    public ErrorInfo faasInitHandler(List<String> args) {
        LOG.info("faas init handler called.");
        if (args == null || args.size() < ARGS_MINIMUM_LENGTH) {
            return new ErrorInfo(new ErrorCode(FAAS_INIT_ERROR.getCode()), ModuleCode.RUNTIME, INVALID_ARGS_EXCEPTION);
        }
        Map<String, String> createParams = null;
        try {
            createParams = GSON.fromJson(args.get(INDEX_USER_HANDLER), Map.class);
            context = initContext(args);
            ((ContextImpl) context).setInstanceLabel(createParams.get(KEY_INSTANCE_LABEL));
        } catch (JsonSyntaxException e) {
            String errorMessage = String.format(Locale.ROOT, "faas failed to convert json to object. err: %s",
                    e.getMessage());
            LOG.error(errorMessage);
            return new ErrorInfo(new ErrorCode(INITIALIZE_FUNCTION_ERROR.getCode()), ModuleCode.RUNTIME, errorMessage);
        }
        LOG.debug("faas succeeds to init context ");

        // loadCallMethod must run before runUserInitHandler
        String userCallEntry = createParams.get(KEY_USER_CALL_ENTRY);
        ErrorInfo loadCallErrorInfo = loadCallMethod(userCallEntry);
        if (!ErrorCode.ERR_OK.equals(loadCallErrorInfo.getErrorCode())) {
            return loadCallErrorInfo;
        }
        String userInitEntry = createParams.get(KEY_USER_INIT_ENTRY);
        ErrorInfo runUserInitErrorInfo = runUserInitHandler(userInitEntry);
        if (!ErrorCode.ERR_OK.equals(runUserInitErrorInfo.getErrorCode())) {
            return runUserInitErrorInfo;
        }
        LOG.info("faas init handler complete.");
        return new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.RUNTIME, "processed init handler successfully.");
    }

    /**
     * FaaS call handler.
     *
     * @param args the args
     * @return String
     * @throws HandlerNotAvailableException the handler not available exception
     */
    public String faasCallHandler(List<String> args) throws HandlerNotAvailableException {
        LOG.info("faas call handler called.");
        ContextImpl callContext = new ContextImpl((ContextImpl) context);
        Object result;
        int innerCode = NONE_ERROR.getCode();
        if (args == null || args.size() < ARGS_MINIMUM_LENGTH) {
            result = INVALID_ARGS_EXCEPTION;
            innerCode = FaasErrorCode.FAAS_INIT_ERROR.getCode();
            return handleResponse(result, innerCode);
        }
        if (args.get(INDEX_CALL_USER_EVENT).getBytes(StandardCharsets.UTF_8).length > USER_EVENT_MAX_SIZE) {
            result = FaasErrorCode.REQUEST_BODY_EXCEED_LIMIT.getErrorMessage();
            innerCode = FaasErrorCode.REQUEST_BODY_EXCEED_LIMIT.getCode();
            return handleResponse(result, innerCode);
        }
        JsonObject jsonObject = GSON.fromJson(args.get(INDEX_CALL_USER_EVENT), JsonObject.class);
        String userEvent = "";

        if (jsonObject.has(BODY_STR) && !jsonObject.get(BODY_STR).isJsonNull()) {
            JsonElement jsonElement = jsonObject.get(BODY_STR);
            userEvent = (jsonElement.isJsonObject()) ? jsonElement.toString() : jsonElement.getAsString();
        }

        boolean isEventEnable = false;
        if (jsonObject.has(HEADER_STR) && !jsonObject.get(HEADER_STR).isJsonNull()) {
            JsonObject headerObj = jsonObject.getAsJsonObject(HEADER_STR);
            if (headerObj.has(X_TRACE_ID) && !headerObj.get(X_TRACE_ID).isJsonNull()) {
                String traceId = headerObj.get(X_TRACE_ID).getAsString();
                callContext.setTraceID(traceId);
            }
            if (headerObj.has(X_INSTANCE_SESSION) && !headerObj.get(X_INSTANCE_SESSION).isJsonNull()) {
                try {
                    String sessionJsonStr = headerObj.get(X_INSTANCE_SESSION).getAsString();
                    JsonObject sessionObj = GSON.fromJson(sessionJsonStr, JsonObject.class);
                    if (sessionObj.has(SESSION_ID) && !sessionObj.get(SESSION_ID).isJsonNull()) {
                        callContext.setSessionId(sessionObj.get(SESSION_ID).getAsString());
                    }
                } catch (JsonSyntaxException | IllegalStateException e) {
                    LOG.warn("Failed to parse session JSON from header, degrade gracefully. cause: {}", e.getMessage());
                }
            }
            if (headerObj.has(EVENT_HEADER) && !headerObj.get(EVENT_HEADER).isJsonNull()) {
                if (EVENT_HEADER_VALUE.equals(headerObj.get(EVENT_HEADER).getAsString())) {
                    isEventEnable = true;
                    Pair<String, String> reqInstanceId = LibRuntime.getRequestAndInstanceID();
                    callContext.getContextInvokeParams().setInvokeID(reqInstanceId.getFirst());
                    callContext.setInstanceID(reqInstanceId.getSecond());
                }
            }
        }

        userEvent = "null".equals(userEvent) ? "" : userEvent;
        String logType = "";
        if (jsonObject.has(HEADER_STR)) {
            JsonObject headerObject = jsonObject.get(HEADER_STR).getAsJsonObject();
            if (headerObject != null && headerObject.has(Constants.CFF_LOG_TYPE)) {
                logType = headerObject.get(Constants.CFF_LOG_TYPE).toString();
            }
        }
        Util.setLogOpts(logType);
        String logGroupId = "";
        String logStreamId = "";
        LogTankService logTankService = callContext.getExtendedMetaData().getLogTankService();
        if (logTankService != null) {
            if (logTankService.getLogGroupId() != null) {
                logGroupId = logTankService.getLogGroupId();
            }
            if (logTankService.getLogStreamId() != null) {
                logStreamId = logTankService.getLogStreamId();
            }
        }

        String[] callInfos = new String[] {
            callContext.getInvokeID(), callContext.getRequestID(), callContext.getInstanceID(),
            Util.getFunctionInfo(callContext),
            logGroupId,
            logStreamId
        };
        Util.setInheritableThreadLocal(callInfos);
        UDFManager udfManager = UDFManager.getUDFManager();
        long startTime = System.currentTimeMillis();
        try {
            result = callMethod.invoke(udfManager.loadInstance(KEY_USER_CALL_ENTRY),
                getInputParameters(userEvent, userEventClazz), callContext);
        } catch (IllegalAccessException | IllegalArgumentException e) {
            String errorMsg = getErrorMsg(e);
            String cause = getCausedByString(e);
            LOG.error("faas run invoke method failed, errorMsg: {}, cause : {}", errorMsg, cause);
            result = errorMsg;
            innerCode = FaasErrorCode.FUNCTION_RUN_ERROR.getCode();
        } catch (InvocationTargetException e) {
            innerCode = FaasErrorCode.FUNCTION_RUN_ERROR.getCode();
            String errorMsg = getErrorMsg(e);
            String cause = getCausedByString(e.getCause());
            LOG.error("faas run invoke user method failed, errorMsg: {}, cause : {}", errorMsg, cause);
            result = errorMsg;
        }
        Util.clearLogOpts();
        Util.clearInheritableThreadLocal();
        if (isEventEnable) {
            try {
                LibRuntime.streamWrite(EVENT_EOF, callContext.getInvokeID(), callContext.getInstanceID());
            } catch (LibRuntimeException e) {
                innerCode = FaasErrorCode.FUNCTION_RUN_ERROR.getCode();
                String errorMsg = getErrorMsg(e);
                String cause = getCausedByString(e.getCause());
                LOG.error("failed to send event EOF, errorMsg: {}, cause : {}", errorMsg, cause);
                result = errorMsg;
            }
        }
        long userFuncTime = System.currentTimeMillis() - startTime;
        return handleResponse(result, innerCode, userFuncTime);
    }

    private static String getCausedByString(Throwable throwable) {
        StringWriter stringWriter = new StringWriter();
        PrintWriter printWriter = new PrintWriter(stringWriter);
        throwable.printStackTrace(printWriter);
        return stringWriter.toString();
    }

    private static String getErrorMsg(Throwable exception) {
        String msg = exception.getMessage();
        if (exception instanceof InvocationTargetException) {
            Throwable cause = exception.getCause();
            if (cause != null) {
                msg = cause.toString();
            }
        }
        return msg;
    }

    /**
     * Faas check point handler.
     *
     * @param args the args
     */
    public void faasCheckPointHandler(List<String> args) {
        return;
    }

    /**
     * faasRecoverHandler
     *
     * @param args the args
     */
    public void faasRecoverHandler(List<String> args) {
        return;
    }

    /**
     * faasShutDownHandler
     *
     * @param gracePeriodSeconds grace period seconds
     * @return ErrorInfo, the ErrorInfo of the execution of preStopHandler.
     */
    public ErrorInfo faasShutDownHandler(int gracePeriodSeconds) {
        LOG.info("faas shut down handler called.");
        UDFManager udfManager = UDFManager.getUDFManager();
        Class<?> userClass = udfManager.loadClass();
        if (preStopHandler != null && !preStopHandler.isEmpty()) {
            try {
                String[] preStopClassMethod = splitUserClassAndMethod(preStopHandler, true);
                Method method = userClass.getMethod(preStopClassMethod[INDEX_USER_ENTRY_METHOD], Context.class);
                method.setAccessible(true);
                method.invoke(udfManager.loadInstance(KEY_USER_CALL_ENTRY), context);
            } catch (NoSuchMethodException e) {
                LOG.error("faas failed to load user preStop code. err: {}", e.getMessage());
                return new ErrorInfo(new ErrorCode(ENTRY_NOT_FOUND.getCode()), ModuleCode.RUNTIME,
                    ENTRY_NOT_FOUND.getErrorMessage());
            } catch (IllegalAccessException | IllegalArgumentException | InvocationTargetException e) {
                String errorMessage = String.format(Locale.ROOT,
                    "faas failed to run user preStop code. err: %s , cause: %s", getErrorMsg(e), getCausedByString(e));
                LOG.error(errorMessage);
                return new ErrorInfo(new ErrorCode(FUNCTION_RUN_ERROR.getCode()), ModuleCode.RUNTIME, errorMessage);
            }
        }
        return new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.RUNTIME, "run shut down handler successfully.");
    }

    /**
     * faasSignalHandler
     *
     * @param args the args
     */
    public void faasSignalHandler(List<String> args) {
        return;
    }

    /**
     * registerClass , registerInstance , set callMethod , set userEventClazz
     *
     * @param userCallEntry userCallEntry
     * @return ErrorInfo
     */
    private ErrorInfo loadCallMethod(String userCallEntry) {
        ClassLoader classLoader = ExtClasspathLoader.getFunctionClassLoader();
        UDFManager udfManager = UDFManager.getUDFManager();
        try {
            String[] callEntryClassMethod = splitUserClassAndMethod(userCallEntry, false);
            Class<?> userClass = classLoader.loadClass(callEntryClassMethod[INDEX_USER_ENTRY_CLASS]);
            udfManager.registerClass(userClass);
            Object entryInstance = userClass.newInstance();
            udfManager.registerInstance(KEY_USER_CALL_ENTRY, entryInstance);
            callMethod = getMethod(userClass, callEntryClassMethod[INDEX_USER_ENTRY_METHOD]);
            userEventClazz = callMethod.getParameterTypes()[FIRST_PARAMETER_TYPE];
            callMethod.setAccessible(true);
        } catch (ClassNotFoundException | NoSuchMethodException | InstantiationException | IllegalAccessException e) {
            LOG.error("faas failed to init user code. err: {}", e.getMessage());
            return new ErrorInfo(new ErrorCode(ENTRY_NOT_FOUND.getCode()), ModuleCode.RUNTIME,
                ENTRY_NOT_FOUND.getErrorMessage());
        } catch (Exception e) {
            String errorMessage = String.format(Locale.ROOT, "faas unexpected exception: %s", e.getMessage());
            LOG.error(errorMessage);
            return new ErrorInfo(new ErrorCode(INITIALIZE_FUNCTION_ERROR.getCode()), ModuleCode.RUNTIME, errorMessage);
        }
        return new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.RUNTIME, "load call method successfully.");
    }

    /**
     * load userClass , loadUserInstance , invoke userInitMethod
     *
     * @param userInitEntry userInitEntry
     * @return InitErrorResponse
     */
    private ErrorInfo runUserInitHandler(String userInitEntry) {
        UDFManager udfManager = UDFManager.getUDFManager();
        Class<?> userClass = udfManager.loadClass();
        // execute user init code if present
        if (userInitEntry != null && !userInitEntry.isEmpty()) {
            try {
                String[] initEntryClassMethod = splitUserClassAndMethod(userInitEntry, true);
                Method method = userClass.getMethod(initEntryClassMethod[INDEX_USER_ENTRY_METHOD], Context.class);
                method.setAccessible(true);
                method.invoke(udfManager.loadInstance(KEY_USER_CALL_ENTRY), context);
            } catch (NoSuchMethodException e) {
                LOG.error("faas failed to load user init code. err: {}", e.getMessage());
                return new ErrorInfo(new ErrorCode(ENTRY_NOT_FOUND.getCode()), ModuleCode.RUNTIME,
                    ENTRY_NOT_FOUND.getErrorMessage());
            } catch (IllegalAccessException | IllegalArgumentException e) {
                String errorMessage = String.format(Locale.ROOT, "faas failed to run user init code. err: %s",
                        getErrorMsg(e));
                LOG.error(errorMessage);
                return new ErrorInfo(new ErrorCode(INIT_FUNCTION_FAIL.getCode()), ModuleCode.RUNTIME, errorMessage);
            } catch (InvocationTargetException e) {
                String errorMessage = String.format(Locale.ROOT, "faas failed to run user init code. err: %s",
                    getErrorMsg(e));
                LOG.error("{}, cause: {}", errorMessage, getCausedByString(e.getCause()));
                return new ErrorInfo(new ErrorCode(INIT_FUNCTION_FAIL.getCode()), ModuleCode.RUNTIME, errorMessage);
            }
        }
        return new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.RUNTIME, "run user init handler successfully.");
    }

    private Context initContext(List<String> args) {
        ContextImpl contextImpl = GSON.fromJson(args.get(INDEX_CALL_CONTEXT), ContextImpl.class);
        String delegateEncryptEnv = System.getenv(ENV_DELEGATE_DECRYPT);
        DelegateDecrypt delegateDecrypt;
        if (delegateEncryptEnv != null && !delegateEncryptEnv.isEmpty()) {
            delegateDecrypt = GSON.fromJson(delegateEncryptEnv, DelegateDecrypt.class);
        } else {
            delegateDecrypt = GSON.fromJson(args.get(args.size() - 1), DelegateDecrypt.class);
        }
        if (delegateDecrypt == null) {
            delegateDecrypt = new DelegateDecrypt();
        }
        contextImpl.setDelegateDecrypt(delegateDecrypt);
        Map<String, String> runtimeUserDataMap = new HashMap<>();
        Map<String, String> envSetMap = new HashMap<>();
        if (delegateDecrypt.getEnvironment() != null && !delegateDecrypt.getEnvironment().isEmpty()) {
            Map<String, String> environmentMap = GSON.fromJson(delegateDecrypt.getEnvironment(), HashMap.class);
            for (Map.Entry<String, String> entry : environmentMap.entrySet()) {
                String key = entry.getKey();
                String value = entry.getValue();
                if (!LD_LIBRARY_PATH.equals(key)) {
                    envSetMap.put(key, value);
                }
                runtimeUserDataMap.put(key, value);
            }
        }

        if (delegateDecrypt.getEncryptedUserData() != null && !delegateDecrypt.getEncryptedUserData().isEmpty()) {
            Map<String, String> userDataMap = GSON.fromJson(delegateDecrypt.getEncryptedUserData(), HashMap.class);
            runtimeUserDataMap.putAll(userDataMap);
        }
        String pathValue = runtimeUserDataMap.getOrDefault(LD_LIBRARY_PATH, "");
        envSetMap.put(LD_LIBRARY_PATH,
                System.getenv(LD_LIBRARY_PATH) + String.format(":%s", pathValue));
        contextImpl.getFuncMetaData().setUserData(runtimeUserDataMap);
        String userDataString = GSON.toJson(runtimeUserDataMap);
        envSetMap.put("RUNTIME_USERDATA", userDataString);
        setEnvContext(contextImpl, envSetMap);
        this.initializerTimeout = Integer.parseInt(envSetMap.get("RUNTIME_INITIALIZER_TIMEOUT"));
        this.preStopHandler = envSetMap.get("PRE_STOP_HANDLER");
        this.preStopTimeout = Integer.parseInt(envSetMap.getOrDefault("PRE_STOP_TIMEOUT", "0"));
        return contextImpl;
    }

    private String handleResponse(Object result, int innerCode) {
        return handleResponse(result, innerCode, 0);
    }

    private String handleResponse(Object result, int innerCode, long userFuncTime) {
        Response response;
        if (result instanceof JsonObject) {
            response = new CallResponseJsonObject();
        } else {
            response = new CallResponse();
        }
        response.setBody(result);
        response.setBillingDuration("this is billing duration TODO");
        response.setInnerCode(String.valueOf(innerCode));
        response.setInvokerSummary("this is summary TODO");
        response.setLogResult(Base64.getEncoder()
            .encodeToString("this is user log TODO".getBytes(StandardCharsets.UTF_8)));
        if (userFuncTime != 0) {
            response.setUserFuncTime(userFuncTime);
        }
        String resultJson = GSON.toJson(response);

        int respLength = resultJson.getBytes(StandardCharsets.UTF_8).length;
        if (respLength > RESPONSE_MAX_SIZE) {
            response.setBody(String.format(Locale.ROOT, "response body size %d exceeds the limit of %d",
                respLength, RESPONSE_MAX_SIZE));
            response.setInnerCode(String.valueOf(RESPONSE_EXCEED_LIMIT.getCode()));
            resultJson = GSON.toJson(response);
        }
        return resultJson;
    }

    private void setEnvContext(ContextImpl contextImpl, Map<String, String> envSetMap) {
        envSetMap.put("RUNTIME_PROJECT_ID", contextImpl.getFuncMetaData().getTenantId());
        envSetMap.put("RUNTIME_PACKAGE", contextImpl.getFuncMetaData().getService());
        envSetMap.put("RUNTIME_FUNC_NAME", contextImpl.getFuncMetaData().getFuncName());
        envSetMap.put("RUNTIME_FUNC_VERSION", contextImpl.getFuncMetaData().getVersion());
        envSetMap.put("RUNTIME_HANDLER", contextImpl.getFuncMetaData().getHandler());
        envSetMap.put("RUNTIME_TIMEOUT", Integer.toString(contextImpl.getFuncMetaData().getTimeout()));
        envSetMap.put("RUNTIME_CPU", Integer.toString(contextImpl.getResourceMetaData().getCpu()));
        envSetMap.put("RUNTIME_MEMORY", Integer.toString(contextImpl.getResourceMetaData().getMemory()));
        envSetMap.put("RUNTIME_MAX_RESP_BODY_SIZE", Integer.toString(USER_EVENT_MAX_SIZE));
        if (contextImpl.getExtendedMetaData() != null && contextImpl.getExtendedMetaData().getInitializer() != null) {
            if (contextImpl.getExtendedMetaData().getInitializer().getInitializerHandler() != null) {
                envSetMap.put("RUNTIME_INITIALIZER_HANDLER",
                    contextImpl.getExtendedMetaData().getInitializer().getInitializerHandler());
            }
            envSetMap.put("RUNTIME_INITIALIZER_TIMEOUT",
                    Integer.toString(contextImpl.getExtendedMetaData().getInitializer().getInitializerTimeout()));
        }
        if (contextImpl.getExtendedMetaData() != null && contextImpl.getExtendedMetaData().getPreStop() != null) {
            if (contextImpl.getExtendedMetaData().getPreStop().getPreStopHandler() != null) {
                envSetMap.put("PRE_STOP_HANDLER",
                    contextImpl.getExtendedMetaData().getPreStop().getPreStopHandler());
            }
            envSetMap.put("PRE_STOP_TIMEOUT",
                Integer.toString(contextImpl.getExtendedMetaData().getPreStop().getPreStopTimeout()));
        }
        envSetMap.put("RUNTIME_ROOT", RUNTIME_ROOT);
        envSetMap.put("RUNTIME_CODE_ROOT", RUNTIME_CODE_ROOT);
        envSetMap.put("RUNTIME_LOG_DIR", RUNTIME_LOG_DIR);
        setJavaProcessEnvMap(envSetMap);
    }

    /**
     * Sets java process env map.
     *
     * @param envMap the env map
     */
    private static void setJavaProcessEnvMap(Map<String, String> envMap) {
        // keep process previous env map configs
        try {
            Class<?> processEnvironmentClass = Class.forName("java.lang.ProcessEnvironment");
            updateJavaEnvMap(processEnvironmentClass, "theCaseInsensitiveEnvironment", envMap);
            updateJavaEnvMap(processEnvironmentClass, "theUnmodifiableEnvironment", envMap);
        } catch (ClassNotFoundException e) {
            LOG.error("get field: theEnvironment has an error: ", e);
        }
    }

    /**
     * Update java env map.
     *
     * @param cls the cls
     * @param filedName the filed name
     * @param envMap the env map
     */
    private static void updateJavaEnvMap(Class<?> cls, String filedName, Map<String, String> envMap) {
        try {
            // get field and access
            Field oldFiled = cls.getDeclaredField(filedName);
            oldFiled.setAccessible(true);
            // get Filed map
            Object unmodifiableMap = oldFiled.get(null);
            for (Map.Entry<String, String> entry : envMap.entrySet()) {
                LOG.debug("updateJavaEnvMap key: {}, value:{}", entry.getKey(), entry.getValue());
                injectIntoUnmodifiableMap(entry.getKey(), entry.getValue(), unmodifiableMap);
            }
        } catch (ReflectiveOperationException e) {
            LOG.error("get field: {} has an error: {}", filedName, e);
        }
    }

    private static void injectIntoUnmodifiableMap(String key, String value, Object map)
            throws ReflectiveOperationException {
        Class<?> unmodifiableMap = Class.forName("java.util.Collections$UnmodifiableMap");
        Field field = unmodifiableMap.getDeclaredField("m");
        field.setAccessible(true);
        Object obj = field.get(map);
        ((Map<String, String>) obj).put(key, value);
    }
}
