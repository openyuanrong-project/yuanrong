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

#include "gmock/gmock.h"
#include "src/libruntime/fsclient/fs_intf_manager.h"

namespace YR {
namespace test {
using YR::Libruntime::FSIntfReaderWriter;
using YR::Libruntime::ReaderWriterClientOption;
using YR::Libruntime::ProtocolType;
class MockFSIntfManager : public YR::Libruntime::FSIntfManager {
public:
    MockFSIntfManager() : YR::Libruntime::FSIntfManager(nullptr) {}
    ~MockFSIntfManager() = default;
    MOCK_METHOD(std::shared_ptr<FSIntfReaderWriter>, NewFsIntfClient,
                (const std::string &srcInstance, const std::string &dstInstance, const std::string &runtimeID,
                 const ReaderWriterClientOption &option, ProtocolType type),
                (override));
    MOCK_METHOD(std::shared_ptr<FSIntfReaderWriter>, NewEventIntfClient,
            (const std::string &srcInstance, const std::string &dstInstance, const std::string &runtimeID,
             const ReaderWriterClientOption &option, ProtocolType type),
            (override));
    MOCK_METHOD(std::shared_ptr<FSIntfReaderWriter>, TryGet, (const std::string &instanceID), (override));
    MOCK_METHOD(std::shared_ptr<FSIntfReaderWriter>, Get, (const std::string &instanceID), (override));
    MOCK_METHOD(std::shared_ptr<FSIntfReaderWriter>, TryGetEventIntfs, (const std::string &instanceID), (override));
    MOCK_METHOD(std::shared_ptr<FSIntfReaderWriter>, GetEventIntfs, (const std::string &instanceID), (override));
    MOCK_METHOD(bool, EmplaceEventIntfs, (const std::string &instanceID,
                const std::shared_ptr<FSIntfReaderWriter> &intf), (override));
};
}  // namespace test
}  // namespace YR