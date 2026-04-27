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

package org.yuanrong.call;

import org.yuanrong.InvokeOptions;
import org.yuanrong.api.YR;
import org.yuanrong.exception.YRException;
import org.yuanrong.function.YRFunc0;
import org.yuanrong.function.YRFunc1;
import org.yuanrong.function.YRFunc2;
import org.yuanrong.function.YRFunc3;
import org.yuanrong.function.YRFunc4;
import org.yuanrong.function.YRFunc5;
import org.yuanrong.function.YRFuncVoid0;
import org.yuanrong.function.YRFuncVoid1;
import org.yuanrong.function.YRFuncVoid2;
import org.yuanrong.function.YRFuncVoid3;
import org.yuanrong.function.YRFuncVoid4;
import org.yuanrong.function.YRFuncVoid5;
import org.yuanrong.libruntime.generated.Libruntime.ApiType;
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
 * A helper class to serializer InstanceHandler.
 *
 * @since 2024/05/26
 */
class InstanceHandlerHelper {
    private static final Logger LOGGER = LoggerFactory.getLogger(InstanceHandlerHelper.class);

    /**
     * The serializer of class InstanceHandler.
     *
     * @since 2024 /05/25
     */
    public static class InstanceHandlerSerializer extends JsonSerializer<InstanceHandler> {
        @Override
        public void serialize(InstanceHandler instanceHandler, JsonGenerator jsonGenerator,
            SerializerProvider serializerProvider) {
            try {
                Map<String, String> handlerMap = instanceHandler.exportHandler();
                jsonGenerator.writeObject(handlerMap);
            } catch (IOException e) {
                LOGGER.error("Error while serialize object: ", e);
            } catch (YRException e) {
                LOGGER.error("Error while expoert handler: ", e);
            }
        }
    }

    /**
     * The deserializer of class InstanceHandler .
     *
     * @since 2024 /05/16
     */
    public static class InstanceHandlerDeserializer extends JsonDeserializer<InstanceHandler> {
        @Override
        public InstanceHandler deserialize(JsonParser jsonParser, DeserializationContext deserializationContext) {
            try {
                JsonToken currentToken = jsonParser.nextToken();
                Map<String, String> input = new HashMap<>();
                for (; currentToken == JsonToken.FIELD_NAME; currentToken = jsonParser.nextToken()) {
                    String fieldName = jsonParser.getCurrentName();
                    currentToken = jsonParser.nextToken();
                    if (Constants.INSTANCE_KEY.equals(fieldName)) {
                        String instanceID = jsonParser.getValueAsString();
                        LOGGER.debug("set instanceID = {}", instanceID);
                        input.put(Constants.INSTANCE_KEY, instanceID);
                    } else if (Constants.INSTANCE_ID.equals(fieldName)) {
                        String realInstanceID = jsonParser.getValueAsString();
                        LOGGER.debug("set real instanceID = {}", realInstanceID);
                        input.put(Constants.INSTANCE_ID, realInstanceID);
                    } else if (Constants.API_TYPE.equals(fieldName)) {
                        String apiType = jsonParser.getValueAsString();
                        LOGGER.debug("set apiType = {}", apiType);
                        input.put(Constants.API_TYPE, apiType);
                    } else {
                        LOGGER.debug("get fieldName = {}", fieldName);
                    }
                }
                InstanceHandler handler = new InstanceHandler();
                handler.importHandler(input);
                return handler;
            } catch (IOException e) {
                LOGGER.error("Error while deserializing object", e);
            } catch (YRException e) {
                LOGGER.error("Error while expoert handler: ", e);
            }
            return new InstanceHandler();
        }
    }
}

/**
 * Create an operation class for creating a Java stateful function instance.
 *
 * @note The class InstanceHandler is the handle returned after creating a Java class instance;
 *       it is the return type of the interface `InstanceCreator.invoke` after creating a Java class instance.\n
 *       Users can use the function method of the built-in InstanceHandler to create a Java class instance member method
 *       handle and return the handle class InstanceFunctionHandler.
 *
 * @since 2022/08/30
 */
