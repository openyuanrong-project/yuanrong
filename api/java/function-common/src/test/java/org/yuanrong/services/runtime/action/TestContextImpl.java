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

package org.yuanrong.services.runtime.action;

import org.yuanrong.services.logger.UserFunctionLogger;

import org.junit.Assert;
import org.junit.Test;

import java.util.HashMap;
import java.util.Map;

public class TestContextImpl {
    @Test
    public void testContextImpl() {
        ContextImpl context = new ContextImpl();
        context.setFuncMetaData(new FunctionMetaData());
        context.setResourceMetaData(new ResourceMetaData());
        context.setDelegateDecrypt(new DelegateDecrypt());
        context.getFuncMetaData().setTenantId("project1");
        Assert.assertEquals("project1", context.getProjectID());

        context.getFuncMetaData().setFuncName("functionname");
        Assert.assertEquals("functionname", context.getFunctionName());

        context.getFuncMetaData().setVersion("functionversion");
        Assert.assertEquals("functionversion", context.getVersion());

        Map map = new HashMap<String, String>(1) ;
        map.put("test", "success");
        context.getFuncMetaData().setUserData(map);
        Assert.assertEquals("success", context.getUserData("test"));

        context.getResourceMetaData().setMemory(500);
        Assert.assertEquals(500, context.getMemorySize());

        context.getResourceMetaData().setCpu(500);
        Assert.assertEquals(500, context.getCPUNumber());

        ContextInvokeParams params = new ContextInvokeParams();
        params.setRequestID("requestID");
        context.setContextInvokeParams(params);

        LogTankService logTankService = new LogTankService();
        ExtendedMetaData extendedMetaData = new ExtendedMetaData();
        extendedMetaData.setLogTankService(logTankService);
        context.setExtendedMetaData(extendedMetaData);

        Assert.assertEquals(context.getRequestID(), "requestID");
        Assert.assertEquals(context.getTraceID(), "requestID");
        Assert.assertEquals(context.getInvokeID(), "");
        Assert.assertNull(context.getAccessKey());
        Assert.assertNull(context.getSecretKey());
        Assert.assertNull(context.getToken());
        Assert.assertNull(context.getSecurityToken());
        Assert.assertNull(context.getAlias());
        Assert.assertNull(context.getInvokeProperty());
        Assert.assertNotNull(context.getLogger());
        Assert.assertNotNull(context.hashCode());
        Assert.assertNotNull(context.toString());
        Assert.assertNotNull(context.getRemainingTimeInMilliSeconds());
        Assert.assertNotNull(context.getRunningTimeInSeconds());

        context.setLogger(new UserFunctionLogger());
        Assert.assertNotNull(context.getLogger());
        context.setInstanceID("instanceID");
        Assert.assertNotNull(context.getInstanceID());
        context.setStartTime(100L);
        Assert.assertNotNull(context.getStartTime());
        Assert.assertNotNull(context.getContextInvokeParams());
        Assert.assertNotNull(context.getDelegateDecrypt());
        context.setState(new String("stateID"));
        Assert.assertNull(context.getState());
        context.setTraceID("traceID");

        ContextImpl context1 = new ContextImpl();
        context1.canEqual(context);
        Assert.assertNotEquals(context, context1);
        Assert.assertNotEquals(context1, null);

        ContextImpl context2 = new ContextImpl();
        context2.setFuncMetaData(context.getFuncMetaData());
        context2.setResourceMetaData(context.getResourceMetaData());
        context2.setDelegateDecrypt(context.getDelegateDecrypt());
        context2.setContextInvokeParams(context.getContextInvokeParams());
        context2.setLogger((UserFunctionLogger) context.getLogger());
        context2.setInstanceID(context.getInstanceID());
        context2.setExtendedMetaData(context.getExtendedMetaData());
        context2.setStartTime(context.getStartTime());
        context2.setState(context.getState());
        context2.setStream(context.getStream());
        context2.setTraceID(context.getTraceID());

        ContextImpl context3 = new ContextImpl(context);

        Assert.assertTrue(context.canEqual(context2));
        Assert.assertEquals(context, context2);
        Assert.assertNotEquals(context, context3);
    }

    @Test
    public void testInitializer() {
        Initializer initializer1 = new Initializer();
        Initializer initializer2 = new Initializer();
        initializer1.canEqual(initializer2);
        Assert.assertTrue(initializer1.equals(initializer2));
        initializer1.setInitializerHandler("init");
        Assert.assertFalse(initializer1.equals(initializer2));
        initializer1.setInitializerTimeout(100);
        initializer2.setInitializerHandler("init");
        Assert.assertFalse(initializer1.equals(initializer2));
        Assert.assertNotNull(initializer1.toString());
        Assert.assertNotNull(initializer1.hashCode());
    }

    @Test
    public void testContextInvokeParams() {
        ContextInvokeParams param = new ContextInvokeParams();
        param.setAccessKey("ak");
        param.setSecretKey("sk");
        param.setSecurityAccessKey("sak");
        param.setSecuritySecretKey("ssk");
        param.setRequestID("reqID");
        param.setInvokeID("invokeID");
        param.setToken("token");
        param.setSecurityToken("sToken");
        param.setAlias("alias");
        param.setWorkflowID("workflowID");
        param.setWorkflowRunID("runid");
        param.setWorkflowStateID("statID");
        param.setReqStreamName("reqStreamName");
        param.setRespStreamName("respStreamName");
        Assert.assertEquals(param.getAccessKey(), "ak");
        Assert.assertEquals(param.getSecretKey(), "sk");
        Assert.assertEquals(param.getSecurityAccessKey(), "sak");
        Assert.assertEquals(param.getSecuritySecretKey(), "ssk");
        Assert.assertEquals(param.getRequestID(), "reqID");
        Assert.assertEquals(param.getInvokeID(), "invokeID");
        Assert.assertEquals(param.getToken(), "token");
        Assert.assertEquals(param.getSecurityToken(), "sToken");
        Assert.assertEquals(param.getAlias(), "alias");
        Assert.assertEquals(param.getWorkflowID(), "workflowID");
        Assert.assertEquals(param.getWorkflowRunID(), "runid");
        Assert.assertEquals(param.getWorkflowStateID(), "statID");
        Assert.assertEquals(param.getReqStreamName(), "reqStreamName");
        Assert.assertEquals(param.getRespStreamName(), "respStreamName");
    }

    @Test
    public void testExtendedMetaData() {
        ExtendedMetaData data1 = new ExtendedMetaData();
        ExtendedMetaData data2 = new ExtendedMetaData();
        data1.canEqual(data2);
        Assert.assertTrue(data1.equals(data2));
        data1.setInitializer(new Initializer());
        Assert.assertFalse(data1.equals(data2));
        Assert.assertNotNull(data1.toString());
        Assert.assertNotNull(data1.hashCode());
    }
}
