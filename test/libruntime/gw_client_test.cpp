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
#include <boost/beast/http.hpp>
#include <cctype>
#include "src/libruntime/utils/utils.h"
#define private public
#define protected public
#include "src/dto/data_object.h"
#include "src/libruntime/gwclient/gw_client.h"
using namespace testing;
using namespace YR::Libruntime;
using namespace YR::utility;

namespace {
std::string GwTestGetHeader(const std::unordered_map<std::string, std::string> &headers, const std::string &name)
{
    auto it = headers.find(name);
    if (it != headers.end()) {
        return it->second;
    }
    for (const auto &kv : headers) {
        if (kv.first.size() != name.size()) {
            continue;
        }
        bool same = true;
        for (size_t i = 0; i < name.size(); ++i) {
            if (std::tolower(static_cast<unsigned char>(kv.first[i])) !=
                std::tolower(static_cast<unsigned char>(name[i]))) {
                same = false;
                break;
            }
        }
        if (same) {
            return kv.second;
        }
    }
    return {};
}
}  // namespace

namespace YR {
namespace test {
std::vector<std::shared_ptr<YR::Libruntime::Buffer>> BuildBuffers();
std::vector<std::shared_ptr<YR::Libruntime::Buffer>> BuildEmptyBuffer();
bool ValidBuffers(std::shared_ptr<std::vector<std::shared_ptr<YR::Libruntime::Buffer>>> buffers);

enum ResponseType {
    HTTP_ERROR_COMMUNICATION = 0,
    HTTP_BAD_REQUEST = 1,
    HTTP_OK_AND_SUCCESS = 2,
    HTTP_OK_BUT_FAILED = 3,
    HTTP_TIMEOUT = 4,
};
const size_t getSize = 1;
const uint HTTP_BAD_REQUEST_CODE = 400;
const uint HTTP_OK_CODE = 200;
static std::unordered_map<ResponseType, std::string> responseTypeMap = {
    {ResponseType::HTTP_ERROR_COMMUNICATION, "TEST_HTTP_ERROR_COMMUNICATION"},
    {ResponseType::HTTP_BAD_REQUEST, "TEST_HTTP_BAD_REQUEST"},
    {ResponseType::HTTP_OK_AND_SUCCESS, "TEST_HTTP_OK_AND_SUCCESS"},
    {ResponseType::HTTP_OK_BUT_FAILED, "TEST_HTTP_OK_BUT_FAILED"},
    {ResponseType::HTTP_TIMEOUT, "TEST_HTTP_TIMEOUT"}};

const int32_t testTimeoutMs = 50;
const int32_t connectTimeoutS = 1;
const std::string errorCommunicationMsg = "error_code: ";
const std::string badRequestMsg = "failed response status_code: ";
const std::string timeoutErr = "http request timeout s: " + std::to_string(connectTimeoutS);
const std::string failedMsg = "system error";
const std::string g_requestId = "cae7c30c8d63f5ed00";
const std::string g_instanceId = "cae7c30c8d63f5ed";
void CommonCallback(const std::string &rspType, const HttpCallbackFunction &receiver, const std::string &body)
{
    if (responseTypeMap[ResponseType::HTTP_ERROR_COMMUNICATION] == rspType) {
        receiver(body, boost::asio::error::make_error_code(boost::asio::error::fault), 0);
    } else if (responseTypeMap[ResponseType::HTTP_BAD_REQUEST] == rspType) {
        receiver(body, boost::beast::error_code(), HTTP_BAD_REQUEST_CODE);
    } else if (responseTypeMap[ResponseType::HTTP_OK_AND_SUCCESS] == rspType) {
        receiver(body, boost::beast::error_code(), HTTP_OK_CODE);
    } else if (responseTypeMap[ResponseType::HTTP_OK_BUT_FAILED] == rspType) {
        receiver(body, boost::beast::error_code(), HTTP_OK_CODE);
    } else {
        std::cout << "unknown rspType:" << rspType;
    }
}
class MockHttpClient : public HttpClient {
public:
    MockHttpClient() = default;
    ~MockHttpClient() = default;
    ErrorInfo Init(const ConnectionParam &param) override
    {
        return ErrorInfo();
    }

