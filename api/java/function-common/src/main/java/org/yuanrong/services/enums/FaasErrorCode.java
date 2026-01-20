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

package org.yuanrong.services.enums;

/**
 * The faaS error code
 *
 * @since 2024-07-20
 */
public enum FaasErrorCode {
    NONE_ERROR(0, "no error"),

    FAAS_ERROR(500, "error in faaS executor"),

    FAAS_INIT_ERROR(6001, "illegal argument exception"),

    ENTRY_NOT_FOUND(4001, "user entry not found"),

    FUNCTION_RUN_ERROR(4002, "user function failed to run"),

    RESPONSE_EXCEED_LIMIT(4004, "response body size exceeds the limit of 6291456"),

    INITIALIZE_FUNCTION_ERROR(4009, "function initialization exception"),

    INVOKE_FUNCTION_TIMEOUT(4010, "invoke timed out"),

    INIT_FUNCTION_FAIL(4201, "function initialization failed"),

    INIT_FUNCTION_TIMEOUT(4211, "runtime initialization timed out"),

    REQUEST_BODY_EXCEED_LIMIT(4140, "request body exceeds limit");

    private final int code;

    private final String errorMessage;

    FaasErrorCode(int code, String errorMessage) {
        this.code = code;
        this.errorMessage = errorMessage;
    }

    /**
     * getCode
     *
     * @return code
     */
    public int getCode() {
        return code;
    }

    /**
     * getErrorMessage
     *
     * @return errorMessage
     */
    public String getErrorMessage() {
        return errorMessage;
    }
}
