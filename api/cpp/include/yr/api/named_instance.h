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

#include <boost/callable_traits.hpp>
#include <msgpack.hpp>
#include "yr/api/instance_function_handler.h"
#include "yr/api/invoke_options.h"
#include "yr/api/local_instance_manager.h"
#include "yr/api/runtime_manager.h"
namespace YR {
#define THROW_WHEN_IS_RANGE(groupName)                                                                            \
    if (!groupName.empty()) {                                                                                     \
        std::string msg =                                                                                         \
            "unsupproted invoke type: range instance handler cannont be used to invoke directly, please execute " \
            "'GetInstances' first.";                                                                              \
        throw Exception::IncorrectInvokeUsageException(msg);                                                      \
    }

#define THROW_WHEN_IS_NOT_RANGE(groupName)                                                                        \
    if (groupName.empty()) {                                                                                      \
        std::string msg =                                                                                         \
            "unsupproted function type: this function can only be used for range instance handler, please check " \
            "whether range is enabled in InvokeOptions.";                                                         \
        throw Exception::IncorrectFunctionUsageException(msg);                                                    \
    }

const std::string INSTANCE_KEY = "instanceKey";
const std::string INSTANCE_ID = "instanceID";
const std::string CLASS_NAME = "className";
const std::string FUNCTION_URN = "functionUrn";
const std::string NEED_ORDER = "needOrder";
const std::string GROUP_NAME = "groupName";
const std::string GROUP_INS_IDS = "groupInsIds";
const std::string INSTANCE_ROUTE = "instanceRoute";

using FormattedMap = std::unordered_map<std::string, std::string>;

/*!
 * @class NamedInstance
 * @brief Named instance that can invoke the Function method of this class to construct member functions of the
 * instance's class, suitable for object-oriented programming scenarios.
 * @tparam InstanceType The type of the instance.
 */
template <typename InstanceType>
class NamedInstance {
public:
    NamedInstance() = default;

    NamedInstance(const std::string &instanceId) : instanceId(instanceId) {}

    ~NamedInstance() = default;

    NamedInstance(const NamedInstance &) = default;

    NamedInstance(NamedInstance &&) = default;

    NamedInstance &operator=(const NamedInstance &) = default;

    NamedInstance &operator=(NamedInstance &&) = default;

    void SetClassName(const std::string &name);

    void SetFunctionUrn(const std::string &urn);

    void SetNeedOrder(bool needOrder);

    void SetGroupName(const std::string &groupName);

    /*!
     * @brief Constructs a Function for a named instance's function call, preparing to execute the function.
     * @tparam F The type of the member function.
     * @param memberFunc The function to be executed within the instance. If the instance is bound to a class object,
     * the function must be a member function of the class.
     * @return An instance of the `InstanceFunctionHandler` class, which provides member methods for executing the
     * function call.
     *
     * @snippet{trimleft} instance_example.cpp instance function
     */
    template <typename F>
    InstanceFunctionHandler<F, InstanceType> Function(F memberFunc);

    /**
     * @brief Sends a request to invoke a C++ class member function to a remote backend, which executes the user
     * function call.
     * @tparam R The return type of the function.
     * @param functionName The name of the C++ class member function.
     * @return An `InstanceFunctionHandler` object, which can be used to invoke the instance by calling its `Invoke`
     * member function.
     *
     * @snippet{trimleft} cpp_instance_example.cpp cpp instance function
     */
    template <typename R>
    InstanceFunctionHandler<CppClassMethod<R>, InstanceType> CppFunction(const std::string &functionName);

    /**
     * @brief Sends a request to invoke a Python class member function to a remote backend, which executes the user
     * function call.
     * @tparam R The return type of the function.
     * @param functionName The name of the Python class member function.
     * @return An `InstanceFunctionHandler` object, which can be used to invoke the instance by calling its `Invoke`
     * member function.
     *
     * @snippet{trimleft} py_instance_example.cpp py instance function
     */
    template <typename R>
    InstanceFunctionHandler<PyClassMethod<R>, InstanceType> PyFunction(const std::string &functionName);

