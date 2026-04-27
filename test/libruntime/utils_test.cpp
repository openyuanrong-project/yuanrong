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

#include <gmock/gmock.h>
#include <gtest/gtest.h>
#include <unistd.h>
#include <cstdlib>
#include <cstring>
#include <fstream>

#include "src/libruntime/err_type.h"
#include "src/libruntime/fsclient/protobuf/common.pb.h"
#include "src/libruntime/fsclient/protobuf/runtime_service.pb.h"
#include "src/libruntime/utils/datasystem_utils.h"
#include "src/libruntime/utils/security.h"
#include "src/libruntime/utils/utils.h"
#include "src/proto/libruntime.pb.h"

using namespace testing;
using namespace YR::Libruntime;

namespace YR {
namespace test {
class UtilsTest : public ::testing::Test {
public:
    UtilsTest() {}
    ~UtilsTest() {}
};

TEST_F(UtilsTest, SetCallResultWithStackTraceInfoSuccessfully)
{
    CallResult callResult;
    callResult.set_requestid("requestid");
    callResult.set_instanceid("instanceid");

    callResult.set_code(static_cast<common::ErrorCode>(common::ErrorCode::ERR_INNER_COMMUNICATION));
    callResult.set_message("success");

    YR::Libruntime::StackTraceElement element = {"IOException", "testException", "filename", 8};
    std::vector<YR::Libruntime::StackTraceElement> elements = {element};
    YR::Libruntime::StackTraceInfo stackTraceInfo("type", "message", elements, "JAVA");
    std::vector<YR::Libruntime::StackTraceInfo> stackTraces = {stackTraceInfo};

    YR::SetCallResultWithStackTraceInfo(stackTraces, callResult);

    ASSERT_EQ(callResult.stacktraceinfos().size(), 1);
    ASSERT_EQ(callResult.stacktraceinfos()[0].stacktraceelements().size(), 1);
    ASSERT_EQ(callResult.stacktraceinfos()[0].stacktraceelements()[0].classname(), "IOException");
    ASSERT_EQ(callResult.stacktraceinfos()[0].type(), "type");
}

TEST_F(UtilsTest, GetStackTraceInfosSuccessfully)
{
    NotifyRequest notifyRequest;
    notifyRequest.set_requestid("requestid");
    notifyRequest.set_code(static_cast<common::ErrorCode>(common::ErrorCode::ERR_INNER_COMMUNICATION));
    notifyRequest.set_message("success");

    YR::Libruntime::StackTraceElement element = {"IOException", "testException", "filename", 8};
    std::vector<YR::Libruntime::StackTraceElement> elements = {element};
    YR::Libruntime::StackTraceInfo stackTraceInfo("type", "message", elements, "JAVA");
    std::vector<YR::Libruntime::StackTraceInfo> stackTraces = {stackTraceInfo};

    for (size_t i = 0; i < stackTraces.size(); i++) {
        auto setInfo = notifyRequest.add_stacktraceinfos();
        auto eles = stackTraces[i].StackTraceElements();
        for (size_t j = 0; j < eles.size(); j++) {
            auto stackTraceEle = setInfo->add_stacktraceelements();
            stackTraceEle->set_classname(eles[j].className);
            stackTraceEle->set_methodname(eles[j].methodName);
            stackTraceEle->set_filename(eles[j].fileName);
            stackTraceEle->set_linenumber(eles[j].lineNumber);
        }
        setInfo->set_type(stackTraces[i].Type());
        setInfo->set_message(stackTraces[i].Message());
    }

    std::vector<YR::Libruntime::StackTraceInfo> infos = YR::GetStackTraceInfos(notifyRequest);
    ASSERT_EQ(infos.size(), 1);
    ASSERT_EQ(infos[0].StackTraceElements().size(), 1);
    ASSERT_EQ(infos[0].Type(), "type");
}

TEST_F(UtilsTest, SensitiveDataTest)
{
    std::string value = "123";
    SensitiveData sensitiveData;
    SensitiveData sensitiveData2("1234");
    sensitiveData = value.c_str();
    EXPECT_FALSE(sensitiveData == sensitiveData2);

    sensitiveData = sensitiveData2;
    sensitiveData = "";
    sensitiveData2 = "";
    EXPECT_TRUE(sensitiveData == sensitiveData2);

    std::string str = "data";
    std::unique_ptr<char[]> charArray = std::make_unique<char[]>(str.size() + 1);
    std::copy(str.begin(), str.end(), charArray.get());
    charArray[str.size()] = '\0';
    SensitiveData d1(std::move(charArray), str.size() + 1);
    SensitiveData d2 = SensitiveData(d1);
    SensitiveData d3 = d1;
    std::unique_ptr<char[]> outData;
    size_t outSize;
    d1.MoveTo(outData, outSize);
    ASSERT_EQ(outSize, 5);
}

TEST_F(UtilsTest, GenerateErrorInfoTest)
{
    auto status = datasystem::Status(datasystem::StatusCode::K_NOT_READY, "ERROR MESSAGE");
    std::vector<std::string> remainIds = {"remainid"};
    std::vector<std::string> ids;
    auto err = GenerateErrorInfo(1, status, 1000, remainIds, ids);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    err = GenerateErrorInfo(0, status, 1000, remainIds, ids);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_DATASYSTEM_FAILED);
    status = datasystem::Status(datasystem::StatusCode::K_OUT_OF_MEMORY, "ERROR MESSAGE");
    err = GenerateErrorInfo(0, status, 1000, remainIds, ids);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_SHARED_MEMORY_LIMITED);
    err = GenerateSetErrorInfo(status);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_SHARED_MEMORY_LIMITED);
}

