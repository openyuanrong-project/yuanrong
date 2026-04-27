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

#include <gtest/gtest.h>
#include <unordered_map>

#include "src/libruntime/fsclient/fs_intf_manager.h"

namespace YR {
namespace Libruntime {

class FSIntfManagerTest : public ::testing::Test {
protected:
    std::shared_ptr<Security> security;
    std::shared_ptr<ClientsManager> clientsMgr;
    std::shared_ptr<FSIntfManager> fsIntfManager;

    void SetUp() override
    {
        security = std::make_shared<Security>();
        clientsMgr = std::make_shared<ClientsManager>();
        fsIntfManager = std::make_shared<FSIntfManager>(clientsMgr);
    }
};

TEST_F(FSIntfManagerTest, NewFsIntfClient_ShouldReturnValidInterface_WhenOptionIsValid)
{
    ReaderWriterClientOption option;
    ProtocolType type = ProtocolType::GRPC;

    std::shared_ptr<FSIntfReaderWriter> intf =
        fsIntfManager->NewFsIntfClient("srcInstance", "dstInstance", "runtimeID", option, type);

    ASSERT_NE(intf, nullptr);
}

TEST_F(FSIntfManagerTest, TryGet_ShouldReturnSystemInterface_WhenDirectRuntimeRwNotExist)
{
    std::string instanceID = "instanceID";
    ReaderWriterClientOption option;
    ProtocolType type = ProtocolType::GRPC;
    std::shared_ptr<FSIntfReaderWriter> system =
        fsIntfManager->NewFsIntfClient("srcInstance", "function-proxy", "runtimeID", option, type);
    fsIntfManager->UpdateSystemIntf(system);
    std::shared_ptr<FSIntfReaderWriter> intf = fsIntfManager->Get(instanceID);

    ASSERT_EQ(intf, fsIntfManager->GetSystemIntf());
    ASSERT_EQ(system, fsIntfManager->GetSystemIntf());
}

TEST_F(FSIntfManagerTest, Get_ShouldReturnValidInterface_WhenInstanceIDExist)
{
    ReaderWriterClientOption option;
    ProtocolType type = ProtocolType::GRPC;
    std::shared_ptr<FSIntfReaderWriter> intf =
        fsIntfManager->NewFsIntfClient("srcInstance", "dstInstance", "runtimeID", option, type);

    ASSERT_TRUE(fsIntfManager->Emplace("dstInstance", intf));

    std::shared_ptr<FSIntfReaderWriter> retrievedIntf = fsIntfManager->TryGet("dstInstance");

    ASSERT_EQ(intf, retrievedIntf);
}

TEST_F(FSIntfManagerTest, Remove_ShouldRemoveInterface_WhenInstanceIDExist)
{
    ReaderWriterClientOption option;
    ProtocolType type = ProtocolType::GRPC;
    std::shared_ptr<FSIntfReaderWriter> intf =
        fsIntfManager->NewFsIntfClient("srcInstance", "dstInstance", "runtimeID", option, type);

    ASSERT_TRUE(fsIntfManager->Emplace("dstInstance", intf));

    fsIntfManager->Remove("dstInstance");

    std::shared_ptr<FSIntfReaderWriter> retrievedIntf = fsIntfManager->Get("dstInstance");

    ASSERT_EQ(retrievedIntf, nullptr);
}

TEST_F(FSIntfManagerTest, Clear_ShouldRemoveAllInterfaces)
{
    std::string instanceID1 = "instanceID1";
    std::string instanceID2 = "instanceID2";
    ReaderWriterClientOption option;
    ProtocolType type = ProtocolType::GRPC;
    std::shared_ptr<FSIntfReaderWriter> intf1 =
        fsIntfManager->NewFsIntfClient("srcInstance", instanceID1, "runtimeID", option, type);
    std::shared_ptr<FSIntfReaderWriter> intf2 =
        fsIntfManager->NewFsIntfClient("srcInstance", instanceID2, "runtimeID", option, type);

    ASSERT_TRUE(fsIntfManager->Emplace(instanceID1, intf1));
    ASSERT_TRUE(fsIntfManager->Emplace(instanceID2, intf2));

    fsIntfManager->Clear();

    ASSERT_EQ(fsIntfManager->Get(instanceID1), nullptr);
    ASSERT_EQ(fsIntfManager->Get(instanceID2), nullptr);
}

TEST_F(FSIntfManagerTest, NewEventIntfClient_EmplaceEventIntfs_TryGetEventIntfs)
{
    std::shared_ptr<FSIntfReaderWriter> intf = fsIntfManager->NewEventIntfClient(
        "srcInstance", "eventInstance", "runtimeID", ReaderWriterClientOption(), ProtocolType::GRPC);
    ASSERT_TRUE(fsIntfManager->EmplaceEventIntfs("eventInstance", intf));

    std::shared_ptr<FSIntfReaderWriter> retrievedIntf1 = fsIntfManager->TryGetEventIntfs("eventInstance");
    std::shared_ptr<FSIntfReaderWriter> retrievedIntf2 = fsIntfManager->GetEventIntfs("eventInstance");
    ASSERT_NE(intf, nullptr);
    ASSERT_NE(retrievedIntf1, nullptr);
    ASSERT_EQ(retrievedIntf2, nullptr);

    fsIntfManager->Remove("eventInstance");
    retrievedIntf1 = fsIntfManager->TryGetEventIntfs("eventInstance");
    ASSERT_EQ(retrievedIntf1, nullptr);
}

TEST_F(FSIntfManagerTest, NewEventIntfClient_ShouldReturnExistingEventIntf_WhenAvailable)
{
    std::string dstInstance = "eventInstance2";
    std::shared_ptr<FSIntfReaderWriter> intf = fsIntfManager->NewEventIntfClient(
        "srcInstance", dstInstance, "runtimeID", ReaderWriterClientOption(), ProtocolType::GRPC);
    std::shared_ptr<FSIntfReaderWriter> intf2 = fsIntfManager->NewEventIntfClient(
        "srcInstance", dstInstance, "runtimeID", ReaderWriterClientOption(), ProtocolType::GRPC);
    ASSERT_NE(intf, intf2);

    std::shared_ptr<FSIntfReaderWriter> intf3 = fsIntfManager->NewEventIntfClient(
    "srcInstance", "eventInstance", "runtimeID", ReaderWriterClientOption(), static_cast<ProtocolType>(999));
    ASSERT_EQ(intf3, nullptr);
}

}  // namespace Libruntime
}  // namespace YR