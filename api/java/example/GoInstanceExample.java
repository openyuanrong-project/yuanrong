/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
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

public class GoInstanceExample {
    public static void main(String[] args) throws Exception {
        //! [GoFunctionHandle options 样例代码]
        InvokeOptions invokeOptions = new InvokeOptions();
        invokeOptions.setCpu(1500);
        invokeOptions.setMemory(1500);
        GoFunction<Integer> goFunction = GoFunction.of("PlusOne", int.class, 1);
        GoFunctionHandler<Integer> goFuncHandler = YR.function(goFunction).options(invokeOptions);
        ObjectRef ref = goFuncHandler.invoke();
        int result = YR.get(ref, 15);
        //! [GoFunctionHandle options 样例代码]

        //! [GoInstanceCreator options 样例代码]
        InvokeOptions invokeOptions = new InvokeOptions();
        invokeOptions.setCpu(1500);
        invokeOptions.setMemory(1500);
        GoInstanceCreator goInstanceCreator = YR.instance(GoInstanceClass.of("Counter")).options(invokeOptions);
        GoInstanceHandler goInstanceHandler = goInstanceCreator.invoke(1);
        ObjectRef ref = goInstanceHandler.function(GoInstanceMethod.of("Add", int.class)).invoke(5);
        int res = (int)YR.get(ref, 100);
        //! [GoInstanceCreator options 样例代码]

        //! [GoInstanceHandler function example]
        GoInstanceCreator goInstanceCreator = YR.instance(GoInstanceClass.of("Counter"));
        GoInstanceHandler goInstanceHandler = goInstanceCreator.invoke(1);
        GoInstanceFunctionHandler goInstFuncHandler = goInstanceHandler.function(GoInstanceMethod.of("Add", int.class));
        ObjectRef ref = goInstFuncHandler.invoke(5);
        int res = (int)YR.get(ref, 100);
        //! [GoInstanceHandler function example]

        //! [GoInstanceFunctionHandler options 样例代码]
        InvokeOptions invokeOptions = new InvokeOptions();
        invokeOptions.addCustomExtension("app_name", "myApp");
        GoInstanceHandler goInstanceHandler = YR.instance(GoInstanceClass.of("Counter")).invoke(1);
        GoInstanceFunctionHandler goInstFuncHandler = goInstanceHandler.function(GoInstanceMethod.of("Add", int.class)).options(invokeOptions);
        ObjectRef ref = goInstFuncHandler.invoke(5);
        String res = (String)YR.get(ref, 100);
        //! [GoInstanceFunctionHandler options 样例代码]

        //! [GoInstanceHandler terminate example]
        GoInstanceCreator goInstanceCreator = YR.instance(GoInstanceClass.of("Counter"));
        GoInstanceHandler goInstanceHandler = goInstanceCreator.invoke(1);
        goInstanceHandler.terminate();
        //! [GoInstanceHandler terminate example]

        //! [GoInstanceHandler terminate sync example]
        GoInstanceCreator goInstanceCreator = YR.instance(GoInstanceClass.of("Counter"));
        GoInstanceHandler goInstanceHandler = goInstanceCreator.invoke(1);
        goInstanceHandler.terminate(true);
        //! [GoInstanceHandler terminate sync example]
    }
}