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

package org.yuanrong.stream;

import org.junit.Assert;
import org.junit.Test;

import java.nio.ByteBuffer;

public class TestElement {
    @Test
    public void testInitElement() {
        Element element = new Element();
        Element element1 = new Element(1L, ByteBuffer.allocate(10));
        element.setId(2L);
        element1.setBuffer(ByteBuffer.allocate(20));
        element.getBuffer();
        element.getId();
        element.toString();
        element1.hashCode();

        Assert.assertTrue(element1.equals(element1));
        Assert.assertFalse(element1.equals(null));
        Assert.assertFalse(element1.equals(element));
    }
}
