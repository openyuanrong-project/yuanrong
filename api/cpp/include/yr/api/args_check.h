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
#include <functional>
#include <type_traits>
#include "yr/api/cross_lang.h"
#include "yr/api/object_ref.h"

namespace YR {
namespace internal {

template <typename T>
struct RemoveObjectRef {
    using type = T;
};

template <typename T>
struct RemoveObjectRef<ObjectRef<T>> {
    using type = T;
};

template <typename T>
struct RemoveObjectRef<YR::ObjectRef<T> &> {
    using type = T;
};

template <typename T>
struct RemoveObjectRef<YR::ObjectRef<T> &&> {
    using type = T;
};

template <typename T>
struct RemoveObjectRef<const YR::ObjectRef<T> &> {
    using type = T;
};

template <typename F, typename... Args>
struct ConstructibleCheck : std::is_constructible<std::function<void(Args...)>,
                                                  std::reference_wrapper<typename std::remove_reference<F>::type>> {
};

template <typename Function, typename... Args>
inline std::enable_if_t<!std::is_member_function_pointer<Function>::value> ArgumentsCheck()
{
    static_assert(ConstructibleCheck<Function, typename RemoveObjectRef<Args>::type...>::value,
                  "arguments type could not be constructed");
}

template <typename Function, typename... Args>
inline std::enable_if_t<std::is_member_function_pointer<Function>::value> ArgumentsCheck()
{
    using ClassType = boost::callable_traits::class_of_t<Function>;
    static_assert(ConstructibleCheck<Function, ClassType &, typename RemoveObjectRef<Args>::type...>::value,
                  "arguments type could not be constructed");
}

template <typename Function, typename... Args>
inline std::enable_if_t<!internal::IsCrossLang<Function>::value> ArgumentsCheckWrapper()
{
    ArgumentsCheck<Function, Args...>();
}

template <typename Function, typename... Args>
inline std::enable_if_t<internal::IsCrossLang<Function>::value> ArgumentsCheckWrapper()
{
}
}  // namespace internal
}  // namespace YR
