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

#include "jni_errorinfo.h"
#include "jni_types.h"
#include "src/utility/logger/logger.h"

namespace YR {
namespace jni {
void JNIErrorInfo::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "org/yuanrong/errorcode/ErrorInfo");
    init_ = GetJMethod(env, clz_, "<init>",
                       "(Lorg/yuanrong/errorcode/ErrorCode;Lorg/yuanrong/errorcode/ModuleCode;Ljava/lang/"
                       "String;Ljava/util/List;)V");
    getMsg_ = GetJMethod(env, clz_, "getErrorMessage", "()Ljava/lang/String;");
    getCode_ = GetJMethod(env, clz_, "getErrorCode", "()Lorg/yuanrong/errorcode/ErrorCode;");
    getMCode_ = GetJMethod(env, clz_, "getModuleCode", "()Lorg/yuanrong/errorcode/ModuleCode;");
    getStackTraceInfos_ = GetJMethod(env, clz_, "getStackTraceInfos", "()Ljava/util/List;");
}

void JNIErrorInfo::Recycle(JNIEnv *env)
{
    if (clz_ != nullptr) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

jobject JNIErrorInfo::FromCc(JNIEnv *env, const YR::Libruntime::ErrorInfo &errorInfo)
{
    jstring jmsg = JNIString::FromCc(env, errorInfo.Msg());
    if (jmsg == nullptr) {
        YRLOG_ERROR("Failed to convert jmsg from Cc code to Java");
        return nullptr;
    }

    jobject jerrorCode = JNIErrorCode::FromCc(env, errorInfo.Code());
    if (jerrorCode == nullptr) {
        YRLOG_ERROR("Failed to convert jerrorCode from Cc code to Java");
        return nullptr;
    }

    jobject jmoduleCode = JNIModuleCode::FromCc(env, errorInfo.MCode());
    if (jmoduleCode == nullptr) {
        YRLOG_ERROR("Failed to convert jmoduleCode from Cc code to Java");
        return nullptr;
    }

    jobject jstackTraceInfos = JNIStackTraceInfo::ListFromCc(env, errorInfo.GetStackTraceInfos());
    jobject jerrorInfo = (jobject)env->NewObject(clz_, init_, jerrorCode, jmoduleCode, jmsg, jstackTraceInfos);

    env->DeleteLocalRef(jmsg);
    env->DeleteLocalRef(jerrorCode);
    env->DeleteLocalRef(jmoduleCode);
    env->DeleteLocalRef(jstackTraceInfos);
    return jerrorInfo;
}

YR::Libruntime::ErrorInfo JNIErrorInfo::FromJava(JNIEnv *env, jobject o)
{
    jstring jstr = static_cast<jstring>(env->CallObjectMethod(o, getMsg_));
    std::string cmsg = JNIString::FromJava(env, jstr);
    env->DeleteLocalRef(jstr);

    jobject errorCodeObj = env->CallObjectMethod(o, getCode_);
    YR::Libruntime::ErrorCode errorCode = JNIErrorCode::FromJava(env, errorCodeObj);
    env->DeleteLocalRef(errorCodeObj);

    jobject moduleCodeObj = env->CallObjectMethod(o, getMCode_);
    YR::Libruntime::ModuleCode moduleCode = JNIModuleCode::FromJava(env, moduleCodeObj);
    env->DeleteLocalRef(moduleCodeObj);

    jobject objList = env->CallObjectMethod(o, getStackTraceInfos_);
    std::vector<YR::Libruntime::StackTraceInfo> stackTraceInfos = JNIStackTraceInfo::ListFromJava(env, objList);
    env->DeleteLocalRef(objList);

    YR::Libruntime::ErrorInfo errorInfo;
    if (!stackTraceInfos.empty()) {
        errorInfo = YR::Libruntime::ErrorInfo(errorCode, moduleCode, cmsg, stackTraceInfos);
    } else {
        errorInfo = YR::Libruntime::ErrorInfo(errorCode, moduleCode, cmsg);
    }
    return errorInfo;
}
}  // namespace jni
}  // namespace YR
