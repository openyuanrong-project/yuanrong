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

import org.yuanrong.runtime.config.RuntimeContext;
import org.yuanrong.runtime.util.Constants;
import org.yuanrong.utils.SdkUtils;

import lombok.Data;

import java.util.List;
import java.util.ArrayList;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * Singleton class for manage user input Config
 * make sure that input config won't be changed after intitialization
 *
 * @since 2023 /10/16
 */
@Data
public class ConfigManager {
    private boolean isInitialized = false;
    private String functionURN;
    private String serverAddress;
    private int serverAddressPort;
    private String dataSystemAddress;
    private int dataSystemAddressPort;
    private String jobId = "";
    private String runtimeId = "";
    private String cppFunctionURN;
    private String goFunctionURN;
    private int recycleTime;
    private boolean isInCluster;
    private int maxTaskInstanceNum = -1;
    private boolean enableMetrics = true;
    private boolean enableMTLS = false;
    private String certificateFilePath;
    private String privateKeyPath;
    private String verifyFilePath;
    private String privateKeyPaaswd;
    private String serverName;
    private String driverServerIP;
    private boolean isDriver = true;
    private int maxConcurrencyCreateNum = 100;
    private List<String> codePath;
    private String logLevel;
    private String logDir;
    private String ns = "";
    private int logFileSizeMax;
    private int logFileNumMax;
    private boolean isLogMerge = false;
    private int threadPoolSize;
    private ArrayList<String> loadPaths;
    private int httpIocThreadsNum = Constants.DEFAULT_HTTP_IO_THREAD_CNT;
    private int httpIdleTime = Constants.DEFAULT_HTTP_IDLE_TIME;
    private boolean enableDisConvCallStack;
    private int rpcTimeout;
    private String tenantId = "";
    private Map<String, String> customEnvs;
    private boolean enableDsEncrypt = false;
    private String dsPublicKeyContextPath = "";
    private String runtimePublicKeyContextPath = "";
    private String runtimePrivateKeyContextPath = "";
    private boolean enableFrontendTLS;

    private static class ConfigManagerHolder {
        static final ConcurrentHashMap<String, ConfigManager> INSTANCE_CACHE = new ConcurrentHashMap<>();
    }

    /**
     * Gets an instance of ConfigManager. If the user uses the isThreadLocal mode,
     * an instance of ThreadLocal is returned.
     *
     * @return the instance of ConfigManager
     */
    public static ConfigManager getInstance() {
        String runtimeCtx = RuntimeContext.RUNTIME_CONTEXT.get();
        ConfigManagerHolder.INSTANCE_CACHE.putIfAbsent(runtimeCtx, new ConfigManager());
        return ConfigManagerHolder.INSTANCE_CACHE.get(runtimeCtx);
    }

    /**
     * init
     *
     * @param config config
     */
    public synchronized void init(Config config) {
        this.jobId = SdkUtils.generateJobId();
        this.runtimeId = SdkUtils.generateRuntimeId();
        this.functionURN = config.getFunctionURN();
        this.serverAddress = config.getServerAddress();
        this.dataSystemAddress = config.getDataSystemAddress();
        this.cppFunctionURN = config.getCppFunctionURN();
        this.goFunctionURN = config.getGoFunctionURN();
        this.recycleTime = config.getRecycleTime();
        this.isInCluster = config.isInCluster();
        this.maxTaskInstanceNum = config.getMaxTaskInstanceNum();
        this.enableMetrics = config.isEnableMetrics();
        this.maxConcurrencyCreateNum = config.getMaxConcurrencyCreateNum();
        this.logLevel = config.getLogLevel();
        this.logDir = config.getLogDir();
        this.logFileSizeMax = config.getMaxLogSizeMb();
        this.logFileNumMax = config.getMaxLogFileNum();
        this.isLogMerge = config.isEnableSetContext();
        this.threadPoolSize = config.getThreadPoolSize();
        this.loadPaths = config.getLoadPaths();
        this.serverAddressPort = config.getServerAddressPort();
        this.dataSystemAddressPort = config.getDataSystemAddressPort();
        this.httpIocThreadsNum = config.getHttpIocThreadsNum();
        this.ns = config.getNs();
        this.enableDisConvCallStack = config.isEnableDisConvCallStack();
        this.rpcTimeout = config.getRpcTimeout();
        this.tenantId = config.getTenantId();
        this.customEnvs = config.getCustomEnvs();
        this.enableMTLS = config.isEnableMTLS();
        this.certificateFilePath = config.getCertificateFilePath();
        this.privateKeyPath = config.getPrivateKeyPath();
        this.verifyFilePath = config.getVerifyFilePath();
        this.privateKeyPaaswd = config.getPrivateKeyPaaswd();
        this.serverName = config.getServerName();
        this.isInitialized = true;
        this.codePath = config.getCodePath();
        this.enableDsEncrypt = config.isEnableDsEncrypt();
        this.dsPublicKeyContextPath = config.getDsPublicKeyContextPath();
        this.runtimePublicKeyContextPath = config.getRuntimePublicKeyContextPath();
        this.runtimePrivateKeyContextPath = config.getRuntimePrivateKeyContextPath();
        this.enableFrontendTLS = config.isEnableFrontendTLS();
    }
}
