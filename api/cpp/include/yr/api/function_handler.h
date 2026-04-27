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

#pragma once

#include <future>
#include <vector>

#include <boost/type_traits/is_same.hpp>

#include "yr/api/args_check.h"
#include "yr/api/cross_lang.h"
#include "yr/api/function_manager.h"
#include "yr/api/invoke_arg.h"
#include "yr/api/invoke_options.h"
#include "yr/api/local_mode_runtime.h"
#include "yr/api/runtime.h"
namespace YR {
template <typename F>

/**
 * @brief Type alias for the return type of the function
 *
 * ReturnType is an alias for the type returned by the function F. It is
 * defined using boost::callable_traits::return_type_t to automatically
 * extract the return type from the function signature.
 */
using ReturnType = boost::callable_traits::return_type_t<F>;

const std::string PY_PLACEHOLDER = "__YR_PLACEHOLDER__";
namespace internal {

static inline void AddPythonPlaceholder(YR::internal::FunctionLanguage language, std::vector<InvokeArg> &invokeArgs)
{
    if (language != YR::internal::FunctionLanguage::FUNC_LANG_PYTHON) {
        return;
    }
    InvokeArg invokeArg{};
    invokeArg.buf = std::move(Serialize(PY_PLACEHOLDER));
    invokeArg.isRef = false;
    invokeArgs.emplace_back(std::move(invokeArg));
}

template <typename F>
using GetClassOfWorker = std::remove_pointer_t<boost::callable_traits::return_type_t<F>>;

template <typename T>
using Promise = std::promise<std::shared_ptr<T>>;

template <typename T>
static inline T ParseArg(ObjectRef<T> &arg)
{
    if (!arg.IsLocal()) {
        throw YR::Exception::InvalidParamException("cannot pass cluster object ref as local invoke args");
    }
    return *YR::internal::GetLocalModeRuntime()->Get(arg, -1);
}

template <typename T>
static inline T &ParseArg(T &arg)
{
    return arg;
}

template <typename R, typename Func, typename... Args, size_t... I>
static std::enable_if_t<std::is_void<R>::value, std::shared_ptr<R>> LocalCall(Func func, std::tuple<Args...> &&args,
                                                                              std::index_sequence<I...>)
{
    using ArgsTuple = boost::callable_traits::args_t<Func>;
    func(((typename std::tuple_element<I, ArgsTuple>::type)ParseArg(std::get<I>(args)))...);
    return std::make_shared<int>(0);
}

template <typename R, typename Func, typename... Args, size_t... I>
static std::enable_if_t<!std::is_void<R>::value, std::shared_ptr<R>> LocalCall(Func func, std::tuple<Args...> &&args,
                                                                               std::index_sequence<I...>)
{
    using ArgsTuple = boost::callable_traits::args_t<Func>;
    R ret = func(((typename std::tuple_element<I, ArgsTuple>::type)ParseArg(std::get<I>(args)))...);
    return std::make_shared<R>(ret);
}

template <typename F, typename... Args>
static std::enable_if_t<IsCrossLang<F>::value> ExecuteFunction(const std::string &id,
                                                               std::shared_ptr<Promise<ReturnType<F>>> p, F func,
                                                               std::tuple<Args...> args)
{
}

template <typename F, typename... Args>
static std::enable_if_t<!IsCrossLang<F>::value> ExecuteFunction(const std::string &id,
                                                                std::shared_ptr<Promise<ReturnType<F>>> p, F func,
                                                                std::tuple<Args...> args)
{
    try {
        auto ret = LocalCall<ReturnType<F>>(func, std::move(args), std::index_sequence_for<Args...>{});
        p->set_value(ret);
        YR::internal::GetLocalModeRuntime()->SetReady(id);
    } catch (YR::Exception &e) {
        auto exceptionPtr = std::make_exception_ptr(e);
        p->set_exception(exceptionPtr);
        YR::internal::GetLocalModeRuntime()->SetException(id, exceptionPtr);
    } catch (std::exception &e) {
        std::stringstream ss;
        ss << "exception happens when executing user's function: " << e.what();
        auto exceptionPtr = std::make_exception_ptr(Exception::UserCodeException(ss.str()));
        p->set_exception(exceptionPtr);
        YR::internal::GetLocalModeRuntime()->SetException(id, exceptionPtr);
    } catch (...) {
        auto exceptionPtr = std::make_exception_ptr(Exception::UserCodeException("non-standard exception is thrown"));
        p->set_exception(exceptionPtr);
        YR::internal::GetLocalModeRuntime()->SetException(id, exceptionPtr);
    }
}

template <typename ArgType>
static void PackInvokeArgsImpl(YR::internal::FunctionLanguage language, std::vector<InvokeArg> &invokeArgs,
                               ArgType &arg)
{
    AddPythonPlaceholder(language, invokeArgs);
    InvokeArg invokeArg{};
    localNestedObjList.clear();
    if constexpr (boost::is_same<ArgType, Buffer>::value) {
        invokeArg.yrBuf = std::move(arg);
    } else {
        invokeArg.buf = std::move(Serialize(arg));  // Serialize add nested objects to localNestedObjList
        invokeArg.nestedObjects.swap(localNestedObjList);
    }
    localNestedObjList.clear();
    invokeArg.isRef = false;
    invokeArgs.emplace_back(std::move(invokeArg));
}

template <typename ArgType>
static void PackInvokeArgsImpl(YR::internal::FunctionLanguage language, std::vector<InvokeArg> &invokeArgs,
                               ArgType &&arg)
{
    PackInvokeArgsImpl(language, invokeArgs, arg);
}

template <typename ArgType>
static void PackInvokeArgsImpl(YR::internal::FunctionLanguage language, std::vector<InvokeArg> &invokeArgs,
                               ObjectRef<ArgType> &arg)
{
    AddPythonPlaceholder(language, invokeArgs);
    InvokeArg invokeArg{};
    invokeArg.buf = std::move(Serialize(arg));
    invokeArg.isRef = true;
    invokeArg.objId = arg.ID();
    invokeArgs.emplace_back(std::move(invokeArg));
}

template <typename ArgType>
static void PackInvokeArgsImpl(YR::internal::FunctionLanguage language, std::vector<InvokeArg> &invokeArgs,
                               ObjectRef<ArgType> &&arg)
{
    PackInvokeArgsImpl(language, invokeArgs, arg);
}

template <typename ArgType>
static void PackInvokeArgsImpl(YR::internal::FunctionLanguage language, std::vector<InvokeArg> &invokeArgs,
                               std::vector<ObjectRef<ArgType>> &arg)
{
    AddPythonPlaceholder(language, invokeArgs);
    InvokeArg invokeArg{};
    invokeArg.buf = std::move(Serialize(arg));
    invokeArg.isRef = false;
    invokeArg.nestedObjects.clear();
    for (auto &a : arg) {
        invokeArg.nestedObjects.insert(a.ID());
    }
    invokeArgs.emplace_back(std::move(invokeArg));
}

template <typename ArgType>
static void PackInvokeArgsImpl(YR::internal::FunctionLanguage language, std::vector<InvokeArg> &invokeArgs,
                               std::vector<ObjectRef<ArgType>> &&arg)
{
    PackInvokeArgsImpl(language, invokeArgs, arg);
}

template <typename... OtherArgTypes>
static void PackInvokeArgs(YR::internal::FunctionLanguage language, std::vector<InvokeArg> &invokeArgs,
                           OtherArgTypes &&...args)
{
    std::initializer_list<int>{(PackInvokeArgsImpl(language, invokeArgs, std::forward<OtherArgTypes>(args)), 0)...};
}
}  // namespace internal

template <typename F>
class FunctionHandler {
public:
    explicit FunctionHandler(F f) : func(f) {}
    FunctionHandler(const internal::FuncMeta &funcMeta, F f) : func(f), funcMeta(funcMeta) {}
    FunctionHandler(const FunctionHandler &other)
    {
        this->func = other.func;
        this->funcMeta = other.funcMeta;
        this->opts = other.opts;
    }
    FunctionHandler &operator=(const FunctionHandler &other)
    {
        if (&other != this) {
            this->func = other.func;
            this->funcMeta = other.funcMeta;
            this->opts = other.opts;
        }
    }
    ~FunctionHandler() = default;

