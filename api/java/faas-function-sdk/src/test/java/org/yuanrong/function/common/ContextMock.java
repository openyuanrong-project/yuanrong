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

import org.yuanrong.services.runtime.Context;
import org.yuanrong.services.runtime.RuntimeLogger;
import org.yuanrong.services.runtime.action.ExtendedMetaData;
import org.yuanrong.services.runtime.action.Stream;
import org.yuanrong.services.session.SessionService;

import java.util.Map;

/*
 * when we run unit test of Mock, it would be triggered
 */
public class ContextMock implements Context {
    private static final String INVOKE_ID = "invokeID";
    private static final String INVOKE_STATE = "invokeState";
    private static String TRACE_ID = "traceID";
    private static String INVOKE_PROPERTY = "invokeProperty";
    private Object state;

    @Override
    public String getRequestID() {
        return null;
    }

    @Override
    public int getRemainingTimeInMilliSeconds() {
        return 0;
    }

    @Override
    public String getAccessKey() {
        return null;
    }

    @Override
    public String getSecretKey() {
        return null;
    }

    @Override
    public String getSecurityAccessKey() {
        return null;
    }

    @Override
    public String getSecuritySecretKey() {
        return null;
    }

    @Override
    public String getUserData(String s) {
        return null;
    }

    @Override
    public String getFunctionName() {
        return null;
    }

    @Override
    public int getRunningTimeInSeconds() {
        return 0;
    }

    @Override
    public String getVersion() {
        return null;
    }

    @Override
    public int getMemorySize() {
        return 0;
    }

    @Override
    public int getCPUNumber() {
        return 0;
    }

    @Override
    public String getProjectID() {
        return null;
    }

    @Override
    public String getPackage() {
        return null;
    }

    @Override
    public String getToken() {
        return null;
    }

    @Override
    public String getAlias() {
        return null;
    }

    @Override
    public String getSecurityToken() {
        return null;
    }

    @Override
    public RuntimeLogger getLogger() {
        return null;
    }

    @Override
    public Object getState() {
        return state;
    }

    @Override
    public void setState(Object state) {
        this.state = state;
    }

    @Override
    public String getInstanceID() {
        return "001";
    }

    @Override
    public String getInstanceLabel() {
        return "";
    }

    public void setInstanceLabel(String instanceLabel) {
    }

    @Override
    public String getInvokeProperty() {
        return INVOKE_PROPERTY;
    }

    public void setInvokeProperty(String invokeProperty) {
        INVOKE_PROPERTY = invokeProperty;
    }

    @Override
    public String getTraceID() {
        return TRACE_ID;
    }

    public void setTraceID(String traceID) {
        TRACE_ID = traceID;
    }

    @Override
    public String getInvokeID() {
        return INVOKE_ID;
    }

    @Override
    public String getWorkflowID() {
        return null;
    }

    @Override
    public String getWorkflowRunID() {
        return null;
    }

    @Override
    public String getWorkflowStateID() {
        return null;
    }

    @Override
    public String getReqStreamName() {
        return null;
    }

    @Override
    public String getRespStreamName() {
        return null;
    }

    @Override
    public String getFrontendResponseStreamName() {
        return null;
    }

    @Override
    public String getIAMToken() {
        return "";
    }

    @Override
    public Map<String, String> getExtraMap() {
        return null;
    }

    @Override
    public void setExtraMap(Map<String, String> extraMap) {
    }

    public static class State {

        private String key = "invokeProperty";

        public State(String key) {
            this.key = key;
        }

        public State() {

        }

        public String getKey() {
            return key;
        }

        public void setKey(String key) {
            this.key = key;
        }
    }

    @Override
    public Stream getStream() {
        return null;
    }

    @Override
    public String getSessionId() {
        return "";
    }

    @Override
    public SessionService getSessionService() {
        return null;
    }
}