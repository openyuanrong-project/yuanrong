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
 * Description: ResourceAffinity.
 *
 * Specified with different selector for different affinity type.
 *
 * @since 2024-09-11
 */
@Data
public class ResourceAffinity {
    /**
     * The selector for PreferredAffinity.
     */
    private Selector preferredAffinity;

    /**
     * The selector for PreferredAntiAffinity.
     */
    private Selector preferredAntiAffinity;

    /**
     * The selector for RequiredAffinity.
     */
    private Selector requiredAffinity;

    /**
     * The selector for RequiredAntiAffinity.
     */
    private Selector requiredAntiAffinity;

    /**
     * Init with one label key, operator type
     *
     * @param key the key
     * @param operatorType the operatorType
     */
    public ResourceAffinity(String key, OperatorType operatorType) {
        this(new Selector(new Condition(key, operatorType)));
    }

    /**
     * Init with one label key, operator type and label values
     *
     * @param key the key
     * @param operatorType the operatorType
     * @param values the values
     */
    public ResourceAffinity(String key, OperatorType operatorType, List<String> values) {
        this(new Selector(new Condition(key, operatorType, values)));
    }

    /**
     * Init ResourceAffinity with default PreferredAffinity
     *
     * @param expression the expression
     */
    public ResourceAffinity(LabelExpression expression) {
        this(new Selector(new Condition(expression)));
    }

    /**
     * Init ResourceAffinity with default PreferredAffinity
     *
     * @param expressions the expressions with AND relation
     */
    public ResourceAffinity(List<LabelExpression> expressions) {
        this(new Selector(new Condition(expressions)));
    }

    /**
     * Init ResourceAffinity with default PreferredAffinity
     *
     * @param subConditions the sub conditions with OR relation
     * @param orderPriority the order priority
     */
    public ResourceAffinity(List<SubCondition> subConditions, boolean orderPriority) {
        this(new Selector(new Condition(subConditions, orderPriority)));
    }

    /**
     * Init ResourceAffinity with default PreferredAffinity
     *
     * @param preferredAffinity the selector using PreferredAffinity
     */
    public ResourceAffinity(Selector preferredAffinity) {
        this(AffinityType.PREFERRED, preferredAffinity);
    }

    /**
     * Init ResourceAffinity
     *
     * @param type the affinity type
     * @param selector the selector for given type
     */
    public ResourceAffinity(AffinityType type, Selector selector) {
        this.preferredAffinity = null;
        this.preferredAntiAffinity = null;
        this.requiredAffinity = null;
        this.requiredAntiAffinity = null;
        switch (type) {
            case PREFERRED:
                this.preferredAffinity = selector;
                break;
            case PREFERRED_ANTI:
                this.preferredAntiAffinity = selector;
                break;
            case REQUIRED:
                this.requiredAffinity = selector;
                break;
            case REQUIRED_ANTI:
                this.requiredAntiAffinity = selector;
                break;
        }
    }

    /**
     * Init ResourceAffinity
     *
     * @param preferredAffinity the selector using PreferredAffinity
     * @param preferredAntiAffinity the selector using PreferredAntiAffinity
     * @param requiredAffinity the selector using RequiredAffinity
     * @param requiredAntiAffinity the selector using RequiredAntiAffinity
     */
    public ResourceAffinity(Selector preferredAffinity, Selector preferredAntiAffinity,
                            Selector requiredAffinity, Selector requiredAntiAffinity) {
        this.preferredAffinity = preferredAffinity;
        this.preferredAntiAffinity = preferredAntiAffinity;
        this.requiredAffinity = requiredAffinity;
        this.requiredAntiAffinity = requiredAntiAffinity;
    }

    /**
     * Init ResourceAffinity
     */
    public ResourceAffinity() {}
}
