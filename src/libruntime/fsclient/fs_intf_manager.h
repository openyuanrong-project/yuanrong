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

#pragma once

#include <unordered_map>

#include "absl/synchronization/mutex.h"
#include "src/libruntime/clientsmanager/clients_manager.h"
#include "src/libruntime/fsclient/fs_intf_reader_writer.h"

namespace YR {
namespace Libruntime {
enum class ProtocolType {
    GRPC = 0,
    LITEBUS = 1,
};
class FSIntfManager {
public:
    FSIntfManager(const std::shared_ptr<ClientsManager> &clientsMgr)
        : clientsMgr(clientsMgr)
    {
    }
    virtual ~FSIntfManager() = default;
    virtual std::shared_ptr<FSIntfReaderWriter> NewFsIntfClient(const std::string &srcInstance,
                                                                const std::string &dstInstance,
                                                                const std::string &runtimeID,
                                                                const ReaderWriterClientOption &option,
                                                                ProtocolType type);
    virtual std::shared_ptr<FSIntfReaderWriter> NewEventIntfClient(const std::string &srcInstance,
                                                                const std::string &dstInstance,
                                                                const std::string &runtimeID,
                                                                const ReaderWriterClientOption &option,
                                                                ProtocolType type);
    // if direct runtime rw is not exist, return system intf rw
    virtual std::shared_ptr<FSIntfReaderWriter> TryGet(const std::string &instanceID);
    virtual std::shared_ptr<FSIntfReaderWriter> Get(const std::string &instanceID);
    virtual bool Emplace(const std::string &instanceID, const std::shared_ptr<FSIntfReaderWriter> &intf);
    virtual void Remove(const std::string &instanceID);
    virtual void Clear();
    virtual std::shared_ptr<FSIntfReaderWriter> TryGetEventIntfs(const std::string &instanceID);
    virtual std::shared_ptr<FSIntfReaderWriter> GetEventIntfs(const std::string &instanceID);
    virtual bool EmplaceEventIntfs(const std::string &instanceID, const std::shared_ptr<FSIntfReaderWriter> &intf);

    void UpdateSystemIntf(const std::shared_ptr<FSIntfReaderWriter> &intf);
    std::shared_ptr<FSIntfReaderWriter> GetSystemIntf();

protected:
    mutable absl::Mutex mu;
    mutable absl::Mutex eventMu;
    std::unordered_map<std::string, std::shared_ptr<FSIntfReaderWriter>> rtIntfs;
    std::unordered_map<std::string, std::shared_ptr<FSIntfReaderWriter>> eventIntfs;
    std::shared_ptr<FSIntfReaderWriter> systemIntf;
    std::shared_ptr<ClientsManager> clientsMgr;
};
}  // namespace Libruntime
}  // namespace YR
