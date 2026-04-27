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
#include <cstdlib>
#include <string>

#include "user_common_func.h"
#include "utils.h"
#include "yr/yr.h"

using testing::HasSubstr;

class CollectiveTest : public testing::Test {
public:
    CollectiveTest() {};
    ~CollectiveTest() {};
    static void SetUpTestCase() {};
    static void TearDownTestCase() {};

    void SetUp()
    {
        YR::Config config;
        config.mode = YR::Config::Mode::CLUSTER_MODE;
        auto info = YR::Init(config);
        std::cout << "job id: " << info.jobId << std::endl;
    };

    void TearDown()
    {
        YR::Finalize();
    };
};

/**
 * @title: 试图创建非法groupName group
 * @step:  1.创建CollectiveActor实例
 * @step:  2.调用CollectiveActor::InitCollectiveGroup，在acto中创建非法group
 * @step:  3.调用CreateCollectiveGroup，创建非法group
 *
 * @expect: 2.预期异常抛出
 * @expect: 3.预期异常抛出
 */
TEST_F(CollectiveTest, InvalidGroupNameTest)
{
    auto ins = YR::Instance(CollectiveActor::FactoryCreate).Invoke();
    std::string groupName1 = "test@group1";
    std::string groupName2 = "test/group1";
    std::string groupName3 = "test-group1";
    auto res = ins.Function(&CollectiveActor::InitCollectiveGroup).Invoke(groupName1, 0, 1);
    EXPECT_THROW_WITH_CODE_AND_MSG(YR::Get(res), 2002, "groupName is invalid. It should match the regex");

    res = ins.Function(&CollectiveActor::InitCollectiveGroup).Invoke(groupName2, 0, 1);
    EXPECT_THROW_WITH_CODE_AND_MSG(YR::Get(res), 2002, "groupName is invalid. It should match the regex");

    YR::Collective::CollectiveGroupSpec spec1{
        .worldSize = 1,
        .groupName = groupName1,
    };
    EXPECT_THROW_WITH_CODE_AND_MSG(YR::Collective::CreateCollectiveGroup(spec1, {ins.GetInstanceId()}, {0}), 1001,
                                   "groupName is invalid. It should match the regex");

    YR::Collective::CollectiveGroupSpec spec2{
        .worldSize = 1,
        .groupName = groupName2,
    };
    EXPECT_THROW_WITH_CODE_AND_MSG(YR::Collective::CreateCollectiveGroup(spec2, {ins.GetInstanceId()}, {0}), 1001,
                                   "groupName is invalid. It should match the regex");

    YR::Collective::CollectiveGroupSpec spec3{
        .worldSize = 1,
        .groupName = groupName3,
    };
    YR::Collective::CreateCollectiveGroup(spec3, {ins.GetInstanceId()}, {0});
    EXPECT_THROW_WITH_CODE_AND_MSG(YR::Collective::CreateCollectiveGroup(spec3, {ins.GetInstanceId()}, {0}), 1001,
                                   "already existed, please destroy it first");
    YR::Collective::DestroyCollectiveGroup(groupName3);
}

/**
 * @title: InitCollectiveGroup函数调用成功
 * @step:  1.创建CollectiveActor实例
 * @step:  2.调用CollectiveActor::InitCollectiveGroup，在acto中创建对应group
 * @step:  3.调用CollectiveActor::Compute sum 函数，传入{1, 2, 3, 4, 5}
 * @step:  4.调用CollectiveActor::Compute min 函数，传入{1, 2, 3, 4, 5}, {5, 4, 3, 2, 1}
 * @step:  5.调用CollectiveActor::Compute max 函数，传入{5, 4, 3, 2, 1}, {1, 2, 3, 4, 5}
 * @step:  6.调用CollectiveActor::Compute product 函数，传入{5, 4, 3, 2, 1}, {1, 2, 3, 4, 5}
 *
 * @expect: 3.预期无异常抛出，返回值为两个actor中各项之和，30
 * @expect: 4.预期无异常抛出，返回值为两个actor中各项取最小值之和，9
 * @expect: 5.预期无异常抛出，返回值为两个actor中各项取最大值之和，21
 * @expect: 6.预期无异常抛出，返回值为两个actor中各项相乘之和，35
 */
