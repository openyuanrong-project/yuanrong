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
import org.yuanrong.exception.handler.traceback.StackTraceUtils;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.List;

/**
 * The type User stack trace filter.
 *
 * @since 2023/12/26
 */
public class UserStackTraceFilter extends StackTraceFilter {
    private static final Logger LOG = LoggerFactory.getLogger(UserStackTraceFilter.class);

    private Throwable cause;

    /**
     * Instantiates a new User stack trace filter.
     *
     * @param exception the exception
     */
    public UserStackTraceFilter(Throwable exception) {
        super(exception);
        this.cause = exception;
    }

    /**
     * Filter stack trace stack trace info.
     *
     * @param clzName the clzName
     * @param funcName the funcName
     * @return the stack trace info
     */
    @Override
    public StackTraceInfo generateStackTraceInfo(String clzName, String funcName) {
        StackTraceInfo stackTraceInfo = new StackTraceInfo(this.cause.getClass().getName(), this.cause.getMessage());

        String printedStackTrace = super.getPrintStackTrace(this.cause);
        LOG.debug("Get printedStackTrace UserStackTraceFilter FuncCallStackTraceFilter for {}.{}, causedBy {}", clzName,
            funcName, printedStackTrace);
        List<StackTraceElement> stackTraceElements = StackTraceUtils.processValidAndDuplicate(
            StackTraceUtils.getListTraceElements(printedStackTrace));
        LOG.debug("Stack trace from stacktraceUtils stackTraceElements.size():{}, content {} ",
            stackTraceElements.size(), stackTraceElements);
        stackTraceInfo.setStackTraceElements(stackTraceElements);
        return stackTraceInfo;
    }
}
