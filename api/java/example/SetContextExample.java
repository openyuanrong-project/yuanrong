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
import org.yuanrong.api.ClientInfo;
import org.yuanrong.api.YR;

public class SetContextExample {
    public static class MyYRApp {
        public static String smallCall() {
            return "aaa";
        }
    }

    public static void main(String[] args) throws Exception {
        Config conf = new Config();
        Config conf1 = new Config();
        Config conf2 = new Config();

        //! [change context]
        ClientInfo info1 = YR.init(conf1);
        String ctx1 = info1.getContext();
        ClientInfo info2 = YR.init(conf2);
        String ctx2 = info2.getContext();
        // The tenant context switches from ctx2 to ctx1
        YR.setContext(ctx1);
        //! [change context]

        //! [change context and invoke]
        ClientInfo info1 = YR.init(conf);
        String ctx1 = info1.getContext();
        InstanceHandler instanceHandler = YR.instance(SetContextExample.MyYRApp::new).invoke();
        ObjectRef ref1 = instanceHandler.function(MyYRApp::smallCall).invoke();
        String res = (String)YR.get(ref1, 100);
        System.out.println(res);

        ClientInfo info2 = YR.init(conf);
        String ctx2 = info2.getContext();
        ObjectRef ref2 = instanceHandler.function(MyYRApp::smallCall).invoke();  // Call failed
        res = (String)YR.get(ref2, 100);

        // The tenant context switches from ctx2 to ctx1
        YR.setContext(ctx1);
        ObjectRef ref3 = instanceHandler.function(MyYRApp::smallCall).invoke();  // Call successful
        res = (String)YR.get(ref3, 100);
        //! [change context and invoke]

        //! [change context and invoke with export and import]
        ClientInfo info1 = YR.init(conf);
        String ctx1 = info1.getContext();
        InstanceHandler instanceHandler = YR.instance(SetContextExample.MyYRApp::new).invoke();
        Map<String, String> out = instanceHandler.exportHandler();
        ObjectRef ref1 = instanceHandler.function(MyYRApp::smallCall).invoke();
        String res = (String)YR.get(ref1, 100);
        System.out.println(res);

        ClientInfo info2 = YR.init(conf);
        String ctx2 = info2.getContext();
        instanceHandler.importHandler(out);
        ObjectRef ref2 = instanceHandler.function(MyYRApp::smallCall).invoke();  // 调用成功
        res = (String)YR.get(ref2, 100);

        // The tenant context switches from ctx2 to ctx1
        YR.setContext(ctx1);
        ObjectRef ref3 = instanceHandler.function(MyYRApp::smallCall).invoke();  // 调用成功
        res = (String)YR.get(ref3, 100);
        //! [change context and invoke with export and import]
    }
}