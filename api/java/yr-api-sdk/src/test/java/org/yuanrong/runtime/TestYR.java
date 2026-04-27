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

package org.yuanrong.runtime;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyBoolean;
import static org.mockito.ArgumentMatchers.anyInt;
import static org.mockito.ArgumentMatchers.anyList;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.when;

import org.yuanrong.Config;
import org.yuanrong.ConfigManager;
import org.yuanrong.api.ClientInfo;
import org.yuanrong.api.Node;
import org.yuanrong.api.YR;
import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.errorcode.Pair;
import org.yuanrong.exception.YRException;
import org.yuanrong.jni.LibRuntime;
import org.yuanrong.jobexecutor.YRJobParam;
import org.yuanrong.runtime.client.ObjectRef;
import org.yuanrong.storage.InternalWaitResult;
import org.yuanrong.storage.WaitResult;
import org.yuanrong.stream.Consumer;
import org.yuanrong.stream.Producer;
import org.yuanrong.stream.ProducerConfig;
import org.yuanrong.stream.SubscriptionConfig;
import org.yuanrong.utils.SdkUtils;

import org.junit.Assert;
import org.junit.Before;
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

/**
 * The type yr test.
 *
 * @since 2024 /03/21
 */
@RunWith(PowerMockRunner.class)
@PrepareForTest({SdkUtils.class, LibRuntime.class})
@SuppressStaticInitializationFor({"org.yuanrong.jni.LibRuntime"})
@PowerMockIgnore("javax.management.*")
public class TestYR {
    Config conf = new Config(
            "sn:cn:yrk:default:function:0-crossyrlib-helloworld:$latest",
            "127.0.0.0",
            "127.0.0.0",
            "sn:cn:yrk:default:function:0-test-hello:$latest",
            true);

    @Before
    public void init() throws Exception {
        PowerMockito.mockStatic(LibRuntime.class);
        when(LibRuntime.IsInitialized()).thenReturn(true);
        when(LibRuntime.Init(any())).thenReturn(new ErrorInfo());
    }

    @Test
    public void testInitTwice() throws Exception {
        boolean isException = false;
        try {
            ClientInfo info = YR.init(conf);
            Assert.assertNotEquals(info.getJobID(), "");
        } catch(Exception e) {
            e.printStackTrace();
            isException = true;
        }
        Assert.assertFalse(isException);

        try {
            YR.init(conf);
        } catch(YRException e) {
            isException = true;
            Assert.assertEquals(e.getErrorCode(), ErrorCode.ERR_INCORRECT_INIT_USAGE);
            Assert.assertEquals(e.getModuleCode(), ModuleCode.RUNTIME);
        }
        Assert.assertTrue(isException);
        YR.Finalize();
    }

    @Test
    public void testInitAfterInitFailure() throws Exception {
        Config invalidConf = new Config(
            "sn:cn:yrk:default:function:0-crossyrlib-helloworld:$latest",
            "127.0.0.0",
            "127.0.0.0",
            "sn:cn:yrk:default:function:0-test-hello:$latest",
            true);
        invalidConf.setDataSystemAddress("invalidaddress");
        boolean isException = false;
        try {
            YR.init(invalidConf);
        } catch (Exception e) {
            isException = true;
            Assert.assertTrue(e.toString().contains("dataSystemAddress is invalid"));
        }
        Assert.assertTrue(isException);

        isException = false;
        invalidConf.setDataSystemAddress("127.0.0.0");
        try {
            YR.init(invalidConf);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);
        Config conf = new Config("sn:cn:yrk:default:function:0-st-stjava:$latest",
                "127.0.0.1", 304822, "127.0.0.1", 290811, "", true);
        try {
            YR.init(invalidConf);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);
        YR.Finalize();
    }

    @Test
    public void testFinalize() {
        boolean isException = false;
        try {
            YR.init(conf);
            YR.Finalize();
        } catch(YRException e) {
            e.printStackTrace();
            isException = true;
        }
        Assert.assertFalse(isException);

        when(LibRuntime.IsInitialized()).thenReturn(false);
        try {
            YR.Finalize();
        } catch(YRException e) {
            Assert.assertEquals(ErrorCode.ERR_FINALIZED, e.getErrorCode());
            Assert.assertEquals(ModuleCode.RUNTIME, e.getModuleCode());
        }
    }

