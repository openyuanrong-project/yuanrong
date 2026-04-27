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

#pragma once

#include <future>
#include <json.hpp>
#include "src/dto/accelerate.h"
#include "src/libruntime/fsclient/fs_intf.h"
#include "src/utility/logger/logger.h"
#include "gmock/gmock.h"

using namespace YR::Libruntime;
namespace YR {
namespace test {
const std::string META_PREFIX = "0000000000000000";
class MockFsIntf : public YR::Libruntime::FSIntf {
public:
    MockFsIntf(){};
    ~MockFsIntf(){};
    void CreateAsync(const CreateRequest &req, CreateRespCallback respCallback, CreateCallBack callback,
                     int timeoutSec) override
    {
        CreateResponse response;
        response.set_instanceid("58f32000-0000-4000-8000-0ecfe00dd5e5");
        if (isReqNormal) {
            response.set_code(::common::ErrorCode::ERR_NONE);
        } else {
            response.set_code(::common::ErrorCode::ERR_PARAM_INVALID);
            respCallback(response);
            callbackPromise.set_value(1);
            return;
        }
        respCallback(response);
        NotifyRequest notifyReq;
        notifyReq.set_requestid("34a3b92ad0a6b79900");
        if (isReqNormal) {
            notifyReq.set_code(::common::ErrorCode::ERR_NONE);
        } else {
            notifyReq.set_code(::common::ErrorCode::ERR_PARAM_INVALID);
        }
        callback(notifyReq);
        callbackPromise.set_value(1);
    };
    void GroupCreateAsync(const CreateRequests &reqs, CreateRespsCallback respCallback, CreateCallBack callback,
                          int timeoutSec = -1) override
    {
        CreateResponses responses;
        for (int i = 0; i < reqs.requests_size(); i++) {
            responses.add_instanceids("58f32000-0000-4000-8000-0ecfe00dd5e5-" + std::to_string(i));
        }
        if (isReqNormal) {
            responses.set_code(::common::ErrorCode::ERR_NONE);
        } else {
            responses.set_code(::common::ErrorCode::ERR_PARAM_INVALID);
        }
        respCallback(responses);
        NotifyRequest notifyReq;
        notifyReq.set_requestid("34a3b92ad0a6b79900");
        if (isReqNormal) {
            notifyReq.set_code(::common::ErrorCode::ERR_NONE);
        } else {
            notifyReq.set_code(::common::ErrorCode::ERR_PARAM_INVALID);
        }
        callback(notifyReq);
        callbackPromise.set_value(1);
    };
    ErrorInfo Start(const std::string &, const std::string &, const std::string &, const std::string &,
                    const YR::Libruntime::SubscribeFunc &) override
    {
        return ErrorInfo();
    };
    void Stop(void) override{};
    bool IsHealth(void) override { return true; };
    void InvokeAsync(const std::shared_ptr<InvokeMessageSpec> &req, InvokeCallBack callback, int timeout) override
    {
        NotifyRequest notifyReq;
        if (isAcquireResponse) {
            nlohmann::json instanceResp;
            if (isReqNormal) {
                notifyReq.set_code(common::ERR_NONE);
                instanceResp["errorCode"] = 6030;
            } else {
                notifyReq.set_code(common::ERR_PARAM_INVALID);
                instanceResp["errorCode"] = 6031;
            }
            instanceResp["errorMessage"] = "msg";
            instanceResp["funcKey"] = "funcKey";
            instanceResp["funcSig"] = "funcSig";
            instanceResp["instanceID"] = "instanceID";
            instanceResp["threadID"] = "leaseId";
            instanceResp["leaseInterval"] = 5000;
            instanceResp["schedulerTime"] = 1.0;
            auto smallObject = notifyReq.add_smallobjects();
            smallObject->set_value(META_PREFIX + instanceResp.dump());
        } else if (isBatchRenew) {
            auto smallObject = notifyReq.add_smallobjects();
            if (isReqNormal) {
                smallObject->set_value(META_PREFIX + "{\"errorCode\": 6030,\"instanceAllocSucceed\": {},\"instanceAllocFailed\": {\"leaseId\": {\"errorCode\": 111111,\"errorMessage\": \"sssss\"}},\"leaseInterval\": 1000,\"schedulerTime\": 0.000123337}");
            } else {
                smallObject->set_value("");
            }
        } else {
            if (isReqNormal) {
                notifyReq.set_code(::common::ErrorCode::ERR_NONE);
                libruntime::FunctionMeta funcMeta;
                funcMeta.set_classname("classname");
                std::string serializeFuncMeta;
                funcMeta.SerializeToString(&serializeFuncMeta);
                auto smallObject = notifyReq.add_smallobjects();
                smallObject->set_id(req->Immutable().requestid());
                smallObject->set_value(serializeFuncMeta);
            } else if (needCheckArgs) {
                std::string value = req->Mutable().args(1).value();
                std::size_t found = value.find("instanceInvokeLabel");
                nlohmann::json instanceResp;
                if (found != std::string::npos) {
                    notifyReq.set_code(common::ERR_NONE);
                    instanceResp["errorCode"] = 6030;
                } else {
                    notifyReq.set_code(common::ERR_PARAM_INVALID);
                    instanceResp["errorCode"] = 6031;
                }
                std::cout << "Value: " << value << std::endl;
                instanceResp["errorMessage"] = "msg";
                instanceResp["funcKey"] = "funcKey";
                instanceResp["funcSig"] = "funcSig";
                instanceResp["instanceID"] = "instanceID";
                instanceResp["threadID"] = "leaseId";
                instanceResp["leaseInterval"] = 5000;
                instanceResp["schedulerTime"] = 1.0;
                auto smallObject = notifyReq.add_smallobjects();
                smallObject->set_value(META_PREFIX + instanceResp.dump());
            } else {
                notifyReq.set_code(::common::ErrorCode::ERR_PARAM_INVALID);
            }
        }
        callback(notifyReq, ErrorInfo());
    };
    void CallResultAsync(const std::shared_ptr<CallResultMessageSpec> req, CallResultCallBack callback) override{};
    void KillAsync(const KillRequest &req, KillCallBack callback, int timeout) override
    {
        KillResponse resp;
        if (isReqNormal) {
            resp.set_code(::common::ErrorCode::ERR_NONE);
        } else {
            resp.set_code(::common::ErrorCode::ERR_SCHEDULE_PLUGIN_CONFIG);
        }
        if (isGetInstance) {
            // Must match libruntime.FunctionMeta JSON (field className); JsonStringToMessage in GetInstance rejects binary.
            resp.set_message(R"({"className":"classname"})");
        } else {
            AccelerateMsgQueueHandle handler{.name = "name"};
            resp.set_message(handler.ToJson());
        }
        callback(resp, ErrorInfo());
        try {
            killCallbackPromise.set_value(1);
        } catch (const std::future_error &e) {
            YRLOG_DEBUG("killCallbackPromise has already set value");
        }
    };
    void ExitAsync(const ExitRequest &req, ExitCallBack callback) override{};
    void StateSaveAsync(const StateSaveRequest &req, StateSaveCallBack callback) override{};
    void StateLoadAsync(const StateLoadRequest &req, StateLoadCallBack callback) override{};
    void CreateRGroupAsync(const CreateResourceGroupRequest &req, CreateResourceGroupCallBack callback,
                             int timeout) override{};
    
    MOCK_METHOD(void, ReturnCallResult,
                (const std::shared_ptr<CallResultMessageSpec> result, bool isCreate, CallResultCallBack callback),
                (override));

    bool isReqNormal = true;
    bool isGetInstance = false;
    bool isAcquireResponse = false;
    bool isBatchRenew = false;
    bool needCheckArgs = false;
    std::promise<int> callbackPromise = std::promise<int>();
    std::promise<int> killCallbackPromise = std::promise<int>();
    std::shared_future<int> callbackFuture = callbackPromise.get_future();
    std::shared_future<int> killCallbackFuture = killCallbackPromise.get_future();
};
}  // namespace test
}  // namespace YR
