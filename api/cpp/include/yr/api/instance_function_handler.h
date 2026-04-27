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
#include <type_traits>

#include <boost/callable_traits.hpp>

#include "yr/api/args_check.h"
#include "yr/api/function_manager.h"
#include "yr/api/object_ref.h"
#include "yr/api/runtime.h"

namespace YR {

namespace internal {
template <typename F, typename T, typename... Args, size_t... I>
static std::enable_if_t<std::is_void<ReturnType<F>>::value, std::shared_ptr<int>> LocalIntanceFuncCall(
    F func, T *instance, std::tuple<Args...> &&args, std::index_sequence<I...>)
{
    using ArgsTuple = boost::callable_traits::args_t<F>;
    (instance->*func)(((typename std::tuple_element<I + 1, ArgsTuple>::type) ParseArg(std::get<I>(args)))...);
    return std::make_shared<int>(0);
}

template <typename F, typename T, typename... Args, size_t... I>
static std::enable_if_t<!std::is_void<ReturnType<F>>::value, std::shared_ptr<ReturnType<F>>> LocalIntanceFuncCall(
    F func, T *instance, std::tuple<Args...> &&args, std::index_sequence<I...>)
{
    using ArgsTuple = boost::callable_traits::args_t<F>;
    auto res =
        (instance->*func)(((typename std::tuple_element<I + 1, ArgsTuple>::type) ParseArg(std::get<I>(args)))...);
    return std::make_shared<ReturnType<F>>(res);
}
}  // namespace internal
template <typename F, typename T>
class InstanceFunctionHandler {
public:
    InstanceFunctionHandler(F f, std::string id, bool isLocal)
        : func(f), instanceId(std::move(id)), alwaysLocalMode(isLocal)
    {
    }
    InstanceFunctionHandler(const internal::FuncMeta &funcMeta, const std::string &id, Runtime *runtime)
        : funcMeta(funcMeta), instanceId(id), yrRuntime(runtime)
    {
    }

    InstanceFunctionHandler(const InstanceFunctionHandler &other)
    {
        this->func = other.func;
        this->funcMeta = other.funcMeta;
        this->instanceId = other.instanceId;
        this->yrRuntime = other.yrRuntime;
        this->opts = other.opts;
    }
    InstanceFunctionHandler &operator=(const InstanceFunctionHandler &other)
    {
        if (&other != this) {
            this->func = other.func;
            this->funcMeta = other.funcMeta;
            this->instanceId = other.instanceId;
            this->yrRuntime = other.yrRuntime;
            this->opts = other.opts;
        }
    }
    ~InstanceFunctionHandler() = default;

    /*!
       @brief Invokes a remote function by sending the request to a remote backend for execution.

       @tparam Args The types of the arguments passed to the function.

       @param args The arguments to be passed to the remote function. The types and number of arguments must match the
       function's definition exactly. Ensure that the argument types match the expected types precisely
       to avoid issues caused by implicit type conversions.

       @return ObjectRef<ReturnType<F>>, A reference to the result object, which is essentially a
       key. To retrieve the actual value, use the `YR::Get` method.

       @throws Exception If the invocation fails, such as due to the function instance exiting
       abnormally or the user code executing abnormally.

       @details This function sends a request to a remote backend to execute the specified function. The function's
       arguments are serialized and transmitted to the backend, and the result is returned as an `ObjectRef`.

       The function supports dependency resolution based on the argument types:
       1. If the function's parameter is of type `ArgType` and the argument passed is of type `ObjectRef<ArgType>`,
       the request will only be sent after the corresponding computation for `obj` is completed, and the `obj` must be
       from the same client.
       2. If the function's parameter is of type `vector<ObjectRef<ArgType>> objs` and the argument passed is of type
       `vector<ObjectRef<ArgType>>`, the request will only be sent after all computations corresponding to the `objs`
       are completed, and all `ObjectRef` instances must be from the same client.
       3. Arguments of other types do not support dependency resolution.

       @code{.cpp}
       int main(void) {
           YR::Config conf;
           YR::Init(conf);

           auto ins = YR::Instance(SimpleCaculator::Constructor).Invoke();
           auto r3 = ins.Function(&SimpleCaculator::Plus).Invoke(1, 1);
           int res = *(YR::Get(r3));

           return 0;
       }
       @endcode
       In this example, the `Invoke` method is used to call the `Plus` function of the `SimpleCaculator` instance with
       arguments `1` and `1`. The result is retrieved using `YR::Get(r3)`.
    */
    template <typename... Args>
    ObjectRef<ReturnType<F>> Invoke(Args &&...args)
    {
        CheckInitialized();
        if (internal::IsLocalMode() || alwaysLocalMode) {
            auto resultPromise = std::make_shared<internal::Promise<ReturnType<F>>>();
            std::shared_future<std::shared_ptr<ReturnType<F>>> fu(resultPromise->get_future());
            auto res = internal::GetLocalModeRuntime()->PutFuture(std::move(fu));
            auto f = std::bind(
                InstanceFunctionHandler::LocalExecutionWrapper<F,
                                                               std::remove_const_t<std::remove_reference_t<Args>>...>,
                res.ID(), func, instanceId, resultPromise, std::make_tuple(std::forward<Args>(args)...));
            internal::GetLocalModeRuntime()->LocalSubmit(std::move(f));
            return res;
        }
        YR::internal::ArgumentsCheckWrapper<F, Args...>();
        std::vector<YR::internal::InvokeArg> invokeArgs;
        PackInvokeArgs(funcMeta.language, invokeArgs, std::forward<Args>(args)...);
        std::string objId = yrRuntime->InvokeInstance(funcMeta, instanceId, invokeArgs, this->opts);
        return ObjectRef<ReturnType<F>>(objId, false);
    }

