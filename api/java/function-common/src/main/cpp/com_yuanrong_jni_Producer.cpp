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

#include "com_yuanrong_jni_Producer.h"
#include <jni.h>

#include "jni_types.h"

#include "src/dto/stream_conf.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/streamstore/stream_producer_consumer.h"

#ifdef __cplusplus
extern "C" {
#endif

JNIEXPORT jobject JNICALL Java_com_yuanrong_jni_JniProducer_sendHeapBufferDefaultTimeout(JNIEnv *env, jclass,
                                                                                         jlong handle, jbyteArray bytes,
                                                                                         jlong len)
{
    auto producer = reinterpret_cast<std::shared_ptr<YR::Libruntime::StreamProducer> *>(handle);
    jbyte *bytekey = env->GetByteArrayElements(bytes, 0);
    if (bytekey == nullptr) {
        YR::jni::JNILibruntimeException::ThrowNew(env, "An empty array is passed in the Java layer.");
        return nullptr;
    } else {
        YR::Libruntime::Element element(reinterpret_cast<uint8_t *>(bytekey), len);
        auto err = (*producer)->Send(element);
        env->ReleaseByteArrayElements(bytes, bytekey, 0);
        jobject jerr = YR::jni::JNIErrorInfo::FromCc(env, err);
        if (jerr == nullptr) {
            YR::jni::JNILibruntimeException::ThrowNew(env, "failed to convert jerr when Producer_send, get null");
            return nullptr;
        }
        return jerr;
    }
}

JNIEXPORT jobject JNICALL Java_com_yuanrong_jni_JniProducer_sendDirectBufferDefaultTimeout(JNIEnv *env, jclass,
                                                                                           jlong handle, jobject buf)
{
    auto producer = reinterpret_cast<std::shared_ptr<YR::Libruntime::StreamProducer> *>(handle);
    auto body = env->GetDirectBufferAddress(buf);
    if (!body) {
        YR::jni::JNILibruntimeException::ThrowNew(env, "cannot get element address");
        return nullptr;
    }
    int limit = YR::jni::JNIByteBuffer::GetByteBufferLimit(env, buf);
    YR::Libruntime::Element element(static_cast<uint8_t *>(body), limit);
    auto err = (*producer)->Send(element);
    body = NULL;
    jobject jerr = YR::jni::JNIErrorInfo::FromCc(env, err);
    if (jerr == nullptr) {
        YR::jni::JNILibruntimeException::ThrowNew(env, "failed to convert jerr when Producer_send, get null");
        return nullptr;
    }
    return jerr;
}

JNIEXPORT jobject JNICALL Java_com_yuanrong_jni_JniProducer_sendHeapBuffer(JNIEnv *env, jclass, jlong handle,
                                                                           jbyteArray bytes, jlong len, jint timeoutMs)
{
    auto producer = reinterpret_cast<std::shared_ptr<YR::Libruntime::StreamProducer> *>(handle);
    jbyte *bytekey = env->GetByteArrayElements(bytes, 0);
    if (bytekey == nullptr) {
        YR::jni::JNILibruntimeException::ThrowNew(env, "An empty array is passed in the Java layer.");
        return nullptr;
    } else {
        YR::Libruntime::Element element(reinterpret_cast<uint8_t *>(bytekey), len);
        auto err = (*producer)->Send(element, timeoutMs);
        env->ReleaseByteArrayElements(bytes, bytekey, 0);
        jobject jerr = YR::jni::JNIErrorInfo::FromCc(env, err);
        if (jerr == nullptr) {
            YR::jni::JNILibruntimeException::ThrowNew(env, "failed to convert jerr when Producer_send, get null");
            return nullptr;
        }
        return jerr;
    }
}

JNIEXPORT jobject JNICALL Java_com_yuanrong_jni_JniProducer_sendDirectBuffer(JNIEnv *env, jclass, jlong handle,
                                                                             jobject buf, jint timeoutMs)
{
    auto producer = reinterpret_cast<std::shared_ptr<YR::Libruntime::StreamProducer> *>(handle);
    auto body = env->GetDirectBufferAddress(buf);
    if (!body) {
        YR::jni::JNILibruntimeException::ThrowNew(env, "cannot get element address");
        return nullptr;
    }
    int limit = YR::jni::JNIByteBuffer::GetByteBufferLimit(env, buf);
    YR::Libruntime::Element element(static_cast<uint8_t *>(body), limit);
    auto err = (*producer)->Send(element, timeoutMs);
    body = NULL;
    jobject jerr = YR::jni::JNIErrorInfo::FromCc(env, err);
    if (jerr == nullptr) {
        YR::jni::JNILibruntimeException::ThrowNew(env, "failed to convert jerr when Producer_send, get null");
        return nullptr;
    }
    return jerr;
}

JNIEXPORT jobject JNICALL Java_com_yuanrong_jni_JniProducer_close(JNIEnv *env, jclass, jlong handle)
{
    auto producer = reinterpret_cast<std::shared_ptr<YR::Libruntime::StreamProducer> *>(handle);
    auto err = (*producer)->Close();
    delete (producer);
    jobject jerr = YR::jni::JNIErrorInfo::FromCc(env, err);
    if (jerr == nullptr) {
        YR::jni::JNILibruntimeException::ThrowNew(env, "failed to convert jerr when Producer_close, get null");
        return nullptr;
    }
    return jerr;
}

JNIEXPORT void JNICALL Java_com_yuanrong_jni_JniProducer_freeJNIPtrNative(JNIEnv *, jclass, jlong handle)
{
    auto producer = reinterpret_cast<std::shared_ptr<YR::Libruntime::StreamProducer> *>(handle);
    delete producer;
}

#ifdef __cplusplus
}
#endif
