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

package org.yuanrong.serialization.strategy;

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.exception.YRException;
import org.yuanrong.runtime.client.ObjectRef;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.nio.ByteBuffer;
import java.util.ArrayList;
import java.util.List;
import java.util.Objects;
import java.util.stream.Collectors;

/**
 * Strategies for serialization and deserialization of objects.
 * It provides two serialization strategies: GeneralSerializationStrategy and
 * ByteBufferSerializationStrategy.
 *
 * @since 2023 /12/01
 */
public class Strategy {
    private static final Logger LOGGER = LoggerFactory.getLogger(Strategy.class);

    /**
     * This is a constant field that represents the GeneralSerializationStrategy.
     */
    public static final GeneralSerializationStrategy GENERAL_SERIALIZATION_STRATEGY =
        new GeneralSerializationStrategy();

    /**
     * This is a constant field that represents the ByteBufferSerializationStrategy.
     */
    public static final ByteBufferSerializationStrategy BUFFER_SERIALIZATION_STRATEGY =
        new ByteBufferSerializationStrategy();

    /**
     * Selects the appropriate deserialization strategy based on the given
     * object. If the object that the objectRef representing is a ByteBuffer, the method
     * returns the ByteBufferSerializationStrategy.
     * Otherwise, it returns the GeneralSerializationStrategy.
     *
     * @param objectRef the objectRef to select the serialization strategy for
     * @return the appropriate serialization strategy for the given objectRef
     */
    public static SerializationStrategy selectDeserializationStrategy(ObjectRef objectRef) {
        if (objectRef.isByteBuffer()) {
            return BUFFER_SERIALIZATION_STRATEGY;
        }
        return GENERAL_SERIALIZATION_STRATEGY;
    }

    /**
     * This method selects the appropriate serialization strategy based on the given
     * object. If the object is a ByteBuffer, the method returns the
     * ByteBufferSerializationStrategy.
     * Otherwise, it returns the GeneralSerializationStrategy.
     *
     * @param object the object to select the serialization strategy for
     * @return the appropriate serialization strategy for the given object
     */
    public static SerializationStrategy selectSerializationStrategy(Object object) {
        if (object instanceof ByteBuffer) {
            return BUFFER_SERIALIZATION_STRATEGY;
        }
        return GENERAL_SERIALIZATION_STRATEGY;
    }

    /**
     * This method retrieves a list of objects from the given byte array list and
     * object reference list.
     * It uses the appropriate deserialization strategy to deserialize each byte
     * array into an object.
     *
     * @param getRes the byte array list to retrieve objects from
     * @param refs   the object reference list to use for deserialization
     * @return a list of deserialized objects
     * @throws YRException if there is an error retrieving or deserializing
     *                            the objects
     */
    public static List<Object> getObjects(List<byte[]> getRes, List<ObjectRef> refs) throws YRException {
        LOGGER.debug("getting objects {}", refs.size());
        List<String> objectIdList = refs.stream().map(ObjectRef::getObjectID).collect(Collectors.toList());
        List<Object> resultList = new ArrayList<>();
        try {
            for (int objectIndex = 0; objectIndex < getRes.size(); objectIndex++) {
                if (getRes.get(objectIndex) == null) {
                    LOGGER.warn("unable to get object from datasystem, id: {}", objectIdList.get(objectIndex));
                    resultList.add(null);
                    continue;
                }
                if (getRes.get(objectIndex).length == 0) {
                    throw new YRException(ErrorCode.ERR_GET_OPERATION_FAILED, ModuleCode.RUNTIME,
                            "get empty return value from get interface");
                }
                ByteBuffer byteBuffer = ByteBuffer.wrap(getRes.get(objectIndex));
                ObjectRef objectRef = refs.get(objectIndex);

                Class<?> clz = objectRef.getType();
                if (Objects.isNull(clz)) {
                    // set deserialization strategy for current object, List<ObjectRef> holds same
                    // kind of objects, so set deserialization strategy according to the first
                    // objectRef.
                    clz = Object.class;
                }

                Object resObject = Strategy.selectDeserializationStrategy(objectRef).deserialize(byteBuffer, clz);
                resultList.add(resObject);
            }
        } catch (IOException exp) {
            throw new YRException(ErrorCode.ERR_GET_OPERATION_FAILED, ModuleCode.RUNTIME, exp);
        } catch (Exception e) {
            LOGGER.error("Error while serialize object", e);
        }
        return resultList;
    }
}
