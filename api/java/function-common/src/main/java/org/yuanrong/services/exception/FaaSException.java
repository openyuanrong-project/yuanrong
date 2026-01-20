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

package org.yuanrong.services.exception;

import java.util.Objects;

/**
 * The faaS executor exception
 *
 * @since 2024-07-06
 */
public class FaaSException extends Exception {
    private final String errorMessage;

    /**
     * constructor
     *
     * @param message message
     * @param exception exception
     */
    public FaaSException(String message, Throwable exception) {
        super(exception);
        this.errorMessage = message;
    }

    /**
     * constructor
     *
     * @param message message
     */
    public FaaSException(String message) {
        this.errorMessage = message;
    }

    /**
     * getMessage
     *
     * @return errorMessage
     */
    public String getMessage() {
        return this.errorMessage;
    }

    /**
     * equals
     *
     * @param obj the Object
     * @return boolean
     */
    @Override
    public boolean equals(Object obj) {
        if (this == obj) {
            return true;
        }
        if (obj == null || getClass() != obj.getClass()) {
            return false;
        }
        FaaSException that = (FaaSException) obj;
        return errorMessage.equals(that.errorMessage);
    }

    /**
     * hashCode
     *
     * @return hash value
     */
    @Override
    public int hashCode() {
        return Objects.hash(errorMessage);
    }
}
