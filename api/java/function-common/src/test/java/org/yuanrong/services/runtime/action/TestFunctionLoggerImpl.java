/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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

package org.yuanrong.services.runtime.action;

import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.jni.LibRuntime;

import org.junit.Assert;
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
public class TestFunctionLoggerImpl {
    @Test
    public void testFunctionLogger() throws Exception {
        FunctionLoggerImpl functionLogger = new FunctionLoggerImpl();
        functionLogger.setInvokeID("1");
        functionLogger.setRequestID("2");
        Assert.assertNotNull(functionLogger);
    }

    @Test
    public void testCustomLogger() throws Exception {
        CustomLoggerStream customLogger = new CustomLoggerStream("INFO");
        Assert.assertNotNull(customLogger);
    }

    /**
     * test set FunctionLoggerImpl level
     */
    @Test
    public void testSetLevel() {
        FunctionLoggerImpl functionLogger = new FunctionLoggerImpl();
        functionLogger.setLevel("INFO");
        Assert.assertNotNull(functionLogger);
    }

    @Test
    public void testLog() throws Exception {
        PowerMockito.mockStatic(LibRuntime.class);
        PowerMockito.doReturn(new ErrorInfo()).when(LibRuntime.class);

        FunctionLoggerImpl functionLogger = new FunctionLoggerImpl();
        functionLogger.setLevel("WARN");
        functionLogger.log("");
        Assert.assertNotNull(functionLogger);

        FunctionLoggerImpl functionLogger1 = new FunctionLoggerImpl();
        functionLogger1.setInvokeID("001");
        functionLogger1.setRequestID("002");
        functionLogger1.log("ss");
        functionLogger1.debug("");
        Assert.assertNotNull(functionLogger1);
    }

    @Test
    public void testDebug() throws Exception {
        PowerMockito.mockStatic(LibRuntime.class);
        PowerMockito.doReturn(new ErrorInfo()).when(LibRuntime.class);

        FunctionLoggerImpl functionLogger = new FunctionLoggerImpl();
        functionLogger.setLevel("DEBUG");
        functionLogger.debug("");
        functionLogger.debug("ss");
        Assert.assertNotNull(functionLogger);

        FunctionLoggerImpl functionLogger1 = new FunctionLoggerImpl();
        functionLogger1.setInvokeID("001");
        functionLogger1.setRequestID("002");
        functionLogger1.debug("ss");
        Assert.assertNotNull(functionLogger1);
    }

    @Test
    public void testWarn() throws Exception {
        PowerMockito.mockStatic(LibRuntime.class);
        PowerMockito.doReturn(new ErrorInfo()).when(LibRuntime.class);

        FunctionLoggerImpl functionLogger = new FunctionLoggerImpl();
        functionLogger.setLevel("ERROR");
        functionLogger.warn("");
        Assert.assertNotNull(functionLogger);

        FunctionLoggerImpl functionLogger1 = new FunctionLoggerImpl();
        functionLogger1.setInvokeID("001");
        functionLogger1.setRequestID("002");
        functionLogger1.warn("ss");
        functionLogger1.warn("");
        Assert.assertNotNull(functionLogger1);
    }

    @Test
    public void testError() throws Exception {
        PowerMockito.mockStatic(LibRuntime.class);
        PowerMockito.doReturn(new ErrorInfo()).when(LibRuntime.class);

        FunctionLoggerImpl functionLogger = new FunctionLoggerImpl();
        functionLogger.setLevel("ERROR");
        functionLogger.error("");
        functionLogger.error("ss");
        Assert.assertNotNull(functionLogger);
    }
}
