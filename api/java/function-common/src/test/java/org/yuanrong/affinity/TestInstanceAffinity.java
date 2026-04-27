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

public class TestInstanceAffinity {
    @Test
    public void testInitInstanceAffinity() {
        InstanceAffinity instanceAffinity = new InstanceAffinity();
        InstanceAffinity instanceAffinity1 = new InstanceAffinity(new LabelExpression("testKey", new LabelOperator()));
        InstanceAffinity instanceAffinity2 = new InstanceAffinity(new Selector());
        InstanceAffinity instanceAffinity3 = new InstanceAffinity(new ArrayList<>());
        InstanceAffinity instanceAffinity4 = new InstanceAffinity("testInstance", OperatorType.LABEL_IN);
        InstanceAffinity instanceAffinity5 = new InstanceAffinity(new Selector(), AffinityScope.POD);
        InstanceAffinity instanceAffinity6 = new InstanceAffinity(new ArrayList<>(), false);
        InstanceAffinity instanceAffinity7 = new InstanceAffinity(AffinityType.REQUIRED, new Selector(),
            AffinityScope.NODE);
        InstanceAffinity instanceAffinity8 = new InstanceAffinity(new Selector(),
            new Selector(new Condition("testCondition", OperatorType.LABEL_EXISTS)), new Selector(), new Selector());
        InstanceAffinity instanceAffinity9 = new InstanceAffinity(new Selector(),
            new Selector(new Condition("testCondition", OperatorType.LABEL_EXISTS)), new Selector(), new Selector(),
            AffinityScope.NODE);
        InstanceAffinity instanceAffinity10 = new InstanceAffinity("testAffinity", OperatorType.LABEL_NOT_IN,
            new ArrayList<>());
        InstanceAffinity instanceAffinity11 = new InstanceAffinity(AffinityType.PREFERRED, new Selector(),
            AffinityScope.NODE);
        InstanceAffinity instanceAffinity12 = new InstanceAffinity(AffinityType.PREFERRED_ANTI, new Selector(),
            AffinityScope.NODE);
        InstanceAffinity instanceAffinity13 = new InstanceAffinity(AffinityType.REQUIRED, new Selector(),
            AffinityScope.NODE);
        InstanceAffinity instanceAffinity14 = new InstanceAffinity(AffinityType.REQUIRED_ANTI, new Selector(),
            AffinityScope.NODE);
        instanceAffinity.setScope(AffinityScope.POD);
        instanceAffinity.setPreferredAffinity(new Selector());
        instanceAffinity.setPreferredAntiAffinity(new Selector());
        instanceAffinity.setRequiredAffinity(new Selector());
        instanceAffinity.setRequiredAntiAffinity(new Selector());

        instanceAffinity.getPreferredAffinity();
        instanceAffinity.getRequiredAffinity();
        instanceAffinity.getPreferredAntiAffinity();
        instanceAffinity.getRequiredAntiAffinity();
        instanceAffinity.hashCode();
        instanceAffinity6.toString();
        instanceAffinity.canEqual(instanceAffinity2);

        Assert.assertEquals(2, instanceAffinity11.getScope().getScope());
        Assert.assertFalse(instanceAffinity3.equals(instanceAffinity8));
        Assert.assertEquals(AffinityScope.POD, instanceAffinity.getScope());
    }

    @Test
    public void testInitInstanceAffinityPreferredAffinity() {
        InstanceAffinity instanceAffinity = new InstanceAffinity(AffinityType.PREFERRED, new Selector(),
                AffinityScope.NODE);

        Assert.assertNotNull(instanceAffinity.getPreferredAffinity());

        Assert.assertNull(instanceAffinity.getPreferredAntiAffinity());
        Assert.assertNull(instanceAffinity.getRequiredAffinity());
        Assert.assertNull(instanceAffinity.getRequiredAntiAffinity());

        Assert.assertEquals(AffinityScope.NODE, instanceAffinity.getScope());
    }

    @Test
    public void testInitInstanceAffinityPreferredAntiAffinity() {
        InstanceAffinity instanceAffinity = new InstanceAffinity(AffinityType.PREFERRED_ANTI, new Selector(),
                AffinityScope.NODE);

        Assert.assertNotNull(instanceAffinity.getPreferredAntiAffinity());

        Assert.assertNull(instanceAffinity.getPreferredAffinity());
        Assert.assertNull(instanceAffinity.getRequiredAffinity());
        Assert.assertNull(instanceAffinity.getRequiredAntiAffinity());

        Assert.assertEquals(AffinityScope.NODE, instanceAffinity.getScope());
    }

