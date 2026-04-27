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

package org.yuanrong.api;

import org.junit.Assert;
import org.junit.Test;

import java.util.HashSet;

public class TestInvokeArg {
    @Test
    public void testInitInvokeArg() {
        InvokeArg invokeArg = new InvokeArg();
        InvokeArg invokeArg1 = new InvokeArg(new byte[] {});
        invokeArg.setData(new byte[]{});
        invokeArg.setNestedObjects(new HashSet<>());
        invokeArg.setObjectRef(true);
        invokeArg.setObjId("test-id");

        invokeArg1.getData();
        invokeArg1.getNestedObjects();
        invokeArg1.getObjectId();
        invokeArg1.getObjId();
        invokeArg1.toString();
        invokeArg1.canEqual(invokeArg);

        Assert.assertTrue(invokeArg.isObjectRef());
        Assert.assertFalse(invokeArg.equals(invokeArg1));
    }
}
