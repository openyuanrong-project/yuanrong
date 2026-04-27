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

package org.yuanrong.affinity;

import org.junit.Assert;
import org.junit.Test;

public class TestSelector {
    @Test
    public void testInitSelector() {
        Selector selector = new Selector();

        Condition cond = new Condition();
        Selector selector1 = new Selector(cond);
        Assert.assertEquals(cond, selector1.getCondition());

        selector.setCondition(new Condition());
        selector1.getCondition();
        selector1.hashCode();
        selector1.toString();
        selector1.canEqual(selector);

        Assert.assertNotEquals(selector, selector1);
        Assert.assertFalse(selector.equals(selector1));
    }

}
