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

#include <thread>

#include <gmock/gmock.h>
#include <gtest/gtest.h>
#include "common/mock_libruntime.h"
#define private public
#include "api/go/libruntime/cpplibruntime/clibruntime.h"
#include "datasystem/kv_client.h"
#include "src/libruntime/libruntime_manager.h"

using namespace YR::utility;

class FakeStreamProducer : public YR::Libruntime::StreamProducer {
public:
    MOCK_METHOD1(Send, YR::Libruntime::ErrorInfo(const YR::Libruntime::Element &element));
    MOCK_METHOD2(Send, YR::Libruntime::ErrorInfo(const YR::Libruntime::Element &element, int64_t timeoutMs));
    MOCK_METHOD0(Flush, YR::Libruntime::ErrorInfo());
    MOCK_METHOD0(Close, YR::Libruntime::ErrorInfo());
};

class FakeStreamConsumer : public YR::Libruntime::StreamConsumer {
public:
    MOCK_METHOD3(Receive, YR::Libruntime::ErrorInfo(uint32_t expectNum, uint32_t timeoutMs,
                                                    std::vector<YR::Libruntime::Element> &outElements));
    MOCK_METHOD2(Receive,
                 YR::Libruntime::ErrorInfo(uint32_t timeoutMs, std::vector<YR::Libruntime::Element> &outElements));
    MOCK_METHOD1(Ack, YR::Libruntime::ErrorInfo(uint64_t elementId));
    MOCK_METHOD0(Close, YR::Libruntime::ErrorInfo());
};

void SafeFreeCErr(CErrorInfo cErr)
{
    if (cErr.message != nullptr) {
        free(cErr.message);
    }
}

CErrorInfo *GoLoadFunctions(char **codePaths, int size_codePaths)
{
    return nullptr;
}

CErrorInfo *GoFunctionExecution(CFunctionMeta *, CInvokeType, CArg *, int, CDataObject *, int)
{
    return nullptr;
}

CErrorInfo *GoCheckpoint(char *checkpointId, CBuffer *buffer)
{
    return nullptr;
}

CErrorInfo *GoRecover(CBuffer *buffer)
{
    return nullptr;
}

CErrorInfo *GoShutdown(uint64_t gracePeriodSeconds)
{
    return nullptr;
}

CErrorInfo *GoSignal(int sigNo, CBuffer *payload)
{
    return nullptr;
}

CHealthCheckCode GoHealthCheck(void)
{
    return CHealthCheckCode::HEALTHY;
}

char GoHasHealthCheck(void)
{
    return 0;
}

void GoFunctionExecutionPoolSubmit(void *ptr) {}

void GoRawCallback(char *cKey, CErrorInfo cErr, CBuffer cResultRaw) {}

void GoGetAsyncCallback(char *cObjectID, CBuffer cBuf, CErrorInfo *cErr, void *userData) {}

void GoWaitAsyncCallback(char *cObjectID, CErrorInfo *cErr, void *userData) {}

void GoGetEventCallback(char *cObjectID, CBuffer cBuf, CErrorInfo *cErr, void *userData) {}

void freeCStrings(char **cStrings, size_t cStringsLen)
{
    for (size_t i = 0; i < cStringsLen; i++) {
        free(cStrings[i]);
    }
    free(cStrings);
}

void freeCErrorIds(CErrorObject **errorIds, int size_errorIds)
{
    for (size_t i = 0; i < size_errorIds; i++) {
        free(errorIds[i]->objectId);
        SafeFreeCErr(*(errorIds[i]->errorInfo));
    }
    free(errorIds);
}

class CLibruntimeTest : public testing::Test {
public:
    CLibruntimeTest() {};
    ~CLibruntimeTest() {};
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
        auto lc = std::make_shared<YR::Libruntime::LibruntimeConfig>();
        lc->jobId = "111";
        auto clientsMgr = std::make_shared<YR::Libruntime::ClientsManager>();
        auto metricsAdaptor = std::make_shared<YR::Libruntime::MetricsAdaptor>();
        auto sec = std::make_shared<YR::Libruntime::Security>();
        auto socketClient = std::make_shared<YR::Libruntime::DomainSocketClient>("/home/snuser/socket/runtime.sock");
        lr = std::make_shared<YR::Libruntime::MockLibruntime>(lc, clientsMgr, metricsAdaptor, sec, socketClient);
        YR::Libruntime::LibruntimeManager::Instance().SetLibRuntime(lr);
    }

    char *GetStr(std::string str)
    {
        char *cStr = (char *)malloc(str.size() + 1);
        (void)memcpy_s(cStr, str.size(), str.data(), str.size());
        cStr[str.size()] = 0;
        tmpStrs.push_back(cStr);
        return cStr;
    }

    void TearDown() override
    {
        lr.reset();
        YR::Libruntime::LibruntimeManager::Instance().Finalize();
        for (auto ptr : tmpStrs) {
            if (ptr != nullptr) {
                free(ptr);
                ptr = nullptr;
            }
        }
        tmpStrs.clear();
    }

    std::shared_ptr<YR::Libruntime::MockLibruntime> lr;
    std::vector<char *> tmpStrs;
};

TEST_F(CLibruntimeTest, CCreateStateStoreTest)
{
    std::shared_ptr<YR::Libruntime::StateStore> stateStore;
    stateStore = std::make_shared<YR::Libruntime::DSCacheStateStore>();
    EXPECT_CALL(*lr.get(), CreateStateStore(_, _))
        .WillOnce(DoAll(SetArgReferee<1>(stateStore), Return(YR::Libruntime::ErrorInfo())));
    CConnectArguments arguments{"127.0.0.1", 0,       100, nullptr, 0, nullptr, 0, nullptr, 0, nullptr, 0, nullptr,
                                0,           nullptr, 0,   nullptr, 0, nullptr, 0, nullptr, 0, nullptr, 0, 0};
    CStateStorePtr stateStore2 = nullptr;
    auto cErr = CCreateStateStore(&arguments, &stateStore2);
    ASSERT_TRUE(cErr.code == 0);
    ASSERT_TRUE(stateStore2 != nullptr);
    CDestroyStateStore(stateStore2);
    SafeFreeCErr(cErr);
}

