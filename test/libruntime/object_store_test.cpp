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

#include "mock/mock_datasystem_client.h"
#include "src/libruntime/libruntime.h"
#include "src/libruntime/objectstore/datasystem_object_client_wrapper.h"
#include "src/libruntime/objectstore/datasystem_object_store.h"
#include "src/libruntime/objectstore/object_store.h"

using namespace testing;
using namespace YR::Libruntime;
using namespace YR::utility;

namespace YR {
namespace test {
class ObjectStoreTest : public testing::Test {
public:
    ObjectStoreTest(){};
    ~ObjectStoreTest(){};
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
        objectStore_ = std::make_shared<DSCacheObjectStore>();
        objectStore_->Init("127,0,0,1", 11111);
        InitGlobalTimer();
    }
    void TearDown() override
    {
        CloseGlobalTimer();
        objectStore_->Shutdown();
    }

    std::shared_ptr<DSCacheObjectStore> objectStore_;
};

TEST_F(ObjectStoreTest, CreateBufferTest)
{
    std::shared_ptr<Buffer> dataBuf;
    CreateParam createParam;
    auto err = objectStore_->CreateBuffer("objID", 1000, dataBuf, createParam);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    err = objectStore_->CreateBuffer("repeatedObjId", 1000, dataBuf, createParam);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    err = objectStore_->CreateBuffer("errObjId", 1000, dataBuf, createParam);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_INNER_COMMUNICATION);
}

TEST_F(ObjectStoreTest, GetBuffersWithoutRetryTest)
{
    std::vector<std::string> ids = {"objID"};
    auto [info, buffer] = objectStore_->GetBuffersWithoutRetry(ids, 1000);
    ASSERT_EQ(info.errorInfo.Code(), ErrorCode::ERR_OK);
}

TEST_F(ObjectStoreTest, PutTest)
{
    std::shared_ptr<Buffer> data = std::make_shared<NativeBuffer>(100);
    std::unordered_set<std::string> nestedId;
    CreateParam createParam;
    auto err = objectStore_->Put(data, "objID", nestedId, createParam);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    err = objectStore_->Put(data, "repeatedObjId", nestedId, createParam);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    err = objectStore_->Put(data, "errObjId", nestedId, createParam);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_INNER_COMMUNICATION);
}

TEST_F(ObjectStoreTest, GetTest)
{
    auto res = objectStore_->Get("objID", 1000);
    ASSERT_EQ(res.first.Code(), ErrorCode::ERR_OK);
}

TEST_F(ObjectStoreTest, UpdateTokenAndAKSKTest)
{
    datasystem::SensitiveValue token("token");
    ErrorInfo err = objectStore_->UpdateToken(token);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    datasystem::SensitiveValue sk("sk");
    err = objectStore_->UpdateAkSk("ak", sk);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    std::shared_ptr<Buffer> dataBuf = std::make_shared<NativeBuffer>(1000);
    CreateParam createParam;
    err = objectStore_->CreateBuffer("objID", 1000, dataBuf, createParam);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    err = objectStore_->UpdateToken(token);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    err = objectStore_->UpdateAkSk("ak", sk);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
}

TEST_F(ObjectStoreTest, IncreGlobalReferenceTest)
{
    std::vector<std::string> objectIds = {"objID"};
    auto err = objectStore_->IncreGlobalReference(objectIds);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    std::vector<std::string> errObjectIds = {"objID1", "objID2"};
    err = objectStore_->IncreGlobalReference(errObjectIds);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_INNER_COMMUNICATION);
    auto res = objectStore_->IncreGlobalReference(objectIds, "remoteID");
    ASSERT_EQ(res.first.Code(), ErrorCode::ERR_OK);
}

TEST_F(ObjectStoreTest, DecreGlobalReferenceTest)
{
    std::vector<std::string> objectIds = {"objID"};
    auto err = objectStore_->DecreGlobalReference(objectIds);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    auto res = objectStore_->DecreGlobalReference(objectIds, "remoteID");
    ASSERT_EQ(res.first.Code(), ErrorCode::ERR_OK);
}

TEST_F(ObjectStoreTest, QueryGlobalReferenceTest)
{
    std::vector<std::string> objectIds = {"objID"};
    auto res = objectStore_->QueryGlobalReference(objectIds);
    ASSERT_EQ(res.at(0), 1);
}

TEST_F(ObjectStoreTest, ReleaseGRefsTest)
{
    auto err = objectStore_->ReleaseGRefs("remoteID");
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
}

TEST_F(ObjectStoreTest, GenerateKeyTest)
{
    std::string key;
    auto err = objectStore_->GenerateKey(key, "prefix", true);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
}

TEST_F(ObjectStoreTest, ClearTest)
{
    EXPECT_NO_THROW(objectStore_->SetTenantId("tenantId"));
    EXPECT_NO_THROW(objectStore_->Clear());
}

TEST_F(ObjectStoreTest, DataSystemBufferTest)
{
    std::shared_ptr<Buffer> dataBuf;
    CreateParam createParam;
    auto err = objectStore_->CreateBuffer("objID", 1000, dataBuf, createParam);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    void* ptr = reinterpret_cast<void*>(11111);
    std::unordered_set<std::string> nestedIds = {"nestedId"};
    err = dataBuf->MemoryCopy(ptr, 7);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    err = dataBuf->Seal(nestedIds);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    err = dataBuf->WriterLatch();
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    err = dataBuf->WriterUnlatch();
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    err = dataBuf->ReaderLatch();
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    err = dataBuf->ReaderUnlatch();
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    err = dataBuf->Publish();
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
}

TEST_F(ObjectStoreTest, DatasystemObjectClientWrapperTest)
{
    std::vector<std::string> objectIds = {"objID"};
    std::vector<std::string> failedObjectIds;
    auto err = objectStore_->DecreGlobalReference(objectIds);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    ds::ConnectOptions connectOpts;
    auto dsClient = std::make_shared<ds::ObjectClient>(connectOpts);
    DatasystemObjectClientWrapper wrapper(dsClient);
    auto status = wrapper.GDecreaseRef(objectIds, failedObjectIds);
    ASSERT_EQ(failedObjectIds.size(), 0);
    wrapper.SetTenantId("tenantId");
}

}  // namespace test
}  // namespace YR