TEST_F(UtilsTest, InitWithDriverTest)
{
    auto security = std::make_shared<Security>();
    auto librtConfig = std::make_shared<LibruntimeConfig>();
    librtConfig->enableMTLS = true;
    librtConfig->verifyFilePath = "test";
    librtConfig->certificateFilePath = "test";
    librtConfig->privateKeyPath = "test";
    std::strcpy(librtConfig->privateKeyPaaswd, "paaswd");
    librtConfig->serverName = "test";
    librtConfig->encryptEnable = "test";
    librtConfig->encryptEnable = "test";
    librtConfig->runtimePublicKey = "test";
    librtConfig->runtimePrivateKey = "test";
    librtConfig->ak_ = "ak";
    librtConfig->sk_ = "sk";
    auto err = security->InitWithDriver(librtConfig);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
}

TEST_F(UtilsTest, GetAuthConnectOptsTest)
{
    ConnectOptions connnections;
    std::string ak;
    datasystem::SensitiveValue sk;
    datasystem::SensitiveValue token = "token";
    GetAuthConnectOpts(connnections, ak, sk, token);
    ASSERT_EQ(connnections.accessKey, ak);
}

TEST_F(UtilsTest, unhexlifyTest)
{
    char ascii[2];
    auto res = unhexlify("00", ascii);
    ASSERT_EQ(res, 0);
    ASSERT_EQ(ascii[0], '\0');
    res = unhexlify("FF", ascii);
    ASSERT_EQ(res, 0);
    ASSERT_EQ(ascii[0], '\xFF');
    res = unhexlify("1a", ascii);
    ASSERT_EQ(res, 0);
    ASSERT_EQ(ascii[0], '\x1a');
    res = unhexlify("AB", ascii);
    ASSERT_EQ(res, 0);
    ASSERT_EQ(ascii[0], '\xAB');
    res = unhexlify("G1", ascii);
    ASSERT_EQ(res, -1);
    res = unhexlify("1G", ascii);
    ASSERT_EQ(res, -1);
    res = unhexlify("1", ascii);
    ASSERT_EQ(res, -1);
}

TEST_F(UtilsTest, LoadEnvFromFile)
{
    // Create temporary .env file with various scenarios
    std::string tempFile = "/tmp/test_env_" + std::to_string(getpid()) + ".env";
    std::ofstream file(tempFile);
    file << "# Comment line\n";
    file << "TEST_KEY1=value1\n";
    file << "TEST_KEY2=value with spaces\n";
    file << "TEST_SINGLE_QUOTE='quoted value'\n";
    file << "TEST_DOUBLE_QUOTE=\"quoted value\"\n";
    file << "TEST_WITH_EQUALS=key=value\n";
    file << "TEST_WHITESPACE= value with spaces \n";
    file << "\n";  // Empty line
    file.close();

    // Clear test environment variables
    unsetenv("TEST_KEY1");
    unsetenv("TEST_KEY2");
    unsetenv("TEST_SINGLE_QUOTE");
    unsetenv("TEST_DOUBLE_QUOTE");
    unsetenv("TEST_WITH_EQUALS");
    unsetenv("TEST_WHITESPACE");

    {
        YR::LoadEnvFromFile(tempFile);

        ASSERT_STREQ(std::getenv("TEST_KEY1"), "value1");
        ASSERT_STREQ(std::getenv("TEST_KEY2"), "value with spaces");
        ASSERT_STREQ(std::getenv("TEST_SINGLE_QUOTE"), "quoted value");
        ASSERT_STREQ(std::getenv("TEST_DOUBLE_QUOTE"), "quoted value");
        ASSERT_STREQ(std::getenv("TEST_WITH_EQUALS"), "key=value");
        ASSERT_STREQ(std::getenv("TEST_WHITESPACE"), "value with spaces");
    }

    // Cleanup
    unlink(tempFile.c_str());
    unsetenv("TEST_KEY1");
    unsetenv("TEST_KEY2");
    unsetenv("TEST_SINGLE_QUOTE");
    unsetenv("TEST_DOUBLE_QUOTE");
    unsetenv("TEST_WITH_EQUALS");
    unsetenv("TEST_WHITESPACE");
}

}  // namespace test
}  // namespace YR