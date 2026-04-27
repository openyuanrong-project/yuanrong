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
#include "jni_function_meta.h"

namespace YR {
namespace jni {
void JNIFunctionMeta::Init(JNIEnv *env)
{
    clz_ = LoadClass(env, "org/yuanrong/libruntime/generated/Libruntime$FunctionMeta");
    factoryClz_ = LoadClass(env, "org/yuanrong/instance/FunctionMetaFactory");
    clzLanguageType_ = LoadClass(env, "org/yuanrong/libruntime/generated/Libruntime$LanguageType");
    clzApiType_ = LoadClass(env, "org/yuanrong/libruntime/generated/Libruntime$ApiType");
    init_ = GetStaticMethodID(
        env, factoryClz_, "getFunctionMeta",
        "(Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Lorg/yuanrong/"
        "libruntime/generated/Libruntime$LanguageType;Lorg/yuanrong/libruntime/generated/Libruntime$ApiType;"
        "Ljava/lang/String;)Lorg/yuanrong/libruntime/generated/Libruntime$FunctionMeta;");
    getFuncName_ = GetJMethod(env, clz_, "getFunctionName", "()Ljava/lang/String;");
    getFunctionID_ = GetJMethod(env, clz_, "getFunctionID", "()Ljava/lang/String;");
    getSignature_ = GetJMethod(env, clz_, "getSignature", "()Ljava/lang/String;");
    getClassName_ = GetJMethod(env, clz_, "getClassName", "()Ljava/lang/String;");
    getAppName_ = GetJMethod(env, clz_, "getApplicationName", "()Ljava/lang/String;");
    getModuleName_ = GetJMethod(env, clz_, "getModuleName", "()Ljava/lang/String;");
    getLangaugeType_ =
        GetJMethod(env, clz_, "getLanguage", "()Lorg/yuanrong/libruntime/generated/Libruntime$LanguageType;");
    getApiType_ =
        GetJMethod(env, clz_, "getApiType", "()Lorg/yuanrong/libruntime/generated/Libruntime$ApiType;");
    getName_ = GetJMethod(env, clz_, "getName", "()Ljava/lang/String;");
    getNs_ = GetJMethod(env, clz_, "getNs", "()Ljava/lang/String;");
}

void JNIFunctionMeta::Recycle(JNIEnv *env)
{
    if (clz_) {
        env->DeleteGlobalRef(clz_);
        clz_ = nullptr;
    }
    if (factoryClz_) {
        env->DeleteGlobalRef(factoryClz_);
        factoryClz_ = nullptr;
    }
}

YR::Libruntime::FunctionMeta JNIFunctionMeta::FromJava(JNIEnv *env, jobject o)
{
    RETURN_IF_NULL(o, YR::Libruntime::FunctionMeta{});
    YR::Libruntime::FunctionMeta m;
    m.appName = GetAppName(env, o);
    m.moduleName = GetModuleName(env, o);
    m.funcName = GetFuncName(env, o);
    m.className = GetClassName(env, o);
    m.languageType = GetLanguageType(env, o);
    m.signature = GetSignature(env, o);
    m.apiType = GetApiType(env, o);
    m.functionId = GetFunctionID(env, o);
    m.name = StringGetter(env, o, getName_);
    m.ns = StringGetter(env, o, getNs_);
    return m;
}

jobject JNIFunctionMeta::FromCc(JNIEnv *env, const YR::Libruntime::FunctionMeta &meta)
{
    jstring jappName = JNIString::FromCc(env, meta.appName);
    jstring jmoduleName = JNIString::FromCc(env, meta.moduleName);
    jstring jfuncName = JNIString::FromCc(env, meta.funcName);
    jstring jclassName = JNIString::FromCc(env, meta.className);
    jobject jlanguage = JNILanguageType::FromCc(env, meta.languageType);
    jstring jsignature = JNIString::FromCc(env, meta.signature);
    jobject japiType = JNIApiType::FromCc(env, meta.apiType);
    jobject obj = env->CallStaticObjectMethod(factoryClz_, init_, jappName, jmoduleName, jfuncName, jclassName,
                                              jlanguage, japiType, jsignature);
    env->DeleteLocalRef(jappName);
    env->DeleteLocalRef(jmoduleName);
    env->DeleteLocalRef(jfuncName);
    env->DeleteLocalRef(jclassName);
    env->DeleteLocalRef(jlanguage);
    env->DeleteLocalRef(jsignature);
    env->DeleteLocalRef(japiType);
    return obj;
}

std::string JNIFunctionMeta::StringGetter(JNIEnv *env, jobject o, jmethodID jm)
{
    jstring jstr = static_cast<jstring>(env->CallObjectMethod(o, jm));
    std::string ret = JNIString::FromJava(env, jstr);
    return ret;
}

std::string JNIFunctionMeta::GetAppName(JNIEnv *env, jobject o)
{
    return StringGetter(env, o, getAppName_);
}

std::string JNIFunctionMeta::GetModuleName(JNIEnv *env, jobject o)
{
    return StringGetter(env, o, getModuleName_);
}

libruntime::LanguageType JNIFunctionMeta::GetLanguageType(JNIEnv *env, jobject o)
{
    jobject languageType = env->CallObjectMethod(o, JNIFunctionMeta::getLangaugeType_);
    return JNILanguageType::FromJava(env, languageType);
}

libruntime::ApiType JNIFunctionMeta::GetApiType(JNIEnv *env, jobject o)
{
    jobject apiType = env->CallObjectMethod(o, JNIFunctionMeta::getApiType_);
    return JNIApiType::FromJava(env, apiType);
}

std::string JNIFunctionMeta::GetClassName(JNIEnv *env, jobject o)
{
    return StringGetter(env, o, getClassName_);
}

std::string JNIFunctionMeta::GetFuncName(JNIEnv *env, jobject o)
{
    return StringGetter(env, o, getFuncName_);
}

std::string JNIFunctionMeta::GetFunctionID(JNIEnv *env, jobject o)
{
    return StringGetter(env, o, getFunctionID_);
}

std::string JNIFunctionMeta::GetSignature(JNIEnv *env, jobject o)
{
    return StringGetter(env, o, getSignature_);
}
}  // namespace jni
}  // namespace YR