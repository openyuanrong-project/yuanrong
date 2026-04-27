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

#include "spd_logger.h"

#include <chrono>
#include <iomanip>
#include <iostream>
#include <sstream>
#include <mutex>

#include "spdlog/async.h"
#include "spdlog/sinks/dup_filter_sink.h"

namespace YR {
namespace utility {
static const uint32_t DUP_FILTER_TIME = 60;
const std::string DEFAULT_LOG_NAME = "driver";
const std::string DEFAULT_LOG_NAME_PREFIX = "libruntime";
static const int LOG_NOT_MERGE_TYPE = 0;
static const int LOG_MERGE_TYPE = 1;

yr_spdlog::level::level_enum GetLogLevel(const std::string &level)
{
    static std::map<std::string, yr_spdlog::level::level_enum> logLevelMap = {
        {"TRACE", yr_spdlog::level::trace}, {"DEBUG", yr_spdlog::level::debug}, {"INFO", yr_spdlog::level::info},
        {"WARN", yr_spdlog::level::warn},   {"ERR", yr_spdlog::level::err},     {"FATAL", yr_spdlog::level::critical}};
    auto iter = logLevelMap.find(level);
    return iter == logLevelMap.end() ? yr_spdlog::level::info : iter->second;
}

std::string FormatTimePoint()
{
    using std::chrono::system_clock;
    std::time_t tt = system_clock::to_time_t(system_clock::now());
    struct std::tm ptm {};
    (void)::localtime_r(&tt, &ptm);

    std::stringstream ss;
    ss << std::put_time(&ptm, "%Y%m%d%H%M%S");

    return ss.str();
}

SpdLogger::~SpdLogger() {}

std::string SpdLogger::GetLogDir(void) const
{
    return this->logDir;
}

std::string SpdLogger::GetNodeName(void) const
{
    return this->nodeName;
}

std::string SpdLogger::GetModelName(void) const
{
    return this->modelName;
}

std::pair<std::shared_ptr<yr_spdlog::logger>, std::string> SpdLogger::GetLogger()
{
    std::string loggerName = this->getLoggerNameFunc ? getLoggerNameFunc() : LOGGER_NAME;
    auto logger = yr_spdlog::get(loggerName);
    std::string logPrefix = "";
    if (logMergeType_.load() == LOG_MERGE_TYPE) {
        GetLogPrefix(loggerName, logPrefix);
    }
    return std::make_pair(logger, logPrefix);
}

void SpdLogger::SetGetLoggerNameFunc(GetLoggerNameFunc func)
{
    this->getLoggerNameFunc = func;
}

void SpdLogger::ConstructLoggerInfo(const LogParam &logParam)
{
    this->logDir = logParam.logDir;
    this->nodeName = logParam.nodeName;
    this->modelName = logParam.modelName;
}

std::string SpdLogger::GetLogFile(const LogParam &logParam)
{
    auto logFile = logParam.logDir + "/";
    if (logParam.isLogMerge) {
        pid_t pid = getpid();
        logFile += Join({DEFAULT_LOG_NAME, std::to_string(pid)}, "-") + LOG_SUFFIX;
    } else if (logParam.logFileWithTime) {
        logFile += Join({logParam.nodeName, logParam.modelName, FormatTimePoint()}, "-") + LOG_SUFFIX;
    } else {
        if (logParam.modelName == DEFAULT_LOG_NAME) {
            logFile += Join({logParam.nodeName, logParam.modelName}, "-") + LOG_SUFFIX;
        } else {
            if (logParam.withLogPrefix) {
                logFile += Join({logParam.loggerId, DEFAULT_LOG_NAME_PREFIX}, "_") + LOG_SUFFIX;
            } else {
                logFile += Join({logParam.nodeName, logParam.modelName}, "-") + LOG_SUFFIX;
            }
        }
    }

    return logFile;
}

void SpdLogger::CreateLogger(const LogParam &logParam, const std::string &nodeName, const std::string &modelName)
{
    ConstructLoggerInfo(logParam);
    try {
        std::string logFile = GetLogFile(logParam);
        if (logParam.isLogMerge) {
            if (logMergeType_.load() == LOG_NOT_MERGE_TYPE) {
                Clear();
                InitAsyncThread(logParam);
            }
            logMergeType_.store(LOG_MERGE_TYPE);
            RegisterLogger(logParam, nodeName, nodeName, modelName, logFile);
            AddLogPrefix(nodeName, nodeName + "," + modelName + "]");
            RegisterLogger(logParam, LOGGER_NAME, DEFAULT_JOB_ID, "driver", logFile);
            AddLogPrefix(LOGGER_NAME, DEFAULT_JOB_ID + ",driver]");
        } else {
            if (logMergeType_.load() == LOG_MERGE_TYPE) {
                std::cout << "failed to init logger when log that support log merge has been initialized" << std::endl
                          << std::flush;
                return;
            }
            logMergeType_.store(LOG_NOT_MERGE_TYPE);
            // clear old loggers, include default logger
            Clear();
            InitAsyncThread(logParam);
            RegisterLogger(logParam, LOGGER_NAME, nodeName, modelName, logFile);
        }
    } catch (const yr_spdlog::spdlog_ex &ex) {
        std::cout << "failed to init logger:" << ex.what() << std::endl << std::flush;
    }
}

void SpdLogger::RegisterLogger(const LogParam &logParam, const std::string &loggerName, const std::string &nodeName,
                               const std::string &modelName, const std::string &logFile)
{
    absl::WriterMutexLock lock(&spdLoggerMu_);
    auto logger = yr_spdlog::get(loggerName);
    if (logger) {
        if (!logParam.isLogMerge) {
            yr_spdlog::drop(loggerName);
        } else {
            return;
        }
    }

    if (sinks.empty()) {
        if (logParam.onlyStdout) {
            // Only output to stdout, no file sink
            auto consoleSink = std::make_shared<yr_spdlog::sinks::stdout_color_sink_mt>();
            sinks = {consoleSink};
        } else {
            auto rotatingSink = std::make_shared<yr_spdlog::sinks::rotating_file_sink_mt>(logFile,
                logParam.maxSize * SIZE_MEGA_BYTES, logParam.maxFiles);
            auto dupFilter = std::make_shared<yr_spdlog::sinks::dup_filter_sink_mt>(
                std::chrono::seconds(DUP_FILTER_TIME));
            sinks = {rotatingSink, dupFilter};
            if (logParam.alsoLog2Stderr) {
                auto consoleSink = std::make_shared<yr_spdlog::sinks::stdout_color_sink_mt>();
                (void)sinks.emplace_back(consoleSink);
            }
        }
    }
    logger = std::make_shared<yr_spdlog::async_logger>(loggerName, sinks.begin(), sinks.end(),
                                                       yr_spdlog::thread_pool(),
                                                       yr_spdlog::async_overflow_policy::block);

    logLevel = GetLogLevel(logParam.logLevel);
    logger->set_level(logLevel);
    setenv("DATASYSTEM_CLIENT_LOG_DIR", logParam.logDir.c_str(), 1);

    auto pattern = "%L%m%d %H:%M:%S.%f %t %s:%#] %P,%!]" + nodeName + "," + modelName + "]%v";
    if (logMergeType_.load() == LOG_MERGE_TYPE) {
        pattern = "%L%m%d %H:%M:%S.%f %t %s:%#] %P,%!]%v";
    }
    auto timeType = logParam.useUtcTime ? yr_spdlog::pattern_time_type::utc : yr_spdlog::pattern_time_type::local;
    logger->set_pattern(pattern, timeType);

    yr_spdlog::register_logger(logger);
}

void SpdLogger::Flush()
{
    std::string loggerName = this->getLoggerNameFunc ? getLoggerNameFunc() : LOGGER_NAME;
    auto logger = yr_spdlog::get(loggerName);
    if (logger) {
        logger->flush();
    }
}

yr_spdlog::level::level_enum SpdLogger::level()
{
    return logLevel;
}

void SpdLogger::AddLogPrefix(const std::string &key, const std::string &value)
{
    absl::WriterMutexLock lock(&mu_);
    logPrefixMap_.emplace(key, value);
}

void SpdLogger::RemoveLogPrefix(const std::string &key)
{
    absl::WriterMutexLock lock(&mu_);
    logPrefixMap_.erase(key);
}

void SpdLogger::GetLogPrefix(const std::string &key, std::string &value)
{
    absl::ReaderMutexLock lock(&mu_);
    if (logPrefixMap_.find(key) != logPrefixMap_.end()) {
        value = logPrefixMap_[key];
    }
}

void SpdLogger::Clear()
{
    Flush();
    yr_spdlog::drop_all();
    sinks.clear();
}

void SpdLogger::InitAsyncThread(const LogParam &logParam)
{
    static std::once_flag onceflag;
    std::call_once(onceflag, [logParam]() {
        try {
            if (!yr_spdlog::thread_pool()) {
                yr_spdlog::init_thread_pool(static_cast<size_t>(logParam.maxAsyncQueueSize),
                                            static_cast<size_t>(logParam.asyncThreadCount));
            }
            yr_spdlog::flush_every(std::chrono::seconds(logParam.logBufSecs));
        } catch (const yr_spdlog::spdlog_ex &ex) {
            std::cout << "failed to init logger thread pool:" << ex.what() << std::endl << std::flush;
        }
    });
}
}  // namespace utility
}  // namespace YR
