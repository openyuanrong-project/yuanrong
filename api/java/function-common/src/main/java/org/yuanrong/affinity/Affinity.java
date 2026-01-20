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
 * Description:
 *
 * @since 2023-11-26
 */
@Data
public class Affinity {
    private AffinityKind affinityKind;

    private AffinityType affinityType;

    private List<LabelOperator> labelOperators;

    private AffinityScope affinityScope;

    private ResourceAffinity resourceAffinity;

    private InstanceAffinity instanceAffinity;

    /**
     * Init Affinity
     *
     * @param affinityKind the affinityKind
     * @param affinityType the affinityType
     * @param labelOperators the labelOperators
     */
    public Affinity(AffinityKind affinityKind, AffinityType affinityType, List<LabelOperator> labelOperators) {
        this.affinityType = affinityType;
        this.affinityKind = affinityKind;
        this.labelOperators = labelOperators;
    }

    /**
     * Init Affinity with affinityScope
     *
     * @param affinityKind the affinityKind
     * @param affinityType the affinityType
     * @param labelOperators the labelOperators
     * @param affinityScope the affinityScope
     */
    public Affinity(AffinityKind affinityKind, AffinityType affinityType, List<LabelOperator> labelOperators,
        AffinityScope affinityScope) {
        this(affinityKind, affinityType, labelOperators);
        this.affinityScope = affinityScope;
    }

    /**
     * Init Affinity
     *
     * @param resourceAffinity the resource affinity
     */
    public Affinity(ResourceAffinity resourceAffinity) {
        this(resourceAffinity, null);
    }

    /**
     * Init Affinity
     *
     * @param instanceAffinity the instance affinity
     */
    public Affinity(InstanceAffinity instanceAffinity) {
        this(null, instanceAffinity);
    }

    /**
     * Init Affinity
     *
     * @param resourceAffinity the resource affinity
     * @param instanceAffinity the instance affinity
     */
    public Affinity(ResourceAffinity resourceAffinity, InstanceAffinity instanceAffinity) {
        this.resourceAffinity = resourceAffinity;
        this.instanceAffinity = instanceAffinity;
    }

    /**
     * get affinityKind value
     *
     * @return value
     */
    public int getAffinityValue() {
        if (this.affinityKind == null || this.affinityType == null) {
            return 0;
        }
        return this.affinityKind.getKind() + this.affinityType.getType();
    }

    /**
     * get affinityKind value
     *
     * @return value
     */
    public int getAffinityScopeValue() {
        if (this.affinityScope == null) {
            return 0;
        }
        return this.affinityScope.getScope();
    }
}
