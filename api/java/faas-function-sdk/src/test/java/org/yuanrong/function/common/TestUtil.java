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

package org.yuanrong.function.common;

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.function.runtime.exception.InvokeException;
import org.yuanrong.services.runtime.action.ContextImpl;
import org.yuanrong.services.runtime.action.FunctionMetaData;

import com.google.gson.Gson;

import org.junit.Assert;
import org.junit.Test;

public class TestUtil {
    private static final Gson GSON = new Gson();

    @Test
    public void testCheckFuncName() {
        // func name is empty
        try {
            Util.checkFuncName(null);
        } catch (InvokeException e) {
            Assert.assertEquals(RspErrorCode.INVALID_PARAMETER.getErrorCode(), e.getErrorCode());
            Assert.assertEquals("invalid funcName, expect not null", e.getMessage());
        }

        // funcName contains multiple ':'
        try {
            Util.checkFuncName("name1:name2:name3");
        } catch (InvokeException e) {
            Assert.assertEquals(RspErrorCode.INVALID_PARAMETER.getErrorCode(), e.getErrorCode());
            Assert.assertEquals("invalid funcName, not match regular expression", e.getMessage());
        }

        // funcName not match regular
        try {
            Util.checkFuncName("+++:name2");
        } catch (InvokeException e) {
            Assert.assertEquals(RspErrorCode.INVALID_PARAMETER.getErrorCode(), e.getErrorCode());
            Assert.assertEquals("invalid funcName, not match regular expression", e.getMessage());
        }

        // funcName contains ':' and version not regular
        try {
            Util.checkFuncName("funcName:!v1");
        } catch (InvokeException e) {
            Assert.assertEquals(RspErrorCode.INVALID_PARAMETER.getErrorCode(), e.getErrorCode());
            Assert.assertEquals("invalid funcName, not match regular expression", e.getMessage());
        }

        // funcName contains ':' and version not regular
        try {
            Util.checkFuncName("funcName:v1@2");
        } catch (InvokeException ie) {
            Assert.assertEquals(RspErrorCode.INVALID_PARAMETER.getErrorCode(), ie.getErrorCode());
            Assert.assertEquals("invalid funcName, not match regular expression", ie.getMessage());
        }

        // funcName contains ':' and version not regular
        try {
            Util.checkFuncName("funcName");
        } catch (InvokeException e) {
            Assert.assertEquals(RspErrorCode.INVALID_PARAMETER.getErrorCode(), e.getErrorCode());
            Assert.assertEquals("invalid funcName, not match regular expression", e.getMessage());
        }

        // funcName not contains ':' and version not regular
        try {
            Util.checkFuncName("funcName@");
        } catch (InvokeException e) {
            Assert.assertEquals(RspErrorCode.INVALID_PARAMETER.getErrorCode(), e.getErrorCode());
            Assert.assertEquals("invalid funcName, not match regular expression", e.getMessage());
        }

        // funcName check right
        String funcName = "funcName";
        String version = "v1";
        String[] strings = Util.checkFuncName(funcName + ":" + version);
        Assert.assertEquals(funcName, strings[0]);
        Assert.assertEquals(version, strings[1]);
    }

    @Test
    public void testCheckPayload() {
        try {
            Util.checkPayload("");
        } catch (InvokeException e) {
            Assert.assertEquals(RspErrorCode.INVALID_PARAMETER.getErrorCode(), e.getErrorCode());
            Assert.assertEquals("invalid payload, expect not null", e.getMessage());
        }

        try {
            Util.checkPayload("abc");
        } catch (InvokeException e) {
            Assert.assertEquals(RspErrorCode.INVALID_PARAMETER.getErrorCode(), e.getErrorCode());
            Assert.assertEquals("invalid payload, invalid json string", e.getMessage());
        }
    }

    @Test
    public void testInitUtil() {
        Util util = new Util();
        ContextImpl context = new ContextImpl();
        FunctionMetaData functionMetaData = new FunctionMetaData();
        context.setFuncMetaData(functionMetaData);
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_ETCD_OPERATION_ERROR, ModuleCode.CORE, "test");
        boolean isException = false;
        try {
            Util.checkErrorAndThrow(errorInfo, "test");
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }

    @Test
    public void testRspErrorCode() {
        Assert.assertNotNull(RspErrorCode.INVALID_PARAMETER);
    }
}
