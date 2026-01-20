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

package org.yuanrong.function;

import org.yuanrong.api.FunctionLanguage;

/**
 * Java function calls Java function.
 *
 * @since 2024/04/16
 */
public class JavaFunction<R> extends CrossLangFunction<R> {
    private JavaFunction(String className, String functionName, Class<R> returnType) {
        super(FunctionLanguage.FUNC_LANG_JAVA, className, functionName, returnType);
    }

    /**
     * Java function calls Java function.
     *
     * @param className Java class name.
     * @param functionName Java function name.
     * @return JavaFunction Instance.
     */
    public static JavaFunction<Object> of(String className, String functionName) {
        return of(className, functionName, Object.class);
    }

    /**
     * Java function calls Java function.
     *
     * @param className Java class name.
     * @param functionName Java function name.
     * @param returnType The return type of a Java method.
     * @param <R> Return value type.
     * @return JavaFunction Instance.
     *
     * @snippet{trimleft} JavaInstanceExample.java JavaFunction 样例代码
     */
    public static <R> JavaFunction<R> of(String className, String functionName, Class<R> returnType) {
        return new JavaFunction<>(className, functionName, returnType);
    }
}
