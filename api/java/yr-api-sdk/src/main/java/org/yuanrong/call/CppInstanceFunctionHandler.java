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
import org.yuanrong.api.InvokeArg;
import org.yuanrong.api.YR;
import org.yuanrong.exception.YRException;
import org.yuanrong.function.CppInstanceMethod;
import org.yuanrong.libruntime.generated.Libruntime.ApiType;
import org.yuanrong.libruntime.generated.Libruntime.FunctionMeta;
import org.yuanrong.libruntime.generated.Libruntime.LanguageType;
import org.yuanrong.runtime.client.ObjectRef;
import org.yuanrong.utils.SdkUtils;

import lombok.Getter;

import java.util.List;

/**
 * Operation class that calls a C++ stateful function instance member function.
 *
 * @note The CppInstanceFunctionHandler class is the handle of the member function of the C++ class instance after the
 *       Java function creates the C++ class instance.
 *       It is the return value type of the interface `CppInstanceHandler.function`.\n
 *       Users can use the invoke method of CppInstanceFunctionHandler to call member functions of C++ class instances.
 *
 * @param <R> The specific type of a member function.
 *
 * @since 2022/08/26
 */
@Getter
public class CppInstanceFunctionHandler<R> {
    private CppInstanceMethod<R> cppInstanceMethod;

    private InvokeOptions options = new InvokeOptions();

    private String instanceId;

    private String functionId;

    private String className;

    /**
     * The constructor of CppInstanceFunctionHandler.
     *
     * @param instanceId cpp function instance ID, used for invoke function.
     * @param functionId cpp function deployment returns ID.
     * @param className cpp function class name.
     * @param cppInstanceMethod cppInstanceMethod class instance.
     */
    CppInstanceFunctionHandler(String instanceId, String functionId, String className,
            CppInstanceMethod<R> cppInstanceMethod) {
        this.className = className;
        this.cppInstanceMethod = cppInstanceMethod;
        this.instanceId = instanceId;
        this.functionId = functionId;
    }

    /**
     * The member method of the CppInstanceFunctionHandler class is used to call the member function of
     * a cpp class instance.
     *
     * @param args The input parameters required to call the specified method.
     * @return ObjectRef: The "id" of the method's return value in the data system. Use YR.get() to get the actual
     *         return value of the method.
     * @throws YRException Unified exception types thrown.
     *
     * @snippet{trimleft} CppFunctionExample.java CppInstanceFunctionHandler invoke 样例代码
     */
    public ObjectRef invoke(Object... args) throws YRException {
        FunctionMeta functionMeta = FunctionMeta.newBuilder()
                .setClassName("")
                .setFunctionName(String.format("&%s::%s", className, cppInstanceMethod.methodName))
                .setFunctionID(functionId)
                .setSignature("")
                .setLanguage(LanguageType.Cpp)
                .setApiType(ApiType.Function)
                .build();
        List<InvokeArg> packedArgs = SdkUtils.packInvokeArgs(args);
        String objId = YR.getRuntime().invokeInstance(functionMeta, this.instanceId, packedArgs, options);
        return new ObjectRef(objId, cppInstanceMethod.returnType);
    }

    /**
     * Member method of the CppInstanceFunctionHandler class, used to dynamically modify the parameters of the called
     * function.
     *
     * @param opt Function call options, used to specify functions such as calling resources.
     * @return CppInstanceFunctionHandler Class handle.
     */
    public CppInstanceFunctionHandler<R> options(InvokeOptions opt) {
        this.options = new InvokeOptions(opt);
        return this;
    }
}
