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

package org.yuanrong.errorcode;

/**
 * The enum Error code.
 *
 * @since 2023/11/6
 */
public class ErrorCode {
    /**
     * ERR_OK
     */
    public static final ErrorCode ERR_OK = new ErrorCode(0);

    /**
     * ERR_PARAM_INVALID
     */
    public static final ErrorCode ERR_PARAM_INVALID = new ErrorCode(1001);

    /**
     * ERR_RESOURCE_NOT_ENOUGH
     */
    public static final ErrorCode ERR_RESOURCE_NOT_ENOUGH = new ErrorCode(1002);

    /**
     * ERR_INSTANCE_NOT_FOUND
     */
    public static final ErrorCode ERR_INSTANCE_NOT_FOUND = new ErrorCode(1003);

    /**
     * ERR_INSTANCE_DUPLICATED
     */
    public static final ErrorCode ERR_INSTANCE_DUPLICATED = new ErrorCode(1004);

    /**
     * ERR_INVOKE_RATE_LIMITED
     */
    public static final ErrorCode ERR_INVOKE_RATE_LIMITED = new ErrorCode(1005);

    /**
     * ERR_RESOURCE_CONFIG_ERROR
     */
    public static final ErrorCode ERR_RESOURCE_CONFIG_ERROR = new ErrorCode(1006);

    /**
     * ERR_INSTANCE_EXITED
     */
    public static final ErrorCode ERR_INSTANCE_EXITED = new ErrorCode(1007);

    /**
     * ERR_EXTENSION_META_ERROR
     */
    public static final ErrorCode ERR_EXTENSION_META_ERROR = new ErrorCode(1008);

    /**
     * ERR_USER_CODE_LOAD
     */
    public static final ErrorCode ERR_USER_CODE_LOAD = new ErrorCode(2001);

    /**
     * ERR_USER_FUNCTION_EXCEPTION
     */
    public static final ErrorCode ERR_USER_FUNCTION_EXCEPTION = new ErrorCode(2002);

    /**
     * ERR_REQUEST_BETWEEN_RUNTIME_BUS
     */
    public static final ErrorCode ERR_REQUEST_BETWEEN_RUNTIME_BUS = new ErrorCode(3001);

    /**
     * ERR_INNER_COMMUNICATION
     */
    public static final ErrorCode ERR_INNER_COMMUNICATION = new ErrorCode(3002);

    /**
     * ERR_INNER_SYSTEM_ERROR
     */
    public static final ErrorCode ERR_INNER_SYSTEM_ERROR = new ErrorCode(3003);

    /**
     * ERR_DISCONNECT_FRONTEND_BUS
     */
    public static final ErrorCode ERR_DISCONNECT_FRONTEND_BUS = new ErrorCode(3004);

    /**
     * ERR_ETCD_OPERATION_ERROR
     */
    public static final ErrorCode ERR_ETCD_OPERATION_ERROR = new ErrorCode(3005);

    /**
     * ERR_BUS_DISCONNECTION
     */
    public static final ErrorCode ERR_BUS_DISCONNECTION = new ErrorCode(3006);

    /**
     * ERR_REDIS_OPERATION_ERROR
     */
    public static final ErrorCode ERR_REDIS_OPERATION_ERROR = new ErrorCode(3007);

    /**
     * ERR_INCORRECT_INIT_USAGE
     */
    public static final ErrorCode ERR_INCORRECT_INIT_USAGE = new ErrorCode(4001);

    /**
     * ERR_INIT_CONNECTION_FAILED
     */
    public static final ErrorCode ERR_INIT_CONNECTION_FAILED = new ErrorCode(4002);

    /**
     * ERR_DESERIALIZATION_FAILED
     */
    public static final ErrorCode ERR_DESERIALIZATION_FAILED = new ErrorCode(4003);

    /**
     * ERR_INSTANCE_ID_EMPTY
     */
    public static final ErrorCode ERR_INSTANCE_ID_EMPTY = new ErrorCode(4004);

    /**
     * ERR_GET_OPERATION_FAILED
     */
    public static final ErrorCode ERR_GET_OPERATION_FAILED = new ErrorCode(4005);

    /**
     * ERR_INCORRECT_FUNCTION_USAGE
     */
    public static final ErrorCode ERR_INCORRECT_FUNCTION_USAGE = new ErrorCode(4006);

