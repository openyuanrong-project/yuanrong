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
 * Description: Label Operator
 *
 * Currently one Label Operator has one key field.
 *
 * The key field will be removed in future to align with POSIX.
 *
 * Use LabelExpression instead in future to specify the key.
 *
 * @since 2023-11-26
 */
@Data
public class LabelOperator {
    /**
     * The operator type.
     *
     */
    private OperatorType operatorType;

    /**
     * The key to apply this operator.
     *
     * This field will be removed in future as one common key can be associated with multiple label operators.
     *
     */
    private String key;

    /**
     * The values set to check for operator type LABEL_IN and LABEL_NOT_IN.
     */
    private List<String> values;

    /**
     * Init LabelOperator
     *
     * @param operatorType the operatorType
     * @param key the key
     */
    public LabelOperator(OperatorType operatorType, String key) {
        this(operatorType, key, new ArrayList<>());
    }

    /**
     * Init LabelOperator
     *
     * @param operatorType the operatorType
     * @param key the key
     * @param values the values
     */
    public LabelOperator(OperatorType operatorType, String key, List<String> values) {
        this.operatorType = operatorType;
        this.key = key;
        this.values = values;
    }

    /**
     * Init LabelOperator
     *
     * @param operatorType the operatorType
     * @param values the values
     */
    public LabelOperator(OperatorType operatorType, List<String> values) {
        this.operatorType = operatorType;
        this.values = values;
    }

    /**
     * Init LabelOperator
     *
     * @param operatorType the operatorType
     */
    public LabelOperator(OperatorType operatorType) {
        this(operatorType, new ArrayList<>());
    }

    /**
     * Init LabelOperator
     */
    public LabelOperator() {}

    /**
     * get operatorType type
     *
     * @return int
     */
    public int getOperateTypeValue() {
        if (this.operatorType == null) {
            return 0;
        }
        return this.operatorType.getType();
    }
}
