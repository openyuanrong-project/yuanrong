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

#include "jni_init.h"
#include "jni_types.h"

JavaVM *jvm;

/// Load and cache frequently-used Java classes and methods
jint JNI_OnLoad(JavaVM *vm, void *reserved)
{
    JNIEnv *env;
    if (vm->GetEnv(reinterpret_cast<void **>(&env), CURRENT_JNI_VERSION) != JNI_OK) {
        return JNI_ERR;
    }

    jvm = vm;

    // JNIApacheCommonsExceptionUtils depends on JNIString
    // Rest of them depends on JNIApacheCommonsExceptionUtils for any exception in JVM
    bool isOnCloud = true;
    jclass tempLocalClassRef = env->FindClass("org/yuanrong/runtime/server/RuntimeServer");
    if (tempLocalClassRef == nullptr) {
        isOnCloud = false;
        env->ExceptionClear();
        env->DeleteLocalRef(tempLocalClassRef);
    } else {
        env->DeleteLocalRef(tempLocalClassRef);
    }
    YR::jni::JNIString::Init(env);
    YR::jni::JNIApacheCommonsExceptionUtils::Init(env);
    YR::jni::JNIList::Init(env);
    YR::jni::JNIArrayList::Init(env);
    YR::jni::JNIByteBuffer::Init(env);
    YR::jni::JNIFloat::Init(env);
    YR::jni::JNIIterator::Init(env);
    YR::jni::JNIMapEntry::Init(env);
    YR::jni::JNISet::Init(env);
    YR::jni::JNIMap::Init(env);
    YR::jni::JNIHashMap::Init(env);
    YR::jni::JNIApiType::Init(env);
    YR::jni::JNILanguageType::Init(env);
    YR::jni::JNIInvokeType::Init(env);
    YR::jni::JNILabelOperator::Init(env);
    YR::jni::JNIAffinity::Init(env);
    YR::jni::JNIFunctionMeta::Init(env);
    YR::jni::JNILibRuntimeConfig::Init(env);
    YR::jni::JNIInvokeArg::Init(env);
    YR::jni::JNIInvokeOptions::Init(env);
    YR::jni::JNIGroupOptions::Init(env);
    YR::jni::JNIDataObject::Init(env);
    YR::jni::JNIErrorInfo::Init(env);
    YR::jni::JNIErrorCode::Init(env);
    YR::jni::JNIModuleCode::Init(env);
    YR::jni::JNIUnorderedSet::Init(env);
    YR::jni::JNILibruntimeException::Init(env);
    YR::jni::JNIExistenceOpt::Init(env);
    YR::jni::JNIWriteMode::Init(env);
    YR::jni::JNIConsistencyType::Init(env);
    YR::jni::JNICacheType::Init(env);
    YR::jni::JNISetParam::Init(env);
    YR::jni::JNIMSetParam::Init(env);
    YR::jni::JNICreateParam::Init(env);
    YR::jni::JNIGetParam::Init(env);
    YR::jni::JNIGetParams::Init(env);
    YR::jni::JNIInternalWaitResult::Init(env);
    YR::jni::JNIPair::Init(env);
    YR::jni::JNIStackTraceInfo::Init(env);
    YR::jni::JNIStackTraceElement::Init(env);
    YR::jni::JNIYRAutoInitInfo::Init(env);
    YR::jni::JNIFunctionLog::Init(env);
    YR::jni::JNIProducerConfig::Init(env);
    YR::jni::JNINode::Init(env);
    if (isOnCloud) {
        YR::jni::JNIReturnType::Init(env);
        YR::jni::JNICodeLoader::Init(env);
        YR::jni::JNICodeExecutor::Init(env);
    }
    if (env->ExceptionOccurred()) {
        std::abort();
    }
    return CURRENT_JNI_VERSION;
}

/// Load and cache frequently-used Java classes and methods
void JNI_OnUnload(JavaVM *vm, void *reserved)
{
    JNIEnv *env;
    vm->GetEnv(reinterpret_cast<void **>(&env), CURRENT_JNI_VERSION);

    YR::jni::JNIString::Recycle(env);
    YR::jni::JNIApacheCommonsExceptionUtils::Recycle(env);
    YR::jni::JNIList::Recycle(env);
    YR::jni::JNIFloat::Recycle(env);
    YR::jni::JNIArrayList::Recycle(env);
    YR::jni::JNIByteBuffer::Recycle(env);
    YR::jni::JNIIterator::Recycle(env);
    YR::jni::JNIMapEntry::Recycle(env);
    YR::jni::JNISet::Recycle(env);
    YR::jni::JNIMap::Recycle(env);
    YR::jni::JNIHashMap::Recycle(env);
    YR::jni::JNIInvokeType::Recycle(env);
    YR::jni::JNIApiType::Recycle(env);
    YR::jni::JNILanguageType::Recycle(env);
    YR::jni::JNILabelOperator::Recycle(env);
    YR::jni::JNIAffinity::Recycle(env);
    YR::jni::JNIFunctionMeta::Recycle(env);
    YR::jni::JNICodeLoader::Recycle(env);
    YR::jni::JNICodeExecutor::Recycle(env);
    YR::jni::JNIInvokeArg::Recycle(env);
    YR::jni::JNIInvokeOptions::Recycle(env);
    YR::jni::JNIGroupOptions::Recycle(env);
    YR::jni::JNIDataObject::Recycle(env);
    YR::jni::JNIErrorInfo::Recycle(env);
    YR::jni::JNILibRuntimeConfig::Recycle(env);
    YR::jni::JNIErrorCode::Recycle(env);
    YR::jni::JNIModuleCode::Recycle(env);
    YR::jni::JNIReturnType::Recycle(env);
    YR::jni::JNIUnorderedSet::Recycle(env);
    YR::jni::JNIExistenceOpt::Recycle(env);
    YR::jni::JNIWriteMode::Recycle(env);
    YR::jni::JNIConsistencyType::Recycle(env);
    YR::jni::JNICacheType::Recycle(env);
    YR::jni::JNISetParam::Recycle(env);
    YR::jni::JNIMSetParam::Recycle(env);
    YR::jni::JNICreateParam::Recycle(env);
    YR::jni::JNIGetParam::Recycle(env);
    YR::jni::JNIGetParams::Recycle(env);
    YR::jni::JNIInternalWaitResult::Recycle(env);
    YR::jni::JNIPair::Recycle(env);
    YR::jni::JNIStackTraceInfo::Recycle(env);
    YR::jni::JNIStackTraceElement::Recycle(env);
    YR::jni::JNIYRAutoInitInfo::Recycle(env);
    YR::jni::JNIFunctionLog::Recycle(env);
    YR::jni::JNIProducerConfig::Recycle(env);
    YR::jni::JNINode::Recycle(env);
}