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
import org.yuanrong.runtime.client.ObjectRef;

public class OptionsExample {
    public static class MyYRApp {
        public static String myFunction(){
            return "hello world";
        }
    }

    public static void main(String[] args) throws YRException {
        //! [function options 样例代码]
        Config conf = new Config("FunctionURN", "ip", "ip", "", false);
        YR.init(conf);
        InvokeOptions opts = new InvokeOptions();
        opts.setCpu(300);
        FunctionHandler<String> f_h = YR.function(MyYRApp::myFunction).options(opts);
        ObjectRef res = f_h.invoke();
        System.out.println("myFunction invoke ref:" + res.getObjId());
        YR.Finalize();
        //! [function options 样例代码]

        //! [instanceCreator options 样例代码]
        InvokeOptions invokeOptions = new InvokeOptions();
        invokeOptions.setCpu(1500);
        invokeOptions.setMemory(1500);
        InstanceCreator instanceCreator = YR.instance(Counter::new).options(invokeOptions);
        InstanceHandler instanceHandler = instanceCreator.invoke(1);
        //! [instanceCreator options 样例代码]

        //! [instance options 样例代码]
        YR.init(conf);
        InvokeOptions options = new InvokeOptions();
        options.setConcurrency(100);
        options.getCustomResources().put("nvidia.com/gpu", 100F);
        InstanceCreator<MyYRApp> f_instance = YR.instance(MyYRApp::new);
        InstanceHandler f_myHandler = f_instance.invoke();
        InstanceFunctionHandler<String> f_h = f_myHandler.function(MyYRApp::myFunction);
        ObjectRef res = f_h.options(options).invoke();
        System.out.println("myFunction invoke ref:" + res.getObjId());
        //expected result:  "myFunction invoke ref: obj-***-***"
        YR.Finalize();
        //! [instance options 样例代码]
    }
}