@JsonSerialize(using = InstanceHandlerHelper.InstanceHandlerSerializer.class)
@JsonDeserialize(using = InstanceHandlerHelper.InstanceHandlerDeserializer.class)
@Getter
public class InstanceHandler {
    private String instanceId;

    private String realInstanceId = "";

    private ApiType apiType;

    private boolean needOrder = true;

    /**
     * The constructor of the InstanceHandler.
     *
     * @param instanceId Java function instance ID.
     * @param apiType The enumeration class has three values: Function, Faas, and Posix.
     *                It is used internally by Yuanrong to distinguish function types. The default is Actor.
     */
    public InstanceHandler(String instanceId, ApiType apiType) {
        this.instanceId = instanceId;
        this.apiType = apiType;
    }

    /**
     * Default constructor of InstanceHandler.
     */
    public InstanceHandler() {}

    /**
     * Java functions under the cloud call member functions of Java class instances on the cloud,
     * supporting user functions with 0 input parameters and 1 return value.
     *
     * @param <R> Return value type.
     * @param func YRFunc0 Class instance.
     * @return InstanceFunctionHandler Instance.
     */
    public <R> InstanceFunctionHandler<R> function(YRFunc0<R> func) {
        return new InstanceFunctionHandler<>(func, this.instanceId, this.apiType);
    }

    /**
     * Java functions under the cloud call member functions of Java class instances on the cloud,
     * supporting user functions with 1 input parameter and 1 return value.
     *
     * @param <T0> Input parameter type.
     * @param <R> Return value type.
     * @param func YRFunc1 Class instance.
     * @return InstanceFunctionHandler Instance.
     *
     * @snippet{trimleft} InstanceExample.java InstanceHandler function example
     */
    public <T0, R> InstanceFunctionHandler<R> function(YRFunc1<T0, R> func) {
        return new InstanceFunctionHandler<>(func, this.instanceId, this.apiType);
    }

    /**
     * Java functions in the cloud can call member functions of Java class instances in the cloud,
     * supporting user functions with 2 input parameters and 1 return value.
     *
     * @param <T0> Input parameter type.
     * @param <T1> Input parameter type.
     * @param <R> Return value type.
     * @param func YRFunc2 Class instance.
     * @return InstanceFunctionHandler Instance.
     */
    public <T0, T1, R> InstanceFunctionHandler<R> function(YRFunc2<T0, T1, R> func) {
        return new InstanceFunctionHandler<>(func, this.instanceId, this.apiType);
    }

    /**
     * Java functions under the cloud call member functions of Java class instances on the cloud,
     * supporting user functions with 3 parameters and 1 return value.
     *
     * @param <T0> Input parameter type.
     * @param <T1> Input parameter type.
     * @param <T2> Input parameter type.
     * @param <R> Return value type.
     * @param func YRFunc3 Class instance.
     * @return InstanceFunctionHandler Instance.
     */
    public <T0, T1, T2, R> InstanceFunctionHandler<R> function(YRFunc3<T0, T1, T2, R> func) {
        return new InstanceFunctionHandler<>(func, this.instanceId, this.apiType);
    }

    /**
     * Java functions under the cloud call member functions of Java class instances on the cloud,
     * supporting user functions with 4 parameters and 1 return value.
     *
     * @param <T0> Input parameter type.
     * @param <T1> Input parameter type.
     * @param <T2> Input parameter type.
     * @param <T3> Input parameter type.
     * @param <R> Return value type.
     * @param func YRFunc4 Class instance.
     * @return InstanceFunctionHandler Instance.
     */
    public <T0, T1, T2, T3, R> InstanceFunctionHandler<R> function(YRFunc4<T0, T1, T2, T3, R> func) {
        return new InstanceFunctionHandler<>(func, this.instanceId, this.apiType);
    }

