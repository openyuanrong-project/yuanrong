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

#include "jni_types.h"

#include <jni.h>
#include <cassert>
#include <cstddef>
#include <cstdint>
#include <functional>
#include <iostream>
#include <map>
#include <memory>
#include <msgpack.hpp>
#include <ostream>
#include <sstream>
#include <string>
#include <thread>
#include <unordered_set>
#include <vector>

#include "src/libruntime/err_type.h"
#include "src/libruntime/libruntime_manager.h"
#include "src/libruntime/libruntime_options.h"
#include "src/proto/libruntime.pb.h"
#include "src/utility/logger/logger.h"

namespace YR {
namespace jni {
const int LABEL_IN = 1;
const int LABEL_NOT_IN = 2;
const int LABEL_EXISTS = 3;
const int LABEL_DOES_NOT_EXIST = 4;
const int RESOURCE_PREFERRED = 11;
const int RESOURCE_PREFERRED_ANTI = 12;
const int RESOURCE_REQUIRED = 13;
const int RESOURCE_REQUIRED_ANTI = 14;
const int INSTANCE_PREFERRED = 21;
const int INSTANCE_PREFERRED_ANTI = 22;
const int INSTANCE_REQUIRED = 23;
const int INSTANCE_REQUIRED_ANTI = 24;
const int MAX_PASSWD_LENGTH = 100;
const int POD = 1;
const int NODE = 2;

inline jfieldID GetJStaticField(JNIEnv *env, const jclass &clz, const std::string &fieldName, const std::string &sig)
{
    jfieldID f = env->GetStaticFieldID(clz, fieldName.c_str(), sig.c_str());
    LOG_IF_JAVA_OBJECT_IS_NULL(f, "Failed to load " + fieldName + " { " + sig + " }");
    return f;
}

inline jobject GetJStaticObjectField(JNIEnv *env, const jclass &clz, const jfieldID &jfId)
{
    jobject o = env->GetStaticObjectField(clz, jfId);
    LOG_IF_JAVA_OBJECT_IS_NULL(o, "Failed to load static object");
    return o;
}

inline jobject GetJObjectField(JNIEnv *env, const jobject &obj, const jfieldID &jfId)
{
    jobject o = env->GetObjectField(obj, jfId);
    LOG_IF_JAVA_OBJECT_IS_NULL(o, "Failed to load object field");
    return o;
}

RAIIJavaObject::RAIIJavaObject(JNIEnv *env, jobject o) : env_(env)
{
    obj_ = o;
}

jobject RAIIJavaObject::GetJObject()
{
    return obj_;
}

RAIIJavaObject::~RAIIJavaObject()
{
    if (env_ && obj_) {
        env_->DeleteLocalRef(obj_);
    }
}

void JNILibruntimeException::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/exception/LibRuntimeException");
}

void JNILibruntimeException::Recycle(JNIEnv *env) {}

void JNILibruntimeException::ThrowNew(JNIEnv *env, const std::string &msg)
{
    env->ThrowNew(clz_, msg.c_str());
}

void JNILibruntimeException::Throw(JNIEnv *env, const YR::Libruntime::ErrorCode &errorCode,
                                   const YR::Libruntime::ModuleCode &moduleCode, const std::string &msg)
{
    jmethodID constructorId = env->GetMethodID(
        clz_, "<init>", "(Lcom/yuanrong/errorcode/ErrorCode;Lcom/yuanrong/errorcode/ModuleCode;Ljava/lang/String;)V");

    jobject jerrorCode = JNIErrorCode::FromCc(env, errorCode);
    if (jerrorCode == nullptr) {
        YRLOG_WARN("Failed to convert jerrorCode from Cc code to Java");
        jerrorCode = JNIErrorCode::FromCc(env, YR::Libruntime::ErrorCode::ERR_PARAM_INVALID);
    }

    jobject jmoduleCode = JNIModuleCode::FromCc(env, moduleCode);
    if (jmoduleCode == nullptr) {
        YRLOG_WARN("Failed to convert jmoduleCode from Cc code to Java");
        jmoduleCode = JNIModuleCode::FromCc(env, YR::Libruntime::ModuleCode::RUNTIME);
    }

    jstring jmessage = JNIString::FromCc(env, msg);
    if (jmessage == nullptr) {
        YRLOG_WARN("Failed to convert jmessage from Cc code to Java");
    }

    jthrowable exception = (jthrowable)env->NewObject(clz_, constructorId, jerrorCode, jmoduleCode, jmessage);
    env->Throw(exception);
    env->DeleteLocalRef(jerrorCode);
    env->DeleteLocalRef(jmoduleCode);
    env->DeleteLocalRef(jmessage);
}

void JNIString::Init(JNIEnv *env) {}

void JNIString::Recycle(JNIEnv *env) {}

std::string JNIString::FromJava(JNIEnv *env, jstring jstr)
{
    RETURN_IF_NULL(jstr, "");
    const char *cstr = env->GetStringUTFChars(jstr, nullptr);
    std::string result(cstr);
    env->ReleaseStringUTFChars(jstr, cstr);
    return result;
}

const char *JNIString::FromJavaToCharArray(JNIEnv *env, jstring jstr)
{
    RETURN_IF_NULL(jstr, "");
    const char *cstr = env->GetStringUTFChars(jstr, nullptr);
    return cstr;
}

jstring JNIString::FromCc(JNIEnv *env, const std::string &str)
{
    jstring s = env->NewStringUTF(str.c_str());
    return s;
}

jobject JNIString::FromCcVectorToList(JNIEnv *env, const std::vector<std::string> &vector)
{
    return YR::jni::JNIArrayList::FromCc<std::string>(
        env, vector, [](JNIEnv *env, const std::string &s) { return env->NewStringUTF(s.c_str()); });
}

std::unordered_set<std::string> JNIString::FromJArrayToUnorderedSet(JNIEnv *env, const jobject &objs)
{
    return YR::jni::JNIUnorderedSet::FromJava<std::string>(env, objs, [](JNIEnv *e, jobject obj) -> std::string {
        return YR::jni::JNIString::FromJava(e, static_cast<jstring>(obj));
    });
}

void JNIFloat::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "java/lang/Float");
    jmInit_ = env->GetMethodID(clz_, "<init>", "(F)V");
}

void JNIFloat::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
    }
}

jobject JNIFloat::FromCc(JNIEnv *env, const float &arg)
{
    float val = arg;
    return env->NewObject(clz_, jmInit_, val);
}

void JNIList::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "java/util/List");

    jmSize_ = GetJMethod(env, clz_, "size", "()I");
    jmGet_ = GetJMethod(env, clz_, "get", "(I)Ljava/lang/Object;");
    jmAdd_ = GetJMethod(env, clz_, "add", "(Ljava/lang/Object;)Z");
}

void JNIList::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

jobject JNIList::Get(JNIEnv *env, jobject inst, int idx)
{
    jobject ret = env->CallObjectMethod(inst, jmGet_, (jint)idx);
    return ret;
}

int JNIList::Size(JNIEnv *env, jobject inst)
{
    RETURN_IF_NULL(inst, 0);
    int s = static_cast<int>(env->CallIntMethod(inst, jmSize_));
    return s;
}

void JNIList::Add(JNIEnv *env, jobject inst, jobject ele)
{
    RETURN_VOID_IF_NULL(inst);
    env->CallVoidMethod(inst, jmAdd_, ele);
}

void JNIArrayList::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "java/util/ArrayList");
    jmInit_ = GetJMethod(env, clz_, "<init>", "()V");
    jmInitWithCapacity_ = GetJMethod(env, clz_, "<init>", "(I)V");
}

void JNIArrayList::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

void JNIApacheCommonsExceptionUtils::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "org/apache/commons/lang3/exception/ExceptionUtils");
    jmGetStackTrace_ = GetStaticMethodID(env, clz_, "getStackTrace", "(Ljava/lang/Throwable;)Ljava/lang/String;");
}

void JNIApacheCommonsExceptionUtils::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

std::string JNIApacheCommonsExceptionUtils::GetStackTrace(JNIEnv *env, jthrowable throwable)
{
    if (clz_ == nullptr || jmGetStackTrace_ == nullptr) {
        YRLOG_ERROR("failed to get stack trace, since ExceptionUtils is not init, or init failed");
        return "";
    }
    env->ExceptionClear();
    auto jst = static_cast<jstring>(env->CallStaticObjectMethod(clz_, jmGetStackTrace_, throwable));
    if (env->ExceptionOccurred()) {
        YRLOG_ERROR("Exception occurred when trying to get exception information.");
        return "exception occurred when trying to get exception information";
    }
    RETURN_IF_NULL(jst, "failed to get stacktrace when exception occurred");
    const char *cstr = env->GetStringUTFChars(jst, nullptr);
    std::string result(cstr);
    env->ReleaseStringUTFChars(jst, cstr);
    if (env->ExceptionOccurred()) {
        YRLOG_ERROR("Exception occurred when convert exception info to C string.");
        env->DeleteLocalRef(jst);
        return "exception occurred when convert exception info to C string";
    }
    env->DeleteLocalRef(jst);
    return result;
}

// Between ByteBuffer and YR::Libruntime::Buffer
void JNIByteBuffer::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "java/nio/ByteBuffer");
    jmClear_ = GetJMethod(env, clz_, "clear", "()Ljava/nio/Buffer;");
}

void JNIByteBuffer::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

jobject JNIByteBuffer::FromCc(JNIEnv *env, const std::shared_ptr<YR::Libruntime::DataObject> &dataObj)
{
    jobject obj = env->NewDirectByteBuffer(dataObj->data->MutableData(), dataObj->data->GetSize());
    return obj;
}

jobject JNIByteBuffer::FromCcPrtVectorToList(JNIEnv *env,
                                             const std::vector<std::shared_ptr<YR::Libruntime::Buffer>> &vector)
{
    return YR::jni::JNIArrayList::FromCc<std::shared_ptr<YR::Libruntime::Buffer>>(
        env, vector, [](JNIEnv *env, const std::shared_ptr<YR::Libruntime::Buffer> &sbuf_ptr) {
            jbyteArray element = nullptr;
            if (sbuf_ptr != nullptr) {
                element = env->NewByteArray(sbuf_ptr->GetSize());
                env->SetByteArrayRegion(element, 0, sbuf_ptr->GetSize(),
                                        reinterpret_cast<const jbyte *>(sbuf_ptr->ImmutableData()));
            }
            return element;
        });
}

std::shared_ptr<YR::Libruntime::Buffer> JNIByteBuffer::FromJava(JNIEnv *env, const jbyteArray &byteArray)
{
    ASSERT_NOT_NULL(byteArray);
    // Convert Java byte array to C++ char array
    size_t length = env->GetArrayLength(byteArray);
    jbyte *native_bytes = env->GetByteArrayElements(byteArray, nullptr);
    // Convert C++ char array to Buffer

    auto buf = std::make_shared<YR::Libruntime::NativeBuffer>(length);
    buf->MemoryCopy(reinterpret_cast<void *>(native_bytes), length);

    // Release Java byte array
    env->ReleaseByteArrayElements(byteArray, native_bytes, JNI_ABORT);
    env->DeleteLocalRef(byteArray);
    return buf;
}

int JNIByteBuffer::GetByteBufferLimit(JNIEnv *env, jobject buffer)
{
    jclass bufferJC = env->GetObjectClass(buffer);
    jmethodID limitId = env->GetMethodID(bufferJC, "limit", "()I");
    if (limitId == nullptr) {
        std::string msg = "Failed to get the field ID of the limit member variable of the ByteBuffer class.";
        YR::jni::JNILibruntimeException::ThrowNew(env, msg);
        return 0;
    } else {
        return env->CallIntMethod(buffer, limitId);
    }
}

void JNIByteBuffer::WriteByteArray(JNIEnv *env, std::shared_ptr<YR::Libruntime::Buffer> &sb, const jbyteArray &byteBfr)
{
    RETURN_VOID_IF_NULL(byteBfr);
    jbyte *native_bytes = env->GetByteArrayElements(byteBfr, nullptr);
    size_t capacity = env->GetArrayLength(byteBfr);
    sb = std::make_shared<YR::Libruntime::NativeBuffer>(capacity);
    sb->MemoryCopy((const void *)native_bytes, capacity);
    env->ReleaseByteArrayElements(byteBfr, native_bytes, JNI_ABORT);
    env->DeleteLocalRef(byteBfr);
}

void JNIByteBuffer::Clear(JNIEnv *env, const jobject &byteBfr)
{
    RETURN_VOID_IF_NULL(byteBfr);
    env->CallObjectMethod(byteBfr, jmClear_);
}

void JNIInvokeType::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/libruntime/generated/Libruntime$InvokeType");
    jmGetNumber_ = GetJMethod(env, clz_, "getNumber", "()I");
    jmForNumber_ =
        GetStaticMethodID(env, clz_, "forNumber", "(I)Lcom/yuanrong/libruntime/generated/Libruntime$InvokeType;");
}

void JNIInvokeType::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

jobject JNIInvokeType::FromCc(JNIEnv *env, const libruntime::InvokeType &type)
{
    return env->CallStaticObjectMethod(clz_, jmForNumber_, static_cast<jint>(type));
}

libruntime::InvokeType JNIInvokeType::FromJava(JNIEnv *env, const jobject obj)
{
    return static_cast<libruntime::InvokeType>(env->CallIntMethod(obj, jmGetNumber_));
}

void JNILanguageType::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/libruntime/generated/Libruntime$LanguageType");
    jmGetNumber_ = GetJMethod(env, clz_, "getNumber", "()I");
    jmForNumber_ =
        GetStaticMethodID(env, clz_, "forNumber", "(I)Lcom/yuanrong/libruntime/generated/Libruntime$LanguageType;");
}

void JNILanguageType::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

jobject JNILanguageType::FromCc(JNIEnv *env, const libruntime::LanguageType &type)
{
    return env->CallStaticObjectMethod(clz_, jmForNumber_, static_cast<jint>(type));
}

