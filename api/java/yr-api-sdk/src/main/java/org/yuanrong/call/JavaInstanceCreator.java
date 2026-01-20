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
import org.yuanrong.function.JavaInstanceClass;
import org.yuanrong.libruntime.generated.Libruntime.ApiType;
import org.yuanrong.libruntime.generated.Libruntime.FunctionMeta;
import org.yuanrong.libruntime.generated.Libruntime.LanguageType;
import org.yuanrong.runtime.Runtime;
import org.yuanrong.runtime.util.Constants;
import org.yuanrong.utils.SdkUtils;

/**
 * Creates an operation class for creating a Java stateful function instance.
 *
 * @note The JavaInstanceCreator class is a creator of Java class instances; it is the return type of the interface
 *       `YR.instance(JavaInstanceClass javaInstanceClass)`.\n
 *       Users can use the invoke method of JavaInstanceCreator to create Java class instances and return the handle
 *       class JavaInstanceHandler.
 *
 * @since 2024/04/16
 */
public class JavaInstanceCreator {
    private final JavaInstanceClass javaInstanceClass;

    private FunctionMeta functionMeta;

    private InvokeOptions options = new InvokeOptions();

    private String name;

    private String nameSpace;

    private String functionUrn = "";

    /**
     * The constructor of JavaInstanceCreator.
     *
     * @param javaInstanceClass JavaInstanceClass class instance.
     */
    public JavaInstanceCreator(JavaInstanceClass javaInstanceClass) {
        this(javaInstanceClass, "", "");
    }

    /**
     * The constructor of JavaInstanceCreator.
     *
     * @param javaInstanceClass JavaInstanceClass class instance.
     * @param name The instance name of a named instance. When only name exists, the instance name will be set to name
     * @param nameSpace Namespace of the named instance. When both name and nameSpace exist, the instance name will be
     *                  concatenated into nameSpace-name. This field is currently only used for concatenation, and
     *                  namespace isolation and other related functions will be completed later.
     */
    public JavaInstanceCreator(JavaInstanceClass javaInstanceClass, String name, String nameSpace) {
        this.javaInstanceClass = javaInstanceClass;
        this.name = name;
        this.nameSpace = nameSpace;
    }

    /**
     * The member method of the JavaInstanceCreator class is used to create a Java class instance.
     *
     * @param args The input parameters required to create an instance of the class.
     * @return JavaInstanceHandler Class handle.
     * @throws YRException Unified exception types thrown.
     */
    public JavaInstanceHandler invoke(Object... args) throws YRException {
        String funcName = "<init>";
        String functionID = "";
        if (!functionUrn.isEmpty()) {
            functionID = SdkUtils.reformatFunctionUrn(functionUrn);
        }
        this.functionMeta = FunctionMeta.newBuilder()
                .setClassName(javaInstanceClass.className)
                .setFunctionName(funcName)
                .setFunctionID(functionID)
                .setApiType(ApiType.Function)
                .setSignature("")
                .setName(name)
                .setNs(nameSpace)
                .setLanguage(LanguageType.Java)
                .build();
        String concurrency = options.getCustomExtensions().getOrDefault(Constants.CONCURRENCY, "1");
        if ("1".equals(concurrency)) {
            options.setNeedOrder(true);
        }
        Runtime runtime = YR.getRuntime();
        String instanceId = runtime.createInstance(this.functionMeta, SdkUtils.packInvokeArgs(args), options);
        JavaInstanceHandler javaInstanceHandler = new JavaInstanceHandler(instanceId, functionID,
            this.javaInstanceClass.className);
        javaInstanceHandler.setNeedOrder(options.isNeedOrder());
        return javaInstanceHandler;
    }

    /**
     * When Java calls a Java stateful function, set the functionUrn for the function.
     *
     * @param urn functionUrn, can be obtained after the function is deployed. The tenant ID in the function urn must be
     *            consistent with the tenant ID configured in the config. For information about tenant ID configuration,
     *            see "About tenant ID" in Config.
     * @return JavaInstanceCreator, with built-in invoke method, can create instances of this Java function class.
     *
     * @snippet{trimleft} SetUrnExample.java set urn of java invoke java stateful function
     */
    public JavaInstanceCreator setUrn(String urn) {
        this.functionUrn = urn;
        return this;
    }

    /**
     * The member method of the JavaInstanceCreator class is used to dynamically modify the parameters for creating a
     * Java function instance.
     *
     * @param opt Function call options, used to specify functions such as calling resources.
     * @return JavaInstanceCreator Class handle.
     *
     * @snippet{trimleft} JavaInstanceExample.java JavaInstanceCreator options 样例代码
     */
    public JavaInstanceCreator options(InvokeOptions opt) {
        this.options = new InvokeOptions(opt);
        return this;
    }

    /**
     * The member method of the JavaInstanceCreator class is used to return function metadata.
     *
     * @return FunctionMeta class instance: function metadata.
     */
    public FunctionMeta getFunctionMeta() {
        return this.functionMeta;
    }
}
