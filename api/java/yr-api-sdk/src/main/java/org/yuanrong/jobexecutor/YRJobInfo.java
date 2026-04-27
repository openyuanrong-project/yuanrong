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

import lombok.AccessLevel;
import lombok.Setter;

/**
 * The type YRJobInfo
 *
 * @since 2023 /06/06
 */
@Setter(value = AccessLevel.PACKAGE)
public class YRJobInfo {
    private String jobName;

    private String userJobID;

    private String jobStartTime;

    private String jobEndTime;

    private RuntimeEnv runtimeEnv;

    private YRJobStatus status;

    private String errorMessage;

    /**
     * The copy constructor of type YRJobInfo.
     *
     * @param yrJobInfo the YRJobInfo object.
     */
    public YRJobInfo(YRJobInfo yrJobInfo) {
        this.userJobID = yrJobInfo.userJobID;
        this.jobName = yrJobInfo.jobName;
        this.jobStartTime = yrJobInfo.jobStartTime;
        this.jobEndTime = yrJobInfo.jobEndTime;
        if (yrJobInfo.runtimeEnv != null) {
            this.runtimeEnv = new RuntimeEnv(yrJobInfo.runtimeEnv);
        }
        this.status = yrJobInfo.status;
        this.errorMessage = yrJobInfo.errorMessage;
    }

    /**
     * The constructor of type YRJobInfo.
     */
    public YRJobInfo() {}

    /**
     * Gets the user-defined job name set in YRJobParam.
     *
     * @return the user-defined job name String.
     */
    public String getJobName() {
        return jobName;
    }

    /**
     * Gets the job instanceID return by YRJobExecutor.sumitJob().
     *
     * @return unique userJobID String.
     */
    public String getUserJobID() {
        return userJobID;
    }

    /**
     * Gets the start time of remote attaced-runtime process of formatt
     * "yyyy-MM-dd'T'HH:mm:ss.SSSSS"
     *
     * @return the start time String.
     */
    public String getJobStartTime() {
        return jobStartTime;
    }

    /**
     * Gets the end time of remote attaced-runtime process of formatt
     * "yyyy-MM-dd'T'HH:mm:ss.SSSSS".
     * The attaced-runtime process may be stopped by the user or finished the job
     * successfully/unsuccessfully.
     *
     * @return the end time String.
     */
    public String getJobEndTime() {
        return jobEndTime;
    }

    /**
     * Gets the environment to be installed for the execution of user jobs.
     *
     * @return RuntimeEnv object.
     */
    public RuntimeEnv getRuntimeEnv() {
        return runtimeEnv;
    }

    /**
     * Gets the current job's status.
     *
     * @return YRJobStatus enum type.
     */
    public YRJobStatus getStatus() {
        return status;
    }

    /**
     * Get the error message raised by the attached-runtime.
     * The error message defaults to an empty string "".
     *
     * @return the error message String.
     */
    public String getErrorMessage() {
        return errorMessage;
    }

    /**
     * Whether the job is in final states including STOPPED/SUCCEEDED/FAILED.
     *
     * @return a boolean. It is true when the job is in final state.
     */
    public boolean ifFinalized() {
        return this.status != YRJobStatus.RUNNING && this.status != null;
    }

    /**
     * Updates members' values from another YRJobInfo object's NOT null values.
     *
     * @param source another YRJobInfo object.
     */
    public void update(YRJobInfo source) {
        this.userJobID = updateNotNull(this.userJobID, source.userJobID);
        this.jobName = updateNotNull(this.jobName, source.jobName);
        this.jobStartTime = updateNotNull(this.jobStartTime, source.jobStartTime);
        this.jobEndTime = updateNotNull(this.jobEndTime, source.jobEndTime);
        this.runtimeEnv = updateNotNull(this.runtimeEnv, source.runtimeEnv);
        this.status = updateNotNull(this.status, source.status);
        this.errorMessage = updateNotNull(this.errorMessage, source.errorMessage);
    }

    private <T> T updateNotNull(T target, T source) {
        if (source != null) {
            return source;
        }
        return target;
    }
}
