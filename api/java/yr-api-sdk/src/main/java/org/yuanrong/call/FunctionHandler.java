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
import org.yuanrong.function.YRFuncR;
import org.yuanrong.libruntime.generated.Libruntime.ApiType;
import org.yuanrong.libruntime.generated.Libruntime.FunctionMeta;
import org.yuanrong.runtime.client.ObjectRef;

/**
 * Create an operation class for creating a stateless Java function instance.
 *
 * @note The FunctionHandler class is the handle for Java functions on the cloud; it is the return type of the
 *       interface `YR.function(YRFuncR<R> func)`.\n
 *       Users can use FunctionHandler to create and invoke Java function instances.
 *
 * @since 2022/08/30
 */
public class FunctionHandler<R> extends Handler {
    private YRFuncR<R> func;
    private InvokeOptions options = new InvokeOptions();

    /**
     * FunctionHandler Constructor.
     *
     * @param func Java function name, Yuanrong currently supports 0 ~ 5 parameters and one return value for
     *             user functions.
     */
    public FunctionHandler(YRFuncR<R> func) {
        this.func = func;
    }

    /**
     * Java function call interface.
     *
     * @param args The input parameters required to call the specified method.
     * @return ObjectRef: The "id" of the method's return value in the data system. Use YR.get() to get the actual
     *         return value of the method.
     * @throws YRException the YR exception.
     */
    public ObjectRef invoke(Object... args) throws YRException {
        FunctionMeta functionMeta = getFunctionMeta(func, ApiType.Function);

        return YR.getRuntime().invokeJavaByName(functionMeta, options, args);
    }

    /**
     * Gets func.
     *
     * @return the func.
     */
    public YRFuncR<R> getFunc() {
        return func;
    }

    /**
     * Gets options.
     *
     * @return the options.
     */
    public InvokeOptions getOptions() {
        return options;
    }

    /**
     * Used to dynamically modify the parameters of the called function.
     *
     * @param opt Function call options, used to specify functions such as calling resources.
     * @return FunctionHandler Class handle.
     *
     * @snippet{trimleft} OptionsExample.java function options 样例代码
     */
    public FunctionHandler<R> options(InvokeOptions opt) {
        this.options = new InvokeOptions(opt);
        return this;
    }
}
