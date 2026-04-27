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

import org.yuanrong.InvokeOptions;
import org.yuanrong.api.YR;
import org.yuanrong.exception.YRException;
import org.yuanrong.function.JavaInstanceMethod;
import org.yuanrong.runtime.Runtime;
import org.yuanrong.runtime.util.Constants;
import org.yuanrong.utils.SdkUtils;

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
import java.util.Collections;
import java.util.HashMap;
import java.util.Map;

/**
 * A helper class to serializer JavaInstanceHandler.
 *
 * @since 2024.04.16
 */
class JavaInstanceHandlerHelper {
    private static final Logger LOGGER = LoggerFactory.getLogger(JavaInstanceHandlerHelper.class);

    /**
     * The serializer of class JavaInstanceHandler.
     *
     * @since 2024.04.16
     */
    public static class JavaInstanceHandlerSerializer extends JsonSerializer<JavaInstanceHandler> {
        @Override
        public void serialize(JavaInstanceHandler javaInstanceHandler, JsonGenerator jsonGenerator,
            SerializerProvider serializerProvider) {
            try {
                Map<String, String> handlerMap = javaInstanceHandler.exportHandler();
                jsonGenerator.writeObject(handlerMap);
            } catch (IOException e) {
                LOGGER.error("Error while serialize object: ", e);
            } catch (YRException e) {
                LOGGER.error("Error while expoert handler: ", e);
            }
        }
    }

    /**
     * The deserializer of class JavaInstanceHandler.
     *
     * @since 2024.04.16
     */
    public static class JavaInstanceHandlerDeserializer extends JsonDeserializer<JavaInstanceHandler> {
        @Override
        public JavaInstanceHandler deserialize(JsonParser jsonParser, DeserializationContext deserializationContext) {
            try {
                JsonToken currentToken = jsonParser.nextToken();
                Map<String, String> input = new HashMap<>();
                for (; currentToken == JsonToken.FIELD_NAME; currentToken = jsonParser.nextToken()) {
                    String fieldName = jsonParser.getCurrentName();
                    currentToken = jsonParser.nextToken();
                    switch (fieldName) {
                        case Constants.CLASS_NAME:
                        case Constants.FUNCTION_ID:
                        case Constants.INSTANCE_ID:
                        case Constants.INSTANCE_KEY:
                            String value = jsonParser.getValueAsString();
                            LOGGER.debug("set {} = {}", fieldName, value);
                            input.put(fieldName, value);
                            break;
                        default:
                            LOGGER.debug("get fieldName = {}", fieldName);
                    }
                }
                JavaInstanceHandler handler = new JavaInstanceHandler();
                handler.importHandler(input);
                return handler;
            } catch (IOException e) {
                LOGGER.error("Error while deserializing object: ", e);
            } catch (YRException e) {
                LOGGER.error("Error while expoert handler: ", e);
            }
            return new JavaInstanceHandler();
        }
    }
}

/**
 * Create an operation class for creating a Java stateful function instance.
 *
 * @note The JavaInstanceHandler class is the handle returned after creating a Java class instance;
 *       it is the return type of the `JavaInstanceCreator.invoke` interface after creating a Java class instance.\n
 *       Users can use the function method of JavaInstanceHandler to create a Java class instance member method handle
 *       and return the handle class JavaInstanceFunctionHandler.
 *
 * @since 2024/04/16
 */
@JsonSerialize(using = JavaInstanceHandlerHelper.JavaInstanceHandlerSerializer.class)
@JsonDeserialize(using = JavaInstanceHandlerHelper.JavaInstanceHandlerDeserializer.class)
@Getter
public class JavaInstanceHandler {
    /**
     * instanceId to invoke function.
     */
    private String instanceId;

    /**
     * real instanceId to invoke function.
     */
    private String realInstanceId = "";

    /**
     * functionId to invoke function.
     */
    private String functionId = "";

    /**
     * the name of class.
     */
    private String className = "";

    /**
     * indicates wheather need order.
     */
    private boolean needOrder = true;

    /**
     * The constructor of JavaInstanceHandler.
     *
     * @param instanceId Java function instance ID.
     * @param functionId Java function deployment returns an ID.
     * @param className Java function class name.
     */
    public JavaInstanceHandler(String instanceId, String functionId, String className) {
        this.instanceId = instanceId;
        this.functionId = functionId;
        this.className = className;
    }

    /**
     * Default constructor of JavaInstanceHandler.
     */
    public JavaInstanceHandler() {}

    /**
     * The member method of the JavaInstanceHandler class is used to return the member function handle of the Java class
     * instance on the cloud.
     *
     * @param <R> the type of the object.
     * @param javaInstanceMethod JavaInstanceMethod Class instance.
     * @return JavaInstanceFunctionHandler Instance.
     *
     * @snippet{trimleft} JavaInstanceExample.java JavaInstanceHandler function example
     */
    public <R> JavaInstanceFunctionHandler<R> function(JavaInstanceMethod<R> javaInstanceMethod) {
        return new JavaInstanceFunctionHandler<>(this.instanceId, this.functionId, className, javaInstanceMethod);
    }

