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
#include "api/cpp/src/config_manager.h"
#include "api/cpp/src/internal.h"
#include "api/cpp/src/read_only_buffer.h"
#include "api/cpp/src/stream_pubsub.h"
#include "src/dto/buffer.h"
#include "yr/api/config.h"
#include "yr/api/hetero_manager.h"
#include "yr/api/mutable_buffer.h"
#include "yr/api/object_store.h"
#include "yr/yr.h"

using namespace testing;

class MockRuntime : public YR::Runtime {
public:
    MOCK_METHOD0(Init, void());
    MOCK_METHOD0(GetServerVersion, std::string());
    MOCK_METHOD2(Put, std::string(std::shared_ptr<msgpack::sbuffer>, const std::unordered_set<std::string> &));
    MOCK_METHOD2(Put, std::string(std::shared_ptr<YR::Buffer>, const std::unordered_set<std::string> &));
    MOCK_METHOD3(Put, std::string(std::shared_ptr<msgpack::sbuffer>, const std::unordered_set<std::string> &,
                                  const YR::CreateParam &));
    MOCK_METHOD3(Put, std::string(std::shared_ptr<YR::Buffer>, const std::unordered_set<std::string> &,
                                  const YR::CreateParam &));
    MOCK_METHOD3(Put,
                 void(const std::string &, std::shared_ptr<msgpack::sbuffer>, const std::unordered_set<std::string> &));
    MOCK_METHOD3(KVMSetTx, void(const std::vector<std::string> &,
                                const std::vector<std::shared_ptr<msgpack::sbuffer>> &, YR::ExistenceOpt));
    MOCK_METHOD3(KVMSetTx, void(const std::vector<std::string> &,
                                const std::vector<std::shared_ptr<msgpack::sbuffer>> &, const YR::MSetParam &));
    MOCK_METHOD3(Get, std::pair<YR::internal::RetryInfo, std::vector<std::shared_ptr<YR::Buffer>>>(
                          const std::vector<std::string> &, int, int &));
    MOCK_METHOD3(Wait, YR::internal::WaitResult(const std::vector<std::string> &, std::size_t, int));
    MOCK_METHOD3(KVWrite, void(const std::string &, std::shared_ptr<msgpack::sbuffer>, YR::SetParam));
    MOCK_METHOD3(KVWrite, void(const std::string &, std::shared_ptr<msgpack::sbuffer>, YR::SetParamV2));
    MOCK_METHOD3(KVWrite, void(const std::string &key, const char *value, YR::SetParam setParam));
    MOCK_METHOD4(KVWrite, void(const std::string &key, const char *value, size_t len, YR::SetParam setParam));
    MOCK_METHOD2(KVRead, std::shared_ptr<YR::Buffer>(const std::string &, int));
    MOCK_METHOD3(KVRead, std::vector<std::shared_ptr<YR::Buffer>>(const std::vector<std::string> &, int, bool));
    MOCK_METHOD3(KVGetWithParam, std::vector<std::shared_ptr<YR::Buffer>>(const std::vector<std::string> &keys,
                                                                          const YR::GetParams &params, int timeout));
    MOCK_METHOD2(KVDel, void(const std::string &, const YR::DelParam &));
    MOCK_METHOD2(KVDel, std::vector<std::string>(const std::vector<std::string> &, const YR::DelParam &));
    MOCK_METHOD1(KVExist, std::vector<bool>(const std::vector<std::string> &));
    MOCK_METHOD2(CreateStreamProducer, std::shared_ptr<YR::Producer>(const std::string &, YR::ProducerConf));
    MOCK_METHOD3(CreateStreamConsumer,
                 std::shared_ptr<YR::Consumer>(const std::string &, const YR::SubscriptionConfig &, bool));
    MOCK_METHOD1(DeleteStream, void(const std::string &));
    MOCK_METHOD2(QueryGlobalProducersNum, void(const std::string &, uint64_t &));
    MOCK_METHOD2(QueryGlobalConsumersNum, void(const std::string &, uint64_t &));
    MOCK_METHOD2(IncreGlobalReference, void(const std::vector<std::string> &, bool));
    MOCK_METHOD1(DecreGlobalReference, void(const std::vector<std::string> &));
    MOCK_METHOD3(InvokeByName, std::string(const YR::internal::FuncMeta &, std::vector<YR::internal::InvokeArg> &,
                                           const YR::InvokeOptions &));
    MOCK_METHOD3(CreateInstance, std::string(const YR::internal::FuncMeta &, std::vector<YR::internal::InvokeArg> &,
                                             YR::InvokeOptions &));
    MOCK_METHOD4(InvokeInstance, std::string(const YR::internal::FuncMeta &, const std::string &,
                                             std::vector<YR::internal::InvokeArg> &, const YR::InvokeOptions &));
    MOCK_METHOD1(GetRealInstanceId, std::string(const std::string &));
    MOCK_METHOD3(SaveRealInstanceId, void(const std::string &, const std::string &, const YR::InvokeOptions &));
    MOCK_METHOD1(GetGroupInstanceIds, std::string(const std::string &));
    MOCK_METHOD3(SaveGroupInstanceIds, void(const std::string &, const std::string &, const YR::InvokeOptions &));
    MOCK_METHOD3(Cancel, void(const std::vector<std::string> &, bool, bool));
    MOCK_METHOD1(TerminateInstance, void(const std::string &));
    MOCK_METHOD2(TerminateInstanceAsync, std::shared_future<void>(const std::string &instanceId, bool isSync));
    MOCK_METHOD0(Exit, void());
    MOCK_METHOD0(IsOnCloud, bool());
    MOCK_METHOD2(GroupCreate, void(const std::string &, YR::GroupOptions &));
    MOCK_METHOD1(GroupTerminate, void(const std::string &));
    MOCK_METHOD1(GroupWait, void(const std::string &));
    MOCK_METHOD2(GetInstances, std::vector<std::string>(const std::string &, int));
    MOCK_METHOD2(GetInstances, std::vector<std::string>(const std::string &, const std::string &));
    MOCK_METHOD0(GenerateGroupName, std::string());
    MOCK_METHOD1(SaveState, void(const int &));
    MOCK_METHOD1(LoadState, void(const int &));
    MOCK_METHOD3(WaitBeforeGet, int64_t(const std::vector<std::string> &ids, int timeoutMs, bool allowPartial));
    MOCK_METHOD(void, DevDelete,
                (const std::vector<std::string> &objectIds, std::vector<std::string> &failedObjectIds));

