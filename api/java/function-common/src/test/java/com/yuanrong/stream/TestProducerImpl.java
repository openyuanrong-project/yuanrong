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

package com.yuanrong.stream;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyInt;
import static org.mockito.ArgumentMatchers.anyLong;
import static org.mockito.Mockito.when;

import com.yuanrong.errorcode.ErrorCode;
import com.yuanrong.errorcode.ErrorInfo;
import com.yuanrong.errorcode.ModuleCode;
import com.yuanrong.jni.JniProducer;

import org.junit.Assert;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.powermock.api.mockito.PowerMockito;
import org.powermock.core.classloader.annotations.PowerMockIgnore;
import org.powermock.core.classloader.annotations.PrepareForTest;
import org.powermock.core.classloader.annotations.SuppressStaticInitializationFor;
import org.powermock.modules.junit4.PowerMockRunner;

import java.nio.ByteBuffer;

@RunWith(PowerMockRunner.class)
@PrepareForTest( {JniProducer.class})
@SuppressStaticInitializationFor( {"com.yuanrong.jni.JniProducer"})
@PowerMockIgnore("javax.management.*")
public class TestProducerImpl {
    @Test
    public void testInitProducerImpl() {
        ProducerImpl producer = new ProducerImpl(10L);
        producer.toString();
        producer.hashCode();
        ProducerImpl producer1 = new ProducerImpl(20L);
        Assert.assertTrue(producer1.equals(producer1));
        Assert.assertFalse(producer1.equals(null));
        Assert.assertFalse(producer.equals(producer1));
    }

    @Test
    public void testSend() {
        ProducerImpl producer = new ProducerImpl(10L);
        ByteBuffer byteBuffer = ByteBuffer.allocateDirect(100);
        byteBuffer.put(new byte[] {1, 2, 3, 4, 5});
        byteBuffer.flip();

        PowerMockito.mockStatic(JniProducer.class);
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.CORE, "");
        when(JniProducer.sendDirectBufferDefaultTimeout(anyLong(), any())).thenReturn(errorInfo);
        when(JniProducer.sendDirectBuffer(anyLong(), any(), anyInt())).thenReturn(errorInfo);

        boolean isException = false;
        try {
            producer.send(new Element(10L, byteBuffer));
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        isException = false;
        try {
            producer.send(new Element(10L, byteBuffer), 10);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        ByteBuffer byteBuffer1 = ByteBuffer.allocate(100);
        when(JniProducer.sendHeapBufferDefaultTimeout(anyLong(), any(), anyLong())).thenReturn(errorInfo);
        when(JniProducer.sendHeapBuffer(anyLong(), any(), anyLong(), anyInt())).thenReturn(errorInfo);

        isException = false;
        try {
            producer.send(new Element(10L, byteBuffer1));
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        isException = false;
        try {
            producer.send(new Element(10L, byteBuffer1), 10);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testClose() {
        ProducerImpl producer = new ProducerImpl(10L);
        ProducerImpl producer1 = new ProducerImpl(0L);
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.CORE, "");
        PowerMockito.mockStatic(JniProducer.class);
        when(JniProducer.close(anyLong())).thenReturn(errorInfo);
        boolean isException = false;

        try {
            producer1.close();
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            producer.close();
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testFinalize() throws Exception {
        ProducerImpl producer = new ProducerImpl(10L);
        PowerMockito.mockStatic(JniProducer.class);
        PowerMockito.doNothing().when(JniProducer.class, "freeJNIPtrNative",anyLong());
        producer.finalize();
    }
}
