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

package org.yuanrong;

import org.yuanrong.function.CppInstanceClass;
import org.yuanrong.function.YRFunc0;
import org.yuanrong.function.YRFunc1;
import org.yuanrong.function.YRFunc2;
import org.yuanrong.function.YRFunc3;
import org.yuanrong.function.YRFunc4;
import org.yuanrong.function.YRFunc5;

import org.junit.Assert;
import org.junit.Test;

public class TestYRGetInstance {
    @Test
    public void testInitYRGetInstance() {
        YRGetInstance yrGetInstance = new YRGetInstance();
        YRGetInstance.getInstance((YRFunc0<Object>) () -> null, "test0");
        YRGetInstance.getInstance((YRFunc0<Object>) () -> null, "test0", "test00");
        YRGetInstance.getInstance((YRFunc1<Object, Object>) o -> null, "test1");
        YRGetInstance.getInstance((YRFunc1<Object, Object>) o -> null, "test1", "test11");
        YRGetInstance.getInstance((YRFunc2<Object, Object, Object>) (o, o2) -> null, "test2");
        YRGetInstance.getInstance((YRFunc2<Object, Object, Object>) (o, o2) -> null, "test2", "test22");
        YRGetInstance.getInstance((YRFunc3<Object, Object, Object, Object>) (o, o2, o3) -> null, "test3");
        YRGetInstance.getInstance((YRFunc3<Object, Object, Object, Object>) (o, o2, o3) -> null, "test3", "test33");
        YRGetInstance.getInstance((YRFunc4<Object, Object, Object, Object, Object>) (o, o2, o3, o4) -> null, "test4");
        YRGetInstance.getInstance((YRFunc4<Object, Object, Object, Object, Object>) (o, o2, o3, o4) -> null, "test4",
            "test44");
        YRGetInstance.getInstance((YRFunc5<Object, Object, Object, Object, Object, Object>) (o, o2, o3, o4, o5) -> null,
            "test5");
        YRGetInstance.getInstance((YRFunc5<Object, Object, Object, Object, Object, Object>) (o, o2, o3, o4, o5) -> null,
            "test5", "test55");
        YRGetInstance.getCppInstance(CppInstanceClass.of("testCpp", "Add"), "test");
        YRGetInstance.getCppInstance(CppInstanceClass.of("testCpp", "Add"), "test", "testNameSpace");
        Assert.assertNotNull(yrGetInstance);
    }
}