    MOCK_METHOD(void, DevLocalDelete,
                (const std::vector<std::string> &objectIds, std::vector<std::string> &failedObjectIds));

    MOCK_METHOD3(DevSubscribe,
                 void(const std::vector<std::string> &keys, const std::vector<YR::DeviceBlobList> &blob2dList,
                      std::vector<std::shared_ptr<YR::Future>> &futureVec));
    MOCK_METHOD3(DevPublish,
                 void(const std::vector<std::string> &keys, const std::vector<YR::DeviceBlobList> &blob2dList,
                      std::vector<std::shared_ptr<YR::Future>> &futureVec));
    MOCK_METHOD3(DevMSet, void(const std::vector<std::string> &keys, const std::vector<YR::DeviceBlobList> &blob2dList,
                               std::vector<std::string> &failedKeys));
    MOCK_METHOD4(DevMGet, void(const std::vector<std::string> &keys, const std::vector<YR::DeviceBlobList> &blob2dList,
                               std::vector<std::string> &failedKeys, int32_t timeoutMs));
    MOCK_METHOD3(GetInstance,
                 YR::internal::FuncMeta(const std::string &name, const std::string &nameSpace, int timeoutSec));
    MOCK_METHOD1(GetInstanceRoute, std::string(const std::string &objectId));
    MOCK_METHOD2(SaveInstanceRoute, void(const std::string &objectId, const std::string &instanceRoute));
    MOCK_METHOD1(TerminateInstanceSync, void(const std::string &instanceId));
    MOCK_METHOD0(Nodes, std::vector<YR::Node>());
    MOCK_METHOD1(CreateMutableBuffer, std::shared_ptr<YR::MutableBuffer>(uint64_t size));
    MOCK_METHOD2(GetMutableBuffer,
                 std::vector<std::shared_ptr<YR::MutableBuffer>>(const std::vector<std::string> &ids, int timeout));
};

