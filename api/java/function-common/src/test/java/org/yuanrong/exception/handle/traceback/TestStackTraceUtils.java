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

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.exception.YRException;
import org.yuanrong.exception.handler.traceback.StackTraceInfo;
import org.yuanrong.exception.handler.traceback.StackTraceUtils;

import org.junit.Assert;
import org.junit.Test;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

public class TestStackTraceUtils {
    @Test
    public void testCheckErrorAndThrowForInvokeException() {
        YRException yrException = new YRException(ErrorCode.ERR_ETCD_OPERATION_ERROR,
            ModuleCode.CORE, "test-exception");
        StackTraceElement[] stackTrace = yrException.getStackTrace();
        ArrayList<StackTraceElement> list = new ArrayList<>(Arrays.asList(stackTrace));
        StackTraceInfo stackTraceInfo = new StackTraceInfo(yrException.getClass().getName(),
            yrException.getMessage(), list);
        ArrayList<StackTraceInfo> stackTraceInfoArrayList = new ArrayList<>();
        stackTraceInfoArrayList.add(stackTraceInfo);

        boolean isException = false;
        try {
            StackTraceUtils.checkErrorAndThrowForInvokeException(
                new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.CORE, "msg"), "test");
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        try {
            StackTraceUtils.checkErrorAndThrowForInvokeException(
                new ErrorInfo(ErrorCode.ERR_EXTENSION_META_ERROR, ModuleCode.CORE, "msg"), "test");
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            StackTraceUtils.checkErrorAndThrowForInvokeException(
                new ErrorInfo(ErrorCode.ERR_USER_FUNCTION_EXCEPTION, ModuleCode.CORE, "errorInfo-empty-list",
                    new ArrayList<>()), "test");
        } catch (YRException e) {
            Assert.assertTrue(e.getMessage().contains("errorInfo-empty-list"));
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            StackTraceUtils.checkErrorAndThrowForInvokeException(
                new ErrorInfo(ErrorCode.ERR_USER_FUNCTION_EXCEPTION, ModuleCode.RUNTIME, "test",
                    stackTraceInfoArrayList), "test");
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }

    @Test
    public void testProcessValidAndDuplicate() {
        YRException yrException = new YRException(ErrorCode.ERR_ETCD_OPERATION_ERROR,
            ModuleCode.CORE, "test-exception");
        StackTraceElement[] stackTrace = yrException.getStackTrace();
        ArrayList<StackTraceElement> list = new ArrayList<>(Arrays.asList(stackTrace));
        List<StackTraceElement> stackTraceElements = StackTraceUtils.processValidAndDuplicate(list);
        List<StackTraceElement> traceElements = StackTraceUtils.processValidAndDuplicate(new ArrayList<>());
        Assert.assertNotEquals(stackTraceElements, traceElements);
    }

    @Test
    public void testGetListTraceElementsWithEmptyValue() {
        List<StackTraceElement> elementList = StackTraceUtils.getListTraceElements("");
        Assert.assertTrue(elementList.isEmpty());
    }

    @Test
    public void testGetListTraceElementsWithNoBeginTagValue() {
        String stackTrace = "java.lang.RuntimeException: An error occurred\n"
            + "\tat com.ClassA.methodA(ClassA.java:10)\n" + "\tat com.ClassC.methodC(ClassC.java:20)\n";
        List<StackTraceElement> elementList = StackTraceUtils.getListTraceElements(stackTrace);
        Assert.assertFalse(elementList.isEmpty());
    }

    @Test
    public void testGetListTraceElementsWithBeginTagValue() {
        String stackTrace = "java.lang.RuntimeException: An error occurred\n"
            + "\tat com.ClassA.methodA(ClassA.java:10)\n"
            + "\tCaused by: java.lang.NoSuchMethodException: com.exception\n"
            + "\tat com.ClassE.methodE(Class.java:70)";
        List<StackTraceElement> elementList = StackTraceUtils.getListTraceElements(stackTrace);
        Assert.assertFalse(elementList.isEmpty());
    }

    @Test
    public void testGetListTraceElementsWithRuntimeValue() {
        String stackTrace = "java.lang.RuntimeException\n"
            + "\tat org.yuanrong.codemanager.CodeExecutor.Execute\n"
            + "\tat com.example.ClassB.methodB(ClassB.java:20)\n" + "\tat com.example.ClassC.methodC(ClassC.java:30)\n"
            + "\tat com.example.ClassD.methodD(ClassD.java:40)\n" + "\tat com.example.ClassE.methodE(ClassE.java:50)\n"
            + "\tat com.example.ClassE.methodF(ClassE.java:50)";
        List<StackTraceElement> elementList = StackTraceUtils.getListTraceElements(stackTrace);
        Assert.assertTrue(elementList.isEmpty());
    }

    @Test
    public void testStringToStackTraceEle() {
        boolean isException = false;
        try {
            StackTraceUtils.stringToStackTraceEle("");
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }
}
