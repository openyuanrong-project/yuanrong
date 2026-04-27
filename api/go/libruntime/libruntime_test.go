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
Package libruntime
This package provides a set of functions to interface with the clibruntime.
*/
package libruntime

import (
	"os"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/runtime/libruntime/api"
	"yuanrong.org/kernel/runtime/libruntime/config"
)

func TestInitAndReceiveRequestLoop(t *testing.T) {
	convey.Convey(
		"Test Init and ReceiveRequestLoop", t, func() {
			conf := config.Config{
				GrpcAddress:       "127.0.0.1",
				DataSystemAddress: os.Getenv("DATASYSTEM_ADDR"),
				JobID:             "jobId",
				RuntimeID:         "runtimeId",
				LogDir:            "logDir",
				LogLevel:          "DEBUG",
				InCluster:         true,
				Api:               api.PosixApi,
				Hooks: config.HookIntfs{
					LoadFunctionCb:      nil,
					FunctionExecutionCb: nil,
					CheckpointCb:        nil,
					RecoverCb:           nil,
					ShutdownCb:          nil,
					SignalCb:            nil,
					HealthCheckCb:       nil,
				},
			}
			convey.Convey(
				"Test Init Success", func() {
					err := Init(conf)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"Test ReceiveRequestLoop should not be panic", func() {
					convey.So(ReceiveRequestLoop, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestAllocReturnObject(t *testing.T) {
	convey.Convey(
		"Test AllocReturnObject", t, func() {
			convey.Convey(
				"AllocReturnObject Success", func() {
					var n uint = 8
					err := AllocReturnObject(&config.DataObject{}, 8, []string{""}, &n)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestSetReturnObject(t *testing.T) {
	convey.Convey(
		"Test SetReturnObject ", t, func() {
			convey.Convey(
				"SetReturnObject  Success", func() {
					convey.So(func() {
						SetReturnObject(&config.DataObject{}, 8)
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestWriterLatch(t *testing.T) {
	convey.Convey(
		"Test WriterLatch", t, func() {
			convey.Convey(
				"WriterLatch Success", func() {
					err := WriterLatch(&config.DataObject{})
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestMemoryCopy(t *testing.T) {
	convey.Convey(
		"Test MemoryCopy", t, func() {
			convey.Convey(
				"MemoryCopy Success", func() {
					err := MemoryCopy(&config.DataObject{}, []byte{0})
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestSeal(t *testing.T) {
	convey.Convey(
		"Test Seal", t, func() {
			convey.Convey(
				"Seal Success", func() {
					err := Seal(&config.DataObject{})
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestWriterUnlatch(t *testing.T) {
	convey.Convey(
		"Test WriterUnlatch", t, func() {
			convey.Convey(
				"WriterUnlatch Success", func() {
					err := WriterUnlatch(&config.DataObject{})
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}
