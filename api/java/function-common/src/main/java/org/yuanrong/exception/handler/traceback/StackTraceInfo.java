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

package org.yuanrong.exception.handler.traceback;

import lombok.Data;

import java.io.Serializable;
import java.util.ArrayList;
import java.util.List;

/**
 * Used to organize stacktrace information
 *
 * @since 2023 /12/25
 */
@Data
public class StackTraceInfo implements Serializable {
    private static final long serialVersionUID = 2363655188479713663L;

    private static final String LANGUAGE = "java";

    /**
     * Type of user exception or error, get from exception.getClass().getName()
     */
    private String type = "";

    /**
     * Message of the exception or error, get from exception.getMessage()
     */
    private String message = "";

    /**
     * StackTrace of the exception or error, get from exception.getStackTrace()
     */
    private List<StackTraceElement> stackTraceElements = new ArrayList<>();

    private String language = LANGUAGE;

    /**
     * Instantiates a new Stack trace info.
     */
    public StackTraceInfo() {}

    public StackTraceInfo(String type, String message) {
        this(type, message, new ArrayList<>());
    }

    public StackTraceInfo(String type, String message, List<StackTraceElement> stackTraceElements) {
        this(type, message, stackTraceElements, LANGUAGE);
    }

    public StackTraceInfo(String type, String message, List<StackTraceElement> stackTraceElements, String language) {
        this.type = type;
        this.message = message;
        this.stackTraceElements = stackTraceElements;
        this.language = language;
    }

    @Override
    public String toString() {
        return "type:" + type + System.lineSeparator() + "message:" + message + System.lineSeparator() + "stacktraces:"
            + stackTraceElements;
    }
}
