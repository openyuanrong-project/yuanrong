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

#pragma once

#include <jni.h>
#include <cassert>
#include <cstdint>
#include <functional>
#include <iostream>
#include <iterator>
#include <list>
#include <map>
#include <memory>
#include <msgpack.hpp>
#include <sstream>
#include <string>
#include <thread>
#include <unordered_map>
#include <utility>

#include "jni_errorinfo.h"
#include "jni_function_meta.h"
#include "src/dto/data_object.h"
#include "src/dto/internal_wait_result.h"
#include "src/dto/invoke_arg.h"
#include "src/dto/invoke_options.h"
#include "src/dto/resource_unit.h"
#include "src/dto/status.h"
#include "src/dto/stream_conf.h"
#include "src/libruntime/auto_init.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/libruntime_config.h"
#include "src/libruntime/libruntime_options.h"
#include "src/libruntime/objectstore/object_store.h"
#include "src/libruntime/statestore/state_store.h"
#include "src/proto/socket.pb.h"

namespace YR {
namespace jni {
using FunctionLog = ::libruntime::FunctionLog;

#define STDERR std::cerr << __FILE__ << ":" << __LINE__ << "[" << __func__ << "]"

#define CHECK_JAVA_EXCEPTION(env)                                                                              \
    do {                                                                                                       \
        jthrowable throwable = env->ExceptionOccurred();                                                       \
        if (throwable) {                                                                                       \
            std::string errorMessage = YR::jni::JNIApacheCommonsExceptionUtils::GetStackTrace(env, throwable); \
            STDERR << "An unexpected exception occurred while executing Java code "                            \
                      "from JNI ("                                                                             \
                   << __FILE__ << ":" << __LINE__ << " " << __func__ << ")."                                   \
                   << "\n"                                                                                     \
                   << errorMessage;                                                                            \
            env->DeleteLocalRef(throwable);                                                                    \
        }                                                                                                      \
    } while (false)

#define CHECK_JAVA_EXCEPTION_AND_RETURN_IF_OCCUR(env, returnValue)                                             \
    do {                                                                                                       \
        jthrowable throwable = env->ExceptionOccurred();                                                       \
        if (throwable) {                                                                                       \
            std::string errorMessage = YR::jni::JNIApacheCommonsExceptionUtils::GetStackTrace(env, throwable); \
            STDERR << "An unexpected exception occurred while executing Java code "                            \
                      "from JNI ("                                                                             \
                   << __FILE__ << ":" << __LINE__ << " " << __func__ << ")."                                   \
                   << "\n"                                                                                     \
                   << errorMessage;                                                                            \
            env->DeleteLocalRef(throwable);                                                                    \
            return (returnValue);                                                                              \
        }                                                                                                      \
    } while (false)

#define CHECK_JAVA_EXCEPTION_AND_THROW_NEW_AND_RETURN_IF_OCCUR(env, returnValue, msg)                            \
    do {                                                                                                         \
        if (auto jt = env->ExceptionOccurred(); jt != nullptr) {                                                 \
            YR::jni::JNILibruntimeException::ThrowNew(                                                           \
                env, std::string(msg) + ", " + YR::jni::JNIApacheCommonsExceptionUtils::GetStackTrace(env, jt)); \
            env->DeleteLocalRef(jt);                                                                             \
            return (returnValue);                                                                                \
        }                                                                                                        \
    } while (false)

#define RETURN_IF_NULL(obj, returnValue) \
    {                                    \
        if ((obj) == nullptr) {          \
            return (returnValue);        \
        }                                \
    }

#define RETURN_VOID_IF_NULL(obj) \
    {                            \
        if ((obj) == nullptr) {  \
            return;              \
        }                        \
    }

#define LOG_IF_JAVA_OBJECT_IS_NULL(obj, msg)                \
    if ((obj) == nullptr) {                                 \
        STDERR << #obj << " is null, " << msg << std::endl; \
    }

#define ASSERT_NOT_NULL(ptr)                                                      \
    do {                                                                          \
        if ((ptr) == nullptr) {                                                   \
            std::cerr << "Assertion failed: " << #ptr << " is null" << std::endl; \
            (void)raise(SIGINT);                                                  \
        }                                                                         \
    } while (false)

#define CHECK_NULL_THROW_NEW_AND_RETURN(env, ptr, returnValue, msg)                                    \
    if ((ptr) == nullptr) {                                                                            \
        YR::jni::JNILibruntimeException::Throw(env, YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, \
                                               YR::Libruntime::ModuleCode::RUNTIME, std::string(msg)); \
        return returnValue;                                                                            \
    }

#define CHECK_NULL_THROW_NEW_AND_RETURN_VOID(env, ptr, msg)                                            \
    if ((ptr) == nullptr) {                                                                            \
        YR::jni::JNILibruntimeException::Throw(env, YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, \
                                               YR::Libruntime::ModuleCode::RUNTIME, std::string(msg)); \
        return;                                                                                        \
    }

inline jclass LoadClass(JNIEnv *env, const std::string &className)
{
    jclass tempLocalClassRef = env->FindClass(className.c_str());
    if (tempLocalClassRef == nullptr) {
        env->DeleteLocalRef(tempLocalClassRef);
        return nullptr;
    }
    jclass ret = (jclass)env->NewGlobalRef(tempLocalClassRef);
    if (ret == nullptr) {
        env->DeleteLocalRef(tempLocalClassRef);
        return nullptr;
    }
    env->DeleteLocalRef(tempLocalClassRef);
    return ret;
}

inline jmethodID GetJMethod(JNIEnv *env, const jclass &clz, const std::string &methodName, const std::string &sig)
{
    jmethodID m = env->GetMethodID(clz, methodName.c_str(), sig.c_str());
    LOG_IF_JAVA_OBJECT_IS_NULL(m, "Failed to load " + methodName + " { " + sig + " }");
    return m;
}

inline jmethodID GetStaticMethodID(JNIEnv *env, jclass clz, const std::string &methodName, const std::string &sig)
{
    jmethodID m = env->GetStaticMethodID(clz, methodName.c_str(), sig.c_str());
    LOG_IF_JAVA_OBJECT_IS_NULL(m, "Failed to load " + methodName + " { " + sig + " }");
    return m;
}

class RAIIJavaObject {
public:
    RAIIJavaObject(JNIEnv *env, jobject o);
    jobject GetJObject();
    ~RAIIJavaObject();

protected:
    JNIEnv *env_;
    jobject obj_;
};

class JNILibruntimeException {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);

