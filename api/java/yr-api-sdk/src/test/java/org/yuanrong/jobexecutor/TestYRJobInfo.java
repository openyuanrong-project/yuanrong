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

import org.junit.Assert;
import org.junit.Test;

public class TestYRJobInfo {
    @Test
    public void testInitYRJobInfo() {
        YRJobInfo yrJobInfo = new YRJobInfo();
        yrJobInfo.setRuntimeEnv(new RuntimeEnv());
        YRJobInfo testJobInfo = new YRJobInfo(yrJobInfo);
        testJobInfo.setJobName("test-name");
        testJobInfo.setUserJobID("test-id");
        testJobInfo.setJobStartTime("2024-01-01");
        testJobInfo.setJobEndTime("2025-01-01");
        testJobInfo.setRuntimeEnv(new RuntimeEnv());
        testJobInfo.setStatus(YRJobStatus.SUCCEEDED);
        testJobInfo.setErrorMessage("failed");

        testJobInfo.getJobName();
        testJobInfo.getJobEndTime();
        testJobInfo.getUserJobID();
        testJobInfo.getStatus();
        testJobInfo.getErrorMessage();
        testJobInfo.getJobStartTime();
        testJobInfo.getRuntimeEnv();
        Assert.assertTrue(testJobInfo.ifFinalized());
    }

    @Test
    public void testUpdate() {
        YRJobInfo yrJobInfo = new YRJobInfo();
        YRJobInfo testJobInfo = new YRJobInfo();
        testJobInfo.setJobName("test-name");
        testJobInfo.setUserJobID("test-id");
        testJobInfo.setJobStartTime("2024-01-01");
        testJobInfo.setJobEndTime("2025-01-01");
        testJobInfo.setRuntimeEnv(new RuntimeEnv());
        testJobInfo.setStatus(YRJobStatus.SUCCEEDED);
        testJobInfo.setErrorMessage(null);
        yrJobInfo.update(testJobInfo);
        Assert.assertTrue(yrJobInfo.ifFinalized());
    }
}
