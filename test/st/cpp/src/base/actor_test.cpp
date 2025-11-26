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
#include <string>
#include "base/utils.h"
#include "gmock/gmock.h"
#include "gtest/gtest.h"
#include "user_common_func.h"
#include "yr/yr.h"
#include "utils.h"

const static size_t g_threadPoolSize = 4;
class ActorTest : public testing::Test {
public:
    ActorTest(){};
    ~ActorTest(){};
    static void SetUpTestCase(){};
    static void TearDownTestCase(){};
    void SetUp()
    {
        YR::Config config;
        config.mode = YR::Config::Mode::CLUSTER_MODE;
        config.threadPoolSize = g_threadPoolSize;
        config.logLevel = "DEBUG";
        auto info = YR::Init(config);
        std::cout << "job id: " << info.jobId << std::endl;
    };
    void TearDown()
    {
        YR::Finalize();
    };
};

/*case
 * @title: 创建actor成功，并发起调用
 * @precondition:
 * @step:  创建actor
 * @expect:  1.预期无异常抛出，
 */
TEST_F(ActorTest, CreateSuccessful)
{
    auto creator = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto ret = creator.Function(&Counter::Add).Invoke(1);
    EXPECT_EQ(*YR::Get(ret), 2);
}

/*case
 * @title: direct call test
 * @precondition:
 * @step:  创建actor，连续调用
 */
TEST_F(ActorTest, DirectCall)
{
    auto creator = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto ret = creator.Function(&Counter::Add).Invoke(1);
    EXPECT_EQ(*YR::Get(ret), 2);
    int count = 20;
    auto start = std::chrono::steady_clock::now();
    for (int i = 0; i < count; i++) {
        auto ret = creator.Function(&Counter::Add).Invoke(1);
        YR::Wait(ret);
    }
    auto end = std::chrono::steady_clock::now();
    auto duration = std::chrono::duration_cast<std::chrono::microseconds>(end - start).count();
    std::cout << "duration is " << duration / static_cast<double>(count) << " microseconds per invoke" << std::endl;
}

TEST_F(ActorTest, DISABLED_DetachedTestWith2Jobs)
{
    std::string name = "name001_1";
    YR::InvokeOptions option1;
    option1.cpu = 500.0;
    option1.memory = 500.0;
    option1.customExtensions["lifecycle"] = "detached";
    option1.preferredAntiOtherLabels = false;

    YR::NamedInstance instance = YR::Instance(Counter::FactoryCreate, name).Options(std::move(option1)).Invoke(1);
    auto res = instance.Function(&Counter::Add).Invoke(1);
    std::cout << "res is " << *YR::Get(res) << std::endl;

    YR::Finalize();
    YR::Config config;
    config.mode = YR::Config::Mode::CLUSTER_MODE;
    config.threadPoolSize = g_threadPoolSize;
    config.logLevel = "DEBUG";
    auto info = YR::Init(config);
    std::cout << "job id: " << info.jobId << std::endl;

    auto instance2 = YR::Instance(Counter::FactoryCreate, name).Options(std::move(option1)).Invoke(1);
    auto res2 = instance2.Function(&Counter::Add).Invoke(1);
    std::cout << "res2 is " << *YR::Get(res2) << std::endl;
    ASSERT_TRUE(*YR::Get(res2) == 3);
    instance2.Terminate();
}

/*case
 * @title: kill创建好的actor，调用失败
 * @precondition:
 * @step:  1. 创建多个actor
 * @step:  2. 每个actor发起100个成员函数调用
 * @step:  3. kill实例
 * @step:  4. 每个actor发起10个成员函数调用
 * @expect:  1.每个actor的100个invoke无异常抛出
 * @expect:  2.每个actor的10个invoke抛出异常
 */
TEST_F(ActorTest, InvokeFailedWhenKillRuntime)
{
    std::vector<YR::NamedInstance<Counter>> instances;

    for (int i = 0; i < 2; i++) {
        YR::InvokeOptions option;
        option.cpu = 500;
        option.memory = 500.0;
        auto creator = YR::Instance(Counter::FactoryCreate).Options(std::move(option)).Invoke(1);
        instances.push_back(creator);
    }
    std::vector<YR::ObjectRef<int>> addResults;
    for (int j = 0; j < 100; j++) {
        for (auto i : instances) {
            auto ret = i.Function(&Counter::Add).Invoke(1);
            addResults.push_back(ret);
        }
    }
    ASSERT_NO_THROW(YR::Get(addResults));
    std::vector<YR::ObjectRef<int>> addResults2;
    for (int j = 0; j < 10; j++) {
        for (auto i : instances) {
            auto ret = i.Function(&Counter::Sleep).Invoke();
            addResults2.push_back(ret);
        }
    }
    system("kill -6 $(ps -ef|grep cppruntime |grep -v grep |awk '{print $2}')");
    ASSERT_THROW(YR::Get(addResults2), YR::Exception);
}

/*case
 * @title: 手动注入kill -9 创建好的actor，调用失败
 * @precondition:
 * @step:  1. 创建actor
 * @step:  2. 每个actor发起10个成员函数调用
 * @step:  3. kill -9 实例
 * @step:  4. 每个actor发起10个成员函数调用
 * @expect:  1.actor的10个invoke抛出异常
 */
TEST_F(ActorTest, InvokeFailedWhenKillSig9Runtime)
{
    YR::InvokeOptions option;
    option.cpu = 500;
    option.memory = 500.0;
    auto creator = YR::Instance(Counter::FactoryCreate).Options(std::move(option)).Invoke(1);
    std::vector<YR::ObjectRef<int>> addResults;
    auto ret = creator.Function(&Counter::Add).Invoke(1);
    YR::Get(ret);
    for (int j = 0; j < 10; j++) {
        auto ret = creator.Function(&Counter::Sleep).Invoke();
        addResults.push_back(ret);
    }
    system("kill -9 $(ps -ef|grep cppruntime |grep -v grep |awk '{print $2}')");
    ASSERT_THROW(YR::Get(addResults), YR::Exception);
}

/*case
 * @title: 串行依赖  ins.a->ins.b->ins.c
 * @precondition:
 * @step:  1. 创建actor，调用actor成员函数Add
 * @step:  2. 以1的返回objref作为入参，调用actor成员函数Add
 * @step:  3. 以2的返回objref作为入参，调用actor成员函数Add
 * @expect:  预期无异常抛出
 */
TEST_F(ActorTest, DependentMutliRef)
{
    // 三次依赖
    auto creator = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto ret = creator.Function(&Counter::Add).Invoke(1);
    auto ret2 = creator.Function(&Counter::Add).Invoke(ret);
    auto ret3 = creator.Function(&Counter::Add).Invoke(ret2);
    ASSERT_NO_THROW(YR::Get(ret3));
}

/*case
 * @title: 串行依赖  ins.a->ins.b->ins.c ins.a返回错误
 * @precondition:
 * @step:  1. 创建actor，调用actor成员函数Raise
 * @step:  2. 以1的返回objref作为入参，调用actor成员函数Add
 * @step:  3. 以2的返回objref作为入参，调用actor成员函数Add
 * @expect:  预期无异常抛出
 */
TEST_F(ActorTest, DependentMutliRefError)
{
    // 三次依赖，1次报错
    auto creator = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto ret = creator.Function(&Counter::Raise).Invoke();
    auto ret2 = creator.Function(&Counter::Add).Invoke(ret);
    auto ret3 = creator.Function(&Counter::Add).Invoke(ret2);
    ASSERT_THROW(YR::Get(ret3), YR::Exception);
}

