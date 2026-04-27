/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
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

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;

import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.exception.YRException;
import org.yuanrong.exception.LibRuntimeException;
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

@RunWith(PowerMockRunner.class)
@PrepareForTest({LibRuntime.class})
@SuppressStaticInitializationFor( {"org.yuanrong.jni.LibRuntime"})
@PowerMockIgnore("javax.management.*")
public class TestStream {
    @Test
    public void testWriteStream() throws Exception {
        PowerMockito.mockStatic(LibRuntime.class);
        when(LibRuntime.streamWrite(any(), any(), any())).thenReturn(new ErrorInfo());

        Stream stream = new Stream();
        boolean isException = false;
        try {
            stream.write(new JsonObject());
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
        stream.setMaxWriteSize(0);
        stream.setRequestId("request_id");
        stream.setInstanceId("instance_id");
        JsonObject jsonObject = new JsonObject();
        jsonObject.addProperty("testStream","aaaaaaaaaaaaaaaaaaaaaa");
        try {
            stream.write(jsonObject);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        when(LibRuntime.streamWrite(any(), any(), any())).thenThrow(new LibRuntimeException("test_exception"));
        try {
            stream.write(jsonObject);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }
}

