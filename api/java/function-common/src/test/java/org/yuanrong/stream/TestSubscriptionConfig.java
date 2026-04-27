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

package org.yuanrong.stream;

import org.junit.Assert;
import org.junit.Test;

public class TestSubscriptionConfig {
    @Test
    public void testInitSubscriptionConfig() {
        SubscriptionConfig.Builder builder = SubscriptionConfig.builder();
        SubscriptionConfig.Builder builder1 = new SubscriptionConfig.Builder();
        builder1.subscriptionName("test");
        builder1.subscriptionType(SubscriptionType.STREAM);
        SubscriptionConfig subscriptionConfig = builder1.build();
        Assert.assertEquals("test", subscriptionConfig.getSubscriptionName());
    }
}