TEST_F(CollectiveTest, InitGroupInActorTest)
{
    auto ins1 = YR::Instance(CollectiveActor::FactoryCreate).Invoke();
    auto ins2 = YR::Instance(CollectiveActor::FactoryCreate).Invoke();

    std::string groupName = "test-group1";
    YR::Collective::DestroyCollectiveGroup(groupName);

    ins1.Function(&CollectiveActor::InitCollectiveGroup).Invoke(groupName, 0, 2);
    auto ret = ins2.Function(&CollectiveActor::InitCollectiveGroup).Invoke(groupName, 1, 2);
    YR::Get(ret);

    std::vector<int> input = {1, 2, 3, 4, 5};
    auto res1 =
        ins1.Function(&CollectiveActor::Compute).Invoke(input, groupName, static_cast<uint8_t>(YR::ReduceOp::SUM));
    auto res2 =
        ins2.Function(&CollectiveActor::Compute).Invoke(input, groupName, static_cast<uint8_t>(YR::ReduceOp::SUM));
    EXPECT_EQ(*YR::Get(res1), 30);
    EXPECT_EQ(*YR::Get(res2), 30);

    input = {1, 2, 3, 4, 5};
    std::vector<int> input2 = {5, 4, 3, 2, 1};
    res1 = ins1.Function(&CollectiveActor::Compute).Invoke(input, groupName, static_cast<uint8_t>(YR::ReduceOp::MIN));
    res2 = ins2.Function(&CollectiveActor::Compute).Invoke(input2, groupName, static_cast<uint8_t>(YR::ReduceOp::MIN));
    EXPECT_EQ(*YR::Get(res1), 9);
    EXPECT_EQ(*YR::Get(res2), 9);

    res1 = ins1.Function(&CollectiveActor::Compute).Invoke(input2, groupName, static_cast<uint8_t>(YR::ReduceOp::MAX));
    res2 = ins2.Function(&CollectiveActor::Compute).Invoke(input, groupName, static_cast<uint8_t>(YR::ReduceOp::MAX));
    EXPECT_EQ(*YR::Get(res1), 21);
    EXPECT_EQ(*YR::Get(res2), 21);

    res1 =
        ins1.Function(&CollectiveActor::Compute).Invoke(input2, groupName, static_cast<uint8_t>(YR::ReduceOp::PRODUCT));
    res2 =
        ins2.Function(&CollectiveActor::Compute).Invoke(input, groupName, static_cast<uint8_t>(YR::ReduceOp::PRODUCT));
    EXPECT_EQ(*YR::Get(res1), 35);
    EXPECT_EQ(*YR::Get(res2), 35);

    res1 = ins1.Function(&CollectiveActor::DestroyCollectiveGroup).Invoke(groupName);
    res2 = ins2.Function(&CollectiveActor::DestroyCollectiveGroup).Invoke(groupName);
    YR::Get(res1);
    YR::Get(res2);
}

/**
 * @title: CreateCollectiveGroup函数调用成功
 * @step:  1.创建CollectiveActor实例
 * @step:  2.调用CreateCollectiveGroup，在driver中创建对应group
 * @step:  3.调用CollectiveActor::Compute函数，传入{1, 2, 3, 4}
 * @step:  4.调用CollectiveActor::ComputeDouble函数，传入{1.1, 2.2, 3.3, 4.4}
 * @step:  5.调用DestroyCollectiveGroup，清理CollectiveGroup
 *
 * @expect: 预期无异常抛出，返回值为各个actor中各项之和，40 和 44
 */
TEST_F(CollectiveTest, CreateGroupInDriverTest)
{
    std::vector<YR::NamedInstance<CollectiveActor>> instances;
    std::vector<std::string> instanceIDs;
    for (int i = 0; i < 4; ++i) {
        auto ins = YR::Instance(CollectiveActor::FactoryCreate).Invoke();
        instances.push_back(ins);
        instanceIDs.push_back(ins.GetInstanceId());
    }

    std::string groupName = "test-group2";
    YR::Collective::DestroyCollectiveGroup(groupName);
    YR::Collective::CollectiveGroupSpec spec{
        .worldSize = 4,
        .groupName = groupName,
    };
    YR::Collective::CreateCollectiveGroup(spec, instanceIDs, {0, 1, 2, 3});

    std::vector<int> input = {1, 2, 3, 4};
    std::vector<YR::ObjectRef<int>> res;
    for (int i = 0; i < 4; ++i) {
        res.push_back(instances[i]
                          .Function(&CollectiveActor::Compute)
                          .Invoke(input, groupName, static_cast<uint8_t>(YR::ReduceOp::SUM)));
    }

    for (int i = 0; i < 4; ++i) {
        EXPECT_EQ(*YR::Get(res[i]), 40);
    }

    std::vector<double> input2 = {1.1, 2.2, 3.3, 4.4};
    std::vector<YR::ObjectRef<double>> res2;
    for (int i = 0; i < 4; ++i) {
        res2.push_back(instances[i]
                           .Function(&CollectiveActor::ComputeDouble)
                           .Invoke(input2, groupName, static_cast<uint8_t>(YR::ReduceOp::SUM)));
    }

    for (int i = 0; i < 4; ++i) {
        EXPECT_EQ(*YR::Get(res2[i]), 44);
    }

    YR::Collective::DestroyCollectiveGroup(groupName);
}

