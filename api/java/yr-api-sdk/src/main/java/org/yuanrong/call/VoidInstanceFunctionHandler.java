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
import org.yuanrong.api.InvokeArg;
import org.yuanrong.api.YR;
import org.yuanrong.exception.YRException;
import org.yuanrong.function.YRFuncVoid;
import org.yuanrong.libruntime.generated.Libruntime.ApiType;
import org.yuanrong.libruntime.generated.Libruntime.FunctionMeta;
import org.yuanrong.utils.SdkUtils;

import lombok.Getter;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.List;

/**
 * Operation class that calls a Java class instance member function without a return value.
 *
 * @note The VoidInstanceFunctionHandler class is a handle to the member function of a class instance
 *       after creating a Java class instance without a return value; it is the return type of the
 *       interface `InstanceHandler.function`.\n
 *       Users can use the invoke method of VoidInstanceFunctionHandler to call member functions of Java class instances
 *       without return values.
 *
 * @since 2022/08/30
 */
@Getter
public class VoidInstanceFunctionHandler extends Handler {
    private static Logger LOGGER = LoggerFactory.getLogger(VoidInstanceFunctionHandler.class);

    private final YRFuncVoid func;

    private InvokeOptions options = new InvokeOptions();

    private String instanceId;

    private ApiType apiType;

    /**
     * The constructor of VoidInstanceFunctionHandler.
     *
     * @param func Java function name, supports 0 ~ 5 parameters, no return value user function.
     * @param instanceId Java function instance ID.
     * @param apiType The enumeration class has three values: Function, Faas, and Posix.
     *                It is used internally by Yuanrong to distinguish function types. The default is Actor.
     */
    public VoidInstanceFunctionHandler(YRFuncVoid func, String instanceId, ApiType apiType) {
        this.func = func;
        this.instanceId = instanceId;
        this.apiType = apiType;
    }

    /**
     * Member method of the VoidInstanceFunctionHandler class, used to call member functions of void class instances.
     *
     * @param args The input parameters required to call the specified method.
     * @throws YRException Unified exception types thrown.
     */
    public void invoke(Object... args) throws YRException {
        FunctionMeta functionMeta = getFunctionMeta(func, this.apiType);
        List<InvokeArg> packedArgs = SdkUtils.packInvokeArgs(args);
        YR.getRuntime().invokeInstance(functionMeta, this.instanceId, packedArgs, options);
        LOGGER.debug("Succeeded to invoke instance({}) function:{}", this.instanceId, functionMeta.getFunctionName());
    }

    /**
     * Member method of the VoidInstanceFunctionHandler class, used to dynamically modify the parameters of the called
     * void function.
     *
     * @param options Function call options, used to specify functions such as calling resources.
     * @return VoidInstanceFunctionHandler Class handle.
     *
     * @snippet{trimleft} VoidFunctionExample.java VoidInstanceFunctionHandler options 样例代码
     */
    public VoidInstanceFunctionHandler options(InvokeOptions options) {
        this.options = new InvokeOptions(options);
        return this;
    }
}
