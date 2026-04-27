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

package org.yuanrong.stream;

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.exception.YRException;
import org.yuanrong.exception.handler.traceback.StackTraceUtils;
import org.yuanrong.jni.JniProducer;

import java.nio.ByteBuffer;
import java.util.Objects;
import java.util.concurrent.locks.Lock;
import java.util.concurrent.locks.ReentrantReadWriteLock;

/**
 * The producer implement, visible in this package.
 *
 * @since 2024-09-04
 */
public class ProducerImpl implements Producer {
    // for producerPtr.
    private final ReentrantReadWriteLock rwLock = new ReentrantReadWriteLock();
    private final Lock rLock = rwLock.readLock();
    private final Lock wLock = rwLock.writeLock();

    // Point to jni producer object.
    private long producerPtr;

    /**
     * set the producerPtr
     *
     * @param producerPtr the producerPtr
     */
    public ProducerImpl(long producerPtr) {
        this.producerPtr = producerPtr;
    }

    @Override
    public void send(Element element) throws YRException {
        ByteBuffer buffer = element.getBuffer();
        rLock.lock();
        try {
            ensureOpen();
            ErrorInfo err = null;
            if (buffer.isDirect()) {
                err = JniProducer.sendDirectBufferDefaultTimeout(producerPtr, buffer);
            } else {
                err = JniProducer.sendHeapBufferDefaultTimeout(producerPtr, buffer.array(), buffer.limit());
            }
            StackTraceUtils.checkErrorAndThrowForInvokeException(err, err.getErrorMessage());
        } finally {
            rLock.unlock();
        }
    }

    @Override
    public void send(Element element, int timeoutMs) throws YRException {
        ByteBuffer buffer = element.getBuffer();
        rLock.lock();
        try {
            ensureOpen();
            ErrorInfo err;
            if (buffer.isDirect()) {
                err = JniProducer.sendDirectBuffer(producerPtr, buffer, timeoutMs);
            } else {
                err = JniProducer.sendHeapBuffer(producerPtr, buffer.array(), buffer.limit(), timeoutMs);
            }
            if (err != null) {
                StackTraceUtils.checkErrorAndThrowForInvokeException(err, err.getErrorMessage());
            }
        } finally {
            rLock.unlock();
        }
    }

    /**
     * Checks to make sure that producer has not been closed.
     *
     * @throws YRException the YRException
     */
    private void ensureOpen() throws YRException {
        if (producerPtr == 0) {
            throw new YRException(ErrorCode.ERR_INNER_SYSTEM_ERROR, ModuleCode.RUNTIME,
                this.getClass().getName() + ": Producer closed");
        }
    }

    @Override
    public void close() throws YRException {
        wLock.lock();
        try {
            if (producerPtr == 0) {
                throw new YRException(ErrorCode.ERR_INNER_SYSTEM_ERROR, ModuleCode.RUNTIME,
                    "Consumer has been closed");
            }
            ErrorInfo err = JniProducer.close(this.producerPtr);
            producerPtr = 0;
            StackTraceUtils.checkErrorAndThrowForInvokeException(err, err.getErrorMessage());
        } finally {
            wLock.unlock();
        }
    }

    @Override
    protected void finalize() {
        if (producerPtr != 0) {
            JniProducer.freeJNIPtrNative(producerPtr);
        }
    }

    @Override
    public String toString() {
        return super.toString();
    }

    @Override
    public int hashCode() {
        return Objects.hash(producerPtr);
    }

    @Override
    public boolean equals(Object other) {
        if (this == other) {
            return true;
        }
        if (other == null || getClass() != other.getClass()) {
            return false;
        }
        ProducerImpl otherProducerImpl = (ProducerImpl) other;
        return producerPtr == otherProducerImpl.producerPtr;
    }
}