/**
 * @title: Reduce函数调用成功
 * @step:  1.创建CollectiveActor实例
 * @step:  2.调用CreateCollectiveGroup，在driver中创建对应group
 * @step:  3.调用CollectiveActor::Reduce函数，传入{1, 2, 3, 4}
 * @step:  5.调用DestroyCollectiveGroup，清理CollectiveGroup
 *
 * @expect: 预期无异常抛出，rank 0的实例返回值为各个actor中各项之和，40
 */
TEST_F(CollectiveTest, ReduceTest)
{
    std::vector<YR::NamedInstance<CollectiveActor>> instances;
    std::vector<std::string> instanceIDs;
    for (int i = 0; i < 4; ++i) {
        auto ins = YR::Instance(CollectiveActor::FactoryCreate).Invoke();
        instances.push_back(ins);
        instanceIDs.push_back(ins.GetInstanceId());
    }

    std::string groupName = "test-group2";
    YR::Collective::DestroyCollectiveGroup(groupName);
    YR::Collective::CollectiveGroupSpec spec{
        .worldSize = 4,
        .groupName = groupName,
    };
    YR::Collective::CreateCollectiveGroup(spec, instanceIDs, {0, 1, 2, 3});

    std::vector<int> input = {1, 2, 3, 4};
    std::vector<YR::ObjectRef<int>> res;
    for (int i = 0; i < 4; ++i) {
        res.push_back(instances[i]
                          .Function(&CollectiveActor::Reduce)
                          .Invoke(input, groupName, static_cast<uint8_t>(YR::ReduceOp::SUM)));
    }

    EXPECT_EQ(*YR::Get(res[0]), 40);
    YR::Collective::DestroyCollectiveGroup(groupName);
}

/**
 * @title: Send Recv函数调用成功
 * @step:  1.创建CollectiveActor实例
 * @step:  2.调用CreateCollectiveGroup，在driver中创建对应group
 * @step:  3.调用Send函数，传入{1, 2, 3, 4}，另一个actor调用Recv
 *
 * @expect: 预期无异常抛出，返回值为各项之和，10
 */
TEST_F(CollectiveTest, SendRecvTest)
{
    auto ins1 = YR::Instance(CollectiveActor::FactoryCreate).Invoke();
    auto ins2 = YR::Instance(CollectiveActor::FactoryCreate).Invoke();

    std::string groupName = "test-group3";
    YR::Collective::DestroyCollectiveGroup(groupName);
    YR::Collective::CollectiveGroupSpec spec{
        .worldSize = 2,
        .groupName = groupName,
    };
    YR::Collective::CreateCollectiveGroup(spec, {ins1.GetInstanceId(), ins2.GetInstanceId()}, {0, 1});
    std::vector<int> input = {1, 2, 3, 4};
    ins1.Function(&CollectiveActor::Send).Invoke(groupName, input, 1, 1234);
    auto ret = ins2.Function(&CollectiveActor::Recv).Invoke(groupName, 0, 1234, 4);
    EXPECT_EQ(*YR::Get(ret), 10);

    std::vector<int> input2 = {2, 2, 3, 4, 5};
    ret = ins1.Function(&CollectiveActor::Recv).Invoke(groupName, 1, 1234, 5);
    ins2.Function(&CollectiveActor::Send).Invoke(groupName, input2, 0, 1234);
    EXPECT_EQ(*YR::Get(ret), 16);

    YR::Collective::DestroyCollectiveGroup(groupName);
}

/**
 * @title: AllGather函数调用成功
 * @step:  1.创建CollectiveActor实例
 * @step:  2.调用CreateCollectiveGroup，在driver中创建对应group
 * @step:  3.调用CollectiveActor::AllGather函数，传入{1, 2, 3, 4}
 *
 * @expect: 预期无异常抛出，返回值为各个actor中各项之和40, 总数 16个数据
 */
