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

import org.yuanrong.exception.FunctionRuntimeException;
import org.yuanrong.runtime.client.ObjectRef;

import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.core.JsonToken;
import com.fasterxml.jackson.databind.DeserializationContext;
import com.fasterxml.jackson.databind.JsonDeserializer;

import lombok.SneakyThrows;

import java.io.IOException;

/**
 * The type Object ref deserializer.
 *
 * @since 2023 /10/16
 */
public class ObjectRefDeserializer extends JsonDeserializer<ObjectRef> {
    private static final String DESERIALIZE_OBJREF_EXCEPTION =
            "deserialize objectRef occurs exception in function-runtime, currentToken is not String";

    /**
     * ObjectRef deserializer
     *
     * @param jsonParser jsonParser
     * @param deserializationContext context to deserialize
     *
     * @return the ObjectRef object
     */
    @SneakyThrows
    @Override
    public ObjectRef deserialize(JsonParser jsonParser, DeserializationContext deserializationContext) {
        String objId = "";
        try {
            JsonToken currentToken = jsonParser.nextToken();
            if (currentToken != null) {
                if (currentToken == JsonToken.VALUE_STRING) {
                    objId = jsonParser.getText();
                    jsonParser.nextToken();
                } else {
                    throw new FunctionRuntimeException(
                            DESERIALIZE_OBJREF_EXCEPTION
                                    + " current token class: "
                                    + currentToken.getClass().getName());
                }
            }
        } catch (IOException | FunctionRuntimeException e) {
            throw new FunctionRuntimeException(e);
        }
        Serializer.CONTAINED_OBJECT_IDS.get().add(objId);
        /**
         * when current objectRef is detected when deserializing, then it must already be a nestedObjectRef of
         * outside objectRef, there is no need to decrease this objectRef when YR.Finalize is called,
         * so the parameter isNeedDecrement is set to FALSE.
         */
        return new ObjectRef(objId);
    }
}
