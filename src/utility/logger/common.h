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

#include <string>

namespace YR {
namespace utility {
const int DEFAULT_MAX_SIZE = 100;  // 100 MB
const uint32_t DEFAULT_MAX_FILES = 3;
const int DEFAULT_RETENTION_DAYS = 30;
const int64_t SIZE_MEGA_BYTES = 1024 * 1024;  // 1 MB
const int DEFAULT_LOG_BUF_SECONDS = 10;
const std::string LOGGER_NAME = "Logger";
const std::string REPORT_LOGGER_NAME = "interface";
const std::string DEFAULT_JOB_ID = "job-ffffffff";
const std::string LOG_SUFFIX = ".log";
const unsigned int DEFAULT_MAX_ASYNC_QUEUE_SIZE = 1024;  // 1024(every log length)
const unsigned int DEFAULT_ASYNC_THREAD_COUNT = 1;

struct LogParam {
    std::string logLevel;
    std::string logDir;
    std::string nodeName;
    std::string modelName;
    int maxSize = DEFAULT_MAX_SIZE;
    uint32_t maxFiles = DEFAULT_MAX_FILES;
    int retentionDays = DEFAULT_RETENTION_DAYS;
    bool logFileWithTime = false;
    int logBufSecs = DEFAULT_LOG_BUF_SECONDS;
    unsigned int maxAsyncQueueSize = DEFAULT_MAX_ASYNC_QUEUE_SIZE;
    unsigned int asyncThreadCount = DEFAULT_ASYNC_THREAD_COUNT;
    bool alsoLog2Stderr = false;
    bool isLogMerge = false;
    bool withLogPrefix = false;
    bool onlyStdout = false;
    std::string loggerId;
    bool useUtcTime = false;  // default local time, consistent with common/logs config
};
}  // namespace utility
}  // namespace YR