class ApiTest : public ::testing::Test {
public:
    ApiTest() {}
    ~ApiTest() {}
    void SetUp() override
    {
        this->runtime = std::make_shared<MockRuntime>();
        YR::internal::RuntimeManager::GetInstance().mode_ = YR::Config::Mode::CLUSTER_MODE;
        YR::internal::RuntimeManager::GetInstance().yrRuntime = runtime;
        YR::SetInitialized(true);
    }

    void TearDown() override
    {
        try {
            YR::Finalize();
        } catch (YR::Exception &e) {
            std::cout << "Exception occurs when Finalize: " << e.what() << std::endl;
        }
    }

    std::shared_ptr<MockRuntime> runtime;
};

TEST_F(ApiTest, WaitTwoObjectRefSuccessfully)
{
    YR::internal::WaitResult waitResult;
    waitResult.readyIds.push_back("ready");
    waitResult.unreadyIds.push_back("unready");
    EXPECT_CALL(*this->runtime, Wait(_, _, _)).WillOnce(testing::Return(waitResult));

    std::vector<YR::ObjectRef<int>> objs{YR::ObjectRef<int>("ready"), YR::ObjectRef<int>("unready")};

    auto result = YR::Wait(objs, 1);
    ASSERT_EQ(result.first.size(), 1);
    ASSERT_EQ(result.first[0].ID(), "ready");
    ASSERT_EQ(result.second.size(), 1);
    ASSERT_EQ(result.second[0].ID(), "unready");
}

TEST_F(ApiTest, WaitNumIllegalFailed)
{
    std::vector<YR::ObjectRef<int>> objs{YR::ObjectRef<int>("ready"), YR::ObjectRef<int>("unready")};
    ASSERT_THROW(YR::Wait(objs, 0), YR::Exception);
    YR::internal::WaitResult waitResult;
    waitResult.readyIds.push_back("ready");
    waitResult.unreadyIds.push_back("unready");
    EXPECT_CALL(*this->runtime, Wait(_, 2, _)).WillOnce(testing::Return(waitResult));
    ASSERT_NO_THROW(YR::Wait(objs, 3));
}

struct Defer {
    explicit Defer(std::function<void()> &&f)
    {
        deferFunc = std::move(f);
    }
    ~Defer()
    {
        deferFunc();
    }
    std::function<void()> deferFunc = nullptr;
};

void Init(int &cnt)
{
    static bool hasRegisteredAtExit = false;
    Defer registerAtExitHandler([&cnt]() {
        if (!hasRegisteredAtExit) {
            cnt++;
            hasRegisteredAtExit = true;
        }
    });
}

TEST_F(ApiTest, DuplicateRegisterAtExit)
{
    int cnt = 0;
    for (int i = 0; i < 10; i++) {
        Init(cnt);
    }
    EXPECT_EQ(cnt, 1);
}

TEST_F(ApiTest, KVSetTest)
{
    YR::SetParam param;
    param.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
    ASSERT_NO_THROW(YR::KV().Set("key", "value", param));

    std::string result{"result"};
    ASSERT_NO_THROW(YR::KV().Set("key", result.c_str()));

    auto size = result.size() + 1;
    ASSERT_NO_THROW(YR::KV().Set("key", result.c_str(), size));

    ASSERT_NO_THROW(YR::KV().Set("key", "value"));

    ASSERT_NO_THROW(YR::KV().Set("key", result.c_str(), param));

    ASSERT_NO_THROW(YR::KV().Set("key", result.c_str(), size, param));
}

