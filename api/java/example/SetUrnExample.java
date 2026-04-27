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

import org.yuanrong.api.YR;

public class SetUrnExample {
    public static class MyYRApp {
        public static String smallCall() {
            return "aaa";
        }
    }

    public static void main(String[] args) throws Exception {
        //! [set urn of java invoke cpp stateless function]
        ObjectRef ref1 = YR.function(CppFunction.of("PlusOne", int.class))
            .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke(5);
        int res = (int)YR.get(ref1, 100);
        //! [set urn of java invoke cpp stateless function]

        //! [set urn of java invoke cpp stateful function]
        CppInstanceHandler cppInstance = YR.instance(CppInstanceClass.of("Counter","FactoryCreate"))
            .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke(1);
        ObjectRef ref1 = cppInstance.function(CppInstanceMethod.of("Add", int.class)).invoke(5);
        int res = (int)YR.get(ref1, 100);
        //! [set urn of java invoke cpp stateful function]

        //! [set urn of java invoke java stateless function]
        ObjectRef ref1 = YR.function(JavaFunction.of("com.example.YrlibHandler$MyYRApp", "smallCall", String.class))
            .setUrn("sn:cn:yrk:default:function:0-perf-callee:$latest").invoke();
        String res = (String)YR.get(ref1, 100);
        //! [set urn of java invoke java stateless function]

        //! [set urn of java invoke java stateful function]
        JavaInstanceHandler javaInstance = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp"))
            .setUrn("sn:cn:yrk:default:function:0-perf-callee:$latest").invoke();
        ObjectRef ref1 = javaInstance.function(JavaInstanceMethod.of("smallCall", String.class)).invoke();
        String res = (String)YR.get(ref1, 100);
        //! [set urn of java invoke java stateful function]
    }
}