    void SubmitInvokeRequest(const http::verb &method, const std::string &target,
                             const std::unordered_map<std::string, std::string> &headers, const std::string &body,
                             const std::shared_ptr<std::string> requestId,
                             const HttpCallbackFunction &receiver) override
    {
        std::string rspType = GwTestGetHeader(headers, REMOTE_CLIENT_ID_KEY);
        if (rspType.empty()) {
            rspType = GwTestGetHeader(headers, REMOTE_CLIENT_ID_KEY_NEW);
        }
        if (rspType.empty()) {
            receiver(body, boost::asio::error::make_error_code(boost::asio::error::fault), 0);
            return;
        }
        auto code = common::ERR_NONE;
        std::string msg = "";
        if (responseTypeMap[ResponseType::HTTP_TIMEOUT] == rspType) {
            return;
        }
        if (!(responseTypeMap[ResponseType::HTTP_OK_AND_SUCCESS] == rspType)) {
            code = common::ERR_INNER_SYSTEM_ERROR;
            msg = failedMsg;
        }
        if (POSIX_LEASE == target || POSIX_LEASE_KEEPALIVE == target) {
            auto rsp = BuildLeaseResponse(code, msg);
            std::string rspBody;
            rsp.SerializeToString(&rspBody);
            CommonCallback(rspType, receiver, rspBody);
        } else if (POSIX_CREATE == target) {
            std::vector<std::shared_ptr<DataObject>> returnObjects = {};
            std::vector<YR::Libruntime::StackTraceInfo> infos = {};
            auto rsp = BuildNotifyRequest(*requestId, code, msg, returnObjects, infos);
            std::string rspBody;
            rsp.SerializeToString(&rspBody);
            CommonCallback(rspType, receiver, rspBody);
        } else if (POSIX_INVOKE == target) {
            std::vector<std::shared_ptr<DataObject>> returnObjects = {};
            std::vector<YR::Libruntime::StackTraceInfo> infos = {};
            auto rsp = BuildNotifyRequest(*requestId, code, msg, returnObjects, infos);
            std::string rspBody;
            rsp.SerializeToString(&rspBody);
            CommonCallback(rspType, receiver, rspBody);
        } else if (target.find("invocation") != std::string::npos) {
            std::string rspBody = "{\"code\":150444,\"message\":\"instance label not found\"}";
            CommonCallback(rspType, receiver, rspBody);
        } else if (POSIX_KILL == target) {
            auto rsp = BuildKillResponse(code, msg);
            std::string rspBody;
            rsp.SerializeToString(&rspBody);
            CommonCallback(rspType, receiver, rspBody);
        } else if (POSIX_OBJ_GET == target) {
            auto rsp = BuildObjGetResponse(code, msg, BuildBuffers());
            std::string rspBody;
            rsp.SerializeToString(&rspBody);
            CommonCallback(rspType, receiver, rspBody);
        } else if (POSIX_OBJ_PUT == target) {
            auto rsp = BuildObjPutResponse(code, msg);
            std::string rspBody;
            rsp.SerializeToString(&rspBody);
            CommonCallback(rspType, receiver, rspBody);
        } else if (POSIX_OBJ_DECREASE == target) {
            std::vector<std::string> failedObjectIds = {};
            auto rsp = BuildDecreaseRefResponse(code, msg, failedObjectIds);
            std::string rspBody;
            rsp.SerializeToString(&rspBody);
            CommonCallback(rspType, receiver, rspBody);
        } else if (POSIX_OBJ_INCREASE == target) {
            std::vector<std::string> failedObjectIds = {"failed_obj-1"};
            auto rsp = BuildIncreaseRefResponse(code, msg, failedObjectIds);
            std::string rspBody;
            rsp.SerializeToString(&rspBody);
            CommonCallback(rspType, receiver, rspBody);
        } else if (POSIX_KV_GET == target) {
            auto rsp = BuildKvGetResponse(code, msg, BuildBuffers());
            std::string rspBody;
            rsp.SerializeToString(&rspBody);
            CommonCallback(rspType, receiver, rspBody);
        } else if (POSIX_KV_SET == target) {
            auto rsp = BuildKvSetResponse(code, msg);
            std::string rspBody;
            rsp.SerializeToString(&rspBody);
            CommonCallback(rspType, receiver, rspBody);
        } else if (POSIX_KV_MSET_TX == target) {
            auto rsp = BuildKvMSetTxResponse(code, msg);
            std::string rspBody;
            rsp.SerializeToString(&rspBody);
            CommonCallback(rspType, receiver, rspBody);
        } else if (POSIX_KV_DEL == target) {
            std::vector<std::string> failedKeys = {};
            auto rsp = BuildKvDelResponse(code, msg, failedKeys);
            std::string rspBody;
            rsp.SerializeToString(&rspBody);
            CommonCallback(rspType, receiver, rspBody);
        } else {
            std::cout << "unknown target:" << target;
        }
    }

    ErrorInfo ReInit(const std::shared_ptr<std::string> requestId) override
    {
        return ErrorInfo();
    }

    LeaseResponse BuildLeaseResponse(const common::ErrorCode code, const std::string &msg)
    {
        LeaseResponse rsp;
        rsp.set_code(code);
        rsp.set_message(msg);
        return rsp;
    }

    CreateResponse BuildCreateResponse(const common::ErrorCode code, const std::string &msg,
                                       const std::string &instanceID)
    {
        CreateResponse rsp;
        rsp.set_code(code);
        rsp.set_message(msg);
        rsp.set_instanceid(instanceID);
        return rsp;
    }

    NotifyRequest BuildNotifyRequest(const std::string &requestID, const common::ErrorCode code, const std::string &msg,
                                     const std::vector<std::shared_ptr<DataObject>> &returnObjects,
                                     const std::vector<YR::Libruntime::StackTraceInfo> &infos)
    {
        NotifyRequest req;
        req.set_requestid(requestID);
        req.set_code(code);
        req.set_message(msg);
        for (size_t i = 0; i < returnObjects.size(); i++) {
            if (returnObjects[i]->data != nullptr && returnObjects[i]->data->IsNative() &&
                returnObjects[i]->putDone == false) {
                auto smallObject = req.add_smallobjects();
                smallObject->set_id(returnObjects[i]->id);
                smallObject->set_value(returnObjects[i]->data->ImmutableData(), returnObjects[i]->data->GetSize());
            }
        }
        for (size_t i = 0; i < infos.size(); i++) {
            auto setInfo = req.add_stacktraceinfos();
            auto eles = infos[i].StackTraceElements();
            for (size_t j = 0; j < eles.size(); j++) {
                auto stackTraceEle = setInfo->add_stacktraceelements();
                stackTraceEle->set_classname(eles[j].className);
                stackTraceEle->set_methodname(eles[j].methodName);
                stackTraceEle->set_filename(eles[j].fileName);
                stackTraceEle->set_linenumber(eles[j].lineNumber);
                for (auto &it : eles[j].extensions) {
                    stackTraceEle->mutable_extensions()->insert({it.first, it.second});
                }
            }
            setInfo->set_type(infos[i].Type());
            setInfo->set_message(infos[i].Message());
        }
        return req;
    }

    KillResponse BuildKillResponse(const common::ErrorCode code, const std::string &msg)
    {
        KillResponse rsp;
        rsp.set_code(code);
        rsp.set_message(msg);
        return rsp;
    }

    GetResponse BuildObjGetResponse(const common::ErrorCode code, const std::string &msg,
                                    const std::vector<std::shared_ptr<YR::Libruntime::Buffer>> &values)
    {
        GetResponse rsp;
        rsp.set_code(code);
        rsp.set_message(msg);
        for (const auto &value : values) {
            rsp.add_buffers(value->ImmutableData(), value->GetSize());
        }
        return rsp;
    }

    PutResponse BuildObjPutResponse(const common::ErrorCode code, const std::string &msg)
    {
        PutResponse rsp;
        rsp.set_code(code);
        rsp.set_message(msg);
        return rsp;
    }

    KvSetResponse BuildKvSetResponse(const common::ErrorCode code, const std::string &msg)
    {
        KvSetResponse rsp;
        rsp.set_code(code);
        rsp.set_message(msg);
        return rsp;
    }

