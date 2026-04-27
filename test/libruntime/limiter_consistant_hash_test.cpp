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

#include <gmock/gmock.h>
#include <gtest/gtest.h>

#include "src/libruntime/invoke_spec.h"
#include "src/libruntime/utils/utils.h"
#include "src/utility/logger/logger.h"
#define private public
#include "src/libruntime/invokeadaptor/limiter_consistant_hash.h"

namespace YR {
namespace test {
using namespace YR::Libruntime;
using namespace YR::utility;

class LimiterCsHashTest : public ::testing::Test {
public:
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
        std::shared_ptr<LoadBalancer> lb(LoadBalancer::Factory(LoadBalancerType::ConsistantRoundRobin));
        limiterHash = std::make_shared<LimiterCsHash>(lb);
    }

    void TearDown() override
    {
        limiterHash.reset();
    }
    std::shared_ptr<LimiterCsHash> limiterHash;
};

TEST_F(LimiterCsHashTest, AddTest)
{
    limiterHash->Add("schedulerName", "schedulerId");
    EXPECT_FALSE(limiterHash->Next("funcId", false).empty());
    EXPECT_EQ(limiterHash->Next("funcId"), "schedulerId");

    limiterHash->Add("schedulerName", "schedulerId1");
    EXPECT_EQ(limiterHash->Next("funcId", false), "schedulerId1");
}

TEST_F(LimiterCsHashTest, RemoveTest)
{
    limiterHash->Add("schedulerName", "schedulerId");
    limiterHash->Remove("schedulerName");
    // remove after add, hash has no scheduler, then res is empty
    EXPECT_TRUE(limiterHash->Next("funcId").empty());

    limiterHash->Add("schedulerName", "schedulerId");
    // add schedulername <-> schedulerid, then res is schedulerid
    EXPECT_EQ(limiterHash->Next("funcId", false), "schedulerId");
    limiterHash->Add("schedulerName", "anotherSchedulerId");
    // update schedulername <-> anotherSchedulerId, then res is anotherSchedulerId
    EXPECT_EQ(limiterHash->Next("funcId", false), "anotherSchedulerId");

    limiterHash->Add("schedulername_1", "anotherSchedulerId_1");
    // add new schedulername_1 <-> anotherSchedulerId_1, if move flag is false, then res is anotherSchedulerId,
    // otherwise need move to next anchorpoint and res is anotherSchedulerId_1
    EXPECT_EQ(limiterHash->Next("funcId", false), "anotherSchedulerId");
    EXPECT_EQ(limiterHash->Next("funcId", true), "anotherSchedulerId_1");

    limiterHash->Remove("schedulerName");
    // remove schedulername <-> schedulerid, hash has only schedulername_1 <-> anotherSchedulerId_1, then  Whether move
    // is false or true, the result is always anotherSchedulerId_1
    EXPECT_EQ(limiterHash->Next("funcId", false), "anotherSchedulerId_1");
    EXPECT_EQ(limiterHash->Next("funcId", true), "anotherSchedulerId_1");
}

TEST_F(LimiterCsHashTest, RemoveTest2)
{
    limiterHash->Add("schedulerName", "schedulerId");
    limiterHash->Add("schedulerName1", "schedulerId1");
    limiterHash->Add("schedulerName2", "schedulerId2");
    auto schedulerID = limiterHash->Next("funcId", true);
    std::cout << "schedulerid is : " << schedulerID << std::endl;
    EXPECT_EQ(schedulerID.empty(), false);
    std::vector<std::shared_ptr<SchedulerInstance>> vec{
        {std::make_shared<SchedulerInstance>(SchedulerInstance{
             .InstanceID = "schedulerId", .updateTime = YR::GetCurrentTimestampNs(), .isAvailable = false}),
         std::make_shared<SchedulerInstance>(SchedulerInstance{
             .InstanceID = "schedulerId1", .updateTime = YR::GetCurrentTimestampNs(), .isAvailable = false}),
         std::make_shared<SchedulerInstance>(SchedulerInstance{
             .InstanceID = "schedulerId2", .updateTime = YR::GetCurrentTimestampNs(), .isAvailable = false})}};
    auto schedulerInfo = std::make_shared<AvailableSchedulerInfos>();
    schedulerInfo->schedulerInstanceList = vec;
    EXPECT_EQ(limiterHash->Next("funcId", schedulerInfo, true), ALL_SCHEDULER_UNAVAILABLE);
}

TEST_F(LimiterCsHashTest, RemoveAllTest)
{
    limiterHash->Add("schedulerName0", "schedulerId0");
    limiterHash->Add("schedulerName1", "schedulerId1");
    limiterHash->Add("schedulerName2", "schedulerId2");
    limiterHash->Add("schedulerName3", "schedulerId3");
    limiterHash->RemoveAll();
    EXPECT_TRUE(limiterHash->Next("scheduleId").empty());
}

TEST_F(LimiterCsHashTest, NextRetryTest)
{
    limiterHash->Add("schedulerName", "");
    limiterHash->Add("schedulerName1", "schedulerId1");
    auto schedulerId = limiterHash->NextRetry("scheduleId");
    std::cout << "scheduler id is " << schedulerId << std::endl;
    EXPECT_TRUE(!schedulerId.empty());
}

TEST_F(LimiterCsHashTest, NextTest1)
{
    limiterHash->Add("schedulerName", "");
    auto schedulerId = limiterHash->NextRetry("scheduleId");
    std::cout << "scheduler id1 is : " << schedulerId << std::endl;
    EXPECT_TRUE(schedulerId.empty());
    // 添加 "schedulerName1" <-> "schedulerId1", 返回schedulerId1
    limiterHash->Add("schedulerName1", "schedulerId1");
    schedulerId = limiterHash->NextRetry("func1", true);
    std::cout << "scheduler id2 is : " << schedulerId << std::endl;
    EXPECT_TRUE(!schedulerId.empty());
    // 入参中增加"schedulerId1", 更新时间晚于add时间，返回"AllSchedulerUnavailable"
    auto updateTime = YR::GetCurrentTimestampNs();
    auto schedulerInfo = std::make_shared<AvailableSchedulerInfos>();
    std::vector<std::shared_ptr<SchedulerInstance>> vec1{std::make_shared<SchedulerInstance>(
        SchedulerInstance{.InstanceID = "schedulerId1", .updateTime = updateTime, .isAvailable = false})};
    schedulerInfo->schedulerInstanceList = vec1;
    schedulerId = limiterHash->NextRetry("func1", schedulerInfo, true);
    std::cout << "scheduler id2 is : " << schedulerId << std::endl;
    EXPECT_EQ(schedulerId, "AllSchedulerUnavailable");
    // 重新添加"schedulerName1" <-> "schedulerId1", hash环中add时间更新，返回"schedulerId1"
    std::vector<std::shared_ptr<SchedulerInstance>> vec2{std::make_shared<SchedulerInstance>(
        SchedulerInstance{.InstanceID = "schedulerId1", .updateTime = updateTime, .isAvailable = false})};
    limiterHash->Add("schedulerName1", "schedulerId1");
    schedulerInfo->schedulerInstanceList = vec2;
    schedulerId = limiterHash->NextRetry("func1", schedulerInfo, true);
    std::cout << "scheduler id3 is : " << schedulerId << std::endl;
    EXPECT_TRUE(schedulerId != "AllSchedulerUnavailable");
    for (auto scheduler : schedulerInfo->schedulerInstanceList) {
        EXPECT_TRUE(scheduler->updateTime > updateTime);
        EXPECT_TRUE(scheduler->isAvailable);
    }
    // 再次Next获取，更新入参vec中的update time，返回"AllSchedulerUnavailable"
    std::vector<std::shared_ptr<SchedulerInstance>> vec{std::make_shared<SchedulerInstance>(SchedulerInstance{
        .InstanceID = "schedulerId1", .updateTime = YR::GetCurrentTimestampNs(), .isAvailable = false})};
    schedulerInfo->schedulerInstanceList = vec;
    schedulerId = limiterHash->NextRetry("func1", schedulerInfo, true);
    std::cout << "scheduler id3 is : " << schedulerId << std::endl;
    EXPECT_TRUE(schedulerId == "AllSchedulerUnavailable");
    limiterHash->RemoveAll();
}

TEST_F(LimiterCsHashTest, NextTest2)
{
    std::unordered_map<std::string, std::string> idMap = {
        {"schedulerId1", "schedulerName1"},
        {"schedulerId2", "schedulerName2"},
        {"schedulerId3", "schedulerName3"},
    };
    limiterHash->Add(idMap["schedulerId1"], "schedulerId1");
    limiterHash->Add(idMap["schedulerId2"], "schedulerId2");
    limiterHash->Add(idMap["schedulerId3"], "schedulerId3");
    auto spec = std::make_shared<YR::Libruntime::InvokeSpec>();
    auto schedulerId4 = limiterHash->NextRetry("func1", true);
    std::cout << "scheduler id4 is : " << schedulerId4 << std::endl;
    EXPECT_TRUE(!schedulerId4.empty());
    EXPECT_TRUE(schedulerId4 != "AllSchedulerUnavailable");
    spec->schedulerInfos->schedulerInstanceList.push_back(std::make_shared<SchedulerInstance>(SchedulerInstance{
        .InstanceID = schedulerId4, .updateTime = YR::GetCurrentTimestampNs(), .isAvailable = false}));
    auto schedulerId5 = limiterHash->NextRetry("func1", spec->schedulerInfos, true);
    std::cout << "scheduler id5 is : " << schedulerId5 << std::endl;
    EXPECT_TRUE(!schedulerId5.empty());
    EXPECT_TRUE(schedulerId5 != "AllSchedulerUnavailable");
    EXPECT_TRUE(schedulerId5 != schedulerId4);
    EXPECT_EQ(spec->schedulerInfos->schedulerInstanceList.size(), 2);
    EXPECT_TRUE(spec->schedulerInfos->schedulerInstanceList[1]->isAvailable);

    spec->schedulerInfos->schedulerInstanceList[1]->isAvailable = false;
    auto schedulerId6 = limiterHash->NextRetry("func1", spec->schedulerInfos, true);
    std::cout << "scheduler id6 is : " << schedulerId6 << std::endl;
    EXPECT_TRUE(!schedulerId6.empty());
    EXPECT_TRUE(schedulerId6 != "AllSchedulerUnavailable");
    EXPECT_TRUE(schedulerId5 != schedulerId6);
    EXPECT_EQ(spec->schedulerInfos->schedulerInstanceList.size(), 3);
    for (auto &scheduler : spec->schedulerInfos->schedulerInstanceList) {
        if (scheduler->InstanceID == schedulerId6) {
            EXPECT_TRUE(scheduler->isAvailable);
            scheduler->isAvailable = false;
        }
    }
    auto schedulerI7 = limiterHash->NextRetry("func1", spec->schedulerInfos, true);
    std::cout << "scheduler id7 is : " << schedulerI7 << std::endl;
    EXPECT_TRUE(!schedulerI7.empty());
    EXPECT_TRUE(schedulerI7 == "AllSchedulerUnavailable");
    for (auto scheduler : spec->schedulerInfos->schedulerInstanceList) {
        if (scheduler->InstanceID == schedulerId5) {
            limiterHash->Add(idMap[schedulerId5], schedulerId5);
        }
    }
    auto schedulerI8 = limiterHash->NextRetry("func1", spec->schedulerInfos, true);
    std::cout << "scheduler id8 is : " << schedulerI8 << std::endl;
    EXPECT_TRUE(!schedulerI8.empty());
    EXPECT_TRUE(schedulerI8 != "AllSchedulerUnavailable");
}

TEST_F(LimiterCsHashTest, ResetAllTest)
{
    limiterHash->Add("schedulerName1", "schedulerId1");
    limiterHash->Add("schedulerName2", "schedulerId2");
    limiterHash->Add("schedulerName3", "schedulerId3");

    auto vec = std::vector<SchedulerInstance>{
        SchedulerInstance{.InstanceName = "schedulerName1", .InstanceID = "schedulerId1"},
        SchedulerInstance{.InstanceName = "schedulerName2", .InstanceID = "schedulerId2"},
        SchedulerInstance{.InstanceName = "schedulerName3", .InstanceID = "schedulerId3"}};
    limiterHash->ResetAll(vec, 0);
    EXPECT_TRUE(limiterHash->schedulerInfoMap.find("schedulerName3") != limiterHash->schedulerInfoMap.end());

    auto vec1 = std::vector<SchedulerInstance>{
        SchedulerInstance{.InstanceName = "schedulerName1", .InstanceID = "schedulerId1"},
        SchedulerInstance{.InstanceName = "schedulerName2", .InstanceID = "schedulerId2"}};
    limiterHash->ResetAll(vec1, 0);
    EXPECT_TRUE(limiterHash->schedulerInfoMap.find("schedulerName3") == limiterHash->schedulerInfoMap.end());
}
}  // namespace test
}  // namespace YR
