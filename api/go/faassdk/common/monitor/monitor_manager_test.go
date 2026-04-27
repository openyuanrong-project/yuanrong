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
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/runtime/faassdk/types"
)

func TestCreateFunctionMonitor(t *testing.T) {
	convey.Convey("CreateFunctionMonitor", t, func() {
		spec := &types.FuncSpec{ResourceMetaData: types.ResourceMetaData{Memory: 500}}
		convey.Convey("create monitor custom image disk", func() {
			monitor, err := CreateFunctionMonitor(spec, make(chan struct{}))
			convey.So(monitor, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("success", func() {
			monitor, err := CreateFunctionMonitor(spec, make(chan struct{}))
			convey.So(monitor, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
			close(monitor.MemoryManager.OOMChan)
			err = <-monitor.ErrChan
			convey.So(err, convey.ShouldEqual, ErrExecuteOOM)
		})
	})
}
