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
#include <future>
#include "wait_request_manager.h"
#include "yr/api/exception.h"
#include "yr/api/local_state_store.h"
#include "yr/api/object_ref.h"
namespace YR {

namespace utility {
class ThreadPool;
}
namespace internal {
class LocalModeRuntime {
public:
    LocalModeRuntime() = default;
    ~LocalModeRuntime() = default;
    void Init();
    void Stop();
    template <typename T>
    ObjectRef<T> Put(const T &val);

    template <typename T>
    bool Wait(const ObjectRef<T> &obj, int timeout);

    template <typename T>
    std::vector<bool> Wait(const std::vector<ObjectRef<T>> &objs, std::size_t waitNum, int timeout);

    template <typename T>
    std::shared_ptr<T> Get(const ObjectRef<T> &obj, int timeout);

    template <typename T>
    std::vector<std::shared_ptr<T>> Get(const std::vector<ObjectRef<T>> &objs, int timeout, bool allowPartial = false);

    // KV
    void KVWrite(const std::string &key, std::shared_ptr<msgpack::sbuffer> value, ExistenceOpt existence);

    void KVMSetTx(const std::vector<std::string> &keys, const std::vector<std::shared_ptr<msgpack::sbuffer>> &vals,
                  ExistenceOpt existence);

    std::shared_ptr<Buffer> KVRead(const std::string &key, int timeoutMs);

    std::vector<std::shared_ptr<Buffer>> KVRead(const std::vector<std::string> &keys, int timeoutMs,
                                                bool allowPartial = false);

    void KVDel(const std::string &key);

    std::vector<std::string> KVDel(const std::vector<std::string> &keys);

    std::vector<bool> KVExist(const std::vector<std::string> &keys);

    template <typename T>
    bool IsAllFail(const std::vector<std::shared_ptr<T>> &results);

    template <typename T>
    ObjectRef<T> PutFuture(std::shared_future<std::shared_ptr<T>> &&fut);

    std::string GenerateObjId();

    void LocalSubmit(std::function<void()> &&func);

    bool SetReady(const std::string &id);

    void SetException(const std::string &id, const std::exception_ptr &exceptionPtr);

private:
    size_t threads_;
    std::shared_ptr<WaitRequestManager> waitRequestManager_;
    std::shared_ptr<LocalStateStore> stateStore_;
    std::shared_ptr<YR::utility::ThreadPool> pool_;
    std::atomic<bool> initPool_{false};
};

template <typename T>
ObjectRef<T> LocalModeRuntime::PutFuture(std::shared_future<std::shared_ptr<T>> &&fut)
{
    std::string id_ = GenerateObjId();
    auto obj = ObjectRef<T>(id_, false, true);
    obj.PutFuture(std::move(fut));
    return obj;
}

template <typename T>
ObjectRef<T> LocalModeRuntime::Put(const T &val)
{
    std::string id_ = GenerateObjId();
    auto obj = ObjectRef<T>(id_, false, true);
    obj.Put(val);
    return obj;
}

template <typename T>
std::vector<bool> LocalModeRuntime::Wait(const std::vector<ObjectRef<T>> &objs, std::size_t waitNum, int timeout)
{
    auto isReady = [](const ObjectRef<T> &obj) -> bool { return obj.IsReady(); };
    auto getId = [](const ObjectRef<T> &obj) -> std::string { return obj.ID(); };
    return waitRequestManager_->Wait<ObjectRef<T>>(objs, waitNum, timeout, isReady, getId);
}

template <typename T>
bool LocalModeRuntime::Wait(const ObjectRef<T> &obj, int timeout)
{
    return obj.Wait(timeout);
}

template <typename T>
std::shared_ptr<T> LocalModeRuntime::Get(const ObjectRef<T> &obj, int timeout)
{
    return obj.Get(timeout, false);
}

template <typename T>
std::vector<std::shared_ptr<T>> LocalModeRuntime::Get(const std::vector<ObjectRef<T>> &objs, int timeout,
                                                      bool allowPartial)
{
    int originalTimeout = timeout;
    std::vector<std::shared_ptr<T>> results(objs.size());
    for (std::size_t i = 0; i < objs.size(); ++i) {
        auto start = std::chrono::high_resolution_clock::now();
        results[i] = objs[i].Get(timeout, allowPartial);
        auto end = std::chrono::high_resolution_clock::now();
        timeout = (timeout == -1) ? timeout
                                  : (timeout - std::chrono::duration_cast<std::chrono::seconds>(end - start).count());
        if (timeout != -1 && timeout < 0) {
            if (!allowPartial) {
                std::string msg = "WaitFor wait result timeout -- " + std::to_string(originalTimeout);
                // do not allow return partial successful result
                throw YR::Exception::InvalidParamException(msg);
            }
            //  allow return partial successful result
            if (!IsAllFail(results)) {
                return results;
            }
            // get all fail
            std::string msg = "All objectRef get failed";
            throw YR::Exception::InvalidParamException(msg);
        }
    }
    return results;
}

template <typename T>
bool LocalModeRuntime::IsAllFail(const std::vector<std::shared_ptr<T>> &results)
{
    for (std::size_t i = 0; i < results.size(); ++i) {
        if (results[i] != nullptr) {
            return false;
        }
    }
    return true;
}
}  // namespace internal

}  // namespace YR
