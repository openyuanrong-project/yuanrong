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

// InstanceFunctionHandler provides method to execute the function
type InstanceFunctionHandler struct {
	instanceId string
	funcMeta   api.FunctionMeta
}

// NewInstanceFunctionHandler return a *InstanceFunctionHandler
func NewInstanceFunctionHandler(fn any, instanceId string) *InstanceFunctionHandler {
	if !IsFunction(fn) {
		panic(fmt.Errorf("fn parameter is not a function"))
	}
	functionName, _ := GetFunctionName(fn)
	meta := api.FunctionMeta{
		FuncName: functionName,
		FuncID:   GetConfigManager().FunctionId,
		Language: api.Golang,
	}
	return &InstanceFunctionHandler{funcMeta: meta, instanceId: instanceId}
}

// Invoke function invoke
func (handler *InstanceFunctionHandler) Invoke(args ...any) (refs []*ObjectRef) {
	packedArgs, err := PackInvokeArgs(args...)
	if err != nil {
		panic(fmt.Errorf(err.Error()))
	}
	objId, err := GetRuntimeHolder().GetRuntime().InvokeByInstanceId(handler.funcMeta, handler.instanceId,
		packedArgs, api.InvokeOptions{})
	if err != nil {
		panic(fmt.Sprintf("invoke failed, err: %v", err))
	}

	objRef := ObjectRef{objId: objId}
	refs = append(refs, &objRef)
	return
}
