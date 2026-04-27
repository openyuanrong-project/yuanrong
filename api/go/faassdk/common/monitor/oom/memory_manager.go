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

// Package oom
package oom

import (
	"os"
	"path/filepath"
	"time"

	"yuanrong.org/kernel/runtime/libruntime/common/faas/logger"
)

const (
	mebibyte  = 1024 * 1024
	memoryGap = 5

	memoryPath = "/runtime/memory"
	memoryFile = "memory.stat"

	readMemoryInternal = 2 * time.Millisecond
)

// MemoryManager -
type MemoryManager struct {
	OOMChan   chan struct{}
	stopCh    chan struct{}
	PodMemory int
}

// NewMemoryManager -
func NewMemoryManager(podMemory int, stopCh chan struct{}) *MemoryManager {
	mm := &MemoryManager{
		OOMChan:   make(chan struct{}, 1),
		stopCh:    stopCh,
		PodMemory: podMemory,
	}
	go mm.WatchingRuntimeMemoryUsedLoop()
	return mm
}

// WatchingRuntimeMemoryUsedLoop - watch memory.stat until oom or process exit
func (mm *MemoryManager) WatchingRuntimeMemoryUsedLoop() {
	lastMemoryUsed := 0.0
	podMemory := mm.PodMemory
	limit := float64(podMemory - memoryGap)
	path := filepath.Join(memoryPath, os.Getenv("DELEGATE_CONTAINER_ID"), memoryFile)
	parser, err := NewCGroupMemoryParserWithPath(path)
	if err != nil {
		logger.GetLogger().Warnf("failed to create cgroup memory parser: %s", err.Error())
		return
	}
	defer parser.Close()
	logger.GetLogger().Infof("start to watch memory, path:%s, limit %f", path, limit)
	for {
		select {
		case _, ok := <-mm.stopCh:
			if !ok {
				logger.GetLogger().Warnf("context canceled")
				return
			}
		default:
			flag := mm.checkRuntimeMemory(&lastMemoryUsed, limit, parser)
			if flag {
				logger.GetLogger().Warnf("the runtime oom check process exited")
				return
			}
			time.Sleep(readMemoryInternal)
		}
	}
}

func (mm *MemoryManager) checkRuntimeMemory(lastMemoryUsed *float64, limit float64,
	parser *Parser) bool {
	memoryUsedBytes, err := parser.Read()
	if err != nil {
		logger.GetLogger().Warnf("failed to parse memory used, err: %s", err.Error())
		return false
	}
	tmpMemoryUsed := float64(memoryUsedBytes) / mebibyte
	// update used memory
	if tmpMemoryUsed != *lastMemoryUsed {
	}
	*lastMemoryUsed = tmpMemoryUsed
	if tmpMemoryUsed > limit {
		close(mm.OOMChan)
		return true
	}
	return false
}
