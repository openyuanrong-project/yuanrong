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

#include "api/cpp/include/yr/api/err_type.h"
#include "api/cpp/include/yr/api/exception.h"
#include "api/cpp/include/yr/api/stream.h"
#include "api/cpp/src/stream_pubsub.h"

namespace YR {
namespace Libruntime {
class MockStreamProducer : public YR::Libruntime::StreamProducer {
public:
    MOCK_METHOD1(Send, YR::Libruntime::ErrorInfo(const YR::Libruntime::Element &element));
    MOCK_METHOD2(Send, YR::Libruntime::ErrorInfo(const YR::Libruntime::Element &element, int64_t timeoutMs));
    MOCK_METHOD0(Close, YR::Libruntime::ErrorInfo());
};

class MockStreamConsumer : public YR::Libruntime::StreamConsumer {
public:
    MOCK_METHOD3(Receive, YR::Libruntime::ErrorInfo(uint32_t expectNum, uint32_t timeoutMs,
                                                    std::vector<YR::Libruntime::Element> &outElements));
    MOCK_METHOD2(Receive,
                 YR::Libruntime::ErrorInfo(uint32_t timeoutMs, std::vector<YR::Libruntime::Element> &outElements));
    MOCK_METHOD1(Ack, YR::Libruntime::ErrorInfo(uint64_t elementId));
    MOCK_METHOD0(Close, YR::Libruntime::ErrorInfo());
};
}  // namespace Libruntime

namespace test {
using namespace YR::utility;
using namespace testing;

class StreamPubSubTest : public testing::Test {
public:
    StreamPubSubTest() {};
    ~StreamPubSubTest() {};
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
        this->producer = std::make_shared<YR::Libruntime::MockStreamProducer>();
        this->consumer = std::make_shared<YR::Libruntime::MockStreamConsumer>();
        this->streamProducer = std::make_shared<YR::StreamProducer>(this->producer);
        this->streamConsumer = std::make_shared<YR::StreamConsumer>(this->consumer);
    }
    std::shared_ptr<YR::Libruntime::MockStreamProducer> producer;
    std::shared_ptr<YR::Libruntime::MockStreamConsumer> consumer;
    std::shared_ptr<YR::StreamProducer> streamProducer;
    std::shared_ptr<YR::StreamConsumer> streamConsumer;
};

TEST_F(StreamPubSubTest, SendFailedTest)
{
    Element ele(nullptr, 10, 10);
    YR::Libruntime::ErrorInfo err(YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED,
                                  YR::Libruntime::ModuleCode::DATASYSTEM, "Send failed.");
    EXPECT_CALL(*this->producer, Send(_)).WillOnce(testing::Return(err));
    bool isThrow = false;
    try {
        streamProducer->Send(ele);
    } catch (YR::Exception &err) {
        ASSERT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED);
        EXPECT_THAT(err.Msg(), testing::HasSubstr("Send failed."));
        isThrow = true;
    }
    EXPECT_TRUE(isThrow);

    EXPECT_CALL(*this->producer, Send(_, _)).WillOnce(testing::Return(err));
    isThrow = false;
    try {
        streamProducer->Send(ele, 1000);
    } catch (YR::Exception &err) {
        ASSERT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED);
        EXPECT_THAT(err.Msg(), testing::HasSubstr("Send failed."));
        isThrow = true;
    }
    EXPECT_TRUE(isThrow);
}

TEST_F(StreamPubSubTest, SendSuccessfullyTest)
{
    Element ele(nullptr, 10, 10);
    EXPECT_CALL(*this->producer, Send(_)).WillOnce(testing::Return(YR::Libruntime::ErrorInfo()));
    EXPECT_NO_THROW(streamProducer->Send(ele));

    EXPECT_CALL(*this->producer, Send(_, _)).WillOnce(testing::Return(YR::Libruntime::ErrorInfo()));
    EXPECT_NO_THROW(streamProducer->Send(ele, 1000));
}

