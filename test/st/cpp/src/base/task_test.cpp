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
#include <stdlib.h>
#include <string>

#include "base/utils.h"
#include "gmock/gmock.h"
#include "gtest/gtest.h"
#include "user_common_func.h"
#include "yr/yr.h"

using testing::HasSubstr;

class TaskTest : public testing::Test {
public:
    TaskTest() {};
    ~TaskTest() {};
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

/*case
 * @title: task函数调用成功
 * @precondition:
 * @step:  调用openYuanRong task 函数 AddOne
 * @expect:  1.预期无异常抛出，
 * @expect:  2.返回值为2
 */
TEST_F(TaskTest, InvokeSuccessfully)
{
    auto ret = YR::Function(&AddOne).Invoke(1);
    EXPECT_EQ(*YR::Get(ret), 2);
}

TEST_F(TaskTest, InvokeDirectReturnBig)
{
    const std::vector<char> bigArgs(101 * 1024, 'a');
    for (int i = 0; i < 10; i++) {
        auto ret = YR::Function(&BigBox).Invoke(bigArgs);
        ASSERT_EQ(*YR::Get(ret, 10), bigArgs);
    }
}
/*case
 * @title: 以不同资源请求进行task调用
 * @precondition:
 * @step:  1.[cpu mem] = [300 500] 请求资源运行task函数，发起8次调用
 * @step:  2.以默认[cpu mem] = [500 500] 请求资源运行task函数，发起8次调用
 * @expect:  1.预期无异常抛出
 * @expect:  2.返回值为2
 */
TEST_F(TaskTest, InvokeSuccessfullyWithDifferentResource)
{
    auto start = std::chrono::steady_clock::now();
    std::vector<YR::ObjectRef<int>> rets;
    YR::InvokeOptions option;
    option.cpu = 300;
    option.memory = 500.0;
    for (int i = 0; i < 8; i++) {
        rets.push_back(YR::Function(&AddAfterSleep).Options(option).Invoke(1));
    }
    for (int i = 0; i < 8; i++) {
        rets.push_back(YR::Function(&AddAfterSleep).Invoke(1));
    }
    auto x = YR::Get(rets);
    auto end = std::chrono::steady_clock::now();
    auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(end - start);
    std::cout << "invoke cost time: " << duration.count() << "ms" << std::endl;
    EXPECT_EQ(*x[0], 2);
}

/*case
 * @title: 非法的资源请求进行task调用
 * @precondition:
 * @step:  1. 请求资源cpu=1，进行task函数调用
 * @step:  2. 请求资源memory=1，进行task函数调用
 * @expect:  1.预期成功
 * @expect:  2.预期成功
 */
TEST_F(TaskTest, TestResource)
{
    // cpu
    YR::InvokeOptions option;
    option.cpu = 1.0;
    try {
        auto r1 = YR::Function(&AddOne).Options(std::move(option)).Invoke(2);
        auto res = *YR::Get(r1);
    } catch (YR::Exception &e) {
        ASSERT_EQ(0, 1);
    }
    // memory
    YR::InvokeOptions optionMem;
    optionMem.memory = 1.0;
    try {
        auto r1 = YR::Function(&AddOne).Options(std::move(optionMem)).Invoke(2);
        auto res = *YR::Get(r1);
    } catch (YR::Exception &e) {
        ASSERT_EQ(0, 1);
    }
}

/*case
 * @title: 1k次task并发调用
 * @precondition:
 * @step:  并发1k次task函数调用
 * @expect:  1.预期无异常抛出
 * @expect:  2.返回值为2
 */
TEST_F(TaskTest, DISABLED_Invoke1kSuccessfully)
{
    auto start = std::chrono::steady_clock::now();
    std::vector<YR::ObjectRef<int>> rets;
    for (int i = 0; i < 1000; i++) {
        rets.push_back(YR::Function(&Add).Invoke(1, 1));
    }
    auto x = YR::Get(rets);
    auto end = std::chrono::steady_clock::now();
    auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(end - start);
    std::cout << "invoke cost time: " << duration.count() << "ms" << std::endl;
    EXPECT_EQ(*x[0], 2);
}

/*case
 * @title: 配置task实例并发度为5 task并发调用5
 * @precondition:
 * @step:  1. 配置task实例并发度为5
 * @step:  2. task并发调用5
 * @expect:  1.预期无异常抛出
 * @expect:  2.返回值为6
 */
TEST_F(TaskTest, ConcurrencyInvokeMulti)
{
    printf("=====注册函数，云下调用,并发度为5，发送5个请求\n");
    YR::InvokeOptions option;
    option.customExtensions.insert({YR::CONCURRENCY_KEY, "5"});
    std::vector<YR::ObjectRef<int>> rets;
    for (int i = 0; i < 5; i++) {
        rets.push_back(YR::Function(&AddOne).Options(option).Invoke(5));
    }
    auto x = YR::Get(rets);
    EXPECT_EQ(*x[0], 6) << "YR Get failed，expect result : 7";
}

/*case
 * @title: 配置非法的并发度调用task函数
 * @precondition:
 * @step:  1. 配置task实例并发度为0，并调用函数
 * @step:  2. 配置task实例并发度为101，并调用函数
 * @step:  3. 配置task实例并发度为-1，并调用函数
 * @expect:  1.预期异常抛出
 * @expect:  2.预期异常抛出
 * @expect:  3.预期异常抛出
 */
TEST_F(TaskTest, InvalidConcurrency)
{
    printf("====设置无效concurrency====\n");
    YR::InvokeOptions option;
    try {
        option.customExtensions.insert({YR::CONCURRENCY_KEY, "0"});
        auto r1 = YR::Function(&AddOne).Options(option).Invoke(1);
        int r2 = *YR::Get(r1);
    } catch (YR::Exception &e) {
        printf("Exception:%s,\n", e.what());
        std::string errorCode = "1001";
        std::string errorMsg = "invalid opts concurrency";
        std::string excepMsg = e.what();
        std::cout << "exception: " << excepMsg << std::endl;
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
    try {
        option.customExtensions[YR::CONCURRENCY_KEY] = "101";
        auto r3 = YR::Function(&AddOne).Options(option).Invoke(1);
        int r4 = *YR::Get(r3);
    } catch (YR::Exception &e) {
        printf("Exception:%s,\n", e.what());
        std::string errorCode = "1001";
        std::string errorMsg = "invalid opts concurrency";
        std::string excepMsg = e.what();
        std::cout << "exception: " << excepMsg << std::endl;
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
    try {
        option.customExtensions[YR::CONCURRENCY_KEY] = "-1";
        auto r5 = YR::Function(&AddOne).Options(option).Invoke(1);
        int r6 = *YR::Get(r5);
    } catch (YR::Exception &e) {
        printf("Exception:%s,\n", e.what());
        std::string errorCode = "1001";
        std::string errorMsg = "invalid opts concurrency";
        std::string excepMsg = e.what();
        std::cout << "exception: " << excepMsg << std::endl;
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
    std::cout << "test case end" << std::endl;
}

/*case
 * @title: task 函数Add 依赖函数AddAfterSleep的输出
 * @precondition:
 * @step:  1. 调用AddAfterSleep
 * @step:  2. 以1的objref作为参数，调用Add
 * @expect:  返回值为4
 */
TEST_F(TaskTest, DependentOneFuncRetRef)
{
    auto r1 = YR::Function(&AddAfterSleep).Invoke(1);
    auto r2 = YR::Function(&Add).Invoke(r1, 2);
    int n = *YR::Get(r2);
    EXPECT_EQ(n, 4) << "case run failed! expect result : 4";
}

/*case
 * @title: task 函数Add 依赖函数AddAfterSleep和AddTwo的输出
 * @precondition:
 * @step:  1. 调用AddAfterSleep
 * @step:  2. 调用AddTwo
 * @step:  2. 以1，2的objref作为参数，调用Add
 * @expect:  返回值为6
 */
TEST_F(TaskTest, DependentTwoFuncRetRef)
{
    auto r1 = YR::Function(&AddAfterSleep).Invoke(1);
    auto r2 = YR::Function(&AddTwo).Invoke(2);
    auto r3 = YR::Function(&Add).Invoke(r1, r2);
    int n = *YR::Get(r3);
    EXPECT_EQ(n, 6) << "case run failed! expect result : 6";
}

/*case
 * @title: task 函数Add 依赖函数RaiseRuntimeError和AddTwo的输出
 * @precondition:
 * @step:  1. RaiseRuntimeError
 * @step:  2. 调用AddTwo
 * @step:  2. 以1，2的objref作为参数，调用Add
 * @expect:  返回值为6
 */
TEST_F(TaskTest, DependentTwoFuncRetRefError)
{
    auto r1 = YR::Function(&RaiseRuntimeError).Invoke();
    auto r2 = YR::Function(&AddTwo).Invoke(2);
    auto r3 = YR::Function(&Add).Invoke(r1, r2);
    ASSERT_THROW(YR::Get(r3), YR::Exception);
}

/*case
 * @title: 串行依赖  a->b->c
 * @precondition:
 * @step:  1. AddOne
 * @step:  2. 以1的返回objref作为入参，调用AddTwo
 * @step:  3. 以2的返回objref作为入参，调用AddTwo
 * @expect:  预期无异常抛出
 */
TEST_F(TaskTest, DependentMutliRef)
{
    // 三次依赖
    YR::ObjectRef<int> ret = YR::Function(&AddOne).Invoke(1);
    YR::ObjectRef<int> ret2 = YR::Function(&AddOne).Invoke(ret);
    YR::ObjectRef<int> ret3 = YR::Function(&AddOne).Invoke(ret2);
    ASSERT_NO_THROW(YR::Get(ret3));
}

/*case
 * @title: 串行依赖  a->b->c  a报错
 * @precondition:
 * @step:  1. RaiseRuntimeError
 * @step:  2. 以1的返回objref作为入参，调用AddOne
 * @step:  3. 以2的返回objref作为入参，调用AddOne
 * @expect:  预期异常抛出
 */
TEST_F(TaskTest, DependentMutliRefError)
{
    // 三次依赖，1次报错
    YR::ObjectRef<int> ret = YR::Function(&RaiseRuntimeError).Invoke();
    YR::ObjectRef<int> ret2 = YR::Function(&AddOne).Invoke(ret);
    YR::ObjectRef<int> ret3 = YR::Function(&AddOne).Invoke(ret2);
    ASSERT_THROW(YR::Get(ret3), YR::Exception);
}

/*case
 * @title: 同一个返回值依赖两次
 * @precondition:
 * @step:  1. AddOne
 * @step:  2. 以1的返回objref作为入参，调用AddOne
 * @step:  3. 以1的返回objref作为入参，调用AddOne
 * @expect:  预期返回值为3
 */
TEST_F(TaskTest, DependentSameRef)
{
    YR::InvokeOptions option;
    option.customExtensions.insert({"GRACEFUL_SHUTDOWN_TIME", "1"});
    YR::ObjectRef<int> ret = YR::Function(&AddOne).Options(option).Invoke(1);
    YR::ObjectRef<int> ret2 = YR::Function(&AddOne).Options(option).Invoke(ret);
    YR::ObjectRef<int> ret3 = YR::Function(&AddOne).Options(option).Invoke(ret);
    int n = *YR::Get(ret2);
    int m = *YR::Get(ret3);
    EXPECT_EQ(n, 3) << "case run failed! expect result : 3";
    EXPECT_EQ(n, m);
}

/*case
 * @title: 同一个错误返回值依赖两次
 * @precondition:
 * @step:  1. AddOne
 * @step:  2. 以1的返回objref作为入参，调用AddOne
 * @step:  3. 以1的返回objref作为入参，调用AddOne
 * @expect:  预期异常抛出
 */
TEST_F(TaskTest, DependentSameErrorRef)
{
    YR::ObjectRef<int> ret = YR::Function(&RaiseRuntimeError).Invoke();
    YR::ObjectRef<int> ret2 = YR::Function(&AddOne).Invoke(ret);
    YR::ObjectRef<int> ret3 = YR::Function(&AddOne).Invoke(ret);
    ASSERT_THROW(YR::Get(ret2), YR::Exception);
    ASSERT_THROW(YR::Get(ret3), YR::Exception);
}

/*case
 * @title: 函数调函数抛出SIGFPE
 * @precondition:
 * @step:  1.调用函数ExcChain
 * @expect:  1.异常抛出
 */
TEST_F(TaskTest, ExceptionChain)
{
    printf("=====云上invoke 错误的算术运算=====\n");
    auto r1 = YR::Function(ExcChain).Invoke();
    try {
        int n1 = *YR::Get(r1);
    } catch (YR::Exception &e) {
        printf("error: %s\n", e.what());
        std::string errorCode = "ErrCode: 2002";
        std::string errorMsg = "SIGFPE";
        std::string excepMsg = e.what();
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
}

/*case
 * @title: 函数调函数抛出SIGABRT
 * @precondition:
 * @step:  1.调用函数Excdying
 * @expect:  1.异常抛出
 */
TEST_F(TaskTest, ExceptionDying)
{
    printf("=====云上invoke 程序的异常终止=====\n");
    auto r1 = YR::Function(Excdying).Invoke();
    try {
        int n1 = *YR::Get(r1);
    } catch (YR::Exception &e) {
        printf("error: %s\n", e.what());
        std::string errorCode = "ErrCode: 2002";
        std::string errorMsg = "SIGABRT";
        std::string excepMsg = e.what();
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
}

/*case
 * @title: 函数调函数抛出异常
 * @precondition:
 * @step:  1.调用函数ExcMethod
 * @expect:  1.异常抛出
 */
TEST_F(TaskTest, ExceptionMethod)
{
    printf("=====云上invoke 用户函数构造异常=====\n");
    auto r1 = YR::Function(ExcMethod).Invoke();
    try {
        int n1 = *YR::Get(r1);
    } catch (YR::Exception &e) {
        printf("error: %s\n", e.what());
        std::string errorCode = "ErrCode: 2002";
        std::string errorMsg = "exception happens when executing user's function";
        std::string excepMsg = e.what();
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
}

/*case
 * @title: 函数调函数,参数包含vector
 * @precondition:
 * @step:  1.调用函数ExcVector
 * @expect:  1.正常返回
 */
TEST_F(TaskTest, ExecWithVector)
{
    printf("=====函数调函数,参数包含vector=====\n");
    YR::InvokeOptions option;
    option.customExtensions.insert({"GRACEFUL_SHUTDOWN_TIME", "1"});
    std::vector<int> nums = {1, 2, 3, 4};
    auto r1 = YR::Function(Sum).Options(option).Invoke(nums);
    EXPECT_EQ(*YR::Get(r1), 10);

    std::vector<YR::ObjectRef<int>> nums2;
    for (int i = 0; i < 10; i++) {
        nums2.push_back(YR::Function(Add).Options(option).Invoke(1, 1));
    }
    auto r2 = YR::Function(SumWithObjectRef).Options(option).Invoke(nums2);
    EXPECT_EQ(*YR::Get(r2), 20);
}

/*case
 * @title: a->b,a调用发出后返回
 * @precondition:
 * @step:  1.调用函数ExcVector
 * @expect:  1.正常返回
 */
TEST_F(TaskTest, ExecWithDirectReturn)
{
    printf("=====a->b,a调用发出后返回=====\n");
    auto r1 = YR::Function(DirectReturn).Invoke();
    EXPECT_EQ((*YR::Get(*YR::Get(r1))[0]), 2);
}

/*case
 * @title: a->b,a调用发出后返回
 * @precondition:
 * @step:  1.调用函数ExcVector
 * @expect:  1.正常返回
 */
TEST_F(TaskTest, DISABLED_PutObjWithObjectRef)
{
    printf("=====a->b,a调用发出后返回=====\n");
    std::vector<YR::ObjectRef<int>> nums2;

    for (int i = 0; i < 10; i++) {
        nums2.push_back(YR::Function(Add).Invoke(1, 1));
    }
    // 当前memoryStore中没处理嵌套对象，导致嵌套对象不会被put到数据系统
    // 数据系统seal时如果传入不存在元数据的对象，不会报错
    auto ret = YR::Put(nums2);
    // SumWithObjectRef中Get嵌套对象会卡主
    auto r2 = YR::Function(SumWithObjectRef).Invoke(ret);
    EXPECT_EQ(*YR::Get(r2, -1), 20);
}

TEST_F(TaskTest, YR_MSetTx)
{
    std::vector<std::string> keys, vals;
    int totalNum = 8;
    for (int i = 0; i < totalNum; i++) {
        std::string key = "Key" + std::to_string(i);
        std::string value = "Value" + std::to_string(i);
        keys.emplace_back(key);
        vals.emplace_back(value);
    }
    YR::KV().MSetTx(keys, vals, YR::ExistenceOpt::NX);
    std::vector<std::string> actualVals = YR::KV().Get(keys, 300);
    for (int i = 0; i < keys.size(); i++) {
        if (vals[i] != actualVals[i]) {
            std::cout << "failed, value unexcepted." << vals[i] << ", " << actualVals[i] << std::endl;
            EXPECT_EQ(vals[i], actualVals[i]);
        }
    }
    YR::KV().Del(keys);

    std::vector<char *> valsPtr;
    std::vector<size_t> lens;
    for (const auto &val : vals) {
        valsPtr.push_back(const_cast<char *>(val.c_str()));
        lens.push_back(val.size());
    }
    YR::KV().MSetTx(keys, valsPtr, lens, YR::ExistenceOpt::NX);
    actualVals = YR::KV().Get(keys, 300);
    for (int i = 0; i < keys.size(); i++) {
        if (vals[i] != actualVals[i]) {
            std::cout << "failed, value unexcepted." << vals[i] << ", " << actualVals[i] << std::endl;
            EXPECT_EQ(vals[i], actualVals[i]);
        }
    }
    YR::KV().Del(keys);

    YR::KV().MWriteTx(keys, vals, YR::ExistenceOpt::NX);
    std::vector<std::shared_ptr<std::string>> actualVals2 = YR::KV().Read<std::string>(keys, 300, false);
    for (int i = 0; i < keys.size(); i++) {
        if (vals[i] != *(actualVals2[i])) {
            std::cout << "failed, value unexcepted." << vals[i] << ", " << actualVals[i] << std::endl;
            EXPECT_EQ(vals[i], actualVals[i]);
        }
    }
    YR::KV().Del(keys);
    std::cout << "kv mset test done." << std::endl;
}

TEST_F(TaskTest, DeliverObjectRefCall)
{
    YR::ObjectRef<int> num = YR::Put(1);
    std::vector<YR::ObjectRef<int>> nums;
    nums.emplace_back(num);
    auto ret = YR::Function(RemoteAdd).Invoke(nums);
    EXPECT_EQ(*YR::Get(ret, -1), 1);
}

/*case
 * @title: 在函数执行期间，kill bus，以测试 NotifyAllDisconnected 回调是否能正常工作
 * @precondition: This test case should manually modify deploy.sh
 * @step: kill bus
 * @expect:  kill bus 后 DISCONNECT_TIMEOUT_MS 后，Get 抛出 3006 异常。
 */
TEST_F(TaskTest, AfterSleepKillBusTest)
{
    auto obj = YR::Function(AfterSleepSec).Invoke(1);  // should change to 20
    std::cout << "you should manually kill bus proxy.\n";
    try {
        int ret = *YR::Get(obj, 930);
        std::cout << "ret is " << ret << std::endl;
    } catch (YR::Exception &e) {
        std::cout << e.what() << std::endl;
    }

    this->TearDown();
    this->SetUp();

    auto obj2 = YR::Function(AfterSleepSec).Invoke(1);  // should change to 22
    try {
        int ret = *YR::Get(obj2, 930);
        std::cout << "ret is " << ret << std::endl;
    } catch (YR::Exception &e) {
        std::cout << e.what() << std::endl;
    }
}

bool Retry(const YR::Exception &e) noexcept
{
    if (e.Code() == 2002) {
        std::string msg = e.what();
        if (msg.find("failed for") != std::string::npos) {
            return true;
        }
    }
    return false;
}

bool RetryForNothing(const YR::Exception &e) noexcept
{
    if (e.Code() == 2002) {
        std::string msg = e.what();
        if (msg.find("nothing") != std::string::npos) {
            return true;
        }
    }
    return false;
}

TEST_F(TaskTest, RetryChecker)
{
    std::string key = "counter";
    int n = 3;
    YR::InvokeOptions opt;

    // retry success
    opt.retryTimes = n;
    opt.retryChecker = Retry;
    auto obj = YR::Function(FailedForNTimesAndThenSuccess).Options(opt).Invoke(n);
    EXPECT_EQ(*YR::Get(obj), 0);
    YR::KV().Del(key);

    // retry success with no retry checker
    opt.retryTimes = n;
    opt.retryChecker = nullptr;
    obj = YR::Function(FailedForNTimesAndThenSuccess).Options(opt).Invoke(n);
    EXPECT_EQ(*YR::Get(obj), 0);
    YR::KV().Del(key);

    // too few retry time
    opt.retryTimes = n - 1;
    obj = YR::Function(FailedForNTimesAndThenSuccess).Options(opt).Invoke(n);
    EXPECT_THROW_WITH_CODE_AND_MSG(YR::Get(obj), 2002, "failed for " + std::to_string(n) + " times");
    YR::KV().Del(key);

    // wrong retry checker, failed for the 1st time
    opt.retryTimes = n;
    opt.retryChecker = RetryForNothing;
    obj = YR::Function(FailedForNTimesAndThenSuccess).Options(opt).Invoke(n);
    EXPECT_THROW_WITH_CODE_AND_MSG(YR::Get(obj), 2002, "failed for 1 times");
    YR::KV().Del(key);
}

/*case
 * @title: 设置重试次数为1，调用会抛出异常的函数
 * @precondition:
 * @step:  1. 设置重试次数为1
 * @step:  2. 构造大参数vector
 * @step:  3. 调用会抛出异常的函数，vector作为入参
 * @step:  4. 调用YR::Get
 * @expect:  1.应该抛出execBigArgsError的异常
 * @expect:  2.不应该抛出Get timeout 300000ms from datasystem的异常（参数被减计数为0导致）
 */
TEST_F(TaskTest, Test_After_Retry_Args_Should_Not_DecreaseRef)
{
    printf("=====云下invoke 大参数 用户函数构造异常=====\n");
    std::vector<char> v(512 * 1024, 'a');  // 512K
    YR::InvokeOptions option;
    option.retryTimes = 1;
    auto r1 = YR::Function(ExecBigArgsAndFailed).Options(option).Invoke(v);
    try {
        int n1 = *YR::Get(r1);
    } catch (YR::Exception &e) {
        printf("error: %s\n", e.what());
        std::string errorCode = "ErrCode: 2002";
        std::string errorMsg = execBigArgsError;
        std::string excepMsg = e.what();
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
}

/*case
 * @title: cpp task函数调用成功
 * @precondition:
 * @step:  调用openYuanRong task 函数 AddOne
 * @expect:  1.预期无异常抛出，
 * @expect:  2.返回值为2
 */
TEST_F(TaskTest, InvokeCppFuncSuccessfully)
{
    auto ret = YR::CppFunction<int>("AddOne")
                   .SetUrn("sn:cn:yrk:default:function:0-yr-stcpp:$latest")
                   .Invoke(1);
    EXPECT_EQ(*YR::Get(ret), 2);
}

/*case
 * @title: cpp task函数调用失败
 * @precondition:
 * @step:  调用openYuanRong task 函数 AddOne
 * @expect:  1.预期有异常抛出
 */
TEST_F(TaskTest, InvokeCppFuncFailed)
{
    try {
        YR::CppFunction<int>("AddOne").SetUrn("abc123").Invoke(1);
    } catch (YR::Exception &e) {
        printf("error: %s\n", e.what());
        std::string errorCode = "ErrCode: 1001";
        std::string errorMsg = "Failed to split functionUrn: split num 1 is expected to be 7";
        std::string excepMsg = e.what();
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
    try {
        auto ret = YR::CppFunction<std::string>("AddOne")
                       .SetUrn("sn:cn:yrk:default:function:0-yr-stcpp:$latest")
                       .Invoke(1);
        YR::Get(ret);
    } catch (YR::Exception &e) {
        printf("error: %s\n", e.what());
        std::string errorCode = "ErrCode: 4003";
        std::string errorMsg = "std::bad_cast";
        std::string excepMsg = e.what();
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
    try {
        auto ret = YR::CppFunction<int>("AddTen")
                       .SetUrn("sn:cn:yrk:default:function:0-yr-stcpp:$latest")
                       .Invoke(1);
        YR::Get(ret);
    } catch (YR::Exception &e) {
        printf("error: %s\n", e.what());
        std::string errorCode = "ErrCode: 2002";
        std::string errorMsg = "AddTen is not found in FunctionHelper";
        std::string excepMsg = e.what();
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
    try {
        auto ret = YR::CppFunction<int>("AddOne")
                       .SetUrn("sn:cn:yrk:default:function:0-yr-stcpp:$latest")
                       .Invoke(std::string("one"));
        YR::Get(ret);
    } catch (YR::Exception &e) {
        printf("error: %s\n", e.what());
        std::string errorCode = "ErrCode: 4003";
        std::string errorMsg = "std::bad_cast";
        std::string excepMsg = e.what();
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
    ASSERT_EQ(1, 1);
}

/*case
 * @title: python task函数调用成功
 * @precondition:
 * @step:  调用openYuanRong task 函数 returnInt
 * @expect:  1.预期无异常抛出，
 * @expect:  2.返回值为1
 */
TEST_F(TaskTest, InvokePythonFuncSuccessfully)
{
    auto ret = YR::PyFunction<int>("common", "add_one")
                   .SetUrn("sn:cn:yrk:default:function:0-yr-stpython:$latest")
                   .Invoke(10);  // moduleName, functionName
    EXPECT_EQ(*YR::Get(ret), 11);
}

/*case
 * @title: python task函数调用成功
 * @precondition:
 * @step:  调用openYuanRong task 函数 returnInt
 * @expect:  1.预期无异常抛出，
 * @expect:  2.返回值为1
 */
TEST_F(TaskTest, InvokePythonFuncWithRefSuccessfully)
{
    auto obj = YR::Put(10);
    auto ret = YR::PyFunction<int>("common", "add_one")
                   .SetUrn("sn:cn:yrk:default:function:0-yr-stpython:$latest")
                   .Invoke(obj);  // moduleName, functionName
    EXPECT_EQ(*YR::Get(ret), 11);
}

/*case
 * @title: java task函数调用成功
 * @precondition:
 * @step:  调用openYuanRong task 函数 returnInt
 * @expect:  1.预期无异常抛出，
 * @expect:  2.返回值为1
 */
TEST_F(TaskTest, DISABLED_InvokeJavaFuncSuccessfully)
{
    auto ret = YR::JavaFunction<int>("org.yuanrong.testutils.TestUtils", "returnInt")
                   .SetUrn("sn:cn:yrk:default:function:0-yr-stjava:$latest")
                   .Invoke(1);
    EXPECT_EQ(*YR::Get(ret), 1);
}

/*case
 * @title: java task函数调用失败
 * @precondition:
 * @step:  调用openYuanRong task 函数 returnInt
 * @expect:  1.预期有异常抛出，
 */
TEST_F(TaskTest, DISABLED_InvokeJavaFuncFailed)
{
    try {
        YR::JavaFunction<int>("org.yuanrong.testutils.TestUtils", "returnInt").SetUrn("abc123").Invoke(1);
    } catch (YR::Exception &e) {
        printf("error: %s\n", e.what());
        std::string errorCode = "ErrCode: 1001";
        std::string errorMsg = "Failed to split functionUrn: split num 1 is expected to be 7";
        std::string excepMsg = e.what();
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
    try {
        auto ret = YR::JavaFunction<std::string>("org.yuanrong.testutils.TestUtils", "returnInt")
                       .SetUrn("sn:cn:yrk:default:function:0-yr-stjava:$latest")
                       .Invoke(1);
        YR::Get(ret);
    } catch (YR::Exception &e) {
        printf("error: %s\n", e.what());
        std::string errorCode = "ErrCode: 4003";
        std::string errorMsg = "std::bad_cast";
        std::string excepMsg = e.what();
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
    try {
        auto ret = YR::JavaFunction<int>("TestUtils", "returnInt")
                       .SetUrn("sn:cn:yrk:default:function:0-yr-stjava:$latest")
                       .Invoke(1);
        YR::Get(ret);
    } catch (YR::Exception &e) {
        printf("error: %s\n", e.what());
        std::string errorCode = "ErrCode: 3003";
        std::string errorMsg = "ClassNotFoundException";
        std::string excepMsg = e.what();
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
    try {
        auto ret = YR::JavaFunction<int>("org.yuanrong.testutils.TestUtils", "addOne")
                       .SetUrn("sn:cn:yrk:default:function:0-yr-stjava:$latest")
                       .Invoke(1);
        YR::Get(ret);
    } catch (YR::Exception &e) {
        printf("error: %s\n", e.what());
        std::string errorCode = "ErrCode: 3003";
        std::string errorMsg = "IllegalArgumentException";
        std::string excepMsg = e.what();
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
}

TEST_F(TaskTest, FunctionNotRegisteredTest)
{
    try {
        auto ret = YR::Function(&FunctionNotRegistered).Invoke();
        YR::Wait(ret);
        EXPECT_EQ(0, 1);
    } catch (const YR::Exception &e) {
        const std::string msg = e.what();
        std::cerr << msg << std::endl;
        EXPECT_TRUE(msg.find(YR::FUNCTION_NOT_REGISTERED_ERROR_MSG) != std::string::npos);
    }
}

TEST_F(TaskTest, CloudFunctionNotRegisteredTest)
{
    try {
        auto ret = YR::Function(&FunctionRegistered).Invoke();
        YR::Wait(ret);
        EXPECT_EQ(0, 1);
    } catch (const YR::Exception &e) {
        const std::string msg = e.what();
        std::cerr << msg << std::endl;
        EXPECT_TRUE(msg.find(YR::FUNCTION_NOT_REGISTERED_ERROR_MSG) != std::string::npos);
    }
}

/*case
 * @title: 创建task函数，objID携带workerID
 * @precondition:
 * @step:  调用openYuanRong task 函数 AddOne
 * @expect:  1.预期无异常抛出，
 * @expect:  2.objID携带ds workerID
 * @expect:  3.返回值为2
 */
TEST_F(TaskTest, CheckTaskObjIdSuccessfully)
{
    auto ret = YR::Function(&AddOne).Invoke(1);
    EXPECT_EQ(ret.ID().size(), 20);
    EXPECT_EQ(*YR::Get(ret), 2);
    auto obj = YR::Put(3);
    auto ret1 = YR::Function(&AddOne).Invoke(obj);
    EXPECT_EQ(ret1.ID().size(), 20);
    EXPECT_EQ(*YR::Get(ret1), 4);
}

/*case
 * @title: put, 生成objID携带workerID
 * @precondition:
 * @step:  1.调用函数，put，get
 * @expect:  1.put生成携带workerID的objID
 */
TEST_F(TaskTest, CheckPutObjIdSuccessfully)
{
    auto r1 = YR::Function(Add).Invoke(1, 1);
    auto ret = YR::Put(r1);
    EXPECT_EQ(ret.ID().size(), 57);
    EXPECT_EQ(*YR::Get(*YR::Get(ret, -1), -1), 2);
}

/*
 * Check whether customextension has been written into the request body by viewing log file.
 * Therefore this case is disabled.
 */
TEST_F(TaskTest, DISABLED_InvokeFunctionWithCustomextensionTest)
{
    YR::InvokeOptions opt;
    opt.customExtensions = {
        {"endpoint", "InvokeFunction1"}, {"app_name", "InvokeFunction2"}, {"tenant_id", "InvokeFunction3"}};
    auto ret = YR::Function(&AddTwo).Options(opt).Invoke(1);
    EXPECT_EQ(*YR::Get(ret), 3);
}

/*
 * Check whether preferredAntiOtherLabels has been written into the request body by viewing log file.
 * Therefore this case is disabled.
 */
TEST_F(TaskTest, DISABLED_AntiOtherLabelsSuccess)
{
    YR::InvokeOptions opt;
    auto aff1 = YR::ResourcePreferredAffinity(YR::LabelExistsOperator("label_1"));
    opt.AddAffinity(aff1);
    opt.preferredAntiOtherLabels = true;
    auto ret = YR::Function(&AddTwo).Options(opt).Invoke(1);
    EXPECT_EQ(*YR::Get(ret), 3);
}

/*case
 * @title: 测试KVSet
 * @precondition:
 * @step:  1.调用KVSet
 * @step:  2.调用KVGet
 * @expect:  1.get到正确的数据
 */
TEST_F(TaskTest, KVSetAndGetSuccessfully)
{
    std::string key = "kv-key";
    std::string value = "kv-value";
    YR::SetParam param;
    param.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
    YR::KV().Set(key, value.c_str(), param);
    std::string result = YR::KV().Get(key);
    EXPECT_EQ(result, value);
    YR::KV().Del(key);

    YR::SetParamV2 paramV2;
    paramV2.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
    // Check whether shared disk is enabled.
    YR::KV().Set(key, value.c_str(), paramV2);
    auto resultV2 = YR::KV().Get(key);
    EXPECT_EQ(resultV2, value);
    YR::KV().Del(key);
}

/*case
 * @title: 测试MSetTx with param
 * @precondition:
 * @step:  1.调用MSetTx
 * @step:  2.调用KVGet
 * @expect:  1.get到正确的数据
 */
TEST_F(TaskTest, KVMSetTxWithParamSuccessfully)
{
    std::vector<std::string> keys;
    std::vector<std::string> values;
    std::string key = "kv-key";
    std::string value = "kv-value";
    for (int i = 0; i < 6; i++) {
        std::string key1 = key + std::to_string(i);
        keys.emplace_back(key1);
        std::string value1 = value + std::to_string(i);
        values.emplace_back(value1);
    }
    YR::MSetParam param;
    param.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
    param.ttlSecond = 10;
    // Check whether shared disk is enabled.
    YR::KV().MSetTx(keys, values, param);
    for (int i = 0; i < 6; i++) {
        std::string key1 = key + std::to_string(i);
        std::string result = YR::KV().Get(key1);
        EXPECT_EQ(result, value + std::to_string(i));
    }
    YR::KV().Del(keys);

    std::vector<char *> valsPtr;
    std::vector<size_t> lens;
    for (const auto &val : values) {
        valsPtr.push_back(const_cast<char *>(val.c_str()));
        lens.push_back(val.size());
    }
    YR::KV().MSetTx(keys, valsPtr, lens, param);
    auto actualVals = YR::KV().Get(keys, 300);
    for (int i = 0; i < 6; i++) {
        if (values[i] != actualVals[i]) {
            std::cout << "failed, value unexcepted." << values[i] << ", " << actualVals[i] << std::endl;
            EXPECT_EQ(values[i], actualVals[i]);
        }
    }
    YR::KV().Del(keys);
}

/*case
 * @title: 调用Set或者MSetTx失败
 * @precondition:
 * @step:  Set或者MSetTx
 * @step:  传入非法参数
 * @expect:  抛出异常
 */
TEST_F(TaskTest, TaskSetOrMSetTxFailed)
{
    std::vector<std::string> keys;
    std::vector<std::string> values;
    try {
        YR::MSetParam param;
        param.cacheType = YR::CacheType::DISK;
        YR::KV().MSetTx(keys, values, param);
    } catch (YR::Exception &e) {
        std::string errorCode = "ErrCode: 1001";
        std::string errorMsg = "The keys should not be empty";
        std::string excepMsg = e.what();
        std::cout << "exception: " << excepMsg << std::endl;
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
    try {
        YR::MSetParam param;
        keys.emplace_back("key1");
        YR::KV().MSetTx(keys, values, param);
    } catch (YR::Exception &e) {
        std::string errorCode = "ErrCode: 1001";
        std::string errorMsg = "input vector size not equal";
        std::string excepMsg = e.what();
        std::cout << "exception: " << excepMsg << std::endl;
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
    try {
        values.emplace_back("value1");
        YR::MSetParam param;
        param.existence = YR::ExistenceOpt::NONE;
        YR::KV().MSetTx(keys, values, param);
    } catch (YR::Exception &e) {
        std::string errorCode = "ErrCode: 1001";
        std::string errorMsg = "ExistenceOpt should be NX";
        std::string excepMsg = e.what();
        std::cout << "exception: " << excepMsg << std::endl;
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
}

/*case
 * @title: 测试Put
 * @precondition:
 * @step:  1.调用Put
 * @step:  2.调用Get
 * @expect:  1.get到正确的数据
 */
TEST_F(TaskTest, PutAndGetSuccessfully)
{
    YR::CreateParam param;
    param.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
    param.consistencyType = YR::ConsistencyType::PRAM;
    // Check whether shared disk is enabled.
    std::string res = "success";
    auto resRef = YR::Put(res, param);
    std::string value = *YR::Get(resRef);
    EXPECT_EQ(res, value);
}

/*case
 * @title: 测试实例数量约束为1，调用不同options无状态函数请求，请求结果不卡住
 * @expect:  1.get到正确的数据
 */
TEST_F(TaskTest, TestDifferentResourceTask)
{
    YR::Finalize();
    YR::Config config;
    config.mode = YR::Config::Mode::CLUSTER_MODE;
    config.maxTaskInstanceNum = 1;
    auto info = YR::Init(config);
    YR::InvokeOptions opt;
    opt.cpu = 600;
    auto r1 = YR::Function(&AddTwo).Options(opt).Invoke(1);
    auto r2 = YR::Function(&AddTwo).Invoke(1);
    EXPECT_EQ(*YR::Get(r1), 3);
    EXPECT_EQ(*YR::Get(r2), 3);
}

/*case
 * @title: 中途手动kill proxy进程，invoke请求不卡住
 * @expect:  1.get到正确的数据
 */
TEST_F(TaskTest, DISABLED_TestGrpcClientReconnect)
{
    auto r1 = YR::Function(&AddAfterSleepTen).Invoke(2);
    sleep(1);
    system("ps -ef | grep function_proxy | grep -v grep | awk {'print $2'} | xargs kill -9");
    EXPECT_EQ(*YR::Get(r1), 3);
}

TEST_F(TaskTest, TestCancel)
{
    auto r1 = YR::Function(&InvokeAndCancel_AddAfterSleepTen).Invoke(2);
    EXPECT_EQ(*YR::Get(r1), 1);
}

TEST_F(TaskTest, cpp_kv_get_oncloud_part_keys_success_APT)
{
    printf("=========云上kv.get多个key部分成功,传入allowPartial参数true===========\n");
    bool ap = true;
    auto r1 = YR::Function(KVGetPartKeysSuccess).Invoke(ap);
    printf("result is %d\n", *YR::Get(r1, 30));
    EXPECT_EQ(*YR::Get(r1), 1) << "YR put Get failed，expect result : 1";
}

/*case
 * @title: 测试Config中新增的customEnvs参数是否生效
 * @precondition:
 * @step:  1.Config中设置customEnvs
 * @step:  2.发起函数Invoke
 * @expect:  1.customEnvs参数生效
 */
TEST_F(TaskTest, TestCustomEnvsConfig)
{
    YR::Finalize();
    YR::Config config;
    std::string key = "LD_LIBRARY_PATH";
    std::string value = "${LD_LIBRARY_PATH}:${YR_FUNCTION_LIB_PATH}/depend";
    config.customEnvs[key] = value;
    config.mode = YR::Config::Mode::CLUSTER_MODE;
    YR::Init(config);
    auto ref = YR::Function(InvokeReturnCustomEnvs).Invoke(key);
    auto customEnv = *YR::Get(ref);
    std::cout << "customEnv: " << customEnv << std::endl;
    ASSERT_TRUE(customEnv.find("depend") != std::string::npos);
    ASSERT_TRUE(customEnv.find("YR_FUNCTION_LIB_PATH") == std::string::npos);
    ASSERT_TRUE(customEnv.find("LD_LIBRARY_PATH") == std::string::npos);
}

TEST_F(TaskTest, HybridClusterCallLocal)
{
    auto obj = YR::Function(CallLocal).Invoke(1);
    EXPECT_EQ(*YR::Get(obj), 2);
}

TEST_F(TaskTest, HybridLocalCallCluster)
{
    YR::InvokeOptions opt;
    opt.alwaysLocalMode = true;
    auto obj = YR::Function(CallCluster).Options(opt).Invoke(1);
    EXPECT_EQ(*YR::Get(obj), 2);
}

TEST_F(TaskTest, HybridLocalCallClusterEmptyThreadPool)
{
    YR::Finalize();
    YR::Config config;
    config.mode = YR::Config::Mode::CLUSTER_MODE;
    config.localThreadPoolSize = 0;
    YR::Init(config);
    auto obj = YR::Function(CallLocal).Invoke(1);
    EXPECT_THROW(
        {
            try {
                YR::Get(obj);
            } catch (const YR::Exception &e) {
                std::cout << "exception: " << e.what() << std::endl;
                EXPECT_THAT(e.what(), HasSubstr("cannot submit task to empty thread pool"));
                throw;
            }
        },
        YR::Exception);
}

TEST_F(TaskTest, CancelUnfinishedTask)
{
    auto r4 = YR::Function(AddAfterSleep).Invoke(2);
    try {
        YR::Cancel(r4);
        auto r1Resp = *YR::Get(r4, 20);
    } catch (YR::Exception &e) {
        printf("error: %s\n", e.what());
        std::string errorCode = "ErrCode: 3003, ModuleCode: 20";
        std::string errorMsg = "invalid get obj, the obj has been cancelled.";
        std::string excepMsg = e.what();
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
}

TEST_F(TaskTest, RepeatPutShouldNotOOM)
{
    std::vector<char> param;
    param.resize(1024 * 1024 * 1024);
    for (int i = 0; i < 10; i++) {
        auto paramRef = YR::Put(param);
        auto value = YR::Get(paramRef);
    }
    ASSERT_EQ(1, 1);
}

/*
并发get put数据
*/
TEST_F(TaskTest, DISABLED_ConcurrencyCall)
{
    int num = 5;
    std::vector<YR::ObjectRef<std::string>> objs_str;
    objs_str.reserve(num);
    YR::InvokeOptions option;
    option.customExtensions.insert({YR::CONCURRENCY_KEY, "5"});
    for (int i = 0; i < num; i++) {
        objs_str.push_back(YR::Function(PutOneData).Options(option).Invoke());
    }
    auto res = YR::Get(objs_str);
    for (int i = 0; i < num; i++) {
        ASSERT_TRUE((*res[i]).size() == 1024);
    }
}

/*case
 * @title: 测试开启spill后可一次性Get超过共享内存的数据
 * @precondition:
 * @step:  1.Invoke 20次 函数
 * @step:  2.批量Get返回值
 * @expect:  1.批量Get返回值成功
 */
TEST_F(TaskTest, test_open_spill_2G_data)
{
    // 测试该用例需要：
    // 1.将bigArgs改成100M，for循环是20次，以使总数据为2G
    // 2.执行yr_master时调低共享内存为1G  （-s 1024）
    // 3.执行yr_master.sh时开启spill并设置4G （--ds_spill_enable true --ds_spill_directory
    // ${DEPLOY_PATH}/yr_master--ds_spill_size_limit 4096）
    printf("----读写大数据,该条用例需要环境中开启spill----\n");
    const std::vector<char> bigArgs(1 * 1024 * 1024, 'a');  // 1 修改为 100
    auto bigObj = YR::Put(bigArgs);
    YR::InvokeOptions option;
    option.customExtensions.insert({"GRACEFUL_SHUTDOWN_TIME", "1"});
    option.cpu = 1000;
    option.memory = 500;
    std::vector<YR::ObjectRef<std::vector<char>>> rets;
    for (int i = 0; i < 20; i++) {
        auto r1 = YR::Function(BigBox).Options(option).Invoke(bigObj);
        rets.push_back(r1);
    }
    std::vector<std::shared_ptr<std::vector<char>>> res = YR::Get(rets, -1);
    for (int i = 0; i < 20; i++) {
        EXPECT_EQ(*(res[i]), bigArgs) << "Get param error!";
    }
}

TEST_F(TaskTest, cpp_refcount_submit_data)
{
    auto r1 = YR::Put(100);
    int n1 = *YR::Get(r1);
    YR::Finalize();

    YR::Config config;
    config.mode = YR::Config::Mode::CLUSTER_MODE;
    auto info = YR::Init(config);
    std::cout << "job id: " << info.jobId << std::endl;
    try {
        int n1 = *YR::Get(r1, 1);
    } catch (YR::Exception &e) {
        printf("error: %s\n", e.what());
        std::string error_code = "ErrCode: 4005, ModuleCode: 30";
        std::string error_msg = "Get timeout 1000ms";
        std::string excep_msg = e.what();
        ErrorMsgCheck(error_code, error_msg, excep_msg);
    }
}

TEST_F(TaskTest, TestEnvVars)
{
    YR::InvokeOptions opts;
    std::string key = "A";
    std::string value = "A_VARS";
    opts.envVars[key] = value;
    auto ref = YR::Function(ReturnCustomEnvs).Options(opts).Invoke(key);
    std::string res = *YR::Get(ref);
    ASSERT_EQ(res, value);
}

/*case
 * @title: 云上实例finalize失败
 * @precondition:
 * @step:  invoke无状态函数，云上调用finalize方法
 * @expect: 客户端抛出异常
 */
TEST_F(TaskTest, CppFinalizeFailedCloud)
{
    int res = 0;
    try {
        auto r1 = YR::Function(PlusOneFinalize).Invoke(1);
        res = *YR::Get(r1);
    } catch (YR::Exception &e) {
        printf("error: %s\n", e.what());
        std::string error_code = "ErrCode: 4006, ModuleCode: 20";
        std::string error_msg = "ErrMsg: Finalize is not allowed to use on cloud";
        std::string excep_msg = e.what();
        ErrorMsgCheck(error_code, error_msg, excep_msg);
    }
    ASSERT_FALSE(res == 1);
}