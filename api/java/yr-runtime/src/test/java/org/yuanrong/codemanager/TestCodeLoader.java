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

package org.yuanrong.codemanager;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertTrue;

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ErrorInfo;

import org.junit.Assert;
import org.junit.Test;

import java.util.ArrayList;

public class TestCodeLoader {
    @Test
    public void testCodeLoader() {
        final CodeLoader codeLoader = new CodeLoader();
        final ArrayList<String> stringArrayList = new ArrayList<>();
        stringArrayList.add("test");
        boolean isException = false;
        try {
            ErrorInfo err = CodeLoader.Load(stringArrayList);
            assertEquals(err.getErrorCode(), ErrorCode.ERR_PARAM_INVALID);
            assertTrue(err.getErrorMessage().contains("failed to load code in specified path"));
        } catch (Throwable e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }
}
