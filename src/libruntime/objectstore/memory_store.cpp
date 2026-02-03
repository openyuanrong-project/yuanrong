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

#include "memory_store.h"
#include <algorithm>
#include "datasystem_object_store.h"
#include "src/dto/buffer.h"
#include "src/utility/id_generator.h"
#include "src/utility/logger/logger.h"
namespace YR {
namespace Libruntime {

void MemoryStore::Init(std::shared_ptr<ObjectStore> _dsObjectStore,
                       std::shared_ptr<WaitingObjectManager> _waitingObjectManager)
{
    std::lock_guard<std::mutex> lock(mu);
    dsObjectStore = _dsObjectStore;
    waitingObjectManager = _waitingObjectManager;
    timerWorker_ = std::make_shared<TimerWorker>();
}

ErrorInfo MemoryStore::GenerateKey(std::string &key, const std::string &prefix, bool isPut)
{
    return dsObjectStore->GenerateKey(key, prefix, isPut);
}

ErrorInfo MemoryStore::GenerateReturnObjectIds(const std::string &requestId,
                                               std::vector<YR::Libruntime::DataObject> &returnObjs)
{
    for (size_t i = 0; i < returnObjs.size(); i++) {
        returnObjs[i].id = YR::utility::IDGenerator::GenObjectId(requestId, i);
    }
    return ErrorInfo();
}

ErrorInfo MemoryStore::GetPrefix(const std::string &key, std::string &prefix)
{
    return dsObjectStore->GetPrefix(key, prefix);
}

// Default Put to datasystem
ErrorInfo MemoryStore::Put(std::shared_ptr<Buffer> data, const std::string &objID,
                           const std::unordered_set<std::string> &nestedID, const CreateParam &createParam)
{
    return this->Put(data, objID, nestedID, true, createParam);
}

ErrorInfo MemoryStore::Put(std::shared_ptr<Buffer> data, const std::string &objID,
                           const std::unordered_set<std::string> &nestedID, bool toDataSystem,
                           const CreateParam &createParam, const ErrorInfo &err)
{
    std::unique_lock<std::mutex> lock(mu);
    std::unique_lock<std::mutex> objectDetailLock;
    if (nestedID.find(objID) != nestedID.end()) {
        YRLOG_ERROR("Circular references detected! objID: {}", objID);
        ErrorInfo e;
        std::string msg = "MemoryStore::Put id: " + objID + " has circular reference in its nestedID.";
        e.SetErrorCode(ErrorCode::ERR_PARAM_INVALID);
        e.SetErrorMsg(msg);
        return e;
    }
    auto it = storeMap.find(objID);
    if (it != storeMap.end()) {
        std::shared_ptr<ObjectDetail> objDetail = it->second;
        objectDetailLock = std::unique_lock<std::mutex>(objDetail->_mu);
        // To prevent twice Put, check whether objID is already store in mem or ds
        if (objDetail->storeInMemory || objDetail->storeInDataSystem) {
            // fail, return err
            ErrorInfo e;
            std::string msg = "MemoryStore::Put id: " + objID + " is already exist in " +
                              (objDetail->storeInMemory ? "memory" : "datasystem");
            e.SetErrorCode(ErrorCode::ERR_KEY_ALREADY_EXIST);
            e.SetErrorMsg(msg);
            return e;
        }
        // forceToDatasystem (yr::Put)
        if (toDataSystem) {
            // unlock storeMap, and still lock objectDetail.
            lock.unlock();
            ErrorInfo dsErr = AlsoPutToDS(nestedID, createParam);
            if (dsErr.Code() != ErrorCode::ERR_OK) {
                YRLOG_ERROR("AlsoPutToDS for nestedIDs error.");
                return dsErr;
            }
            dsErr = dsObjectStore->Put(data, objID, nestedID, createParam);
            if (dsErr.Code() == ErrorCode::ERR_OK) {
                objDetail->storeInDataSystem = true;
            }
            return dsErr;
        }
        // check before save to memory
        if (!nestedID.empty()) {
            ErrorInfo e;
            std::string msg = "MemoryStore::Put putting id: " + objID + " to memory, should not have nestedID.";
            e.SetErrorCode(ErrorCode::ERR_PARAM_INVALID);
            e.SetErrorMsg(msg);
            return e;
        }
        // save to memory
        objDetail->data = data;
        objDetail->storeInMemory = true;
        objDetail->err = err;
        totalInMemBufSize += data->GetSize();
        return ErrorInfo();
    } else {
        // fail, return err
        ErrorInfo e;
        std::string msg =
            "MemoryStore::Put id: " + objID + " haven't Incre ref by this runtime. You should Incre ref before Put.";
        e.SetErrorCode(ErrorCode::ERR_PARAM_INVALID);
        e.SetErrorMsg(msg);
        return e;
    }
}

ErrorInfo MemoryStore::Put(std::shared_ptr<Buffer> data, const std::string &objID,
                           const std::unordered_set<std::string> &nestedID, bool toDataSystem, const ErrorInfo &err)
{
    return this->Put(data, objID, nestedID, toDataSystem, {}, err);
}

SingleResult MemoryStore::Get(const std::string &objID, int timeoutMS)
{
    std::unique_lock<std::mutex> lock(mu);
    auto it = storeMap.find(objID);
    if (it != storeMap.end()) {
        std::shared_ptr<ObjectDetail> objDetail = it->second;
        std::unique_lock<std::mutex> objectDetailLock(objDetail->_mu);
        if (objDetail->storeInMemory) {
            auto buffer = objDetail->data;
            return std::make_pair(ErrorInfo(), buffer);
        } else {
            lock.unlock();
            return dsObjectStore->Get(objID, timeoutMS);
        }
    } else {
        // Not tracked by MemoryStore. Directly fetch from DS.
        YRLOG_DEBUG("id {} not exist in storeMap. will Get from DS.", objID);
        return DSDirectGet(objID, timeoutMS);
    }
}

// Without storeMap management. Directly Get from DS
SingleResult MemoryStore::DSDirectGet(const std::string &objID, int timeoutMS)
{
    return dsObjectStore->Get(objID, timeoutMS);
}

MultipleResult MemoryStore::Get(const std::vector<std::string> &ids, int timeoutMS)
{
    std::vector<size_t> notInMemIndex;
    std::vector<std::string> vecGetFromDS;
    YR::Libruntime::ErrorInfo lastErr;
    std::vector<std::shared_ptr<Buffer>> result(ids.size());
    lastErr = GetBuffersFromMem(ids, result, notInMemIndex, vecGetFromDS);
    if (vecGetFromDS.size() == 0) {
        return std::make_pair(lastErr, result);
    }
    MultipleResult dsMultiRes = dsObjectStore->Get(vecGetFromDS, timeoutMS);
    if (dsMultiRes.first.Code() != ErrorCode::ERR_OK) {
        lastErr = dsMultiRes.first;
    }
    if (dsMultiRes.second.empty()) {
        return std::make_pair(lastErr, result);
    }
    for (size_t i = 0; i < vecGetFromDS.size(); i++) {
        int bufIndex = notInMemIndex[i];
        result[bufIndex] = dsMultiRes.second[i];
    }
    return std::make_pair(lastErr, result);
}

ErrorInfo MemoryStore::IncreGlobalReference(const std::vector<std::string> &objectIds)
{
    return this->IncreGlobalReference(objectIds, true);
}

std::pair<ErrorInfo, std::vector<std::string>> MemoryStore::IncreaseGRefInMemoryAndDs(
    const std::vector<std::string> &objectIds, bool toDataSystem, const std::string &remoteId)
{
    std::unique_lock<std::mutex> lock(mu);
    std::vector<std::string> shouldIncreInDS;
    std::vector<std::shared_ptr<ObjectDetail>> increseObjectDetails;
    std::vector<std::shared_ptr<ObjectDetail>> waitObjectDetails;
    for (const std::string &id : objectIds) {
        auto [it, insertSuccess] = storeMap.emplace(id, std::make_shared<ObjectDetail>());
        (void)insertSuccess;
        std::shared_ptr<ObjectDetail> objDetail = it->second;
        std::unique_lock<std::mutex> objectDetailLock(objDetail->_mu);
        if (toDataSystem) {
            if (objDetail->increInDataSystemEnum == IncreInDataSystemEnum::INCREASING_IN_DS) {
                waitObjectDetails.push_back(objDetail);
            }
            if (objDetail->increInDataSystemEnum == IncreInDataSystemEnum::NOT_INCREASE_IN_DS) {
                shouldIncreInDS.push_back(id);
                increseObjectDetails.push_back(objDetail);
                objDetail->increInDataSystemEnum = IncreInDataSystemEnum::INCREASING_IN_DS;
            }
        }
        objDetail->localRefCount++;
        if (objDetail->localRefCount == 1) {
            YRLOG_DEBUG("Incred id {} localRefCount {}", id, objDetail->localRefCount);
        } else {
            YRLOG_TRACE("Incred id {} localRefCount {}", id, objDetail->localRefCount);
        }
        objectDetailLock.unlock();
    }
    lock.unlock();

    auto result = std::make_pair(ErrorInfo(), std::vector<std::string>());
    if (!shouldIncreInDS.empty()) {
        YRLOG_DEBUG("ds increase id {}..., objs size {}", shouldIncreInDS[0], shouldIncreInDS.size());
        if (!remoteId.empty()) {
            result = dsObjectStore->IncreGlobalReference(shouldIncreInDS, remoteId);
        } else {
            auto err = dsObjectStore->IncreGlobalReference(shouldIncreInDS);
            if (!err.OK()) {
                YRLOG_ERROR("id [{}, ...] datasystem IncreGlobalReference failed. Code: {}, MCode: {}, Msg: {}",
                            shouldIncreInDS[0], fmt::underlying(err.Code()), fmt::underlying(err.MCode()), err.Msg());
                result.first = err;
            }
        }
        for (auto objDetail : increseObjectDetails) {
            std::unique_lock<std::mutex> objectDetailLock(objDetail->_mu);
            if (!result.first.OK()) {
                objDetail->increInDataSystemEnum = IncreInDataSystemEnum::NOT_INCREASE_IN_DS;
            } else {
                objDetail->increInDataSystemEnum = IncreInDataSystemEnum::INCREASE_IN_DS;
            }
            objDetail->notification.Notify();
            objectDetailLock.unlock();
        }
        if (!result.first.OK()) {
            YRLOG_WARN("increase global reference failed, ds increase id {}..., objs size is {}, remote id is {}",
                       shouldIncreInDS[0], shouldIncreInDS.size(), remoteId);
            return result;
        }
    }
    for (auto objDetail : waitObjectDetails) {
        objDetail->notification.WaitForNotificationWithTimeout(
            absl::Seconds(Config::Instance().DS_CONNECT_TIMEOUT_SEC()));
    }
    return result;
}

ErrorInfo MemoryStore::IncreDSGlobalReference(const std::vector<std::string> &objectIds)
{
    std::unique_lock<std::mutex> lock(mu);
    std::vector<std::string> shouldIncreInDS;
    std::vector<std::shared_ptr<ObjectDetail>> increaseObjectDetails;
    std::vector<std::shared_ptr<ObjectDetail>> waitObjectDetails;
    for (const std::string &id : objectIds) {
        auto [it, insertSuccess] = storeMap.emplace(id, std::make_shared<ObjectDetail>());
        (void)insertSuccess;
        std::shared_ptr<ObjectDetail> objDetail = it->second;
        std::unique_lock<std::mutex> objectDetailLock(objDetail->_mu);
        if (objDetail->increInDataSystemEnum == IncreInDataSystemEnum::INCREASING_IN_DS) {
            waitObjectDetails.push_back(objDetail);
        }
        if (objDetail->increInDataSystemEnum == IncreInDataSystemEnum::NOT_INCREASE_IN_DS) {
            // Not exist before, should Incre in DS
            shouldIncreInDS.push_back(id);
            objDetail->increInDataSystemEnum = IncreInDataSystemEnum::INCREASING_IN_DS;
            increaseObjectDetails.push_back(objDetail);
        }
        objectDetailLock.unlock();
    }
    lock.unlock();

    if (!shouldIncreInDS.empty()) {
        YRLOG_DEBUG("ds increase id {}..., objs size {}", shouldIncreInDS[0], shouldIncreInDS.size());
        auto err = dsObjectStore->IncreGlobalReference(shouldIncreInDS);

        for (auto objDetail : increaseObjectDetails) {
            std::unique_lock<std::mutex> objectDetailLock(objDetail->_mu);
            if (!err.OK()) {
                objDetail->increInDataSystemEnum = IncreInDataSystemEnum::NOT_INCREASE_IN_DS;
            } else {
                objDetail->increInDataSystemEnum = IncreInDataSystemEnum::INCREASE_IN_DS;
            }
            objDetail->notification.Notify();
            objectDetailLock.unlock();
        }

        if (!err.OK()) {
            YRLOG_ERROR("id [{}, ...] datasystem IncreGlobalReference failed. Code: {}, MCode: {}, Msg: {}",
                        shouldIncreInDS[0], fmt::underlying(err.Code()), fmt::underlying(err.MCode()), err.Msg());
            return err;
        }
    }

    for (auto detail : waitObjectDetails) {
        detail->notification.WaitForNotificationWithTimeout(absl::Seconds(Config::Instance().DS_CONNECT_TIMEOUT_SEC()));
    }
    return ErrorInfo();
}

// toDataSystem true: FORCE incre in datasystem; false: incre just in memory (known small object).
ErrorInfo MemoryStore::IncreGlobalReference(const std::vector<std::string> &objectIds, bool toDataSystem)
{
    return IncreaseGRefInMemoryAndDs(objectIds, toDataSystem).first;
}

std::pair<ErrorInfo, std::vector<std::string>> MemoryStore::IncreGlobalReference(
    const std::vector<std::string> &objectIds, const std::string &remoteId)
{
    return IncreaseGRefInMemoryAndDs(objectIds, true, remoteId);
}

std::vector<std::string> MemoryStore::DecreaseGRefInMemory(const std::vector<std::string> &objectIds)
{
    std::vector<std::string> shouldDecreInDS;
    {
        std::lock_guard<std::mutex> lock(mu);
        for (const std::string &id : objectIds) {
            YRLOG_TRACE("Decre id {}", id);
            if (id == "") {
                continue;
            }
            auto it = storeMap.find(id);
            if (it == storeMap.end()) {
                YRLOG_DEBUG("Decre id {} not exist in storeMap. Will force decre in DS.", id);
                shouldDecreInDS.push_back(id);
                continue;
            }
            std::shared_ptr<ObjectDetail> objDetail = it->second;
            std::unique_lock<std::mutex> objectDetailLock(objDetail->_mu);
            if (objDetail->localRefCount == 0) {
                ErrorInfo e;
                e.SetErrorCode(ErrorCode::ERR_PARAM_INVALID);
                e.SetErrorMsg("Decre an id " + id + " ref is 0 in storeMap.");
                continue;
            }
            objDetail->localRefCount--;
            if (objDetail->localRefCount == 0) {
                YRLOG_DEBUG("Decre id {} localRefCount {}", id, objDetail->localRefCount);
            } else {
                YRLOG_TRACE("Decre id {} localRefCount {}", id, objDetail->localRefCount);
            }
            if (objDetail->localRefCount == 0) {
                if (objDetail->storeInDataSystem ||
                    objDetail->increInDataSystemEnum == IncreInDataSystemEnum::INCREASE_IN_DS) {
                    YRLOG_DEBUG("Will Decre id {} in ds", id);
                    shouldDecreInDS.push_back(id);
                }
                // remove from storeMap
                storeMap.erase(it);
            }
        }
    }
    return shouldDecreInDS;
}

ErrorInfo MemoryStore::DecreGlobalReference(const std::vector<std::string> &objectIds)
{
    std::vector<std::string> shouldDecreInDS = DecreaseGRefInMemory(objectIds);
    // remove from datasystem
    if (!shouldDecreInDS.empty()) {
        return dsObjectStore->DecreGlobalReference(shouldDecreInDS);
    }
    return ErrorInfo();
}

std::pair<ErrorInfo, std::vector<std::string>> MemoryStore::DecreGlobalReference(
    const std::vector<std::string> &objectIds, const std::string &remoteId)
{
    std::vector<std::string> shouldDecreInDS = DecreaseGRefInMemory(objectIds);
    // remove from datasystem
    if (!shouldDecreInDS.empty()) {
        return dsObjectStore->DecreGlobalReference(shouldDecreInDS, remoteId);
    }
    return std::make_pair(ErrorInfo(), std::vector<std::string>());
}

std::vector<int> MemoryStore::QueryGlobalReference(const std::vector<std::string> &objectIds)
{
    std::vector<int> globalRef(objectIds.size());
    std::vector<std::string> shouldQueryFromDS;
    std::vector<size_t> shouldQueryFromDSIndex;
    size_t i = 0;
    {
        std::lock_guard<std::mutex> lock(mu);
        for (const std::string &id : objectIds) {
            auto it = storeMap.find(id);
            std::unique_lock<std::mutex> objectDetailLock;
            std::shared_ptr<ObjectDetail> objDetail;
            if (it != storeMap.end()) {
                objDetail = it->second;
                objectDetailLock = std::unique_lock<std::mutex>(objDetail->_mu);
            }
            // not in storeMap, or local ref sum is 0
            if (it == storeMap.end() || !(objDetail->storeInMemory) || objDetail->localRefCount == 0) {
                shouldQueryFromDS.push_back(id);
                shouldQueryFromDSIndex.push_back(i);
            } else {
                // Not really get from global
                globalRef[i] = objDetail->localRefCount;
            }
            i++;
        }
    }
    if (shouldQueryFromDS.empty()) {
        return globalRef;
    }
    std::vector<int> dsGlobalRef = dsObjectStore->QueryGlobalReference(shouldQueryFromDS);
    for (i = 0; i < dsGlobalRef.size(); i++) {
        globalRef[shouldQueryFromDSIndex[i]] = dsGlobalRef[i];
    }
    return globalRef;
}

ErrorInfo MemoryStore::ReleaseGRefs(const std::string &remoteId)
{
    return dsObjectStore->ReleaseGRefs(remoteId);
}

void MemoryStore::Clear()
{
    std::lock_guard<std::mutex> lock(mu);
    for (auto &it : storeMap) {
        if (it.second->increInDataSystemEnum == IncreInDataSystemEnum::INCREASE_IN_DS) {
            dsObjectStore->DecreGlobalReference({it.first});
        }
    }
    dsObjectStore->Clear();
    storeMap.clear();
}

ErrorInfo MemoryStore::DoPutToDS(const std::string &id, const CreateParam &createParam)
{
    std::unique_lock<std::mutex> lock(mu);
    auto it = storeMap.find(id);
    if (it == storeMap.end()) {
        YRLOG_DEBUG("id {} not exist in storeMap.", id);
        return ErrorInfo();
    }
    std::shared_ptr<ObjectDetail> objDetail = it->second;
    lock.unlock();  // unlock storeMap, and still lock objectDetail.
    std::unique_lock<std::mutex> objectDetailLock(objDetail->_mu);
    if (objDetail->storeInDataSystem) {
        YRLOG_DEBUG("id {} is already in datasystem.", id);
        return ErrorInfo();
    }
    if (!objDetail->storeInMemory) {
        YRLOG_DEBUG("id {} not store in mem.", id);
        return ErrorInfo();
    }
    if (objDetail->increInDataSystemEnum == IncreInDataSystemEnum::NOT_INCREASE_IN_DS) {
        ErrorInfo dsErr = dsObjectStore->IncreGlobalReference({id});
        objDetail->notification.Notify();
        if (dsErr.Code() != ErrorCode::ERR_OK) {
            YRLOG_ERROR("id {} datasystem IncreGlobalReference failed. Code: {}, MCode: {}, Msg: {}", id,
                        fmt::underlying(dsErr.Code()), fmt::underlying(dsErr.MCode()), dsErr.Msg());
            return dsErr;
        }
        objDetail->increInDataSystemEnum = IncreInDataSystemEnum::INCREASE_IN_DS;
    }
    if (objDetail->increInDataSystemEnum == IncreInDataSystemEnum::INCREASING_IN_DS) {
        if (!objDetail->notification.WaitForNotificationWithTimeout(
                absl::Seconds(Config::Instance().DS_CONNECT_TIMEOUT_SEC()))) {
            objDetail->increInDataSystemEnum = IncreInDataSystemEnum::NOT_INCREASE_IN_DS;
            YRLOG_ERROR("objid {} increase global failed", id);
        }
        objDetail->increInDataSystemEnum = IncreInDataSystemEnum::INCREASE_IN_DS;
    }
    YRLOG_DEBUG("try put id {} to dsObjectStore", id);
    ErrorInfo dsErr = dsObjectStore->Put(objDetail->data, id, std::unordered_set<std::string>(), createParam);
    if (dsErr.Code() != ErrorCode::ERR_OK) {
        YRLOG_ERROR("id {} datasystem Put failed. Code: {}, MCode: {}, Msg: {}", id, fmt::underlying(dsErr.Code()),
                    fmt::underlying(dsErr.MCode()), dsErr.Msg());
        dsObjectStore->DecreGlobalReference({id});
        return dsErr;
    }
    objDetail->storeInDataSystem = true;
    return ErrorInfo();
}

ErrorInfo MemoryStore::AlsoPutToDS(const std::string &id, const CreateParam &createParam)
{
    ErrorInfo err = DoPutToDS(id, createParam);
    if (err.Code() != ErrorCode::ERR_OK) {
        return err;
    }

    return ErrorInfo();
}

ErrorInfo MemoryStore::AlsoPutToDS(const std::unordered_set<std::string> &ids, const CreateParam &createParam)
{
    for (const std::string &id : ids) {
        ErrorInfo err = DoPutToDS(id, createParam);
        if (err.Code() != ErrorCode::ERR_OK) {
            return err;
        }
    }
    return ErrorInfo();
}

ErrorInfo MemoryStore::AlsoPutToDS(const std::vector<std::string> &ids, const CreateParam &createParam)
{
    for (const std::string &id : ids) {
        ErrorInfo err = DoPutToDS(id, createParam);
        if (err.Code() != ErrorCode::ERR_OK) {
            return err;
        }
    }
    return ErrorInfo();
}

ErrorInfo MemoryStore::IncreaseObjRef(const std::vector<std::string> &objectIds)
{
    std::unique_lock<std::mutex> lock(mu);
    for (auto &objectId : objectIds) {
        if (storeMap.find(objectId) == storeMap.end()) {
            YRLOG_DEBUG("id {} not exist in storeMap.", objectId);
            return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, ModuleCode::RUNTIME,
                             "id " + objectId + " not exist in storeMap");
        }
    }
    std::vector<std::string> objectIdsNeedIncre;
    std::vector<std::shared_ptr<ObjectDetail>> increseObjectDetails;
    std::vector<std::shared_ptr<ObjectDetail>> waitObjectDetails;
    for (auto &objectId : objectIds) {
        auto it = storeMap.find(objectId);
        std::shared_ptr<ObjectDetail> objDetail = it->second;
        std::unique_lock<std::mutex> objectDetailLock(objDetail->_mu);
        if (objDetail->increInDataSystemEnum == IncreInDataSystemEnum::INCREASING_IN_DS) {
            waitObjectDetails.push_back(objDetail);
        }
        if (objDetail->increInDataSystemEnum == IncreInDataSystemEnum::NOT_INCREASE_IN_DS) {
            objectIdsNeedIncre.push_back(objectId);
            increseObjectDetails.push_back(objDetail);
            objDetail->increInDataSystemEnum = IncreInDataSystemEnum::INCREASING_IN_DS;
        }
        it->second->localRefCount++;
        objectDetailLock.unlock();
    }
    lock.unlock();
    if (!objectIdsNeedIncre.empty()) {
        ErrorInfo dsErr = dsObjectStore->IncreGlobalReference(objectIdsNeedIncre);
        for (auto objectDetail : increseObjectDetails) {
            std::unique_lock<std::mutex> objectDetailLock(objectDetail->_mu);
            if (dsErr.Code() != ErrorCode::ERR_OK) {
                objectDetail->increInDataSystemEnum = IncreInDataSystemEnum::NOT_INCREASE_IN_DS;
            }
            objectDetail->notification.Notify();
            objectDetailLock.unlock();
        }
        if (dsErr.Code() != ErrorCode::ERR_OK) {
            YRLOG_ERROR("id [{}, ...] datasystem IncreGlobalReference failed. Code: {}, MCode: {}, Msg: {}",
                        objectIdsNeedIncre[0], fmt::underlying(dsErr.Code()), fmt::underlying(dsErr.MCode()),
                        dsErr.Msg());
            return dsErr;
        }
    }
    for (auto objectDetail : waitObjectDetails) {
        objectDetail->notification.WaitForNotificationWithTimeout(
            absl::Seconds(Config::Instance().DS_CONNECT_TIMEOUT_SEC()));
    }
    return ErrorInfo();
}

void MemoryStore::BindObjRefInReq(const std::string &requestId, const std::vector<std::string> &objectIds)
{
    std::unique_lock<std::mutex> lock(reqMu);
    reqObjStore.emplace(requestId, objectIds);
}

std::vector<std::string> MemoryStore::UnbindObjRefInReq(const std::string &requestId)
{
    std::lock_guard<std::mutex> lock(reqMu);
    std::vector<std::string> ids;
    auto it = reqObjStore.find(requestId);
    if (it != reqObjStore.end()) {
        ids = std::move(it->second);
        reqObjStore.erase(it);
    }
    return ids;
}

ErrorInfo MemoryStore::CreateBuffer(const std::string &objectId, size_t dataSize, std::shared_ptr<Buffer> &dataBuf,
                                    const CreateParam &createParam)
{
    return dsObjectStore->CreateBuffer(objectId, dataSize, dataBuf, createParam);
}

std::pair<ErrorInfo, std::shared_ptr<Buffer>> MemoryStore::GetBuffer(const std::string &id, int timeoutMS)
{
    std::vector<std::string> ids{id};
    auto [err, results] = GetBuffers(ids, timeoutMS);
    if (results.size() < ids.size()) {
        return std::make_pair(err, nullptr);
    }
    return std::make_pair(err, results[0]);
}

std::pair<ErrorInfo, std::vector<std::shared_ptr<Buffer>>> MemoryStore::GetBuffers(const std::vector<std::string> &ids,
                                                                                   int timeoutMS)
{
    std::vector<size_t> notInMemIndex;
    std::vector<std::string> vecGetFromDS;
    YR::Libruntime::ErrorInfo lastErr;
    std::vector<std::shared_ptr<Buffer>> result(ids.size());
    lastErr = this->GetBuffersFromMem(ids, result, notInMemIndex, vecGetFromDS);
    if (vecGetFromDS.size() > 0) {
        auto dsMultiRes = dsObjectStore->GetBuffers(vecGetFromDS, timeoutMS);
        if (dsMultiRes.first.Code() != ErrorCode::ERR_OK) {
            lastErr = dsMultiRes.first;
        }
        if (dsMultiRes.second.empty()) {
            return std::make_pair(lastErr, result);
        }
        for (size_t i = 0; i < vecGetFromDS.size(); i++) {
            int bufIndex = notInMemIndex[i];
            result[bufIndex] = dsMultiRes.second[i];
        }
    }
    return std::make_pair(lastErr, result);
}

ErrorInfo MemoryStore::GetBuffersFromMem(const std::vector<std::string> &ids,
                                         std::vector<std::shared_ptr<Buffer>> &result,
                                         std::vector<size_t> &notInMemIndex, std::vector<std::string> &vecGetFromDS)
{
    YR::Libruntime::ErrorInfo lastErr;
    {
        std::lock_guard<std::mutex> lock(mu);
        for (size_t i = 0; i < ids.size(); i++) {
            std::string id = ids[i];
            auto it = storeMap.find(id);
            if (it == storeMap.end()) {
                // Not tracked by MemoryStore. Directly fetch from DS.
                YRLOG_DEBUG("id {} not exist in storeMap. will Get from DS.", id);
                notInMemIndex.push_back(i);
                vecGetFromDS.push_back(id);
            } else {
                std::shared_ptr<ObjectDetail> objDetail = it->second;
                std::unique_lock<std::mutex> objectDetailLock(objDetail->_mu);
                if (objDetail->err.Code() != ErrorCode::ERR_OK) {
                    lastErr = objDetail->err;
                } else if (objDetail->storeInMemory) {
                    result[i] = objDetail->data;
                } else {
                    notInMemIndex.push_back(i);
                    vecGetFromDS.push_back(id);
                }
            }
        }
    }
    return lastErr;
}

std::pair<RetryInfo, std::vector<std::shared_ptr<Buffer>>> MemoryStore::GetBuffersWithoutRetry(
    const std::vector<std::string> &ids, int timeoutMS)
{
    YR::Libruntime::ErrorInfo lastErr;
    std::vector<size_t> notInMemIndex;
    std::vector<std::string> vecGetFromDS;
    std::vector<std::shared_ptr<Buffer>> result(ids.size());
    lastErr = this->GetBuffersFromMem(ids, result, notInMemIndex, vecGetFromDS);
    RetryInfo retryInfo;
    retryInfo.retryType = RetryType::UNLIMITED_RETRY;
    if (vecGetFromDS.size() > 0) {
        auto dsResult = dsObjectStore->GetBuffersWithoutRetry(vecGetFromDS, timeoutMS);
        if (dsResult.second.empty()) {
            retryInfo = dsResult.first;
            return std::make_pair(retryInfo, result);
        }
        for (size_t i = 0; i < vecGetFromDS.size(); i++) {
            int bufIndex = notInMemIndex[i];
            result[bufIndex] = dsResult.second[i];
        }
        retryInfo = dsResult.first;
        if (!dsResult.first.errorInfo.OK()) {
            lastErr = dsResult.first.errorInfo;
        }
    }
    retryInfo.errorInfo = lastErr;
    return std::make_pair(retryInfo, result);
}

bool MemoryStore::SetReady(const std::vector<DataObject> &objs)
{
    for (const auto &obj : objs) {
        if (!SetReady(obj.id)) {
            return false;
        }
    }
    return true;
}

bool MemoryStore::SetReady(const std::string &id)
{
    std::list<YR::Libruntime::ObjectReadyCallback> callbacks;
    std::list<YR::Libruntime::ObjectReadyCallbackWithData> callbacksWithData;
    std::shared_ptr<Buffer> data;
    {
        std::lock_guard<std::mutex> lock(mu);
        YRLOG_DEBUG("SetReady id {}.", id);
        auto it = storeMap.find(id);
        if (it == storeMap.end()) {
            YRLOG_DEBUG("id {} not exist in storeMap.", id);
            return false;
        }
        std::shared_ptr<ObjectDetail> objDetail = it->second;
        std::unique_lock<std::mutex> objectDetailLock(objDetail->_mu);
        if (objDetail->ready) {
            YRLOG_DEBUG("id {} more than once. Id is already READY.", id);
            return false;
        }
        objDetail->ready = true;
        callbacks = std::move(objDetail->callbacks);
        callbacksWithData = std::move(objDetail->callbacksWithData);
        data = objDetail->data;
    }
    waitingObjectManager->SetReady(id);
    for (ObjectReadyCallback &func : callbacks) {
        func(ErrorInfo());
    }
    auto err = ErrorInfo();
    if (!callbacksWithData.empty() && !data) {
        auto res = dsObjectStore->Get(id, -1);
        err = res.first;
        data = res.second;
    }
    for (auto &f : callbacksWithData) {
        f(err, data);
    }
    return true;
}

bool MemoryStore::SetError(const std::vector<DataObject> &objs, const ErrorInfo &err)
{
    for (const auto &obj : objs) {
        if (!SetError(obj.id, err)) {
            return false;
        }
    }
    return true;
}

bool MemoryStore::SetError(const std::string &id, const ErrorInfo &err)
{
    YRLOG_DEBUG("set id {}, msg: {}", id, err.Msg());
    std::list<YR::Libruntime::ObjectReadyCallback> callbacks;
    std::list<YR::Libruntime::ObjectReadyCallbackWithData> callbacksWithData;
    {
        std::lock_guard<std::mutex> lock(mu);
        auto it = storeMap.find(id);
        if (it == storeMap.end()) {
            YRLOG_DEBUG("id {} not exist in storeMap.", id);
            return false;
        }
        std::shared_ptr<ObjectDetail> objDetail = it->second;
        std::unique_lock<std::mutex> objectDetailLock(objDetail->_mu);
        if (objDetail->ready) {
            YRLOG_ERROR("id {} more than once. Id is already READY.", id);
            return false;
        }
        objDetail->ready = true;
        objDetail->err = err;
        callbacks = std::move(objDetail->callbacks);
        callbacksWithData = std::move(objDetail->callbacksWithData);
    }
    waitingObjectManager->SetError(id, err);
    for (ObjectReadyCallback &func : callbacks) {
        func(err);
    }
    for (auto &f : callbacksWithData) {
        f(err, nullptr);
    }
    return true;
}

bool MemoryStore::AddGenerator(const std::string &generatorId)
{
    std::unique_lock<std::mutex> lock(mu);
    auto [it, insertSuccess] = storeMap.emplace(generatorId, std::make_shared<ObjectDetail>());
    (void)it;
    return insertSuccess;
}

void MemoryStore::AddOutput(const std::string &generatorId, const std::string &objectId, uint64_t index,
                            const ErrorInfo &errInfo)
{
    std::unique_lock<std::mutex> lock(mu);
    auto it = storeMap.find(generatorId);
    if (it != storeMap.end()) {
        auto detail = it->second;
        lock.unlock();
        {
            std::unique_lock<std::mutex> objectDetailLock(detail->_mu);
            YRLOG_DEBUG(
                "start add object id into generator res map, id is {}, index is {}, err code is {}, err msg is {}",
                objectId, index, fmt::underlying(errInfo.Code()), errInfo.Msg());
            auto [iterator, insertSuccess] =
                detail->generatorResMap.emplace(index, GeneratorRes{.objectId = objectId, .err = errInfo});
            if (!insertSuccess) {
                YRLOG_WARN("duplicated add output of generator id: {}, object id: {}, index: {}", generatorId, objectId,
                           index);
            }
            (void)iterator;
        }
        detail->cv.notify_all();
    } else {
        YRLOG_WARN("generator id {} does not exist in store map, object id is {}, index is {}, ec is {}, msg is {}",
                   generatorId, objectId, index, fmt::underlying(errInfo.Code()), errInfo.Msg());
    }
}

void MemoryStore::GeneratorFinished(const std::string &generatorId)
{
    std::unique_lock<std::mutex> lock(mu);
    auto it = storeMap.find(generatorId);
    if (it != storeMap.end()) {
        auto detail = it->second;
        lock.unlock();
        {
            std::unique_lock<std::mutex> objectDetailLock(detail->_mu);
            detail->finished = true;
        }
        detail->cv.notify_all();
    }
}

std::pair<ErrorInfo, std::string> MemoryStore::GetOutput(const std::string &generatorId, bool blocking)
{
    std::unique_lock<std::mutex> lock(mu);
    auto it = storeMap.find(generatorId);
    if (it == storeMap.end()) {
        auto errMsg = "there is no info of generator: " + generatorId + ", please check the parameter";
        return std::make_pair(ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, ModuleCode::RUNTIME, errMsg), "");
    }
    auto detail = it->second;
    lock.unlock();