    /**
     * @brief Sends a request to invoke a Java class member function to a remote backend, which executes the user
     * function call.
     * @tparam R The return type of the function.
     * @param functionName The name of the Java class member function.
     * @return An `InstanceFunctionHandler` object, which can be used to invoke the instance by calling its `Invoke`
     * member function.
     *
     * @snippet{trimleft} java_instance_example.cpp java instance function
     */
    template <typename R>
    InstanceFunctionHandler<JavaClassMethod<R>, InstanceType> JavaFunction(const std::string &functionName);

    const std::string GetInstanceId();

    const std::string GetObjectId();

    void SetAlwaysLocalMode(bool isLocalMode);

    void SetName(const std::string &instanceName);

    void SetNs(const std::string &nsInput);

    /**
     * @brief For an instance handle, this indicates deleting an already created function instance. The default timeout
     * for the current kill request is 30 seconds. In scenarios such as high disk load or etcd failures, the kill
     * request processing time may exceed 30 seconds, causing the interface to throw a timeout exception. Since the kill
     * request has a retry mechanism, users can choose to ignore the timeout exception or retry after capturing it. For
     * Range scheduling handles, this indicates deleting a group of already created function instances.
     * @throws Exception An exception is thrown if the function instance deletion fails, along with corresponding
     * error messages, such as deleting a function instance that has already exited prematurely or receiving a timeout
     * when deleting the function instance response.
     *
     * @snippet{trimleft} instance_example.cpp terminate instance
     */
    void Terminate();

    /**
     * @brief Supports synchronous or asynchronous termination. For an instance handle, this indicates deleting an
     * already created function instance. When synchronous termination is not enabled, the default timeout for the
     * current kill request is 30 seconds. In scenarios such as high disk load or etcd failures, the kill request
     * processing time may exceed 30 seconds, causing the interface to throw a timeout exception. Since the kill request
     * has a retry mechanism, users can choose to ignore the timeout exception or retry after capturing it. When
     * synchronous termination is enabled, this interface will block until the instance completely exits. For Range
     * scheduling handles, this indicates deleting a group of already created function instances.
     * @param isSync Whether to enable synchronous termination. If `true`, it sends a kill request with the signal
     * `killInstanceSync` to the function-proxy, and the kernel synchronously kills the instance. If `false`, it sends a
     * kill request with the signal `killInstance` to the function-proxy, and the kernel asynchronously kills the
     * instance.
     * @throws Exception An exception is thrown if the function instance deletion fails, along with corresponding
     * error messages, such as deleting a function instance that has already exited prematurely or receiving a timeout
     * when deleting the function instance response.
     *
     * @snippet{trimleft} instance_example.cpp terminate instance sync
     */
    void Terminate(bool isSync);

    /**
     * @brief Supports synchronous or asynchronous termination. For an instance handle, this indicates deleting an
     * already created function instance. When synchronous termination is not enabled, the default timeout for the
     * current kill request is 30 seconds. In scenarios such as high disk load or etcd failures, the kill request
     * processing time may exceed 30 seconds, causing the interface to throw a timeout exception. Since the kill request
     * has a retry mechanism, users can choose to ignore the timeout exception or retry after capturing it. When
     * synchronous termination is enabled, this interface will block until the instance completely exits.
     * @param isSync Whether to enable synchronous termination. If `true`, it sends a kill request with the signal
     * `killInstanceSync` to the function-proxy, and the kernel synchronously kills the instance. If `false`, it sends a
     * kill request with the signal `killInstance` to the function-proxy, and the kernel asynchronously kills the
     * instance.
     * @return A future for the kill result.
     *
     * @snippet{trimleft} instance_example.cpp terminate instance sync
     */
    std::shared_future<void> AsyncTerminate(bool isSync);

    /**
     * @brief Specifies the timeout in seconds for waiting until a set of instances in Range scheduling are scheduled
     * and their instance IDs are returned, generating a list of `NamedInstance` objects. This parameter is optional. If
     * provided, it uses the specified timeout; otherwise, it uses `-1` to indicate an infinite wait.
     * @return A list of `NamedInstance` objects, which can be iterated to obtain the handles of individual instances
     * for invocation.
     * @throws Exception An exception is thrown if the backend interface call fails, such as due to insufficient
     * resources, where multiple instances within the range are not scheduled, leading to a timeout when obtaining the
     * instance ID list.
     *
     * @snippet{trimleft} range_example.cpp GetInstances
     */
    std::vector<NamedInstance<InstanceType>> GetInstances(int timeoutSec = NO_TIMEOUT);

