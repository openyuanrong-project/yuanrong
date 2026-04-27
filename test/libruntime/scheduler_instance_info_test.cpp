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
#include <chrono>
#include <future>

#include <gmock/gmock.h>
#include <gtest/gtest.h>
#include <json.hpp>
#include <string>

#include "src/libruntime/invokeadaptor/scheduler_instance_info.h"

namespace YR {
namespace test {
class SchedulerInstanceInfoTest : public ::testing::Test {
    void SetUp() override {}

    void TearDown() override {}
};

std::string g_scheduler_info = R"(
{
    "schedulerFuncKey": "0/faasscheduler/0",
    "schedulerInstanceList": [
        {
            "instanceName": "faasscheduler0",
            "instanceId": "faasscheduler0-id"
        },
        {
            "instanceName": "faasscheduler1",
            "instanceId": "faasscheduler1-id"
        },
        {
            "instanceName": "faasscheduler2",
            "instanceId": "faasscheduler2-id"
        }
    ]
}
)";

TEST_F(SchedulerInstanceInfoTest, ParseSchedulerInfoTest)
{
    Libruntime::SchedulerInfo schedulerInfo;
    ParseSchedulerInfo(g_scheduler_info, schedulerInfo);

    ASSERT_EQ(schedulerInfo.schedulerFuncKey, "0/faasscheduler/0");
    ASSERT_EQ(schedulerInfo.schedulerInstanceList.size(), 3);
    ASSERT_EQ(schedulerInfo.schedulerInstanceList[0].InstanceName, "faasscheduler0");
    ASSERT_EQ(schedulerInfo.schedulerInstanceList[0].InstanceID, "faasscheduler0-id");

    ASSERT_EQ(schedulerInfo.schedulerInstanceList[1].InstanceName, "faasscheduler1");
    ASSERT_EQ(schedulerInfo.schedulerInstanceList[1].InstanceID, "faasscheduler1-id");

    ASSERT_EQ(schedulerInfo.schedulerInstanceList[2].InstanceName, "faasscheduler2");
    ASSERT_EQ(schedulerInfo.schedulerInstanceList[2].InstanceID, "faasscheduler2-id");

    auto err = ParseSchedulerInfo("err json", schedulerInfo);
    ASSERT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_PARAM_INVALID);
}
}
}