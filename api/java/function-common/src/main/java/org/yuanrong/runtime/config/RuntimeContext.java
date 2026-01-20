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

package org.yuanrong.runtime.config;

/**
 * The type RuntimeContext
 *
 * @since 2024 /02/30
 */
public class RuntimeContext {
    /**
     * Flag indicates runtime's runtime id.
     */
    public static String runtimeId = "";

    /**
     * Flag indicates runtime's runtime context, include threadid and tenant context.
     */
    public static final ThreadLocal<String> RUNTIME_CONTEXT = new ThreadLocal<String>() {
        @Override
        protected String initialValue() {
            return "";
        }
    };
}