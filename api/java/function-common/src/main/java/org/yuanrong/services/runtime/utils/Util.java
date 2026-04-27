/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package org.yuanrong.services.runtime.utils;

import org.yuanrong.jni.LibRuntime;
import org.yuanrong.libruntime.generated.Socket.FunctionLog;
import org.yuanrong.runtime.util.Constants;
import org.yuanrong.runtime.util.ExtClasspathLoader;
import org.yuanrong.runtime.util.Utils;
import org.yuanrong.services.runtime.Context;
import org.yuanrong.services.runtime.action.CustomLoggerStream;

import lombok.extern.slf4j.Slf4j;

import java.io.PrintStream;
import java.io.UnsupportedEncodingException;
import java.text.DateFormat;
import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Date;
import java.util.Locale;

/**
 * The type Util.
 *
 * @since 2024 /08/20
 */
@Slf4j
public class Util {
    private static final int MAX_SINGLE_LOG_SIZE = 90 * 1024;
    private static final CustomLoggerStream SYS_OUT_CUSTOM_LOGGER_STREAM = new CustomLoggerStream("INFO");
    private static final CustomLoggerStream SYS_ERR_CUSTOM_LOGGER_STREAM = new CustomLoggerStream("INFO");

    private static InheritableThreadLocal<String[]> inheritableThreadLocal = new InheritableThreadLocal<>();
    private static InheritableThreadLocal<String[]> invokeLogOptsThreadLocal = new InheritableThreadLocal<>();

    /**
     * Set the loggers based on ENABLE_LOG_PATH
     */
    public static void setLoggers() {
        String enableLogPath = System.getenv("ENABLE_LOG_PATH");
        ClassLoader classLoader = ExtClasspathLoader.getFunctionClassLoader();
        if (!"true".equals(enableLogPath)) {
            try {
                System.setOut(new PrintStream(SYS_OUT_CUSTOM_LOGGER_STREAM, true, "UTF-8"));
                System.setErr(new PrintStream(SYS_ERR_CUSTOM_LOGGER_STREAM, true, "UTF-8"));
            } catch (UnsupportedEncodingException e) {
                log.warn("UnsupportedEncodingException: {}", e.getMessage());
            } catch (Exception e) {
                log.error("load log config catch a exception: {}", e.getMessage());
            }
        } else {
            log.info("enableLogPath is true");
        }
    }

    /**
     * substring
     *
     * @param str string
     * @param fi  start index
     * @param ti  end index
     * @return substring
     */
    public static String substring(String str, int fi, int ti) {
        int strLength = str.length();
        if (fi > strLength) {
            return "";
        }
        if (ti > strLength) {
            return str.substring(fi, strLength);
        } else {
            return str.substring(fi, ti);
        }
    }

    /**
     * splitMessage
     *
     * @param messageString message string
     * @return lines ArrayList
     */
    public static ArrayList<String> splitMessage(String messageString) {
        ArrayList<String> lines = new ArrayList<String>();
        int length = messageString.length();
        int splitTimes = length / MAX_SINGLE_LOG_SIZE;
        int remainder = length % MAX_SINGLE_LOG_SIZE;
        if (remainder == 0) {
            if (splitTimes == 0) {
                lines.add(messageString);
            }
            for (int i = 0; i < splitTimes; i++) {
                String childStr = substring(messageString, i * MAX_SINGLE_LOG_SIZE, (i + 1) * MAX_SINGLE_LOG_SIZE);
                lines.add(childStr);
            }
        } else {
            for (int i = 0; i <= splitTimes; i++) {
                if (i < splitTimes) {
                    String childStr = substring(messageString, i * MAX_SINGLE_LOG_SIZE, (i + 1) * MAX_SINGLE_LOG_SIZE);
                    lines.add(childStr);
                } else {
                    String childStr = substring(messageString, i * MAX_SINGLE_LOG_SIZE, length);
                    lines.add(childStr);
                }
            }
        }
        return lines;
    }

    /**
     * set logType and livedataTraceID into invokeLogOptsThreadLocal
     *
     * @param logType logType
     */
    public static void setLogOpts(String logType) {
        invokeLogOptsThreadLocal.set(new String[]{logType});
    }

