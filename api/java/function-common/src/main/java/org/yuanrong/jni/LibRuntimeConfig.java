/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
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

package org.yuanrong.jni;

import org.yuanrong.libruntime.generated.Libruntime;

import lombok.Data;

import java.util.List;
import java.util.Map;

/**
 * Translated from `YR::LibRuntime::LibRuntimeConfig`
 * One missing `libruntimeOptions` will be filled by JNI code.
 *
 * @since 2023/10/23
 */
@Data
public class LibRuntimeConfig {
    private String functionSystemIpAddr = "";
    private int functionSystemPort = 0;
    private String functionSystemRtServerIpAddr = "";
    private int functionSystemRtServerPort = 0;
    private String dataSystemIpAddr = "";
    private int dataSystemPort = 0;
    private boolean isDriver = false;
    private String jobId;
    private String runtimeId;
    private String functionUrn = "";
    private String logLevel;
    private String logDir;
    private int logFileSizeMax = 1024;
    private int logFileNumMax = 1024;
    private int logFlushInterval = 1;
    private boolean isLogMerge = false;
    private String metaConfig;
    private int recycleTime;
    private int maxTaskInstanceNum;
    private int maxConcurrencyCreateNum;
    private boolean enableMetrics;
    private int threadPoolSize;
    private Map<Libruntime.LanguageType, String> functionIds;
    private List<String> loadPaths;
    private boolean enableMTLS = false;
    private String certificateFilePath;
    private String privateKeyPath;
    private String verifyFilePath;
    private String privateKeyPaaswd;
    private String serverName;
    private int httpIocThreadsNum;
    private int httpIdleTime;
    private boolean inCluster = true;
    private int rpcTimeout = 60;
    private String tenantId = "";
    private String ns = "";
    private Map<String, String> customEnvs;
    private List<String> codePath;
    private boolean enableDsEncrypt = false;
    private String dsPublicKeyContextPath = "";
    private String runtimePublicKeyContextPath = "";
    private String runtimePrivateKeyContextPath = "";
    private boolean enableFrontendTLS;
}