    /**
     * The member method of the JavaInstanceHandler class is used to recycle cloud Java function instances. It supports
     * synchronous or asynchronous termination.
     *
     * @note When synchronous termination is not enabled, the default timeout for the current kill request is
     *       30 seconds. In scenarios such as high disk load or etcd failure, the kill request processing time may
     *       exceed 30 seconds, causing the interface to throw a timeout exception. Since the kill request has a retry
     *       mechanism, users can choose not to handle or retry after capturing the timeout exception. When synchronous
     *       termination is enabled, the interface will block until the instance is completely exited.
     *
     * @param isSync Whether to enable synchronization. If true, it indicates sending a kill request with the signal
     *               quantity of killInstanceSync to the function-proxy, and the kernel synchronously kills the
     *               instance; if false, it indicates sending a kill request with the signal quantity of killInstance
     *               to the function-proxy, and the kernel asynchronously kills the instance.
     * @throws YRException Unified exception types thrown.
     *
     * @snippet{trimleft} JavaInstanceExample.java JavaInstanceHandler terminate sync example
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
     * The member method of the JavaInstanceHandler class is used to recycle cloud Java function instances.
     *
     * @note The default timeout for the current kill request is 30 seconds. In scenarios such as high disk load and
     *       etcd failure, the kill request processing time may exceed 30 seconds, causing the interface to throw a
     *       timeout exception. Since the kill request has a retry mechanism, users can choose not to handle or retry
     *       after capturing the timeout exception.
     *
     * @throws YRException Unified exception types thrown.
     *
     * @snippet{trimleft} JavaInstanceExample.java JavaInstanceHandler terminate example
     */
    public void terminate() throws YRException {
        YR.getRuntime().terminateInstance(instanceId);
    }

    /**
     * Both the user and the runtime java hold JavaInstanceHandler. To ensure that JavaInstanceHandler cannot be used by
     * the user after calling `Finalize`, the member variables of JavaInstanceHandler should be cleared.
     */
    public void clearHandlerInfo() {
        instanceId = "";
        className = "";
    }

    /**
     * The member method of the JavaInstanceHandler class allows users to obtain handle information, which can be
     * serialized and stored in a database or other persistence tools. When the tenant context is enabled, handle
     * information can also be obtained through this method and used across tenants.
     *
     * @return Map<String, String> type, storing the handle information of JavaInstanceHandler through kv
     *         key-value pairs.
     * @throws YRException Unified exception types thrown.
     *
     * @snippet{trimleft} JavaInstanceExample.java JavaInstanceHandler exportHandler example
     */
    public Map<String, String> exportHandler() throws YRException {
        Map<String, String> out = new HashMap<>();
        out.put(Constants.NEED_ORDER, this.needOrder ? "true" : "false");
        out.put(Constants.FUNCTION_ID, this.functionId);
        out.put(Constants.CLASS_NAME, this.className);
        out.put(Constants.INSTANCE_KEY, this.instanceId);
        Runtime runtime = YR.getRuntime();
        if (this.realInstanceId.isEmpty()) {
            out.put(Constants.INSTANCE_ID, runtime.getRealInstanceId(this.instanceId));
        } else {
            out.put(Constants.INSTANCE_ID, this.realInstanceId);
        }
        out.put(Constants.INSTANCE_ROUTE, runtime.getInstanceRoute(this.instanceId));
        return out;
    }

    /**
     * The member method of the JavaInstanceHandler class allows users to import handle information, which can be
     * obtained and deserialized from persistent tools such as databases. When the tenant context is enabled, this
     * method can also be used to import handle information from other tenant contexts.
     *
     * @param input Store the handle information of JavaInstanceHandler through the kv key-value pair.
     * @throws YRException Unified exception types thrown.
     *
     * @snippet{trimleft} JavaInstanceExample.java JavaInstanceHandler importHandler example
     */
    public void importHandler(Map<String, String> input) throws YRException {
        this.needOrder = !"false".equals(SdkUtils.defaultIfNotFound(input, Constants.NEED_ORDER, "false"));
        this.functionId = SdkUtils.defaultIfNotFound(input, Constants.FUNCTION_ID, "");
        this.className = SdkUtils.defaultIfNotFound(input, Constants.CLASS_NAME, "");
        this.realInstanceId = SdkUtils.defaultIfNotFound(input, Constants.INSTANCE_ID, "");
        this.instanceId = SdkUtils.defaultIfNotFound(input, Constants.INSTANCE_KEY, "");
        String instanceRoute = SdkUtils.defaultIfNotFound(input, Constants.INSTANCE_ROUTE, "");
        InvokeOptions opts = new InvokeOptions();
        opts.setNeedOrder(needOrder);
        Runtime runtime = YR.getRuntime();
        runtime.saveRealInstanceId(this.instanceId, this.realInstanceId, opts);
        if (instanceRoute != null && !instanceRoute.isEmpty()) {
            runtime.saveInstanceRoute(this.instanceId, instanceRoute);
        }
    }

    /**
     * set need order.
     *
     * @param needOrder indicates whether need order.
     *
     */
    void setNeedOrder(boolean needOrder) {
        this.needOrder = needOrder;
    }

    /**
     * Release JavaInstanceHandler, decrease reference.
     *
     * @throws YRException Unified exception types thrown.
     */
    public void release() throws YRException {
        YR.getRuntime().decreaseReference(Collections.singletonList(instanceId));
    }
}
