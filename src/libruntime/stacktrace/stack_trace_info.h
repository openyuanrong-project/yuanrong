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
#include <vector>

namespace YR {
namespace Libruntime {
struct StackTraceElement {
    std::string className;
    std::string methodName;
    std::string fileName;
    int lineNumber;
    // extensions: extension field which is used to transfer the unique fields of each language stack
    // for example, the parameters and address offset in the go stack
    std::unordered_map<std::string, std::string> extensions;
};

class StackTraceInfo {
public:
    StackTraceInfo() {}
    StackTraceInfo(std::string type, std::string message) : type_(type), message_(message) {}
    StackTraceInfo(std::string type, std::string message, std::string language)
        : type_(type), message_(message), language_(language)
    {
    }
    StackTraceInfo(std::string type, std::string message, std::vector<StackTraceElement> stackTraceElements)
        : type_(type), message_(message), stackTraceElements_(std::move(stackTraceElements))
    {
    }
    StackTraceInfo(std::string type, std::string message, std::vector<StackTraceElement> stackTraceElements,
                   std::string language)
        : type_(type), message_(message), stackTraceElements_(std::move(stackTraceElements)), language_(language)
    {
    }
    StackTraceInfo &operator=(const StackTraceInfo &info)
    {
        type_ = info.Type();
        message_ = info.Message();
        stackTraceElements_ = std::move(info.StackTraceElements());
        return *this;
    }

    StackTraceInfo(const StackTraceInfo &info)
    {
        type_ = info.Type();
        message_ = info.Message();
        stackTraceElements_ = std::move(info.StackTraceElements());
        language_ = info.Language();
    }

    std::string Type() const
    {
        return type_;
    }

    std::string Message() const
    {
        return message_;
    }
    
    void SetMsg(const std::string &errMsg)
    {
        message_ = errMsg;
    }

    std::vector<StackTraceElement> StackTraceElements() const
    {
        return stackTraceElements_;
    }

    void SetStackTraceElements(std::vector<StackTraceElement> &stackTraceElements)
    {
        stackTraceElements_ = std::move(stackTraceElements);
    }

    std::string Language() const
    {
        return language_;
    }

private:
    std::string type_;
    std::string message_;
    std::vector<StackTraceElement> stackTraceElements_;
    std::string language_;
};

}
}
