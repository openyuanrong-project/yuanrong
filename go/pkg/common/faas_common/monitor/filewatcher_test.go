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
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

func buildTestFile() string {
	path, _ := os.Getwd()
	if strings.Contains(path, "\\") {
		path = path + "\\test.json"
	} else {
		path = path + "/test.json"
	}

	return path
}

func TestCreateFileWatcher(t *testing.T) {
	convey.Convey("TestCreateFileWatcher error", t, func() {
		defer gomonkey.ApplyFunc(createDefaultFileWatcher, func(stopCh <-chan struct{}) (FileWatcher, error) {
			return nil, fmt.Errorf("fsnotify.NewWatcher error")
		}).Reset()
		stopCh := make(chan struct{}, 1)
		_, err := CreateFileWatcher(stopCh)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestInitFileWatcher(t *testing.T) {
	convey.Convey("TestInitFileWatcher", t, func() {
		stopCh := make(chan struct{}, 1)
		watcher, err := CreateFileWatcher(stopCh)
		convey.So(err, convey.ShouldBeNil)

		filename := buildTestFile()
		handler, _ := os.Create(filename)
		defer func() {
			handler.Close()
			os.Remove(filename)
		}()
		callbackChan := make(chan int, 5)
		watcher.RegisterCallback("", nil)
		watcher.RegisterCallback(filename, func(filename string, t OpType) {
			callbackChan <- 1
		})

		os.WriteFile(filename, []byte{'a'}, os.ModePerm)
		res := <-callbackChan
		convey.So(res, convey.ShouldBeGreaterThan, 0)
		time.Sleep(5 * time.Millisecond)
		close(stopCh)
	})
}

func TestInitFileWatcherWithInvalidCreator(t *testing.T) {
	convey.Convey("TestInitFileWatcherWithInvalidCreator", t, func() {
		defer SetCreator(createDefaultFileWatcher)
		SetCreator(func(stopCh <-chan struct{}) (FileWatcher, error) {
			return nil, errors.New("error")
		})

		stopCh := make(chan struct{}, 1)
		_, err := CreateFileWatcher(stopCh)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestFileWatcher_Start(t *testing.T) {
	convey.Convey("TestFileWatcher_Start", t, func() {
		stopCh := make(chan struct{}, 1)
		watcher, _ := CreateFileWatcher(stopCh)
		filename := "./TestFileWatcher_Start.tmp"
		f, _ := os.Create(filename)
		f.Close()
		tmp := hashRetry
		hashRetry = 1
		defer func() {
			hashRetry = tmp
		}()
		callbackChan := make(chan int, 1)
		watcher.RegisterCallback(filename, func(filename string, t OpType) {
			callbackChan <- 1
		})
		os.Remove(filename)
		time.AfterFunc(3*time.Second, func() {
			close(callbackChan)
		})
		convey.So(<-callbackChan, convey.ShouldEqual, 1)
		time.Sleep(50 * time.Millisecond)
		close(stopCh)
	})
}
