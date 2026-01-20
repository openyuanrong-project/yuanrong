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
 * The interface Yr func 5.
 *
 * @param <T0> the type parameter
 * @param <T1> the type parameter
 * @param <T2> the type parameter
 * @param <T3> the type parameter
 * @param <T4> the type parameter
 * @param <R> the type parameter
 * @since 2022 /08/30
 */
@FunctionalInterface
public interface YRFunc5<T0, T1, T2, T3, T4, R> extends YRFuncR<R> {
    /**
     * Apply r.
     *
     * @param t0 the t 0
     * @param t1 the t 1
     * @param t2 the t 2
     * @param t3 the t 3
     * @param t4 the t 4
     * @return the r
     * @throws Exception the exception
     */
    R apply(T0 t0, T1 t1, T2 t2, T3 t3, T4 t4) throws Exception;
}
