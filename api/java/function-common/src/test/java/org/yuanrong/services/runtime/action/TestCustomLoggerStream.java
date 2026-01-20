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

package org.yuanrong.services.runtime.action;

import org.yuanrong.runtime.util.Constants;

import org.junit.Assert;
import org.junit.Test;

import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;

public class TestCustomLoggerStream {
    @Test
    public void testWrite() {
        CustomLoggerStream customLoggerStream = new CustomLoggerStream("LOG");
        byte[] bt = {1, 2};
        try {
            customLoggerStream.write(bt, 5, 3);
        } catch (Exception e) {
            Assert.assertEquals(IndexOutOfBoundsException.class, e.getClass());
        }
        customLoggerStream.write(bt, 1, 1);
    }

    /**
     * test cutRequestIDAndInvokeID is success
     */
    @Test
    public void testCutRequestIDAndInvokeID() {
        CustomLoggerStream customLoggerStream = new CustomLoggerStream("LOG");
        Class[] cArg = new Class[1];
        cArg[0] = String.class;
        String result = "";
        String result2 = "";
        try {
            Method method = customLoggerStream.getClass().getDeclaredMethod("cutRequestIDAndInvokeID", cArg);
            method.setAccessible(true);
            result = (String) method.invoke(customLoggerStream,
                "[2025-02-06 16:46:54.307 INFO  YRLogRequestID:f93aa503-f151-4374-8622-67c7ee3a835c|123 main.java:26] 测试日志");
            result2 = (String) method.invoke(customLoggerStream,
                "[2025-02-06 16:46:54.307 ERROR YRLogRequestID:f93aa503-f151-4374-8622-67c7ee3a835c|123 main.java:26] 测试日志");
        } catch (NoSuchMethodException | IllegalAccessException | InvocationTargetException e) {
            Assert.fail();
        }
        Assert.assertEquals("[2025-02-06 16:46:54.307 INFO  main.java:26] 测试日志", result);
        Assert.assertEquals("[2025-02-06 16:46:54.307 ERROR main.java:26] 测试日志", result2);

        try {
            Method method = customLoggerStream.getClass().getDeclaredMethod("cutRequestIDAndInvokeID", cArg);
            method.setAccessible(true);
            result = (String) method.invoke(customLoggerStream,
                "[2025-02-06 16:46:54.307 INFO  main.java:26] 测试日志");
            result2 = (String) method.invoke(customLoggerStream,
                "[2025-02-06 16:46:54.307 ERROR main.java:26] 测试日志");
        } catch (NoSuchMethodException | IllegalAccessException | InvocationTargetException e) {
            Assert.fail();
        }
        Assert.assertEquals("[2025-02-06 16:46:54.307 INFO  main.java:26] 测试日志", result);
        Assert.assertEquals("[2025-02-06 16:46:54.307 ERROR main.java:26] 测试日志", result2);
    }

    /**
     * test getRequestID is success
     */
    @Test
    public void testHandleRequestIDAndInvokeID() {
        CustomLoggerStream customLoggerStream = new CustomLoggerStream("LOG");
        Class[] cArg = new Class[1];
        cArg[0] = String.class;
        String[] result = null;
        String[] result2 = null;
        try {
            Method method = customLoggerStream.getClass().getDeclaredMethod("handleRequestIDAndInvokeID", cArg);
            method.setAccessible(true);
            result = (String[]) method.invoke(customLoggerStream,
                "[2022-11-19 16:46:54.307 INFO  YRLogRequestID:f93aa503-f151-4374-8622-67c7ee3a835c|123 main.java:26] 测试日志");
            result2 = (String[]) method.invoke(customLoggerStream,
                "[2022-11-19 16:46:54.307 ERROR YRLogRequestID:f93aa503-f151-4374-8622-67c7ee3a835c|123 main.java:26] 测试日志");
        } catch (NoSuchMethodException | IllegalAccessException | InvocationTargetException e) {
            Assert.fail();
        }
        Assert.assertEquals("f93aa503-f151-4374-8622-67c7ee3a835c", result[0]);
        Assert.assertEquals("123", result[1]);
        Assert.assertEquals("f93aa503-f151-4374-8622-67c7ee3a835c", result2[0]);
        Assert.assertEquals("123", result2[1]);

        try {
            Method method = customLoggerStream.getClass().getDeclaredMethod("handleRequestIDAndInvokeID", cArg);
            method.setAccessible(true);
            result = (String[]) method.invoke(customLoggerStream,
                "[2022-11-19 16:46:54.307 INFO  main.java:26] 测试日志");
            result2 = (String[]) method.invoke(customLoggerStream,
                "[2022-11-19 16:46:54.307 ERROR main.java:26] 测试日志");
        } catch (NoSuchMethodException | IllegalAccessException | InvocationTargetException e) {
            Assert.fail();
        }
        Assert.assertEquals(0, result.length);
        Assert.assertEquals(0, result2.length);
    }

    @Test
    public void testHandleLargeLog() {
        try {
            CustomLoggerStream customLoggerStream = new CustomLoggerStream("INFO");
            Method handleLargeLog = customLoggerStream.getClass().getDeclaredMethod("handleLargeLog", String.class);
            handleLargeLog.setAccessible(true);
            handleLargeLog.invoke(customLoggerStream, Constants.LOG_REQUEST_ID + ":aaaa|bbbb");
        } catch (Exception e) {
            Assert.assertEquals(InvocationTargetException.class.getName(), e.toString());
        }
    }
}