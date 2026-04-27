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

package org.yuanrong.call;

import org.yuanrong.InvokeOptions;
import org.yuanrong.api.YR;
import org.yuanrong.exception.YRException;
import org.yuanrong.function.CppInstanceMethod;
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
 * A helper class to serializer CppInstanceHandler.
 *
 * @since 2022/08/26
 */
class CppInstanceHandlerHelper {
    private static final Logger LOGGER = LoggerFactory.getLogger(CppInstanceHandlerHelper.class);

    /**
     * The serializer of class CppInstanceHandler.
     *
     * @since 2023/01/16
     */
    public static class CppInstanceHandlerSerializer extends JsonSerializer<CppInstanceHandler> {
        @Override
        public void serialize(CppInstanceHandler cppInstanceHandler, JsonGenerator jsonGenerator,
            SerializerProvider serializerProvider) {
            try {
                Map<String, String> handlerMap = cppInstanceHandler.exportHandler();
                jsonGenerator.writeObject(handlerMap);
            } catch (IOException e) {
                LOGGER.error("Error while serialize object: ", e);
            } catch (YRException e) {
                LOGGER.error("Error while expoert handler: ", e);
            }
        }
    }

    /**
     * The deserializer of class CppInstanceHandler.
     *
     * @since 2023/01/16
     */
    public static class CppInstanceHandlerDeserializer extends JsonDeserializer<CppInstanceHandler> {
        @Override
        public CppInstanceHandler deserialize(JsonParser jsonParser, DeserializationContext deserializationContext) {
            try {
                JsonToken currentToken = jsonParser.nextToken();
                Map<String, String> input = new HashMap<>();
                for (; currentToken == JsonToken.FIELD_NAME; currentToken = jsonParser.nextToken()) {
                    String fieldName = jsonParser.getCurrentName();
                    currentToken = jsonParser.nextToken();
                    switch (fieldName) {
                        case Constants.INSTANCE_KEY:
                        case Constants.INSTANCE_ID:
                        case Constants.FUNCTION_ID:
                        case Constants.CLASS_NAME:
                            String value = jsonParser.getValueAsString();
                            LOGGER.debug("set {} = {}", fieldName, value);
                            input.put(fieldName, value);
                            break;
                        default:
                            LOGGER.debug("get fieldName = {}", fieldName);
                    }
                }
                CppInstanceHandler handler = new CppInstanceHandler();
                handler.importHandler(input);
                return handler;
            } catch (IOException e) {
                LOGGER.error("Error while deserializing object", e);
            } catch (YRException e) {
                LOGGER.error("Error while expoert handler: ", e);
            }
            return new CppInstanceHandler();
        }
    }
}

/**
 * Create an operation class for creating a cpp stateful function instance.
 *
 * @note The CppInstanceHandler class is the handle returned after a Java function creates a cpp class instance;
 *       it is the return type of the `CppInstanceCreator.invoke` interface after creating a cpp class instance.\n
 *       Users can use the function method of CppInstanceHandler to create a cpp class instance member method handle
 *       and return a handle class CppInstanceFunctionHandler.
 *
 * @since 2022/08/26
 */
@JsonSerialize(using = CppInstanceHandlerHelper.CppInstanceHandlerSerializer.class)
@JsonDeserialize(using = CppInstanceHandlerHelper.CppInstanceHandlerDeserializer.class)
@Getter
public class CppInstanceHandler {
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
     * indicates whether need order.
     */
    private boolean needOrder = true;

    /**
     * The constructor of CppInstanceHandler.
     *
     * @param instanceId cpp function instance ID.
     * @param functionId cpp function deployment returns ID.
     * @param className  cpp function class name.
     */
    public CppInstanceHandler(String instanceId, String functionId, String className) {
        this.instanceId = instanceId;
        this.functionId = functionId;
        this.className = className;
    }

    /**
     * Default constructor of CppInstanceHandler.
     */
    public CppInstanceHandler() {}

    /**
     * The member method of the CppInstanceHandler class is used to return the handle of the member function of the
     * cloud C++ class instance.
     *
     * @param <R> the type parameter.
     * @param cppInstanceMethod CppInstanceMethod class instance.
     * @return CppInstanceFunctionHandler Instance.
     *
     * @snippet{trimleft} CppFunctionExample.java CppInstanceHandler function 样例代码
     */
    public <R> CppInstanceFunctionHandler<R> function(CppInstanceMethod<R> cppInstanceMethod) {
        return new CppInstanceFunctionHandler<>(this.instanceId, this.functionId, className, cppInstanceMethod);
    }

