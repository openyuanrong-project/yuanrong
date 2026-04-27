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
#include <initializer_list>

#include <boost/callable_traits.hpp>

#include "yr/api/args_check.h"
#include "yr/api/constant.h"
#include "yr/api/cross_lang.h"
#include "yr/api/function_manager.h"
#include "yr/api/invoke_arg.h"
#include "yr/api/invoke_options.h"
#include "yr/api/local_instance_manager.h"
#include "yr/api/named_instance.h"
#include "yr/api/runtime.h"
#include "yr/api/serdes.h"

namespace YR {
namespace internal {
template <typename Func, typename... Args, size_t... I>
static std::enable_if_t<std::is_void<ReturnType<Func>>::value, std::shared_ptr<int>> LocalCreatorCall(
    Func func, std::tuple<Args...> &&args, std::index_sequence<I...>)
{
    using ArgsTuple = boost::callable_traits::args_t<Func>;
    func(((typename std::tuple_element<I, ArgsTuple>::type)ParseArg(std::get<I>(args)))...);
    return std::make_shared<int>(0);
}

template <typename Func, typename... Args, size_t... I>
static std::enable_if_t<!std::is_void<ReturnType<Func>>::value, std::shared_ptr<GetClassOfWorker<Func>>>
LocalCreatorCall(Func func, std::tuple<Args...> &&args, std::index_sequence<I...>)
{
    using ArgsTuple = boost::callable_traits::args_t<Func>;
    auto res = func(((typename std::tuple_element<I, ArgsTuple>::type)ParseArg(std::get<I>(args)))...);
    return std::shared_ptr<GetClassOfWorker<Func>>(res);
}

template <typename F, typename... Args>
static std::enable_if_t<IsCrossLang<F>::value> ExecuteCreatorFunction(const std::string &id,
                                                                      const std::string &instanceId,
                                                                      std::shared_ptr<Promise<GetClassOfWorker<F>>> p,
                                                                      F func, std::tuple<Args...> args)
{
}

template <typename F, typename... Args>
static std::enable_if_t<!IsCrossLang<F>::value> ExecuteCreatorFunction(const std::string &id,
                                                                       const std::string &instanceId,
                                                                       std::shared_ptr<Promise<GetClassOfWorker<F>>> p,
                                                                       F func, std::tuple<Args...> args)
{
    try {
        auto res = LocalCreatorCall(func, std::move(args), std::index_sequence_for<Args...>{});
        p->set_value(res);
        internal::GetLocalModeRuntime()->SetReady(id);
    } catch (YR::Exception &e) {
        auto exceptionPtr = std::make_exception_ptr(e);
        p->set_exception(exceptionPtr);
        internal::GetLocalModeRuntime()->SetException(id, exceptionPtr);
    } catch (std::exception &e) {
        std::stringstream ss;
        ss << "exception happens when executing user's function: " << e.what();
        auto exceptionPtr = std::make_exception_ptr(Exception::UserCodeException(ss.str()));
        p->set_exception(exceptionPtr);
        internal::GetLocalModeRuntime()->SetException(id, exceptionPtr);
    } catch (...) {
        auto exceptionPtr = std::make_exception_ptr(Exception::UserCodeException("non-standard exception is thrown"));
        p->set_exception(exceptionPtr);
        internal::GetLocalModeRuntime()->SetException(id, exceptionPtr);
    }
}

static inline bool InstanceRangeEnabled(InstanceRange instanceRange)
{
    if (instanceRange.min == DEFAULT_INSTANCE_RANGE_NUM && instanceRange.max == DEFAULT_INSTANCE_RANGE_NUM) {
        return false;
    }
    return true;
}

template <typename Creator>
class InstanceCreator {
public:
    explicit InstanceCreator(Creator constructor) : creator(constructor) {}
    InstanceCreator(const internal::FuncMeta &funcMeta, Runtime *runtime, Creator constructor)
        : creator(constructor), funcMeta(funcMeta), yrRuntime(runtime)
    {
    }

    InstanceCreator(const InstanceCreator &other)
        : funcMeta(other.funcMeta), opts(other.opts), yrRuntime(other.yrRuntime), creator(other.creator)
    {
    }

    InstanceCreator &operator=(const InstanceCreator &other)
    {
        if (this != &other) {
            this->creator = other.creator;
            this->funcMeta = other.funcMeta;
            this->opts = other.opts;
            this->yrRuntime = other.yrRuntime;
        }
        return *this;
    }

