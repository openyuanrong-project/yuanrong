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

#pragma once
#ifdef __cplusplus
extern "C" {
#endif

JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_JniConsumer_receive(JNIEnv *, jclass, jlong, jlong, jint,
                                                                           jboolean);

JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_JniConsumer_ack(JNIEnv *, jclass, jlong, jlong);

JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_JniConsumer_close(JNIEnv *, jclass, jlong);

JNIEXPORT void JNICALL Java_org_yuanrong_jni_JniConsumer_freeJNIPtrNative(JNIEnv *, jclass, jlong);

#ifdef __cplusplus
}
#endif
