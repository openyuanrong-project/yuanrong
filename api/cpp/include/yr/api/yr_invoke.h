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

#include <string>
#include <vector>

#include <boost/utility/string_view.hpp>
#include <msgpack.hpp>

#include "yr/api/macro.h"

namespace YR {
static std::vector<std::string> YR_UNUSED ParseFuncNames(boost::string_view names)
{
    std::vector<std::string> namesStr;
    size_t begin = 0;
    while (begin < names.size()) {
        size_t end = names.find_first_of(",", begin);
        boost::string_view subStr;
        if (end == boost::string_view::npos) {
            subStr = names.substr(begin);
        } else {
            subStr = names.substr(begin, end - begin);
        }

        auto left = subStr.find_first_not_of(' ');
        auto right = subStr.find_last_not_of(' ');

        auto compactStr = subStr.substr(left, (right - left) + 1);
        namesStr.emplace_back(std::string(compactStr.data(), compactStr.size()));

        if (end == boost::string_view::npos) {
            break;
        }
        begin = end + 1;
    }

    return namesStr;
}

template <typename... Funcs>
static int RegisterInvokeFunctions(boost::string_view names, Funcs... funcs)
{
    auto funcNames = ParseFuncNames(names);

    size_t i = 0;
    std::initializer_list<size_t>{
        (internal::FunctionManager::Singleton().RegisterInvokeFunction(funcNames[i], funcs), i++)...};
    return 0;
}

template <typename... Funcs>
static int RegisterRecoverFunctions(boost::string_view names, Funcs... funcs)
{
    auto funcNames = ParseFuncNames(names);
    size_t i = 0;
    std::initializer_list<size_t>{
        (internal::FunctionManager::Singleton().RegisterRecoverFunction(funcNames[i], funcs), i++)...};
    return 0;
}

template <typename... Funcs>
static int RegisterShutdownFunctions(boost::string_view names, Funcs... funcs)
{
    auto funcNames = ParseFuncNames(names);
    size_t i = 0;
    std::initializer_list<size_t>{
        (internal::FunctionManager::Singleton().RegisterShutdownFunctions(funcNames[i], funcs), i++)...};
    return 0;
}

#define COMBINE(a, b) a##b

#define COMBINE_WRAPPER(a, b) COMBINE(a, b)

#define UNIQ_VAR_NAME(v) COMBINE_WRAPPER(v, __LINE__)

#define AUTO_VAR(v) UNIQ_VAR_NAME(v)

/**
 * @def YR_INVOKE
 * @brief Register functions for Yuanrong distributed invocation. In local mode, registered functions execute
 * within the current process. In cluster mode, functions execute remotely.
 * @details All functions intended for remote execution must be registered using the YR_INVOKE interface.
 * If a function is registered multiple times, the program will throw an exception and terminate during runtime.
 * @throws Exception
 * If a function is detected to be registered multiple times, the program will exit with an error
 * message.
 * @note Specifically, when using `printf` within remotely registered functions via YR_INVOKE,
 * note that Yuanrong's runtime kernel redirects standard output, switching it to full buffering mode.
 * This means output will only be written to disk when the buffer is full or `fflush` is explicitly called,
 * and newline characters will not be immediately printed.
 * For example:
 * @code
 * int HelloWorld()
 * {
 *    printf("helloworld\n"); // Not recommended; output may not immediately appear due to remote runtime buffering
 *    return 0;
 * }
 * YR_INVOKE(HelloWorld)
 * @endcode
 * Therefore, it is recommended to use `std::cout` or explicitly call `fflush`:
 * @code
 * int HelloWorld()
 * {
 *     std::cout << "helloworld" << std::endl; // Recommended
 *     // Or explicitly call fflush
 *     printf("helloworld\n");
 *     fflush(stdout);
 *     return 0;
 * }
 * YR_INVOKE(HelloWorld)
 * @endcode
 *
 * @snippet{trimleft} instance_example.cpp yr invoke
 */
#define YR_INVOKE(...) static auto AUTO_VAR(x) = YR::RegisterInvokeFunctions(#__VA_ARGS__, __VA_ARGS__);

/**
 * @brief Users can use this interface to perform data recovery operations.
 * @details Functions to be executed during instance recovery must be decorated with YR_RECOVER.
 * This macro is used to decorate user-defined state member functions within actors executing in the cloud.
 * Functions decorated with YR_RECOVER can utilize Yuanrong interfaces.
 * These functions will be executed in the following scenarios: during runtime recovery requests to restore instances.
 * @throws Exception
 * If a function is detected to be registered multiple times,
 * the program will exit with an error message.
 *
 * @snippet{trimleft} yr_recover_example.cpp yr recover
 */
#define YR_RECOVER(...) static auto AUTO_VAR(x) = YR::RegisterRecoverFunctions(#__VA_ARGS__, __VA_ARGS__);

/*!
   @def YR_SHUTDOWN
   @brief Users can use this interface to perform data cleanup or data persistence operations.
   @details Functions to be executed during graceful shutdown must be decorated with YR_SHUTDOWN.
   The function must have exactly one parameter of type uint64_t. Otherwise, the function will fail to execute
   on the cloud due to parameter mismatch. These functions are used to decorate user-defined graceful shutdown
   member functions within actors executing in the cloud. These functions can utilize Yuanrong interfaces.
   Functions decorated with YR_SHUTDOWN will be executed in the following scenarios:
   - When the runtime receives a shutdown request (e.g., when a user calls the terminate interface in the cloud
   environment).
   - When the runtime captures a Sigterm signal (e.g., during cluster scaling down, where the runtime-manager captures
   the Sigterm signal and forwards it to the runtime).

   @note
   - When defining custom graceful shutdown functions, the function must have exactly one parameter
   named gracePeriodSeconds of type uint64_t. Otherwise, the function will fail to execute on the cloud due to
   parameter mismatch.
   - If the execution time of the custom graceful shutdown function exceeds gracePeriodSeconds
   seconds, the cloud instance will not wait for the function to complete and will be immediately recycled.

   @snippet{trimleft} yr_shutdown_example.cpp yr shutdown
*/
#define YR_SHUTDOWN(...) static auto AUTO_VAR(x) = YR::RegisterShutdownFunctions(#__VA_ARGS__, __VA_ARGS__);

/*!
   @def YR_STATE
   @brief Marks a class parameter as state, and Yuanrong will automatically save and recover the
   state member variables of the class.
   @details Class member variables that need to be marked as state should be decorated with YR_STATE.
   Constraint: YR_STATE must be declared in the public section.

   @snippet{trimleft} yr_state_example.cpp yr state
*/
#define YR_STATE(...) MSGPACK_DEFINE(__VA_ARGS__)
}  // namespace YR