/*case
 * @title: actor函数依赖成员函数的输出
 * @precondition:
 * @step:  1. 创建actor，调用actor成员函数Add
 * @step:  2. 调用actor成员函数Add
 * @step:  3. 以1,2的返回objref作为入参，调用actor成员函数Add
 * @expect:  预期无异常抛出
 */
TEST_F(ActorTest, DependentTwoMemberRetRef)
{
    // 依赖两个对象
    auto creator = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto ret = creator.Function(&Counter::Add).Invoke(1);
    auto ret2 = creator.Function(&Counter::Add).Invoke(1);
    auto ret3 = creator.Function(&Counter::AddTwo).Invoke(ret, ret2);
    ASSERT_NO_THROW(YR::Get(ret3));
}

/*case
 * @title: actor函数依赖成员函数的错误输出
 * @precondition:
 * @step:  1. 创建actor，调用actor成员函数Raise
 * @step:  2. 调用actor成员函数Add
 * @step:  3. 以1,2的返回objref作为入参，调用actor成员函数Add
 * @expect:  预期异常抛出
 */
TEST_F(ActorTest, DependentTwoMemberRetRefError)
{
    // 依赖两个对象
    auto creator = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto ret = creator.Function(&Counter::Raise).Invoke();
    auto ret2 = creator.Function(&Counter::Add).Invoke(1);
    auto ret3 = creator.Function(&Counter::AddTwo).Invoke(ret, ret2);
    ASSERT_THROW(YR::Get(ret3), YR::Exception);
}

/*case
 * @title: 同一个错误返回值依赖两次
 * @precondition:
 * @step:  1. 创建actor，调用actor成员函数Raise
 * @step:  2. 以1的返回objref作为入参，调用actor成员函数Add
 * @step:  3. 以1的返回objref作为入参，调用actor成员函数Add
 * @expect:  预期异常抛出
 */
TEST_F(ActorTest, DependentSameErrorRef)
{
    // 同个对象依赖两次
    auto creator = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto ret = creator.Function(&Counter::Raise).Invoke();
    auto ret2 = creator.Function(&Counter::Add).Invoke(ret);
    auto ret3 = creator.Function(&Counter::Add).Invoke(ret);
    ASSERT_THROW(YR::Get(ret2), YR::Exception);
    ASSERT_THROW(YR::Get(ret3), YR::Exception);
}

/*case
 * @title: 同一个返回值依赖两次
 * @precondition:
 * @step:  1. 创建actor，调用actor成员函数Add
 * @step:  2. 以1的返回objref作为入参，调用actor成员函数Add
 * @step:  3. 以1的返回objref作为入参，调用actor成员函数Add
 * @expect:  预期异常抛出
 */
TEST_F(ActorTest, DependentSameRef)
{
    // 同个对象依赖两次
    auto creator = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto ret = creator.Function(&Counter::Add).Invoke(1);
    auto ret2 = creator.Function(&Counter::Add).Invoke(ret);
    auto ret3 = creator.Function(&Counter::Add).Invoke(ret);
    ASSERT_NO_THROW(YR::Get(ret2));
    ASSERT_NO_THROW(YR::Get(ret3));
}

/*case
 * @title: 创建actor成功，并发起调用引起异常的函数
 * @precondition:
 * @step:  1.创建actor，调用成员函数Add
 * @step:  2.创建actor，调用成员函数Raise
 * @expect:  1.预期无异常抛出
 * @expect:  1.预期异常抛出
 */
TEST_F(ActorTest, InvokeFailedWhenRuntimeSEGV)
{
    YR::InvokeOptions option;
    option.cpu = 500;
    option.memory = 500.0;
    auto creator = YR::Instance(Counter::FactoryCreate).Options(std::move(option)).Invoke(1);

    auto ret = creator.Function(&Counter::Add).Invoke(1);

    ASSERT_NO_THROW(YR::Get(ret));

    auto ret2 = creator.Function(&Counter::SEGV).Invoke();
    bool raise = false;
    try {
        YR::Get(ret2);
    } catch (YR::Exception &e) {
        std::cout << e.Msg() << std::endl;
        ASSERT_FALSE(e.Msg().find("SEGV") == std::string::npos);
        raise = true;
    }
    ASSERT_TRUE(raise);
}

/*case
 * @title: 创建指定gpuactor失败
 * @precondition:
 * @step:  1.指定gpu=1.0，创建actor
 * @expect:  1.预期异常抛出
 */
