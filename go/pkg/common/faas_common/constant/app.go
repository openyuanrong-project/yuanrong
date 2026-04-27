/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
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

// Package constant -
package constant

// used for app-job
const (
	// UserMetadataKey key used for the app createOpts
	UserMetadataKey = "USER_PROVIDED_METADATA"
	// EntryPointKey entrypoint for starting app
	EntryPointKey = "ENTRYPOINT"
	// AppFN -
	FunctionNameApp = "app"
	// AppFuncId -
	AppFuncId = "default/0-system-faasExecutorPosixCustom/$latest"
	// AppType type for invoking create-app
	AppType = "SUBMISSION"
	// AppStatusPending -
	AppStatusPending = "PENDING"
	// AppStatusRunning -
	AppStatusRunning = "RUNNING"
	// AppStatusSucceeded -
	AppStatusSucceeded = "SUCCEEDED"
	// AppStatusFailed -
	AppStatusFailed = "FAILED"
	// AppStatusStopped -
	AppStatusStopped = "STOPPED"

	// AppInvokeTimeout 30min
	AppInvokeTimeout = 1800
)

// AppInfo - Ray job JobDetails
type AppInfo struct {
	Key string `json:"key"`
	// Enum: "SUBMISSION" "DRIVER"
	Type         string     `json:"type"`
	Entrypoint   string     `json:"entrypoint"`
	SubmissionID string     `json:"submission_id"`
	DriverInfo   DriverInfo `json:"driver_info"  valid:",optional"`
	// Status Enum: "PENDING" "RUNNING" "STOPPED" "SUCCEEDED" "FAILED"
	Status                 string                 `json:"status"  valid:",optional"`
	StartTime              string                 `json:"start_time"  valid:",optional"`
	EndTime                string                 `json:"end_time"  valid:",optional"`
	Metadata               map[string]string      `json:"metadata"  valid:",optional"`
	RuntimeEnv             map[string]interface{} `json:"runtime_env"  valid:",optional"`
	DriverAgentHttpAddress string                 `json:"driver_agent_http_address"  valid:",optional"`
	DriverNodeID           string                 `json:"driver_node_id"  valid:",optional"`
	DriverExitCode         int32                  `json:"driver_exit_code"  valid:",optional"`
	ErrorType              string                 `json:"error_type"  valid:",optional"`
}

// DriverInfo -
type DriverInfo struct {
	ID            string `json:"id"`
	NodeIPAddress string `json:"node_ip_address"`
	PID           string `json:"pid"`
}
