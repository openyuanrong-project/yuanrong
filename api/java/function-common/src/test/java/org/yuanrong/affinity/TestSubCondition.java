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

public class TestSubCondition {
    @Test
    public void testInitSubCondition() {

        String labelKey = "key";
        List<String> values = new ArrayList<>();
        values.add("value1");
        values.add("value2");

        SubCondition subCondition = new SubCondition(labelKey, OperatorType.LABEL_IN, values);
        LabelExpression expression = subCondition.getExpressions().get(0);
        Assert.assertEquals(labelKey, expression.getKey());
        LabelOperator operator = expression.getOperator();
        Assert.assertEquals(values, operator.getValues());
        SubCondition subCondition1 = new SubCondition("test", new LabelOperator());
        SubCondition subCondition2 = new SubCondition();
        subCondition2.setExpressions(new ArrayList<>());
        subCondition2.setWeight(1);
        subCondition1.toString();
        subCondition2.hashCode();
        Assert.assertFalse(subCondition1.equals(subCondition2));
    }

    @Test
    public void testInitSubConditionOperator() {
        String labelKey = "key";
        LabelOperator operator = new LabelOperator();
        SubCondition subCondition = new SubCondition(labelKey, operator);
        LabelExpression ex = subCondition.getExpressions().get(0);
        Assert.assertEquals(labelKey, ex.getKey());
        Assert.assertEquals(operator, ex.getOperator());
        Assert.assertEquals(0, subCondition.getWeight());
    }

    @Test
    public void testInitSubConditionExpression() {
        LabelExpression ex = new LabelExpression();
        SubCondition subCondition = new SubCondition(ex);
        Assert.assertEquals(ex, subCondition.getExpressions().get(0));
        Assert.assertEquals(0, subCondition.getWeight());
    }

    @Test
    public void testInitSubConditionExpressions() {
        LabelExpression ex1 = new LabelExpression();
        LabelExpression ex2 = new LabelExpression();
        List<LabelExpression> expressions = new ArrayList<>();
        expressions.add(ex1);
        expressions.add(ex2);
        SubCondition subCondition = new SubCondition(expressions);
        Assert.assertEquals(expressions, subCondition.getExpressions());
        Assert.assertEquals(0, subCondition.getWeight());
    }

    @Test
    public void testInitSubConditionExpressionWeight() {
        String labelKey = "key";
        List<String> values = new ArrayList<>();
        values.add("value1");
        values.add("value2");
        LabelExpression ex = new LabelExpression(labelKey, OperatorType.LABEL_IN, values);
        int weight = 10;
        SubCondition subCondition = new SubCondition(ex, weight);

        LabelExpression expression = subCondition.getExpressions().get(0);
        Assert.assertEquals(labelKey, expression.getKey());
        LabelOperator operator = expression.getOperator();
        Assert.assertEquals(values, operator.getValues());
        Assert.assertEquals(weight, subCondition.getWeight());
    }

    @Test
    public void testInitSubConditionExpressionsWeight() {
        LabelExpression ex1 = new LabelExpression();
        LabelExpression ex2 = new LabelExpression();
        List<LabelExpression> expressions = new ArrayList<>();
        expressions.add(ex1);
        expressions.add(ex2);

        int weight = 10;

        SubCondition subCondition = new SubCondition(expressions, weight);

        Assert.assertEquals(expressions, subCondition.getExpressions());
        Assert.assertEquals(weight, subCondition.getWeight());
    }

}
