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

package org.yuanrong.exception.handle.filter;

import org.yuanrong.exception.handler.filter.FuncCallStackTraceFilter;
import org.yuanrong.exception.handler.traceback.StackTraceInfo;

import org.junit.Assert;
import org.junit.Test;

public class TestFuncCallStackTraceFilter {
    @Test
    public void testInitFunc() {
        FuncCallStackTraceFilter funcCallStackTraceFilter = new FuncCallStackTraceFilter(new Throwable());
        FuncCallStackTraceFilter funcCallStackTraceFilter1 = new FuncCallStackTraceFilter(new Throwable("test"));
        FuncCallStackTraceFilter funcCallStackTraceFilter2 = new FuncCallStackTraceFilter(null);

        StackTraceInfo stackTraceInfo1 = funcCallStackTraceFilter1.generateStackTraceInfo("testClz1", "testFunc1");
        StackTraceInfo stackTraceInfo2 = funcCallStackTraceFilter2.generateStackTraceInfo("testClz2", "testFunc2");

        Assert.assertEquals(stackTraceInfo1.getMessage(),"test");
        Assert.assertEquals(stackTraceInfo2.getMessage(),"testFunc2");
        Assert.assertNotEquals(funcCallStackTraceFilter, funcCallStackTraceFilter1);
    }
}
