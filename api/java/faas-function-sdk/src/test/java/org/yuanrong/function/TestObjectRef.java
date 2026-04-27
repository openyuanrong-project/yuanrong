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

import static org.mockito.ArgumentMatchers.anyBoolean;
import static org.mockito.ArgumentMatchers.anyInt;
import static org.mockito.ArgumentMatchers.anyList;
import static org.mockito.Mockito.when;

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.errorcode.Pair;
import org.yuanrong.exception.LibRuntimeException;
import org.yuanrong.function.runtime.exception.InvokeException;
import org.yuanrong.jni.LibRuntime;

import com.google.gson.JsonObject;

import org.junit.Assert;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.powermock.api.mockito.PowerMockito;
import org.powermock.core.classloader.annotations.PowerMockIgnore;
import org.powermock.core.classloader.annotations.PrepareForTest;
import org.powermock.core.classloader.annotations.SuppressStaticInitializationFor;
import org.powermock.modules.junit4.PowerMockRunner;

import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.List;

@RunWith(PowerMockRunner.class)
@PrepareForTest( {LibRuntime.class})
@SuppressStaticInitializationFor( {"org.yuanrong.jni.LibRuntime"})
@PowerMockIgnore("javax.management.*")
public class TestObjectRef {
    @Test
    public void testCheckAndGetTimeoutMs() {
        ObjectRef<String> objectRef = new ObjectRef<String>("testID");
        boolean isException = false;
        try {
            objectRef.get(-2);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }

    @Test
    public void testGet() throws LibRuntimeException {
        ObjectRef<String> objectRef = new ObjectRef<String>("testID");
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.RUNTIME, "");
        ArrayList<byte[]> bytes = new ArrayList<>();
        String jsonString = "{\"innerCode\": \"0\", \"body\": {}}";
        bytes.add(jsonString.getBytes(StandardCharsets.UTF_8));
        Pair<ErrorInfo, List<byte[]>> getRes = new Pair<>(errorInfo, bytes);
        PowerMockito.mockStatic(LibRuntime.class);
        when(LibRuntime.Get(anyList(), anyInt(), anyBoolean())).thenReturn(getRes);
        boolean isException = false;
        try {
            objectRef.get(10);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        isException = false;
        try {
            objectRef.get();
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        ObjectRef<JsonObject> objectRef2 = new ObjectRef<>("1");
        isException = false;
        try {
            objectRef2.get(JsonObject.class);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        isException = false;
        try {
            objectRef2.get(String.class,1);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testGetInvokeException() throws LibRuntimeException {
        ObjectRef<String> objectRef = new ObjectRef<String>("testID");
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.RUNTIME, "");
        ArrayList<byte[]> bytes = new ArrayList<>();
        String jsonString = "{\"innerCode\": \"4004\", \"body\": \"response body size 6400000 exceeds the limit of 6291456\"}";
        bytes.add(jsonString.getBytes(StandardCharsets.UTF_8));
        Pair<ErrorInfo, List<byte[]>> getRes = new Pair<>(errorInfo, bytes);
        PowerMockito.mockStatic(LibRuntime.class);
        when(LibRuntime.Get(anyList(), anyInt(), anyBoolean())).thenReturn(getRes);
        boolean isException = false;
        try {
            objectRef.get(10);
        } catch (InvokeException e) {
            isException = true;
            Assert.assertEquals(4004, e.getErrorCode());
            Assert.assertTrue(e.getMessage().contains("exceeds the limit of 6291456"));
        }
        Assert.assertTrue(isException);

        ObjectRef<JsonObject> objectRef2 = new ObjectRef<>("1");
        isException = false;
        try {
            objectRef2.get(JsonObject.class);
        } catch (InvokeException e) {
            isException = true;
            Assert.assertEquals(4004, e.getErrorCode());
            Assert.assertTrue(e.getMessage().contains("exceeds the limit of 6291456"));
        }
        Assert.assertTrue(isException);
    }
}