    static void ThrowNew(JNIEnv *env, const std::string &msg);
    static void Throw(JNIEnv *env, const YR::Libruntime::ErrorCode &errorCode,
                      const YR::Libruntime::ModuleCode &moduleCode, const std::string &msg);

private:
    inline static jclass clz_ = nullptr;
};

class JNIString {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static std::string FromJava(JNIEnv *env, jstring jstr);
    static const char *FromJavaToCharArray(JNIEnv *env, jstring jstr);

    static jstring FromCc(JNIEnv *env, const std::string &str);
    static jobject FromCcVectorToList(JNIEnv *env, const std::vector<std::string> &vector);
    static std::unordered_set<std::string> FromJArrayToUnorderedSet(JNIEnv *env, const jobject &objs);

private:
    inline static jclass clz_ = nullptr;
};

class JNIFloat {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);

    static jobject FromCc(JNIEnv *env, const float &arg);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID jmInit_ = nullptr;
};

class JNIList {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);

public:
    static jobject Get(JNIEnv *env, jobject inst, int idx);

    static int Size(JNIEnv *env, jobject inst);

    static void Add(JNIEnv *env, jobject inst, jobject ele);

    template <typename NativeT>
    static std::vector<NativeT> FromJava(JNIEnv *env, jobject inst,
                                         std::function<NativeT(JNIEnv *, const jobject &)> converter)
    {
        RETURN_IF_NULL(inst, std::vector<NativeT>());
        int size = Size(env, inst);
        std::vector<NativeT> ret;
        ret.reserve(size);
        for (int i = 0; i < size; i++) {
            auto element = Get(env, inst, i);
            ret.emplace_back(converter(env, element));
        }
        return ret;
    }

    template <typename NativeT>
    static std::list<NativeT> FromJavaToList(JNIEnv *env, jobject inst,
                                             std::function<NativeT(JNIEnv *, const jobject &)> converter)
    {
        RETURN_IF_NULL(inst, std::list<NativeT>());
        int size = Size(env, inst);
        std::list<NativeT> ret;

        for (int i = 0; i < size; i++) {
            auto element = Get(env, inst, i);
            ret.push_back(converter(env, element));
        }
        return ret;
    }

    // Cannot fromCc since List is an interface which cannot be instantiated.
    // Use JNIArrayList::FromCc instead.
private:
    inline static jmethodID jmSize_ = nullptr;
    inline static jmethodID jmGet_ = nullptr;
    inline static jmethodID jmAdd_ = nullptr;
    inline static jclass clz_ = nullptr;
};