libruntime::LanguageType JNILanguageType::FromJava(JNIEnv *env, const jobject obj)
{
    return static_cast<libruntime::LanguageType>(env->CallIntMethod(obj, jmGetNumber_));
}

void JNIApiType::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/libruntime/generated/Libruntime$ApiType");
    jmGetNumber_ = GetJMethod(env, clz_, "getNumber", "()I");
    jmForNumber_ =
        GetStaticMethodID(env, clz_, "forNumber", "(I)Lcom/yuanrong/libruntime/generated/Libruntime$ApiType;");
}

void JNIApiType::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

jobject JNIApiType::FromCc(JNIEnv *env, const libruntime::ApiType &type)
{
    return env->CallStaticObjectMethod(clz_, jmForNumber_, static_cast<jint>(type));
}

libruntime::ApiType JNIApiType::FromJava(JNIEnv *env, const jobject obj)
{
    return static_cast<libruntime::ApiType>(env->CallIntMethod(obj, jmGetNumber_));
}

void JNICodeLoader::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/codemanager/CodeLoader");
    if (clz_ == nullptr) {
        // client don't need this package
        return;
    }
    jmLoad_ = GetStaticMethodID(env, clz_, "Load", "(Ljava/util/List;)Lcom/yuanrong/errorcode/ErrorInfo;");
}

void JNICodeLoader::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

YR::Libruntime::ErrorInfo JNICodeLoader::Load(JNIEnv *env, const std::vector<std::string> &codePaths)
{
    auto jPaths = JNIArrayList::FromCc<std::string>(
        env, codePaths, [](JNIEnv *env, const std::string &s) { return env->NewStringUTF(s.c_str()); });
    CHECK_JAVA_EXCEPTION_AND_RETURN_IF_OCCUR(
        env, YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, errorMessage));
    env->CallStaticVoidMethod(clz_, jmLoad_, jPaths);
    CHECK_JAVA_EXCEPTION_AND_RETURN_IF_OCCUR(
        env, YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, errorMessage));
    return YR::Libruntime::ErrorInfo();
}

void JNICodeExecutor::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/codemanager/CodeExecutor");
    if (clz_ == nullptr) {
        // client don't need this package
        return;
    }
    jmExecute_ = GetStaticMethodID(
        env, clz_, "execute",
        "(Lcom/yuanrong/libruntime/generated/Libruntime$FunctionMeta;Lcom/yuanrong/libruntime/generated/"
        "Libruntime$InvokeType;Ljava/util/List;)Lcom/yuanrong/executor/ReturnType;");
    jmDumpInstance_ = GetStaticMethodID(env, clz_, "dumpInstance", "(Ljava/lang/String;)Lcom/yuanrong/errorcode/Pair;");

    jmLoadInstance_ = GetStaticMethodID(env, clz_, "loadInstance", "([B[B)V");

    jmShutdown_ = GetStaticMethodID(env, clz_, "shutdown", "(I)Lcom/yuanrong/errorcode/ErrorInfo;");
    jmRecover_ = GetStaticMethodID(env, clz_, "recover", "()Lcom/yuanrong/errorcode/ErrorInfo;");
}

void JNICodeExecutor::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

YR::Libruntime::ErrorInfo JNICodeExecutor::Execute(
    JNIEnv *env, const YR::Libruntime::FunctionMeta &meta, const libruntime::InvokeType &invokeType,
    const std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &rawArgs,
    std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &returnValues)
{
    auto nativeMeta = JNIFunctionMeta::FromCc(env, meta);
    RETURN_IF_NULL(nativeMeta, YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                                                         "failed to convert func meta to java in jni"));
    auto nativeInvokeType = JNIInvokeType::FromCc(env, invokeType);
    RETURN_IF_NULL(nativeInvokeType, YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                                                               "failed to convert InvokeType to java in jni"));
    auto nativeArgs = JNIArrayList::FromCc<std::shared_ptr<YR::Libruntime::DataObject>>(
        env, rawArgs, [](JNIEnv *env, const std::shared_ptr<YR::Libruntime::DataObject> &arg) {
            return JNIByteBuffer::FromCc(env, arg);
        });
    RETURN_IF_NULL(nativeArgs, YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                                                         "failed to convert buffer to java in jni"));
    jobject jReturnType =
        (jobject)env->CallStaticObjectMethod(clz_, jmExecute_, nativeMeta, nativeInvokeType, nativeArgs);
    env->DeleteLocalRef(nativeMeta);
    env->DeleteLocalRef(nativeInvokeType);
    env->DeleteLocalRef(nativeArgs);

    CHECK_JAVA_EXCEPTION_AND_RETURN_IF_OCCUR(
        env, YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, errorMessage));

    auto [errorInfo, retVal] = JNIReturnType::FromJava(env, jReturnType);
    env->DeleteLocalRef(jReturnType);

    CHECK_JAVA_EXCEPTION_AND_RETURN_IF_OCCUR(
        env, YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, errorMessage));

    if (invokeType == libruntime::InvokeType::CreateInstance ||
        invokeType == libruntime::InvokeType::CreateInstanceStateless) {
        return errorInfo;
    }

    if (!errorInfo.OK()) {
        return errorInfo;
    }

    YR::Libruntime::ErrorInfo finalRes = JNICodeExecutor::ProcessInvokeResult(env, retVal, returnValues);
    return finalRes;
}

YR::Libruntime::ErrorInfo JNICodeExecutor::ProcessInvokeResult(
    JNIEnv *env, std::shared_ptr<YR::Libruntime::Buffer> retVal,
    std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &returnValues)
{
    uint64_t totalNativeBufferSize = 0;
    auto errorInfo = YR::Libruntime::LibruntimeManager::Instance().GetLibRuntime()->AllocReturnObject(
        returnValues[0], 0, retVal->GetSize(), {}, totalNativeBufferSize);
    if (!errorInfo.OK()) {
        return errorInfo;
    }

    if (returnValues.size() <= 0) {
        return YR::Libruntime::ErrorInfo(Libruntime::ErrorCode::ERR_PARAM_INVALID, "return value size < 0");
    }

    errorInfo = returnValues[0]->buffer->WriterLatch();
    if (!errorInfo.OK()) {
        return errorInfo;
    }
    (void)memset_s(returnValues[0]->meta->MutableData(), returnValues[0]->meta->GetSize(), 0,
                   returnValues[0]->meta->GetSize());
    errorInfo = returnValues[0]->data->MemoryCopy(retVal->ImmutableData(), retVal->GetSize());
    if (!errorInfo.OK()) {
        return errorInfo;
    }

    errorInfo = returnValues[0]->buffer->Seal({});
    if (!errorInfo.OK()) {
        return errorInfo;
    }

    errorInfo = returnValues[0]->buffer->WriterUnlatch();
    if (!errorInfo.OK()) {
        return errorInfo;
    }

    return YR::Libruntime::ErrorInfo();
}

YR::Libruntime::ErrorInfo JNICodeExecutor::DumpInstance(JNIEnv *env, const std::string &instanceID,
                                                        std::shared_ptr<YR::Libruntime::Buffer> &data)
{
    jobject jpair = (jobject)env->CallStaticObjectMethod(clz_, jmDumpInstance_, env->NewStringUTF(instanceID.c_str()));
    CHECK_JAVA_EXCEPTION_AND_RETURN_IF_OCCUR(
        env, YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, errorMessage));

    jbyteArray instanceBytes = (jbyteArray)JNIPair::GetFirst(env, jpair);
    jbyteArray clzNameBytes = (jbyteArray)JNIPair::GetSecond(env, jpair);

    ASSERT_NOT_NULL(instanceBytes);
    ASSERT_NOT_NULL(clzNameBytes);

    jbyte *instanceBytesPtr = env->GetByteArrayElements(instanceBytes, JNI_FALSE);
    size_t instanceBufSize = env->GetArrayLength(instanceBytes);

    jbyte *clzNameBytesPtr = env->GetByteArrayElements(clzNameBytes, JNI_FALSE);
    size_t clzNameSize = env->GetArrayLength(clzNameBytes);
    // data buffer format: [uint_8(size of buf1)|buf1(instanceBuf)|buf2(clsName)]
    // nativeBuffer is the combination of instanceBuf and clsName
    if (instanceBufSize > (std::numeric_limits<size_t>::max() - sizeof(size_t)) ||
        (sizeof(size_t) + instanceBufSize) > (std::numeric_limits<size_t>::max() - clzNameSize)) {
        return YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                                         "nativeBufferSize exceeds maximum allowed size");
    }
    size_t nativeBufferSize = sizeof(size_t) + instanceBufSize + clzNameSize;
    std::shared_ptr<YR::Libruntime::NativeBuffer> nativeBuffer =
        std::make_shared<YR::Libruntime::NativeBuffer>(nativeBufferSize);
    char *ptr = reinterpret_cast<char *>(nativeBuffer->MutableData());
    std::copy_n(reinterpret_cast<char *>(&instanceBufSize), sizeof(size_t), ptr);
    ptr += sizeof(size_t);
    std::copy_n(reinterpret_cast<char *>(instanceBytesPtr), instanceBufSize, ptr);
    ptr += instanceBufSize;
    std::copy_n(reinterpret_cast<char *>(clzNameBytesPtr), clzNameSize, ptr);
    YRLOG_DEBUG("Succeeded to copy instance byteArray and class name byteArray to Buffer data");

    env->ReleaseByteArrayElements(instanceBytes, instanceBytesPtr, JNI_ABORT);
    env->ReleaseByteArrayElements(clzNameBytes, clzNameBytesPtr, JNI_ABORT);
    env->DeleteLocalRef(instanceBytes);
    env->DeleteLocalRef(clzNameBytes);

    data = nativeBuffer;
    return YR::Libruntime::ErrorInfo();
}

YR::Libruntime::ErrorInfo JNICodeExecutor::LoadInstance(JNIEnv *env, std::shared_ptr<YR::Libruntime::Buffer> data)
{
    size_t totalSize = data->GetSize();
    if (totalSize == 0) {
        YRLOG_WARN("Failed to load instance, empty buffer");
        return YR::Libruntime::ErrorInfo();
    }

    // deserialize data buffer format: [uint_8(size of buf1)|buf1(instanceBuf)|buf2(clsName)]
    char *ptr = reinterpret_cast<char *>(data->MutableData());

    size_t instanceBufSize;
    std::copy_n(ptr, sizeof(size_t), reinterpret_cast<char *>(&instanceBufSize));
    ptr += sizeof(size_t);

    jbyteArray instanceBytes = env->NewByteArray(instanceBufSize);
    env->SetByteArrayRegion(instanceBytes, 0, instanceBufSize, reinterpret_cast<const jbyte *>(ptr));
    ptr += instanceBufSize;

    size_t clzNameSize = totalSize - sizeof(size_t) - instanceBufSize;
    jbyteArray clzNameBytes = env->NewByteArray(clzNameSize);
    env->SetByteArrayRegion(clzNameBytes, 0, clzNameSize, reinterpret_cast<const jbyte *>(ptr));
    YRLOG_DEBUG("Succeeded to split instance byteArray and class name byteArray from Buffer data");

    env->CallStaticObjectMethod(clz_, jmLoadInstance_, instanceBytes, clzNameBytes);
    CHECK_JAVA_EXCEPTION_AND_RETURN_IF_OCCUR(
        env, YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, errorMessage));

    jobject jerrorInfo = (jobject)env->CallStaticObjectMethod(clz_, jmRecover_);
    auto errorInfo = JNIErrorInfo::FromJava(env, jerrorInfo);
    env->DeleteLocalRef(jerrorInfo);
    CHECK_JAVA_EXCEPTION_AND_RETURN_IF_OCCUR(
        env, YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, errorMessage));

    return errorInfo;
}

YR::Libruntime::ErrorInfo JNICodeExecutor::Shutdown(JNIEnv *env, uint64_t gracePeriodSeconds)
{
    if (gracePeriodSeconds > static_cast<uint64_t>(std::numeric_limits<jint>::max())) {
        return YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "Grace period is too large.");
    }
    jint jgracePeriodSeconds = static_cast<jint>(gracePeriodSeconds);

    jobject jerrorInfo = (jobject)env->CallStaticObjectMethod(clz_, jmShutdown_, jgracePeriodSeconds);

    auto errorInfo = JNIErrorInfo::FromJava(env, jerrorInfo);
    env->DeleteLocalRef(jerrorInfo);

    CHECK_JAVA_EXCEPTION_AND_RETURN_IF_OCCUR(
        env, YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, errorMessage));

    return errorInfo;
}