    KvMSetTxResponse BuildKvMSetTxResponse(const common::ErrorCode code, const std::string &msg)
    {
        KvMSetTxResponse rsp;
        rsp.set_code(code);
        rsp.set_message(msg);
        return rsp;
    }

    KvGetResponse BuildKvGetResponse(const common::ErrorCode code, const std::string &msg,
                                     const std::vector<std::shared_ptr<YR::Libruntime::Buffer>> &values)
    {
        KvGetResponse rsp;
        rsp.set_code(code);
        rsp.set_message(msg);
        for (const auto &value : values) {
            rsp.add_values(value->ImmutableData(), value->GetSize());
        }
        return rsp;
    }

    KvDelResponse BuildKvDelResponse(const common::ErrorCode code, const std::string &msg,
                                     const std::vector<std::string> &failedKeys)
    {
        KvDelResponse rsp;
        rsp.set_code(code);
        rsp.set_message(msg);
        for (const auto &key : failedKeys) {
            rsp.add_failedkeys(key);
        }
        return rsp;
    }

    IncreaseRefResponse BuildIncreaseRefResponse(const common::ErrorCode code, const std::string &msg,
                                                 const std::vector<std::string> &failedObjectIds)
    {
        IncreaseRefResponse rsp;
        rsp.set_code(code);
        rsp.set_message(msg);
        for (const auto &objId : failedObjectIds) {
            rsp.add_failedobjectids(objId);
        }
        return rsp;
    }

    DecreaseRefResponse BuildDecreaseRefResponse(const common::ErrorCode code, const std::string &msg,
                                                 const std::vector<std::string> &failedObjectIds)
    {
        DecreaseRefResponse rsp;
        rsp.set_code(code);
        rsp.set_message(msg);
        for (const auto &objId : failedObjectIds) {
            rsp.add_failedobjectids(objId);
        }
        return rsp;
    }
};

std::vector<std::shared_ptr<YR::Libruntime::Buffer>> BuildBuffers()
{
    std::vector<std::shared_ptr<YR::Libruntime::Buffer>> buffers;
    const char *str = "value";
    size_t len = std::strlen(str);
    auto buf = std::make_shared<YR::Libruntime::NativeBuffer>(len);
    buf->MemoryCopy(str, len);
    buffers.emplace_back(std::move(buf));
    return buffers;
}

std::vector<std::shared_ptr<YR::Libruntime::Buffer>> BuildEmptyBuffer()
{
    size_t number = 1;
    std::vector<std::shared_ptr<YR::Libruntime::Buffer>> buffers(number);
    buffers[0] = std::make_shared<YR::Libruntime::NativeBuffer>(0);
    return buffers;
}

bool ValidBuffers(std::shared_ptr<std::vector<std::shared_ptr<YR::Libruntime::Buffer>>> buffers)
{
    EXPECT_TRUE(buffers->size() == getSize);
    const size_t firstIndex = 0;
    auto buffersOne = BuildBuffers();
    for (const auto &sbuf : *buffers) {
        if (buffersOne[firstIndex]->GetSize() != sbuf->GetSize()) {
            return false;
        }
        if (!std::equal(
                static_cast<const char *>(buffersOne[firstIndex]->ImmutableData()),
                static_cast<const char *>(buffersOne[firstIndex]->ImmutableData()) + buffersOne[firstIndex]->GetSize(),
                static_cast<const char *>(sbuf->ImmutableData()))) {
            return false;
        }
    }
    return true;
}

class GwClientTest : public testing::Test {
public:
    GwClientTest(){
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
    };
    ~GwClientTest(){};
    void SetUp() override
    {
        auto librtCfg = std::make_shared<LibruntimeConfig>();
        librtCfg->functionSystemIpAddr = "";
        librtCfg->functionSystemPort = 0;
        librtCfg->functionIds[libruntime::LanguageType::Cpp] =
            "default/0-function-function/$latest";
        FSIntfHandlers handlers;
        auto httpClient = std::make_unique<MockHttpClient>();
        httpClient->Init(ConnectionParam{librtCfg->functionSystemIpAddr, std::to_string(librtCfg->functionSystemPort)});
        auto security = std::make_shared<Security>();
        std::string accessKey = "access_key";
        SensitiveValue secretKey = std::string("secret_key");
        security->SetAKSKAndCredential(accessKey, secretKey);
        gwClient_ =
            std::make_shared<GwClient>(librtCfg->functionIds[libruntime::LanguageType::Cpp], handlers, security);
        gwClient_->Init(std::move(httpClient), connectTimeoutS);
    }
    void TearDown() override
    {
        gwClient_.reset();
    }
    std::shared_ptr<GwClient> gwClient_;
};

void ErrorMsgCheck(const ErrorInfo &err, ErrorCode code, const std::string &msg)
{
    EXPECT_EQ(err.Code(), code);
    if (msg.empty()) {
        return;
    }
    std::string::size_type code_idx = err.Msg().find(msg);
    EXPECT_TRUE(code_idx != std::string::npos);
}

TEST_F(GwClientTest, TestLease)
{
    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_ERROR_COMMUNICATION]).OK());
    ErrorMsgCheck(gwClient_->Lease(), ErrorCode::ERR_INNER_COMMUNICATION, errorCommunicationMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_BAD_REQUEST]).OK());
    ErrorMsgCheck(gwClient_->Lease(), ErrorCode::ERR_PARAM_INVALID, badRequestMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_BUT_FAILED]).OK());
    ErrorMsgCheck(gwClient_->Lease(), ErrorCode::ERR_INNER_SYSTEM_ERROR, failedMsg);

    EXPECT_TRUE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_AND_SUCCESS]).OK());
    ErrorMsgCheck(gwClient_->Lease(), ErrorCode::ERR_OK, "");

    ErrorMsgCheck(gwClient_->KeepLease(), ErrorCode::ERR_OK, "");

    ErrorMsgCheck(gwClient_->Release(), ErrorCode::ERR_OK, "");

    gwClient_->Stop();
}

