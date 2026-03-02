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

#include <jni.h>
/* Header for class org_yuanrong_jni_LibRuntime */

#pragma once

#include <string>

#ifdef __cplusplus
extern "C" {
#endif

static std::string get_runtime_context_callback(JNIEnv *env, jclass c);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    Init
 * Signature: (Lorg/yuanrong/jni/LibRuntimeConfig;)Lorg/yuanrong/errorcode/ErrorInfo;
 */
JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_LibRuntime_Init(JNIEnv *, jclass, jobject);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    CreateInstance
 * Signature:
 * (Lorg/yuanrong/libruntime/generated/Libruntime$FunctionMeta;Ljava/util/List;Lorg/yuanrong/InvokeOptions;)Lorg/yuanrong/errorcode/Pair;
 */
JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_LibRuntime_CreateInstance(JNIEnv *, jclass, jobject, jobject,
                                                                                 jobject);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    InvokeInstance
 * Signature:
 * (Lorg/yuanrong/libruntime/generated/Libruntime$FunctionMeta;Ljava/lang/String;Ljava/util/List;Lorg/yuanrong/InvokeOptions;)Lorg/yuanrong/errorcode/Pair;
 */
JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_LibRuntime_InvokeInstance(JNIEnv *, jclass, jobject, jstring,
                                                                                 jobject, jobject);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    Get
 * Signature: (Ljava/util/List;IZ)Lorg/yuanrong/errorcode/Pair;
 */
JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_LibRuntime_Get(JNIEnv *, jclass, jobject, jint, jboolean);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    Put
 * Signature: ([BLjava/util/List;)Lorg/yuanrong/errorcode/Pair;
 */
JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_LibRuntime_Put(JNIEnv *, jclass, jbyteArray, jobject);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    PutWithParam
 * Signature: ([BLjava/util/List;Lorg/yuanrong/Createparam;)Lorg/yuanrong/errorcode/Pair;
 */
JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_LibRuntime_PutWithParam(JNIEnv *, jclass, jbyteArray, jobject,
                                                                               jobject);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    Wait
 * Signature:(Ljava/util/List;II)Lorg/yuanrong/storage/InternalWaitResult;
 */
JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_LibRuntime_Wait(JNIEnv *, jclass, jobject, jint, jint);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    DecreaseReference
 * Signature: ()V
 */
JNIEXPORT void JNICALL Java_org_yuanrong_jni_LibRuntime_DecreaseReference(JNIEnv *, jclass, jobject);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    receiveRequestLoop
 * Signature: ()V
 */
JNIEXPORT void JNICALL Java_org_yuanrong_jni_LibRuntime_ReceiveRequestLoop(JNIEnv *, jclass);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    Kill
 * Signature: (Ljava/lang/String;)Lorg/yuanrong/errorcode/ErrorInfo;
 */
JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_LibRuntime_Kill(JNIEnv *, jclass, jstring);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    FinalizeWithCtx
 * Signature: (Ljava/lang/String;)V
 */
JNIEXPORT void JNICALL Java_org_yuanrong_jni_LibRuntime_FinalizeWithCtx(JNIEnv *, jclass, jstring);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    Finalize
 * Signature: ()V
 */
JNIEXPORT void JNICALL Java_org_yuanrong_jni_LibRuntime_Finalize(JNIEnv *, jclass);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    Exit
 * Signature: ()V
 */
JNIEXPORT void JNICALL Java_org_yuanrong_jni_LibRuntime_Exit(JNIEnv *, jclass);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    AutoInitYR
 * Signature: ()V
 */
JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_LibRuntime_AutoInitYR(JNIEnv *, jclass, jobject);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    IsInitialized
 * Signature: ()Z
 */
JNIEXPORT jboolean JNICALL Java_org_yuanrong_jni_LibRuntime_IsInitialized(JNIEnv *, jclass);

/*
 * Class:     org_yuanrong_jni_NativeLibRuntime
 * Method:    setRuntimeContext
 * Signature: (Ljava/lang/String;)V
 */
JNIEXPORT void JNICALL Java_org_yuanrong_jni_LibRuntime_setRuntimeContext(JNIEnv *, jclass, jstring);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    GetRealInstanceId
 * Signature: (Ljava/lang/String;)Ljava/lang/String;
 */
JNIEXPORT jstring JNICALL Java_org_yuanrong_jni_LibRuntime_GetRealInstanceId(JNIEnv *, jclass, jstring);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    SaveRealInstanceId
 * Signature: (Ljava/lang/String;Ljava/lang/String;Lorg/yuanrong/InvokeOptions;)V
 */
JNIEXPORT void JNICALL Java_org_yuanrong_jni_LibRuntime_SaveRealInstanceId(JNIEnv *, jclass, jstring, jstring,
                                                                                  jobject);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    KVWrite
 * Signature: (Ljava/lang/String;[BLorg/yuanrong/ExistenceOpt;)Lorg/yuanrong/errorcode/ErrorInfo;
 */
JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_LibRuntime_KVWrite(JNIEnv *, jclass, jstring, jbyteArray,
                                                                          jobject);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    KVRead
 * Signature: (Ljava/lang/String;I)Lorg/yuanrong/errorcode/Pair;
 */
JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_LibRuntime_KVRead__Ljava_lang_String_2I(JNIEnv *, jclass,
                                                                                               jstring, jint);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    KVRead
 * Signature: (Ljava/util/List;IZ)Lorg/yuanrong/errorcode/Pair;
 */
JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_LibRuntime_KVRead__Ljava_util_List_2IZ(JNIEnv *, jclass, jobject,
                                                                                              jint, jboolean);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    KVGetWithParam
 * Signature: (Ljava/util/List;Lorg/yuanrong/Getparams;I)Lorg/yuanrong/errorcode/Pair;
 */
JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_LibRuntime_KVGetWithParam(JNIEnv *env, jclass c, jobject keys,
                                                                                 jobject getParams, jint timeoutMS);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    KVDel
 * Signature: (Ljava/lang/String;)Lorg/yuanrong/errorcode/ErrorInfo;
 */
JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_LibRuntime_KVDel__Ljava_lang_String_2(JNIEnv *, jclass, jstring);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    KVDel
 * Signature: (Ljava/util/List;)Lorg/yuanrong/errorcode/Pair;
 */
JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_LibRuntime_KVDel__Ljava_util_List_2(JNIEnv *, jclass, jobject);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    SaveState
 * Signature: (I;)Lorg/yuanrong/errorcode/ErrorInfo;
 */
JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_LibRuntime_SaveState(JNIEnv *, jclass, jint);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    loadState
 * Signature: (I;)Lorg/yuanrong/errorcode/ErrorInfo;
 */
JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_LibRuntime_LoadState(JNIEnv *, jclass, jint);

JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_LibRuntime_GroupCreate(JNIEnv *, jclass, jstring, jobject);

JNIEXPORT void JNICALL Java_org_yuanrong_jni_LibRuntime_GroupTerminate(JNIEnv *, jclass, jstring);

JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_LibRuntime_GroupWait(JNIEnv *, jclass, jstring);

JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_LibRuntime_processLog(JNIEnv *, jclass, jobject);

JNIEXPORT jlong JNICALL Java_org_yuanrong_jni_LibRuntime_CreateStreamProducer(JNIEnv *, jclass, jstring, jlong,
                                                                                     jlong, jlong, jboolean, jboolean,
                                                                                     jlong, jlong);

JNIEXPORT jlong JNICALL Java_org_yuanrong_jni_LibRuntime_CreateStreamConsumer(JNIEnv *, jclass, jstring, jstring,
                                                                                     jobject, jboolean);

JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_LibRuntime_DeleteStream(JNIEnv *, jclass, jstring);

JNIEXPORT jlong JNICALL Java_org_yuanrong_jni_LibRuntime_QueryGlobalProducersNum(JNIEnv *, jclass, jstring);

JNIEXPORT jlong JNICALL Java_org_yuanrong_jni_LibRuntime_QueryGlobalConsumersNum(JNIEnv *, jclass, jstring);

JNIEXPORT jlong JNICALL Java_org_yuanrong_jni_LibRuntime_CreateStreamProducerWithConfig(JNIEnv *, jclass,
                                                                                               jstring, jobject);
/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    GetInstanceRoute
 * Signature: (Ljava/lang/String;)Ljava/lang/String;
 */
JNIEXPORT jstring JNICALL Java_org_yuanrong_jni_LibRuntime_GetInstanceRoute(JNIEnv *, jclass, jstring);

/*
 * Class:     org_yuanrong_jni_LibRuntime
 * Method:    SaveInstanceRoute
 * Signature: (Ljava/lang/String;Ljava/lang/String;)V
 */
JNIEXPORT void JNICALL Java_org_yuanrong_jni_LibRuntime_SaveInstanceRoute(JNIEnv *, jclass, jstring, jstring);

JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_LibRuntime_KillSync(JNIEnv *, jclass, jstring);

JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_LibRuntime_streamWrite(JNIEnv *, jclass, jstring,
                                                                       jstring jrequestId, jstring jinstanceId);

JNIEXPORT jobject JNICALL Java_org_yuanrong_jni_LibRuntime_getRequestAndInstanceID(JNIEnv *, jclass);

#ifdef __cplusplus
}
#endif
