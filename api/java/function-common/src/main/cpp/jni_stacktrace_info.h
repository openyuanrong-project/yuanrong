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

#include <jni.h>
#include <vector>
#include <string>

#include "src/libruntime/err_type.h"
#include "src/libruntime/stacktrace/stack_trace_info.h"
#include "jni_types.h"
#include "jni_stacktrace_element.h"

namespace YR {
namespace jni {

class JNIStackTraceInfo {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static jobject FromCc(JNIEnv *env, const YR::Libruntime::StackTraceInfo &stackTraceInfo);
    static YR::Libruntime::StackTraceInfo FromJava(JNIEnv *env, jobject o);
    static jobject ListFromCc(JNIEnv *env, const std::vector<YR::Libruntime::StackTraceInfo> stackTraceInfos);
    static std::vector<YR::Libruntime::StackTraceInfo> ListFromJava(JNIEnv *env, jobject objList);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID init_ = nullptr;
    inline static jmethodID getType_ = nullptr;
    inline static jmethodID getMessage_ = nullptr;
    inline static jmethodID getStackTraceElements_ = nullptr;
    inline static jmethodID getLanguage_ = nullptr;
};
}
}