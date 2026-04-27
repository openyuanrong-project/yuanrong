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
#include <functional>
#include <type_traits>
#include <unordered_map>

#include <boost/callable_traits.hpp>
#include <boost/optional.hpp>
#include <msgpack.hpp>

#include "yr/api/exception.h"
#include "yr/api/macro.h"
#include "yr/api/object_ref.h"
#include "yr/api/runtime.h"
#include "yr/api/runtime_manager.h"
#include "yr/api/serdes.h"

extern thread_local std::unordered_set<std::string> localNestedObjList;

namespace YR {
namespace internal {

using WrappedRetType = std::pair<std::shared_ptr<msgpack::sbuffer>, bool>;
using WrapperFunction = std::function<WrappedRetType(const std::string &, const std::vector<msgpack::sbuffer> &)>;
using WrapperMemberFunction =
    std::function<WrappedRetType(const std::string &, const msgpack::sbuffer &, const std::vector<msgpack::sbuffer> &)>;
using WrapperShutdownCaller = std::function<void(const msgpack::sbuffer &, uint64_t gracePeriodSecond)>;
using CheckpointFunction = std::function<msgpack::sbuffer(const msgpack::sbuffer &)>;
using RecoverFunction = std::function<msgpack::sbuffer(const msgpack::sbuffer &)>;
using RecoverCallbackFunction = std::function<void(const msgpack::sbuffer &)>;

using boost::callable_traits::args_t;
using boost::callable_traits::class_of_t;
using boost::callable_traits::return_type_t;

template <typename Func, typename T = void>
using enable_if_member_func_t = std::enable_if_t<std::is_member_function_pointer<Func>::value, T>;

template <typename Func, typename T = void>
using enable_if_not_member_func_t = std::enable_if_t<!std::is_member_function_pointer<Func>::value, T>;

using namespace std::placeholders;

template <typename>
struct NonConstRefType {
};

template <typename... Ts>
struct NonConstRefType<std::tuple<Ts...>> {
    using type = std::tuple<std::remove_const_t<std::remove_reference_t<Ts>>...>;
};

template <typename Tuple>
using RemoveConstRef = typename NonConstRefType<Tuple>::type;

template <typename>
struct RemoveFirstOfTuple {
};

template <typename T, typename... Ts>
struct RemoveFirstOfTuple<std::tuple<T, Ts...>> {
    using type = std::tuple<Ts...>;
};

template <typename Tuple>
using RemoveFirstElement = typename RemoveFirstOfTuple<Tuple>::type;

static msgpack::sbuffer YR_UNUSED VoidReturn()
{
    return Serialize(msgpack::type::nil_t());
}

template <typename Arg>
static Arg ParseArgValue(const msgpack::sbuffer &arg)
{
    localNestedObjList.clear();
    Arg retArg = Deserialize<Arg>(arg);
    return retArg;
}

template <typename... Args, size_t... I>
static void ParseArgsValue(std::tuple<Args...> &argsTuple, const std::vector<msgpack::sbuffer> &args,
                           std::index_sequence<I...>)
{
    std::initializer_list<int>{(std::get<I>(argsTuple) = ParseArgValue<Args>(args[I]), 0)...};
}

template <typename T>
static std::enable_if_t<!std::is_pointer<T>::value, msgpack::sbuffer> NonVoidReturn(const T &value)
{
    return Serialize(std::move(value));
}

template <typename T>
static std::enable_if_t<std::is_pointer<T>::value, msgpack::sbuffer> NonVoidReturn(const T &value)
{
    return Serialize((uint64_t)value);
}

template <typename R, typename Func, typename... Args, size_t... I>
R CallInternal(Func func, std::tuple<Args...> args, std::index_sequence<I...>)
{
    using ArgsTuple = args_t<Func>;
    return func(((typename std::tuple_element<I, ArgsTuple>::type)std::get<I>(args))...);
}

inline bool processRetNestedObj(std::shared_ptr<msgpack::sbuffer> retBufPtr, const std::string &retValObjId,
                                std::unordered_set<std::string> &nestedObjList)
{
    // Wait until them all ready
    std::vector<std::string> nestedObjListVec;
    nestedObjListVec.reserve(localNestedObjList.size());
    for (const std::string &e : localNestedObjList) {
        nestedObjListVec.push_back(e);
    }
    internal::GetRuntime()->Wait(nestedObjListVec, nestedObjListVec.size(), NO_TIMEOUT);
    internal::GetRuntime()->IncreGlobalReference({retValObjId});
    internal::GetRuntime()->Put(retValObjId, retBufPtr, localNestedObjList);  // Put to datasystem
    return true;
}

// return pair<return object SBufferPtr, putDone>
template <typename R, typename Func, typename... Args>
static std::enable_if_t<std::is_void<R>::value, WrappedRetType> Call(Func func, const std::string &retValObjId,
                                                                     std::tuple<Args...> args)
{
    CallInternal<R>(func, args, std::index_sequence_for<Args...>{});
    localNestedObjList.clear();
    auto bufPtr = std::make_shared<msgpack::sbuffer>(std::move(VoidReturn()));
    return std::make_pair(bufPtr, false);
}

// return pair<return object SBufferPtr, putDone>
template <typename R, typename Func, typename... Args>
static std::enable_if_t<!std::is_void<R>::value, WrappedRetType> Call(Func func, const std::string &retValObjId,
                                                                      std::tuple<Args...> args)
{
    bool putDone = false;
    R ret = CallInternal<R>(func, args, std::index_sequence_for<Args...>{});
    localNestedObjList.clear();
    auto bufPtr = std::make_shared<msgpack::sbuffer>(std::move(NonVoidReturn(ret)));
    if (!retValObjId.empty() && !localNestedObjList.empty()) {
        putDone = processRetNestedObj(bufPtr, retValObjId, localNestedObjList);
    }
    return std::make_pair(bufPtr, putDone);
}

template <typename R, typename Func, typename Cls, typename... Args, size_t... I>
R MemberCallInternal(Func func, Cls &cls, std::tuple<Args...> args, std::index_sequence<I...>)
{
    using AllArgsTuple = args_t<Func>;
    (void)args;
    return (cls.*func)(((typename std::tuple_element<I + 1, AllArgsTuple>::type) std::get<I>(args))...);
}

// return pair<return object SBufferPtr, putDone>
template <typename R, typename Func, typename Cls, typename... Args>
static std::enable_if_t<std::is_void<R>::value, WrappedRetType> MemberCall(Func func, const std::string &retValObjId,
                                                                           Cls &cls, std::tuple<Args...> args)
{
    MemberCallInternal<R>(func, cls, args, std::index_sequence_for<Args...>{});
    localNestedObjList.clear();
    auto bufPtr = std::make_shared<msgpack::sbuffer>(std::move(VoidReturn()));
    return std::make_pair(bufPtr, false);
}

// return pair<return object SBufferPtr, putDone>
template <typename R, typename Func, typename Cls, typename... Args>
static std::enable_if_t<!std::is_void<R>::value, WrappedRetType> MemberCall(Func func, const std::string &retValObjId,
                                                                            Cls &cls, std::tuple<Args...> args)
{
    bool putDone = false;
    R ret = MemberCallInternal<R>(func, cls, args, std::index_sequence_for<Args...>{});
    localNestedObjList.clear();
    auto bufPtr = std::make_shared<msgpack::sbuffer>(std::move(NonVoidReturn(ret)));
    if (!localNestedObjList.empty()) {
        putDone = processRetNestedObj(bufPtr, retValObjId, localNestedObjList);
    }
    return std::make_pair(bufPtr, putDone);
}

template <class T>
T &ParseClassRef(const msgpack::sbuffer &cls)
{
    uint64_t clsHandler = Deserialize<uint64_t>(cls);
    return *(reinterpret_cast<T *>(clsHandler));
}

template <typename Func>
static std::pair<std::shared_ptr<msgpack::sbuffer>, bool> FunctionCaller(Func func, const std::string &retValObjId,
                                                                         const std::vector<msgpack::sbuffer> &args)
{
    using ArgsTuple = RemoveConstRef<args_t<Func>>;
    using RetType = return_type_t<Func>;

    ArgsTuple argsValue;
    ParseArgsValue(argsValue, args, std::make_index_sequence<std::tuple_size<ArgsTuple>::value>{});
    return Call<RetType>(func, retValObjId, std::move(argsValue));
}

template <typename Func>
static std::pair<std::shared_ptr<msgpack::sbuffer>, bool> MemberFunctionCaller(
    Func func, const std::string &retValObjId, const msgpack::sbuffer &cls, const std::vector<msgpack::sbuffer> &args)
{
    using ArgsTuple = RemoveFirstElement<RemoveConstRef<args_t<Func>>>;
    using RetType = return_type_t<Func>;
    using ClassType = class_of_t<Func>;

    ArgsTuple argsValue;
    ParseArgsValue(argsValue, args, std::make_index_sequence<std::tuple_size<ArgsTuple>::value>{});
    ClassType &clsRef = ParseClassRef<ClassType>(cls);
    return MemberCall<RetType>(func, retValObjId, clsRef, std::move(argsValue));
}

template <typename Func>
static void RecoverCallback(Func func, const msgpack::sbuffer &cls)
{
    using ClassType = class_of_t<Func>;

    ClassType &clsRef = ParseClassRef<ClassType>(cls);
    MemberCall<void>(func, "", clsRef, {});
}

template <typename ClassType>
static msgpack::sbuffer Checkpoint(const msgpack::sbuffer &cls)
{
    ClassType &clsRef = ParseClassRef<ClassType>(cls);
    return Serialize(clsRef);
}

template <typename Func>
static void ShutdownCaller(Func func, const msgpack::sbuffer &cls, uint64_t gracePeriodSecond)
{
    using ClassType = class_of_t<Func>;

    ClassType &clsRef = ParseClassRef<ClassType>(cls);
    std::tuple<uint64_t> arg(gracePeriodSecond);
    MemberCall<void>(func, "", clsRef, arg);
}

template <typename ClassType>
static msgpack::sbuffer Recover(const msgpack::sbuffer &data)
{
    void *ptr = operator new(sizeof(ClassType));
    if (ptr == nullptr) {
        return msgpack::sbuffer();
    }
    auto clsPtr = new (ptr) ClassType();
    *clsPtr = Deserialize<ClassType>(data);
    return Serialize((uint64_t)clsPtr);
}

template <typename Func>
uint64_t GetUniqueFuncId(Func func)
{
    union {
        Func ptr;
        uint64_t id;
    } uniqueId;
    uniqueId.ptr = func;
    return uniqueId.id;
}

class FunctionManager {
public:
    static FunctionManager &Singleton();

