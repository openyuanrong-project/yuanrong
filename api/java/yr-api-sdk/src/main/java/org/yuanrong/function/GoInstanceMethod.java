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
 * @since 2024.03.12
 */
public class GoInstanceMethod<R> {
    /**
     * the name of this instance method.
     */
    public final String methodName;

    /**
     * Type of the return value of the instance method.
     */
    public final Class<R> returnType;

    /**
     * Number of the return value of the instance method
     */
    private int returnNum = 1;

    /**
     * Constructor of GoInstanceMethod.
     *
     * @param methodName the name of the instance method.
     * @param returnType the return type of method function.
     * @param returnNum Number of the return values of this function.
     */
    GoInstanceMethod(String methodName, Class<R> returnType, int returnNum) {
        this.methodName = methodName;
        this.returnType = returnType;
        this.returnNum = returnNum;
    }

    /**
     * Create a go instance method.
     *
     * @param methodName The name of the instance method.
     * @param returnNum Number of the return values of this function.
     * @return a go instance method.
     */
    public static GoInstanceMethod<Object> of(String methodName, int returnNum) {
        return new GoInstanceMethod<>(methodName, Object.class, returnNum);
    }

    /**
     * Create a go instance method.
     *
     * @param methodName The name of this instance method.
     * @param returnType Class of the return value of the instance method.
     * @param returnNum Number of the return values of this function.
     * @return a go instance method.
     */
    public static <R> GoInstanceMethod<R> of(String methodName, Class<R> returnType, int returnNum) {
        return new GoInstanceMethod<>(methodName, returnType, returnNum);
    }
}