class JNIArrayList : public JNIList {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);

    template <typename T>
    static jobject FromCc(JNIEnv *env, const std::vector<T> &vect,
                          std::function<jobject(JNIEnv *, const T &)> converter)
    {
        auto jlst = env->NewObject(clz_, jmInitWithCapacity_, vect.size());
        for (const T &ele : vect) {
            auto eleTmp = converter(env, ele);
            JNIList::Add(env, jlst, eleTmp);
            env->DeleteLocalRef(eleTmp);
        }
        return jlst;
    }

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID jmInit_ = nullptr;
    inline static jmethodID jmInitWithCapacity_ = nullptr;
};

class JNIApacheCommonsExceptionUtils {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static std::string GetStackTrace(JNIEnv *env, jthrowable throwable);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID jmInit_ = nullptr;
    inline static jmethodID jmGetStackTrace_ = nullptr;
};

// Between ByteBuffer and YR::Libruntime::Buffer
class JNIByteBuffer {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static jobject FromCc(JNIEnv *env, const std::shared_ptr<YR::Libruntime::DataObject> &dataObj);
    static jobject FromCcPrtVectorToList(JNIEnv *env,
                                         const std::vector<std::shared_ptr<YR::Libruntime::Buffer>> &vector);
    static std::shared_ptr<YR::Libruntime::Buffer> FromJava(JNIEnv *env, const jbyteArray &byteArray);
    static int GetByteBufferLimit(JNIEnv *env, jobject buffer);
    static void Write(JNIEnv *env, std::shared_ptr<YR::Libruntime::Buffer> &sb, const jobject &byteBfr);
    static void WriteByteArray(JNIEnv *env, std::shared_ptr<YR::Libruntime::Buffer> &sb, const jbyteArray &byteBfr);
    static void Clear(JNIEnv *env, const jobject &byteBfr);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID jmClear_ = nullptr;
};

class JNIIterator {
public:
    static void Init(JNIEnv *env);

    static void Recycle(JNIEnv *env);
    static bool HasNext(JNIEnv *env, const jobject &inst);

    static jobject Next(JNIEnv *env, const jobject &inst);

    static void ForEachJavaObject(JNIEnv *env, const jobject &inst,
                                  std::function<void(JNIEnv *, const jobject &)> traversal);
    template <typename T, typename CT>
    static CT FromJava(JNIEnv *env, const jobject &inst, std::function<T(JNIEnv *, const jobject &)> eleConvert,
                       CT &container, std::function<void(CT &, const T &)> addToContainer);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID jmHasNext_ = nullptr;
    inline static jmethodID jmNext_ = nullptr;
};

class JNIMapEntry {
public:
    static void Init(JNIEnv *env);

    static void Recycle(JNIEnv *env);

    template <typename T>
    static T GetKey(JNIEnv *env, const jobject &inst, std::function<T(JNIEnv *env, const jobject &inst)> convert);

    template <typename T>
    static T GetVal(JNIEnv *env, const jobject &inst, std::function<T(JNIEnv *env, const jobject &inst)> convert);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID jmGetKey_ = nullptr;
    inline static jmethodID jmGetVal_ = nullptr;
};

class JNISet {
public:
    static void Init(JNIEnv *env);

    static void Recycle(JNIEnv *env);

    template <typename T>
    static std::set<T> FromJava(JNIEnv *env, const jobject &obj, std::function<T(JNIEnv *, jobject)> convertElement);

    inline static jclass clz_ = nullptr;
    inline static jmethodID jmIterator_ = nullptr;
};

class JNIUnorderedSet {
public:
    static void Init(JNIEnv *env);

    static void Recycle(JNIEnv *env);

    template <typename T>
    static std::unordered_set<T> FromJava(JNIEnv *env, const jobject &obj,
                                          std::function<T(JNIEnv *, jobject)> convertElement);

    inline static jclass clz_ = nullptr;
    inline static jmethodID jmIterator_ = nullptr;
};

class JNIMap {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);

    template <typename K, typename V>
    static std::unordered_map<K, V> FromJava(JNIEnv *env, const jobject &jmap,
                                             std::function<K(JNIEnv *, const jobject &)> convertKey,
                                             std::function<V(JNIEnv *, const jobject &)> convertVal);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID jmEntrySet_ = nullptr;
};

class JNIHashMap {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    template <typename K, typename V>
    static jobject FromCc(JNIEnv *env, const std::unordered_map<K, V> &map,
                          std::function<jobject(JNIEnv *, const K &)> converterK,
                          std::function<jobject(JNIEnv *, const V &)> converterV);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID init_ = nullptr;
    inline static jmethodID jmPut_ = nullptr;
};

class JNIInvokeType {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);

    static jobject FromCc(JNIEnv *env, const libruntime::InvokeType &type);
    static libruntime::InvokeType FromJava(JNIEnv *env, const jobject obj);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID jmGetNumber_ = nullptr;
    inline static jmethodID jmForNumber_ = nullptr;
};

