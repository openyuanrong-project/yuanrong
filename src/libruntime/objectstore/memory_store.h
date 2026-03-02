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

#pragma once

#include <deque>
#include <exception>
#include <functional>
#include <memory>
#include <mutex>
#include <unordered_map>
#include <unordered_set>
#include <vector>

#include <msgpack.hpp>
#include "absl/synchronization/notification.h"

#include "object_store.h"
#include "src/dto/data_object.h"
#include "src/libruntime/waiting_object_manager.h"
#include "src/utility/timer_worker.h"

namespace YR {
namespace Libruntime {
class WaitingObjectManager;
using ObjectReadyCallback = std::function<void(const ErrorInfo &)>;
using ObjectReadyCallbackWithData = std::function<void(const ErrorInfo &, std::shared_ptr<Buffer>)>;
using EventCallbackWithData = std::function<void(const ErrorInfo &, std::shared_ptr<Buffer>)>;
using YR::utility::TimerWorker;
const int EVENT_TIMER_EXEC_TIMES = 1;
struct GeneratorRes {
    std::string objectId;
    ErrorInfo err;
};

struct EventCallbackInfo {
    EventCallbackWithData eventCallback;
    std::deque<std::shared_ptr<std::string>> buffers;
    std::shared_ptr<YR::utility::Timer> timer;
};

enum IncreInDataSystemEnum : int { INCREASE_IN_DS = 1, INCREASING_IN_DS = 0, NOT_INCREASE_IN_DS = -1 };
struct ObjectDetail {
    std::mutex _mu;
    std::shared_ptr<Buffer> data;
    int localRefCount = 0;
    bool storeInMemory = false;
    bool storeInDataSystem =
        false;  // for obj in ds, when RefCount Decre to 0, SHOULD also decrease dsObjectStore refcount.
    bool increInDataSystem = false;
    IncreInDataSystemEnum increInDataSystemEnum = IncreInDataSystemEnum::NOT_INCREASE_IN_DS;
    bool ready = true;  // only return value of yr::Invoke or yr::Create is UNREADY. Otherwise, ready.
    ErrorInfo err;
    std::list<ObjectReadyCallback> callbacks;
    std::list<ObjectReadyCallbackWithData> callbacksWithData;
    std::promise<std::vector<std::string>> instanceIds;
    std::shared_future<std::vector<std::string>> instanceIdsFuture;
    std::atomic<uint64_t> getIndex{0};
    std::unordered_map<uint64_t, GeneratorRes> generatorResMap;
    std::condition_variable cv;
    bool finished = false;
    absl::Notification notification;
    std::promise<std::string> instanceRoute;
    std::shared_future<std::string> instanceRouteFuture;

    ObjectDetail()
    {
        instanceIds = std::promise<std::vector<std::string>>();
        instanceIdsFuture = instanceIds.get_future();
        instanceRoute = std::promise<std::string>();
        instanceRouteFuture = instanceRoute.get_future();
    }
};

class MemoryStore {
public:
    ErrorInfo Init(const std::string &addr, int port, std::int32_t connectTimeout);
    void Init(std::shared_ptr<ObjectStore> _dsObjectStore, std::shared_ptr<WaitingObjectManager> _waitingObjectManager);
    // Default Put to datasystem
    ErrorInfo Put(std::shared_ptr<Buffer> data, const std::string &objID,
                  const std::unordered_set<std::string> &nestedID, const CreateParam &createParam = {});
    ErrorInfo Put(std::shared_ptr<Buffer> data, const std::string &objID,
                  const std::unordered_set<std::string> &nestedID, bool toDataSystem,
                  const CreateParam &createParam = {}, const ErrorInfo &err = {});
    ErrorInfo Put(std::shared_ptr<Buffer> data, const std::string &objID,
                  const std::unordered_set<std::string> &nestedID, bool toDataSystem, const ErrorInfo &err);
    SingleResult Get(const std::string &objID, int timeoutMS);
    MultipleResult Get(const std::vector<std::string> &ids, int timeoutMS);
    ErrorInfo IncreGlobalReference(const std::vector<std::string> &objectIds);
    ErrorInfo IncreGlobalReference(const std::vector<std::string> &objectIds, bool toDataSystem);
    ErrorInfo IncreDSGlobalReference(const std::vector<std::string> &objectIds);
    std::pair<ErrorInfo, std::vector<std::string>> IncreGlobalReference(const std::vector<std::string> &objectIds,
                                                                        const std::string &remoteId);
    ErrorInfo DecreGlobalReference(const std::vector<std::string> &objectIds);
    std::pair<ErrorInfo, std::vector<std::string>> DecreGlobalReference(const std::vector<std::string> &objectIds,
                                                                        const std::string &remoteId);
    std::vector<int> QueryGlobalReference(const std::vector<std::string> &objectIds);
    ErrorInfo ReleaseGRefs(const std::string &remoteId);
    ErrorInfo GenerateKey(std::string &key, const std::string &prefix, bool isPut = true);
    ErrorInfo GenerateReturnObjectIds(const std::string &requestId,
                                      std::vector<YR::Libruntime::DataObject> &returnObjs);
    ErrorInfo GetPrefix(const std::string &key, std::string &prefix);

