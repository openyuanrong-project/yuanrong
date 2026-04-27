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

package org.yuanrong.jni;

import org.junit.Assert;
import org.junit.Test;

public class TestLibRunTimeConfig {
    @Test
    public void testInitLibRunTimeConfig() {
        boolean expectedValue = true;
        LibRuntimeConfig libRuntimeConfig = new LibRuntimeConfig();
        libRuntimeConfig.setFunctionUrn("sn:cn:yrk:default:function:0-test-hello:$latest");
        libRuntimeConfig.setLogFlushInterval(1);
        libRuntimeConfig.setMetaConfig("test-meta");
        libRuntimeConfig.setEnableMTLS(expectedValue);
        libRuntimeConfig.setNs("test-ns");

        libRuntimeConfig.getJobId();
        libRuntimeConfig.isDriver();
        libRuntimeConfig.getRuntimeId();
        libRuntimeConfig.getFunctionUrn();
        libRuntimeConfig.getLogLevel();
        libRuntimeConfig.getLogFileNumMax();
        libRuntimeConfig.getLogFileSizeMax();
        libRuntimeConfig.getLogFlushInterval();
        libRuntimeConfig.getMetaConfig();
        libRuntimeConfig.getRecycleTime();
        libRuntimeConfig.getMaxTaskInstanceNum();
        libRuntimeConfig.isEnableMetrics();
        libRuntimeConfig.getThreadPoolSize();
        libRuntimeConfig.getFunctionIds();
        libRuntimeConfig.getLoadPaths();
        libRuntimeConfig.isInCluster();
        libRuntimeConfig.getRpcTimeout();
        libRuntimeConfig.getTenantId();
        libRuntimeConfig.getNs();
        libRuntimeConfig.getCustomEnvs();
        libRuntimeConfig.hashCode();
        libRuntimeConfig.toString();

        LibRuntimeConfig testLibRunTimeConfig = new LibRuntimeConfig();
        libRuntimeConfig.canEqual(testLibRunTimeConfig);
        Assert.assertFalse(libRuntimeConfig.equals(testLibRunTimeConfig));
        Assert.assertEquals(expectedValue, libRuntimeConfig.isEnableMTLS());

    }
}
