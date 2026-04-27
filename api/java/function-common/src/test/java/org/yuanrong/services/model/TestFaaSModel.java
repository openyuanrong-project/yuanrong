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
package org.yuanrong.services.model;

import com.google.gson.JsonObject;

import org.junit.Assert;
import org.junit.Test;

import java.util.HashMap;
import java.util.Map;

public class TestFaaSModel {
    @Test
    public void testCallRequest() {
        CallRequest callRequest = new CallRequest();
        callRequest.setHeader(new HashMap<String, String>(){
            {
                put("key1", "val1");
            }
        });
        callRequest.setBody("body1");
        Assert.assertEquals(callRequest.getHeader().get("key1"), "val1");
        Assert.assertEquals(callRequest.getBody(), "body1");
    }

    @Test
    public void testCallResponseJsonObject() {
        CallResponseJsonObject callResponse = new CallResponseJsonObject();
        JsonObject jsonObject = new JsonObject();
        jsonObject.addProperty("name", "faas");
        callResponse.setBody(jsonObject);
        callResponse.setInnerCode("200");
        callResponse.setBillingDuration("x-bill-duration");
        callResponse.setLogResult("x-log-result");
        callResponse.setInvokerSummary("x-invoke-summary");
        Assert.assertEquals(callResponse.getInnerCode(), "200");
        Assert.assertEquals(callResponse.getBillingDuration(), "x-bill-duration");
        Assert.assertEquals(callResponse.getLogResult(), "x-log-result");
        Assert.assertEquals(callResponse.getInvokerSummary(), "x-invoke-summary");
        Assert.assertEquals(callResponse.getBody().get("name").getAsString(), "faas");
    }
}
