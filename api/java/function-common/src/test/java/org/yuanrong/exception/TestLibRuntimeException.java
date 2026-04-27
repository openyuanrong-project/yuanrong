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

package org.yuanrong.exception;

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ModuleCode;

import org.junit.Assert;
import org.junit.Test;

public class TestLibRuntimeException {
    @Test
    public void testInitException() {
        LibRuntimeException libRuntimeException = new LibRuntimeException("test-LibRuntime");
        libRuntimeException.setErrorCode(ErrorCode.ERR_ETCD_OPERATION_ERROR);
        libRuntimeException.setModuleCode(ModuleCode.CORE);
        libRuntimeException.setMessage("test");

        libRuntimeException.getMessage();
        libRuntimeException.getModuleCode();
        libRuntimeException.getErrorCode();

        LibRuntimeException libRuntimeException1 = new LibRuntimeException(ErrorCode.ERR_ETCD_OPERATION_ERROR,
            ModuleCode.CORE, "test");
        libRuntimeException1.hashCode();
        libRuntimeException1.toString();
        libRuntimeException.canEqual(libRuntimeException1);

        Assert.assertTrue(libRuntimeException.equals(libRuntimeException1));
    }

}
