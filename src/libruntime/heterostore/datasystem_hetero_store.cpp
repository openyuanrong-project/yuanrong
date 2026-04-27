/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
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

#include "datasystem_hetero_store.h"
#include "src/libruntime/utils/datasystem_utils.h"

namespace YR {
namespace Libruntime {

std::vector<datasystem::DeviceBlobList> BuildDsDeviceBlobList(const std::vector<DeviceBlobList> &devBlobList)
{
    std::vector<datasystem::DeviceBlobList> dsDevBlobList;
    dsDevBlobList.reserve(devBlobList.size());
    for (const auto &devBlob : devBlobList) {
        datasystem::DeviceBlobList dsDevBlob;
        dsDevBlob.deviceIdx = devBlob.deviceIdx;

        dsDevBlob.blobs.reserve(devBlob.blobs.size());
        for (const auto &blob : devBlob.blobs) {
            datasystem::Blob dsBlob;
            dsBlob.pointer = blob.pointer;
            dsBlob.size = blob.size;
            dsDevBlob.blobs.push_back(dsBlob);
        }
        dsDevBlobList.push_back(dsDevBlob);
    }
    return dsDevBlobList;
}

ErrorInfo DatasystemHeteroStore::Init(datasystem::ConnectOptions &options)
{
    connectOptions.host = options.host;
    connectOptions.port = options.port;
    connectOptions.clientPublicKey = options.clientPublicKey;
    connectOptions.clientPrivateKey = options.clientPrivateKey;
    connectOptions.serverPublicKey = options.serverPublicKey;
    connectOptions.accessKey = options.accessKey;
    connectOptions.secretKey = options.secretKey;
    connectOptions.token = options.token;
    connectOptions.connectTimeoutMs = options.connectTimeoutMs;
    connectOptions.tenantId = options.tenantId;
    return ErrorInfo();
}

ErrorInfo DatasystemHeteroStore::DoInitOnce()
{
    this->dsHeteroClient = std::make_shared<datasystem::HeteroClient>(connectOptions);
    datasystem::Status status = dsHeteroClient->Init();
    std::string msg = "failed to init hetero client, errMsg:" + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    isInit = true;
    return ErrorInfo();
}

void DatasystemHeteroStore::InitOnce(void)
{
    std::call_once(this->initFlag, [this]() { this->initErr = this->DoInitOnce(); });
}

void DatasystemHeteroStore::Shutdown()
{
    if (!dsHeteroClient) {
        return;
    }
    datasystem::Status status = dsHeteroClient->ShutDown();
    std::string msg = "shutdown hetero client failed, errMsg:" + status.ToString();
    if (!status.IsOk()) {
        YRLOG_WARN("hetero object client Shutdown fail. Status code: {}, Msg: {}", fmt::underlying(status.GetCode()),
                   status.ToString());
    }
}

ErrorInfo DatasystemHeteroStore::DevDelete(const std::vector<std::string> &objectIds,
                                        std::vector<std::string> &failedObjectIds)
{
    HETERO_STORE_INIT_ONCE();
    datasystem::Status status = dsHeteroClient->DevDelete(objectIds, failedObjectIds);
    std::string msg = "delete hetero object failed, errMsg:" + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    return ErrorInfo();
}

ErrorInfo DatasystemHeteroStore::DevLocalDelete(const std::vector<std::string> &objectIds,
                                             std::vector<std::string> &failedObjectIds)
{
    HETERO_STORE_INIT_ONCE();
    datasystem::Status status = dsHeteroClient->DevLocalDelete(objectIds, failedObjectIds);
    std::string msg = "local delete hetero object failed, errMsg:" + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    return ErrorInfo();
}

ErrorInfo DatasystemHeteroStore::DevSubscribe(const std::vector<std::string> &keys,
                                              const std::vector<DeviceBlobList> &blob2dList,
                                              std::vector<std::shared_ptr<YR::Libruntime::HeteroFuture>> &futureVec)
{
    HETERO_STORE_INIT_ONCE();
    YRLOG_DEBUG("start DevSubscribe, keys size is {}, blob2dList size is {}, futureVec size is {}", keys.size(),
                blob2dList.size(), futureVec.size());
    auto dsDevBlobList = BuildDsDeviceBlobList(blob2dList);
    std::vector<datasystem::Future> dsFutureVec{};
    datasystem::Status status = dsHeteroClient->DevSubscribe(keys, dsDevBlobList, dsFutureVec);
    std::string msg = "DevSubscribe failed, errMsg:" + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    for (datasystem::Future dsFuture : dsFutureVec) {
        auto heteroFuture =
            std::make_shared<YR::Libruntime::HeteroFuture>(std::make_shared<datasystem::Future>(dsFuture));
        futureVec.push_back(heteroFuture);
    }
    return ErrorInfo();
}

ErrorInfo DatasystemHeteroStore::DevPublish(const std::vector<std::string> &keys,
                                            const std::vector<DeviceBlobList> &blob2dList,
                                            std::vector<std::shared_ptr<YR::Libruntime::HeteroFuture>> &futureVec)
{
    HETERO_STORE_INIT_ONCE();
    YRLOG_DEBUG("start DevPublish, keys size is {}, blob2dList size is {}, futureVec size is {}", keys.size(),
                blob2dList.size(), futureVec.size());
    auto dsDevBlobList = BuildDsDeviceBlobList(blob2dList);
    std::vector<datasystem::Future> dsFutureVec{};
    datasystem::Status status = dsHeteroClient->DevPublish(keys, dsDevBlobList, dsFutureVec);
    std::string msg = "DevPublish failed, errMsg:" + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    for (datasystem::Future dsFuture : dsFutureVec) {
        auto heteroFuture =
            std::make_shared<YR::Libruntime::HeteroFuture>(std::make_shared<datasystem::Future>(dsFuture));
        futureVec.push_back(heteroFuture);
    }
    return ErrorInfo();
}

ErrorInfo DatasystemHeteroStore::DevMSet(const std::vector<std::string> &keys,
                                         const std::vector<DeviceBlobList> &blob2dList,
                                         std::vector<std::string> &failedKeys)
{
    HETERO_STORE_INIT_ONCE();
    YRLOG_DEBUG("start DevMSet, keys size is {}, blob2dList size is {}, failedKeys size is {}", keys.size(),
                blob2dList.size(), failedKeys.size());
    auto dsDevBlobList = BuildDsDeviceBlobList(blob2dList);
    datasystem::Status status = dsHeteroClient->DevMSet(keys, dsDevBlobList, failedKeys);
    std::string msg = "devmset failed, errMsg:" + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    return ErrorInfo();
}

ErrorInfo DatasystemHeteroStore::DevMGet(const std::vector<std::string> &keys,
                                         const std::vector<DeviceBlobList> &blob2dList,
                                         std::vector<std::string> &failedKeys, int32_t timeoutMs)
{
    HETERO_STORE_INIT_ONCE();
    YRLOG_DEBUG("start DevMGet, keys size is {}, blob2dList size is {}, failedKeys size is {}", keys.size(),
                blob2dList.size(), failedKeys.size());
    auto dsDevBlobList = BuildDsDeviceBlobList(blob2dList);
    datasystem::Status status = dsHeteroClient->DevMGet(keys, dsDevBlobList, failedKeys, timeoutMs);
    std::string msg = "DevMGet failed, errMsg:" + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    return ErrorInfo();
}
}  // namespace Libruntime
}  // namespace YR