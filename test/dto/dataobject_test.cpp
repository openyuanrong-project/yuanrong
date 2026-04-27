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
#include <cstdlib>

#include <gmock/gmock.h>
#include <gtest/gtest.h>

#define private public
#include "src/dto/data_object.h"
using namespace YR::Libruntime;
class DTODataObjectTest : public testing::Test {
public:
    DTODataObjectTest(){};
    ~DTODataObjectTest(){};
    void SetUp() override {}
    void TearDown() override {}
};

TEST_F(DTODataObjectTest, TestNUllArgs)
{
    auto argBuf = std::make_shared<YR::Libruntime::NativeBuffer>(0);
    auto sendDataObj = std::make_shared<DataObject>(0, 0);
    sendDataObj->data->MemoryCopy(nullptr, 0);
    EXPECT_TRUE(sendDataObj->data->GetSize() == 0);

    auto rawArg = std::make_shared<DataObject>("abc", sendDataObj->buffer);
    EXPECT_TRUE(rawArg->data);
    EXPECT_TRUE(rawArg->data->ImmutableData());
    EXPECT_TRUE(rawArg->data->GetSize() == 0);
}

TEST_F(DTODataObjectTest, TestConstructor)
{
    auto argBuf = std::make_shared<YR::Libruntime::NativeBuffer>(26);
    auto rawArg = std::make_shared<DataObject>("abc", argBuf);
    EXPECT_TRUE(rawArg->data->GetSize() == 10);
}
