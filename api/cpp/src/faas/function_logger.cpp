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

#include "FunctionLogger.h"

#include <cstdarg>
#include <iostream>

#include <securec.h>

using namespace std;
namespace Function {
const std::string LOG_INFO = "INFO";
const std::string LOG_WARN = "WARN";
const std::string LOG_DEBUG = "DEBUG";
const std::string LOG_ERROR = "ERROR";
const size_t PRINTF_BUFFER_LENGTH = 2048;

string AdaptPrintf(const string &message, va_list &args)
{
    char buf[PRINTF_BUFFER_LENGTH] = {0};
    int ret = vsprintf_s(buf, sizeof(buf), message.c_str(), args);
    if (ret < 0) {
        return "";
    }
    return string(buf);
}

void FunctionLogger::Info(string message, ...)
{
    if (logLevel == LOG_WARN || logLevel == LOG_ERROR) {
        return;
    }
    if (sendEmptyLog(message, LOG_INFO)) {
        return;
    }
    va_list args;
    va_start(args, message);
    string logMessage = AdaptPrintf(message, args);
    va_end(args);
}

void FunctionLogger::Warn(string message, ...)
{
    if (logLevel == LOG_ERROR) {
        return;
    }
    if (sendEmptyLog(message, LOG_WARN)) {
        return;
    }
    va_list args;
    va_start(args, message);
    string logMessage = AdaptPrintf(message, args);
    va_end(args);
}

void FunctionLogger::Debug(string message, ...)
{
    if (logLevel == LOG_INFO || logLevel == LOG_WARN || logLevel == LOG_ERROR) {
        return;
    }
    if (sendEmptyLog(message, LOG_DEBUG)) {
        return;
    }
    va_list args;
    va_start(args, message);
    string logMessage = AdaptPrintf(message, args);
    va_end(args);
}

void FunctionLogger::Error(string message, ...)
{
    if (sendEmptyLog(message, LOG_ERROR)) {
        return;
    }
    va_list args;
    va_start(args, message);
    string logMessage = AdaptPrintf(message, args);
    va_end(args);
}

void FunctionLogger::Log(const string &level, const string &logMessage)
{
    return;
}

bool FunctionLogger::sendEmptyLog(const string &message, const string &level)
{
    if (message.empty()) {
        return true;
    }
    return false;
}

void FunctionLogger::setLevel(const string &level)
{
    if (level == LOG_DEBUG || level == LOG_INFO || level == LOG_WARN || level == LOG_ERROR) {
        this->logLevel = level;
    }
}
void FunctionLogger::SetInvokeID(std::string invokeID)
{
    this->invokeId = invokeID;
}
void FunctionLogger::SetTraceID(std::string traceID)
{
    this->traceId = traceID;
}
}  // namespace Function
