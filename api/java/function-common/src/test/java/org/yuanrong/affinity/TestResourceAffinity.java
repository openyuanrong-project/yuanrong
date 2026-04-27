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

public class TestResourceAffinity {
    @Test
    public void testInitResourceAffinity() {
        ResourceAffinity resourceAffinity = new ResourceAffinity();
        ResourceAffinity resourceAffinity1 = new ResourceAffinity(new Selector());
        ResourceAffinity resourceAffinity2 = new ResourceAffinity(new LabelExpression());
        ResourceAffinity resourceAffinity3 = new ResourceAffinity(new ArrayList<>());
        ResourceAffinity resourceAffinity4 = new ResourceAffinity(AffinityType.REQUIRED, new Selector());
        ResourceAffinity resourceAffinity5 = new ResourceAffinity("testKey", OperatorType.LABEL_IN);
        ResourceAffinity resourceAffinity6 = new ResourceAffinity(new Selector(), new Selector(), new Selector(),
            new Selector());
        ResourceAffinity resourceAffinity7 = new ResourceAffinity(new ArrayList<>(), false);
        ResourceAffinity resourceAffinity8 = new ResourceAffinity("test", OperatorType.LABEL_NOT_IN, new ArrayList<>());
        ResourceAffinity resourceAffinity9 = new ResourceAffinity(AffinityType.PREFERRED, new Selector());
        ResourceAffinity resourceAffinity10 = new ResourceAffinity(AffinityType.PREFERRED_ANTI, new Selector());
        ResourceAffinity resourceAffinity11 = new ResourceAffinity(AffinityType.REQUIRED, new Selector());
        ResourceAffinity resourceAffinity12 = new ResourceAffinity(AffinityType.REQUIRED_ANTI, new Selector());

        resourceAffinity.setPreferredAffinity(new Selector());
        resourceAffinity.setRequiredAffinity(new Selector());
        resourceAffinity.setRequiredAntiAffinity(new Selector());
        resourceAffinity.setPreferredAntiAffinity(new Selector());

        resourceAffinity.getPreferredAffinity();
        resourceAffinity.getPreferredAntiAffinity();
        resourceAffinity.getRequiredAffinity();
        resourceAffinity.getRequiredAntiAffinity();
        resourceAffinity11.hashCode();
        resourceAffinity6.toString();
        resourceAffinity.canEqual(resourceAffinity2);

        Assert.assertFalse(resourceAffinity10.equals(resourceAffinity11));
    }

    @Test
    public void testInitResourceAffinityPreferredAffinity() {
        ResourceAffinity resourceAffinity = new ResourceAffinity(AffinityType.PREFERRED, new Selector());

        Assert.assertNotNull(resourceAffinity.getPreferredAffinity());

        Assert.assertNull(resourceAffinity.getPreferredAntiAffinity());
        Assert.assertNull(resourceAffinity.getRequiredAffinity());
        Assert.assertNull(resourceAffinity.getRequiredAntiAffinity());
    }

    @Test
    public void testInitResourceAffinityPreferredAntiAffinity() {
        ResourceAffinity resourceAffinity = new ResourceAffinity(AffinityType.PREFERRED_ANTI, new Selector());

        Assert.assertNotNull(resourceAffinity.getPreferredAntiAffinity());

        Assert.assertNull(resourceAffinity.getPreferredAffinity());
        Assert.assertNull(resourceAffinity.getRequiredAffinity());
        Assert.assertNull(resourceAffinity.getRequiredAntiAffinity());
    }

    @Test
    public void testInitResourceAffinityRequiredAffinity() {
        ResourceAffinity resourceAffinity = new ResourceAffinity(AffinityType.REQUIRED, new Selector());

        Assert.assertNotNull(resourceAffinity.getRequiredAffinity());

        Assert.assertNull(resourceAffinity.getPreferredAffinity());
        Assert.assertNull(resourceAffinity.getPreferredAntiAffinity());
        Assert.assertNull(resourceAffinity.getRequiredAntiAffinity());
    }

    @Test
    public void testInitResourceAffinityRequiredAntiAffinity() {
        ResourceAffinity resourceAffinity = new ResourceAffinity(AffinityType.REQUIRED_ANTI, new Selector());

        Assert.assertNotNull(resourceAffinity.getRequiredAntiAffinity());

        Assert.assertNull(resourceAffinity.getPreferredAffinity());
        Assert.assertNull(resourceAffinity.getPreferredAntiAffinity());
        Assert.assertNull(resourceAffinity.getRequiredAffinity());
    }

    @Test
    public void testInitResourceAffinitySelector() {
        Selector pa = new Selector();
        ResourceAffinity resourceAffinity = new ResourceAffinity(pa);

        Assert.assertEquals(pa, resourceAffinity.getPreferredAffinity());

        Assert.assertNull(resourceAffinity.getPreferredAntiAffinity());
        Assert.assertNull(resourceAffinity.getRequiredAffinity());
        Assert.assertNull(resourceAffinity.getRequiredAntiAffinity());
    }

    @Test
    public void testInitResourceAffinitySelectors() {
        Selector pa = new Selector();
        Selector paa = new Selector();
        Selector ra = new Selector();
        Selector raa = new Selector();
        ResourceAffinity resourceAffinity = new ResourceAffinity(pa, paa, ra, raa);

        Assert.assertEquals(pa, resourceAffinity.getPreferredAffinity());
        Assert.assertEquals(paa, resourceAffinity.getPreferredAntiAffinity());
        Assert.assertEquals(ra, resourceAffinity.getRequiredAffinity());
        Assert.assertEquals(raa, resourceAffinity.getRequiredAntiAffinity());
    }
}