TEST_F(CLibruntimeTest, CSetTraceIdTest)
{
    std::string traceId = "traceId";
    EXPECT_CALL(*lr.get(), SetTraceId(_)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    auto cErr = CSetTraceId(traceId.c_str(), traceId.size());
    ASSERT_TRUE(cErr.code == 0);
    SafeFreeCErr(cErr);
}

TEST_F(CLibruntimeTest, CGenerateKeyTest)
{
    char *key = nullptr;
    int cKeyLen = 0;
    CStateStorePtr nullStateStorePtr = nullptr;
    auto cErr = CGenerateKey(nullStateStorePtr, &key, &cKeyLen);
    ASSERT_TRUE(cErr.code == 3003);
    ASSERT_TRUE(std::string(cErr.message) == "the state store is empty");
    SafeFreeCErr(cErr);

    auto stateStore = YR::Libruntime::DSCacheStateStore();
    auto stateStorePtr = reinterpret_cast<CStateStorePtr>(&stateStore);
    EXPECT_CALL(*lr.get(), GenerateKeyByStateStore(_, _))
        .WillOnce(DoAll(SetArgReferee<1>("genKey"), Return(YR::Libruntime::ErrorInfo())));
    cErr = CGenerateKey(stateStorePtr, &key, &cKeyLen);
    ASSERT_TRUE(cErr.code == 0);
    ASSERT_TRUE(strcmp(key, "genKey") == 0);
    ASSERT_TRUE(cKeyLen == strlen("genKey"));
    SafeFreeCErr(cErr);
    free(key);

    YR::Libruntime::LibruntimeManager::Instance().SetLibRuntime(nullptr);
    cErr = CGenerateKey(stateStorePtr, &key, &cKeyLen);
    ASSERT_TRUE(cErr.code == 9000);
    ASSERT_TRUE(std::string(cErr.message) == "libRuntime empty");
}

TEST_F(CLibruntimeTest, CSetByStateStoreTest)
{
    std::string rightKey = "rightKey";
    std::string value = "value";
    CBuffer buffer{static_cast<void *>(const_cast<char *>(value.c_str())), static_cast<int64_t>(value.size())};
    CSetParam param;

    CStateStorePtr nullStateStorePtr = nullptr;
    auto cErr = CSetByStateStore(nullStateStorePtr, const_cast<char *>(rightKey.c_str()), buffer, param);
    ASSERT_EQ(cErr.code, 3003);
    ASSERT_TRUE(std::string(cErr.message) == "the state store is empty");
    SafeFreeCErr(cErr);

    auto stateStore = YR::Libruntime::DSCacheStateStore();
    auto stateStorePtr = reinterpret_cast<CStateStorePtr>(&stateStore);
    EXPECT_CALL(*lr.get(), SetByStateStore(_, _, _, _)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    cErr = CSetByStateStore(stateStorePtr, const_cast<char *>(rightKey.c_str()), buffer, param);
    ASSERT_TRUE(cErr.code == 0);
    SafeFreeCErr(cErr);

    auto err = YR::Libruntime::ErrorInfo();
    err.SetErrorCode(YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED);
    err.SetDsStatusCode(datasystem::StatusCode::K_RUNTIME_ERROR);
    EXPECT_CALL(*lr.get(), SetByStateStore(_, _, _, _)).WillOnce(Return(err));
    std::string wrongKey = "wrongKey";
    cErr = CSetByStateStore(stateStorePtr, const_cast<char *>(wrongKey.c_str()), buffer, param);
    ASSERT_TRUE(cErr.code != 0);
    ASSERT_TRUE(cErr.dsStatusCode == datasystem::StatusCode::K_RUNTIME_ERROR);
    SafeFreeCErr(cErr);
}

TEST_F(CLibruntimeTest, CSetValueByStateStoreTest)
{
    std::string value = "value";
    CSetParam param;
    CBuffer buffer{static_cast<void *>(const_cast<char *>(value.c_str())), static_cast<int64_t>(value.size())};
    char *key = nullptr;
    int keyLen = 0;

    CStateStorePtr nullStateStorePtr = nullptr;
    auto cErr = CSetValueByStateStore(nullStateStorePtr, buffer, param, &key, &keyLen);
    ASSERT_TRUE(cErr.code == 3003);
    ASSERT_TRUE(std::string(cErr.message) == "the state store is empty");
    SafeFreeCErr(cErr);

    auto stateStore = YR::Libruntime::DSCacheStateStore();
    auto stateStorePtr = reinterpret_cast<CStateStorePtr>(&stateStore);
    EXPECT_CALL(*lr.get(), SetValueByStateStore(_, _, _, _))
        .WillOnce(DoAll(SetArgReferee<3>("returnKey"), Return(YR::Libruntime::ErrorInfo())));
    cErr = CSetValueByStateStore(stateStorePtr, buffer, param, &key, &keyLen);
    ASSERT_TRUE(cErr.code == 0);
    ASSERT_TRUE(strcmp(key, "returnKey") == 0);
    ASSERT_TRUE(keyLen == strlen("returnKey"));
    free(key);
    SafeFreeCErr(cErr);
}

TEST_F(CLibruntimeTest, GetByStateStoreTest)
{
    std::string rightKey = "rightKey";
    CBuffer buffer;

    CStateStorePtr nullStateStorePtr = nullptr;
    auto cErr = CGetByStateStore(nullStateStorePtr, const_cast<char *>(rightKey.c_str()), &buffer, 0);
    ASSERT_TRUE(cErr.code == 3003);
    ASSERT_TRUE(std::string(cErr.message) == "the state store is empty");
    SafeFreeCErr(cErr);

    auto stateStore = YR::Libruntime::DSCacheStateStore();
    auto stateStorePtr = reinterpret_cast<CStateStorePtr>(&stateStore);
    EXPECT_CALL(*lr.get(), GetByStateStore(_, _, _))
        .WillOnce(Return(std::make_pair<std::shared_ptr<YR::Libruntime::Buffer>, YR::Libruntime::ErrorInfo>(
            std::make_shared<YR::Libruntime::NativeBuffer>(1), YR::Libruntime::ErrorInfo())));
    cErr = CGetByStateStore(stateStorePtr, const_cast<char *>(rightKey.c_str()), &buffer, 0);
    ASSERT_TRUE(cErr.code == 0);
    SafeFreeCErr(cErr);
    free(buffer.buffer);

    std::string wrongKey = "wrongKey";
    CBuffer bufferTwo{nullptr, 0};
    auto err = YR::Libruntime::ErrorInfo();
    err.SetErrorCode(YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED);
    err.SetDsStatusCode(datasystem::StatusCode::K_OUT_OF_MEMORY);
    EXPECT_CALL(*lr.get(), GetByStateStore(_, _, _))
        .WillOnce(Return(std::make_pair<std::shared_ptr<YR::Libruntime::Buffer>, YR::Libruntime::ErrorInfo>(
            std::make_shared<YR::Libruntime::NativeBuffer>(0), std::move(err))));
    cErr = CGetByStateStore(stateStorePtr, const_cast<char *>(wrongKey.c_str()), &bufferTwo, 0);
    ASSERT_TRUE(cErr.code != 0);
    ASSERT_TRUE(cErr.dsStatusCode == datasystem::StatusCode::K_OUT_OF_MEMORY);
    SafeFreeCErr(cErr);
    free(bufferTwo.buffer);
}

TEST_F(CLibruntimeTest, GetArrayByStateStoreTest)
{
    char *keys[2];
    keys[0] = GetStr("key1");
    keys[1] = GetStr("key2");
    CBuffer buffer[2];
    for (size_t i = 0; i < 2; i++) {
        buffer[i].buffer = nullptr;
        buffer[i].size_buffer = 0;
    }

    CStateStorePtr nullStateStorePtr = nullptr;
    auto cErr = CGetArrayByStateStore(nullStateStorePtr, keys, 2, buffer, 0);
    ASSERT_TRUE(cErr.code == 3003);
    ASSERT_TRUE(std::string(cErr.message) == "the state store is empty");
    SafeFreeCErr(cErr);

    auto stateStore = YR::Libruntime::DSCacheStateStore();
    auto stateStorePtr = reinterpret_cast<CStateStorePtr>(&stateStore);
    auto err = YR::Libruntime::ErrorInfo();
    err.SetErrorCode(YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED);
    err.SetDsStatusCode(datasystem::StatusCode::K_OUT_OF_MEMORY);
    EXPECT_CALL(*lr.get(), GetArrayByStateStore(_, _, _, _))
        .WillOnce(
            Return(std::make_pair<std::vector<std::shared_ptr<YR::Libruntime::Buffer>>, YR::Libruntime::ErrorInfo>(
                {}, std::move(err))));
    cErr = CGetArrayByStateStore(stateStorePtr, keys, 2, buffer, 0);
    ASSERT_TRUE(cErr.code != 0);
    ASSERT_EQ(cErr.dsStatusCode, datasystem::StatusCode::K_OUT_OF_MEMORY);
    SafeFreeCErr(cErr);
    for (size_t i = 0; i < 2; i++) {
        free(buffer[i].buffer);
    }
}

TEST_F(CLibruntimeTest, QuerySizeByStateStoreTest)
{
    char *fake[1];
    fake[0] = GetStr("query_1");
    uint64_t fakeSizes[1];

    CStateStorePtr nullStateStorePtr = nullptr;
    auto cErr = CQuerySizeByStateStore(nullStateStorePtr, fake, 1, fakeSizes);
    ASSERT_TRUE(cErr.code == 3003);
    ASSERT_TRUE(std::string(cErr.message) == "the state store is empty");
    SafeFreeCErr(cErr);

    auto stateStore = YR::Libruntime::DSCacheStateStore();
    auto stateStorePtr = reinterpret_cast<CStateStorePtr>(&stateStore);
    EXPECT_CALL(*lr.get(), QuerySizeByStateStore(_, _, _))
        .WillOnce([=](std::shared_ptr<YR::Libruntime::StateStore>, const std::vector<std::string> &,
                      std::vector<uint64_t> &param) {
            param = {10};
            return YR::Libruntime::ErrorInfo();
        });
    cErr = CQuerySizeByStateStore(stateStorePtr, fake, 1, fakeSizes);
    ASSERT_EQ(0, cErr.code);
    ASSERT_EQ(fakeSizes[0], 10);
}

TEST_F(CLibruntimeTest, DelByStateStoreTest)
{
    CStateStorePtr nullStateStorePtr = nullptr;
    auto stateStore = YR::Libruntime::DSCacheStateStore();
    auto stateStorePtr = reinterpret_cast<CStateStorePtr>(&stateStore);
    EXPECT_CALL(*lr.get(), DelByStateStore(_, _)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    std::string rightKey = "rightKey";

    auto cErr = CDelByStateStore(nullStateStorePtr, const_cast<char *>(rightKey.c_str()));
    ASSERT_TRUE(cErr.code == 3003);
    ASSERT_TRUE(std::string(cErr.message) == "the state store is empty");
    SafeFreeCErr(cErr);

    cErr = CDelByStateStore(stateStorePtr, const_cast<char *>(rightKey.c_str()));
    ASSERT_TRUE(cErr.code == 0);
    SafeFreeCErr(cErr);

    auto err = YR::Libruntime::ErrorInfo();
    err.SetErrorCode(YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED);
    err.SetDsStatusCode(datasystem::StatusCode::K_RUNTIME_ERROR);
    EXPECT_CALL(*lr.get(), DelByStateStore(_, _)).WillOnce(Return(err));
    std::string wrongKey = "wrongKey";
    cErr = CDelByStateStore(stateStorePtr, const_cast<char *>(wrongKey.c_str()));
    ASSERT_TRUE(cErr.code != 0);
    ASSERT_EQ(cErr.dsStatusCode, datasystem::StatusCode::K_RUNTIME_ERROR);
    SafeFreeCErr(cErr);
}

TEST_F(CLibruntimeTest, DelArrayByStateStoreTest)
{
    char *keys[2];
    keys[0] = GetStr("wrongKey");
    keys[1] = GetStr("rightKey");
    char **failedKeys = nullptr;
    int failedKeysLen = 0;

    CStateStorePtr nullStateStorePtr = nullptr;
    auto cErr = CDelArrayByStateStore(nullStateStorePtr, keys, 2, &failedKeys, &failedKeysLen);
    ASSERT_TRUE(cErr.code == 3003);
    ASSERT_TRUE(std::string(cErr.message) == "the state store is empty");
    SafeFreeCErr(cErr);

    auto stateStore = YR::Libruntime::DSCacheStateStore();
    auto stateStorePtr = reinterpret_cast<CStateStorePtr>(&stateStore);
    auto err = YR::Libruntime::ErrorInfo();
    err.SetErrorCode(YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED);
    err.SetDsStatusCode(datasystem::StatusCode::K_RUNTIME_ERROR);
    EXPECT_CALL(*lr.get(), DelArrayByStateStore(_, _))
        .WillOnce(
            Return(std::make_pair<std::vector<std::string>, YR::Libruntime::ErrorInfo>({"wrongKey"}, std::move(err))));
    cErr = CDelArrayByStateStore(stateStorePtr, keys, 2, &failedKeys, &failedKeysLen);
    ASSERT_TRUE(cErr.code != 0);
    ASSERT_TRUE(cErr.dsStatusCode == datasystem::StatusCode::K_RUNTIME_ERROR);
    ASSERT_TRUE(failedKeysLen == 1);
    ASSERT_TRUE(strcmp(failedKeys[0], "wrongKey") == 0);
    SafeFreeCErr(cErr);
    freeCStrings(failedKeys, failedKeysLen);
}

TEST_F(CLibruntimeTest, CGetCredentialTest)
{
    YR::Libruntime::Credential credential = {.ak = "ak", .sk = "sk", .dk = "dk"};
    EXPECT_CALL(*lr.get(), GetCredential()).WillOnce(Return(credential));
    auto cCredential = CGetCredential();
    ASSERT_TRUE(std::string(cCredential.ak) == "ak");
    ASSERT_TRUE(std::string(cCredential.sk) == "sk");
    ASSERT_TRUE(std::string(cCredential.dk) == "dk");
    free(cCredential.ak);
    free(cCredential.sk);
    free(cCredential.dk);
}

TEST_F(CLibruntimeTest, CInitTest)
{
    CLibruntimeConfig config{
        "127.0.0.1:11111",
        "127.0.0.1:11112",
        "127.0.0.1:11113",
        "127.0.0.1:11114",
        "jobId",
        "runtimeId",
        "instanceId",
        "functionName",
        "DEBUG",
        "./",
        CApiType::ACTOR,
        1,
        0,
        0,
        "privateKeyPath",
        "certificateFilePath",
        "verifyFilePath",
        "privateKeyPaaswd",
        GetStr("functionId"),
        "systemAuthAccessKey",
        "systemAuthSecretKey",
        2,
        "systemAuthDataKey",
        2,
        "EncryptPrivateKeyPasswd",
        "PrimaryKeyStoreFile",
        "StandbyKeyStoreFile",
        0,
        "RuntimePublicKeyContext",
        "RuntimePrivateKeyContext",
        "DsPublicKeyContext",
    };
    auto cErr = CInit(&config);
    ASSERT_TRUE(cErr.code == 0);
    SafeFreeCErr(cErr);
}

TEST_F(CLibruntimeTest, CWaitTest)
{
    CWaitResult result{nullptr, 0, nullptr, 0, nullptr, 0};
    char *objectIds[3];
    objectIds[0] = GetStr("obj1");
    objectIds[1] = GetStr("obj2");
    objectIds[2] = GetStr("obj3");
    char *readyIds[1];
    readyIds[0] = GetStr("obj1");
    char *unreadyIds[1];
    unreadyIds[0] = GetStr("obj2");
    std::vector<YR::Libruntime::StackTraceInfo> stackTraceInfos;
    YR::Libruntime::StackTraceInfo stackTraceInfo("type", "err msg");
    stackTraceInfos.push_back(stackTraceInfo);
    auto err = YR::Libruntime::ErrorInfo();
    err.SetErrorCode(YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED);
    err.SetDsStatusCode(datasystem::StatusCode::K_RUNTIME_ERROR);
    err.SetErrorMsg("failed to wait");
    err.SetStackTraceInfos(stackTraceInfos);
    auto ret = std::make_shared<YR::InternalWaitResult>();
    ret->readyIds.push_back(readyIds[0]);
    ret->unreadyIds.push_back(unreadyIds[0]);
    ret->exceptionIds["obj3"] = err;
    EXPECT_CALL(*lr.get(), Wait(_, _, _)).WillOnce(Return(ret));
    ASSERT_NO_THROW(CWait(objectIds, 3, 3, 1, &result));
    ASSERT_EQ(result.size_readyIds, 1);
    ASSERT_EQ(result.size_unreadyIds, 1);
    ASSERT_EQ(result.size_errorIds, 1);
    freeCStrings(result.unreadyIds, result.size_unreadyIds);
    freeCStrings(result.readyIds, result.size_readyIds);
    freeCErrorIds(result.errorIds, result.size_errorIds);
}

TEST_F(CLibruntimeTest, CCreateStreamProducerTest)
{
    CProducerConfig config{0, 0, 0};
    config.traceId = GetStr("trace_id");
    std::string streamName = "stream_001";
    Producer_p producer = nullptr;
    EXPECT_CALL(*lr.get(), CreateStreamProducer(_, _, _)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    auto cErr = CCreateStreamProducer(const_cast<char *>(streamName.c_str()), &config, &producer);
    ASSERT_EQ(cErr.code, 0);
    ASSERT_TRUE(producer != nullptr);
    SafeFreeCErr(cErr);
    cErr = CProducerClose(producer);
    SafeFreeCErr(cErr);
}

TEST_F(CLibruntimeTest, CCreateInstanceTest)
{
    EXPECT_CALL(*lr.get(), CreateInstance(_, _, _))
        .WillOnce(
            Return(std::pair<YR::Libruntime::ErrorInfo, std::string>(YR::Libruntime::ErrorInfo(), "instance_id")));
    CFunctionMeta meta{GetStr("app_name"),
                       GetStr("module_name"),
                       GetStr("func_name"),
                       GetStr("class_name"),
                       1,
                       GetStr("code_id"),
                       GetStr("signature"),
                       GetStr("pool_label"),
                       CApiType::ACTOR,
                       GetStr("function_id"),
                       '1',
                       GetStr("name"),
                       '1',
                       GetStr("namespace")};

    char *fake[2];
    fake[0] = GetStr("obj_1");
    fake[1] = GetStr("obj_2");
    CInvokeArg arg{(void *)GetStr("buf"), 3, '0', GetStr("obj_id"), GetStr("tenant_id"), fake, 2};
    CCustomResource res{GetStr("name"), 1.0};
    CCustomExtension extension{GetStr("key"), GetStr("value")};
    CCreateOpt opt{GetStr("key"), GetStr("value")};
    CLabelOperator labelOperator{CLabelOpType::EXISTS, GetStr("label_key"), fake, 2};
    CAffinity affinity{CAffinityKind::INSTANCE, CAffinityType::PREFERRED, '1', '1', &labelOperator, 1};
    CInvokeOptions option{500,
                          500,
                          &res,
                          1,
                          &extension,
                          1,
                          &opt,
                          1,
                          fake,
                          2,
                          &affinity,
                          0,
                          0,
                          1,
                          fake,
                          2,
                          GetStr("scheduler_id"),
                          fake,
                          2,
                          GetStr("trace_id"),
                          1};
    char *instanceId;
    int invokeArgsSize = 1;
    auto cErr = CCreateInstance(&meta, &arg, invokeArgsSize, &option, &instanceId);
    ASSERT_EQ(0, cErr.code);
    free(instanceId);
}

TEST_F(CLibruntimeTest, CInvokeByInstanceIdTest)
{
    EXPECT_CALL(*lr.get(), InvokeByInstanceId(_, _, _, _, _)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    CFunctionMeta meta{GetStr("app_name"),
                       GetStr("module_name"),
                       GetStr("func_name"),
                       GetStr("class_name"),
                       1,
                       GetStr("code_id"),
                       GetStr("signature"),
                       GetStr("pool_label"),
                       CApiType::ACTOR,
                       GetStr("function_id"),
                       '1',
                       GetStr("name"),
                       '1',
                       GetStr("namespace")};

    char *fake[2];
    fake[0] = GetStr("obj_1");
    fake[1] = GetStr("obj_2");
    CInvokeArg arg{(void *)GetStr("buf"), 3, '0', GetStr("obj_id"), GetStr("tenant_id"), fake, 2};
    CCustomResource res{GetStr("name"), 1.0};
    CCustomExtension extension{GetStr("key"), GetStr("value")};
    CCreateOpt opt{GetStr("key"), GetStr("value")};
    CLabelOperator labelOperator{CLabelOpType::EXISTS, GetStr("label_key"), fake, 2};
    CAffinity affinity{CAffinityKind::INSTANCE, CAffinityType::PREFERRED, '1', '1', &labelOperator, 1};
    CInvokeOptions option{500,
                          500,
                          &res,
                          1,
                          &extension,
                          1,
                          &opt,
                          1,
                          fake,
                          2,
                          &affinity,
                          0,
                          0,
                          1,
                          fake,
                          2,
                          GetStr("scheduler_id"),
                          fake,
                          2,
                          GetStr("trace_id"),
                          1};
    char *returnObjId;
    auto cErr = CInvokeByInstanceId(&meta, GetStr("instance_id"), &arg, 1, &option, &returnObjId);
    ASSERT_EQ(0, cErr.code);
    free(returnObjId);
}

TEST_F(CLibruntimeTest, CInvokeByFunctionNameTest)
{
    EXPECT_CALL(*lr.get(), InvokeByFunctionName(_, _, _, _)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    CFunctionMeta meta{GetStr("app_name"),
                       GetStr("module_name"),
                       GetStr("func_name"),
                       GetStr("class_name"),
                       1,
                       GetStr("code_id"),
                       GetStr("signature"),
                       GetStr("pool_label"),
                       CApiType::ACTOR,
                       GetStr("function_id"),
                       '1',
                       GetStr("name"),
                       '1',
                       GetStr("namespace")};

    char *fake[2];
    fake[0] = GetStr("obj_1");
    fake[1] = GetStr("obj_2");
    CInvokeArg arg{(void *)GetStr("buf"), 3, '0', GetStr("obj_id"), GetStr("tenant_id"), fake, 2};
    CCustomResource res{GetStr("name"), 1.0};
    CCustomExtension extension{GetStr("key"), GetStr("value")};
    CCreateOpt opt{GetStr("key"), GetStr("value")};
    CLabelOperator labelOperator{CLabelOpType::EXISTS, GetStr("label_key"), fake, 2};
    CAffinity affinity{CAffinityKind::INSTANCE, CAffinityType::PREFERRED, '1', '1', &labelOperator, 1};
    CInvokeOptions option{500,
                          500,
                          &res,
                          1,
                          &extension,
                          1,
                          &opt,
                          1,
                          fake,
                          2,
                          &affinity,
                          0,
                          0,
                          1,
                          fake,
                          2,
                          GetStr("scheduler_id"),
                          fake,
                          2,
                          GetStr("trace_id"),
                          1};
    char *returnObjId;
    auto cErr = CInvokeByFunctionName(&meta, &arg, 1, &option, &returnObjId);
    ASSERT_EQ(0, cErr.code);
    free(returnObjId);
}

TEST_F(CLibruntimeTest, CIncreaseReferenceTest)
{
    EXPECT_CALL(*lr.get(), IncreaseReference(_, Matcher<const std::string &>(_)))
        .WillOnce(Return(std::pair<YR::Libruntime::ErrorInfo, std::vector<std::string>>(YR::Libruntime::ErrorInfo(),
                                                                                        {"failed_id1"})));
    char *fake[2];
    fake[0] = GetStr("obj_1");
    fake[1] = GetStr("obj_2");
    char **failedIds;
    int failedIdSize;
    auto cErr = CIncreaseReferenceCommon(fake, 2, GetStr("remote_id"), &failedIds, &failedIdSize, 0);
    ASSERT_EQ(0, cErr.code);
    ASSERT_EQ(std::string(failedIds[0]), "failed_id1");
    free(failedIds[0]);
    free(failedIds);
}

TEST_F(CLibruntimeTest, CDecreaseReferenceTest)
{
    EXPECT_CALL(*lr.get(), DecreaseReference(_, _))
        .WillOnce(Return(std::pair<YR::Libruntime::ErrorInfo, std::vector<std::string>>(YR::Libruntime::ErrorInfo(),
                                                                                        {"failed_id1"})));
    char *fake[2];
    fake[0] = GetStr("obj_1");
    fake[1] = GetStr("obj_2");
    char **failedIds;
    int failedIdSize;
    auto cErr = CDecreaseReferenceCommon(fake, 2, GetStr("remote_id"), &failedIds, &failedIdSize, 0);
    ASSERT_EQ(0, cErr.code);
    ASSERT_EQ(std::string(failedIds[0]), "failed_id1");
    free(failedIds[0]);
    free(failedIds);
}

TEST_F(CLibruntimeTest, CKVWriteTest)
{
    EXPECT_CALL(*lr.get(), KVWrite(_, _, _)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    CBuffer data{(void *)"123", 3, nullptr};
    CSetParam param;
    auto cErr = CKVWrite(GetStr("key"), data, param);
    ASSERT_EQ(0, cErr.code);
}

TEST_F(CLibruntimeTest, CKVMSetTxTest)
{
    EXPECT_CALL(*lr.get(), KVMSetTx(_, _, _)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    char *fake[1];
    fake[0] = GetStr("obj_1");
    CBuffer data{(void *)"123", 3, nullptr};
    CMSetParam param;
    auto cErr = CKVMSetTx(fake, 1, &data, param);
    ASSERT_EQ(0, cErr.code);
}

TEST_F(CLibruntimeTest, CKVReadTest)
{
    auto returnBuffer = std::make_shared<YR::Libruntime::NativeBuffer>(3);
    returnBuffer->MemoryCopy("123", 3);
    EXPECT_CALL(*lr.get(), KVRead(_, _))
        .WillOnce(Return(std::pair<std::shared_ptr<YR::Libruntime::Buffer>, YR::Libruntime::ErrorInfo>(
            returnBuffer, YR::Libruntime::ErrorInfo())));
    CBuffer data;
    auto cErr = CKVRead(GetStr("key"), 1000, &data);
    ASSERT_EQ(0, cErr.code);
    free(data.buffer);
}

TEST_F(CLibruntimeTest, CKVMultiReadTest)
{
    auto returnBuffer = std::make_shared<YR::Libruntime::NativeBuffer>(3);
    returnBuffer->MemoryCopy("123", 3);
    EXPECT_CALL(*lr.get(), KVRead(_, _, _))
        .WillOnce(Return(std::pair<std::vector<std::shared_ptr<YR::Libruntime::Buffer>>, YR::Libruntime::ErrorInfo>(
            {returnBuffer}, YR::Libruntime::ErrorInfo())));
    CBuffer data;
    char *fake[1];
    fake[0] = GetStr("obj_1");
    auto cErr = CKVMultiRead(fake, 1, 1000, 1, &data);
    ASSERT_EQ(0, cErr.code);
    free(data.buffer);
}

TEST_F(CLibruntimeTest, CGetTest)
{
    auto returnDataObj = std::make_shared<YR::Libruntime::DataObject>(0, 3);
    returnDataObj->data->MemoryCopy("123", 3);
    EXPECT_CALL(*lr.get(), Get(_, _, _))
        .WillOnce(Return(std::pair<YR::Libruntime::ErrorInfo, std::vector<std::shared_ptr<YR::Libruntime::DataObject>>>(
            YR::Libruntime::ErrorInfo(), {returnDataObj})));
    CBuffer data;
    auto cErr = CGet(GetStr("obj_id"), 1000, &data);
    ASSERT_EQ(0, cErr.code);
    free(data.buffer);
}

TEST_F(CLibruntimeTest, CKillTest)
{
    EXPECT_CALL(*lr.get(), Kill(_, _, _)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    CBuffer data;
    auto cErr = CKill(GetStr("instance_id"), 1, data);
    ASSERT_EQ(0, cErr.code);
}

TEST_F(CLibruntimeTest, CExitTest)
{
    EXPECT_CALL(*lr.get(), Exit(_, _)).WillOnce(Return());
    EXPECT_NO_THROW(CExit(0, ""));
}

TEST_F(CLibruntimeTest, CPutTest)
{
    EXPECT_CALL(*lr.get(), Put(_, _, _, _)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    CBuffer data{(void *)"123", 3, nullptr};
    char *fake[2];
    fake[0] = GetStr("obj_1");
    fake[1] = GetStr("obj_2");
    CCreateParam param;
    auto cErr = CPutCommon(GetStr("obj_id"), data, fake, 2, 0, param);
    ASSERT_EQ(0, cErr.code);
}

TEST_F(CLibruntimeTest, CGetMultiTest)
{
    auto returnDataObj = std::make_shared<YR::Libruntime::DataObject>(0, 3);
    returnDataObj->data->MemoryCopy("123", 3);
    EXPECT_CALL(*lr.get(), Get(_, _, _))
        .WillOnce(Return(std::pair<YR::Libruntime::ErrorInfo, std::vector<std::shared_ptr<YR::Libruntime::DataObject>>>(
            YR::Libruntime::ErrorInfo(), {returnDataObj})));
    CBuffer data;
    char *fake[1];
    fake[0] = GetStr("obj_1");
    auto cErr = CGetMultiCommon(fake, 1, 1000, false, &data, 0);
    ASSERT_EQ(0, cErr.code);
    free(data.buffer);
}

TEST_F(CLibruntimeTest, CKVDelTest)
{
    EXPECT_CALL(*lr.get(), KVDel(Matcher<const std::string &>(_))).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    auto cErr = CKVDel(GetStr("obj_id"));
    ASSERT_EQ(0, cErr.code);
}

TEST_F(CLibruntimeTest, CKVMultiDelTest)
{
    EXPECT_CALL(*lr.get(), KVDel(Matcher<const std::vector<std::string> &>(_)))
        .WillOnce(Return(
            std::pair<std::vector<std::string>, YR::Libruntime::ErrorInfo>({"key"}, YR::Libruntime::ErrorInfo())));
    char *fake[1];
    fake[0] = GetStr("obj_1");
    char **failedKeys;
    int failedKeysSize;
    auto cErr = CKMultiVDel(fake, 1, &failedKeys, &failedKeysSize);
    ASSERT_EQ(0, cErr.code);
    freeCStrings(failedKeys, failedKeysSize);
}

TEST_F(CLibruntimeTest, CCreateStreamConsumerTest)
{
    CSubscriptionConfig config{GetStr("name"), CSubscriptionType::KEY_PARTITIONS};
    config.traceId = GetStr("trace_id");
    std::string streamName = "stream_001";
    Consumer_p consumer = nullptr;
    EXPECT_CALL(*lr.get(), CreateStreamConsumer(_, _, _, _)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    auto cErr = CCreateStreamConsumer(const_cast<char *>(streamName.c_str()), &config, &consumer);
    ASSERT_EQ(cErr.code, 0);
    ASSERT_TRUE(consumer != nullptr);
    SafeFreeCErr(cErr);
    cErr = CConsumerClose(consumer);
    SafeFreeCErr(cErr);
}

TEST_F(CLibruntimeTest, CDeleteStreamTest)
{
    EXPECT_CALL(*lr.get(), DeleteStream(_)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    auto cErr = CDeleteStream(GetStr("stream_001"));
    ASSERT_EQ(cErr.code, 0);
    SafeFreeCErr(cErr);
}

TEST_F(CLibruntimeTest, CQueryGlobalProducersNumTest)
{
    EXPECT_CALL(*lr.get(), QueryGlobalProducersNum(_, _)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    uint64_t num;
    auto cErr = CQueryGlobalProducersNum(GetStr("stream_001"), &num);
    ASSERT_EQ(cErr.code, 0);
    SafeFreeCErr(cErr);
}

TEST_F(CLibruntimeTest, CQueryGlobalConsumersNumTest)
{
    EXPECT_CALL(*lr.get(), QueryGlobalConsumersNum(_, _)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    uint64_t num;
    auto cErr = CQueryGlobalConsumersNum(GetStr("stream_001"), &num);
    ASSERT_EQ(cErr.code, 0);
    SafeFreeCErr(cErr);
}

TEST_F(CLibruntimeTest, CProducerTest)
{
    std::shared_ptr<FakeStreamProducer> fkProducer = std::make_shared<FakeStreamProducer>();
    std::unique_ptr<std::shared_ptr<YR::Libruntime::StreamProducer>> pProducer =
        std::make_unique<std::shared_ptr<YR::Libruntime::StreamProducer>>(std::move(fkProducer));
    Producer_p producer = reinterpret_cast<void *>(pProducer.release());
    uint8_t array[10] = {0};
    uint8_t *ptr = array;
    uint64_t size = 10;
    uint64_t id = 111;
    auto cErr = CProducerSend(producer, ptr, size, id);
    ASSERT_EQ(cErr.code, 0);
    SafeFreeCErr(cErr);

    cErr = CProducerSendWithTimeout(producer, ptr, size, id, 1000);
    ASSERT_EQ(cErr.code, 0);
    SafeFreeCErr(cErr);

    ASSERT_EQ(cErr.code, 0);
    SafeFreeCErr(cErr);

    cErr = CProducerClose(producer);
    ASSERT_EQ(cErr.code, 0);
    SafeFreeCErr(cErr);
}

TEST_F(CLibruntimeTest, CConsumerTest)
{
    std::shared_ptr<FakeStreamConsumer> fkConsumer = std::make_shared<FakeStreamConsumer>();
    std::unique_ptr<std::shared_ptr<YR::Libruntime::StreamConsumer>> pConsumer =
        std::make_unique<std::shared_ptr<YR::Libruntime::StreamConsumer>>(fkConsumer);
    Consumer_p consumer = reinterpret_cast<void *>(pConsumer.release());

    uint64_t elementId = 111;
    auto cErr = CConsumerAck(consumer, elementId);
    ASSERT_EQ(cErr.code, 0);
    SafeFreeCErr(cErr);

    cErr = CConsumerClose(consumer);
    ASSERT_EQ(cErr.code, 0);
    SafeFreeCErr(cErr);
}

TEST_F(CLibruntimeTest, CAllocReturnObjectTest)
{
    EXPECT_CALL(*lr.get(), AllocReturnObject(Matcher<YR::Libruntime::DataObject *>(_), _, _, _, _))
        .WillOnce(Return(YR::Libruntime::ErrorInfo()));
    CDataObject obj;
    auto dataObj = std::make_shared<YR::Libruntime::DataObject>();
    dataObj->data = std::make_shared<YR::Libruntime::NativeBuffer>(1);
    obj.selfSharedPtr = dataObj.get();
    char *fake[1];
    fake[0] = GetStr("obj_1");
    uint64_t totalNativeBufferSize;
    auto cErr = CAllocReturnObject(&obj, 1, fake, 1, &totalNativeBufferSize);
    ASSERT_EQ(cErr.code, 0);
    SafeFreeCErr(cErr);
}

TEST_F(CLibruntimeTest, CSetReturnObjectTest)
{
    CDataObject obj;
    auto dataObj = std::make_shared<YR::Libruntime::DataObject>();
    obj.selfSharedPtr = dataObj.get();
    EXPECT_NO_THROW(CSetReturnObject(&obj, 10));
    auto cBuffer = obj.buffer;
    auto cErr = CWriterLatch(&cBuffer);
    ASSERT_EQ(cErr.code, 0);
    SafeFreeCErr(cErr);
    char *cSrc = GetStr("abc");
    cErr = CMemoryCopy(&cBuffer, (void *)cSrc, 3);
    ASSERT_EQ(cErr.code, 0);
    SafeFreeCErr(cErr);
    cErr = CSeal(&cBuffer);
    ASSERT_EQ(cErr.code, 0);
    SafeFreeCErr(cErr);
    cErr = CWriterUnlatch(&cBuffer);
    ASSERT_EQ(cErr.code, 0);
    SafeFreeCErr(cErr);
}

TEST_F(CLibruntimeTest, CAcquireInstanceTest)
{
    YR::Libruntime::InstanceAllocation insAlloc{"functionId", "funcSig", "instanceId", "leaseId", 0};
    EXPECT_CALL(*lr.get(), AcquireInstance(_, _, _))
        .WillOnce(Return(std::make_pair<YR::Libruntime::InstanceAllocation, YR::Libruntime::ErrorInfo>(
            std::move(insAlloc), YR::Libruntime::ErrorInfo())));
    char *stateId = GetStr("aaa");

    CFunctionMeta meta{GetStr("app_name"),
                       GetStr("module_name"),
                       GetStr("func_name"),
                       GetStr("class_name"),
                       1,
                       GetStr("code_id"),
                       GetStr("signature"),
                       GetStr("pool_label"),
                       CApiType::ACTOR,
                       GetStr("function_id"),
                       '1',
                       GetStr("name"),
                       '1',
                       GetStr("namespace")};

    char *fake[2];
    fake[0] = GetStr("obj_1");
    fake[1] = GetStr("obj_2");
    CCustomResource res{GetStr("name"), 1.0};
    CCustomExtension extension{GetStr("key"), GetStr("value")};
    CCreateOpt opt{GetStr("key"), GetStr("value")};
    CLabelOperator labelOperator{CLabelOpType::EXISTS, GetStr("label_key"), fake, 2};
    CAffinity affinity{CAffinityKind::INSTANCE, CAffinityType::PREFERRED, '1', '1', &labelOperator, 1};
    char *schdulerId = GetStr("scheduler_id");
    char *traceId = GetStr("trace_id");
    CInvokeOptions option{500, 500, &res, 1,    &extension, 1,          &opt, 1, fake,    2, &affinity,
                          0,   0,   1,    fake, 2,          schdulerId, fake, 2, traceId, 1};
    std::cout << option.customExtensions[0].value << std::endl;
    CInstanceAllocation cInsAlloc;
    auto cErr = CAcquireInstance(stateId, &meta, &option, &cInsAlloc);
    ASSERT_EQ(cErr.code, 0);
    auto err = YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "aaa");
    YR::Libruntime::InstanceAllocation insAlloc2{"functionId", "funcSig", "instanceId", "leaseId", 0};
    EXPECT_CALL(*lr.get(), AcquireInstance(_, _, _))
        .WillOnce(Return(std::make_pair<YR::Libruntime::InstanceAllocation, YR::Libruntime::ErrorInfo>(
            std::move(insAlloc2), std::move(err))));
    cErr = CAcquireInstance(stateId, &meta, &option, &cInsAlloc);
    ASSERT_EQ(cErr.code, 1001);
    ASSERT_EQ(std::string(cErr.message), "aaa");
    free(cErr.message);
}
