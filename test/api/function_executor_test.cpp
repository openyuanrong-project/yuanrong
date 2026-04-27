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
#if __has_include(<filesystem>)
#include <filesystem>
#else
#include <experimental/filesystem>
#endif
#include <stdexcept>
#include <string>

#include <gmock/gmock.h>
#include <gtest/gtest.h>
#include <msgpack.hpp>
#define private public
#include "api/cpp/src/executor/function_executor.h"
#include "api/cpp/src/executor/executor_holder.h"
#include "api/cpp/src/executor/posix_executor.h"
#include "src/utility/timer_worker.h"
#include "yr/api/function_manager.h"

namespace YR {
namespace test {
using namespace testing;
using namespace YR::internal;
using YR::Libruntime::ErrorCode;
using YR::Libruntime::ModuleCode;
using YR::utility::CloseGlobalTimer;
using YR::utility::InitGlobalTimer;
namespace fs = std::filesystem;
class FunctionExecutorTest : public testing::Test {
public:
    FunctionExecutorTest(){};
    ~FunctionExecutorTest(){};

    static void SetUpTestCase()
    {
        ClearFunctions();
    }

    void SetUp() override
    {
        exec_ = std::make_shared<FunctionExecutor>();
    }

    void TearDown() override
    {
        exec_.reset();
        ClearFunctions();
    }

    static void ClearFunctions()
    {
        FunctionManager::Singleton().funcIdToName_.clear();
        FunctionManager::Singleton().memberFuncIdToName_.clear();
        FunctionManager::Singleton().clsMap_.clear();
        FunctionManager::Singleton().funcMap_.clear();
        FunctionManager::Singleton().memberFuncMap_.clear();
        FunctionManager::Singleton().shutdownCallerMap_.clear();
        FunctionManager::Singleton().ckptFuncMap_.clear();
        FunctionManager::Singleton().recoverFuncMap_.clear();
        FunctionManager::Singleton().recoverCallbackFuncMap_.clear();
    }

private:
    std::shared_ptr<FunctionExecutor> exec_;
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

TEST_F(FunctionExecutorTest, LoadFunctionsSuccessfullyTest)
{
    auto libPath = FindDummyLibPath();
    ASSERT_FALSE(libPath.empty()) << "Cannot find dummy lib directory";
    InitGlobalTimer();
    YR::Libruntime::ErrorInfo err;
    err = LoadFunctions({libPath});
    ASSERT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_OK) << err.Msg();

    err = exec_->LoadFunctions({libPath + "/test_dummy_lib.so"});
    ASSERT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_OK) << err.Msg();
    CloseGlobalTimer();
}

TEST_F(FunctionExecutorTest, LoadFunctionsFailedTest)
{
    YR::Libruntime::ErrorInfo err;
    err = exec_->LoadFunctions({});
    ASSERT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_USER_CODE_LOAD);

    err = exec_->LoadFunctions({"/ddd"});
    ASSERT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_USER_CODE_LOAD);

    fs::path currentPath = fs::current_path();
    err = exec_->LoadFunctions({currentPath.string()});
    ASSERT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_USER_CODE_LOAD);

    err = exec_->LoadFunctions({(currentPath / "code_manager_test.cpp").string()});
    ASSERT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_USER_CODE_LOAD);
}

