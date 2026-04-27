/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2024. All rights reserved.
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

import static org.mockito.ArgumentMatchers.anyBoolean;
import static org.mockito.ArgumentMatchers.anyInt;
import static org.mockito.ArgumentMatchers.anyList;
import static org.mockito.ArgumentMatchers.anyLong;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.when;

import org.yuanrong.CreateParam;
import org.yuanrong.GetParam;
import org.yuanrong.GetParams;
import org.yuanrong.GroupOptions;
import org.yuanrong.InvokeOptions;
import org.yuanrong.MSetParam;
import org.yuanrong.SetParam;
import org.yuanrong.api.InvokeArg;
import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.errorcode.Pair;
import org.yuanrong.exception.YRException;
import org.yuanrong.exception.LibRuntimeException;
import org.yuanrong.call.CppInstanceHandler;
import org.yuanrong.call.GoInstanceHandler;
import org.yuanrong.call.InstanceHandler;
import org.yuanrong.call.JavaInstanceHandler;
import org.yuanrong.jni.LibRuntime;
import org.yuanrong.libruntime.generated.Libruntime.ApiType;
import org.yuanrong.libruntime.generated.Libruntime.FunctionMeta;
import org.yuanrong.libruntime.generated.Libruntime.LanguageType;
import org.yuanrong.runtime.client.ObjectRef;
import org.yuanrong.storage.InternalWaitResult;
import org.yuanrong.stream.ProducerConfig;
import org.yuanrong.stream.SubscriptionConfig;
import org.yuanrong.utils.SdkUtils;

import org.junit.Assert;
import org.junit.BeforeClass;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.powermock.api.mockito.PowerMockito;
import org.powermock.core.classloader.annotations.PowerMockIgnore;
import org.powermock.core.classloader.annotations.PrepareForTest;
import org.powermock.core.classloader.annotations.SuppressStaticInitializationFor;
import org.powermock.modules.junit4.PowerMockRunner;

import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * The type Http task test.
 *
 * @since 2024 /03/06
 */
@RunWith(PowerMockRunner.class)
@PrepareForTest({LibRuntime.class})
@SuppressStaticInitializationFor({"org.yuanrong.jni.LibRuntime"})
@PowerMockIgnore("javax.management.*")
public class TestClusterModeRuntime {

    @BeforeClass
    public static void SetUp() {}

    @Test
    public void testIsDriver() {
        Runtime runtime = new ClusterModeRuntime();
        Assert.assertTrue(runtime.isDriver());
    }

    @Test
    public void testGetKVManager() {
        Runtime runtime = new ClusterModeRuntime();
        Assert.assertNotNull(runtime.getKVManager());
    }

    @Test
    public void testSaveAndGetRealInstanceId() throws Exception {
        PowerMockito.mockStatic(LibRuntime.class);
        Runtime runtime = new ClusterModeRuntime();
        when(LibRuntime.GetRealInstanceId(anyString())).thenReturn("instanceID");
        runtime.saveRealInstanceId("objID", "instanceID", new InvokeOptions());
        Assert.assertTrue(runtime.getRealInstanceId("objID").equals("instanceID"));
    }

