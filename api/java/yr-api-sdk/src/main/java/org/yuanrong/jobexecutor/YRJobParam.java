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

import org.yuanrong.InvokeOptions;
import org.yuanrong.affinity.Affinity;
import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.exception.YRException;
import org.yuanrong.runtime.client.ObjectRef;
import org.yuanrong.runtime.util.Constants;

import com.google.gson.Gson;

import lombok.Getter;
import lombok.Setter;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * The type YRJobParam used as the parameter for YR.submitJob
 *
 * @since 2023 /06/06
 */
@Getter
@Setter
public class YRJobParam {
    private static final Logger LOGGER = LoggerFactory.getLogger(YRJobParam.class);

    /**
     * The index of code path in user defined entryPoint.
     */
    private static final int CODE_PATH_INDEX = 1;

    /**
     * The Minimum length of user defined entryPoint command after seperated by
     * space.
     */
    private static final int ENTRYPOINT_MIN_LENGTH = 2;

    /**
     * The default CPU value for the JobExecutor runtime.
     */
    private static final int DEFAULT_CPU_NUM = 500;

    /**
     * The default memory value for the JobExecutor runtime.
     */
    private static final int DEFAULT_MEM_NUM = 500;

    /**
     * The default concurrency value for the JobExecutor runtime.
     */
    private static final String DEFAULT_CONCURRENCY = "100";

    /**
     * The maximum value of memory for the JobExecutor runtime.
     */
    private static final int MAX_MEMORY = 65536;

    /**
     * The minimum value of memory for the JobExecutor runtime.
     */
    private static final int MIN_MEMORY = 128;

    /**
     * The maximum value of CPU for the JobExecutor runtime.
     */
    private static final int MAX_CPU = 16000;

    /**
     * The minimum value of CPU for the JobExecutor runtime.
     */
    private static final int MIN_CPU = 300;

    /**
     * The user-defined job name.
     */
    private String jobName;
    private ArrayList<String> entryPoint;
    private RuntimeEnv runtimeEnv = new RuntimeEnv();
    private OBSoptions obsOptions;
    private String localCodePath;
    private int cpu = DEFAULT_CPU_NUM;
    private int memory = DEFAULT_MEM_NUM;
    private List<Affinity> scheduleAffinities = new ArrayList<Affinity>();
    private boolean preferredPriority = true;
    private boolean requiredPriority = false;
    private ObjectRef entryPointObjRef;
    private String entryPointFileName;

    /**
     * The Custom resources."nvidia.com/gpu"
     */
    private Map<String, Float> customResources = new HashMap<>();

    /**
     * The Custom extensions. concurrency, a int in the range of [1, 1000]
     */
    private Map<String, String> customExtensions = new HashMap<>();

    /**
     * entryPoint is an ArrayList represents the command for starting a subprocess
     * (attached-runtime) running the driver code. For example:
     * <p>
     * {"python", "my_script.py"}
     * </p>
     *
     * @param entryPoint the command ArrayList.
     * @throws YRException the actor task exception.
     */
    public void setEntryPoint(ArrayList<String> entryPoint) throws YRException {
        if (entryPoint.size() < ENTRYPOINT_MIN_LENGTH) {
            throw new YRException(ErrorCode.ERR_JOB_USER_CODE_EXCEPTION, ModuleCode.RUNTIME,
                    "Length of the entryPoint(=" + entryPoint.size() + ") is not valid, it should be >= "
                            + ENTRYPOINT_MIN_LENGTH);
        }

        this.entryPoint = entryPoint;
    }

    /**
     * Sets the CPU value for the JobExecutor runtime. It is in 1/1024 cpu core, 300
     * to 16000 supported
     *
     * @param cpu the CPU value.
     * @throws YRException the actor task exception.
     */
    public void setCpu(int cpu) throws YRException {
        if (cpu < MIN_CPU || cpu > MAX_CPU) {
            throw new YRException(ErrorCode.ERR_JOB_USER_CODE_EXCEPTION, ModuleCode.RUNTIME,
                    "The CPU value(" + cpu + ") is not in [" + MIN_CPU + ", " + MAX_CPU + "]");
        }
        this.cpu = cpu;
    }

    /**
     * Sets the memory value for the JobExecutor runtime. It is in 1MB, 128 to 65536
     * supported.
     *
     * @param memory the memory value.
     * @throws YRException the actor task exception.
     */
    public void setMemory(int memory) throws YRException {
        if (memory < MIN_MEMORY || memory > MAX_MEMORY) {
            throw new YRException(ErrorCode.ERR_JOB_USER_CODE_EXCEPTION, ModuleCode.RUNTIME,
                    "The memory value(" + memory + ") is not in [" + MIN_MEMORY + ", " + MAX_MEMORY + "]");
        }
        this.memory = memory;
    }

