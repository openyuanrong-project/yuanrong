/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
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

import org.yuanrong.services.UDFManager;
import org.yuanrong.services.enums.FaasErrorCode;
import org.yuanrong.services.exception.FaaSException;
import org.yuanrong.services.model.CallRequest;
import org.yuanrong.services.model.CallResponse;
import org.yuanrong.services.runtime.action.ContextImpl;
import org.yuanrong.services.runtime.action.ContextInvokeParams;
import org.yuanrong.services.runtime.action.DelegateDecrypt;
import org.yuanrong.services.runtime.action.ExtendedMetaData;
import org.yuanrong.services.runtime.action.FunctionMetaData;
import org.yuanrong.services.runtime.action.Initializer;
import org.yuanrong.services.runtime.action.LogTankService;
import org.yuanrong.services.runtime.action.PreStop;
import org.yuanrong.services.runtime.action.ResourceMetaData;
import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.Pair;
import org.yuanrong.jni.LibRuntime;
import org.yuanrong.libruntime.generated.Libruntime;
import org.yuanrong.libruntime.generated.Libruntime.FunctionMeta;
import org.yuanrong.runtime.util.Constants;

import com.google.gson.Gson;

import org.junit.Assert;
import org.junit.Test;

import java.lang.reflect.Field;
import java.nio.ByteBuffer;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class TestFaaSHandler {
    private static void setJavaProcessEnvMap(Map<String, String> envMap) {
        // keep process previous env map configs
        try {
            Class<?> processEnvironmentClass = Class.forName("java.lang.ProcessEnvironment");
            updateJavaEnvMap(processEnvironmentClass, "theUnmodifiableEnvironment", envMap);
        } catch (ClassNotFoundException e) {
            e.printStackTrace();
        }
    }

    private static void updateJavaEnvMap(Class<?> cls, String filedName, Map<String, String> envMap) {
        try {
            // get field and access
            Field oldFiled = cls.getDeclaredField(filedName);
            oldFiled.setAccessible(true);
            // get Filed map
            Object unmodifiableMap = oldFiled.get(null);
            for (Map.Entry<String, String> entry : envMap.entrySet()) {
                injectIntoUnmodifiableMap(entry.getKey(), entry.getValue(), unmodifiableMap);
            }
        } catch (ReflectiveOperationException e) {
            e.printStackTrace();
        }
    }

    private static void injectIntoUnmodifiableMap(String key, String value, Object map)
            throws ReflectiveOperationException {
        if (key == null || value == null) {
            return;
        }
        Class<?> unmodifiableMap = Class.forName("java.util.Collections$UnmodifiableMap");
        Field field = unmodifiableMap.getDeclaredField("m");
        field.setAccessible(true);
        Object obj = field.get(map);
        ((Map<String, String>) obj).put(key, value);
    }

    private List<String> generateInitArgs() {
        List<String> args = new ArrayList<>();
        ContextImpl contextImpl = getContext();
        Gson gson = new Gson();
        String contextConfig = gson.toJson(contextImpl);
        String udfEntry = "{\"userInitEntry\":\"org.yuanrong.executor.UserTestHandler.initializer\"," +
                "\"userCallEntry\":\"org.yuanrong.executor.UserTestHandler::handler\"}";
        args.add(contextConfig);
        args.add(udfEntry);
        args.add(getSchedulerData());
        return args;
    }

    private List<String> generateInitArgsWiThPreStopHandler(String testCase) {
        List<String> args = new ArrayList<>();
        ContextImpl contextImpl = getContext();
        PreStop preStop = new PreStop();
        if ("NoSuchMethodException".equals(testCase)) {
            preStop.setPreStopHandler("org.yuanrong.executor.UserTestHandler.wrongPreStop");
            preStop.setPreStopTimeout(60);
        } else if ("InvocationTargetException".equals(testCase)) {
            preStop.setPreStopHandler("org.yuanrong.executor.UserTestHandler.failedPreStop");
            preStop.setPreStopTimeout(60);
        } else if ("InterruptedException".equals(testCase)) {
            preStop.setPreStopHandler("org.yuanrong.executor.UserTestHandler.timeoutPreStop");
            preStop.setPreStopTimeout(1);
        } else {
            preStop.setPreStopHandler("org.yuanrong.executor.UserTestHandler.preStop");
            preStop.setPreStopTimeout(60);
        }
        contextImpl.getExtendedMetaData().setPreStop(preStop);
        Gson gson = new Gson();
        String contextConfig = gson.toJson(contextImpl);
        String udfEntry = "{\"userInitEntry\":\"org.yuanrong.executor.UserTestHandler.initializer\"," +
            "\"userCallEntry\":\"org.yuanrong.executor.UserTestHandler::handler\"}";
        args.add(contextConfig);
        args.add(udfEntry);
        args.add(getSchedulerData());
        return args;
    }

    private List<String> generateCallTimeoutArgs() {
        List<String> args = new ArrayList<>();
        ContextImpl contextImpl = getContext();
        Gson gson = new Gson();
        String contextConfig = gson.toJson(contextImpl);
        String udfEntry = "{\"userInitEntry\":\"org.yuanrong.executor.UserTestHandler.initializer\"," +
                "\"userCallEntry\":\"org.yuanrong.executor.UserTestHandler::timeoutHandler\"}";
        args.add(contextConfig);
        args.add(udfEntry);
        return args;
    }

    private List<String> generateCallResponseLargeSizeArgs() {
        List<String> args = new ArrayList<>();
        ContextImpl contextImpl = getContext();
        Gson gson = new Gson();
        String contextConfig = gson.toJson(contextImpl);
        String udfEntry = "{\"userInitEntry\":\"org.yuanrong.executor.UserTestHandler.initializer\"," +
                "\"userCallEntry\":\"org.yuanrong.executor.UserTestHandler::largeResponse\"}";
        args.add(contextConfig);
        args.add(udfEntry);
        return args;
    }

    private List<String> generateErrorInitArgs(String errorType) {
        Gson gson = new Gson();
        List<String> args = new ArrayList<>();
        ContextImpl contextImpl = getContext();
        String contextConfig = gson.toJson(contextImpl);
        args.add(contextConfig);
        switch (errorType) {
            case "function initialization exception": {
                args.add("Strings cannot unmarshal with map");
                break;
            }
            case "runtime initialization timed out": {
                String udfEntry = "{\"userInitEntry\":\"org.yuanrong.executor.UserTestHandler.timeoutInitializer\"," +
                        "\"userCallEntry\":\"org.yuanrong.executor.UserTestHandler::handler\"}";
                args.add(udfEntry);
                break;
            }
            case "call user entry not found": {
                String wrongUdfEntry = "{\"userInitEntry\":\"org.yuanrong.executor.UserTestHandler.initializer\"," +
                        "\"userCallEntry\":\"org.yuanrong.executor.UserTestHandler::wrongHandler\"}";
                args.add(wrongUdfEntry);
                break;
            }
            case "init user entry not found": {
                String wrongUdfEntry = "{\"userInitEntry\":\"org.yuanrong.executor.UserTestHandler.wrongInitializer\"," +
                        "\"userCallEntry\":\"org.yuanrong.executor.UserTestHandler::handler\"}";
                args.add(wrongUdfEntry);
                break;
            }
            case "IllegalArgumentException" : {
                String wrongUdfEntry =
                    "{\"userInitEntry\":\"org.yuanrong.executor.UserTestHandler.illegalArgumentInitializer\","
                        + "\"userCallEntry\":\"org.yuanrong.executor.UserTestHandler::handler\"}";
                args.add(wrongUdfEntry);
                break;
            }
            case "InvocationTargetException" : {
                String wrongUdfEntry =
                    "{\"userInitEntry\":\"org.yuanrong.executor.UserTestHandler.failedInitializer\","
                        + "\"userCallEntry\":\"org.yuanrong.executor.UserTestHandler::handler\"}";
                args.add(wrongUdfEntry);
                break;
            }
        }
        return args;
    }

    private ContextImpl getContext() {
        ContextImpl contextImpl = new ContextImpl();
        FunctionMetaData funcMetaData = new FunctionMetaData();
        ResourceMetaData resourceMetaData = new ResourceMetaData();
        ExtendedMetaData extendedMetaData = new ExtendedMetaData();
        Initializer initializer = new Initializer();
        DelegateDecrypt delegateDecrypt = new DelegateDecrypt();
        LogTankService logTankService = new LogTankService();
        funcMetaData.setTenantId("12345678910123456789");
        funcMetaData.setFuncName("testpythonbase001");
        funcMetaData.setService("base");
        funcMetaData.setVersion("latest");
        funcMetaData.setHandler("handler");
        funcMetaData.setTimeout(3);
        resourceMetaData.setCpu(500);
        resourceMetaData.setMemory(500);
        initializer.setInitializerHandler("programmingmodel.TestJavaBase002::init");
        initializer.setInitializerTimeout(3);
        extendedMetaData.setInitializer(initializer);
        logTankService.setLogGroupId("groupID");
        logTankService.setLogStreamId("streamID");
        extendedMetaData.setLogTankService(logTankService);
        delegateDecrypt.setAccessKey("accessKey123");
        delegateDecrypt.setSecretKey("secretKey123");
        delegateDecrypt.setAuthToken("authToken123");
        delegateDecrypt.setSecurityToken("securityToken123");
        contextImpl.setFuncMetaData(funcMetaData);
        contextImpl.setResourceMetaData(resourceMetaData);
        contextImpl.setExtendedMetaData(extendedMetaData);
        contextImpl.setDelegateDecrypt(delegateDecrypt);
        ContextInvokeParams params = new ContextInvokeParams();
        params.setRequestID("request-123456789");
        contextImpl.setContextInvokeParams(params);
        return contextImpl;
    }

    private DelegateDecrypt getDelegateDecrypt() {
        Gson gson = new Gson();
        DelegateDecrypt delegateDecrypt = new DelegateDecrypt();
        Map<String, String> map = new HashMap<>(2);
        map.put("key1", "val1");
        map.put("key2", "val2");
        delegateDecrypt.setEnvironment(gson.toJson(map));
        return delegateDecrypt;
    }

    private String getSchedulerData() {
        return "{\"schedulerFuncKey\":\"default/0-system-faasscheduler/$latest\"," +
                "\"schedulerIDList\":[\"2238fb12-0000-4000-8000-00abc0d9cc91\"]}";
    }

    private List<String> generateCallArgs() {
        Gson gson = new Gson();
        List<String> args = new ArrayList<>();
        args.add("{\"codeID\":\"\",\"config\":{\"functionID\":{\"cpp\":\"\"," +
                "\"python\":\"sn:cn:yrk:default:function:0-he-he:$latest\"}," +
                "\"jodID\":\"96f2fc5e-c9ab-4d83-9aa2-89579a29ff4a\",\"logLevel\":30,\"recycleTime\":2}," +
                "\"invokeType\":3,\"objectDescriptor\":{\"className\":\"\",\"functionName\":\"execute\"," +
                "\"moduleName\":\"faasexecutor\",\"srcLanguage\":\"python\",\"targetLanguage\":\"python\"}}");
        TestRequestEvent testRequestEvent = new TestRequestEvent("yuanrong", 1);
        CallRequest callRequest = new CallRequest();
        callRequest.setBody(testRequestEvent);
        callRequest.setHeader(new HashMap<String, String>(){
            {
                put(Constants.CFF_LOG_TYPE, "tail");
            }
        });
        args.add(gson.toJson(callRequest));
        return args;
    }

    public List<String> generateCallArgsWithHeader() {
        Gson gson = new Gson();
        List<String> args = new ArrayList<>();
        args.add("{\"codeID\":\"\",\"config\":{\"functionID\":{\"cpp\":\"\"," +
                "\"python\":\"sn:cn:yrk:12345678901234561234567890123456:function:0-he-he:$latest\"}," +
                "\"jodID\":\"96f2fc5e-c9ab-4d83-9aa2-89579a29ff4a\",\"logLevel\":30,\"recycleTime\":2}," +
                "\"invokeType\":3,\"objectDescriptor\":{\"className\":\"\",\"functionName\":\"execute\"," +
                "\"moduleName\":\"faasexecutor\",\"srcLanguage\":\"python\",\"targetLanguage\":\"python\"}}");
        TestRequestEvent testRequestEvent = new TestRequestEvent("yuanrong", 1);
        CallRequest callRequest = new CallRequest();
        callRequest.setBody(testRequestEvent);
        callRequest.setHeader(new HashMap<String, String>(){
            {
                put("Accept", "text/event-stream");
                put("X-Trace-Id", "test_trace_id");
            }
        });
        args.add(gson.toJson(callRequest));
        return args;
    }

    public List<String> generateErrorCallArgs() {
        Gson gson = new Gson();
        List<String> args = new ArrayList<>();
        args.add("{\"codeID\":\"\",\"config\":{\"functionID\":{\"cpp\":\"\"," +
                "\"python\":\"sn:cn:yrk:default:function:0-he-he:$latest\"}," +
                "\"jodID\":\"96f2fc5e-c9ab-4d83-9aa2-89579a29ff4a\",\"logLevel\":30,\"recycleTime\":2}," +
                "\"invokeType\":3,\"objectDescriptor\":{\"className\":\"\",\"functionName\":\"execute\"," +
                "\"moduleName\":\"faasexecutor\",\"srcLanguage\":\"python\",\"targetLanguage\":\"python\"}}");
        TestRequestEvent testRequestEvent = new TestRequestEvent("yuanrong", 0);
        args.add(gson.toJson(testRequestEvent));
        return args;
    }

    @Test
    public void testFaaSHandlerInitSuccess() throws Exception {
        FaaSHandler faaSHandler = new FaaSHandler();
        List<String> args = generateInitArgs();
        Assert.assertTrue(faaSHandler.faasInitHandler(args).getErrorMessage().contains("processed init handler successfully."));
    }

    @Test
    public void testFaaSHandlerInitWithDelegate() throws Exception {
        Gson gson = new Gson();
        DelegateDecrypt delegateDecrypt = new DelegateDecrypt();
        Map<String, String> map = new HashMap<>();
        map.put("spring_start_class", "com.inventory.InventoryApplication");
        delegateDecrypt.setEnvironment(gson.toJson(map));
        Map<String, String> envMap = new HashMap<>();
        envMap.put("ENV_DELEGATE_DECRYPT", gson.toJson(delegateDecrypt));
        setJavaProcessEnvMap(envMap);
        FaaSHandler faaSHandler = new FaaSHandler();
        List<String> args = new ArrayList<>();
        ContextImpl contextImpl = getContext();
        String contextConfig = gson.toJson(contextImpl);
        String udfEntry = "{\"userInitEntry\":\"org.yuanrong.executor.UserTestHandler.initializer\"," +
                "\"userCallEntry\":\"org.yuanrong.executor.UserTestHandler.contextHandler\"}";
        args.add(contextConfig);
        args.add(udfEntry);
        args.add(getSchedulerData());
        Assert.assertTrue(faaSHandler.faasInitHandler(args).getErrorMessage().contains("processed init handler successfully."));
        envMap.put("ENV_DELEGATE_DECRYPT", "");
        setJavaProcessEnvMap(envMap);
    }

    @Test
    public void testFaaSHandlerInitWithInstanceLabel() throws Exception {
        Gson gson = new Gson();
        DelegateDecrypt delegateDecrypt = new DelegateDecrypt();
        Map<String, String> map = new HashMap<>();
        map.put("spring_start_class", "com.inventory.InventoryApplication");
        delegateDecrypt.setEnvironment(gson.toJson(map));
        Map<String, String> envMap = new HashMap<>();
        envMap.put("ENV_DELEGATE_DECRYPT", gson.toJson(delegateDecrypt));
        setJavaProcessEnvMap(envMap);
        FaaSHandler faaSHandler = new FaaSHandler();
        List<String> args = new ArrayList<>();
        ContextImpl contextImpl = getContext();
        String contextConfig = gson.toJson(contextImpl);
        String udfEntry = "{\"userInitEntry\":\"org.yuanrong.executor.UserTestHandler.initializer\"," +
                "\"userCallEntry\":\"org.yuanrong.executor.UserTestHandler.instanceLabelHandler\", \"instanceLabel\":\"aaaaa\"}";
        args.add(contextConfig);
        args.add(udfEntry);
        args.add(getSchedulerData());
        Assert.assertTrue(faaSHandler.faasInitHandler(args).getErrorMessage().contains("processed init handler successfully."));
        List<String> callArgs = generateCallArgs();
        Assert.assertTrue(faaSHandler.faasCallHandler(callArgs).contains("aaaaa"));
    }

    @Test
    public void testFaaSHandlerInitFail() throws Exception {
        FaaSHandler faaSHandler = new FaaSHandler();
        Gson gson = new Gson();
        List<String> args = new ArrayList<>();
        ErrorInfo initErrorInfo = faaSHandler.faasInitHandler(args);
        Assert.assertEquals(FaasErrorCode.FAAS_INIT_ERROR.getCode(), initErrorInfo.getErrorCode().getValue());

        List<String> wrongArgs = generateErrorInitArgs(FaasErrorCode.INITIALIZE_FUNCTION_ERROR.getErrorMessage());
        initErrorInfo = faaSHandler.faasInitHandler(wrongArgs);
        Assert.assertEquals(FaasErrorCode.INITIALIZE_FUNCTION_ERROR.getCode(), initErrorInfo.getErrorCode().getValue());

        wrongArgs = generateErrorInitArgs("call " + FaasErrorCode.ENTRY_NOT_FOUND.getErrorMessage());
        initErrorInfo = faaSHandler.faasInitHandler(wrongArgs);
        Assert.assertEquals(FaasErrorCode.ENTRY_NOT_FOUND.getCode(), initErrorInfo.getErrorCode().getValue());

        wrongArgs = generateErrorInitArgs("init " + FaasErrorCode.ENTRY_NOT_FOUND.getErrorMessage());
        initErrorInfo = faaSHandler.faasInitHandler(wrongArgs);
        Assert.assertEquals(FaasErrorCode.ENTRY_NOT_FOUND.getCode(), initErrorInfo.getErrorCode().getValue());

        wrongArgs = generateErrorInitArgs("IllegalArgumentException");
        initErrorInfo = faaSHandler.faasInitHandler(wrongArgs);
        Assert.assertEquals(FaasErrorCode.INIT_FUNCTION_FAIL.getCode(), initErrorInfo.getErrorCode().getValue());

        wrongArgs = generateErrorInitArgs("InvocationTargetException");
        initErrorInfo = faaSHandler.faasInitHandler(wrongArgs);
        Assert.assertEquals(FaasErrorCode.INIT_FUNCTION_FAIL.getCode(), initErrorInfo.getErrorCode().getValue());
    }

    @Test
    public void testFaaSHandlerCallSuccess() throws Exception {
        FaaSHandler faaSHandler = new FaaSHandler();
        List<String> initArgs = generateInitArgs();
        faaSHandler.faasInitHandler(initArgs);
        List<String> callArgs = generateCallArgs();
        String response = faaSHandler.faasCallHandler(callArgs);
        Gson gson = new Gson();
        CallResponse response2 = gson.fromJson(response, CallResponse.class);
        Assert.assertEquals("true", response2.getBody().toString());
    }

    @Test
    public void testFaaSHandlerCallSuccessWithSerializeBody() throws Exception {
        FaaSHandler faaSHandler = new FaaSHandler();
        List<String> initArgs = generateInitArgs();
        faaSHandler.faasInitHandler(initArgs);
        Gson gson = new Gson();
        List<String> args = new ArrayList<>();
        args.add("{\"codeID\":\"\",\"config\":{\"functionID\":{\"cpp\":\"\"," +
                "\"python\":\"sn:cn:yrk:default:function:0-he-he:$latest\"}," +
                "\"jodID\":\"96f2fc5e-c9ab-4d83-9aa2-89579a29ff4a\",\"logLevel\":30,\"recycleTime\":2}," +
                "\"invokeType\":3,\"objectDescriptor\":{\"className\":\"\",\"functionName\":\"execute\"," +
                "\"moduleName\":\"faasexecutor\",\"srcLanguage\":\"python\",\"targetLanguage\":\"python\"}}");
        TestRequestEvent testRequestEvent = new TestRequestEvent("{\"id\":\"aaa\"}", 1);
        CallRequest callRequest = new CallRequest();
        callRequest.setBody(testRequestEvent);
        args.add(gson.toJson(callRequest));
        String response = (String) faaSHandler.faasCallHandler(args);
        CallResponse response2 = gson.fromJson(response, CallResponse.class);
        Assert.assertEquals("true", response2.getBody().toString());
    }

    @Test
    public void testFaaSHandlerCallFailedWithEventHeader() throws Exception {
        FaaSHandler faaSHandler = new FaaSHandler();
        List<String> initArgs = generateInitArgs();
        faaSHandler.faasInitHandler(initArgs);
        List<String> callArgs = generateCallArgsWithHeader();
        boolean isException = false;
        try {
            faaSHandler.faasCallHandler(callArgs);
        } catch (Throwable e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }

    @Test
    public void testFaaSHandlerCallWithJsonObjectBody() throws Exception {
        FaaSHandler faaSHandler = new FaaSHandler();
        Gson gson = new Gson();
        List<String> initArgs = new ArrayList<>();
        String contextConfig = gson.toJson(getContext());
        String udfEntry = "{\"userInitEntry\":\"org.yuanrong.executor.UserTestHandler.initializer\","
            + "\"userCallEntry\":\"org.yuanrong.executor.UserTestHandler::jsonHandler\"}";
        initArgs.add(contextConfig);
        initArgs.add(udfEntry);
        initArgs.add(getSchedulerData());
        faaSHandler.faasInitHandler(initArgs);

        String arg1 = "{\"codeID\":\"\",\"config\":{\"functionID\":{\"cpp\":\"\","
            + "\"python\":\"sn:cn:yrk:default:function:0-he-he:$latest\"},"
            + "\"jodID\":\"96f2fc5e-c9ab-4d83-9aa2-89579a29ff4a\",\"logLevel\":30,\"recycleTime\":2},"
            + "\"invokeType\":3,\"objectDescriptor\":{\"className\":\"\",\"functionName\":\"execute\","
            + "\"moduleName\":\"faasexecutor\",\"srcLanguage\":\"python\",\"targetLanguage\":\"python\"}}";
        List<String> callArgs1 = new ArrayList<>();
        callArgs1.add(arg1);
        CallRequest callRequest1 = new CallRequest.Builder().withBody("{}").build();
        callArgs1.add(gson.toJson(callRequest1));
        CallResponse response1 = gson.fromJson(faaSHandler.faasCallHandler(callArgs1), CallResponse.class);
        Assert.assertEquals("{}", response1.getBody().toString());

        List<String> callArgs2 = new ArrayList<>();
        callArgs2.add(arg1);
        CallRequest callRequest2 = new CallRequest.Builder().withBody("{\"id\":\"aaa\"}").build();
        callArgs2.add(gson.toJson(callRequest2));
        CallResponse response2 = gson.fromJson(faaSHandler.faasCallHandler(callArgs2), CallResponse.class);
        Assert.assertEquals("{\"id\":\"aaa\"}", response2.getBody().toString());
    }

    @Test
    public void testFaaSHandlerCallFail() throws Exception {
        Gson gson = new Gson();
        FaaSHandler faaSHandler = new FaaSHandler();
        List<String> initArgs = generateCallTimeoutArgs();
        Assert.assertTrue(faaSHandler.faasInitHandler(initArgs).getErrorMessage().contains("processed init handler successfully."));

        List<String> wrongArgs = new ArrayList<>();
        String resp1 = faaSHandler.faasCallHandler(wrongArgs);
        CallResponse response1 = gson.fromJson(resp1, CallResponse.class);
        Assert.assertEquals(String.valueOf(FaasErrorCode.FAAS_INIT_ERROR.getCode()), response1.getInnerCode());

        List<String> callArgs = generateErrorCallArgs();
        UDFManager.getUDFManager().registerInstance("userCallEntry", "");
        String resp3 = faaSHandler.faasCallHandler(callArgs);
        CallResponse response3 = gson.fromJson(resp3, CallResponse.class);
        Assert.assertEquals("4002", response3.getInnerCode());
    }

    @Test
    public void testFaaSHandlerCallUserFail() throws Exception {
        FaaSHandler faaSHandler = new FaaSHandler();
        Gson gson = new Gson();
        List<String> initArgs = new ArrayList<>();
        String contextConfig = gson.toJson(getContext());
        String udfEntry = "{\"userInitEntry\":\"org.yuanrong.executor.UserTestHandler.initializer\","
            + "\"userCallEntry\":\"org.yuanrong.executor.UserTestHandler::failedHandler\"}";
        initArgs.add(contextConfig);
        initArgs.add(udfEntry);
        initArgs.add(getSchedulerData());
        faaSHandler.faasInitHandler(initArgs);

        String arg1 = "{\"codeID\":\"\",\"config\":{\"functionID\":{\"cpp\":\"\","
            + "\"python\":\"sn:cn:yrk:default:function:0-he-he:$latest\"},"
            + "\"jodID\":\"96f2fc5e-c9ab-4d83-9aa2-89579a29ff4a\",\"logLevel\":30,\"recycleTime\":2},"
            + "\"invokeType\":3,\"objectDescriptor\":{\"className\":\"\",\"functionName\":\"execute\","
            + "\"moduleName\":\"faasexecutor\",\"srcLanguage\":\"python\",\"targetLanguage\":\"python\"}}";
        List<String> callArgs1 = new ArrayList<>();
        callArgs1.add(arg1);
        CallRequest callRequest1 = new CallRequest.Builder().withBody("abc").build();
        callArgs1.add(gson.toJson(callRequest1));
        CallResponse response1 = gson.fromJson(faaSHandler.faasCallHandler(callArgs1), CallResponse.class);
        Assert.assertEquals("4002", response1.getInnerCode());

    }

    @Test
    public void testFaaSHandlerResponseExceedSize() throws Exception {
        Gson gson = new Gson();
        FaaSHandler faaSHandler = new FaaSHandler();
        List<String> initArgs = generateCallResponseLargeSizeArgs();
        Assert.assertTrue(faaSHandler.faasInitHandler(initArgs).getErrorMessage().contains("processed init handler successfully."));

        List<String> callArgs = generateCallArgs();
        String response = faaSHandler.faasCallHandler(callArgs);
        CallResponse response2 = gson.fromJson(response, CallResponse.class);
        Assert.assertEquals("4004", response2.getInnerCode());
    }

    @Test
    public void testFaaSException() {
        FaaSException ex1 = new FaaSException("this is faas exception1");
        FaaSException ex2 = new FaaSException("this is faas exception2");
        Assert.assertEquals(ex1.equals(ex2), false);
        int hash1 = ex1.hashCode();
        int hash2 = ex2.hashCode();
        Assert.assertEquals(hash1 != hash2, true);
        Assert.assertTrue(ex1.equals(ex1));
        Assert.assertFalse(ex1.equals(null));
    }

    @Test
    public void testFaasCheckPointHandler() {
        FaaSHandler faaSHandler = new FaaSHandler();
        faaSHandler.faasCheckPointHandler(null);
    }

    @Test
    public void testFaasRecoverHandler() {
        FaaSHandler faaSHandler = new FaaSHandler();
        faaSHandler.faasRecoverHandler(null);
    }

    @Test
    public void testFaasShutDownHandlerDoNothing() {
        FaaSHandler faaSHandler = new FaaSHandler();
        ErrorInfo response = faaSHandler.faasShutDownHandler(0);
        Assert.assertTrue(response.getErrorMessage().contains("run shut down handler successfully."));
    }

    @Test
    public void testFaasShutDownHandlerSuccess() {
        FaaSHandler faaSHandler = new FaaSHandler();
        List<String> initArgs = generateInitArgsWiThPreStopHandler("Success");
        faaSHandler.faasInitHandler(initArgs);
        ErrorInfo response = faaSHandler.faasShutDownHandler(0);
        Assert.assertTrue(response.getErrorMessage().contains("run shut down handler successfully."));
    }

    @Test
    public void testFaasShutDownHandlerNotExist() {
        FaaSHandler faaSHandler = new FaaSHandler();
        List<String> initArgs = generateInitArgsWiThPreStopHandler("NoSuchMethodException");
        faaSHandler.faasInitHandler(initArgs);
        ErrorInfo errorInfo = faaSHandler.faasShutDownHandler(0);
        Assert.assertEquals(FaasErrorCode.ENTRY_NOT_FOUND.getCode(), errorInfo.getErrorCode().getValue());
    }

    @Test
    public void testFaasShutDownHandlerFailed() {
        FaaSHandler faaSHandler = new FaaSHandler();
        List<String> initArgs = generateInitArgsWiThPreStopHandler("InvocationTargetException");
        faaSHandler.faasInitHandler(initArgs);
        ErrorInfo errorInfo = faaSHandler.faasShutDownHandler(0);
        Assert.assertEquals(FaasErrorCode.FUNCTION_RUN_ERROR.getCode(), errorInfo.getErrorCode().getValue());
    }

    @Test
    public void testFaasSignalHandler() {
        FaaSHandler faaSHandler = new FaaSHandler();
        boolean isException = false;
        try {
            faaSHandler.faasSignalHandler(null);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testExecute() {
        boolean isException = false;
        FaaSHandler faaSHandler = new FaaSHandler();
        FunctionMeta meta = FunctionMeta.newBuilder()
            .setClassName("MockClass")
            .setFunctionName("mockMethod")
            .setSignature("()V")
            .build();
        ArrayList<ByteBuffer> byteBuffers = new ArrayList<>();
        ByteBuffer buffer = ByteBuffer.allocateDirect(10);
        buffer.put((byte)0x01);
        byteBuffers.add(buffer);
        try {
            ReturnType returnType = faaSHandler.execute(meta, Libruntime.InvokeType.InvokeFunction, byteBuffers);
            Assert.assertTrue(returnType.getErrorInfo().getErrorMessage().contains("faas get args invalid"));
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);


        ArrayList<ByteBuffer> byteBuffers2 = new ArrayList<>();
        ByteBuffer buffer2 = ByteBuffer.allocateDirect(10);
        buffer2.put((byte)0x01);
        byteBuffers2.add(buffer2);
        try {
            ReturnType returnType = faaSHandler.execute(meta, Libruntime.InvokeType.CreateInstance, byteBuffers2);
            Assert.assertTrue(returnType.getErrorInfo().getErrorMessage().contains("faas get args invalid"));
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        try {
            ReturnType returnType = faaSHandler.execute(meta, Libruntime.InvokeType.InvokeFunctionStateless,
                new ArrayList<>());
            Assert.assertEquals("call handler arg list is empty.",
                returnType.getErrorInfo().getErrorMessage());
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        ByteBuffer buffer3 = ByteBuffer.allocate(10);
        buffer3.put((byte)0x01);
        byteBuffers2.add(buffer3);
        try {
            ReturnType returnType = faaSHandler.execute(meta, Libruntime.InvokeType.GetNamedInstanceMeta, byteBuffers2);
            Assert.assertEquals("invalid invoke type", returnType.getErrorInfo().getErrorMessage());
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testShutdown() {
        FaaSHandler faaSHandler = new FaaSHandler();
        ErrorInfo errorInfo = faaSHandler.shutdown(10);
        Assert.assertEquals(ErrorCode.ERR_OK, errorInfo.getErrorCode());

        List<String> initArgs = generateInitArgsWiThPreStopHandler("Success");
        faaSHandler.faasInitHandler(initArgs);
        ErrorInfo errorInfo1 = faaSHandler.shutdown(10);
        Assert.assertEquals(ErrorCode.ERR_OK, errorInfo1.getErrorCode());

        List<String> initArgs2 = generateInitArgsWiThPreStopHandler("InterruptedException");
        faaSHandler.faasInitHandler(initArgs2);
        ErrorInfo errorInfo2 = faaSHandler.shutdown(1);
        Assert.assertEquals(FaasErrorCode.INVOKE_FUNCTION_TIMEOUT.getCode(), errorInfo2.getErrorCode().getValue());

        Pair<byte[], byte[]> pair = null;
        boolean isException = false;
        try {
            pair = faaSHandler.dumpInstance("testID");
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        faaSHandler.loadInstance(null,null);
        Assert.assertNotNull(pair);

        Assert.assertEquals(ErrorCode.ERR_OK, faaSHandler.recover().getErrorCode());
    }
}
