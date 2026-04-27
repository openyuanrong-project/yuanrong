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

// Package monitor provide memory and file monitor
package monitor

import (
	"errors"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

// OpType describes file operation type
type OpType uint32

const (
	// Create op type
	Create OpType = 1 << iota
	// Write op type
	Write
	// Remove op type
	Remove
	// Rename op type
	Rename
	// Chmod op type
	Chmod
)

var (
	creator Creator = createDefaultFileWatcher
)

// FileChangedCallback describes callback function, when file changed, callback function will be invoked
type FileChangedCallback func(filename string, opType OpType)

// Creator describes watcher create function
type Creator func(stopCh <-chan struct{}) (FileWatcher, error)

// FileWatcher describes interface of general FileWatcher
type FileWatcher interface {
	Start()
	RegisterCallback(filename string, callback FileChangedCallback)
}

// SetCreator set file watcher creator func, if not set, use createDefaultFileWatcher
func SetCreator(newCreator Creator) {
	creator = newCreator
}

// CreateFileWatcher create a file watcher
// notice: one FileWatcher can only watcher one file
func CreateFileWatcher(stopCh <-chan struct{}) (FileWatcher, error) {
	watcher, err := creator(stopCh)
	if err != nil {
		log.GetLogger().Errorf("create watcher failed %s", err.Error())
		return nil, err
	}
	if watcher == nil {
		log.GetLogger().Errorf("watcher is nil")
		return nil, errors.New("watcher is nil")
	}
	go watcher.Start()
	return watcher, nil
}
