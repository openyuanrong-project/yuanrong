/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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

public class TestJobExecutor {
    @Test
    public void testGetJobInfo() throws Exception {
        RuntimeEnv runtimeEnv = new RuntimeEnv();
        runtimeEnv.setPackages("test-packages");
        runtimeEnv.setPackageManager("pip3.9");
        ArrayList<String> testEntryPoint = new ArrayList<>();
        testEntryPoint.add("python3.9");

        ArrayList<String> testEntryPointWithWrongValue = new ArrayList<>();
        testEntryPointWithWrongValue.add("point01");
        testEntryPointWithWrongValue.add("point02");

        boolean isException = false;
        JobExecutor jobExecutor;
        YRJobParam yrJobParam = new YRJobParam();
        yrJobParam.setJobName("test-job");
        yrJobParam.setRuntimeEnv(runtimeEnv);
        yrJobParam.setEntryPoint(testEntryPointWithWrongValue);

        try {
            jobExecutor = new JobExecutor(yrJobParam);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        yrJobParam.setEntryPoint(testEntryPoint);
        try {
            jobExecutor = new JobExecutor(yrJobParam);
            jobExecutor.getJobInfo(true);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        try {
            jobExecutor = new JobExecutor(yrJobParam);
            jobExecutor.stop();
            jobExecutor.getJobInfo(false);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }
}