    ~InstanceCreator() = default;
    /*!
      @brief Execute the creation of an instance and construct an object of the class.

      @tparam F The type of the worker function used to construct the instance.
      @tparam Args The types of the arguments required by the worker's constructor.

      @param args The arguments required by the worker's constructor. The types and number of arguments must match the
      constructor's definition exactly. Ensure that the arguments passed match the expected types precisely to avoid
      issues caused by implicit type conversions.

      @returns NamedInstance<GetClassOfInstance<F>>, A named instance that can be used to call member functions of the
      class using the `Function` method, suitable for object-oriented programming scenarios.

      @note This method constructs an instance of the class and ensures that the instance can only execute member
      function requests of the class. The returned named instance allows you to call member functions of the class.

      @code{.cpp}
      // Example: Create an instance and call a member function
      int main(void) {
          YR::Config conf;
          YR::Init(conf);

          auto ins = YR::Instance(SimpleCaculator::Constructor).Invoke(); // Create an instance
          auto r3 = ins.Function(&SimpleCaculator::Plus).Invoke(1, 1);    // Call a member function
          int res = *(YR::Get(r3));                                     // Get the result
          return 0;
     }
     @endcode

     @warning Ensure that the provided arguments match the constructor's expected types and number exactly to avoid
     compilation errors. The instance created by this method is designed to work with member functions of the class
     only.
    */
    template <typename... Args>
    NamedInstance<GetClassOfWorker<Creator>> Invoke(Args &&...args)
    {
        CheckInitialized();
        if (IsLocalMode() || opts.alwaysLocalMode) {
            auto instanceId = GetLocalModeRuntime()->GenerateObjId();
            NamedInstance<GetClassOfWorker<Creator>> handler(instanceId);
            handler.SetAlwaysLocalMode(true);
            auto resultPromise = std::make_shared<Promise<GetClassOfWorker<Creator>>>();
            std::shared_future<std::shared_ptr<GetClassOfWorker<Creator>>> fu(resultPromise->get_future());
            auto obj = GetLocalModeRuntime()->PutFuture(std::move(fu));
            auto f = std::bind(ExecuteCreatorFunction<Creator, std::remove_const_t<std::remove_reference_t<Args>>...>,
                               obj.ID(), instanceId, std::move(resultPromise), creator,
                               std::make_tuple(std::forward<Args>(args)...));
            GetLocalModeRuntime()->LocalSubmit(std::move(f));
            LocalInstanceManager<GetClassOfWorker<Creator>>::Singleton().SetResult(instanceId, std::move(obj));
            return handler;
        }
        ArgumentsCheckWrapper<Creator, Args...>();
        std::vector<InvokeArg> invokeArgs;
        PackInvokeArgs(funcMeta.language, invokeArgs, std::forward<Args>(args)...);
        auto it = opts.customExtensions.find(CONCURRENCY_KEY);
        if (it == opts.customExtensions.end() || it->second == "1") {
            opts.needOrder = true;
        }
        auto instanceId = yrRuntime->CreateInstance(funcMeta, invokeArgs, opts);
        NamedInstance<GetClassOfWorker<Creator>> handler(instanceId);
        handler.SetAlwaysLocalMode(false);
        handler.SetClassName(funcMeta.className);
        handler.SetFunctionUrn(funcMeta.funcUrn);
        handler.SetNeedOrder(opts.needOrder);
        handler.SetName(funcMeta.name);
        handler.SetNs(funcMeta.ns);
        if (InstanceRangeEnabled(this->opts.instanceRange)) {
            handler.SetGroupName(this->opts.groupName);
        }
        return handler;
    }

    /*!
     @brief Set the function URN for the instance creation.

     @tparam Creator The type of the creator used to construct the instance.

     @param urn The function URN to be set. This URN should be obtained after the function is deployed. The tenant ID in
     the URN must match the tenant ID configured in the `Config`. For details on tenant ID configuration, refer to the
     "About Tenant ID" section in the `Config` documentation.

     @returns InstanceCreator<Creator> & A reference to the `InstanceCreator` object, allowing for a fluent interface to
     call the `Invoke` method directly.

     @note This method is designed to be used in conjunction with `CppInstanceClass` or `JavaInstanceClass`. The URN
     format should follow the specified structure, and the tenant ID must be correctly configured to ensure proper
     functionality.

     @code{.cpp}
     // Example: Setting the URN for an instance
     int main(void) {
         YR::Config conf;
         YR::Init(conf);

         auto cppCls = YR::CppInstanceClass::FactoryCreate("Counter::FactoryCreate");
         auto cppIns = YR::Instance(cppCls)
                           .SetUrn("sn:cn:yrk:default:function:0-opc-opc:$latest")
                           .Invoke(1);
         auto obj = cppIns.CppFunction<int>("&Counter::Add").Invoke(1);

         return 0;
     }
     @endcode

     @warning Ensure that the tenant ID in the provided URN matches the tenant ID configured in the `Config`. Mismatched
     tenant IDs may lead to errors or unexpected behavior.
    */
    InstanceCreator<Creator> &SetUrn(const std::string &urn)
    {
        this->funcMeta.funcUrn = urn;
        return *this;
    }

