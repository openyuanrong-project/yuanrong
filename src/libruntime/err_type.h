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

#pragma once
#include <iostream>
#include <sstream>
#include <string>
#include <unordered_map>
#include <vector>

#include "src/libruntime/stacktrace/stack_trace_info.h"

namespace YR {
namespace Libruntime {
enum ErrorCode : int {
    ERR_OK = 0,

    ERR_PARAM_INVALID = 1001,
    ERR_RESOURCE_NOT_ENOUGH = 1002,
    ERR_INSTANCE_NOT_FOUND = 1003,
    ERR_INSTANCE_DUPLICATED = 1004,
    ERR_INVOKE_RATE_LIMITED = 1005,
    ERR_RESOURCE_CONFIG_ERROR = 1006,
    ERR_INSTANCE_EXITED = 1007,
    ERR_EXTENSION_META_ERROR = 1008,
    ERR_INSTANCE_SUB_HEALTH = 1009,
    ERR_GROUP_SCHEDULE_FAILED = 1010,
    ERR_INSTANCE_EVICTED = 1013,
    ERR_FUNCTION_META_NOT_FOUND = 1015,

    ERR_INSTANCE_SUSPEND = 1020,
    ERR_INSTANCE_BUSY = 1021,
    ERR_USER_CODE_LOAD = 2001,
    ERR_USER_FUNCTION_EXCEPTION = 2002,

    ERR_REQUEST_BETWEEN_RUNTIME_BUS = 3001,
    ERR_INNER_COMMUNICATION = 3002,
    ERR_INNER_SYSTEM_ERROR = 3003,
    ERR_DISCONNECT_FRONTEND_BUS = 3004,
    ERR_ETCD_OPERATION_ERROR = 3005,
    ERR_BUS_DISCONNECTION = 3006,
    ERR_REDIS_OPERATION_ERROR = 3007,
    ERR_REQUEST_BETWEEN_RUNTIME_FRONTEND = 3008,

    ERR_INCORRECT_INIT_USAGE = 4001,
    ERR_INIT_CONNECTION_FAILED = 4002,
    ERR_DESERIALIZATION_FAILED = 4003,
    ERR_INSTANCE_ID_EMPTY = 4004,
    ERR_GET_OPERATION_FAILED = 4005,
    ERR_INCORRECT_FUNCTION_USAGE = 4006,
    ERR_INCORRECT_CREATE_USAGE = 4007,
    ERR_INCORRECT_INVOKE_USAGE = 4008,
    ERR_INCORRECT_KILL_USAGE = 4009,

    ERR_ROCKSDB_FAILED = 4201,
    ERR_SHARED_MEMORY_LIMITED = 4202,
    ERR_OPERATE_DISK_FAILED = 4203,
    ERR_INSUFFICIENT_DISK_SPACE = 4204,
    ERR_CONNECTION_FAILED = 4205,
    ERR_KEY_ALREADY_EXIST = 4206,
    ERR_CLIENT_ALREADY_CLOSED = 4298,
    ERR_DATASYSTEM_FAILED = 4299,
    ERR_DEPENDENCY_FAILED = 4306,

    ERR_ACQUIRE_TIMEOUT = 6038,

    ERR_FINALIZED = 9000,
    ERR_CREATE_RETURN_BUFFER = 9001,
    ERR_HEALTH_CHECK_HEALTHY = 9002,
    ERR_HEALTH_CHECK_FAILED = 9003,
    ERR_HEALTH_CHECK_SUBHEALTH = 9004,
    ERR_GENERATOR_FINISHED = 9005,
    ERR_FUNCTION_MASTER_NOT_CONFIGURED = 9006,
    ERR_FUNCTION_MASTER_TIMEOUT = 9007,
    ERR_CLIENT_TERMINAL_KILLED = 9008,
    ERR_ALL_SCHEDULER_UNAVALIABLE = 9009,
    ERR_REQ_TIMEOUT_WITHOUT_ACK = 9010,

