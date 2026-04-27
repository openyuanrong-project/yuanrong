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
import org.yuanrong.Group;
import org.yuanrong.GroupOptions;
import org.yuanrong.call.InstanceHandler;

public class GroupExample {
    public static void main(String[] args) throws Exception {
        //! [Group invoke example]
        Group g = new Group();
        g.setGroupName("groupName");
        GroupOptions gOpts = new GroupOptions();
        gOpts.setTimeout(60);
        g.setGroupOpts(gOpts);
        InvokeOptions opts = new InvokeOptions();
        opts.setGroupName("groupName");
        InstanceHandler res1 = YR.instance(MyClass::new).options(opts).invoke();
        InstanceHandler res2 = YR.instance(MyClass::new).options(opts).invoke();
        g.invoke();
        //! [Group invoke example]

        //! [Group terminate example]
        Group g = new Group();
        g.setGroupName("groupName");
        InvokeOptions opts = new InvokeOptions();
        opts.setGroupName("groupName");
        InstanceHandler res1 = YR.instance(MyClass::new).options(opts).invoke();
        InstanceHandler res2 = YR.instance(MyClass::new).options(opts).invoke();
        g.invoke();
        g.terminate();
        //! [Group terminate example]
    }
}