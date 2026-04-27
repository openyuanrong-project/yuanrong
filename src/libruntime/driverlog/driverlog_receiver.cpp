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

#include "driverlog_receiver.h"

#include "src/libruntime/streamstore/stream_producer_consumer.h"

namespace YR {
namespace Libruntime {
const uint32_t CONSUMER_TIMEOUT = 1000;
const int AGG_WINDOWS_S = 5;
const int LOGINFO_MATCH_LEN = 2;
const size_t MAX_LOG_SCAN_LENGTH = 256;  // only scan 256 charactors in log message to parse runtime id

DriverLogReceiver::DriverLogReceiver() {}

void DriverLogReceiver::Init(std::shared_ptr<StreamStore> store, std::string &jobId, bool dedup)
{
    if (jobId.empty() || !store) {
        YRLOG_ERROR("failed to init driverlog receiver {}", jobId);
    }
    jobId_ = "/log/runtime/std/" + jobId;
    dedup_ = dedup;
    dsStreamStore_ = store;
    receiverThread_ = std::thread(&DriverLogReceiver::Receive, this);
}

DriverLogReceiver::~DriverLogReceiver()
{
    Stop();
    if (receiverThread_.joinable()) {
        receiverThread_.join();
    }
}

void DriverLogReceiver::Stop()
{
    stopped_ = true;
    if (consumer_ != nullptr) {
        auto err = consumer_->Close();
        if (!err.OK()) {
            YRLOG_WARN("failed to close log consumer {}, err code: {}, err message: {}", jobId_,
                       fmt::underlying(err.Code()), err.Msg());
        }
        YRLOG_DEBUG("log consummer {} closed", jobId_);
        if (!jobId_.empty()) {
            err = dsStreamStore_->DeleteStream(jobId_);
            if (!err.OK()) {
                YRLOG_WARN("failed to delete stream {}, err code: {}, err message: {}", jobId_,
                           fmt::underlying(err.Code()), err.Msg());
            }
            YRLOG_DEBUG("log stream {} delete", jobId_);
        }
    }
    Flush(-1);
}

std::string DriverLogReceiver::Format(const DedupState &state)
{
    std::ostringstream oss;
    oss << "[repeated " << state.count << "x across cluster] " << state.line;
    return oss.str();
}

std::pair<std::string, std::string> DriverLogReceiver::ParseLine(const std::string &input)
{
    size_t scanLen = input.length() > MAX_LOG_SCAN_LENGTH ? MAX_LOG_SCAN_LENGTH : input.length();
    int leftBraceCounter = 0;
    int rightBraceCounter = 0;
    int runtimeIdStartPos = -1;
    int runtimeIdEndPos = -1;
    for (size_t i = 0; i < scanLen; i++) {
        auto c = input[i];
        if (c == '(') {
            leftBraceCounter++;
            if (leftBraceCounter == 1) {
                runtimeIdStartPos = i + 1;
            }
        } else if (c == ')') {
            rightBraceCounter++;
            if (rightBraceCounter == leftBraceCounter) {
                runtimeIdEndPos = i;
                break;
            }
        }
    }
    if (runtimeIdStartPos == -1 || runtimeIdEndPos == -1) {
        return {"", ""};
    }
    return {input.substr(runtimeIdStartPos, runtimeIdEndPos - runtimeIdStartPos), input.substr(runtimeIdEndPos + 1)};
}

void DriverLogReceiver::Deduplicate(std::shared_ptr<std::vector<std::string>> lines)
{
    auto now = std::chrono::time_point_cast<std::chrono::seconds>(std::chrono::steady_clock::now())
               .time_since_epoch()
               .count();
    Flush(now);
    std::lock_guard<std::mutex> lk(mtx);
    for (const auto &line : *lines) {
        auto [runtimeId, dedupKey] = ParseLine(line);
        if (runtimeId.empty() || dedupKey.empty()) {
            std::cout << line << std::flush;
            continue;
        }
        if (recent_.find(dedupKey) != recent_.end()) {
            recent_[dedupKey].sources.insert(runtimeId);
            if (recent_[dedupKey].sources.size() > 1) {
                recent_[dedupKey].timestamp = now;
                recent_[dedupKey].count++;
                recent_[dedupKey].line = line;
            } else {
                std::cout << line << std::flush;
            }
        } else {
            recent_[dedupKey] = DedupState{now, 0, line, {runtimeId}};
            std::cout << line << std::flush;
        }
    }
}

void DriverLogReceiver::Flush(int64_t now)
{
    std::lock_guard<std::mutex> lk(mtx);
    auto it = recent_.begin();
    while (it != recent_.end()) {
        auto duration = now - it->second.timestamp;
        if (duration >= 0 && duration < AGG_WINDOWS_S) {
            ++it;
            continue;
        }
        if (it->second.count > 1) {
            std::cout << Format(it->second) << std::flush;
            auto parseRes = ParseLine(it->second.line);
            it->second.timestamp = now;
            it->second.count = 0;
            it->second.sources.clear();
            it->second.sources.insert(std::move(parseRes.first));
        } else if (it->second.count > 0) {
            std::cout << it->second.line << std::flush;
            // Aggregation wasn't fruitful, print the line and stop aggregating.
            it = recent_.erase(it);
            continue;
        } else if (it->second.count == 0) {
            it = recent_.erase(it);
            continue;
        }
        ++it;
    }
}

void DriverLogReceiver::Receive()
{
    SubscriptionConfig config;
    auto consumer = std::make_shared<StreamConsumer>();
    auto initErr = dsStreamStore_->CreateStreamConsumer(jobId_, config, consumer, true);
    if (!initErr.OK()) {
        YRLOG_ERROR("failed to create log stream consumer {}, err code: {}, err message: {}", jobId_,
                    fmt::underlying(initErr.Code()), initErr.Msg());
        return;
    }
    consumer_ = consumer;
    YRLOG_INFO("begin to receive log from topic: {}, dedup {}", jobId_, dedup_);

    while (!stopped_) {
        std::vector<Element> elements;
        auto err = consumer_->Receive(CONSUMER_TIMEOUT, elements);
        if (!err.OK()) {
            YRLOG_ERROR("failed to receive log from consumer, err code: {}, err message: {}",
                        fmt::underlying(err.Code()), err.Msg());
            return;
        }
        auto lines = std::make_shared<std::vector<std::string>>();
        for (auto &ele : elements) {
            auto logInfo = std::string(reinterpret_cast<char *>(ele.ptr), ele.size);
            if (!dedup_) {
                std::cout << logInfo << std::flush;
            } else {
                lines->push_back(logInfo);
            }
        }
        if (dedup_) {
            Deduplicate(lines);
        }
    }
    YRLOG_INFO("finishe to receive message from topic: {}", jobId_);
}
}  // namespace Libruntime
}  // namespace YR