    void Clear();
    // check whether ids in DS, if not, put to DS.
    ErrorInfo AlsoPutToDS(const std::string &id, const CreateParam &createParam = {});
    ErrorInfo AlsoPutToDS(const std::vector<std::string> &ids, const CreateParam &createParam = {});
    ErrorInfo AlsoPutToDS(const std::unordered_set<std::string> &ids, const CreateParam &createParam = {});
    ErrorInfo CreateBuffer(const std::string &objectId, size_t dataSize, std::shared_ptr<Buffer> &dataBuf,
                           const CreateParam &createParam = {});
    std::pair<ErrorInfo, std::shared_ptr<Buffer>> GetBuffer(const std::string &ids, int timeoutMS);
    std::pair<ErrorInfo, std::vector<std::shared_ptr<Buffer>>> GetBuffers(const std::vector<std::string> &ids,
                                                                          int timeoutMS);
    std::pair<RetryInfo, std::vector<std::shared_ptr<Buffer>>> GetBuffersWithoutRetry(
        const std::vector<std::string> &ids, int timeoutMS);

    bool SetReady(const std::vector<DataObject> &objs);
    bool SetReady(const std::string &id);
    bool SetError(const std::vector<DataObject> &objs, const ErrorInfo &err);
    bool SetError(const std::string &id, const ErrorInfo &err);
    bool AddReadyCallback(const std::string &id, ObjectReadyCallback callback);
    bool AddReadyCallbackWithData(const std::string &id, ObjectReadyCallbackWithData callback);
    void AddEventCallbackWithData(const std::string &id, EventCallbackWithData callback);
    bool AddEventData(const std::string &reqId, const std::string &eventData);
    void AddEventTimer(const std::string &reqId, int timeoutSec);
    bool AddReturnObject(const std::vector<DataObject> &returnObjs);
    bool AddReturnObject(const std::string &objId);
    bool SetInstanceId(const std::string &id, const std::string &instanceId);
    std::string GetInstanceId(const std::string &objId);
    bool SetInstanceIds(const std::string &id, const std::vector<std::string> &instanceIds);
    std::pair<std::vector<std::string>, ErrorInfo> GetInstanceIds(const std::string &objId, int timeoutSec);
    bool SetInstanceRoute(const std::string &id, const std::string &instanceRoute);
    std::string GetInstanceRoute(const std::string &objId, int timeoutSec = ZERO_TIMEOUT);
    // Directly Get from datasystem
    SingleResult DSDirectGet(const std::string &objID, int timeoutMS);
    ErrorInfo GetLastError(const std::string &objId);
    ErrorInfo IncreaseObjRef(const std::vector<std::string> &objectIds);
    void BindObjRefInReq(const std::string &requestId, const std::vector<std::string> &objectIds);
    std::vector<std::string> UnbindObjRefInReq(const std::string &requestId);
    bool IsExistedInLocal(const std::string &objId);
    bool AddGenerator(const std::string &generatorId);
    std::pair<ErrorInfo, std::string> GetOutput(const std::string &generatorId, bool block);

    void AddOutput(const std::string &generatorId, const std::string &objectId, uint64_t index,
                   const ErrorInfo &err = ErrorInfo());
    void GeneratorFinished(const std::string &generatorId);
    std::string GenerateObjectId(const std::string &generatorId, uint64_t index);
    bool IsReady(const std::string &objId);

private:
    ErrorInfo DoPutToDS(const std::string &id, const CreateParam &createParam);
    ErrorInfo GetBuffersFromMem(const std::vector<std::string> &ids, std::vector<std::shared_ptr<Buffer>> &result,
                                std::vector<size_t> &notInMemIndex, std::vector<std::string> &vecGetFromDS);
    std::pair<ErrorInfo, std::vector<std::string>> IncreaseGRefInMemoryAndDs(const std::vector<std::string> &objectIds,
                                                                             bool toDataSystem,
                                                                             const std::string &remoteId = "");
    std::vector<std::string> DecreaseGRefInMemory(const std::vector<std::string> &objectIds);
    void DoEventCallback(EventCallbackWithData callback, const std::string &reqId, const std::string &eventData);

    std::mutex mu;
    std::mutex reqMu;
    std::mutex eventMu;
    std::shared_ptr<ObjectStore> dsObjectStore;
    std::shared_ptr<WaitingObjectManager> waitingObjectManager;
    std::unordered_map<std::string, std::shared_ptr<ObjectDetail>> storeMap;
    std::unordered_map<std::string, std::vector<std::string>> reqObjStore;
    std::unordered_map<std::string, EventCallbackInfo> reqEventCallbackMap_;
    std::shared_ptr<TimerWorker> timerWorker_;
    std::condition_variable eventCond_;
    size_t totalInMemBufSize;
};

}  // namespace Libruntime
}  // namespace YR