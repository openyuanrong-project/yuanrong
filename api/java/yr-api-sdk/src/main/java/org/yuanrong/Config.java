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

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.exception.YRException;
import org.yuanrong.runtime.util.Constants;
import org.yuanrong.utils.SdkUtils;
import org.yuanrong.jni.YRAutoInitInfo;
import org.yuanrong.runtime.util.Utils;

import lombok.Data;

import java.util.List;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.Map;

/**
 * The Config class is the initialization data structure of Yuanrong, used to store basic information such as
 * IP, port, and URN needed when initializing the Yuanrong system. The Config instance is the input parameter of the
 * init interface. Except for functionURN, serverAddress, dataSystemAddress, cppFunctionURN, and isInCluster (the top
 * five in the table), which are mandatory configurations and supported through the constructor, the rest of the
 * parameters are default and set through setters. For specific interfaces, please refer to the table at the end.
 *
 * @since 2023/10/16
 */
@Data
public class Config {
    private static final int DEFAULT_RECYCLETIME = 10;
    private static final int DEFAULT_SERVER_PORT = 31222;
    private static final int DEFAULT_DS_PORT = 31222;
    private static final int DEFAULT_TIMEOUT = 30 * 60;
    private static final String DEFAULT_LOG_DIR = System.getProperty("user.dir");
    private static final String DEFAULT_FUNC_URN
        = "sn:cn:yrk:default:function:0-defaultservice-java:$latest";
    private static final String DEFAULT_CPP_URN
        = "sn:cn:yrk:default:function:0-defaultservice-cpp:$latest";
    private static final String DEFAULT_GO_URN
        = "sn:cn:yrk:default:function:0-defaultservice-go:$latest";

    /**
     * Specify the so path. If not specified, it is specified by services.yaml.
     */
    private ArrayList<String> loadPaths = new ArrayList<>();

    /**
     * The functionURN returned by the deployment function.
     * It can be set through the YRFUNCID environment variable. The current functionURN generation logic is
     * ``sn:cn:yrk:default:function:0-{ServiceName}-{FunctionName}:$latest``.
     */
    private String functionURN = Utils.getEnvWithDefualtValue("YRFUNCID", DEFAULT_FUNC_URN, "config-");

    /**
     * The functionID returned by the deployment cpp function.
     */
    private String cppFunctionURN = DEFAULT_CPP_URN;

    /**
     * The functionID returned by the deployment go function.
     */
    private String goFunctionURN = DEFAULT_GO_URN;

    /**
     * Cluster IP (Yuanrong cluster master node).
     */
    private String serverAddress = "";

    /**
     * Data system IP (Yuanrong cluster master node).
     */
    private String dataSystemAddress = "";

    /**
     * rpc timeout.
     */
    private int rpcTimeout = DEFAULT_TIMEOUT;

    /**
     * Cluster port number, default ``31222``.
     */
    private int serverAddressPort = DEFAULT_SERVER_PORT;

    /**
     * DataSystem port number, default ``31222``.
     */
    private int dataSystemAddressPort = DEFAULT_DS_PORT;

    /**
     * When a user uses a named stateful function, `ns` is the namespace of the stateful function. The default value
     * is an empty string. If `ns` is not an empty string, the stateful function name is ``namespace-name``.
     */
    private String ns = "";

    /**
     * Thread pool size, used only in ParallelFor.
     */
    private int threadPoolSize = 0;

    /**
     * Inside/Outside the cluster, determine whether the runtime is connected to the bus, default is ``false``;
     * clients outside the cluster have no retry mechanism.
     */
    private boolean isInCluster = true;

    /**
     * Is driver or not, determines whether the API is used on the cloud or locally, default is ``true``.
     */
    private boolean isDriver = true;

    /**
     * Recycling time, default is ``10``, value must be greater than 0.
     */
    private int recycleTime = DEFAULT_RECYCLETIME;

    /**
     * Whether to enable metric collection and reporting. The default value is ``false``.
     */
    private boolean enableMetrics = true;

    /**
     * Whether to enable two-way authentication for external cloud clients, default is ``off``.
     */
    private boolean enableMTLS = false;

    /**
     * Client certificate file path.
     */
    private String certificateFilePath;

    /**
     * Client private key file path.
     */
    private String privateKeyPath;

    /**
     * Server certificate file path.
     */
    private String verifyFilePath;