void JNILibRuntimeConfig::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/jni/LibRuntimeConfig");
    jmIsDriver_ = GetJMethod(env, clz_, "isDriver", "()Z");
    jmIsInCluster_ = GetJMethod(env, clz_, "isInCluster", "()Z");
    jmIsEnableMetrics_ = GetJMethod(env, clz_, "isEnableMetrics", "()Z");
    jmIsEnableMTLS_ = GetJMethod(env, clz_, "isEnableMTLS", "()Z");
    jmIsEnableDsEncrypt_ = GetJMethod(env, clz_, "isEnableDsEncrypt", "()Z");
    jmGetCertificateFilePath_ = GetJMethod(env, clz_, "getCertificateFilePath", "()Ljava/lang/String;");
    jmGetPrivateKeyPath_ = GetJMethod(env, clz_, "getPrivateKeyPath", "()Ljava/lang/String;");
    jmGetDsPublicKeyContextPath_ = GetJMethod(env, clz_, "getDsPublicKeyContextPath", "()Ljava/lang/String;");
    jmGetRuntimePublicKeyContextPath_ = GetJMethod(env, clz_, "getRuntimePublicKeyContextPath", "()Ljava/lang/String;");
    jmGetRuntimePrivateKeyContextPath_ =
        GetJMethod(env, clz_, "getRuntimePrivateKeyContextPath", "()Ljava/lang/String;");
    jmGetVerifyFilePath_ = GetJMethod(env, clz_, "getVerifyFilePath", "()Ljava/lang/String;");
    jmGetPrivateKeyPaaswd_ = GetJMethod(env, clz_, "getPrivateKeyPaaswd", "()Ljava/lang/String;");
    jmGetServerName_ = GetJMethod(env, clz_, "getServerName", "()Ljava/lang/String;");
    jmGetFunctionSystemIpAddr_ = GetJMethod(env, clz_, "getFunctionSystemIpAddr", "()Ljava/lang/String;");
    jmGetFunctionSystemPort_ = GetJMethod(env, clz_, "getFunctionSystemPort", "()I");
    jmGetFunctionSystemRtServerIpAddr_ =
        GetJMethod(env, clz_, "getFunctionSystemRtServerIpAddr", "()Ljava/lang/String;");
    jmGetFunctionSystemRtServerPort_ = GetJMethod(env, clz_, "getFunctionSystemRtServerPort", "()I");
    jmGetDataSystemIpAddr_ = GetJMethod(env, clz_, "getDataSystemIpAddr", "()Ljava/lang/String;");
    jmGetDataSystemPort_ = GetJMethod(env, clz_, "getDataSystemPort", "()I");
    jmGetJobId_ = GetJMethod(env, clz_, "getJobId", "()Ljava/lang/String;");
    jmGetRuntimeId_ = GetJMethod(env, clz_, "getRuntimeId", "()Ljava/lang/String;");
    jmGetFunctionIds_ = GetJMethod(env, clz_, "getFunctionIds", "()Ljava/util/Map;");
    jmGetFunctionUrn_ = GetJMethod(env, clz_, "getFunctionUrn", "()Ljava/lang/String;");
    jmGetLogLevel_ = GetJMethod(env, clz_, "getLogLevel", "()Ljava/lang/String;");
    jmGetLogDir_ = GetJMethod(env, clz_, "getLogDir", "()Ljava/lang/String;");
    jmGetLogFileSizeMax_ = GetJMethod(env, clz_, "getLogFileSizeMax", "()I");
    jmGetLogFileNumMax_ = GetJMethod(env, clz_, "getLogFileNumMax", "()I");
    jmGetLogFlushInterval_ = GetJMethod(env, clz_, "getLogFlushInterval", "()I");
    jmIsLogMerge_ = GetJMethod(env, clz_, "isLogMerge", "()Z");
    jmGetMetaConfig_ = GetJMethod(env, clz_, "getMetaConfig", "()Ljava/lang/String;");
    jmGetRecycleTime_ = GetJMethod(env, clz_, "getRecycleTime", "()I");
    jmGetMaxTaskInstanceNum_ = GetJMethod(env, clz_, "getMaxTaskInstanceNum", "()I");
    jmGetMaxConcurrencyCreateNum_ = GetJMethod(env, clz_, "getMaxConcurrencyCreateNum", "()I");
    jmGetThreadPoolSize_ = GetJMethod(env, clz_, "getThreadPoolSize", "()I");
    jmGetLoadPaths_ = GetJMethod(env, clz_, "getLoadPaths", "()Ljava/util/List;");
    jGetHttpIocThreadsNum_ = GetJMethod(env, clz_, "getHttpIocThreadsNum", "()I");
    jGetHttpIdleTime_ = GetJMethod(env, clz_, "getHttpIdleTime", "()I");
    jGetRpcTimeout_ = GetJMethod(env, clz_, "getRpcTimeout", "()I");
    jGetTenantId_ = GetJMethod(env, clz_, "getTenantId", "()Ljava/lang/String;");
    jGetNs_ = GetJMethod(env, clz_, "getNs", "()Ljava/lang/String;");
    jmGetCustomEnvs_ = GetJMethod(env, clz_, "getCustomEnvs", "()Ljava/util/Map;");
    jmGetCodePath_ = GetJMethod(env, clz_, "getCodePath", "()Ljava/util/List;");
}

void JNILibRuntimeConfig::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

YR::Libruntime::LibruntimeConfig JNILibRuntimeConfig::FromJava(JNIEnv *env, const jobject &meta)
{
    RETURN_IF_NULL(meta, YR::Libruntime::LibruntimeConfig{});
    YR::Libruntime::LibruntimeConfig libConfig;
    libConfig.functionSystemIpAddr =
        JNIString::FromJava(env, static_cast<jstring>(env->CallObjectMethod(meta, jmGetFunctionSystemIpAddr_)));
    libConfig.functionSystemPort = env->CallIntMethod(meta, jmGetFunctionSystemPort_);
    libConfig.functionSystemRtServerIpAddr =
        JNIString::FromJava(env, static_cast<jstring>(env->CallObjectMethod(meta, jmGetFunctionSystemRtServerIpAddr_)));
    libConfig.functionSystemRtServerPort = env->CallIntMethod(meta, jmGetFunctionSystemRtServerPort_);
    libConfig.dataSystemIpAddr =
        JNIString::FromJava(env, static_cast<jstring>(env->CallObjectMethod(meta, jmGetDataSystemIpAddr_)));
    libConfig.dataSystemPort = env->CallIntMethod(meta, jmGetDataSystemPort_);
    libConfig.isDriver = static_cast<bool>(env->CallBooleanMethod(meta, jmIsDriver_));
    libConfig.jobId = JNIString::FromJava(env, static_cast<jstring>(env->CallObjectMethod(meta, jmGetJobId_)));
    libConfig.runtimeId = JNIString::FromJava(env, static_cast<jstring>(env->CallObjectMethod(meta, jmGetRuntimeId_)));
    libConfig.selfLanguage = libruntime::LanguageType::Java;
    // LanguageType -> std::string
    libConfig.functionIds = JNIMap::FromJava<libruntime::LanguageType, std::string>(
        env, env->CallObjectMethod(meta, jmGetFunctionIds_),
        [](JNIEnv *env, const jobject &ko) -> libruntime::LanguageType { return JNILanguageType::FromJava(env, ko); },
        [](JNIEnv *env, const jobject &vo) -> std::string {
            return JNIString::FromJava(env, static_cast<jstring>(vo));
        });
    libConfig.logLevel = JNIString::FromJava(env, static_cast<jstring>(env->CallObjectMethod(meta, jmGetLogLevel_)));
    libConfig.logDir = JNIString::FromJava(env, static_cast<jstring>(env->CallObjectMethod(meta, jmGetLogDir_)));
    libConfig.logFileSizeMax = static_cast<uint32_t>(env->CallIntMethod(meta, jmGetLogFileSizeMax_));
    libConfig.logFileNumMax = static_cast<uint32_t>(env->CallIntMethod(meta, jmGetLogFileNumMax_));
    libConfig.logFlushInterval = env->CallIntMethod(meta, jmGetLogFlushInterval_);
    libConfig.isLogMerge = static_cast<bool>(env->CallBooleanMethod(meta, jmIsLogMerge_));
    libConfig.libruntimeOptions = Libruntime::LibruntimeOptions{},
    libConfig.recycleTime = env->CallIntMethod(meta, jmGetRecycleTime_);
    libConfig.maxTaskInstanceNum = env->CallIntMethod(meta, jmGetMaxTaskInstanceNum_);
    libConfig.maxConcurrencyCreateNum = env->CallIntMethod(meta, jmGetMaxConcurrencyCreateNum_);
    libConfig.enableMetrics = static_cast<bool>(env->CallBooleanMethod(meta, jmIsEnableMetrics_));
    libConfig.threadPoolSize = static_cast<uint32_t>(env->CallIntMethod(meta, jmGetThreadPoolSize_));
    libConfig.loadPaths = JNIList::FromJava<std::string>(
        env, env->CallObjectMethod(meta, jmGetLoadPaths_),
        [](JNIEnv *env, jobject obj) -> std::string { return JNIString::FromJava(env, static_cast<jstring>(obj)); });
    libConfig.tenantId = JNIString::FromJava(env, static_cast<jstring>(env->CallObjectMethod(meta, jGetTenantId_)));
    libConfig.enableMTLS = static_cast<bool>(env->CallBooleanMethod(meta, jmIsEnableMTLS_));
    libConfig.encryptEnable = static_cast<bool>(env->CallBooleanMethod(meta, jmIsEnableDsEncrypt_));
    libConfig.dsPublicKeyPath =
        JNIString::FromJava(env, static_cast<jstring>(env->CallObjectMethod(meta, jmGetDsPublicKeyContextPath_)));
    libConfig.runtimePublicKeyPath =
        JNIString::FromJava(env, static_cast<jstring>(env->CallObjectMethod(meta, jmGetRuntimePublicKeyContextPath_)));
    libConfig.runtimePrivateKeyPath =
        JNIString::FromJava(env, static_cast<jstring>(env->CallObjectMethod(meta, jmGetRuntimePrivateKeyContextPath_)));
    libConfig.privateKeyPath =
        JNIString::FromJava(env, static_cast<jstring>(env->CallObjectMethod(meta, jmGetPrivateKeyPath_)));
    libConfig.certificateFilePath =
        JNIString::FromJava(env, static_cast<jstring>(env->CallObjectMethod(meta, jmGetCertificateFilePath_)));
    libConfig.verifyFilePath =
        JNIString::FromJava(env, static_cast<jstring>(env->CallObjectMethod(meta, jmGetVerifyFilePath_)));
    libConfig.serverName =
        JNIString::FromJava(env, static_cast<jstring>(env->CallObjectMethod(meta, jmGetServerName_)));
    libConfig.ns = JNIString::FromJava(env, static_cast<jstring>(env->CallObjectMethod(meta, jGetNs_)));
    libConfig.customEnvs = JNIMap::FromJava<std::string, std::string>(
        env, env->CallObjectMethod(meta, jmGetCustomEnvs_),
        [](JNIEnv *env, const jobject &ko) -> std::string {
            return JNIString::FromJava(env, static_cast<jstring>(ko));
        },
        [](JNIEnv *env, const jobject &vo) -> std::string {
            return JNIString::FromJava(env, static_cast<jstring>(vo));
        });

    jstring jStrPasswd = static_cast<jstring>(env->CallObjectMethod(meta, jmGetPrivateKeyPaaswd_));
    if (jStrPasswd != nullptr) {
        const char *passwd = env->GetStringUTFChars(jStrPasswd, nullptr);
        if (passwd != nullptr) {
            size_t passwdLen = strlen(passwd) + 1;
            memcpy_s(libConfig.privateKeyPaaswd, passwdLen, passwd, passwdLen);
            env->ReleaseStringUTFChars(jStrPasswd, passwd);
        }
    }
    libConfig.inCluster = static_cast<bool>(env->CallBooleanMethod(meta, jmIsInCluster_));
    libConfig.httpIocThreadsNum = static_cast<uint32_t>(env->CallIntMethod(meta, jGetHttpIocThreadsNum_));
    libConfig.httpIdleTime = env->CallIntMethod(meta, jGetHttpIdleTime_);
    libConfig.rpcTimeout = static_cast<uint32_t>(env->CallIntMethod(meta, jGetRpcTimeout_));

    auto codePath = JNIList::FromJava<std::string>(
        env, env->CallObjectMethod(meta, jmGetCodePath_),
        [](JNIEnv *env, jobject obj) -> std::string { return JNIString::FromJava(env, static_cast<jstring>(obj)); });

    libConfig.loadPaths.insert(libConfig.loadPaths.end(), codePath.begin(), codePath.end());
    return libConfig;
}

void JNIMap::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "java/util/Map");
    jmEntrySet_ = GetJMethod(env, clz_, "entrySet", "()Ljava/util/Set;");
}

void JNIMap::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

template <typename K, typename V>
std::unordered_map<K, V> JNIMap::FromJava(JNIEnv *env, const jobject &jmap,
                                          std::function<K(JNIEnv *, const jobject &)> convertKey,
                                          std::function<V(JNIEnv *, const jobject &)> convertVal)
{
    std::unordered_map<K, V> cmap;
    RETURN_IF_NULL(jmap, cmap);

    jobject jEntrySet = env->CallObjectMethod(jmap, jmEntrySet_);
    std::set<std::pair<K, V>> entrySet = JNISet::FromJava<std::pair<K, V>>(
        env, jEntrySet, [&convertKey, &convertVal](JNIEnv *e, jobject jEntry) -> std::pair<K, V> {
            // here the element is a java Map$Entry
            return std::make_pair(JNIMapEntry::GetKey(e, jEntry, convertKey),
                                  JNIMapEntry::GetVal(e, jEntry, convertVal));
        });

    for (auto e : entrySet) {
        cmap.emplace(e);
    }
    return cmap;
}

void JNIHashMap::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "java/util/HashMap");
    init_ = GetJMethod(env, clz_, "<init>", "()V");
    jmPut_ = GetJMethod(env, clz_, "put", "(Ljava/lang/Object;Ljava/lang/Object;)Ljava/lang/Object;");
}

void JNIHashMap::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
    }
}

template <typename K, typename V>
jobject JNIHashMap::FromCc(JNIEnv *env, const std::unordered_map<K, V> &map,
                           std::function<jobject(JNIEnv *, const K &)> converterK,
                           std::function<jobject(JNIEnv *, const V &)> converterV)
{
    jobject hashMap = env->NewObject(clz_, init_);
    for (const auto &kv : map) {
        auto tmpKey = converterK(env, kv.first);
        auto tmpValue = converterV(env, kv.second);
        env->CallObjectMethod(hashMap, jmPut_, tmpKey, tmpValue);

        env->DeleteLocalRef(tmpKey);
        env->DeleteLocalRef(tmpValue);
    }
    return hashMap;
}

void JNISet::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "java/util/Set");
    jmIterator_ = GetJMethod(env, clz_, "iterator", "()Ljava/util/Iterator;");
}

void JNISet::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

template <typename T>
std::set<T> JNISet::FromJava(JNIEnv *env, const jobject &obj, std::function<T(JNIEnv *, jobject)> convertElement)
{
    std::set<T> ret;
    RETURN_IF_NULL(obj, ret);
    jobject jIter = env->CallObjectMethod(obj, jmIterator_);
    JNIIterator::FromJava<T, std::set<T>>(env, jIter, convertElement, ret,
                                          [](std::set<T> &s, const T &t) { s.emplace(t); });
    return ret;
}