    std::unique_lock<std::mutex> objectDetailLock(detail->_mu);

    if (blocking) {
        detail->cv.wait(objectDetailLock, [&detail] {
            return detail->generatorResMap.find(detail->getIndex) != detail->generatorResMap.end() || detail->finished;
        });

        if (detail->generatorResMap.find(detail->getIndex) != detail->generatorResMap.end()) {
            auto res = detail->generatorResMap[detail->getIndex];
            YRLOG_DEBUG(
                "succeed to get generator res, res object id is {}, err code is {}, err msg is {}, index is {}, "
                "generator id is {}",
                res.objectId, fmt::underlying(res.err.Code()), res.err.Msg(), detail->getIndex.load(), generatorId);
            detail->getIndex++;
            return std::make_pair(res.err, res.objectId);
        }
        auto errMsg = "generator: " + generatorId + " has already end, but no result have been received";
        YRLOG_ERROR(errMsg);
        return std::make_pair(ErrorInfo(ErrorCode::ERR_GENERATOR_FINISHED, ModuleCode::RUNTIME, errMsg), "");
    }
    auto genObjectID = GenerateObjectId(generatorId, detail->getIndex);
    if (detail->generatorResMap.find(detail->getIndex) != detail->generatorResMap.end()) {
        // has received stream result
        auto res = detail->generatorResMap[detail->getIndex];
        YRLOG_DEBUG("{} has received", genObjectID);
        detail->getIndex++;
        return std::make_pair(res.err, res.objectId);
    }
    // has not received stream result and generator has finished
    if (detail->finished) {
        YRLOG_DEBUG("{} has finished", genObjectID);
        auto errMsg = "generator: " + generatorId + " has already end, but no result have been received";
        YRLOG_ERROR(errMsg);
        return std::make_pair(ErrorInfo(ErrorCode::ERR_GENERATOR_FINISHED, ModuleCode::RUNTIME, errMsg), "");
    }
    detail->getIndex++;
    objectDetailLock.unlock();