    template <typename Func>
    enable_if_not_member_func_t<Func, std::string> GetFunctionName(const Func &func)
    {
        auto funcId = GetUniqueFuncId(func);
        auto iter = funcIdToName_.find(funcId);
        if (iter == funcIdToName_.end()) {
            return "";
        }
        return iter->second;
    }

    template <typename Func>
    enable_if_member_func_t<Func, std::string> GetFunctionName(const Func &func)
    {
        auto funcId = GetUniqueFuncId(func);
        auto iter = memberFuncIdToName_.find(funcId);
        if (iter == memberFuncIdToName_.end()) {
            return "";
        }
        return iter->second;
    }

    boost::optional<WrapperFunction &> GetNormalFunction(const std::string &funcName)
    {
        auto iter = funcMap_.find(funcName);
        if (iter == funcMap_.end()) {
            return boost::none;
        }
        return iter->second;
    }

    boost::optional<WrapperMemberFunction &> GetInstanceFunction(const std::string &funcName)
    {
        auto iter = memberFuncMap_.find(funcName);
        if (iter == memberFuncMap_.end()) {
            return boost::none;
        }
        return iter->second;
    }

    template <typename Func>
    enable_if_not_member_func_t<Func> RegisterInvokeFunction(const std::string &name, Func func)
    {
        auto funcId = GetUniqueFuncId(func);
        if (!funcIdToName_.emplace(funcId, name).second) {
            throw YR::Exception::RegisterFunctionException(name);
        }

        auto happened = funcMap_.emplace(name, std::bind(FunctionCaller<Func>, func, _1, _2)).second;
        if (!happened) {
            throw YR::Exception::RegisterFunctionException(name);
        }
    }

