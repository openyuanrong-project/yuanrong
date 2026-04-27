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
#include "datasystem/utils/sensitive_value.h"
#include "hetero_store.h"
#include "src/libruntime/utils/constants.h"

namespace YR {
namespace Libruntime {
class DatasystemHeteroStore : public HeteroStore {
public:
    DatasystemHeteroStore() = default;
    ~DatasystemHeteroStore() override = default;
    void Shutdown() override;
    ErrorInfo Init(datasystem::ConnectOptions &connectOptions) override;
    ErrorInfo DevDelete(const std::vector<std::string> &objectIds, std::vector<std::string> &failedObjectIds) override;
    ErrorInfo DevLocalDelete(const std::vector<std::string> &objectIds,
                             std::vector<std::string> &failedObjectIds) override;
    ErrorInfo DevSubscribe(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                           std::vector<std::shared_ptr<YR::Libruntime::HeteroFuture>> &futureVec) override;
    ErrorInfo DevPublish(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                         std::vector<std::shared_ptr<YR::Libruntime::HeteroFuture>> &futureVec) override;
    ErrorInfo DevMSet(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                      std::vector<std::string> &failedKeys) override;
    ErrorInfo DevMGet(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                      std::vector<std::string> &failedKeys, int32_t timeoutMs) override;

private:
    void InitOnce(void);
    ErrorInfo DoInitOnce(void);
    bool isInit = false;
    std::once_flag initFlag;
    ErrorInfo initErr;
    datasystem::ConnectOptions connectOptions;
    std::shared_ptr<datasystem::HeteroClient> dsHeteroClient{nullptr};
};

#define HETERO_STORE_INIT_ONCE() \
    do {                         \
        InitOnce();              \
        if (!initErr.OK()) {     \
            return initErr;      \
        }                        \
    } while (0)

#define HETERO_STORE_INIT_ONCE_RETURN(_ret_) \
    do {                                     \
        InitOnce();                          \
        if (!initErr.OK()) {                 \
            return (_ret_);                  \
        }                                    \
    } while (0)

}  // namespace Libruntime
}  // namespace YR