TEST_F(ApiTest, KVMSetTxTest)
{
    YR::MSetParam param1{.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT};
    YR::MSetParam param2{.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT,
                         .ttlSecond = 10,
                         .existence = YR::ExistenceOpt::NONE,
                         .cacheType = YR::CacheType::DISK};
    std::vector<std::string> keys = {"key1", "key2"};
    std::string val1{"val1"};
    std::string val2{"val2"};
    std::vector<char *> cvals = {const_cast<char *>(val1.data()), const_cast<char *>(val2.data())};
    std::vector<std::string> svals = {val1, val2};
    std::vector<size_t> lengths = {val1.size() + 1, val2.size() + 1};
    ASSERT_NO_THROW(YR::KV().MSetTx({"key1", "key2"}, cvals, lengths, YR::ExistenceOpt::NX));
    ASSERT_THROW(YR::KV().MSetTx({"key1"}, cvals, lengths, YR::ExistenceOpt::NX), YR::Exception);
    ASSERT_THROW(YR::KV().MSetTx({"key1", "key2"}, cvals, lengths, YR::ExistenceOpt::NONE), YR::Exception);

    ASSERT_NO_THROW(YR::KV().MSetTx({"key1", "key2"}, svals, YR::ExistenceOpt::NX));
    ASSERT_THROW(YR::KV().MSetTx({"key1"}, svals, YR::ExistenceOpt::NX), YR::Exception);
    ASSERT_THROW(YR::KV().MSetTx({"key1", "key2"}, svals, YR::ExistenceOpt::NONE), YR::Exception);

    ASSERT_NO_THROW(YR::KV().MSetTx({"key1", "key2"}, cvals, lengths, param1));
    ASSERT_THROW(YR::KV().MSetTx({"key1"}, cvals, lengths, param1), YR::Exception);
    ASSERT_THROW(YR::KV().MSetTx({"key1", "key2"}, cvals, lengths, param2), YR::Exception);

    ASSERT_NO_THROW(YR::KV().MSetTx({"key1", "key2"}, svals, param1));
    ASSERT_THROW(YR::KV().MSetTx({"key1"}, svals, param1), YR::Exception);
    ASSERT_THROW(YR::KV().MSetTx({"key1", "key2"}, svals, param2), YR::Exception);
}

TEST_F(ApiTest, CheckInitializedTest)
{
    YR::SetInitialized(false);
    ASSERT_THROW(YR::CheckInitialized(), YR::Exception);
    YR::SetInitialized(true);
    ASSERT_NO_THROW(YR::CheckInitialized());
}

TEST_F(ApiTest, ExitTest)
{
    EXPECT_CALL(*this->runtime, Exit()).WillOnce(testing::Return());
    ASSERT_NO_THROW(YR::Exit());
}

TEST_F(ApiTest, ExistTest)
{
    std::vector<bool> exists = {false, false, true};
    EXPECT_CALL(*this->runtime, KVExist(_)).WillOnce(testing::Return(exists));
    ASSERT_EQ(YR::KV().Exist({"key1", "key2", "key3"}), exists);
}

TEST_F(ApiTest, IsOnCloudTest)
{
    ASSERT_TRUE(YR::IsOnCloud());
}

TEST_F(ApiTest, IsLocalModeTest)
{
    YR::SetInitialized(false);
    ASSERT_THROW(YR::IsLocalMode(), YR::Exception);
    YR::SetInitialized(true);
    ASSERT_FALSE(YR::IsLocalMode());
}

TEST_F(ApiTest, LocalStreamTest)
{
    YR::Config conf;
    conf.mode = YR::Config::Mode::LOCAL_MODE;
    int mockArgc = 5;
    char *mockArgv[] = {"--logDir=/tmp/log", "--logLevel=DEBUG", "--grpcAddress=12.34.56.78:1234", "--runtimeId=driver",
                        "jobId=job123"};
    YR::ConfigManager::Singleton().Init(conf, mockArgc, mockArgv);
    YR::ProducerConf pConf;
    YR::SubscriptionConfig sConf;
    ASSERT_THROW(YR::CreateProducer("streamName", pConf), YR::Exception);
    ASSERT_THROW(YR::Subscribe("streamName", sConf, true), YR::Exception);
    ASSERT_THROW(YR::DeleteStream("streamName"), YR::Exception);
}

TEST_F(ApiTest, StreamTest)
{
    YR::Config conf;
    conf.mode = YR::Config::Mode::CLUSTER_MODE;
    int mockArgc = 5;
    char *mockArgv[] = {"--logDir=/tmp/log", "--logLevel=DEBUG", "--grpcAddress=12.34.56.78:1234", "--runtimeId=driver",
                        "jobId=job123"};
    YR::ConfigManager::Singleton().Init(conf, mockArgc, mockArgv);
    YR::ProducerConf pConf;
    YR::SubscriptionConfig sConf;
    auto streamProducer = std::make_shared<YR::Libruntime::StreamProducer>();
    std::shared_ptr<YR::Producer> producer = std::make_shared<YR::StreamProducer>(streamProducer);
    EXPECT_CALL(*this->runtime, CreateStreamProducer(_, _)).WillOnce(testing::Return(producer));
    ASSERT_NO_THROW(YR::CreateProducer("streamName", pConf));

    auto streamConsumer = std::make_shared<YR::Libruntime::StreamConsumer>();
    std::shared_ptr<YR::Consumer> consumer = std::make_shared<YR::StreamConsumer>(streamConsumer);
    EXPECT_CALL(*this->runtime, CreateStreamConsumer(_, _, _)).WillOnce(testing::Return(consumer));
    ASSERT_NO_THROW(YR::Subscribe("streamName", sConf, true));

    EXPECT_CALL(*this->runtime, DeleteStream(_)).WillOnce(testing::Return());
    ASSERT_NO_THROW(YR::DeleteStream("streamName"));
}

