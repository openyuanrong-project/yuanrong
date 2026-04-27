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

#include <gmock/gmock.h>
#include <gtest/gtest.h>
#include <chrono>
#include <future>
#include <iostream>
#include <thread>

#include "mock/mock_datasystem.h"
#include "mock/mock_datasystem_client.h"
#include "src/libruntime/objectstore/datasystem_object_store.h"
#include "src/libruntime/objectstore/memory_store.h"
#include "src/libruntime/waiting_object_manager.h"
#include "src/utility/logger/logger.h"
#define private public
#include "src/libruntime/objectstore/object_id_pool.h"

using namespace testing;
using namespace YR::Libruntime;
using namespace YR::utility;

namespace YR {
namespace test {
class ObjectIdPoolTest : public testing::Test {
public:
    ObjectIdPoolTest(){};
    ~ObjectIdPoolTest(){};
    void SetUp() override
    {
        Mkdir("/tmp/log");
        LogParam g_logParam = {
            .logLevel = "DEBUG",
            .logDir = "/tmp/log",
            .nodeName = "test-runtime",
            .modelName = "test",
            .maxSize = 100,
            .maxFiles = 1,
            .logFileWithTime = false,
            .logBufSecs = 30,
            .maxAsyncQueueSize = 1048510,
            .asyncThreadCount = 1,
            .alsoLog2Stderr = true,
        };
        InitLog(g_logParam);
        memoryStore = std::make_shared<MemoryStore>();
        dsObjectStore = std::make_shared<MockObjectStore>();
        auto wom = std::make_shared<WaitingObjectManager>();
        memoryStore->Init(dsObjectStore, wom);
        objectPool = std::make_shared<ObjectIdPool>(memoryStore, 100);
    }

    void TearDown() override
    {
        dsObjectStore.reset();
        memoryStore.reset();
        objectPool.reset();
    }

    std::shared_ptr<MemoryStore> memoryStore;
    std::shared_ptr<MockObjectStore> dsObjectStore;
    std::shared_ptr<ObjectIdPool> objectPool;
};

TEST_F(ObjectIdPoolTest, ScaleTest)
{
    auto err = objectPool->Scale();
    ASSERT_EQ(err.OK(), true);
    objectPool->Clear();

}
}  // namespace test
}  // namespace YR