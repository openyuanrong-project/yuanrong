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

package monitor

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
)

func TestDefaultFileWatcherStart(t *testing.T) {
	eventChan := make(chan fsnotify.Event)
	stopCh := make(chan struct{})
	errorChan := make(chan error)
	mockWatcher := &fsnotify.Watcher{
		Events: eventChan,
		Errors: errorChan,
	}

	invokeCallbackCh := make(chan bool)
	closeCh := make(chan bool)

	watcher := &defaultFileWatcher{
		watcher:  mockWatcher,
		filename: "/path/testfile.txt",
		stopCh:   stopCh,
		callback: func(filename string, opType OpType) {
			invokeCallbackCh <- true
		},
	}
	patches := []*gomonkey.Patches{
		gomonkey.ApplyFunc(ioutil.ReadFile, func(filename string) ([]byte, error) {
			return []byte("mock content"), nil
		}),
		gomonkey.ApplyMethod(reflect.TypeOf(watcher.watcher), "Remove", func(_ *fsnotify.Watcher, _ string) error {
			return nil
		}),
		gomonkey.ApplyFunc(filepath.EvalSymlinks, func(path string) (string, error) {
			return "/mock/symlink/path", nil
		}),
		gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
			return nil, nil
		}),
		gomonkey.ApplyMethod(reflect.TypeOf(watcher.watcher), "Add", func(_ *fsnotify.Watcher, _ string) error {
			return nil
		}),
		gomonkey.ApplyMethod(reflect.TypeOf(watcher.watcher), "Close", func(_ *fsnotify.Watcher) error {
			closeCh <- true
			return nil
		}),
	}
	defer func() {
		for _, patch := range patches {
			patch.Reset()
		}
	}()

	go watcher.Start()

	errorChan <- fmt.Errorf("err")
	eventChan <- fsnotify.Event{Name: "/path/test1.txt", Op: fsnotify.Remove}

	invokeCallback := <-invokeCallbackCh
	close(stopCh)

	assert.Equal(t, invokeCallback, true)
	isClosed := <-closeCh
	assert.Equal(t, isClosed, true)
}
