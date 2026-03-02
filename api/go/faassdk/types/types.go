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

// Package types -
package types

import (
	"encoding/json"

	"yuanrong.org/kernel/runtime/libruntime/api"
	"yuanrong.org/kernel/runtime/libruntime/common/logger/config"
)

// CreateParams for RequestArgs
type CreateParams struct {
	InitEntry string `json:"userInitEntry"`
	CallEntry string `json:"userCallEntry"`
}

// HttpCreateParams is used for http function creation
type HttpCreateParams struct {
	Port      int    `json:"port"`
	InitRoute string `json:"initRoute"`
	CallRoute string `json:"callRoute"`
}

// CreateOptions context
type CreateOptions struct {
	FuncMetaDataContext     string
	ResourceMetaDataContext string
}

// FuncSpec -
type FuncSpec struct {
	FuncMetaData     FuncMetaData     `json:"funcMetaData"`
	ResourceMetaData ResourceMetaData `json:"resourceMetaData"`
	ExtendedMetaData ExtendedMetaData `json:"extendedMetaData"`
}

// FuncMetaData -
type FuncMetaData struct {
	FunctionName       string `json:"name"`
	Service            string `json:"service"`
	Runtime            string `json:"runtime"`
	TenantId           string `json:"tenantId"`
	Version            string `json:"version"`
	Timeout            int    `json:"timeout"`
	Handler            string `json:"handler"`
	FunctionVersionURN string `json:"functionVersionUrn"`
}

// ResourceMetaData -
type ResourceMetaData struct {
	Cpu                 int    `json:"cpu"`
	Memory              int    `json:"memory"`
	GpuMemory           int64  `json:"gpu_memory"`
	EnableDynamicMemory bool   `json:"enable_dynamic_memory" valid:",optional"`
	CustomResources     string `json:"customResources" valid:",optional"`
	EnableTmpExpansion  bool   `json:"enable_tmp_expansion" valid:",optional"`
	EphemeralStorage    int    `json:"ephemeral_storage" valid:"int,optional"`
}

// ExtendedMetaData -
type ExtendedMetaData struct {
	Initializer            Initializer            `json:"initializer" valid:",optional"`
	LogTankService         LogTankService         `json:"log_tank_service" valid:",optional"`
	CustomHealthCheck      CustomHealthCheck      `json:"custom_health_check" valid:",optional"`
	CustomGracefulShutdown CustomGracefulShutdown `json:"runtime_graceful_shutdown"`
}

// LogTankService -
type LogTankService struct {
	GroupID  string `json:"logGroupId" valid:",optional"`
	StreamID string `json:"logStreamId" valid:",optional"`
}

// Initializer -
type Initializer struct {
	Handler string `json:"initializer_handler" valid:",optional"`
	Timeout int    `json:"initializer_timeout" valid:",optional"`
}

// CustomHealthCheck -
type CustomHealthCheck struct {
	TimeoutSeconds   int `json:"timeoutSeconds" valid:",optional"`
	PeriodSeconds    int `json:"periodSeconds" valid:",optional"`
	FailureThreshold int `json:"failureThreshold" valid:",optional"`
}

// CustomGracefulShutdown define the option of custom container's runtime graceful shutdown
type CustomGracefulShutdown struct {
	MaxShutdownTimeout int `json:"maxShutdownTimeout"`
}

// CallRequest -
type CallRequest struct {
	Header map[string]string `json:"header"`
	Path   string            `json:"path"`
	Method string            `json:"method"`
	Query  string            `json:"query"`
	Body   json.RawMessage   `json:"body"`
	Timeout int64 `json:"timeout"`
}

// CallResponse -
type CallResponse struct {
	Headers         map[string]string `json:"headers"`
	Body            json.RawMessage   `json:"body"`
	BillingDuration string            `json:"billingDuration"`
	InnerCode       string            `json:"innerCode"`
	InvokerSummary  string            `json:"invokerSummary"`
	LogResult       string            `json:"logResult"`
	UserFuncTime    float64           `json:"userFuncTime"`
	ExecutorTime    float64           `json:"executorTime"`
}

// InitResponse -
type InitResponse struct {
	ErrorCode string          `json:"errorCode"`
	Message   json.RawMessage `json:"message"`
}

// InvokeRequest -
type InvokeRequest struct {
	FuncName       string            `json:"funcName"`
	FuncVersion    string            `json:"funcVersion"`
	Payload        string            `json:"payload"`
	TraceID        string            `json:"traceID"`
	Timeout        int64             `json:"timeout"`
	AcquireTimeout int64             `json:"acquireTimeout"`
	StateID        string            `json:"stateID"`
	Params         map[string]string `json:"params"`
	FuncUrn        string
}

// InvokeResponse -
type InvokeResponse struct {
	ObjectID     string `json:"objectID"`
	StatusCode   int    `json:"statusCode"`
	ErrorMessage string `json:"errorMessage"`
}