    /**
     * Client private key encryption password.
     */
    private String privateKeyPaaswd;

    /**
     * Server name.
     */
    private String serverName;

    /**
     * Whether to enable data system authentication, default is ``false``.
     */
    private boolean enableDsAuth = false;

    /**
     * Input directory for client logs. Default is the current directory. If the directory does not exist, create the
     * directory and generate the log file in the directory.
     */
    private String logDir = DEFAULT_LOG_DIR;

    /**
     * Maximum size of individual client log files, in MB, default value is ``0`` (if the default value is ``0``,
     * it will eventually be set to ``40``). When the size limit is exceeded, the log file will be rolled over and
     * split. The system checks for rolling over every 30 seconds. A negative value will throw an exception.
     */
    private int maxLogSizeMb = 0;

    /**
     * The maximum number of files to be retained after client log rolling, default value ``0`` (if the default value
     * is ``0``, it will be set to ``20``), and the oldest logs will be deleted when the number exceeds the limit,
     * checked and deleted every 30 seconds.
     */
    private int maxLogFileNum = 0;

    /**
     * Log level, with values of DEBUG, INFO, WARN, and ERROR. If not set or set to an invalid value,
     * it will be set to the default value ``INFO``.
     */
    private String logLevel = "";

    /**
     * Limit the maximum number of instances that can be created for stateless functions. The valid range is 1 to 65536.
     * If not configured, the default value is ``-1``,
     * which means no limit. If an invalid value is entered, the Init interface will throw an exception.
     */
    private int maxTaskInstanceNum = -1;

    /**
     * Set the maximum number of concurrent stateless instance creations. The value must be greater than 0. The default
     * value is ``100``.
     */
    private int maxConcurrencyCreateNum = 100;

    /**
     * If true, enables distributed collection call stack, and exceptions thrown by user code will be
     * accurately located. Users can use the distributed collection call stack to locate exceptions and errors that
     * occur in the distributed call chain in a simpler way, just like in a single process.
     */
    private boolean enableDisConvCallStack = true;

    /**
     * When the user configures iamAuthToken, the token is included in the header of each request sent to the cluster.
     * The remote runtime will obtain the token ciphertext from RuntimeManager and then carry the token ciphertext
     * in the runtime communication request.
     */
    private String iamAuthToken = "";

    /**
     * Whether to enable multi-cluster mode. The default is ``false``. If isThreadLocal is true, call YR.init in
     * different threads and set different cluster addresses. The runtime Java SDK can connect to different clusters.
     */
    private boolean isThreadLocal = false;

    /**
     * Whether to enable tenant context. The default value is ``false``. If enableSetContext is true, tenant context
     * switching is supported.
     */
    private boolean enableSetContext = false;

    /**
     * Tenant ID, default is ``empty``.
     */
    private String tenantId = "";

    /**
     * Number of HTTP connection working threads. Default value: ``200``.
     */
    private int httpIocThreadsNum = Constants.DEFAULT_HTTP_IO_THREAD_CNT;

    /**
     * HTTP connection idle time, default value ``120``.
     */
    private int httpIdleTime = Constants.DEFAULT_HTTP_IDLE_TIME;

    /**
     * Used to set custom environment variables for the runtime. Currently, only LD_LIBRARY_PATH is supported.
     */
    private Map<String, String> customEnvs = new HashMap<>();

    /**
     * Code path.
     */
    private List<String> codePath;
    private boolean enableDsEncrypt = false;
    private String dsPublicKeyContextPath = "";
    private String runtimePublicKeyContextPath = "";
    private String runtimePrivateKeyContextPath = "";
    private boolean enableFrontendTLS;

    /**
     * The constructor of Config.
     *
     * @param functionURN The functionURN returned by the deployment function.
     * @param serverAddress Cluster IP (Yuanrong cluster master node).
     * @param dataSystemAddress Data system IP (Yuanrong cluster master node).
     * @param cppFunctionURN The functionID returned by the deployment cpp function.
     * @param isInCluster Inside/Outside the cluster.
     */
    public Config(String functionURN, String serverAddress, String dataSystemAddress, String cppFunctionURN,
            boolean isInCluster) {
        this(functionURN, serverAddress, DEFAULT_SERVER_PORT, dataSystemAddress, DEFAULT_DS_PORT, cppFunctionURN,
                isInCluster, true);
    }

