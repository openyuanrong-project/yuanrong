/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
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

package org.yuanrong.services.runtime.utils;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertNotNull;

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

import java.util.ArrayList;

@RunWith(PowerMockRunner.class)
@PrepareForTest({LibRuntime.class})
@SuppressStaticInitializationFor({"org.yuanrong.jni.LibRuntime"})
@PowerMockIgnore("javax.management.*")
public class TestUtil {
    @Test
    public void testSubstring_fiGreaterThanLength() {
        String str = "Hello, World!";
        int fi = 20;
        int ti = 10;
        String result = Util.substring(str, fi, ti);
        Assert.assertEquals("", result);
    }

    @Test
    public void testSubstring_tiGreaterThanLength() {
        String str = "Hello, World!";
        int fi = 5;
        int ti = 20;
        String result = Util.substring(str, fi, ti);
        Assert.assertEquals(", World!", result);
    }

    @Test
    public void testSubstring_fiAndTiInRange() {
        String str = "Hello, World!";
        int fi = 5;
        int ti = 10;
        String result = Util.substring(str, fi, ti);
        Assert.assertEquals(", Wor", result);
    }

    @Test
    public void testLogOpts() {
        Util.setLogOpts("test");
        assertNotNull(Util.getLogOpts());
        assertEquals("test", Util.getLogOpts()[0]);
        Util.clearLogOpts();
        assertEquals("", Util.getLogOpts()[0]);
    }

    @Test
    public void testSplitMessage() {
        ArrayList<String> lines = Util.splitMessage("test1");
        assertEquals(lines.get(0), "test1");
        String message = "";
        for (int i = 0; i < 90 * 1024 + 100; i++) {
            message = message + "a";
        }
        ArrayList<String> lines1 = Util.splitMessage(message);
        assertEquals(lines1.get(0).length(), 90 * 1024);
        String message1 = "";
        for (int i = 0; i < 90 * 1024; i++) {
            message1 = message1 + "a";
        }
        ArrayList<String> lines2 = Util.splitMessage(message1);
        assertEquals(lines2.get(0).length(), 90 * 1024);

        ArrayList<String> lines3 = Util.splitMessage("");
        assertEquals(lines3.get(0), "");
    }

    @Test
    public void testSendFinishLog() {
        int innerCode = 0;
        String logType = "testLogType";
        String[] callInfos = {"info1", "info2", "info3", "info4", "info5", "info6"};

        PowerMockito.mockStatic(LibRuntime.class);
        PowerMockito.doReturn(new ErrorInfo()).when(LibRuntime.class);
        boolean isException = false;
        try {
            Util.sendFinishLog(innerCode, logType, callInfos);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testSendStartLog() {
        String logType = "testLogType";
        String[] callInfos = {"info1", "info2", "info3", "info4", "info5", "info6"};

        PowerMockito.mockStatic(LibRuntime.class);
        PowerMockito.doReturn(new ErrorInfo()).when(LibRuntime.class);
        boolean isException = false;
        try {
            Util.sendStartLog(logType, callInfos);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testSetLoggers() {
        boolean isException = false;
        try {
            Util.setLoggers();
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }
}
