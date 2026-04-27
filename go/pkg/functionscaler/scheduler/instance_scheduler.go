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

// Package scheduler -
package scheduler

import (
	"errors"

	"yuanrong.org/kernel/pkg/functionscaler/scaler"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

// InstanceTopic defines topic in instanceScheduler
type InstanceTopic string

const (
	// AvailInsThdTopic is the topic of available instance thread
	AvailInsThdTopic InstanceTopic = "availInsThd"
	// InUseInsThdTopic is the topic of in-use instance thread
	InUseInsThdTopic InstanceTopic = "inUseInsThd"
	// TotalInsThdTopic is the topic of total instance thread
	TotalInsThdTopic InstanceTopic = "totalInsThd"
	// TriggerScaleTopic is the topic of trigger scale
	TriggerScaleTopic InstanceTopic = "triggerScale"
	// CreateErrorTopic is the topic of instance thread create error
	CreateErrorTopic InstanceTopic = "createErr"
)

var (
	// ErrUnsupported is the error of operation unsupported
	ErrUnsupported = errors.New("operation unsupported")
	// ErrInternal is the error of internal error
	ErrInternal = errors.New("internal error")
	// ErrTypeConvertFail is the error of type convert failed
	ErrTypeConvertFail = errors.New("type convert failed")
	// ErrInsNotExist is the error of instance does not exist
	ErrInsNotExist = errors.New("instance does not exist in queue")
	// ErrInsAlreadyExist is the error of instance already exist
	ErrInsAlreadyExist = errors.New("instance already exist in queue")
	// ErrInsSubHealthy is the error of instance already exist
	ErrInsSubHealthy = errors.New("instance is not healthy")
	// ErrNoInsAvailable is the error of no instance available
	ErrNoInsAvailable = errors.New("no instance available")
	// ErrDesignateInsNotAvailable is the error of designateInstance not available
	ErrDesignateInsNotAvailable = errors.New("designateInstance not available")
	// ErrInsReqTimeout is the error of no instance request timeout
	ErrInsReqTimeout = errors.New("instance request timeout")
	// ErrInvalidSession is the error of invalid session parameter
	ErrInvalidSession = errors.New("invalid session parameter")
	// ErrFuncSigMismatch is the error of "function signature mismatch"
	ErrFuncSigMismatch = errors.New("function signature mismatch")
	// ErrFunctionDeleted is the error of "function is deleted"
	ErrFunctionDeleted = errors.New("function is deleted")
	// ErrSessionNotFound is the error of session not found
	ErrSessionNotFound = errors.New("session not found")
	// ErrOverAcqLimitExceeded is the error of over-acquire limit exceeded for AI Agent session
	ErrOverAcqLimitExceeded = errors.New("over-acquire limit exceeded for AI Agent session")
)

// SignalInstanceFunc sends certain signal to instance
type SignalInstanceFunc func(*types.Instance)

// InstanceScheduler schedules instance
type InstanceScheduler interface {
	GetInstanceNumber(onlySelf bool) int
	AcquireInstance(insAcqReq *types.InstanceAcquireRequest) (*types.InstanceAllocation, error)
	ReleaseInstance(insAlloc *types.InstanceAllocation) error
	AddInstance(instance *types.Instance) error
	DelInstance(instance *types.Instance) error
	PopInstance(force bool) *types.Instance
	ConnectWithInstanceScaler(instanceScaler scaler.InstanceScaler)
	HandleFuncSpecUpdate(funcSpec *types.FunctionSpecification)
	HandleInstanceUpdate(instance *types.Instance)
	HandleFuncOwnerUpdate(isFuncOwner bool)
	HandleCreateError(createErr error)
	SignalAllInstances(signalFunc SignalInstanceFunc)
	ReassignInstanceWhenGray(ratio int)
	Destroy()
}
