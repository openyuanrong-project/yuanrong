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
#include <filesystem>
#include <stdexcept>
#include <string>

#include <gmock/gmock.h>
#include <gtest/gtest.h>
#define private public
#include "api/cpp/include/yr/api/function_manager.h"
#include "api/cpp/src/code_manager.h"
#include "src/utility/timer_worker.h"

namespace YR {
namespace test {
using namespace testing;
using namespace YR::internal;
using YR::Libruntime::ErrorCode;
using YR::Libruntime::ModuleCode;
using YR::utility::CloseGlobalTimer;
using YR::utility::InitGlobalTimer;
namespace fs = std::filesystem;
class CodeManagerTest : public testing::Test {
public:
    CodeManagerTest(){};
    ~CodeManagerTest(){};
    void SetUp() override {}
    void TearDown() override {}
};

static std::string FindDummyLibPath()
{
    // In bazel runfiles, look for test_dummy_lib.so alongside the test binary
    fs::path currentPath = fs::current_path();
    for (auto &entry : fs::recursive_directory_iterator(currentPath)) {
        if (entry.is_regular_file() && entry.path().filename() == "test_dummy_lib.so") {
            return entry.path().parent_path().string();
        }
    }
    // Fallback: source tree layout
    auto path = currentPath.string();
    auto idx = path.find("yuanrong/");
    if (idx != std::string::npos) {
        return path.substr(0, idx) + "yuanrong/test";
    }
    return "";
}

TEST_F(CodeManagerTest, LoadFunctionsSuccessfullyTest)
{
    auto libPath = FindDummyLibPath();
    ASSERT_FALSE(libPath.empty()) << "Cannot find dummy lib directory";
    InitGlobalTimer();
    YR::Libruntime::ErrorInfo err;
    err = YR::internal::CodeManager::Singleton().LoadFunctions({libPath});
    ASSERT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_OK);

    err = YR::internal::CodeManager::Singleton().LoadFunctions({libPath + "/test_dummy_lib.so"});
    ASSERT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_OK);
    CloseGlobalTimer();
}

TEST_F(CodeManagerTest, LoadFunctionsFailedTest)
{
    YR::Libruntime::ErrorInfo err;
    err = YR::internal::CodeManager::Singleton().LoadFunctions({});
    ASSERT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_USER_CODE_LOAD);

    err = YR::internal::CodeManager::Singleton().LoadFunctions({"/ddd"});
    ASSERT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_USER_CODE_LOAD);

    fs::path currentPath = fs::current_path();
    err = YR::internal::CodeManager::Singleton().LoadFunctions({currentPath.string()});
    ASSERT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_USER_CODE_LOAD);

    err = YR::internal::CodeManager::Singleton().LoadFunctions({currentPath.string() + "code_manager_test.cpp"});
    ASSERT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_USER_CODE_LOAD);
}

