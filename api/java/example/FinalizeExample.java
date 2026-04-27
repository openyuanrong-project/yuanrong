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
import org.yuanrong.exception.YRException;
import org.yuanrong.runtime.client.ObjectRef;

public class FinalizeExample {
    //! [Finalize class]
    public static class Counter {
        private int value = 0;

        public int increment() {
            this.value += 1;
            return this.value;
        }

        public static int functionWithAnArgument(int value) {
            return value + 1;
        }
    }

    public static class MyYRApp {
        public static int myFunction() {
            return 1;
        }

        public static int functionWithAnArgument(int value) {
            return value + 1;
        }
    }
    //! [Finalize class]

    public static void main(String[] args) throws YRException {
        //! [Finalize 样例代码]
        // Instance example
        InstanceHandler counter = YR.instance(Counter::new).invoke();
        ObjectRef objectRef = counter.function(Counter::increment).invoke();
        System.out.println(YR.get(objectRef, 3000));

        // Function example
        ObjectRef res = YR.function(MyYRApp::myFunction).invoke();
        System.out.println(YR.get(res, 3000));
        ObjectRef objRef2 = YR.function(MyYRApp::functionWithAnArgument).invoke(1);
        System.out.println(YR.get(objRef2, 3000));
        YR.Finalize();
        //! [Finalize 样例代码]
        //! [finalize ctx 样例代码]
        try {
            ClientInfo info = YR.init(conf);
            String ctx = info.getContext();
            InstanceHandler instanceHandler = YR.instance(YrlibHandler.MyYRApp::new).invoke();
            ObjectRef ref1 = instanceHandler.function(MyYRApp::longCall).invoke();
            String res = (String)YR.get(ref1, 10000);
            System.out.println(res);
            YR.finalize(ctx);
        } catch (YRException exp) {
            System.out.println(exp.getMessage());
        }
        //! [finalize ctx 样例代码]
    }
}