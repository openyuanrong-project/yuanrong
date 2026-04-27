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

// Package rpcclient -
package rpcclient

import (
	"time"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/common"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	"yuanrong.org/kernel/pkg/common/faas_common/statuscode"
)

const (
	defaultTimeout = 900 * time.Second
)

var (
	// ErrKernelClientTimeout -
	ErrKernelClientTimeout = snerror.New(statuscode.InternalErrorCode, "kernel rpcclient timeout")
)

// TransportParams -
type TransportParams struct {
	Timeout       time.Duration
	RetryInterval time.Duration
	RetryNumber   int
}

// AffinityType -
type AffinityType int32

// SchedulingOptions -
type SchedulingOptions struct {
	Priority         int32
	Resources        map[string]float64
	Extension        map[string]string
	Affinity         map[string]AffinityType
	ScheduleAffinity []byte
}

// CreateParams -
type CreateParams struct {
	TransportParams
	DesignatedInstanceID string
	Label                []string
	CreateOption         map[string]string
	ScheduleOption       SchedulingOptions
}

// InvokeParams -
type InvokeParams struct {
	TransportParams
	InvokeOptions map[string]string
	RequestID     string
	TraceID       string
}

// KernelClientCallback -
type KernelClientCallback = func(result []byte, err snerror.SNError)

// KernelClientAsyncCreate -
type KernelClientAsyncCreate = func(function string, args []string, createParams CreateParams,
	callback KernelClientCallback) (string, snerror.SNError)

// KernelClientAsyncInvoke _
type KernelClientAsyncInvoke = func(function string, instanceID string, args []string, invokeParams InvokeParams,
	callback KernelClientCallback) snerror.SNError

// KernelClient defines basic POSIX client methods, it's worth noting that Create and
// Invoke are original async calls while others are sync calls
type KernelClient interface {
	Create(funcKey string, args []*api.Arg, createParams CreateParams, callback KernelClientCallback) (string,
		snerror.SNError)

	Invoke(funcKey string, instanceID string, args []*api.Arg, invokeParams InvokeParams,
		callback KernelClientCallback) (string, snerror.SNError)

	SaveState(state []byte) (string, snerror.SNError)

	LoadState(checkpointID string) ([]byte, snerror.SNError)

	Kill(instanceID string, signal int32, payload []byte) snerror.SNError

	Exit()
}

// SyncInvoke will call invoke synchronously
func SyncInvoke(asyncInvoke KernelClientAsyncInvoke, funcKey string, instanceID string, args []string,
	invokeParams InvokeParams) ([]byte, snerror.SNError) {
	CalibrateTransportParams(&invokeParams.TransportParams)
	var (
		resultData  []byte
		resultError snerror.SNError
	)
	waitCh := make(chan struct{}, 1)
	callback := func(result []byte, err snerror.SNError) {
		resultData, resultError = result, err
		waitCh <- struct{}{}
	}
	invokeErr := asyncInvoke(funcKey, instanceID, args, invokeParams, callback)
	if invokeErr != nil {
		return nil, invokeErr
	}
	timer := time.NewTimer(invokeParams.Timeout)
	defer timer.Stop()
	retryCount := 0
	for {
		select {
		case <-timer.C:
			log.GetLogger().Errorf("sync invoke times out after %ds for function %s traceID %s",
				invokeParams.Timeout.Seconds(), funcKey, invokeParams.TraceID)
			return nil, ErrKernelClientTimeout
		case <-waitCh:
			if resultError == nil {
				return resultData, nil
			}
			retryCount++
			if retryCount <= invokeParams.RetryNumber {
				time.Sleep(invokeParams.RetryInterval)
				log.GetLogger().Errorf("sync invoke reties count %d after %ds for function %s traceID %s",
					retryCount, invokeParams.RetryInterval.Seconds(), funcKey, invokeParams.TraceID)
				invokeErr = asyncInvoke(funcKey, instanceID, args, invokeParams, callback)
				if invokeErr != nil {
					return nil, invokeErr
				}
				continue
			}
			log.GetLogger().Errorf("sync invoke reties reach limit %d for function %s traceID %s",
				invokeParams.RetryNumber, funcKey, invokeParams.TraceID)
			return resultData, resultError
		}
	}
}

func pb2Arg(args []*api.Arg) []*common.Arg {
	length := len(args)
	newArgs := make([]*common.Arg, 0, length)
	for _, arg := range args {
		newArgs = append(newArgs, &common.Arg{Type: common.Arg_ArgType(arg.Type), Value: arg.Data})
	}
	return newArgs
}

// CalibrateTransportParams calibrates transport params
func CalibrateTransportParams(params *TransportParams) {
	if params.Timeout == 0 {
		params.Timeout = defaultTimeout
	}
}