    /**
     * The member method of the CppInstanceHandler class is used to recycle the cloud-based cpp function instance.
     *
     * @note The default timeout for the current kill request is 30 seconds. In scenarios such as high disk load and
     *       etcd failure, the kill request processing time may exceed 30 seconds, causing the interface to throw
     *       a timeout exception. Since the kill request has a retry mechanism, users can choose not to handle or retry
     *       after capturing the timeout exception.
     *
     * @throws YRException If the recycling instance fails, an `YRException` exception is thrown.
     *                            For specific errors, see the error message description.
     *
     * @snippet{trimleft} CppFunctionExample.java CppInstanceHandler terminate 样例代码
     */
    public void terminate() throws YRException {
        YR.getRuntime().terminateInstance(instanceId);
    }

    /**
     * The member method of the CppInstanceHandler class is used to recycle cloud-based cpp function instances.
     * It supports synchronous or asynchronous termination.
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
     * @throws YRException If the recycling instance fails, an `YRException` exception is thrown.
     *                            For specific errors, see the error message description.
     *
     * @snippet{trimleft} CppFunctionExample.java CppInstanceHandler terminate sync 样例代码
     */
    public void terminate(boolean isSync) throws YRException {
        Runtime runtime = YR.getRuntime();
        if (isSync) {
            runtime.terminateInstanceSync(instanceId);
        } else {
            runtime.terminateInstance(instanceId);
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
     * Clear handler info.
     */
    public void clearHandlerInfo() {
        instanceId = "";
        className = "";
    }

    /**
     * Release CppInstanceHandler, decrease reference.
     *
     * @throws YRException Unified exception types thrown.
     */
    public void release() throws YRException {
        YR.getRuntime().decreaseReference(Collections.singletonList(instanceId));
    }

    /**
     * Obtain instance handle information. CppInstanceHandler class member method.
     * Users can obtain handle information through this method, which can be serialized and stored in a database or
     * other persistent tools. When the tenant context is enabled, handle information can also be obtained through this
     * method and used across tenants.
     *
     * @note If you want to enable tenant context, you can refer to the use case 2 of setContext.
     *
     * @return The handle information of CppInstanceHandler is stored in the kv key-value pair.
     * @throws YRException The export information failed and threw an YRException exception. For specific
     *                            errors, see the error message description.
     *
     * @snippet{trimleft} CppFunctionExample.java cpp exportHandler 样例代码
     */
    public Map<String, String> exportHandler() throws YRException {
        Map<String, String> output = new HashMap<>();
        output.put(Constants.INSTANCE_KEY, this.instanceId);
        output.put(Constants.CLASS_NAME, this.className);
        output.put(Constants.FUNCTION_ID, this.functionId);
        output.put(Constants.NEED_ORDER, this.needOrder ? "true" : "false");
        Runtime runtime = YR.getRuntime();
        if (this.realInstanceId.isEmpty()) {
            output.put(Constants.INSTANCE_ID, runtime.getRealInstanceId(this.instanceId));
        } else {
            output.put(Constants.INSTANCE_ID, this.realInstanceId);
        }
        output.put(Constants.INSTANCE_ROUTE, runtime.getInstanceRoute(this.instanceId));
        return output;
    }

    /**
     * The member method of the CppInstanceHandler class allows users to import handle information,
     * which can be obtained from a database or other persistent tools and deserialized.
     *
     * @param input The handle information of CppInstanceHandler is stored in the kv key-value pair.
     * @throws YRException The import information failed and threw an YRException exception. For specific
     *                            errors, see the error message description.
     *
     * @snippet{trimleft} CppFunctionExample.java CppInstanceHandler importHandler 样例代码
     */
    public void importHandler(Map<String, String> input) throws YRException {
        this.instanceId = SdkUtils.defaultIfNotFound(input, Constants.INSTANCE_KEY, "");
        this.className = SdkUtils.defaultIfNotFound(input, Constants.CLASS_NAME, "");
        this.functionId = SdkUtils.defaultIfNotFound(input, Constants.FUNCTION_ID, "");
        this.needOrder = !"false".equals(SdkUtils.defaultIfNotFound(input, Constants.NEED_ORDER, "false"));
        this.realInstanceId = SdkUtils.defaultIfNotFound(input, Constants.INSTANCE_ID, "");
        String instanceRoute = SdkUtils.defaultIfNotFound(input, Constants.INSTANCE_ROUTE, "");
        InvokeOptions opts = new InvokeOptions();
        opts.setNeedOrder(needOrder);
        Runtime runtime = YR.getRuntime();
        runtime.saveRealInstanceId(this.instanceId, this.realInstanceId, opts);
        if (instanceRoute != null && !instanceRoute.isEmpty()) {
            runtime.saveInstanceRoute(this.instanceId, instanceRoute);
        }
    }
}