TEST_F(StreamPubSubTest, ProducerCloseFailedTest)
{
    YR::Libruntime::ErrorInfo err(YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED,
                                  YR::Libruntime::ModuleCode::DATASYSTEM, "Close failed.");
    EXPECT_CALL(*(this->producer), Close()).WillOnce(testing::Return(err));
    bool isThrow = false;
    try {
        this->streamProducer->Close();
    } catch (YR::Exception &err) {
        ASSERT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED);
        EXPECT_THAT(err.Msg(), testing::HasSubstr("Close failed."));
        isThrow = true;
    }
    EXPECT_TRUE(isThrow);
}

TEST_F(StreamPubSubTest, ProducerCloseSuccessfullyTest)
{
    EXPECT_CALL(*(this->producer), Close()).WillOnce(testing::Return(YR::Libruntime::ErrorInfo()));
    EXPECT_NO_THROW(this->streamProducer->Close());
}

TEST_F(StreamPubSubTest, ReceiveTest)
{
    YR::Libruntime::ErrorInfo err(YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED,
                                  YR::Libruntime::ModuleCode::DATASYSTEM, "Receive failed.");
    EXPECT_CALL(*(this->consumer), Receive(_, _))
        .WillOnce(testing::Return(err))
        .WillOnce(testing::Return(YR::Libruntime::ErrorInfo()));
    std::vector<Element> outElements;
    bool isThrow = false;
    try {
        this->streamConsumer->Receive(1000, outElements);
    } catch (YR::Exception &err) {
        ASSERT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED);
        EXPECT_THAT(err.Msg(), testing::HasSubstr("Receive failed."));
        isThrow = true;
    }
    EXPECT_TRUE(isThrow);
    EXPECT_NO_THROW(this->streamConsumer->Receive(1000, outElements));

    EXPECT_CALL(*(this->consumer), Receive(_, _, _))
        .WillOnce(testing::Return(err))
        .WillOnce(testing::Return(YR::Libruntime::ErrorInfo()));
    isThrow = false;
    try {
        this->streamConsumer->Receive(1, 1000, outElements);
    } catch (YR::Exception &err) {
        ASSERT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED);
        EXPECT_THAT(err.Msg(), testing::HasSubstr("Receive failed."));
        isThrow = true;
    }
    EXPECT_TRUE(isThrow);
    EXPECT_NO_THROW(this->streamConsumer->Receive(1, 1000, outElements));
}

TEST_F(StreamPubSubTest, AckFailedTest)
{
    YR::Libruntime::ErrorInfo err(YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED,
                                  YR::Libruntime::ModuleCode::DATASYSTEM, "Ack failed.");
    EXPECT_CALL(*(this->consumer), Ack(_)).WillOnce(testing::Return(err));
    bool isThrow = false;
    try {
        this->streamConsumer->Ack(111);
    } catch (YR::Exception &err) {
        ASSERT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED);
        EXPECT_THAT(err.Msg(), testing::HasSubstr("Ack failed."));
        isThrow = true;
    }
    EXPECT_TRUE(isThrow);
}

TEST_F(StreamPubSubTest, AckSuccessfullyTest)
{
    EXPECT_CALL(*(this->consumer), Ack(_)).WillOnce(testing::Return(YR::Libruntime::ErrorInfo()));
    EXPECT_NO_THROW(this->streamConsumer->Ack(111));
}

TEST_F(StreamPubSubTest, ConsumerCloseFailedTest)
{
    YR::Libruntime::ErrorInfo err(YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED,
                                  YR::Libruntime::ModuleCode::DATASYSTEM, "Close failed.");
    EXPECT_CALL(*(this->consumer), Close()).WillOnce(testing::Return(err));
    bool isThrow = false;
    try {
        this->streamConsumer->Close();
    } catch (YR::Exception &err) {
        ASSERT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED);
        EXPECT_THAT(err.Msg(), testing::HasSubstr("Close failed."));
        isThrow = true;
    }
    EXPECT_TRUE(isThrow);
}

TEST_F(StreamPubSubTest, ConsumerCloseSuccessfullyTest)
{
    EXPECT_CALL(*(this->consumer), Close()).WillOnce(testing::Return(YR::Libruntime::ErrorInfo()));
    EXPECT_NO_THROW(this->streamConsumer->Close());
}
}  // namespace test
}  // namespace YR