    AddReturnObject(genObjectID);
    YRLOG_DEBUG("{} peek not received", genObjectID);
    return std::make_pair(ErrorInfo(), genObjectID);
}

std::string MemoryStore::GenerateObjectId(const std::string &generatorId, uint64_t index)
{
    return "gen_" + generatorId + "_" + std::to_string(index);
}

bool MemoryStore::AddReadyCallback(const std::string &id, ObjectReadyCallback callback)
{
    std::unique_lock<std::mutex> lock(mu);
    auto it = storeMap.find(id);
    if (it == storeMap.end()) {
        lock.unlock();
        callback(ErrorInfo());
        YRLOG_WARN("id {} does not exist in storeMap, exec callback directly.", id);
        return false;
    }
    std::shared_ptr<ObjectDetail> objDetail = it->second;
    std::unique_lock<std::mutex> objectDetailLock(objDetail->_mu);
    if (!objDetail->err.OK()) {
        YRLOG_DEBUG("id {} already has exception.", id);
        lock.unlock();
        objectDetailLock.unlock();
        callback(objDetail->err);
        return false;
    }
    if (objDetail->ready || objDetail->storeInMemory || objDetail->storeInDataSystem) {
        YRLOG_DEBUG("id {} already READY. RDY {}, StoreInMem {}, StoreInDS {}.", id, objDetail->ready,
                    objDetail->storeInMemory, objDetail->storeInDataSystem);
        lock.unlock();
        objectDetailLock.unlock();
        callback(ErrorInfo());
        return false;
    }
    objDetail->callbacks.push_back(callback);
    return true;
}

