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

package org.yuanrong.jni;

import org.yuanrong.errorcode.ErrorInfo;

import java.nio.ByteBuffer;

/**
 * NativeProducer represents methods of libRuntime
 *
 * @since 2024-09-04
 */
public class JniProducer {
    /**
     * sendHeapBufferDefaultTimeout
     *
     * @param producerPtr producerPtr
     * @param bytes bytes
     * @param len len
     * @return ErrorInfo
     */
    public static native ErrorInfo sendHeapBufferDefaultTimeout(long producerPtr, byte[] bytes, long len);

    /**
     * sendDirectBufferDefaultTimeout
     *
     * @param producerPtr producerPtr
     * @param buffers buffers
     * @return ErrorInfo
     */
    public static native ErrorInfo sendDirectBufferDefaultTimeout(long producerPtr, ByteBuffer buffers);

    /**
     * sendHeapBuffer
     *
     * @param producerPtr producerPtr
     * @param bytes bytes
     * @param len len
     * @param timeoutMs timeoutMs
     * @return ErrorInfo
     */
    public static native ErrorInfo sendHeapBuffer(long producerPtr, byte[] bytes, long len, int timeoutMs);

    /**
     * sendDirectBuffer
     *
     * @param producerPtr producerPtr
     * @param buffers buffers
     * @param timeoutMs timeoutMs
     * @return ErrorInfo
     */
    public static native ErrorInfo sendDirectBuffer(long producerPtr, ByteBuffer buffers, int timeoutMs);

    /**
     * close
     *
     * @param producerPtr producerPtr
     * @return ErrorInfo
     */
    public static native ErrorInfo close(long producerPtr);

    /**
     * freeJNIPtrNative
     *
     * @param producerPtr producerPtr
     */
    public static native void freeJNIPtrNative(long producerPtr);
}