class JNILanguageType {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);

    static jobject FromCc(JNIEnv *env, const libruntime::LanguageType &type);
    static libruntime::LanguageType FromJava(JNIEnv *env, const jobject obj);

    // no need `fromJava` since this is a single direction invocation [from libRuntime => java]
private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID jmGetNumber_ = nullptr;
    inline static jmethodID jmForNumber_ = nullptr;
};

class JNICodeLoader {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static YR::Libruntime::ErrorInfo Load(JNIEnv *env, const std::vector<std::string> &codePaths);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID jmLoad_ = nullptr;
};

class JNICodeExecutor {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static YR::Libruntime::ErrorInfo Execute(JNIEnv *env, const YR::Libruntime::FunctionMeta &meta,
                                             const libruntime::InvokeType &invokeType,
                                             const std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &rawArgs,
                                             std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &returnValues);

    static YR::Libruntime::ErrorInfo Shutdown(JNIEnv *env, uint64_t gracePeriodSeconds);
    static YR::Libruntime::ErrorInfo ProcessInvokeResult(
        JNIEnv *env, std::shared_ptr<YR::Libruntime::Buffer> retVal,
        std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &returnValues);
    static YR::Libruntime::ErrorInfo DumpInstance(JNIEnv *env, const std::string &instanceID,
                                                  std::shared_ptr<YR::Libruntime::Buffer> &data);
    static YR::Libruntime::ErrorInfo LoadInstance(JNIEnv *env, std::shared_ptr<YR::Libruntime::Buffer> data);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID jmExecute_ = nullptr;
    inline static jmethodID jmDumpInstance_ = nullptr;
    inline static jmethodID jmLoadInstance_ = nullptr;
    inline static jmethodID jmShutdown_ = nullptr;
    inline static jmethodID jmRecover_ = nullptr;
};

class JNILibRuntimeConfig {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static YR::Libruntime::LibruntimeConfig FromJava(JNIEnv *env, const jobject &meta);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID jmIsDriver_ = nullptr;
    inline static jmethodID jmIsInCluster_ = nullptr;
    inline static jmethodID jmIsEnableMetrics_ = nullptr;
    inline static jmethodID jmIsEnableMTLS_ = nullptr;
    inline static jmethodID jmIsEnableDsEncrypt_ = nullptr;
    inline static jmethodID jmIsEnableFrontendTLS_ = nullptr;
    inline static jmethodID jmGetCertificateFilePath_ = nullptr;
    inline static jmethodID jmGetPrivateKeyPath_ = nullptr;
    inline static jmethodID jmGetDsPublicKeyContextPath_ = nullptr;
    inline static jmethodID jmGetRuntimePublicKeyContextPath_ = nullptr;
    inline static jmethodID jmGetRuntimePrivateKeyContextPath_ = nullptr;
    inline static jmethodID jmGetVerifyFilePath_ = nullptr;
    inline static jmethodID jmGetPrivateKeyPaaswd_ = nullptr;
    inline static jmethodID jmGetServerName_ = nullptr;
    inline static jmethodID jmGetFunctionSystemIpAddr_ = nullptr;
    inline static jmethodID jmGetFunctionSystemPort_ = nullptr;
    inline static jmethodID jmGetFunctionSystemRtServerIpAddr_ = nullptr;
    inline static jmethodID jmGetFunctionSystemRtServerPort_ = nullptr;
    inline static jmethodID jmGetDataSystemIpAddr_ = nullptr;
    inline static jmethodID jmGetDataSystemPort_ = nullptr;
    inline static jmethodID jmGetJobId_ = nullptr;
    inline static jmethodID jmGetRuntimeId_ = nullptr;
    inline static jmethodID jmGetFunctionIds_ = nullptr;
    inline static jmethodID jmGetFunctionUrn_ = nullptr;
    inline static jmethodID jmGetLogLevel_ = nullptr;
    inline static jmethodID jmGetLogDir_ = nullptr;
    inline static jmethodID jmGetLogFileSizeMax_ = nullptr;
    inline static jmethodID jmGetLogFileNumMax_ = nullptr;
    inline static jmethodID jmGetLogFlushInterval_ = nullptr;
    inline static jmethodID jmIsLogMerge_ = nullptr;
    inline static jmethodID jmGetMetaConfig_ = nullptr;
    inline static jmethodID jmGetRecycleTime_ = nullptr;
    inline static jmethodID jmGetMaxTaskInstanceNum_ = nullptr;
    inline static jmethodID jmGetMaxConcurrencyCreateNum_ = nullptr;
    inline static jmethodID jmGetThreadPoolSize_ = nullptr;
    inline static jmethodID jmGetLoadPaths_ = nullptr;
    inline static jmethodID jGetHttpIocThreadsNum_ = nullptr;
    inline static jmethodID jGetHttpIdleTime_ = nullptr;
    inline static jmethodID jGetRpcTimeout_ = nullptr;
    inline static jmethodID jGetTenantId_ = nullptr;
    inline static jmethodID jGetNs_ = nullptr;
    inline static jmethodID jmGetCustomEnvs_ = nullptr;
    inline static jmethodID jmGetCodePath_ = nullptr;
};

