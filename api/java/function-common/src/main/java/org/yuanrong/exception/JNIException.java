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

package org.yuanrong.exception;

import java.util.Objects;

/**
 * The type YR exception.
 *
 * @since 2022 /08/30
 */
public class JNIException extends Exception {
    private final String exceptionMessage;

    private FuncErrorCode errorCode;

    /**
     * Instantiates a new YR exception.
     *
     * @param exception the exception
     */
    public JNIException(Exception exception) {
        this.exceptionMessage = exception.getMessage();
    }

    public JNIException(FuncErrorCode errorCode) {
        this.errorCode = errorCode;
        this.exceptionMessage = errorCode.getDesc();
    }

    /**
     * Instantiates a new YR exception.
     *
     * @param message the message
     */
    public JNIException(String message) {
        this.exceptionMessage = message;
    }

    @Override
    public String getMessage() {
        return this.exceptionMessage;
    }

    public FuncErrorCode getErrorCode() {
        return errorCode;
    }

    public void setErrorCode(FuncErrorCode errorCode) {
        this.errorCode = errorCode;
    }

    @Override
    public boolean equals(Object obj) {
        if (this == obj) {
            return true;
        }
        if (obj == null || getClass() != obj.getClass()) {
            return false;
        }
        JNIException that = (JNIException) obj;
        return errorCode == that.errorCode && Objects.equals(exceptionMessage, that.exceptionMessage);
    }

    @Override
    public int hashCode() {
        return Objects.hash(exceptionMessage, errorCode);
    }
}
