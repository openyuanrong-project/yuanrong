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
 
#pragma once

#include <sstream>
#include <vector>

#include <boost/algorithm/string.hpp>
#include <boost/beast/core/detail/base64.hpp>
#include "re2/re2.h"
namespace YR {
namespace utility {
inline void Split(const std::string &source, std::vector<std::string> &result, const char sep)
{
    result.clear();
    std::istringstream iss(source);
    std::string temp;
    while (std::getline(iss, temp, sep)) {
        result.emplace_back(std::move(temp));
    }
}

inline std::string Join(std::vector<std::string> &&parts, std::string &&delimiter)
{
    return boost::algorithm::join(parts, delimiter);
}

inline std::string Join(const std::vector<std::string> &parts, const std::string &delimiter)
{
    return boost::algorithm::join(parts, delimiter);
}

inline std::string TrimSpace(const std::string &s)
{
    auto begin = s.find_first_not_of(' ');
    if (begin == std::string::npos) {
        return std::string();
    }
    auto end = s.find_last_not_of(' ');
    if (end == std::string::npos) {
        return std::string();
    }
    return s.substr(begin, end - begin + 1);
}

inline std::pair<int, std::string> CompareIntFromString(const std::string &l, const std::string &r)
{
    try {
        auto lv = std::stoll(l);
        auto rv = std::stoll(r);
        if (lv > rv) {
            return std::make_pair(1, "");
        }
        if (lv < rv) {
            return std::make_pair(-1, "");
        }
        return std::make_pair(0, "");
    } catch (std::exception &e) {
        return std::make_pair(0, e.what());
    }
}

inline std::string DecodedToString(const std::string &inStr)
{
    auto len = boost::beast::detail::base64::decoded_size(inStr.size());
    uint8_t data[len + 1] = {0};
    boost::beast::detail::base64::decode(data, inStr.data(), inStr.size());
    return reinterpret_cast<const char *>(data);
}

inline std::vector<std::string> SplitToStr(const std::string &info, const std::string &pattern)
{
    if (info.empty()) {
        return {};
    }
    re2::StringPiece text(info);
    re2::RE2 re2Pattern(pattern);
    std::vector<std::string> result;
    size_t lastPos = 0;
 
    re2::StringPiece match;
    while (re2Pattern.Match(text, lastPos, text.size(), RE2::UNANCHORED, &match, 1)) {
        uint64_t splitIndex = match.data() - text.data();
        if (match.data() - (text.data() + lastPos) > 0) {
            result.push_back(std::string(text.data() + lastPos, match.data() - (text.data() + lastPos)));
        }
        lastPos = splitIndex + match.size();
    }
 
    if (lastPos < text.size()) {
        result.push_back(std::string(text.data() + lastPos));
    }
    return result;
}
}  // namespace utility
}  // namespace YR