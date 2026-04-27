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

#include <iostream>

#include <jni.h>
#include "jni_types.h"
#include "jni_errorinfo.h"
#include "jni_stacktrace_info.h"
#include "jni_stacktrace_element.h"
#include "jni_function_meta.h"

#define CURRENT_JNI_VERSION JNI_VERSION_1_8

#ifdef __cplusplus
extern "C" {
#endif

jint JNI_OnLoad(JavaVM *vm, void *reserved);
void JNI_OnUnload(JavaVM *vm, void *reserved);

#ifdef __cplusplus
}
#endif