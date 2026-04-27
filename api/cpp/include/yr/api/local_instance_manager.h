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
#include <mutex>
#include <unordered_map>

#include <boost/any.hpp>

#include "yr/api/exception.h"
#include "yr/api/object_ref.h"
namespace YR {
namespace internal {
const std::uint64_t FETCH_INTERVAL_US = 5000;  // 5ms
const std::uint64_t RETRYNUM = 1000;

template <typename T>
class LocalInstanceManager {
public:
    static LocalInstanceManager &Singleton()
    {
        static LocalInstanceManager InstanceMgr;
        return InstanceMgr;
    }

    /* Note: Instance must exist */
    void DelLocalInstance(const std::string &instanceId)
    {
        std::lock_guard<std::mutex> lock(InstanceObjMtx);
        auto iter = InstanceObjMap.find(instanceId);
        if (iter != InstanceObjMap.end()) {
            InstanceObjMap.erase(iter);
        }
    }

    uint8_t *GetLocalInstance(const std::string &instanceId)
    {
        int retryNum = RETRYNUM;
        while (retryNum--) {
            InstanceObjMtx.lock();
            auto objIter = InstanceObjMap.find(instanceId);
            if (objIter != InstanceObjMap.end()) {
                ObjectRef<T> result = boost::any_cast<ObjectRef<T>>(objIter->second);
                InstanceObjMtx.unlock();
                uint8_t *instance = (uint8_t *)YR::internal::GetLocalModeRuntime()->Get(result, -1).get();
                return instance;
            }
            InstanceObjMtx.unlock();
            usleep(FETCH_INTERVAL_US);
        }
        std::string msg = "YR_INVOKE instance is empty, instanceId: " + instanceId;
        throw YR::Exception::InvalidParamException(msg);
    }

    void SetResult(const std::string &instanceId, ObjectRef<T> &&res)
    {
        std::unique_lock<std::mutex> lk(InstanceObjMtx);
        auto happened = InstanceObjMap.emplace(instanceId, res).second;
        if (!happened) {
            std::string msg = "YR_INVOKE instance is duplicated, instanceId: " + instanceId;
            throw YR::Exception::InvalidParamException(msg);
        }
    }

private:
    LocalInstanceManager() = default;
    ~LocalInstanceManager() = default;

    std::mutex InstanceObjMtx;
    std::unordered_map<std::string, boost::any> InstanceObjMap;
};
}  // namespace internal

}  // namespace YR