TEST_F(GwClientTest, TestCreate)
{
    CreateRequest req;
    req.set_requestid(g_requestId);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_ERROR_COMMUNICATION]).OK());
    gwClient_->CreateAsync(
        req, [](const CreateResponse &rsp) { EXPECT_TRUE(rsp.code() == common::ERR_INNER_COMMUNICATION); },
        [](const NotifyRequest &req) { EXPECT_TRUE(0 == 1); });

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_BAD_REQUEST]).OK());
    gwClient_->CreateAsync(
        req, [](const CreateResponse &rsp) { EXPECT_TRUE(rsp.code() == common::ERR_PARAM_INVALID); },
        [](const NotifyRequest &req) { EXPECT_TRUE(0 == 1); });

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_BUT_FAILED]).OK());
    gwClient_->CreateAsync(
        req, [](const CreateResponse &rsp) { EXPECT_TRUE(rsp.code() == common::ERR_INNER_SYSTEM_ERROR); },
        [](const NotifyRequest &req) { EXPECT_TRUE(0 == 1); });

    EXPECT_TRUE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_AND_SUCCESS]).OK());
    gwClient_->CreateAsync(
        req, [](const CreateResponse &rsp) { EXPECT_TRUE(rsp.code() == common::ERR_NONE); },
        [](const NotifyRequest &req) { EXPECT_TRUE(req.code() == common::ERR_NONE); });
}

TEST_F(GwClientTest, TestInvoke)
{
    InvokeRequest req;
    req.set_requestid(g_requestId);
    auto messageSpec = std::make_shared<InvokeMessageSpec>(std::move(req));
    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_ERROR_COMMUNICATION]).OK());
    gwClient_->InvokeAsync(messageSpec, [](const NotifyRequest &req, const ErrorInfo &err) {
        EXPECT_TRUE(req.code() == common::ERR_INNER_COMMUNICATION);
    });

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_BAD_REQUEST]).OK());
    gwClient_->InvokeAsync(messageSpec, [](const NotifyRequest &req, const ErrorInfo &err) {
        EXPECT_TRUE(req.code() == common::ERR_PARAM_INVALID);
    });

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_BUT_FAILED]).OK());
    gwClient_->InvokeAsync(messageSpec, [](const NotifyRequest &req, const ErrorInfo &err) {
        EXPECT_TRUE(req.code() == common::ERR_INNER_SYSTEM_ERROR);
    });

    EXPECT_TRUE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_AND_SUCCESS]).OK());
    gwClient_->InvokeAsync(messageSpec, [](const NotifyRequest &req, const ErrorInfo &err) {
        EXPECT_TRUE(req.code() == common::ERR_NONE);
    });
}

TEST_F(GwClientTest, TestInvocation)
{
    const std::string url =
        "/serverless/v2/functions/"
        "wisefunction:cn:iot:8d86c63b22e24d9ab650878b75408ea6:function:test-jiuwen-session-004-bj:$latest/invocations";
    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_ERROR_COMMUNICATION]).OK());
    auto spec = std::make_shared<InvokeSpec>();
    spec->traceId = "traceId";
    spec->requestId = "requestId";
    std::string arg = "{\"key\":\"invoke\"}";
    InvokeArg libArg;
    libArg.dataObj = std::make_shared<YR::Libruntime::DataObject>(0, arg.size());
    libArg.dataObj->data->MemoryCopy(arg.data(), arg.size());
    spec->invokeArgs.emplace_back(std::move(libArg));
    gwClient_->InvocationAsync(url, spec,
                               [](const std::string &requestId, Libruntime::ErrorCode code, const std::string &result) {
                                   EXPECT_TRUE(code == Libruntime::ErrorCode::ERR_INNER_COMMUNICATION);
                               });

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_BAD_REQUEST]).OK());
    gwClient_->InvocationAsync(url, spec,
                               [](const std::string &requestId, Libruntime::ErrorCode code, const std::string &result) {
                                   EXPECT_TRUE(code == Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR);
                               });

    EXPECT_TRUE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_AND_SUCCESS]).OK());
    gwClient_->InvocationAsync(url, spec,
                               [](const std::string &requestId, Libruntime::ErrorCode code, const std::string &result) {
                                   EXPECT_TRUE(code == Libruntime::ErrorCode::ERR_OK);
                               });
}

TEST_F(GwClientTest, TestKill)
{
    KillRequest req;
    req.set_instanceid(g_instanceId);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_ERROR_COMMUNICATION]).OK());
    gwClient_->KillAsync(req, [](const KillResponse &rsp, const ErrorInfo &err) {
        EXPECT_TRUE(rsp.code() == common::ERR_INNER_COMMUNICATION);
    });

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_BAD_REQUEST]).OK());
    gwClient_->KillAsync(req, [](const KillResponse &rsp, const ErrorInfo &err) {
        EXPECT_TRUE(rsp.code() == common::ERR_PARAM_INVALID);
    });

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_BUT_FAILED]).OK());
    gwClient_->KillAsync(req, [](const KillResponse &rsp, const ErrorInfo &err) {
        EXPECT_TRUE(rsp.code() == common::ERR_INNER_SYSTEM_ERROR);
    });

    EXPECT_TRUE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_AND_SUCCESS]).OK());
    gwClient_->KillAsync(
        req, [](const KillResponse &rsp, const ErrorInfo &err) { EXPECT_TRUE(rsp.code() == common::ERR_NONE); });
}

TEST_F(GwClientTest, TestPosixObjGet)
{
    std::vector<std::string> objIds = {"obj"};
    auto result = std::make_shared<std::vector<std::shared_ptr<YR::Libruntime::Buffer>>>();
    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_ERROR_COMMUNICATION]).OK());
    ErrorMsgCheck(gwClient_->PosixObjGet(objIds, result, testTimeoutMs), ErrorCode::ERR_INNER_COMMUNICATION,
                  errorCommunicationMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_BAD_REQUEST]).OK());
    ErrorMsgCheck(gwClient_->PosixObjGet(objIds, result, testTimeoutMs), ErrorCode::ERR_PARAM_INVALID, badRequestMsg);

    result->resize(getSize);
    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_BUT_FAILED]).OK());
    ErrorMsgCheck(gwClient_->PosixObjGet(objIds, result, testTimeoutMs), ErrorCode::ERR_INNER_SYSTEM_ERROR, failedMsg);

    result->resize(getSize);
    EXPECT_TRUE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_AND_SUCCESS]).OK());
    ErrorMsgCheck(gwClient_->PosixObjGet(objIds, result, testTimeoutMs), ErrorCode::ERR_OK, "");
    EXPECT_TRUE(ValidBuffers(result));

    EXPECT_TRUE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_TIMEOUT]).OK());
    ErrorMsgCheck(gwClient_->PosixObjGet(objIds, result, testTimeoutMs), ErrorCode::ERR_INNER_COMMUNICATION,
                  timeoutErr);
}