TEST_F(ApiTest, SaveLoadStateThrowTest)
{
    YR::Config conf;
    conf.mode = YR::Config::Mode::LOCAL_MODE;
    conf.isDriver = true;
    int mockArgc = 5;
    char *mockArgv[] = {"--logDir=/tmp/log", "--logLevel=DEBUG", "--grpcAddress=127.0.0.1:1234", "--runtimeId=driver",
                        "jobId=job123"};
    YR::ConfigManager::Singleton().Init(conf, mockArgc, mockArgv);
    ASSERT_THROW(YR::SaveState(10), YR::Exception);
    ASSERT_THROW(YR::LoadState(10), YR::Exception);
}

TEST_F(ApiTest, SaveLoadStateFailedTest)
{
    YR::Config conf;
    conf.mode = YR::Config::Mode::CLUSTER_MODE;
    conf.isDriver = true;
    int mockArgc = 5;
    char *mockArgv[] = {"--logDir=/tmp/log", "--logLevel=DEBUG", "--grpcAddress=127.0.0.1:1234", "--runtimeId=driver",
                        "jobId=job123"};
    YR::ConfigManager::Singleton().Init(conf, mockArgc, mockArgv);
    EXPECT_THROW(YR::SaveState(10), YR::Exception);
    EXPECT_THROW(YR::LoadState(10), YR::Exception);
}

TEST_F(ApiTest, GroupTest)
{
    YR::GroupOptions gOpts;
    std::string gName = "gName";
    YR::Group group(gName, gOpts);
    YR::Group group2(gName, std::move(gOpts));
    EXPECT_CALL(*this->runtime, GroupCreate(_, _)).WillOnce(testing::Return());
    EXPECT_CALL(*this->runtime, GroupTerminate(_)).WillOnce(testing::Return());
    EXPECT_CALL(*this->runtime, GroupWait(_)).WillOnce(testing::Return());
    ASSERT_NO_THROW(group.Invoke());
    ASSERT_NO_THROW(group.Wait());
    ASSERT_NO_THROW(group.Terminate());
    ASSERT_EQ(group.GetGroupName(), "gName");
}

TEST_F(ApiTest, GroupBindOptionsTest)
{
    YR::GroupOptions gOpts;
    std::string gName = "gName";
    gOpts.bind.resource = "NUMA";
    gOpts.bind.strategy = "SPREAD";
    YR::Group group(gName, gOpts);
    ASSERT_EQ(group.GetGroupName(), "gName");
    ASSERT_EQ(group.groupOpts.bind.resource, "NUMA");
    ASSERT_EQ(group.groupOpts.bind.strategy, "SPREAD");
}

TEST_F(ApiTest, HeteroDeleteTest)
{
    std::vector<std::string> objectIds;
    std::vector<std::string> failedObjectIds;
    YR::internal::RuntimeManager::GetInstance().mode_ = YR::Config::Mode::LOCAL_MODE;
    ASSERT_THROW(YR::HeteroManager().DevDelete(objectIds, failedObjectIds), YR::HeteroException);
    YR::internal::RuntimeManager::GetInstance().mode_ = YR::Config::Mode::CLUSTER_MODE;
    ASSERT_NO_THROW(YR::HeteroManager().DevDelete(objectIds, failedObjectIds));
}

