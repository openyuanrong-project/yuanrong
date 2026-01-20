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
 * Description: Condition.
 *
 * Combined with multiple SubConditions.
 *
 * It's OR realtion between diffrent SubConditions.
 *
 * And there's a flag to mean in order of priority instead of weights rank.
 *
 * @since 2024-09-11
 */
@Data
public class Condition {
    /**
     * The values set to check for operator type LABEL_IN and LABEL_NOT_IN.
     */
    private List<SubCondition> subConditions;

    /**
     * The flag to mean order priority used instead of weights rank.
     */
    private boolean orderPriority;

    /**
     * Init with one label key, operator type
     *
     * @param key the key
     * @param operatorType the operatorType
     */
    public Condition(String key, OperatorType operatorType) {
        this(new SubCondition(key, operatorType, new ArrayList<>()));
    }

    /**
     * Init with one label key, operator type and label values
     *
     * @param key the key
     * @param operatorType the operatorType
     * @param values the values
     */
    public Condition(String key, OperatorType operatorType, List<String> values) {
        this(new SubCondition(key, operatorType, values));
    }

    /**
     * Init with one LabelExpression
     *
     * @param expression only one LabelExpression
     */
    public Condition(LabelExpression expression) {
        this(new SubCondition(expression));
    }

    /**
     * Init with multiple LabelExpressions
     *
     * @param expressions expressions with AND relation
     */
    public Condition(List<LabelExpression> expressions) {
        this(new SubCondition(expressions));
    }

    /**
     * Init with one SubCondition
     *
     * @param subCondition only one SubCondition
     */
    public Condition(SubCondition subCondition) {
        this.subConditions = new ArrayList<>();
        this.subConditions.add(subCondition);
        this.orderPriority = false;
    }

    /**
     * Init Condition
     *
     * @param subConditions the sub conditions with OR relation
     * @param orderPriority the order priority
     */
    public Condition(List<SubCondition> subConditions, boolean orderPriority) {
        this.subConditions = subConditions;
        this.orderPriority = orderPriority;
    }

    /**
     * Init Condition
     */
    public Condition() {}
}
