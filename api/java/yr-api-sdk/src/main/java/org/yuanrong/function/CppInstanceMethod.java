/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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
 * @since 2022/08/26
 */
public class CppInstanceMethod<R> {
    /**
     * the name of this instance method.
     */
    public final String methodName;

    /**
     * Type of the return value of the instance method.
     */
    public final Class<R> returnType;

    /**
     * Constructor of CppInstanceMethod.
     *
     * @param methodName the name of the instance method.
     * @param returnType the return type of method function.
     */
    CppInstanceMethod(String methodName, Class<R> returnType) {
        this.methodName = methodName;
        this.returnType = returnType;
    }

    /**
     * Java function calls C++ class instance member function.
     *
     * @param methodName C++ instance method name.
     * @return Object Instance of type CppInstanceMethod.
     */
    public static CppInstanceMethod<Object> of(String methodName) {
        return new CppInstanceMethod<>(methodName, Object.class);
    }

    /**
     * Java function calls C++ class instance member function.
     *
     * @param methodName C++ instance method name.
     * @param returnType Return type of C++ class instance method.
     * @param <R> The type of the return value of the instance method.
     * @return CppInstanceMethod Instance.
     *
     * @snippet{trimleft} CppFunctionExample.java CppInstanceMethod 样例代码
     */
    public static <R> CppInstanceMethod<R> of(String methodName, Class<R> returnType) {
        return new CppInstanceMethod<>(methodName, returnType);
    }
}
