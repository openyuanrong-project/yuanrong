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

/*
Package execution
This package provides methods to obtain the execution interface.
*/
package execution

import (
	"yuanrong.org/kernel/runtime/libruntime/api"
	"yuanrong.org/kernel/runtime/libruntime/config"
)

// FunctionExecutionIntfs This interface defines several callback functions, for example, FunctionExecute.
type FunctionExecutionIntfs interface {
	GetExecType() ExecutionType
	LoadFunction(codePaths []string) error
	FunctionExecute(
		funcMeta api.FunctionMeta, invokeType config.InvokeType, args []api.Arg, returnobjs []config.DataObject,
	) error
	Checkpoint(checkpointID string) ([]byte, error)
	Recover(state []byte) error
	Shutdown(gracePeriod uint64) error
	Signal(sig int, data []byte) error
	HealthCheck() (api.HealthType, error)
}

// ExecutionType execution type, for example, posix
type ExecutionType int32

const (
	// ExecutionTypeInvalid invalid type
	ExecutionTypeInvalid ExecutionType = 0
	// ExecutionTypeActor actor type
	ExecutionTypeActor ExecutionType = 1
	// ExecutionTypeFaaS faas type
	ExecutionTypeFaaS ExecutionType = 2
	// ExecutionTypePosix posix type
	ExecutionTypePosix ExecutionType = 3
)
