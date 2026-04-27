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

#include <gmock/gmock.h>
#include <gtest/gtest.h>

#include "mock/mock_datasystem_client_wrapper.h"
#include "src/libruntime/objectstore/async_decre_ref.h"

using namespace YR::Libruntime;
using namespace testing;
namespace YR {
namespace test {
class AsyncDecreRefTest : public testing::Test {
public:
    AsyncDecreRefTest(){};
    ~AsyncDecreRefTest(){};

    void SetUp() override
    {
        wrapper = std::make_shared<MockDatasystemClientWrapper>();
        asyncDecreRef = std::make_shared<AsyncDecreRef>();
        asyncDecreRef->Init(wrapper);
    }

    void TearDown() override
    {
        if (asyncDecreRef) {
            asyncDecreRef.reset();
        }
        if (wrapper) {
            wrapper.reset();
        }
    }
    std::shared_ptr<MockDatasystemClientWrapper> wrapper;
    std::shared_ptr<AsyncDecreRef> asyncDecreRef;
};

void checkResult(std::shared_ptr<AsyncDecreRef> asyncDecreRef)
{
    bool flag = false;
    auto time = 0;
    while (true) {
        // wait process done, timeout: 10s
        flag = asyncDecreRef->IsEmpty();
        if (flag || time >= 100) {
            break;
        }
        ++time;
        std::this_thread::sleep_for(std::chrono::milliseconds(100));
    }
    ASSERT_TRUE(flag);
}

TEST_F(AsyncDecreRefTest, PushPopSuccessfullyTest)
{
    EXPECT_CALL(*this->wrapper, GDecreaseRef(_, _)).WillRepeatedly(testing::Return(datasystem::Status()));
    std::vector<std::string> objs1;
    objs1.push_back("tenantID1-obj1");
    asyncDecreRef->Push(objs1, "tenantID1");
    std::vector<std::string> objs2;
    for (int i = 0; i < 1002; i++) {
        objs2.push_back("tenantID2-obj" + std::to_string(i));
    }
    asyncDecreRef->Push(objs2, "tenantID2");
    checkResult(asyncDecreRef);
}

TEST_F(AsyncDecreRefTest, PushPopFailedTest)
{
    EXPECT_CALL(*this->wrapper, GDecreaseRef(_, _))
        .WillRepeatedly(
            testing::Return(datasystem::Status(datasystem::StatusCode::K_RUNTIME_ERROR, "failed to decrease ref")));
    std::vector<std::string> objs1;
    objs1.push_back("tenantID1-obj1");
    asyncDecreRef->Push(objs1, "tenantID1");
    std::vector<std::string> objs2;
    for (int i = 0; i < 1002; i++) {
        objs2.push_back("tenantID2-obj" + std::to_string(i));
    }
    asyncDecreRef->Push(objs2, "tenantID2");
    checkResult(asyncDecreRef);
}
}  // namespace test
}  // namespace YR