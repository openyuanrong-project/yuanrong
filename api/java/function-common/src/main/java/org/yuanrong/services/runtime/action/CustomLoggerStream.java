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

package org.yuanrong.services.runtime.action;

import org.yuanrong.jni.LibRuntime;
import org.yuanrong.libruntime.generated.Socket.FunctionLog;
import org.yuanrong.runtime.util.Constants;
import org.yuanrong.services.runtime.utils.Util;

import lombok.SneakyThrows;

import java.nio.charset.StandardCharsets;
import java.io.ByteArrayOutputStream;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * CustomLoggerStream catches stdout and stderr, and write back to runtime-manager.
 *
 * @since 2024/8/22
 */
public class CustomLoggerStream extends ByteArrayOutputStream {
    private static final int DEFAULT_BUFFER_SIZE = 32;
    private static final String REQUEST_NAME = Constants.LOG_REQUEST_ID + ":";
    private static final String REQUEST_REGEX = Constants.LOG_REQUEST_ID + ":(\\S)+";
    private static final String REQUEST_REPLACE = REQUEST_REGEX + " ";
    private static final Pattern PATTERN = Pattern.compile(REQUEST_REGEX);

    private final String logLevel;

    /**
     * The type CustomLoggerStream.
     *
     * @param logLevel standard log level
     */
    public CustomLoggerStream(String logLevel) {
        this.logLevel = logLevel;
    }

    @SneakyThrows
    @Override
    public synchronized void write(byte[] bt, int off, int len) {
        if ((off < 0) || (off > bt.length) || (len < 0) || ((off + len) - bt.length > 0)) {
            throw new IndexOutOfBoundsException();
        }
        if (Arrays.equals(System.lineSeparator().getBytes(StandardCharsets.UTF_8), bt)) {
            return;
        }
        // update buffer
        if (buf.length < count + len) {
            buf = Arrays.copyOf(buf, count + len);
        }
        System.arraycopy(bt, off, buf, count, len);
        count += len;

        String message = new String(buf, StandardCharsets.UTF_8).trim();
        if (message.isEmpty()) {
            return;
        }

        // reset buffer
        buf = new byte[DEFAULT_BUFFER_SIZE];
        count = 0;

        if (!message.contains(System.lineSeparator())) {
            handleLargeLog(message);
            return;
        }

        // handle multiple logs are merged
        String[] splitMessage = message.split(System.lineSeparator());
        for (int i = 0; i < splitMessage.length; i++) {
            String splitMessageStr = splitMessage[i];
            handleLargeLog(splitMessageStr);
        }
    }

    private String cutRequestIDAndInvokeID(String str) {
        return str.replaceAll(REQUEST_REPLACE, "");
    }

    private String[] handleRequestIDAndInvokeID(String str) {
        if (getRequestIDAndInvokeID(str).isEmpty()) {
            return new String[]{};
        }
        return getRequestIDAndInvokeID(str).split("\\|");
    }

    private String getRequestIDAndInvokeID(String str) {
        Matcher matcher = PATTERN.matcher(str);
        if (matcher.find()) {
            return matcher.group().substring(REQUEST_NAME.length());
        }
        return "";
    }

    private void handleLargeLog(String message) {
        String requestID = "";
        String invokeID = "";
        String[] requestIDAndInvokeID = handleRequestIDAndInvokeID(message);
        String handleMessage = cutRequestIDAndInvokeID(message);

        if (requestIDAndInvokeID.length == 0 && Util.getInheritableThreadLocal() == null) {
            return;
        }

        if (requestIDAndInvokeID.length != 2) {
            invokeID = Util.getInheritableThreadLocal()[Constants.INDEX_FIRST];
            requestID = Util.getInheritableThreadLocal()[Constants.INDEX_SECOND];
        } else {
            requestID = requestIDAndInvokeID[Constants.INDEX_FIRST];
            invokeID = requestIDAndInvokeID[Constants.INDEX_SECOND];
        }

        ArrayList<String> lines = Util.splitMessage(handleMessage);
        for (String line : lines) {
            buildFunctionLog(line, requestID, invokeID, Util.getInheritableThreadLocal(), Util.getLogOpts());
        }
    }

    private void buildFunctionLog(String message, String requestID, String invokeID, String[] callInfos,
                                  String[] logOpts) {
        FunctionLog rpcLog = FunctionLog.newBuilder()
                .setLevel(logLevel)
                .setContent(message)
                .setInvokeID(invokeID)
                .setTraceID(requestID)
                .setLogType(logOpts[Constants.INDEX_FIRST])
                .setInstanceId(callInfos[Constants.INDEX_THIRD])
                .setFunctionInfo(callInfos[Constants.INDEX_FORTH])
                .setLogSource(Constants.STD_TYPE)
                .setLogGroupId(callInfos[Constants.INDEX_FIFTH])
                .setLogStreamId(callInfos[Constants.INDEX_SIXTH])
                .build();
        LibRuntime.processLog(rpcLog);
    }
}
