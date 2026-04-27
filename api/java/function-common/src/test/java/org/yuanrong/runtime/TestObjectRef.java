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

import static org.mockito.ArgumentMatchers.anyBoolean;
import static org.mockito.ArgumentMatchers.anyInt;
import static org.mockito.ArgumentMatchers.anyList;
import static org.mockito.Mockito.when;

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.errorcode.Pair;
import org.yuanrong.exception.LibRuntimeException;
import org.yuanrong.jni.LibRuntime;
import org.yuanrong.runtime.client.ObjectRef;
import org.yuanrong.storage.InternalWaitResult;

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
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RunWith(PowerMockRunner.class)
@PrepareForTest( {LibRuntime.class})
@SuppressStaticInitializationFor( {"org.yuanrong.jni.LibRuntime"})
@PowerMockIgnore("javax.management.*")
public class TestObjectRef {
    @Test
    public void testGetObjectRef() throws LibRuntimeException {
        ObjectRef objectRef = new ObjectRef("testID");
        PowerMockito.mockStatic(LibRuntime.class);
        List<String> readyIds = Arrays.asList("2", "1");
        List<String> unreadyIds = new ArrayList<String>();
        Map<String, ErrorInfo> exceptionIds = new HashMap<String, ErrorInfo>();
        InternalWaitResult waitResult = new InternalWaitResult(readyIds, unreadyIds, exceptionIds);
        when(LibRuntime.Wait(anyList(), anyInt(), anyInt())).thenReturn(waitResult);

        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.RUNTIME, "");
        List<byte[]> ok = new ArrayList<>();
        ok.add("result1".getBytes(StandardCharsets.UTF_8));
        ok.add("result2".getBytes(StandardCharsets.UTF_8));
        Pair<ErrorInfo, List<byte[]>> getRes = new Pair<>(errorInfo, ok);
        when(LibRuntime.Get(anyList(), anyInt(), anyBoolean())).thenReturn(getRes);

        boolean isException = false;
        try {
            objectRef.get();
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        isException = false;
        try {
            objectRef.get(1);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        isException = false;
        try {
            objectRef.get(ObjectRef.class);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testGetNullObjectRef() throws LibRuntimeException {
        ObjectRef objectRef = new ObjectRef("testID");
        PowerMockito.mockStatic(LibRuntime.class);
        List<String> readyIds = Arrays.asList("2", "1");
        List<String> unreadyIds = new ArrayList<String>();
        Map<String, ErrorInfo> exceptionIds = new HashMap<String, ErrorInfo>();
        InternalWaitResult waitResult = new InternalWaitResult(readyIds, unreadyIds, exceptionIds);
        when(LibRuntime.Wait(anyList(), anyInt(), anyInt())).thenReturn(waitResult);

        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.RUNTIME, "");
        List<byte[]> ok = new ArrayList<>();
        Pair<ErrorInfo, List<byte[]>> getRes = new Pair<>(errorInfo, ok);
        when(LibRuntime.Get(anyList(), anyInt(), anyBoolean())).thenReturn(getRes);
        boolean isException = false;
        try {
            objectRef.get();
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            objectRef.get(1);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            objectRef.get(ObjectRef.class);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }
}