    /**
     * @brief Users can use this method to obtain handle information, which can be serialized and stored in databases or
     * other persistent tools.
     * @return Handle information of `NamedInstance` stored in key-value pairs.
     *
     * @snippet{trimleft} instance_example.cpp Export
     */
    FormattedMap Export() const;

    /**
     * @brief Users can use this method to import handle information, which can be retrieved from databases or other
     * persistent tools and deserialized for import.
     * @param input Handle information of `NamedInstance` stored in key-value pairs.
     *
     * @snippet{trimleft} instance_example.cpp Import
     */
    void Import(FormattedMap &input);

private:
    std::string DefaultIfNotFound(FormattedMap &input, const std::string &key, const std::string &def);
    std::vector<NamedInstance<InstanceType>> BuildInstanceHandlers(std::vector<std::string> instanceIds);
    std::string instanceId;
    std::string realInstanceId;
    std::string className;
    std::string functionUrn;
    bool needOrder{false};
    std::string groupName;  // used for range scheduling.
    std::string groupInsIds;
    bool alwaysLocalMode = false;
    std::string name;
    std::string ns;
};

template <typename InstanceType>
void NamedInstance<InstanceType>::SetClassName(const std::string &name)
{
    this->className = name;
}

template <typename InstanceType>
void NamedInstance<InstanceType>::SetFunctionUrn(const std::string &urn)
{
    this->functionUrn = urn;
}

template <typename InstanceType>
void NamedInstance<InstanceType>::SetNeedOrder(bool needOrder)
{
    this->needOrder = needOrder;
}

template <typename InstanceType>
void NamedInstance<InstanceType>::SetGroupName(const std::string &groupName)
{
    this->groupName = groupName;
}

template <typename InstanceType>
void NamedInstance<InstanceType>::SetAlwaysLocalMode(bool isLocalMode)
{
    this->alwaysLocalMode = isLocalMode;
}

template <typename InstanceType>
void NamedInstance<InstanceType>::SetName(const std::string &instanceName)
{
    this->name = instanceName;
}

template <typename InstanceType>
void NamedInstance<InstanceType>::SetNs(const std::string &nsInput)
{
    this->ns = nsInput;
}

template <typename InstanceType>
template <typename F>
InstanceFunctionHandler<F, InstanceType> NamedInstance<InstanceType>::Function(F memberFunc)
{
    THROW_WHEN_IS_RANGE(groupName);
    internal::FuncMeta funcMeta;
    funcMeta.funcName = internal::FunctionManager::Singleton().GetFunctionName(memberFunc);
    funcMeta.language = YR::internal::FunctionLanguage::FUNC_LANG_CPP;
    funcMeta.name = name;
    funcMeta.ns = ns;
    return alwaysLocalMode
               ? InstanceFunctionHandler<F, InstanceType>(memberFunc, instanceId, true)
               : InstanceFunctionHandler<F, InstanceType>(funcMeta, instanceId, YR::internal::GetRuntime().get());
}

template <typename InstanceType>
template <typename R>
InstanceFunctionHandler<CppClassMethod<R>, InstanceType> NamedInstance<InstanceType>::CppFunction(
    const std::string &functionName)
{
    THROW_WHEN_IS_RANGE(groupName);
    internal::FuncMeta funcMeta;
    funcMeta.funcName = functionName;
    funcMeta.funcUrn = functionUrn;
    funcMeta.language = YR::internal::FunctionLanguage::FUNC_LANG_CPP;
    funcMeta.name = name;
    funcMeta.ns = ns;
    return InstanceFunctionHandler<CppClassMethod<R>, InstanceType>(funcMeta, instanceId,
                                                                    YR::internal::GetRuntime().get());
}

template <typename InstanceType>
template <typename R>
InstanceFunctionHandler<PyClassMethod<R>, InstanceType> NamedInstance<InstanceType>::PyFunction(
    const std::string &functionName)
{
    THROW_WHEN_IS_RANGE(groupName);
    internal::FuncMeta funcMeta;
    funcMeta.funcName = functionName;
    funcMeta.className = className;
    funcMeta.funcUrn = functionUrn;
    funcMeta.language = YR::internal::FunctionLanguage::FUNC_LANG_PYTHON;
    funcMeta.name = name;
    funcMeta.ns = ns;
    return InstanceFunctionHandler<PyClassMethod<R>, InstanceType>(funcMeta, instanceId,
                                                                   YR::internal::GetRuntime().get());
}

template <typename InstanceType>
template <typename R>
InstanceFunctionHandler<JavaClassMethod<R>, InstanceType> NamedInstance<InstanceType>::JavaFunction(
    const std::string &functionName)
{
    THROW_WHEN_IS_RANGE(groupName);
    internal::FuncMeta funcMeta;
    funcMeta.className = className;
    funcMeta.funcName = functionName;
    funcMeta.funcUrn = functionUrn;
    funcMeta.language = YR::internal::FunctionLanguage::FUNC_LANG_JAVA;
    funcMeta.name = name;
    funcMeta.ns = ns;
    return InstanceFunctionHandler<JavaClassMethod<R>, InstanceType>(funcMeta, instanceId,
                                                                     YR::internal::GetRuntime().get());
}

template <typename InstanceType>
FormattedMap NamedInstance<InstanceType>::Export() const
{
    CheckInitialized();
    FormattedMap out;
    out[INSTANCE_KEY] = instanceId;
    out[CLASS_NAME] = className;
    out[FUNCTION_URN] = functionUrn;
    out[NEED_ORDER] = needOrder ? "true" : "false";
    out[GROUP_NAME] = groupName;
    if (!groupName.empty() && groupInsIds.empty()) {
        out[GROUP_INS_IDS] = YR::internal::GetRuntime()->GetGroupInstanceIds(instanceId);
    } else if (!groupName.empty() && !groupInsIds.empty()) {
        out[GROUP_INS_IDS] = groupInsIds;
    } else if (realInstanceId.empty()) {
        out[INSTANCE_ID] = YR::internal::GetRuntime()->GetRealInstanceId(instanceId);
        out[INSTANCE_ROUTE] = YR::internal::GetRuntime()->GetInstanceRoute(instanceId);
    } else {
        out[INSTANCE_ID] = realInstanceId;
        out[INSTANCE_ROUTE] = YR::internal::GetRuntime()->GetInstanceRoute(instanceId);
    }
    return out;
}

template <typename InstanceType>
void NamedInstance<InstanceType>::Import(FormattedMap &input)
{
    CheckInitialized();
    instanceId = DefaultIfNotFound(input, INSTANCE_KEY, "");
    className = DefaultIfNotFound(input, CLASS_NAME, "");
    functionUrn = DefaultIfNotFound(input, FUNCTION_URN, "");
    realInstanceId = DefaultIfNotFound(input, INSTANCE_ID, "");
    groupName = DefaultIfNotFound(input, GROUP_NAME, "");
    needOrder = DefaultIfNotFound(input, NEED_ORDER, "false") == "false" ? false : true;
    groupInsIds = DefaultIfNotFound(input, GROUP_INS_IDS, "");
    auto instanceRoute = DefaultIfNotFound(input, INSTANCE_ROUTE, "");
    InvokeOptions instOpts;
    instOpts.needOrder = needOrder;
    if (!groupName.empty()) {
        YR::internal::GetRuntime()->SaveGroupInstanceIds(instanceId, groupInsIds, instOpts);
    } else {
        YR::internal::GetRuntime()->SaveRealInstanceId(instanceId, realInstanceId, instOpts);
    }
    if (!instanceRoute.empty()) {
        YR::internal::GetRuntime()->SaveInstanceRoute(instanceId, instanceRoute);
    }
}

template <typename InstanceType>
std::string NamedInstance<InstanceType>::DefaultIfNotFound(FormattedMap &input, const std::string &key,
                                                           const std::string &def)
{
    std::string ret;
    auto iter = input.find(key);
    if (iter != input.end()) {
        ret = input[key];
    } else {
        ret = def;
    }
    return ret;
}

template <typename InstanceType>
std::vector<NamedInstance<InstanceType>> NamedInstance<InstanceType>::GetInstances(int timeoutSec)
{
    CheckInitialized();
    THROW_WHEN_IS_NOT_RANGE(groupName);
    auto instanceIds = YR::internal::GetRuntime()->GetInstances(instanceId, timeoutSec);
    return BuildInstanceHandlers(instanceIds);
}

template <typename InstanceType>
std::vector<NamedInstance<InstanceType>> NamedInstance<InstanceType>::BuildInstanceHandlers(
    std::vector<std::string> instanceIds)
{
    std::vector<NamedInstance<InstanceType>> res;
    for (auto &insId : instanceIds) {
        NamedInstance<InstanceType> handler(insId);
        handler.SetClassName(className);
        handler.SetFunctionUrn(functionUrn);
        handler.SetNeedOrder(needOrder);
        res.push_back(handler);
    }
    return res;
}

template <typename InstanceType>
const std::string NamedInstance<InstanceType>::GetInstanceId()
{
    return YR::internal::GetRuntime()->GetRealInstanceId(instanceId);
}

template <typename InstanceType>
const std::string NamedInstance<InstanceType>::GetObjectId()
{
    return instanceId;
}

template <typename InstanceType>
void NamedInstance<InstanceType>::Terminate()
{
    CheckInitialized();
    if (internal::IsLocalMode() || this->alwaysLocalMode) {
        YR::internal::LocalInstanceManager<InstanceType>::Singleton().DelLocalInstance(instanceId);
        return;
    }
    if (groupName.empty()) {
        YR::internal::GetRuntime()->TerminateInstance(instanceId);
    } else {
        YR::internal::GetRuntime()->GroupTerminate(groupName);
    }
}

template <typename InstanceType>
void NamedInstance<InstanceType>::Terminate(bool isSync)
{
    CheckInitialized();
    if (internal::IsLocalMode() || this->alwaysLocalMode) {
        YR::internal::LocalInstanceManager<InstanceType>::Singleton().DelLocalInstance(instanceId);
        return;
    }
    if (!groupName.empty()) {
        YR::internal::GetRuntime()->GroupTerminate(groupName);
    }
    if (isSync) {
        YR::internal::GetRuntime()->TerminateInstanceSync(instanceId);
    } else {
        YR::internal::GetRuntime()->TerminateInstance(instanceId);
    }
}

template <typename InstanceType>
std::shared_future<void> NamedInstance<InstanceType>::AsyncTerminate(bool isSync)
{
    CheckInitialized();
    if (internal::IsLocalMode() || this->alwaysLocalMode) {
        YR::internal::LocalInstanceManager<InstanceType>::Singleton().DelLocalInstance(instanceId);
        auto promise = std::make_shared<std::promise<void>>();
        promise->set_value();
        return promise->get_future().share();
    }
    if (!groupName.empty()) {
        throw Exception::IncorrectFunctionUsageException("range instance does not support async terminate");
    }
    return YR::internal::GetRuntime()->TerminateInstanceAsync(instanceId, isSync);
}

}  // namespace YR

