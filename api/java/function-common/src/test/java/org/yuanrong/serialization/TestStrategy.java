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

package org.yuanrong.serialization;

import org.yuanrong.runtime.client.ObjectRef;
import org.yuanrong.serialization.strategy.Strategy;

import org.junit.Assert;
import org.junit.Test;

import java.nio.ByteBuffer;
import java.util.ArrayList;

public class TestStrategy {
    @Test
    public void testInitStrategy() {
        Strategy strategy = new Strategy();
        ObjectRef objectRef = new ObjectRef("test");
        objectRef.setByteBuffer(true);
        Strategy.selectDeserializationStrategy(objectRef);

        ByteBuffer byteBuffer = ByteBuffer.allocate(10);
        Strategy.selectSerializationStrategy(byteBuffer);

        ArrayList<ObjectRef> objectRefs = new ArrayList<>();
        boolean isException = false;
        try {
            Strategy.getObjects(null, objectRefs);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }
}