TEST_F(GwClientTest, TestObjGet)
{
    std::string obj = "obj";
    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_ERROR_COMMUNICATION]).OK());
    ErrorMsgCheck(gwClient_->Get(obj, testTimeoutMs).first, ErrorCode::ERR_INNER_COMMUNICATION, errorCommunicationMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_BAD_REQUEST]).OK());
    ErrorMsgCheck(gwClient_->Get(obj, testTimeoutMs).first, ErrorCode::ERR_PARAM_INVALID, badRequestMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_BUT_FAILED]).OK());
    ErrorMsgCheck(gwClient_->Get(obj, testTimeoutMs).first, ErrorCode::ERR_INNER_SYSTEM_ERROR, failedMsg);

    EXPECT_TRUE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_AND_SUCCESS]).OK());
    ErrorMsgCheck(gwClient_->Get(obj, testTimeoutMs).first, ErrorCode::ERR_OK, "");
    ErrorMsgCheck(gwClient_->GetBuffers({obj}, testTimeoutMs).first, ErrorCode::ERR_OK, "");
    ErrorMsgCheck(gwClient_->GetBuffersWithoutRetry({obj}, testTimeoutMs).first.errorInfo, ErrorCode::ERR_OK, "");

    EXPECT_TRUE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_TIMEOUT]).OK());
    ErrorMsgCheck(gwClient_->Get(obj, testTimeoutMs).first, ErrorCode::ERR_INNER_COMMUNICATION, timeoutErr);
}

TEST_F(GwClientTest, TestPosixObjPut)
{
    PutRequest req;
    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_ERROR_COMMUNICATION]).OK());
    ErrorMsgCheck(gwClient_->PosixObjPut(req), ErrorCode::ERR_INNER_COMMUNICATION, errorCommunicationMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_BAD_REQUEST]).OK());
    ErrorMsgCheck(gwClient_->PosixObjPut(req), ErrorCode::ERR_PARAM_INVALID, badRequestMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_BUT_FAILED]).OK());
    ErrorMsgCheck(gwClient_->PosixObjPut(req), ErrorCode::ERR_INNER_SYSTEM_ERROR, failedMsg);

    EXPECT_TRUE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_AND_SUCCESS]).OK());
    ErrorMsgCheck(gwClient_->PosixObjPut(req), ErrorCode::ERR_OK, "");
}

TEST_F(GwClientTest, TestObjPut)
{
    auto result = std::make_shared<YR::Libruntime::NativeBuffer>(1);
    std::string objID = "obj_id";
    std::unordered_set<std::string> nestedID = {"nested_id"};
    CreateParam param;
    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_ERROR_COMMUNICATION]).OK());
    ErrorMsgCheck(gwClient_->Put(result, objID, nestedID, param), ErrorCode::ERR_INNER_COMMUNICATION,
                  errorCommunicationMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_BAD_REQUEST]).OK());
    ErrorMsgCheck(gwClient_->Put(result, objID, nestedID, param), ErrorCode::ERR_PARAM_INVALID, badRequestMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_BUT_FAILED]).OK());
    ErrorMsgCheck(gwClient_->Put(result, objID, nestedID, param), ErrorCode::ERR_INNER_SYSTEM_ERROR, failedMsg);

    EXPECT_TRUE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_AND_SUCCESS]).OK());
    ErrorMsgCheck(gwClient_->Put(result, objID, nestedID, param), ErrorCode::ERR_OK, "");
}

TEST_F(GwClientTest, TestPosixKvGet)
{
    std::vector<std::string> keys = {"key"};
    auto result = std::make_shared<std::vector<std::shared_ptr<YR::Libruntime::Buffer>>>();
    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_ERROR_COMMUNICATION]).OK());
    ErrorMsgCheck(gwClient_->PosixKvGet(keys, result, testTimeoutMs), ErrorCode::ERR_INNER_COMMUNICATION,
                  errorCommunicationMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_BAD_REQUEST]).OK());
    ErrorMsgCheck(gwClient_->PosixKvGet(keys, result, testTimeoutMs), ErrorCode::ERR_PARAM_INVALID, badRequestMsg);

    result->resize(getSize);
    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_BUT_FAILED]).OK());
    ErrorMsgCheck(gwClient_->PosixKvGet(keys, result, testTimeoutMs), ErrorCode::ERR_INNER_SYSTEM_ERROR, failedMsg);

    result->resize(getSize);
    EXPECT_TRUE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_AND_SUCCESS]).OK());
    ErrorMsgCheck(gwClient_->PosixKvGet(keys, result, testTimeoutMs), ErrorCode::ERR_OK, "");
    EXPECT_TRUE(ValidBuffers(result));

    EXPECT_TRUE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_TIMEOUT]).OK());
    ErrorMsgCheck(gwClient_->PosixKvGet(keys, result, testTimeoutMs), ErrorCode::ERR_INNER_COMMUNICATION, timeoutErr);
}

TEST_F(GwClientTest, TestKVGetRead)
{
    std::vector<std::string> keys = {"key"};
    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_ERROR_COMMUNICATION]).OK());
    ErrorMsgCheck(gwClient_->Read(keys, testTimeoutMs, true).second, ErrorCode::ERR_INNER_COMMUNICATION,
                  errorCommunicationMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_BAD_REQUEST]).OK());
    ErrorMsgCheck(gwClient_->Read(keys, testTimeoutMs, true).second, ErrorCode::ERR_PARAM_INVALID, badRequestMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_BUT_FAILED]).OK());
    ErrorMsgCheck(gwClient_->Read(keys, testTimeoutMs, true).second, ErrorCode::ERR_INNER_SYSTEM_ERROR, failedMsg);

    EXPECT_TRUE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_AND_SUCCESS]).OK());
    ErrorMsgCheck(gwClient_->Read(keys, testTimeoutMs, false).second, ErrorCode::ERR_OK, "");
    ErrorMsgCheck(gwClient_->Read("key", testTimeoutMs).second, ErrorCode::ERR_OK, "");

    EXPECT_TRUE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_TIMEOUT]).OK());
    ErrorMsgCheck(gwClient_->Read(keys, testTimeoutMs, true).second, ErrorCode::ERR_INNER_COMMUNICATION, timeoutErr);
}