    /**
     * ERR_INCORRECT_CREATE_USAGE
     */
    public static final ErrorCode ERR_INCORRECT_CREATE_USAGE = new ErrorCode(4007);

    /**
     * ERR_INCORRECT_INVOKE_USAGE
     */
    public static final ErrorCode ERR_INCORRECT_INVOKE_USAGE = new ErrorCode(4008);

    /**
     * ERR_INCORRECT_KILL_USAGE
     */
    public static final ErrorCode ERR_INCORRECT_KILL_USAGE = new ErrorCode(4009);

    /**
     * ERR_PARSE_INVOKE_RESPONSE_ERROR
     */
    public static final ErrorCode ERR_PARSE_INVOKE_RESPONSE_ERROR = new ErrorCode(4010);

    /**
     * ERR_ROCKSDB_FAILED
     */
    public static final ErrorCode ERR_ROCKSDB_FAILED = new ErrorCode(4201);

    /**
     * ERR_SHARED_MEMORY_LIMITED
     */
    public static final ErrorCode ERR_SHARED_MEMORY_LIMITED = new ErrorCode(4202);

    /**
     * ERR_OPERATE_DISK_FAILED
     */
    public static final ErrorCode ERR_OPERATE_DISK_FAILED = new ErrorCode(4203);

    /**
     * ERR_INSUFFICIENT_DISK_SPACE
     */
    public static final ErrorCode ERR_INSUFFICIENT_DISK_SPACE = new ErrorCode(4204);

    /**
     * ERR_CONNECTION_FAILED
     */
    public static final ErrorCode ERR_CONNECTION_FAILED = new ErrorCode(4205);

    /**
     * ERR_KEY_ALREADY_EXIST
     */
    public static final ErrorCode ERR_KEY_ALREADY_EXIST = new ErrorCode(4206);

    /**
     * ERR_DEPENDENCY_FAILED
     */
    public static final ErrorCode ERR_DEPENDENCY_FAILED = new ErrorCode(4306);

    /**
     *
     */
    public static final ErrorCode ERR_DATASYSTEM_FAILED = new ErrorCode(4299);

    /**
     * ERR_FINALIZED
     */
    public static final ErrorCode ERR_FINALIZED = new ErrorCode(9000);

    /**
     * ERR_CREATE_RETURN_BUFFER
     */
    public static final ErrorCode ERR_CREATE_RETURN_BUFFER = new ErrorCode(9001);

    /**
     * ERR_SESSION_TIMEOUT
     */
    public static final ErrorCode ERR_SESSION_TIMEOUT = new ErrorCode(5001);

    /**
     * ERR_SESSION_INTERRUPTED
     */
    public static final ErrorCode ERR_SESSION_INTERRUPTED = new ErrorCode(5002);

    /**
     * ERR_SESSION_NOT_WAITING
     */
    public static final ErrorCode ERR_SESSION_NOT_WAITING = new ErrorCode(5003);

    // ErrorCode provided for the jobExecutor interface:

    /**
     * ERR_JOB_USER_CODE_EXCEPTION
     */
    public static final ErrorCode ERR_JOB_USER_CODE_EXCEPTION = new ErrorCode(50001);

    /**
     * ERR_JOB_RUNTIME_EXCEPTION
     */
    public static final ErrorCode ERR_JOB_RUNTIME_EXCEPTION = new ErrorCode(50002);

    /**
     * ERR_JOB_INNER_SYSTEM_EXCEPTION
     */
    public static final ErrorCode ERR_JOB_INNER_SYSTEM_EXCEPTION = new ErrorCode(50003);


    private int code;

    /**
     * Init ErrorCode
     *
     * @param code the code
     */
    public ErrorCode(int code) {
        this.code = code;
    }

    /**
     * Gets code.
     *
     * @return the code
     */
    public int getValue() {
        return this.code;
    }

    /**
     * toString
     *
     * @return String
     */
    @Override
    public String toString() {
        return String.valueOf(this.code);
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
        if (obj == null || this.getClass() != obj.getClass()) {
            return false;
        }
        ErrorCode person = (ErrorCode) obj;
        return this.code == person.code;
    }

    /**
     * hashCode
     *
     * @return code
     */
    @Override
    public int hashCode() {
        return this.code;
    }
};