    template <typename Func>
    enable_if_member_func_t<Func> RegisterInvokeFunction(const std::string &name, Func func)
    {
        using ClassType = class_of_t<Func>;
        auto cls = typeid(ClassType).name();
        std::string clsName = GetClsName(cls, name);

        auto funcId = GetUniqueFuncId(func);
        if (!memberFuncIdToName_.emplace(funcId, name).second) {
            throw YR::Exception::RegisterFunctionException(name);
        }

        auto happened = memberFuncMap_.emplace(name, std::bind(MemberFunctionCaller<Func>, func, _1, _2, _3)).second;
        if (!happened) {
            throw YR::Exception::RegisterFunctionException(name);
        }
        if (ckptFuncMap_.find(clsName) == ckptFuncMap_.end()) {
            ckptFuncMap_.emplace(clsName, std::bind(Checkpoint<ClassType>, _1));
        }
        if (recoverFuncMap_.find(clsName) == recoverFuncMap_.end()) {
            recoverFuncMap_.emplace(clsName, std::bind(Recover<ClassType>, _1));
        }
    }

    template <typename Func>
    enable_if_member_func_t<Func> RegisterRecoverFunction(const std::string &name, Func func)
    {
        using ClassType = class_of_t<Func>;
        auto cls = typeid(ClassType).name();
        std::string clsName = GetClsName(cls, name);
        if (recoverCallbackFuncMap_.find(clsName) == recoverCallbackFuncMap_.end()) {
            recoverCallbackFuncMap_.emplace(clsName, std::bind(RecoverCallback<Func>, func, _1));
        }
    }