    /*!
       @brief Sets options for the function invocation, such as timeout and retry count.

       @tparam Args The types of the arguments passed to the function.

       @param optsInput The invocation options, including timeout, retry count, etc. For detailed descriptions, refer to
       the `struct InstanceInvokeOptions`.

       @returns InstanceFunctionHandler<F, T> &, Returns a reference to the `InstanceFunctionHandler` object,
       facilitating direct invocation of the `Invoke` interface.

       @details This method allows you to configure specific options for the function invocation, such as setting a
       timeout or specifying the number of retries. These options are crucial for controlling the behavior of remote
       function calls. However, please note that this method does not take effect in local mode.

       @code{.cpp}
       int main(void) {
           YR::Config conf;
           YR::Init(conf);

           YR::InvokeOptions opts;
           opts.retryTimes = 5;
           auto ins = YR::Instance(SimpleCaculator::Constructor).Invoke();
           auto r3 = ins.Function(&SimpleCaculator::Plus).Options(opts).Invoke(1, 1);

           return 0;
       }
       @endcode

       @warning This method is designed for distributed environments and does not function in local mode. Ensure that
       you are using it in the appropriate context where invocation options are applicable.
    */
    InstanceFunctionHandler<F, T> &Options(InvokeOptions &&optsInput)
    {
        this->opts = std::move(optsInput);
        this->opts.CheckOptionsValid();
        if (opts.retryChecker) {
            throw YR::Exception::InvalidParamException("retry checker is not yet supported for stateful functions");
        }
        return *this;
    }

    InstanceFunctionHandler<F, T> &Options(const InvokeOptions &optsInput)
    {
        this->opts = optsInput;
        this->opts.CheckOptionsValid();
        if (opts.retryChecker) {
            throw YR::Exception::InvalidParamException("retry checker is not yet supported for stateful functions");
        }
        return *this;
    }

private:
    template <typename U = F, typename... Args>
    static std::enable_if_t<internal::IsCrossLang<U>::value> LocalExecutionWrapper(
        const std::string &id, F func, const std::string &instanceId,
        std::shared_ptr<internal::Promise<ReturnType<F>>> promise, std::tuple<Args...> args)
    {
    }

    template <typename U = F, typename... Args>
    static std::enable_if_t<!internal::IsCrossLang<U>::value> LocalExecutionWrapper(
        const std::string &id, F func, const std::string &instanceId,
        std::shared_ptr<internal::Promise<ReturnType<F>>> promise, std::tuple<Args...> args)
    {
        try {
            T *funcInstance = (T *)internal::LocalInstanceManager<T>::Singleton().GetLocalInstance(instanceId);
            if (funcInstance == nullptr) {
                std::string msg = "instance is nullptr, instanceId: " + instanceId;
                throw YR::Exception::InstanceIdEmptyException(msg);
            }
            auto ret =
                internal::LocalIntanceFuncCall(func, funcInstance, std::move(args), std::index_sequence_for<Args...>{});
            promise->set_value(ret);
            internal::GetLocalModeRuntime()->SetReady(id);
        } catch (YR::Exception &e) {
            auto exceptionPtr = std::make_exception_ptr(e);
            promise->set_exception(exceptionPtr);
            internal::GetLocalModeRuntime()->SetException(id, exceptionPtr);
        } catch (std::exception &e) {
            std::stringstream ss;
            ss << "exception happens when executing user's function: " << e.what();
            auto exceptionPtr = std::make_exception_ptr(Exception::UserCodeException(ss.str()));
            promise->set_exception(exceptionPtr);
            internal::GetLocalModeRuntime()->SetException(id, exceptionPtr);
        } catch (...) {
            auto exceptionPtr =
                std::make_exception_ptr(Exception::UserCodeException("non-standard exception is thrown"));
            promise->set_exception(exceptionPtr);
            internal::GetLocalModeRuntime()->SetException(id, exceptionPtr);
        }
    }

    F func;
    internal::FuncMeta funcMeta;
    std::string instanceId;
    Runtime *yrRuntime;
    InvokeOptions opts;
    bool alwaysLocalMode = false;
};
}  // namespace YR
