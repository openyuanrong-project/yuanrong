/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.
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

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.exception.YRException;
import org.yuanrong.serialization.Serializer;

import com.fasterxml.jackson.core.JsonProcessingException;

import java.io.IOException;
import java.nio.ByteBuffer;

/**
 * General class serialization strategy.
 *
 * @since 2023 /1/9
 */
public class GeneralSerializationStrategy implements SerializationStrategy<Object, ByteBuffer, ByteBuffer> {
    @Override
    public ByteBuffer serialize(Object value) throws JsonProcessingException {
        return ByteBuffer.wrap(Serializer.serialize(value));
    }

    @Override
    public Object deserialize(ByteBuffer serializedData, Class<Object> type)
            throws IOException, YRException {
        if (serializedData == null) {
            throw new YRException(ErrorCode.ERR_DESERIALIZATION_FAILED, ModuleCode.RUNTIME,
                    "The input parameter 'serializedData' is null.");
        }
        byte[] data;
        if (serializedData.hasArray()) {
            data = serializedData.array();
        } else {
            data = new byte[serializedData.remaining()];
            serializedData.get(data);
        }
        return Serializer.deserialize(data, type);
    }
}
