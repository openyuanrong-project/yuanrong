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

import org.yuanrong.api.YR;
import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.exception.YRException;
import org.yuanrong.runtime.client.ObjectRef;
import org.yuanrong.runtime.util.Constants;
import org.yuanrong.runtime.util.Utils;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.lang.reflect.Field;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.InvalidPathException;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.SynchronousQueue;
import java.util.concurrent.ThreadPoolExecutor;
import java.util.concurrent.TimeUnit;

/**
 * The JobExecutor actor runs in remote runtime to manage driver code.
 *
 * @since 2023 /06/06
 */
public class JobExecutor {
    private static final Logger LOGGER = LoggerFactory.getLogger(JobExecutor.class);

    /**
     * The key to yuanrong functionProxy port in remote JobExecutor runtime
     * environment parameter map. The value of the port will be pass to the
     * attached-runtime process as it' environment parameter.
     */
    private static final String DRIVER_SERVER_PORT = "DRIVER_SERVER_PORT";

    /**
     * The key to downloaded driver code path in JobExcutor actor runtime envronment.
     */
    private static final String ENV_DELEGATE_DOWNLOAD = "ENV_DELEGATE_DOWNLOAD";

    /**
     * The key to JobExcutor actor instance ID, which is also called 'userJobID'
     * in the context of JobExecutor.
     */
    private static final String INSTANCE_ID = "INSTANCE_ID";
    private static final String LOG_PREFIX = "(JobExecutor) ";
    private static final String DATASYSTEM_ADDR = "DATASYSTEM_ADDR";
    private static final String HOST_IP = "HOST_IP";
    private static final String YR_SPARK_HOME = "YR_SPARK_HOME";
    private static final String YR_SPARK_FUNC_URN = "YR_SPARK_FUNC_URN";
    private static final String POSIX_LISTEN_ADDR = "POSIX_LISTEN_ADDR";
    private static final String JAVA_HOME = "JAVA_HOME";
    private static final String YR_JOB_AFFINITY = "YR_JOB_AFFINITY";
    private static final String DEFAULT_DS_ADDRESS_PORT = "31501";
    private static final String DEFAULT_DRIVER_SERVER_PORT = "22771";
    private static final String DEFAULT_JAVA_HOME = "/opt/function/runtime/java8/rtsp/jre";
    private static final String DEFAULT_SPARK_HOME = "/dcache/layer/func/bucket-jobexecutor-test/spark_test.zip";
    private static final String DEFAULT_YR_SPARK_FUNC_URN =
            "sn:cn:yrk:default:function:0-sparkonyr-core:$latest";
    private static final int CODE_PATH_INDEX = 1;
    private static final int DEFAULT_GET_TIMEOUT_MS = 30000;

    /**
     * The thread pool for asynchronous log printing and waiting
     * for the end of the attached-runtime.
     */
    private ThreadPoolExecutor attachedRuntimeThreadPool = new ThreadPoolExecutor(3, 3, 60L,
            TimeUnit.SECONDS, new SynchronousQueue<>());
    private int attachedRuntimePid = -1;
    private Process attachedRuntime;
    private String jobName = "";
    private String userJobID;
    private LocalDateTime jobStartTime;
    private volatile LocalDateTime jobEndTime;
    private volatile String errorMessage;
    private RuntimeEnv runtimeEnv;
    private volatile YRJobStatus attachedRuntimeStatus = YRJobStatus.RUNNING;
    private int exitCode;
    private BufferedReader inputReader;
    private BufferedReader errReader;

