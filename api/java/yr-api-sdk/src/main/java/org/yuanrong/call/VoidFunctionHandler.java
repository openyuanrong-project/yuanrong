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
import org.yuanrong.function.YRFuncVoid;
import org.yuanrong.libruntime.generated.Libruntime.ApiType;
import org.yuanrong.libruntime.generated.Libruntime.FunctionMeta;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Operation class that calls a void function with no return value.
 *
 * @note The VoidFunctionHandler class is a handle for Java functions with no return value; it is the return type of the
 *       interface `YR.function(YRFuncVoid func)`.\n
 *       Users can use VoidFunctionHandler to create and invoke Java functions that do not return values.
 *
 * @since 2022/08/30
 */
public class VoidFunctionHandler extends Handler {
    private static Logger LOGGER = LoggerFactory.getLogger(VoidFunctionHandler.class);

    private final YRFuncVoid func;

    private InvokeOptions options = new InvokeOptions();

    /**
     * The constructor of VoidFunctionHandler.
     *
     * @param func Java function name, supports 0 ~ 5 parameters, no return value user function.
     */
    public VoidFunctionHandler(YRFuncVoid func) {
        this.func = func;
    }

    /**
     * Member method of the VoidFunctionHandler class, used to call void functions.
     *
     * @param args The input parameters required to call the specified method.
     * @throws YRException Unified exception types thrown.
     */
    public void invoke(Object... args) throws YRException {
        FunctionMeta functionMeta = getFunctionMeta(func, ApiType.Function);
        YR.getRuntime().invokeJavaByName(functionMeta, options, args);
    }

    /**
     * The member method of the VoidFunctionHandler class is used to dynamically modify the parameters of the called
     * void function.
     *
     * @param opt Function call options, used to specify functions such as calling resources.
     * @return VoidFunctionHandler Class handle.
     *
     * @snippet{trimleft} VoidFunctionExample.java VoidFunctionHandler options 样例代码
     */
    public VoidFunctionHandler options(InvokeOptions opt) {
        this.options = new InvokeOptions(opt);
        return this;
    }
}
