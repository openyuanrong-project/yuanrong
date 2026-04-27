/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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

package org.yuanrong.function.runtime.exception;

/**
 * Description:
 *
 * @since 2024-07-20
 */
public class InvokeException extends RuntimeException {
    private int errorCode;
    private String message;

    /**
     * Instantiates a new InvokeException exception.
     *
     * @param errorCode the error code
     */
    public InvokeException(int errorCode) {
        this(errorCode, null, null);
    }

    /**
     * Instantiates a new InvokeException exception.
     *
     * @param errorCode the error code
     * @param message   the message
     */
    public InvokeException(int errorCode, String message) {
        this(errorCode, message, null);
    }

    /**
     * Instantiates a new InvokeException exception.
     *
     * @param errorCode the error code
     * @param message   the message
     * @param cause     the cause
     */
    public InvokeException(int errorCode, String message, Throwable cause) {
        super(cause);
        this.errorCode = errorCode;
        this.message = message;
    }

    /**
     * Gets error code.
     *
     * @return the error code
     */
    public int getErrorCode() {
        return errorCode;
    }

    /**
     * Sets error code.
     *
     * @param errorCode the error code
     */
    public void setErrorCode(int errorCode) {
        this.errorCode = errorCode;
    }

    /**
     * Gets message.
     *
     * @return the message
     */
    @Override
    public String getMessage() {
        return message;
    }

    /**
     * Sets message.
     *
     * @param message the message
     */
    public void setMessage(String message) {
        this.message = message;
    }

    /**
     * To string string.
     *
     * @return the string
     */
    @Override
    public String toString() {
        return "{" + "\"code\":\"" + errorCode + "\", \"message\":\"" + message + "\"}";
    }
}
