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

public class InstanceExample {
    public static void main(String[] args) throws Exception {
        //! [InstanceHandler terminate example]
        InstanceHandler InstanceHandler = YR.instance(Counter::new).invoke(1);
        InstanceHandler.terminate();
        //! [InstanceHandler terminate example]

        //! [InstanceHandler terminate sync example]
        InstanceHandler InstanceHandler = YR.instance(Counter::new).invoke(1);
        InstanceHandler.terminate(true);
        //! [InstanceHandler terminate sync example]

        //! [InstanceHandler exportHandler example]
        InstanceHandler instHandler = YR.instance(MyYRApp::new).invoke();
        Map<String, String> out = instHandler.exportHandler();
        // Serialize out and store in a database or other persistence tool.
        //! [InstanceHandler exportHandler example]

        //! [InstanceHandler importHandler example]
        InstanceHandler instanceHandler = new InstanceHandler();
        instanceHandler.importHandler(in);
        // The input parameter is obtained by retrieving and deserializing from a database or other persistent storage.
        //! [InstanceHandler importHandler example]

        //! [InstanceHandler function example]
        InstanceHandler InstanceHandler = YR.instance(Counter::new).invoke(1);
        ObjectRef ref = InstanceHandler.function(Counter::Add).invoke(5);
        int res = (int)YR.get(ref, 100);
        InstanceHandler.terminate();
        //! [InstanceHandler function example]

        //! [Instance example]
        public static class MyYRApp{
            public static String myFunction() throws YRException, ExecutionException{
                return "hello world";
            }
        }
        public static void main(String[] args) throws Exception {
            Config conf = new Config("FunctionURN", "ip", "ip", ""， false);
            YR.init(conf);
            MyYRApp myapp = new MyYRApp();
            InstanceCreator<MyYRApp> myYRapp = YR.instance(MyYRApp::new);
            InstanceHandler h_myYRapp = myYRapp.invoke();
            System.out.println("InstanceHandler:" + h_myYRapp);
            InstanceFunctionHandler<String> f_h_myYRapp = h_myYRapp.function(MyYRApp::myFunction);
            ObjectRef res = f_h_myYRapp.invoke();
            System.out.println("myFunction invoke ref:" + res.getObjId());
            YR.Finalize();
        }
        //! [Instance example]

        //! [Instance name example]
        public static class MyYRApp{
            public static String myFunction() throws YRException, ExecutionException{
                return "hello world";
            }
        }
        public static void main(String[] args) throws Exception {
            Config conf = new Config("FunctionURN", "ip", "ip", "", false);
            YR.init(conf);
            MyYRApp myapp = new MyYRApp();
            // The instance name of this named instance is funcB
            InstanceCreator<MyYRApp> myYRapp = YR.instance(MyYRApp::new, "funcB");
            InstanceHandler h_myYRapp = myYRapp.invoke();
            System.out.println("InstanceHandler:" + h_myYRapp);
            InstanceFunctionHandler<String> f_h_myYRapp = h_myYRapp.function(MyYRApp::myFunction);
            ObjectRef res = f_h_myYRapp.invoke();
            System.out.println("myFunction invoke ref:" + res.getObjId());
            // A handle to a named instance of funcB will be generated, and the funcB instance will be reused.
            InstanceCreator<MyYRApp> creator = YR.instance(MyYRApp::new, "funcB");
            InstanceHandler handler = creator.invoke();
            System.out.println("InstanceHandler:" + handler);
            InstanceFunctionHandler<String> funcHandler = handler.function(MyYRApp::myFunction);
            res = funcHandler.invoke();
            System.out.println("myFunction invoke ref:" + res.getObjId());
            YR.Finalize();
        }
        //! [Instance name example]

        //! [Instance nameSpace example]
        public static class MyYRApp{
            public static String myFunction() throws YRException, ExecutionException{
                return "hello world";
            }
        }
        public static void main(String[] args) throws Exception {
            Config conf = new Config("FunctionURN", "ip", "ip", "", false);
            YR.init(conf);
            MyYRApp myapp = new MyYRApp();
            // The instance name of this named instance is nsA-funcB
            InstanceCreator<MyYRApp> myYRapp = YR.instance(MyYRApp::new, "funcB", "nsA");
            InstanceHandler h_myYRapp = myYRapp.invoke();
            System.out.println("InstanceHandler:" + h_myYRapp);
            InstanceFunctionHandler<String> f_h_myYRapp = h_myYRapp.function(MyYRApp::myFunction);
            ObjectRef res = f_h_myYRapp.invoke();
            System.out.println("myFunction invoke ref:" + res.getObjId());
            // A handle to the named instance of nsA-funcB will be generated, reusing the nsA-funcB instance
            InstanceCreator<MyYRApp> creator = YR.instance(MyYRApp::new, "funcB", "nsA");
            InstanceHandler handler = creator.invoke();
            System.out.println("InstanceHandler:" + handler);
            InstanceFunctionHandler<String> funcHandler = handler.function(MyYRApp::myFunction);
            res = funcHandler.invoke();
            System.out.println("myFunction invoke ref:" + res.getObjId());
            YR.Finalize();
        }
        //! [Instance nameSpace example]
    }
}