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

package org.yuanrong.executor;

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ModuleCode;

import org.junit.Assert;
import org.junit.Test;

import java.nio.ByteBuffer;
import java.util.ArrayList;

public class TestReturnType {
    @Test
    public void testInitReturnType() {
        ReturnType returnType = new ReturnType(ErrorCode.ERR_JOB_USER_CODE_EXCEPTION);
        ReturnType returnType1 = new ReturnType(ErrorCode.ERR_USER_CODE_LOAD, ByteBuffer.allocate(10));
        ReturnType returnType2 = new ReturnType(ErrorCode.ERR_OK, ModuleCode.CORE, "test2");
        ReturnType returnType3 = new ReturnType(ErrorCode.ERR_CREATE_RETURN_BUFFER, new byte[] {});
        ReturnType returnType4 = new ReturnType(ErrorCode.ERR_USER_CODE_LOAD, "test4");
        ReturnType returnType5 = new ReturnType(ErrorCode.ERR_DATASYSTEM_FAILED, ModuleCode.CORE, "test5",
            new byte[] {});
        ReturnType returnType6 = new ReturnType(ErrorCode.ERR_BUS_DISCONNECTION, ModuleCode.CORE, "test6",
            new ArrayList<>());
        returnType6.getBytes();
        Assert.assertNotEquals(returnType1, returnType2);
        Assert.assertNotEquals(returnType, returnType3);
        Assert.assertNotEquals(returnType4, returnType5);
    }
}
