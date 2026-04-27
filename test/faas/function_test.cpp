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

#include <gmock/gmock.h>
#include <gtest/gtest.h>

#include "Function.h"
#include "FunctionError.h"
#include "common/mock_libruntime.h"
#include "api/cpp/src/faas/context_impl.h"
#include "src/libruntime/libruntime_manager.h"
#include "src/utility/id_generator.h"

namespace YR {
namespace test {
using namespace YR::utility;
class FunctionTest : public testing::Test {
public:
    FunctionTest(){};
    ~FunctionTest(){};

    void SetUp() override
    {
        auto lc = std::make_shared<YR::Libruntime::LibruntimeConfig>();
        lc->jobId = YR::utility::IDGenerator::GenApplicationId();
        auto clientsMgr = std::make_shared<YR::Libruntime::ClientsManager>();
        auto metricsAdaptor = std::make_shared<YR::Libruntime::MetricsAdaptor>();
        auto sec = std::make_shared<YR::Libruntime::Security>();
        auto socketClient = std::make_shared<YR::Libruntime::DomainSocketClient>("/home/snuser/socket/runtime.sock");
        lr = std::make_shared<YR::Libruntime::MockLibruntime>(lc, clientsMgr, metricsAdaptor, sec, socketClient);
        YR::Libruntime::LibruntimeManager::Instance().SetLibRuntime(lr);
    }

    void TearDown() override
    {
        YR::Libruntime::LibruntimeManager::Instance().Finalize();
        lr.reset();
    }

    std::shared_ptr<YR::Libruntime::MockLibruntime> lr;
};

TEST_F(FunctionTest, InvokeTest)
{
    auto contextEnv = std::make_shared<Function::ContextEnv>();
    contextEnv->SetFunctionName("func");
    contextEnv->SetFuncPackage("service");
    contextEnv->SetProjectID("123");
    std::unordered_map<std::string, std::string> params;
    auto contextInvokeParams = std::make_shared<Function::ContextInvokeParams>(params);
    auto context = Function::ContextImpl(contextInvokeParams, contextEnv);
    std::string result = "{\"innerCode\":\"0\", \"body\":\"result\"}";
    auto returnObj = std::make_shared<YR::Libruntime::DataObject>(0, result.size());
    returnObj->data->MemoryCopy(static_cast<void *>(result.data()), result.size());

    std::vector<std::shared_ptr<Function::Function>> funcs;
    funcs.emplace_back(std::make_shared<Function::Function>(context));
    funcs.emplace_back(std::make_shared<Function::Function>(context, "func"));
    funcs.emplace_back(std::make_shared<Function::Function>(context, "func:latest"));
    for (auto &func : funcs) {
        std::vector<std::shared_ptr<YR::Libruntime::DataObject>> rets;
        rets.push_back(returnObj);
        EXPECT_CALL(*lr.get(), InvokeByFunctionName(_, _, _, _)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
        EXPECT_CALL(*lr.get(), Get(_, _, _))
            .WillOnce(Return(
                std::make_pair<YR::Libruntime::ErrorInfo, std::vector<std::shared_ptr<YR::Libruntime::DataObject>>>(
                    YR::Libruntime::ErrorInfo(), std::move(rets))));
        auto ref = func->Invoke("aaa");
        auto ret = func->GetObjectRef(ref);
        ASSERT_EQ(ret, "result");
        ASSERT_EQ(func->GetContext() != nullptr, true);
    }
}

TEST_F(FunctionTest, InvokeFailedTest)
{
    auto contextEnv = std::make_shared<Function::ContextEnv>();
    contextEnv->SetFunctionName("func");
    contextEnv->SetFuncPackage("service");
    contextEnv->SetProjectID("123");
    std::unordered_map<std::string, std::string> params;
    auto contextInvokeParams = std::make_shared<Function::ContextInvokeParams>(params);
    auto context = Function::ContextImpl(contextInvokeParams, contextEnv);
    std::string result = "result";
    auto returnObj = std::make_shared<YR::Libruntime::DataObject>(0, result.size());
    returnObj->data->MemoryCopy(static_cast<void *>(result.data()), result.size());

    auto func = Function::Function(context);
    std::make_shared<Function::Function>(context, "func:latest");
    std::vector<std::shared_ptr<YR::Libruntime::DataObject>> rets;
    rets.push_back(returnObj);
    EXPECT_CALL(*lr.get(), InvokeByFunctionName(_, _, _, _))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "aaa")));
    EXPECT_THROW(func.Invoke("aaa"), Function::FunctionError);

    auto func2 = Function::Function(context, "");
    EXPECT_THROW(func2.Invoke("aaa"), Function::FunctionError);

    auto func3 = Function::Function(context, "func:func:lateset");
    EXPECT_THROW(func3.Invoke("aaa"), Function::FunctionError);

    auto func4 = Function::Function(context, "func:xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx");
    EXPECT_THROW(func4.Invoke("aaa"), Function::FunctionError);

    auto func5 = Function::Function(context, "func&^%$:latest");
    EXPECT_THROW(func5.Invoke("aaa"), Function::FunctionError);

    auto func6 = Function::Function(context, "func&^%$:latest");
    EXPECT_THROW(func6.Invoke("aaa"), Function::FunctionError);

    auto func7 = Function::Function(context, "func&^%$");
    EXPECT_THROW(func7.Invoke("aaa"), Function::FunctionError);

    auto func8 = Function::Function(context, "!func&^%$:latest", "instanceName");
    EXPECT_THROW(func8.Invoke("aaa"), Function::FunctionError);
}

TEST_F(FunctionTest, MemberFunctionTest)
{
    auto contextEnv = std::make_shared<Function::ContextEnv>();
    contextEnv->SetFunctionName("func");
    contextEnv->SetFuncPackage("service");
    contextEnv->SetProjectID("123");
    std::unordered_map<std::string, std::string> params;
    auto contextInvokeParams = std::make_shared<Function::ContextInvokeParams>(params);
    auto context = Function::ContextImpl(contextInvokeParams, contextEnv);
    auto func = Function::Function(context, "func:latest", "instanceName");
    EXPECT_THROW(func.GetInstance("func:latest", "instanceName"), Function::FunctionError);
    EXPECT_THROW(func.GetLocalInstance("func:latest", "instanceName"), Function::FunctionError);
    EXPECT_THROW(func.Terminate(), Function::FunctionError);
    EXPECT_THROW(func.SaveState(), Function::FunctionError);
    EXPECT_THROW(func.GetInstanceId(), Function::FunctionError);
}

}  // namespace test
}  // namespace YR