// GetStatusCode InvokeResponse get status code
func (i *InvokeResponse) GetStatusCode() int {
	return i.StatusCode

}

// GetErrorMessage InvokeResponse get error message
func (i *InvokeResponse) GetErrorMessage() string {
	return i.ErrorMessage

}

// ExitRequest -
type ExitRequest struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// GetFutureRequest -
type GetFutureRequest struct {
	ObjectID string `json:"objectID"`
	TraceID  string `json:"traceID"`
}

// CircuitBreakRequest -
type CircuitBreakRequest struct {
	Switch bool `json:"switch"`
}

// GetFutureResponse -
type GetFutureResponse struct {
	ObjectID     string `json:"objectID"`
	Content      string `json:"content"`
	StatusCode   int    `json:"statusCode"`
	ErrorMessage string `json:"errorMessage"`
}

// GetStatusCode GetFutureResponse get status code
func (f *GetFutureResponse) GetStatusCode() int {
	return f.StatusCode
}

// GetErrorMessage GetFutureResponse get error message
func (f *GetFutureResponse) GetErrorMessage() string {
	return f.ErrorMessage
}

// StateRequest -
type StateRequest struct {
	FuncName    string            `json:"funcName"`
	FuncVersion string            `json:"funcVersion"`
	Params      map[string]string `json:"params"`
	StateID     string            `json:"stateID"`
	TraceID     string            `json:"traceID"`
}

// StateResponse -
type StateResponse struct {
	StateID      string `json:"stateID"`
	StatusCode   int    `json:"statusCode"`
	ErrorMessage string `json:"errorMessage"`
}

// GetStatusCode StateNewResponse get status code
func (i *StateResponse) GetStatusCode() int {
	return i.StatusCode
}

// GetErrorMessage StateNewResponse get error message
func (f *StateResponse) GetErrorMessage() string {
	return f.ErrorMessage
}

// TerminateResponse -
type TerminateResponse struct {
	ObjectID     string `json:"objectID"`
	StatusCode   int    `json:"statusCode"`
	ErrorMessage string `json:"errorMessage"`
}

// GetStatusCode StateNewResponse get status code
func (i *TerminateResponse) GetStatusCode() int {
	return i.StatusCode
}

// GetErrorMessage StateNewResponse get error message
func (f *TerminateResponse) GetErrorMessage() string {
	return f.ErrorMessage
}

// ExitResponse -
type ExitResponse struct {
	StatusCode   int    `json:"statusCode"`
	ErrorMessage string `json:"errorMessage"`
}

// GetStatusCode ExitResponse get status code
func (i *ExitResponse) GetStatusCode() int {
	return i.StatusCode
}

// GetErrorMessage ExitResponse get error message
func (f *ExitResponse) GetErrorMessage() string {
	return f.ErrorMessage
}

// Response interface
type Response interface {
	GetStatusCode() int
	GetErrorMessage() string
}

// AuthConfig represents configurations of local auth
type AuthConfig struct {
	AKey     string `json:"aKey" yaml:"aKey" valid:"optional"`
	SKey     string `json:"sKey" yaml:"sKey" valid:"optional"`
	Duration int    `json:"duration" yaml:"duration" valid:"optional"`
}

// CustomUserArgs -
type CustomUserArgs struct {
	AlarmConfig       AlarmConfig     `json:"alarmConfig" valid:"optional"`
	StsServerConfig   StsServerConfig `json:"stsServerConfig"`
	ClusterName       string          `json:"clusterName"`
	DiskMonitorEnable bool            `json:"diskMonitorEnable"`
	LocalAuth         AuthConfig      `json:"localAuth"`
}

// StsServerConfig -
type StsServerConfig struct {
	Domain string `json:"domain,omitempty" validate:"max=255"`
	Path   string `json:"path,omitempty" validate:"max=255"`
}

// AlarmConfig -
type AlarmConfig struct {
	EnableAlarm         bool               `json:"enableAlarm"`
	AlarmLogConfig      config.CoreInfo    `json:"alarmLogConfig" valid:"optional"`
	XiangYunFourConfig  XiangYunFourConfig `json:"xiangYunFourConfig" valid:"optional"`
	MinInsStartInterval int                `json:"minInsStartInterval"`
	MinInsCheckInterval int                `json:"minInsCheckInterval"`
}

// XiangYunFourConfig -
type XiangYunFourConfig struct {
	Site          string `json:"site"`
	TenantID      string `json:"tenantID"`
	ApplicationID string `json:"applicationID"`
	ServiceID     string `json:"serviceID"`
}

// CredentialResponse -
type CredentialResponse struct {
	StatusCode   int    `json:"statusCode"`
	ErrorMessage string `json:"errorMessage"`
	api.Credential
}

// GetStatusCode CredentialResponse get status code
func (c *CredentialResponse) GetStatusCode() int {
	return c.StatusCode
}

// GetErrorMessage CredentialResponse get error message
func (c *CredentialResponse) GetErrorMessage() string {
	return c.ErrorMessage
}