    /**
     * Sets the value of runtimeEnv, which is the environment to be installed for
     * the execution of user jobs.
     *
     * @param packageManager the package manager. For example: "pip"
     * @param packages       packages to be installed. For example: {"numpy",
     *                       "pandas"}
     */
    public void setRuntimeEnv(String packageManager, String[] packages) {
        this.runtimeEnv.setPackageManager(packageManager);
        this.runtimeEnv.setPackages(packages);
    }

    /**
     * Sets the value of runtimeEnv, which is the environment to be installed for
     * the execution of user jobs.
     *
     * @param runtimeEnv the RuntimeEnv object contains environment infomation.
     */
    public void setRuntimeEnv(RuntimeEnv runtimeEnv) {
        this.runtimeEnv = runtimeEnv;
    }

    /**
     * The options for downloading user driver code from OBS.
     * The setting of obsOptions is optional, but either obsOptions or
     * localCodePath must be set and only one can be set.
     *
     * @param obsOptions the OBSoptions object.
     */
    public void setObsOptions(OBSoptions obsOptions) {
        if (this.localCodePath != null && !this.localCodePath.isEmpty()) {
            LOGGER.warn("(JobExecutor) localCodePath has been set, OBS setting will not work.");
            return;
        }
        this.obsOptions = obsOptions;
    }

    /**
     * Adds the given Affinity object to the list of schedule affinities.
     *
     * @param affinity the Affinity object to be added.
     */
    public void addScheduleAffinity(Affinity affinity) {
        this.scheduleAffinities.add(affinity);
    }

    /**
     * Adds the given CustomExtensions.
     *
     * @param key the key of CustomExtensions.
     * @param value the value of customExtensions.
     */
    public void addCustomExtensions(String key, String value) {
        this.customExtensions.put(key, value);
    }

    /**
     * Sets the preferred priority.
     *
     * @param isPreferred the boolean value indicating the preferred priority.
     */
    public void preferredPriority(boolean isPreferred) {
        this.preferredPriority = isPreferred;
    }

    /**
     * Sets the required priority.
     *
     * @param isRequired the boolean value indicating the required priority.
     */
    public void requiredPriority(boolean isRequired) {
        this.requiredPriority = isRequired;
    }

    /**
     * The path to driver code in remote runtime.
     * The setting of localCodePath is optional, but either obsOptions or
     * localCodePath must be set and only one can be set.
     *
     * @param path the path to driver code in remote runtime.
     * @throws YRException the actor task exception.
     */
    public void setLocalCodePath(String path) throws YRException {
        if (this.obsOptions != null) {
            LOGGER.warn("(JobExecutor) obsOptions has been set, localCodePath setting will not work.");
            return;
        }
        if (path.isEmpty()) {
            throw new YRException(ErrorCode.ERR_JOB_USER_CODE_EXCEPTION, ModuleCode.RUNTIME,
                    "localCodePath cannot be an empty String.");
        }

        this.localCodePath = path;
    }

    /**
     * The localEntryPoint is converted from the entryPoint, which is to be
     * performed by remote JobExecutor runtime. It provides the path in the
     * JobExecutor runtime of the driver code if localCodePath is set.
     *
     * @param ArrayList<String> the actual entryPoint executed in remote JobExecutor
     *                          actor runtime.
     * @return the localEntryPoint performed in remote JobExecutor runtime if
     *         localCodepath is set.
     * @throws YRException the actor task exception.
     */
    public ArrayList<String> getLocalEntryPoint() throws YRException {
        if (this.entryPoint == null) {
            throw new YRException(ErrorCode.ERR_JOB_USER_CODE_EXCEPTION, ModuleCode.RUNTIME,
                    "Failed to get entryPoint. EntryPoint should be set.");
        }
        if (this.localCodePath != null && !this.localCodePath.isEmpty()) {
            ArrayList<String> localEntryPoint = new ArrayList<>(this.entryPoint);
            localEntryPoint.set(CODE_PATH_INDEX,
                    String.join(Constants.BACKSLASH, this.localCodePath, this.entryPoint.get(CODE_PATH_INDEX)));
            return localEntryPoint;
        }
        return this.entryPoint;
    }

