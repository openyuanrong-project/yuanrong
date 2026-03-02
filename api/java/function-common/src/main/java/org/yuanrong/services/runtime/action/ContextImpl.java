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
import org.yuanrong.services.runtime.Context;
import org.yuanrong.services.runtime.RuntimeLogger;

import com.google.gson.annotations.Expose;

import lombok.Data;

import java.util.Map;

/**
 * The type Context.
 *
 * @since 2024-07-20
 */
@Data
public class ContextImpl implements Context {
    /**
     * function logger
     */
    @Expose(serialize = false, deserialize = false)
    private UserFunctionLogger logger = new UserFunctionLogger();

    /**
     * function meta data
     */
    private FunctionMetaData funcMetaData;

    /**
     * function delegate decrypt info
     */
    private DelegateDecrypt delegateDecrypt;

    /**
     * function resource meta data
     */
    private ResourceMetaData resourceMetaData;

    /**
     * function extended meta data
     */
    private ExtendedMetaData extendedMetaData;

    /**
     * context invoke params
     */
    private ContextInvokeParams contextInvokeParams = new ContextInvokeParams();

    /**
     *  The Instance id
     */
    private String instanceID = "";

    /**
     *  The Instance Label
     */
    private String instanceLabel = "";

    /**
     * The Start time.
     */
    private long startTime = System.currentTimeMillis();

    /**
     * Stream event handler.
     */
    private Stream stream = new Stream();

    public ContextImpl() {}

    public ContextImpl(ContextImpl contextImpl) {
        if (contextImpl != null) {
            this.funcMetaData = contextImpl.getFuncMetaData();
            this.delegateDecrypt = contextImpl.getDelegateDecrypt();
            this.resourceMetaData = contextImpl.getResourceMetaData();
            this.extendedMetaData = contextImpl.getExtendedMetaData();
            this.contextInvokeParams = new ContextInvokeParams(contextImpl.getContextInvokeParams());
            this.instanceID = contextImpl.getInstanceID();
            this.instanceLabel = contextImpl.getInstanceLabel();
            this.startTime = contextImpl.getStartTime();
        }
    }

    /**
     * Set context from invoke parameters
     *
     * @param contextInvokeParams context from invoke parameters
     */
    public void setContextInvokeParams(ContextInvokeParams contextInvokeParams) {
        this.contextInvokeParams = contextInvokeParams;
    }

    /**
     * Get the extended meta data
     *
     * @return extended meta data
     */
    public ExtendedMetaData getExtendedMetaData() {
        return this.extendedMetaData;
    }

    @Override
    public String getRequestID() {
        return this.contextInvokeParams.getRequestID();
    }

    @Override
    public int getRemainingTimeInMilliSeconds() {
        return (int) (this.funcMetaData.getTimeout() * 1000L - System.currentTimeMillis() + this.startTime);
    }

    @Override
    public String getAccessKey() {
        return this.delegateDecrypt.getAccessKey();
    }

    @Override
    public String getSecretKey() {
        return this.delegateDecrypt.getSecretKey();
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
    public String getUserData(String key) {
        return this.funcMetaData.getUserData().get(key);
    }

    @Override
    public String getFunctionName() {
        return this.funcMetaData.getFuncName();
    }

    /**
     * getRunningTimeInSeconds is second in java, millSecond in other language
     *
     * @return the int
     */
    @Override
    public int getRunningTimeInSeconds() {
        return this.funcMetaData.getTimeout();
    }

    @Override
    public String getVersion() {
        return this.funcMetaData.getVersion();
    }

    @Override
    public int getMemorySize() {
        return this.resourceMetaData.getMemory();
    }

    @Override
    public int getCPUNumber() {
        return this.resourceMetaData.getCpu();
    }

    @Override
    public String getProjectID() {
        return this.funcMetaData.getTenantId();
    }

    @Override
    public String getPackage() {
        return this.funcMetaData.getService();
    }

    @Override
    public String getToken() {
        return this.delegateDecrypt.getAuthToken();
    }

    @Override
    public String getSecurityToken() {
        return this.delegateDecrypt.getSecurityToken();
    }

    @Override
    public String getAlias() {
        return this.funcMetaData.getAlias();
    }

    @Override
    public RuntimeLogger getLogger() {
        return logger;
    }

    @Override
    public Object getState() {
        return null;
    }

    @Override
    public void setState(Object state) {}

    @Override
    public String getInstanceID() {
        return this.instanceID;
    }

    @Override
    public String getInstanceLabel() {
        return this.instanceLabel;
    }

    public void setInstanceLabel(String instanceLabel) {
        this.instanceLabel = instanceLabel;
    }

    @Override
    public String getInvokeProperty() {
        return null;
    }

    @Override
    public String getTraceID() {
        return this.contextInvokeParams.getRequestID();
    }

    @Override
    public void setTraceID(String requestID) {
        this.contextInvokeParams.setRequestID(requestID);
    }

    @Override
    public String getInvokeID() {
        return this.contextInvokeParams.getInvokeID();
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

    @Override
    public Stream getStream() {
        stream.setRequestId(this.getInvokeID());
        stream.setInstanceId(this.getInstanceID());
        return this.stream;
    }
}
