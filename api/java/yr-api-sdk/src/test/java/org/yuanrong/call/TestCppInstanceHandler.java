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

package org.yuanrong.call;

import org.yuanrong.Config;
import org.yuanrong.api.YR;
import org.yuanrong.exception.YRException;
import org.yuanrong.function.CppInstanceMethod;
import org.yuanrong.serialization.Serializer;
import org.yuanrong.runtime.ClusterModeRuntime;

import org.junit.After;
import org.junit.Assert;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.powermock.api.mockito.PowerMockito;
import org.powermock.core.classloader.annotations.PowerMockIgnore;
import org.powermock.core.classloader.annotations.PrepareForTest;
import org.powermock.modules.junit4.PowerMockRunner;

import java.io.IOException;
import java.util.Map;

/**
 * Description: test case of CppInstanceHandler.
*
* @since 2024/05/20
*/
@RunWith(PowerMockRunner.class)
@PowerMockIgnore({"Cppx.net.ssl.*", "javax.management.*"})
@PrepareForTest(YR.class)
public class TestCppInstanceHandler {
    ClusterModeRuntime runtime = PowerMockito.mock(ClusterModeRuntime.class);

    @Before
    public void init() throws Exception {
        Config conf = new Config(
                "sn:cn:yrk:default:function:0-crossyrlib-helloworld:$latest",
                "127.0.0.0",
                "127.0.0.0",
                "",
                "sn:cn:yrk:default:function:0-test-hello:$latest",
                true);
        PowerMockito.whenNew(ClusterModeRuntime.class).withAnyArguments().thenReturn(runtime);
        YR.init(conf);
    }

    @After
    public void finalizeYR() throws YRException {
        YR.Finalize();
    }

    @Test
    public void testCppInstanceHandler() throws YRException {
        String instanceId = "instanceId";
        String functionId = "functionId";
        String className = "Counter";
        CppInstanceHandler instance = new CppInstanceHandler(instanceId, functionId, className);
        CppInstanceHandler cppInstanceHandler = new CppInstanceHandler();
        cppInstanceHandler.release();
        cppInstanceHandler.clearHandlerInfo();
        CppInstanceHandlerHelper cppInstanceHandlerHelper = new CppInstanceHandlerHelper();
        CppInstanceFunctionHandler<Integer> functionHandler = instance.function(
                CppInstanceMethod.of("Add", int.class));
        CppInstanceMethod<Object> add = CppInstanceMethod.of("Add");

        Assert.assertEquals(functionHandler.getClassName(), className);
        Assert.assertEquals(functionHandler.getFunctionId(), functionId);
        Assert.assertEquals(functionHandler.getInstanceId(), instanceId);
        Assert.assertNotNull(functionHandler.getOptions());
        Assert.assertNotNull(cppInstanceHandlerHelper);
        boolean isException = false;
        try {
            instance.terminate();
            instance.terminate(true);
            instance.terminate(false);
        } catch (YRException exp) {
            exp.printStackTrace();
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testCppInstanceHandlerDefault() {
        CppInstanceHandler instance = new CppInstanceHandler();
        CppInstanceFunctionHandler<Integer> functionHandler = instance.function(
                CppInstanceMethod.of("Add", int.class));

        Assert.assertEquals(functionHandler.getClassName(), "");
        Assert.assertEquals(functionHandler.getFunctionId(), "");
        Assert.assertNull(functionHandler.getInstanceId());
        Assert.assertNotNull(functionHandler.getOptions());
    }

    @Test
    public void testCppInstanceHandlerPacker() throws IOException {
        String instanceId = "instanceId";
        String functionId = "functionId";
        String className = "Counter";
        CppInstanceHandler instance = new CppInstanceHandler(instanceId, functionId, className);
        byte[] bytes = Serializer.serialize(instance);
        boolean isException = false;
        Object obj = new Object();
        try {
            obj = Serializer.deserialize(bytes, CppInstanceHandler.class);
        } catch (Exception e) {
            e.printStackTrace();
            isException = true;
        }
        Assert.assertFalse(isException);
        if (obj instanceof CppInstanceHandler) {
            CppInstanceHandler unpackHandler = (CppInstanceHandler) Serializer.deserialize(bytes,
                    CppInstanceHandler.class);
            Assert.assertEquals(unpackHandler.getClassName(), className);
            Assert.assertEquals(unpackHandler.getInstanceId(), instanceId);
        }
    }

    @Test
    public void testExportImportCppInstanceHandler() {
        String instanceId = "instanceId";
        String functionId = "functionId";
        String className = "Counter";
        boolean isException = false;
        CppInstanceHandler instanceHandler = new CppInstanceHandler(instanceId, functionId, className);
        instanceHandler.setNeedOrder(true);
        try {
            Map<String, String> mp = instanceHandler.exportHandler();
            Assert.assertEquals(mp.get("instanceKey"), instanceId);
            Assert.assertEquals(mp.get("functionId"), functionId);
            Assert.assertEquals(mp.get("className"), className);
            Assert.assertEquals(mp.get("needOrder"), "true");
            mp.put("instanceID", "realInsID");
            mp.put("instanceRoute", "route");
            CppInstanceHandler newHandler = new CppInstanceHandler();
            newHandler.importHandler(mp);
            newHandler.exportHandler();
            Assert.assertEquals(newHandler.getClassName(), className);
            Assert.assertEquals(newHandler.getFunctionId(), functionId);
            Assert.assertEquals(newHandler.getInstanceId(), instanceId);
            Assert.assertEquals(newHandler.getRealInstanceId(), "realInsID");
            Assert.assertEquals(newHandler.isNeedOrder(), true);
            newHandler.release();
        } catch (YRException exp) {
            exp.printStackTrace();
            isException = true;
        }
        Assert.assertFalse(isException);
    }
}
