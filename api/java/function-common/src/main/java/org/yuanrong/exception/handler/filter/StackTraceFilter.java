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

package org.yuanrong.exception.handler.filter;

import org.yuanrong.exception.handler.traceback.StackTraceInfo;

import lombok.Data;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.PrintWriter;
import java.io.StringWriter;
import java.util.ArrayList;
import java.util.Arrays;

/**
 * Used to filter stacktrace information
 *
 * @since 2023 /12/25
 */
@Data
public abstract class StackTraceFilter {
    /**
     * The constant REGEX.
     */
    public static final String REGEX = "((\\w+\\.?)+\\w+)\\.(\\w+)\\((\\w+\\.java):(\\d+)\\)";

    private static final Logger LOG = LoggerFactory.getLogger(StackTraceFilter.class);

    private static final String SUPPRESSED_EXCEPTION_BEGIN_TAG = "Caused by:";

    private static final String AT_PATTERN = "^\\tat.*$";

    /**
     * The Exception.
     */
    protected Throwable cause;

    /**
     * Instantiates a new Stack trace filter.
     *
     * @param exception the exception
     */
    public StackTraceFilter(Throwable exception) {
        this.cause = exception;
    }

    /**
     * Filter stack trace info.
     *
     * @param clzName the clzName
     * @param funcName the funcName
     * @return the stack trace info
     */
    public StackTraceInfo generateStackTraceInfo(String clzName, String funcName) {
        StackTraceInfo stackTraceInfo;
        if (this.cause == null) {
            stackTraceInfo = new StackTraceInfo(clzName, funcName);
        } else {
            stackTraceInfo = new StackTraceInfo(this.cause.getClass().getName(), this.cause.getMessage());
            /*
             * StackTraceInfo object cannot be serialized by ByteArrayOutputStream or messagepack
             * with exception that inner parameter class for example ErrorCode has no default constructor
             */
            String causedBy = getCausedByString(this.cause);
            LOG.debug("Get causeByString in StackTraceFilter for {}.{}, causedBy {}", clzName, funcName, causedBy);
            stackTraceInfo.setStackTraceElements(new ArrayList<>(Arrays.asList(this.cause.getStackTrace())));
            stackTraceInfo.setMessage(this.cause.getMessage());
        }
        return stackTraceInfo;
    }

    /**
     * Gets printed stacktrace string.
     *
     * @param throwable the throwable
     * @return the caused by string
     */
    public String getPrintStackTrace(Throwable throwable) {
        StringWriter stringWriter = new StringWriter();
        PrintWriter printWriter = new PrintWriter(stringWriter);
        throwable.printStackTrace(printWriter);
        String printedStackTraceStr = stringWriter.toString();
        printWriter.close();
        LOG.debug("throwable printedStackTraceStr in StackTraceFilter: {}", printedStackTraceStr);
        return printedStackTraceStr;
    }

    /**
     * Gets caused by string.
     *
     * @param throwable the throwable
     * @return the caused by string
     */
    public String getCausedByString(Throwable throwable) {
        StringWriter stringWriter = new StringWriter();
        PrintWriter printWriter = new PrintWriter(stringWriter);
        throwable.printStackTrace(printWriter);
        String causedBy = stringWriter.toString();
        printWriter.close();
        LOG.debug("throwable stack trace string in StackTraceFilter: {}", causedBy);
        return causedBy;
    }
}
