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

// FunctionHandler provides method to execute the function
type FunctionHandler struct {
	funcMeta      api.FunctionMeta
	invokeOptions api.InvokeOptions
}

// Function return a *FunctionHandler
func Function(fn any) *FunctionHandler {
	if !IsFunction(fn) {
		panic(fmt.Errorf("fn parament is not a function"))
	}
	functionName, _ := GetFunctionName(fn)
	funcMeta := api.FunctionMeta{
		FuncName: functionName,
		FuncID:   GetConfigManager().FunctionId,
		Language: api.Golang,
	}

	invokeOptions := api.InvokeOptions{
		Cpu:                500,
		Memory:             500,
		InvokeLabels:       make(map[string]string),
		CustomResources:    make(map[string]float64),
		CustomExtensions:   make(map[string]string),
		Labels:             make([]string, 0),
		Affinities:         make(map[string]string),
		ScheduleAffinities: make([]api.Affinity, 0),
		RetryTimes:         0,
	}
	return &FunctionHandler{
		funcMeta:      funcMeta,
		invokeOptions: invokeOptions,
	}
}

// Options set function option
func (handler *FunctionHandler) Options(options *InvokeOptions) *FunctionHandler {
	invokeOptions := api.InvokeOptions{
		Cpu:                options.Cpu,
		Memory:             options.Memory,
		InvokeLabels:       options.InvokeLabels,
		CustomResources:    options.CustomResources,
		CustomExtensions:   options.CustomExtensions,
		Labels:             options.Labels,
		Affinities:         options.Affinity,
		ScheduleAffinities: options.ScheduleAffinities,
		RetryTimes:         int(options.RetryTime),
	}
	handler.invokeOptions = invokeOptions
	return handler
}

// Invoke function invoke
func (handler *FunctionHandler) Invoke(args ...any) (refs []*ObjectRef) {
	packedArgs, err := PackInvokeArgs(args...)
	if err != nil {
		panic(fmt.Errorf(err.Error()))
	}

	objId, err := GetRuntimeHolder().GetRuntime().InvokeByFunctionName(
		handler.funcMeta, packedArgs, handler.invokeOptions,
	)
	if err != nil {
		panic(fmt.Sprintf("invoke failed, err: %v", err))
	}

	objRef := ObjectRef{objId: objId}
	refs = append(refs, &objRef)
	return
}
