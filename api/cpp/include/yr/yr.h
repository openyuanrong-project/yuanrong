/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
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
#include <chrono>
#include <memory>
#include <unordered_map>
#include "yr/api/check_initialized.h"
#include "yr/api/client_info.h"
#include "yr/api/config.h"
#include "yr/api/constant.h"
#include "yr/api/cross_lang.h"
#include "yr/api/err_type.h"
#include "yr/api/exception.h"
#include "yr/api/function_handler.h"
#include "yr/api/function_manager.h"
#include "yr/api/group.h"
#include "yr/api/hetero_exception.h"
#include "yr/api/hetero_manager.h"
#include "yr/api/instance_creator.h"
#include "yr/api/kv.h"
#include "yr/api/mutable_buffer.h"
#include "yr/api/node.h"
#include "yr/api/object_ref.h"
#include "yr/api/object_store.h"
#include "yr/api/runtime_manager.h"
#include "yr/api/serdes.h"
#include "yr/api/stream.h"
#include "yr/api/wait_result.h"
#include "yr/api/yr_core.h"
#include "yr/api/yr_invoke.h"
#include "yr/collective/collective.h"

extern thread_local std::unordered_set<std::string> localNestedObjList;
namespace YR {
/**
 * @brief Represents a pair of `vector<ObjectRef>`, where the former represents completed ObjectRef and the latter
 * represents pending ObjectRef.
 * @tparam T The type of the objects being waited for.
 */
template <typename T>
using WaitResult = std::pair<std::vector<ObjectRef<T>>, std::vector<ObjectRef<T>>>;

void Run(int argc, char *argv[]);


/*!
  @brief Exit the current function instance

  This function is not supported for local calls (either in LOCAL_MODE or local calls within CLUSTER_MODE).
  If called locally, it will throw a "Not support exit out of cluster" exception.

  @throw Exception If called in a local environment,
  it throws an exception with the message "Not support exit out of cluster".

  @code
  void AddOneAndPut(int x) {
      YR::KV().Set("key", std::to_string(x));
      // The function instance will exit after setting the key and value
      YR::Exit();
  }

  YR_INVOKE(AddOneAndPut);
  @endcode
*/
void Exit();

/*!
  @brief Determine if Yuanrong is in local mode.

  @return `true` if Yuanrong is in local mode, `false` otherwise.

  @throws Exception If CheckInitialized is called before the system is initialized,
  the exception "Please init YR first" will be thrown.
 */
bool IsLocalMode();

template <typename T>
using InstanceHandler = NamedInstance<T>;

/*!
  @brief Get the value of an ObjectRef.
  @details Get the value of the object stored in the backend storage according to the key of the object. The interface
  call will block until the value of the object is obtained or the timeout occurs.
  @tparam T the type of the object, Must not be void.
  @param obj the obj's ObjectRef.
  @param timeout the timeout.
  @return the shared pointer of the real value.
  @throw Exception if the object doesn't exist or timeout.

  @snippet{trimleft} get_put_example.cpp Get and put a single object
 */
template <typename T>
std::shared_ptr<T> Get(const ObjectRef<T> &obj, int timeout = DEFAULT_GET_TIMEOUT_SEC);

/*!
  @brief Get the value of an ObjectRef.

  @tparam T the type of the object, Must not be void.
  @param objs the objs' ObjectRef as a vector, should less than ``10000``.
  @param timeout the timeout.Default to ``300``.
  @param allowPartial if set to true, partialially ok will not throw an error.
  @return the shared pointers of the real values.
  @throw Exception if the all object don't exist or timeout and not allowPartial.

  @snippet{trimleft} get_put_example.cpp Get and put multiple objects
 */
template <typename T>
std::vector<std::shared_ptr<T>> Get(const std::vector<ObjectRef<T>> &objs, int timeout = DEFAULT_GET_TIMEOUT_SEC,
                                    bool allowPartial = false);

/*!
  @brief Waits for the value of an object in the data system to be ready based on the object's key.
  @tparam T The type of the object to wait for.
  @param obj A reference to the object in the data system.
  @param timeoutSec Timeout limit in milliseconds. ``-1`` indicates no time limit.

  @snippet{trimleft} wait_example.cpp Wait a single object
 */
template <typename T>
void Wait(const ObjectRef<T> &obj, int timeoutSec = -1);

/*!
  @brief Waits for the values of a set of objects in the data system to be ready based on their keys.
  @details When waiting for a group of ObjectRef, you can specify to wait for at least waitNum objects to be calculated.
  @tparam T The type of the objects to wait for.
  @param objs A collection of object references in the data system.
  @param waitNum The minimum number of `ObjectRef` to wait for.
  @param timeoutSec Timeout limit in milliseconds. `-1` indicates no time limit.
  @return A pair of `vector<ObjectRef>`, where the first vector contains completed `ObjectRef` and the second contains
  pending `ObjectRef`. The following invariant always holds:
  1. `waitNum <= waitResult.first.size() <= objs.size()`
  2. `waitResult.first.size() + waitResult.second.size() == objs.size()`

  @snippet{trimleft} wait_example.cpp Wait multiple objects
 */
template <typename T>
WaitResult<T> Wait(const std::vector<ObjectRef<T>> &objs, std::size_t waitNum, int timeoutSec = -1);

/*!
  @brief Cancel the corresponding function call by specifying an ObjectRef.
  @tparam T The type of the object to cancel.
  @param obj A collection of object references in the data system.
  @param isForce When set to `true`, if a function call is currently executing, the function instance process
  corresponding to the function call will be forcibly terminated. Note: Currently, forcibly terminating function
  instances with a concurrency configuration other than 1 is not supported. The default value is `true`.
  @param isRecursive Setting it to `true` will cancel nested function calls. The default value is `false`.

  @snippet{trimleft} cancel_example.cpp Cancel a single object
 */
template <typename T>
void Cancel(const ObjectRef<T> &obj, bool isForce = true, bool isRecursive = false);

/*!
  @brief Cancel the corresponding function call by specifying a set of ObjectRef.
  @tparam T The type of the object to cancel.
  @param objs A reference to the object in the data system.
  @param isForce When set to `true`, if a function call is currently executing, the function instance process
  corresponding to the function call will be forcibly terminated. Note: Currently, forcibly terminating function
  instances with a concurrency configuration other than 1 is not supported. The default value is `true`.
  @param isRecursive Setting it to `true` will cancel nested function calls. The default value is `false`.
  @throws Exception
  1. Local mode does not support `Cancel`, and an exception "local mode does not support cancel" will be thrown.
  2. Objs should not be empty, and an exception "Cancel does not accept empty object list" will be thrown.

  @snippet{trimleft} cancel_example.cpp Cancel multiple objects
 */
template <typename T>
void Cancel(const std::vector<ObjectRef<T>> &objs, bool isForce = true, bool isRecursive = false);

/*!
   @brief Determine if the current code is running on the remote server

   @return true if the current code is running on the remote server, false otherwise.

   @code
      int f(int x) {
      bool onCloud = YR::IsOnCloud(); // true
      return x;
   }

   YR_INVOKE(f);

   int main(void) {
       YR::Config conf;
       conf.mode = Config::Mode::LOCAL_MODE;
       YR::Init(conf);
       bool onCloud = YR::IsOnCloud(); // false

       return 0;
   }
   @endcode
*/
bool IsOnCloud();

/**
 * @brief Saves the instance state.
 * @param timeout Timeout in seconds, default value: 30.
 * @throws Exception
 * 1. Local mode does not support `SaveState`, and an exception "SaveState is not supported in local mode" will be
 * thrown.
 * 2. Cluster mode only supports this operation in remote code; if called in local code, an exception "SaveState is only
 * supported on cloud with posix api" will be thrown.
 *
 * @snippet{trimleft} save_and_load_state.cpp save and load state
 */
void SaveState(const int &timeout = DEFAULT_SAVE_LOAD_STATE_TIMEOUT);

/**
 * @brief Loads the saved instance state.
 * @param timeout Timeout in seconds, default value: 30.
 * @throws Exception
 * 1. Local mode does not support `LoadState`, and an exception "LoadState is not supported in local mode" will be
 * thrown.
 * 2. Cluster mode only supports this operation in remote code; if called in local code, an exception "LoadState is only
 * supported on cloud with posix api" will be thrown.
 *
 * @snippet{trimleft} save_and_load_state.cpp save and load state
 */
void LoadState(const int &timeout = DEFAULT_SAVE_LOAD_STATE_TIMEOUT);

/**
 * @brief Creates a producer.
 * @param streamName The name of the stream.
 * @param producerConf Configuration information for the producer.
 * @return A pointer to the created producer.
 * @throws Exception **4006**: not support local mode.
 *
 * @snippet{trimleft} stream_example.cpp create producer
 */
std::shared_ptr<Producer> CreateProducer(const std::string &streamName, ProducerConf producerConf = {});

/**
 * @brief Create a consumer.
 * @param streamName The name of the stream. Must be less than 256 characters and contain only the following characters:
 * (a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;).
 * @param config Configuration information for the consumer.
 * @param autoAck If `autoAck` is true, the consumer will automatically send an Acknowledgment (Ack) for received
 * messages to the data system. Default value: false.
 * @return A pointer to the created consumer.
 * @throws Exception **4006**: not support local mode.
 *
 * @snippet{trimleft} stream_example.cpp create consumer
 */
std::shared_ptr<Consumer> Subscribe(const std::string &streamName, const SubscriptionConfig &config,
                                    bool autoAck = false);

/**
 * @brief Deletes a stream. When the global count of producers and consumers for the stream reaches zero, this stream is
 * no longer in use, and all related metadata on workers and master nodes will be cleaned up. This function can be
 * called on any Host node.
 * @param streamName The name of the stream. Must be less than 256 characters and contain only the following characters:
 * (a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;).
 * @throws Exception **4006**: not support local mode.
 *
 * @snippet{trimleft} stream_example.cpp delete stream
 */
void DeleteStream(const std::string &streamName);

/*!
 @brief Create an InstanceCreator for constructing an instance of a class.

 @tparam F The class type that must not inherit from `internal::CrossLangClass`.

 @param constructor A function that constructs an instance of the class and returns a pointer to the constructed object.

 @returns InstanceCreator<Cls>, A creator object that can be used to create the instance and specify additional options
 for the instance's resources.

 @code{.cpp}
 int main(void) {
    YR::Config conf;
    YR::Init(conf);

    auto counter = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto c = counter.Function(&Counter::Add).Invoke(1);
    std::cout << "counter is " << *YR::Get(c) << std::endl;

    return 0;
 }
 @endcode
*/
template <typename F>
std::enable_if_t<!std::is_base_of<internal::CrossLangClass, F>::value, YR::internal::InstanceCreator<F>> Instance(
    F constructor)
{
    internal::FuncMeta funcMeta;
    funcMeta.className = internal::GetClassName<F>();
    funcMeta.funcName = internal::FunctionManager::Singleton().GetFunctionName(constructor);
    funcMeta.language = YR::internal::FunctionLanguage::FUNC_LANG_CPP;
    return IsLocalMode() ? YR::internal::InstanceCreator<F>(constructor)
                         : YR::internal::InstanceCreator<F>(funcMeta, YR::internal::GetRuntime().get(), constructor);
}

template <typename Cls>
std::enable_if_t<std::is_base_of<internal::CrossLangClass, Cls>::value, YR::internal::InstanceCreator<Cls>> Instance(
    Cls cls)
{
    internal::FuncMeta funcMeta;
    funcMeta.className = cls.GetClassName();
    funcMeta.moduleName = cls.GetModuleName();
    funcMeta.funcName = cls.GetFuncName();
    funcMeta.language = cls.GetLangType();
    return YR::internal::InstanceCreator<Cls>(funcMeta, YR::internal::GetRuntime().get(), cls);
}

/*!
 @brief Create an InstanceCreator for constructing an instance of a class.

 @tparam F The class type that must not inherit from `internal::CrossLangClass`.

 @param constructor A function that constructs an instance of the class and returns a pointer to the constructed object.
 @param name Optional. The name of the named instance. If provided, the instance can be reused by this name.

 @returns InstanceCreator<Cls>, A creator object that can be used to create the instance and specify additional options
 for the instance's resources.

 @code{.cpp}
 int main(void) {
    YR::Config conf;
    YR::Init(conf);

    // The instance name of this named instance is name_1
    auto counter = YR::Instance(Counter::FactoryCreate, "name_1").Invoke(1);
    auto c = counter.Function(&Counter::Add).Invoke(1);
    std::cout << "counter is " << *YR::Get(c) << std::endl;

    // A handle to a named instance of name_1 will be generated, reusing the name_1 instance
    counter = YR::Instance(Counter::FactoryCreate, "name_1").Invoke(1);
    c = counter.Function(&Counter::Add).Invoke(1);
    std::cout << "counter is " << *YR::Get(c) << std::endl;

    return 0;
 }
 @endcode
*/
template <typename F>
std::enable_if_t<!std::is_base_of<internal::CrossLangClass, F>::value, YR::internal::InstanceCreator<F>> Instance(
    F constructor, const std::string &name)
{
    internal::FuncMeta funcMeta;
    funcMeta.className = internal::GetClassName<F>();
    funcMeta.funcName = internal::FunctionManager::Singleton().GetFunctionName(constructor);
    funcMeta.language = YR::internal::FunctionLanguage::FUNC_LANG_CPP;
    funcMeta.name = name;
    return IsLocalMode() ? YR::internal::InstanceCreator<F>(constructor)
                         : YR::internal::InstanceCreator<F>(funcMeta, YR::internal::GetRuntime().get(), constructor);
}

template <typename Cls>
std::enable_if_t<std::is_base_of<internal::CrossLangClass, Cls>::value, YR::internal::InstanceCreator<Cls>> Instance(
    Cls cls, const std::string &name)
{
    internal::FuncMeta funcMeta;
    funcMeta.className = cls.GetClassName();
    funcMeta.moduleName = cls.GetModuleName();
    funcMeta.funcName = cls.GetFuncName();
    funcMeta.language = cls.GetLangType();
    funcMeta.name = name;
    return YR::internal::InstanceCreator<Cls>(funcMeta, YR::internal::GetRuntime().get(), cls);
}

/*!
 @brief Create an InstanceCreator for constructing an instance of a class.

 @tparam F The class type that must not inherit from `internal::CrossLangClass`.

 @param constructor A function that constructs an instance of the class and returns a pointer to the constructed object.
 @param name Optional. The name of the named instance. If provided, the instance can be reused by this name.
 @param ns Optional. The namespace for the named instance. If both `name` and `ns` are provided, the instance name
 will be concatenated as `ns-name`.

 @returns InstanceCreator<Cls>, A creator object that can be used to create the instance and specify additional options
 for the instance's resources.

 @note This function is designed for object-oriented programming scenarios. The `InstanceCreator` returned by
 this function can only execute member functions of the class constructed by the provided function `cls`.
 If name is passed, or both name and nameSpace are passed, the instance is a named instance,
 and the instance name is specified by the user. The instance can be reused by the instance name.

 @code{.cpp}
 int main(void) {
     YR::Config conf;
     YR::Init(conf);

     // The instance name of this named instance is ns_1-name_1
     auto counter = YR::Instance(Counter::FactoryCreate, "name_1", "ns_1").Invoke(1);
     auto c = counter.Function(&Counter::Add).Invoke(1);
     std::cout << "counter is " << *YR::Get(c) << std::endl;

     // A handle to the named instance ns_1-name_1 will be generated, and the ns_1-name_1 instance will be reused.
     counter = YR::Instance(Counter::FactoryCreate, "name_1", "ns_1").Invoke(1);
     c = counter.Function(&Counter::Add).Invoke(1);
     std::cout << "counter is " << *YR::Get(c) << std::endl;

     return 0;
 }
 @endcode
*/
template <typename F>
std::enable_if_t<!std::is_base_of<internal::CrossLangClass, F>::value, YR::internal::InstanceCreator<F>> Instance(
    F constructor, const std::string &name, const std::string &ns)
{
    internal::FuncMeta funcMeta;
    funcMeta.className = internal::GetClassName<F>();
    funcMeta.funcName = internal::FunctionManager::Singleton().GetFunctionName(constructor);
    funcMeta.language = YR::internal::FunctionLanguage::FUNC_LANG_CPP;
    funcMeta.name = name;
    funcMeta.ns = ns;
    return IsLocalMode() ? YR::internal::InstanceCreator<F>(constructor)
                         : YR::internal::InstanceCreator<F>(funcMeta, YR::internal::GetRuntime().get(), constructor);
}

template <typename Cls>
std::enable_if_t<std::is_base_of<internal::CrossLangClass, Cls>::value, YR::internal::InstanceCreator<Cls>> Instance(
    Cls cls, const std::string &name, const std::string &ns)
{
    internal::FuncMeta funcMeta;
    funcMeta.className = cls.GetClassName();
    funcMeta.moduleName = cls.GetModuleName();
    funcMeta.funcName = cls.GetFuncName();
    funcMeta.language = cls.GetLangType();
    funcMeta.name = name;
    funcMeta.ns = ns;
    return YR::internal::InstanceCreator<Cls>(funcMeta, YR::internal::GetRuntime().get(), cls);
}

/*!
  @brief Constructs a FunctionHandler for a given function

  This function template creates a FunctionHandler object for a specified
  static function. The FunctionHandler object can then be used to execute the
  function and set various options for the function call, such as resource
  allocation.

  @tparam F The type of the function to be executed.
  @param f The static function to be called.
  @returns FunctionHandler<F>: A FunctionHandler object that provides methods
  to execute the function and set options for the function call.

  @throws Exception If the function has not been registered using the
  YR_INVOKE macro, an exception will be thrown.

  @snippet{trimleft} function_example.cpp Function
*/
template <typename F>
FunctionHandler<F> Function(F f)
{
    CheckInitialized();
    internal::FuncMeta funcMeta;
    funcMeta.funcName = YR::internal::FunctionManager::Singleton().GetFunctionName(f);
    funcMeta.language = YR::internal::FunctionLanguage::FUNC_LANG_CPP;
    return IsLocalMode() ? FunctionHandler<F>(f) : FunctionHandler<F>(funcMeta, f);
}

/*!
    @brief Create a FunctionHandler for invoking a C++ function by name.

    @tparam R The return type of the C++ function.
    @param funcName The name of the C++ function to invoke.
    @return A FunctionHandler object that can be used to execute the C++ function. The return type can be accessed
   through the CppFunctionHandler template class.

    @code
    // C++ function to be invoked
    int PlusOne(int x)
    {
        return x + 1;
    }

    YR_INVOKE(PlusOne);

    // C++ code to invoke the function
    int main(void)
    {
        YR::Config conf;
        YR::Init(conf);
        auto ref = YR::CppFunction<int>("PlusOne").Invoke(1);
        int res = *YR::Get(ref);   // get 2
        return 0;
    }
    @endcode
*/
template <typename R>
FunctionHandler<CppFunctionHandler<R>> CppFunction(const std::string &funcName)
{
    CheckInitialized();
    internal::FuncMeta funcMeta;
    funcMeta.funcName = funcName;
    funcMeta.language = YR::internal::FunctionLanguage::FUNC_LANG_CPP;
    return FunctionHandler<CppFunctionHandler<R>>(funcMeta, CppFunctionHandler<R>());
}

/*!
  @brief Used for C++ to call Python functions, constructs a call to a Python function.
  @tparam R The return type of the function.
  @param moduleName The name of the Python module where the function resides.
  @param functionName The name of the Python function.
  @return A `FunctionHandler` object, which provides methods to execute the function. The `PyFunctionHandler` is a
  template class within `FunctionHandler` that can be used to obtain the return type.

  @snippet{trimleft} py_function_example.cpp py function
 */
template <typename R>
FunctionHandler<PyFunctionHandler<R>> PyFunction(const std::string &moduleName, const std::string &functionName)
{
    CheckInitialized();
    internal::FuncMeta funcMeta;
    funcMeta.funcName = functionName;
    funcMeta.moduleName = moduleName;
    funcMeta.language = YR::internal::FunctionLanguage::FUNC_LANG_PYTHON;
    return FunctionHandler<PyFunctionHandler<R>>(funcMeta, PyFunctionHandler<R>());
}

/*!
   @brief Create a FunctionHandler for invoking a Java function from C++

   @tparam R The return type of the Java function.
   @param className The fully qualified class name of the Java function, including package name.
   If the class is an inner static class, use '$' to connect the outer class and inner class.
   @param functionName The name of the Java function to invoke.
   @return A FunctionHandler object that can be used to execute the Java function.
   The return type can be accessed through the JavaFunctionHandler template class.

   @code
   // Java code:
   package com.example;

   public class YrlibHandler {
   public static class MyYRApp {
       public static int add_one(int x) {
               return x + 1;
           }
       }
   }

   // C++ code:
   int main(void) {
       YR::Config conf;
       YR::Init(conf);
       auto ref = YR::JavaFunction<int>("com.example.YrlibHandler$MyYRApp", "add_one").Invoke(1);
       int res = *YR::Get(ref);   // get 2
       return 0;
   }
   @endcode
*/
template <typename R>
FunctionHandler<JavaFunctionHandler<R>> JavaFunction(const std::string &className, const std::string &functionName)
{
    CheckInitialized();
    internal::FuncMeta funcMeta;
    funcMeta.funcName = functionName;
    funcMeta.className = className;
    funcMeta.language = YR::internal::FunctionLanguage::FUNC_LANG_JAVA;
    return FunctionHandler<JavaFunctionHandler<R>>(funcMeta, JavaFunctionHandler<R>());
}

/*!
 @brief Put an object to datasystem

 @tparam T the template.
 @param val the object.
 @returns ObjectRef<T> the object ref of the object.

 @throw Exception if Yuanrong is not initialized.

 @snippet{trimleft} get_put_example.cpp Get and put a single object
 */
template <typename T>
ObjectRef<T> Put(const T &val)
{
    CheckInitialized();
    if (YR::internal::IsLocalMode()) {
        return YR::internal::GetLocalModeRuntime()->Put(val);
    }
    std::string objId;
    localNestedObjList.clear();
    if constexpr (boost::is_same<T, Buffer>::value) {
        objId = YR::internal::GetRuntime()->Put(std::make_shared<Buffer>(val), localNestedObjList);
    } else {
        std::shared_ptr<msgpack::sbuffer> data = std::make_shared<msgpack::sbuffer>(YR::internal::Serialize(val));
        objId = YR::internal::GetRuntime()->Put(data, localNestedObjList);
    }
    localNestedObjList.clear();
    return ObjectRef<T>(objId, false);
}

/*!
 @brief Put an object to datasystem

 @tparam T the template.
 @param val the object.
 @param createParam the create param, Refer to the parameter structure supplement below for details.
 @returns ObjectRef<T> the object ref of the object.

 @throw Exception if Yuanrong is not initialized.

 @snippet{trimleft} get_put_example.cpp Get and put a single object with param
 */
template <typename T>
ObjectRef<T> Put(const T &val, const CreateParam &createParam)
{
    CheckInitialized();
    if (YR::internal::IsLocalMode()) {
        return YR::internal::GetLocalModeRuntime()->Put(val);
    }
    std::string objId;
    localNestedObjList.clear();
    if constexpr (boost::is_same<T, Buffer>::value) {
        objId = YR::internal::GetRuntime()->Put(std::make_shared<Buffer>(val), localNestedObjList, createParam);
    } else {
        std::shared_ptr<msgpack::sbuffer> data = std::make_shared<msgpack::sbuffer>(YR::internal::Serialize(val));
        objId = YR::internal::GetRuntime()->Put(data, localNestedObjList, createParam);
    }
    localNestedObjList.clear();
    return ObjectRef<T>(objId, false);
}

template <typename T>
std::vector<std::shared_ptr<T>> Get(const std::vector<std::string> &objIds, int timeoutSec, bool allowPartial)
{
    internal::CheckTimeout(timeoutSec);
    std::vector<std::string> remainIds;
    std::unordered_map<std::string, std::list<size_t>> idToIndex;
    for (size_t i = 0; i < objIds.size(); i++) {
        remainIds.push_back(objIds[i]);
        idToIndex[remainIds[i]].push_back(i);
    }
    std::vector<std::shared_ptr<T>> returnObjects;
    returnObjects.resize(remainIds.size());
    int timeoutSecTmp = timeoutSec > TIMEOUT_MAX ? TIMEOUT_MAX : timeoutSec;
    int timeoutMs = timeoutSecTmp != NO_TIMEOUT ? timeoutSecTmp * S_TO_MS : NO_TIMEOUT;
    auto remainTimeoutMs = YR::internal::GetRuntime()->WaitBeforeGet(remainIds, timeoutMs, allowPartial);
    int limitedRetryTime = 0;
    internal::GetStatus status;
    ErrorInfo err;
    auto start = std::chrono::high_resolution_clock::now();
    auto getElapsedTime = [start]() {
        return std::chrono::duration_cast<std::chrono::milliseconds>(std::chrono::high_resolution_clock::now() - start)
            .count();
    };
    while ((getElapsedTime() <= remainTimeoutMs) || (remainTimeoutMs == NO_TIMEOUT) || (remainTimeoutMs == 0)) {
        int to = (remainTimeoutMs == NO_TIMEOUT) ? (DEFAULT_TIMEOUT_MS)
                                                 : (remainTimeoutMs - static_cast<int>(getElapsedTime()));
        to = to < 0 ? 0 : to;
        auto ret = YR::internal::GetRuntime()->Get(remainIds, to, limitedRetryTime);
        auto retryInfo = ret.first;
        auto remainBuffers = ret.second;
        auto needRetry = retryInfo.needRetry;
        err = retryInfo.errorInfo;
        internal::ExtractSuccessObjects(remainIds, remainBuffers, returnObjects, idToIndex);
        if (remainIds.empty()) {
            status = internal::GetStatus::ALL_SUCCESS;
            break;
        }
        if (!needRetry) {
            status = remainIds.size() == objIds.size() ? internal::GetStatus::ALL_FAILED
                                                     : internal::GetStatus::PARTIAL_SUCCESS;
            break;
        }
        if ((remainTimeoutMs != NO_TIMEOUT && getElapsedTime() > remainTimeoutMs) || (remainTimeoutMs == 0)) {
            status = remainIds.size() == objIds.size() ? internal::GetStatus::ALL_FAILED_AND_TIMEOUT
                                                     : internal::GetStatus::PARTIAL_SUCCESS_AND_TIMEOUT;
            break;
        }
        std::this_thread::sleep_for(std::chrono::seconds(GET_RETRY_INTERVAL));
        if ((remainTimeoutMs != NO_TIMEOUT && getElapsedTime() > remainTimeoutMs)) {
            status = remainIds.size() == objIds.size() ? internal::GetStatus::ALL_FAILED_AND_TIMEOUT
                                                     : internal::GetStatus::PARTIAL_SUCCESS_AND_TIMEOUT;
            break;
        }
    }
    ThrowExceptionBasedOnStatus(status, err, remainIds, timeoutMs, allowPartial);
    return returnObjects;
}

template <typename T>
std::shared_ptr<T> Get(const ObjectRef<T> &obj, int timeoutSec)
{
    CheckInitialized();
    if (obj.IsLocal()) {
        return internal::GetLocalModeRuntime()->Get(obj, timeoutSec);
    }
    auto result = Get<T>(std::vector<std::string>{obj.ID()}, timeoutSec, false);
    if (!result.empty()) {
        return result[0];
    }
    return nullptr;
}

template <typename T>
std::vector<std::shared_ptr<T>> Get(const std::vector<ObjectRef<T>> &objs, int timeoutSec, bool allowPartial)
{
    internal::CheckObjs(objs);
    if (objs[0].IsLocal()) {
        return YR::internal::GetLocalModeRuntime()->Get(objs, timeoutSec, allowPartial);
    }
    std::vector<std::string> objIds;
    for (auto obj: objs) {
        objIds.push_back(obj.ID());
    }
    return Get<T>(objIds, timeoutSec, allowPartial);
}

template <typename T>
void Wait(const ObjectRef<T> &obj, int timeoutSec)
{
    CheckInitialized();
    if (timeoutSec <= 0 && timeoutSec != -1) {
        throw Exception::InvalidParamException("timeout should be larger than 0 or be -1");
    }
    if (obj.IsLocal()) {
        YR::internal::GetLocalModeRuntime()->Wait(obj, timeoutSec);
    } else {
        YR::internal::GetRuntime()->Wait({obj.ID()}, 1, timeoutSec);
    }
}

template <typename T>
WaitResult<T> Wait(const std::vector<ObjectRef<T>> &objs, std::size_t waitNum, int timeoutSec)
{
    CheckInitialized();
    if (timeoutSec < 0 && timeoutSec != -1) {
        throw Exception::InvalidParamException("timeout should be larger than 0 or be -1");
    }

    if (objs.empty()) {
        throw Exception::InvalidParamException("Wait does not accept empty object list");
    }

    if (CheckRepeat(objs)) {
        throw Exception::InvalidParamException("duplicate objectRef exists in objs vector");
    }

    internal::CheckIfObjectRefsHomogeneous(objs);

    if (waitNum == 0) {
        throw Exception::InvalidParamException("waitNum cannot be 0");
    }

    waitNum = waitNum > objs.size() ? objs.size() : waitNum;
    if (objs[0].IsLocal()) {
        auto result = internal::GetLocalModeRuntime()->Wait(objs, waitNum, timeoutSec);
        WaitResult<T> ret;
        ret.first.reserve(objs.size());
        ret.second.reserve(objs.size());
        for (std::size_t i = 0; i < result.size(); ++i) {
            if (result[i]) {
                ret.first.push_back(objs[i]);
            } else {
                ret.second.push_back(objs[i]);
            }
        }
        return ret;
    }

    std::vector<std::string> ids;
    std::unordered_map<std::string, ObjectRef<T>> idObjMap;
    ids.reserve(objs.size());
    for (auto &obj : objs) {
        ids.emplace_back(obj.ID());
        idObjMap[obj.ID()] = obj;
    }
    YR::internal::WaitResult result = YR::internal::GetRuntime()->Wait(ids, waitNum, timeoutSec);

    WaitResult<T> ret;
    ret.first.reserve(result.readyIds.size());
    ret.second.reserve(result.unreadyIds.size());

    for (std::string id : result.readyIds) {
        ret.first.push_back(idObjMap[id]);
    }
    for (std::string id : result.unreadyIds) {
        ret.second.push_back(idObjMap[id]);
    }
    return ret;
}

void ReceiveRequestLoop(void);

template <typename T>
void Cancel(const ObjectRef<T> &obj, bool isForce, bool isRecursive)
{
    CheckInitialized();
    std::vector<std::string> objIDs{obj.ID()};
    YR::internal::GetRuntime()->Cancel(objIDs, isForce, isRecursive);
}

template <typename T>
void Cancel(const std::vector<ObjectRef<T>> &objs, bool isForce, bool isRecursive)
{
    CheckInitialized();
    if (objs.empty()) {
        throw Exception::InvalidParamException("Cancel does not accept empty object list");
    }
    internal::CheckIfObjectRefsHomogeneous(objs);
    if (objs[0].IsLocal()) {
        throw Exception::IncorrectFunctionUsageException("local mode does not support cancel");
    }
    std::vector<std::string> objIDs;
    for (auto &obj : objs) {
        objIDs.emplace_back(obj.ID());
    }
    YR::internal::GetRuntime()->Cancel(objIDs, isForce, isRecursive);
}

/*!
  @brief Get instance associated with the specified name and nameSpace within a timeout.
  @tparam F The type of the instance.
  @param name The instance name of the named instance.
  @param nameSpace Namespace of the named instance. The default value is `""`.
  @param timeoutSec The timeout in seconds. The default value is `60`.
  @return A named instance that can be used to call member functions of the class using the Function method.
  @throws Exception
  1. Name should not be empty, and an exception "name should not be empty" will be thrown.
  2. Timeout should be greater than `0`, and an exception "timeout should be greater than 0" will be thrown.

  @snippet{trimleft} get_instnce.cpp get instance
*/
template <typename F>
NamedInstance<F> GetInstance(const std::string &name, const std::string &nameSpace = "", int timeoutSec = 60)
{
    CheckInitialized();
    if (name.empty()) {
        throw Exception::InvalidParamException("name should not be empty");
    }
    if (timeoutSec < 0) {
        throw Exception::InvalidParamException("timeout should be greater than 0");
    }

    auto funcMeta = YR::internal::GetRuntime()->GetInstance(name, nameSpace, timeoutSec);
    NamedInstance<F> handler(nameSpace.empty() ? name : nameSpace + "-" + name);
    handler.SetAlwaysLocalMode(false);
    handler.SetClassName(funcMeta.className);
    handler.SetFunctionUrn(funcMeta.funcUrn);
    handler.SetName(funcMeta.name);
    handler.SetNs(funcMeta.ns);
    return handler;
}

/**
 * @brief Get node information in the cluster.
 * @return std::vector<Node>: node information, include id, alive, resources and labels.
 * @throws Exception if Yuanrong is not initialized or failed to get node information.
 */
std::vector<Node> Nodes();

std::shared_ptr<MutableBuffer> CreateBuffer(uint64_t size);

std::vector<std::shared_ptr<MutableBuffer>> Get(const std::vector<ObjectRef<MutableBuffer>> &objs,
                                                int timeoutSec = DEFAULT_GET_TIMEOUT_SEC);

std::string Serialize(ObjectRef<MutableBuffer> &obj);

ObjectRef<MutableBuffer> Deserialize(const void *value, int size);
}  // namespace YR