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

/**
 * Used to create certain StackTrace Filter
 *
 * @since 2023 /12/25
 */
public class FilterFactory {
    private StackTraceFilter filter;
    private Throwable throwable;

    /**
     * Instantiates a new Filter factory.
     *
     * @param exception the exception
     * @param isUserStackTrace  the isUserStackTrace
     */
    public FilterFactory(Throwable exception, boolean isUserStackTrace) {
        this.throwable = exception;
        if (isUserStackTrace) {
            this.filter = new UserStackTraceFilter(exception);
        } else {
            this.filter = new FuncCallStackTraceFilter(exception);
        }
    }

    /**
     * Gets stack trace info.
     *
     * @param clzName the clzName
     * @param funcName the funcName
     * @return the stack trace info
     */
    public StackTraceInfo generateStackTraceInfo(String clzName, String funcName) {
        return filter.generateStackTraceInfo(clzName, funcName);
    }

    public Throwable getThrowable() {
        return this.throwable;
    }
}
