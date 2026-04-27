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

#include <tuple>

#include <gmock/gmock.h>
#include <gtest/gtest.h>

#include "src/utility/memory.h"

namespace YR {
namespace test {
using namespace YR::utility;
class MemoryUtilityTest : public ::testing::Test {
    void SetUp() override {}

    void TearDown() override {}
};

TEST_F(MemoryUtilityTest, CopyInParallel)
{
    const size_t dataSize = 1024;
    const size_t blockSize = 64;
    const int64_t totalBytes = dataSize;

    uint8_t *src = new uint8_t[dataSize];
    uint8_t *dst = new uint8_t[dataSize];

    for (size_t i = 0; i < dataSize; ++i) {
        src[i] = static_cast<uint8_t>(i);
    }

    CopyInParallel(dst, src, totalBytes, blockSize);

    for (size_t i = 0; i < dataSize; ++i) {
        EXPECT_EQ(src[i], dst[i]);
    }
    delete[] src;
    delete[] dst;
}
}  // namespace test
}  // namespace YR