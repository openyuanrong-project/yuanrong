/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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

package org.yuanrong.function;

import org.yuanrong.api.FunctionLanguage;

/**
 * Java function calls cross-language function.
 *
 * @since 2022/08/26
 */
public class CrossLangFunction<R> {
    /**
     * The name of this class.
     */
    public final String className;

    /**
     * The name of this function.
     */
    public final String functionName;

    /**
     * Type of the return value of this function.
     */
    public final Class<R> returnType;

    /**
     * Language type of Function.
     */
    public final FunctionLanguage language;

    /**
     * Constructor of CrossLangFunction.
     *
     * @param language cross-language type.
     * @param className the class name.
     * @param functionName the method name.
     * @param returnType return type of the method.
     */
    protected CrossLangFunction(FunctionLanguage language, String className, String functionName, Class<R> returnType) {
        this.language = language;
        this.className = className;
        this.functionName = functionName;
        this.returnType = returnType;
    }
}

