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

#include "logger.h"

#include <fstream>
#if __has_include(<filesystem>)
#include <filesystem>
#else
#include <experimental/filesystem>
#endif

#include "absl/debugging/failure_signal_handler.h"
#include "absl/debugging/symbolize.h"

#include "src/utility/string_utility.h"

namespace YR {
namespace utility {
const std::string exceptionLogSubdir = "exception";
const std::string backtracePrefix = "BackTrace";
const std::string mapinfoPrefix = "Mapinfo";

#ifdef __cpp_lib_filesystem
using fspath = std::filesystem::path;
#elif __cpp_lib_experimental_filesystem
using fspath = std::experimental::filesystem::path;
#endif

void InitLog(const LogParam &logParam)
{
    SpdLogger::GetInstance().CreateLogger(logParam, logParam.nodeName, logParam.modelName);
}

std::string MakeExceptionHeader(const std::string &nodeName, const std::string &modelName, const fspath &exceptionPath,
                                const fspath &mapInfoPath)
{
    std::stringstream ss;
    ss << "job: " << nodeName << ", runtime: " << modelName << ", receive signal\n";
    ss << "Record Exception in remote path:" << exceptionPath;
    ss << "\nRecord MapInfo in remote path: " << mapInfoPath << "\n";
    return ss.str();
}

void WriteMapinfo(void)
{
    // support later
}

void FailureSignalWriter(const char *data)
{
    // Null `data` is a hint to flush any buffered data before the program may be terminated.
    if (!data) {
        SpdLogger::GetInstance().Flush();
        return;
    }
    auto logDir = SpdLogger::GetInstance().GetLogDir();
    auto exceptionDir = fspath(logDir) / fspath(exceptionLogSubdir);
    if (!ExistPath(exceptionDir.string())) {
        Mkdir(exceptionDir.string());
    }

    auto nodeName = SpdLogger::GetInstance().GetNodeName();
    auto modelName = SpdLogger::GetInstance().GetModelName();

    auto backtraceFilename = Join({backtracePrefix, modelName}, "_") + LOG_SUFFIX;
    auto backtracePath = fspath(logDir) / fspath(exceptionLogSubdir) / fspath(backtraceFilename);

    auto mapinfoFilename = Join({mapinfoPrefix, modelName}, "_") + LOG_SUFFIX;
    auto mapinfoPath = fspath(logDir) / fspath(exceptionLogSubdir) / fspath(mapinfoFilename);

    std::ofstream file(backtracePath.string(), std::ios::out | std::ios::app);
    bool firstTime = file.tellp() == 0;
    if (firstTime) {
        auto header = MakeExceptionHeader(nodeName, modelName, backtracePath, mapinfoPath);
        file << header;
    }
    file << data;
    file.close();

    if (firstTime) {
        WriteMapinfo();
    }
}

void InstallFailureSignalHandler(const char *arg0)
{
    absl::InitializeSymbolizer(arg0);

    absl::FailureSignalHandlerOptions options;
    options.call_previous_handler = true;
    options.writerfn = FailureSignalWriter;
    options.alarm_on_failure_secs = 0;
    absl::InstallFailureSignalHandler(options);
}

void SetGetLoggerNameFunc(GetLoggerNameFunc func)
{
    SpdLogger::GetInstance().SetGetLoggerNameFunc(func);
}
}  // namespace utility
}  // namespace YR