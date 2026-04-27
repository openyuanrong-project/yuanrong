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

// Package yr for actor
package yr

import (
	"fmt"

	"yuanrong.org/kernel/runtime/libruntime/api"
)

// InstanceCreator provides method to create instance
type InstanceCreator struct {
	meta    api.FunctionMeta
	options api.InvokeOptions
}

// Options set function option
func (instanceCreator *InstanceCreator) Options(options *InvokeOptions) *InstanceCreator {
	invokeOptions := api.InvokeOptions{
		Cpu:                options.Cpu,
		Memory:             options.Memory,
		InvokeLabels:       options.InvokeLabels,
		CustomResources:    options.CustomResources,
		CustomExtensions:   options.CustomExtensions,
		Labels:             options.Labels,
		ScheduleAffinities: options.ScheduleAffinities,
		Affinities:         options.Affinity,
		RetryTimes:         int(options.RetryTime),
		RecoverRetryTimes:  options.RecoverRetryTimes,
	}
	instanceCreator.options = invokeOptions
	return instanceCreator
}

// Invoke create instance
func (instanceCreator *InstanceCreator) Invoke(args ...any) *NamedInstance {
	packedArgs, err := PackInvokeArgs(args...)
	if err != nil {
		panic(fmt.Errorf(err.Error()))
	}
	instanceId, err := GetRuntimeHolder().GetRuntime().CreateInstance(
		instanceCreator.meta, packedArgs, instanceCreator.options,
	)
	if err != nil {
		panic(err)
	}
	handler := NewNamedInstance(instanceId)
	return handler
}

// Instance return a *InstanceCreator
func Instance(fn any) *InstanceCreator {
	if !IsFunction(fn) {
		panic(fmt.Errorf("paramenter is not a function"))
	}
	functionName, _ := GetFunctionName(fn)
	funcMeta := api.FunctionMeta{
		FuncName: functionName,
		Language: api.Golang,
		FuncID:   GetConfigManager().FunctionId,
	}
	invokeOptions := api.InvokeOptions{
		Cpu:                500,
		Memory:             500,
		InvokeLabels:       make(map[string]string),
		CustomResources:    make(map[string]float64),
		CustomExtensions:   make(map[string]string),
		Labels:             make([]string, 0),
		ScheduleAffinities: make([]api.Affinity, 0),
		Affinities:         make(map[string]string),
		RetryTimes:         0,
		Timeout:            60,
	}
	return &InstanceCreator{meta: funcMeta, options: invokeOptions}
}