    @Test
    public void testFinalizeAndExit() throws YRException {
        PowerMockito.mockStatic(LibRuntime.class);
        Runtime runtime = new ClusterModeRuntime();
        JavaInstanceHandler jHander = new JavaInstanceHandler("jInstanceId", "jFunctionId", "jClassName");
        CppInstanceHandler cHandler = new CppInstanceHandler("cInstanceId", "cFunctionId", "cClassName");
        InstanceHandler handler = new InstanceHandler("instance", ApiType.Function);
        boolean isException = false;
        try {
            runtime.Finalize("ctx", 0);
            runtime.Finalize();
            runtime.exit();
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testMultithreadCollectInstanceHandler() {
        PowerMockito.mockStatic(LibRuntime.class);
        Runtime runtime = new ClusterModeRuntime();
        int numberOfThreads = 100;
        boolean isException = false;
        ExecutorService executor = Executors.newFixedThreadPool(numberOfThreads);
        try {
            for (int i = 0; i < numberOfThreads; i++) {
                executor.submit(() -> {
                    InstanceHandler handler = new InstanceHandler("instance", ApiType.Function);
                });
            }
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);
        executor.shutdown();
        runtime.decreaseReference(Collections.singletonList("instanceID"));
        runtime.Finalize("ctx", 0);
        runtime.Finalize();
        runtime.exit();
    }

    @Test
    public void testKVWrite() throws Exception {
        String key = "key1";
        byte[] value = "value1".getBytes();
        PowerMockito.mockStatic(LibRuntime.class);
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_DATASYSTEM_FAILED, ModuleCode.DATASYSTEM, "");
        when(LibRuntime.KVWrite(anyString(), any(), any())).thenReturn(errorInfo);
        boolean isException = false;
        Runtime runtime = new ClusterModeRuntime();
        try {
            runtime.KVWrite(key, null, new SetParam());
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);
        isException = false;
        try {
            runtime.KVWrite(key, value, new SetParam());
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        when(LibRuntime.KVWrite(anyString(), any(), any())).thenReturn(new ErrorInfo());
        isException = false;
        try {
            runtime.KVWrite(key, value, new SetParam());
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testKVRead() throws Exception {
        String key = "key1";
        PowerMockito.mockStatic(LibRuntime.class);
        byte[] failedKey = ("key1").getBytes();
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_DATASYSTEM_FAILED, ModuleCode.DATASYSTEM, "");
        Pair<byte[], ErrorInfo> mockResult = new Pair<>(failedKey, errorInfo);
        when(LibRuntime.KVRead(anyString(), anyInt())).thenReturn(mockResult);
        boolean isException = false;
        Runtime runtime = new ClusterModeRuntime();
        try {
            runtime.KVRead(key, 100);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);
        isException = false;
        errorInfo = new ErrorInfo();
        Pair<byte[], ErrorInfo> mockResult2 = new Pair<>(failedKey, errorInfo);
        when(LibRuntime.KVRead(anyString(), anyInt())).thenReturn(mockResult2);
        try {
            runtime.KVRead(key, 100);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testKVReadList() throws Exception {
        List<String> keys = new ArrayList<String>();
        keys.add("key1");
        PowerMockito.mockStatic(LibRuntime.class);
        List<byte[]> failedKeys = new ArrayList<byte[]>();
        failedKeys.add(("key1").getBytes());
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_DATASYSTEM_FAILED, ModuleCode.DATASYSTEM, "");
        Pair<List<byte[]>, ErrorInfo> mockResult = new Pair<>(failedKeys, errorInfo);
        when(LibRuntime.KVRead(anyList(), anyInt(), anyBoolean())).thenReturn(mockResult);
        boolean isException = false;
        Runtime runtime = new ClusterModeRuntime();
        try {
            runtime.KVRead(keys, 100, false);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);
        isException = false;
        errorInfo = new ErrorInfo();
        Pair<List<byte[]>, ErrorInfo> mockResult2 = new Pair<>(failedKeys, errorInfo);
        when(LibRuntime.KVRead(anyList(), anyInt(), anyBoolean())).thenReturn(mockResult2);
        try {
            List<byte[]> res = runtime.KVRead(keys, 100, false);
            Assert.assertTrue(res.size() > 0);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testKVGetWithParam() throws Exception {
        List<String> keys = new ArrayList<String>(){{add("key1");}};
        PowerMockito.mockStatic(LibRuntime.class);
        List<byte[]> vals = new ArrayList<byte[]>();
        vals.add(("val1").getBytes());
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_DATASYSTEM_FAILED, ModuleCode.DATASYSTEM, "");
        Pair<List<byte[]>, ErrorInfo> mockResult = new Pair<>(vals, errorInfo);
        when(LibRuntime.KVGetWithParam(anyList(), any(), anyInt())).thenReturn(mockResult);
        GetParam param = new GetParam.Builder().offset(0).size(2).build();
        List<GetParam> paramList = new ArrayList<GetParam>(){{add(param);}};
        GetParams params = new GetParams.Builder().getParams(paramList).build();
        boolean isException = false;
        Runtime runtime = new ClusterModeRuntime();
        try {
            runtime.KVGetWithParam(keys, params, 10);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);
        isException = false;
        errorInfo = new ErrorInfo();
        Pair<List<byte[]>, ErrorInfo> mockResult2 = new Pair<>(vals, errorInfo);
        when(LibRuntime.KVGetWithParam(anyList(), any(), anyInt())).thenReturn(mockResult2);
        try {
            List<byte[]> res = runtime.KVGetWithParam(keys, params, 10);
            Assert.assertTrue(res.size() > 0);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testKVdel() throws Exception {
        String key = "key1";
        PowerMockito.mockStatic(LibRuntime.class);
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_DATASYSTEM_FAILED, ModuleCode.DATASYSTEM, "");
        when(LibRuntime.KVDel(anyString())).thenReturn(errorInfo);
        boolean isException = false;
        Runtime runtime = new ClusterModeRuntime();
        try {
            runtime.KVDel(key);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);
        isException = false;
        errorInfo = new ErrorInfo();
        when(LibRuntime.KVDel(anyString())).thenReturn(errorInfo);
        try {
            runtime.KVDel(key);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testKVdelList() throws Exception {
        List<String> keys = new ArrayList<String>();
        keys.add("key1");

        // mock 'LibRuntime.KVDel'
        PowerMockito.mockStatic(LibRuntime.class);
        List<String> failedKeys = new ArrayList<String>();
        failedKeys.add("key1");
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_DATASYSTEM_FAILED, ModuleCode.DATASYSTEM, "");
        Pair<List<String>, ErrorInfo> mockResult = new Pair<>(failedKeys, errorInfo);
        when(LibRuntime.KVDel(anyList())).thenReturn(mockResult);

        // assert that there is not YRException thrown by KVDel method and returned failed keys
        Runtime runtime = new ClusterModeRuntime();
        List<String> result = runtime.KVDel(keys);
        Assert.assertEquals(failedKeys, result);
    }

    @Test
    public void testLoadState() throws Exception {
        PowerMockito.mockStatic(LibRuntime.class);
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME, "");
        when(LibRuntime.LoadState(anyInt())).thenReturn(errorInfo);
        boolean isException = false;
        Runtime runtime = new ClusterModeRuntime();
        try {
            runtime.loadState(100);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        when(LibRuntime.LoadState(anyInt())).thenReturn(new ErrorInfo());
        isException = false;
        try {
            runtime.loadState(100);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testSaveState() throws Exception {
        PowerMockito.mockStatic(LibRuntime.class);
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME, "");
        when(LibRuntime.SaveState(anyInt())).thenReturn(errorInfo);
        boolean isException = false;
        Runtime runtime = new ClusterModeRuntime();
        try {
            runtime.saveState(100);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        when(LibRuntime.SaveState(anyInt())).thenReturn(new ErrorInfo());
        isException = false;
        try {
            runtime.saveState(100);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testGet() throws Exception {
        ClusterModeRuntime runtime = new ClusterModeRuntime();
        boolean isException = false;
        try {
            runtime.get(null, 10);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        ObjectRef ref = new ObjectRef("objID", String.class);
        PowerMockito.mockStatic(LibRuntime.class);
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.RUNTIME, "");
        List<byte[]> ok = new ArrayList<byte[]>();
        ok.add("result1".getBytes(StandardCharsets.UTF_8));
        ok.add("result2".getBytes(StandardCharsets.UTF_8));
        Pair<ErrorInfo, List<byte[]>> getRes = new Pair<ErrorInfo,List<byte[]>>(errorInfo, ok);
        when(LibRuntime.Get(anyList(), anyInt(), anyBoolean())).thenReturn(getRes);
        Object result = runtime.get(ref, 10);
        Assert.assertNotNull(result);

        when(LibRuntime.Get(anyList(), anyInt(), anyBoolean())).thenThrow(new LibRuntimeException("err"));
        isException = false;
        try {
            runtime.get(ref, 10);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }

    @Test
    public void testGetMulti() throws Exception {
        List<ObjectRef> refs = new ArrayList<ObjectRef>();
        List<String> objIds = Arrays.asList("1", "2");
        for (String i: objIds) {
            refs.add(new ObjectRef(i, String.class));
        }
        ClusterModeRuntime runtime = new ClusterModeRuntime();
        boolean isException = false;
        try {
            runtime.get(refs, -10, false);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        // Mock a 'LibRuntime.Wait'
        PowerMockito.mockStatic(LibRuntime.class);
        List<String> readyIds = Arrays.asList("2", "1");
        List<String> unreadyIds = new ArrayList<String>();
        Map<String, ErrorInfo> exceptionIds = new HashMap<String, ErrorInfo>();
        InternalWaitResult waitResult = new InternalWaitResult(readyIds, unreadyIds, exceptionIds);
        when(LibRuntime.Wait(anyList(), anyInt(), anyInt())).thenReturn(waitResult);

        // Mock 'LibRuntime.Get' (check getting "objIds")
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.RUNTIME, "");
        List<byte[]> ok = new ArrayList<byte[]>();
        ok.add("result1".getBytes(StandardCharsets.UTF_8));
        ok.add("result2".getBytes(StandardCharsets.UTF_8));
        Pair<ErrorInfo, List<byte[]>> getRes = new Pair<ErrorInfo,List<byte[]>>(errorInfo, ok);
        when(LibRuntime.Get(eq(objIds), anyInt(), anyBoolean())).thenReturn(getRes);
        List<Object> result = runtime.get(refs, 10, false);
        Assert.assertEquals(2, result.size());

        when(LibRuntime.Get(eq(objIds), anyInt(), anyBoolean())).thenThrow(new LibRuntimeException("err"));
        isException = false;
        try {
            runtime.get(refs, 10, false);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }

    @Test
    public void testPut() throws Exception {
        PowerMockito.mockStatic(LibRuntime.class);
        ClusterModeRuntime runtime = new ClusterModeRuntime();
        Pair<ErrorInfo, String> mockRes = new Pair<ErrorInfo, String>(new ErrorInfo(), "success");
        when(LibRuntime.Put(any(), anyList())).thenReturn(mockRes);
        Object obj = new String("obj");
        boolean isException = false;
        try {
            runtime.put(obj, 10L, 20L);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        when(LibRuntime.PutWithParam(any(), anyList(), any())).thenReturn(mockRes);
        try {
            runtime.put(obj, 10L, 20L, new CreateParam());
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        when(LibRuntime.Put(any(), anyList())).thenThrow(new LibRuntimeException("error occurred"));
        try {
            runtime.put(obj, 10L, 20L);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        when(LibRuntime.PutWithParam(any(), anyList(), any())).thenThrow(new LibRuntimeException("error occurred"));
        isException = false;
        try {
            runtime.put(obj, 10L, 20L, new CreateParam());
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }

    @Test
    public void testWait() throws Exception {
        PowerMockito.mockStatic(LibRuntime.class);
        ClusterModeRuntime runtime = new ClusterModeRuntime();
        List<String> repeatedObjIds = new ArrayList<String>() {
            {
                add("obj1");
                add("obj1");
            }
        };
        List<String> normalObjIds = new ArrayList<String>() {
            {
                add("obj1");
                add("obj2");
            }
        };
        boolean isException = false;
        try {
            runtime.wait(normalObjIds, 2, -10);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);
        isException = false;
        try {
            runtime.wait(repeatedObjIds, 2, 10);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);
        isException = false;
        try {
            runtime.wait(normalObjIds, 0, 10);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        when(LibRuntime.Wait(anyList(), anyInt(), anyInt())).thenReturn(null);
        isException = false;
        try {
            runtime.wait(normalObjIds, 2, 10);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        List<String> readyIds = Arrays.asList("2", "1");
        List<String> unreadyIds = new ArrayList<String>();
        Map<String, ErrorInfo> exceptionIds = new HashMap<String, ErrorInfo>() {
            {
                put("err", new ErrorInfo(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME, ""));
            }
        };
        InternalWaitResult waitResult = new InternalWaitResult(readyIds, unreadyIds, exceptionIds);
        when(LibRuntime.Wait(anyList(), anyInt(), anyInt())).thenReturn(waitResult);
        isException = false;
        try {
            runtime.wait(normalObjIds, 2, 10);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        exceptionIds = new HashMap<String, ErrorInfo>();
        waitResult = new InternalWaitResult(readyIds, unreadyIds, exceptionIds);
        when(LibRuntime.Wait(anyList(), anyInt(), anyInt())).thenReturn(waitResult);
        InternalWaitResult res = runtime.wait(normalObjIds, 2, 10);
        Assert.assertNotNull(res);

        exceptionIds = new HashMap<String, ErrorInfo>() {
            {
                put("err", null);
            }
        };
        waitResult = new InternalWaitResult(readyIds, unreadyIds, exceptionIds);
        when(LibRuntime.Wait(anyList(), anyInt(), anyInt())).thenReturn(waitResult);
        isException = false;
        try {
            runtime.wait(normalObjIds, 2, 10);
        } catch (YRException e) {
            isException = true;
            Assert.assertTrue(e.getMessage().contains("errorInfo is null"));
        }
        Assert.assertTrue(isException);
    }

    @Test
    public void testInstance() throws Exception {
        Runtime runtime = new ClusterModeRuntime();
        PowerMockito.mockStatic(LibRuntime.class);
        FunctionMeta meta = FunctionMeta.newBuilder().setClassName("className").setFunctionName("methodName")
                                        .setApiType(ApiType.Function).setLanguage(LanguageType.Java).build();
        String arg = new String("arg");
        List<InvokeArg> args = SdkUtils.packInvokeArgs(arg);
        InvokeOptions opts = new InvokeOptions();
        Pair<ErrorInfo, String> mockRes = new Pair<ErrorInfo, String>(new ErrorInfo(), "objID");
        when(LibRuntime.CreateInstance(any(), anyList(), any())).thenReturn(mockRes);
        String instanceID = runtime.createInstance(meta, args, opts);
        Assert.assertEquals("objID", instanceID);

        when(LibRuntime.CreateInstance(any(), anyList(), any())).thenThrow(
            new LibRuntimeException("error occurred"));
        boolean isException = false;
        try {
            runtime.createInstance(meta, args, opts);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        when(LibRuntime.InvokeInstance(any(), anyString(), anyList(), any())).thenReturn(mockRes);
        String objID = runtime.invokeInstance(meta, instanceID, args, opts);
        Assert.assertEquals("objID", objID);

        when(LibRuntime.InvokeInstance(any(), anyString(), anyList(), any())).thenThrow(
            new LibRuntimeException("error occurred"));
        isException = false;
        try {
            runtime.invokeInstance(meta, instanceID, args, opts);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        when(LibRuntime.Kill(anyString())).thenReturn(new ErrorInfo());
        isException = false;
        try {
            runtime.terminateInstance(instanceID);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        when(LibRuntime.KillSync(anyString())).thenReturn(new ErrorInfo());
        isException = false;
        try {
            runtime.terminateInstanceSync(instanceID);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testGroupCreate() throws Exception {
        Runtime runtime = new ClusterModeRuntime();
        GroupOptions opt = new GroupOptions();
        opt.setTimeout(-2);
        boolean isException = false;
        try {
            runtime.groupCreate("group1", opt);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        PowerMockito.mockStatic(LibRuntime.class);
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME, "");
        when(LibRuntime.GroupCreate(anyString(), any())).thenReturn(errorInfo);
        opt.setTimeout(100);
        isException = false;
        try {
            runtime.groupCreate("group1", opt);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        errorInfo = new ErrorInfo();
        when(LibRuntime.GroupCreate(anyString(), any())).thenReturn(errorInfo);
        opt.setTimeout(100);
        isException = false;
        try {
            runtime.groupCreate("group1", opt);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testGroupWait() throws Exception {
        Runtime runtime = new ClusterModeRuntime();
        PowerMockito.mockStatic(LibRuntime.class);
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME, "");
        when(LibRuntime.GroupWait(anyString())).thenReturn(errorInfo);
        boolean isException = false;
        try {
            runtime.groupWait("group1");
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);
        when(LibRuntime.GroupWait(anyString())).thenReturn(new ErrorInfo());
        isException = false;
        try {
            runtime.groupWait("group1");
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testGroupTerminate() throws YRException {
        Runtime runtime = new ClusterModeRuntime();
        PowerMockito.mockStatic(LibRuntime.class);
        runtime.groupTerminate("group1");
    }

    @Test
    public void testCreateStreamProducer() throws Exception {
        ProducerConfig cfg = new ProducerConfig();
        Runtime runtime = new ClusterModeRuntime();
        PowerMockito.mockStatic(LibRuntime.class);
        when(LibRuntime.CreateStreamProducerWithConfig(anyString(), any())).thenReturn(10086L);
        boolean isException = false;
        try {
            runtime.createStreamProducer("stream", cfg);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
        when(LibRuntime.CreateStreamProducerWithConfig(anyString(), any())).thenThrow(new LibRuntimeException("err"));
        isException = false;
        try {
            runtime.createStreamProducer("stream", cfg);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }

    @Test
    public void testCreateStreamProducerFailed() throws Exception {
        Runtime runtime = new ClusterModeRuntime();
        boolean isException = false;
        try {
            ProducerConfig cfg = new ProducerConfig();
            cfg.setMaxStreamSize(-1L);
            runtime.createStreamProducer("aaaaaaaa", cfg);
        } catch (YRException e) {
            Assert.assertTrue(e.getErrorCode().toString().equals("1001"));
            Assert.assertTrue(e.getErrorMessage().contains("is invalid, expect >= 0"));
            isException = true;
        }
        Assert.assertTrue(isException);
        isException = false;
        try {
            ProducerConfig cfg = new ProducerConfig();
            cfg.setRetainForNumConsumers(-1L);
            runtime.createStreamProducer("aaaaaaaa", cfg);
        } catch (YRException e) {
            Assert.assertTrue(e.getErrorCode().toString().equals("1001"));
            Assert.assertTrue(e.getErrorMessage().contains("is invalid, expect >= 0"));
            isException = true;
        }
        Assert.assertTrue(isException);
        isException = false;
        try {
            ProducerConfig cfg = new ProducerConfig();
            cfg.setReserveSize(-1L);
            runtime.createStreamProducer("aaaaaaaa", cfg);
        } catch (YRException e) {
            Assert.assertTrue(e.getErrorCode().toString().equals("1001"));
            Assert.assertTrue(e.getErrorMessage().contains("is invalid, expect >= 0"));
            isException = true;
        }
        Assert.assertTrue(isException);
    }

    @Test
    public void testCreateStreamConsumer() throws Exception {
        SubscriptionConfig cfg = new SubscriptionConfig();
        Runtime runtime = new ClusterModeRuntime();
        PowerMockito.mockStatic(LibRuntime.class);
        when(LibRuntime.CreateStreamConsumer(anyString(), anyString(), any(), anyBoolean())).thenReturn(10086L);
        boolean isException = false;
        try {
            runtime.createStreamConsumer("stream", cfg, false);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
        when(LibRuntime.CreateStreamConsumer(anyString(), anyString(), any(), anyBoolean()))
                .thenThrow(new LibRuntimeException("err"));
        isException = false;
        try {
            runtime.createStreamConsumer("stream", cfg, false);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }

    @Test
    public void testDeleteStream() throws Exception {
        Runtime runtime = new ClusterModeRuntime();
        PowerMockito.mockStatic(LibRuntime.class);
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME, "");
        when(LibRuntime.DeleteStream(anyString())).thenReturn(errorInfo);
        boolean isException = false;
        try {
            runtime.deleteStream("stream");
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);
        when(LibRuntime.DeleteStream(anyString())).thenReturn(new ErrorInfo());
        isException = false;
        try {
            runtime.deleteStream("stream");
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testQueryGlobalConsumersNum() throws Exception {
        Runtime runtime = new ClusterModeRuntime();
        PowerMockito.mockStatic(LibRuntime.class);
        when(LibRuntime.QueryGlobalConsumersNum(anyString())).thenReturn(10086L);
        boolean isException = false;
        try {
            runtime.queryGlobalConsumersNum("stream");
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
        when(LibRuntime.QueryGlobalConsumersNum(anyString())).thenThrow(new LibRuntimeException("err"));
        isException = false;
        try {
            runtime.queryGlobalConsumersNum("stream");
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }

    @Test
    public void testQueryGlobalProducersNum() throws Exception {
        Runtime runtime = new ClusterModeRuntime();
        PowerMockito.mockStatic(LibRuntime.class);
        when(LibRuntime.QueryGlobalProducersNum(anyString())).thenReturn(10086L);
        boolean isException = false;
        try {
            runtime.queryGlobalProducersNum("stream");
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
        when(LibRuntime.QueryGlobalProducersNum(anyString())).thenThrow(new LibRuntimeException("err"));
        isException = false;
        try {
            runtime.queryGlobalProducersNum("stream");
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }

    @Test
    public void testKVMSetTx() throws Exception {
        PowerMockito.mockStatic(LibRuntime.class);
        ClusterModeRuntime runtime = new ClusterModeRuntime();
        ErrorInfo mockRes = new ErrorInfo(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME, "");
        when(LibRuntime.KVMSetTx(anyList(), anyList(), any())).thenReturn(mockRes);
        boolean isException = false;
        try {
            runtime.KVMSetTx(new ArrayList<>(), new ArrayList<>(), new MSetParam());
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        when(LibRuntime.KVMSetTx(anyList(), anyList(), any())).thenThrow(new LibRuntimeException("error occurred"));
        isException = false;
        try {
            runtime.KVMSetTx(new ArrayList<>(), new ArrayList<>(), new MSetParam());
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }

    @Test
    public void testGetInstanceRoute() throws Exception {
        PowerMockito.mockStatic(LibRuntime.class);
        ClusterModeRuntime runtime = new ClusterModeRuntime();
        when(LibRuntime.GetInstanceRoute(anyString())).thenReturn("getInstanceRoute");
        boolean isException = false;
        try {
            runtime.getInstanceRoute("objId");
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        when(LibRuntime.GetInstanceRoute(anyString())).thenThrow(new LibRuntimeException("error occurred"));
        try {
            runtime.getInstanceRoute("objId");
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }
}