TEST_F(GwClientTest, TestKvSet)
{
    std::string key = "";
    auto value = std::make_shared<YR::Libruntime::NativeBuffer>(0);
    SetParam setParam;
    auto result = std::make_shared<std::vector<std::shared_ptr<YR::Libruntime::Buffer>>>();
    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_ERROR_COMMUNICATION]).OK());
    ErrorMsgCheck(gwClient_->PosixKvSet(key, value, setParam), ErrorCode::ERR_INNER_COMMUNICATION,
                  errorCommunicationMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_BAD_REQUEST]).OK());
    ErrorMsgCheck(gwClient_->PosixKvSet(key, value, setParam), ErrorCode::ERR_PARAM_INVALID, badRequestMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_BUT_FAILED]).OK());
    ErrorMsgCheck(gwClient_->PosixKvSet(key, value, setParam), ErrorCode::ERR_INNER_SYSTEM_ERROR, failedMsg);

    EXPECT_TRUE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_AND_SUCCESS]).OK());
    ErrorMsgCheck(gwClient_->PosixKvSet(key, value, setParam), ErrorCode::ERR_OK, "");
    ErrorMsgCheck(gwClient_->Write(key, value, setParam), ErrorCode::ERR_OK, "");
}

TEST_F(GwClientTest, TestMKvSetTx)
{
    std::vector<std::string> keys = {"", ""};
    auto value1 = std::make_shared<YR::Libruntime::NativeBuffer>(0);
    auto value2 = std::make_shared<YR::Libruntime::NativeBuffer>(1);
    std::vector<std::shared_ptr<Buffer>> vals = {value1, value2};
    MSetParam mSetParam;
    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_ERROR_COMMUNICATION]).OK());
    ErrorMsgCheck(gwClient_->PosixKvMSetTx(keys, vals, mSetParam), ErrorCode::ERR_INNER_COMMUNICATION,
                  errorCommunicationMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_BAD_REQUEST]).OK());
    ErrorMsgCheck(gwClient_->PosixKvMSetTx(keys, vals, mSetParam), ErrorCode::ERR_PARAM_INVALID, badRequestMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_BUT_FAILED]).OK());
    ErrorMsgCheck(gwClient_->PosixKvMSetTx(keys, vals, mSetParam), ErrorCode::ERR_INNER_SYSTEM_ERROR, failedMsg);

    EXPECT_TRUE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_AND_SUCCESS]).OK());
    ErrorMsgCheck(gwClient_->PosixKvMSetTx(keys, vals, mSetParam), ErrorCode::ERR_OK, "");
    ErrorMsgCheck(gwClient_->MSetTx(keys, vals, mSetParam), ErrorCode::ERR_OK, "");
}

TEST_F(GwClientTest, TestPosixKvDel)
{
    std::vector<std::string> keys = {};
    auto failedKeys = std::make_shared<std::vector<std::string>>();
    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_ERROR_COMMUNICATION]).OK());
    ErrorMsgCheck(gwClient_->PosixKvDel(keys, failedKeys), ErrorCode::ERR_INNER_COMMUNICATION, errorCommunicationMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_BAD_REQUEST]).OK());
    ErrorMsgCheck(gwClient_->PosixKvDel(keys, failedKeys), ErrorCode::ERR_PARAM_INVALID, badRequestMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_BUT_FAILED]).OK());
    ErrorMsgCheck(gwClient_->PosixKvDel(keys, failedKeys), ErrorCode::ERR_INNER_SYSTEM_ERROR, failedMsg);

    EXPECT_TRUE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_AND_SUCCESS]).OK());
    ErrorMsgCheck(gwClient_->PosixKvDel(keys, failedKeys), ErrorCode::ERR_OK, "");
}

TEST_F(GwClientTest, TestDel)
{
    std::string key = "key";
    auto failedKeys = std::make_shared<std::vector<std::string>>();
    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_ERROR_COMMUNICATION]).OK());
    ErrorMsgCheck(gwClient_->Del(key), ErrorCode::ERR_INNER_COMMUNICATION, errorCommunicationMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_BAD_REQUEST]).OK());
    ErrorMsgCheck(gwClient_->Del(key), ErrorCode::ERR_PARAM_INVALID, badRequestMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_BUT_FAILED]).OK());
    ErrorMsgCheck(gwClient_->Del(key), ErrorCode::ERR_INNER_SYSTEM_ERROR, failedMsg);

    EXPECT_TRUE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_AND_SUCCESS]).OK());
    ErrorMsgCheck(gwClient_->Del(key), ErrorCode::ERR_OK, "");
}

TEST_F(GwClientTest, TestPosixIncreaseRef)
{
    std::vector<std::string> objIds = {};
    auto failedObjIds = std::make_shared<std::vector<std::string>>();
    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_ERROR_COMMUNICATION]).OK());
    ErrorMsgCheck(gwClient_->PosixGInCreaseRef(objIds, failedObjIds), ErrorCode::ERR_INNER_COMMUNICATION,
                  errorCommunicationMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_BAD_REQUEST]).OK());
    ErrorMsgCheck(gwClient_->PosixGInCreaseRef(objIds, failedObjIds), ErrorCode::ERR_PARAM_INVALID, badRequestMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_BUT_FAILED]).OK());
    ErrorMsgCheck(gwClient_->PosixGInCreaseRef(objIds, failedObjIds), ErrorCode::ERR_INNER_SYSTEM_ERROR, failedMsg);

    EXPECT_TRUE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_AND_SUCCESS]).OK());
    ErrorMsgCheck(gwClient_->PosixGInCreaseRef(objIds, failedObjIds), ErrorCode::ERR_OK, "");
}