void JNIUnorderedSet::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "java/util/HashSet");
    jmIterator_ = GetJMethod(env, clz_, "iterator", "()Ljava/util/Iterator;");
}

void JNIUnorderedSet::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

template <typename T>
std::unordered_set<T> JNIUnorderedSet::FromJava(JNIEnv *env, const jobject &obj,
                                                std::function<T(JNIEnv *, jobject)> convertElement)
{
    std::unordered_set<T> ret;
    RETURN_IF_NULL(obj, ret);
    jobject jIter = env->CallObjectMethod(obj, jmIterator_);
    JNIIterator::FromJava<T, std::unordered_set<T>>(env, jIter, convertElement, ret,
                                                    [](std::unordered_set<T> &s, const T &t) { s.emplace(t); });
    return ret;
}

void JNIInvokeArg::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/api/InvokeArg");
    init_ = GetJMethod(env, clz_, "<init>", "()V");
    getData_ = GetJMethod(env, clz_, "getData", "()[B");
    isObjectRef_ = GetJMethod(env, clz_, "isObjectRef", "()Z");
    getObjectId_ = GetJMethod(env, clz_, "getObjId", "()Ljava/lang/String;");
    getNestedObjects_ = GetJMethod(env, clz_, "getNestedObjects", "()Ljava/util/HashSet;");
}

void JNIInvokeArg::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

std::vector<YR::Libruntime::InvokeArg> JNIInvokeArg::FromJavaList(JNIEnv *env, jobject o, std::string tenantId)
{
    auto ret = YR::jni::JNIList::FromJava<YR::Libruntime::InvokeArg>(
        env, o, [tenantId](JNIEnv *env, const jobject &obj) -> YR::Libruntime::InvokeArg {
            return YR::jni::JNIInvokeArg::FromJava(env, obj, tenantId);
        });
    return ret;
}

YR::Libruntime::InvokeArg JNIInvokeArg::FromJava(JNIEnv *env, jobject o, std::string tenantId)
{
    YR::Libruntime::InvokeArg invokeArg{};
    RETURN_IF_NULL(o, invokeArg);

    jbyteArray jbytes = (jbyteArray)env->CallObjectMethod(o, getData_);
    std::shared_ptr<YR::Libruntime::DataObject> dataObj;
    std::shared_ptr<YR::Libruntime::Buffer> buf;
    JNIDataObject::WriteDataObject(env, dataObj, jbytes);
    invokeArg.dataObj = dataObj;

    invokeArg.nestedObjects = GetNestedObjects(env, o);
    invokeArg.isRef = GetIsRef(env, o);
    invokeArg.objId = GetObjectId(env, o);
    invokeArg.tenantId = tenantId;
    return invokeArg;
}

bool JNIInvokeArg::GetIsRef(JNIEnv *env, jobject o)
{
    jboolean jIsRef = static_cast<jboolean>(env->CallBooleanMethod(o, isObjectRef_));
    bool ret = static_cast<bool>(jIsRef);
    return ret;
}

std::string JNIInvokeArg::GetObjectId(JNIEnv *env, jobject o)
{
    jstring jstr = static_cast<jstring>(env->CallObjectMethod(o, getObjectId_));
    std::string ret = JNIString::FromJava(env, jstr);
    return ret;
}

std::unordered_set<std::string> JNIInvokeArg::GetNestedObjects(JNIEnv *env, jobject o)
{
    jobject jobj = static_cast<jobject>(env->CallObjectMethod(o, getNestedObjects_));
    std::unordered_set<std::string> entrySet = JNIUnorderedSet::FromJava<std::string>(
        env, jobj,
        [](JNIEnv *e, jobject obj) -> std::string { return JNIString::FromJava(e, static_cast<jstring>(obj)); });
    return entrySet;
}

void JNIMapEntry::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "java/util/Map$Entry");
    jmGetKey_ = GetJMethod(env, clz_, "getKey", "()Ljava/lang/Object;");
    jmGetVal_ = GetJMethod(env, clz_, "getValue", "()Ljava/lang/Object;");
}

void JNIMapEntry::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

template <typename T>
T JNIMapEntry::GetKey(JNIEnv *env, const jobject &inst, std::function<T(JNIEnv *env, const jobject &inst)> convert)
{
    jobject obj = env->CallObjectMethod(inst, jmGetKey_);
    T r = convert(env, obj);
    return r;
}

template <typename T>
T JNIMapEntry::GetVal(JNIEnv *env, const jobject &inst, std::function<T(JNIEnv *env, const jobject &inst)> convert)
{
    jobject obj = env->CallObjectMethod(inst, jmGetVal_);
    T r = convert(env, obj);
    return r;
}

void JNIIterator::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "java/util/Iterator");
    jmHasNext_ = GetJMethod(env, clz_, "hasNext", "()Z");
    jmNext_ = GetJMethod(env, clz_, "next", "()Ljava/lang/Object;");
}

void JNIIterator::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

bool JNIIterator::HasNext(JNIEnv *env, const jobject &inst)
{
    RETURN_IF_NULL(inst, false);
    bool has = static_cast<bool>(env->CallBooleanMethod(inst, jmHasNext_));
    return has;
}

jobject JNIIterator::Next(JNIEnv *env, const jobject &inst)
{
    RETURN_IF_NULL(inst, nullptr);
    jobject ret = env->CallObjectMethod(inst, jmNext_);
    return ret;
}

void JNIIterator::ForEachJavaObject(JNIEnv *env, const jobject &inst,
                                    std::function<void(JNIEnv *, const jobject &)> traversal)
{
    if (traversal == nullptr) {
        return;
    }

    // The 'inst' does not need to be passed to the
    // 'traversal' function for execution. The following
    // 'while' loop can process all elements in the 'inst'.
    while (JNIIterator::HasNext(env, inst)) {
        jobject next = JNIIterator::Next(env, inst);
        traversal(env, next);
    }
}

template <typename T, typename CT>
CT JNIIterator::FromJava(JNIEnv *env, const jobject &inst, std::function<T(JNIEnv *, const jobject &)> eleConvert,
                         CT &container, std::function<void(CT &, const T &)> addToContainer)
{
    if (inst == nullptr || eleConvert == nullptr || addToContainer == nullptr) {
        return container;
    }
    ForEachJavaObject(env, inst, [&container, addToContainer, eleConvert](JNIEnv *env, const jobject &ele) {
        addToContainer(container, eleConvert(env, ele));
    });
    return container;
}

void JNIGroupOptions::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

void JNIGroupOptions::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/GroupOptions");
    init_ = GetJMethod(env, clz_, "<init>", "()V");
    getTimeout_ = GetJMethod(env, clz_, "getTimeout", "()I");
    getSameLifecycle_ = GetJMethod(env, clz_, "isSameLifecycle", "()Z");
}

YR::Libruntime::GroupOpts JNIGroupOptions::FromJava(JNIEnv *env, jobject o)
{
    YR::Libruntime::GroupOpts groupOpts{};
    RETURN_IF_NULL(o, groupOpts);
    YR::Libruntime::GroupOpts opts;
    opts.timeout = GetTimeout(env, o);
    opts.sameLifecycle = GetSameLifecycle(env, o);
    return opts;
}

int JNIGroupOptions::GetTimeout(JNIEnv *env, jobject o)
{
    jint timeoutInt = static_cast<jint>(env->CallIntMethod(o, getTimeout_));
    int timeout = timeoutInt;
    return timeout;
}

bool JNIGroupOptions::GetSameLifecycle(JNIEnv *env, jobject o)
{
    jboolean value = static_cast<jboolean>(env->CallBooleanMethod(o, getSameLifecycle_));
    bool result = static_cast<bool>(value);
    return result;
}

void JNIInvokeOptions::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/InvokeOptions");
    init_ = GetJMethod(env, clz_, "<init>", "()V");
    getCpu_ = GetJMethod(env, clz_, "getCpu", "()I");
    getMemory_ = GetJMethod(env, clz_, "getMemory", "()I");
    getCustomResources_ = GetJMethod(env, clz_, "getCustomResources", "()Ljava/util/Map;");
    getCustomExtensions_ = GetJMethod(env, clz_, "getCustomExtensions", "()Ljava/util/Map;");
    getCreateOptions_ = GetJMethod(env, clz_, "getCreateOptions", "()Ljava/util/Map;");
    getPodLabels_ = GetJMethod(env, clz_, "getPodLabels", "()Ljava/util/Map;");
    getLabels_ = GetJMethod(env, clz_, "getLabels", "()Ljava/util/List;");
    getAffinity_ = GetJMethod(env, clz_, "getAffinity", "()Ljava/util/Map;");
    getRetryTimes_ = GetJMethod(env, clz_, "getRetryTimes", "()I");
    getPriority_ = GetJMethod(env, clz_, "getPriority", "()I");
    getInstancePriority_ = GetJMethod(env, clz_, "getInstancePriority", "()I");
    getRecoverRetryTimes_ = GetJMethod(env, clz_, "getRecoverRetryTimes", "()I");
    getInvokeGroupName_ = GetJMethod(env, clz_, "getGroupName", "()Ljava/lang/String;");
    getTraceId_ = GetJMethod(env, clz_, "getTraceId", "()Ljava/lang/String;");
    getNeedOrder_ = GetJMethod(env, clz_, "isNeedOrder", "()Z");
    getScheduleTimeoutMs_ = GetJMethod(env, clz_, "getScheduleTimeoutMs", "()J");
    getPreemptedAllowed_ = GetJMethod(env, clz_, "isPreemptedAllowed", "()Z");
    isPreferredPriority_ = GetJMethod(env, clz_, "isPreferredPriority", "()Z");
    isRequiredPriority_ = GetJMethod(env, clz_, "isRequiredPriority", "()Z");
    isPreferredAntiOtherLabels_ = GetJMethod(env, clz_, "isPreferredAntiOtherLabels", "()Z");
    getScheduleAffinities_ = GetJMethod(env, clz_, "getScheduleAffinities", "()Ljava/util/List;");
    getEnvVars_ = GetJMethod(env, clz_, "getEnvVars", "()Ljava/util/Map;");
    getAliasParams_ = GetJMethod(env, clz_, "getAliasParams", "()Ljava/util/Map;");
}

void JNIInvokeOptions::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

YR::Libruntime::InvokeOptions JNIInvokeOptions::FromJava(JNIEnv *env, jobject o)
{
    YR::Libruntime::InvokeOptions invokeOptions{};
    RETURN_IF_NULL(o, invokeOptions);
    bool preferredPriority = static_cast<bool>(env->CallBooleanMethod(o, isPreferredPriority_));
    bool requiredPriority = static_cast<bool>(env->CallBooleanMethod(o, isRequiredPriority_));
    bool preferredAntiOtherLabels = static_cast<bool>(env->CallBooleanMethod(o, isPreferredAntiOtherLabels_));

    YR::Libruntime::InvokeOptions opts;
    opts.cpu = GetCpu(env, o);
    opts.memory = GetMemory(env, o);
    opts.customResources = GetCustomResources(env, o);
    opts.customExtensions = GetCustomExtensions(env, o);
    opts.createOptions = GetCreateOptions(env, o);
    opts.podLabels = GetPodLabels(env, o);
    opts.labels = GetLabels(env, o);
    opts.affinity = GetAffinity(env, o);
    opts.scheduleAffinities =
        GetScheduleAffinities(env, o, preferredPriority, requiredPriority, preferredAntiOtherLabels);
    opts.retryTimes = GetRetryTimes(env, o);
    opts.priority = GetPriority(env, o);
    opts.instancePriority = GetInstancePriority(env, o);
    opts.groupName = GetInvokeGroupName(env, o);
    opts.needOrder = GetNeedOrder(env, o);
    opts.traceId = GetTraceId(env, o);
    opts.scheduleTimeoutMs = GetScheduleTimeoutMs(env, o);
    opts.preemptedAllowed = GetPreemptedAllowed(env, o);
    opts.recoverRetryTimes = GetRecoverRetryTimes(env, o);
    opts.envVars = GetEnvVars(env, o);
    opts.aliasParams = GetAliasParams(env, o);

    return opts;
}

std::list<std::shared_ptr<YR::Libruntime::Affinity>> JNIInvokeOptions::GetScheduleAffinities(
    JNIEnv *env, jobject o, bool preferredPriority, bool requiredPriority, bool preferredAntiOtherLabels)

{
    jobject jAffinities = env->CallObjectMethod(o, getScheduleAffinities_);
    std::list<std::shared_ptr<YR::Libruntime::Affinity>> affinities;
    int size = JNIList::Size(env, jAffinities);
    for (int i = 0; i < size; i++) {
        auto element = JNIList::Get(env, jAffinities, i);
        auto item = JNIAffinity::FromJava(env, element, preferredPriority, requiredPriority, preferredAntiOtherLabels);
        affinities.push_back(item);
    }
    return affinities;
}

int JNIInvokeOptions::GetCpu(JNIEnv *env, jobject o)
{
    jint cpuInt = static_cast<jint>(env->CallIntMethod(o, getCpu_));
    int result = cpuInt;
    return result;
}

int JNIInvokeOptions::GetMemory(JNIEnv *env, jobject o)
{
    jint memoryInt = static_cast<jint>(env->CallIntMethod(o, getMemory_));
    int result = static_cast<int>(memoryInt);
    return result;
}

std::unordered_map<std::string, float> JNIInvokeOptions::GetCustomResources(JNIEnv *env, jobject o)
{
    std::unordered_map<std::string, float> result = JNIMap::FromJava<std::string, float>(
        env, env->CallObjectMethod(o, getCustomResources_),
        [](JNIEnv *env, const jobject &ko) -> std::string {
            return JNIString::FromJava(env, static_cast<jstring>(ko));
        },
        [](JNIEnv *env, const jobject &vo) -> float {
            jclass cls = env->GetObjectClass(vo);
            jfieldID fid = env->GetFieldID(cls, "value", "F");
            jfloat floatValue = env->GetFloatField(vo, fid);
            float cppFloatValue = static_cast<float>(floatValue);
            return cppFloatValue;
        });
    return result;
}

