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
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/runtime/libruntime/api"
)

var clusterRt = &ClusterModeRuntime{}

func TestClusterModeRuntimeInit(t *testing.T) {
	convey.Convey(
		"Test ClusterModeRuntimeInit", t, func() {
			convey.Convey(
				"Init success", func() {
					err := clusterRt.Init()
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestReleaseInstance(t *testing.T) {
	convey.Convey(
		"Test ReleaseInstance", t, func() {
			convey.Convey(
				"ReleaseInstance success", func() {
					convey.So(func() {
						clusterRt.ReleaseInstance(api.InstanceAllocation{}, "", false, api.InvokeOptions{})
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestAcquireInstance(t *testing.T) {
	convey.Convey(
		"Test AcquireInstance", t, func() {
			convey.Convey(
				"AcquireInstance success", func() {
					instanceAllocation, err := clusterRt.AcquireInstance("", api.FunctionMeta{}, api.InvokeOptions{})
					convey.So(instanceAllocation, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestCreateInstanceRaw(t *testing.T) {
	convey.Convey(
		"Test CreateInstanceRaw", t, func() {
			convey.Convey(
				"CreateInstanceRaw success", func() {
					bytes, err := clusterRt.CreateInstanceRaw([]byte{0})
					convey.So(bytes, convey.ShouldBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestInvokeByInstanceIdRaw(t *testing.T) {
	convey.Convey(
		"Test InvokeByInstanceIdRaw", t, func() {
			convey.Convey(
				"InvokeByInstanceIdRaw success", func() {
					bytes, err := clusterRt.InvokeByInstanceIdRaw([]byte{0})
					convey.So(bytes, convey.ShouldBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestKillRaw(t *testing.T) {
	convey.Convey(
		"Test KillRaw", t, func() {
			convey.Convey(
				"KillRaw success", func() {
					bytes, err := clusterRt.KillRaw([]byte{0})
					convey.So(bytes, convey.ShouldBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestSaveState(t *testing.T) {
	convey.Convey(
		"Test SaveState", t, func() {
			convey.Convey(
				"SaveState success", func() {
					str, err := clusterRt.SaveState([]byte{0})
					convey.So(str, convey.ShouldBeEmpty)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestLoadState(t *testing.T) {
	convey.Convey(
		"Test LoadState", t, func() {
			convey.Convey(
				"LoadState success", func() {
					bytes, err := clusterRt.LoadState("")
					convey.So(bytes, convey.ShouldBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestKVSetWithoutKey(t *testing.T) {
	convey.Convey(
		"Test KVSetWithoutKey", t, func() {
			convey.Convey(
				"KVSetWithoutKey success", func() {
					str, err := clusterRt.KVSetWithoutKey([]byte{0}, api.SetParam{})
					convey.So(str, convey.ShouldBeEmpty)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestSetTraceID(t *testing.T) {
	convey.Convey(
		"Test SetTraceID", t, func() {
			convey.Convey(
				"SetTraceID success", func() {
					convey.So(func() {
						clusterRt.SetTraceID("")
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestSetTenantID(t *testing.T) {
	convey.Convey(
		"Test SetTenantID", t, func() {
			convey.Convey(
				"SetTenantID success", func() {
					err := clusterRt.SetTenantID("")
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestPutRaw(t *testing.T) {
	convey.Convey(
		"Test PutRaw", t, func() {
			convey.Convey(
				"PutRaw success", func() {
					err := clusterRt.PutRaw("", []byte{0}, api.PutParam{}, "")
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestGetRaw(t *testing.T) {
	convey.Convey(
		"Test GetRaw", t, func() {
			convey.Convey(
				"GetRaw success", func() {
					bytesMatrix, err := clusterRt.GetRaw([]string{""}, 3000)
					convey.So(bytesMatrix, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestGIncreaseRefRaw(t *testing.T) {
	convey.Convey(
		"Test GIncreaseRefRaw", t, func() {
			convey.Convey(
				"GIncreaseRefRaw success", func() {
					strs, err := clusterRt.GIncreaseRefRaw([]string{""}, "")
					convey.So(strs, convey.ShouldBeEmpty)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestGDecreaseRef(t *testing.T) {
	convey.Convey(
		"Test GDecreaseRef", t, func() {
			convey.Convey(
				"GDecreaseRef success", func() {
					strs, err := clusterRt.GDecreaseRef([]string{""}, "")
					convey.So(strs, convey.ShouldBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestGDecreaseRefRaw(t *testing.T) {
	convey.Convey(
		"Test GDecreaseRefRaw", t, func() {
			convey.Convey(
				"GDecreaseRefRaw success", func() {
					strs, err := clusterRt.GDecreaseRefRaw([]string{""}, "")
					convey.So(strs, convey.ShouldBeEmpty)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestGetAsync(t *testing.T) {
	convey.Convey(
		"Test GetAsync", t, func() {
			convey.Convey(
				"GetAsync success", func() {
					convey.So(func() {
						clusterRt.GetAsync("", nil)
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestGetEvent(t *testing.T) {
	convey.Convey(
		"Test GetEvent", t, func() {
			convey.Convey(
				"GetEvent success", func() {
					convey.So(func() {
						clusterRt.GetEvent("", nil)
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestDeleteGetEventCallback(t *testing.T) {
	convey.Convey(
		"Test DeleteGetEventCallback", t, func() {
			convey.Convey(
				"DeleteGetEventCallback success", func() {
					convey.So(func() {
						clusterRt.DeleteGetEventCallback("")
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestGetFormatLogger(t *testing.T) {
	convey.Convey(
		"Test GetFormatLogger", t, func() {
			convey.Convey(
				"GetFormatLogger success", func() {
					logger := clusterRt.GetFormatLogger()
					convey.So(logger, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestCreateClient(t *testing.T) {
	convey.Convey(
		"Test CreateClient", t, func() {
			convey.Convey(
				"CreateClient success", func() {
					kvClient, err := clusterRt.CreateClient(api.ConnectArguments{})
					convey.So(kvClient, convey.ShouldBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}
