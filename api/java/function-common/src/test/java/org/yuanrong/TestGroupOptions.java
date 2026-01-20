/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
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

package org.yuanrong;

import org.junit.Assert;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.powermock.core.classloader.annotations.PowerMockIgnore;
import org.powermock.modules.junit4.PowerMockRunner;

@RunWith(PowerMockRunner.class)
@PowerMockIgnore("javax.management.*")
public class TestGroupOptions {
    @Test
    public void testInitGroupOptions() {
        GroupOptions groupOptions = new GroupOptions();
        groupOptions.setTimeout(1);
        groupOptions.getTimeout();
        groupOptions.hashCode();
        groupOptions.toString();
        GroupOptions testGroupOptions = new GroupOptions(2);
        groupOptions.canEqual(testGroupOptions);
        GroupOptions groupOptions1 = new GroupOptions(10, true);
        groupOptions1.setSameLifecycle(false);
        Assert.assertFalse(groupOptions.equals(testGroupOptions));
    }
}
