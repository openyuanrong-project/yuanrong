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
import org.yuanrong.api.YR;
import org.yuanrong.exception.YRException;
import org.yuanrong.function.GoInstanceClass;
import org.yuanrong.libruntime.generated.Libruntime.ApiType;
import org.yuanrong.libruntime.generated.Libruntime.FunctionMeta;
import org.yuanrong.libruntime.generated.Libruntime.LanguageType;
import org.yuanrong.runtime.Runtime;
import org.yuanrong.utils.SdkUtils;

/**
 * Create an operation class for creating a Go stateful function instance.
 *
 * @note The GoInstanceCreator class is a Java function that creates Go class instances;
 *       it is the return type of the interface `YR.instance(GoInstanceClass goInstanceClass)`.\n
 *       Users can use the invoke method of GoInstanceCreator to create a Go class instance and return the
 *       goInstanceHandler class handle.
 *
 * @since 2023/03/12
 */
public class GoInstanceCreator {
    private final GoInstanceClass goInstanceClass;

    private FunctionMeta functionMeta;

    private InvokeOptions options = new InvokeOptions();

    private String name;

    private String nameSpace;

    /**
     * The constructor of GoInstanceCreator.
     *
     * @param goInstanceClass GoInstanceClass class instance.
     */
    public GoInstanceCreator(GoInstanceClass goInstanceClass) {
        this(goInstanceClass, "", "");
    }

    /**
     * The constructor of GoInstanceCreator.
     *
     * @param goInstanceClass GoInstanceClass class instance.
     * @param name The instance name of a named instance. When only name exists, the instance name will be set to name.
     * @param nameSpace Namespace of the named instance. When both name and nameSpace exist, the instance name will be
     *                  concatenated into nameSpace-name. This field is currently only used for concatenation, and
     *                  namespace isolation and other related functions will be completed later.
     */
    public GoInstanceCreator(GoInstanceClass goInstanceClass, String name, String nameSpace) {
        this.goInstanceClass = goInstanceClass;
        this.name = name;
        this.nameSpace = nameSpace;
    }

    /**
     * The member method of the GoInstanceCreator class is used to create a Go class instance.
     *
     * @param args The input parameters required to create an instance of the class.
     * @return GoInstanceHandler class handle.
     * @throws YRException Unified exception types thrown.
     */
    public GoInstanceHandler invoke(Object... args) throws YRException {
        String funcName = goInstanceClass.className;
        this.functionMeta = FunctionMeta.newBuilder()
                .setClassName("")
                .setFunctionName(funcName)
                .setSignature("")
                .setLanguage(LanguageType.Golang)
                .setApiType(ApiType.Function)
                .setName(name)
                .setNs(nameSpace)
                .build();
        Runtime runtime = YR.getRuntime();
        String instanceId = runtime.createInstance(this.functionMeta, SdkUtils.packInvokeArgs(args), options);
        return new GoInstanceHandler(instanceId, this.goInstanceClass.className);
    }

    /**
     * The member method of the GoInstanceCreator class is used to dynamically modify the parameters for creating a Go
     * function instance.
     *
     * @param opt Function call options, used to specify functions such as calling resources.
     * @return GoInstanceCreator Class handle.
     *
     * @snippet{trimleft} GoInstanceExample.java GoInstanceCreator options 样例代码
     */
    public GoInstanceCreator options(InvokeOptions opt) {
        this.options = new InvokeOptions(opt);
        return this;
    }

    /**
     * The member method of the GoInstanceCreator class is used to obtain function metadata.
     *
     * @return FunctionMeta class instance: function metadata.
     */
    public FunctionMeta getFunctionMeta() {
        return this.functionMeta;
    }
}
