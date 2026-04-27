/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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

package example;

import org.yuanrong.Config;

public class JavaInstanceExample {
    public static void main(String[] args) throws Exception {
        //! [JavaInstanceClass 样例代码]
        JavaInstanceHandler javaInstance = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).invoke();
        //! [JavaInstanceClass 样例代码]

        //! [JavaInstanceMethod 样例代码]
        ObjectRef ref1 = javaInstance.function(JavaInstanceMethod.of("smallCall", String.class)).invoke();
        //! [JavaInstanceMethod 样例代码]

        //! [JavaFunction 样例代码]
        ObjectRef ref1 = YR.function(JavaFunction.of("com.example.YrlibHandler$MyYRApp", "smallCall", String.class)).invoke();
        //! [JavaFunction 样例代码]

        //! [JavaFunctionHandle options 样例代码]
        InvokeOptions invokeOptions = new InvokeOptions();
        invokeOptions.setCpu(1500);
        invokeOptions.setMemory(1500);
        JavaFunctionHandler javaFuncHandler = YR.function(JavaFunction.of("com.example.YrlibHandler$MyYRApp", "smallCall", String.class))
                        .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest");
        ObjectRef ref = javaFuncHandler.options(invokeOptions).invoke();
        String result = (String)YR.get(ref, 15);
        //! [JavaFunctionHandle options 样例代码]

        //! [JavaInstanceCreator options 样例代码]
        InvokeOptions invokeOptions = new InvokeOptions();
        invokeOptions.setCpu(1500);
        invokeOptions.setMemory(1500);
        JavaInstanceCreator javaInstanceCreator = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").options(invokeOptions);
        JavaInstanceHandler javaInstanceHandler = javaInstanceCreator.invoke();
        ObjectRef ref = javaInstanceHandler.function(JavaInstanceMethod.of("smallCall", String.class)).invoke();
        String res = (String)YR.get(ref, 100);
        //! [JavaInstanceCreator options 样例代码]

        //! [JavaInstanceFunctionHandler options 样例代码]
        InvokeOptions invokeOptions = new InvokeOptions();
        invokeOptions.addCustomExtensions("app_name", "myApp");
        JavaInstanceHandler javaInstanceHandler = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke();
        JavaInstanceFunctionHandler javaInstFuncHandler = javaInstanceHandler.function(JavaInstanceMethod.of("smallCall", String.class)).options(invokeOptions);
        ObjectRef ref = javaInstFuncHandler.invoke();
        String res = (String)YR.get(ref, 100);
        //! [JavaInstanceFunctionHandler options 样例代码]

        //! [JavaInstanceHandler function example]
        JavaInstanceHandler javaInstanceHandler = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke();
        JavaInstanceFunctionHandler javaInstFuncHandler = javaInstanceHandler.function(JavaInstanceMethod.of("smallCall", String.class))
        ObjectRef ref = javaInstFuncHandler.invoke();
        String res = (String)YR.get(ref, 100);
        //! [JavaInstanceHandler function example]

        //! [JavaInstanceHandler terminate example]
        JavaInstanceHandler javaInstanceHandler = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke();
        javaInstanceHandler.terminate();
        //! [JavaInstanceHandler terminate example]

        //! [JavaInstanceHandler terminate sync example]
        JavaInstanceHandler javaInstanceHandler = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke();
        javaInstanceHandler.terminate(true);
        //! [JavaInstanceHandler terminate sync example]

        //! [JavaInstanceHandler exportHandler example]
        JavaInstanceHandler javaInstanceHandler = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke();
        Map<String, String> out = javaInstanceHandler.exportHandler();
        // Serialize out and store in a database or other persistence tool.
        //! [JavaInstanceHandler exportHandler example]

        //! [JavaInstanceHandler importHandler example]
        JavaInstanceHandler javaInstanceHandler = new JavaInstanceHandler();
        javaInstanceHandler.importHandler(in);
        // The input parameter is obtained by retrieving and deserializing from a database or other persistent storage.
        //! [JavaInstanceHandler importHandler example]
    }
}