class JNIInvokeArg {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static std::vector<YR::Libruntime::InvokeArg> FromJavaList(JNIEnv *env, jobject o, std::string tenantId);
    static YR::Libruntime::InvokeArg FromJava(JNIEnv *env, jobject o, std::string tenantId);
    static bool GetIsRef(JNIEnv *env, jobject o);
    static std::string GetObjectId(JNIEnv *env, jobject o);
    static std::unordered_set<std::string> GetNestedObjects(JNIEnv *env, jobject o);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID init_ = nullptr;
    inline static jmethodID getData_ = nullptr;
    inline static jmethodID isObjectRef_ = nullptr;
    inline static jmethodID getObjectId_ = nullptr;
    inline static jmethodID getNestedObjects_ = nullptr;
};

class JNIGroupOptions {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static YR::Libruntime::GroupOpts FromJava(JNIEnv *env, jobject o);
    static int GetTimeout(JNIEnv *env, jobject o);
    static bool GetSameLifecycle(JNIEnv *env, jobject o);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID init_ = nullptr;
    inline static jmethodID getTimeout_ = nullptr;
    inline static jmethodID getSameLifecycle_ = nullptr;
};

class JNIInvokeOptions {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static YR::Libruntime::InvokeOptions FromJava(JNIEnv *env, jobject o);
    static int GetCpu(JNIEnv *env, jobject o);
    static int GetMemory(JNIEnv *env, jobject o);
    static std::unordered_map<std::string, float> GetCustomResources(JNIEnv *env, jobject o);
    static std::unordered_map<std::string, std::string> GetCustomExtensions(JNIEnv *env, jobject o);
    static std::unordered_map<std::string, std::string> GetCreateOptions(JNIEnv *env, jobject o);
    static std::unordered_map<std::string, std::string> GetPodLabels(JNIEnv *env, jobject o);
    static std::vector<std::string> GetLabels(JNIEnv *env, jobject o);
    static std::unordered_map<std::string, std::string> GetAffinity(JNIEnv *env, jobject o);
    static size_t GetRetryTimes(JNIEnv *env, jobject o);
    static size_t GetPriority(JNIEnv *env, jobject o);
    static int GetInstancePriority(JNIEnv *env, jobject o);
    static int GetRecoverRetryTimes(JNIEnv *env, jobject o);
    static std::string GetInvokeGroupName(JNIEnv *env, jobject o);
    static std::string GetTraceId(JNIEnv *env, jobject o);
    static bool GetRetIsFundamentalType(JNIEnv *env, jobject o);
    static bool GetNeedOrder(JNIEnv *env, jobject o);
    static int64_t GetScheduleTimeoutMs(JNIEnv *env, jobject o);
    static bool GetPreemptedAllowed(JNIEnv *env, jobject o);
    static std::list<std::shared_ptr<YR::Libruntime::Affinity>> GetScheduleAffinities(JNIEnv *env, jobject o,
                                                                                      bool preferredPriority,
                                                                                      bool requiredPriority,
                                                                                      bool preferredAntiOtherLabels);
    static std::unordered_map<std::string, std::string> GetEnvVars(JNIEnv *env, jobject o);
    static std::unordered_map<std::string, std::string> GetAliasParams(JNIEnv *env, jobject o);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID init_ = nullptr;
    inline static jmethodID getCpu_ = nullptr;
    inline static jmethodID getMemory_ = nullptr;
    inline static jmethodID getCustomResources_ = nullptr;
    inline static jmethodID getCustomExtensions_ = nullptr;
    inline static jmethodID getCreateOptions_ = nullptr;
    inline static jmethodID getPodLabels_ = nullptr;
    inline static jmethodID getLabels_ = nullptr;
    inline static jmethodID getAffinity_ = nullptr;
    inline static jmethodID getRetryTimes_ = nullptr;
    inline static jmethodID getPriority_ = nullptr;
    inline static jmethodID getInstancePriority_ = nullptr;
    inline static jmethodID getRecoverRetryTimes_ = nullptr;
    inline static jmethodID getInvokeGroupName_ = nullptr;
    inline static jmethodID getTraceId_ = nullptr;
    inline static jmethodID getRetIsFundamentalType_ = nullptr;
    inline static jmethodID getScheduleAffinities_ = nullptr;
    inline static jmethodID getNeedOrder_ = nullptr;
    inline static jmethodID getScheduleTimeoutMs_ = nullptr;
    inline static jmethodID getPreemptedAllowed_ = nullptr;
    inline static jmethodID isPreferredPriority_ = nullptr;
    inline static jmethodID isRequiredPriority_ = nullptr;
    inline static jmethodID isPreferredAntiOtherLabels_ = nullptr;
    inline static jmethodID getEnvVars_ = nullptr;
    inline static jmethodID getAliasParams_ = nullptr;
};