bool MemoryStore::AddReadyCallbackWithData(const std::string &id, ObjectReadyCallbackWithData callback)
{
    std::unique_lock<std::mutex> lock(mu);
    auto it = storeMap.find(id);
    if (it == storeMap.end()) {
        lock.unlock();
        callback(ErrorInfo(), nullptr);
        YRLOG_WARN("id {} does not exist in storeMap, exec callback directly.", id);
        return false;
    }
    std::shared_ptr<ObjectDetail> objDetail = it->second;
    std::unique_lock<std::mutex> objectDetailLock(objDetail->_mu);
    if (!objDetail->err.OK()) {
        YRLOG_DEBUG("id {} already has exception.", id);
        lock.unlock();
        objectDetailLock.unlock();
        callback(objDetail->err, nullptr);
        return false;
    }
    if (objDetail->ready || objDetail->storeInMemory || objDetail->storeInDataSystem) {
        YRLOG_DEBUG("id {} already READY. RDY {}, StoreInMem {}, StoreInDS {}.", id, objDetail->ready,
                    objDetail->storeInMemory, objDetail->storeInDataSystem);
        std::shared_ptr<Buffer> data;
        ErrorInfo err;
        if (objDetail->storeInMemory) {
            data = objDetail->data;
        } else {
            auto res = dsObjectStore->Get(id, -1);
            err = res.first;
            data = res.second;
        }
        lock.unlock();
        objectDetailLock.unlock();
        callback(err, data);
        return false;
    }
    objDetail->callbacksWithData.push_back(callback);
    return true;
}

