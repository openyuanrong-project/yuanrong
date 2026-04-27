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

#include <chrono>
#include <thread>

#include "mock/mock_datasystem.h"

#define private public
#include "src/libruntime/generator/generator_id_map.h"
#include "src/libruntime/generator/stream_generator_notifier.h"
#include "src/libruntime/generator/stream_generator_receiver.h"
#include "src/utility/logger/logger.h"

namespace YR {
namespace test {
using namespace testing;
using namespace YR::Libruntime;
using namespace YR::utility;
using namespace std::chrono_literals;
class GeneratorTest : public testing::Test {
public:
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
        consumer = std::make_shared<MockStreamConsumer>();
        streamStore = std::make_shared<MockStreamStore>();
        auto dsStore = std::make_shared<MockObjectStore>();
        auto waitManager = std::make_shared<WaitingObjectManager>();
        memStore = std::make_shared<MemoryStore>();
        memStore->Init(dsStore, waitManager);
    }

    void TearDown() override
    {
        map.reset();
        consumer.reset();
        streamStore.reset();
        memStore.reset();
        if (notifier) {
            notifier.reset();
        }
        if (receiver) {
            receiver.reset();
        }
    }

    std::shared_ptr<StreamGeneratorReceiver> receiver = nullptr;
    std::shared_ptr<StreamGeneratorNotifier> notifier = nullptr;
    std::shared_ptr<MockStreamStore> streamStore = nullptr;
    std::shared_ptr<MemoryStore> memStore = nullptr;
    std::shared_ptr<StreamConsumer> consumer = nullptr;
    std::shared_ptr<LibruntimeConfig> config = std::make_shared<LibruntimeConfig>();
    std::shared_ptr<GeneratorIdMap> map = std::make_shared<GeneratorIdMap>();
};

TEST_F(GeneratorTest, DetectHeartbeatTest)
{
    receiver = std::make_shared<StreamGeneratorReceiver>(config, streamStore, memStore);
    auto fakeTimeStamp = GetCurrentTimestampMs() - 40000;
    receiver->map_->AddRecord("objId", std::to_string(fakeTimeStamp));
    ASSERT_NO_THROW(receiver->DetectHeartbeat());
}

TEST_F(GeneratorTest, HandleGeneratorResultTest)
{
    receiver = std::make_shared<StreamGeneratorReceiver>(config, streamStore, memStore);
    libruntime::NotifyGeneratorResult result;
    result.set_genid("genid");
    result.set_errorcode(0);
    result.set_data("result");
    receiver->numGeneratorResults_["genid"] = 1;
    receiver->generatorResultsCounter_["genid"] = 1;
    ASSERT_NO_THROW(receiver->HandleGeneratorResult(result));
}

TEST_F(GeneratorTest, AddRecordTest)
{
    receiver = std::make_shared<StreamGeneratorReceiver>(config, streamStore, memStore);
    receiver->AddRecord("genId");
    std::vector<std::string> keys;
    receiver->map_->GetRecordKeys(keys);
    ASSERT_EQ(keys.size() != 0, true);
}

TEST_F(GeneratorTest, HandleGeneratorHeartbeatTest)
{
    receiver = std::make_shared<StreamGeneratorReceiver>(config, streamStore, memStore);
    receiver->HandleGeneratorHeartbeat("genId");
    std::vector<std::string> keys;
    receiver->map_->GetRecordKeys(keys);
    ASSERT_EQ(keys.size() != 0, true);
}

TEST_F(GeneratorTest, NotifyHeartbeatTest)
{
    notifier = std::make_shared<StreamGeneratorNotifier>(streamStore, map);
    notifier->map_ = nullptr;
    ASSERT_EQ(notifier->NotifyHeartbeat().Code(), ErrorCode::ERR_INNER_SYSTEM_ERROR);
    notifier->map_ = map;
    notifier->map_->AddRecord("genId", std::to_string(GetCurrentTimestampMs()));
    ASSERT_EQ(notifier->NotifyHeartbeat().OK(), true);
}

TEST_F(GeneratorTest, NotifyHeartbeatByStreamTest)
{
    notifier = std::make_shared<StreamGeneratorNotifier>(streamStore, map);
    notifier->map_ = nullptr;
    ASSERT_EQ(notifier->NotifyHeartbeatByStream("genId").Code(), ErrorCode::ERR_INNER_SYSTEM_ERROR);

    notifier->map_ = map;
    ASSERT_EQ(notifier->NotifyHeartbeatByStream("genId").Code(), ErrorCode::ERR_INNER_SYSTEM_ERROR);

    notifier->map_->AddRecord("genId", std::to_string(GetCurrentTimestampMs()));
    ASSERT_EQ(notifier->NotifyHeartbeatByStream("genId").OK(), true);
}

TEST_F(GeneratorTest, NotifyFinishedByStreamTest)
{
    notifier = std::make_shared<StreamGeneratorNotifier>(streamStore, map);
    notifier->map_->AddRecord("genId", std::to_string(GetCurrentTimestampMs()));
    notifier->dsStreamStore_ = nullptr;
    ASSERT_EQ(notifier->NotifyFinishedByStream("genId", 1).Code(), ErrorCode::ERR_INNER_SYSTEM_ERROR);
    std::shared_ptr<DataObject> dataObj = std::make_shared<DataObject>("objId");
    ASSERT_EQ(notifier->NotifyResultByStream("genId", 1, dataObj, ErrorInfo()).Code(),
              ErrorCode::ERR_INNER_SYSTEM_ERROR);
}

TEST_F(GeneratorTest, IncreaseProducerReferenceTest)
{
    notifier = std::make_shared<StreamGeneratorNotifier>(streamStore, map);
    notifier->IncreaseProducerReference("topic");
    ASSERT_EQ(notifier->producerReferences_["topic"], 1);
    notifier->IncreaseProducerReference("topic");
    ASSERT_EQ(notifier->producerReferences_["topic"], 2);
}

TEST_F(GeneratorTest, InitializeTest)
{
    notifier = std::make_shared<StreamGeneratorNotifier>(streamStore, map);
    notifier->map_ = nullptr;
    notifier->stopped = true;
    ASSERT_NO_THROW(notifier->Initialize());
    notifier->timer_->cancel();
    notifier->timer_.reset();
}

TEST_F(GeneratorTest, PopBatchTest)
{
    notifier = std::make_shared<StreamGeneratorNotifier>(streamStore, map);
    std::vector<std::shared_ptr<GeneratorNotifyData>> datas;
    notifier->genQueue_.push_back(std::make_shared<GeneratorNotifyData>());
    notifier->PopBatch(datas);
    ASSERT_EQ(datas.size(), 1);
}

TEST_F(GeneratorTest, GeneratorIdMapTest)
{
    auto map = std::make_shared<GeneratorIdMap>();
    std::string genId("fake_genid");
    std::string rtId("fake_rtid");
    {
        GeneratorIdRecorder r(genId, rtId, map);
        std::string v;
        map->GetRecord(genId, v);
        EXPECT_EQ(v, rtId);
    }
    std::string v1;
    map->GetRecord(genId, v1);
    EXPECT_TRUE(v1 == rtId);
    map->RemoveRecord(genId);
    std::string v2;
    map->GetRecord(genId, v2);
    EXPECT_TRUE(v2.empty());
}

TEST_F(GeneratorTest, GeneratorIdNotifyTest)
{
    std::shared_ptr<MockStreamStore> streamStore = std::make_shared<MockStreamStore>();
    auto map = std::make_shared<GeneratorIdMap>();
    {
        auto n = std::make_shared<StreamGeneratorNotifier>(streamStore, map);
        std::string genId("fake_genid");
        std::string rtId("fake_rtid");
        std::string objId("fake_objid");
        int idx = 0;
        for (int i = 0; i < 2; i++) {
            GeneratorIdRecorder r(genId, rtId, map);
            auto resultObj = std::make_shared<DataObject>();
            resultObj->id = objId;
            resultObj->buffer = std::make_shared<SharedBuffer>(nullptr, 0);
            auto resultErr = ErrorInfo();
            auto err = n->NotifyResult(genId, idx + i, resultObj, resultErr);
            EXPECT_TRUE(err.OK());
        }
        auto err = n->NotifyFinished(genId, 3);
        EXPECT_TRUE(err.OK());
    }
}

TEST_F(GeneratorTest, GeneratorNotifierTest_NotifyResult)
{
    std::shared_ptr<MockStreamStore> streamStore = std::make_shared<MockStreamStore>();
    auto map = std::make_shared<GeneratorIdMap>();

    std::string genId("fake_genid");
    std::string rtId("fake_rtid");
    std::string objId("fake_objid");
    int idx = 0;

    auto p = std::make_shared<MockStreamProducer>();

    EXPECT_CALL(*streamStore, CreateStreamProducer(rtId, _, _))
        .WillOnce(
            [&p](const std::string &streamName, std::shared_ptr<StreamProducer> &producer, ProducerConf producerConf) {
                producer = p;
                return ErrorInfo();
            });

    EXPECT_CALL(*p, Send(_))
        .WillOnce([&genId, &objId, &idx](const Element &element) {
            libruntime::NotifyGeneratorResult res;
            auto success = res.ParseFromArray(element.ptr, element.size);
            EXPECT_TRUE(success);
            EXPECT_EQ(res.genid(), genId);
            EXPECT_EQ(res.objectid(), objId);
            EXPECT_EQ(res.index(), idx);
            EXPECT_EQ(res.errorcode(), 0);
            EXPECT_EQ(res.finished(), false);
            return ErrorInfo();
        })
        .WillOnce([&genId, &objId, &idx](const Element &element) {
            libruntime::NotifyGeneratorResult res;
            auto success = res.ParseFromArray(element.ptr, element.size);
            EXPECT_TRUE(success);
            EXPECT_EQ(res.genid(), genId);
            EXPECT_EQ(res.objectid(), objId);
            EXPECT_EQ(res.index(), (idx + 1));
            EXPECT_EQ(res.errorcode(), 0);
            EXPECT_EQ(res.finished(), false);
            return ErrorInfo();
        });

    {
        auto n = std::make_shared<StreamGeneratorNotifier>(streamStore, map);
        for (int i = 0; i < 2; i++) {
            GeneratorIdRecorder r(genId, rtId, map);
            auto resultObj = std::make_shared<DataObject>();
            resultObj->id = objId;
            resultObj->buffer = std::make_shared<SharedBuffer>(nullptr, 0);
            auto resultErr = ErrorInfo();
            n->NotifyResultByStream(genId, idx + i, resultObj, resultErr);
        }
    }
}

TEST_F(GeneratorTest, GeneratorNotifierTest_NotifyError)
{
    std::shared_ptr<MockStreamStore> streamStore = std::make_shared<MockStreamStore>();
    auto map = std::make_shared<GeneratorIdMap>();

    std::string genId("fake_genid");
    std::string rtId("fake_rtid");
    std::string objId("fake_objid");
    int idx = 0;
    ErrorCode errCode = ErrorCode::ERR_USER_FUNCTION_EXCEPTION;
    std::string errMsg("failed to execute user func");

    auto p = std::make_shared<MockStreamProducer>();

    EXPECT_CALL(*streamStore, CreateStreamProducer(rtId, _, _))
        .WillOnce(
            [&p](const std::string &streamName, std::shared_ptr<StreamProducer> &producer, ProducerConf producerConf) {
                producer = p;
                return ErrorInfo();
            });

    EXPECT_CALL(*p, Send(_)).WillOnce([&genId, &objId, &idx, &errCode, &errMsg](const Element &element) {
        libruntime::NotifyGeneratorResult res;
        auto success = res.ParseFromArray(element.ptr, element.size);
        EXPECT_TRUE(success);
        EXPECT_EQ(res.genid(), genId);
        EXPECT_EQ(res.objectid(), objId);
        EXPECT_EQ(res.index(), idx);
        EXPECT_EQ(res.errorcode(), int64_t(errCode));
        EXPECT_EQ(res.errormessage(), errMsg);
        EXPECT_EQ(res.finished(), false);
        return ErrorInfo();
    });

    {
        auto n = std::make_shared<StreamGeneratorNotifier>(streamStore, map);
        GeneratorIdRecorder r(genId, rtId, map);
        auto resultObj = std::make_shared<DataObject>();
        resultObj->id = objId;
        resultObj->buffer = std::make_shared<SharedBuffer>(nullptr, 0);
        auto resultErr = ErrorInfo(errCode, errMsg);
        n->NotifyResultByStream(genId, idx, resultObj, resultErr);
        EXPECT_TRUE(n->producerReferences_.find(rtId) == n->producerReferences_.end());
    }
}

TEST_F(GeneratorTest, GeneratorNotifierTest_NotifyFinished)
{
    std::shared_ptr<MockStreamStore> streamStore = std::make_shared<MockStreamStore>();
    auto map = std::make_shared<GeneratorIdMap>();

    std::string genId("fake_genid");
    std::string rtId("fake_rtid");
    int numResults = 10;

    auto p = std::make_shared<MockStreamProducer>();

    EXPECT_CALL(*streamStore, CreateStreamProducer(rtId, _, _))
        .WillOnce(
            [&p](const std::string &streamName, std::shared_ptr<StreamProducer> &producer, ProducerConf producerConf) {
                producer = p;
                return ErrorInfo();
            });

    EXPECT_CALL(*p, Send(_)).WillOnce([&genId, &numResults](const Element &element) {
        libruntime::NotifyGeneratorResult res;
        auto success = res.ParseFromArray(element.ptr, element.size);
        EXPECT_TRUE(success);
        EXPECT_EQ(res.genid(), genId);
        EXPECT_EQ(res.finished(), true);
        EXPECT_EQ(res.numresults(), numResults);
        return ErrorInfo();
    });

    {
        auto n = std::make_shared<StreamGeneratorNotifier>(streamStore, map);
        GeneratorIdRecorder r(genId, rtId, map);
        n->NotifyFinishedByStream(genId, numResults);
        EXPECT_TRUE(n->producerReferences_.find(rtId) == n->producerReferences_.end());
    }
}

TEST_F(GeneratorTest, GeneratorReceiverTest_MarkEndOfStream)
{
    auto librtCfg = std::make_shared<LibruntimeConfig>();
    librtCfg->runtimeId = "driver";
    librtCfg->jobId = "56781234";
    std::shared_ptr<MockStreamStore> streamStore = std::make_shared<MockStreamStore>();
    auto memoryStore = std::make_shared<MemoryStore>();
    auto dsObjectStore = std::make_shared<MockObjectStore>();
    auto wom = std::make_shared<WaitingObjectManager>();
    memoryStore->Init(dsObjectStore, wom);
    wom->SetMemoryStore(memoryStore);

    std::string genId("fake_genid");
    std::string rtId("fake_rtid");
    std::string objId("fake_objid");
    int idx = 0;
    std::string errMsg("failed to execute user func");

    auto c = std::make_shared<MockStreamConsumer>();

    EXPECT_CALL(*streamStore, CreateStreamConsumer("driver_56781234", _, _, true))
        .WillOnce([&c](const std::string &streamName, const SubscriptionConfig &config,
                       std::shared_ptr<StreamConsumer> &consumer, bool autoAck) {
            consumer = c;
            return ErrorInfo();
        });

    void *buffer1;

    EXPECT_CALL(*c, Receive(_, _, _))
        .WillOnce([&genId, &idx, &objId, &buffer1](uint32_t expectNum, uint32_t timeoutMs,
                                                   std::vector<Element> &outElements) {
            libruntime::NotifyGeneratorResult res;
            res.set_genid(genId);
            res.set_index(idx);
            res.set_objectid(objId);
            auto rs = res.ByteSizeLong();
            buffer1 = malloc(rs);
            EXPECT_TRUE(buffer1 != nullptr);
            res.SerializeToArray(buffer1, rs);
            outElements.resize(1);
            outElements[0].ptr = static_cast<uint8_t *>(buffer1);
            outElements[0].size = rs;
            return ErrorInfo();
        })
        .WillRepeatedly([](uint32_t expectNum, uint32_t timeoutMs, std::vector<Element> &outElements) {
            std::this_thread::sleep_for(500ms);
            return ErrorInfo();
        });

    {
        auto r = std::make_shared<StreamGeneratorReceiver>(librtCfg, streamStore, memoryStore);
        r->Initialize();
        std::this_thread::sleep_for(50ms);
        EXPECT_EQ(r->generatorResultsCounter_[genId], 1);
        auto err = ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, "some invoke error");
        r->MarkEndOfStream(genId, err);
        std::this_thread::sleep_for(100ms);
        EXPECT_EQ(r->generatorResultsCounter_.find(genId), r->generatorResultsCounter_.end());
        r->Stop();
    }
    free(buffer1);
}

