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

#include <memory>

#include <gmock/gmock.h>
#include <gtest/gtest.h>
#include "src/libruntime/err_type.h"
#include "src/utility/logger/logger.h"
#define private public
#include "api/cpp/include/yr/api/function_manager.h"
#include "api/cpp/src/code_manager.h"
#include "api/cpp/src/state_loader.h"

using namespace testing;
namespace YR {
namespace test {
using namespace testing;
using namespace YR::utility;
using namespace YR::internal;

class StateLoaderTest : public testing::Test {
public:
    StateLoaderTest(){};
    ~StateLoaderTest(){};
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
};

TEST_F(StateLoaderTest, DumpLoadInstanceTest)
{
    std::string instanceId = "abc123";
    std::shared_ptr<YR::Libruntime::Buffer> data;
    // Should dump nothing
    YR::Libruntime::ErrorInfo errInfo = DumpInstance(instanceId, data);
    ASSERT_EQ(errInfo.Code(), YR::Libruntime::ErrorCode::ERR_OK);
    ASSERT_EQ(data, nullptr);

    std::string mockClassName = "MockClass";
    CodeManager::Singleton().SetClassName(mockClassName);

    // mock a instance Buffer in CodeManager
    CodeManager::Singleton().instancePtr = std::make_shared<msgpack::sbuffer>();

    // chkptFuncMap empty, so Should dump fail
    errInfo = DumpInstance(instanceId, data);
    ASSERT_EQ(errInfo.Code(), YR::Libruntime::ErrorCode::ERR_USER_FUNCTION_EXCEPTION);

    // mock a Checkpoint Function for the MockClass in FunctionManager
    FunctionManager::Singleton().ckptFuncMap_.emplace(mockClassName, [](const msgpack::sbuffer &) {
        msgpack::sbuffer sbuf;
        msgpack::packer<msgpack::sbuffer> pk(&sbuf);
        pk.pack(std::string("helloworld"));
        return sbuf;
    });

    // Should dump success
    errInfo = DumpInstance(instanceId, data);
    ASSERT_EQ(errInfo.Code(), YR::Libruntime::ErrorCode::ERR_OK);

    // --- Load Instance ---
    // Should load nothing
    std::shared_ptr<YR::Libruntime::Buffer> loadData = std::make_shared<YR::Libruntime::NativeBuffer>(0);
    errInfo = LoadInstance(loadData);
    ASSERT_EQ(errInfo.Code(), YR::Libruntime::ErrorCode::ERR_OK);

    // recoverFuncMap empty, so Should load fail
    std::string tempName = "tempName";
    CodeManager::Singleton().SetClassName(tempName);
    errInfo = LoadInstance(data);
    ASSERT_EQ(errInfo.Code(), YR::Libruntime::ErrorCode::ERR_USER_FUNCTION_EXCEPTION);

    // mock an empty buffer recover Function for the MockClass in FunctionManager
    FunctionManager::Singleton().recoverFuncMap_[mockClassName] = [](const msgpack::sbuffer &) {
        msgpack::sbuffer sbuf;
        return sbuf;
    };
    errInfo = LoadInstance(data);
    ASSERT_EQ(errInfo.Code(), YR::Libruntime::ErrorCode::ERR_USER_FUNCTION_EXCEPTION);

    // mock a recover Function for the MockClass in FunctionManager
    FunctionManager::Singleton().recoverFuncMap_[mockClassName] = [](const msgpack::sbuffer &) {
        msgpack::sbuffer sbuf;
        msgpack::packer<msgpack::sbuffer> pk(&sbuf);
        pk.pack(std::string("helloworld"));
        return sbuf;
    };

    errInfo = LoadInstance(data);
    ASSERT_EQ(errInfo.Code(), YR::Libruntime::ErrorCode::ERR_OK);
    // Should recover the ClassName from data
    ASSERT_EQ(CodeManager::Singleton().GetClassName(), mockClassName);
}  // namespace test

}  // namespace test
}  // namespace YR