    /**
     * The Constructor of JobExecutor.
     *
     * @param yrJobParam job parameters (runtime, entrypoint, affinity, etc.)
     * @throws YRException the actor task exception.
     */
    public JobExecutor(YRJobParam yrJobParam) throws YRException {
        this.runtimeEnv = yrJobParam.getRuntimeEnv();
        this.jobName = yrJobParam.getJobName();
        this.userJobID = Utils.getEnvWithDefualtValue(INSTANCE_ID, "", LOG_PREFIX);
        LOGGER.debug("(JobExecutor) Job submitted, jobExecutor actor instanceID: {}", userJobID);

        ArrayList<String> localEntryPoint = yrJobParam.getLocalEntryPoint();
        String affinityMsgJsonStr = yrJobParam.extractInvokeOptions().affinityMsgToJsonStr();
        ObjectRef entryPointObjRef = yrJobParam.getEntryPointObjRef();
        String fileName = yrJobParam.getEntryPointFileName();
        String downloadCodePath = System.getenv(ENV_DELEGATE_DOWNLOAD);
        if (downloadCodePath != null && !downloadCodePath.isEmpty()) {
            localEntryPoint.set(CODE_PATH_INDEX,
                    String.join(Constants.BACKSLASH, downloadCodePath, localEntryPoint.get(CODE_PATH_INDEX)));
        }

        if (entryPointObjRef != null) {
            LOGGER.info("(JobExecutor) entryPointObjRef is not null");
            byte[] data = (byte[]) YR.getRuntime().get(entryPointObjRef, DEFAULT_GET_TIMEOUT_MS);

            if (!saveBytesToFile(data, yrJobParam.getLocalCodePath(), fileName)) {
                throw new YRException(ErrorCode.ERR_INNER_SYSTEM_ERROR, ModuleCode.RUNTIME,
                        "failed to load entryPoint file");
            }
        } else {
            LOGGER.info("(JobExecutor) entryPointObjRef is null");
        }

        LOGGER.info("(JobExecutor) Starts attached-runtime process, entrypoint: {}", localEntryPoint);
        startAttachedRuntime(generateAttachedRuntimeEnv(affinityMsgJsonStr), localEntryPoint);
        asyncReadProcessStream();
        asyncWaitForAttachedRuntime();
    }


    private boolean saveBytesToFile(byte[] data, String pathStr, String fileName) {
        LOGGER.debug("begin to save entrypoint file. {}", pathStr + fileName);
        if (data == null) {
            LOGGER.warn("entryPoint data is null");
            return false;
        }

        // Validate fileName to prevent path traversal attacks
        if (fileName == null || fileName.isEmpty()) {
            LOGGER.error("File name is null or empty");
            return false;
        }
        // Check for path traversal attempts
        if (fileName.contains("..") || fileName.contains("/") || fileName.contains("\\")) {
            LOGGER.error("Potentially unsafe file name detected: {}", fileName);
            return false;
        }

        try {
            Path path = Paths.get(pathStr, fileName);
            Path parent = path.getParent();
            if (parent != null && !Files.exists(parent)) {
                Files.createDirectories(parent);
            }
            // 写入文件
            Files.write(path, data);
            LOGGER.info("File saved successfully: {}{}", pathStr, fileName);
            return true;
        } catch (InvalidPathException e) {
            LOGGER.error("Invalid output path: {}{}", pathStr, fileName);
            return false;
        } catch (IOException e) {
            LOGGER.error("Failed to save the file: {}", e.getMessage());
            return false;
        }
    }

    /**
     * Gets the jobInfo of the the attached-runtime.
     *
     * @param isWithStatic whether invoke static infomation from remote JobExecutor
     *                     runtime.
     * @return YRJobInfo
     */
    public YRJobInfo getJobInfo(boolean isWithStatic) {
        YRJobInfo jobInfo = new YRJobInfo();
        if (isWithStatic) {
            jobInfo.setUserJobID(this.userJobID);
            jobInfo.setJobName(this.jobName);
            jobInfo.setRuntimeEnv(this.runtimeEnv);
            if (this.jobStartTime != null) {
                jobInfo.setJobStartTime(this.jobStartTime.toString());
            }
        }
        jobInfo.setStatus(this.attachedRuntimeStatus);
        jobInfo.setErrorMessage(this.errorMessage);
        if (this.jobEndTime != null) {
            jobInfo.setJobEndTime(this.jobEndTime.toString());
        }
        return jobInfo;
    }