    /*!
      @brief Invokes the registered function with the provided arguments

      The Invoke method sends a request to the backend to execute the registered
      function with the specified arguments. The function must have been
      registered using the YR_INVOKE macro. The arguments passed to the Invoke
      method must match the function's parameter types and number, otherwise a
      compile-time error will occur. It is crucial to ensure that the argument
      types match exactly to avoid potential issues caused by implicit type
      conversions.

      @tparam Args The types of the arguments to be passed to the function.

      @param args The arguments to be passed to the function. The types and number
      of arguments must match the function's signature.

      @returns ObjectRef<ReturnType<F>>: An ObjectRef object representing the
      return value of the function. This object acts as a key and requires the
      YR::Get method to be called to obtain the actual value.

      @throws Exception If the backend interface call fails, for example due to
      errors in creating function instances or insufficient resources, an
      exception will be thrown.

      @snippet{trimleft} function_example.cpp Function

      @par Chained Invocation Usage:
      In single-machine mode, the Invoke method can be chained to perform multiple function calls. However, if the total
      number of function execution instances exceeds the thread pool size, the program may become blocked. It is
      recommended to configure the thread pool size according to the number of expected function calls. For example, if
      a series of function calls involves three levels (e.g., AddThree, AddTwo, AddOne), each call will create a new
      thread, so the thread pool size should be configured to at least 3.
      @code
      int AddOne(int x)
      {
          return x + 1;
      }

      int AddTwo(int x)
      {
          auto obj = YR::Function(AddOne).Invoke(x);
          auto res = *YR::Get(obj);
          return res + 1;
      }

      int AddThree(int x)
      {
          auto obj = YR::Function(AddTwo).Invoke(x);
          auto res = *YR::Get(obj);
          return res + 1;
      }

      YR_INVOKE(AddOne, AddTwo, AddThree);

      int main(void) {
          YR::Config conf;
          conf.mode = YR::Config::Mode::LOCAL_MODE;
          conf.threadPoolSize = 3;
          YR::Init(conf);
          auto r = YR::Function(AddThree).Invoke(5);
          int ret = *(YR::Get(r)); // ret = 8
          return 0;
      }
      YR_INVOKE(HelloWorld)
      @endcode
      For more details on configuring the thread pool, refer to the struct-Config.
    */
    template <typename... Args>
    ObjectRef<ReturnType<F>> Invoke(Args &&...args)
    {
        CheckInitialized();
        if (internal::IsLocalMode() || opts.alwaysLocalMode) {
            auto resultPromise = std::make_shared<internal::Promise<ReturnType<F>>>();
            std::shared_future<std::shared_ptr<ReturnType<F>>> fu(resultPromise->get_future());
            auto obj = YR::internal::GetLocalModeRuntime()->PutFuture(std::move(fu));

            auto f = std::bind(internal::ExecuteFunction<F, std::remove_const_t<std::remove_reference_t<Args>>...>,
                               obj.ID(), std::move(resultPromise), func, std::make_tuple(std::forward<Args>(args)...));
            YR::internal::GetLocalModeRuntime()->LocalSubmit(std::move(f));
            return obj;
        }
        YR::internal::ArgumentsCheckWrapper<F, Args...>();
        std::vector<YR::internal::InvokeArg> invokeArgs;
        YR::internal::PackInvokeArgs(funcMeta.language, invokeArgs, std::forward<Args>(args)...);
        std::string objId = YR::internal::GetRuntime()->InvokeByName(funcMeta, invokeArgs, opts);
        //  objId just generated in InvokeByName, haven't Incre yet. Its first Incre time should in remote.
        return ObjectRef<ReturnType<F>>(objId, false);
    }
    /*!
       @brief Set the URN for the current function invocation, which can be used in conjunction with CppFunction or
       JavaFunction.

       @tparam F the function type.
       @param urn the function URN, which should match the tenant ID configured in the config.
       @returns FunctionHandler<F>&: a reference to the function handler object, facilitating direct invocation of the
       Invoke interface.

       @code
       int main(void) {
           YR::Config conf;
           YR::Init(conf);

           auto r1 =
           YR::CppFunction<int>("PlusOne").
           SetUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").Invoke(2);
           YR::Get(r1);
           return 0;
       }
       @endcode
    */
    FunctionHandler<F> &SetUrn(const std::string &urn)
    {
        this->funcMeta.funcUrn = urn;
        return *this;
    }
    /*!
      @brief Set the options for the current invocation, including resources, etc. This interface is not effective in
      local mode.

      @tparam F the function type.
      @param optsInput the invocation options. For detailed descriptions, refer to struct InvokeOptions.
      @returns FunctionHandler<F>&, a reference to the function handler object, facilitating direct invocation of the
      Invoke interface.
    */
    FunctionHandler<F> &Options(InvokeOptions &&optsInput)
    {
        this->opts = std::move(optsInput);
        this->opts.CheckOptionsValid();
        return *this;
    }
    /*!
      @brief Set the options for the current invocation, including resources, etc. This interface is not effective in
      local mode.

      @tparam F the function type.
      @param optsInput the invocation options. For detailed descriptions, refer to struct InvokeOptions.
      @returns FunctionHandler<F>&, a reference to the function handler object, facilitating direct invocation of the
      Invoke interface.

      1. Stateless function retry
      @code
      // static functions
      int AddOne(int x)
      {
          return x + 1;
      }

      void ThrowRuntimeError()
      {
          throw std::runtime_error("123");
      }

      YR_INVOKE(AddOne, ThrowRuntimeError);

      bool RetryChecker(const YR::Exception &e) noexcept
      {
          if (e.Code() == YR::ErrorCode::ERR_USER_FUNCTION_EXCEPTION) {
               std::string msg = e.what();
               if (msg.find("123") != std::string::npos) {
                   return true;
               }
          }
          return false;
      }
      int main(void) {
          YR::Config conf;
          YR::Init(conf);
          YR::InvokeOptions opts;

          opts.cpu = 1000;
          opts.memory = 1000;
          auto r1 = YR::Function(AddOne).Options(opts).Invoke(5);
          YR::Get(r1);

          opts.retryTimes = 1;
          opts.retryChecker = RetryChecker;
          auto r2 = YR::Function(ThrowRuntimeError).Options(opts).Invoke();
          YR::Get(r2);

          return 0;
      }
      @endcode
      2. Stateless functions specify local execution
      @code
      int CallLocal(int x)
      {
          YR::InvokeOptions opt;
          opt.alwaysLocalMode = true;
          auto obj = YR::Function(AddOne).Options(opt).Invoke(x);
          int ret = *YR::Get(obj);
          std::cout << "CallLocal result: " << ret << std::endl;
          return ret;
      }

      int CallCluster(int x)
      {
          auto obj = YR::Function(AddOne).Invoke(x);
          int ret = *YR::Get(obj);
          std::cout << "CallCluster result: " << ret << std::endl;
          return ret;
      }

      YR_INVOKE(CallLocal, CallCluster)

      void HybridClusterCallLocal()
      {
          auto obj = YR::Function(CallLocal).Invoke(1);
          std::cout << *YR::Get(obj) << " == " << 2 << std::endl;
      }

      void HybridLocalCallCluster()
      {
          YR::InvokeOptions opt;
          opt.alwaysLocalMode = true;
          auto obj = YR::Function(CallCluster).Options(opt).Invoke(1);
          std::cout << *YR::Get(obj) << " == " << 2 << std::endl;
      }

      int main()
      {
          HybridClusterCallLocal();
          HybridLocalCallCluster();
          return 0;
      }
      @endcode
    */
    FunctionHandler<F> &Options(const InvokeOptions &optsInput)
    {
        this->opts = optsInput;
        this->opts.CheckOptionsValid();
        return *this;
    }

private:
    F func = nullptr;
    internal::FuncMeta funcMeta;
    InvokeOptions opts;
};

}  // namespace YR
