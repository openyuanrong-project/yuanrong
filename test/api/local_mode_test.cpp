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
#include "yr/api/err_type.h"
#include "yr/parallel/parallel_for.h"
#include "yr/yr.h"
#include "api/cpp/src/config_manager.h"
namespace YR {
namespace test {
using testing::HasSubstr;
const static size_t g_threadPoolSize = 8;
class LocalTest : public testing::Test {
public:
    LocalTest() {}
    ~LocalTest() {}
    void SetUp() override
    {
        YR::Config conf;
        conf.mode = YR::Config::Mode::LOCAL_MODE;
        conf.logLevel = "DEBUG";
        conf.logDir = "/tmp/log";
        conf.threadPoolSize = g_threadPoolSize;
        YR::Init(conf);
    }
    void TearDown() override
    {
        YR::Finalize();
    }
};

int PlusOne(int x)
{
    return x + 1;
}

YR_INVOKE(PlusOne)
TEST_F(LocalTest, When_Invoke_Ten_Task_Should_All_Return_Correct_Result)
{
    std::vector<YR::ObjectRef<int>> n2;
    int k = 10;
    for (int i = 0; i < k; i++) {
        auto r2 = YR::Function(PlusOne).Invoke(i);
        n2.emplace_back(r2);
    }
    for (int i = 0; i < k; i++) {
        auto integer = *YR::Get(n2[i]);
        EXPECT_EQ(i + 1, integer);
    }
}

TEST_F(LocalTest, When_Invoke_Task_ObjID_Not_Contain_WorkerID)
{
    std::vector<YR::ObjectRef<int>> n2;
    auto r2 = YR::Function(PlusOne).Invoke(1);
    auto id = r2.ID();
    EXPECT_EQ(id.size(), 20);
    auto integer = *YR::Get(r2);
    EXPECT_EQ(2, integer);
}

static std::mutex mu;
static bool destructed = false;
static std::condition_variable cv;
class Counter {
public:
    Counter() = default;
    Counter(const Counter &c)
    {
        if (this != &c) {
            *this = c;
        }
    }

    Counter &operator=(const Counter &c)
    {
        count = c.count;
        return *this;
    }

    ~Counter()
    {
        std::unique_lock<std::mutex> lock(mu);
        destructed = true;
        cv.notify_one();
    }

    Counter(int init)
    {
        count = init;
    }

    static Counter *FactoryCreate(int init)
    {
        return new Counter(init);
    }

    int Add(int x)
    {
        count += x;
        return count;
    }

    int ParallelFor()
    {
        auto f = [this](size_t start, size_t end, const YR::Parallel::Context &ctx) {
            for (size_t i = start; i < end; i++) {
                std::lock_guard<std::mutex> lock(ctxIdSetMutex);
                ctxIdSet.emplace(ctx.id);
            }
            std::this_thread::yield();
            std::this_thread::sleep_for(std::chrono::microseconds(10));
        };
        YR::Parallel::ParallelFor<size_t>(0, 1000, f, 1);
        return 0;
    }

    size_t GetCtxIdsSize()
    {
        std::lock_guard<std::mutex> lock(ctxIdSetMutex);
        return ctxIdSet.size();
    }

    int Get(void)
    {
        return count;
    }

    int Throw()
    {
        throw std::runtime_error("runtime error");
    }

    YR_STATE(count);

public:
    std::mutex ctxIdSetMutex;
    std::unordered_set<size_t> ctxIdSet;
    int count;
};

class Foo {
public:
    Foo() = default;
    ~Foo() = default;

    Foo(const Foo &other)
    {
        clue = other.clue + ",CopyCons";
    }

    Foo &operator=(const Foo &other)
    {
        clue = other.clue + ",CopyAss";
        return *this;
    }

    Foo(Foo &&other)
    {
        clue = std::move(other.clue);
        clue += ",MoveCons";
    }

    Foo &operator=(Foo &&other)
    {
        clue = std::move(other.clue);
        clue += ",MoveAss";
        return *this;
    }

    std::string Clue(void) const
    {
        return clue;
    }

