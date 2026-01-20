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
import org.yuanrong.exception.handler.filter.StackTraceFilter;
import org.yuanrong.exception.handler.filter.UserStackTraceFilter;
import org.yuanrong.exception.handler.traceback.StackTraceInfo;

import org.junit.Assert;
import org.junit.Test;

public class TestUserStackTraceFilter {
    @Test
    public void testInitUserStack() {
        UserStackTraceFilter userStackTraceFilter = new UserStackTraceFilter(new Throwable("test-throwable"));
        userStackTraceFilter.setCause(new Throwable());

        userStackTraceFilter.getCause();
        userStackTraceFilter.hashCode();
        userStackTraceFilter.toString();
        UserStackTraceFilter userStackTraceFilter1 = new UserStackTraceFilter(new Throwable("test-user"));

        Assert.assertFalse(userStackTraceFilter.equals(userStackTraceFilter1));
    }

    @Test
    public void testGenerateStackTraceInfo() {
        YRException yrException = new YRException(ErrorCode.ERR_ETCD_OPERATION_ERROR,
            ModuleCode.CORE, "test-exception");
        UserStackTraceFilter userStackTraceFilter = new UserStackTraceFilter(yrException);
        StackTraceInfo stackTraceInfo = userStackTraceFilter.generateStackTraceInfo(
            yrException.getClass().getName(), "YRException");

        Assert.assertFalse(stackTraceInfo.getStackTraceElements().isEmpty());

        StackTraceFilter stackTraceFilter = new StackTraceFilter(yrException) { };
        StackTraceInfo traceInfo = stackTraceFilter.generateStackTraceInfo(yrException.getClass().getName(),
            "YRException");

        Assert.assertFalse(traceInfo.getStackTraceElements().isEmpty());
    }

    @Test
    public void testGetPrintStackTrace() {
        YRException yrException = new YRException(ErrorCode.ERR_ETCD_OPERATION_ERROR,
            ModuleCode.CORE, "test-exception");
        UserStackTraceFilter userStackTraceFilter = new UserStackTraceFilter(yrException);

        Assert.assertFalse(userStackTraceFilter.getPrintStackTrace(yrException).isEmpty());
    }

    @Test
    public void testGetCausedByString() {
        YRException yrException = new YRException(ErrorCode.ERR_ETCD_OPERATION_ERROR,
            ModuleCode.CORE, "test-exception");
        UserStackTraceFilter userStackTraceFilter = new UserStackTraceFilter(yrException);

        Assert.assertFalse(userStackTraceFilter.getCausedByString(yrException).isEmpty());
    }
}