    /**
     * Description:
     *  Test 'Exit' throws exception out of the runtime.
     * Steps:
     *   1. Calls 'Exit' when 'inCluster' is true.
     *   2. Sets 'inCluster' false.
     *   3. Calls 'Exit' again.
     * Expectation:
     *   When 'inCluster' is true, call 'Exit' would not cause exception,
     *   and throws YRException otherwise.
     */
    @Test
    public void testExit() throws Exception {
        when(LibRuntime.IsInitialized()).thenReturn(true);
        boolean isException = false;
        try {
            YR.init(conf);
            YR.exit();
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        conf.setInCluster(false);
        ConfigManager.getInstance().init(conf);
        try {
            YR.exit();
        } catch (YRException e) {
            isException = true;
            Assert.assertTrue(e.getErrorMessage().contains("Not support exit out of cluster"));
        }
        Assert.assertTrue(isException);
        // make other cases available
        YR.Finalize();
    }

    @Test
    public void testTenantContext() {
        Config ctxConf = Config.builder()
                            .functionURN("sn:cn:yrk:default:function:0-j-a:$latest")
                            .serverAddress("127.0.0.1")
                            .serverAddressPort(31222)
                            .dataSystemAddress("127.0.0.1")
                            .dataSystemAddressPort(31501)
                            .isInCluster(false)
                            .tenantId("tenantId1")
                            .enableSetContext(true)
                            .build();
        boolean isException = false;
        String ctx1 = "";
        String ctx2 = "";
        try {
            ClientInfo info = YR.init(ctxConf);
            ctx1 = info.getContext();
            Assert.assertNotEquals(info.getContext(), "");
        } catch(YRException e) {
            e.printStackTrace();
            isException = true;
        }
        Assert.assertFalse(isException);

        try {
            ClientInfo info = YR.init(ctxConf);
            ctx2 = info.getContext();
            Assert.assertNotEquals(info.getContext(), "");
        } catch(YRException e) {
            e.printStackTrace();
            isException = true;
        }
        Assert.assertFalse(isException);

        try {
            YR.setContext(ctx1);
        } catch(YRException e) {
            e.printStackTrace();
            isException = true;
        }
        Assert.assertFalse(isException);

        try {
            YR.finalize(ctx1);
            YR.finalize(ctx2);
        } catch(YRException e) {
            e.printStackTrace();
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testTenantContextFailed() {
        Config ctxConf = Config.builder()
                            .functionURN("sn:cn:yrk:default:function:0-j-a:$latest")
                            .serverAddress("127.0.0.1")
                            .serverAddressPort(31222)
                            .dataSystemAddress("127.0.0.1")
                            .dataSystemAddressPort(31501)
                            .isInCluster(false)
                            .tenantId("tenantId1")
                            .isThreadLocal(true)
                            .enableSetContext(true)
                            .build();
        String ctx = "ctxNotExist";
        boolean isException = false;
        try {
            ClientInfo info = YR.init(ctxConf);
            Assert.assertEquals(0, 1);
        } catch(YRException e) {
            e.printStackTrace();
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            YR.setContext(ctx);
            Assert.assertEquals(0, 1);
        } catch(YRException e) {
            e.printStackTrace();
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            YR.finalize(ctx);
            Assert.assertEquals(0, 1);
        } catch(YRException e) {
            e.printStackTrace();
            isException = true;
        }
        Assert.assertTrue(isException);
    }

    @Test
    public void testStream() throws YRException {
        YR.init(conf);
        ProducerConfig pCfg = new ProducerConfig();
        Producer producer = YR.createProducer("streamName", pCfg);
        Assert.assertNotNull(producer);
        producer = YR.createProducer("streamName");
        Assert.assertNotNull(producer);

        SubscriptionConfig sCfg = new SubscriptionConfig();
        Consumer consumer = YR.subscribe("streamName", sCfg);
        Assert.assertNotNull(consumer);
        consumer = YR.subscribe("streamName", sCfg, false);
        Assert.assertNotNull(consumer);

        when(LibRuntime.DeleteStream(anyString())).thenReturn(new ErrorInfo());
        YR.deleteStream("streamName");
        YR.Finalize();
    }

    @Test
    public void testState() throws Exception {
        YR.init(conf);
        when(LibRuntime.LoadState(anyInt())).thenReturn(new ErrorInfo());
        when(LibRuntime.SaveState(anyInt())).thenReturn(new ErrorInfo());
        boolean isException = false;
        try {
            YR.saveState(20);
            YR.saveState();
            YR.loadState(20);
            YR.loadState();
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);
        YR.Finalize();
    }

    @Test
    public void testPutGet() throws Exception {
        YR.init(conf);
        Pair<ErrorInfo, String> mockRes = new Pair<ErrorInfo, String>(new ErrorInfo(), "success");
        when(LibRuntime.Put(any(), anyList())).thenReturn(mockRes);
        List<byte[]> ok = new ArrayList<byte[]>();
        ok.add("result1".getBytes(StandardCharsets.UTF_8));
        ok.add("result2".getBytes(StandardCharsets.UTF_8));
        ErrorInfo errorInfo = new ErrorInfo();
        Pair<ErrorInfo, List<byte[]>> getRes = new Pair<ErrorInfo,List<byte[]>>(errorInfo, ok);
        when(LibRuntime.Get(anyList(), anyInt(), anyBoolean())).thenReturn(getRes);

        ObjectRef ref = YR.put(10);
        Assert.assertNotNull(ref);
        Object obj = YR.get(ref, 10);
        Assert.assertNotNull(obj);
        obj = YR.get(Arrays.asList(ref), 10);
        Assert.assertNotNull(obj);
        obj = YR.get(Arrays.asList(ref), 10, false);
        Assert.assertNotNull(obj);
        boolean isException = false;
        try {
            YR.put(ref);
        } catch(IllegalArgumentException e) {
            e.printStackTrace();
            isException = true;
        }
        Assert.assertTrue(isException);
        YR.Finalize();
    }

    @Test
    public void testWait() throws Exception {
        YR.init(conf);
        boolean isException = false;
        try {
            YR.wait(null, 2, 10);
        } catch(YRException e) {
            e.printStackTrace();
            isException = true;
        }
        Assert.assertTrue(isException);

        List<String> readyIds = Arrays.asList("2", "1");
        List<String> unreadyIds = new ArrayList<String>();
        Map<String, ErrorInfo> exceptionIds = new HashMap<String, ErrorInfo>();
        InternalWaitResult waitResult = new InternalWaitResult(readyIds, unreadyIds, exceptionIds);
        when(LibRuntime.Wait(anyList(), anyInt(), anyInt())).thenReturn(waitResult);

        ObjectRef ref1 = new ObjectRef("objID1", String.class);
        ObjectRef ref2 = new ObjectRef("objID2", String.class);
        WaitResult res = YR.wait(new ArrayList<ObjectRef>(){{add(ref1);add(ref2);}}, 2, 10);
        Assert.assertTrue(res.getReady().size() == 2);
        YR.Finalize();
    }

    @Test
    public void testJobs() throws Exception {
        YR.init(conf);
        Pair<ErrorInfo, String> mockRes = new Pair<ErrorInfo, String>(new ErrorInfo(), "objID");
        when(LibRuntime.CreateInstance(any(), anyList(), any())).thenReturn(mockRes);
        when(LibRuntime.InvokeInstance(any(), anyString(), anyList(), any())).thenReturn(mockRes);
        List<String> readyIds = Arrays.asList("2", "1");
        List<String> unreadyIds = new ArrayList<String>();
        Map<String, ErrorInfo> exceptionIds = new HashMap<String, ErrorInfo>();
        InternalWaitResult waitResult = new InternalWaitResult(readyIds, unreadyIds, exceptionIds);
        when(LibRuntime.Wait(anyList(), anyInt(), anyInt())).thenReturn(waitResult);
        when(LibRuntime.GetRealInstanceId(anyString())).thenReturn("instanceID");
        List<byte[]> ok = new ArrayList<byte[]>();
        ok.add("result1".getBytes(StandardCharsets.UTF_8));
        ok.add("result2".getBytes(StandardCharsets.UTF_8));
        Pair<ErrorInfo, List<byte[]>> getRes = new Pair<ErrorInfo,List<byte[]>>(new ErrorInfo(), ok);
        when(LibRuntime.Get(anyList(), anyInt(), anyBoolean())).thenReturn(getRes);

        YRJobParam param = new YRJobParam();
        ArrayList<String> entryPoints = new ArrayList<String>(){
            {
                add("java1.8");
                add("/home/snuser/java");
            }
        };
        param.setEntryPoint(entryPoints);
        param.setJobName("jobName");
        param.setLocalCodePath("DEFAULT_LOCAL_PATH");

        String res = YR.submitJob(param);
        Assert.assertNotNull(res);

        YR.Finalize();
    }

    @Test
    public void testNodes() throws Exception {
        YR.init(conf);
        Node node = new Node();
        String nodeId = "function-agent-127.0.0.1";
        node.setId(nodeId);
        node.setAlive(true);
        Map<String, Float> resources = new HashMap<>();
        resources.put("CPU", 1000F);
        resources.put("Memory", 1000F);
        node.setResources(resources);
        Map<String, List<String>> labels = new HashMap<>();
        List<String> label = new ArrayList<>();
        label.add("label1");
        label.add("label2");
        labels.put(nodeId, label);
        node.setLabels(labels);
        List<Node> nodes = Collections.singletonList(node);
        when(LibRuntime.nodes()).thenReturn(new Pair<>(new ErrorInfo(), nodes));

        List<Node> getNodes = YR.nodes();
        Node node1 = getNodes.get(0);
        Assert.assertEquals(node1.getId(), nodeId);
        Assert.assertTrue(node1.isAlive());
        Assert.assertEquals(node1.getResources(), resources);
        Assert.assertEquals(node1.getLabels(), labels);

        Node node2 = new Node(nodeId, true, resources, labels);
        Assert.assertEquals(node, node2);

        YR.Finalize();
    }
}