TEST_F(GeneratorTest, GeneratorReceiverTest_ReceiveError)
{
    auto librtCfg = std::make_shared<LibruntimeConfig>();
    librtCfg->runtimeId = "driver";
    librtCfg->jobId = "56781234";
    std::shared_ptr<MockStreamStore> streamStore = std::make_shared<MockStreamStore>();
    auto memoryStore = std::make_shared<MemoryStore>();
    auto dsObjectStore = std::make_shared<MockObjectStore>();
    auto wom = std::make_shared<WaitingObjectManager>();
    memoryStore->Init(dsObjectStore, wom);
    wom->SetMemoryStore(memoryStore);

    std::string genId("fake_genid");
    std::string rtId("fake_rtid");
    std::string objId("fake_objid");
    int idx = 0;
    ErrorCode errCode = ErrorCode::ERR_USER_FUNCTION_EXCEPTION;
    std::string errMsg("failed to execute user func");

    auto c = std::make_shared<MockStreamConsumer>();

    EXPECT_CALL(*streamStore, CreateStreamConsumer("driver_56781234", _, _, true))
        .WillOnce([&c](const std::string &streamName, const SubscriptionConfig &config,
                       std::shared_ptr<StreamConsumer> &consumer, bool autoAck) {
            consumer = c;
            return ErrorInfo();
        });

    void *buffer1;
    void *buffer2;

    EXPECT_CALL(*c, Receive(_, _, _))
        .WillOnce([&genId, &idx, &objId, &buffer1](uint32_t expectNum, uint32_t timeoutMs,
                                                   std::vector<Element> &outElements) {
            libruntime::NotifyGeneratorResult res;
            res.set_genid(genId);
            res.set_index(idx);
            res.set_objectid(objId);
            auto rs = res.ByteSizeLong();
            buffer1 = malloc(rs);
            EXPECT_TRUE(buffer1 != nullptr);
            res.SerializeToArray(buffer1, rs);
            outElements.resize(1);
            outElements[0].ptr = static_cast<uint8_t *>(buffer1);
            outElements[0].size = rs;
            return ErrorInfo();
        })
        .WillOnce([&genId, &idx, &objId, &errCode, &errMsg, &buffer2](uint32_t expectNum, uint32_t timeoutMs,
                                                                      std::vector<Element> &outElements) {
            std::this_thread::sleep_for(100ms);
            libruntime::NotifyGeneratorResult res;
            res.set_genid(genId);
            res.set_index(idx + 1);
            res.set_objectid(objId);
            res.set_errorcode(int64_t(errCode));
            res.set_errormessage(errMsg);
            auto rs = res.ByteSizeLong();
            buffer2 = malloc(rs);
            EXPECT_TRUE(buffer2 != nullptr);
            res.SerializeToArray(buffer2, rs);
            outElements.resize(1);
            outElements[0].ptr = static_cast<uint8_t *>(buffer2);
            outElements[0].size = rs;
            return ErrorInfo();
        })
        .WillRepeatedly([](uint32_t expectNum, uint32_t timeoutMs, std::vector<Element> &outElements) {
            std::this_thread::sleep_for(500ms);
            return ErrorInfo();
        });

    {
        auto r = std::make_shared<StreamGeneratorReceiver>(librtCfg, streamStore, memoryStore);
        r->Initialize();
        std::this_thread::sleep_for(50ms);
        EXPECT_EQ(r->generatorResultsCounter_[genId], 1);
        std::this_thread::sleep_for(200ms);
        EXPECT_EQ(r->generatorResultsCounter_.find(genId), r->generatorResultsCounter_.end());
        r->Stop();
    }

    free(buffer1);
    free(buffer2);
}

