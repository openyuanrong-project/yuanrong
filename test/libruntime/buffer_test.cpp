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

#include <thread>

#include <gmock/gmock.h>
#include <gtest/gtest.h>

#include "src/dto/buffer.h"

class BufferTest : public testing::Test {
public:
    BufferTest(){};
    ~BufferTest(){};
    void SetUp() override {}
    void TearDown() override {}
};

TEST_F(BufferTest, NativeBufferTest)
{
    uint64_t size = 10;
    auto buf = YR::Libruntime::NativeBuffer(size);
    ASSERT_TRUE(buf.Seal({}).OK());
    ASSERT_TRUE(buf.WriterLatch().OK());
    ASSERT_TRUE(buf.WriterUnlatch().OK());
    ASSERT_TRUE(buf.ReaderLatch().OK());
    ASSERT_TRUE(buf.ReaderUnlatch().OK());
    ASSERT_TRUE(buf.IsNative());
    ASSERT_EQ(buf.GetSize(), size);
    ASSERT_FALSE(buf.MemoryCopy("aaa", 20).OK());
    ASSERT_TRUE(buf.MemoryCopy("0123456789", 10).OK());
    ASSERT_EQ(std::string(static_cast<const char *>(buf.ImmutableData()), buf.GetSize()), "0123456789");
    ASSERT_EQ(std::string(static_cast<char *>(buf.MutableData()), buf.GetSize()), "0123456789");
}

TEST_F(BufferTest, ReadOnlyNativeBufferTest)
{
    const char *test = "0123456789";
    auto buf = YR::Libruntime::ReadOnlyNativeBuffer(test, 10);
    ASSERT_TRUE(buf.IsNative());
    ASSERT_EQ(buf.GetSize(), 10);
    ASSERT_FALSE(buf.MemoryCopy("aaa", 20).OK());
    ASSERT_EQ(std::string(static_cast<const char *>(buf.ImmutableData()), buf.GetSize()), "0123456789");
    ASSERT_EQ(std::string(static_cast<char *>(buf.MutableData()), buf.GetSize()), "0123456789");
}

TEST_F(BufferTest, SharedBufferTest)
{
    uint64_t size = 10;
    void *data = std::malloc(size);
    auto buf = YR::Libruntime::SharedBuffer(data, size);
    ASSERT_FALSE(buf.IsNative());
    ASSERT_EQ(buf.GetSize(), size);
    ASSERT_FALSE(buf.MemoryCopy("aaa", 20).OK());
    ASSERT_TRUE(buf.MemoryCopy("0123456789", 10).OK());
    ASSERT_EQ(std::string(static_cast<const char *>(buf.ImmutableData()), buf.GetSize()), "0123456789");
    ASSERT_EQ(std::string(static_cast<char *>(buf.MutableData()), buf.GetSize()), "0123456789");
    free(data);
}
