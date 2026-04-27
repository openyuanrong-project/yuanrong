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

public class TestAffinity {
    @Test
    public void testInitAffinity() {

        Affinity affinity = new Affinity(AffinityKind.RESOURCE, AffinityType.PREFERRED, new ArrayList<>());
        affinity.setAffinityKind(AffinityKind.RESOURCE);
        affinity.setAffinityType(AffinityType.PREFERRED_ANTI);
        affinity.setLabelOperators(new ArrayList<>());

        affinity.getAffinityKind();
        affinity.getAffinityType();
        affinity.getLabelOperators();

        Affinity affinity1 = new Affinity(AffinityKind.RESOURCE, AffinityType.PREFERRED, new ArrayList<>());
        affinity1.hashCode();
        affinity1.toString();
        affinity1.canEqual(affinity);

        Affinity affinity2 = new Affinity(new InstanceAffinity());
        Affinity affinity3 = new Affinity(new ResourceAffinity());
        Affinity affinity4 = new Affinity(new ResourceAffinity(), new InstanceAffinity());
        affinity2.setResourceAffinity(new ResourceAffinity());
        affinity3.setInstanceAffinity(new InstanceAffinity());
        Assert.assertFalse(affinity.equals(affinity1));
        Assert.assertTrue(affinity.equals(affinity));

        Affinity affinity5 = new Affinity(AffinityKind.INSTANCE, AffinityType.PREFERRED, new ArrayList<>(), AffinityScope.POD);
        Assert.assertEquals(affinity5.getAffinityScope(), AffinityScope.POD);
    }

    @Test
    public void testInitAffinityResource() {

        String labelKey = "blue";
        ResourceAffinity resourceAffinity = new ResourceAffinity(labelKey, OperatorType.LABEL_EXISTS);

        Affinity affinity = new Affinity(resourceAffinity);

        ResourceAffinity ra = affinity.getResourceAffinity();
        Selector pa = ra.getPreferredAffinity();
        Condition condition = pa.getCondition();
        SubCondition subCondition = condition.getSubConditions().get(0);
        LabelExpression expression = subCondition.getExpressions().get(0);
        Assert.assertEquals(labelKey, expression.getKey());
    }

    @Test
    public void testInitAffinityInstance() {

        String labelKey = "green";
        InstanceAffinity instanceAffinity = new InstanceAffinity(labelKey, OperatorType.LABEL_EXISTS);

        Affinity affinity = new Affinity(instanceAffinity);

        InstanceAffinity ra = affinity.getInstanceAffinity();
        Selector pa = ra.getPreferredAffinity();
        Condition condition = pa.getCondition();
        SubCondition subCondition = condition.getSubConditions().get(0);
        LabelExpression expression = subCondition.getExpressions().get(0);
        Assert.assertEquals(labelKey, expression.getKey());
    }

    @Test
    public void testGetAffinityValue() {
        Affinity affinity = new Affinity(AffinityKind.RESOURCE, AffinityType.PREFERRED, new ArrayList<>());
        Assert.assertNotEquals(0, affinity.getAffinityValue());

        Affinity affinity1 = new Affinity(null, null, new ArrayList<>());
        Assert.assertEquals(0, affinity1.getAffinityValue());
    }

    @Test
    public void testGetAffinityScopeValue() {
        Affinity affinity = new Affinity(AffinityKind.RESOURCE, AffinityType.PREFERRED, new ArrayList<>());
        Assert.assertEquals(0, affinity.getAffinityScopeValue());

        Affinity affinity1 = new Affinity(AffinityKind.INSTANCE, AffinityType.PREFERRED_ANTI, new ArrayList<>(),
            AffinityScope.NODE);
        Assert.assertEquals(2, affinity1.getAffinityScopeValue());
    }
}