void MemoryStore::AddEventCallbackWithData(const std::string &id, EventCallbackWithData callback)
{
    std::deque<std::shared_ptr<std::string>> tempBuffers;
    {
        std::unique_lock<std::mutex> lock(eventMu);
        auto it = reqEventCallbackMap_.find(id);
        if (it == reqEventCallbackMap_.end()) {
            EventCallbackInfo eventCallbackInfo;
            eventCallbackInfo.eventCallback = callback;
            reqEventCallbackMap_.emplace(id, eventCallbackInfo);
        } else {
            it->second.eventCallback = callback;
            tempBuffers = std::move(it->second.buffers);
        }
    }
    for (const auto &eventBuffer : tempBuffers) {
        this->DoEventCallback(callback, id, *eventBuffer);
    }
}

bool MemoryStore::AddEventData(const std::string &reqId, const std::string &eventData)
{
    YRLOG_DEBUG("Begin to add event data, size is {}, reqId is {}", eventData.size(), reqId);
    std::unique_lock<std::mutex> lock(eventMu);
    // if the callback function has not been registered, the data will be cached in the buffers,
    // when AddEventCallbackWithData method is called, the buffers will be called back in sequence.
    auto it = reqEventCallbackMap_.find(reqId);
    if (it == reqEventCallbackMap_.end()) {
        EventCallbackInfo eventCallbackInfo;
        eventCallbackInfo.buffers.push_back(std::make_shared<std::string>(eventData));
        reqEventCallbackMap_.emplace(reqId, eventCallbackInfo);
        return false;
    }
    if (it->second.eventCallback == nullptr) {
        it->second.buffers.push_back(std::make_shared<std::string>(eventData));
        return false;
    }
    auto callback = it->second.eventCallback;
    lock.unlock();
    DoEventCallback(callback, reqId, eventData);
    return true;
}

