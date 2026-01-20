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

package org.yuanrong;

import static org.junit.Assert.assertFalse;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.jni.LibRuntime;

import org.junit.Assert;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.powermock.api.mockito.PowerMockito;
import org.powermock.core.classloader.annotations.PowerMockIgnore;
import org.powermock.core.classloader.annotations.PrepareForTest;
import org.powermock.core.classloader.annotations.SuppressStaticInitializationFor;
import org.powermock.modules.junit4.PowerMockRunner;

/**
 * The type Http task test.
 *
 * @since 2024 /04/18
 */
@RunWith(PowerMockRunner.class)
@PrepareForTest({LibRuntime.class})
@SuppressStaticInitializationFor({"org.yuanrong.jni.LibRuntime"})
@PowerMockIgnore("javax.management.*")
public class TestEntrypoint {

    /**
     * Description：
     *   Confirm that the code and message from 'Create' method are correct.
     * Steps:
     *   1. Sets 'Libruntime.Init' method return an ErrorInfo with not 'OK' ErrorCode.
     *   2. Sets 'Libruntime.ReceiveRequestLoop' throw an Exception.
     * Expectation:
     *   Exception should not be thrown.
     * @throws Exception
     */
    @Test
    public void testRuntimeEntrypoint() throws Exception {
        PowerMockito.mockStatic(LibRuntime.class);
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_CONNECTION_FAILED, ModuleCode.CORE, "");
        when(LibRuntime.Init(any())).thenReturn(errorInfo);
        System.setProperty("POSIX_LISTEN_ADDR", "127.0.0.1:1");
        System.setProperty("DATASYSTEM_ADDR", "127.0.0.1:1");
        boolean isException = false;
        try {
            Entrypoint.runtimeEntrypoint(new String[] {"127.0.0.1:2", "arg2"});
        } catch (RuntimeException e) {
            isException = true;
        }
        assertFalse("RuntimeException should not be thrown", isException);
    }

    @Test
    public void testSplitIPAndPortFromAddrException() {
        boolean isException = false;
        try {
            Entrypoint.runtimeEntrypoint(new String[] {"error", "arg2"});
        } catch (RuntimeException e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }

    @Test
    public void testSafeGetEnvWithException() {
        System.setProperty("POSIX_LISTEN_ADDR", "");
        boolean isException = false;
        try {
            Entrypoint.runtimeEntrypoint(new String[] {"127.0.0.1:2", "arg2"});
        } catch (RuntimeException e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }

    @Test
    public void testRuntimeEntrypointErrorOK() throws Exception {
        PowerMockito.mockStatic(LibRuntime.class);
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.CORE, "");
        when(LibRuntime.Init(any())).thenReturn(errorInfo);
        PowerMockito.doNothing().when(LibRuntime.class, "ReceiveRequestLoop");
        PowerMockito.doNothing().when(LibRuntime.class, "Finalize");
        System.setProperty("POSIX_LISTEN_ADDR", "127.0.0.1:1");
        System.setProperty("DATASYSTEM_ADDR", "127.0.0.1:1");
        boolean isException = false;
        try {
            Entrypoint entrypoint = new Entrypoint();
            Entrypoint.runtimeEntrypoint(new String[] {"127.0.0.1:2", "arg2"});
        } catch (RuntimeException e) {
            isException = true;
        }
        assertFalse("RuntimeException should not be thrown", isException);
    }
}