std::unordered_map<std::string, std::string> JNIInvokeOptions::GetCustomExtensions(JNIEnv *env, jobject o)
{
    std::unordered_map<std::string, std::string> result = JNIMap::FromJava<std::string, std::string>(
        env, env->CallObjectMethod(o, getCustomExtensions_),
        [](JNIEnv *env, const jobject &ko) -> std::string {
            return JNIString::FromJava(env, static_cast<jstring>(ko));
        },
        [](JNIEnv *env, const jobject &vo) -> std::string {
            return JNIString::FromJava(env, static_cast<jstring>(vo));
        });
    return result;
}

std::unordered_map<std::string, std::string> JNIInvokeOptions::GetCreateOptions(JNIEnv *env, jobject o)
{
    std::unordered_map<std::string, std::string> result = JNIMap::FromJava<std::string, std::string>(
        env, env->CallObjectMethod(o, getCreateOptions_),
        [](JNIEnv *env, const jobject &ko) -> std::string {
            return JNIString::FromJava(env, static_cast<jstring>(ko));
        },
        [](JNIEnv *env, const jobject &vo) -> std::string {
            return JNIString::FromJava(env, static_cast<jstring>(vo));
        });
    return result;
}

std::unordered_map<std::string, std::string> JNIInvokeOptions::GetAliasParams(JNIEnv *env, jobject o)
{
    std::unordered_map<std::string, std::string> result = JNIMap::FromJava<std::string, std::string>(
        env, env->CallObjectMethod(o, getAliasParams_),
        [](JNIEnv *env, const jobject &ko) -> std::string {
            return JNIString::FromJava(env, static_cast<jstring>(ko));
        },
        [](JNIEnv *env, const jobject &vo) -> std::string {
            return JNIString::FromJava(env, static_cast<jstring>(vo));
        });
    return result;
}

std::unordered_map<std::string, std::string> JNIInvokeOptions::GetPodLabels(JNIEnv *env, jobject o)
{
    std::unordered_map<std::string, std::string> result = JNIMap::FromJava<std::string, std::string>(
        env, env->CallObjectMethod(o, getPodLabels_),
        [](JNIEnv *env, const jobject &ko) -> std::string {
            return JNIString::FromJava(env, static_cast<jstring>(ko));
        },
        [](JNIEnv *env, const jobject &vo) -> std::string {
            return JNIString::FromJava(env, static_cast<jstring>(vo));
        });
    return result;
}

std::vector<std::string> JNIInvokeOptions::GetLabels(JNIEnv *env, jobject o)
{
    std::vector<std::string> result = JNIList::FromJava<std::string>(
        env, env->CallObjectMethod(o, getLabels_),
        [](JNIEnv *env, jobject obj) -> std::string { return JNIString::FromJava(env, static_cast<jstring>(obj)); });
    return result;
}

std::unordered_map<std::string, std::string> JNIInvokeOptions::GetAffinity(JNIEnv *env, jobject o)
{
    std::unordered_map<std::string, std::string> result = JNIMap::FromJava<std::string, std::string>(
        env, env->CallObjectMethod(o, getAffinity_),
        [](JNIEnv *env, const jobject &ko) -> std::string {
            return JNIString::FromJava(env, static_cast<jstring>(ko));
        },
        [](JNIEnv *env, const jobject &vo) -> std::string {
            return JNIString::FromJava(env, static_cast<jstring>(vo));
        });
    return result;
}

size_t JNIInvokeOptions::GetRetryTimes(JNIEnv *env, jobject o)
{
    jint value = static_cast<jint>(env->CallIntMethod(o, getRetryTimes_));
    size_t result = static_cast<size_t>(value);
    return result;
}

size_t JNIInvokeOptions::GetPriority(JNIEnv *env, jobject o)
{
    jint value = static_cast<jint>(env->CallIntMethod(o, getPriority_));
    size_t result = static_cast<size_t>(value);
    return result;
}

int JNIInvokeOptions::GetInstancePriority(JNIEnv *env, jobject o)
{
    jint value = static_cast<jint>(env->CallIntMethod(o, getInstancePriority_));
    int result = static_cast<int>(value);
    return result;
}

int JNIInvokeOptions::GetRecoverRetryTimes(JNIEnv *env, jobject o)
{
    jint value = static_cast<jint>(env->CallIntMethod(o, getRecoverRetryTimes_));
    int result = static_cast<int>(value);
    return result;
}

std::string JNIInvokeOptions::GetInvokeGroupName(JNIEnv *env, jobject o)
{
    jstring jstr = static_cast<jstring>(env->CallObjectMethod(o, getInvokeGroupName_));
    std::string ret = JNIString::FromJava(env, jstr);
    return ret;
}

std::string JNIInvokeOptions::GetTraceId(JNIEnv *env, jobject o)
{
    jstring jstr = static_cast<jstring>(env->CallObjectMethod(o, getTraceId_));
    std::string ret = JNIString::FromJava(env, jstr);
    return ret;
}

std::unordered_map<std::string, std::string> JNIInvokeOptions::GetEnvVars(JNIEnv *env, jobject o)
{
    std::unordered_map<std::string, std::string> result = JNIMap::FromJava<std::string, std::string>(
        env, env->CallObjectMethod(o, getEnvVars_),
        [](JNIEnv *env, const jobject &ko) -> std::string {
            return JNIString::FromJava(env, static_cast<jstring>(ko));
        },
        [](JNIEnv *env, const jobject &vo) -> std::string {
            return JNIString::FromJava(env, static_cast<jstring>(vo));
        });
    return result;
}

bool JNIInvokeOptions::GetRetIsFundamentalType(JNIEnv *env, jobject o)
{
    jboolean value = static_cast<jboolean>(env->CallBooleanMethod(o, getRetIsFundamentalType_));
    bool result = static_cast<bool>(value);
    return result;
}

bool JNIInvokeOptions::GetNeedOrder(JNIEnv *env, jobject o)
{
    jboolean value = static_cast<jboolean>(env->CallBooleanMethod(o, getNeedOrder_));
    bool result = static_cast<bool>(value);
    return result;
}

int64_t JNIInvokeOptions::GetScheduleTimeoutMs(JNIEnv *env, jobject o)
{
    jlong value = static_cast<jlong>(env->CallLongMethod(o, getScheduleTimeoutMs_));
    int64_t result = static_cast<int64_t>(value);
    return result;
}

bool JNIInvokeOptions::GetPreemptedAllowed(JNIEnv *env, jobject o)
{
    jboolean value = static_cast<jboolean>(env->CallBooleanMethod(o, getPreemptedAllowed_));
    bool result = static_cast<bool>(value);
    return result;
}

void JNIDataObject::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/instance/DataObject");
    init_ = GetJMethod(env, clz_, "<init>", "(Ljava/lang/String;)V");
    getId_ = GetJMethod(env, clz_, "getId", "()Ljava/lang/String;");
}

void JNIDataObject::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

std::vector<YR::Libruntime::DataObject> JNIDataObject::FromJavaList(JNIEnv *env, jobject o)
{
    auto ret = YR::jni::JNIList::FromJava<YR::Libruntime::DataObject>(
        env, o, [](JNIEnv *env, const jobject &obj) { return YR::jni::JNIDataObject::FromJava(env, obj); });
    return ret;
}

YR::Libruntime::DataObject JNIDataObject::FromJava(JNIEnv *env, jobject o)
{
    return YR::Libruntime::DataObject{
        .id = GetId(env, o),
    };
}

std::string JNIDataObject::GetId(JNIEnv *env, jobject o)
{
    jstring jstr = static_cast<jstring>(env->CallObjectMethod(o, getId_));
    std::string ret = JNIString::FromJava(env, jstr);
    return ret;
}

YR::Libruntime::ErrorInfo JNIDataObject::WriteDataObject(JNIEnv *env,
                                                         std::shared_ptr<YR::Libruntime::DataObject> &dataObj,
                                                         const jbyteArray &byteBfr)
{
    RETURN_IF_NULL(
        byteBfr, YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "failed to write dataobject"));
    jbyte *native_bytes = env->GetByteArrayElements(byteBfr, nullptr);
    size_t capacity = env->GetArrayLength(byteBfr);
    dataObj = std::make_shared<Libruntime::DataObject>(0, capacity);
    (void)memset_s(dataObj->meta->MutableData(), dataObj->meta->GetSize(), 0, dataObj->meta->GetSize());
    auto errorInfo = dataObj->data->MemoryCopy((const void *)native_bytes, capacity);
    if (!errorInfo.OK()) {
        return errorInfo;
    }
    env->ReleaseByteArrayElements(byteBfr, native_bytes, JNI_ABORT);
    env->DeleteLocalRef(byteBfr);
    return YR::Libruntime::ErrorInfo();
}

jobject JNIDataObject::FromCcPtrVectorToList(JNIEnv *env,
                                             const std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &vector)
{
    return YR::jni::JNIArrayList::FromCc<std::shared_ptr<YR::Libruntime::DataObject>>(
        env, vector, [](JNIEnv *env, const std::shared_ptr<YR::Libruntime::DataObject> &dataObj_ptr) {
            jbyteArray element = nullptr;
            if (dataObj_ptr != nullptr) {
                element = env->NewByteArray(dataObj_ptr->data->GetSize());
                env->SetByteArrayRegion(element, 0, dataObj_ptr->data->GetSize(),
                                        reinterpret_cast<const jbyte *>(dataObj_ptr->data->ImmutableData()));
            }
            return element;
        });
}

void JNIErrorCode::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/errorcode/ErrorCode");
    jfInitWithInt_ = GetJMethod(env, clz_, "<init>", "(I)V");
    jfGetValue_ = GetJMethod(env, clz_, "getValue", "()I");
}

void JNIErrorCode::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

jobject JNIErrorCode::FromCc(JNIEnv *env, const YR::Libruntime::ErrorCode &errorCode)
{
    std::map<YR::Libruntime::ErrorCode, int> fieldMap = {
        {YR::Libruntime::ErrorCode::ERR_OK, 0},
        {YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, 1001},
        {YR::Libruntime::ErrorCode::ERR_RESOURCE_NOT_ENOUGH, 1002},
        {YR::Libruntime::ErrorCode::ERR_INSTANCE_NOT_FOUND, 1003},
        {YR::Libruntime::ErrorCode::ERR_INSTANCE_DUPLICATED, 1004},
        {YR::Libruntime::ErrorCode::ERR_INVOKE_RATE_LIMITED, 1005},
        {YR::Libruntime::ErrorCode::ERR_RESOURCE_CONFIG_ERROR, 1006},
        {YR::Libruntime::ErrorCode::ERR_INSTANCE_EXITED, 1007},
        {YR::Libruntime::ErrorCode::ERR_EXTENSION_META_ERROR, 1008},
        {YR::Libruntime::ErrorCode::ERR_INSTANCE_SUB_HEALTH, 1009},
        {YR::Libruntime::ErrorCode::ERR_GROUP_SCHEDULE_FAILED, 1010},
        {YR::Libruntime::ErrorCode::ERR_USER_CODE_LOAD, 2001},
        {YR::Libruntime::ErrorCode::ERR_USER_FUNCTION_EXCEPTION, 2002},
        {YR::Libruntime::ErrorCode::ERR_REQUEST_BETWEEN_RUNTIME_BUS, 3001},
        {YR::Libruntime::ErrorCode::ERR_INNER_COMMUNICATION, 3002},
        {YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, 3003},
        {YR::Libruntime::ErrorCode::ERR_DISCONNECT_FRONTEND_BUS, 3004},
        {YR::Libruntime::ErrorCode::ERR_ETCD_OPERATION_ERROR, 3005},
        {YR::Libruntime::ErrorCode::ERR_BUS_DISCONNECTION, 3006},
        {YR::Libruntime::ErrorCode::ERR_REDIS_OPERATION_ERROR, 3007},
        {YR::Libruntime::ErrorCode::ERR_INCORRECT_INIT_USAGE, 4001},
        {YR::Libruntime::ErrorCode::ERR_INIT_CONNECTION_FAILED, 4002},
        {YR::Libruntime::ErrorCode::ERR_DESERIALIZATION_FAILED, 4003},
        {YR::Libruntime::ErrorCode::ERR_INSTANCE_ID_EMPTY, 4004},
        {YR::Libruntime::ErrorCode::ERR_GET_OPERATION_FAILED, 4005},
        {YR::Libruntime::ErrorCode::ERR_INCORRECT_FUNCTION_USAGE, 4006},
        {YR::Libruntime::ErrorCode::ERR_INCORRECT_CREATE_USAGE, 4007},
        {YR::Libruntime::ErrorCode::ERR_INCORRECT_INVOKE_USAGE, 4008},
        {YR::Libruntime::ErrorCode::ERR_INCORRECT_KILL_USAGE, 4009},
        {YR::Libruntime::ErrorCode::ERR_ROCKSDB_FAILED, 4201},
        {YR::Libruntime::ErrorCode::ERR_SHARED_MEMORY_LIMITED, 4202},
        {YR::Libruntime::ErrorCode::ERR_OPERATE_DISK_FAILED, 4203},
        {YR::Libruntime::ErrorCode::ERR_INSUFFICIENT_DISK_SPACE, 4204},
        {YR::Libruntime::ErrorCode::ERR_CONNECTION_FAILED, 4205},
        {YR::Libruntime::ErrorCode::ERR_KEY_ALREADY_EXIST, 4206},
        {YR::Libruntime::ErrorCode::ERR_DEPENDENCY_FAILED, 4306},
        {YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, 4299},
        {YR::Libruntime::ErrorCode::ERR_FINALIZED, 9000},
        {YR::Libruntime::ErrorCode::ERR_CREATE_RETURN_BUFFER, 9001},
    };

    if (auto it = fieldMap.find(errorCode); it == fieldMap.end()) {
        YRLOG_ERROR("Failed to match errorcode, code: ", fmt::underlying(errorCode));
        return nullptr;
    }
    return env->NewObject(clz_, jfInitWithInt_, static_cast<jint>(fieldMap[errorCode]));
}