TEST_F(ActorTest, DISABLED_NotEnoughGpuCheck)
{
    YR::InvokeOptions option;
    try {
        option.customResources.insert({"nvidia.com/gpu", 1.0});
        auto creator = YR::Instance(Counter::FactoryCreate).Options(std::move(option)).Invoke(1);
        auto member = creator.Function(&Counter::Add).Invoke(3);
        auto res = *YR::Get(member);
    } catch (YR::Exception &e) {
        std::string errorCode = "ErrCode: 1006";
        std::string errorMsg = "invalid resource parameter, request resource is greater than each node's max resource";
        std::string excepMsg = e.what();
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
}

/*case
 * @title: 创建指定非法资源actor失败
 * @precondition:
 * @step:  1.指定cpu=299，创建actor
 * @step:  2.指定gpu=1.0，创建actor
 * @step:  3.指定cpu=16001.0，创建actor
 * @step:  4.指定mem=127.0，创建actor
 * @expect:  1.预期异常抛出
 * @expect:  2.预期异常抛出
 * @expect:  3.预期异常抛出
 * @expect:  4.预期异常抛出
 */
TEST_F(ActorTest, InvalidResource)
{
    YR::InvokeOptions option;
    option.cpu = 299.0;
    option.memory = 128.0;
    option.customResources.insert({"nvidia.com/gpu", 0.0});
    try {
        auto creator = YR::Instance(Counter::FactoryCreate).Options(std::move(option)).Invoke(1);
        auto member = creator.Function(&Counter::Add).Invoke(3);
        auto res = *YR::Get(member);
    } catch (YR::Exception &e) {
        printf("Exception:%s,\n", e.what());
        std::string errorCode = "ErrCode: 1006";
        std::string errorMsg =
            "Required CPU resource size 299 millicores is invalid. Valid value range is [0,16000] millicores";
        std::string excepMsg = e.what();
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }

    option.cpu = 16001.0;
    option.memory = 128.0;
    try {
        auto creator = YR::Instance(Counter::FactoryCreate).Options(std::move(option)).Invoke(1);
        auto member = creator.Function(&Counter::Add).Invoke(3);
        auto res = *YR::Get(member);
    } catch (YR::Exception &e) {
        printf("Exception:%s,\n", e.what());
        std::string errorCode = "ErrCode: 1006";
        std::string errorMsg =
            "Required CPU resource size 16001 millicores is invalid. Valid value range is [0,16000] millicores";
        std::string excepMsg = e.what();
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }

    option.cpu = 300.0;
    option.memory = 127.0;
    try {
        auto creator = YR::Instance(Counter::FactoryCreate).Options(std::move(option)).Invoke(1);
        auto member = creator.Function(&Counter::Add).Invoke(3);
        auto res = *YR::Get(member);
    } catch (YR::Exception &e) {
        printf("Exception:%s,\n", e.what());
        std::string errorCode = "ErrCode: 1006";
        std::string errorMsg = "Required memory resource size 127 MB is invalid. Valid value range is [128,1073741824] MB";
        std::string excepMsg = e.what();
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
}

/*case
 * @title: 指定gpu=0创建actor成功
 * @precondition:
 * @step:  2.指定gpu=0.0，创建actor
 * @expect:  1.预期无异常抛出
 */
TEST_F(ActorTest, ZeroGPU)
{
    YR::InvokeOptions option;
    option.cpu = 333.0;
    option.memory = 222.0;
    option.customResources.insert({"nvidia.com/gpu", 0.0});
    auto creator = YR::Instance(Counter::FactoryCreate).Options(std::move(option)).Invoke(1);
    auto member = creator.Function(&Counter::Add).Invoke(3);
    auto res = *YR::Get(member);
    printf("instance result is %d\n", *YR::Get(member));
    EXPECT_EQ(res, 4);
    creator.Terminate();
}

/*case
 * @title: actor函数，类成员函数互相调用
 * @precondition:
 * @step:  1.创建actor A，A在init中创建B
 * @step:  2.A的成员函数会调用B的函数进行各自状态累加操作
 * @step:  3.获取A的状态
 * @step:  4.通过A获取B状态
 * @step:  5.通过A删除B
 * @expect:  1.状态值正确，无异常抛出
 */
TEST_F(ActorTest, ActorCordination)
{
    // 初始化两个class，worker正常拉起，A的成员函数会调用B的函数进行各自状态累加操作，
    // 然后A的get函数获取对象状态结果值累加正确，terminate状态值后，实例可以正常缩容
    printf("=========初始化两个class，worker正常拉起，A的成员函数会调用B的函数=========\n");
    auto instance = YR::Instance(CounterA::FactoryCreate).Invoke(1);
    printf("=========调用A的add成员函数，状态值正常累加=========\n");
    auto r1 = instance.Function(&CounterA::Add).Invoke(1);
    auto v1 = *YR::Get(r1);
    printf("result is %d\n", v1);
    EXPECT_LE(v1, 4);
    printf("=========调用A的GetCountB成员函数，获取ClassB的状态值=========\n");
    r1 = instance.Function(&CounterA::GetCountB).Invoke();
    v1 = *YR::Get(r1);
    printf("GetCountB is %d\n", *YR::Get(r1));
    EXPECT_LE(v1, 2);
    printf("=========调用A的GetCountA成员函数，获取状态值=========\n");
    r1 = instance.Function(&CounterA::GetCountA).Invoke();
    v1 = *YR::Get(r1);
    printf("GetCountA is %d\n", *YR::Get(r1));
    EXPECT_LE(v1, 2);
    printf("=========调用A的TerminateB成员函数，清除B的状态值=========\n");
    r1 = instance.Function(&CounterA::TerminateB).Invoke(1);
    v1 = *YR::Get(r1);
    printf("TerminateB is %d\n", *YR::Get(r1));
    EXPECT_LE(v1, 1);
}

/*case
 * @title: actor函数，类成员函数互相调用
 * @precondition:
 * @step:  1.初始化classA，A初始化B，B初始化C
 * @step:  2.使用invoke调用A的成员函数，函数按A调用B，B调用C链式invoke
 * @step:  3.调用A，B，C的查询方法获取状态结果
 * @step:  4.执行terminate方法
 * @expect:  1.状态值正确，无异常抛出
 */
TEST_F(ActorTest, ActorsCordination)
{
    printf("=========函数按A调用B，B调用C链式invoke=========\n");
    auto instance = YR::Instance(CounterC::FactoryCreate).Invoke(1);
    auto r1 = instance.Function(&CounterC::Add).Invoke(1);
    auto v1 = *YR::Get(r1);
    printf("result is %d\n", *YR::Get(r1));
    r1 = instance.Function(&CounterC::GetCountB).Invoke();
    v1 = *YR::Get(r1);
    printf("GetCountB is %d\n", *YR::Get(r1));
    r1 = instance.Function(&CounterC::GetCountA).Invoke();
    v1 = *YR::Get(r1);
    printf("GetCountA is %d\n", *YR::Get(r1));
    EXPECT_LE(v1, 2);
    r1 = instance.Function(&CounterC::GetCountC).Invoke();
    v1 = *YR::Get(r1);
    printf("GetCountA is %d\n", *YR::Get(r1));
    EXPECT_LE(v1, 2);
}

/*case
 * @title: 删除actor实例后调用实例
 * @precondition:
 * @step:  1.创建实例
 * @step:  2.正常调用
 * @step:  3.执行terminate方法
 * @step:  4.继续
 * @expect:  1.异常抛出
 */
TEST_F(ActorTest, NotExistInstanceMsgCheck)
{
    auto instance = YR::Instance(CounterC::FactoryCreate).Invoke(1);
    auto r1 = instance.Function(&CounterC::Add).Invoke(1);
    auto v1 = *YR::Get(r1);
    instance.Terminate();
    sleep(2);
    try {
        auto r2 = instance.Function(&CounterC::Add).Invoke(1);
        auto v2 = *YR::Get(r2);
    } catch (YR::Exception &e) {
        printf("error: %s\n", e.what());
        std::string errorCodeExit = "ErrCode: 1007";
        std::string errorCodeNotFound = "ErrCode: 1003";
        std::string excepMsg = e.what();
        std::string::size_type codeExitIdx = excepMsg.find(errorCodeExit);
        std::string::size_type codeNotFoundIdx = excepMsg.find(errorCodeNotFound);
        ASSERT_TRUE(codeExitIdx != std::string::npos || codeNotFoundIdx != std::string::npos);
    }
}

/*case
 * @title: 创建actor实例后，调用实例成员方法抛出SIGILL
 * @precondition:
 * @step:  1.创建实例
 * @step:  2.调用get_sigill
 * @expect:  1.异常抛出
 */
// The error information is inconsistent.
TEST_F(ActorTest, ExceptionIllegalInstruction)
{
    printf("=====云上invoke 检测非法指令=====\n");
    auto instance = YR::Instance(Actor::FactoryCreate).Invoke(100);
    auto ins = instance.Function(&Actor::get_sigill).Invoke();
    try {
        int n1 = *YR::Get(ins);
    } catch (YR::Exception &e) {
        printf("error: %s\n", e.what());
        std::string errorCode = "ErrCode: 2002";
        std::string errorMsg = "SIGILL";
        std::string excepMsg = e.what();
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
}

/*case
 * @title: 创建actor实例后，调用实例成员方法抛出SIGINT
 * @precondition:
 * @step:  1.创建实例
 * @step:  2.调用get_sigill
 * @expect:  1.异常抛出
 */
TEST_F(ActorTest, ExceptionInterruptSignal)
{
    printf("=====云上invoke 程序终止信号=====\n");
    auto instance = YR::Instance(Actor::FactoryCreate).Invoke(100);
    auto ins = instance.Function(&Actor::get_sigint).Invoke();
    try {
        int n1 = *YR::Get(ins);
    } catch (YR::Exception &e) {
        printf("error: %s\n", e.what());
        std::string errorCode = "ErrCode: 2002";
        std::string errorMsg = "";
        std::string excepMsg = e.what();
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
}

/*case
 * @title: 创建actor,并因为资源不足失败，Wait不卡住
 * @precondition:
 * @step:  1.指定cpu=2000，创建500个actor
 * @step:  2.调用一次invoke
 * @step:  3.调用wait等待结果
 * @expect:  抛出资源不足异常，不卡住
 */
TEST_F(ActorTest, DISABLED_ResourceNotEnough)
{
    std::vector<YR::NamedInstance<Counter>> ins;
    std::vector<YR::ObjectRef<int>> rets;
    for (int i = 0; i < 500; i++) {
        YR::InvokeOptions option;
        option.cpu = 2000;
        auto creator = YR::Instance(Counter::FactoryCreate).Options(std::move(option)).Invoke(1);
        ins.push_back(creator);
        auto member = creator.Function(&Counter::Add).Invoke(3);
        rets.push_back(member);
    }
    try {
        auto ret = YR::Wait(rets, rets.size(), -1);
        std::cout << ret.first.size() << std::endl;
        for (auto i : ret.first) {
            std::cout << i.ID() << std::endl;
        }
        std::cout << ret.second.size() << std::endl;
        for (auto i : ret.second) {
            std::cout << i.ID() << std::endl;
        }
        ASSERT_TRUE(false);  // this case should not goto this code line
    } catch (YR::Exception &e) {
        printf("error: %s\n", e.what());
        std::string errorCode = "ErrCode: 1002";
        std::string errorMsg = "resources not enough";
        std::string excepMsg = e.what();
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
}

/*case
 * @title: 配置子目录，创建actor成功，并发起调用
 * @precondition:
 * @step:  创建actor，获取子目录环境变量，返回子目录
 * @expect:  返回用户定义子目录
 */
TEST_F(ActorTest, SubDir)
{
    YR::InvokeOptions opts;
    opts.customExtensions["DELEGATE_DIRECTORY_INFO"] = "/tmp";
    auto creator = YR::Instance(Counter::FactoryCreate).Options(std::move(opts)).Invoke(1);
    auto r1 = creator.Function(&Counter::GetDir).Invoke();
    auto v1 = *YR::Get(r1);
    std::string::size_type idx = v1.find("/tmp");
    ASSERT_TRUE(idx != std::string::npos);
    creator.Terminate();
}

/*case
 * @title: 创建actor
 * @precondition:
 * @step:  1.设置Concurrency并发度为4
 * @step:  2.调用4次Invoke，执行ParallelFor
 * @step:  3.调用wait等待结果
 * @step:  4.调用GetCtxIdsSize获取runtime启动的线程数量
 * @expect:  不卡住
 * @expect:  GetCtxIdsSize() == threadPoolSize
 */
TEST_F(ActorTest, TestConcurrencyParallelFor)
{
    YR::InvokeOptions option;
    option.customExtensions[YR::CONCURRENCY_KEY] = "4";
    auto counter = YR::Instance(CounterB::FactoryCreate).Options(std::move(option)).Invoke(1);
    std::vector<YR::ObjectRef<int>> rets;
    for (int i = 0; i < 4; i++) {
        rets.emplace_back(counter.Function(&CounterB::ParallelFor).Invoke());
    }
    // test wait not stuck
    YR::Wait(rets, rets.size(), -1);
    counter.Function(&CounterB::ParallelFor).Invoke();
    auto ret = *(YR::Get(counter.Function(&CounterB::GetCtxIdsSize).Invoke()));
    counter.Terminate();
    EXPECT_GE(ret, 1);
}

/*case
 * @title: 创建actor成功，并发起调用
 * @precondition:
 * @step:  创建actor
 * @expect:  1.预期无异常抛出，
 * @expect:  2.返回值为2
 */
TEST_F(ActorTest, CreateCppActorSuccessful)
{
    auto cppCls = YR::CppInstanceClass::FactoryCreate("Counter::FactoryCreate");
    auto creator =
        YR::Instance(cppCls).SetUrn("sn:cn:yrk:12345678901234561234567890123456:function:0-yr-stcpp:$latest").Invoke(1);
    auto ret = creator.CppFunction<int>("&Counter::Add").Invoke(1);
    EXPECT_EQ(*YR::Get(ret), 2);
}

/*case
 * @title: 创建actor成功，并发起调用
 * @precondition:
 * @step:  创建actor
 * @expect:  1.预期有异常抛出
 */
TEST_F(ActorTest, CreateCppActorFailed)
{
    try {
        auto cppCls = YR::CppInstanceClass::FactoryCreate("Counter::FactoryCreate");
        auto creator = YR::Instance(cppCls).SetUrn("abc123").Invoke(1);
    } catch (YR::Exception &e) {
        printf("error: %s\n", e.what());
        std::string errorCode = "ErrCode: 1001";
        std::string errorMsg = "Failed to split functionUrn: split num 1 is expected to be 7";
        std::string excepMsg = e.what();
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
    try {
        auto cppCls = YR::CppInstanceClass::FactoryCreate("Counter");
        auto creator = YR::Instance(cppCls)
                           .SetUrn("sn:cn:yrk:12345678901234561234567890123456:function:0-yr-stcpp:$latest")
                           .Invoke(1);
        auto ret = creator.CppFunction<int>("&Counter::Add").Invoke(1);
        YR::Get(ret);
    } catch (YR::Exception &e) {
        printf("error: %s\n", e.what());
        std::string errorCode = "ErrCode: 2002";
        std::string errorMsg = "Counter is not found in FunctionHelper";
        std::string excepMsg = e.what();
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
    try {
        auto cppCls = YR::CppInstanceClass::FactoryCreate("Counter::FactoryCreate");
        auto creator = YR::Instance(cppCls)
                           .SetUrn("sn:cn:yrk:12345678901234561234567890123456:function:0-yr-stcpp:$latest")
                           .Invoke(std::string("one"));
        auto ret = creator.CppFunction<int>("&Counter::Add").Invoke(1);
        YR::Get(ret);
    } catch (YR::Exception &e) {
        printf("error: %s\n", e.what());
        std::string errorCode = "ErrCode: 2002";
        std::string errorMsg = "std::bad_cast";
        std::string excepMsg = e.what();
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
}

/*case
 * @title: 创建actor成功，并发起调用
 * @precondition:
 * @step:  创建actor
 * @expect:  1.预期无异常抛出，
 * @expect:  2.返回值为1
 */
TEST_F(ActorTest, CreatePythonActorSuccessful)
{
    auto pyCls = YR::PyInstanceClass::FactoryCreate("common", "SimpleInstance");
    auto pyIns = YR::Instance(pyCls)
                     .SetUrn("sn:cn:yrk:12345678901234561234567890123456:function:0-yr-stpython:$latest")
                     .Invoke();
    auto ret = pyIns.PyFunction<int>("add_one").Invoke(1);
    EXPECT_EQ(*YR::Get(ret), 2);
}

/*case
 * @title: 创建actor成功，并发起调用
 * @precondition:
 * @step:  创建actor
 * @expect:  1.预期无异常抛出，
 * @expect:  2.返回值为1
 */
TEST_F(ActorTest, CreatePythonWithRefActorSuccessful)
{
    auto pyCls = YR::PyInstanceClass::FactoryCreate("common", "SimpleInstance");
    auto pyIns = YR::Instance(pyCls)
                     .SetUrn("sn:cn:yrk:12345678901234561234567890123456:function:0-yr-stpython:$latest")
                     .Invoke();
    auto obj = YR::Put(1);
    auto ret = pyIns.PyFunction<int>("add_one").Invoke(obj);
    EXPECT_EQ(*YR::Get(ret), 2);
}

/*case
 * @title: 创建actor成功，并发起调用
 * @precondition:
 * @step:  创建actor
 * @expect:  1.预期无异常抛出，
 * @expect:  2.返回值为1
 */
TEST_F(ActorTest, CreateJavaActorSuccessful)
{
    auto javaCls = YR::JavaInstanceClass::FactoryCreate("com.yuanrong.testutils.TestUtils");
    auto creator = YR::Instance(javaCls)
                       .SetUrn("sn:cn:yrk:12345678901234561234567890123456:function:0-yr-stjava:$latest")
                       .Invoke();
    auto ret = creator.JavaFunction<int>("returnInt").Invoke(1);
    EXPECT_EQ(*YR::Get(ret), 1);
}

/*case
 * @title: 创建actor成功，并发起调用
 * @precondition:
 * @step:  创建actor
 * @expect:  1.预期有异常抛出
 */
TEST_F(ActorTest, CreateJavaActorFailed)
{
    try {
        auto javaCls = YR::JavaInstanceClass::FactoryCreate("com.yuanrong.testutils.TestUtils");
        auto creator = YR::Instance(javaCls).SetUrn("abc123").Invoke();
    } catch (YR::Exception &e) {
        printf("error: %s\n", e.what());
        std::string errorCode = "ErrCode: 1001";
        std::string errorMsg = "Failed to split functionUrn: split num 1 is expected to be 7";
        std::string excepMsg = e.what();
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
    try {
        auto javaCls = YR::JavaInstanceClass::FactoryCreate("TestUtils");
        auto creator = YR::Instance(javaCls)
                           .SetUrn("sn:cn:yrk:12345678901234561234567890123456:function:0-yr-stjava:$latest")
                           .Invoke();
        auto ret = creator.JavaFunction<int>("returnInt").Invoke(1);
        YR::Get(ret);
    } catch (YR::Exception &e) {
        printf("error: %s\n", e.what());
        std::string errorCode = "ErrCode: 2002";
        std::string errorMsg = "ClassNotFoundException";
        std::string excepMsg = e.what();
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
}

TEST_F(ActorTest, OrderedInvocations)
{
    YR::InvokeOptions option;
    option.needOrder = true;
    auto creator = YR::Instance(Counter::FactoryCreate).Options(option).Invoke(1);
    for (int i = 0; i < 3; i++) {
        auto ret = creator.Function(&Counter::Add).Invoke(1);
        EXPECT_EQ(*YR::Get(ret), 2 + i);
    }
}

TEST_F(ActorTest, OrderedInvocations_PassNamedInstance)
{
    YR::InvokeOptions option;
    option.needOrder = true;
    auto adder = YR::Instance(Adder::FactoryCreate).Options(option).Invoke(10);
    auto proxy1 = YR::Instance(AdderProxy::FactoryCreate).Invoke();
    auto ret1 = proxy1.Function(&AdderProxy::SetAdder).Invoke(adder);
    YR::Wait(ret1, 30);
    for (int i = 0; i < 3; i++) {
        auto ret = proxy1.Function(&AdderProxy::Add).Invoke(1);
        EXPECT_EQ(*YR::Get(ret), 11 + i);
    }
}

TEST_F(ActorTest, OrderedInvocations_DesignatedNameOfNamedInstance)
{
    std::string designateName = "my_name";
    YR::InvokeOptions option;
    option.needOrder = true;
    auto adder = YR::Instance(Adder::FactoryCreate, designateName).Options(option).Invoke(9);
    auto ret = adder.Function(&Adder::Add).Invoke(1);
    EXPECT_EQ(*YR::Get(ret, 30), 10);
    auto proxy1 = YR::Instance(AdderProxy::FactoryCreate).Invoke();
    auto ret1 = proxy1.Function(&AdderProxy::SetAdderByName).Invoke(designateName);
    YR::Wait(ret1, 30);
    for (int i = 0; i < 3; i++) {
        auto ret = proxy1.Function(&AdderProxy::Add).Invoke(1);
        EXPECT_EQ(*YR::Get(ret), 11 + i);
    }
}

TEST_F(ActorTest, TestGroup)
{
    YR::GroupOptions groupOpts;
    std::string groupName = "group1";
    auto g = YR::Group(groupName, groupOpts);
    YR::InvokeOptions opt;
    opt.groupName = groupName;
    auto ins = YR::Instance(Counter::FactoryCreate).Options(opt).Invoke(1);
    g.Invoke();
    auto res = ins.Function(&Counter::Add).Invoke(1);
    EXPECT_EQ(*YR::Get(res), 2);
}

TEST_F(ActorTest, TestGroupSameLifecycleTrue)
{
    YR::GroupOptions groupOpts;
    groupOpts.sameLifecycle = false;
    std::string groupName = "group1";
    auto g = YR::Group(groupName, groupOpts);
    YR::InvokeOptions opt;
    opt.groupName = groupName;
    auto ins1 = YR::Instance(Counter::FactoryCreate).Options(opt).Invoke(1);
    auto ins2 = YR::Instance(Counter::FactoryCreate).Options(opt).Invoke(1);
    g.Invoke();
    auto res1 = ins1.Function(&Counter::Add).Invoke(1);
    EXPECT_EQ(*YR::Get(res1), 2);
    ins1.Terminate();
    sleep(1);
    auto res2 = ins2.Function(&Counter::Add).Invoke(1);
    EXPECT_EQ(*YR::Get(res2), 2);
}

TEST_F(ActorTest, GroupWaitTimeoutZero)
{
    YR::GroupOptions groupOpts;
    groupOpts.timeout = 0;
    std::string groupName = "group1";
    auto g = YR::Group(groupName, groupOpts);
    YR::InvokeOptions opt1;
    opt1.groupName = groupName;
    auto inst1 = YR::Instance(Adder::FactoryCreate).Options(opt1).Invoke(10);
    g.Invoke();
    bool raise = false;
    try {
        g.Wait();
    } catch (YR::Exception &e) {
        std::string errorCode = "ErrCode: 2002";
        std::string errorMsg = "group create timeout";
        std::string excepMsg = e.what();
        raise = true;
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
    ASSERT_TRUE(raise);
}

TEST_F(ActorTest, GroupInvokeAfterTerminate)
{
    YR::GroupOptions groupOpts;
    std::string groupName = "group1";
    auto g = YR::Group(groupName, groupOpts);
    YR::InvokeOptions opt;
    opt.groupName = groupName;
    auto ins = YR::Instance(Counter::FactoryCreate).Options(opt).Invoke(1);
    g.Invoke();
    g.Terminate();
    bool raise = false;
    try {
        auto obj = ins.Function(&Counter::Add).Invoke(1);
        auto res = *YR::Get(obj);
    } catch (YR::Exception &e) {
        std::string errorCode = "ErrCode: 9000";
        std::string errorMsg = "group ins had been terminated";
        std::string excepMsg = e.what();
        raise = true;
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
    ASSERT_TRUE(raise);
}

/*case
 * @title: 创建actor函数，objID携带workerID
 * @precondition:
 * @step:  创建actor
 * @expect:  1.预期无异常抛出，
 * @expect:  2.objID不携带ds workerID
 */
TEST_F(ActorTest, CheckActorObjIdSuccessfully)
{
    auto creator = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto id = creator.GetObjectId();
    EXPECT_EQ(id.size(), 20);
    auto ret = creator.Function(&Counter::Add).Invoke(1);
    EXPECT_EQ(*YR::Get(ret), 2);
    auto obj = YR::Put(3);
    auto ret1 = creator.Function(&Counter::Add).Invoke(obj);
    EXPECT_EQ(ret1.ID().size(), 20);
    EXPECT_EQ(*YR::Get(ret1), 5);
}

/*
 * Check whether customextension has been written into the request body by viewing log file.
 * Therefore this case is disabled.
 */
TEST_F(ActorTest, DISABLED_InvokeInstanceWithCustomextensionTest)
{
    YR::InvokeOptions opt;
    opt.customExtensions = {
        {"endpoint", "CreateInstance1"}, {"app_name", "CreateInstance2"}, {"tenant_id", "CreateInstance3"}};
    auto counter = YR::Instance(Counter::FactoryCreate).Options(opt).Invoke(1);

    opt.customExtensions = {
        {"endpoint", "InvokeInstance1"}, {"app_name", "InvokeInstance2"}, {"tenant_id", "InvokeInstance3"}};
    auto ret = counter.Function(&Counter::Add).Options(opt).Invoke(1);
    EXPECT_EQ(*YR::Get(ret), 2);
}

/*
 * Check whether preferredAntiOtherLabels has been written into the request body by viewing log file.
 * Therefore this case is disabled.
 */
TEST_F(ActorTest, DISABLED_AntiOtherLabelsSuccess)
{
    YR::InvokeOptions opt;
    auto aff1 = YR::ResourcePreferredAffinity(YR::LabelExistsOperator("label_1"));
    opt.AddAffinity(aff1);
    opt.preferredAntiOtherLabels = true;
    auto creator = YR::Instance(Counter::FactoryCreate).Options(std::move(opt)).Invoke(1);
}

TEST_F(ActorTest, testGracefulShutdownWithTerminate)
{
    YR::KV().Del("shutdownKey");
    auto counter = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto ret = counter.Function(&Counter::Add).Invoke(1);
    EXPECT_EQ(*YR::Get(ret), 2);
    counter.Terminate();
    std::string result = YR::KV().Get("shutdownKey", 30);
    EXPECT_EQ(result, "shutdownValue");
}

TEST_F(ActorTest, testGetWithParam)
{
    std::string key = "kv-id-888";
    YR::KV().Del(key);
    std::string value = "kv-id-888wqdq";
    YR::KV().Set(key, value);
    YR::GetParam param = {.offset = 0, .size = 0};
    YR::GetParams params;
    params.getParams.emplace_back(param);
    std::vector<std::shared_ptr<YR::Buffer>> res = YR::KV().GetWithParam({key}, params);
    EXPECT_EQ(value, std::string((const char *)(res[0]->ImmutableData())));
}

TEST_F(ActorTest, testGetWithParamPartial)
{
    std::string key = "kv-cpp-id-testGetWithParamPartial";
    std::string value = "kv-value123456";
    std::string traceId = "123333";
    YR::GetParam param = {.offset = 0, .size = 0};
    YR::GetParams p1;
    std::vector<std::string> result;
    std::vector<std::string> keys;
    std::vector<std::shared_ptr<YR::Buffer>> res;
    std::string key1 = key + std::to_string(0);
    keys.emplace_back(key1);
    p1.getParams.emplace_back(param);
    YR::KV().Set(key1, value);

    std::string key2 = key + std::to_string(1);
    keys.emplace_back(key2);
    p1.getParams.emplace_back(param);

    res = YR::KV().GetWithParam(keys, p1, 4);
    for (int i = 0; i < res.size(); i++) {
        if (res[i] != nullptr) {
            auto getVal = std::string((const char *)(res[i]->ImmutableData()));
            result.emplace_back(getVal);
            EXPECT_EQ(getVal, value);
        }
    }
    EXPECT_EQ(result.size(), 1);
}

TEST_F(ActorTest, DISABLED_testGracefulShutdownWithManualSigterm)
{
    YR::KV().Del("shutdownKey");
    auto counter = YR::Instance(Counter::FactoryCreate).Invoke(1);
    sleep(10);
    std::string result = YR::KV().Get("shutdownKey", 30);
    EXPECT_EQ(result, "shutdownValue");
}

/*case
 * @title: range调度成功
 * @precondition:
 * @step:  配置range，创建一组actor实例并调用
 * @expect:  1.一组实例调用实例成功
 */
TEST_F(ActorTest, TestRangeSuccess)
{
    auto ins = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto res = ins.Function(&Counter::AddRange).Invoke(1, 1, 1, true, 20, 20);
    int ret = *YR::Get(res, 20);
    EXPECT_EQ(ret, 21);
}

/*case
 * @title: range调度成功,其中step>max-min
 * @precondition:
 * @step:  配置range，创建一组共计2个actor实例并调用
 * @expect:  1.一组共计2个实例调用实例成功
 */
TEST_F(ActorTest, TestRangeSuccess_BigStep)
{
    auto ins = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto res = ins.Function(&Counter::AddRange).Invoke(256, 2, 300, true, 20, 20);
    int ret = *YR::Get(res, 20);
    EXPECT_EQ(ret, 41);
}

/*case
 * @title: range调度成功, 开启同生共死
 * @precondition:
 * @step:  配置range，创建一组actor实例并调用
 * @expect:  1.一组实例调用实例成功,terminate一个实例，其它实例也被terminate
 */
TEST_F(ActorTest, DISABLED_TestRangeSuccess_SameLifecycle)
{
    YR::InstanceRange range;
    range.max = 5;
    range.min = 2;
    range.step = 2;
    YR::InvokeOptions opt;
    opt.instanceRange = range;
    auto instances = YR::Instance(Counter::FactoryCreate).Options(opt).Invoke(1);
    auto insList = instances.GetInstances(15);
    sleep(1);
    insList[0].Terminate();
    sleep(1);
    bool raise = false;
    try {
        for (int i = 1; i < insList.size(); ++i) {
            auto res = insList[i].Function(&Counter::Add).Invoke(1);
            std::cout << "res is " << *YR::Get(res, 12) << std::endl;
        }
    } catch (YR::Exception &e) {
        std::cout << "exception is: " << e.what() << std::endl;
        std::string errorCode = "1011";
        std::string errorMsg = "instance occurs fatal error";
        std::string excepMsg = e.what();
        raise = true;
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
    ASSERT_TRUE(raise);
}

/*case
 * @title: range调度成功, 关闭同生共死
 * @precondition:
 * @step:  配置range，创建一组actor实例并调用
 * @expect:  1.一组实例调用实例成功,terminate一个实例，其它实例不会被terminate
 */
TEST_F(ActorTest, DISABLED_TestRangeSuccess_NoSameLifecycle)
{
    YR::InstanceRange range;
    range.max = 5;
    range.min = 2;
    range.step = 2;
    range.sameLifecycle = false;
    YR::InvokeOptions opt;
    opt.instanceRange = range;
    auto instances = YR::Instance(Counter::FactoryCreate).Options(opt).Invoke(1);
    auto insList = instances.GetInstances(15);
    bool raise = false;
    try {
        for (auto ins : insList) {
            auto res = ins.Function(&Counter::Add).Invoke(1);
            std::cout << "res is " << *YR::Get(res, 12) << std::endl;
            ins.Terminate();
        }
    } catch (YR::Exception &e) {
        raise = true;
    }
    ASSERT_FALSE(raise);
}

/*case
 * @title: range调度失败：非法调用
 * @precondition:
 * @step:  配置range，创建一组actor实例并调用
 * @expect:  1.普通实例句柄调用GetInstances抛错
 * @expect:  2.range句柄调用Function抛错
 */
TEST_F(ActorTest, TestRangeFailed_IncorrectInvokeUsage)
{
    YR::InstanceRange range;
    range.max = 3;
    range.min = 1;
    YR::InvokeOptions opt;
    opt.instanceRange = range;
    opt.customExtensions.insert({"GRACEFUL_SHUTDOWN_TIME", "1"});
    bool raise = false;
    try {
        auto handler = YR::Instance(Counter::FactoryCreate).Invoke(1);
        auto instances = handler.GetInstances();
    } catch (YR::Exception &e) {
        std::string errorCode = "ErrCode: 4006";
        std::string errorMsg = "this function can only be used for range instance handler";
        std::string excepMsg = e.what();
        raise = true;
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
    ASSERT_TRUE(raise);

    raise = false;
    try {
        auto handler = YR::Instance(Counter::FactoryCreate).Options(opt).Invoke(1);
        auto res = handler.Function(&Counter::Add).Invoke(1);
    } catch (YR::Exception &e) {
        std::string errorCode = "ErrCode: 4008";
        std::string errorMsg = "range instance handler cannont be used to invoke directly";
        std::string excepMsg = e.what();
        raise = true;
        ErrorMsgCheck(errorCode, errorMsg, excepMsg);
    }
    ASSERT_TRUE(raise);
}

/*case
 * @title: range调度失败：非法参数
 * @precondition:
 * @step:  配置非法参数的range，创建一组actor实例并调用
 * @expect:  1.参数校验失败
 */
TEST_F(ActorTest, TestRangeFailed_InvalidParam)
{
    YR::InvokeOptions opt;
    std::vector<std::vector<int>> testArr;  // vector of {max, min}
    testArr.push_back({-2, -1});
    testArr.push_back({0, -1});
    testArr.push_back({-1, -2});
    testArr.push_back({-1, 0});
    testArr.push_back({1, 2});

    for (auto &test : testArr) {
        YR::InstanceRange range;
        range.max = test.at(0);
        range.min = test.at(1);
        opt.instanceRange = range;
        bool raise = false;
        try {
            auto handler = YR::Instance(Counter::FactoryCreate).Options(opt).Invoke(1);
        } catch (YR::Exception &e) {
            std::string errorCode = "ErrCode: 1001";
            std::string errorMsg = "please set the min and the max as follows: max = min = -1 or max >= min > 0";
            std::string excepMsg = e.what();
            raise = true;
            ErrorMsgCheck(errorCode, errorMsg, excepMsg);
        }
        ASSERT_TRUE(raise);
    }

    std::vector<int> testStep;
    testStep.push_back(-2);
    testStep.push_back(-1);
    testStep.push_back(0);
    for (auto &test : testStep) {
        YR::InstanceRange range;
        range.max = 10;
        range.min = 1;
        range.step = test;
        opt.instanceRange = range;
        bool raise = false;
        try {
            auto handler = YR::Instance(Counter::FactoryCreate).Options(opt).Invoke(1);
        } catch (YR::Exception &e) {
            std::string errorCode = "ErrCode: 1001";
            std::string errorMsg = "please set the step > 0";
            std::string excepMsg = e.what();
            raise = true;
            ErrorMsgCheck(errorCode, errorMsg, excepMsg);
        }
        ASSERT_TRUE(raise);
    }

    std::vector<int> testTimeout;
    testTimeout.push_back(-2);
    for (auto &test : testTimeout) {
        YR::InstanceRange range;
        range.max = 10;
        range.min = 1;
        range.step = 2;
        YR::RangeOptions rangeOpts;
        rangeOpts.timeout = test;
        range.rangeOpts = rangeOpts;
        opt.instanceRange = range;
        bool raise = false;
        try {
            auto handler = YR::Instance(Counter::FactoryCreate).Options(opt).Invoke(1);
        } catch (YR::Exception &e) {
            std::string errorCode = "ErrCode: 1001";
            std::string errorMsg = "please set the timeout >= -1";
            std::string excepMsg = e.what();
            raise = true;
            ErrorMsgCheck(errorCode, errorMsg, excepMsg);
        }
        ASSERT_TRUE(raise);
    }
}

/*case
 * @title: range调度成功,其中getInstances超时
 * @precondition:
 * @step:  配置range，创建一组actor实例并调用，设置非法超时
 * @expect:  抛出参数错误的异常。返回-1
 */
TEST_F(ActorTest, TestRangeFailed_InvalidGetInstancesTimeout)
{
    auto ins = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto res = ins.Function(&Counter::AddRange).Invoke(2, 2, 2, true, 20, -2);
    ASSERT_THROW(*YR::Get(res, 20), YR::Exception);
}

/*case
 * @title: range调度失败,其中创建instances超时
 * @precondition:
 * @step:  配置range，创建一组actor实例并调用，设置超时
 * @expect:  抛出超时的异常。返回-1
 */
TEST_F(ActorTest, TestRangeFailed_CreateInstancesTimeout)
{
    auto ins = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto res = ins.Function(&Counter::AddRange).Invoke(256, 256, 1, true, 1, 1);
    ASSERT_THROW(*YR::Get(res, 20), YR::Exception);
}

/*case
 * @title: range调度失败,其中getInstances超时
 * @precondition:
 * @step:  配置range，创建一组actor实例并调用，设置get超时
 * @expect:  抛出超时的异常。返回-1
 */
TEST_F(ActorTest, TestRangeFailed_GetInstancesTimeout)
{
    auto ins = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto res = ins.Function(&Counter::AddRange).Invoke(2, 2, 2, true, 20, 0);
    ASSERT_THROW(*YR::Get(res, 20), YR::Exception);
}

TEST_F(ActorTest, testLogMessageOfSigterm)
{
    auto instance = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto ins = instance.Function(&Counter::GetSigterm).Invoke();
    try {
        int n1 = *YR::Get(ins);
    } catch (YR::Exception &e) {
        printf("error: %s\n", e.what());
        std::string error_code = "ErrCode: 2002";
        std::string error_msg = "SIGTERM";
        std::string excep_msg = e.what();
        ErrorMsgCheck(error_code, error_msg, excep_msg);
    }
}

/*case
 * @title: 创建actor函数，云下put、云上get
 * @precondition:
 * @step:  创建actor
 * @expect:  1.预期无异常抛出，
 * @expect:  2.get成功，返回正确返回值。
 */
TEST_F(ActorTest, DISABLED_TestTenantIdWithSetUrnSuccessfully)
{
    auto cppCls = YR::CppInstanceClass::FactoryCreate("Counter::FactoryCreate");
    auto creator =
        YR::Instance(cppCls).SetUrn("sn:cn:yrk:12345678901234561234567890123456:function:0-yr-stcpp:$latest").Invoke(1);
    auto ret = creator.CppFunction<int>("&Counter::Add").Invoke(1);
    std::vector<YR::ObjectRef<int>> objs;
    objs.push_back(ret);
    auto instance = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto res = instance.Function(&Counter::AddRef).Invoke(objs);
    EXPECT_EQ(*YR::Get(res), 2);
}

/*case
 * @title: 创建actor函数，云下put、云上get
 * @precondition:
 * @step:  创建actor
 * @expect:  1.预期无异常抛出，
 * @expect:  2.get成功，返回正确返回值。
 */
TEST_F(ActorTest, TestTenantIdSuccessfully)
{
    auto instance = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto ret = instance.Function(&Counter::Add).Invoke(1);
    std::vector<YR::ObjectRef<int>> objs;
    objs.push_back(ret);
    auto instance2 = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto res = instance2.Function(&Counter::AddRef).Invoke(objs);
    EXPECT_EQ(*YR::Get(res), 3);
}

/*case
 * @title: 创建actor成功，并调用SaveState、LoadState成功
 * @precondition:
 * @step:  创建Counter
 * @step:  调用SaveState
 * @step:  增加计数
 * @step:  调用LoadState
 * @expect:  1.返回值为0
 */
TEST_F(ActorTest, ActorSaveStateAndLoadStateSuccessfully)
{
    auto instance = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto ret = instance.Function(&Counter::Add).Invoke(1);
    EXPECT_EQ(*YR::Get(ret), 2);
    auto ret1 = instance.Function(&Counter::SaveState).Invoke();
    EXPECT_EQ(*YR::Get(ret1), 2);
    auto ret2 = instance.Function(&Counter::Add).Invoke(1);
    EXPECT_EQ(*YR::Get(ret2), 3);
    auto ret3 = instance.Function(&Counter::LoadState).Invoke();
    EXPECT_EQ(*YR::Get(ret3), 3);
    auto ret4 = instance.Function(&Counter::Add).Invoke(1);
    EXPECT_EQ(*YR::Get(ret4), 3);
    auto ret5 = instance.Function(&Counter::GetRecoverFlag).Invoke();
    EXPECT_EQ(*YR::Get(ret5), 1);
}

/*case
 * @title: 创建actor成功，并Recover成功
 * @precondition:
 * @step:  创建Counter
 * @step:  调用SEGV
 * @step:  调用GetRecoverFlag
 * @expect:  1.recover次数为1
 */
TEST_F(ActorTest, DISABLED_ActorRecoverSuccessfully)
{
    YR::InvokeOptions opt;
    opt.recoverRetryTimes = 1;
    auto instance = YR::Instance(Counter::FactoryCreate).Options(opt).Invoke(1);
    auto ret = instance.Function(&Counter::GetPid).Invoke();
    auto pid = *YR::Get(ret);
    std::cout << "Counter pid: " << pid << std::endl;
    kill(pid, 9);
    auto ret1 = instance.Function(&Counter::GetRecoverFlag).Invoke();
    EXPECT_EQ(*YR::Get(ret1, 30), 1);
}

/*case
 * @title: 创建actor成功，并Recover成功
 * @precondition:
 * @step:  创建Counter
 * @step:  调用SEGV
 * @step:  调用GetGroupRecoverFlag
 * @expect:  1.结果为11
 */
TEST_F(ActorTest, ActorGroupRecoverSuccessfully)
{
    YR::InvokeOptions opt;
    opt.recoverRetryTimes = 1;
    auto instance = YR::Instance(Counter::FactoryCreate).Options(opt).Invoke(1);
    auto ret0 = instance.Function(&Counter::SaveGroupState).Invoke();
    auto state = *YR::Get(ret0);
    std::cout << "Counter state: " << state << std::endl;
    auto ret = instance.Function(&Counter::GetPid).Invoke();
    auto pid = *YR::Get(ret);
    std::cout << "Counter pid: " << pid << std::endl;
    kill(pid, 9);
    auto ret1 = instance.Function(&Counter::GetGroupRecoverFlag).Invoke();
    EXPECT_EQ(*YR::Get(ret1, 30), 11);
}

/*case
 * @title: 创建actor，配置envVars参数
 * @precondition:
 * @step:  创建Counter
 * @step:  调用returnActorEnvVar
 * @expect:  1.返回配置的环境变量值
 */
TEST_F(ActorTest, ActorReturnEnvVars)
{
    YR::InvokeOptions opt;
    std::string key = "A";
    std::string value = "A_VARS";
    opt.envVars[key] = value;
    auto instance = YR::Instance(Counter::FactoryCreate).Options(opt).Invoke(1);
    auto ret = instance.Function(&Counter::returnActorEnvVar).Invoke(key);
    auto res = *YR::Get(ret);
    EXPECT_EQ(res, value);
}

/*case
 * @title: 测试实例保序
 * @precondition:
 * @step:  1.调用Wait接口传入ObjectRef1、ObjectRef2
 * @step:  2.ObjectRef1返回，ObjectRef2不返回
 * @expect:  Wait返回的readyids中仅包含ObjectRef1
 * @title: Wait接口传入多个ObjectRef，其中有返回及不返回
 */
TEST_F(ActorTest, ActorOrderTest)
{
    std::vector<YR::NamedInstance<Counter>> instances;
    for (int i = 0; i < 3; i++) {
        instances.push_back(YR::Instance(Counter::FactoryCreate).Invoke(1));
    }
    std::vector<YR::ObjectRef<int>> vec;
    for (auto &instance : instances) {
        vec.push_back(instance.Function(&Counter::Add).Invoke(1));
        vec.push_back(instance.Function(&Counter::Sleep).Invoke());
    }
    auto res = YR::Wait(vec, vec.size(), 5);
    ASSERT_EQ(res.first.size(), 3);
    ASSERT_EQ(res.first.size() + res.second.size(), vec.size());
}

/*case
 * @title: 测试invoke超时后能够给runtime发送signal请求驱逐pending线程
 * @precondition:
 * @step:  1.创建并发度为1的actor实例
 * @step:  2.调用阻塞方法，设置invoke超时时间
 * @step:  3.invoke请求超时后立即调用另一个函数invoke请求
 * @expect: 第一次invoke请求失败，第二次invoke请求成功
 */
TEST_F(ActorTest, ActorTaskPendingTest)
{
    YR::InvokeOptions opts;
    opts.timeout = 2;
    opts.customExtensions["Concurrency"] = "2";

    auto instance = YR::Instance(Counter::FactoryCreate).Options(opts).Invoke(1);
    auto ret1 = instance.Function(&Counter::Sleep).Options(opts).Invoke();
    auto ret2 = instance.Function(&Counter::Sleep).Options(opts).Invoke();
    ASSERT_THROW(*YR::Get(ret1), YR::Exception);

    auto ret3 = instance.Function(&Counter::Add).Invoke(1);
    auto res = *YR::Get(ret3);
    ASSERT_EQ(res, 2);
}

/*case
 * @title: 测试获取cpp实例句柄
 * @precondition:
 * @step:  1.创建具名实例
 * @step:  2.发起实例函数调用
 * @step:  3.调用getInstance接口获取实例句柄
 * @step:  4.基于获取的实例句柄发起函数调用
 */
TEST_F(ActorTest, ActorGetTest)
{
    std::string name = "test-cpp-get";
    auto ins = YR::Instance(Counter::FactoryCreate, name).Invoke(1);
    auto res = ins.Function(&Counter::Add).Invoke(1);
    EXPECT_EQ(*YR::Get(res), 2);

    std::string ns = "";
    auto insGet = YR::GetInstance<Counter>(name, ns, 60);
    auto resGet = insGet.Function(&Counter::Add).Invoke(1);
    EXPECT_EQ(*YR::Get(resGet), 3);
}

/*case
 * @title: 创建实例CounterA和CounterProxy
 * @precondition:
 * @step:  1. CounterA为保序实例
 * @step:  2.分别对CounterA和CounterProxy发起一次Add invoke调用，保证实例正常拉起
 * @step:  3.对CounterProxy发起函数请求，该函数会在云上发起get_instance请求，获取CoutnerA实例句柄，并通过该实例句柄发起两次invoke调用，返回两次invoke结果之和。
 * @expect: 正常返回函数请求不卡住。
 */
TEST_F(ActorTest, GetOrderInstanceOnCloudTest)
{
    std::string name = "CounterOnCloudTest";
    auto counter = YR::Instance(Counter::FactoryCreate, name).Invoke(1);
    auto counterObj = counter.Function(&Counter::Add).Invoke(1);
    EXPECT_EQ(*YR::Get(counterObj), 2);

    auto counterProxy = YR::Instance(CounterProxy::FactoryCreate).Invoke(1);
    auto proxyObj = counterProxy.Function(&CounterProxy::Add).Invoke();
    EXPECT_EQ(*YR::Get(proxyObj), 1);

    auto onCloudObj = counterProxy.Function(&CounterProxy::GetCounterAndInvoke).Invoke(name);
    EXPECT_EQ(*YR::Get(onCloudObj), 7);
}