    /**
     * getLogOpts
     *
     * @return logType and livedataTraceID
     */
    public static String[] getLogOpts() {
        String[] logOpts = invokeLogOptsThreadLocal.get();
        return logOpts == null ? new String[]{""} : logOpts;
    }

    /**
     * clear invokeLogOptsThreadLocal
     */
    public static void clearLogOpts() {
        invokeLogOptsThreadLocal.remove();
    }

    /**
     * set invokeId and requestId into InheritableThreadLocal
     *
     * @param callInfos the array of invokeId, requestId, instanceId, functionInfo,
     *                  logGroupId and logStreamId.
     */
    public static void setInheritableThreadLocal(String[] callInfos) {
        inheritableThreadLocal.set(callInfos);
    }

    /**
     * getInheritableThreadLocal
     *
     * @return the array of invokeId, requestId, instanceId, functionInfo,
     *         logGroupId and logStreamId.
     */
    public static String[] getInheritableThreadLocal() {
        return inheritableThreadLocal.get();
    }

    /**
     * clearInheritableThreadLocal
     */
    public static void clearInheritableThreadLocal() {
        inheritableThreadLocal.remove();
    }

    /**
     * get ServiceNameFromEnv
     *
     * @param context the context
     * @return service
     */
    public static String getServiceNameFromEnv(Context context) {
        String service = "";
        if (context != null) {
            service = context.getPackage();
        }
        if (service == null || service.isEmpty()) {
            service = Utils.getFromJavaEnv("RUNTIME_PACKAGE");
        }
        return service;
    }

    /**
     * get TenantIdFromEnv
     *
     * @param context the context
     * @return tenantId
     */
    public static String getTenantIdFromEnv(Context context) {
        String tenantId = "";
        if (context != null) {
            tenantId = context.getProjectID();
        }
        if (tenantId == null || tenantId.isEmpty()) {
            tenantId = Utils.getFromJavaEnv("RUNTIME_PROJECT_ID");
        }
        return tenantId;
    }

    /**
     * getFunctionInfo
     *
     * @param context context
     * @return function info
     */
    public static String getFunctionInfo(Context context) {
        return String.format(Locale.ROOT, "%s:function:0@%s@%s:%s", getTenantIdFromEnv(context),
                            getServiceNameFromEnv(context), context.getFunctionName(), context.getVersion());
    }

    /**
     * Get UTC ISO8601 timestamp
     *
     * @param date current time
     * @return string format of date
     */
    public static String getISO8601Timestamp(Date date) {
        DateFormat dateFormat = new SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss.SSS'Z'");
        return dateFormat.format(date);
    }

    /**
     * send start log
     *
     * @param logType log type
     * @param callInfos call infos
     */
    public static void sendStartLog(String logType, String[] callInfos) {
        FunctionLog rpcLog = FunctionLog.newBuilder()
                .setLevel("INFO")
                .setContent("")
                .setTimestamp(Util.getISO8601Timestamp(new Date()))
                .setIsStart(true)
                .setLogType(logType)
                .setFunctionInfo(callInfos[Constants.INDEX_FORTH])
                .setLogSource(Constants.LOGGER_TYPE)
                .setLogGroupId(callInfos[Constants.INDEX_FIFTH])
                .setLogStreamId(callInfos[Constants.INDEX_SIXTH])
                .build();
        LibRuntime.processLog(rpcLog);
    }

    /**
     * send finish log
     *
     * @param innerCode inner error code
     * @param logType log type
     * @param callInfos call infos
     */
    public static void sendFinishLog(int innerCode, String logType, String[] callInfos) {
        String level = innerCode != 0 ? "ERROR" : "INFO";
        FunctionLog rpcLog = FunctionLog.newBuilder()
                .setLevel(level)
                .setContent("")
                .setTimestamp(Util.getISO8601Timestamp(new Date()))
                .setErrorCode(innerCode)
                .setIsFinish(true)
                .setLogType(logType)
                .setFunctionInfo(callInfos[Constants.INDEX_FORTH])
                .setLogSource(Constants.LOGGER_TYPE)
                .setLogGroupId(callInfos[Constants.INDEX_FIFTH])
                .setLogStreamId(callInfos[Constants.INDEX_SIXTH])
                .build();
        LibRuntime.processLog(rpcLog);
    }
}
