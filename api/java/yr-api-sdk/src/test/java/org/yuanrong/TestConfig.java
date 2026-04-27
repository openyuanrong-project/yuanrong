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

import org.yuanrong.exception.YRException;
import org.yuanrong.runtime.util.Constants;

import org.junit.Assert;
import org.junit.Test;

import java.util.ArrayList;
import java.util.HashMap;

public class TestConfig {
    @Test
    public void testInitConfig() {
        Config testConf1 = new Config(
            "sn:cn:yrk:default:function:0-crossyrlib-helloworld:$latest",
            "127.0.0.0",
            "127.0.0.0",
            "sn:cn:yrk:default:function:0-test-hello:$latest",
            true,
            false);
        Config testConf2 = new Config(
            "sn:cn:yrk:default:function:0-crossyrlib-helloworld:$latest",
            "127.0.0.0",
            1,
            "127.0.0.0",
            1,
            "sn:cn:yrk:default:function:0-test-hello:$latest",
            true);
        Config testConf3 = new Config(
            "sn:cn:yrk:default:function:0-crossyrlib-helloworld:$latest",
            "127.0.0.0",
            "127.0.0.0",
            "sn:cn:yrk:default:function:0-test-hello:$latest",
            "test-go-urn",
            true,
            false);
        Config testConf4 = new Config(
            "sn:cn:yrk:default:function:0-crossyrlib-helloworld:$latest",
            "127.0.0.0",
            1,
            "127.0.0.0",
            1,
            "sn:cn:yrk:default:function:0-test-hello:$latest",
            "sn:cn:yrk:default:function:0-test-hello:$latest",
            false);
        Config testConf5 = new Config.Builder().iamAuthToken("test-token")
            .cppFunctionURN("sn:cn:yrk:default:function:0-test-hello:$latest")
            .goFunctionURN("sn:cn:yrk:default:function:0-test-hello:$latest")
            .isDriver(true)
            .ns("test-ns")
            .logDir("/tmp")
            .logLevel("test-level")
            .enableDisConvCallStack(false)
            .rpcTimeout(1)
            .codePath(new ArrayList<>())
            .build();
        Config testConf6 = new Config();

        testConf1.setMaxLogSizeMb(10);
        testConf2.setMaxLogFileNum(10);
        testConf5.toString();
        testConf4.hashCode();
        Assert.assertFalse(testConf3.equals(testConf4));
        Assert.assertFalse(testConf1.equals(testConf2));

        Assert.assertTrue(testConf6.isInCluster());
        Assert.assertTrue(testConf6.isDriver());
        Assert.assertTrue(testConf6.isEnableMetrics());
        Assert.assertFalse(testConf6.isEnableMTLS());
        Assert.assertFalse(testConf6.isEnableDsAuth());
        Assert.assertTrue(testConf6.isEnableDisConvCallStack());
        Assert.assertFalse(testConf6.isThreadLocal());
        Assert.assertFalse(testConf6.isEnableSetContext());
        Assert.assertEquals("", testConf6.getServerAddress());
        Assert.assertEquals("", testConf6.getDataSystemAddress());
        Assert.assertEquals("", testConf6.getNs());
        Assert.assertEquals("", testConf6.getLogLevel());
        Assert.assertEquals("", testConf6.getIamAuthToken());
        Assert.assertEquals("", testConf6.getTenantId());
        Assert.assertEquals("sn:cn:yrk:default:function:0-defaultservice-cpp:$latest",
            testConf6.getCppFunctionURN());
        Assert.assertEquals("sn:cn:yrk:default:function:0-defaultservice-go:$latest",
            testConf6.getGoFunctionURN());
        Assert.assertEquals(30 * 60, testConf6.getRpcTimeout());
        Assert.assertEquals(31222, testConf6.getServerAddressPort());
        Assert.assertEquals(31222, testConf6.getDataSystemAddressPort());
        Assert.assertEquals(0, testConf6.getThreadPoolSize());
        Assert.assertEquals(10, testConf6.getRecycleTime());
        Assert.assertEquals(System.getProperty("user.dir"), testConf6.getLogDir());
        Assert.assertEquals(0, testConf6.getMaxLogSizeMb());
        Assert.assertEquals(0, testConf6.getMaxLogFileNum());
        Assert.assertEquals(-1, testConf6.getMaxTaskInstanceNum());
        Assert.assertEquals(100, testConf6.getMaxConcurrencyCreateNum());
        Assert.assertEquals(Constants.DEFAULT_HTTP_IO_THREAD_CNT, testConf6.getHttpIocThreadsNum());
        Assert.assertEquals(Constants.DEFAULT_HTTP_IDLE_TIME, testConf6.getHttpIdleTime());
        Assert.assertEquals(new ArrayList<>(), testConf6.getLoadPaths());
        Assert.assertEquals(new HashMap<>(), testConf6.getCustomEnvs());
    }

    @Test
    public void testCheckParameter() {
        Config testConf = new Config(
            "sn:cn:yrk:default:function:0-crossyrlib-helloworld:$latest",
            "127.0.0.0",
            "127.0.0.0",
            "sn:cn:yrk:default:function:0-test-hello:$latest",
            "test-go-urn",
            true,
            true);
        boolean isException = false;

        testConf.setCppFunctionURN("test-cppFunction");
        try {
            testConf.checkParameter();
        } catch (YRException e) {
            Assert.assertTrue(e.getMessage().contains("cppFunctionURN is invalid"));
            testConf.setCppFunctionURN(
                "sn:cn:yrk:default:function:0-crossyrlib-helloworld:$latest");
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        testConf.setGoFunctionURN("test-goFunction");
        try {
            testConf.checkParameter();
        } catch (YRException e) {
            Assert.assertTrue(e.getMessage().contains("goFunctionURN is invalid"));
            testConf.setGoFunctionURN(
                "sn:cn:yrk:default:function:0-crossyrlib-helloworld:$latest");
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        testConf.setServerAddress("test-server-address");
        try {
            testConf.checkParameter();
        } catch (YRException e) {
            Assert.assertTrue(e.getMessage().contains("serverAddress is invalid"));
            testConf.setServerAddress("127.0.0.1");
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        testConf.setDataSystemAddress("test-data-address");
        try {
            testConf.checkParameter();
        } catch (YRException e) {
            Assert.assertTrue(e.getMessage().contains("dataSystemAddress is invalid"));
            testConf.setDataSystemAddress("127.0.0.1");
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        testConf.setFunctionURN("test-function-urn");
        try {
            testConf.checkParameter();
        } catch (YRException e) {
            Assert.assertTrue(e.getMessage().contains("functionURN is invalid"));
            testConf.setFunctionURN(
                "sn:cn:yrk:default:function:0-crossyrlib-helloworld:$latest");
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        testConf.setRecycleTime(0);
        try {
            testConf.checkParameter();
        } catch (YRException e) {
            Assert.assertTrue(e.getMessage().contains("recycleTime is invalid"));
            isException = true;
        }
        Assert.assertTrue(isException);
    }
}