    /**
     * The constructor of Config.
     */
    public Config() {}

    /**
     * The constructor of Config.
     *
     * @param functionURN The functionURN returned by the deployment function.
     * @param serverAddress Cluster IP (Yuanrong cluster master node).
     * @param dataSystemAddress Data system IP (Yuanrong cluster master node).
     * @param cppFunctionURN The functionID returned by the deployment cpp function.
     * @param goFunctionURN The functionID returned by the deployment go function.
     * @param isInCluster Inside/Outside the cluster.
     */
    public Config(String functionURN, String serverAddress, String dataSystemAddress, String cppFunctionURN,
                  String goFunctionURN, boolean isInCluster) {
        this(functionURN, serverAddress, DEFAULT_SERVER_PORT, dataSystemAddress, DEFAULT_DS_PORT, cppFunctionURN,
                isInCluster, true);
        this.setGoFunctionURN(goFunctionURN);
    }

    /**
     * The constructor of Config.
     *
     * @param functionURN The functionURN returned by the deployment function.
     * @param serverAddress Cluster IP (Yuanrong cluster master node).
     * @param dataSystemAddress Data system IP (Yuanrong cluster master node).
     * @param cppFunctionURN The functionID returned by the deployment cpp function.
     * @param isInCluster Inside/Outside the cluster.
     * @param isDriver On cloud or off cloud.
     */
    public Config(
            String functionURN,
            String serverAddress,
            String dataSystemAddress,
            String cppFunctionURN,
            boolean isInCluster, boolean isDriver) {
        this(functionURN, serverAddress, DEFAULT_SERVER_PORT, dataSystemAddress, DEFAULT_DS_PORT, cppFunctionURN,
                isInCluster, isDriver);
    }

    /**
     * The constructor of Config.
     *
     * @param functionURN The functionURN returned by the deployment function.
     * @param serverAddress Cluster IP (Yuanrong cluster master node).
     * @param dataSystemAddress Data system IP (Yuanrong cluster master node).
     * @param cppFunctionURN The functionID returned by the deployment cpp function.
     * @param goFunctionURN The functionID returned by the deployment go function.
     * @param isInCluster Inside/Outside the cluster.
     * @param isDriver On cloud or off cloud.
     */
    public Config(
            String functionURN,
            String serverAddress,
            String dataSystemAddress,
            String cppFunctionURN,
            String goFunctionURN,
            boolean isInCluster, boolean isDriver) {
        this(functionURN, serverAddress, DEFAULT_SERVER_PORT, dataSystemAddress, DEFAULT_DS_PORT, cppFunctionURN,
                isInCluster, isDriver);
        this.setGoFunctionURN(goFunctionURN);
    }

    /**
     * The constructor of Config.
     *
     * @param functionUrn The functionURN returned by the deployment function.
     * @param serverAddr Cluster IP (Yuanrong cluster master node).
     * @param serverAddressPort Cluster port number.
     * @param dataSystemAddress Data system IP (Yuanrong cluster master node).
     * @param dataSystemAddressPort DataSystem port number.
     * @param cppFunctionUrn The functionID returned by the deployment cpp function.
     * @param isInCluster Inside/Outside the cluster.
     */
    public Config(
            String functionUrn,
            String serverAddr,
            int serverAddressPort,
            String dataSystemAddr,
            int dataSystemAddressPort,
            String cppFunctionUrn,
            boolean isInCluster) {
        this(
                functionUrn,
                serverAddr,
                serverAddressPort,
                dataSystemAddr,
                dataSystemAddressPort,
                cppFunctionUrn,
                isInCluster,
                true);
    }

    /**
     * The constructor of Config.
     *
     * @param functionUrn The functionURN returned by the deployment function.
     * @param serverAddr Cluster IP (Yuanrong cluster master node).
     * @param serverAddressPort Cluster port number.
     * @param dataSystemAddress Data system IP (Yuanrong cluster master node).
     * @param dataSystemAddressPort DataSystem port number.
     * @param cppFunctionUrn The functionID returned by the deployment cpp function.
     * @param goFunctionUrn The functionID returned by the deployment go function.
     * @param isInCluster Inside/Outside the cluster.
     */
    public Config(
            String functionUrn,
            String serverAddr,
            int serverAddressPort,
            String dataSystemAddr,
            int dataSystemAddressPort,
            String cppFunctionUrn,
            String goFunctionUrn,
            boolean isInCluster) {
        this(
                functionUrn,
                serverAddr,
                serverAddressPort,
                dataSystemAddr,
                dataSystemAddressPort,
                cppFunctionUrn,
                isInCluster,
                true);
        this.setGoFunctionURN(goFunctionUrn);
    }

