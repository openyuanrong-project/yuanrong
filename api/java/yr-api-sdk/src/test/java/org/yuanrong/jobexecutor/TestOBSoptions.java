/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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

package org.yuanrong.jobexecutor;

import org.yuanrong.exception.YRException;

import org.junit.Assert;
import org.junit.Test;

public class TestOBSoptions {
    @Test
    public void testInitOptions() {
        OBSoptions obSoptions = new OBSoptions();
        String expectedValue = "test-ID";
        obSoptions.setAk("test-Ak");
        obSoptions.setSk("test-SK");
        obSoptions.setEndPoint("test-EndPoint");
        obSoptions.setBucketID(expectedValue);
        obSoptions.setObjectID(expectedValue);
        obSoptions.setSecurityToken("test-token");

        obSoptions.getAk();
        obSoptions.getSk();
        obSoptions.getEndPoint();
        obSoptions.getSecurityToken();

        Assert.assertEquals(expectedValue, obSoptions.getBucketID());
        Assert.assertEquals(expectedValue, obSoptions.getObjectID());
    }

    @Test
    public void testToMap() {

        OBSoptions obSoptions = new OBSoptions();
        obSoptions.setEndPoint("https://example.com");
        obSoptions.setBucketID("test-bucket");
        obSoptions.setObjectID("test-object");
        obSoptions.setSecurityToken("test-token");
        obSoptions.setAk("test-ak");
        obSoptions.setSk("test-sk");

        boolean isException = false;

        try {
            obSoptions.toMap();
        } catch (YRException e) {
            Assert.assertTrue(e.getMessage().contains("Failed to convert OBS options to a json string"));
            isException = true;
        }
        Assert.assertFalse(isException);
    }
}
