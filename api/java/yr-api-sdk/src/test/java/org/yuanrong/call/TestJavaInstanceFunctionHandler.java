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
import org.yuanrong.InvokeOptions;
import org.yuanrong.api.YR;
import org.yuanrong.exception.YRException;
import org.yuanrong.function.JavaInstanceMethod;
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
 * Java instance function handler test
 *
 * @since 2024/05/20
 */
@RunWith(PowerMockRunner.class)
@PowerMockIgnore({"javax.net.ssl.*", "javax.management.*"})
@PrepareForTest(YR.class)
public class TestJavaInstanceFunctionHandler {
    ClusterModeRuntime runtime = PowerMockito.mock(ClusterModeRuntime.class);

    @Before
    public void initYR() throws Exception {
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

    @Test
    public void testJavaInstanceHandlerInvoke() {
        String instanceId = "instanceId";
        String functionId = "";
        String className = "com.example.Counter";
        JavaInstanceHandler instance = new JavaInstanceHandler(instanceId, functionId, className);
        Assert.assertEquals(className, instance.getClassName());
        Assert.assertEquals(functionId, instance.getFunctionId());
        Assert.assertEquals(instanceId, instance.getInstanceId());
        JavaInstanceFunctionHandler<Integer> functionHandler = instance.function(JavaInstanceMethod.of("Add", int.class));
        JavaInstanceMethod<Object> add = JavaInstanceMethod.of("Add");
        InvokeOptions invokeOptions = new InvokeOptions();
        invokeOptions.setCpu(1500);
        invokeOptions.setMemory(2500);
        functionHandler.options(invokeOptions);
        Assert.assertEquals(className, functionHandler.getClassName());
        Assert.assertEquals(functionId, functionHandler.getFunctionId());
        Assert.assertEquals(instanceId, functionHandler.getInstanceId());
        Assert.assertEquals(1500, functionHandler.getOptions().getCpu());
        Assert.assertEquals(2500, functionHandler.getOptions().getMemory());
    }

    @Test
    public void testInvokeJavaInstanceFunctionHandler() throws IOException, YRException {
        JavaInstanceHandler javaInstanceHandler = new JavaInstanceHandler();
        JavaInstanceFunctionHandler<Integer> functionHandler = javaInstanceHandler.function(JavaInstanceMethod.of("Add", int.class));
        functionHandler.invoke(javaInstanceHandler);
        functionHandler.getJavaInstanceMethod();
        Assert.assertNotNull(functionHandler);
    }

    @After
    public void finalizeYR() throws YRException {
        YR.Finalize();
    }
}