    /**
     * The constructor of Config.
     *
     * @param functionUrn The functionURN returned by the deployment function.
     * @param serverAddr Cluster IP (Yuanrong cluster master node).
     * @param serverAddressPort Cluster port number.
     * @param dataSystemAddress Data system IP (Yuanrong cluster master node).
     * @param dataSystemAddressPort DataSystem port number.
     * @param cppFunctionUrn The functionID returned by the deployment cpp function.
     * @param isInCluster Inside/Outside the cluster.
     * @param isDriver On cloud or off cloud.
     */
    public Config(
            String functionUrn,
            String serverAddr,
            int serverAddressPort,
            String dataSystemAddr,
            int dataSystemAddressPort,
            String cppFunctionUrn,
            boolean isInCluster,
            boolean isDriver) {
        this.functionURN = functionUrn;
        this.serverAddress = serverAddr;
        this.dataSystemAddress = dataSystemAddr;
        this.cppFunctionURN = cppFunctionUrn;
        this.isInCluster = isInCluster;
        this.isDriver = isDriver;
        this.serverAddressPort = serverAddressPort;
        this.dataSystemAddressPort = dataSystemAddressPort;
    }

    private Config(Builder builder) {
        this.functionURN = builder.functionURN;
        this.serverAddress = builder.serverAddress;
        this.dataSystemAddress = builder.dataSystemAddress;
        this.iamAuthToken = builder.iamAuthToken;
        this.cppFunctionURN = builder.cppFunctionURN;
        this.goFunctionURN = builder.goFunctionURN;
        this.ns = builder.ns;
        this.tenantId = builder.tenantId;
        this.isInCluster = builder.isInCluster;
        this.isDriver = builder.isDriver;
        this.serverAddressPort = builder.serverAddressPort;
        this.dataSystemAddressPort = builder.dataSystemAddressPort;
        this.isThreadLocal = builder.isThreadLocal;
        this.enableSetContext = builder.enableSetContext;
        this.logDir = builder.logDir;
        this.logLevel = builder.logLevel;
        this.enableDisConvCallStack = builder.enableDisConvCallStack;
        this.rpcTimeout = builder.rpcTimeout;
    }

    /**
     * set maxLogSizeMb.
     *
     * @param maxLogSizeMb maxLogSizeMb
     */
    public void setMaxLogSizeMb(int maxLogSizeMb) {
        SdkUtils.validateNonNegative(maxLogSizeMb, "maxLogSizeMb must be non-negative.");
        this.maxLogSizeMb = maxLogSizeMb;
    }

    /**
     * set maxLogFileNum.
     *
     * @param maxLogFileNum maxLogFileNum
     */
    public void setMaxLogFileNum(int maxLogFileNum) {
        SdkUtils.validateNonNegative(maxLogFileNum, "maxLogFileNum must be non-negative.");
        this.maxLogFileNum = maxLogFileNum;
    }

    /**
     * set threadPoolSize.
     *
     * @param threadPoolSize threadPoolSize
     */
    public void setThreadPoolSize(int threadPoolSize) {
        SdkUtils.validateNonNegative(threadPoolSize, "threadPoolSize must be non-negative.");
        this.threadPoolSize = threadPoolSize;
    }

    /**
     * set rpcTimeout.
     *
     * @param rpcTimeout rpcTimeout
     */
    public void setRpcTimeout(int rpcTimeout) {
        SdkUtils.validateNonNegative(rpcTimeout, "rpcTimeout must be non-negative.");
        this.rpcTimeout = rpcTimeout;
    }

