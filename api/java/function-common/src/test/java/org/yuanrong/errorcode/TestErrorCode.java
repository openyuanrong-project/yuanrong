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

public class TestErrorCode {
    @Test
    public void testInitErrorCode() {
        ErrorCode errorCode = new ErrorCode(1);
        errorCode.hashCode();
        Assert.assertTrue(errorCode.equals(errorCode));
        Assert.assertFalse(errorCode.equals(null));
    }

    @Test
    public void testInitModuleCode() {
        ModuleCode.CORE.toString();
        Assert.assertEquals(10, ModuleCode.CORE.getCode());
    }
}
