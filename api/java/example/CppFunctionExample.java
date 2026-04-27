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
import org.yuanrong.api.YR;
import org.yuanrong.exception.YRException;
import org.yuanrong.function.CppFunction;
import org.yuanrong.function.CppInstanceClass;
import org.yuanrong.function.CppInstanceMethod;
import org.yuanrong.runtime.client.ObjectRef;

import java.util.Map;

public class CppFunctionExample {
    public static void main(String[] args) throws YRException {
        Config conf = new Config();
        YR.init(conf);

        //! [CppFunction 样例代码]
        ObjectRef res = YR.function(CppFunction.of("PlusOne", int.class)).invoke(1);
        //! [CppFunction 样例代码]

        //! [CppInstanceClass 样例代码]
        CppInstanceClass cppInstance = CppInstanceClass.of("Counter", "FactoryCreate");
        //! [CppInstanceClass 样例代码]

        //! [CppFunctionHandler invoke 样例代码]
        CppFunctionHandler cppFuncHandler = YR.function(CppFunction.of("Add", int.class));
        ObjectRef ref = cppFuncHandler.invoke(1);
        int result = YR.get(ref, 15);
        //! [CppFunctionHandler invoke 样例代码]

        //! [CppInstanceCreator options 样例代码]
        InvokeOptions invokeOptions = new InvokeOptions();
        invokeOptions.setCpu(1500);
        invokeOptions.setMemory(1500);
        CppInstanceCreator cppInstanceCreator = YR.instance(CppInstanceClass.of("Counter","FactoryCreate")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").options(invokeOptions);
        CppInstanceHandler cppInstanceHandler = cppInstanceCreator.invoke(1);
        ObjectRef ref = cppInstanceHandler.function(CppInstanceMethod.of("Add", int.class)).invoke(5);
        int res = (int)YR.get(ref, 100);
        //! [CppInstanceCreator options 样例代码]

        //! [CppInstanceFunctionHandler invoke 样例代码]
        CppInstanceHandler cppInstanceHandler = YR.instance(CppInstanceClass.of("Counter","FactoryCreate")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke(1);
        CppInstanceFunctionHandler cppInsFuncHandler = cppInstanceHandler.function(CppInstanceMethod.of("Add", int.class));
        ObjectRef ref = cppInsFuncHandler.invoke(5);
        int res = (int)YR.get(ref, 100);
        //! [CppInstanceFunctionHandler invoke 样例代码]

        //! [cpp exportHandler 样例代码]
        CppInstanceHandler cppInstanceHandler = YR.instance(CppInstanceClass.of("Counter", "FactoryCreate"))
            .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest")
            .invoke(1);
        Map<String, String> out = cppInstanceHandler.exportHandler();
        // Serialize out and store in a database or other persistence tool.
        //! [cpp exportHandler 样例代码]

        //! [CppInstanceHandler function 样例代码]
        CppInstanceHandler cppInstanceHandler = YR.instance(CppInstanceClass.of("Counter", "FactoryCreate"))
            .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest")
            .invoke(1);
        CppInstanceFunctionHandler cppInstFuncHandler = cppInstanceHandler.function(
            CppInstanceMethod.of("Add", int.class));
        ObjectRef ref = cppInstFuncHandler.invoke(5);
        int res = (int) YR.get(ref, 100);
        //! [CppInstanceHandler function 样例代码]

        //! [CppInstanceHandler importHandler 样例代码]
        CppInstanceHandler cppInstanceHandler = new CppInstanceHandler();
        cppInstanceHandler.importHandler(in);
        // The input parameter is obtained by retrieving and deserializing from a database or other persistent storage.
        //! [CppInstanceHandler importHandler 样例代码]

        //! [CppInstanceHandler terminate 样例代码]
        CppInstanceHandler cppInstanceHandler = YR.instance(CppInstanceClass.of("Counter", "FactoryCreate"))
            .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest")
            .invoke(1);
        cppInstanceHandler.terminate();
        //! [CppInstanceHandler terminate 样例代码]

        //! [CppInstanceHandler terminate sync 样例代码]
        CppInstanceHandler cppInstanceHandler = YR.instance(CppInstanceClass.of("Counter", "FactoryCreate"))
            .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest")
            .invoke(1);
        cppInstanceHandler.terminate(true);
        //! [CppInstanceHandler terminate sync 样例代码]

        //! [CppInstanceMethod 样例代码]
        ObjectRef<int> res = cppInstance.function(CppInstanceMethod.of("Add", int.class)).invoke(1);
        //! [CppInstanceMethod 样例代码]

        YR.Finalize();
    }
}