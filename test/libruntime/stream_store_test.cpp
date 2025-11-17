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
#include "src/dto/stream_conf.h"
#include "src/libruntime/libruntime.h"
#include "src/libruntime/streamstore/datasystem_stream_store.h"
#include "src/libruntime/streamstore/stream_producer_consumer.h"
#include "src/proto/libruntime.pb.h"
#include "src/utility/logger/logger.h"

using namespace testing;
using namespace YR::Libruntime;
using namespace YR::utility;

namespace YR {
namespace test {
class StreamStoreTest : public testing::Test {
public:
    StreamStoreTest() {};
    ~StreamStoreTest() {};
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
        streamStore_ = std::make_shared<DatasystemStreamStore>();
        streamStore_->Init("127,0,0,1", 11111);
        InitGlobalTimer();
    }
    void TearDown() override
    {
        CloseGlobalTimer();
        streamStore_->Shutdown();
    }

    std::shared_ptr<DatasystemStreamStore> streamStore_;
};

TEST_F(StreamStoreTest, CreateStreamProducer)
{
    std::string streamName = "streamName";
    Libruntime::ProducerConf libProducerConf{};
    libProducerConf.extendConfig.insert({"STREAM_MODE", "MPMC"});
    std::shared_ptr<Libruntime::StreamProducer> emptyProducer = nullptr;
    ErrorInfo err = streamStore_->CreateStreamProducer(streamName, emptyProducer, libProducerConf);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_PARAM_INVALID);
    auto producer = std::make_shared<Libruntime::StreamProducer>();
    err = streamStore_->CreateStreamProducer(streamName, producer, libProducerConf);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    libProducerConf.extendConfig.insert({"XXXXX", "XXXXX"});
    err = streamStore_->CreateStreamProducer(streamName, producer, libProducerConf);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    libProducerConf.extendConfig["STREAM_MODE"] = "XXXXX";
    err = streamStore_->CreateStreamProducer(streamName, producer, libProducerConf);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_PARAM_INVALID);
}

TEST_F(StreamStoreTest, CreateStreamConsumer)
{
    std::string streamName = "streamName";
    Libruntime::SubscriptionConfig subscriptionConfig("subName", libruntime::SubscriptionType::STREAM);
    std::shared_ptr<Libruntime::StreamConsumer> emptyConsumer = nullptr;
    ErrorInfo err = streamStore_->CreateStreamConsumer(streamName, subscriptionConfig, emptyConsumer);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_PARAM_INVALID);
    auto consumer = std::make_shared<Libruntime::StreamConsumer>();
    err = streamStore_->CreateStreamConsumer(streamName, subscriptionConfig, consumer);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
}

TEST_F(StreamStoreTest, DeleteStream)
{
    std::string streamName = "streamName";
    ErrorInfo err = streamStore_->DeleteStream(streamName);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
}

TEST_F(StreamStoreTest, QueryGlobalProducersNum)
{
    std::string streamName = "streamName";
    uint64_t gProducerNum = 0;
    ErrorInfo err = streamStore_->QueryGlobalProducersNum(streamName, gProducerNum);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
}

TEST_F(StreamStoreTest, QueryGlobalConsumersNum)
{
    std::string streamName = "streamName";
    uint64_t gConsumerNum = 0;
    ErrorInfo err = streamStore_->QueryGlobalConsumersNum(streamName, gConsumerNum);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
}

TEST_F(StreamStoreTest, TestProducer)
{
    auto streamProducer = std::make_shared<Libruntime::StreamProducer>();
    std::string str = "hello";
    Libruntime::Element element((uint8_t *)(str.c_str()), str.size());
    ErrorInfo err = streamProducer->Send(element);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    err = streamProducer->Send(element, 1000);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    err = streamProducer->Close();
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
}

TEST_F(StreamStoreTest, TestConsumer)
{
    auto streamConsumer = std::make_shared<Libruntime::StreamConsumer>();
    std::vector<Libruntime::Element> elements;
    ErrorInfo err = streamConsumer->Receive(1, 1000, elements);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    err = streamConsumer->Receive(1000, elements);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    err = streamConsumer->Ack(11);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    err = streamConsumer->Close();
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    setenv("STREAM_RECEIVE_LIMIT", "1000", 1);
    Libruntime::Config::Instance() = Libruntime::Config();
    err = streamConsumer->Receive(999, 1, elements);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_INNER_SYSTEM_ERROR);
    unsetenv("STREAM_RECEIVE_LIMIT");
}

TEST_F(StreamStoreTest, TestUpdateTokenAndAksk)
{
    ErrorInfo err = streamStore_->UpdateAkSk("ak", datasystem::SensitiveValue("sk"));
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    err = streamStore_->UpdateToken(datasystem::SensitiveValue("token"));
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
}

}  // namespace test
}  // namespace YR
