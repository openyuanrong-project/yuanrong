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

/**
 * Helper class to invoke member function.
 *
 * @since 2024/04/16
 */
public class JavaInstanceMethod<R> {
    /**
     * the name of this instance method.
     */
    public final String methodName;

    /**
     * Type of the return value of the instance method.
     */
    public final Class<R> returnType;

    /**
     * Constructor of JavaInstanceMethod.
     *
     * @param methodName the name of the instance method.
     * @param returnType the return type of method function.
     */
    JavaInstanceMethod(String methodName, Class<R> returnType) {
        this.methodName = methodName;
        this.returnType = returnType;
    }

    /**
     * Java function calls Java class instance member function.
     *
     * @param methodName Java instance method name.
     * @return JavaInstanceMethod Instance.
     */
    public static JavaInstanceMethod<Object> of(String methodName) {
        return new JavaInstanceMethod<>(methodName, Object.class);
    }

    /**
     * Java function calls Java class instance member function.
     *
     * @param methodName Java instance method name.
     * @param returnType Java class instance method return value class.
     * @param <R> The return type of a Java class instance method.
     * @return JavaInstanceMethod Instance.
     *
     * @snippet{trimleft} JavaInstanceExample.java JavaInstanceMethod 样例代码
     */
    public static <R> JavaInstanceMethod<R> of(String methodName, Class<R> returnType) {
        return new JavaInstanceMethod<>(methodName, returnType);
    }
}
