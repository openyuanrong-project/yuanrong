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
#include <array>
#include <atomic>
#include <sys/syscall.h>
#include <thread>
#include <vector>
#include "securec.h"
#include "yr/parallel/parallel_for.h"
#include "yr/yr.h"

namespace YR {
namespace Parallel {
const std::string threadNamePrefix;
}
}  // namespace YR
using namespace YR;
using namespace YR::Parallel;

std::vector<uint64_t> g_nodepool;
std::map<std::thread::id, size_t> g_tid_id_map;
std::map<size_t, std::thread::id> g_id_tid_map;
bool g_id_tid_good = true;
std::mutex g_id_tid_check_mu;

const uint32_t g_taskNum = 1000000;
const uint32_t g_threadsNum = 8;
const uint32_t g_chunksize = 100;

class ParallelForLocalTest : public testing::Test {
public:
    ParallelForLocalTest() {}
    ~ParallelForLocalTest() {}
    static void SetUpTestCase()
    {
        Config conf;
        conf.mode = Config::Mode::LOCAL_MODE;
        conf.threadPoolSize = g_threadsNum;
        Init(conf);
    }
    static void TearDownTestCase()
    {
        Finalize();
    }
    void SetUp()
    {
        g_nodepool.resize(g_taskNum);
        g_id_tid_good = true;
        g_tid_id_map.clear();
        g_id_tid_map.clear();
    }
    void TearDown()
    {
        g_nodepool.clear();
        g_tid_id_map.clear();
        g_id_tid_map.clear();
    }
};

void BodyFun(uint32_t start, uint32_t end)
{
    for (uint32_t i = start; i < end; i++) {
        g_nodepool[i] += i;
    }
}

auto BodyLambda = [](size_t start, size_t end) {
    for (size_t i = start; i < end; i++) {
        g_nodepool[i] += i;
    }
};

class BodyOperator {
public:
    void operator()(size_t start, size_t end) const
    {
        for (size_t i = start; i < end; i++) {
            g_nodepool[i] += i;
        }
    }
};

class BodyClassFun {
public:
    void Fun(uint32_t start, uint32_t end)
    {
        for (uint32_t i = start; i < end; i++) {
            g_nodepool[i] += i;
        }
    }

    static void StaticFun(uint32_t start, uint32_t end)
    {
        for (uint32_t i = start; i < end; i++) {
            g_nodepool[i] += i;
        }
    }
};

// 3 arguments handler
// ctx.id and thread_id need to be one-to-one corresponding
void idTidCheck(size_t ctxid)
{
    std::thread::id tid = std::this_thread::get_id();
    std::lock_guard<std::mutex> lock(g_id_tid_check_mu);
    auto it = g_tid_id_map.find(tid);
    if (it != g_tid_id_map.end() && it->second != ctxid) {
        g_id_tid_good = false;
    }
    auto it2 = g_id_tid_map.find(ctxid);
    if (it2 != g_id_tid_map.end() && it2->second != tid) {
        g_id_tid_good = false;
    }
    g_tid_id_map[tid] = ctxid;
    g_id_tid_map[ctxid] = tid;
    EXPECT_TRUE(g_id_tid_good);
}

void BodyFunWithCtx(uint32_t start, uint32_t end, const Context &ctx)
{
    idTidCheck(ctx.id);
    for (uint32_t i = start; i < end; i++) {
        g_nodepool[i] += i;
    }
}

auto BodyLambdaWithCtx = [](size_t start, size_t end, const Context &ctx) {
    idTidCheck(ctx.id);
    for (size_t i = start; i < end; i++) {
        g_nodepool[i] += i;
    }
};

class BodyOperatorWithCtx {
public:
    void operator()(size_t start, size_t end, const Context &ctx) const
    {
        idTidCheck(ctx.id);
        for (size_t i = start; i < end; i++) {
            g_nodepool[i] += i;
        }
    }
};

class BodyClassFunWithCtx {
public:
    void Fun(uint32_t start, uint32_t end, const Context &ctx)
    {
        idTidCheck(ctx.id);
        for (uint32_t i = start; i < end; i++) {
            g_nodepool[i] += i;
        }
    }

    static void StaticFun(uint32_t start, uint32_t end, const Context &ctx)
    {
        idTidCheck(ctx.id);
        for (uint32_t i = start; i < end; i++) {
            g_nodepool[i] += i;
        }
    }
};

