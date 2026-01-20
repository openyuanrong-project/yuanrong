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

package org.yuanrong.exception.handle.traceback;

import org.yuanrong.exception.handler.traceback.StackTraceInfo;

import org.junit.Assert;
import org.junit.Test;

import java.util.ArrayList;

public class TestStackTraceInfo {
    @Test
    public void testInitStackTraceInfo() {
        StackTraceInfo stackTraceInfo = new StackTraceInfo();
        StackTraceInfo stackTraceInfo1 = new StackTraceInfo("test-type", "test-info");
        ArrayList<StackTraceElement> list = new ArrayList<>();
        StackTraceElement stackTraceElement = new StackTraceElement("test-class", "test-method", "test-file", 1);
        list.add(stackTraceElement);
        StackTraceInfo stackTraceInfo2 = new StackTraceInfo("test-type", "test-message", list);
        StackTraceInfo stackTraceInfo3 = new StackTraceInfo("type", "test-message", list, "java");

        stackTraceInfo.setType("type");
        stackTraceInfo1.setMessage("message");
        stackTraceInfo2.setLanguage("java");
        stackTraceInfo3.setStackTraceElements(list);
        stackTraceInfo.getStackTraceElements();
        stackTraceInfo1.getLanguage();
        stackTraceInfo2.getMessage();
        stackTraceInfo3.getType();
        stackTraceInfo3.toString();
        stackTraceInfo3.hashCode();

        Assert.assertFalse(stackTraceInfo.equals(stackTraceInfo1));
        Assert.assertFalse(stackTraceInfo2.equals(stackTraceInfo3));
    }
}