    @Test
    public void testInitInstanceAffinityRequiredAffinity() {
        InstanceAffinity instanceAffinity = new InstanceAffinity(AffinityType.REQUIRED, new Selector(),
                AffinityScope.NODE);

        Assert.assertNotNull(instanceAffinity.getRequiredAffinity());

        Assert.assertNull(instanceAffinity.getPreferredAffinity());
        Assert.assertNull(instanceAffinity.getPreferredAntiAffinity());
        Assert.assertNull(instanceAffinity.getRequiredAntiAffinity());

        Assert.assertEquals(AffinityScope.NODE, instanceAffinity.getScope());
    }

    @Test
    public void testInitInstanceAffinityRequiredAntiAffinity() {
        InstanceAffinity instanceAffinity = new InstanceAffinity(AffinityType.REQUIRED_ANTI, new Selector(),
                AffinityScope.NODE);

        Assert.assertNotNull(instanceAffinity.getRequiredAntiAffinity());

        Assert.assertNull(instanceAffinity.getPreferredAffinity());
        Assert.assertNull(instanceAffinity.getPreferredAntiAffinity());
        Assert.assertNull(instanceAffinity.getRequiredAffinity());

        Assert.assertEquals(AffinityScope.NODE, instanceAffinity.getScope());
    }

    @Test
    public void testInitInstanceAffinitySelector() {
        Selector pa = new Selector();
        InstanceAffinity instanceAffinity = new InstanceAffinity(pa);

        Assert.assertEquals(pa, instanceAffinity.getPreferredAffinity());

        Assert.assertNull(instanceAffinity.getPreferredAntiAffinity());
        Assert.assertNull(instanceAffinity.getRequiredAffinity());
        Assert.assertNull(instanceAffinity.getRequiredAntiAffinity());

        Assert.assertEquals(AffinityScope.NODE, instanceAffinity.getScope());
    }

    @Test
    public void testInitInstanceAffinitySelectorScope() {
        Selector pa = new Selector();
        InstanceAffinity instanceAffinity = new InstanceAffinity(pa, AffinityScope.POD);

        Assert.assertEquals(pa, instanceAffinity.getPreferredAffinity());

        Assert.assertNull(instanceAffinity.getPreferredAntiAffinity());
        Assert.assertNull(instanceAffinity.getRequiredAffinity());
        Assert.assertNull(instanceAffinity.getRequiredAntiAffinity());

        Assert.assertEquals(AffinityScope.POD, instanceAffinity.getScope());
    }

    @Test
    public void testInitInstanceAffinitySelectors() {
        Selector pa = new Selector();
        Selector paa = new Selector();
        Selector ra = new Selector();
        Selector raa = new Selector();
        InstanceAffinity instanceAffinity = new InstanceAffinity(pa, paa, ra, raa);
        Assert.assertEquals(pa, instanceAffinity.getPreferredAffinity());
        Assert.assertEquals(paa, instanceAffinity.getPreferredAntiAffinity());
        Assert.assertEquals(ra, instanceAffinity.getRequiredAffinity());
        Assert.assertEquals(raa, instanceAffinity.getRequiredAntiAffinity());

        Assert.assertEquals(AffinityScope.NODE, instanceAffinity.getScope());
    }

    @Test
    public void testInitInstanceAffinitySelectorsScope() {
        Selector pa = new Selector();
        Selector paa = new Selector();
        Selector ra = new Selector();
        Selector raa = new Selector();
        InstanceAffinity instanceAffinity = new InstanceAffinity(pa, paa, ra, raa, AffinityScope.POD);
        Assert.assertEquals(pa, instanceAffinity.getPreferredAffinity());
        Assert.assertEquals(paa, instanceAffinity.getPreferredAntiAffinity());
        Assert.assertEquals(ra, instanceAffinity.getRequiredAffinity());
        Assert.assertEquals(raa, instanceAffinity.getRequiredAntiAffinity());

        Assert.assertEquals(AffinityScope.POD, instanceAffinity.getScope());
    }
}
