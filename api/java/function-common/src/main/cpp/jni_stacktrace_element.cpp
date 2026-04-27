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

#include "jni_stacktrace_element.h"
#include "src/utility/logger/logger.h"

namespace YR {
namespace jni {

const std::string STACK_TRACE_ELEMENT_CLASS = "java/lang/StackTraceElement";
const std::string INIT_TAG = "<init>";
const std::string CONSTRUCTOR_DESCRIPTION = "(Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;I)V";
const std::string GET_STRING_DESCRIPTION = "()Ljava/lang/String;";
const std::string GET_INT_DESCRIPTION = "()I";

const std::string GET_CLASS_NAME = "getClassName";
const std::string GET_METHOD_NAME = "getMethodName";
const std::string GET_FILE_NAME = "getFileName";
const std::string GET_LINE_NUMBER = "getLineNumber";

void JNIStackTraceElement::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, STACK_TRACE_ELEMENT_CLASS);
    init_ = GetJMethod(env, clz_, INIT_TAG, CONSTRUCTOR_DESCRIPTION);
    getClassName_ = GetJMethod(env, clz_, GET_CLASS_NAME, GET_STRING_DESCRIPTION);
    getMethodName_ = GetJMethod(env, clz_, GET_METHOD_NAME, GET_STRING_DESCRIPTION);
    getFileName_ = GetJMethod(env, clz_, GET_FILE_NAME, GET_STRING_DESCRIPTION);
    getLineNumber_ = GetJMethod(env, clz_, GET_LINE_NUMBER, GET_INT_DESCRIPTION);
}

void JNIStackTraceElement::Recycle(JNIEnv *env)
{
    if (clz_ != nullptr) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

jobject JNIStackTraceElement::FromCc(JNIEnv *env, const YR::Libruntime::StackTraceElement &stackTraceElement)
{
    jstring jclassName = JNIString::FromCc(env, stackTraceElement.className);
    if (jclassName == nullptr) {
        YRLOG_ERROR("Failed to convert jclassName from Cc code to Java");
        return nullptr;
    }

    jstring jmethodName = JNIString::FromCc(env, stackTraceElement.methodName);
    if (jmethodName == nullptr) {
        YRLOG_ERROR("Failed to convert jmethodName from Cc code to Java");
        return nullptr;
    }

    jstring jfileName = JNIString::FromCc(env, stackTraceElement.fileName);
    if (jfileName == nullptr) {
        YRLOG_ERROR("Failed to convert jfileName from Cc code to Java");
        return nullptr;
    }

    jint lineNumberInt = static_cast<jint>(stackTraceElement.lineNumber);
    if (lineNumberInt <= 0) {
        YRLOG_ERROR("Failed to get valid lineNumberInt from Cc code to Java");
        return nullptr;
    }

    jobject jStackTraceElement = env->NewObject(clz_, init_, jclassName, jmethodName, jfileName, lineNumberInt);
    env->DeleteLocalRef(jclassName);
    env->DeleteLocalRef(jmethodName);
    env->DeleteLocalRef(jfileName);
    return jStackTraceElement;
}

YR::Libruntime::StackTraceElement JNIStackTraceElement::FromJava(JNIEnv *env, jobject obj)
{
    jstring jclassName = static_cast<jstring>(env->CallObjectMethod(obj, getClassName_));
    std::string cclassName_ = JNIString::FromJava(env, jclassName);
    if (cclassName_.empty()) {
        YRLOG_ERROR("Failed to convert jclassName from Java code to Cc");
        return YR::Libruntime::StackTraceElement{};
    }

    jstring jmethodName = static_cast<jstring>(env->CallObjectMethod(obj, getMethodName_));
    std::string cmethodName_ = JNIString::FromJava(env, jmethodName);
    if (cmethodName_.empty()) {
        YRLOG_ERROR("Failed to convert jmethodName from Java code to Cc");
        return YR::Libruntime::StackTraceElement{};
    }

    jstring jfileName = static_cast<jstring>(env->CallObjectMethod(obj, getFileName_));
    std::string cfileName_ = JNIString::FromJava(env, jfileName);
    if (cfileName_.empty()) {
        YRLOG_ERROR("Failed to convert jfileName from Java code to Cc");
        return YR::Libruntime::StackTraceElement{};
    }

    jint jlineNumber = static_cast<jint>(env->CallIntMethod(obj, getLineNumber_));
    if (jlineNumber <= 0) {
        YRLOG_ERROR("Failed to get valid jlineNumber from Java code to Cc");
        return YR::Libruntime::StackTraceElement{};
    }

    YR::Libruntime::StackTraceElement stackTraceElement = {cclassName_, cmethodName_, cfileName_, (int)jlineNumber};
    return stackTraceElement;
}

std::vector<YR::Libruntime::StackTraceElement> JNIStackTraceElement::ListFromJava(JNIEnv *env, jobject objList)
{
    std::vector<YR::Libruntime::StackTraceElement> stackTraceElements =
        JNIList::FromJava<YR::Libruntime::StackTraceElement>(
            env, objList, [](JNIEnv *env, jobject obj) -> YR::Libruntime::StackTraceElement {
                return JNIStackTraceElement::FromJava(env, obj);
            });
    return stackTraceElements;
}

jobject JNIStackTraceElement::ListFromCc(JNIEnv *env, std::vector<YR::Libruntime::StackTraceElement> objs)
{
    auto jstackTraceElements = JNIArrayList::FromCc<YR::Libruntime::StackTraceElement>(
        env, objs, [](JNIEnv *env, const YR::Libruntime::StackTraceElement &element) {
            return JNIStackTraceElement::FromCc(env, element);
        });
    return jstackTraceElements;
}
}  // namespace jni
}  // namespace YR