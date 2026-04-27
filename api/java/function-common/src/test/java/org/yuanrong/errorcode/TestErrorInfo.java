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

package org.yuanrong.errorcode;

import org.junit.Assert;
import org.junit.Test;

import java.util.ArrayList;

public class TestErrorInfo {
    @Test
    public void testErrorInfo() {
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_ETCD_OPERATION_ERROR, ModuleCode.RUNTIME_INVOKE, "test");

        errorInfo.setStackTraceInfos(new ArrayList<>());
        errorInfo.toString();
        Assert.assertEquals("test", errorInfo.getErrorMessage());
    }
}
