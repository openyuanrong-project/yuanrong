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

// Package instancequeue -
package instancequeue

import (
	"errors"

	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	"yuanrong.org/kernel/pkg/common/faas_common/statuscode"
	"yuanrong.org/kernel/pkg/functionscaler/metrics"
	"yuanrong.org/kernel/pkg/functionscaler/requestqueue"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

var (
	// ErrInsNotExist is the error of instance does not exist
	ErrInsNotExist = errors.New("instance does not exist in queue")
	// ErrInsSubHealth is the error of instance is not normal
	ErrInsSubHealth = errors.New("instance is subHealth")
	// ErrNoInsThdAvailable is the error of no instance available
	ErrNoInsThdAvailable = errors.New("no instance thread available now")
	// ErrFuncSigMismatch is the error of "function signature mismatch"
	ErrFuncSigMismatch = errors.New("function signature mismatch")
	// ErrFunctionDeleted is the error of "function is deleted"
	ErrFunctionDeleted = errors.New("function is deleted")
)

// InsQueConfig -
type InsQueConfig struct {
	FuncSpec         *types.FunctionSpecification
	InsThdReqQueue   *requestqueue.InsAcqReqQueue
	InstanceType     types.InstanceType
	ResKey           resspeckey.ResSpecKey
	MetricsCollector metrics.Collector
	CreateInstanceFunc
	DeleteInstanceFunc
	SignalInstanceFunc
}

// InstanceOperationFunc contains functions which operates instance
type InstanceOperationFunc struct{}

// CreateInstanceFunc -
type CreateInstanceFunc func(string, string, types.InstanceType, resspeckey.ResSpecKey, []byte) (*types.Instance, error)

// DeleteInstanceFunc -
type DeleteInstanceFunc func(*types.Instance) error

// SignalInstanceFunc -
type SignalInstanceFunc func(*types.Instance, int)

// InstanceQueue stores instances
type InstanceQueue interface {
	AcquireInstance(insAcqReq *types.InstanceAcquireRequest) (*types.InstanceAllocation, snerror.SNError)
	HandleInstanceUpdate(instance *types.Instance)
	HandleInstanceDelete(instance *types.Instance)
	HandleFuncSpecUpdate(funcSpec *types.FunctionSpecification)
	Destroy()
	GetInstanceNumber(onlySelf bool) int
}

func buildSnError(err error) snerror.SNError {
	if snErr, ok := err.(snerror.SNError); ok {
		return snErr
	}
	switch err {
	case nil:
		return nil
	case scheduler.ErrNoInsAvailable:
		return snerror.New(statuscode.NoInstanceAvailableErrCode, err.Error())
	case scheduler.ErrInsNotExist, scheduler.ErrInsSubHealthy:
		return snerror.New(statuscode.InstanceNotFoundErrCode, err.Error())
	case scheduler.ErrInsReqTimeout:
		return snerror.New(statuscode.InsThdReqTimeoutCode, err.Error())
	case scheduler.ErrInvalidSession:
		return snerror.New(statuscode.InstanceSessionInvalidErrCode, err.Error())
	default:
		return snerror.New(statuscode.StatusInternalServerError, err.Error())
	}
}
