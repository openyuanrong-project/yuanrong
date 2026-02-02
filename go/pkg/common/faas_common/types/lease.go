/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
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

package types

import (
	"fmt"

	"yuanrong.org/kernel/runtime/libruntime/api"
)

// AcquireOption holds the options for acquireInstance
type AcquireOption struct {
	DesignateInstanceID string
	SchedulerFuncKey    string
	SchedulerID         string
	SchedulerName       string
	SchedulerAddress    string
	RequestID           string
	TraceID             string
	FuncSig             string
	ResourceSpecs       map[string]int64
	PoolLabel           string
	InvokeTag           map[string]string
	InstanceLabel       string
	Timeout             int64
	TrafficLimited      bool
	InstanceSession     *InstanceSessionConfig
}

// InstanceAllocationSucceedInfo is the response returned by faas scheduler's CallHandler
type InstanceAllocationSucceedInfo struct {
	FuncKey    string `json:"funcKey"`
	FuncSig    string `json:"funcSig"`
	InstanceID string `json:"instanceID"`
	ThreadID   string `json:"threadID"`
}

// InstanceAllocationFailedInfo contains err info for allocation failed info
type InstanceAllocationFailedInfo struct {
	ErrorCode    int    `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

// InstanceAllocationInfo contains instance router info and lease returned to function accessor
type InstanceAllocationInfo struct {
	FuncKey       string `json:"funcKey"`
	FuncSig       string `json:"funcSig"`
	InstanceID    string `json:"instanceID"`
	ThreadID      string `json:"threadID"`
	InstanceIP    string `json:"instanceIP"`
	InstancePort  string `json:"instancePort"`
	NodeIP        string `json:"nodeIP"`
	NodePort      string `json:"nodePort"`
	LeaseInterval int64  `json:"leaseInterval"`
	CPU           int64  `json:"cpu"`
	Memory        int64  `json:"memory"`
	ForceInvoke   bool   `json:"forceInvoke"`
}

// ExtraParams for interface CreateInstance
type ExtraParams struct {
	DesignatedInstanceID string
	Label                []string
	Resources            map[string]float64
	CustomResources      map[string]float64
	CreateOpt            map[string]string
	CustomExtensions     map[string]string
	ScheduleAffinities   []api.Affinity
}

// InstanceResponse is the response returned by faas scheduler's CallHandler
type InstanceResponse struct {
	InstanceAllocationInfo
	ErrorCode     int     `json:"errorCode"`
	ErrorMessage  string  `json:"errorMessage"`
	SchedulerTime float64 `json:"schedulerTime"`
}

// BatchInstanceResponse is the batch response returned by faas scheduler's CallHandler
type BatchInstanceResponse struct {
	InstanceAllocSucceed map[string]InstanceAllocationSucceedInfo `json:"instanceAllocSucceed"`
	InstanceAllocFailed  map[string]InstanceAllocationFailedInfo  `json:"instanceAllocFailed"`
	LeaseInterval        int64                                    `json:"leaseInterval"`
	SchedulerTime        float64                                  `json:"schedulerTime"`
}

// LeaseEvent -
type LeaseEvent struct {
	Type           string `json:"type"`
	RemoteClientID string `json:"remoteClientId"`
	Timestamp      int64  `json:"timestamp"`
	TraceID        string `json:"traceId"`
}

// InstanceSessionConfig -
type InstanceSessionConfig struct {
	SessionID   string `json:"sessionID"`
	SessionTTL  int    `json:"sessionTTL"`
	Concurrency int    `json:"concurrency"`
}

// ToString -
func (session *InstanceSessionConfig) ToString() string {
	return fmt.Sprintf("sessionID: %s, sessionTTL: %d, concurrency: %d",
		session.SessionID, session.SessionTTL, session.Concurrency)
}