TEST_F(GeneratorTest, GeneratorReceiverTest_ReceiveFinished)
{
    auto librtCfg = std::make_shared<LibruntimeConfig>();
    librtCfg->runtimeId = "driver";
    librtCfg->jobId = "56781234";
    std::shared_ptr<MockStreamStore> streamStore = std::make_shared<MockStreamStore>();
    auto memoryStore = std::make_shared<MemoryStore>();
    auto dsObjectStore = std::make_shared<MockObjectStore>();
    auto wom = std::make_shared<WaitingObjectManager>();
    memoryStore->Init(dsObjectStore, wom);
    wom->SetMemoryStore(memoryStore);
    std::string genId("fake_genid");
    std::string rtId("fake_rtid");
    std::string objId("fake_objid");
    int idx = 0;

    auto c = std::make_shared<MockStreamConsumer>();

    EXPECT_CALL(*streamStore, CreateStreamConsumer("driver_56781234", _, _, true))
        .WillOnce([&c](const std::string &streamName, const SubscriptionConfig &config,
                       std::shared_ptr<StreamConsumer> &consumer, bool autoAck) {
            consumer = c;
            return ErrorInfo();
        });

    void *buffer1;
    void *buffer2;

    EXPECT_CALL(*c, Receive(_, _, _))
        .WillOnce([&genId, &idx, &objId, &buffer1](uint32_t expectNum, uint32_t timeoutMs,
                                                   std::vector<Element> &outElements) {
            libruntime::NotifyGeneratorResult res;
            res.set_genid(genId);
            res.set_index(idx);
            res.set_objectid(objId);
            auto rs = res.ByteSizeLong();
            buffer1 = malloc(rs);
            EXPECT_TRUE(buffer1 != nullptr);
            res.SerializeToArray(buffer1, rs);
            outElements.resize(1);
            outElements[0].ptr = static_cast<uint8_t *>(buffer1);
            outElements[0].size = rs;
            return ErrorInfo();
        })
        .WillOnce([&genId, &buffer2](uint32_t expectNum, uint32_t timeoutMs, std::vector<Element> &outElements) {
            std::this_thread::sleep_for(100ms);
            libruntime::NotifyGeneratorResult res;
            res.set_genid(genId);
            res.set_finished(true);
            res.set_numresults(1);
            auto rs = res.ByteSizeLong();
            buffer2 = malloc(rs);
            EXPECT_TRUE(buffer2 != nullptr);
            res.SerializeToArray(buffer2, rs);
            outElements.resize(1);
            outElements[0].ptr = static_cast<uint8_t *>(buffer2);
            outElements[0].size = rs;
            return ErrorInfo();
        })
        .WillRepeatedly([](uint32_t expectNum, uint32_t timeoutMs, std::vector<Element> &outElements) {
            std::this_thread::sleep_for(500ms);
            return ErrorInfo();
        });

    {
        auto r = std::make_shared<StreamGeneratorReceiver>(librtCfg, streamStore, memoryStore);
        r->Initialize();
        std::this_thread::sleep_for(50ms);
        EXPECT_EQ(r->generatorResultsCounter_[genId], 1);
        for (int i = 0; i < 100; i++) {
            std::this_thread::sleep_for(50ms);
            if (r->generatorResultsCounter_.find(genId) == r->generatorResultsCounter_.end()) {
                break;
            }
        }
        EXPECT_EQ(r->generatorResultsCounter_.find(genId), r->generatorResultsCounter_.end());
        r->Stop();
    }

    free(buffer1);
    free(buffer2);
}

TEST_F(GeneratorTest, UpdateRecordAndGetRecordTest)
{
    auto map = std::make_shared<GeneratorIdMap>();
    std::string key = "key";
    std::string value = "value";
    map->UpdateRecord(key, value);
    ASSERT_EQ(map->records_[key], value);

    std::vector<std::string> vec;
    map->GetRecordKeys(vec);
    ASSERT_EQ(vec[0], key);
}

}  // namespace test
}  // namespace YR