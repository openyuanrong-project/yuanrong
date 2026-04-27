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

#define private public
#include "mock/mock_datasystem_client.h"
#include "src/libruntime/heterostore/datasystem_hetero_store.h"
#include "src/utility/logger/logger.h"
#undef private

using namespace testing;
using namespace YR::Libruntime;
using namespace YR::utility;

namespace YR {
namespace test {
class HeteroStoreTest : public testing::Test {
public:
    HeteroStoreTest(){};
    ~HeteroStoreTest(){};
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
        datasystem::ConnectOptions connectOptions;
        heteroStore_ = std::make_shared<DatasystemHeteroStore>();
        heteroStore_->Init(connectOptions);
        heteroStore_->InitOnce();
    }
    void TearDown() override
    {
        heteroStore_.reset();
    }

    std::shared_ptr<DatasystemHeteroStore> heteroStore_;
};

std::vector<DeviceBlobList> BuildDeviceBlobList()
{
    Blob blob;
    blob.pointer = reinterpret_cast<void*>(11111);
    blob.size = 100;
    DeviceBlobList blobList;
    blobList.blobs = {blob, blob};
    blobList.deviceIdx = 0;
    std::vector<DeviceBlobList> devBlobList = {blobList};
    return devBlobList;
}

TEST_F(HeteroStoreTest, ShutdownTest)
{
    EXPECT_NO_THROW(heteroStore_->Shutdown());
}

TEST_F(HeteroStoreTest, DevDeleteTest)
{
    std::vector<std::string> objIds = {"obj1", "obj2"};
    std::vector<std::string> failedObjectIds;
    auto err = heteroStore_->DevDelete(objIds, failedObjectIds);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
}

TEST_F(HeteroStoreTest, DevLocalDeleteTest)
{
    std::vector<std::string> objIds = {"obj1", "obj2"};
    std::vector<std::string> failedObjectIds;
    auto err = heteroStore_->DevLocalDelete(objIds, failedObjectIds);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
}

TEST_F(HeteroStoreTest, DevSubscribeTest)
{
    auto devBlobList = BuildDeviceBlobList();
    std::vector<std::shared_ptr<YR::Libruntime::HeteroFuture>> futureVec;
    auto err = heteroStore_->DevSubscribe({"key1"}, devBlobList, futureVec);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    ASSERT_EQ(futureVec.size(), 1);
}

TEST_F(HeteroStoreTest, DevPublishTest)
{
    auto devBlobList = BuildDeviceBlobList();
    std::vector<std::shared_ptr<YR::Libruntime::HeteroFuture>> futureVec;
    auto err = heteroStore_->DevPublish({"key1"}, devBlobList, futureVec);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    ASSERT_EQ(futureVec.size(), 1);
}

TEST_F(HeteroStoreTest, DevMSetTest)
{
    auto devBlobList = BuildDeviceBlobList();
    std::vector<std::string> failedKeys;
    auto err = heteroStore_->DevMSet({"key1", "key2"}, devBlobList, failedKeys);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    ASSERT_EQ(failedKeys.size(), 0);
}

TEST_F(HeteroStoreTest, DevMGetTest)
{
    auto devBlobList = BuildDeviceBlobList();
    std::vector<std::string> failedKeys;
    auto err = heteroStore_->DevMGet({"key1", "key2"}, devBlobList, failedKeys, 1000);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    ASSERT_EQ(failedKeys.size(), 0);
}
}  // namespace test
}  // namespace YR
