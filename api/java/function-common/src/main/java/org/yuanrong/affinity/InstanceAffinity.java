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
 * Description: InstanceAffinity.
 *
 * Specified with different selector for different affinity type.
 *
 * And an affinity scope is required.
 *
 * @since 2024-09-11
 */
@Data
public class InstanceAffinity {
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
     * The affinity scope.
     */
    private AffinityScope scope;

    /**
     * Init with one label key, operator type
     *
     * @param key the key
     * @param operatorType the operatorType
     */
    public InstanceAffinity(String key, OperatorType operatorType) {
        this(new Selector(new Condition(key, operatorType)));
    }

    /**
     * Init with one label key, operator type and label values
     *
     * @param key the key
     * @param operatorType the operatorType
     * @param values the values
     */
    public InstanceAffinity(String key, OperatorType operatorType, List<String> values) {
        this(new Selector(new Condition(key, operatorType, values)));
    }

    /**
     * Init InstanceAffinity with default PreferredAffinity
     *
     * @param expression the expression
     */
    public InstanceAffinity(LabelExpression expression) {
        this(new Selector(new Condition(expression)));
    }

    /**
     * Init InstanceAffinity with default PreferredAffinity
     *
     * @param expressions the expressions with AND relation
     */
    public InstanceAffinity(List<LabelExpression> expressions) {
        this(new Selector(new Condition(expressions)));
    }

    /**
     * Init InstanceAffinity with default PreferredAffinity
     *
     * @param subConditions the sub conditions with OR relation
     * @param orderPriority the order priority
     */
    public InstanceAffinity(List<SubCondition> subConditions, boolean orderPriority) {
        this(new Selector(new Condition(subConditions, orderPriority)));
    }

    /**
     * Init InstanceAffinity with default PreferredAffinity
     *
     * @param preferredAffinity the selector using PreferredAffinity
     */
    public InstanceAffinity(Selector preferredAffinity) {
        this(AffinityType.PREFERRED, preferredAffinity, AffinityScope.NODE);
    }

    /**
     * Init InstanceAffinity with default PreferredAffinity
     *
     * @param preferredAffinity the selector using PreferredAffinity
     * @param scope the affinity scope
     */
    public InstanceAffinity(Selector preferredAffinity, AffinityScope scope) {
        this(AffinityType.PREFERRED, preferredAffinity, scope);
    }

    /**
     * Init InstanceAffinity
     *
     * @param type the affinity type
     * @param selector the selector for given type
     * @param scope the affinity scope
     */
    public InstanceAffinity(AffinityType type, Selector selector, AffinityScope scope) {
        this.preferredAffinity = null;
        this.preferredAntiAffinity = null;
        this.requiredAffinity = null;
        this.requiredAntiAffinity = null;
        switch (type) {
            case PREFERRED:
                preferredAffinity = selector;
                break;
            case PREFERRED_ANTI:
                preferredAntiAffinity = selector;
                break;
            case REQUIRED:
                requiredAffinity = selector;
                break;
            case REQUIRED_ANTI:
                requiredAntiAffinity = selector;
                break;
        }
        this.scope = scope;
    }

    /**
     * Init InstanceAffinity with default node scope
     *
     * @param preferredAffinity the selector used PreferredAffinity
     * @param preferredAntiAffinity the selector used PreferredAntiAffinity
     * @param requiredAffinity the selector used RequiredAffinity
     * @param requiredAntiAffinity the selector used RequiredAntiAffinity
     */
    public InstanceAffinity(Selector preferredAffinity, Selector preferredAntiAffinity,
                            Selector requiredAffinity, Selector requiredAntiAffinity) {
        this(preferredAffinity, preferredAntiAffinity, requiredAffinity, requiredAntiAffinity, AffinityScope.NODE);
    }

    /**
     * Init InstanceAffinity
     *
     * @param preferredAffinity the selector used PreferredAffinity
     * @param preferredAntiAffinity the selector used PreferredAntiAffinity
     * @param requiredAffinity the selector used RequiredAffinity
     * @param requiredAntiAffinity the selector used RequiredAntiAffinity
     * @param scope the affinity scope
     */
    public InstanceAffinity(Selector preferredAffinity, Selector preferredAntiAffinity,
                            Selector requiredAffinity, Selector requiredAntiAffinity, AffinityScope scope) {
        this.preferredAffinity = preferredAffinity;
        this.preferredAntiAffinity = preferredAntiAffinity;
        this.requiredAffinity = requiredAffinity;
        this.requiredAntiAffinity = requiredAntiAffinity;
        this.scope = scope;
    }

    /**
     * Init InstanceAffinity
     */
    public InstanceAffinity() {}
}
