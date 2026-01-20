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

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.exception.YRException;
import org.yuanrong.runtime.util.Constants;

import lombok.Getter;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Arrays;

/**
 * The environment to be installed in remote runtime, used by the
 * attached-runtime and runtimes started by it.
 *
 * @since 2023 /06/06
 */
@Getter
public class RuntimeEnv {
    private static final Logger LOGGER = LoggerFactory.getLogger(RuntimeEnv.class);

    /**
     * The string indicates "pip3.9", typicallly used to construct a pip installation
     * command.
     */
    private static final String PIP = "pip3.9";

    /**
     * The string indicates "pip3.9 check", typicallly used to construct a pip check
     * command after installation.
     */
    private static final String PIP_CHECK = "&& " + RuntimeEnv.PIP + " check";

    private String packageManager = PIP;

    /**
     * The packages to be installed in runtime. It should be initialized as an empty
     * String array rather than null when calling YR.invoke.
     */
    private String[] packages = new String[0];

    private boolean shouldPipCheck = false;

    private String trustedSource = "";

    /**
     * The default constructor of type RuntimeEnv.
     */
    public RuntimeEnv() {}

    /**
     * The copy constructor of type RuntimeEnv.
     *
     * @param runtimeEnv the RuntimeEnv object.
     */
    public RuntimeEnv(RuntimeEnv runtimeEnv) {
        this.packageManager = runtimeEnv.packageManager;
        this.packages = Arrays.copyOf(runtimeEnv.packages, runtimeEnv.packages.length);
        this.shouldPipCheck = runtimeEnv.shouldPipCheck;
    }

    /**
     * The package manager for environment installation, such as "pip3.9".
     *
     * @param packageManager the package manager String.
     */
    public void setPackageManager(String packageManager) {
        this.packageManager = packageManager;
    }

    /**
     * Sets packages to be installed. Example of the value:
     * <p>{"numpy", "pandas"}</p>
     *
     * @param packages the packages Strings.
     */
    public void setPackages(String... packages) {
        this.packages = packages;
    }

    /**
     * Gets packages to be installed. Example of the value:
     * <p>{"numpy", "pandas"}</p>
     *
     * @return the packages to be installed in remote runtime.
     */
    public String[] getPackages() {
        return Arrays.copyOf(this.packages, this.packages.length);
    }

    /**
     * Sets pipCheck. Pip check would be enable after the pip installation is
     * complete if true is set. The default value is false.
     *
     * @param shouldPipCheck the boolean indicates whether to perform pip check.
     */
    public void setShouldPipCheck(boolean shouldPipCheck) {
        this.shouldPipCheck = shouldPipCheck;
    }

    /**
     * The trusted source to be used. NOTE that an empty string is also allowed.
     *
     * @param trustedSource the trusted source of the format "--trusted-host HOST_NAME -i ADDRESS".
     */
    public void setTrustedSource(String trustedSource) {
        this.trustedSource = trustedSource;
    }

    /**
     * Converts environment infomation to a completed packages installation command.
     *
     * @return the completed packages installation command.
     * @throws YRException the actor task exception.
     */
    public String toCommand() throws YRException {
        if (this.packageManager == null || this.packageManager.isEmpty()) {
            LOGGER.warn("(JobExecutor) PackageManager is empty, no environment would be installed in remote runtime.");
            return "";
        }

        if (this.packages.length == 0) {
            LOGGER.warn("(JobExecutor) Packages array is empty, no environment would be installed in remote runtime.");
            return "";
        }

        String command = "";
        if (PIP.equals(this.packageManager)) {
            String packs = String.join(Constants.SPACE, this.packages);
            command = String.join(Constants.SPACE, this.packageManager, "install", packs, this.trustedSource);
        } else {
            throw new YRException(ErrorCode.ERR_JOB_USER_CODE_EXCEPTION, ModuleCode.RUNTIME,
                    "Only pip3.9 is supported currently.");
        }

        if (this.shouldPipCheck) {
            command = String.join(Constants.SPACE, command, RuntimeEnv.PIP_CHECK);
        }

        return command;
    }
};