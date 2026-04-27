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

package org.yuanrong.serialization;

import org.yuanrong.runtime.client.ObjectRef;

import com.fasterxml.jackson.core.JsonGenerator;
import com.fasterxml.jackson.databind.JsonSerializer;
import com.fasterxml.jackson.databind.SerializerProvider;

import java.io.IOException;
import java.util.Collections;

/**
 * The type Object ref serializer.
 *
 * @since 2023 /10/16
 */
public class ObjectRefSerializer extends JsonSerializer<ObjectRef> {
    /**
     * Serialize for ObjectRef objects.
     *
     * @param objectRef the object to serialize.
     * @param jsonGenerator json generator
     * @param serializerProvider serializer provider
     * @throws IOException the io exception
     */
    @Override
    public void serialize(ObjectRef objectRef, JsonGenerator jsonGenerator, SerializerProvider serializerProvider)
            throws IOException {
        Serializer.CONTAINED_OBJECT_IDS.get().add(objectRef.getObjectID());
        jsonGenerator.writeObject(Collections.singletonList(objectRef.getObjectID()));
    }
}