TEST_F(GwClientTest, TestIncreaseRef)
{
    std::vector<std::string> objIds = {};
    auto failedObjIds = std::make_shared<std::vector<std::string>>();
    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_ERROR_COMMUNICATION]).OK());
    ErrorMsgCheck(gwClient_->IncreGlobalReference(objIds), ErrorCode::ERR_INNER_COMMUNICATION, errorCommunicationMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_BAD_REQUEST]).OK());
    ErrorMsgCheck(gwClient_->IncreGlobalReference(objIds), ErrorCode::ERR_PARAM_INVALID, badRequestMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_BUT_FAILED]).OK());
    ErrorMsgCheck(gwClient_->IncreGlobalReference(objIds), ErrorCode::ERR_INNER_SYSTEM_ERROR, failedMsg);

    EXPECT_TRUE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_AND_SUCCESS]).OK());
    ErrorMsgCheck(gwClient_->IncreGlobalReference(objIds), ErrorCode::ERR_OK, "");
}

TEST_F(GwClientTest, TestPosixDecreaseRef)
{
    std::vector<std::string> objIds = {};
    auto failedObjIds = std::make_shared<std::vector<std::string>>();
    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_ERROR_COMMUNICATION]).OK());
    ErrorMsgCheck(gwClient_->PosixGDecreaseRef(objIds, failedObjIds), ErrorCode::ERR_INNER_COMMUNICATION,
                  errorCommunicationMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_BAD_REQUEST]).OK());
    ErrorMsgCheck(gwClient_->PosixGDecreaseRef(objIds, failedObjIds), ErrorCode::ERR_PARAM_INVALID, badRequestMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_BUT_FAILED]).OK());
    ErrorMsgCheck(gwClient_->PosixGDecreaseRef(objIds, failedObjIds), ErrorCode::ERR_INNER_SYSTEM_ERROR, failedMsg);

    EXPECT_TRUE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_AND_SUCCESS]).OK());
    ErrorMsgCheck(gwClient_->PosixGDecreaseRef(objIds, failedObjIds), ErrorCode::ERR_OK, "");
}

TEST_F(GwClientTest, TestDecreaseRef)
{
    std::vector<std::string> objIds = {"objId"};
    EXPECT_TRUE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_AND_SUCCESS]).OK());
    gwClient_->SetTenantId("tenantId");
    ErrorMsgCheck(gwClient_->DecreGlobalReference(objIds), ErrorCode::ERR_OK, "");
}

TEST_F(GwClientTest, TestUpdateToken)
{
    datasystem::SensitiveValue token("token");
    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_ERROR_COMMUNICATION]).OK());
    ErrorMsgCheck(gwClient_->UpdateToken(token), ErrorCode::ERR_OK, "");
}

TEST_F(GwClientTest, TestUpdateAkSk)
{
    std::string ak = "ak";
    datasystem::SensitiveValue sk("sk");
    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_ERROR_COMMUNICATION]).OK());
    ErrorMsgCheck(gwClient_->UpdateAkSk(ak, sk), ErrorCode::ERR_OK, "");
}

TEST_F(GwClientTest, TestGenerateKey)
{
    std::string key;
    std::string prefix = "591672113dc36b6a0000";
    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_ERROR_COMMUNICATION]).OK());
    ErrorMsgCheck(gwClient_->GenerateKey(key, prefix, false), ErrorCode::ERR_OK, "");
}

TEST_F(GwClientTest, TestGetPrefix)
{
    std::string key = "591672113dc36b6a0000";
    std::string prefix;
    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_ERROR_COMMUNICATION]).OK());
    ErrorMsgCheck(gwClient_->GetPrefix(key, prefix), ErrorCode::ERR_OK, "");
}

TEST_F(GwClientTest, TestParseObjGetResponse)
{
    size_t number = 1;
    auto result = std::make_shared<std::vector<std::shared_ptr<YR::Libruntime::Buffer>>>();
    result->resize(number);
    auto buffers = BuildEmptyBuffer();
    auto client = MockHttpClient();
    auto code = common::ERR_NONE;
    std::string msg = "";
    auto rsp = client.BuildObjGetResponse(code, msg, buffers);
    std::string rspBody;
    rsp.SerializeToString(&rspBody);
    gwClient_->ParseObjGetResponse(rspBody, result);
    EXPECT_TRUE((*result)[0] == nullptr);
}

TEST_F(GwClientTest, TestParseKvGetResponse)
{
    size_t number = 1;
    auto result = std::make_shared<std::vector<std::shared_ptr<YR::Libruntime::Buffer>>>();
    result->resize(number);
    auto buffers = BuildEmptyBuffer();
    auto client = MockHttpClient();
    auto code = common::ERR_NONE;
    std::string msg = "";
    auto rsp = client.BuildKvGetResponse(code, msg, buffers);
    std::string rspBody;
    rsp.SerializeToString(&rspBody);
    gwClient_->ParseKvGetResponse(rspBody, result);
    EXPECT_TRUE((*result)[0] == nullptr);
}

TEST_F(GwClientTest, When_HttpClient_Connecting_Do_CreateBuffer_Should_Return_OK_Test)
{
    std::string objectId = "111";
    std::shared_ptr<Buffer> dataBuf;
    CreateParam param;
    param.cacheType = CacheType::DISK;
    param.consistencyType = ConsistencyType::PRAM;
    param.writeMode = WriteMode::NONE_L2_CACHE_EVICT;
    EXPECT_TRUE(gwClient_->CreateBuffer(objectId, 10, dataBuf, param).OK());
    std::unordered_set<std::string> nestedIds{"222"};

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_ERROR_COMMUNICATION]).OK());
    ErrorMsgCheck(dataBuf->Seal(nestedIds), ErrorCode::ERR_INNER_COMMUNICATION, errorCommunicationMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_BAD_REQUEST]).OK());
    ErrorMsgCheck(dataBuf->Seal(nestedIds), ErrorCode::ERR_PARAM_INVALID, badRequestMsg);

    EXPECT_FALSE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_BUT_FAILED]).OK());
    ErrorMsgCheck(dataBuf->Seal(nestedIds), ErrorCode::ERR_INNER_SYSTEM_ERROR, failedMsg);

    EXPECT_TRUE(gwClient_->Start(responseTypeMap[ResponseType::HTTP_OK_AND_SUCCESS]).OK());
    ErrorMsgCheck(dataBuf->Seal(nestedIds), ErrorCode::ERR_OK, "");
}

