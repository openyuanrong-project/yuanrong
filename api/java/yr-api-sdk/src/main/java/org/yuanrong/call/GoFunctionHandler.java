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
import org.yuanrong.api.InvokeArg;
import org.yuanrong.api.YR;
import org.yuanrong.exception.YRException;
import org.yuanrong.function.GoFunction;
import org.yuanrong.libruntime.generated.Libruntime.ApiType;
import org.yuanrong.libruntime.generated.Libruntime.FunctionMeta;
import org.yuanrong.libruntime.generated.Libruntime.LanguageType;
import org.yuanrong.runtime.client.ObjectRef;
import org.yuanrong.utils.SdkUtils;

import java.util.List;

/**
 * Class for invoking stateless functions in Go.
 *
 * @note The GoFunctionHandler class is the handle for a Go function created on the cloud by a Java function;
 *       it is the return type of the interface `YR.function(GoFunction<R> func)`.\n
 *       Users can use GoFunctionHandler to create and invoke Go function instances.
 *
 * @since 2024/03/12
 */
public class GoFunctionHandler<R> {
    private final GoFunction<R> func;

    private InvokeOptions options = new InvokeOptions();

    /**
     * The constructor of GoFunctionHandler.
     *
     * @param func GoFunction class instance.
     */
    public GoFunctionHandler(GoFunction<R> func) {
        this.func = func;
    }

    /**
     * Member method of the GoFunctionHandler class, used to call a Go function.
     *
     * @param args The input parameters required to call the specified method.
     * @return ObjectRef: The "id" of the method's return value in the data system. Use YR.get() to get the actual
     *         return value of the method.
     * @throws YRException Unified exception types thrown.
     */
    public ObjectRef invoke(Object... args) throws YRException {
        FunctionMeta functionMeta = FunctionMeta.newBuilder()
                .setClassName("")
                .setFunctionName(func.functionName)
                .setSignature("")
                .setLanguage(LanguageType.Golang)
                .setApiType(ApiType.Function)
                .build();
        List<InvokeArg> packedArgs = SdkUtils.packInvokeArgs(args);
        String objId = YR.getRuntime().invokeByName(functionMeta, packedArgs, options);
        return new ObjectRef(objId, func.returnType);
    }

    /**
     * The member method of the GoFunctionHandler class is used to dynamically modify the parameters of the called
     * go function.
     *
     * @param opt Function call options, used to specify functions such as calling resources.
     * @return GoFunctionHandler Class handle.
     *
     * @snippet{trimleft} GoInstanceExample.java GoFunctionHandle options 样例代码
     */
    public GoFunctionHandler<R> options(InvokeOptions opt) {
        this.options = new InvokeOptions(opt);
        return this;
    }
}
