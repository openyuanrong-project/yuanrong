/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
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

import lombok.Data;
import lombok.experimental.Accessors;

import java.util.HashMap;
import java.util.Map;

/**
 * Configuration class for creating the producer
 *
 * @since 2024/09/04
 */
@Data
@Accessors(chain = true)
public class ProducerConfig {
    /**
     * After sending, the flush is triggered after the maximum delay time. <0: No automatic flush. 0: Flush immediately.
     * Otherwise, it indicates the delay time in milliseconds.
     */
    private long delayFlushTimeMs = 5L;

    /**
     * Represents the buffer page size corresponding to the producer, in bytes (B); when the page is full, flush is
     * triggered. The default is ``1MB``, and it must be greater than 0 and a multiple of 4K.
     */
    private long pageSizeByte = 1024 * 1024L;

    /**
     * Specifies the maximum amount of shared memory that a stream can use on a worker, in units of B (bytes).
     * The default is ``100MB``, with a range of [64KB, the size of the worker's shared memory].
     */
    private long maxStreamSize = 100 * 1024 * 1024L;

    /**
     * Specifies whether the stream has the automatic cleanup feature enabled. The default is ``false``, which means it
     * is disabled.
     */
    private boolean autoCleanup = false;

    /**
     * Specifies whether the stream has the content encryption feature enabled. The default is ``false``, which means it
     * is disabled.
     */
    private boolean encryptStream = false;

    /**
     * The data sent by the producer will be retained until the Nth consumer receives it. The default value is ``0``,
     * which means that if there are no consumers when the producer sends the data, the data will not be retained,
     * and the consumer might not receive the data after it is created. This parameter is only effective for the first
     * consumer created, and the current valid range is [0, 1], and it does not support multiple consumers.
     */
    private long retainForNumConsumers = 0L;

    /**
     * Represents the reserved memory size, in units of B (bytes). When creating a producer, it will attempt to reserve
     * reserveSize bytes of memory. If the reservation fails, an exception will be thrown during the creation of the
     * producer. reserveSize must be an integer multiple of pageSize, and its value range is [0, maxStreamSize]. If
     * reserveSize is 0, it will be set to pageSize. The default value is ``0``.
     */
    private long reserveSize = 0L;

    /**
     * Producer expansion configuration. Common configuration items are as follows:\n
     * "STREAM_MODE": Stream mode, can be ``MPMC``, ``MPSC``, or ``SPSC``, default is ``MPMC``. If it is not one of the
     * above modes, an exception will be thrown. ``MPMC`` stands for multiple producers and multiple consumers, ``MPSC``
     * stands for multiple producers and single consumer, and ``SPSC`` stands for single producer and single consumer.
     * If it is ``MPSC`` or ``SPSC`` mode, the data system internally enables the multi-stream shared Page function.
     */
    private Map<String, String> extendConfig = new HashMap<>();

    /**
     * The ProducerConfig class Builder.
     *
     * @return Builder object of ProducerConfig class.
     */
    public static Builder builder() {
        return new Builder();
    }

    /**
     * The ProducerConfig Builder Class.
     */
    public static class Builder {
        private ProducerConfig producerConfig;

        /**
         * The ProducerConfig Builder.
         */
        public Builder() {
            producerConfig = new ProducerConfig();
        }

        /**
         * Sets the delayFlushTimeMs of ProducerConfig class.
         *
         * @param delayFlushTimeMs the delayFlushTimeMs.
         * @return ProducerConfig Builder class object.
         */
        public Builder delayFlushTimeMs(long delayFlushTimeMs) {
            producerConfig.delayFlushTimeMs = delayFlushTimeMs;
            return this;
        }

        /**
         * Sets the pageSizeByte of ProducerConfig class.
         *
         * @param pageSizeByte the pageSizeByte.
         * @return ProducerConfig Builder class object.
         */
        public Builder pageSizeByte(long pageSizeByte) {
            producerConfig.pageSizeByte = pageSizeByte;
            return this;
        }

        /**
         * Sets the maxStreamSize of ProducerConfig class.
         *
         * @param maxStreamSize the maxStreamSize.
         * @return ProducerConfig Builder class object.
         */
        public Builder maxStreamSize(long maxStreamSize) {
            producerConfig.maxStreamSize = maxStreamSize;
            return this;
        }

        /**
         * Sets the autoCleanup of ProducerConfig class.
         *
         * @param autoCleanup the autoCleanup.
         * @return ProducerConfig Builder class object.
         */
        public Builder autoCleanup(boolean autoCleanup) {
            producerConfig.autoCleanup = autoCleanup;
            return this;
        }

        /**
         * Sets the encryptStream of ProducerConfig class.
         *
         * @param encryptStream the encryptStream.
         * @return ProducerConfig Builder class object.
         */
        public Builder encryptStream(boolean encryptStream) {
            producerConfig.encryptStream = encryptStream;
            return this;
        }

        /**
         * Set the retainForNumConsumers.
         *
         * @param retainForNumConsumers retainForNumConsumers.
         * @return ProducerConfig Builder class object.
         */
        public Builder retainForNumConsumers(long retainForNumConsumers) {
            producerConfig.retainForNumConsumers = retainForNumConsumers;
            return this;
        }

        /**
         * Sets the reserveSize of ProducerConfig class.
         *
         * @param reserveSize the reserveSize.
         * @return ProducerConfig Builder class object.
         */
        public Builder reserveSize(long reserveSize) {
            producerConfig.reserveSize = reserveSize;
            return this;
        }

        /**
         * Add the extendConfig of ProducerConfig class.
         *
         * @param key the extendConfig key.
         * @param val the extendConfig value.
         * @return ProducerConfig Builder class object.
         */
        public Builder addExtendConfig(String key, String val) {
            producerConfig.extendConfig.put(key, val);
            return this;
        }

        /**
         * Sets the extendConfig of ProducerConfig class.
         *
         * @param extendConfig the extendConfig map.
         * @return ProducerConfig Builder class object.
         */
        public Builder extendConfig(Map<String, String> extendConfig) {
            producerConfig.extendConfig.putAll(extendConfig);
            return this;
        }

        /**
         * Builds the ProducerConfig object.
         *
         * @return ProducerConfig class object.
         */
        public ProducerConfig build() {
            return producerConfig;
        }
    }
}