TEST_F(CollectiveTest, AllGatherTest)
{
    std::vector<YR::NamedInstance<CollectiveActor>> instances;
    std::vector<std::string> instanceIDs;
    for (int i = 0; i < 4; ++i) {
        auto ins = YR::Instance(CollectiveActor::FactoryCreate).Invoke();
        instances.push_back(ins);
        instanceIDs.push_back(ins.GetInstanceId());
    }

    std::string groupName = "test-group4";
    YR::Collective::DestroyCollectiveGroup(groupName);
    YR::Collective::CollectiveGroupSpec spec{
        .worldSize = 4,
        .groupName = groupName,
    };
    YR::Collective::CreateCollectiveGroup(spec, instanceIDs, {0, 1, 2, 3});

    std::vector<int> input = {1, 2, 3, 4};
    std::vector<YR::ObjectRef<std::pair<int, int>>> res;
    for (int i = 0; i < 4; ++i) {
        res.push_back(instances[i].Function(&CollectiveActor::AllGather).Invoke(groupName, input));
    }

    for (int i = 0; i < 4; ++i) {
        EXPECT_EQ(YR::Get(res[i])->first, 40);
        EXPECT_EQ(YR::Get(res[i])->second, 16);
    }

    YR::Collective::DestroyCollectiveGroup(groupName);
}

/**
 * @title: Broadcast函数调用成功
 * @step:  1.创建CollectiveActor实例
 * @step:  2.调用CreateCollectiveGroup，在driver中创建对应group
 * @step:  3.调用CollectiveActor::Broadcast，ins1 传入{1, 2, 3, 4, 5}
 *
 * @expect: 预期无异常抛出，返回值为各个actor中各项之和15
 */
TEST_F(CollectiveTest, BroadcastTest)
{
    std::vector<YR::NamedInstance<CollectiveActor>> instances;
    std::vector<std::string> instanceIDs;
    for (int i = 0; i < 4; ++i) {
        auto ins = YR::Instance(CollectiveActor::FactoryCreate).Invoke();
        instances.push_back(ins);
        instanceIDs.push_back(ins.GetInstanceId());
    }

    std::string groupName = "test-group5";
    YR::Collective::DestroyCollectiveGroup(groupName);
    YR::Collective::CollectiveGroupSpec spec{
        .worldSize = 4,
        .groupName = groupName,
        .timeout = 10000,
    };
    YR::Collective::CreateCollectiveGroup(spec, instanceIDs, {0, 1, 2, 3});

    std::vector<int> input = {1, 2, 3, 4, 5};
    std::vector<YR::ObjectRef<int>> res;
    for (int i = 0; i < 4; ++i) {
        res.push_back(instances[i].Function(&CollectiveActor::Broadcast).Invoke(groupName, input, 0));
    }

    for (int i = 0; i < 4; ++i) {
        EXPECT_EQ(*YR::Get(res[i]), 15);
    }

    YR::Collective::DestroyCollectiveGroup(groupName);
}

/**
 * @title: Scatter函数调用成功
 * @step:  1.创建CollectiveActor实例
 * @step:  2.调用CreateCollectiveGroup，在driver中创建对应group
 * @step:  3.调用CollectiveActor::Scatter，ins1 传入{1, 2, 3, 4, 5}
 * @step:  4.调用CollectiveActor::Scatter，ins1 传入{1, 1, 2, 2, 3, 3, 4, 4, 5, 5}
 *
 * @expect: 预期无异常抛出，返回值为各个actor中各自分到一项
 */
TEST_F(CollectiveTest, ScatterTest)
{
    std::vector<YR::NamedInstance<CollectiveActor>> instances;
    std::vector<std::string> instanceIDs;
    for (int i = 0; i < 4; ++i) {
        auto ins = YR::Instance(CollectiveActor::FactoryCreate).Invoke();
        instances.push_back(ins);
        instanceIDs.push_back(ins.GetInstanceId());
    }

    std::string groupName = "test-group6";
    YR::Collective::DestroyCollectiveGroup(groupName);
    YR::Collective::CollectiveGroupSpec spec{
        .worldSize = 4,
        .groupName = groupName,
        .timeout = 1000,
    };
    YR::Collective::CreateCollectiveGroup(spec, instanceIDs, {0, 1, 2, 3});

    std::vector<std::vector<int>> input = {{1}, {2}, {3}, {4}};
    std::vector<YR::ObjectRef<int>> res;
    for (int i = 0; i < 4; ++i) {
        res.push_back(instances[i].Function(&CollectiveActor::Scatter).Invoke(groupName, input, 0, 1));
    }

    for (int i = 0; i < 4; ++i) {
        EXPECT_EQ(*YR::Get(res[i]), i + 1);
    }

    input = {{1, 1}, {2, 2}, {3, 3}, {4, 4}};
    res = {};
    for (int i = 0; i < 4; ++i) {
        res.push_back(instances[i].Function(&CollectiveActor::Scatter).Invoke(groupName, input, 0, 2));
    }

    for (int i = 0; i < 4; ++i) {
        EXPECT_EQ(*YR::Get(res[i]), (i + 1) * 2);
    }

    YR::Collective::DestroyCollectiveGroup(groupName);
}