    /**
     * Check parameter.
     *
     * @throws YRException the YR exception.
     */
    public void checkParameter() throws YRException {
        if (!cppFunctionURN.isEmpty() && !SdkUtils.checkURN(cppFunctionURN)) {
            throw new YRException(ErrorCode.ERR_INCORRECT_INIT_USAGE, ModuleCode.RUNTIME,
                    "cppFunctionURN is invalid");
        }
        if (!goFunctionURN.isEmpty() && !SdkUtils.checkURN(goFunctionURN)) {
            throw new YRException(ErrorCode.ERR_INCORRECT_INIT_USAGE, ModuleCode.RUNTIME,
                    "goFunctionURN is invalid");
        }
        if (!SdkUtils.checkIP(serverAddress)) {
            throw new YRException(ErrorCode.ERR_INCORRECT_INIT_USAGE, ModuleCode.RUNTIME,
                    "serverAddress is invalid");
        }
        if (!SdkUtils.checkIP(dataSystemAddress)) {
            throw new YRException(ErrorCode.ERR_INCORRECT_INIT_USAGE, ModuleCode.RUNTIME,
                    "dataSystemAddress is invalid");
        }
        if (!functionURN.isEmpty() && !SdkUtils.checkURN(functionURN)) {
            throw new YRException(ErrorCode.ERR_INCORRECT_INIT_USAGE, ModuleCode.RUNTIME,
                    "functionURN is invalid");
        }
        if (recycleTime <= 0) {
            throw new YRException(ErrorCode.ERR_INCORRECT_INIT_USAGE, ModuleCode.RUNTIME,
                    "recycleTime is invalid");
        }
    }

    /**
     * Returns the Builder object of Config class.
     *
     * @return Builder object of Config class.
     */
    public static Builder builder() {
        return new Builder();
    }

    /**
     * The Builder class of Config class.
     */
    public static class Builder {
        private String functionURN = DEFAULT_FUNC_URN;
        private String serverAddress = "";
        private String dataSystemAddress = "";
        private String iamAuthToken = "";
        private String cppFunctionURN = "";
        private String goFunctionURN = "";
        private String ns = "";
        private String logDir = DEFAULT_LOG_DIR;
        private String logLevel = "";
        private String tenantId = "";
        private boolean isInCluster = true;
        private boolean isDriver = true;
        private int serverAddressPort = DEFAULT_SERVER_PORT;
        private int dataSystemAddressPort = DEFAULT_DS_PORT;
        private boolean isThreadLocal = false;
        private boolean enableSetContext = false;
        private boolean enableDisConvCallStack = true;
        private int rpcTimeout = DEFAULT_TIMEOUT;
        private List<String> codePath;

        /**
         * Sets the functionURN of function of Config class.
         *
         * @param functionURN the functionURN String.
         * @return Config Builder class object.
         */
        public Builder functionURN(String functionURN) {
            this.functionURN = functionURN;
            return this;
        }

        /**
         * Sets the serverAddress of Config class.
         *
         * @param serverAddress the server IP String.
         * @return Config Builder class object.
         */
        public Builder serverAddress(String serverAddress) {
            this.serverAddress = serverAddress;
            return this;
        }

        /**
         * Sets the dataSystemAddress of Config class.
         *
         * @param dataSystemAddress the dataSystem IP String.
         * @return Config Builder class object.
         */
        public Builder dataSystemAddress(String dataSystemAddress) {
            this.dataSystemAddress = dataSystemAddress;
            return this;
        }

        /**
         * Sets the iamAuthToken of Config class.
         *
         * @param iamAuthToken the IAM token String.
         * @return Config Builder class object.
         */
        public Builder iamAuthToken(String iamAuthToken) {
            this.iamAuthToken = iamAuthToken;
            return this;
        }

        /**
         * Sets the cppFunctionURN string of Config class.
         *
         * @param cppFunctionURN the cppFunctionURN string.
         * @return Config Builder class object.
         */
        public Builder cppFunctionURN(String cppFunctionURN) {
            this.cppFunctionURN = cppFunctionURN;
            return this;
        }

        /**
         * Sets the goFunctionURN string of Config class.
         *
         * @param goFunctionURN the goFunctionURN string.
         * @return Config Builder class object.
         */
        public Builder goFunctionURN(String goFunctionURN) {
            this.goFunctionURN = goFunctionURN;
            return this;
        }

        /**
         * Sets the tenantId string of Config class.
         *
         * @param tenantId the tenantId string.
         * @return Config Builder class object.
         */
        public Builder tenantId(String tenantId) {
            this.tenantId = tenantId;
            return this;
        }

        /**
         * Sets the isInCluster boolean of Config class.
         *
         * @param isInCluster the isInCluster boolean. Applying driver mode if it is true.
         * @return Config Builder class object.
         */
        public Builder isInCluster(boolean isInCluster) {
            this.isInCluster = isInCluster;
            return this;
        }