    template <typename Func>
    enable_if_not_member_func_t<Func> RegisterRecoverFunction(const std::string &name, Func func)
    {
        throw YR::Exception::RegisterRecoverFunctionException();
    }

    boost::optional<CheckpointFunction &> GetCheckpointFunction(const std::string &className)
    {
        auto iter = ckptFuncMap_.find(className);
        if (iter == ckptFuncMap_.end()) {
            return boost::none;
        }
        return iter->second;
    }

    template <typename Func>
    enable_if_member_func_t<Func> RegisterShutdownFunctions(const std::string &name, Func func)
    {
        using ClassType = class_of_t<Func>;
        auto cls = typeid(ClassType).name();
        std::string clsName = GetClsName(cls, name);
        if (shutdownCallerMap_.find(clsName) == shutdownCallerMap_.end()) {
            shutdownCallerMap_.emplace(clsName, std::bind(ShutdownCaller<Func>, func, _1, _2));
        }
    }

    template <typename Func>
    enable_if_not_member_func_t<Func> RegisterShutdownFunctions(const std::string &name, Func func)
    {
        throw YR::Exception::RegisterShutdownFunctionException();
    }

    boost::optional<WrapperShutdownCaller &> GetShutdownFunction(const std::string &className)
    {
        auto iter = shutdownCallerMap_.find(className);
        if (iter == shutdownCallerMap_.end()) {
            return boost::none;
        }
        return iter->second;
    }

    boost::optional<RecoverFunction &> GetRecoverFunction(const std::string &className)
    {
        auto iter = recoverFuncMap_.find(className);
        if (iter == recoverFuncMap_.end()) {
            return boost::none;
        }
        return iter->second;
    }

    boost::optional<RecoverCallbackFunction &> GetRecoverCallbackFunction(const std::string &className)
    {
        auto iter = recoverCallbackFuncMap_.find(className);
        if (iter == recoverCallbackFuncMap_.end()) {
            return boost::none;
        }
        return iter->second;
    }

    std::string GetClassName(const std::string &classID)
    {
        auto iter = clsMap_.find(classID);
        if (iter == clsMap_.end()) {
            return "";
        }
        return iter->second;
    }

private:
    std::string GetClsName(const std::string &cls, const std::string &name)
    {
        if (clsMap_.find(cls) == clsMap_.end()) {
            std::string sep("::");
            auto clsIndex = name.find_last_of(sep);
            clsIndex -= sep.length();
            if (name[0] == '&') {
                clsMap_.emplace(cls, name.substr(1, clsIndex));
            } else {
                clsMap_.emplace(cls, name.substr(0, clsIndex));
            }
        }
        return clsMap_[cls];
    }

    std::unordered_map<uint64_t, std::string> funcIdToName_;
    std::unordered_map<uint64_t, std::string> memberFuncIdToName_;
    std::unordered_map<std::string, std::string> clsMap_;
    std::unordered_map<std::string, WrapperFunction> funcMap_;
    std::unordered_map<std::string, WrapperMemberFunction> memberFuncMap_;
    std::unordered_map<std::string, WrapperShutdownCaller> shutdownCallerMap_;
    std::unordered_map<std::string, CheckpointFunction> ckptFuncMap_;
    std::unordered_map<std::string, RecoverFunction> recoverFuncMap_;
    std::unordered_map<std::string, RecoverCallbackFunction> recoverCallbackFuncMap_;
};
}  // namespace internal
}  // namespace YR