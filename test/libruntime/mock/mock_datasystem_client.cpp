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

#include "mock_datasystem_client.h"

#include <cstdlib>
#include <msgpack.hpp>
#include <string>
#include <unordered_set>

#define private public
#include "datasystem/hetero_client.h"
#include "datasystem/kv_client.h"
#include "datasystem/object_client.h"
#include "datasystem/stream_client.h"

namespace datasystem {
class ThreadPool {};

StreamClient::StreamClient(ConnectOptions options) {}

Status StreamClient::Init(bool reportWorkerLost)
{
    return Status::OK();
}

Status StreamClient::CreateProducer(const std::string &streamName, std::shared_ptr<Producer> &outProducer,
                                    ProducerConf producerConf)
{
    return Status::OK();
}

Status StreamClient::Subscribe(const std::string &streamName, const struct SubscriptionConfig &config,
                               std::shared_ptr<Consumer> &outConsumer, bool autoAck)
{
    // autoAck default is false
    return Status::OK();
}

Status StreamClient::DeleteStream(const std::string &streamName)
{
    return Status::OK();
}

Status StreamClient::QueryGlobalProducersNum(const std::string &streamName, uint64_t &gProducerNum)
{
    return Status::OK();
}

Status StreamClient::QueryGlobalConsumersNum(const std::string &streamName, uint64_t &gConsumerNum)
{
    return Status::OK();
}

Status StreamClient::ShutDown()
{
    return Status::OK();
}

Status Producer::Send(const Element &element)
{
    return Status::OK();
}

Status Producer::Send(const Element &element, int64_t timeoutMs)
{
    return Status::OK();
}

Status Producer::Close()
{
    return Status::OK();
}

Status Consumer::Receive(uint32_t expectNum, uint32_t timeoutMs, std::vector<Element> &outElements)
{
    Element element = {.ptr = nullptr, .size = sizeof(int), .id = ULONG_MAX};
    outElements.emplace_back(element);
    if (expectNum == 999) {
        for (int i = 1; i < 999; ++i) {
            Element element = {.ptr = nullptr, .size = 100, .id = ULONG_MAX};
            outElements.emplace_back(element);
        }
    }
    return Status::OK();
}

Status Consumer::Receive(uint32_t timeoutMs, std::vector<Element> &outElements)
{
    Element element = {.ptr = nullptr, .size = sizeof(int), .id = ULONG_MAX};
    outElements.emplace_back(element);
    return Status::OK();
}

Status Consumer::Close()
{
    return Status::OK();
}

Status Consumer::Ack(uint64_t elementId)
{
    return Status::OK();
}

class ObjectClientImpl {};

ObjectClient::ObjectClient(const ConnectOptions &connectOptions) {}

ObjectClient::~ObjectClient() {}

Status ObjectClient::Init()
{
    return Status::OK();
}

Status ObjectClient::Create(const std::string &objectId, uint64_t size, const CreateParam &param,
                            std::shared_ptr<Buffer> &buffer)
{
    if (objectId == "repeatedObjId") {
        return Status(StatusCode::K_OC_ALREADY_SEALED, "repeated seal");
    } else if (objectId == "errObjId") {
        return Status(StatusCode::K_RPC_DEADLINE_EXCEEDED, "error");
    }
    buffer = std::make_shared<Buffer>();
    return Status::OK();
}

Status ObjectClient::Get(const std::vector<std::string> &objectIds, int32_t timeout,
                         std::vector<Optional<Buffer>> &buffer)
{
    // To test the if branch of partial get,
    // if a vector of len = 1, successfully get
    // if a vector of len > 1, only store the first element
    buffer.clear();
    auto buf = std::make_shared<Buffer>();
    if (objectIds.size() == 1) {
        buffer.emplace_back(std::move(*buf));
        return Status::OK();
    }
    // if length >= 2, only store the first object
    buffer.emplace_back(std::move(*buf));            // only first one is non-empty
    for (size_t i = 1; i < objectIds.size(); ++i) {  // others are empty
        buffer.emplace_back();
    }
    auto status = Status(StatusCode::K_OUT_OF_MEMORY, "mock test runtime error");
    return status;
}

Status ObjectClient::GIncreaseRef(const std::vector<std::string> &objectIds, std::vector<std::string> &failedObjectIds,
                                     const std::string &remoteClientId)
{
    if (objectIds.size() == 2) {
        return Status(StatusCode::K_RPC_DEADLINE_EXCEEDED, "error");
    }
    return Status::OK();
}

Status ObjectClient::GDecreaseRef(const std::vector<std::string> &objectIds, std::vector<std::string> &failedObjectIds,
                                     const std::string &remoteClientId)
{
    return Status::OK();
}

int ObjectClient::QueryGlobalRefNum(const std::string &id)
{
    return 1;
}

Status ObjectClient::ReleaseGRefs(const std::string &remoteClientId)
{
    return Status::OK();
}

Status ObjectClient::GenerateKey(const std::string &prefix, std::string &key)
{
    key = prefix;
    return Status::OK();
}

Status ObjectClient::ShutDown()
{
    return Status::OK();
}
constexpr int64_t bufferSize = 32;
void *mutableData = nullptr;

Status Buffer::WLatch(uint64_t timeout)
{
    return Status::OK();
}

Status Buffer::MemoryCopy(const void *data, uint64_t length)
{
    return Status::OK();
}

Status Buffer::Seal(const std::unordered_set<std::string> &nestedIds)
{
    return Status::OK();
}

Status Buffer::UnWLatch()
{
    if (mutableData != nullptr) {
        free(mutableData);
    }
    mutableData = nullptr;
    return Status::OK();
}

Status Buffer::RLatch(uint64_t timeout)
{
    return Status::OK();
}

const void *Buffer::ImmutableData()
{
    return static_cast<const void *>(MutableData());
}

void *Buffer::MutableData()
{
    if (mutableData == nullptr) {
        mutableData = (void *)std::malloc(bufferSize);
        return mutableData;
    }
    return mutableData;
}

int64_t Buffer::GetSize() const
{
    return bufferSize;
}

Status Buffer::UnRLatch()
{
    return Status::OK();
}

Status Buffer::Publish(const std::unordered_set<std::string> &nestedIds)
{
    return Status::OK();
}

class KVClientImpl {};

KVClient::KVClient(const ConnectOptions &connectOptions) {};

Status KVClient::Init()
{
    return Status::OK();
}

Status KVClient::Set(const std::string &key, const StringView &val, const SetParam &param)
{
    if (param.writeMode == WriteMode::NONE_L2_CACHE_EVICT) {
        return Status(StatusCode::K_OUT_OF_MEMORY, "mock test runtime error");
    }
    if (key == "wrongKey") {
        return Status(StatusCode::K_RUNTIME_ERROR, "ERROR MESSAGE");
    }
    return Status::OK();
}

std::string KVClient::Set(const StringView &val, const SetParam &setParam)
{
    if (val.data() == nullptr || val.size() == 0) {
        return "";
    }
    return "returnKey";
}

std::string KVClient::GenerateKey(const std::string &prefixKey)
{
    return "genKey";
}

Status KVClient::Get(const std::string &key, std::string &val, int32_t timeoutMs)
{
    if (key == "wrongKey") {
        return Status(StatusCode::K_RUNTIME_ERROR, "ERROR MESSAGE");
    }
    val = key;
    return Status::OK();
}

Status KVClient::Get(const std::vector<std::string> &keys, std::vector<Optional<ReadOnlyBuffer>> &readOnlyBuffers,
                     int32_t timeoutMs)
{
    readOnlyBuffers.clear();
    if (keys.empty()) {
        return Status::OK();
    }
    if (keys.size() == 1) {
        if (keys[0] == "wrongKey") {
            readOnlyBuffers.emplace_back();
            return Status(StatusCode::K_OUT_OF_MEMORY, "mock test runtime error");
        }
        auto buf = std::make_shared<Buffer>();
        auto rdBuf = std::make_shared<ReadOnlyBuffer>(buf);
        readOnlyBuffers.emplace_back(std::move(*rdBuf));
        return Status::OK();
    }
    std::unordered_set<std::string> uniq(keys.begin(), keys.end());
    if (uniq.size() == keys.size()) {
        for (size_t i = 0; i < keys.size(); ++i) {
            auto buf = std::make_shared<Buffer>();
            auto rdBuf = std::make_shared<ReadOnlyBuffer>(buf);
            readOnlyBuffers.emplace_back(std::move(*rdBuf));
        }
        return Status::OK();
    }
    auto buf = std::make_shared<Buffer>();
    auto rdBuf = std::make_shared<ReadOnlyBuffer>(buf);
    readOnlyBuffers.emplace_back(std::move(*rdBuf));
    for (size_t i = 1; i < keys.size(); ++i) {
        readOnlyBuffers.emplace_back();
    }
    return Status(StatusCode::K_OUT_OF_MEMORY, "mock test runtime error");
}

Status KVClient::Get(const std::vector<std::string> &keys, std::vector<std::string> &vals, int32_t timeoutMs)
{
    char v = 'v';
    vals.emplace_back(reinterpret_cast<const char *>(&v), 1);
    return Status::OK();
}

Status KVClient::Del(const std::string &key)
{
    if (key == "wrongKey") {
        return Status(StatusCode::K_RUNTIME_ERROR, "ERROR MESSAGE");
    }
    return Status::OK();
}

Status KVClient::Del(const std::vector<std::string> &keys, std::vector<std::string> &failedKeys)
{
    std::string wrongKey = "wrongKey";
    auto it = std::find(keys.begin(), keys.end(), wrongKey);
    if (it != keys.end()) {
        failedKeys.emplace_back(wrongKey);
        return Status(StatusCode::K_RUNTIME_ERROR, "ERROR MESSAGE");
    }
    return Status::OK();
}

Status KVClient::Exist(const std::vector<std::string> &keys, std::vector<bool> &exists)
{
    if (keys.empty()) {
        return Status(StatusCode::K_INVALID, "The keys are empty");
    }
    return Status::OK();
}

Status KVClient::QuerySize(const std::vector<std::string> &objectKeys, std::vector<uint64_t> &outSizes)
{
    outSizes.clear();
    for (size_t i = 0; i < objectKeys.size(); ++i) {
        outSizes.emplace_back(0);
    }
    return Status::OK();
}

Status KVClient::ShutDown()
{
    return Status::OK();
}

Status KVClient::HealthCheck()
{
    return Status::OK();
}

Status::Status() noexcept : state_(nullptr) {}

Status::Status(StatusCode code, std::string msg) noexcept
{
    state_ = std::make_unique<State>();
    state_->code = code;
    state_->errMsg = msg;
}

std::string Status::ToString() const
{
    return "code: [" + std::to_string(GetCode()) + "], msg: [" + GetMsg() + "]";
}

StatusCode Status::GetCode() const
{
    return state_ == nullptr ? K_OK : state_->code;
}

Status::Status(const Status &other) noexcept
{
    Assign(other);
}

Status &Status::operator=(const Status &other) noexcept
{
    Assign(other);
    return *this;
}

Status::Status(Status &&other) noexcept
{
    std::swap(state_, other.state_);
}

Status &Status::operator=(Status &&other) noexcept
{
    std::swap(state_, other.state_);
    return *this;
}

std::string Status::GetMsg() const
{
    return state_ == nullptr ? "" : state_->errMsg;
}

void Status::Assign(const Status &other) noexcept
{
    if (other.IsOk()) {
        state_ = nullptr;
        return;
    }
    if (state_ == nullptr) {
        state_ = std::make_unique<State>();
    }
    *state_ = *other.state_;
}

SensitiveValue::SensitiveValue(char const *) {}

SensitiveValue::~SensitiveValue() {}

SensitiveValue &SensitiveValue::operator=(const std::string &str)
{
    return *this;
}

Buffer::Buffer(Buffer &&other) noexcept {}

Buffer::~Buffer() {}

class HeteroClientImpl {};

HeteroClient::HeteroClient(const ConnectOptions &connectOptions) {}

HeteroClient::~HeteroClient() {}

Status HeteroClient::Init()
{
    return Status::OK();
}

Status HeteroClient::ShutDown()
{
    return Status::OK();
}

Status HeteroClient::MGetH2D(const std::vector<std::string> &objectIds, const std::vector<DeviceBlobList> &devBlobList,
                             std::vector<std::string> &failList, int32_t timeoutMs)
{
    return Status::OK();
}

Status HeteroClient::DevDelete(const std::vector<std::string> &objectIds, std::vector<std::string> &failedObjectIds)
{
    return Status::OK();
}

Status HeteroClient::DevLocalDelete(const std::vector<std::string> &objectIds,
                                    std::vector<std::string> &failedObjectIds)
{
    return Status::OK();
}

Status HeteroClient::DevSubscribe(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                                  std::vector<Future> &futureVec)
{
    std::promise<Status> promise;
    auto future = promise.get_future().share();
    std::shared_ptr<DeviceRtEventWrapper> event;
    Future f(future, event, "obj1");
    futureVec.emplace_back(f);
    return Status::OK();
}

Status HeteroClient::DevPublish(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                                std::vector<Future> &futureVec)
{
    std::promise<Status> promise;
    auto future = promise.get_future().share();
    std::shared_ptr<DeviceRtEventWrapper> event;
    Future f(future, event, "obj1");
    futureVec.emplace_back(f);
    return Status::OK();
}

Status HeteroClient::DevMSet(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                             std::vector<std::string> &failedKeys)
{
    return Status::OK();
}

Status HeteroClient::DevMGet(const std::vector<std::string> &keys, std::vector<DeviceBlobList> &blob2dList,
                             std::vector<std::string> &failedKeys, int32_t timeoutMs)
{
    return Status::OK();
}
}  // namespace datasystem
