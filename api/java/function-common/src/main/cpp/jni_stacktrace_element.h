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

namespace YR {
namespace jni {
class JNIStackTraceElement {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static jobject FromCc(JNIEnv *env, const YR::Libruntime::StackTraceElement &stackTraceElement);
    static YR::Libruntime::StackTraceElement FromJava(JNIEnv *env, jobject obj);
    static std::vector<YR::Libruntime::StackTraceElement> ListFromJava(JNIEnv *env, jobject objList);
    static jobject ListFromCc(JNIEnv *env, std::vector<YR::Libruntime::StackTraceElement> objs);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID init_ = nullptr;
    inline static jmethodID getClassName_ = nullptr;
    inline static jmethodID getMethodName_ = nullptr;
    inline static jmethodID getFileName_ = nullptr;
    inline static jmethodID getLineNumber_ = nullptr;
};
}
}