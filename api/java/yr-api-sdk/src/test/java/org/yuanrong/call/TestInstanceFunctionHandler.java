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

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyList;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.when;

import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.Pair;
import org.yuanrong.exception.YRException;
import org.yuanrong.function.YRFunc0;
import org.yuanrong.function.YRFunc1;
import org.yuanrong.jni.LibRuntime;
import org.yuanrong.libruntime.generated.Libruntime.ApiType;
import org.yuanrong.libruntime.generated.Libruntime.FunctionMeta;

import org.junit.Assert;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.powermock.api.mockito.PowerMockito;
import org.powermock.core.classloader.annotations.PowerMockIgnore;
import org.powermock.core.classloader.annotations.PrepareForTest;
import org.powermock.core.classloader.annotations.SuppressStaticInitializationFor;
import org.powermock.modules.junit4.PowerMockRunner;

@RunWith(PowerMockRunner.class)
@PrepareForTest( {LibRuntime.class})
@SuppressStaticInitializationFor( {"org.yuanrong.jni.LibRuntime"})
@PowerMockIgnore("javax.management.*")
public class TestInstanceFunctionHandler {
    @Before
    public void init() throws Exception {
        PowerMockito.mockStatic(LibRuntime.class);
        when(LibRuntime.IsInitialized()).thenReturn(true);
        when(LibRuntime.Init(any())).thenReturn(new ErrorInfo());
        Pair<ErrorInfo, String> mockRes = new Pair<ErrorInfo, String>(new ErrorInfo(), "objID");
        when(LibRuntime.InvokeInstance(any(), anyString(), anyList(), any())).thenReturn(mockRes);
    }

    @Test
    public void testInitInstanceFunctionHandler() {
        InstanceFunctionHandler<String> functionHandler = new InstanceFunctionHandler<String>(
            (YRFunc0<String>) () -> null, "testID", ApiType.Function);
        functionHandler.getInstanceId();
        functionHandler.getFunc();
        functionHandler.getOptions();
        functionHandler.getApiType();
        Assert.assertNotNull(functionHandler);
    }

    @Test
    public void testInvoke() throws YRException {
        InstanceFunctionHandler<String> functionHandler = new InstanceFunctionHandler<String>(
            (YRFunc1<String, String>) o -> "test", "testID", ApiType.Function);
        functionHandler.invoke("test");
        Assert.assertNotNull(functionHandler);
    }
}