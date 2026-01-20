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

import org.yuanrong.FunctionWrapper;
import org.yuanrong.InvokeOptions;
import org.yuanrong.api.InvokeArg;
import org.yuanrong.api.YR;
import org.yuanrong.exception.YRException;
import org.yuanrong.function.YRFunc;
import org.yuanrong.function.YRFuncR;
import org.yuanrong.libruntime.generated.Libruntime.ApiType;
import org.yuanrong.libruntime.generated.Libruntime.FunctionMeta;
import org.yuanrong.runtime.Runtime;
import org.yuanrong.runtime.client.ObjectRef;
import org.yuanrong.utils.SdkUtils;

import lombok.Getter;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.List;

/**
 * Operation class that calls a Java stateful function instance member function.
 *
 * @note The class InstanceFunctionHandler is the handle of the member function of the class instance after the Java
 *       class instance is created; it is the return value type of the interface `InstanceHandler.function`.\n
 *       Users can use the invoke method of InstanceFunctionHandler to call member functions of Java class instances.
 *
 * @since 2022/08/30
 */
@Getter
public class InstanceFunctionHandler<R> extends Handler {
    private static Logger LOGGER = LoggerFactory.getLogger(InstanceFunctionHandler.class);

    private YRFunc func;

    private InvokeOptions options = new InvokeOptions();

    private String instanceId;

    private ApiType apiType;

    /**
     * The constructor of the InstanceFunctionHandler.
     *
     * @param func YRFuncR Class instance.
     * @param instanceId Java function instance ID.
     * @param apiType The enumeration class has three values: Function, Faas, and Posix.
     *                It is used internally by Yuanrong to distinguish function types. The default is Actor.
     */
    public InstanceFunctionHandler(YRFuncR<R> func, String instanceId, ApiType apiType) {
        this.func = func;
        this.instanceId = instanceId;
        this.apiType = apiType;
    }

    /**
     * The member method of the InstanceFunctionHandler class is used to call the member function of a Java class
     * instance.
     *
     * @param args The input parameters required to call the specified method.
     * @return ObjectRef: The "id" of the method's return value in the data system. Use YR.get() to get the actual
     *         return value of the method.
     * @throws YRException Unified exception types thrown.
     *
     * @snippet{trimleft} InstanceFunctionExample.java instanceFunction invoke example
     */
    public ObjectRef invoke(Object... args) throws YRException {
        FunctionMeta functionMeta = getFunctionMeta(this.func, this.apiType);
        Runtime runtime = YR.getRuntime();
        FunctionWrapper function = runtime.getJavaFunction(functionMeta);
        SdkUtils.checkJavaParameterTypes(function, args);
        List<InvokeArg> packedArgs = SdkUtils.packInvokeArgs(args);
        String objId = runtime.invokeInstance(functionMeta, this.instanceId, packedArgs, options);
        LOGGER.debug("Succeeded to invoke instance, objectRefId: {}", objId);
        Class<?> returnType = function.getReturnType().orElse(null);
        return new ObjectRef(objId, returnType);
    }

    /**
     * The member method of the InstanceFunctionHandler class is used to dynamically modify the parameters of the called
     * Java function.
     *
     * @param options Function call options, used to specify functions such as calling resources.
     * @return InstanceFunctionHandler Class handle.
     *
     * @snippet{trimleft} OptionsExample.java instance options 样例代码
     */
    public InstanceFunctionHandler<R> options(InvokeOptions options) {
        this.options = new InvokeOptions(options);
        return this;
    }
}
