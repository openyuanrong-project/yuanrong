/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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

#include "http_utils.h"
namespace YR {
namespace Libruntime {
std::string HashToHex(const std::string &message)
{
    std::stringstream ss;
    SHA256AndHex(message, ss);
    return ss.str();
}

std::string GenerateRequestDigest(const std::unordered_map<std::string, std::string> &headers, const std::string &body,
                                  const std::string &url)
{
    std::stringstream ss;
    ss << url << "\n";
    if (auto iter = headers.find(TIMESTAMP_KEY); iter != headers.end()) {
        ss << iter->first << ": " << iter->second << "\n";
    }
    if (auto iter = headers.find(ACCESS_KEY); iter != headers.end()) {
        ss << iter->first << ": " << iter->second << "\n";
    }
    ss << body;
    return ss.str();
}

void SignHttpRequest(const std::string &accessKey, const SensitiveValue &secretKey,
                     std::unordered_map<std::string, std::string> &headers, const std::string &body,
                     const std::string &url)
{
    auto timestamp = GetCurrentUTCTime();
    headers[TIMESTAMP_KEY] = timestamp;
    headers[ACCESS_KEY] = accessKey;

    auto digest = GenerateRequestDigest(headers, body, url);
    auto digestHashHex = HashToHex(digest);
    auto signature = GetHMACSha256(secretKey, digestHashHex);
    std::stringstream rss;
    rss << "HMAC-SHA256 timestamp=" << headers.at(TIMESTAMP_KEY) << ",access_key=" << headers.at(ACCESS_KEY)
        << ",signature=" << signature;
    headers[AUTHORIZATION_KEY] = rss.str();
}

bool VerifyHttpRequest(const std::string &accessKey, const SensitiveValue &secretKey,
                       std::unordered_map<std::string, std::string> &headers, const std::string &body,
                       const std::string &url)
{
    auto tenantAccessKey = headers.find(ACCESS_KEY);
    if (tenantAccessKey == headers.end() || tenantAccessKey->second.empty()) {
        YRLOG_ERROR("failed to verify http request: failed to find ACCESS_KEY in headers");
        return false;
    }
    auto timestamp = headers.find(TIMESTAMP_KEY);
    if (timestamp == headers.end() || timestamp->second.empty()) {
        YRLOG_ERROR("failed to verify http request: failed to find TIMESTAMP in headers");
        return false;
    }

    auto currentTimestamp = GetCurrentUTCTime();
    if (IsLaterThan(currentTimestamp, timestamp->second, TIMESTAMP_EXPIRE_DURATION_SECONDS)) {
        YRLOG_ERROR("failed to verify http request: failed to verify timestamp, difference is more than 1 min {} vs {}",
                    currentTimestamp, timestamp->second);
        return false;
    }

    auto authorizationValue = headers.find(AUTHORIZATION_KEY);
    if (authorizationValue == headers.end() || authorizationValue->second.empty()) {
        YRLOG_ERROR("failed to verify http request: failed to find Authorization in headers");
        return false;
    }
    std::string key = "signature=";
    size_t pos = authorizationValue->second.find(key);
    if (pos == std::string::npos) {
        return false;
    }
    std::string signature = authorizationValue->second.substr(pos + key.length());
    auto digest = GenerateRequestDigest(headers, body, url);
    auto digestHashHex = HashToHex(digest);
    if (GetHMACSha256(secretKey, digestHashHex) != signature) {
        YRLOG_ERROR("failed to verify http request");
        return false;
    }
    return true;
}

bool ShouldQueryEscape(char c)
{
    if (std::isalnum(c)) {
        // A~Z, a~z, 0~9 not escape
        return false;
    }
    if (c == '-' || c == '_' || c == '.' || c == '~') {
        // -, _, ., ~ not escpae
        return false;
    }
    // encode according to RFC 3986.
    return true;
}

std::string EscapeQuery(const std::string &s)
{
    std::stringstream ss;
    for (const auto &c : s) {
        if (!ShouldQueryEscape(c)) {
            ss << c;
        } else if (' ' == c) {
            ss << '+';
        } else {
            // encode according to RFC 3986.
            ss << '%' <<HEX_STRING_SET_CAP[c >> FIRST_FOUR_BIT_MOVE] << HEX_STRING_SET_CAP[c & 0xf];
        }
    }
    return ss.str();
}

std::string DoReplace(std::string &source, const std::string &replace, const std::string &replacement)
{
    if (replace.empty()) {
        return source;
    }
    size_t pos = 0;
    while ((pos = source.find(replace, pos)) != std::string::npos) {
        source.replace(pos, replace.length(), replacement);
        pos += replacement.length();
    }
    return source;
}

std::string EscapeURL(const std::string &url, bool replacePath)
{
    if (url.empty()) {
        return "";
    }
    std::string encodeurl = EscapeQuery(url);
    encodeurl = DoReplace(encodeurl, "+", "%20");
    encodeurl = DoReplace(encodeurl, "*", "%2A");
    encodeurl = DoReplace(encodeurl, "%7E", "~");
    if (replacePath) {
        encodeurl = DoReplace(encodeurl, "%2F", "/");
    }
    return encodeurl;
}

void ToLower(std::string &source)
{
    (void)std::transform(source.begin(), source.end(), source.begin(), [](unsigned char c) { return std::tolower(c); });
}

std::string &Trim(std::string &s, Mode mode, const std::string &chars)
{
    size_t start = 0;
    // set start pos
    if (mode == Mode::ANY || mode == Mode::PREFIX) {
        start = s.find_first_not_of(chars);
    }
    // if s contains only chars in chars
    if (start == std::string::npos) {
        s.clear();
        return s;
    }
    size_t length = std::string::npos;
    // set end pos
    if (mode == Mode::ANY || mode == Mode::SUFFIX) {
        length = s.find_last_not_of(chars) + 1 - start;
    }
    s = s.substr(start, length);
    return s;
}

std::string GetCanonicalQueries(const std::shared_ptr<std::map<std::string, std::string>> &queries)
{
    if (queries == nullptr) {
        return "";
    }
    bool first = true;
    std::stringstream ss;
    for (const auto &query : *queries) {
        if (first) {
            first = false;
            ss << EscapeURL(query.first, false) << '=' << EscapeURL(query.second, false);
        } else {
            ss << "&" << EscapeURL(query.first, false) << '=' << EscapeURL(query.second, false);
        }
    }
    return ss.str();
}

std::string GetCanonicalHeaders(const std::map<std::string, std::string> &headers)
{
    std::stringstream ss;
    for (const auto &header : headers) {
        std::string lowerKey = header.first;
        ToLower(lowerKey);

        if (lowerKey == HEADER_CONNECTION || lowerKey == HEADER_AUTHORIZATION) {
            continue;
        }

        std::string value = header.second;
        ss << lowerKey << ':' << Trim(value) << "\n";
    }
    return ss.str();
}

std::string GetSignedHeaders(const std::map<std::string, std::string> &headers)
{
    bool first = true;
    std::stringstream ss;
    for (const auto &header : headers) {
        std::string lowerKey = header.first;
        ToLower(lowerKey);

        if (lowerKey == HEADER_CONNECTION || lowerKey == HEADER_AUTHORIZATION) {
            continue;
        }

        if (first) {
            first = false;
            ss << lowerKey;
        } else {
            ss << ";" << lowerKey;
        }
    }
    return ss.str();
}

std::string GetCanonicalRequest(const std::string &method, const std::string &path,
                                const std::shared_ptr<std::map<std::string, std::string>> &queries,
                                const std::map<std::string, std::string> &headers, const std::string &sha256)
{
    std::string canonicalPath = path.empty() ? "/" : EscapeURL(path, true);
    std::string canonicalQueries = GetCanonicalQueries(queries);
    std::string canonicalHeaders = GetCanonicalHeaders(headers);
    std::string signedHeaders = GetSignedHeaders(headers);

    std::stringstream ss;
    ss << method << "\n"
       << canonicalPath << "\n"
       << canonicalQueries << "\n"
       << canonicalHeaders << "\n"
       << signedHeaders << "\n";
    if (sha256.empty()) {
        // default empty canonical sha256
        ss << EMPTY_CONTENT_SHA256;
    } else {
        ss << sha256;
    }
    return ss.str();
}

// 从 std::unordered_map 转换为 std::map ，并对元素进行排序
std::map<std::string, std::string> UnorderedMapToMapSorted(const std::unordered_map<std::string, std::string> &um)
{
    // 将 unordered_map 的元素复制到 vector 中
    std::vector<std::pair<std::string, std::string>> vec(um.begin(), um.end());

    // 对 vector 中的元素进行排序
    std::sort(
        vec.begin(), vec.end(),
        [](const std::pair<std::string, std::string>& a, const std::pair<std::string, std::string>& b) {
            return a.first < b.first;
    });

    // 将排序后的元素插入到 map 中
    std::map<std::string, std::string> m;
    for (const auto& pair : vec) {
        m.insert(pair);
    }
    return m;
}

std::string HashToHex2(const std::string &message)
{
    std::stringstream ss;
    SHA256AndHex(message, ss);
    std::string result = ss.str();
    if (!result.empty() && result.back() == '\n') {
        result.pop_back();
    }
    return result;
}

void SignHttpRequest2(std::shared_ptr<Security> security, const std::string &method,
                      const std::string &path, std::unordered_map<std::string, std::string> &headers,
                      const std::string &body)
{
    std::string ak;
    SensitiveValue sk;
    security->GetAKSK(ak, sk);
    if (ak.empty() || sk.Empty()) {
        YRLOG_ERROR("ak or sk is empty");
        return;
    }
    std::string timestamp = GetCurrentUTCTime();
    auto signedHeaderMap = UnorderedMapToMapSorted(headers);
    std::string encodeBody = HashToHex2(body);
    std::string canonicalRequest = GetCanonicalRequest(method, path, nullptr, signedHeaderMap, encodeBody);
    std::string signatureData = timestamp + " " + HashToHex2(canonicalRequest);
    auto signature = GetHMACSha256(sk, signatureData);
    std::stringstream rss;
    rss << "HmacSha256 timestamp=" << timestamp << ",ak=" << ak << ",signature=" << signature;
    headers[X_SIGNATURE] = rss.str();
    std::string signHeaders = GetSignedHeaders(signedHeaderMap);
    headers[X_SIGNED_HEADER] = signHeaders;
    headers[X_AUTH] = ak;
}
}  // namespace Libruntime
}  // namespace YR