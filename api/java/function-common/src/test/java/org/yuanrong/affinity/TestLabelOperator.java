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

package org.yuanrong.affinity;

import org.junit.Assert;
import org.junit.Test;

import java.util.ArrayList;

public class TestLabelOperator {
    @Test
    public void testInitOperator() {

        LabelOperator labelOperator = new LabelOperator();
        LabelOperator labelOperator1 = new LabelOperator(OperatorType.LABEL_IN, "test-key1");
        LabelOperator labelOperator2 = new LabelOperator(OperatorType.LABEL_NOT_IN, "test-key2", new ArrayList<>());
        LabelOperator labelOperator3 = new LabelOperator(OperatorType.LABEL_IN);

        labelOperator.setOperatorType(OperatorType.LABEL_EXISTS);
        labelOperator.setKey("test-key");
        labelOperator.setValues(new ArrayList<>());
        labelOperator1.getOperatorType();
        labelOperator1.getKey();
        labelOperator1.getValues();
        labelOperator2.hashCode();
        labelOperator2.toString();
        labelOperator2.canEqual(labelOperator1);

        Assert.assertNotEquals(labelOperator, labelOperator2);
        Assert.assertFalse(labelOperator1.equals(labelOperator2));
    }

    @Test
    public void testGetOperateTypeValue() {
        LabelOperator labelOperator = new LabelOperator();
        Assert.assertEquals(0, labelOperator.getOperateTypeValue());

        LabelOperator labelOperator1 = new LabelOperator(OperatorType.LABEL_IN, "test-key1");
        Assert.assertEquals(1, labelOperator1.getOperateTypeValue());
    }
}
