/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
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

package org.yuanrong.jni;

import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.Pair;
import org.yuanrong.stream.Element;

import java.util.List;

/**
 * NativeConsumer represents methods of libRuntime
 *
 * @since 2024-09-04
 */
public class JniConsumer {
    /**
     * receive
     *
     * @param consumerPtr the consumerPtr
     * @param expectNum the expectNum
     * @param timeoutMs the timeoutMs
     * @param hasExpectedNum the hasExpectedNum
     * @return pair
     */
    public static native Pair<ErrorInfo, List<Element>> receive(long consumerPtr, long expectNum, int timeoutMs,
        boolean hasExpectedNum);

    /**
     * ack
     *
     * @param consumerPtr the consumerPtr
     * @param elementId the elementId
     * @return ErrorInfo
     */
    public static native ErrorInfo ack(long consumerPtr, long elementId);

    /**
     * close
     *
     * @param consumerPtr the consumerPtr
     * @return ErrorInfo
     */
    public static native ErrorInfo close(long consumerPtr);

    /**
     * freeJNIPtrNative
     *
     * @param clientPtr the clientPtr
     * @return ErrorInfo
     */
    public static native ErrorInfo freeJNIPtrNative(long clientPtr);
}