    ERR_SESSION_TIMEOUT = 5001,
    ERR_SESSION_INTERRUPTED = 5002,
    ERR_SESSION_NOT_WAITING = 5003,
};

const static std::unordered_map<ErrorCode, std::string> errCodeToString = {
    {ErrorCode::ERR_OK, "ERR_OK"},
    {ErrorCode::ERR_PARAM_INVALID, "ERR_PARAM_INVALID"},
    {ErrorCode::ERR_RESOURCE_NOT_ENOUGH, "ERR_RESOURCE_NOT_ENOUGH"},
    {ErrorCode::ERR_INSTANCE_NOT_FOUND, "ERR_INSTANCE_NOT_FOUND"},
    {ErrorCode::ERR_INSTANCE_DUPLICATED, "ERR_INSTANCE_DUPLICATED"},
    {ErrorCode::ERR_INVOKE_RATE_LIMITED, "ERR_INVOKE_RATE_LIMITED"},
    {ErrorCode::ERR_RESOURCE_CONFIG_ERROR, "ERR_RESOURCE_CONFIG_ERROR"},
    {ErrorCode::ERR_INSTANCE_EXITED, "ERR_INSTANCE_EXITED"},
    {ErrorCode::ERR_EXTENSION_META_ERROR, "ERR_EXTENSION_META_ERROR"},
    {ErrorCode::ERR_INSTANCE_SUB_HEALTH, "ERR_INSTANCE_SUB_HEALTH"},
    {ErrorCode::ERR_GROUP_SCHEDULE_FAILED, "ERR_GROUP_SCHEDULE_FAILED"},
    {ErrorCode::ERR_INSTANCE_EVICTED, "ERR_INSTANCE_EVICTED"},
    {ErrorCode::ERR_FUNCTION_META_NOT_FOUND, "ERR_FUNCTION_META_NOT_FOUND"},
    {ErrorCode::ERR_USER_CODE_LOAD, "ERR_USER_CODE_LOAD"},
    {ErrorCode::ERR_USER_FUNCTION_EXCEPTION, "ERR_USER_FUNCTION_EXCEPTION"},
    {ErrorCode::ERR_REQUEST_BETWEEN_RUNTIME_BUS, "ERR_REQUEST_BETWEEN_RUNTIME_BUS"},
    {ErrorCode::ERR_INNER_COMMUNICATION, "ERR_INNER_COMMUNICATION"},
    {ErrorCode::ERR_INNER_SYSTEM_ERROR, "ERR_INNER_SYSTEM_ERROR"},
    {ErrorCode::ERR_DISCONNECT_FRONTEND_BUS, "ERR_DISCONNECT_FRONTEND_BUS"},
    {ErrorCode::ERR_ETCD_OPERATION_ERROR, "ERR_ETCD_OPERATION_ERROR"},
    {ErrorCode::ERR_BUS_DISCONNECTION, "ERR_BUS_DISCONNECTION"},
    {ErrorCode::ERR_REQUEST_BETWEEN_RUNTIME_FRONTEND, "ERR_REQUEST_BETWEEN_RUNTIME_FRONTEND"},
    {ErrorCode::ERR_REDIS_OPERATION_ERROR, "ERR_REDIS_OPERATION_ERROR"},
    {ErrorCode::ERR_INCORRECT_INIT_USAGE, "ERR_INCORRECT_INIT_USAGE"},
    {ErrorCode::ERR_INIT_CONNECTION_FAILED, "ERR_INIT_CONNECTION_FAILED"},
    {ErrorCode::ERR_DESERIALIZATION_FAILED, "ERR_DESERIALIZATION_FAILED"},
    {ErrorCode::ERR_INSTANCE_ID_EMPTY, "ERR_INSTANCE_ID_EMPTY"},
    {ErrorCode::ERR_GET_OPERATION_FAILED, "ERR_GET_OPERATION_FAILED"},
    {ErrorCode::ERR_INCORRECT_FUNCTION_USAGE, "ERR_INCORRECT_FUNCTION_USAGE"},
    {ErrorCode::ERR_INCORRECT_CREATE_USAGE, "ERR_INCORRECT_CREATE_USAGE"},
    {ErrorCode::ERR_INCORRECT_INVOKE_USAGE, "ERR_INCORRECT_INVOKE_USAGE"},
    {ErrorCode::ERR_INCORRECT_KILL_USAGE, "ERR_INCORRECT_KILL_USAGE"},
    {ErrorCode::ERR_ROCKSDB_FAILED, "ERR_ROCKSDB_FAILED"},
    {ErrorCode::ERR_SHARED_MEMORY_LIMITED, "ERR_SHARED_MEMORY_LIMITED"},
    {ErrorCode::ERR_OPERATE_DISK_FAILED, "ERR_OPERATE_DISK_FAILED"},
    {ErrorCode::ERR_INSUFFICIENT_DISK_SPACE, "ERR_INSUFFICIENT_DISK_SPACE"},
    {ErrorCode::ERR_CONNECTION_FAILED, "ERR_CONNECTION_FAILED"},
    {ErrorCode::ERR_KEY_ALREADY_EXIST, "ERR_KEY_ALREADY_EXIST"},
    {ErrorCode::ERR_CLIENT_ALREADY_CLOSED, "ERR_CLIENT_ALREADY_CLOSED"},
    {ErrorCode::ERR_DATASYSTEM_FAILED, "ERR_DATASYSTEM_FAILED"},
    {ErrorCode::ERR_DEPENDENCY_FAILED, "ERR_DEPENDENCY_FAILED"},
    {ErrorCode::ERR_ACQUIRE_TIMEOUT, "ERR_ACQUIRE_TIMEOUT"},
    {ErrorCode::ERR_FINALIZED, "ERR_FINALIZED"},
    {ErrorCode::ERR_CREATE_RETURN_BUFFER, "ERR_CREATE_RETURN_BUFFER"},
    {ErrorCode::ERR_HEALTH_CHECK_HEALTHY, "ERR_HEALTH_CHECK_HEALTHY"},
    {ErrorCode::ERR_HEALTH_CHECK_FAILED, "ERR_HEALTH_CHECK_FAILED"},
    {ErrorCode::ERR_HEALTH_CHECK_SUBHEALTH, "ERR_HEALTH_CHECK_SUBHEALTH"},
    {ErrorCode::ERR_GENERATOR_FINISHED, "ERR_GENERATOR_FINISHED"},
    {ErrorCode::ERR_FUNCTION_MASTER_NOT_CONFIGURED, "ERR_FUNCTION_MASTER_NOT_CONFIGURED"},
    {ErrorCode::ERR_FUNCTION_MASTER_TIMEOUT, "ERR_FUNCTION_MASTER_TIMEOUT"},
    {ErrorCode::ERR_CLIENT_TERMINAL_KILLED, "ERR_CLIENT_TERMINAL_KILLED"},
    {ErrorCode::ERR_SESSION_TIMEOUT, "ERR_SESSION_TIMEOUT"},
    {ErrorCode::ERR_SESSION_INTERRUPTED, "ERR_SESSION_INTERRUPTED"},
    {ErrorCode::ERR_SESSION_NOT_WAITING, "ERR_SESSION_NOT_WAITING"}};

const static std::unordered_map<uint32_t, ErrorCode> datasystemErrCodeMap = {
    {1, ErrorCode::ERR_PARAM_INVALID},          {2, ErrorCode::ERR_PARAM_INVALID},
    {3, ErrorCode::ERR_GET_OPERATION_FAILED},   {4, ErrorCode::ERR_ROCKSDB_FAILED},
    {5, ErrorCode::ERR_DATASYSTEM_FAILED},      {6, ErrorCode::ERR_SHARED_MEMORY_LIMITED},
    {7, ErrorCode::ERR_OPERATE_DISK_FAILED},    {8, ErrorCode::ERR_DATASYSTEM_FAILED},
    {10, ErrorCode::ERR_GET_OPERATION_FAILED},  {13, ErrorCode::ERR_INSUFFICIENT_DISK_SPACE},
    {25, ErrorCode::ERR_DATASYSTEM_FAILED},     {1000, ErrorCode::ERR_INNER_COMMUNICATION},
    {1001, ErrorCode::ERR_INNER_COMMUNICATION}, {1002, ErrorCode::ERR_INNER_COMMUNICATION},
    {2004, ErrorCode::ERR_KEY_ALREADY_EXIST},   {3009, ErrorCode::ERR_CLIENT_ALREADY_CLOSED},
};

enum ModuleCode : int {
    CORE = 10,
    RUNTIME = 20,
    RUNTIME_CREATE = 21,
    RUNTIME_INVOKE = 22,
    RUNTIME_KILL = 23,
    DATASYSTEM = 30
};

class ErrorInfo {
public:
    ErrorInfo() = default;
    ~ErrorInfo() = default;