void MemoryStore::DoEventCallback(EventCallbackWithData callback, const std::string &reqId,
                                  const std::string &eventData)
{
    std::shared_ptr<Buffer> eventDataBuf = std::make_shared<NativeBuffer>(eventData.size());
    eventDataBuf->MemoryCopy(eventData.data(), eventData.size());
    callback(ErrorInfo(), eventDataBuf);
    if (eventData == YR::Libruntime::EVENT_EOF) {
        YRLOG_DEBUG("Receive event EOF, begin to erase eventCallback.");
        std::unique_lock<std::mutex> lock(eventMu);
        auto it = reqEventCallbackMap_.find(reqId);
        if (it != reqEventCallbackMap_.end()) {
            if (it->second.timer != nullptr) {
                it->second.timer->cancel();
            }
            reqEventCallbackMap_.erase(it);
        }
    }
}

void MemoryStore::AddEventTimer(const std::string &reqId, int timeoutSec)
{
    auto eventTimer = timerWorker_->CreateTimer(
        timeoutSec, EVENT_TIMER_EXEC_TIMES, [this, reqId] { this->AddEventData(reqId, YR::Libruntime::EVENT_EOF); });
    std::unique_lock<std::mutex> lock(eventMu);
    auto it = reqEventCallbackMap_.find(reqId);
    if (it == reqEventCallbackMap_.end()) {
        EventCallbackInfo eventCallbackInfo;
        eventCallbackInfo.timer = eventTimer;
        reqEventCallbackMap_.emplace(reqId, eventCallbackInfo);
    } else {
        if (it->second.timer != nullptr) {
            it->second.timer->cancel();
        }
        it->second.timer = eventTimer;
    }
}