TEST_F(FunctionExecutorTest, ExecuteFunctionTest)
{
    YR::Libruntime::ErrorInfo err;
    YR::Libruntime::FunctionMeta function;
    function.funcName = "funcName";
    libruntime::InvokeType type = libruntime::InvokeType::CreateInstance;
    std::vector<std::shared_ptr<YR::Libruntime::DataObject>> rawArgs;
    std::vector<std::shared_ptr<YR::Libruntime::DataObject>> returnObjects;
    returnObjects.push_back(std::make_shared<YR::Libruntime::DataObject>());

    err = ExecuteFunction(function, type, rawArgs, returnObjects);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_USER_FUNCTION_EXCEPTION);

    // mock a normal Function for the MockClass in FunctionManager
    FunctionManager::Singleton().funcMap_["funcName"] =
        [](const std::string &returnObjId,
           const std::vector<msgpack::sbuffer> &rawBuffers) -> std::pair<std::shared_ptr<msgpack::sbuffer>, bool> {
        throw std::exception();
    };
    err = exec_->ExecuteFunction(function, type, rawArgs, returnObjects);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_USER_FUNCTION_EXCEPTION);

    // mock a normal Function for the MockClass in FunctionManager
    FunctionManager::Singleton().funcMap_["funcName"] =
        [](const std::string &returnObjId,
           const std::vector<msgpack::sbuffer> &rawBuffers) -> std::pair<std::shared_ptr<msgpack::sbuffer>, bool> {
        throw YR::Exception(ErrorCode::ERR_INCORRECT_INVOKE_USAGE, ModuleCode::RUNTIME_INVOKE,
                            "YR_INVOKE function is duplicated");
    };
    err = exec_->ExecuteFunction(function, type, rawArgs, returnObjects);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_USER_FUNCTION_EXCEPTION);

    // mock a normal Function for the MockClass in FunctionManager
    FunctionManager::Singleton().funcMap_["funcName"] =
        [](const std::string &returnObjId,
           const std::vector<msgpack::sbuffer> &rawBuffers) -> std::pair<std::shared_ptr<msgpack::sbuffer>, bool> {
        std::string val = "hello";
        std::shared_ptr<msgpack::sbuffer> data = std::make_shared<msgpack::sbuffer>(YR::internal::Serialize(val));
        return std::make_pair(data, true);
    };
    err = exec_->ExecuteFunction(function, type, rawArgs, returnObjects);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);

    type = libruntime::InvokeType::InvokeFunctionStateless;
    err = exec_->ExecuteFunction(function, type, rawArgs, returnObjects);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);

    type = libruntime::InvokeType::CreateInstanceStateless;
    err = exec_->ExecuteFunction(function, type, rawArgs, returnObjects);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);

    type = libruntime::InvokeType::InvokeFunction;
    err = exec_->ExecuteFunction(function, type, rawArgs, returnObjects);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_USER_FUNCTION_EXCEPTION);

    // mock a member Function for the MockClass in FunctionManager
    FunctionManager::Singleton().memberFuncMap_["funcName"] =
        [](const std::string &returnObjId, const msgpack::sbuffer &sbuffer,
           const std::vector<msgpack::sbuffer> &rawBuffers) -> std::pair<std::shared_ptr<msgpack::sbuffer>, bool> {
        throw std::exception();
    };
    err = exec_->ExecuteFunction(function, type, rawArgs, returnObjects);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_USER_FUNCTION_EXCEPTION);

    // mock a member Function for the MockClass in FunctionManager
    FunctionManager::Singleton().memberFuncMap_["funcName"] =
        [](const std::string &returnObjId, const msgpack::sbuffer &sbuffer,
           const std::vector<msgpack::sbuffer> &rawBuffers) -> std::pair<std::shared_ptr<msgpack::sbuffer>, bool> {
        throw YR::Exception(ErrorCode::ERR_INCORRECT_INVOKE_USAGE, ModuleCode::RUNTIME_INVOKE,
                            "YR_INVOKE function is duplicated");
    };
    err = exec_->ExecuteFunction(function, type, rawArgs, returnObjects);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_USER_FUNCTION_EXCEPTION);

    // mock a member Function for the MockClass in FunctionManager
    FunctionManager::Singleton().memberFuncMap_["funcName"] =
        [](const std::string &returnObjId, const msgpack::sbuffer &sbuffer,
           const std::vector<msgpack::sbuffer> &rawBuffers) -> std::pair<std::shared_ptr<msgpack::sbuffer>, bool> {
        std::string val = "hello";
        std::shared_ptr<msgpack::sbuffer> data = std::make_shared<msgpack::sbuffer>(YR::internal::Serialize(val));
        return std::make_pair(data, true);
    };
    err = exec_->ExecuteFunction(function, type, rawArgs, returnObjects);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
}

class TestCounter {
public:
    static TestCounter *FactoryCreate()
    {
        return new TestCounter();
    }

    int Get()
    {
        return count;
    }

    void ShutDownThrow(uint64_t gracePeriodSecond)
    {
        throw std::exception();
    }

    void ShutDownThrow2(uint64_t gracePeriodSecond)
    {
        throw YR::Exception("msg");
    }

    void ShutDown(uint64_t gracePeriodSecond)
    {
        return;
    }
    int count = 0;
    MSGPACK_DEFINE(count);
};

