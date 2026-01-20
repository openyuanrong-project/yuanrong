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

import org.junit.Assert;
import org.junit.Test;

import java.util.Set;

public class TestStatusCode {
    @Test
    public void testInitStatus() {
        StatusCode statusCode = new StatusCode();
        int status = StatusCode.K_OUT_OF_MEMORY;
        Set<Integer> statusSet = StatusCode.RETRY_ERR_CDOE_SET;
        Assert.assertTrue(statusSet.contains(status));
    }
}
