/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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

package org.yuanrong.function;

import org.yuanrong.InvokeOptions;

import org.junit.Assert;
import org.junit.Test;

import java.util.HashMap;

public class TestCreateOptions {
    @Test
    public void testInitCreateOptions() {
        CreateOptions createOptions = new CreateOptions(10, 10);
        InvokeOptions invokeOptions = createOptions.convertInvokeOptions();
        createOptions.setCpu(10);
        createOptions.setMemory(10);
        Assert.assertNotNull(createOptions);
        Assert.assertNotNull(invokeOptions);

        HashMap<String, String> m = new HashMap<>();
        m.put("a", "b");
        createOptions.setAliasParams(m);
        Assert.assertNotNull(createOptions.getAliasParams().get("a"));

    }
}
