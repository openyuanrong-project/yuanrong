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

#include <unordered_map>
#include "iostream"

#include <gmock/gmock.h>
#include <gtest/gtest.h>

#include "src/libruntime/invokeadaptor/load_balancer.h"
#include "src/libruntime/utils/utils.h"
#include "src/utility/logger/logger.h"

using namespace testing;
using namespace YR::Libruntime;
using namespace YR::utility;

namespace YR {
namespace test {
class LoadBalancerTest : public ::testing::Test {
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
    }

    void TearDown() override {}
};  // namespace test

TEST_F(LoadBalancerTest, WRRTest)
{
    LoadBalancer *lb = LoadBalancer::Factory(LoadBalancerType::WeightedRoundRobin);
    EXPECT_TRUE(lb != nullptr);
    {
        std::vector<std::pair<std::string, int>> nodes = {
            {"a", 40}, {"b", 10}, {"c", 10}, {"d", 10}, {"e", 10},
        };
        std::unordered_map<std::string, int> counter;

        int total = 0;
        for (auto &n : nodes) {
            lb->Add(n.first, n.second);
            total += n.second;
            counter[n.first] = 0;
        }

        for (int i = 0; i < total; ++i) {
            auto n = lb->Next("");
            counter[n]++;
        }

        for (auto &n : nodes) {
            ASSERT_EQ(n.second, counter[n.first]);
        }
    }

    lb->RemoveAll();

    {
        std::vector<std::pair<std::string, int>> nodes = {
            {"ax", 40},
            {"bx", 10},
        };
        std::unordered_map<std::string, int> counter;

        int total = 0;
        for (auto &n : nodes) {
            lb->Add(n.first, n.second);
            total += n.second;
            counter[n.first] = 0;
        }

        for (int i = 0; i < total; ++i) {
            auto n = lb->Next("");
            counter[n]++;
        }

        for (auto &n : nodes) {
            ASSERT_EQ(n.second, counter[n.first]);
        }
    }

    delete lb;
}

TEST_F(LoadBalancerTest, CsHashRoundRobinTest)
{
    LoadBalancer *lb = LoadBalancer::Factory(LoadBalancerType::ConsistantRoundRobin);
    lb->Add("scheduler1", 0);
    lb->Add("scheduler2", 0);
    lb->Add("scheduler3", 0);

    auto res1 = lb->Next("function1", false);
    auto res2 = lb->Next("function1", false);
    ASSERT_EQ(res1, res2);
    std::cout << "res 1 is " << res1 << std::endl;

    auto res3 = lb->Next("function1", true);
    std::cout << "res 3 is " << res3 << std::endl;
    EXPECT_NE(res3, res2);

    auto res4 = lb->Next("function1", true);
    std::cout << "res 4 is " << res4 << std::endl;
    EXPECT_NE(res3, res4);

    lb->RemoveAll();
    auto res6 = lb->Next("function1", false);
    std::cout << "res 6 is " << res6 << std::endl;
    ASSERT_EQ(res6.find("scheduler3") != std::string::npos, false);

    lb->Add("scheduler1", 0);
    lb->Add("scheduler2", 0);
    lb->Add("scheduler3", 0);

    auto res7 = lb->Next("function1", false);
    std::cout << "res 7 is " << res7 << std::endl;

    auto res8 = lb->Next("function1", false);
    std::cout << "res 8 is " << res8 << std::endl;
    ASSERT_EQ(res7, res8);

    auto res9 = lb->Next("function1", true);
    std::cout << "res 9 is " << res9 << std::endl;
    EXPECT_NE(res9, res8);

    auto res10 = lb->Next("function1", true);
    std::cout << "res 10 is " << res10 << std::endl;
    EXPECT_NE(res9, res10);

    delete lb;
}

TEST_F(LoadBalancerTest, CsHashRoundRobinRemoveTest)
{
    LoadBalancer *lb = LoadBalancer::Factory(LoadBalancerType::ConsistantRoundRobin);
    lb->Add("scheduler1", 0);
    auto res1 = lb->Next("function1", true);
    ASSERT_EQ(res1.find("scheduler1") != std::string::npos, true);

    lb->Remove("scheduler1");
    lb->Add("scheduler2", 0);
    auto res3 = lb->Next("function1", true);
    ASSERT_EQ(res3.find("scheduler2") != std::string::npos, true);

    auto res4 = lb->Next("function1", false);
    ASSERT_EQ(res4.find("scheduler2") != std::string::npos, true);

    delete lb;
}
}  // namespace test
}  // namespace YR