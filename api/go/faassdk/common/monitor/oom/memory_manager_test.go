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
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

func TestNewMemoryManager(t *testing.T) {
	convey.Convey("NewMemoryManager", t, func() {
		stopCh := make(chan struct{})
		manager := NewMemoryManager(500, stopCh)
		convey.So(manager, convey.ShouldNotBeNil)
		close(stopCh)
	})
}

func createTempMemoryStatFile(path string, memory int) {
	file, _ := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	file.WriteString(fmt.Sprintf("rss %d", memory))
	file.Close()
}

func TestMemoryManager_WatchingRuntimeMemoryUsedLoop(t *testing.T) {
	convey.Convey("WatchingRuntimeMemoryUsedLoop", t, func() {
		tmpFileMemoryStat := "./memory.stat"
		createTempMemoryStatFile(tmpFileMemoryStat, 400*1024*1024)
		patches := []*gomonkey.Patches{
			gomonkey.ApplyFunc(filepath.Join, func(elem ...string) string {
				return tmpFileMemoryStat
			}),
		}
		defer func() {
			for _, patch := range patches {
				patch.Reset()
			}
			err := os.Remove(tmpFileMemoryStat)
			if err != nil {
				fmt.Println(err)
			}
		}()
		convey.Convey("success", func() {
			stopCh := make(chan struct{})
			manager := NewMemoryManager(500, stopCh)
			convey.So(manager, convey.ShouldNotBeNil)
			defer close(stopCh)

			time.Sleep(4 * time.Millisecond)
			createTempMemoryStatFile(tmpFileMemoryStat, 500*1024*1024)
			_, ok := <-manager.OOMChan
			convey.So(ok, convey.ShouldEqual, false)
		})

		convey.Convey("stop channel close", func() {
			stopCh := make(chan struct{})
			manager := NewMemoryManager(500, stopCh)
			convey.So(manager, convey.ShouldNotBeNil)
			time.Sleep(4 * time.Millisecond)
			close(stopCh)
		})
	})
}