TEST_F(FunctionExecutorTest, ExecuteShutdownFunctionTest)
{
    YR::Libruntime::ErrorInfo err;
    YR::Libruntime::FunctionMeta function;

    err = ExecuteShutdownFunction(1);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_USER_FUNCTION_EXCEPTION) << err.Msg();

    function.funcName = "&TestCounter::FactoryCreate";
    function.className = "TestCounter";
    libruntime::InvokeType type = libruntime::InvokeType::CreateInstance;
    std::vector<std::shared_ptr<YR::Libruntime::DataObject>> rawArgs;
    std::vector<std::shared_ptr<YR::Libruntime::DataObject>> returnObjects;
    returnObjects.push_back(std::make_shared<YR::Libruntime::DataObject>());
    FunctionManager::Singleton().RegisterInvokeFunction("&TestCounter::FactoryCreate", &TestCounter::FactoryCreate);
    err = exec_->ExecuteFunction(function, type, rawArgs, returnObjects);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK) << err.Msg();

    // TestCounter cnt;
    err = exec_->ExecuteShutdownFunction(1);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK) << err.Msg();

    FunctionManager::Singleton().RegisterShutdownFunctions("&TestCounter::ShutDownThrow", &TestCounter::ShutDownThrow);
    err = exec_->ExecuteShutdownFunction(1);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_USER_FUNCTION_EXCEPTION) << err.Msg();

    FunctionManager::Singleton().RegisterShutdownFunctions("&TestCounter::ShutDownThrow2",
                                                           &TestCounter::ShutDownThrow2);
    err = exec_->ExecuteShutdownFunction(1);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_USER_FUNCTION_EXCEPTION) << err.Msg();

    FunctionManager::Singleton().shutdownCallerMap_.clear();
    FunctionManager::Singleton().RegisterShutdownFunctions("&TestCounter::ShutDown", &TestCounter::ShutDown);
    err = exec_->ExecuteShutdownFunction(1);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK) << err.Msg();
}

TEST_F(FunctionExecutorTest, CheckpointRecoverTest)
{
    YR::Libruntime::ErrorInfo err;
    YR::Libruntime::FunctionMeta function;
    function.funcName = "&TestCounter::FactoryCreate";
    function.className = "TestCounter";
    libruntime::InvokeType type = libruntime::InvokeType::CreateInstance;
    std::vector<std::shared_ptr<YR::Libruntime::DataObject>> rawArgs;
    std::vector<std::shared_ptr<YR::Libruntime::DataObject>> returnObjects;
    returnObjects.push_back(std::make_shared<YR::Libruntime::DataObject>());
    FunctionManager::Singleton().RegisterInvokeFunction("&TestCounter::FactoryCreate", &TestCounter::FactoryCreate);
    FunctionManager::Singleton().RegisterInvokeFunction("&TestCounter::Get", &TestCounter::Get);
    err = exec_->ExecuteFunction(function, type, rawArgs, returnObjects);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK) << err.Msg();

    std::shared_ptr<YR::Libruntime::Buffer> emptyData = std::make_shared<YR::Libruntime::NativeBuffer>(0);
    err = Recover(emptyData);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK) << err.Msg();

    std::shared_ptr<YR::Libruntime::Buffer> data;
    err = exec_->Checkpoint("instanceid", data);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK) << err.Msg();

    err = exec_->Recover(data);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK) << err.Msg();

    err = Signal(10, nullptr);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK) << err.Msg();
}

TEST_F(FunctionExecutorTest, PosixExecutorTest)
{
    std::shared_ptr<Executor> posixExecutor = std::make_shared<PosixExecutor>();
    auto err = posixExecutor->LoadFunctions({""});
    ASSERT_EQ(err.Code(), ErrorCode::ERR_PARAM_INVALID) << err.Msg();

    YR::Libruntime::FunctionMeta function;
    function.funcName = "funcName";
    libruntime::InvokeType type = libruntime::InvokeType::CreateInstance;
    std::vector<std::shared_ptr<YR::Libruntime::DataObject>> rawArgs;
    std::vector<std::shared_ptr<YR::Libruntime::DataObject>> returnObjects;
    returnObjects.push_back(std::make_shared<YR::Libruntime::DataObject>());
    err = posixExecutor->ExecuteFunction(function, type, rawArgs, returnObjects);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_PARAM_INVALID) << err.Msg();
    std::shared_ptr<YR::Libruntime::Buffer> data;
    err = posixExecutor->Checkpoint("instanceId", data);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_PARAM_INVALID) << err.Msg();
    err = posixExecutor->Recover(data);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_PARAM_INVALID) << err.Msg();
    err = posixExecutor->ExecuteShutdownFunction(100);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_PARAM_INVALID) << err.Msg();
    err = posixExecutor->Signal(1, data);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_PARAM_INVALID) << err.Msg();
}

}  // namespace test
}  // namespace YR