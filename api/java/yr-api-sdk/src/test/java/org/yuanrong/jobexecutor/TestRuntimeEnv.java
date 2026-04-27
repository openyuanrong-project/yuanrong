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

public class TestRuntimeEnv {
    @Test
    public void testInitRuntimeEnvWithExpectedValue() {
        boolean expectedValue = true;

        RuntimeEnv runtimeEnv = new RuntimeEnv();
        RuntimeEnv testRuntimeEnv = new RuntimeEnv(runtimeEnv);
        testRuntimeEnv.setPackages("test-package");
        testRuntimeEnv.setPackageManager("test-PIP");
        testRuntimeEnv.setShouldPipCheck(expectedValue);
        testRuntimeEnv.setTrustedSource("test-source");
        testRuntimeEnv.getPackages();
        testRuntimeEnv.getPackageManager();
        testRuntimeEnv.getTrustedSource();

        Assert.assertTrue(testRuntimeEnv.isShouldPipCheck());
    }

    @Test
    public void testRuntimeEnvToCommand() {
        RuntimeEnv runtimeEnv = new RuntimeEnv();
        boolean isException = false;

        runtimeEnv.setPackageManager("");
        try {
            runtimeEnv.toCommand();
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        runtimeEnv.setPackageManager(null);
        try {
            runtimeEnv.toCommand();
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
        runtimeEnv.setPackageManager("test-PIP");

        runtimeEnv.setPackages();
        try {
            runtimeEnv.toCommand();
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
        runtimeEnv.setPackages("test-packages");

        runtimeEnv.setPackageManager("pip3.9");
        try {
            runtimeEnv.toCommand();
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        runtimeEnv.setPackageManager("pip3.8");
        try {
            runtimeEnv.toCommand();
        } catch (YRException e) {
            Assert.assertTrue(e.getMessage().contains("Only pip3.9 is supported currently."));
            runtimeEnv.setPackageManager("pip3.9");
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        runtimeEnv.setShouldPipCheck(true);
        try {
            runtimeEnv.toCommand();
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

}
