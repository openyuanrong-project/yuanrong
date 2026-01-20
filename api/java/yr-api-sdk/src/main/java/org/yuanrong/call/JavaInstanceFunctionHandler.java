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
import org.yuanrong.function.JavaInstanceMethod;
import org.yuanrong.libruntime.generated.Libruntime.ApiType;
import org.yuanrong.libruntime.generated.Libruntime.FunctionMeta;
import org.yuanrong.libruntime.generated.Libruntime.LanguageType;
import org.yuanrong.runtime.client.ObjectRef;
import org.yuanrong.utils.SdkUtils;

import lombok.Getter;

import java.util.List;

/**
 * Class that invokes Java instance member functions.
 *
 * @note The JavaInstanceFunctionHandler class is a handle to the member functions of a Java class instance after the
 *       Java class instance is created; it is the return type of the interface `JavaInstanceHandler.function`.\n
 *       Users can use the invoke method of JavaInstanceFunctionHandler to call member functions of
 *       Java class instances.
 *
 * @since 2024/04/16
 */
@Getter
public class JavaInstanceFunctionHandler<R> {
    private JavaInstanceMethod<R> javaInstanceMethod;

    private InvokeOptions options = new InvokeOptions();

    private String instanceId;

    private String functionId;

    private String className;

    /**
     * The constructor of JavaInstanceFunctionHandler.
     *
     * @param <R> the type of the object.
     * @param instanceId Java function instance ID.
     * @param functionId Java function deployment returns an ID.
     * @param className Java function class name.
     * @param javaInstanceMethod JavaInstanceMethod class instance.
     */
    JavaInstanceFunctionHandler(String instanceId, String functionId, String className,
            JavaInstanceMethod<R> javaInstanceMethod) {
        this.className = className;
        this.javaInstanceMethod = javaInstanceMethod;
        this.instanceId = instanceId;
        this.functionId = functionId;
    }

    /**
     * The member method of the JavaInstanceFunctionHandler class is used to call the member function of a Java class
     * instance.
     *
     * @param args The input parameters required to call the specified method.
     * @return ObjectRef: The "id" of the method's return value in the data system. Use YR.get() to get the actual
     *         return value of the method.
     * @throws YRException Unified exception types thrown.
     */
    public ObjectRef invoke(Object... args) throws YRException {
        FunctionMeta functionMeta = FunctionMeta.newBuilder()
                .setClassName(className)
                .setFunctionName(javaInstanceMethod.methodName)
                .setFunctionID(functionId)
                .setSignature("")
                .setLanguage(LanguageType.Java)
                .setApiType(ApiType.Function)
                .build();
        List<InvokeArg> packedArgs = SdkUtils.packInvokeArgs(args);
        String objId = YR.getRuntime().invokeInstance(functionMeta, this.instanceId, packedArgs, options);
        return new ObjectRef(objId, javaInstanceMethod.returnType);
    }

    /**
     * The member method of the JavaInstanceFunctionHandler class is used to dynamically modify the parameters of the
     * called Java function.
     *
     * @param opt Function call options, used to specify functions such as calling resources.
     * @return JavaInstanceFunctionHandler Class handle.
     *
     * @snippet{trimleft} JavaInstanceExample.java JavaInstanceFunctionHandler options 样例代码
     */
    public JavaInstanceFunctionHandler<R> options(InvokeOptions opt) {
        this.options = new InvokeOptions(opt);
        return this;
    }
}
