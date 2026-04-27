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

import lombok.Getter;
import lombok.Setter;

/**
 * Context from Invoke Parameters
 *
 * @since 2024/8/22
 */
@Setter
@Getter
public class ContextInvokeParams {
    private String accessKey = "";
    private String secretKey = "";
    private String securityAccessKey = "";
    private String securitySecretKey = "";
    private String requestID = "";
    private String invokeID = "";
    private String token = "";
    private String securityToken = "";
    private String alias = "";
    private String workflowID = "";
    private String workflowRunID = "";
    private String workflowStateID = "";
    private String reqStreamName = "";
    private String respStreamName = "";

    /**
     * construct method
     */
    public ContextInvokeParams() {
    }

    public ContextInvokeParams(ContextInvokeParams contextInvokeParams) {
        this.accessKey = contextInvokeParams.getAccessKey();
        this.secretKey = contextInvokeParams.getSecretKey();
        this.securityAccessKey = contextInvokeParams.getSecurityAccessKey();
        this.securitySecretKey = contextInvokeParams.getSecuritySecretKey();
        this.requestID = contextInvokeParams.getRequestID();
        this.invokeID = contextInvokeParams.getInvokeID();
        this.token = contextInvokeParams.getToken();
        this.securityToken = contextInvokeParams.getSecurityToken();
        this.alias = contextInvokeParams.getAlias();
        this.workflowID = contextInvokeParams.getWorkflowID();
        this.workflowRunID = contextInvokeParams.getWorkflowRunID();
        this.workflowStateID = contextInvokeParams.getWorkflowStateID();
        this.reqStreamName = contextInvokeParams.getReqStreamName();
        this.respStreamName = contextInvokeParams.getRespStreamName();
    }
}
