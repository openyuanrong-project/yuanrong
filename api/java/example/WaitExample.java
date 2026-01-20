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

public class WaitExample {
    public static void main(String[] args) throws Exception {
        //! [wait result]
        int y = 1;
        // Get the values of multiple object refs in parallel.
        List<ObjectRef> objectRefs = new ArrayList<>();
        for (int i = 0; i < 3; i++) {
            objectRefs.add(YR.put(i));
        }
        WaitResult waitResult = YR.wait(objectRefs, /*num_returns=*/ 1, /*timeoutMs=*/ 1000);
        System.out.println(waitResult.getReady()); // List of ready objects.
        System.out.println(waitResult.getUnready()); // list of unready objects.
        //! [wait result]
    }
}