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

package org.yuanrong.function.common;

/**
 * The response error code
 *
 * @since 2024/07/02
 */
public enum RspErrorCode {
    GET_TIMEOUT_ERROR(133, "Get timeout failed"),
    USER_STATE_LARGE_ERROR(4003, "state content is too large"),
    USER_STATE_UNDEFINED_ERROR(4005, "state is undefined"),
    INVALID_PARAMETER(4040, "invalid input parameter"),
    INTERNAL_ERROR(110500, "internal system error");

    private static final String TIMEOUT_ERROR_CODE = "211408";
    private static final String RUNTIME_INVOKE_TIMEOUT_CODE = "4010";
    private static final String SLA_ERROR_CODE = "216001";
    private static final String QUEUE_TIMEOUT_CODE = "150430";
    private static final String RUNTIME_ERROR_PREFIX = "4";
    private int errorCode;
    private String desc;

    private RspErrorCode(int errorCode, String desc) {
        this.errorCode = errorCode;
        this.desc = desc;
    }

    /**
     * getErrorCode
     *
     * @return errorCode
     */
    public int getErrorCode() {
        return this.errorCode;
    }

    /**
     * getDesc
     *
     * @return desc
     */
    public String getDesc() {
        return this.desc;
    }
}