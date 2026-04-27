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

// Package monitor
package monitor

import (
	"errors"

	"yuanrong.org/kernel/runtime/faassdk/common/monitor/oom"
	"yuanrong.org/kernel/runtime/faassdk/types"
)

var (
	// ErrExecuteDiskLimit is the error of execute function disk limit
	ErrExecuteDiskLimit = errors.New("execute function disk limit")
	// ErrExecuteOOM is the error of exec ute function oom
	ErrExecuteOOM = errors.New("execute function oom")
)

// EnvDelegateContainer is the environment key of delegate-container
const EnvDelegateContainer = "DELEGATE_CONTAINER_ID"

// FunctionMonitorManager -
type FunctionMonitorManager struct {
	MemoryManager *oom.MemoryManager
	ErrChan       chan error
}

// CreateFunctionMonitor create function monitor include disk monitor, memory monitor
func CreateFunctionMonitor(funcSpec *types.FuncSpec, stopCh chan struct{}) (*FunctionMonitorManager, error) {
	memoryManager := oom.NewMemoryManager(funcSpec.ResourceMetaData.Memory, stopCh)

	m := &FunctionMonitorManager{
		MemoryManager: memoryManager,
		ErrChan:       make(chan error, 1),
	}
	go m.receiveMonitorsError()
	return m, nil
}

func (fmm *FunctionMonitorManager) receiveMonitorsError() {
	select {
	case <-fmm.MemoryManager.OOMChan:
		fmm.ErrChan <- ErrExecuteOOM
	}
}
