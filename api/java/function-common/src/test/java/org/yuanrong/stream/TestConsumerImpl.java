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

package org.yuanrong.stream;

import static org.mockito.ArgumentMatchers.anyBoolean;
import static org.mockito.ArgumentMatchers.anyInt;
import static org.mockito.ArgumentMatchers.anyLong;
import static org.mockito.Mockito.when;

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.errorcode.Pair;
import org.yuanrong.jni.JniConsumer;

import org.junit.Assert;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.powermock.api.mockito.PowerMockito;
import org.powermock.core.classloader.annotations.PowerMockIgnore;
import org.powermock.core.classloader.annotations.PrepareForTest;
import org.powermock.core.classloader.annotations.SuppressStaticInitializationFor;
import org.powermock.modules.junit4.PowerMockRunner;

import java.util.ArrayList;
import java.util.List;

@RunWith(PowerMockRunner.class)
@PrepareForTest( {JniConsumer.class})
@SuppressStaticInitializationFor( {"org.yuanrong.jni.JniConsumer"})
@PowerMockIgnore("javax.management.*")
public class TestConsumerImpl {
    @Test
    public void testInitConsumerImpl() {
        ConsumerImpl consumer = new ConsumerImpl(10L);
        ConsumerImpl consumer1 = new ConsumerImpl(20L);
        consumer.toString();
        consumer1.hashCode();
        Assert.assertTrue(consumer.equals(consumer));
        Assert.assertFalse(consumer1.equals(null));
        Assert.assertFalse(consumer.equals(consumer1));
    }

    @Test
    public void testReceive() {
        ConsumerImpl consumer = new ConsumerImpl(10L);
        PowerMockito.mockStatic(JniConsumer.class);
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.CORE, "");
        Pair<ErrorInfo, List<Element>> errorInfoListPair = new Pair<>(errorInfo, new ArrayList<>());
        when(JniConsumer.receive(anyLong(), anyLong(), anyInt(), anyBoolean())).thenReturn(errorInfoListPair);

        boolean isException = false;
        try {
            consumer.receive(10);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        isException = false;
        try {
            consumer.receive(10L, 10);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        ConsumerImpl consumer1 = new ConsumerImpl(0L);
        isException = false;
        try {
            consumer1.receive(10L, 10);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }

    @Test
    public void testAck() {
        ConsumerImpl consumer = new ConsumerImpl(10L);
        PowerMockito.mockStatic(JniConsumer.class);
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.CORE, "");
        when(JniConsumer.ack(anyLong(), anyLong())).thenReturn(errorInfo);
        boolean isException = false;
        try {
            consumer.ack(10);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testClose() {
        ConsumerImpl consumer = new ConsumerImpl(10L);
        PowerMockito.mockStatic(JniConsumer.class);
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.CORE, "");
        when(JniConsumer.close(anyLong())).thenReturn(errorInfo);
        boolean isException = false;
        try {
            consumer.close();
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        ConsumerImpl consumer1 = new ConsumerImpl(0L);
        isException = false;
        try {
            consumer1.close();
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }

    @Test
    public void testFinalize() {
        ConsumerImpl consumer = new ConsumerImpl(10L);
        PowerMockito.mockStatic(JniConsumer.class);
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.CORE, "");
        when(JniConsumer.freeJNIPtrNative(anyLong())).thenReturn(errorInfo);
        boolean isException = false;
        try {
            consumer.finalize();
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }
}
