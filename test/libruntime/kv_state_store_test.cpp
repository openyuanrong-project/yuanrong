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

#include "mock/mock_datasystem_client.h"
#include "src/libruntime/libruntime.h"
#include "src/libruntime/statestore/datasystem_state_store.h"
#include "src/utility/logger/logger.h"

using namespace testing;
using namespace YR::Libruntime;
using namespace YR::utility;

namespace YR {
namespace test {
class KVStateStoreTest : public testing::Test {
public:
    KVStateStoreTest(){};
    ~KVStateStoreTest(){};
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
        stateStore_ = std::make_shared<DSCacheStateStore>();
        stateStore_->Init("127.0.0.1", 11111);
        InitGlobalTimer();
    }
    void TearDown() override
    {
        CloseGlobalTimer();
        stateStore_.reset();
    }

    std::shared_ptr<DSCacheStateStore> stateStore_;
};

TEST_F(KVStateStoreTest, KVWriteReadDelExist)
{
    std::string key = "123";
    std::string key2 = "456";
    std::string value = "this is mock value.";

    // Write
    std::shared_ptr<Buffer> sbuf = std::make_shared<NativeBuffer>(value.size());
    sbuf->MemoryCopy(value.data(), value.size());
    YR::Libruntime::ExistenceOpt existence = ExistenceOpt::NONE;
    YR::Libruntime::WriteMode writeMode = WriteMode::NONE_L2_CACHE;
    YR::Libruntime::SetParam setParam = {
        .writeMode = writeMode,
        .ttlSecond = 10,
        .existence = existence,
    };
    ErrorInfo err = stateStore_->Write(key, sbuf, setParam);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);

    // Read
    SingleReadResult readResult = stateStore_->Read(key, -1);
    ASSERT_EQ(readResult.second.Code(), ErrorCode::ERR_OK);
    MultipleReadResult multiReadResult = stateStore_->Read({key, key}, 100, false);
    ASSERT_EQ(multiReadResult.second.Code(), ErrorCode::ERR_GET_OPERATION_FAILED);

    // Exist
    MultipleExistResult emptyExistResult = stateStore_->Exist({});
    ASSERT_EQ(emptyExistResult.second.Code(), ErrorCode::ERR_PARAM_INVALID);
    MultipleExistResult existResult = stateStore_->Exist({key, key2});
    ASSERT_EQ(existResult.second.Code(), ErrorCode::ERR_OK);

    // Del
    err = stateStore_->Del(key);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    MultipleDelResult mdResult = stateStore_->Del({key, key2});
    ASSERT_EQ(mdResult.second.Code(), ErrorCode::ERR_OK);
}

TEST_F(KVStateStoreTest, MultipleWriteReadTest)
{
    std::vector<std::string> keys = {"key1", "key2", "key3"};
    std::vector<std::string> values = {"value1", "value2", "value3"};
    
    for (size_t i = 0; i < keys.size(); i++) {
        std::shared_ptr<Buffer> sbuf = std::make_shared<NativeBuffer>(values[i].size());
        sbuf->MemoryCopy(values[i].data(), values[i].size());
        YR::Libruntime::SetParam setParam = {
            .writeMode = WriteMode::NONE_L2_CACHE,
            .ttlSecond = 10,
        };
        ErrorInfo err = stateStore_->Write(keys[i], sbuf, setParam);
        ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    }
    
    MultipleReadResult multiReadResult = stateStore_->Read(keys, 100, false);
    ASSERT_EQ(multiReadResult.second.Code(), ErrorCode::ERR_OK);
    ASSERT_EQ(multiReadResult.first.size(), 3);
}

TEST_F(KVStateStoreTest, TtlExpirationTest)
{
    std::string key = "ttl_key";
    std::string value = "ttl_value";
    
    std::shared_ptr<Buffer> sbuf = std::make_shared<NativeBuffer>(value.size());
    sbuf->MemoryCopy(value.data(), value.size());
    YR::Libruntime::SetParam setParam = {
        .writeMode = WriteMode::NONE_L2_CACHE,
        .ttlSecond = 1,
    };
    ErrorInfo err = stateStore_->Write(key, sbuf, setParam);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    
    SingleReadResult readResult = stateStore_->Read(key, -1);
    ASSERT_EQ(readResult.second.Code(), ErrorCode::ERR_OK);
}

TEST_F(KVStateStoreTest, ErrorHandlingTest)
{
    std::string emptyKey = "";
    std::string value = "test";
    
    std::shared_ptr<Buffer> sbuf = std::make_shared<NativeBuffer>(value.size());
    sbuf->MemoryCopy(value.data(), value.size());
    YR::Libruntime::SetParam setParam = {};
    
    ErrorInfo err = stateStore_->Write(emptyKey, sbuf, setParam);
}

TEST_F(KVStateStoreTest, ExistTest)
{
    std::string key = "exist_key";
    std::string value = "exist_value";
    
    std::shared_ptr<Buffer> sbuf = std::make_shared<NativeBuffer>(value.size());
    sbuf->MemoryCopy(value.data(), value.size());
    YR::Libruntime::SetParam setParam = {};
    
    ErrorInfo err = stateStore_->Write(key, sbuf, setParam);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    
    MultipleExistResult existResult = stateStore_->Exist({key});
    ASSERT_EQ(existResult.second.Code(), ErrorCode::ERR_OK);
}

TEST_F(KVStateStoreTest, QuerySizeTest)
{
    std::string key = "size_key";
    std::string value = "size_value";
    
    std::shared_ptr<Buffer> sbuf = std::make_shared<NativeBuffer>(value.size());
    sbuf->MemoryCopy(value.data(), value.size());
    YR::Libruntime::SetParam setParam = {};
    
    ErrorInfo err = stateStore_->Write(key, sbuf, setParam);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    
    std::vector<uint64_t> sizes;
    err = stateStore_->QuerySize({key}, sizes);
}
}  // namespace test
}  // namespace YR