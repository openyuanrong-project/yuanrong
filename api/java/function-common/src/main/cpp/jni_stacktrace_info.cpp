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

#include "jni_stacktrace_info.h"
#include "src/utility/logger/logger.h"

namespace YR {
namespace jni {
void JNIStackTraceInfo::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "org/yuanrong/exception/handler/traceback/StackTraceInfo");
    init_ =
        GetJMethod(env, clz_, "<init>", "(Ljava/lang/String;Ljava/lang/String;Ljava/util/List;Ljava/lang/String;)V");
    getType_ = GetJMethod(env, clz_, "getType", "()Ljava/lang/String;");
    getMessage_ = GetJMethod(env, clz_, "getMessage", "()Ljava/lang/String;");
    getStackTraceElements_ = GetJMethod(env, clz_, "getStackTraceElements", "()Ljava/util/List;");
    getLanguage_ = GetJMethod(env, clz_, "getLanguage", "()Ljava/lang/String;");
}

void JNIStackTraceInfo::Recycle(JNIEnv *env)
{
    if (clz_ != nullptr) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

jobject JNIStackTraceInfo::FromCc(JNIEnv *env, const YR::Libruntime::StackTraceInfo &stackTraceInfo)
{
    jstring jtype = JNIString::FromCc(env, stackTraceInfo.Type());
    if (jtype == nullptr) {
        YRLOG_ERROR("Failed to convert jtype from Cc code to Java");
        return nullptr;
    }

    jstring jmessage = JNIString::FromCc(env, stackTraceInfo.Message());
    if (jmessage == nullptr) {
        YRLOG_ERROR("Failed to convert jmessage from Cc code to Java");
        return nullptr;
    }
    std::vector<YR::Libruntime::StackTraceElement> elements = stackTraceInfo.StackTraceElements();
    auto jstackTraceElements = JNIStackTraceElement::ListFromCc(env, elements);

    jstring jlanguage = JNIString::FromCc(env, stackTraceInfo.Language());
    if (jlanguage == nullptr) {
        YRLOG_ERROR("Failed to convert jlanguage from Cc code to Java");
        return nullptr;
    }

    jobject jstackTraceInfo = (jobject)env->NewObject(clz_, init_, jtype, jmessage, jstackTraceElements, jlanguage);

    env->DeleteLocalRef(jtype);
    env->DeleteLocalRef(jmessage);
    env->DeleteLocalRef(jlanguage);
    env->DeleteLocalRef(jstackTraceElements);
    return jstackTraceInfo;
}

YR::Libruntime::StackTraceInfo JNIStackTraceInfo::FromJava(JNIEnv *env, jobject o)
{
    jstring jtype = static_cast<jstring>(env->CallObjectMethod(o, getType_));
    std::string ctype = JNIString::FromJava(env, jtype);
    env->DeleteLocalRef(jtype);

    jstring jmessage = static_cast<jstring>(env->CallObjectMethod(o, getMessage_));
    std::string cmessage = JNIString::FromJava(env, jmessage);
    env->DeleteLocalRef(jmessage);

    jstring jlanguage = static_cast<jstring>(env->CallObjectMethod(o, getLanguage_));
    std::string clanguage = JNIString::FromJava(env, jlanguage);
    env->DeleteLocalRef(jlanguage);

    jobject objList = env->CallObjectMethod(o, getStackTraceElements_);
    std::vector<YR::Libruntime::StackTraceElement> stackTraceElements =
        JNIStackTraceElement::ListFromJava(env, objList);
    env->DeleteLocalRef(objList);
    YR::Libruntime::StackTraceInfo stackTraceInfo(ctype, cmessage, stackTraceElements, clanguage);
    return stackTraceInfo;
}

jobject JNIStackTraceInfo::ListFromCc(JNIEnv *env, const std::vector<YR::Libruntime::StackTraceInfo> stackTraceInfos)
{
    jobject jstackTraceInfos = JNIArrayList::FromCc<YR::Libruntime::StackTraceInfo>(
        env, stackTraceInfos,
        [](JNIEnv *env, const YR::Libruntime::StackTraceInfo &info) { return JNIStackTraceInfo::FromCc(env, info); });
    return jstackTraceInfos;
}

std::vector<YR::Libruntime::StackTraceInfo> JNIStackTraceInfo::ListFromJava(JNIEnv *env, jobject objList)
{
    if (objList != nullptr) {
        std::vector<YR::Libruntime::StackTraceInfo> stackTraceInfos = JNIList::FromJava<YR::Libruntime::StackTraceInfo>(
            env, objList, [](JNIEnv *env, jobject obj) -> YR::Libruntime::StackTraceInfo {
                return JNIStackTraceInfo::FromJava(env, obj);
            });
        return stackTraceInfos;
    } else {
        return std::vector<YR::Libruntime::StackTraceInfo>();
    }
}
}  // namespace jni
}  // namespace YR