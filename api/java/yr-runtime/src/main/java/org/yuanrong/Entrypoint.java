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

package org.yuanrong;

import org.yuanrong.codemanager.CodeExecutor;
import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.Pair;
import org.yuanrong.jni.LibRuntime;
import org.yuanrong.jni.LibRuntimeConfig;
import org.yuanrong.runtime.config.RuntimeContext;
import org.yuanrong.runtime.util.Utils;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Arrays;
import java.util.HashMap;
import java.util.Map;

/**
 * java
 * -Dlog4j.configurationFile=./java/yr-jni/src/main/resources/log4j2.xml
 * -Djava.library.path=$(pwd)/build/output/execroot/yuanrong_multi_language_runtime/
 * bazel-out/k8-opt/bin/api/java/yr-jni/src/main/cpp/
 * -cp
 * ./build/output/execroot/yuanrong_multi_language_runtime/bazel-out/k8-opt/bin/api/java/yr_deploy.jar
 * org.yuanrong.Entrypoint
 *
 * @since 2023/10/24
 */
public class Entrypoint {
    private static final Logger LOG = LoggerFactory.getLogger(Entrypoint.class);

    /*
     * INSTANCE_ID, DRIVER_SERVER_PORT, HOST_IP, YR_FUNCTION_LIB_PATH, LAYER_LIB_PATH, LD_LIBRARY_PATH is unused.
     * see functionsystem/src/runtime_manager/config/build.cpp for details
     */

    private static final String ENV_POSIX_LISTEN_ADDR = "POSIX_LISTEN_ADDR";

    private static final String ENV_DATASYSTEM_ADDR = "DATASYSTEM_ADDR";

    private static final String ENV_HOME = "HOME";

    private static final String ENV_LOG_LEVEL = "logLevel";

    private static final String ENV_LOG_ID = "logId";

    private static final String ENV_RUNTIME_ID = "runtimeId";

    private static final String ENV_GLOG_DIR = "GLOG_log_dir";

    private static final String ENV_JOB_ID = "jobId";

    private static final int DFAULT_RECYCLE_TIME = 2;

    private static Map<String, String> defaultEnvs = new HashMap<String, String>() {
        {
            put(ENV_HOME, "/home/snuser");
            put(ENV_LOG_LEVEL, "INFO");
            put(ENV_GLOG_DIR, "/home/snuser/log");
            put(ENV_JOB_ID, Utils.generateCloudJobId());
            put(ENV_LOG_ID, "");
            put(ENV_RUNTIME_ID, "");
        }
    };

    /**
     * the entrypoint of runtime
     *
     * @param args : commandline args
     */
    public static void runtimeEntrypoint(String[] args) {
        LOG.debug("input args: {}", Arrays.toString(args));
        Pair<String, Integer> addr = splitIPAndPortFromAddr(args[0]);
        String runtimeId = args[1];
        RuntimeContext.runtimeId = runtimeId;
        LOG.info("service addresses: ip=({}), port=({}), runtimeId=({})", addr.first, addr.second, runtimeId);

        ErrorInfo errorInfo = LibRuntime.Init(makeLibRuntimeConfig(addr, runtimeId));
        if (!ErrorCode.ERR_OK.equals(errorInfo.getErrorCode())) {
            LOG.error("Failed to init LibRuntime, Code:{}, MCode:{}, Msg:{}", errorInfo.getErrorCode(),
                    errorInfo.getModuleCode(), errorInfo.getErrorMessage());
            return;
        }
        CodeExecutor.loadHandler();
        LibRuntime.ReceiveRequestLoop();
        LibRuntime.Finalize();
    }

    /**
     * split ip and port from an colon-connected IP and port address string
     *
     * @param addr : colon-connected IP and port address string, like
     *             "0.0.0.0:12345"
     * @return a pair <"0.0.0.0", 12345>
     */
    private static Pair<String, Integer> splitIPAndPortFromAddr(String addr) {
        String[] parts = addr.split(":");
        if (parts.length != 2) {
            String errMsg = "failed to split addr (" + addr
                    + ") to ip and port, expect to split into 2 parts with \":\", but actually ("
                    + String.valueOf(parts.length) + ")";
            LOG.error(errMsg);
            throw new RuntimeException(errMsg);
        }
        return new Pair<String, Integer>(parts[0], Integer.parseInt(parts[1]));
    }

    /**
     * By default, get from env, is missing, get from jvm property, if both failed,
     * log error and return empty string
     *
     * @param key : the env key
     * @return the env val, or empty string if failed
     */
    private static String safeGetEnv(String key) {
        String envVal = System.getenv(key);
        if (envVal != null && !envVal.isEmpty()) {
            return envVal;
        }
        envVal = System.getProperty(key);
        if (envVal != null && !envVal.isEmpty()) {
            return envVal;
        }
        String res = defaultEnvs.get(key);
        if (res != null) {
            return res;
        }
        String errMsg = "env/property (" + key + ") missed or empty.";
        LOG.error(errMsg);
        throw new RuntimeException(errMsg);
    }

    /**
     * Make LibRuntimeConfig according to command line args, env, and some default
     * values.
     *
     * @param addr      : the runtime server address <ip, port>
     * @param runtimeId : the runtimeID
     * @return the LibRuntimeConfig
     */
    private static LibRuntimeConfig makeLibRuntimeConfig(Pair<String, Integer> addr, String runtimeId) {
        LibRuntimeConfig config = new LibRuntimeConfig();

        // from commandline
        config.setFunctionSystemRtServerIpAddr(addr.first);
        config.setFunctionSystemRtServerPort(addr.second);
        config.setRuntimeId(runtimeId);

        // addresses from ENV
        Pair<String, Integer> posixAddr = splitIPAndPortFromAddr(safeGetEnv(ENV_POSIX_LISTEN_ADDR));
        Pair<String, Integer> dsAddr = splitIPAndPortFromAddr(safeGetEnv(ENV_DATASYSTEM_ADDR));

        config.setFunctionSystemIpAddr(posixAddr.first);
        config.setFunctionSystemPort(posixAddr.second);
        config.setDataSystemIpAddr(dsAddr.first);
        config.setDataSystemPort(dsAddr.second);

        // other configs from ENV. 'ENV_LOG_LEVEL' has been set by 'initLogger'
        config.setLogLevel(safeGetEnv(ENV_LOG_LEVEL));
        config.setLogDir(safeGetEnv(ENV_GLOG_DIR));
        config.setJobId(safeGetEnv(ENV_JOB_ID));
        config.setRecycleTime(DFAULT_RECYCLE_TIME);

        // use default value
        config.setDriver(false);
        config.setEnableMetrics(false);

        // others use default value in libruntime
        return config;
    }
}