    ErrorInfo(const ErrorInfo &err)
        : code(err.Code()),
          mCode(err.MCode()),
          msg(err.Msg()),
          isCreate(err.IsCreate()),
          isTimeout(err.IsTimeout()),
          isAckTimeout(err.IsAckTimeout()),
          stackTraceInfos_(std::move(err.GetStackTraceInfos())),
          dsStatusCode_(err.GetDsStatusCode())
    {
    }
    ErrorInfo &operator=(const ErrorInfo &err)
    {
        code = err.Code();
        mCode = err.MCode();
        msg = err.Msg();
        isCreate = err.IsCreate();
        isTimeout = err.IsTimeout();
        isAckTimeout = err.IsAckTimeout();
        stackTraceInfos_ = std::move(err.GetStackTraceInfos());
        dsStatusCode_ = err.GetDsStatusCode();
        return *this;
    }

    ErrorInfo(ErrorCode errCode, const std::string &errMsg) : code(errCode), msg(errMsg) {}
    ErrorInfo(ErrorCode errCode, const std::string &errMsg, bool isAckTimeoutInput)
        : code(errCode), msg(errMsg), isAckTimeout(isAckTimeoutInput)
    {
    }
    ErrorInfo(ErrorCode errCode, ModuleCode moduleCode, const std::string &errMsg)
        : code(errCode), mCode(moduleCode), msg(errMsg)
    {
    }
    ErrorInfo(ErrorCode errCode, ModuleCode moduleCode, const std::string &errMsg,
              const std::vector<StackTraceInfo> &stackTraceInfos)
        : code(errCode), mCode(moduleCode), msg(errMsg), stackTraceInfos_(std::move(stackTraceInfos))
    {
    }
    ErrorInfo(ErrorCode errCode, ModuleCode moduleCode, const std::string &errMsg, bool isCreateInput)
        : code(errCode), mCode(moduleCode), msg(errMsg), isCreate(isCreateInput)
    {
    }
    ErrorInfo(ErrorCode errCode, ModuleCode moduleCode, const std::string &errMsg, bool isCreateInput,
              const std::vector<StackTraceInfo> &stackTraceInfos)
        : code(errCode),
          mCode(moduleCode),
          msg(errMsg),
          isCreate(isCreateInput),
          stackTraceInfos_(std::move(stackTraceInfos))
    {
    }
    ErrorCode Code() const
    {
        return code;
    }
    ModuleCode MCode() const
    {
        return mCode;
    }
    std::string Msg() const
    {
        return msg;
    }

