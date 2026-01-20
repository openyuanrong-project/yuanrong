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

package org.yuanrong.serialization.strategy;

import org.yuanrong.exception.YRException;

import com.fasterxml.jackson.core.JsonProcessingException;

import java.nio.ByteBuffer;

/**
 * Serialization strategy interface.
 *
 * @since 2023 /10/16
 */
public interface SerializationStrategy<T, S, D> {
    /**
     * Serialize Generic T to serialized S
     *
     * @param value object to serialize
     * @return serialized ByteBuffer data
     * @throws JsonProcessingException json process exception
     */
    ByteBuffer serialize(T value) throws JsonProcessingException;

    /**
     * Deserialize data Generic D to Generic T
     *
     * @param serializedData data to deserialize
     * @param type           data type
     * @return object
     * @throws JsonProcessingException json process exception
     * @throws Exception      YR exception
     * @throws YRException the YRException
     */
    T deserialize(D serializedData, Class<T> type) throws JsonProcessingException, Exception, YRException;
}
