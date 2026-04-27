/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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

#ifdef ENABLE_GLOO

#include "gloo_collective_group.h"

#include "api/cpp/src/utils/utils.h"
#include "gloo/algorithm.h"
#include "gloo/allgather.h"
#include "gloo/barrier.h"
#include "gloo/broadcast.h"
#include "gloo/reduce.h"
#include "gloo/rendezvous/prefix_store.h"
#include "gloo/scatter.h"
#include "gloo/transport/ibverbs/device.h"
#include "gloo/transport/tcp/device.h"
#include "src/dto/config.h"
#include "yr/api/kv_manager.h"

namespace YR::Collective {

#define EXECUTE_BY_TYPE(dType, OPERATION, ...)                                \
    switch (dType) {                                                          \
        case DataType::INT:                                                   \
            OPERATION<int>(__VA_ARGS__);                                      \
            break;                                                            \
        case DataType::DOUBLE:                                                \
            OPERATION<double>(__VA_ARGS__);                                   \
            break;                                                            \
        case DataType::LONG:                                                  \
            OPERATION<long>(__VA_ARGS__);                                     \
            break;                                                            \
        case DataType::FLOAT:                                                 \
            OPERATION<float>(__VA_ARGS__);                                    \
            break;                                                            \
        case DataType::INVALID:                                               \
        default:                                                              \
            throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, \
                                "invalid dType: " + std::to_string(dType));   \
    }

const int DS_STORE_CHECK_INTERVAL = 100;  // ms

std::string ReplaceDsStoreKey(const std::string &str)
{
    std::string editedKey = str;
    // datasystem doesn't support key that contains '/', replace it with '@'
    std::replace(editedKey.begin(), editedKey.end(), '/', '@');
    return editedKey;
}

DsStore::~DsStore()
{
    clear();
}

void DsStore::set(const std::string &key, const std::vector<char> &data)
{
    auto str = std::string(data.begin(), data.end());
    try {
        YR::KVManager::Set(ReplaceDsStoreKey(key), std::string(data.begin(), data.end()));
    } catch (YR::Exception &e) {
        throw YR::Exception(Libruntime::ErrorCode::ERR_USER_FUNCTION_EXCEPTION,
                            "collective group " + groupName_ + " failed to execute, what: " + e.what());
    }
    keys_.emplace(ReplaceDsStoreKey(key));
    std::this_thread::sleep_for(std::chrono::milliseconds(DS_STORE_CHECK_INTERVAL));
}

std::vector<char> DsStore::get(const std::string &key)
{
    std::string result;
    try {
        result = YR::KVManager::Get(ReplaceDsStoreKey(key));
    } catch (YR::Exception &e) {
        throw YR::Exception(Libruntime::ErrorCode::ERR_USER_FUNCTION_EXCEPTION,
                            "collective group " + groupName_ + " failed to execute, what: " + e.what());
    }
    std::vector<char> out;
    out.assign(result.begin(), result.end());
    return out;
}

void DsStore::wait(const std::vector<std::string> &keys)
{
    std::vector<std::string> editedKeys;
    for (const auto &key : keys) {
        editedKeys.push_back(ReplaceDsStoreKey(key));
    }

    try {
        YR::KVManager::Get(editedKeys);
    } catch (YR::Exception &e) {
        throw YR::Exception(Libruntime::ErrorCode::ERR_USER_FUNCTION_EXCEPTION,
                            "collective group " + groupName_ + " failed to execute, what: " + e.what());
    }
}

void DsStore::wait(const std::vector<std::string> &keys, const std::chrono::milliseconds &timeout)
{
    std::vector<std::string> editedKeys;
    for (const auto &key : keys) {
        editedKeys.push_back(ReplaceDsStoreKey(key));
    }
    // Convert timeout to seconds for KVManager, ensuring at least 1s if timeout > 0
    int timeoutSec = static_cast<int>(timeout.count() / S_TO_MS);
    if (timeout.count() > 0 && timeoutSec == 0) {
        timeoutSec = 1;
    }

    try {
        YR::KVManager::Get(editedKeys, timeoutSec);
    } catch (YR::Exception &e) {
        throw YR::Exception(Libruntime::ErrorCode::ERR_USER_FUNCTION_EXCEPTION,
                            "collective group " + groupName_ + " failed to execute, what: " + e.what());
    }
}