    std::string CodeAndMsg() const
    {
        return "ErrCode: " + std::to_string(code) + ", ModuleCode: " + std::to_string(mCode) + ", ErrMsg: " + msg;
    }

    void SetErrorCode(ErrorCode errCode)
    {
        code = errCode;
    }
    void SetErrorMsg(const std::string &errMsg)
    {
        msg = errMsg;
    }
    void SetErrCodeAndMsg(const ErrorCode &errCode, const ModuleCode &moduleCode, const std::string &errMsg,
                          const int dsStatusCode = 0)
    {
        code = errCode;
        mCode = moduleCode;
        msg = errMsg;
        dsStatusCode_ = dsStatusCode;
    }

    int GetDsStatusCode() const
    {
        return dsStatusCode_;
    }

    void SetDsStatusCode(int dsStatusCode)
    {
        dsStatusCode_ = dsStatusCode;
    }

    std::string GetExceptionMsg(const std::vector<std::string> &failIds, int timeoutMs) const
    {
        std::ostringstream oss;
        if (msg.empty()) {
            oss << "Get timeout " << timeoutMs << "ms from datasystem, ";
        } else {
            oss << msg;
        }
        oss << " partial failed: "
            << "(" << failIds.size() << "). ";
        oss << "Failed objects: [ ";
        oss << failIds[0] << " ... "
            << "]";
        return oss.str();
    }

    bool OK() const
    {
        return code == ErrorCode::ERR_OK;
    }

    bool operator==(const ErrorInfo &other) const
    {
        return code == other.code && mCode == other.mCode && msg == other.msg;
    }

    bool Finalized()
    {
        return code == ErrorCode::ERR_FINALIZED;
    }

    bool IsCreate() const
    {
        return isCreate;
    }

    void SetIsCreate(bool isCreateInput)
    {
        isCreate = isCreateInput;
    }

    bool IsTimeout() const
    {
        return isTimeout;
    }

    bool IsAckTimeout() const
    {
        return isAckTimeout;
    }

    void SetIsTimeout(bool isTimeoutInput)
    {
        isTimeout = isTimeoutInput;
    }

    void SetIsAckTimeout(bool isTimeoutInput)
    {
        isAckTimeout = isTimeoutInput;
    }

    std::vector<StackTraceInfo> GetStackTraceInfos() const
    {
        return stackTraceInfos_;
    }

    void SetStackTraceInfos(std::vector<StackTraceInfo> &stackTraceInfos)
    {
        stackTraceInfos_ = std::move(stackTraceInfos);
    }

private:
    ErrorCode code = ErrorCode::ERR_OK;
    ModuleCode mCode = ModuleCode::RUNTIME;
    std::string msg;
    bool isCreate = false;
    bool isTimeout = false;  // This information is used to exclude the timeout error when the get operation fails due
                             // to exception IDs.
    bool isAckTimeout = false;
    std::vector<StackTraceInfo> stackTraceInfos_;
    int dsStatusCode_{0};
};
}  // namespace Libruntime
}  // namespace YR