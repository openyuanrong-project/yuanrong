/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
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

package org.yuanrong.storage;

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.ModuleCode;

import org.junit.Assert;
import org.junit.Test;

import java.util.ArrayList;
import java.util.HashMap;

public class TestInternalWaitResult {
    @Test
    public void testInitResult() {
        InternalWaitResult internalWaitResult = new InternalWaitResult(new ArrayList<>(), new ArrayList<>(),
            new HashMap<>());
        InternalWaitResult internalWaitResult1 = new InternalWaitResult(new ArrayList<>(), new ArrayList<>(),
            new HashMap<>());
        HashMap<String, ErrorInfo> errorInfoHashMap = new HashMap<>();
        errorInfoHashMap.put("test",
            new ErrorInfo(ErrorCode.ERR_EXTENSION_META_ERROR, ModuleCode.RUNTIME_INVOKE, "test"));
        internalWaitResult.setExceptionIds(errorInfoHashMap);
        ArrayList<String> stringArrayList = new ArrayList<>();
        stringArrayList.add("test-ready");
        internalWaitResult.setReadyIds(stringArrayList);
        ArrayList<String> stringArrayList1 = new ArrayList<>();
        stringArrayList1.add("test-unready");
        internalWaitResult.setUnreadyIds(stringArrayList1);
        internalWaitResult1.getExceptionIds();
        internalWaitResult1.getUnreadyIds();
        internalWaitResult1.getReadyIds();
        internalWaitResult.hashCode();
        internalWaitResult1.hashCode();
        internalWaitResult1.toString();
        internalWaitResult.canEqual(internalWaitResult1);
        Assert.assertFalse(internalWaitResult.equals(internalWaitResult1));
    }

}
