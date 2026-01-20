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

/**
 * Description: Function Error Code
 *
 * @since 2022-10-18
 */
public enum FuncErrorCode {
    ERR_DATA_SYSTEM_EXCEPTION(1000, "data system call exception"),
    ERR_GET_DATA_TIMEOUT(1001, "get data timeout"),
    ERR_GET_DATA_INTERRUPT(1002, "get data interrupted"),
    ERR_OBJECT_REF_IS_NULL(1003, "object ref is null"),
    ERR_OBJECT_REF_LIST_IS_NULL(1004, "object ref list is null"),
    ERR_EXECUTE_REQUEST_ERROR(1005, "fail to execute request"),
    ERR_PARSE_INVOKE_RESPONSE_ERROR(1006, "fail to parse invoke response"),
    ERR_INVOKE_FUNCTION_ERR(1007, "invoke function error"),
    ERR_TIMEOUT_INVALID(1008, "timout is invalid, it needs greater than 0"),
    ERR_PUT_INTERVAL_INVALID(1009, "put interval is invalid, it must be a non negative integer."),
    ERR_PUT_RETRY_TIME_INVALID(1010, "put retryTime is invalid, it must be a non negative integer.");

    private final int value;
    private final String desc;

    FuncErrorCode(int value, String desc) {
        this.value = value;
        this.desc = desc;
    }

    public int getValue() {
        return value;
    }

    public String getDesc() {
        return desc;
    }
}