TEST_F(CodeManagerTest, ExecuteFunctionTest)
{
    YR::Libruntime::ErrorInfo err;
    YR::Libruntime::FunctionMeta function;
    function.funcName = "funcName";
    libruntime::InvokeType type = libruntime::InvokeType::CreateInstance;
    std::vector<std::shared_ptr<YR::Libruntime::DataObject>> rawArgs;
    std::vector<std::shared_ptr<YR::Libruntime::DataObject>> returnObjects;
    returnObjects.push_back(std::make_shared<YR::Libruntime::DataObject>());

    err = YR::internal::CodeManager::Singleton().ExecuteFunction(function, type, rawArgs, returnObjects);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_USER_FUNCTION_EXCEPTION);

    // mock a normal Function for the MockClass in FunctionManager
    FunctionManager::Singleton().funcMap_["funcName"] =
        [](const std::string &returnObjId,
           const std::vector<msgpack::sbuffer> &rawBuffers) -> std::pair<std::shared_ptr<msgpack::sbuffer>, bool> {
        throw std::exception();
    };
    err = YR::internal::CodeManager::Singleton().ExecuteFunction(function, type, rawArgs, returnObjects);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_USER_FUNCTION_EXCEPTION);

    // mock a normal Function for the MockClass in FunctionManager
    FunctionManager::Singleton().funcMap_["funcName"] =
        [](const std::string &returnObjId,
           const std::vector<msgpack::sbuffer> &rawBuffers) -> std::pair<std::shared_ptr<msgpack::sbuffer>, bool> {
        throw YR::Exception(ErrorCode::ERR_INCORRECT_INVOKE_USAGE, ModuleCode::RUNTIME_INVOKE,
                            "YR_INVOKE function is duplicated");
    };
    err = YR::internal::CodeManager::Singleton().ExecuteFunction(function, type, rawArgs, returnObjects);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_USER_FUNCTION_EXCEPTION);

    // mock a normal Function for the MockClass in FunctionManager
    FunctionManager::Singleton().funcMap_["funcName"] =
        [](const std::string &returnObjId,
           const std::vector<msgpack::sbuffer> &rawBuffers) -> std::pair<std::shared_ptr<msgpack::sbuffer>, bool> {
        std::string val = "hello";
        std::shared_ptr<msgpack::sbuffer> data = std::make_shared<msgpack::sbuffer>(YR::internal::Serialize(val));
        return std::make_pair(data, true);
    };
    err = YR::internal::CodeManager::Singleton().ExecuteFunction(function, type, rawArgs, returnObjects);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);

    type = libruntime::InvokeType::InvokeFunctionStateless;
    err = YR::internal::CodeManager::Singleton().ExecuteFunction(function, type, rawArgs, returnObjects);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);

    type = libruntime::InvokeType::CreateInstanceStateless;
    err = YR::internal::CodeManager::Singleton().ExecuteFunction(function, type, rawArgs, returnObjects);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);

    type = libruntime::InvokeType::InvokeFunction;
    err = YR::internal::CodeManager::Singleton().ExecuteFunction(function, type, rawArgs, returnObjects);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_USER_FUNCTION_EXCEPTION);

    // mock a member Function for the MockClass in FunctionManager
    FunctionManager::Singleton().memberFuncMap_["funcName"] =
        [](const std::string &returnObjId, const msgpack::sbuffer &sbuffer,
           const std::vector<msgpack::sbuffer> &rawBuffers) -> std::pair<std::shared_ptr<msgpack::sbuffer>, bool> {
        throw std::exception();
    };
    err = YR::internal::CodeManager::Singleton().ExecuteFunction(function, type, rawArgs, returnObjects);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_USER_FUNCTION_EXCEPTION);

    // mock a member Function for the MockClass in FunctionManager
    FunctionManager::Singleton().memberFuncMap_["funcName"] =
        [](const std::string &returnObjId, const msgpack::sbuffer &sbuffer,
           const std::vector<msgpack::sbuffer> &rawBuffers) -> std::pair<std::shared_ptr<msgpack::sbuffer>, bool> {
        throw YR::Exception(ErrorCode::ERR_INCORRECT_INVOKE_USAGE, ModuleCode::RUNTIME_INVOKE,
                            "YR_INVOKE function is duplicated");
    };
    err = YR::internal::CodeManager::Singleton().ExecuteFunction(function, type, rawArgs, returnObjects);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_USER_FUNCTION_EXCEPTION);

    // mock a member Function for the MockClass in FunctionManager
    FunctionManager::Singleton().memberFuncMap_["funcName"] =
        [](const std::string &returnObjId, const msgpack::sbuffer &sbuffer,
           const std::vector<msgpack::sbuffer> &rawBuffers) -> std::pair<std::shared_ptr<msgpack::sbuffer>, bool> {
        std::string val = "hello";
        std::shared_ptr<msgpack::sbuffer> data = std::make_shared<msgpack::sbuffer>(YR::internal::Serialize(val));
        return std::make_pair(data, true);
    };
    err = YR::internal::CodeManager::Singleton().ExecuteFunction(function, type, rawArgs, returnObjects);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
}

TEST_F(CodeManagerTest, ExecuteShutdownFunctionTest)
{
    YR::internal::CodeManager::Singleton().instancePtr = nullptr;
    YR::Libruntime::ErrorInfo err = YR::internal::CodeManager::Singleton().ExecuteShutdownFunction(100);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_USER_FUNCTION_EXCEPTION);

    YR::internal::CodeManager::Singleton().instancePtr = std::make_shared<msgpack::sbuffer>(100);
    YR::internal::CodeManager::Singleton().className = "clsName";
    err = YR::internal::CodeManager::Singleton().ExecuteShutdownFunction(100);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);

    FunctionManager::Singleton().shutdownCallerMap_["clsName"] = [](const msgpack::sbuffer &,
                                                                   uint64_t gracePeriodSecond) { return; };
    err = YR::internal::CodeManager::Singleton().ExecuteShutdownFunction(100);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);

    FunctionManager::Singleton().shutdownCallerMap_["clsName"] =
        [](const msgpack::sbuffer &, uint64_t gracePeriodSecond) { throw std::exception(); };
    err = YR::internal::CodeManager::Singleton().ExecuteShutdownFunction(100);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_USER_FUNCTION_EXCEPTION);

    FunctionManager::Singleton().shutdownCallerMap_["clsName"] =
        [](const msgpack::sbuffer &, uint64_t gracePeriodSecond) { throw YR::Exception("msg"); };
    err = YR::internal::CodeManager::Singleton().ExecuteShutdownFunction(100);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_USER_FUNCTION_EXCEPTION);

    YR::internal::CodeManager::Singleton().instancePtr = nullptr;
    YR::internal::CodeManager::Singleton().className = "";
}
}  // namespace test
}  // namespace YR