    /**
     * Java functions under the cloud call member functions of Java class instances on the cloud,
     * supporting user functions with 5 parameters and 1 return value.
     *
     * @param <T0> Input parameter type.
     * @param <T1> Input parameter type.
     * @param <T2> Input parameter type.
     * @param <T3> Input parameter type.
     * @param <T4> Input parameter type.
     * @param <R> Return value type.
     * @param func YRFunc5 Class instance.
     * @return InstanceFunctionHandler Instance.
     */
    public <T0, T1, T2, T3, T4, R> InstanceFunctionHandler<R> function(YRFunc5<T0, T1, T2, T3, T4, R> func) {
        return new InstanceFunctionHandler<>(func, this.instanceId, this.apiType);
    }

    /**
     * Java functions under the cloud call member functions of Java class instances on the cloud,
     * supporting user functions with 0 input parameters and 0 return values.
     *
     * @param func YRFuncVoid0 Class instance.
     * @return VoidInstanceFunctionHandler instance.
     */
    public VoidInstanceFunctionHandler function(YRFuncVoid0 func) {
        return new VoidInstanceFunctionHandler(func, this.instanceId, this.apiType);
    }

    /**
     * Java functions under the cloud call member functions of Java class instances on the cloud,
     * supporting user functions with 1 input parameter and 0 return values.
     *
     * @param <T0> Input parameter type.
     * @param func YRFuncVoid1 Class instance.
     * @return VoidInstanceFunctionHandler Instance.
     */
    public <T0> VoidInstanceFunctionHandler function(YRFuncVoid1<T0> func) {
        return new VoidInstanceFunctionHandler(func, this.instanceId, this.apiType);
    }

    /**
     * Java functions under the cloud call member functions of Java class instances on the cloud,
     * supporting user functions with 2 input parameters and 0 return values.
     *
     * @param <T0> Input parameter type.
     * @param <T1> Input parameter type.
     * @param func YRFuncVoid2 Class instance.
     * @return VoidInstanceFunctionHandler Instance.
     */
    public <T0, T1> VoidInstanceFunctionHandler function(YRFuncVoid2<T0, T1> func) {
        return new VoidInstanceFunctionHandler(func, this.instanceId, this.apiType);
    }

    /**
     * Java functions under the cloud call member functions of Java class instances on the cloud,
     * supporting user functions with 3 parameters and 0 return values.
     *
     * @param <T0> Input parameter type.
     * @param <T1> Input parameter type.
     * @param <T2> Input parameter type.
     * @param func YRFuncVoid3 Class instance.
     * @return VoidInstanceFunctionHandler Instance.
     */
    public <T0, T1, T2> VoidInstanceFunctionHandler function(YRFuncVoid3<T0, T1, T2> func) {
        return new VoidInstanceFunctionHandler(func, this.instanceId, this.apiType);
    }

    /**
     * Java functions under the cloud call member functions of Java class instances on the cloud,
     * supporting user functions with 4 parameters and 0 return values.
     *
     * @param <T0> Input parameter type.
     * @param <T1> Input parameter type.
     * @param <T2> Input parameter type.
     * @param <T4> Input parameter type.
     * @param func YRFuncVoid4 Class instance.
     * @return VoidInstanceFunctionHandler Instance.
     */
    public <T0, T1, T2, T4> VoidInstanceFunctionHandler function(YRFuncVoid4<T0, T1, T2, T4> func) {
        return new VoidInstanceFunctionHandler(func, this.instanceId, this.apiType);
    }