    /**
     * Generates an InvokeOptions for YR.instance invocation.
     *
     * @return InvokeOptions
     * @throws YRException the actor task exception.
     */
    public InvokeOptions extractInvokeOptions() throws YRException {
        // sets runtime env to invokeOptions
        String runtimeEnvStr = this.runtimeEnv.toCommand();
        if (!runtimeEnvStr.isEmpty()) {
            this.customExtensions.put(Constants.POST_START_EXEC, runtimeEnvStr);
        }
        // sets default Concurrency
        if (!this.customExtensions.containsKey(Constants.CONCURRENCY)) {
            this.customExtensions.put(Constants.CONCURRENCY, DEFAULT_CONCURRENCY);
        }
        // sets OBS options to invokeOptions
        if (this.obsOptions == null && this.localCodePath == null) {
            throw new YRException(ErrorCode.ERR_JOB_USER_CODE_EXCEPTION, ModuleCode.RUNTIME,
                    "Either YRJobParam.obsOptions or YRJobParam.localCodePath should be set. Job name: "
                    + this.jobName);
        }
        if (this.obsOptions != null) {
            this.customExtensions.putAll(obsOptions.toMap());
        }

        InvokeOptions option = InvokeOptions.builder()
                .cpu(this.cpu)
                .memory(this.memory)
                .customResources(this.customResources)
                .customExtensions(this.customExtensions)
                .preferredPriority(this.preferredPriority)
                .requiredPriority(this.requiredPriority)
                .scheduleAffinity(this.scheduleAffinities)
                .build();

        LOGGER.debug("Succeeded to extract InvokeOption from YRJobparam: {}", new Gson().toJson(option));
        return option;
    }

    /**
     * The YRJobParam Builder
     *
     * @return Builder for YRJobParam
     */
    public static Builder builder() {
        return new Builder();
    }

    /**
     * Builder for YRJobParam
     *
     * @since 2023 /06/06
     */
    public static class Builder {
        private YRJobParam param;

        /**
         * Builder
         */
        public Builder() {
            this.param = new YRJobParam();
        }

        /**
         * set job name
         *
         * @param jobName the job name
         * @return Builder
         */

        public Builder jobName(String jobName) {
            this.param.setJobName(jobName);
            return this;
        }

        /**
         * set entryPoint value
         *
         * @param entryPoint the command ArrayList.
         * @return Builder
         * @throws YRException the actor task exception.
         */
        public Builder entryPoint(ArrayList<String> entryPoint) throws YRException {
            this.param.setEntryPoint(entryPoint);
            return this;
        }

        /**
         * set runtimeEnv value
         *
         * @param runtimeEnv the runtime env
         * @return Builder
         */
        public Builder runtimeEnv(RuntimeEnv runtimeEnv) {
            this.param.setRuntimeEnv(runtimeEnv);
            return this;
        }

        /**
         * set obsOptions value
         *
         * @param obsOptions the OBSOptions object.
         * @return Builder
         */
        public Builder obsOptions(OBSoptions obsOptions) {
            this.param.setObsOptions(obsOptions);
            return this;
        }

        /**
         * set localCodePath value
         *
         * @param localCodePath the path to driver code in remote runtime.
         * @return Builder
         * @throws YRException the actor task exception.
         */
        public Builder localCodePath(String localCodePath) throws YRException {
            this.param.setLocalCodePath(localCodePath);
            return this;
        }

        /**
         * set cpu value
         *
         * @param cpu the cpu value
         * @return Builder
         * @throws YRException the actor task exception.
         */
        public Builder cpu(int cpu) throws YRException {
            this.param.setCpu(cpu);
            return this;
        }

        /**
         * set memory value
         *
         * @param memory the memory value.
         * @return Builder
         * @throws YRException the actor task exception.
         */
        public Builder memory(int memory) throws YRException {
            this.param.setMemory(memory);
            return this;
        }

        /**
         * set addScheduleAffinity value
         *
         * @param affinity the Affinity object to be added.
         * @return Builder
         */
        public Builder addScheduleAffinity(Affinity affinity) {
            this.param.scheduleAffinities.add(affinity);
            return this;
        }

        /**
         * set addScheduleAffinity value
         *
         * @param affinities the Affinity object to be added.
         * @return Builder
         */
        public Builder scheduleAffinity(List<Affinity> affinities) {
            this.param.scheduleAffinities = affinities;
            return this;
        }

        /**
         * set preferredPriority value
         *
         * @param isPreferred the boolean value indicating the preferred priority.
         * @return Builder
         */
        public Builder preferredPriority(boolean isPreferred) {
            this.param.preferredPriority = isPreferred;
            return this;
        }

        /**
         * set requiredPriority value
         *
         * @param isRequired the boolean value indicating the required priority.
         * @return Builder
         */
        public Builder requiredPriority(boolean isRequired) {
            this.param.requiredPriority = isRequired;
            return this;
        }

        /**
         * YRJobParam build
         *
         * @return YRJobParam
         */
        public YRJobParam build() {
            return this.param;
        }
    }
}
