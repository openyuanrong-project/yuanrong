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

#include <signal.h>
#include <fstream>
#include <string>

#include "common.h"
#include "counter.h"
#include "fileutils.h"
#include "spd_logger.h"

namespace YR {
namespace utility {
void InitLog(const LogParam &logParam);
void InstallFailureSignalHandler(const char *arg0);
void FailureSignalWriter(const char *data);
void SetGetLoggerNameFunc(GetLoggerNameFunc func);

// DON'T need to know whether to write files or whether to stdout
#define YRLOG_TRACE(format, ...) YRLOG_ASYNC_WRAPPER("TRACE", format, ##__VA_ARGS__)
#define YRLOG_DEBUG(format, ...) YRLOG_ASYNC_WRAPPER("DEBUG", format, ##__VA_ARGS__)
#define YRLOG_INFO(format, ...) YRLOG_ASYNC_WRAPPER("INFO", format, ##__VA_ARGS__)
#define YRLOG_WARN(format, ...) YRLOG_ASYNC_WRAPPER("WARN", format, ##__VA_ARGS__)
#define YRLOG_ERROR(format, ...) YRLOG_ASYNC_WRAPPER("ERR", format, ##__VA_ARGS__)
#define YRLOG_FATAL(format, ...) YRLOG_ASYNC_WRAPPER("FATAL", format, ##__VA_ARGS__)
#define YRLOG_DEBUG_COUNT_60(...) YRLOG_DEBUG_COUNT(60, __VA_ARGS__)
#define YRLOG_DEBUG_COUNT(frequent, ...)                                                          \
    do {                                                                                          \
        if (YR::utility::SpdLogger::GetInstance().level() <= YR::utility::GetLogLevel("DEBUG")) { \
            static YR::utility::Counter counter(frequent);                                        \
            if (counter.Proc()) {                                                                 \
                YRLOG_DEBUG(__VA_ARGS__);                                                         \
            }                                                                                     \
        }                                                                                         \
    } while (false)

// write log if expression is true
#define YRLOG_DEBUG_IF(expr, ...)     \
    do {                              \
        if ((expr)) {                 \
            YRLOG_DEBUG(__VA_ARGS__); \
        }                             \
    } while (false)

inline void KillProcess(const std::string &ret)
{
    YRLOG_ERROR("Function System Exit Tip: {}", ret);

    // flush the log in cache to disk before exiting.
    SpdLogger::GetInstance().Flush();
    (void)raise(SIGKILL);
}

#define FS_EXIT(ret)                                                                \
    do {                                                                            \
        std::stringstream ss;                                                       \
        ss << (ret) << "  ( file: " << __FILE__ << ", line: " << __LINE__ << " )."; \
        KillProcess(ss.str());                                                      \
    } while (0)

#define EXIT_IF_NULL(ptr)                                          \
    {                                                              \
        if ((ptr) == nullptr) {                                    \
            YRLOG_ERROR("ptr{} null, will exit", #ptr);            \
            FS_EXIT("Exit for Bad alloc or Dynamic cast failed."); \
        }                                                          \
    }

#define EXIT_IF_FALSE(condition, ...) \
    {                                 \
        if (!condition) {             \
            YRLOG_FATAL(__VA_ARGS__); \
        }                             \
    }
}  // namespace utility
}  // namespace YR
