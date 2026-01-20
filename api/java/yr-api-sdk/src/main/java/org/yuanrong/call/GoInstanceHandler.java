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

package org.yuanrong.call;

import org.yuanrong.api.YR;
import org.yuanrong.exception.YRException;
import org.yuanrong.function.GoInstanceMethod;
import org.yuanrong.runtime.Runtime;
import org.yuanrong.runtime.util.Constants;

import com.fasterxml.jackson.core.JsonGenerator;
import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.core.JsonToken;
import com.fasterxml.jackson.databind.DeserializationContext;
import com.fasterxml.jackson.databind.JsonDeserializer;
import com.fasterxml.jackson.databind.JsonSerializer;
import com.fasterxml.jackson.databind.SerializerProvider;
import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import com.fasterxml.jackson.databind.annotation.JsonSerialize;

import lombok.Getter;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.util.HashMap;

/**
 * A helper class to serializer GoInstanceHandler.
 *
 * @since 2024.03.12
 */
class GoInstanceHandlerHelper {
    private static final Logger LOGGER = LoggerFactory.getLogger(GoInstanceHandlerHelper.class);

    /**
     * The serializer of class GoInstanceHandler.
     *
     * @since 2024.03.12
     */
    public static class GoInstanceHandlerSerializer extends JsonSerializer<GoInstanceHandler> {
        @Override
        public void serialize(GoInstanceHandler goInstanceHandler, JsonGenerator jsonGenerator,
            SerializerProvider serializerProvider) {
            try {
                HashMap<String, String> handlerMap = new HashMap<>();
                handlerMap.put(Constants.INSTANCE_KEY, "");
                handlerMap.put(Constants.INSTANCE_ID, goInstanceHandler.getInstanceId());
                handlerMap.put(Constants.FUNCTION_KEY, "");
                handlerMap.put(Constants.CLASS_NAME, goInstanceHandler.getClassName());
                handlerMap.put(Constants.MODULE_NAME, "");
                handlerMap.put(Constants.LANGUAGE_KEY, "");
                jsonGenerator.writeObject(handlerMap);
            } catch (IOException e) {
                LOGGER.error("Error while serialize object", e);
            }
        }
    }

    /**
     * The deserializer of class GoInstanceHandler .
     *
     * @since 2024.03.12
     */
    public static class GoInstanceHandlerDeserializer extends JsonDeserializer<GoInstanceHandler> {
        @Override
        public GoInstanceHandler deserialize(JsonParser jsonParser, DeserializationContext deserializationContext) {
            try {
                JsonToken currentToken = jsonParser.nextToken();
                String instanceID = "";
                String className = "";
                for (; currentToken == JsonToken.FIELD_NAME; currentToken = jsonParser.nextToken()) {
                    String fieldName = jsonParser.getCurrentName();
                    currentToken = jsonParser.nextToken();
                    if (Constants.INSTANCE_ID.equals(fieldName)) {
                        instanceID = jsonParser.getValueAsString();
                        LOGGER.debug("set instanceID = {}", instanceID);
                    } else if (Constants.CLASS_NAME.equals(fieldName)) {
                        className = jsonParser.getValueAsString();
                        LOGGER.debug("set className = {}", className);
                    } else {
                        LOGGER.debug("get fieldName = {}", fieldName);
                    }
                }
                return new GoInstanceHandler(instanceID, className);
            } catch (IOException e) {
                LOGGER.error("Error while deserializing object", e);
            }
            return new GoInstanceHandler();
        }
    }
}

/**
 * Create an operation class for creating a Go stateful function instance.
 *
 * @note The GoInstanceHandler class is the handle returned after a Java function creates a Go class instance;
 *       it is the return type of the Go class instance created by the `GoInstanceCreator.invoke` interface.\n
 *       Users can use the function method of GoInstanceHandler to create a Go class instance member method handle and
 *       return the handle class GoInstanceFunctionHandler.
 *
 * @since 2024/03/12
 */
@JsonSerialize(using = GoInstanceHandlerHelper.GoInstanceHandlerSerializer.class)
@JsonDeserialize(using = GoInstanceHandlerHelper.GoInstanceHandlerDeserializer.class)
@Getter
public class GoInstanceHandler {
    private String instanceId;

    private String className = "";

    /**
     * The constructor of GoInstanceHandler.
     *
     * @param instanceId Go function instance id.
     * @param className Go function class name.
     */
    public GoInstanceHandler(String instanceId, String className) {
        this.instanceId = instanceId;
        this.className = className;
    }

    /**
     * Default constructor of GoInstanceHandler.
     */
    public GoInstanceHandler() {}

    /**
     * The member method of the GoInstanceHandler class is used to return the member function handle of the cloud Go
     * class instance.
     *
     * @param <R> the type of the object.
     * @param goInstanceMethod GoInstanceMethod class instance.
     * @return GoInstanceFunctionHandler Instance.
     *
     * @snippet{trimleft} GoInstanceExample.java GoInstanceHandler function example
     */
    public <R> GoInstanceFunctionHandler<R> function(GoInstanceMethod<R> goInstanceMethod) {
        return new GoInstanceFunctionHandler<>(this.instanceId, className, goInstanceMethod);
    }

    /**
     * The member method of the GoInstanceHandler class is used to recycle cloud Go function instances.
     *
     * @note The default timeout for the current kill request is 30 seconds. In scenarios such as high disk load and
     *       etcd failure, the kill request processing time may exceed 30 seconds, causing the interface to throw a
     *       timeout exception. Since the kill request has a retry mechanism, users can choose not to handle or retry
     *       after capturing the timeout exception.
     *
     * @throws YRException Unified exception types thrown.
     *
     * @snippet{trimleft} GoInstanceExample.java GoInstanceHandler terminate example
     */
    public void terminate() throws YRException {
        YR.getRuntime().terminateInstance(instanceId);
    }

    /**
     * The member method of the GoInstanceHandler class is used to recycle cloud Go function instances. It supports
     * synchronous or asynchronous termination.
     *
     * @note When synchronous termination is not enabled, the default timeout for the current kill request is
     *       30 seconds. In scenarios such as high disk load or etcd failure, the kill request processing time may
     *       exceed 30 seconds, causing the interface to throw a timeout exception. Since the kill request has a retry
     *       mechanism, users can choose not to handle or retry after capturing the timeout exception. When synchronous
     *       termination is enabled, this interface will block until the instance is completely exited.
     *
     * @param isSync Whether to enable synchronization. If true, it indicates sending a kill request with the signal
     *               quantity of killInstanceSync to the function-proxy, and the kernel synchronously kills the
     *               instance; if false, it indicates sending a kill request with the signal quantity of killInstance to
     *               the function-proxy, and the kernel asynchronously kills the instance.
     * @throws YRException Unified exception types thrown.
     *
     * @snippet{trimleft} GoInstanceExample.java GoInstanceHandler terminate sync example
     */
    public void terminate(boolean isSync) throws YRException {
        Runtime runtime = YR.getRuntime();
        if (isSync) {
            runtime.terminateInstanceSync(instanceId);
            return;
        }
        runtime.terminateInstance(instanceId);
    }

    /**
     * The member method of the GoInstanceHandler class is used to clear the information in the handle.
     */
    public void clearHandlerInfo() {
        instanceId = "";
        className = "";
    }
}