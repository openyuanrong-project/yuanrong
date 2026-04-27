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

package org.yuanrong.errorcode;

import lombok.Data;

/**
 * The type Pair.
 *
 * @param <T> the type parameter
 * @param <U> the type parameter
 * @since 2023/11/6
 */
@Data
public class Pair<T, U> {
    /**
     * The First.
     */
    public T first;

    /**
     * The Second.
     */
    public U second;

    /**
     * Instantiates a new Pair.
     *
     * @param f the f
     * @param s the s
     */
    public Pair(T f, U s) {
        first = f;
        second = s;
    }
}