namespace msgpack {
MSGPACK_API_VERSION_NAMESPACE(MSGPACK_DEFAULT_API_NS)
{
    namespace adaptor {

    template <typename InstanceType>
    struct convert<YR::NamedInstance<InstanceType>> {
        msgpack::object const &operator()(msgpack::object const &o, YR::NamedInstance<InstanceType> &t) const
        {
            if (o.type != msgpack::type::MAP) {
                throw msgpack::type_error();
            }
            msgpack::object_kv *p(o.via.map.ptr);
            msgpack::object_kv *const pend(o.via.map.ptr + o.via.map.size);
            YR::FormattedMap infoMap;
            for (; p != pend; ++p) {
                if ((p->key.type != msgpack::type::STR && p->key.type != msgpack::type::BIN) ||
                    (p->val.type != msgpack::type::STR && p->val.type != msgpack::type::BIN)) {
                    continue;
                }
                std::string key;
                p->key.convert(key);
                p->val.convert(infoMap[std::move(key)]);
            }
            t.Import(infoMap);
            return o;
        }
    };

    template <typename InstanceType>
    struct pack<YR::NamedInstance<InstanceType>> {
        template <typename Stream>
        msgpack::packer<Stream> &operator()(msgpack::packer<Stream> &p, YR::NamedInstance<InstanceType> const &t) const
        {
            auto infoMap = t.Export();
            p.pack(infoMap);
            return p;
        }
    };
    }  // namespace adaptor
}  // MSGPACK_API_VERSION_NAMESPACE(MSGPACK_DEFAULT_API_NS)
}  // namespace msgpack
