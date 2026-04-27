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

package org.yuanrong.function;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyList;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.when;

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.errorcode.Pair;
import org.yuanrong.exception.LibRuntimeException;
import org.yuanrong.function.common.ContextMock;
import org.yuanrong.function.runtime.exception.InvokeException;
import org.yuanrong.jni.LibRuntime;

import org.junit.Assert;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.powermock.api.mockito.PowerMockito;
import org.powermock.core.classloader.annotations.PowerMockIgnore;
import org.powermock.core.classloader.annotations.PrepareForTest;
import org.powermock.core.classloader.annotations.SuppressStaticInitializationFor;
import org.powermock.modules.junit4.PowerMockRunner;

import java.lang.reflect.Field;

@RunWith(PowerMockRunner.class)
@PrepareForTest( {LibRuntime.class})
@SuppressStaticInitializationFor( {"org.yuanrong.jni.LibRuntime"})
@PowerMockIgnore("javax.management.*")
public class TestFunction {
    private ContextMock context = null;

    @Before
    public void setup() {
        context = new ContextMock();
        context.setInvokeProperty("{'1':'aa'}");
    }

    @Test
    public void testFunctionSecond() {
        new Thread(() -> {
            try {
                Function function = new Function(context, "ss");
                Assert.assertNotNull(function);
            } catch (InvokeException e) {
                Assert.fail();
            }
        }).start();
        try {
            Thread.sleep(1000);
        } catch (InterruptedException e) {
            Assert.fail();
        }
    }

    @Test
    public void testFunctionThird() {
        Thread p =new Thread(() -> {
            try {
                Function function = new Function(context, "ss:alias");
                Assert.assertNotNull(function);
            } catch (InvokeException e) {
                Assert.fail();
            }
        });
        p.start();
        try {
            p.join(1000);
        } catch (InterruptedException e) {
            Assert.fail();
        }
    }

    @Test
    public void testSaveState() {
        new Thread(() -> {
            Function function = new Function(context);
            try {
                function.saveState();
            } catch (InvokeException e) {
                Assert.fail();
            }
        }).start();
        try {
            Thread.sleep(1000);
        } catch (InterruptedException e) {
            Assert.fail();
        }
    }

    @Test
    public void testTerminate() {
        new Thread(() -> {
            try {
                Function function = new Function(context, "demo:latest");
                Class cl = function.getClass();
                Field field = cl.getDeclaredField("instanceID");
                field.setAccessible(true);
                field.set(function, "30");
                ObjectRef terminate = function.terminate();
                Assert.assertNull(terminate);
            } catch (InvokeException | NoSuchFieldException | IllegalAccessException ignored) {
                Assert.assertNull(ignored);
            }
        }).start();
        try {
            Thread.sleep(1000);
        } catch (InterruptedException e) {
            Assert.fail();
        }
        try {
            Thread.sleep(500);
        } catch (InterruptedException e) {
            Assert.fail();
        }
    }

    @Test
    public void testFunctionWithInvalidResource() {
        CreateOptions createOptions = new CreateOptions(-1);
        boolean isException = false;
        try {
            Function f = new Function(context).options(createOptions);
        } catch (InvokeException e) {
            isException = true;
            Assert.assertTrue(e.getMessage().contains("Invalid dynamic resource options, not allow negative number"));
        }
        Assert.assertTrue(isException);
    }

    @Test
    public void testInitFunction() throws Exception {
        Function testFunc = new Function("testFunc");
        testFunc.options(new CreateOptions(10, 10));
        testFunc.getContext();
        testFunc.getInstanceID();
        testFunc.terminate();

        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.RUNTIME, "");
        Pair<ErrorInfo, String> getRes = new Pair<>(errorInfo, "ok");
        PowerMockito.mockStatic(LibRuntime.class);
        when(LibRuntime.InvokeInstance(any(),anyString(), anyList(), any())).thenReturn(getRes);
        testFunc.invoke("{\n" + "  \"name\": \"test\"\n" + "}");
    }
}
