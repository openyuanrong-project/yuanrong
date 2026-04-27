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

import java.util.Arrays;
import java.util.Objects;

/**
 * The type Handler not available exception.
 * <p>
 * * @since 2022 -07-20
 *
 * @since 2022 /09/26
 */
public class HandlerNotAvailableException extends Exception {
    /**
     * The Error message.
     */
    private final String errorMessage;

    /**
     * Instantiates a new Handler not available exception.
     *
     * @param exception the exception
     */
    public HandlerNotAvailableException(Exception exception) {
        this.errorMessage = getRealMessage(exception);
    }

    @Override
    public String getMessage() {
        return this.errorMessage;
    }

    /**
     * Gets real message.
     *
     * @param err the err
     * @return the real message
     */
    private static String getRealMessage(Throwable err) {
        StringBuilder errMeaasge = new StringBuilder();
        Throwable exp = err;
        while (exp != null) {
            Throwable cause = exp.getCause();
            if (cause == null) {
                return errMeaasge.append(exp.getClass()).append(' ').append(exp.getMessage()).append(". ")
                        .append(Arrays.toString(exp.getStackTrace())).toString();
            }
            exp = cause;
        }
        return errMeaasge.toString();
    }

    @Override
    public boolean equals(Object obj) {
        if (this == obj) {
            return true;
        }
        if (obj == null || getClass() != obj.getClass()) {
            return false;
        }
        HandlerNotAvailableException that = (HandlerNotAvailableException) obj;
        return Objects.equals(errorMessage, that.errorMessage);
    }

    @Override
    public int hashCode() {
        return Objects.hash(errorMessage);
    }
}
