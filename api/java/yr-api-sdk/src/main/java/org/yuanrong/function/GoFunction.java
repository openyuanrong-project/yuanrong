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
 * Helper class to save function name and return type.
 *
 * @since 2024.03.12
 */
public class GoFunction<R> {
    /**
     * The name of this function
     */
    public final String functionName;

    /**
     * Type of the return value of this function
     */
    public final Class<R> returnType;

    /**
     * Number of the return value of this function
     */
    public final int returnNum;

    private GoFunction(String functionName, Class<R> returnType, int returnNum) {
        this.functionName = functionName;
        this.returnType = returnType;
        this.returnNum = returnNum;
    }

    /**
     * Create a go function.
     *
     * @param functionName The name of this function
     * @param returnNum Number of the return values of this function
     * @return a go function.
     */
    public static GoFunction<Object> of(String functionName, int returnNum) {
        return of(functionName, Object.class, returnNum);
    }

    /**
     * Create a go function.
     *
     * @param functionName The name of this function
     * @param returnType Class of the return value of this function
     * @param returnNum Number of the return values of this function
     * @return a go function.
     */
    public static <R> GoFunction<R> of(String functionName, Class<R> returnType, int returnNum) {
        return new GoFunction<>(functionName, returnType, returnNum);
    }
}
