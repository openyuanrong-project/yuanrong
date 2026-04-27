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
#include <spdlog/sinks/rotating_file_sink.h>
#include <spdlog/sinks/stdout_color_sinks.h>
#include <spdlog/spdlog.h>

#include <unistd.h>
#include <map>

#include "absl/synchronization/mutex.h"

#include "common.h"
#include "src/utility/id_generator.h"
#include "src/utility/macro.h"
#include "src/utility/singleton.h"

namespace YR {
namespace utility {
extern const std::string DEFAULT_LOG_NAME;
yr_spdlog::level::level_enum GetLogLevel(const std::string &level);
using GetLoggerNameFunc = std::function<std::string()>;

class SpdLogger : public Singleton<SpdLogger> {
public:
    SpdLogger() = default;
    virtual ~SpdLogger();

    std::pair<std::shared_ptr<yr_spdlog::logger>, std::string> GetLogger();
    void CreateLogger(const LogParam &logParam, const std::string &nodeName, const std::string &modelName);
    void Flush();
    yr_spdlog::level::level_enum level();
    std::string GetLogDir(void) const;
    std::string GetNodeName(void) const;
    std::string GetModelName(void) const;
    void SetGetLoggerNameFunc(GetLoggerNameFunc func);

private:
    std::string GetLogFile(const LogParam &logParam);
    void ConstructLoggerInfo(const LogParam &logParam);
    void RegisterLogger(const LogParam &logParam, const std::string &loggerName, const std::string &nodeName,
                        const std::string &modelName, const std::string &logFile);
    void AddLogPrefix(const std::string &key, const std::string &value);
    void RemoveLogPrefix(const std::string &key);
    void GetLogPrefix(const std::string &key, std::string &value);
    void Clear();
    void InitAsyncThread(const LogParam &logParam);
    std::string logDir;
    std::string nodeName;
    std::string modelName;
    yr_spdlog::level::level_enum logLevel;
    std::vector<yr_spdlog::sink_ptr> sinks;
    std::atomic<int> logMergeType_{0};  // 0: not merge log;1: merge log
    std::unordered_map<std::string, std::string> logPrefixMap_ ABSL_GUARDED_BY(mu_);
    absl::Mutex mu_;
    GetLoggerNameFunc getLoggerNameFunc = nullptr;
    absl::Mutex spdLoggerMu_;
};
}  // namespace utility
}  // namespace YR

// async mode to print log according to log level
#define YRLOG_ASYNC(lvl, logPrefix, logger, format,  ...)                             \
    do {                                                                              \
        if (logger && logger->should_log(lvl)) {                                      \
            if (YR_LIKELY(logPrefix == "")) {                                         \
                SPDLOG_LOGGER_CALL(logger, lvl, format, ##__VA_ARGS__);               \
            } else {                                                                  \
                SPDLOG_LOGGER_CALL(logger, lvl, logPrefix + format, ##__VA_ARGS__);   \
            }                                                                         \
        }                                                                             \
    } while (0)

#define YRLOG_ASYNC_WRAPPER(level, format, ...)                                       \
    do {                                                                              \
        auto [logger, logPrefix] = YR::utility::SpdLogger::GetInstance().GetLogger(); \
        auto lvl = YR::utility::GetLogLevel(level);                                   \
        YRLOG_ASYNC(lvl, logPrefix, logger, format, ##__VA_ARGS__);                   \
        if (strcmp(level, "FATAL") == 0) {                                            \
            (void)raise(SIGINT);                                                                  \
        }                                                                             \
    } while (0)
