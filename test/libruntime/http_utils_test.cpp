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

#include "src/libruntime/utils/http_utils.h"
#include <gmock/gmock.h>
#include <gtest/gtest.h>

namespace YR {
namespace test {
using namespace YR::utility;
using namespace YR::Libruntime;
class HttpUtilTest : public ::testing::Test {
protected:
    void SetUp() override
    {
        Mkdir("/tmp/log");
        LogParam g_logParam = {
            .logLevel = "DEBUG",
            .logDir = "/tmp/log",
            .nodeName = "test-runtime",
            .modelName = "test",
            .maxSize = 100,
            .maxFiles = 1,
            .logFileWithTime = false,
            .logBufSecs = 30,
            .maxAsyncQueueSize = 1048510,
            .asyncThreadCount = 1,
            .alsoLog2Stderr = true,
        };
        InitLog(g_logParam);
    }

    void TearDown() override {}
};

TEST_F(HttpUtilTest, SignHttpRequest)
{
    std::string url =
        "/serverless/v2/functions/"
        "wisefunction:cn:iot:8d86c63b22e24d9ab650878b75408ea6:function:0@faas@python:latest/invocations";
    std::string accessKey = "access_key";
    SensitiveValue secretKey = std::string("secret_key");
    std::unordered_map<std::string, std::string> headers;
    headers[TRACE_ID_KEY_NEW] = "traceId";
    headers[INSTANCE_CPU_KEY] = "500";
    headers[INSTANCE_MEMORY_KEY] = "300";
    std::string body = "123";
    SignHttpRequest(accessKey, secretKey, headers, body, url);
    YRLOG_DEBUG(headers[AUTHORIZATION_KEY]);
    ASSERT_FALSE(headers[AUTHORIZATION_KEY].empty());
    bool ok = VerifyHttpRequest(accessKey, secretKey, headers, body, url);
    ASSERT_TRUE(ok);
    auto digest = GenerateRequestDigest(headers, body, url);
    YRLOG_DEBUG(digest);
    ASSERT_TRUE(!digest.empty());
}

TEST_F(HttpUtilTest, ShouldQueryEscape)
{
    for (char c = 'A'; c <= 'Z'; ++c) {
        EXPECT_FALSE(ShouldQueryEscape(c));
    }
    for (char c = 'a'; c <= 'z'; ++c) {
        EXPECT_FALSE(ShouldQueryEscape(c));
    }
    for (char c = '0'; c <= '9'; ++c) {
        EXPECT_FALSE(ShouldQueryEscape(c));
    }
    EXPECT_FALSE(ShouldQueryEscape('_'));
    EXPECT_FALSE(ShouldQueryEscape('-'));
    EXPECT_FALSE(ShouldQueryEscape('.'));
    EXPECT_FALSE(ShouldQueryEscape('~'));
    EXPECT_TRUE(ShouldQueryEscape('='));
    EXPECT_TRUE(ShouldQueryEscape('&'));
}

TEST_F(HttpUtilTest, EscapeQuery)
{
    EXPECT_EQ(EscapeQuery("HelloWorld"), "HelloWorld");
    EXPECT_EQ(EscapeQuery("hello-world_123"), "hello-world_123");
    EXPECT_EQ(EscapeQuery("a-b_c.d~e"), "a-b_c.d~e");
    EXPECT_EQ(EscapeQuery("hello world"), "hello+world");
    EXPECT_EQ(EscapeQuery("  double  space  "), "++double++space++");
    EXPECT_EQ(EscapeQuery(""), "");
    EXPECT_EQ(EscapeQuery("!@#$%^&*()"), "%21%40%23%24%25%5E%26%2A%28%29");
    EXPECT_EQ(EscapeQuery("/?\\<>,.'\""), "%2F%3F%5C%3C%3E%2C.%27%22");
    EXPECT_EQ(EscapeQuery("[]{}|;"), "%5B%5D%7B%7D%7C%3B");
}

TEST_F(HttpUtilTest, DoReplace)
{
    std::string source = "hello world";
    EXPECT_EQ(DoReplace(source, "world", "universe"), "hello universe");

    source = "hello world";
    EXPECT_EQ(DoReplace(source, "", "universe"), "hello world");
}

TEST_F(HttpUtilTest, EscapeURL)
{
    EXPECT_EQ(EscapeURL("", false), "");
    EXPECT_EQ(EscapeURL("", true), "");
    EXPECT_EQ(EscapeURL("http://example.com", false), "http%3A%2F%2Fexample.com");
    EXPECT_EQ(EscapeURL("http://example.com", true), "http%3A//example.com");
    EXPECT_EQ(EscapeURL("http://example.com/hello*world", false), "http%3A%2F%2Fexample.com%2Fhello%2Aworld");
    EXPECT_EQ(EscapeURL("http://example.com/hello*world", true), "http%3A//example.com/hello%2Aworld");
}

TEST_F(HttpUtilTest, ToLower)
{
    std::string source = "";
    ToLower(source);
    EXPECT_EQ(source, "");

    source = "HELLO WORLD";
    ToLower(source);
    EXPECT_EQ(source, "hello world");
}

TEST_F(HttpUtilTest, Trim)
{
    std::string s = "";
    Trim(s, Mode::ANY, " \t\n\r\f\v");
    EXPECT_EQ(s, "");

    s = "   hello world";
    Trim(s, Mode::PREFIX, " \t\n\r\f\v");
    EXPECT_EQ(s, "hello world");

    s = "   \t\n\r\f\v   ";
    Trim(s, Mode::ANY, " \t\n\r\f\v");
    EXPECT_EQ(s, "");
}

TEST_F(HttpUtilTest, GetCanonicalQueries)
{
    std::shared_ptr<std::map<std::string, std::string>> queries = nullptr;
    EXPECT_EQ(GetCanonicalQueries(queries), "");

    queries = std::make_shared<std::map<std::string, std::string>>();
    EXPECT_EQ(GetCanonicalQueries(queries), "");

    queries = std::make_shared<std::map<std::string, std::string>>();
    queries->emplace("key1", "");
    queries->emplace("key2", "value2");
    EXPECT_EQ(GetCanonicalQueries(queries), "key1=&key2=value2");

    queries = std::make_shared<std::map<std::string, std::string>>();
    queries->emplace("key with spaces", "value with spaces");
    queries->emplace("key&special", "value=special");
    EXPECT_EQ(GetCanonicalQueries(queries), "key%20with%20spaces=value%20with%20spaces&key%26special=value%3Dspecial");
}

TEST_F(HttpUtilTest, GetCanonicalHeaders)
{
    std::map<std::string, std::string> headers;
    EXPECT_EQ(GetCanonicalHeaders(headers), "");

    headers["Content-Type"] = "application/json";
    headers["Authorization"] = "Bear token";
    headers["Connection"] = "keep-alive";
    headers["Accept"] = "text/plain";
    EXPECT_EQ(GetCanonicalHeaders(headers), "accept:text/plain\ncontent-type:application/json\n");
}

TEST_F(HttpUtilTest, GetSignedHeaders)
{
    std::map<std::string, std::string> headers;
    EXPECT_EQ(GetSignedHeaders(headers), "");

    headers["Connection"] = "keep-alive";
    headers["Authorization"] = "Bear token";
    headers["Content-Type"] = "application/json";
    headers["ACCEPT"] = "text/plain";
    EXPECT_EQ(GetSignedHeaders(headers), "accept;content-type");
}

TEST_F(HttpUtilTest, GetCanonicalRequest)
{
    std::string method = "GET";
    std::string path = "";
    auto queries = std::make_shared<std::map<std::string, std::string>>();
    std::map<std::string, std::string> headers;
    std::string sha256 = "";
    std::string expected = "GET\n/\n\n\n\n" + EMPTY_CONTENT_SHA256;
    EXPECT_EQ(GetCanonicalRequest(method, path, queries, headers, sha256), expected);

    method = "POST";
    path = "/api/v1/resource";
    queries->emplace("key1", "value1");
    queries->emplace("key2", "value2");
    headers["Content-Type"] = "application/json";
    headers["Accept"] = "text/plain";
    sha256 = "abc123";
    expected = "POST\n/api/v1/resource\nkey1=value1&key2=value2\naccept:text/plain\ncontent-type:application/json\n\naccept;content-type\nabc123";
    EXPECT_EQ(GetCanonicalRequest(method, path, queries, headers, sha256), expected);
}

TEST_F(HttpUtilTest, UnorderedMapToMapSorted)
{
    std::unordered_map<std::string, std::string> um;
    std::map<std::string, std::string> m = UnorderedMapToMapSorted(um);
    EXPECT_TRUE(m.empty());

    um[""] = "empty";
    um["key3"] = "value3";
    um["key2"] = "value2";
    um["key1"] = "value1";
    m = UnorderedMapToMapSorted(um);
    EXPECT_EQ(m.size(), 4);
    EXPECT_EQ(m[""], "empty");
    EXPECT_EQ(m["key1"], "value1");
    EXPECT_EQ(m["key2"], "value2");
    EXPECT_EQ(m["key3"], "value3");
}

TEST_F(HttpUtilTest, HashToHex2)
{
    std::string message = "";
    std::string expected = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855";
    EXPECT_EQ(HashToHex2(message), expected);

    message = "hello world";
    expected = "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9";
    EXPECT_EQ(HashToHex2(message), expected);
}

TEST_F(HttpUtilTest, SignHttpRequest2Test)
{
    auto security = std::make_shared<Security>();
    std::string method = "GET";
    std::string path = "/api/v1/resource";
    std::unordered_map<std::string, std::string> headers;
    headers["Content-type"] = "applocation/json";
    headers["Accept"] = "text/plain";
    std::string body = "test body";

    SignHttpRequest2(security, method, path, headers, body);
    EXPECT_TRUE(headers[X_SIGNATURE].empty());

    std::string ak = "ak";
    SensitiveValue sk = std::string("sk");
    security->SetAKSKAndCredential(ak, sk);
    SignHttpRequest2(security, method, path, headers, body);
    EXPECT_FALSE(headers[X_SIGNATURE].empty());
    EXPECT_FALSE(headers[X_SIGNED_HEADER].empty());
    EXPECT_EQ(headers[X_AUTH], "ak");
}
}  // namespace test
}  // namespace YR