    /*!
     @brief Set options for the instance creation, including resource usage settings.

     @tparam Creator The type of the worker function used to construct the instance.

     @param optsInput The options for the invocation, including resource usage and execution mode. For detailed
     descriptions, refer to the `struct InvokeOptions`.

     @returns InstanceCreator<F> & A reference to the `InstanceCreator` object, allowing for a fluent interface to call
     the `Invoke` method directly.

     @note This method does not take effect in local mode. It is designed for distributed scenarios where resource
     management and execution options are crucial.

     @code{.cpp}
     int main(void) {
        YR::Config conf;
        YR::Init(conf);

        YR::InvokeOptions opts;
        opts.cpu = 1000;
        opts.memory = 1000;
        auto ins = YR::Instance(SimpleCaculator::Constructor).Options(opts).Invoke();
        auto r3 = ins.Function(&SimpleCaculator::Plus).Invoke(1, 1);

        return 0;
     }
     @endcode
     Stateful functions specify local execution (nested stateless).
     @code{.cpp}
     // Example: Setting options for instance creation
     int incre(int x)
     {
          return x + 1;
     }

     YR_INVOKE(incre)

     struct Actor {
         int a;

         Actor() {}
         Actor(int init) { a = init; }
         static Actor *Create(int init) {
             return new Actor(init);
         }
         int InvokeIncre(bool local) {
             auto obj = YR::Function(incre).Options({ .alwaysLocalMode = local }).Invoke(a);
             return *YR::Get(obj);
         }
         MSGPACK_DEFINE(a)
     };

     YR_INVOKE(Actor::Create, &Actor::InvokeIncre)

     void Forward(bool local)
     {
         auto ins = YR::Instance(Actor::Create).Options({ .alwaysLocalMode = local }).Invoke(0);
         auto obj = ins.Function(&Actor::InvokeIncre).Invoke(true);
         std::cout << *YR::Get(obj) << std::endl;
         auto obj2 = ins.Function(&Actor::InvokeIncre).Invoke(false);
         std::cout << *YR::Get(obj2) << std::endl;
     }

     int main()
     {
         Forward(true);
         Forward(false);
         return 0;
     }
     @endcode

     @warning This method is intended for distributed environments and does not function in local mode. Ensure that you
     are using it in the appropriate context where resource management and execution options are applicable.
    */
    InstanceCreator<Creator> &Options(InvokeOptions &&optsInput)
    {
        this->opts = std::move(optsInput);
        this->opts.CheckOptionsValid();
        if (opts.retryChecker) {
            throw YR::Exception::InvalidParamException("retry checker is not yet supported for stateful functions");
        }
        if (InstanceRangeEnabled(this->opts.instanceRange)) {
            this->opts.groupName = yrRuntime->GenerateGroupName();
        }
        return *this;
    }

    /*!
     @brief Set options for the instance creation, including resource usage settings.

     @tparam Creator The type of the worker function used to construct the instance.

     @param optsInput The options for the invocation, including resource usage and execution mode. For detailed
     descriptions, refer to the `struct InvokeOptions`.

     @returns InstanceCreator<F> & A reference to the `InstanceCreator` object, allowing for a fluent interface to call
     the `Invoke` method directly.

     @note This method does not take effect in local mode. It is designed for distributed scenarios where resource
     management and execution options are crucial.

     @warning This method is intended for distributed environments and does not function in local mode. Ensure that you
     are using it in the appropriate context where resource management and execution options are applicable.

    */
    InstanceCreator<Creator> &Options(const InvokeOptions &optsInput)
    {
        this->opts = optsInput;
        this->opts.CheckOptionsValid();
        if (opts.retryChecker) {
            throw YR::Exception::InvalidParamException("retry checker is not yet supported for stateful functions");
        }
        if (InstanceRangeEnabled(this->opts.instanceRange)) {
            this->opts.groupName = yrRuntime->GenerateGroupName();
        }
        return *this;
    }

private:
    Creator creator = nullptr;
    internal::FuncMeta funcMeta;
    InvokeOptions opts;
    Runtime *yrRuntime;
};

template <typename Creator>
std::string GetClassName()
{
    std::string cls = typeid(GetClassOfWorker<Creator>).name();
    return FunctionManager::Singleton().GetClassName(cls);
}
}  // namespace internal
}  // namespace YR
