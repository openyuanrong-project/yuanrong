/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
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

#include "jni_types.h"
#include "src/dto/invoke_options.h"
#include "src/proto/libruntime.pb.h"

namespace YR {
namespace jni {
class JNIFunctionMeta {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static YR::Libruntime::FunctionMeta FromJava(JNIEnv *env, jobject obj);
    static jobject FromCc(JNIEnv *env, const YR::Libruntime::FunctionMeta &meta);
    static std::string StringGetter(JNIEnv *env, jobject obj, jmethodID jm);
    static std::string GetAppName(JNIEnv *env, jobject obj);
    static std::string GetModuleName(JNIEnv *env, jobject obj);
    static libruntime::LanguageType GetLanguageType(JNIEnv *env, jobject obj);
    static std::string GetClassName(JNIEnv *env, jobject obj);
    static std::string GetFuncName(JNIEnv *env, jobject obj);
    static std::string GetFunctionID(JNIEnv *env, jobject obj);
    static std::string GetSignature(JNIEnv *env, jobject obj);
    static libruntime::ApiType GetApiType(JNIEnv *env, jobject obj);

private:
    inline static jclass clz_ = nullptr;
    inline static jclass factoryClz_ = nullptr;
    inline static jclass clzLanguageType_ = nullptr;
    inline static jclass clzApiType_ = nullptr;
    inline static jclass clzAccurateStackTraceEle_ = nullptr;
    inline static jmethodID init_ = nullptr;
    inline static jmethodID getFuncName_ = nullptr;
    inline static jmethodID getFunctionID_ = nullptr;
    inline static jmethodID getSignature_ = nullptr;
    inline static jmethodID getClassName_ = nullptr;
    inline static jmethodID getAppName_ = nullptr;
    inline static jmethodID getModuleName_ = nullptr;
    inline static jmethodID getLangaugeType_ = nullptr;
    inline static jmethodID getApiType_ = nullptr;
    inline static jmethodID getNs_ = nullptr;
    inline static jmethodID getName_ = nullptr;
};
}
}