class JNIDataObject {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static std::vector<YR::Libruntime::DataObject> FromJavaList(JNIEnv *env, jobject o);
    static YR::Libruntime::DataObject FromJava(JNIEnv *env, jobject o);
    static std::string GetId(JNIEnv *env, jobject o);
    static YR::Libruntime::ErrorInfo WriteDataObject(JNIEnv *env, std::shared_ptr<YR::Libruntime::DataObject> &dataObj,
                                                     const jbyteArray &byteBfr);
    static jobject FromCcPtrVectorToList(JNIEnv *env,
                                         const std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &vector);
    static bool GetIsFundamentalType(JNIEnv *env, jobject o);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID init_ = nullptr;
    inline static jmethodID getId_ = nullptr;
    inline static jmethodID getIsFundamentalType_ = nullptr;
};

class JNIErrorCode {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static jobject FromCc(JNIEnv *env, const YR::Libruntime::ErrorCode &errorCode);
    static YR::Libruntime::ErrorCode FromJava(JNIEnv *env, jobject o);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID jfInitWithInt_ = nullptr;
    inline static jmethodID jfGetValue_ = nullptr;
};

class JNIModuleCode {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static jobject FromCc(JNIEnv *env, const YR::Libruntime::ModuleCode &moduleCode);
    static YR::Libruntime::ModuleCode FromJava(JNIEnv *env, jobject o);

private:
    inline static jclass clz_ = nullptr;

    inline static jfieldID jfCORE_ = nullptr;            // 10
    inline static jfieldID jfRUNTIME_ = nullptr;         // 20
    inline static jfieldID jfRUNTIME_CREATE_ = nullptr;  // 21
    inline static jfieldID jfRUNTIME_INVOKE_ = nullptr;  // 22
    inline static jfieldID jfRUNTIME_KILL_ = nullptr;    // 23
    inline static jfieldID jfDATASYSTEM_ = nullptr;      // 30

    inline static jobject joCORE_ = nullptr;            // 10
    inline static jobject joRUNTIME_ = nullptr;         // 20
    inline static jobject joRUNTIME_CREATE_ = nullptr;  // 21
    inline static jobject joRUNTIME_INVOKE_ = nullptr;  // 22
    inline static jobject joRUNTIME_KILL_ = nullptr;    // 23
    inline static jobject joDATASYSTEM_ = nullptr;      // 30
};

class JNIApiType {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static jobject FromCc(JNIEnv *env, const libruntime::ApiType &type);
    static libruntime::ApiType FromJava(JNIEnv *env, const jobject obj);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID jmGetNumber_ = nullptr;
    inline static jmethodID jmForNumber_ = nullptr;
};

class JNILabelOperator {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static std::shared_ptr<YR::Libruntime::LabelOperator> FromJava(JNIEnv *env, jobject o);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID getType_ = nullptr;
    inline static jmethodID getKey_ = nullptr;
    inline static jmethodID getValues_ = nullptr;
};

class JNIAffinity {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static std::shared_ptr<YR::Libruntime::Affinity> FromJava(JNIEnv *env, jobject o, bool preferredPriority,
                                                              bool requiredPriority, bool preferredAntiOtherLabels);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID getValue_ = nullptr;
    inline static jmethodID getOperators_ = nullptr;
    inline static jmethodID getAffinityScopeValue_ = nullptr;
};

class JNIReturnType {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);

    static std::pair<YR::Libruntime::ErrorInfo, std::shared_ptr<YR::Libruntime::Buffer>> FromJava(JNIEnv *env,
                                                                                                  jobject o);

public:
    inline static jclass clz_ = nullptr;

    inline static jmethodID getErrorInfo_ = nullptr;
    inline static jmethodID getBytes_ = nullptr;
};

class JNIExistenceOpt {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static YR::Libruntime::ExistenceOpt FromJava(JNIEnv *env, jobject o);

private:
    inline static jclass clz_ = nullptr;
    inline static jfieldID j_field_NONE_ = nullptr;
    inline static jfieldID j_field_NX_ = nullptr;