    // private:
    std::string clue;

public:
    MSGPACK_DEFINE(clue);
};

std::string TestFoo(const Foo &f)
{
    return f.Clue();
}

YR_INVOKE(Counter::FactoryCreate, &Counter::Add, &Counter::Get, &Counter::Throw, &Counter::ParallelFor,
          &Counter::GetCtxIdsSize, TestFoo);

TEST_F(LocalTest, cpp_invoke_classinstance_foo_localmode)
{
    Foo foo;
    auto r = YR::Function(TestFoo).Invoke(foo);
    auto v = *YR::Get(r);

    std::string::size_type n = 0;
    int count = 0;
    do {
        n = v.find("Copy", n);
        if (n != std::string::npos) {
            count++;
            n = n + std::string("Copy").length();
            if (n >= v.length()) {
                break;
            }
        }
    } while (n != std::string::npos);

    EXPECT_LE(count, 2);
}

TEST_F(LocalTest, When_Invoke_Actor_Should_Return_Final_Correct_Result)
{
    auto counter = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto res = counter.Function(&Counter::Add).Invoke(3);
    auto v = *YR::Get(res);
    EXPECT_EQ(4, v);
    res = counter.Function(&Counter::Add).Invoke(3);
    v = *YR::Get(res);
    EXPECT_EQ(7, v);
    res = counter.Function(&Counter::Get).Invoke();
    v = *YR::Get(res);
    EXPECT_EQ(7, v);
}

TEST_F(LocalTest, When_Invoke_Actor_Should_Not_Contain_WorkerId)
{
    auto counter = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto id = counter.GetObjectId();
    EXPECT_EQ(id.size(), 20);
    auto res = counter.Function(&Counter::Add).Invoke(3);
    auto v = *YR::Get(res);
    EXPECT_EQ(4, v);
}

TEST_F(LocalTest, When_Put_Should_Return_Correct_Result)
{
    int val = 123;
    auto ref = YR::Put<int>(val);
    EXPECT_EQ(val, *YR::Get(ref));

    int init = 10;
    Counter c = Counter(10);
    auto ref_ = YR::Put(c);
    auto c_ = *YR::Get(ref_);
    EXPECT_EQ(init, c.count);
}

TEST_F(LocalTest, When_Put_Should_Not_Contain_WorkerId)
{
    int val = 123;
    auto ref = YR::Put<int>(val);
    EXPECT_EQ(ref.ID().size(), 20);
    EXPECT_EQ(val, *YR::Get(ref));
}

TEST_F(LocalTest, When_Do_KV_Should_Return_Correct_Result)
{
    std::string key = "kv-id-888";
    std::string value = "kv-value-888";
    YR::KV().Write(key, value);

    std::shared_ptr<std::string> result = YR::KV().Read<std::string>(key);
    EXPECT_EQ(value, *result);

    YR::KV().Del(key);

    // Legacy API
    YR::KV().Set(key, value);

    std::string result2 = YR::KV().Get(key);
    EXPECT_EQ(value, result2);

    YR::KV().Del(key);
}

TEST_F(LocalTest, When_Get_Repeated_Keys_Should_Return_Success_Test)
{
    std::string key = "key";
    std::string value = "value";
    KV().Set(key, value);
    std::vector<std::string> keys = {key, key};
    auto values = KV().Get(keys);
    for (const auto &v : values) {
        EXPECT_EQ(v, value);
    }
}

TEST_F(LocalTest, MSetTxTest)
{
    // case 1
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
    for (size_t i = 0; i < keys.size(); i++) {
        EXPECT_EQ(vals[i], actualVals[i]);
    }
    YR::KV().Del(keys);

    // case 2
    std::vector<char *> valsPtr;
    std::vector<size_t> lens;
    for (const auto &val : vals) {
        valsPtr.push_back(const_cast<char *>(val.c_str()));
        lens.push_back(val.size());
    }
    YR::KV().MSetTx(keys, valsPtr, lens, YR::ExistenceOpt::NX);
    actualVals = YR::KV().Get(keys, 300);
    for (size_t i = 0; i < keys.size(); i++) {
        EXPECT_EQ(vals[i], actualVals[i]);
    }
    YR::KV().Del(keys);

    // case 3
    YR::KV().MWriteTx(keys, vals, YR::ExistenceOpt::NX);
    std::vector<std::shared_ptr<std::string>> actualVals2 = YR::KV().Read<std::string>(keys, 300, false);
    for (size_t i = 0; i < keys.size(); i++) {
        EXPECT_EQ(vals[i], *(actualVals2[i]));
    }
    YR::KV().Del(keys);

    // case 4
    YR::KV().Set(keys[1], vals[1]);
    // key key1 already exist
    EXPECT_THROW(YR::KV().MSetTx(keys, vals, YR::ExistenceOpt::NX), Exception);
    YR::KV().Del(keys[1]);
    EXPECT_NO_THROW(YR::KV().MSetTx(keys, vals, YR::ExistenceOpt::NX));
    YR::KV().Del(keys);

    // case 5
    std::vector<std::string> keys2, vals2;
    EXPECT_THROW(YR::KV().MSetTx(keys2, vals2, YR::ExistenceOpt::NX), Exception);

    // case 6
    std::vector<std::string> keys3, vals3;
    int exceedNum = 10;
    for (int i = 0; i < exceedNum; i++) {
        std::string key = "Key" + std::to_string(i);
        std::string value = "Value" + std::to_string(i);
        keys3.emplace_back(key);
        vals3.emplace_back(value);
    }
    EXPECT_THROW(YR::KV().MSetTx(keys3, vals3, YR::ExistenceOpt::NX), Exception);

    // case 7
    std::vector<std::string> keys4, vals4;
    keys4.emplace_back("");
    vals4.emplace_back("emptyKey");
    EXPECT_THROW(YR::KV().MSetTx(keys4, vals4, YR::ExistenceOpt::NX), Exception);
}

TEST_F(LocalTest, Test_When_Actor_Currency_Call_ParallelFor_Should_Not_Be_Stuck)
{
    auto counter = YR::Instance(Counter::FactoryCreate).Invoke(1);
    std::vector<YR::ObjectRef<int>> rets;
    for (int i = 0; i < 4; i++) {
        rets.emplace_back(counter.Function(&Counter::ParallelFor).Invoke());
    }
    // test wait not stuck
    YR::Wait(rets, rets.size(), -1);
    counter.Function(&Counter::ParallelFor).Invoke();
    auto ret = *(YR::Get(counter.Function(&Counter::GetCtxIdsSize).Invoke()));
    EXPECT_GE(ret, 1);
}

TEST_F(LocalTest, TestActorTerminate)
{
    auto counter = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto res = counter.Function(&Counter::Add).Invoke(3);
    auto v = *YR::Get(res);
    EXPECT_EQ(4, v);
    counter.Terminate();
    res = counter.Function(&Counter::Add).Invoke(3);
    EXPECT_THROW(YR::Get(res, 1), Exception);
}

TEST_F(LocalTest, cpp_local_kv_read_error_keys_allowpatital_true)
{
    std::string key;
    std::string value;
    std::vector<std::string> keys;
    for (int i = 0; i < 10; ++i) {
        key = "cpp_local_kv_read_error_keys_allowpatital_true" + std::to_string(i);
        value = "value" + std::to_string(i);
        try {
            YR::KV().Write(key, value);
        } catch (YR::Exception &e) {
            std::cout << e.what() << std::endl;
            EXPECT_EQ(0, 1);
        }
        keys.push_back(key);
    }
    keys.push_back("noValueKey1");
    keys.push_back("noValueKey2");
    keys.push_back("noValueKey3");
    auto returnVal = YR::KV().Read<std::string>(keys, 1, true);
    for (int i = 0; i < returnVal.size(); ++i) {
        if (returnVal[i]) {
            std::cout << i << "-> kv read value is: " << *returnVal[i] << std::endl;
        }
    }
    EXPECT_EQ(returnVal.size(), 13) << "KV Read failed";
    YR::KV().Del(keys);
}

TEST_F(LocalTest, cpp_local_kv_exist)
{
    std::string key;
    std::string value;
    std::vector<std::string> keys;
    for (int i = 0; i < 8; ++i) {
        if (i % 2 == 0) {
            key = "cpp_local_kv_exist" + std::to_string(i);
            value = "value" + std::to_string(i);
            YR::KV().Set(key, value);
            keys.push_back(key);
        } else {
            keys.push_back("noValueKey" + std::to_string(i));
        }

    }
    auto exists = YR::KV().Exist(keys);
    for (int i = 0; i < 8; ++i) {
        std::cout << keys[i] << "-> kv exist is: " << exists[i] << std::endl;
        if (i % 2 == 0) {
            EXPECT_EQ(exists[i], true);
        } else {
            EXPECT_EQ(exists[i], false);
        }
    }
}

int func_throw()
{
    throw std::runtime_error("runtime error");
}

int func_throw_string()
{
    throw std::string("something");
}

const static std::string customExceptionMsg = "a custom exception";
int func_throw_exception()
{
    throw std::invalid_argument(customExceptionMsg);
    return 1;
}

YR_INVOKE(func_throw, func_throw_string, func_throw_exception);

TEST_F(LocalTest, CatchException)
{
    auto obj = Function(func_throw).Invoke();
    EXPECT_THROW(
        {
            try {
                Wait(obj);
            } catch (const Exception &e) {
                EXPECT_EQ(e.Code(), ErrorCode::ERR_USER_FUNCTION_EXCEPTION);
                EXPECT_THAT(e.what(), HasSubstr("exception happens when executing user's function"));
                throw;
            }
        },
        Exception);
    EXPECT_THROW(
        {
            try {
                Get(obj);
            } catch (const Exception &e) {
                EXPECT_EQ(e.Code(), ErrorCode::ERR_USER_FUNCTION_EXCEPTION);
                EXPECT_THAT(e.what(), HasSubstr("exception happens when executing user's function"));
                throw;
            }
        },
        Exception);

    obj = Function(func_throw_string).Invoke();
    EXPECT_THROW(
        {
            try {
                Wait(obj);
            } catch (const Exception &e) {
                EXPECT_THAT(e.what(), HasSubstr("non-standard exception is thrown"));
                throw;
            }
        },
        Exception);
    EXPECT_THROW(
        {
            try {
                Get(obj);
            } catch (const Exception &e) {
                EXPECT_THAT(e.what(), HasSubstr("non-standard exception is thrown"));
                throw;
            }
        },
        Exception);

    destructed = false;
    auto instance = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto obj2 = instance.Function(&Counter::Throw).Invoke();
    EXPECT_THROW(
        {
            try {
                Wait(obj2);
            } catch (const Exception &e) {
                EXPECT_EQ(e.Code(), ErrorCode::ERR_USER_FUNCTION_EXCEPTION);
                EXPECT_THAT(e.what(), HasSubstr("exception happens when executing user's function"));
                throw;
            }
        },
        Exception);
    instance.Terminate();
    {
        std::unique_lock<std::mutex> lock(mu);
        cv.wait_for(lock, std::chrono::milliseconds(10), []{ return destructed; });
    }
    EXPECT_EQ(destructed, true);
    // Sometimes there are some of invoke not processed in threadpool,
    // It should not throw an exception before wait or get.
    instance.Function(&Counter::Throw).Invoke();
}

TEST_F(LocalTest, WaitConcurrencyCatchException)
{
    std::vector<ObjectRef<int>> vec;
    int num = 20;
    int timeout = 1;
    for (int i = 0; i < num; ++i) {
        auto obj = Function(func_throw).Invoke();
        vec.emplace_back(std::move(obj));
    }

    EXPECT_THROW(Wait(vec, num, timeout), std::exception);
    EXPECT_THROW(Get(vec, timeout), std::exception);
}

TEST_F(LocalTest, ExceptionShouldShowDetailMsg)
{
    try {
        auto obj = Function(func_throw_exception).Invoke();
        auto ret = *YR::Get(obj);
    } catch (const std::exception &e) {
        const std::string msg = e.what();
        std::cout << msg << std::endl;
        EXPECT_TRUE(msg.find(customExceptionMsg) != std::string::npos);
    }
}

TEST_F(LocalTest, StopLocalModeRuntime)
{
    YR::internal::LocalModeRuntime runtime;
    runtime.Init();
    ASSERT_NO_THROW(runtime.Stop());
}

TEST_F(LocalTest, LocalModeThreadPoolSize)
{
    YR::Finalize();
    YR::Config conf;
    conf.mode = YR::Config::Mode::LOCAL_MODE;
    conf.logLevel = "DEBUG";
    conf.logDir = "/tmp/log";
    conf.localThreadPoolSize = 65;
    YR::Init(conf);
    ASSERT_EQ(ConfigManager::Singleton().localThreadPoolSize, conf.localThreadPoolSize);
}
}  // namespace test
}  // namespace YR