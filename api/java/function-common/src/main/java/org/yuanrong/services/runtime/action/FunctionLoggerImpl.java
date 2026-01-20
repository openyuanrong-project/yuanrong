/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2024. All rights reserved.
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

package org.yuanrong.services.runtime.action;

import org.yuanrong.jni.LibRuntime;
import org.yuanrong.libruntime.generated.Socket.FunctionLog;
import org.yuanrong.runtime.util.Constants;
import org.yuanrong.services.runtime.RuntimeLogger;
import org.yuanrong.services.runtime.utils.Util;

import org.apache.commons.lang3.StringUtils;

import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Date;

/**
 * context.getLogger() returns a instance of FunctionLoggerImpl, which users can use to print logs with request id.
 *
 * @since 2024/8/22
 */
public class FunctionLoggerImpl implements RuntimeLogger {
    private String requestID = "";
    private String invokeID = "";
    private String instanceID = "";
    private String functionInfo = "";
    private String logGroupId = "";
    private String logStreamId = "";
    private String logLevel = "INFO";

    /**
     * Set request ID
     *
     * @param requestID request ID
     */
    void setRequestID(String requestID) {
        this.requestID = requestID;
    }

    /**
     * Set invoke ID
     *
     * @param invokeID invoke ID
     */
    void setInvokeID(String invokeID) {
        this.invokeID = invokeID;
    }

    /**
     * Set instance ID
     *
     * @param instanceID instance ID
     */
    void setInstanceID(String instanceID) {
        this.instanceID = instanceID;
    }

    /**
     * Set function info
     *
     * @param functionInfo function info
     */
    void setFunctionInfo(String functionInfo) {
        this.functionInfo = functionInfo;
    }

    /**
     * Set log group id
     *
     * @param logGroupId log group id
     */
    void setLogGroupId(String logGroupId) {
        this.logGroupId = logGroupId;
    }

    /**
     * Set log stream id
     *
     * @param logStreamId log stream id
     */
    void setLogStreamId(String logStreamId) {
        this.logStreamId = logStreamId;
    }

    @Override
    public void setLevel(String level) {
        if ("DEBUG".equals(level) || "INFO".equals(level) || "WARN".equals(level) || "ERROR".equals(level)) {
            this.logLevel = level;
        }
    }

    @Override
    public void log(String message) {
        info(message);
    }

    @Override
    public void info(String message) {
        if ("WARN".equals(this.logLevel) || "ERROR".equals(this.logLevel)) {
            return;
        }
        if (StringUtils.isBlank(message)) {
            sendLog("log an empty string", "INFO");
        } else {
            sendLog(message, "INFO");
        }
    }

    @Override
    public void debug(String message) {
        if (!"DEBUG".equals(this.logLevel)) {
            return;
        }
        if (StringUtils.isBlank(message)) {
            sendLog("log an empty string", "DEBUG");
        } else {
            sendLog(message, "DEBUG");
        }
    }

    @Override
    public void warn(String message) {
        if ("ERROR".equals(this.logLevel)) {
            return;
        }
        if (StringUtils.isBlank(message)) {
            sendLog("log an empty string", "WARN");
        } else {
            sendLog(message, "WARN");
        }
    }

    @Override
    public void error(String message) {
        if (StringUtils.isBlank(message)) {
            sendLog("log an empty string", "ERROR");
        } else {
            sendLog(message, "ERROR");
        }
    }

    private void sendLog(String message, String level) {
        SimpleDateFormat format = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss.SSSSSSSSSXXX");
        String dateString = format.format(new Date());
        ArrayList<String> lines = Util.splitMessage(message);
        for (int i = 0; i < lines.size(); i++) {
            FunctionLog rpcLog = FunctionLog.newBuilder()
                    .setLevel(level)
                    .setTimestamp(dateString)
                    .setContent(lines.get(i))
                    .setInvokeID(invokeID)
                    .setTraceID(requestID)
                    .setLogType(Util.getLogOpts()[0])
                    .setFunctionInfo(functionInfo)
                    .setLogSource(Constants.LOGGER_TYPE)
                    .setLogGroupId(logGroupId)
                    .setLogStreamId(logStreamId)
                    .build();
            LibRuntime.processLog(rpcLog);
        }
    }
}
