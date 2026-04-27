/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2024. All rights reserved.
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

package org.yuanrong.exception;

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ModuleCode;

import org.junit.Assert;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.powermock.core.classloader.annotations.PowerMockIgnore;
import org.powermock.modules.junit4.PowerMockRunner;

import java.io.PrintWriter;
import java.io.StringWriter;

@RunWith(PowerMockRunner.class)
@PowerMockIgnore("javax.management.*")
public class TestYRException {
    /**
     * Feature: ModuleCode
     * Description:
     * In the message of the exception thrown, the value of ModuleCode is a number.
     * Steps:
     * 1. Throws an YRException
     * 2. Check whether the message of the exception contains the ModuleCode number.
     * Expectation:
     * The exception stack trace contains ModuleCode number.
     */
    @Test
    public void testModuleCodeIsNumber() {
        String causedBy = "";
        try {
            throw new YRException(ErrorCode.ERR_BUS_DISCONNECTION, ModuleCode.CORE, "MockYRException");
        } catch (YRException e) {
            StringWriter stringWriter = new StringWriter();
            PrintWriter printWriter = new PrintWriter(stringWriter);
            e.printStackTrace(printWriter);
            causedBy = stringWriter.toString();
            printWriter.close();
        }
        System.err.println(causedBy);
        Assert.assertTrue(causedBy.contains("ModuleCode: 10"));
    }

    @Test
    public void testInitYRException() {
        YRException yrException = new YRException(ErrorCode.ERR_INCORRECT_INIT_USAGE,
            ModuleCode.CORE, "InitYRException");
        YRException yrException1 = new YRException(ErrorCode.ERR_ETCD_OPERATION_ERROR,
            ModuleCode.CORE, yrException);
        Exception exception = new Exception("test-Exception");
        YRException yrException2 = new YRException(ErrorCode.ERR_ETCD_OPERATION_ERROR,
            ModuleCode.CORE, exception);
        Throwable throwable = new Throwable("test-throwable");
        YRException yrException3 = new YRException(ErrorCode.ERR_ETCD_OPERATION_ERROR,
            ModuleCode.DATASYSTEM, "test-yrException", throwable);

        yrException3.hashCode();
        yrException3.setStackTrace(new StackTraceElement[] {});

        Assert.assertNotEquals(yrException1, yrException2);
        Assert.assertEquals(yrException1.getErrorCode().getValue(),
            yrException2.getErrorCode().getValue());
    }

    @Test
    public void testEquals() {
        YRException yrException = new YRException(ErrorCode.ERR_INCORRECT_INIT_USAGE,
            ModuleCode.CORE, "InitYRException");

        Assert.assertTrue(yrException.equals(yrException));
        Assert.assertFalse(yrException.equals(null));

        YRException yrException1 = new YRException(ErrorCode.ERR_INCORRECT_INIT_USAGE,
            ModuleCode.DATASYSTEM, "InitYRException");

        Assert.assertFalse(yrException.equals(yrException1));
    }
}
