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

package org.yuanrong.services.runtime;

import org.yuanrong.services.runtime.action.Stream;

import java.util.Map;

/**
 * The context object allows you to access useful information available within
 * the function stage runtime execution environment
 *
 * @since 2024/7/14
 */
public interface Context {
    /**
     * Get the request ID associated with the request.
     * This is the same ID returned to the client that called invoke(). This ID
     * is reused for retries on the same request.
     *
     * @return request ID
     */
    String getRequestID();

    /**
     * Get the time remaining for this execution in milliseconds
     *
     * @return remaining time in milliseconds
     */
    int getRemainingTimeInMilliSeconds();

    /**
     * Get the Access key information of the tenant
     *
     * @return access key
     */
    String getAccessKey();

    /**
     * Get the Secret Key information of the tenant
     *
     * @return secret key
     */
    String getSecretKey();

    /**
     * Get the Security Access key information of the tenant
     *
     * @return access key
     */
    String getSecurityAccessKey();

    /**
     * Get the Security Secret Key information of the tenant
     *
     * @return secret key
     */
    String getSecuritySecretKey();

    /**
     * Get the user data, which saved in a map
     *
     * @param key user key
     * @return user data
     */
    String getUserData(String key);

    /**
     * Get the name of the function
     *
     * @return function name
     */
    String getFunctionName();

    /**
     * Get the time distributed to the running of the function, when exceed the
     * specified time, the running of the function would be stopped by force
     *
     * @return running time in seconds
     */
    int getRunningTimeInSeconds();

    /**
     * Get version of the running function
     *
     * @return version
     */
    String getVersion();

    /**
     * Get the memory size distributed the running function
     *
     * @return memory size
     */
    int getMemorySize();

    /**
     * Get the number of cpu distributed to the running function the cpu number
     * scale by millicores, one cpu cores equals 1000 millicores. In function
     * stage runtime, every function have base of 200 millicores,and increased
     * by memory size distributed to function. the offset is about Memory
     * Size(M)/128 *100.
     *
     * @return cpu number
     */
    int getCPUNumber();

    /**
     * Get project ID of the tenant
     *
     * @return project ID
     */
    String getProjectID();

    /**
     * Get package of function owner
     *
     * @return package
     */
    String getPackage();

    /**
     * Get token of the tenant
     *
     * @return token
     */
    String getToken();

    /**
     * Get security token. In order to invoke interface of other service,
     * AK,SK and security token should been provided.
     *
     * @return security token
     */
    String getSecurityToken();

    /**
     * Get function alias.
     *
     * @return function alias
     */
    String getAlias();

    /**
     * Gets the logger for user to log out in standard output, The Logger
     * interface must be provided in SDK
     *
     * @return runtime logger
     */
    RuntimeLogger getLogger();

    /**
     * Gets the persistent state
     *
     * @return the persistent state
     */
    Object getState();

    /**
     * Sets the persistent state
     *
     * @param state for the persistent state
     */
    void setState(Object state);

    /**
     * Gets the function instance id
     *
     * @return function instance id
     */
    String getInstanceID();

    /**
     * Gets the function instance label
     *
     * @return function instance label
     */
    default String getInstanceLabel() {
        return "";
    }

    /**
     * Gets the property in ondemond message
     *
     * @return function invoke property
     */
    String getInvokeProperty();

    /**
     * Gets the traceID
     *
     * @return traceID
     */
    String getTraceID();

    /**
     * Set the traceID
     *
     * @param traceID for the user traceId
     */
    void setTraceID(String traceID);

    /**
     * Gets the invoke id
     *
     * @return invokeID
     */
    String getInvokeID();

    /**
     * Gets the workflow id
     *
     * @return workflow id
     */
    String getWorkflowID();

    /**
     * Gets the workflow run id
     *
     * @return workflow run id
     */
    String getWorkflowRunID();

    /**
     * Gets the workflow state id
     *
     * @return workflow state id
     */
    String getWorkflowStateID();

    /**
     * Get the request streamName
     *
     * @return request streamName
     */
    String getReqStreamName();

    /**
     * Get the response streamName
     *
     * @return response streamName
     */
    String getRespStreamName();

    /**
     * Get frontend response streamName
     *
     * @return frontend response streamName
     */
    String getFrontendResponseStreamName();

    /**
     * Get iam token
     *
     * @return iam token
     */
    String getIAMToken();

    /**
     * Get extra map
     *
     * @return extra map
     */
    Map<String, String> getExtraMap();

    /**
     * Set extra map
     *
     * @param extraMap for the user custom param
     */
    void setExtraMap(Map<String, String> extraMap);

    /**
     * Get stream handler of context.
     *
     * @return stream handler of context.
     */
    Stream getStream();

    /**
     * Get the agent session ID associated with the current invocation.
     *
     * <p>Returns an empty string when the request does not carry a session ID
     * or {@code use_agent_session=false}.</p>
     *
     * @return session ID, never null
     */
    String getSessionId();

    /**
     * Get the session service for the current invocation.
     *
     * <p>Returns {@code null} when no session is associated with this request.</p>
     *
     * @return {@link org.yuanrong.services.session.SessionService}, or {@code null}
     */
    org.yuanrong.services.session.SessionService getSessionService();
}
