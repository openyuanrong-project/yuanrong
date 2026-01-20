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

public class InstanceFunctionExample {
    public static void main(String[] args) throws Exception {
        //! [instanceFunction invoke example]
        InvokeOptions invokeOptions = new InvokeOptions();
        invokeOptions.addCustomExtensions("app_name", "myApp");
        InstanceHandler instanceHandler = YR.instance(Counter::new).invoke(1);
        InstanceFunctionHandler insFuncHandler = instanceHandler.function(Counter::Add);
        ObjectRef ref = insFuncHandler.options(invokeOptions).invoke(5);
        int res = (int)YR.get(ref, 100);
        //! [instanceFunction invoke example]
    }
}