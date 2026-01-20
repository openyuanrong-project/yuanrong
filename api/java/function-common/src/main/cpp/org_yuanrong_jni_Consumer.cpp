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

#include <jni.h>
#include "org_yuanrong_jni_Consumer.h"

#include "jni_types.h"

#include "src/dto/stream_conf.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/streamstore/stream_producer_consumer.h"

#ifdef __cplusplus
extern "C" {
#endif

using YR::Libruntime::ErrorInfo;

JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_JniConsumer_receive(JNIEnv *env, jclass, jlong consumerPtr,
                                                                           jlong expectNum, jint timeoutMs,
                                                                           jboolean hasExpectedNum)
{
    auto consumer = reinterpret_cast<std::shared_ptr<YR::Libruntime::StreamConsumer> *>(consumerPtr);
    std::vector<YR::Libruntime::Element> elements;
    ErrorInfo err;
    if (hasExpectedNum == JNI_TRUE) {
        err = (*consumer)->Receive(expectNum, timeoutMs, elements);
    } else {
        err = (*consumer)->Receive(timeoutMs, elements);
    }

    jobject jerr = YR::jni::JNIErrorInfo::FromCc(env, err);
    if (jerr == nullptr) {
        YR::jni::JNILibruntimeException::ThrowNew(env, "failed to convert jerr when Consumer_receive, get null");
        return nullptr;
    }
    jclass arrayListClass = env->FindClass("java/util/ArrayList");
    jobject elementList = env->NewObject(arrayListClass, env->GetMethodID(arrayListClass, "<init>", "()V"));
    jmethodID elementAdd = env->GetMethodID(arrayListClass, "add", "(Ljava/lang/Object;)Z");

    jclass elementClass = env->FindClass("org/yuanrong/stream/Element");
    jmethodID elementInit = env->GetMethodID(elementClass, "<init>", "(JLjava/nio/ByteBuffer;)V");
    jobject elementBuffer = NULL;
    jobject elementObject = NULL;
    for (const auto &element : elements) {
        elementBuffer = env->NewDirectByteBuffer(element.ptr, element.size);
        elementObject = env->NewObject(elementClass, elementInit, element.id, elementBuffer);
        env->CallBooleanMethod(elementList, elementAdd, elementObject);
    }
    // Clean up
    env->DeleteLocalRef(arrayListClass);
    env->DeleteLocalRef(elementClass);
    env->DeleteLocalRef(elementBuffer);
    env->DeleteLocalRef(elementObject);
    return YR::jni::JNIPair::CreateJPair(env, jerr, elementList);
}

JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_JniConsumer_ack(JNIEnv *env, jclass, jlong consumerPtr,
                                                                       jlong elementId)
{
    auto consumer = reinterpret_cast<std::shared_ptr<YR::Libruntime::StreamConsumer> *>(consumerPtr);
    auto err = (*consumer)->Ack(static_cast<uint64_t>(elementId));
    jobject jerr = YR::jni::JNIErrorInfo::FromCc(env, err);
    if (jerr == nullptr) {
        YR::jni::JNILibruntimeException::ThrowNew(env, "failed to convert jerr when Consumer_ack, get null");
        return nullptr;
    }
    return jerr;
}

JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_JniConsumer_close(JNIEnv *env, jclass, jlong consumerPtr)
{
    auto consumer = reinterpret_cast<std::shared_ptr<YR::Libruntime::StreamConsumer> *>(consumerPtr);
    auto err = (*consumer)->Close();
    jobject jerr = YR::jni::JNIErrorInfo::FromCc(env, err);
    if (jerr == nullptr) {
        YR::jni::JNILibruntimeException::ThrowNew(env, "failed to convert jerr when Consumer_close, get null");
        return nullptr;
    }
    delete consumer;
    return jerr;
}

JNIEXPORT void JNICALL Java_org_datasystem_streamcache_ConsumerImpl_freeJNIPtrNative(JNIEnv *, jclass, jlong handle)
{
    auto consumer = reinterpret_cast<std::shared_ptr<YR::Libruntime::StreamConsumer> *>(handle);
    delete consumer;
}

#ifdef __cplusplus
}
#endif
