/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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

import lombok.Data;

import java.util.ArrayList;
import java.util.List;

/**
 * Description: Sub condition.
 *
 * Combined with multiple LabelExpressions.
 *
 * It's AND realtion between diffrent LabelExpressions.
 *
 * And there's a weight for this Sub Condition.
 *
 * @since 2024-09-11
 */
@Data
public class SubCondition {
    /**
     * The values set to check for operator type LABEL_IN and LABEL_NOT_IN.
     */
    private List<LabelExpression> expressions;

    /**
     * The weight.
     */
    private int weight;

    /**
     * Init with one label key, operator type and label values
     *
     * @param key the key
     * @param operatorType the operatorType
     * @param values the values
     */
    public SubCondition(String key, OperatorType operatorType, List<String> values) {
        this(new LabelExpression(key, operatorType, values));
    }

    /**
     * Init with one label key and label operator
     *
     * @param key the key
     * @param operator the operator
     */
    public SubCondition(String key, LabelOperator operator) {
        this(new LabelExpression(key, operator));
    }

    /**
     * Init with one LabelExpression and no weight
     *
     * @param expression only one LabelExpression
     */
    public SubCondition(LabelExpression expression) {
        this(expression, 0);
    }

    /**
     * Init SubCondition with no weight
     *
     * @param expressions the expressions
     */
    public SubCondition(List<LabelExpression> expressions) {
        this(expressions, 0);
    }

    /**
     * Init with one LabelExpression and weight
     *
     * @param expression only one LabelExpression
     * @param weight the weight
     */
    public SubCondition(LabelExpression expression, int weight) {
        this.expressions = new ArrayList<>();
        this.expressions.add(expression);
        this.weight = weight;
    }

    /**
     * Init SubCondition
     *
     * @param expressions the expressions
     * @param weight the weight
     */
    public SubCondition(List<LabelExpression> expressions, int weight) {
        this.expressions = expressions;
        this.weight = weight;
    }

    /**
     * Init SubCondition
     */
    public SubCondition() {}
}