TEST_F(GwClientTest, TestBuildHeader)
{
    auto header = gwClient_->BuildHeaders("1", "2", "3");
    auto c = header.at("tenantId");
    EXPECT_TRUE(c == "3");
}

TEST_F(GwClientTest, TestCreateStreamConsumerWillReturnError)
{
    auto c = std::make_shared<StreamConsumer>();
    auto err = gwClient_->CreateStreamConsumer("stream", SubscriptionConfig{}, c);
    ASSERT_EQ(err.Code(), ERR_INNER_SYSTEM_ERROR);
}

TEST_F(GwClientTest, TestQueryGlobalReferenceWillThrowLibRuntimeException)
{
    try {
        gwClient_->QueryGlobalReference(std::vector<std::string>());
        ASSERT_TRUE(false);
    } catch (YR::Libruntime::Exception &e) {
        ASSERT_EQ(e.Code(), ERR_INNER_SYSTEM_ERROR);
        ASSERT_EQ(e.Msg(), "QueryGlobalReference method is not supported when inCluster is false");
    }
}

TEST_F(GwClientTest, TestUnsupportedReq1)
{
    ErrorInfo err;
    datasystem::ConnectOptions options;
    DsConnectOptions connOptions;
    CreateResourceGroupRequest req;
    std::vector<std::string> objIds;
    std::vector<uint64_t> outSizes;
    std::shared_ptr<StreamProducer> producer;
    std::string returnKey;
    uint64_t gNum;
    try {
        gwClient_->CallResultAsync(nullptr, nullptr);
        ASSERT_TRUE(false);
    } catch (YR::Libruntime::Exception &e) {
        ASSERT_EQ(e.Code(), ERR_INNER_SYSTEM_ERROR);
    }
    try {
        gwClient_->CreateRGroupAsync(req, nullptr);
        ASSERT_TRUE(false);
    } catch (YR::Libruntime::Exception &e) {
        ASSERT_EQ(e.Code(), ERR_INNER_SYSTEM_ERROR);
    }

    err = gwClient_->Init("127.0.0.1", 11111, false, false, "runtimePublicKey",
                          datasystem::SensitiveValue("runtimePrivateKey"), "dsPublicKey",
                          datasystem::SensitiveValue("token"), "ak", datasystem::SensitiveValue("sk"));
    ASSERT_EQ(err.Code(), ERR_INNER_SYSTEM_ERROR);

    err = gwClient_->Init(options);
    ASSERT_EQ(err.Code(), ERR_INNER_SYSTEM_ERROR);

    try {
        gwClient_->QuerySize(objIds, outSizes);
        ASSERT_TRUE(false);
    } catch (YR::Libruntime::Exception &e) {
        ASSERT_EQ(e.Code(), ERR_INNER_SYSTEM_ERROR);
    }

    gwClient_->Shutdown();
    err = gwClient_->Init(connOptions);
    ASSERT_EQ(err.Code(), ERR_INNER_SYSTEM_ERROR);

    err = gwClient_->GenerateKey(returnKey);
    ASSERT_EQ(err.Code(), ERR_INNER_SYSTEM_ERROR);

    err = gwClient_->CreateStreamProducer("streamName", producer);
    ASSERT_EQ(err.Code(), ERR_INNER_SYSTEM_ERROR);

    err = gwClient_->DeleteStream("streamName");
    ASSERT_EQ(err.Code(), ERR_INNER_SYSTEM_ERROR);

    err = gwClient_->QueryGlobalProducersNum("streamName", gNum);
    ASSERT_EQ(err.Code(), ERR_INNER_SYSTEM_ERROR);

    err = gwClient_->QueryGlobalConsumersNum("streamName", gNum);
    ASSERT_EQ(err.Code(), ERR_INNER_SYSTEM_ERROR);
}

TEST_F(GwClientTest, TestUnsupportedReq3)
{
    ErrorInfo err;
    ExitRequest exitReq;
    StateSaveRequest saveStateReq;
    StateLoadRequest loadStateReq;
    try {
        gwClient_->ExitAsync(exitReq, [](const ExitResponse &) {});
        ASSERT_TRUE(false);
    } catch (Exception &e) {
        std::string errMsg = e.what();
        ASSERT_EQ(errMsg,
                  "ErrCode: 3003, ModuleCode: 20, ErrMsg: ExitAsync method not implemented when inCluster is false");
    }
    try {
        gwClient_->StateSaveAsync(saveStateReq, [](const StateSaveResponse) {});
        ASSERT_TRUE(false);
    } catch (Exception &e) {
        std::string errMsg = e.what();
        ASSERT_EQ(
            errMsg,
            "ErrCode: 3003, ModuleCode: 20, ErrMsg: StateSaveAsync method not implemented when inCluster is false");
    }
    try {
        gwClient_->StateLoadAsync(loadStateReq, [](const StateLoadResponse &) {});
        ASSERT_TRUE(false);
    } catch (Exception &e) {
        std::string errMsg = e.what();
        ASSERT_EQ(errMsg, "ErrCode: 3003, ModuleCode: 20, ErrMsg: StateLoadAsync is not supported with gateway client");
    }

    err = gwClient_->ReleaseGRefs("remoteId");
    ASSERT_EQ(err.Code(), ERR_PARAM_INVALID);
}

class VerbTest : public ::testing::Test {
    void SetUp() override {}
    void TearDown() override {}
};

TEST_F(VerbTest, GetMethodTest)
{
    std::string method = "GET";
    EXPECT_EQ(StringToVerb(method), boost::beast::http::verb::get);
}

TEST_F(VerbTest, PostMethodTest)
{
    std::string method = "POST";
    EXPECT_EQ(StringToVerb(method), boost::beast::http::verb::post);
}

TEST_F(VerbTest, PutMethodTest)
{
    std::string method = "PUT";
    EXPECT_EQ(StringToVerb(method), boost::beast::http::verb::put);
}

TEST_F(VerbTest, DeleteMethodTest)
{
    std::string method = "DELETE";
    EXPECT_EQ(StringToVerb(method), boost::beast::http::verb::delete_);
}

TEST_F(VerbTest, UnknownMethodTest)
{
    std::string method = "INVALID_METHOD";
    EXPECT_EQ(StringToVerb(method), boost::beast::http::verb::unknown);
}
}  // namespace test
}  // namespace YR