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

/**
 * The type Func call stack trace filter.
 *
 * @since 2023/12/26
 */
public class FuncCallStackTraceFilter extends StackTraceFilter {
    private Throwable cause;

    /**
     * Instantiates a new Func call stack trace filter.
     *
     * @param exception the exception
     */
    public FuncCallStackTraceFilter(Throwable exception) {
        super(exception);
        this.cause = exception;
    }
}
