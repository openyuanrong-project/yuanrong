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

package org.yuanrong.runtime.server;

import org.yuanrong.Entrypoint;
import org.yuanrong.utils.RuntimeUtils;

/**
 * There is no real server right now, but in order to adapt to the old
 * functionsystem code, we added this file to let the old start command
 * ```
 * java -Dxxx -Dxxx -cp /path/to/yr-runtime-1.0.0.jar
 * org.yuanrong.runtime.server.RuntimeServer
 * ```
 * to run properly
 *
 * @since 2024/4/25
 */
public class RuntimeServer {
    /**
     * The number of args should not be less than 2.
     */
    private static final int ARGS_NUM = 2;

    public static void main(String[] args) {
        if (args == null) {
            throw new IllegalArgumentException("error: init parameters null");
        }

        if (args.length < ARGS_NUM) {
            String errorMessage = String.format("The number of args is %d, expecting %d args <address> <runtimeId>",
                args.length, ARGS_NUM);
            throw new IllegalArgumentException("error: " + errorMessage);
        }

        String runtimeId = args[1];
        RuntimeLogger.initLogger(runtimeId);

        // If YR_SEED_FILE is set, read the specified file to block.
        RuntimeUtils.readSeedFile();

        // Get and set the environment variables from the env file of `YR_ENV_FILE`.
        // Since Java cannot set global environment variables, write into the global properties.
        // If System.getEnv fails, then execute System.getProperty.
        RuntimeUtils.loadEnvFromFile(System.getenv("YR_ENV_FILE"));

        Entrypoint.runtimeEntrypoint(args);
    }
}