bool MemoryStore::AddReturnObject(const std::vector<DataObject> &returnObjs)
{
    for (const auto &obj : returnObjs) {
        auto ret = AddReturnObject(obj.id);
        if (!ret) {
            YRLOG_WARN("obj id already exist in storeMap, id is {}", obj.id);
            return false;
        }
    }
    return true;
}

bool MemoryStore::AddReturnObject(const std::string &objId)
{
    {
        std::lock_guard<std::mutex> lock(mu);
        auto [it, insertSuccess] = storeMap.emplace(objId, std::make_shared<ObjectDetail>());
        std::shared_ptr<ObjectDetail> objDetail = it->second;
        std::unique_lock<std::mutex> objectDetailLock(objDetail->_mu);
        objDetail->localRefCount++;
        if (!insertSuccess) {
            return false;
        }
        objDetail->ready = false;
    }
    waitingObjectManager->SetUnready(objId);
    return true;
}

// Here haven't Incre Ref
bool MemoryStore::SetInstanceId(const std::string &id, const std::string &instanceId)
{
    return SetInstanceIds(id, {instanceId});
}

bool MemoryStore::SetInstanceIds(const std::string &id, const std::vector<std::string> &instanceIds)
{
    std::lock_guard<std::mutex> lock(mu);
    auto it = storeMap.find(id);
    if (it == storeMap.end()) {
        return false;
    }
    std::shared_ptr<ObjectDetail> objDetail = it->second;
    std::unique_lock<std::mutex> objectDetailLock(objDetail->_mu);
    try {
        objDetail->instanceIds.set_value(instanceIds);
    } catch (const std::future_error &e) {
        YRLOG_DEBUG("has already set value of objid : {}", id);
    }
    return true;
}