    /**
     * Terminates the attached-runtime process.
     * This method does NOT immediately free the resources occupied by the
     * attached-runtime process.
     */
    public void stop() {
        // If 'stop()' is invoked more than once,
        // 'jobEndTime' should not be changed.
        // Therefore, an 'if' judgment is required here.
        if (this.jobEndTime == null) {
            this.jobEndTime = LocalDateTime.now();
        }
        writeStatus(YRJobStatus.STOPPED);
        if (this.attachedRuntime != null) {
            this.attachedRuntime.destroy();
        }
        this.clearThreadPool();
        LOGGER.info("(JobExecutor) Stop the attached-runtime process (pid: {}), job name: {}."
                + " Current job status: {}, stop time: {}.", this.attachedRuntimePid, this.jobName,
                this.attachedRuntimeStatus, this.jobEndTime);
    }

    private Map<String, String> generateAttachedRuntimeEnv(String affinityMsgJsonStr) throws YRException {
        String defaultDsAddr = System.getenv(HOST_IP) + ":" + DEFAULT_DS_ADDRESS_PORT;

        Map<String, String> arEnv = new HashMap<String, String>();
        arEnv.put(YR_SPARK_HOME, Utils.getEnvWithDefualtValue(YR_SPARK_HOME, DEFAULT_SPARK_HOME, LOG_PREFIX));
        arEnv.put(YR_SPARK_FUNC_URN,
                Utils.getEnvWithDefualtValue(YR_SPARK_FUNC_URN, DEFAULT_YR_SPARK_FUNC_URN, LOG_PREFIX));
        arEnv.put(DATASYSTEM_ADDR, Utils.getEnvWithDefualtValue(DATASYSTEM_ADDR, defaultDsAddr, LOG_PREFIX));
        arEnv.put(POSIX_LISTEN_ADDR, Utils.getEnvWithDefualtValue(POSIX_LISTEN_ADDR, "", LOG_PREFIX));
        arEnv.put(DRIVER_SERVER_PORT,
                Utils.getEnvWithDefualtValue(DRIVER_SERVER_PORT, DEFAULT_DRIVER_SERVER_PORT, LOG_PREFIX));
        arEnv.put(JAVA_HOME, Utils.getEnvWithDefualtValue(JAVA_HOME, DEFAULT_JAVA_HOME, LOG_PREFIX));
        arEnv.put(Constants.AUTHORIZATION, Utils.getEnvWithDefualtValue(Constants.AUTHORIZATION, "", LOG_PREFIX));

        String runtimeEnvCommand = this.runtimeEnv.toCommand();
        arEnv.put(Constants.POST_START_EXEC, runtimeEnvCommand);
        LOGGER.debug("(JobExecutor) Sets environment key-value pair({}: {}) for job (jobName: {}).",
                Constants.POST_START_EXEC, runtimeEnvCommand, this.jobName);

        arEnv.put(YR_JOB_AFFINITY, affinityMsgJsonStr);
        LOGGER.debug("(JobExecutor) Sets environment key-value pair({}: {}) for job (jobName: {}).",
                YR_JOB_AFFINITY, affinityMsgJsonStr, this.jobName);

        return arEnv;
    }

    private void startAttachedRuntime(Map<String, String> attachedRuntimeEnv, List<String> entryPoint) {
        ProcessBuilder processBuilder = new ProcessBuilder(entryPoint);
        Map<String, String> env = processBuilder.environment();
        env.putAll(attachedRuntimeEnv);

        try {
            this.attachedRuntime = processBuilder.start();
            this.jobStartTime = LocalDateTime.now();
        } catch (IOException e) {
            LOGGER.error("(JobExecutor) Failed to start the attached-runtime process for the job: {}. Exception: ",
                    this.jobName, e);
            this.errorMessage = "Failed to start the attached-runtime process, userJobID: " + this.userJobID;
            writeStatus(YRJobStatus.FAILED);
            return;
        }

        this.attachedRuntimePid = getProcessPid(this.attachedRuntime);
        LOGGER.info(
                "(JobExecutor) Succeeded to start a process as an attached-runtime(pid: {}) to run driver code: {}",
                this.attachedRuntimePid, this.jobName);
    }

