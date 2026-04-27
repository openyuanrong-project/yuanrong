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

import java.util.List;

/**
 * Description: Label Expression.
 *
 * Combined with label key and one Label Operator.
 *
 * @since 2024-09-11
 */
@Data
public class LabelExpression {
    /**
     * The label key to apply the operator.
     *
     */
    private String key;

    /**
     * The label operator
     */
    private LabelOperator operator;

    /**
     * Init LabelExpression
     *
     * @param key the key
     * @param operatorType the operatorType
     */
    public LabelExpression(String key, OperatorType operatorType) {
        this(key, new LabelOperator(operatorType));
    }

    /**
     * Init LabelExpression
     *
     * @param key the key
     * @param operatorType the operatorType
     * @param values the values
     */
    public LabelExpression(String key, OperatorType operatorType, List<String> values) {
        this(key, new LabelOperator(operatorType, values));
    }

    /**
     * Init LabelExpression
     *
     * @param key the key
     * @param operator the operator
     */
    public LabelExpression(String key, LabelOperator operator) {
        this.key = key;
        this.operator = operator;
    }

    /**
     * Init LabelExpression
     */
    public LabelExpression() {}
}