std::string MemoryStore::GetInstanceId(const std::string &objId)
{
    return GetInstanceIds(objId, NO_TIMEOUT).first[0];
}

std::pair<std::vector<std::string>, ErrorInfo> MemoryStore::GetInstanceIds(const std::string &objId, int timeoutSec)
{
    std::shared_future<std::vector<std::string>> f;
    std::vector<std::string> retInstances;
    std::shared_ptr<ObjectDetail> objDetail;
    {
        std::lock_guard<std::mutex> lock(mu);
        auto it = storeMap.find(objId);
        if (it == storeMap.end()) {
            std::string msg = "objId " + objId + " does not exist in storeMap.";
            YRLOG_INFO("{} Return objId as instanceId.", msg);
            retInstances.push_back(objId);
            return std::make_pair(retInstances, ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, ModuleCode::RUNTIME, msg));
        }
        objDetail = it->second;
        std::unique_lock<std::mutex> objectDetailLock(objDetail->_mu);
        f = objDetail->instanceIdsFuture;
    }
    if (timeoutSec != NO_TIMEOUT && f.wait_for(std::chrono::seconds(timeoutSec)) != std::future_status::ready) {
        std::string msg = "get instances timeout, failed objectID: " + objId + ".";
        YRLOG_ERROR(msg);
        return std::make_pair(retInstances, ErrorInfo(ErrorCode::ERR_GET_OPERATION_FAILED, ModuleCode::RUNTIME, msg));
    }
    return std::make_pair(f.get(), objDetail->err);
}

bool MemoryStore::SetInstanceRoute(const std::string &id, const std::string &instanceRoute)
{
    std::lock_guard<std::mutex> lock(mu);
    auto it = storeMap.find(id);
    if (it == storeMap.end()) {
        return false;
    }
    std::shared_ptr<ObjectDetail> objDetail = it->second;
    std::unique_lock<std::mutex> objectDetailLock(objDetail->_mu);
    try {
        objDetail->instanceRoute.set_value(instanceRoute);
    } catch (const std::future_error &e) {
        YRLOG_DEBUG("has already set value of objid : {}", id);
    }
    return true;
}

std::string MemoryStore::GetInstanceRoute(const std::string &objId, int timeoutSec)
{
    std::shared_future<std::string> f;
    std::string retInstanceRoute;
    std::shared_ptr<ObjectDetail> objDetail;
    {
        std::lock_guard<std::mutex> lock(mu);
        auto it = storeMap.find(objId);
        if (it == storeMap.end()) {
            std::string msg = "objId " + objId + " does not exist in storeMap.";
            YRLOG_INFO("{} Return empty string as instanceRoute.", msg);
            return retInstanceRoute;
        }
        objDetail = it->second;
        std::unique_lock<std::mutex> objectDetailLock(objDetail->_mu);
        f = objDetail->instanceRouteFuture;
    }
    if (timeoutSec != NO_TIMEOUT && f.wait_for(std::chrono::seconds(timeoutSec)) != std::future_status::ready) {
        if (timeoutSec != ZERO_TIMEOUT) {
            YRLOG_WARN("get instance route timeout, return empty string as instanceRoute. objectID is: {}.", objId);
        }
        return retInstanceRoute;
    }
    return f.get();
}

// Get the Last ErrorInfo of an object. Default is empty.
ErrorInfo MemoryStore::GetLastError(const std::string &objId)
{
    std::lock_guard<std::mutex> lock(mu);
    auto it = storeMap.find(objId);
    if (it == storeMap.end()) {
        YRLOG_ERROR("objId {} does not exist in storeMap. Return default empty ErrorInfo.", objId);
        return ErrorInfo();
    }
    std::shared_ptr<ObjectDetail> objDetail = it->second;
    std::unique_lock<std::mutex> objectDetailLock(objDetail->_mu);
    return objDetail->err;
}

bool MemoryStore::IsReady(const std::string &objId)
{
    std::unique_lock<std::mutex> lock(mu);
    auto it = storeMap.find(objId);
    if (it == storeMap.end()) {
        YRLOG_ERROR("objId {} does not exist in storeMap", objId);
        return false;
    }
    std::shared_ptr<ObjectDetail> objDetail = it->second;
    lock.unlock();
    std::unique_lock<std::mutex> objectDetailLock(objDetail->_mu);
    return objDetail->ready;
}

bool MemoryStore::IsExistedInLocal(const std::string &objId)
{
    std::lock_guard<std::mutex> lock(mu);
    auto it = storeMap.find(objId);
    if (it == storeMap.end()) {
        return false;
    }
    return true;
}

}  // namespace Libruntime
}  // namespace YR