YR::Libruntime::ErrorCode JNIErrorCode::FromJava(JNIEnv *env, jobject o)
{
    jint errCodeValue = env->CallIntMethod(o, jfGetValue_);
    return static_cast<YR::Libruntime::ErrorCode>(errCodeValue);
}

void JNIModuleCode::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/errorcode/ModuleCode");
    jfCORE_ = GetJStaticField(env, clz_, "CORE", "Lcom/yuanrong/errorcode/ModuleCode;");
    jfRUNTIME_ = GetJStaticField(env, clz_, "RUNTIME", "Lcom/yuanrong/errorcode/ModuleCode;");
    jfRUNTIME_CREATE_ = GetJStaticField(env, clz_, "RUNTIME_CREATE", "Lcom/yuanrong/errorcode/ModuleCode;");
    jfRUNTIME_INVOKE_ = GetJStaticField(env, clz_, "RUNTIME_INVOKE", "Lcom/yuanrong/errorcode/ModuleCode;");
    jfRUNTIME_KILL_ = GetJStaticField(env, clz_, "RUNTIME_KILL", "Lcom/yuanrong/errorcode/ModuleCode;");
    jfDATASYSTEM_ = GetJStaticField(env, clz_, "DATASYSTEM", "Lcom/yuanrong/errorcode/ModuleCode;");

    joCORE_ = GetJStaticObjectField(env, clz_, jfCORE_);
    joRUNTIME_ = GetJStaticObjectField(env, clz_, jfRUNTIME_);
    joRUNTIME_CREATE_ = GetJStaticObjectField(env, clz_, jfRUNTIME_CREATE_);
    joRUNTIME_INVOKE_ = GetJStaticObjectField(env, clz_, jfRUNTIME_INVOKE_);
    joRUNTIME_KILL_ = GetJStaticObjectField(env, clz_, jfRUNTIME_KILL_);
    joDATASYSTEM_ = GetJStaticObjectField(env, clz_, jfDATASYSTEM_);
}

void JNIModuleCode::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

jobject JNIModuleCode::FromCc(JNIEnv *env, const YR::Libruntime::ModuleCode &moduleCode)
{
    std::map<YR::Libruntime::ModuleCode, jfieldID> fieldMap = {
        {YR::Libruntime::ModuleCode::CORE, jfCORE_},
        {YR::Libruntime::ModuleCode::RUNTIME, jfRUNTIME_},
        {YR::Libruntime::ModuleCode::RUNTIME_CREATE, jfRUNTIME_CREATE_},
        {YR::Libruntime::ModuleCode::RUNTIME_INVOKE, jfRUNTIME_INVOKE_},
        {YR::Libruntime::ModuleCode::RUNTIME_KILL, jfRUNTIME_KILL_},
        {YR::Libruntime::ModuleCode::DATASYSTEM, jfDATASYSTEM_}};

    if (auto it = fieldMap.find(moduleCode); it == fieldMap.end()) {
        return nullptr;
    }
    return env->GetStaticObjectField(clz_, fieldMap[moduleCode]);
}

YR::Libruntime::ModuleCode JNIModuleCode::FromJava(JNIEnv *env, jobject o)
{
    std::map<YR::Libruntime::ModuleCode, jobject> fieldMap = {
        {YR::Libruntime::ModuleCode::CORE, joCORE_},
        {YR::Libruntime::ModuleCode::RUNTIME, joRUNTIME_},
        {YR::Libruntime::ModuleCode::RUNTIME_CREATE, joRUNTIME_CREATE_},
        {YR::Libruntime::ModuleCode::RUNTIME_INVOKE, joRUNTIME_INVOKE_},
        {YR::Libruntime::ModuleCode::RUNTIME_KILL, joRUNTIME_KILL_},
        {YR::Libruntime::ModuleCode::DATASYSTEM, joDATASYSTEM_},
    };

    for (const auto &it : fieldMap) {
        if (env->IsSameObject(o, it.second)) {
            return it.first;
        }
    }
    return YR::Libruntime::ModuleCode::RUNTIME;
}

void JNILabelOperator::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/affinity/LabelOperator");
    getType_ = GetJMethod(env, clz_, "getOperateTypeValue", "()I");
    getKey_ = GetJMethod(env, clz_, "getKey", "()Ljava/lang/String;");
    getValues_ = GetJMethod(env, clz_, "getValues", "()Ljava/util/List;");
}

void JNILabelOperator::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

std::shared_ptr<YR::Libruntime::LabelOperator> JNILabelOperator::FromJava(JNIEnv *env, jobject o)
{
    int operateType = static_cast<int>(env->CallIntMethod(o, getType_));
    jstring jkey = static_cast<jstring>(env->CallObjectMethod(o, getKey_));
    auto key = JNIString::FromJava(env, jkey);
    auto values = JNIList::FromJavaToList<std::string>(
        env, env->CallObjectMethod(o, getValues_),
        [](JNIEnv *env, jobject obj) -> std::string { return JNIString::FromJava(env, static_cast<jstring>(obj)); });
    std::shared_ptr<YR::Libruntime::LabelOperator> labelOpt = nullptr;
    switch (operateType) {
        case LABEL_IN:
            labelOpt = std::make_shared<YR::Libruntime::LabelInOperator>();
            break;
        case LABEL_NOT_IN:
            labelOpt = std::make_shared<YR::Libruntime::LabelNotInOperator>();
            break;
        case LABEL_EXISTS:
            labelOpt = std::make_shared<YR::Libruntime::LabelExistsOperator>();
            break;
        case LABEL_DOES_NOT_EXIST:
            labelOpt = std::make_shared<YR::Libruntime::LabelDoesNotExistOperator>();
            break;
        default:
            YRLOG_ERROR("invalid operator type:{} ", operateType);
            YR::jni::JNILibruntimeException::ThrowNew(env, "invalid label operator type " + operateType);
            break;
    }
    labelOpt->SetKey(key);
    labelOpt->SetValues(values);
    return labelOpt;
}

void JNIAffinity::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/affinity/Affinity");
    getValue_ = GetJMethod(env, clz_, "getAffinityValue", "()I");
    getOperators_ = GetJMethod(env, clz_, "getLabelOperators", "()Ljava/util/List;");
    getAffinityScopeValue_ = GetJMethod(env, clz_, "getAffinityScopeValue", "()I");
}

std::shared_ptr<YR::Libruntime::Affinity> JNIAffinity::FromJava(JNIEnv *env, jobject o, bool preferredPriority,
                                                                bool requiredPriority, bool preferredAntiOtherLabels)
{
    int affinityValue = static_cast<int>(env->CallIntMethod(o, getValue_));
    auto labelOperators = JNIList::FromJavaToList<std::shared_ptr<YR::Libruntime::LabelOperator>>(
        env, env->CallObjectMethod(o, getOperators_),
        [](JNIEnv *env, jobject obj) -> std::shared_ptr<YR::Libruntime::LabelOperator> {
            return JNILabelOperator::FromJava(env, obj);
        });
    std::shared_ptr<YR::Libruntime::Affinity> affinity = nullptr;
    switch (affinityValue) {
        case RESOURCE_PREFERRED:
            affinity = std::make_shared<YR::Libruntime::ResourcePreferredAffinity>();
            break;
        case RESOURCE_PREFERRED_ANTI:
            affinity = std::make_shared<YR::Libruntime::ResourcePreferredAntiAffinity>();
            break;
        case RESOURCE_REQUIRED:
            affinity = std::make_shared<YR::Libruntime::ResourceRequiredAffinity>();
            break;
        case RESOURCE_REQUIRED_ANTI:
            affinity = std::make_shared<YR::Libruntime::ResourceRequiredAntiAffinity>();
            break;
        case INSTANCE_PREFERRED:
            affinity = std::make_shared<YR::Libruntime::InstancePreferredAffinity>();
            break;
        case INSTANCE_PREFERRED_ANTI:
            affinity = std::make_shared<YR::Libruntime::InstancePreferredAntiAffinity>();
            break;
        case INSTANCE_REQUIRED:
            affinity = std::make_shared<YR::Libruntime::InstanceRequiredAffinity>();
            break;
        case INSTANCE_REQUIRED_ANTI:
            affinity = std::make_shared<YR::Libruntime::InstanceRequiredAntiAffinity>();
            break;
        default:
            YRLOG_ERROR("invalid affinity type:{} ", affinityValue);
            YR::jni::JNILibruntimeException::ThrowNew(env, "invalid affinity type " + affinityValue);
            break;
    }
    if (affinity != nullptr) {
        affinity->SetLabelOperators(labelOperators);
        affinity->SetPreferredPriority(preferredPriority);
        affinity->SetRequiredPriority(requiredPriority);
        affinity->SetPreferredAntiOtherLabels(preferredAntiOtherLabels);
    }
    int affinityScope = static_cast<int>(env->CallIntMethod(o, getAffinityScopeValue_));
    switch (affinityScope) {
        case POD:
            affinity->SetAffinityScope(YR::Libruntime::AFFINITYSCOPE_POD);
            break;
        case NODE:
            affinity->SetAffinityScope(YR::Libruntime::AFFINITYSCOPE_NODE);
            break;
        default:
            YRLOG_DEBUG("affinityScope is not set.");
            break;
    }
    return affinity;
}

void JNIAffinity::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

void JNIReturnType::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/executor/ReturnType");
    if (clz_ == nullptr) {
        // client don't need this package
        return;
    }
    getErrorInfo_ = GetJMethod(env, clz_, "getErrorInfo", "()Lcom/yuanrong/errorcode/ErrorInfo;");
    getBytes_ = GetJMethod(env, clz_, "getBytes", "()[B");
}

void JNIReturnType::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

std::pair<YR::Libruntime::ErrorInfo, std::shared_ptr<YR::Libruntime::Buffer>> JNIReturnType::FromJava(JNIEnv *env,
                                                                                                      jobject o)
{
    jbyteArray jbytes = (jbyteArray)env->CallObjectMethod(o, getBytes_);

    std::shared_ptr<YR::Libruntime::Buffer> buf;
    JNIByteBuffer::WriteByteArray(env, buf, jbytes);
    env->DeleteLocalRef(jbytes);

    jobject jerrorInfo = env->CallObjectMethod(o, getErrorInfo_);
    auto errorInfo = JNIErrorInfo::FromJava(env, jerrorInfo);
    env->DeleteLocalRef(jerrorInfo);

    return std::make_pair(errorInfo, buf);
}

void JNIExistenceOpt::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/ExistenceOpt");
    j_field_NONE_ = GetJStaticField(env, clz_, "NONE", "Lcom/yuanrong/ExistenceOpt;");
    j_field_NX_ = GetJStaticField(env, clz_, "NX", "Lcom/yuanrong/ExistenceOpt;");

    j_object_field_NONE_ = GetJStaticObjectField(env, clz_, j_field_NONE_);
    j_object_field_NX_ = GetJStaticObjectField(env, clz_, j_field_NX_);
}

void JNIExistenceOpt::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

YR::Libruntime::ExistenceOpt JNIExistenceOpt::FromJava(JNIEnv *env, jobject obj)
{
    std::map<std::string, YR::Libruntime::ExistenceOpt> fieldMap = {
        {"NONE", YR::Libruntime::ExistenceOpt::NONE},
        {"NX", YR::Libruntime::ExistenceOpt::NX},
    };

    jmethodID jEnumNameMethod = GetJMethod(env, clz_, "name", "()Ljava/lang/String;");
    jstring jEnumName = (jstring)env->CallObjectMethod(obj, jEnumNameMethod);
    auto cEnumName = YR::jni::JNIString::FromJava(env, jEnumName);
    if (auto it = fieldMap.find(cEnumName); it == fieldMap.end()) {
        YRLOG_ERROR("Failed to match the java object to cpp ExistenceOpt");
        return YR::Libruntime::ExistenceOpt::NONE;
    }

    return fieldMap[cEnumName];
}

void JNIWriteMode::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/WriteMode");
    j_field_NONE_L2_CACHE_ = GetJStaticField(env, clz_, "NONE_L2_CACHE", "Lcom/yuanrong/WriteMode;");
    j_field_WRITE_THROUGH_L2_CACHE_ = GetJStaticField(env, clz_, "WRITE_THROUGH_L2_CACHE", "Lcom/yuanrong/WriteMode;");
    j_field_WRITE_BACK_L2_CACHE_ = GetJStaticField(env, clz_, "WRITE_BACK_L2_CACHE", "Lcom/yuanrong/WriteMode;");

    j_object_field_NONE_L2_CACHE_ = GetJStaticObjectField(env, clz_, j_field_NONE_L2_CACHE_);
    j_object_field_WRITE_THROUGH_L2_CACHE_ = GetJStaticObjectField(env, clz_, j_field_WRITE_THROUGH_L2_CACHE_);
    j_object_field_WRITE_BACK_L2_CACHE_ = GetJStaticObjectField(env, clz_, j_field_WRITE_BACK_L2_CACHE_);
}

void JNIWriteMode::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

YR::Libruntime::WriteMode JNIWriteMode::FromJava(JNIEnv *env, jobject obj)
{
    std::map<std::string, YR::Libruntime::WriteMode> fieldMap = {
        {"NONE_L2_CACHE", YR::Libruntime::WriteMode::NONE_L2_CACHE},
        {"WRITE_THROUGH_L2_CACHE", YR::Libruntime::WriteMode::WRITE_THROUGH_L2_CACHE},
        {"WRITE_BACK_L2_CACHE", YR::Libruntime::WriteMode::WRITE_BACK_L2_CACHE},
        {"NONE_L2_CACHE_EVICT", YR::Libruntime::WriteMode::NONE_L2_CACHE_EVICT},
    };

    jmethodID jEnumNameMethod = GetJMethod(env, clz_, "name", "()Ljava/lang/String;");
    jstring jEnumName = (jstring)env->CallObjectMethod(obj, jEnumNameMethod);
    const char *cEnumNamePtr = env->GetStringUTFChars(jEnumName, NULL);
    std::string cEnumName(cEnumNamePtr);
    env->ReleaseStringUTFChars(jEnumName, cEnumNamePtr);

    if (auto it = fieldMap.find(cEnumName); it == fieldMap.end()) {
        YRLOG_ERROR("Failed to match the java object to cpp WriteMode");
        return YR::Libruntime::WriteMode::NONE_L2_CACHE;
    }

    return fieldMap[cEnumName];
}

