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

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.errorcode.Pair;
import org.yuanrong.exception.YRException;
import org.yuanrong.exception.handler.traceback.StackTraceUtils;
import org.yuanrong.jni.JniConsumer;

import java.util.List;
import java.util.Objects;
import java.util.concurrent.locks.Lock;
import java.util.concurrent.locks.ReentrantReadWriteLock;

/**
 * The consumer implement, visible in this package.
 *
 * @since 2024-09-04
 */
public class ConsumerImpl implements Consumer {
    // for consumerPtr.
    private final ReentrantReadWriteLock rwLock = new ReentrantReadWriteLock();

    private final Lock rLock = rwLock.readLock();

    private final Lock wLock = rwLock.writeLock();

    // Point to jni consumer object.
    private long consumerPtr;

    /**
     * set the consumerPtr
     *
     * @param consumerPtr the consumerPtr
     */
    public ConsumerImpl(long consumerPtr) {
        this.consumerPtr = consumerPtr;
    }

    @Override
    public List<Element> receive(long expectNum, int timeoutMs) throws YRException {
        rLock.lock();
        try {
            ensureOpen();
            Pair<ErrorInfo, List<Element>> res = JniConsumer.receive(consumerPtr, expectNum, timeoutMs, true);
            StackTraceUtils.checkErrorAndThrowForInvokeException(res.getFirst(), res.getFirst().getErrorMessage());
            return res.getSecond();
        } finally {
            rLock.unlock();
        }
    }

    @Override
    public List<Element> receive(int timeoutMs) throws YRException {
        rLock.lock();
        try {
            ensureOpen();
            Pair<ErrorInfo, List<Element>> res = JniConsumer.receive(consumerPtr, 0L, timeoutMs, false);
            StackTraceUtils.checkErrorAndThrowForInvokeException(res.getFirst(), res.getFirst().getErrorMessage());
            return res.getSecond();
        } finally {
            rLock.unlock();
        }
    }

    @Override
    public void ack(long elementId) throws YRException {
        rLock.lock();
        try {
            ensureOpen();
            ErrorInfo err = JniConsumer.ack(consumerPtr, elementId);
            StackTraceUtils.checkErrorAndThrowForInvokeException(err, err.getErrorMessage());
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
        if (consumerPtr == 0) {
            throw new YRException(ErrorCode.ERR_INNER_SYSTEM_ERROR, ModuleCode.RUNTIME,
                this.getClass().getName() + ": Consumer closed");
        }
    }

    @Override
    public void close() throws YRException {
        wLock.lock();
        try {
            if (consumerPtr == 0) {
                throw new YRException(ErrorCode.ERR_INNER_SYSTEM_ERROR, ModuleCode.RUNTIME,
                    "Consumer has been closed");
            }
            ErrorInfo err = JniConsumer.close(consumerPtr);
            consumerPtr = 0;
            StackTraceUtils.checkErrorAndThrowForInvokeException(err, err.getErrorMessage());
        } finally {
            wLock.unlock();
        }
    }

    @Override
    public String toString() {
        return super.toString();
    }

    @Override
    protected void finalize() {
        if (consumerPtr != 0) {
            JniConsumer.freeJNIPtrNative(consumerPtr);
        }
    }

    @Override
    public boolean equals(Object other) {
        if (this == other) {
            return true;
        }
        if (other == null || getClass() != other.getClass()) {
            return false;
        }
        ConsumerImpl otherConsumerImpl = (ConsumerImpl) other;
        return consumerPtr == otherConsumerImpl.consumerPtr;
    }

    @Override
    public int hashCode() {
        return Objects.hash(consumerPtr);
    }
}
