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

package org.yuanrong.errorcode;

import org.yuanrong.exception.handler.traceback.StackTraceInfo;

import java.util.List;

/**
 * This class represents error information, which contains an error code, a
 * module code, and an error message.
 *
 * @since 2023/11/11
 */
public class ErrorInfo {
    private ErrorCode errorCode = ErrorCode.ERR_OK;
    private ModuleCode moduleCode = ModuleCode.RUNTIME;
    private String errorMessage;
    private List<StackTraceInfo> stackTraceInfos;

    /**
     * Instantiates a new Error info.
     */
    public ErrorInfo() {}

    /**
     * Instantiates a new Error info.
     *
     * @param errorCode    the error code
     * @param moduleCode   the module code
     * @param errorMessage the error message
     */
    public ErrorInfo(ErrorCode errorCode, ModuleCode moduleCode, String errorMessage) {
        this(errorCode, moduleCode, errorMessage, null);
    }

    public ErrorInfo(ErrorCode errorCode, ModuleCode moduleCode, String errorMessage,
        List<StackTraceInfo> stackTraceInfos) {
        this.errorCode = errorCode;
        this.moduleCode = moduleCode;
        this.errorMessage = errorMessage;
        this.stackTraceInfos = stackTraceInfos;
    }

    /**
     * Gets the error code.
     *
     * @return the error code
     */
    public ErrorCode getErrorCode() {
        return this.errorCode;
    }

    /**
     * Gets the module code.
     *
     * @return the module code
     */
    public ModuleCode getModuleCode() {
        return this.moduleCode;
    }

    /**
     * Gets the error message.
     *
     * @return the error message
     */
    public String getErrorMessage() {
        return this.errorMessage;
    }

    public void setStackTraceInfos(List<StackTraceInfo> infos) {
        this.stackTraceInfos = infos;
    }

    public List<StackTraceInfo> getStackTraceInfos() {
        return this.stackTraceInfos;
    }

    @Override
    public String toString() {
        return "ErrorInfo(errCode=" + String.valueOf(this.errorCode)
            + ",moduleCode=" + String.valueOf(this.moduleCode)
            + ",msg=" + this.errorMessage
            + ",stackTraceInfo=" + (this.stackTraceInfos == null ? "" : this.stackTraceInfos.toString()) + ")";
    }
}