        /**
         * Sets the isDriver boolean of Config class.
         *
         * @param isDriver the isDriver boolean.
         * @return Config Builder class object.
         */
        public Builder isDriver(boolean isDriver) {
            this.isDriver = isDriver;
            return this;
        }

        /**
         * Sets the serverAddressPort of Config class.
         *
         * @param serverAddressPort the server port int.
         * @return Config Builder class object.
         */
        public Builder serverAddressPort(int serverAddressPort) {
            this.serverAddressPort = serverAddressPort;
            return this;
        }

        /**
         * Sets the dataSystemAddressPort of Config class.
         *
         * @param dataSystemAddressPort the dataSystem port int.
         * @return Config Builder class object.
         */
        public Builder dataSystemAddressPort(int dataSystemAddressPort) {
            this.dataSystemAddressPort = dataSystemAddressPort;
            return this;
        }

        /**
         * If isThreadLocal is true, calling YR.init in different
         * threads and setting different addresses,
         * runtime-java sdk can connect to different clusters.
         *
         * @param isThreadLocal boolean indicates that each YR sdk is threadLocal.
         * @return Config Builder class object.
         */
        public Builder isThreadLocal(boolean isThreadLocal) {
            this.isThreadLocal = isThreadLocal;
            return this;
        }

        /**
         * If enableSetContext is true, tenant context switching is allowed,
         * including tenantid, cluster and so on.
         *
         * @param enableSetContext boolean indicates that whether support tenant context.
         * @return Config Builder class object.
         */
        public Builder enableSetContext(boolean enableSetContext) {
            this.enableSetContext = enableSetContext;
            return this;
        }

        /**
         * Sets the namespace for the builder.
         *
         * @param namespace the namespace to set.
         * @return Config Builder class object
         */
        public Builder ns(String namespace) {
            this.ns = namespace;
            return this;
        }

        /**
         * Sets the log file directory path for the builder.
         *
         * @param logDir the directory path to set.
         * @return Config Builder class object.
         */
        public Builder logDir(String logDir) {
            this.logDir = logDir;
            return this;
        }

        /**
         * Sets the log level for the builder.
         *
         * @param logLevel the log level to set.
         * @return Config Builder class object.
         */
        public Builder logLevel(String logLevel) {
            this.logLevel = logLevel;
            return this;
        }

        /**
         * Sets the enableDisConvCallStack for the builder.
         *
         * @param enableDisConvCallStack the enableDisConvCallStack.
         * @return Config Builder class object.
         */
        public Builder enableDisConvCallStack(boolean enableDisConvCallStack) {
            this.enableDisConvCallStack = enableDisConvCallStack;
            return this;
        }

        /**
         * Sets the rpcTimeout for the builder.
         *
         * @param rpcTimeout the rpcTimeout.
         * @return Config Builder class object.
         */
        public Builder rpcTimeout(int rpcTimeout) {
            SdkUtils.validateNonNegative(rpcTimeout, "rpcTimeout must be non-negative.");
            this.rpcTimeout = rpcTimeout;
            return this;
        }

        /**
         * Sets the codePath for the builder.
         *
         * @param codePath the codePath.
         * @return Config Builder class object.
         */
        public Builder codePath(List<String> codePath) {
            this.codePath = codePath;
            return this;
        }

        /**
         * Builds the Config object.
         *
         * @return Config class object.
         */
        public Config build() {
            return new Config(this);
        }
    }

    /**
     * Build the cluster access info.
     *
     * @return YRAutoInitInfo.
     */
    public YRAutoInitInfo buildClusterAccessInfo() {
        YRAutoInitInfo info = new YRAutoInitInfo();
        info.setInCluster(this.isInCluster);
        info.setFunctionSystemServerIpAddr(this.serverAddress);
        if (this.serverAddress != null && !this.serverAddress.isEmpty()) {
            info.setFunctionSystemServerPort(this.serverAddressPort);
            info.setServerAddr(this.serverAddress + ":" + this.serverAddressPort);
        }
        info.setDataSystemIpAddr(this.dataSystemAddress);
        if (this.dataSystemAddress != null && !this.dataSystemAddress.isEmpty()) {
            info.setDataSystemPort(this.dataSystemAddressPort);
            info.setDsAddr(this.dataSystemAddress + ":" + this.dataSystemAddressPort);
        }
        return info;
    }
}