TEST_F(ParallelForLocalTest, CallBodyOperator)
{
    BodyOperator body;
    ParallelFor<uint32_t>(0, g_taskNum, body, g_chunksize);
    for (uint32_t i = 0; i < g_taskNum; i++) {
        EXPECT_EQ(g_nodepool[i], i);
    }
}

TEST_F(ParallelForLocalTest, CallBodyOperatorWithCtx)
{
    BodyOperatorWithCtx body;
    ParallelFor<uint32_t>(0, g_taskNum, body, g_chunksize);
    for (uint32_t i = 0; i < g_taskNum; i++) {
        EXPECT_EQ(g_nodepool[i], i);
    }
}

TEST_F(ParallelForLocalTest, CallBodyFun)
{
    ParallelFor<uint32_t>(0, g_taskNum, &BodyFun, g_chunksize);
    for (uint32_t i = 0; i < g_taskNum; i++) {
        EXPECT_EQ(g_nodepool[i], i);
    }
}

TEST_F(ParallelForLocalTest, CallBodyFunWithCtx)
{
    ParallelFor<uint32_t>(0, g_taskNum, &BodyFunWithCtx, g_chunksize);
    for (uint32_t i = 0; i < g_taskNum; i++) {
        EXPECT_EQ(g_nodepool[i], i);
    }
}

TEST_F(ParallelForLocalTest, CallBodyLambda)
{
    ParallelFor<uint32_t>(0, g_taskNum, BodyLambda, g_chunksize);
    for (uint32_t i = 0; i < g_taskNum; i++) {
        EXPECT_EQ(g_nodepool[i], i);
    }
}

TEST_F(ParallelForLocalTest, CallBodyLambdaWithCtx)
{
    ParallelFor<uint32_t>(0, g_taskNum, BodyLambdaWithCtx, g_chunksize);
    for (uint32_t i = 0; i < g_taskNum; i++) {
        EXPECT_EQ(g_nodepool[i], i);
    }
}

TEST_F(ParallelForLocalTest, CallBodyClassFun)
{
    BodyClassFun body;
    ParallelFor<uint32_t>(
        0, g_taskNum, std::bind(&BodyClassFun::Fun, &body, std::placeholders::_1, std::placeholders::_2), g_chunksize);
    for (uint32_t i = 0; i < g_taskNum; i++) {
        EXPECT_EQ(g_nodepool[i], i);
    }
}

TEST_F(ParallelForLocalTest, CallBodyClassFunWithCtx)
{
    BodyClassFunWithCtx body;
    ParallelFor<uint32_t>(0, g_taskNum,
                          std::bind(&BodyClassFunWithCtx::Fun, &body, std::placeholders::_1, std::placeholders::_2,
                                    std::placeholders::_3),
                          g_chunksize);
    for (uint32_t i = 0; i < g_taskNum; i++) {
        EXPECT_EQ(g_nodepool[i], i);
    }
}

TEST_F(ParallelForLocalTest, CallBodyClassStaticFun)
{
    ParallelFor<uint32_t>(0, g_taskNum, &BodyClassFun::StaticFun, g_chunksize);
    for (uint32_t i = 0; i < g_taskNum; i++) {
        EXPECT_EQ(g_nodepool[i], i);
    }
}

TEST_F(ParallelForLocalTest, CallBodyClassStaticFunWithCtx)
{
    ParallelFor<uint32_t>(0, g_taskNum, &BodyClassFunWithCtx::StaticFun, g_chunksize);
    for (uint32_t i = 0; i < g_taskNum; i++) {
        EXPECT_EQ(g_nodepool[i], i);
    }
}

TEST_F(ParallelForLocalTest, CallBodyLambdaChunkSizeIsBigger)
{
    auto chunksize = g_taskNum + 100;
    ParallelFor<uint32_t>(0, g_taskNum, BodyLambda, chunksize);
    for (uint32_t i = 0; i < g_taskNum; i++) {
        EXPECT_EQ(g_nodepool[i], i);
    }
}

/* Muti threads tests */
const uint32_t g_teamsNum = 4;
std::array<std::vector<uint64_t>, g_teamsNum> g_nodepools;

