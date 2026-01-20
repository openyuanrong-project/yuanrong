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

import java.util.ArrayList;
import java.util.HashMap;

public class TestYRJobParam {
    @Test
    public void testInitYRJobParam() {
        boolean isException = false;
        YRJobParam yrJobParam = new YRJobParam();
        yrJobParam.setJobName("test-job");

        try {
            yrJobParam.setCpu(100);
        } catch (YRException e) {
            Assert.assertTrue(e.getMessage().contains("The CPU value"));
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            yrJobParam.setCpu(500);
        } catch (YRException e) {
            Assert.assertTrue(e.getMessage().contains("The CPU value"));
            isException = true;
        }
        Assert.assertFalse(isException);

        try {
            yrJobParam.setMemory(100);
        } catch (YRException e) {
            Assert.assertTrue(e.getMessage().contains("The memory value"));
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            yrJobParam.setMemory(500);
        } catch (YRException e) {
            Assert.assertTrue(e.getMessage().contains("The memory value"));
            isException = true;
        }
        Assert.assertFalse(isException);

        try {
            yrJobParam.setEntryPoint(new ArrayList<>());
        } catch (YRException e) {
            Assert.assertTrue(e.getMessage().contains("Length of the entryPoint"));
            isException = true;
        }
        Assert.assertTrue(isException);

        ArrayList<String> testEntryPoint = new ArrayList<>();
        testEntryPoint.add("test-entryPoint1");
        testEntryPoint.add("test-entryPoint2");
        testEntryPoint.add("test-entryPoint3");
        testEntryPoint.add("test-entryPoint4");
        isException = false;
        try {
            yrJobParam.setEntryPoint(testEntryPoint);
        } catch (YRException e) {
            Assert.assertTrue(e.getMessage().contains("Length of the entryPoint"));
            isException = true;
        }
        Assert.assertFalse(isException);

        yrJobParam.setRuntimeEnv(new RuntimeEnv());
        yrJobParam.setRuntimeEnv("test-packageManager", new String[] {"test-package"});
        yrJobParam.setObsOptions(null);
        yrJobParam.addScheduleAffinity(null);
        yrJobParam.setCustomExtensions(new HashMap<>());
        yrJobParam.setCustomResources(new HashMap<>());
        yrJobParam.setScheduleAffinities(new ArrayList<>());
        yrJobParam.preferredPriority(true);

        try {
            yrJobParam.setLocalCodePath("");
        } catch (YRException e) {
            Assert.assertTrue(e.getMessage().contains("localCodePath cannot be an empty String."));
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            yrJobParam.setLocalCodePath("/tmp");
        } catch (YRException e) {
            Assert.assertTrue(e.getMessage().contains("localCodePath cannot be an empty String."));
            isException = true;
        }
        Assert.assertFalse(isException);

        isException = false;
        try {
            OBSoptions obSoptions = new OBSoptions();
            obSoptions.setAk("test-Ak");
            obSoptions.setSk("test-SK");
            yrJobParam.setObsOptions(obSoptions);
            yrJobParam.setLocalCodePath("test");
        } catch (YRException e) {
            Assert.assertTrue(e.getMessage().contains("localCodePath cannot be an empty String."));
            isException = true;
        }
        Assert.assertFalse(isException);

        yrJobParam.setObsOptions(new OBSoptions());
        yrJobParam.setPreferredPriority(true);
        yrJobParam.requiredPriority(true);
        yrJobParam.getJobName();
        yrJobParam.getCpu();
        yrJobParam.getMemory();
        yrJobParam.getEntryPoint();
        yrJobParam.getCustomResources();
        yrJobParam.getCustomExtensions();
        yrJobParam.getLocalCodePath();
        yrJobParam.getObsOptions();
        yrJobParam.getScheduleAffinities();
        yrJobParam.getObsOptions();
        yrJobParam.getRuntimeEnv();
        Assert.assertTrue(yrJobParam.isPreferredPriority());
    }

    @Test
    public void testBuilder() {
        YRJobParam testYRJobparam = YRJobParam.builder().build();
        Assert.assertTrue(testYRJobparam.isPreferredPriority());

        ArrayList<String> testEntryPoint = new ArrayList<>();
        testEntryPoint.add("test-entryPoint1");
        testEntryPoint.add("test-entryPoint2");
        testEntryPoint.add("test-entryPoint3");
        testEntryPoint.add("test-entryPoint4");

        boolean isException = false;
        try {
            YRJobParam yrJobParam = new YRJobParam.Builder().cpu(500)
                .memory(500)
                .entryPoint(testEntryPoint)
                .localCodePath("/tmp")
                .jobName("test-job")
                .addScheduleAffinity(null)
                .obsOptions(new OBSoptions())
                .runtimeEnv(new RuntimeEnv())
                .preferredPriority(true)
                .scheduleAffinity(new ArrayList<>())
                .requiredPriority(true)
                .build();
            Assert.assertTrue(yrJobParam.isPreferredPriority());
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testGetLocalEntryPoint() {
        YRJobParam yrJobParam = new YRJobParam();
        boolean isException = false;

        try {
            yrJobParam.getLocalEntryPoint();
        } catch (YRException e) {
            Assert.assertTrue(e.getMessage().contains("Failed to get entryPoint. EntryPoint should be set."));
            isException = true;
        }
        Assert.assertTrue(isException);

        ArrayList<String> testEntryPoint = new ArrayList<>();
        testEntryPoint.add("test-entryPoint1");
        testEntryPoint.add("test-entryPoint2");
        testEntryPoint.add("test-entryPoint3");
        testEntryPoint.add("test-entryPoint4");
        isException = false;

        try {
            yrJobParam.setEntryPoint(testEntryPoint);
            yrJobParam.getLocalEntryPoint();
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        try {
            yrJobParam.setLocalCodePath("/tmp");
            yrJobParam.getLocalEntryPoint();
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testExtractInvokeOptions() {
        YRJobParam yrJobParam = new YRJobParam();

        RuntimeEnv runtimeEnv = new RuntimeEnv();
        runtimeEnv.setPackages("test-packages");
        runtimeEnv.setPackageManager("pip3.9");

        OBSoptions obSoptions = new OBSoptions();
        obSoptions.setAk("test-Ak");
        obSoptions.setSk("test-SK");
        obSoptions.setEndPoint("test-EndPoint");
        obSoptions.setBucketID("test-ID");
        obSoptions.setObjectID("test-ID");
        obSoptions.setSecurityToken("test-token");

        boolean isException = false;

        try {
            yrJobParam.setRuntimeEnv(runtimeEnv);
            yrJobParam.extractInvokeOptions();
        } catch (YRException e) {
            Assert.assertTrue(
                e.getMessage().contains("Either YRJobParam.obsOptions or YRJobParam.localCodePath should be set"));
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;

        try {
            yrJobParam.setObsOptions(obSoptions);
            yrJobParam.extractInvokeOptions();
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

}
