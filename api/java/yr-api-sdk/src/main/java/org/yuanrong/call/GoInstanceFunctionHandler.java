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
import org.yuanrong.function.GoInstanceMethod;
import org.yuanrong.libruntime.generated.Libruntime.ApiType;
import org.yuanrong.libruntime.generated.Libruntime.FunctionMeta;
import org.yuanrong.libruntime.generated.Libruntime.LanguageType;
import org.yuanrong.runtime.client.ObjectRef;
import org.yuanrong.utils.SdkUtils;

import lombok.Getter;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

/**
 * The operation class that calls the go stateful function instance member function.
 *
 * @note The GoInstanceFunctionHandler class is the handle of the member function of the Go class instance after the
 *       Java function creates the Go class instance; it is the return value type of the interface
 *       `GoInstanceHandler.function`.\n
 *       Users can use the invoke method of GoInstanceFunctionHandler to call member functions of Go class instances.
 *
 * @since 2024/03/12
 */
@Getter
public class GoInstanceFunctionHandler<R> {
    private GoInstanceMethod<R> goInstanceMethod;

    private InvokeOptions options = new InvokeOptions();

    private String instanceId;

    private String className;

    /**
     * The constructor of GoInstanceFunctionHandler.
     *
     * @param instanceId Go function instance id.
     * @param className Go function class name.
     * @param goInstanceMethod GoInstanceMethod class instance.
     */
    GoInstanceFunctionHandler(String instanceId, String className, GoInstanceMethod<R> goInstanceMethod) {
        this.className = className;
        this.goInstanceMethod = goInstanceMethod;
        this.instanceId = instanceId;
    }

    /**
     * The member method of the GoInstanceFunctionHandler class is used to call the member function of a Go class
     * instance.
     *
     * @param args The input parameters required to call the specified method.
     * @return ObjectRef: The "id" of the method's return value in the data system. Use YR.get() to get the actual
     *         return value of the method.
     * @throws YRException Unified exception types thrown.
     */
    public List<ObjectRef> invoke(Object... args) throws YRException {
        FunctionMeta functionMeta = FunctionMeta.newBuilder()
                .setClassName("")
                .setFunctionName(this.goInstanceMethod.methodName)
                .setSignature("")
                .setLanguage(LanguageType.Golang)
                .setApiType(ApiType.Function)
                .build();
        List<InvokeArg> packedArgs = SdkUtils.packInvokeArgs(args);
        String objId = YR.getRuntime().invokeInstance(functionMeta, this.instanceId, packedArgs, options);
        return new ArrayList<>(Arrays.asList(new ObjectRef(objId, goInstanceMethod.returnType)));
    }

    /**
     * The member method of the GoInstanceFunctionHandler class is used to dynamically modify the parameters of the
     * called GO function.
     *
     * @param opt Function call options, used to specify functions such as calling resources.
     * @return GoInstanceFunctionHandler Class handle.
     *
     * @snippet{trimleft} GoInstanceExample.java GoInstanceFunctionHandler options 样例代码
     */
    public GoInstanceFunctionHandler<R> options(InvokeOptions opt) {
        this.options = new InvokeOptions(opt);
        return this;
    }
}