void JNIConsistencyType::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/ConsistencyType");
    j_field_PRAM_ = GetJStaticField(env, clz_, "PRAM", "Lcom/yuanrong/ConsistencyType;");
    j_field_CAUSAL_ = GetJStaticField(env, clz_, "CAUSAL", "Lcom/yuanrong/ConsistencyType;");

    j_object_field_PRAM_ = GetJStaticObjectField(env, clz_, j_field_PRAM_);
    j_object_field_CAUSAL_ = GetJStaticObjectField(env, clz_, j_field_CAUSAL_);
}

void JNIConsistencyType::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

YR::Libruntime::ConsistencyType JNIConsistencyType::FromJava(JNIEnv *env, jobject obj)
{
    std::map<std::string, YR::Libruntime::ConsistencyType> fieldMap = {
        {"PRAM", YR::Libruntime::ConsistencyType::PRAM},
        {"CAUSAL", YR::Libruntime::ConsistencyType::CAUSAL},
    };

    jmethodID jEnumNameMethod = GetJMethod(env, clz_, "name", "()Ljava/lang/String;");
    jstring jEnumName = (jstring)env->CallObjectMethod(obj, jEnumNameMethod);
    auto cEnumName = YR::jni::JNIString::FromJava(env, jEnumName);
    if (auto it = fieldMap.find(cEnumName); it == fieldMap.end()) {
        YRLOG_ERROR("Failed to match the java object to cpp ConsistencyType");
        return YR::Libruntime::ConsistencyType::PRAM;
    }

    return fieldMap[cEnumName];
}

void JNICacheType::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/CacheType");
    j_field_MEMORY_ = GetJStaticField(env, clz_, "MEMORY", "Lcom/yuanrong/CacheType;");
    j_field_DISK_ = GetJStaticField(env, clz_, "DISK", "Lcom/yuanrong/CacheType;");

    j_object_field_MEMORY_ = GetJStaticObjectField(env, clz_, j_field_MEMORY_);
    j_object_field_DISK_ = GetJStaticObjectField(env, clz_, j_field_DISK_);
}

void JNICacheType::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

YR::Libruntime::CacheType JNICacheType::FromJava(JNIEnv *env, jobject obj)
{
    std::map<std::string, YR::Libruntime::CacheType> fieldMap = {
        {"MEMORY", YR::Libruntime::CacheType::MEMORY},
        {"DISK", YR::Libruntime::CacheType::DISK},
    };

    jmethodID jEnumNameMethod = GetJMethod(env, clz_, "name", "()Ljava/lang/String;");
    jstring jEnumName = (jstring)env->CallObjectMethod(obj, jEnumNameMethod);
    auto cEnumName = YR::jni::JNIString::FromJava(env, jEnumName);
    if (auto it = fieldMap.find(cEnumName); it == fieldMap.end()) {
        YRLOG_ERROR("Failed to match the java object to cpp CacheType");
        return YR::Libruntime::CacheType::MEMORY;
    }

    return fieldMap[cEnumName];
}

void JNISetParam::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/SetParam");
    jGetExistence_ = GetJMethod(env, clz_, "getExistence", "()Lcom/yuanrong/ExistenceOpt;");
    jGetWriteMode_ = GetJMethod(env, clz_, "getWriteMode", "()Lcom/yuanrong/WriteMode;");
    jGetTtlSecond_ = GetJMethod(env, clz_, "getTtlSecond", "()I");
    jGetCacheType_ = GetJMethod(env, clz_, "getCacheType", "()Lcom/yuanrong/CacheType;");
}

YR::Libruntime::SetParam JNISetParam::FromJava(JNIEnv *env, jobject o)
{
    YR::Libruntime::SetParam setParam{};
    RETURN_IF_NULL(o, setParam);

    return YR::Libruntime::SetParam{
        .writeMode = GetWriteMode(env, o),
        .ttlSecond = static_cast<uint32_t>(env->CallIntMethod(o, jGetTtlSecond_)),
        .existence = GetExistenceOpt(env, o),
        .cacheType = GetCacheType(env, o),
    };
}

void JNISetParam::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

YR::Libruntime::ExistenceOpt JNISetParam::GetExistenceOpt(JNIEnv *env, jobject o)
{
    jobject existenceOpt = env->CallObjectMethod(o, JNISetParam::jGetExistence_);
    return JNIExistenceOpt::FromJava(env, existenceOpt);
}

YR::Libruntime::WriteMode JNISetParam::GetWriteMode(JNIEnv *env, jobject o)
{
    jobject writeMode = env->CallObjectMethod(o, JNISetParam::jGetWriteMode_);
    return JNIWriteMode::FromJava(env, writeMode);
}

YR::Libruntime::CacheType JNISetParam::GetCacheType(JNIEnv *env, jobject o)
{
    jobject cacheType = env->CallObjectMethod(o, JNISetParam::jGetCacheType_);
    return JNICacheType::FromJava(env, cacheType);
}

void JNIMSetParam::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/MSetParam");
    jGetExistence_ = GetJMethod(env, clz_, "getExistence", "()Lcom/yuanrong/ExistenceOpt;");
    jGetWriteMode_ = GetJMethod(env, clz_, "getWriteMode", "()Lcom/yuanrong/WriteMode;");
    jGetTtlSecond_ = GetJMethod(env, clz_, "getTtlSecond", "()I");
    jGetCacheType_ = GetJMethod(env, clz_, "getCacheType", "()Lcom/yuanrong/CacheType;");
}

YR::Libruntime::MSetParam JNIMSetParam::FromJava(JNIEnv *env, jobject o)
{
    YR::Libruntime::MSetParam mSetParam{};
    RETURN_IF_NULL(o, mSetParam);

    return YR::Libruntime::MSetParam{
        .writeMode = GetWriteMode(env, o),
        .ttlSecond = static_cast<uint32_t>(env->CallIntMethod(o, jGetTtlSecond_)),
        .existence = GetExistenceOpt(env, o),
        .cacheType = GetCacheType(env, o),
    };
}

void JNIMSetParam::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

YR::Libruntime::ExistenceOpt JNIMSetParam::GetExistenceOpt(JNIEnv *env, jobject o)
{
    jobject existenceOpt = env->CallObjectMethod(o, JNIMSetParam::jGetExistence_);
    return JNIExistenceOpt::FromJava(env, existenceOpt);
}

YR::Libruntime::WriteMode JNIMSetParam::GetWriteMode(JNIEnv *env, jobject o)
{
    jobject writeMode = env->CallObjectMethod(o, JNIMSetParam::jGetWriteMode_);
    return JNIWriteMode::FromJava(env, writeMode);
}

YR::Libruntime::CacheType JNIMSetParam::GetCacheType(JNIEnv *env, jobject o)
{
    jobject cacheType = env->CallObjectMethod(o, JNIMSetParam::jGetCacheType_);
    return JNICacheType::FromJava(env, cacheType);
}

void JNICreateParam::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/CreateParam");
    jGetConsistencyType_ = GetJMethod(env, clz_, "getConsistencyType", "()Lcom/yuanrong/ConsistencyType;");
    jGetCacheType_ = GetJMethod(env, clz_, "getCacheType", "()Lcom/yuanrong/CacheType;");
}

YR::Libruntime::CreateParam JNICreateParam::FromJava(JNIEnv *env, jobject o)
{
    YR::Libruntime::CreateParam createParam{};
    RETURN_IF_NULL(o, createParam);

    return YR::Libruntime::CreateParam{
        .consistencyType = GetConsistencyType(env, o),
        .cacheType = GetCacheType(env, o),
    };
}

void JNICreateParam::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

YR::Libruntime::ConsistencyType JNICreateParam::GetConsistencyType(JNIEnv *env, jobject o)
{
    jobject consistencyType = env->CallObjectMethod(o, JNICreateParam::jGetConsistencyType_);
    return JNIConsistencyType::FromJava(env, consistencyType);
}

YR::Libruntime::CacheType JNICreateParam::GetCacheType(JNIEnv *env, jobject o)
{
    jobject cacheType = env->CallObjectMethod(o, JNICreateParam::jGetCacheType_);
    return JNICacheType::FromJava(env, cacheType);
}

void JNIGetParam::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/GetParam");
    jGetOffset_ = GetJMethod(env, clz_, "getOffset", "()J");
    jGetSize_ = GetJMethod(env, clz_, "getSize", "()J");
}

YR::Libruntime::GetParam JNIGetParam::FromJava(JNIEnv *env, jobject o)
{
    return YR::Libruntime::GetParam{
        .offset = static_cast<uint64_t>(env->CallLongMethod(o, jGetOffset_)),
        .size = static_cast<uint64_t>(env->CallLongMethod(o, jGetSize_)),
    };
}

void JNIGetParam::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

void JNIGetParams::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/GetParams");
    jGetGetParams_ = GetJMethod(env, clz_, "getGetParams", "()Ljava/util/List;");
}

YR::Libruntime::GetParams JNIGetParams::FromJava(JNIEnv *env, jobject o)
{
    return YR::Libruntime::GetParams{
        .getParams = GetGetParamList(env, o),
    };
}

void JNIGetParams::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

std::vector<YR::Libruntime::GetParam> JNIGetParams::GetGetParamList(JNIEnv *env, jobject o)
{
    auto ret = YR::jni::JNIList::FromJava<YR::Libruntime::GetParam>(
        env, env->CallObjectMethod(o, jGetGetParams_),
        [](JNIEnv *env, const jobject &obj) { return YR::jni::JNIGetParam::FromJava(env, obj); });
    return ret;
}

void JNIInternalWaitResult::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/storage/InternalWaitResult");
    init_ = GetJMethod(env, clz_, "<init>", "(Ljava/util/List;Ljava/util/List;Ljava/util/Map;)V");
    mClz_ = LoadClass(env, "java/util/HashMap");
    mInit_ = GetJMethod(env, mClz_, "<init>", "()V");
    mPut_ = GetJMethod(env, mClz_, "put", "(Ljava/lang/Object;Ljava/lang/Object;)Ljava/lang/Object;");
}

void JNIInternalWaitResult::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
    if (mClz_) {
        env->DeleteGlobalRef(mClz_);
        mClz_ = nullptr;
    }
}

jobject JNIInternalWaitResult::FromCc(JNIEnv *env, const std::shared_ptr<YR::InternalWaitResult> waitResult)
{
    auto readyIds = waitResult->readyIds;
    auto unreadyIds = waitResult->unreadyIds;
    auto exceptionIds = waitResult->exceptionIds;
    auto jreadyList = JNIArrayList::FromCc<std::string>(
        env, readyIds, [](JNIEnv *env, const std::string &s) { return env->NewStringUTF(s.c_str()); });
    auto junreadyList = JNIArrayList::FromCc<std::string>(
        env, unreadyIds, [](JNIEnv *env, const std::string &s) { return env->NewStringUTF(s.c_str()); });
    jobject jmap = env->NewObject(mClz_, mInit_);
    for (auto it = exceptionIds.begin(); it != exceptionIds.end(); ++it) {
        jstring key = JNIString::FromCc(env, it->first);
        jobject err = JNIErrorInfo::FromCc(env, it->second);
        env->CallObjectMethod(jmap, mPut_, key, err);
        env->DeleteLocalRef(key);
        env->DeleteLocalRef(err);
    }
    jobject internalWaitRes = env->NewObject(clz_, init_, jreadyList, junreadyList, jmap);
    env->DeleteLocalRef(jreadyList);
    env->DeleteLocalRef(junreadyList);
    env->DeleteLocalRef(jmap);
    return internalWaitRes;
}

void JNIPair::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/errorcode/Pair");
    jmGetFirst_ = GetJMethod(env, clz_, "getFirst", "()Ljava/lang/Object;");
    jmGetSecond_ = GetJMethod(env, clz_, "getSecond", "()Ljava/lang/Object;");
}

void JNIPair::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

jobject JNIPair::CreateJPair(JNIEnv *env, jobject first, jobject second)
{
    jmethodID pairConstructor = env->GetMethodID(clz_, "<init>", "(Ljava/lang/Object;Ljava/lang/Object;)V");
    jobject pairObject = env->NewObject(clz_, pairConstructor, first, second);
    if (pairObject == nullptr) {
        YR::jni::JNILibruntimeException::ThrowNew(env, "Failed to create java Pair object");
        return pairObject;
    }
    return pairObject;
}

jobject JNIPair::GetFirst(JNIEnv *env, jobject jpair)
{
    return env->CallObjectMethod(jpair, jmGetFirst_);
}

jobject JNIPair::GetSecond(JNIEnv *env, jobject jpair)
{
    return env->CallObjectMethod(jpair, jmGetSecond_);
}

void JNIYRAutoInitInfo::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/jni/YRAutoInitInfo");
    init_ = GetJMethod(env, clz_, "<init>", "(Ljava/lang/String;Ljava/lang/String;Z)V");
    jmGetServerAddr = GetJMethod(env, clz_, "getServerAddr", "()Ljava/lang/String;");
    jmGetDsAddr = GetJMethod(env, clz_, "getDsAddr", "()Ljava/lang/String;");
    jmGetInCluster = GetJMethod(env, clz_, "isInCluster", "()Z");
}

void JNIYRAutoInitInfo::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

jobject JNIYRAutoInitInfo::FromCc(JNIEnv *env, YR::Libruntime::ClusterAccessInfo info)
{
    jstring jServerAddr = env->NewStringUTF(info.serverAddr.c_str());
    jstring jDsAddr = env->NewStringUTF(info.dsAddr.c_str());

    jobject yrAutoInitInfo = env->NewObject(clz_, init_, jServerAddr, jDsAddr, info.inCluster);
    if (yrAutoInitInfo == nullptr) {
        YRLOG_WARN("Failed to create Java object of com/yuanrong/jni/YRAutoInitInfo");
    }

    env->DeleteLocalRef(jServerAddr);
    env->DeleteLocalRef(jDsAddr);
    return yrAutoInitInfo;
}