    private void asyncWaitForAttachedRuntime() {
        if (this.attachedRuntime == null) {
            LOGGER.warn("(JobExecutor) Attached-runtime has not been started.");
            return;
        }
        this.attachedRuntimeThreadPool.execute(() -> {
            try {
                this.exitCode = this.attachedRuntime.waitFor();
                if (this.jobEndTime == null) {
                    this.jobEndTime = LocalDateTime.now();
                }
            } catch (InterruptedException e) {
                if (this.jobEndTime != null) {
                    // If 'this.jobEndTime' has not null value,
                    // attached-runtime is stopped by calling 'stop()',
                    // which is not regarded as an unexpected error.
                    return;
                }
                LOGGER.warn("(JobExecutor) Attached-runtime process (pid: {}) exits abnormally, Exception: ",
                        this.attachedRuntimePid, e);
                writeStatus(YRJobStatus.FAILED);
                this.clearThreadPool();
                return;
            }
            LOGGER.info("(JobExecutor) Attached-runtime process(pid: {}) exit. Exit code: {}",
                    this.attachedRuntimePid, this.exitCode);
            if (this.exitCode != 0) {
                writeStatus(YRJobStatus.FAILED);
                this.clearThreadPool();
                return;
            }
            writeStatus(YRJobStatus.SUCCEEDED);
            // One JobExecutor runtime maps to one attached-runtime; close its thread pool when done.
            this.clearThreadPool();
        });
    }

    private int getProcessPid(Process process) {
        int pid = -1;
        Field field;
        try {
            field = process.getClass().getDeclaredField("pid");
            field.setAccessible(true);
            pid = (int) field.get(process);
        } catch (NoSuchFieldException | SecurityException | IllegalArgumentException | IllegalAccessException e) {
            LOGGER.warn("(JobExecutor) Failed to get pid of driver code (jobName: {}) process. Exception: ",
                    this.jobName, e);
        }
        return pid;
    }

    private void writeStatus(YRJobStatus status) {
        switch (status) {
            case RUNNING:
                this.attachedRuntimeStatus = status;
                break;
            default:
                if (this.attachedRuntimeStatus == null || this.attachedRuntimeStatus == YRJobStatus.RUNNING) {
                    this.attachedRuntimeStatus = status;
                }
        }
    }

    // 异步读取输出流和错误流
    private void asyncReadProcessStream() {
        if (this.attachedRuntime == null) {
            LOGGER.warn("(JobExecutor) Attached-runtime has not been started.");
            return;
        }

        this.errReader = new BufferedReader(
            new InputStreamReader(this.attachedRuntime.getErrorStream(), StandardCharsets.UTF_8));
        this.attachedRuntimeThreadPool.execute(() -> {
            String line;
            this.errorMessage = "";
            try {
                while ((line = errReader.readLine()) != null) {
                    this.errorMessage = this.errorMessage + System.lineSeparator() + line;
                }
            } catch (IOException e) {
                LOGGER.error("(JobExecutor) Failed to read attached-runtime error message, Exception: ", e);
                this.errorMessage = "Failed to read attached-runtime error message. Exception: " + e.getMessage();
            }
            if (!this.errorMessage.isEmpty()) {
                LOGGER.error("(Attached-runtime) {}", this.errorMessage);
            }
        });

        this.inputReader = new BufferedReader(
            new InputStreamReader(this.attachedRuntime.getInputStream(), StandardCharsets.UTF_8));
        this.attachedRuntimeThreadPool.execute(() -> {
            String line;
            try {
                while ((line = inputReader.readLine()) != null) {
                    LOGGER.info("(Attached-runtime) {}", line);
                }
            } catch (IOException e) {
                LOGGER.error("(JobExecutor) Failed to read attached-runtime inputStream message. Exception: ", e);
            }
        });
    }

    private void clearThreadPool() {
        try {
            if (this.inputReader != null) {
                this.inputReader.close();
            }
        } catch (IOException e) {
            LOGGER.error("(JobExecutor) Failed to close attached-runtime input stream, Exception: ", e);
        }

        try {
            if (errReader != null) {
                errReader.close();
            }
        } catch (IOException e) {
            LOGGER.error("(JobExecutor) Failed to close attached-runtime error stream, Exception: ", e);
        }

        this.attachedRuntimeThreadPool.shutdownNow();
    }
}
