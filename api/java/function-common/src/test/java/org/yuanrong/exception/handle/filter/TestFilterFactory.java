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

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.exception.YRException;
import org.yuanrong.exception.handler.filter.FilterFactory;
import org.yuanrong.exception.handler.traceback.StackTraceInfo;

import org.junit.Assert;
import org.junit.Test;

public class TestFilterFactory {
    @Test
    public void testInitFilterFactory() {
        FilterFactory filterFactory = new FilterFactory(new Throwable("test-user"), true);
        FilterFactory filterFactory1 = new FilterFactory(new Throwable("test-func"), false);

        Assert.assertNotEquals(filterFactory, filterFactory1);
    }

    @Test
    public void testGetThrowable() {
        FilterFactory filterFactory = new FilterFactory(new Throwable("test-user"), true);

        Assert.assertFalse(filterFactory.getThrowable().getMessage().isEmpty());
    }

    @Test
    public void testGenerateStackTraceInfo() {
        FilterFactory filterFactory = new FilterFactory(new Throwable("test-user"), true);
        YRException yrException = new YRException(ErrorCode.ERR_ETCD_OPERATION_ERROR,
            ModuleCode.CORE, "test-exception");
        StackTraceInfo stackTraceInfo = filterFactory.generateStackTraceInfo(
            yrException.getClass().getName(), "YRException");

        Assert.assertFalse(stackTraceInfo.getStackTraceElements().isEmpty());
    }
}
