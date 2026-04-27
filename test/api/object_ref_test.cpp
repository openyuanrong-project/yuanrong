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
#include <iostream>
#include "yr/yr.h"
#include "src/utility/logger/logger.h"
#include "src/libruntime/invokeadaptor/task_submitter.h"

namespace YR {
namespace test {
using namespace testing;
using namespace YR::utility;
using namespace YR::Libruntime;

class MockFsIntf2 : public FSIntf {
public:
    MockFsIntf2() {};
    ~MockFsIntf2(){};
    void GroupCreateAsync(const CreateRequests &reqs, CreateRespsCallback respCallback, int timeoutSec = -1) override{};
    void CreateAsync(const CreateRequest &req, CreateCallBack callback, std::string &instanceId) override{};
    YR::Status Start(const std::string &jobId) override
    {
        return YR::Status{};
    };
    void Stop(void) override {};
    void InvokeAsync(const InvokeRequest &req, InvokeCallBack callback, int) override {};
    void CallResultAsync(const CallResult &req, CallResultCallBack callback) override {};
    void KillAsync(const KillRequest &req, KillCallBack callback) override {};
    void ExitAsync(const ExitRequest &req, ExitCallBack callback) override {};
    void StateSaveAsync(const StateSaveRequest &req, StateSaveCallBack callback) override {};
    void StateLoadAsync(const StateLoadRequest &req, StateLoadCallBack callback) override {};
    bool IsHealth() override
    {
        return true;
    };
};

class ObjectRefTest : public testing::Test {
public:
    ObjectRefTest(){};
    ~ObjectRefTest(){};
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
};

//  TODO. This API layer testcase cannot work on well.
TEST_F(ObjectRefTest, PutGetTest)
{
    YR::Config conf;
    conf.inCluster = true;
    conf.functionUrn = "sn:cn:yrk:default:function:0-x-x:$latest";
    conf.serverAddr = "10.1.1.1:12345";
    conf.dataSystemAddr = "10.1.1.1:12346";
    YR::Init(conf);

    int a = 42;
    ObjectRef<int> objRef1 = YR::Put(a);
    ObjectRef<int> objRef2 = std::move(objRef1);

    std::cout << "end0" << objRef1.ID() << std::endl;
    YR::Finalize();
}

}  // namespace test
}  // namespace YR