void DsStore::clear()
{
    // clear ds kv
    std::vector<std::string> dels(keys_.begin(), keys_.end());
    YR::KVManager::Del(dels);
}

GlooCollectiveGroup::GlooCollectiveGroup(const CollectiveGroupSpec &groupSpec, int rank, std::string storePrefix)
    : CollectiveGroup(groupSpec, rank, std::move(storePrefix))
{
    auto dsStore = std::make_shared<DsStore>(groupName_);
    std::string prefixKey = groupName_;
    if (!storePrefix_.empty()) {
        prefixKey = groupName_ + "-" + storePrefix_;
    }
    store_ = dsStore;
    auto prefixStore = std::make_shared<gloo::rendezvous::PrefixStore>(prefixKey, dsStore);
    std::shared_ptr<gloo::transport::Device> dev;
    auto backend = GetEnv("GLOO_BACKEND_TYPE");

    // Enable Gloo internal logs
    setenv("GLOO_LOG_LEVEL", "INFO", 1);
    if (backend.empty() || backend == "TCP") {
        gloo::transport::tcp::attr attr;
        attr.hostname = YR::Libruntime::Config::Instance().HOST_IP();
        std::cout << "[DEBUG] Gloo TCP Device binding to HOST_IP: " << attr.hostname << std::endl;
        dev = gloo::transport::tcp::CreateDevice(attr);
    } else if (backend == "IBVERBS") {
        gloo::transport::ibverbs::attr attr;
        attr.name = GetEnv("GLOO_IBVERBS_NAME");

        auto indexStr = GetEnv("GLOO_IBVERBS_INDEX");
        auto portStr = GetEnv("GLOO_IBVERBS_PORT");
        try {
            attr.index = indexStr.empty() ? 0 : std::stoi(indexStr);
            attr.port = portStr.empty() ? 0 : std::stoi(portStr);
        } catch (const std::exception &e) {
            throw YR::Exception(
                YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                "invalid GLOO_IBVERBS_INDEX: " + indexStr + ", or invalid GLOO_IBVERBS_PORT: " + portStr);
        }
        dev = gloo::transport::ibverbs::CreateDevice(attr);
    } else {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "invalid backend type: " + backend);
    }

    context_ = std::make_shared<gloo::rendezvous::Context>(rank_, worldSize_);
    context_->setTimeout(std::chrono::milliseconds(timeout_));
    context_->connectFullMesh(prefixStore, dev);
}

GlooCollectiveGroup::~GlooCollectiveGroup()
{
    if (store_) {
        store_->clear();
    }
}

void GlooCollectiveGroup::AllReduce(const void *sendbuf, void *recvbuf, int count, DataType dtype, const ReduceOp &op)
{
    EXECUTE_BY_TYPE(dtype, DoAllReduce, sendbuf, recvbuf, count, op);
}

void GlooCollectiveGroup::Reduce(const void *sendbuf, void *recvbuf, int count, DataType dtype, const ReduceOp &op,
                                 int dstRank)
{
    EXECUTE_BY_TYPE(dtype, DoReduce, sendbuf, recvbuf, count, op, dstRank);
}

void GlooCollectiveGroup::AllGather(const void *sendbuf, void *recvbuf, int count, DataType dtype)
{
    EXECUTE_BY_TYPE(dtype, DoAllGather, sendbuf, recvbuf, count);
}

void GlooCollectiveGroup::Barrier()
{
    std::lock_guard<std::recursive_mutex> lock(mtx_);
    gloo::BarrierOptions opts(context_);
    gloo::barrier(opts);
}

void GlooCollectiveGroup::Scatter(const std::vector<void *> sendbuf, void *recvbuf, int count, DataType dtype,
                                  int srcRank)
{
    EXECUTE_BY_TYPE(dtype, DoScatter, sendbuf, recvbuf, count, srcRank);
}

void GlooCollectiveGroup::Broadcast(const void *sendbuf, void *recvbuf, int count, DataType dtype, int srcRank)
{
    EXECUTE_BY_TYPE(dtype, DoBroadcast, sendbuf, recvbuf, count, srcRank);
}

void GlooCollectiveGroup::Recv(void *recvbuf, int count, DataType dtype, int srcRank, int tag)
{
    ThrowIfTrue(DATA_TYPE_SIZE_MAP.find(dtype) == DATA_TYPE_SIZE_MAP.end(),
                YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "invalid dtype: " + std::to_string(dtype));
    std::lock_guard<std::recursive_mutex> lock(mtx_);
    auto ubuf = context_->createUnboundBuffer(recvbuf, count * DATA_TYPE_SIZE_MAP.at(dtype));
    ubuf->recv(srcRank, tag);
    ubuf->waitRecv(std::chrono::milliseconds(timeout_));
}

