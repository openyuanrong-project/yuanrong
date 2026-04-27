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

import static org.mockito.ArgumentMatchers.anyList;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;

import org.yuanrong.Config;
import org.yuanrong.InvokeOptions;
import org.yuanrong.api.YR;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.Pair;
import org.yuanrong.exception.YRException;
import org.yuanrong.jni.LibRuntime;
import org.yuanrong.runtime.client.ObjectRef;

import org.junit.After;
import org.junit.Assert;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.powermock.api.mockito.PowerMockito;
import org.powermock.core.classloader.annotations.PowerMockIgnore;
import org.powermock.core.classloader.annotations.PrepareForTest;
import org.powermock.core.classloader.annotations.SuppressStaticInitializationFor;
import org.powermock.modules.junit4.PowerMockRunner;

import java.io.IOException;

@RunWith(PowerMockRunner.class)
@PrepareForTest({LibRuntime.class, YR.class})
@SuppressStaticInitializationFor({"org.yuanrong.jni.LibRuntime"})
@PowerMockIgnore({"javax.net.ssl.*", "javax.management.*"})
public class TestFunctionHandler {
    public static class MyYRApp {
        public static String call() {
            return "aaa";
        }
    }

    @Before
    public void initYR() throws Exception {
        Config conf = new Config(
            "sn:cn:yrk:default:function:0-crossyrlib-helloworld:$latest",
            "127.0.0.0",
            "127.0.0.0",
            "",
            "sn:cn:yrk:default:function:0-test-hello:$latest",
            true);
        PowerMockito.mockStatic(LibRuntime.class);
        when(LibRuntime.IsInitialized()).thenReturn(true);
        when(LibRuntime.Init(any())).thenReturn(new ErrorInfo());
        YR.init(conf);
    }

    @After
    public void finalizeYR() throws YRException {
        YR.Finalize();
    }

    @Test
    public void testFunctionHandler() throws Exception {
        Pair<ErrorInfo, String> mockResult = new Pair<ErrorInfo, String>(new ErrorInfo(), "objID");
        when(LibRuntime.InvokeInstance(any(), anyString(), anyList(), any())).thenReturn(mockResult);
        InvokeOptions options= new InvokeOptions();
        FunctionHandler handler = YR.function(MyYRApp::call);
        ObjectRef ref = handler.options(options).invoke();
        Assert.assertTrue(ref.getObjId().equals("objID"));
        Assert.assertNotNull(handler.getFunc());
        Assert.assertNotNull(handler.getOptions());
    }
}
