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
import java.util.List;

public class TestCondition {
    @Test
    public void testInitCondition() {

        String labelKey = "key";
        List<String> values = new ArrayList<>();
        values.add("value1");
        values.add("value2");

        Condition condition = new Condition(labelKey, OperatorType.LABEL_IN, values);
        SubCondition subCondition = condition.getSubConditions().get(0);
        LabelExpression expression = subCondition.getExpressions().get(0);
        Assert.assertEquals(labelKey, expression.getKey());
        LabelOperator operator = expression.getOperator();
        Assert.assertEquals(values, operator.getValues());
    }

    @Test
    public void testInitCondition2() {

        String labelKey1 = "key1";
        List<String> values1 = new ArrayList<>();
        values1.add("value11");
        values1.add("value12");
        SubCondition subCondition1 = new SubCondition(labelKey1, OperatorType.LABEL_IN, values1);

        String labelKey2 = "key2";
        List<String> values2 = new ArrayList<>();
        values1.add("value21");
        values1.add("value22");
        SubCondition subCondition2 = new SubCondition(labelKey2, OperatorType.LABEL_NOT_IN, values2);

        List<SubCondition> subConditions = new ArrayList<>();
        subConditions.add(subCondition1);
        subConditions.add(subCondition2);

        Condition condition = new Condition(subConditions, true);
        SubCondition subCondition = condition.getSubConditions().get(0);
        LabelExpression expression = subCondition.getExpressions().get(0);
        Assert.assertEquals(labelKey1, expression.getKey());
        LabelOperator operator = expression.getOperator();
        Assert.assertEquals(values1, operator.getValues());
        Assert.assertTrue(condition.isOrderPriority());
    }

    @Test
    public void testInitSubCondition3() {

        String labelKey1 = "key1";
        List<String> values1 = new ArrayList<>();
        values1.add("value11");
        values1.add("value12");
        LabelExpression expression1 = new LabelExpression(labelKey1, OperatorType.LABEL_IN, values1);
        int weight1 = 10;
        SubCondition subCondition1 = new SubCondition(expression1, weight1);

        String labelKey2 = "key2";
        List<String> values2 = new ArrayList<>();
        values1.add("value21");
        values1.add("value22");
        LabelExpression expression2 = new LabelExpression(labelKey2, OperatorType.LABEL_NOT_IN, values2);
        int weight2 = 20;
        SubCondition subCondition2 = new SubCondition(expression2, weight2);

        List<SubCondition> subConditions = new ArrayList<>();
        subConditions.add(subCondition1);
        subConditions.add(subCondition2);

        Condition condition = new Condition(subConditions, false);
        SubCondition subCondition = condition.getSubConditions().get(0);
        LabelExpression expression = subCondition.getExpressions().get(0);
        int weight = subCondition.getWeight();
        Assert.assertEquals(weight1, weight);
        Assert.assertEquals(labelKey1, expression.getKey());
        LabelOperator operator = expression.getOperator();
        Assert.assertEquals(values1, operator.getValues());
        Assert.assertFalse(condition.isOrderPriority());
    }

    @Test
    public void testInitCondition3() {
        Condition condition = new Condition();
        Condition condition1 = new Condition(new ArrayList<>());
        Condition condition2 = new Condition(new LabelExpression());
        Condition condition3 = new Condition("testKey", OperatorType.LABEL_NOT_IN);
        Assert.assertNotNull(condition);
        Assert.assertNotNull(condition1);
        Assert.assertNotNull(condition2);
        Assert.assertNotNull(condition3);

        condition.setSubConditions(new ArrayList<>());
        condition.setOrderPriority(false);
        condition.hashCode();
        condition.toString();
        condition1.canEqual(condition2);
        Assert.assertFalse(condition1.equals(condition3));
    }
}