TEST_F(ApiTest, HeteroDevLocalDeleteTest)
{
    std::vector<std::string> objIds;
    std::vector<std::string> failedObjIds;
    YR::internal::RuntimeManager::GetInstance().mode_ = YR::Config::Mode::LOCAL_MODE;
    ASSERT_THROW(YR::HeteroManager().DevLocalDelete(objIds, failedObjIds), YR::HeteroException);
    YR::internal::RuntimeManager::GetInstance().mode_ = YR::Config::Mode::CLUSTER_MODE;
    ASSERT_NO_THROW(YR::HeteroManager().DevLocalDelete(objIds, failedObjIds));
}

TEST_F(ApiTest, HeteroDevSubscribeTest)
{
    std::vector<std::string> keys;
    std::vector<YR::DeviceBlobList> blob2dList;
    std::vector<std::shared_ptr<YR::Future>> futureVec;
    YR::internal::RuntimeManager::GetInstance().mode_ = YR::Config::Mode::LOCAL_MODE;
    ASSERT_THROW(YR::HeteroManager().DevSubscribe(keys, blob2dList, futureVec), YR::HeteroException);
    YR::internal::RuntimeManager::GetInstance().mode_ = YR::Config::Mode::CLUSTER_MODE;
    ASSERT_NO_THROW(YR::HeteroManager().DevSubscribe(keys, blob2dList, futureVec));
    keys.push_back("key");
    ASSERT_THROW(YR::HeteroManager().DevSubscribe(keys, blob2dList, futureVec), YR::HeteroException);
}

TEST_F(ApiTest, HeteroDevPublishTest)
{
    std::vector<std::string> keyList;
    std::vector<YR::DeviceBlobList> blobTodList;
    std::vector<std::shared_ptr<YR::Future>> futureVec;
    YR::internal::RuntimeManager::GetInstance().mode_ = YR::Config::Mode::LOCAL_MODE;
    ASSERT_THROW(YR::HeteroManager().DevPublish(keyList, blobTodList, futureVec), YR::HeteroException);
    YR::internal::RuntimeManager::GetInstance().mode_ = YR::Config::Mode::CLUSTER_MODE;
    ASSERT_NO_THROW(YR::HeteroManager().DevPublish(keyList, blobTodList, futureVec));
    keyList.push_back("key");
    ASSERT_THROW(YR::HeteroManager().DevPublish(keyList, blobTodList, futureVec), YR::HeteroException);
}

TEST_F(ApiTest, HeteroDevMSetTest)
{
    std::vector<std::string> keysList;
    std::vector<YR::DeviceBlobList> blobsTodList;
    std::vector<std::string> failedKeys;
    YR::internal::RuntimeManager::GetInstance().mode_ = YR::Config::Mode::LOCAL_MODE;
    ASSERT_THROW(YR::HeteroManager().DevMSet(keysList, blobsTodList, failedKeys), YR::HeteroException);
    YR::internal::RuntimeManager::GetInstance().mode_ = YR::Config::Mode::CLUSTER_MODE;
    ASSERT_NO_THROW(YR::HeteroManager().DevMSet(keysList, blobsTodList, failedKeys));
    keysList.push_back("key");
    ASSERT_THROW(YR::HeteroManager().DevMSet(keysList, blobsTodList, failedKeys), YR::HeteroException);
}

TEST_F(ApiTest, HeteroDevMGetTest)
{
    std::vector<std::string> keyLists;
    std::vector<YR::DeviceBlobList> blobs2dLists;
    std::vector<std::string> failedKeys;
    ASSERT_THROW(YR::HeteroManager().DevMGet(keyLists, blobs2dLists, failedKeys, 0), YR::HeteroException);
    YR::internal::RuntimeManager::GetInstance().mode_ = YR::Config::Mode::LOCAL_MODE;
    ASSERT_THROW(YR::HeteroManager().DevMGet(keyLists, blobs2dLists, failedKeys, 1), YR::HeteroException);
    YR::internal::RuntimeManager::GetInstance().mode_ = YR::Config::Mode::CLUSTER_MODE;
    ASSERT_NO_THROW(YR::HeteroManager().DevMGet(keyLists, blobs2dLists, failedKeys, 1));
    keyLists.push_back("key");
    ASSERT_THROW(YR::HeteroManager().DevMGet(keyLists, blobs2dLists, failedKeys, 1), YR::HeteroException);
}

