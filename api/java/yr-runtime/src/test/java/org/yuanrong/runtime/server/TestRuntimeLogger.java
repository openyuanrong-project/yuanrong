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
package org.yuanrong.runtime.server;

import org.junit.Assert;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.mockito.Mockito;
import org.powermock.api.mockito.PowerMockito;
import org.powermock.core.classloader.annotations.PowerMockIgnore;
import org.powermock.core.classloader.annotations.PrepareForTest;
import org.powermock.modules.junit4.PowerMockRunner;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;

import java.io.File;

@RunWith(PowerMockRunner.class)
@PrepareForTest({LogManager.class})
@PowerMockIgnore("javax.management.*")
public class TestRuntimeLogger {
    private static final String EXCEPTION_LOG_PATH = "exception";
    private static final String ENV_LOG_LEVEL = "logLevel";
    private static final String ENV_JAVA_LOG_PATH = "logPath";
    private static final String ENV_EXCEPTION_LOG_DIR = "java.io.tmpdir";

    private Logger mockLogger;

    @Before
    public void setup() {
        mockLogger = Mockito.mock(Logger.class);
        PowerMockito.mockStatic(LogManager.class);
        Mockito.when(mockLogger.getName()).thenReturn("TestRuntimeLogger");
        Mockito.when(mockLogger.isInfoEnabled()).thenReturn(true);
        Mockito.when(mockLogger.isDebugEnabled()).thenReturn(true);
        Mockito.when(LogManager.getLogger(RuntimeLogger.class)).thenReturn(mockLogger);
    }

    /**
     * Description：
     *   Confirm that the code and message from 'Create' method are correct.
     * Steps:
     *   1. Sets property 'logLevel' with 'test_log_level'.
     *   2. Runs 'initLogger' method.
     * Expectation:
     *   1. Logs should contain constents which have been set in property.
     * @throws Exception
     */
    @Test
    public void testInitLogger() {
        String runtimeID = "test_runtime_id";
        String logDir ="/home/snuser/log/";
        String logLevel = "test_log_level";

        System.setProperty(ENV_LOG_LEVEL, logLevel);

        RuntimeLogger.initLogger(runtimeID);
        RuntimeLogger logger = new RuntimeLogger();
        boolean isException = false;
        try {
            RuntimeServer.main(null);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            RuntimeServer.main(new String[] {"length1"});
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            RuntimeServer.main(new String[] {runtimeID, "args2"});
        } catch (Throwable e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        Mockito.verify(mockLogger, Mockito.times(1)).info("runtime ID {}", runtimeID);
        Mockito.verify(mockLogger, Mockito.times(2)).debug("current log path = {}",
                String.format("%s", logDir));
        Mockito.verify(mockLogger, Mockito.times(2)).debug("current log level = {}", logLevel);

        Assert.assertNotNull(System.getProperty(ENV_LOG_LEVEL));
        Assert.assertEquals(logLevel, System.getProperty(ENV_LOG_LEVEL));

        Assert.assertNotNull(System.getProperty(ENV_JAVA_LOG_PATH));

        Assert.assertNotNull(System.getProperty(ENV_EXCEPTION_LOG_DIR));
        Assert.assertEquals(String.format("%s%s%s", logDir, File.separator, EXCEPTION_LOG_PATH),
                System.getProperty(ENV_EXCEPTION_LOG_DIR));
    }
}