    inline static jobject j_object_field_NONE_ = nullptr;
    inline static jobject j_object_field_NX_ = nullptr;
};

class JNIWriteMode {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static YR::Libruntime::WriteMode FromJava(JNIEnv *env, jobject o);

private:
    inline static jclass clz_ = nullptr;
    inline static jfieldID j_field_NONE_L2_CACHE_ = nullptr;
    inline static jfieldID j_field_WRITE_THROUGH_L2_CACHE_ = nullptr;
    inline static jfieldID j_field_WRITE_BACK_L2_CACHE_ = nullptr;

    inline static jobject j_object_field_NONE_L2_CACHE_ = nullptr;
    inline static jobject j_object_field_WRITE_THROUGH_L2_CACHE_ = nullptr;
    inline static jobject j_object_field_WRITE_BACK_L2_CACHE_ = nullptr;
};

class JNIConsistencyType {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static YR::Libruntime::ConsistencyType FromJava(JNIEnv *env, jobject o);

private:
    inline static jclass clz_ = nullptr;
    inline static jfieldID j_field_PRAM_ = nullptr;
    inline static jfieldID j_field_CAUSAL_ = nullptr;

    inline static jobject j_object_field_PRAM_ = nullptr;
    inline static jobject j_object_field_CAUSAL_ = nullptr;
};

class JNICacheType {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static YR::Libruntime::CacheType FromJava(JNIEnv *env, jobject o);

private:
    inline static jclass clz_ = nullptr;
    inline static jfieldID j_field_MEMORY_ = nullptr;
    inline static jfieldID j_field_DISK_ = nullptr;

    inline static jobject j_object_field_MEMORY_ = nullptr;
    inline static jobject j_object_field_DISK_ = nullptr;
};

class JNISetParam {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static YR::Libruntime::SetParam FromJava(JNIEnv *env, jobject o);
    static YR::Libruntime::ExistenceOpt GetExistenceOpt(JNIEnv *env, jobject o);
    static YR::Libruntime::WriteMode GetWriteMode(JNIEnv *env, jobject o);
    static YR::Libruntime::CacheType GetCacheType(JNIEnv *env, jobject o);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID jGetExistence_ = nullptr;
    inline static jmethodID jGetWriteMode_ = nullptr;
    inline static jmethodID jGetTtlSecond_ = nullptr;
    inline static jmethodID jGetCacheType_ = nullptr;
};

class JNIMSetParam {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static YR::Libruntime::MSetParam FromJava(JNIEnv *env, jobject o);
    static YR::Libruntime::ExistenceOpt GetExistenceOpt(JNIEnv *env, jobject o);
    static YR::Libruntime::WriteMode GetWriteMode(JNIEnv *env, jobject o);
    static YR::Libruntime::CacheType GetCacheType(JNIEnv *env, jobject o);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID jGetExistence_ = nullptr;
    inline static jmethodID jGetWriteMode_ = nullptr;
    inline static jmethodID jGetTtlSecond_ = nullptr;
    inline static jmethodID jGetCacheType_ = nullptr;
};

class JNICreateParam {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static YR::Libruntime::CreateParam FromJava(JNIEnv *env, jobject o);
    static YR::Libruntime::ConsistencyType GetConsistencyType(JNIEnv *env, jobject o);
    static YR::Libruntime::CacheType GetCacheType(JNIEnv *env, jobject o);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID jGetConsistencyType_ = nullptr;
    inline static jmethodID jGetCacheType_ = nullptr;
};

class JNIGetParam {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static YR::Libruntime::GetParam FromJava(JNIEnv *env, jobject o);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID jGetOffset_ = nullptr;
    inline static jmethodID jGetSize_ = nullptr;
};

class JNIGetParams {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static YR::Libruntime::GetParams FromJava(JNIEnv *env, jobject o);
    static std::vector<YR::Libruntime::GetParam> GetGetParamList(JNIEnv *env, jobject o);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID jGetGetParams_ = nullptr;
};

class JNIInternalWaitResult {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static jobject FromCc(JNIEnv *env, const std::shared_ptr<YR::InternalWaitResult> waitResult);

private:
    inline static jclass clz_ = nullptr;
    inline static jclass mClz_ = nullptr;
    inline static jmethodID init_ = nullptr;
    inline static jmethodID mInit_ = nullptr;
    inline static jmethodID mPut_ = nullptr;
};

class JNIPair {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static jobject CreateJPair(JNIEnv *env, jobject first, jobject second);
    static jobject GetFirst(JNIEnv *env, jobject jpair);
    static jobject GetSecond(JNIEnv *env, jobject jpair);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID jmGetFirst_ = nullptr;
    inline static jmethodID jmGetSecond_ = nullptr;
};

