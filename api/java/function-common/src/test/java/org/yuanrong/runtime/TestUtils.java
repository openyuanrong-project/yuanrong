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

package org.yuanrong.runtime;

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.exception.HandlerNotAvailableException;
import org.yuanrong.runtime.util.Utils;

import org.junit.Assert;
import org.junit.Test;

import java.io.IOException;
import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.HashMap;

public class TestUtils {
    @Test
    public void testSplitEntryWithSeparators() {
        String udfEntry = "{\"userInitEntry\":\"org.yuanrong.executor.UserTestHandler.timeoutInitializer\","
            + "\"userCallEntry\":\"org.yuanrong.executor.UserTestHandler-handler\"}";
        String[] entry = null;
        try {
            entry = Utils.splitEntryWithSeparators(udfEntry, new String[] {"-"});
        } catch (NoSuchMethodException e) {
            Assert.assertEquals(null, entry);
        }
    }

    @Test
    public void testInitUtils() throws NoSuchMethodException {
        Utils utils = new Utils();
        Utils.isEmptyStr("test");
        String generateUUID = Utils.generateUUID();
        String generateTraceId = Utils.generateTraceId("test");
        Utils.invalidDesignatedInstanceID("test");
        Utils.generateRequestId("testJob", "testRuntime", 10);
        Utils.getMethodSignature(Utils.class.getConstructor(null));
        boolean isException = false;
        try {
            Utils.splitMethodName("");
        } catch (HandlerNotAvailableException e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            Utils.splitMethodName("123");
        } catch (HandlerNotAvailableException e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            Utils.splitMethodName("a.b.c");
        } catch (HandlerNotAvailableException e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        HashMap<String, String> map = new HashMap<>();
        map.put("className", "ErrorCode");
        map.put("functionName", "equals");

        ArrayList<Class<?>> classes = new ArrayList<>();
        classes.add(String.class);
        ArrayList<ByteBuffer> byteBuffers = new ArrayList<>();
        byteBuffers.add(ByteBuffer.wrap("test".getBytes(StandardCharsets.UTF_8)));
        isException = false;
        try {
            Utils.deserializeObjects(classes, byteBuffers);
        } catch (IOException e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        Utils.getObjectString(1);
        Utils.getObjectString("1");
        Utils.getObjectMap(1);
        Utils.getObjectMap(new HashMap<String, String>(10));
        Utils.getProcessedExceptionMsg(new Throwable("testMsg"));
        StringBuilder stringBuilder = new StringBuilder(1001);
        for (int i = 0; i < 1111; i++) {
            stringBuilder.append("a");
        }
        String stringWithLen = stringBuilder.toString();
        Utils.getProcessedExceptionMsg(new Throwable(stringWithLen));
        Utils.put("test", null, new ArrayList<>(), null, 1L, 1L);

        isException = false;
        try {
            Utils.get(new ArrayList<>(), 1, 1);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);
        Utils.sleepSeconds(10);


        isException = false;
        try {
            Utils.splitUserClassAndMethod(null,true);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }
}
