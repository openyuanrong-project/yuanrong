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

#pragma once
#include <future>
#include <memory>
#include "datasystem/hetero_client.h"
#include "device_util.h"
#include "src/dto/buffer.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/heterostore/hetero_future.h"
#include "src/libruntime/utils/exception.h"

namespace YR {
namespace Libruntime {
class HeteroStore {
public:
    virtual ~HeteroStore() = default;

    /**
     * @brief Init HeteroStore object.
     * @param[in] options The param of connect datasystem.
     */
    virtual ErrorInfo Init(datasystem::ConnectOptions &options) = 0;

    /**
     * @brief Shutdown the state client.
     * @return ERR_OK on success; the error code otherwise.
     */
    virtual void Shutdown() = 0;

    /**
     * @brief Invoke worker client to delete all the given objectId.
     * @param[in] objectIds The vector of the objId.
     * @param[out] failedObjectIds The failed delete objIds.
     * @return ERR_OK on any key success; the error code otherwise.
     */
    virtual ErrorInfo Delete(const std::vector<std::string> &objectIds, std::vector<std::string> &failedObjectIds) = 0;

    /**
     * @brief LocalDelete interface. After calling this interface, the data replica stored in the data system by the
     * current client connection will be deleted.
     * @param[in] objectIds The objectIds of the data expected to be deleted.
     * @param[out] failedObjectIds Partial failures will be returned through this parameter.
     * @return ERR_OK on when return sucesss; the error code otherwise.
     */
    virtual ErrorInfo LocalDelete(const std::vector<std::string> &objectIds,
                                  std::vector<std::string> &failedObjectIds) = 0;

    /**
     * @brief Subscribe data from device.
     * @param[in] keys A list of keys corresponding to the blob2dList.
     * @param[in] blob2dList A list of structures describing the Device memory.
     * @param[out] futureVec A list of futures to track the  operation.
     * @return ERR_OK on when return all futures sucesss; the error code otherwise.
     */
    virtual ErrorInfo DevSubscribe(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                                   std::vector<std::shared_ptr<YR::Libruntime::HeteroFuture>> &futureVec) = 0;

    /**
     * @brief Publish data to device.
     * @param[in] keys A list of keys corresponding to the blob2dList.
     * @param[in] blob2dList A list of structures describing the Device memory.
     * @param[out] futureVec A list of futures to track the  operation.
     * @return ERR_OK on when return all futures sucesss; the error code otherwise.
     */
    virtual ErrorInfo DevPublish(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                                 std::vector<std::shared_ptr<YR::Libruntime::HeteroFuture>> &futureVec) = 0;

    /**
     * @brief Store Device cache through the data system, caching corresponding keys-blob2dList metadata to the data
     * system
     * @param[in] keys Keys corresponding to blob2dList
     * @param[in] blob2dList List describing the structure of Device memory
     * @param[out] failedKeys Returns failed keys if caching fails
     * @return ERR_OK on when return sucesssfully; the error code otherwise.
     */
    virtual ErrorInfo DevMSet(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                              std::vector<std::string> &failedKeys) = 0;

    /**
     * @brief Retrieves data from the Device through the data system, storing it in the corresponding DeviceBlobList
     * @param[in] keys Keys corresponding to blob2dList
     * @param[in] blob2dList List describing the structure of Device memory
     * @param[out] failedKeys Returns failed keys if retrieval fails
     * @param[in] timeoutMs Provides a timeout time, defaulting to 0
     * @return ERR_OK on when return sucesssfully; the error code otherwise.
     */
    virtual ErrorInfo DevMGet(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                              std::vector<std::string> &failedKeys, int32_t timeoutMs) = 0;
};

}  // namespace Libruntime
}  // namespace YR