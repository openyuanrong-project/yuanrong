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

package org.yuanrong.serialization;

import org.yuanrong.runtime.client.ObjectRef;

import com.fasterxml.jackson.annotation.JsonAutoDetect;
import com.fasterxml.jackson.annotation.PropertyAccessor;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.module.SimpleModule;

import org.msgpack.jackson.dataformat.JsonArrayFormat;
import org.msgpack.jackson.dataformat.MessagePackFactory;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

/**
 * The type Serializer.
 *
 * @since 2023 /10/16
 */
public class Serializer {
    /**
     * The Contained object ids.
     */
    public static final ThreadLocal<Set<String>> CONTAINED_OBJECT_IDS = ThreadLocal.withInitial(HashSet::new);

    private static final Logger LOGGER = LoggerFactory.getLogger(Serializer.class);

    /**
     * The Module.
     */
    private static SimpleModule module = new SimpleModule();

    private static ObjectMapper mapper = new ObjectMapper(new MessagePackFactory());

    static {
        module.addSerializer(ObjectRef.class, new ObjectRefSerializer());
        module.addDeserializer(ObjectRef.class, new ObjectRefDeserializer());
    }

    static {
        mapper.registerModule(module);
        mapper.setAnnotationIntrospector(new JsonArrayFormat());
        mapper.setVisibility(PropertyAccessor.FIELD, JsonAutoDetect.Visibility.ANY);
    }

    /**
     * Serializes the given object to a byte array using the Jackson
     * ObjectMapper.
     *
     * @param obj The object to be serialized
     * @return A byte array representing the serialized object
     * @throws JsonProcessingException If an error occurs during serialization
     */
    public static byte[] serialize(Object obj) throws JsonProcessingException {
        return mapper.writeValueAsBytes(obj);
    }

    /**
     * Deserialize object.
     *
     * @param bs   the bs
     * @param type the type
     * @return the object
     * @throws IOException the IOException
     */
    public static Object deserialize(byte[] bs, Class<?> type) throws IOException {
        if (List.class.equals(type)) {
            return mapper.readValue(bs, new TypeReference<List<ObjectRef>>() {});
        } else {
            return mapper.readValue(bs, type);
        }
    }
}