class JNIYRAutoInitInfo {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static jobject FromCc(JNIEnv *env, YR::Libruntime::ClusterAccessInfo info);
    static YR::Libruntime::ClusterAccessInfo FromJava(JNIEnv *env, jobject obj);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID init_ = nullptr;
    inline static jmethodID jmGetServerAddr = nullptr;
    inline static jmethodID jmGetDsAddr = nullptr;
    inline static jmethodID jmGetInCluster = nullptr;
};

class JNIFunctionLog {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static FunctionLog FromJava(JNIEnv *env, jobject obj);
    static std::string GetLevel(JNIEnv *env, jobject o);
    static std::string GetTimestamp(JNIEnv *env, jobject obj);
    static std::string GetContent(JNIEnv *env, jobject obj);
    static std::string GetInvokeID(JNIEnv *env, jobject obj);
    static std::string GetTraceID(JNIEnv *env, jobject obj);
    static std::string GetStage(JNIEnv *env, jobject obj);
    static std::string GetLogType(JNIEnv *env, jobject obj);
    static std::string GetFunctionInfo(JNIEnv *env, jobject obj);
    static std::string GetInstanceId(JNIEnv *env, jobject obj);
    static std::string GetLogSource(JNIEnv *env, jobject obj);
    static std::string GetLogGroupId(JNIEnv *env, jobject obj);
    static std::string GetLogStreamId(JNIEnv *env, jobject obj);
    static int GetErrorCode(JNIEnv *env, jobject o);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID jmGetLevel_ = nullptr;
    inline static jmethodID jmGetTimestamp_ = nullptr;
    inline static jmethodID jmGetContent_ = nullptr;
    inline static jmethodID jmGetInvokeID_ = nullptr;
    inline static jmethodID jmGetTraceID_ = nullptr;
    inline static jmethodID jmGetStage_ = nullptr;
    inline static jmethodID jmGetLogType_ = nullptr;
    inline static jmethodID jmGetFunctionInfo_ = nullptr;
    inline static jmethodID jmGetInstanceId_ = nullptr;
    inline static jmethodID jmGetLogSource_ = nullptr;
    inline static jmethodID jmGetLogGroupId_ = nullptr;
    inline static jmethodID jmGetLogStreamId_ = nullptr;
    inline static jmethodID jmGetErrorCode_ = nullptr;
    inline static jmethodID jmIsStart_ = nullptr;
    inline static jmethodID jmIsFinish_ = nullptr;
};

class JNIProducerConfig {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static YR::Libruntime::ProducerConf FromJava(JNIEnv *env, jobject obj);
    static int64_t GetDelayFlushTimeMs(JNIEnv *env, jobject o);
    static int64_t GetPageSizeByte(JNIEnv *env, jobject obj);
    static uint64_t GetMaxStreamSize(JNIEnv *env, jobject obj);
    static bool GetAutoCleanup(JNIEnv *env, jobject obj);
    static bool GetEncryptStream(JNIEnv *env, jobject obj);
    static uint64_t GetRetainForNumConsumers(JNIEnv *env, jobject obj);
    static uint64_t GetReserveSize(JNIEnv *env, jobject obj);
    static std::unordered_map<std::string, std::string> GetExtendConfig(JNIEnv *env, jobject obj);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID jmGetDelayFlushTimeMs_ = nullptr;
    inline static jmethodID jmGetPageSizeByte_ = nullptr;
    inline static jmethodID jmGetMaxStreamSize_ = nullptr;
    inline static jmethodID jmGetAutoCleanup_ = nullptr;
    inline static jmethodID jmGetEncryptStream_ = nullptr;
    inline static jmethodID jmGetRetainForNumConsumers_ = nullptr;
    inline static jmethodID jmGetReserveSize_ = nullptr;
    inline static jmethodID jmGetExtendConfig_ = nullptr;
};

class JNINode {
public:
    static void Init(JNIEnv *env);
    static void Recycle(JNIEnv *env);
    static jobject GetResourcesFromResourceUnit(JNIEnv *env, const YR::Libruntime::ResourceUnit &resourceUnit);
    static jobject GetLabelsFromResourceUnit(JNIEnv *env, const YR::Libruntime::ResourceUnit &resourceUnit);
    static jobject FromCc(JNIEnv *env, const YR::Libruntime::ResourceUnit &resourceUnit);

private:
    inline static jclass clz_ = nullptr;
    inline static jmethodID init_ = nullptr;
    inline static jmethodID jmInit_ = nullptr;
};

}  // namespace jni
}  // namespace YR
