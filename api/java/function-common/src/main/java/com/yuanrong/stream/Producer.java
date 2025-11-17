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

package com.yuanrong.stream;

import com.yuanrong.exception.YRException;

/**
 * Producer interface class.
 *
 * @class Producer Producer.java "function-common/src/main/java/com/yuanrong/stream/Producer.java".
 *
 * @since 2024/09/04
 */
public interface Producer {
    /**
     * The producer sends data, which is first placed in a buffer. The buffer is flushed according to the configured
     * automatic flush policy (send at a certain interval or when the buffer is full) or by actively calling flush to
     * allow consumers to access it.
     *
     * @param element The Element data to be sent. Element can refer to the Element object structure in the public
     *                structure.
     * @throws YRException Unified exception types thrown.
     */
    void send(Element element) throws YRException;

    /**
     * The producer sends data, which is first placed in a buffer. The buffer is flushed according to the configured
     * automatic flush policy (send at a certain interval or when the buffer is full) or by actively calling flush to
     * allow consumers to access it.
     *
     * @param element The Element data to be sent. Element can refer to the Element object structure in the public
     *                structure.
     * @param timeoutMs Timeout period.
     * @throws YRException Unified exception types thrown.
     */
    void send(Element element, int timeoutMs) throws YRException;

    /**
     * Closing a producer triggers an automatic flush of the data buffer and indicates that the data buffer is no longer
     * in use. Once closed, the producer can no longer be used.
     *
     * @throws YRException Unified exception types thrown.
     */
    void close() throws YRException;
}
