/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
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

package org.yuanrong;

import org.junit.Assert;
import org.junit.Test;

import java.util.ArrayList;
import java.util.HashMap;

public class TestConfigManager {
    @Test
    public void testInitConfigManager() {

        Config config = new Config(
            "sn:cn:yrk:default:function:0-crossyrlib-helloworld:$latest",
            "127.0.0.0",
            1,
            "127.0.0.0",
            1,
            "sn:cn:yrk:default:function:0-test-hello:$latest",
            "sn:cn:yrk:default:function:0-test-hello:$latest",
            true);

        ConfigManager configManager = new ConfigManager();
        configManager.init(config);
        configManager.setDriverServerIP("127.0.0.0");
        boolean expectedValue = true;
        configManager.setDriver(expectedValue);
        configManager.setCodePath(new ArrayList<>());
        configManager.setEnableDisConvCallStack(expectedValue);
        configManager.getDriverServerIP();
        configManager.getCodePath();

        ConfigManager testConfigManager = new ConfigManager();
        testConfigManager.setInitialized(expectedValue);
        testConfigManager.setFunctionURN("sn:cn:yrk:default:function:0-test-hello:$latest");
        testConfigManager.setServerAddress("127.0.0.0");
        testConfigManager.setServerAddressPort(1);
        testConfigManager.setDataSystemAddress("127.0.0.0");
        testConfigManager.setDataSystemAddressPort(22);
        testConfigManager.setJobId("test-id");
        testConfigManager.setRuntimeId("test-runtime");
        testConfigManager.setCppFunctionURN("sn:cn:yrk:default:function:0-test-hello:$latest");
        testConfigManager.setGoFunctionURN("sn:cn:yrk:default:function:0-test-hello:$latest");
        testConfigManager.setRecycleTime(1);
        testConfigManager.setInitialized(expectedValue);
        testConfigManager.setMaxTaskInstanceNum(10);
        testConfigManager.setEnableMetrics(expectedValue);
        testConfigManager.setMaxConcurrencyCreateNum(10);
        testConfigManager.setLogLevel("test-level");
        testConfigManager.setLogDir("/tmp");
        testConfigManager.setNs("test-ns");
        testConfigManager.setLogFileNumMax(10);
        testConfigManager.setLogFileSizeMax(1024);
        testConfigManager.setThreadPoolSize(1);
        testConfigManager.setLoadPaths(new ArrayList<>());
        testConfigManager.setHttpIocThreadsNum(1);
        testConfigManager.setHttpIdleTime(60);
        testConfigManager.setRpcTimeout(10);
        testConfigManager.setTenantId("test-tenantID");
        testConfigManager.setCustomEnvs(new HashMap<>());
        testConfigManager.canEqual(configManager);
        testConfigManager.setInCluster(expectedValue);
        testConfigManager.toString();
        testConfigManager.hashCode();

        Assert.assertFalse(configManager.equals(testConfigManager));
        Assert.assertEquals(expectedValue, configManager.isDriver());
        Assert.assertEquals(expectedValue, configManager.isEnableDisConvCallStack());
    }
}
