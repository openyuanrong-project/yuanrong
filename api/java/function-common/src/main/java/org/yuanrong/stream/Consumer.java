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

package org.yuanrong.stream;

import org.yuanrong.exception.YRException;

import java.util.List;

/**
 * Consumer interface class.
 *
 * @class Consumer Consumer.java "function-common/src/main/java/com/yuanrong/stream/Consumer.java".
 *
 * @since 2024/09/04
 */
public interface Consumer {
    /**
     * The consumer receives data with a subscription function. The consumer waits for expectNum elements.
     * The call returns when the timeout time timeoutMs is reached or the expected number of data is received.
     *
     * @param expectNum The number of elements expected to be received.
     * @param timeoutMs Timeout for receiving.
     * @return List<Element> A list of Elements that store data.
     * @throws YRException Unified exception types thrown.
     */
    List<Element> receive(long expectNum, int timeoutMs) throws YRException;

    /**
     * The consumer receives data with a subscription function. The call returns when the timeout time timeoutMs is
     * reached.
     *
     * @param timeoutMs Timeout for receiving.
     * @return List<Element> A list of Elements that store data.
     * @throws YRException Unified exception types thrown.
     */
    List<Element> receive(int timeoutMs) throws YRException;

    /**
     * After a consumer finishes using an element identified by a certain elementId, it needs to confirm that it has
     * finished consuming, so that each worker can obtain information on whether all consumers have finished consuming.
     * If a certain page has been consumed, the internal memory recovery mechanism can be triggered. If not ack it will
     * be automatically ack when the consumer exits.
     *
     * @param elementId The id of the consumed element to be confirmed.
     * @throws YRException Unified exception types thrown.
     */
    void ack(long elementId) throws YRException;

    /**
     * Close the consumer. Once closed, the consumer cannot be used.
     *
     * @throws YRException Unified exception types thrown.
     */
    void close() throws YRException;
}
