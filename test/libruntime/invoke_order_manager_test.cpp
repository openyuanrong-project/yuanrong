/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
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

#define private public

#include "src/libruntime/invoke_order_manager.h"

#undef private

using namespace YR::Libruntime;
using namespace YR::utility;

namespace YR {
namespace test {
class InvokeOrderManagerTest : public testing::Test {
};

TEST_F(InvokeOrderManagerTest, OrderTest)
{
    std::string instanceId = "fake_idd";
    auto invokeOrderMgr = std::make_shared<InvokeOrderManager>(std::make_shared<LibruntimeConfig>());
    auto specCreate = std::make_shared<InvokeSpec>();
    specCreate->invokeType = libruntime::InvokeType::CreateInstance;
    specCreate->opts.needOrder = true;
    specCreate->returnIds = std::vector<DataObject>{DataObject(instanceId)};

    invokeOrderMgr->CreateInstance(specCreate);

    auto specInvoke = std::make_shared<InvokeSpec>();
    specInvoke->invokeType = libruntime::InvokeType::InvokeFunction;
    specInvoke->instanceId = instanceId;
    invokeOrderMgr->Invoke(specInvoke);

    invokeOrderMgr->NotifyInvokeSuccess(specCreate);
    ASSERT_EQ(invokeOrderMgr->instances[instanceId]->unfinishedSeqNo, 1);
    invokeOrderMgr->NotifyInvokeSuccess(specInvoke);
    ASSERT_EQ(invokeOrderMgr->instances[instanceId]->unfinishedSeqNo, 2);
}

TEST_F(InvokeOrderManagerTest, CreateAndNotifyAndRomveGroupInstanceTest)
{
    auto invokeOrderMgr = std::make_shared<InvokeOrderManager>(std::make_shared<LibruntimeConfig>());
    invokeOrderMgr->CreateGroupInstance("");
    ASSERT_EQ(invokeOrderMgr->instances.find("instanceId") == invokeOrderMgr->instances.end(), true);

    invokeOrderMgr->CreateGroupInstance("instanceId");
    ASSERT_EQ(invokeOrderMgr->instances.find("instanceId") == invokeOrderMgr->instances.end(), false);
    ASSERT_EQ(invokeOrderMgr->instances.find("instanceId")->second->unfinishedSeqNo == 0, true);

    invokeOrderMgr->CreateGroupInstance("instanceId");
    ASSERT_EQ(invokeOrderMgr->instances.find("instanceId")->second->unfinishedSeqNo == 0, true);

    invokeOrderMgr->NotifyGroupInstance("instanceId");
    ASSERT_EQ(invokeOrderMgr->instances.find("instanceId")->second->unfinishedSeqNo == 1, true);

    invokeOrderMgr->RemoveGroupInstance("instanceId");
    ASSERT_EQ(invokeOrderMgr->instances.find("instanceId") == invokeOrderMgr->instances.end(), true);
}

TEST_F(InvokeOrderManagerTest, ClearInsOrderMsgTest)
{
    auto invokeOrderMgr = std::make_shared<InvokeOrderManager>(std::make_shared<LibruntimeConfig>());
    invokeOrderMgr->instances["instanceId"] = invokeOrderMgr->ConstuctInstOrder();
    invokeOrderMgr->ClearInsOrderMsg("", libruntime::Signal::KillInstance);
    ASSERT_EQ(invokeOrderMgr->instances.find("instanceId") == invokeOrderMgr->instances.end(), false);

    invokeOrderMgr->ClearInsOrderMsg("instanceId", libruntime::Signal::KillInstance);
    ASSERT_EQ(invokeOrderMgr->instances.find("instanceId") == invokeOrderMgr->instances.end(), true);
}

TEST_F(InvokeOrderManagerTest, InvokeOrderInvokeTest)
{
    auto invokeOrderMgr = std::make_shared<InvokeOrderManager>(std::make_shared<LibruntimeConfig>());
    auto spec = std::make_shared<InvokeSpec>();
    invokeOrderMgr->Invoke(spec);
    ASSERT_EQ(invokeOrderMgr->instances.find("instanceId") == invokeOrderMgr->instances.end(), true);

    InvokeOptions opts;
    opts.isGetInstance = true;
    FunctionMeta meta;
    meta.name = "instanceId";
    spec->opts = opts;
    spec->functionMeta = meta;
    invokeOrderMgr->Invoke(spec);
    ASSERT_EQ(invokeOrderMgr->instances.find("yr_defalut_namespace-instanceId")->second->orderingCounter, 1);
}

TEST_F(InvokeOrderManagerTest, RemoveInstanceTest)
{
    auto invokeOrderMgr = std::make_shared<InvokeOrderManager>(std::make_shared<LibruntimeConfig>());
    auto spec = std::make_shared<InvokeSpec>();
    spec->opts.needOrder = true;
    ASSERT_NO_THROW(invokeOrderMgr->RemoveInstance(spec));

    spec->returnIds = {DataObject{"id"}};
    invokeOrderMgr->instances["id"] = std::make_shared<InstanceOrdering>();
    invokeOrderMgr->RemoveInstance(spec);
    ASSERT_EQ(invokeOrderMgr->instances.find("id") == invokeOrderMgr->instances.end(), true);
}

TEST_F(InvokeOrderManagerTest, RegisterInstanceAndUpdateOrderTest)
{
    auto invokeOrderMgr = std::make_shared<InvokeOrderManager>(std::make_shared<LibruntimeConfig>());
    auto insId = "instanceId";
    invokeOrderMgr->RegisterInstanceAndUpdateOrder(insId);
    ASSERT_EQ(invokeOrderMgr->instances.find("instanceId") != invokeOrderMgr->instances.end(), true);
    ASSERT_EQ(invokeOrderMgr->instances["instanceId"]->orderingCounter, 1);
}

TEST_F(InvokeOrderManagerTest, RegisterInstance)
{
    auto invokeOrderMgr = std::make_shared<InvokeOrderManager>(std::make_shared<LibruntimeConfig>());
    auto insId = "instanceId";
    invokeOrderMgr->RegisterInstance(insId);
    ASSERT_EQ(invokeOrderMgr->instances.find("instanceId") != invokeOrderMgr->instances.end(), true);
    ASSERT_EQ(invokeOrderMgr->instances["instanceId"]->orderingCounter, 0);
}
}  // namespace test
}  // namespace YR