YR::Libruntime::ClusterAccessInfo JNIYRAutoInitInfo::FromJava(JNIEnv *env, jobject obj)
{
    YR::Libruntime::ClusterAccessInfo info{
        .serverAddr = JNIString::FromJava(env, static_cast<jstring>(env->CallObjectMethod(obj, jmGetServerAddr))),
        .dsAddr = JNIString::FromJava(env, static_cast<jstring>(env->CallObjectMethod(obj, jmGetDsAddr))),
        .inCluster = static_cast<bool>(env->CallBooleanMethod(obj, jmGetInCluster))};
    return info;
}

void JNIFunctionLog::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/libruntime/generated/Socket$FunctionLog");
    jmGetLevel_ = GetJMethod(env, clz_, "getLevel", "()Ljava/lang/String;");
    jmGetTimestamp_ = GetJMethod(env, clz_, "getTimestamp", "()Ljava/lang/String;");
    jmGetContent_ = GetJMethod(env, clz_, "getContent", "()Ljava/lang/String;");
    jmGetInvokeID_ = GetJMethod(env, clz_, "getInvokeID", "()Ljava/lang/String;");
    jmGetTraceID_ = GetJMethod(env, clz_, "getTraceID", "()Ljava/lang/String;");
    jmGetStage_ = GetJMethod(env, clz_, "getStage", "()Ljava/lang/String;");
    jmGetLogType_ = GetJMethod(env, clz_, "getLogType", "()Ljava/lang/String;");
    jmGetFunctionInfo_ = GetJMethod(env, clz_, "getFunctionInfo", "()Ljava/lang/String;");
    jmGetInstanceId_ = GetJMethod(env, clz_, "getInstanceId", "()Ljava/lang/String;");
    jmGetLogSource_ = GetJMethod(env, clz_, "getLogSource", "()Ljava/lang/String;");
    jmGetLogGroupId_ = GetJMethod(env, clz_, "getLogGroupId", "()Ljava/lang/String;");
    jmGetLogStreamId_ = GetJMethod(env, clz_, "getLogStreamId", "()Ljava/lang/String;");
    jmGetErrorCode_ = GetJMethod(env, clz_, "getErrorCode", "()I");
    jmIsStart_ = GetJMethod(env, clz_, "getIsStart", "()Z");
    jmIsFinish_ = GetJMethod(env, clz_, "getIsFinish", "()Z");
}

void JNIFunctionLog::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
}

FunctionLog JNIFunctionLog::FromJava(JNIEnv *env, jobject obj)
{
    bool isStart = static_cast<bool>(env->CallBooleanMethod(obj, jmIsStart_));
    bool isFinish = static_cast<bool>(env->CallBooleanMethod(obj, jmIsFinish_));
    FunctionLog functionLog;
    functionLog.set_level(GetLevel(env, obj));
    functionLog.set_timestamp(GetTimestamp(env, obj));
    functionLog.set_content(GetContent(env, obj));
    functionLog.set_invokeid(GetInvokeID(env, obj));
    functionLog.set_traceid(GetTraceID(env, obj));
    functionLog.set_stage(GetStage(env, obj));
    functionLog.set_logtype(GetLogType(env, obj));
    functionLog.set_functioninfo(GetFunctionInfo(env, obj));
    functionLog.set_instanceid(GetInstanceId(env, obj));
    functionLog.set_logsource(GetLogSource(env, obj));
    functionLog.set_loggroupid(GetLogGroupId(env, obj));
    functionLog.set_logstreamid(GetLogStreamId(env, obj));
    functionLog.set_errorcode(GetErrorCode(env, obj));
    functionLog.set_isstart(isStart);
    functionLog.set_isfinish(isFinish);
    return functionLog;
}

std::string JNIFunctionLog::GetLevel(JNIEnv *env, jobject obj)
{
    jstring jstr = static_cast<jstring>(env->CallObjectMethod(obj, jmGetLevel_));
    std::string ret = JNIString::FromJava(env, jstr);
    return ret;
}

std::string JNIFunctionLog::GetTimestamp(JNIEnv *env, jobject obj)
{
    jstring jstr = static_cast<jstring>(env->CallObjectMethod(obj, jmGetTimestamp_));
    std::string ret = JNIString::FromJava(env, jstr);
    return ret;
}

std::string JNIFunctionLog::GetContent(JNIEnv *env, jobject obj)
{
    jstring jstr = static_cast<jstring>(env->CallObjectMethod(obj, jmGetContent_));
    std::string ret = JNIString::FromJava(env, jstr);
    return ret;
}

std::string JNIFunctionLog::GetInvokeID(JNIEnv *env, jobject obj)
{
    jstring jstr = static_cast<jstring>(env->CallObjectMethod(obj, jmGetInvokeID_));
    std::string ret = JNIString::FromJava(env, jstr);
    return ret;
}

std::string JNIFunctionLog::GetTraceID(JNIEnv *env, jobject obj)
{
    jstring jstr = static_cast<jstring>(env->CallObjectMethod(obj, jmGetTraceID_));
    std::string ret = JNIString::FromJava(env, jstr);
    return ret;
}

std::string JNIFunctionLog::GetStage(JNIEnv *env, jobject obj)
{
    jstring jstr = static_cast<jstring>(env->CallObjectMethod(obj, jmGetStage_));
    std::string ret = JNIString::FromJava(env, jstr);
    return ret;
}

std::string JNIFunctionLog::GetLogType(JNIEnv *env, jobject obj)
{
    jstring jstr = static_cast<jstring>(env->CallObjectMethod(obj, jmGetLogType_));
    std::string ret = JNIString::FromJava(env, jstr);
    return ret;
}

std::string JNIFunctionLog::GetFunctionInfo(JNIEnv *env, jobject obj)
{
    jstring jstr = static_cast<jstring>(env->CallObjectMethod(obj, jmGetFunctionInfo_));
    std::string ret = JNIString::FromJava(env, jstr);
    return ret;
}

std::string JNIFunctionLog::GetInstanceId(JNIEnv *env, jobject obj)
{
    jstring jstr = static_cast<jstring>(env->CallObjectMethod(obj, jmGetInstanceId_));
    std::string ret = JNIString::FromJava(env, jstr);
    return ret;
}

std::string JNIFunctionLog::GetLogSource(JNIEnv *env, jobject obj)
{
    jstring jstr = static_cast<jstring>(env->CallObjectMethod(obj, jmGetLogSource_));
    std::string ret = JNIString::FromJava(env, jstr);
    return ret;
}

std::string JNIFunctionLog::GetLogGroupId(JNIEnv *env, jobject obj)
{
    jstring jstr = static_cast<jstring>(env->CallObjectMethod(obj, jmGetLogGroupId_));
    std::string ret = JNIString::FromJava(env, jstr);
    return ret;
}

std::string JNIFunctionLog::GetLogStreamId(JNIEnv *env, jobject obj)
{
    jstring jstr = static_cast<jstring>(env->CallObjectMethod(obj, jmGetLogStreamId_));
    std::string ret = JNIString::FromJava(env, jstr);
    return ret;
}

int JNIFunctionLog::GetErrorCode(JNIEnv *env, jobject obj)
{
    jint value = static_cast<jint>(env->CallIntMethod(obj, jmGetErrorCode_));
    int result = static_cast<int>(value);
    return result;
}

void JNIProducerConfig::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/stream/ProducerConfig");
    jmGetDelayFlushTimeMs_ = GetJMethod(env, clz_, "getDelayFlushTimeMs", "()J");
    jmGetPageSizeByte_ = GetJMethod(env, clz_, "getPageSizeByte", "()J");
    jmGetMaxStreamSize_ = GetJMethod(env, clz_, "getMaxStreamSize", "()J");
    jmGetAutoCleanup_ = GetJMethod(env, clz_, "isAutoCleanup", "()Z");
    jmGetEncryptStream_ = GetJMethod(env, clz_, "isEncryptStream", "()Z");
    jmGetRetainForNumConsumers_ = GetJMethod(env, clz_, "getRetainForNumConsumers", "()J");
    jmGetReserveSize_ = GetJMethod(env, clz_, "getReserveSize", "()J");
    jmGetExtendConfig_ = GetJMethod(env, clz_, "getExtendConfig", "()Ljava/util/Map;");
}

void JNIProducerConfig::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
    }
}

YR::Libruntime::ProducerConf JNIProducerConfig::FromJava(JNIEnv *env, jobject obj)
{
    YR::Libruntime::ProducerConf producerConf;
    producerConf.delayFlushTime = GetDelayFlushTimeMs(env, obj);
    producerConf.pageSize = GetPageSizeByte(env, obj);
    producerConf.maxStreamSize = GetMaxStreamSize(env, obj);
    producerConf.autoCleanup = GetAutoCleanup(env, obj);
    producerConf.encryptStream = GetEncryptStream(env, obj);
    producerConf.retainForNumConsumers = GetRetainForNumConsumers(env, obj);
    producerConf.reserveSize = GetReserveSize(env, obj);
    producerConf.extendConfig = GetExtendConfig(env, obj);
    return producerConf;
}

int64_t JNIProducerConfig::GetDelayFlushTimeMs(JNIEnv *env, jobject obj)
{
    jlong value = static_cast<jlong>(env->CallLongMethod(obj, jmGetDelayFlushTimeMs_));
    int64_t result = static_cast<int64_t>(value);
    return result;
}

int64_t JNIProducerConfig::GetPageSizeByte(JNIEnv *env, jobject obj)
{
    jlong value = static_cast<jlong>(env->CallLongMethod(obj, jmGetPageSizeByte_));
    int64_t result = static_cast<int64_t>(value);
    return result;
}

uint64_t JNIProducerConfig::GetMaxStreamSize(JNIEnv *env, jobject obj)
{
    jlong value = static_cast<jlong>(env->CallLongMethod(obj, jmGetMaxStreamSize_));
    uint64_t result = static_cast<uint64_t>(value);
    return result;
}

bool JNIProducerConfig::GetAutoCleanup(JNIEnv *env, jobject obj)
{
    jboolean value = static_cast<jboolean>(env->CallBooleanMethod(obj, jmGetAutoCleanup_));
    bool result = static_cast<bool>(value);
    return result;
}

bool JNIProducerConfig::GetEncryptStream(JNIEnv *env, jobject obj)
{
    jboolean value = static_cast<jboolean>(env->CallBooleanMethod(obj, jmGetEncryptStream_));
    bool result = static_cast<bool>(value);
    return result;
}

uint64_t JNIProducerConfig::GetRetainForNumConsumers(JNIEnv *env, jobject obj)
{
    jlong value = static_cast<jlong>(env->CallLongMethod(obj, jmGetRetainForNumConsumers_));
    uint64_t result = static_cast<uint64_t>(value);
    return result;
}

uint64_t JNIProducerConfig::GetReserveSize(JNIEnv *env, jobject obj)
{
    jlong value = static_cast<jlong>(env->CallLongMethod(obj, jmGetReserveSize_));
    uint64_t result = static_cast<uint64_t>(value);
    return result;
}

std::unordered_map<std::string, std::string> JNIProducerConfig::GetExtendConfig(JNIEnv *env, jobject obj)
{
    std::unordered_map<std::string, std::string> result = JNIMap::FromJava<std::string, std::string>(
        env, env->CallObjectMethod(obj, jmGetExtendConfig_),
        [](JNIEnv *env, const jobject &ko) -> std::string {
            return JNIString::FromJava(env, static_cast<jstring>(ko));
        },
        [](JNIEnv *env, const jobject &vo) -> std::string {
            return JNIString::FromJava(env, static_cast<jstring>(vo));
        });
    return result;
}

void JNINode::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "com/yuanrong/api/Node");
    init_ = GetJMethod(env, clz_, "<init>", "()V");
    jmInit_ = GetJMethod(env, clz_, "<init>", "(Ljava/lang/String;ZLjava/util/Map;Ljava/util/Map;)V");
}

void JNINode::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
    }
}

jobject JNINode::GetResourcesFromResourceUnit(JNIEnv *env, const YR::Libruntime::ResourceUnit &resourceUnit)
{
    return JNIHashMap::FromCc<std::string, float>(
        env, resourceUnit.capacity, [](JNIEnv *env, const std::string &key) { return env->NewStringUTF(key.c_str()); },
        [](JNIEnv *env, const float &value) { return JNIFloat::FromCc(env, value); });
}

jobject JNINode::GetLabelsFromResourceUnit(JNIEnv *env, const YR::Libruntime::ResourceUnit &resourceUnit)
{
    return JNIHashMap::FromCc<std::string, std::vector<std::string>>(
        env, resourceUnit.nodeLabels,
        [](JNIEnv *env, const std::string &key) { return env->NewStringUTF(key.c_str()); },
        [](JNIEnv *env, const std::vector<std::string> &value) {
            return JNIArrayList::FromCc<std::string>(
                env, value, [](JNIEnv *env, const std::string &s) { return env->NewStringUTF(s.c_str()); });
        });
}

jobject JNINode::FromCc(JNIEnv *env, const YR::Libruntime::ResourceUnit &resourceUnit)
{
    jstring jId = JNIString::FromCc(env, resourceUnit.id);
    jboolean jAlive = (resourceUnit.status == 0) ? JNI_TRUE : JNI_FALSE;
    jobject jResources = JNINode::GetResourcesFromResourceUnit(env, resourceUnit);
    jobject jLabels = JNINode::GetLabelsFromResourceUnit(env, resourceUnit);
    jobject result = env->NewObject(clz_, jmInit_, jId, jAlive, jResources, jLabels);

    env->DeleteLocalRef(jId);
    env->DeleteLocalRef(jResources);
    env->DeleteLocalRef(jLabels);
    return result;
}
}  // namespace jni
}  // namespace YR