    /**
     * Java functions under the cloud call member functions of Java class instances on the cloud,
     * supporting user functions with 5 parameters and 0 return values.
     *
     * @param <T0> Input parameter type.
     * @param <T1> Input parameter type.
     * @param <T2> Input parameter type.
     * @param <T4> Input parameter type.
     * @param <T5> Input parameter type.
     * @param func YRFuncVoid5 Class instance.
     * @return VoidInstanceFunctionHandler Instance.
     */
    public <T0, T1, T2, T4, T5> VoidInstanceFunctionHandler function(YRFuncVoid5<T0, T1, T2, T4, T5> func) {
        return new VoidInstanceFunctionHandler(func, this.instanceId, this.apiType);
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
     * The member method of the InstanceHandler class is used to recycle cloud Java function instances. It supports
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
     * @snippet{trimleft} InstanceExample.java InstanceHandler terminate sync example
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
     * The member method of the InstanceHandler class is used to recycle cloud Java function instances.
     *
     * @note The default timeout for the current kill request is 30 seconds. In scenarios such as high disk load and
     *       etcd failure, the kill request processing time may exceed 30 seconds, causing the interface to throw a
     *       timeout exception. Since the kill request has a retry mechanism, users can choose not to handle or retry
     *       after capturing the timeout exception.
     *
     * @throws YRException Unified exception types thrown.
     *
     * @snippet{trimleft} InstanceExample.java InstanceHandler terminate example
     */
    public void terminate() throws YRException {
        YR.getRuntime().terminateInstance(instanceId);
    }

    /**
     * Both the user and the runtime java hold the instanceHandler. To ensure that the instanceHandler cannot be used by
     * the user after calling `Finalize`, the member variables of the instanceHandler should be cleared.
     */
    public void clearHandlerInfo() {
        this.instanceId = "";
        this.apiType = null;
    }

    /**
     * Release InstanceHandler, decrease reference.
     *
     * @throws YRException Unified exception types thrown.
     */
    public void release() throws YRException {
        YR.getRuntime().decreaseReference(Collections.singletonList(instanceId));
    }

    /**
     * The member method of the InstanceHandler class allows users to obtain handle information, which can be serialized
     * and stored in a database or other persistence tools. When the tenant context is enabled, handle information can
     * also be obtained through this method and used across tenants.
     *
     * @return Map<String, String> type, The handle information of the InstanceHandler is stored in the kv key-value
     *         pair.
     * @throws YRException Unified exception types thrown.
     *
     * @snippet{trimleft} InstanceExample.java InstanceHandler exportHandler example
     */
    public Map<String, String> exportHandler() throws YRException {
        Map<String, String> out = new HashMap<>();
        out.put(Constants.INSTANCE_KEY, this.instanceId);
        out.put(Constants.API_TYPE, String.valueOf(apiType.getNumber()));
        out.put(Constants.NEED_ORDER, this.needOrder ? "true" : "false");
        Runtime runtime = YR.getRuntime();
        out.put(Constants.INSTANCE_ROUTE, runtime.getInstanceRoute(this.instanceId));
        if (this.realInstanceId.isEmpty()) {
            out.put(Constants.INSTANCE_ID, runtime.getRealInstanceId(this.instanceId));
        } else {
            out.put(Constants.INSTANCE_ID, this.realInstanceId);
        }
        return out;
    }

    /**
     * The member method of the InstanceHandler class allows users to import handle information, which can be obtained
     * and deserialized from persistent tools such as databases. When the tenant context is enabled, this method can
     * also be used to import handle information from other tenant contexts.
     *
     * @param input The handle information of the InstanceHandler is stored in the kv key-value pair.
     * @throws YRException Unified exception types thrown.
     *
     * @snippet{trimleft} InstanceExample.java InstanceHandler importHandler example
     */
    public void importHandler(Map<String, String> input) throws YRException {
        this.instanceId = SdkUtils.defaultIfNotFound(input, Constants.INSTANCE_KEY, "");
        this.apiType = ApiType.valueOf(Integer.parseInt(SdkUtils.defaultIfNotFound(input, Constants.API_TYPE, "0")));
        this.realInstanceId = SdkUtils.defaultIfNotFound(input, Constants.INSTANCE_ID, "");
        this.needOrder = !"false".equals(SdkUtils.defaultIfNotFound(input, Constants.NEED_ORDER, "false"));
        String instanceRoute = SdkUtils.defaultIfNotFound(input, Constants.INSTANCE_ROUTE, "");
        InvokeOptions opts = new InvokeOptions();
        opts.setNeedOrder(needOrder);
        Runtime runtime = YR.getRuntime();
        if (instanceRoute != null && !instanceRoute.isEmpty()) {
            runtime.saveInstanceRoute(this.instanceId, instanceRoute);
        }
        runtime.saveRealInstanceId(this.instanceId, this.realInstanceId, opts);
    }
}