TEST_F(ApiTest, APIGetInstanceTest)
{
    std::string name = "name";
    std::string ns = "ns";
    YR::internal::FuncMeta funcMeta;
    funcMeta.name = "ins-name";
    EXPECT_CALL(*this->runtime, GetInstance(_, _, _)).WillOnce(testing::Return(funcMeta));
    auto handler = YR::GetInstance<int>(name, ns, 60);
    ASSERT_EQ(handler.name, "ins-name");
}

TEST_F(ApiTest, APIPutAndGetSuccessfully)
{
    std::string objId = "abc";
    EXPECT_CALL(*this->runtime,
                Put(testing::An<std::shared_ptr<YR::Buffer>>(), testing::An<const std::unordered_set<std::string> &>()))
        .WillOnce(testing::Return(objId));
    std::string str = "success";
    YR::Buffer buf(str.data(), str.size());
    auto obj = YR::Put(buf);
    ASSERT_EQ(obj.objId, objId);

    auto bufPtr = std::make_shared<YR::Buffer>(buf);
    YR::internal::RetryInfo returnRetryInfo;
    returnRetryInfo.needRetry = true;
    std::vector<std::shared_ptr<YR::Buffer>> buffers;
    buffers.push_back(bufPtr);
    EXPECT_CALL(*this->runtime, Get(_, _, _))
        .WillOnce(testing::Return(
            std::pair<YR::internal::RetryInfo, std::vector<std::shared_ptr<YR::Buffer>>>(returnRetryInfo, buffers)));
    auto value = YR::Get(obj);
    std::string result = std::string(static_cast<const char *>(value->ImmutableData()), value->GetSize());
    ASSERT_EQ(result, str);
}

TEST_F(ApiTest, TestInvokeByName)
{
    YR::SetInitialized(true);
    std::string str = "success";
    YR::Buffer buf(str.data(), str.size());

    std::string objId = "abc";
    std::vector<YR::internal::InvokeArg> invokeArgs;
    EXPECT_CALL(*this->runtime, InvokeByName(_, _, _))
        .WillOnce(::testing::Invoke([this, objId, &invokeArgs](auto &&arg1, auto &&arg2, auto &&arg3) {
            invokeArgs = std::move(arg2);
            return objId;
        }));
    auto ret = YR::PyFunction<YR::Buffer>("common", "echo")
                   .SetUrn("sn:cn:yrk:default:function:0-yr-stpython:$latest")
                   .Invoke(buf);
    ASSERT_EQ(ret.ID(), objId);
    // 校验mock函数入参
    std::string result =
        std::string(static_cast<const char *>(invokeArgs[1].yrBuf.ImmutableData()), invokeArgs[1].yrBuf.GetSize());
    ASSERT_EQ(result, str);
}

TEST_F(ApiTest, TestNodes)
{
    // 定义资源单元数据
    std::unordered_map<std::string, float> resourceMap = {{"cpu", 500.0f}, {"memory", 500.0f}};
    std::unordered_map<std::string, std::vector<std::string>> labelsMap = {{"NODE_ID", {"node1", "node2"}}};
    std::vector<YR::Node> nodesVector = {{.id = "node1", .alive = true, .resources = resourceMap, .labels = labelsMap}};
    EXPECT_CALL(*this->runtime, Nodes()).WillOnce(testing::Return(nodesVector));

    std::vector<YR::Node> nodes = YR::Nodes();
    EXPECT_EQ(nodes.size(), 1);
    EXPECT_EQ(nodes[0].id, "node1");
    EXPECT_TRUE(nodes[0].alive);
    EXPECT_EQ(nodes[0].resources["cpu"], 500.0f);
    EXPECT_EQ(nodes[0].resources["memory"], 500.0f);
    EXPECT_EQ(nodes[0].labels["NODE_ID"][0], "node1");
}

TEST_F(ApiTest, TestCreateMutableBuffer)
{
    ASSERT_NO_THROW(YR::CreateBuffer(60));
}

TEST_F(ApiTest, TestGetMutableBuffer)
{
    std::vector<YR::ObjectRef<YR::MutableBuffer>> objs;
    objs.push_back(YR::ObjectRef<YR::MutableBuffer>("id"));
    ASSERT_NO_THROW(YR::Get(objs, 60));
}

TEST_F(ApiTest, TestSerializeMutableBuffer)
{
    YR::ObjectRef<YR::MutableBuffer> obj("id");
    ASSERT_NO_THROW(YR::Serialize(obj));
}