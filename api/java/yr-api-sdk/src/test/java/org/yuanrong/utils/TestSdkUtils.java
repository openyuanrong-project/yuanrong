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

package org.yuanrong.utils;

import org.yuanrong.Config;

import org.junit.Assert;
import org.junit.Test;

public class TestSdkUtils {
    @Test
    public void testInitSdkUtils() {
        SdkUtils sdkUtils = new SdkUtils();
        SdkUtils.getObjectRefId();
        SdkUtils.checkInvokeAccurateExceptionLocation(
            new StackTraceElement[] {new StackTraceElement("test", "testMethod", "testFile", 1)});
        SdkUtils.getTenantContext(new Config("testURN", "testAddr", "testData", "testCPPURN", true), "testID");
        Assert.assertNotNull(sdkUtils);
    }
}
