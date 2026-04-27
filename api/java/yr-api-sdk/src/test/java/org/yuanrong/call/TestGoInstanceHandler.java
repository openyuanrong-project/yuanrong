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
import org.yuanrong.function.GoInstanceMethod;
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

/**
 * Description: test case of GoInstanceHandler.
 *
 * @since 2024/03/20
 */
@RunWith(PowerMockRunner.class)
@PowerMockIgnore({"javax.net.ssl.*", "javax.management.*"})
@PrepareForTest(YR.class)
public class TestGoInstanceHandler {
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
    public void testGoInstanceHandler() {
        String instanceId = "instanceId";
        String className = "Counter";
        GoInstanceHandler instance = new GoInstanceHandler(instanceId, className);
        GoInstanceHandlerHelper goInstanceHandlerHelper = new GoInstanceHandlerHelper();
        GoInstanceHandler goInstanceHandler = new GoInstanceHandler();
        goInstanceHandler.clearHandlerInfo();
        GoInstanceFunctionHandler<Integer> functionHandler = instance.function(
                GoInstanceMethod.of("Add", int.class, 1));
        GoInstanceMethod<Object> add = GoInstanceMethod.of("Add", 1);

        Assert.assertNotNull(goInstanceHandlerHelper);
        Assert.assertEquals(functionHandler.getClassName(), className);
        Assert.assertEquals(functionHandler.getInstanceId(), instanceId);
        Assert.assertNotNull(functionHandler.getOptions());
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
    public void testGoInstanceHandlerDefault() {
        GoInstanceHandler instance = new GoInstanceHandler();
        GoInstanceFunctionHandler<Integer> functionHandler = instance.function(
                GoInstanceMethod.of("Add", int.class, 1));

        Assert.assertEquals(functionHandler.getClassName(), "");
        Assert.assertNull(functionHandler.getInstanceId());
        Assert.assertNotNull(functionHandler.getOptions());
    }

    @Test
    public void testGoInstanceHandlerPacker() throws IOException {
        String instanceId = "instanceId";
        String className = "Counter";
        GoInstanceHandler instance = new GoInstanceHandler(instanceId, className);
        byte[] bytes = Serializer.serialize(instance);
        if (Serializer.deserialize(bytes, GoInstanceHandler.class) instanceof GoInstanceHandler) {
            GoInstanceHandler unpackHandler = (GoInstanceHandler) Serializer.deserialize(bytes,
                    GoInstanceHandler.class);
            Assert.assertEquals(unpackHandler.getClassName(), className);
            Assert.assertEquals(unpackHandler.getInstanceId(), instanceId);
        }
    }
}