void ThreadParallelForFunc(int teamid)
{
    auto fun = [&](uint32_t start, uint32_t end) {
        for (uint32_t i = start; i < end; i++) {
            g_nodepools[teamid][i] += i;
        }
    };
    // 1 master and parallelDegree-1 workers
    ParallelFor<uint32_t>(0, g_taskNum, fun, g_chunksize);
    for (uint32_t i = 0; i < g_taskNum; i++) {
        EXPECT_EQ(g_nodepools[teamid][i], i);
    }
}

TEST_F(ParallelForLocalTest, DISABLED_MutiThreadCallBodyOperator)
{
    // allocate global nodes
    for (uint32_t i = 0; i < g_teamsNum; i++) {
        g_nodepools[i].resize(g_taskNum);
    }

    // launch threads to run parallel-for
    std::vector<std::thread> workers;
    for (uint32_t teamid = 0; teamid < g_teamsNum; teamid++) {
        workers.emplace_back(std::thread(&ThreadParallelForFunc, teamid));
    }
    // wait threads done
    for (uint32_t teamid = 0; teamid < workers.size(); teamid++) {
        workers[teamid].join();
    }

    // free memorys
    for (uint32_t i = 0; i < g_teamsNum; i++) {
        g_nodepools[i].clear();
    }
}

TEST_F(ParallelForLocalTest, Test_ParallelFor_Thread_Name_Prefix_Is_yrParallelFor)
{
    auto promise = std::make_shared<std::promise<std::string>>();
    auto future = promise->get_future();
    auto verifyThreadNamePrefix = [&promise](uint32_t start, uint32_t end) {
        auto length = pthread_getname_np(pthread_self(), nullptr, 0);
        if (length <= 0) {
            promise->set_exception(std::make_exception_ptr(YR::Exception("length < 0")));
        }
        char *name = new char[length + 1];
        memset_s(name, length + 1, 0, length + 1);
        if (pthread_getname_np(pthread_self(), name, length + 1) != 0) {
            delete[] name;
            promise->set_exception(std::make_exception_ptr(YR::Exception("get name failed")));
        }
        std::string threadName(name);
        delete[] name;
        promise->set_value(threadName);
    };
    ParallelFor<uint32_t>(0, 1, verifyThreadNamePrefix, 1);
    EXPECT_EQ(threadNamePrefix, future.get().substr(0, threadNamePrefix.length()));
}

TEST_F(ParallelForLocalTest, When_Worker_Thread_Size_Is_One_And_TaskNum_Is_Not_One_Should_Do_Ok)
{
    const int chunkSize = 1;
    const int want = 100;
    const int length = 2;
    std::vector<uint64_t> arr;
    arr.resize(length);
    auto useStart = [&arr](size_t start, size_t end) { arr[start] = want; };
    ParallelFor<uint32_t>(0, length, useStart, chunkSize, 1);
    for (size_t i = 0; i < length; i++) {
        EXPECT_EQ(arr[i], want);
    }
}

TEST_F(ParallelForLocalTest, When_Input_UINT32_MAX_Should_Do_Ok)
{
    std::atomic<size_t> get;
    get.store(0);
    size_t want = 1;
    auto addLambda = [&get](size_t start, size_t end) { get++; };
    ParallelFor<size_t>(0, UINT32_MAX, addLambda, UINT32_MAX);
    EXPECT_EQ(want, get);
}

TEST_F(ParallelForLocalTest, NestedParallelFor)
{
    size_t n = 5;
    std::mutex mu;
    size_t cnt = 0;
    ParallelFor<size_t>(0, n, [n, &mu, &cnt](size_t i1, size_t j1) {
            for (auto i = i1; i < j1; i++) {
                ParallelFor<size_t>(0, n, [i, n, &mu, &cnt](size_t i2, size_t j2) {
                        for (auto j = i2; j < j2; j++) {
                            ParallelFor<size_t>(0, n, [i, j, n, &mu, &cnt](size_t i3, size_t j3) {
                                    for (auto k = i3; k < j3; k++) {
                                        std::lock_guard<std::mutex> lk(mu);
                                        std::cout << cnt++ << ": [" << syscall(SYS_gettid) << "] "
                                                  << i << " " << j << " " << k << std::endl;
                                    }
                                }, 1);
                        }
                    }, 1);
            }
        }, 1);
}