void GlooCollectiveGroup::Send(const void *sendbuf, int count, DataType dtype, int dstRank, int tag)
{
    ThrowIfTrue(DATA_TYPE_SIZE_MAP.find(dtype) == DATA_TYPE_SIZE_MAP.end(),
                YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "invalid dtype: " + std::to_string(dtype));
    std::lock_guard<std::recursive_mutex> lock(mtx_);
    auto ubuf = context_->createUnboundBuffer(const_cast<void *>(sendbuf), count * DATA_TYPE_SIZE_MAP.at(dtype));
    ubuf->send(dstRank, tag);
    ubuf->waitSend(std::chrono::milliseconds(timeout_));
}

template <typename T>
gloo::AllreduceOptions::Func GlooCollectiveGroup::GetReduceOp(const ReduceOp &op)
{
    void (*fn)(void *, const void *, const void *, long unsigned int) = &gloo::sum<T>;
    switch (op) {
        case ReduceOp::SUM:
            fn = &gloo::sum<T>;
            break;
        case ReduceOp::PRODUCT:
            fn = &gloo::product<T>;
            break;
        case ReduceOp::MIN:
            fn = &gloo::min<T>;
            break;
        case ReduceOp::MAX:
            fn = &gloo::max<T>;
            break;
    }
    return fn;
}

template <typename T>
void GlooCollectiveGroup::DoAllReduce(const void *sendbuf, void *recvbuf, int count, const ReduceOp &op)
{
    std::lock_guard<std::recursive_mutex> lock(mtx_);
    gloo::AllreduceOptions opts(context_);
    opts.setInput(const_cast<T *>((T *)sendbuf), count);
    opts.setOutput(static_cast<T *>(recvbuf), count);
    opts.setReduceFunction(GetReduceOp<T>(op));
    gloo::allreduce(opts);
}

template <typename T>
void GlooCollectiveGroup::DoReduce(const void *sendbuf, void *recvbuf, int count, const ReduceOp &op, int dstRank)
{
    std::lock_guard<std::recursive_mutex> lock(mtx_);
    gloo::ReduceOptions opts(context_);
    opts.setInput(const_cast<T *>((T *)sendbuf), count);
    opts.setOutput(static_cast<T *>(recvbuf), count);
    opts.setRoot(dstRank);
    opts.setReduceFunction(GetReduceOp<T>(op));
    gloo::reduce(opts);
}

template <typename T>
void GlooCollectiveGroup::DoBroadcast(const void *sendbuf, void *recvbuf, int count, int srcRank)
{
    std::lock_guard<std::recursive_mutex> lock(mtx_);
    gloo::BroadcastOptions opts(context_);
    if (srcRank == GetRank()) {
        opts.setInput(const_cast<T *>((T *)sendbuf), count);
    }

    opts.setOutput(static_cast<T *>(recvbuf), count);
    opts.setRoot(srcRank);
    gloo::broadcast(opts);
}

template <typename T>
void GlooCollectiveGroup::DoScatter(const std::vector<void *> sendbuf, void *recvbuf, int count, int srcRank)
{
    std::lock_guard<std::recursive_mutex> lock(mtx_);
    gloo::ScatterOptions opts(context_);
    if (srcRank == GetRank()) {
        std::vector<T *> inputs;
        for (auto item : sendbuf) {
            inputs.push_back(const_cast<T *>((T *)item));
        }
        opts.setInputs<T>(inputs, count);
    }

    opts.setOutput(static_cast<T *>(recvbuf), count);
    opts.setRoot(srcRank);
    gloo::scatter(opts);
}

template <typename T>
void GlooCollectiveGroup::DoAllGather(const void *sendbuf, void *recvbuf, int count)
{
    std::lock_guard<std::recursive_mutex> lock(mtx_);
    gloo::AllgatherOptions opts(context_);
    opts.setInput(const_cast<T *>((T *)sendbuf), count);
    opts.setOutput(static_cast<T *>(recvbuf), count * GetWorldSize());
    gloo::allgather(opts);
}
}  // namespace YR::Collective

#endif  // ENABLE_GLOO