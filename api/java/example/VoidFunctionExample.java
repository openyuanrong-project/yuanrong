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
import org.yuanrong.invokeOptions;
import org.yuanrong.call.VoidFunctionHandler;
import org.yuanrong.call.VoidInstanceFunctionHandler;


public class VoidFunctionExample {
    public static void main(String[] args) throws Exception {
        //! [VoidFunctionHandler options 样例代码]
        Config conf = new Config("FunctionURN", "ip", "ip", "", false);
        YR.init(conf);

        InvokeOptions invokeOptions = new InvokeOptions();
        invokeOptions.setCpu(1500);
        invokeOptions.setMemory(1500);
        VoidFunctionHandler<String> functionHandler = YR.function(MyYRApp::myVoidFunction).options(invokeOptions);
        functionHandler.invoke();
        YR.Finalize();
        //! [VoidFunctionHandler options 样例代码]

        //! [VoidInstanceFunctionHandler options 样例代码]
        Config conf = new Config("FunctionURN", "ip", "ip", "", false);
        YR.init(conf);

        InvokeOptions invokeOptions = new InvokeOptions();
        invokeOptions.addCustomExtensions("app_name", "myApp");
        InstanceHandler instanceHandler = YR.instance(MyYRApp::new).invoke();
        VoidInstanceFunctionHandler insFuncHandler = instanceHandler.function(MyYRApp::myVoidFunction).options(invokeOptions);
        insFuncHandler.invoke();
        YR.Finalize();
        //! [VoidInstanceFunctionHandler options 样例代码]
    }
}