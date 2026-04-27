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

package rollout

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"yuanrong.org/kernel/runtime/libruntime/api"

	mockUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
)

func TestGetGlobalUpdateHandler(t *testing.T) {
	handler := GetGlobalRolloutHandler()
	assert.NotNil(t, handler)
}

func TestProcessRatio(t *testing.T) {
	handler := GetGlobalRolloutHandler()
	handler.UpdateForwardInstance("scheduler2")
	assert.Equal(t, handler.ForwardInstance, "scheduler2")
	rolloutRatio := &RolloutRatio{
		RolloutRatio: "aa%",
	}
	ratio, _ := json.Marshal(rolloutRatio)
	err := GetGlobalRolloutHandler().ProcessRatioUpdate(ratio)
	assert.NotNil(t, err)

	rolloutRatio = &RolloutRatio{
		RolloutRatio: "200%",
	}
	ratio, _ = json.Marshal(rolloutRatio)
	err = GetGlobalRolloutHandler().ProcessRatioUpdate(ratio)
	assert.NotNil(t, err)

	rolloutRatio = &RolloutRatio{
		RolloutRatio: "20%",
	}
	ratio, _ = json.Marshal(rolloutRatio)
	_ = GetGlobalRolloutHandler().ProcessRatioUpdate(ratio)
	forwardCount := 0
	notForwardCount := 0
	totalCount := 100
	for i := 0; i < totalCount; i++ {
		if handler.ShouldForwardRequest() {
			forwardCount++
		} else {
			notForwardCount++
		}
	}
	assert.Equal(t, 20, forwardCount)
	assert.Equal(t, 80, notForwardCount)

	GetGlobalRolloutHandler().ProcessRatioDelete()

	forwardCount = 0
	notForwardCount = 0
	totalCount = 100
	for i := 0; i < totalCount; i++ {
		if handler.ShouldForwardRequest() {
			forwardCount++
		} else {
			notForwardCount++
		}
	}
	assert.Equal(t, 0, forwardCount)
	assert.Equal(t, 100, notForwardCount)
}

func TestProcessAllocRecordSync(t *testing.T) {
	var (
		invokeRes []byte
		invokeErr error
	)
	defer gomonkey.ApplyFunc(InvokeByInstanceId, func(args []api.Arg, instanceID string, traceID string) ([]byte, error) {
		return invokeRes, invokeErr
	}).Reset()
	convey.Convey("Test ProcessAllocRecordSync", t, func() {
		invokeErr = errors.New("some error")
		globalRolloutHandler.ProcessAllocRecordSync("instance1", "instance2")
		convey.So(len(globalRolloutHandler.allocRecordSyncCh), convey.ShouldEqual, 0)
		invokeRes = []byte("error data")
		invokeErr = nil
		globalRolloutHandler.ProcessAllocRecordSync("instance1", "instance2")
		convey.So(len(globalRolloutHandler.allocRecordSyncCh), convey.ShouldEqual, 0)
		invokeRes = []byte(`{}`)
		globalRolloutHandler.ProcessAllocRecordSync("instance1", "instance2")
		convey.So(len(globalRolloutHandler.allocRecordSyncCh), convey.ShouldEqual, 1)
	})
}

func TestInvokeByInstanceId(t *testing.T) {
	SetRolloutSdkClient(&mockUtils.FakeLibruntimeSdkClient{})
	convey.Convey("test InvokeByInstanceId", t, func() {
		convey.Convey("invoke error", func() {
			defer gomonkey.ApplyMethod(reflect.TypeOf(&mockUtils.FakeLibruntimeSdkClient{}),
				"InvokeByInstanceId", func(_ *mockUtils.FakeLibruntimeSdkClient,
					funcMeta api.FunctionMeta, instanceID string, args []api.Arg,
					invokeOpt api.InvokeOptions) (string, error) {
					return "", errors.New("invoke error")
				}).Reset()
			_, err := InvokeByInstanceId([]api.Arg{}, "testInstance", "123")
			convey.So(err.Error(), convey.ShouldContainSubstring, "invoke error")
		})
		convey.Convey("success invoke", func() {
			defer gomonkey.ApplyMethod(reflect.TypeOf(&mockUtils.FakeLibruntimeSdkClient{}),
				"InvokeByInstanceId", func(_ *mockUtils.FakeLibruntimeSdkClient,
					funcMeta api.FunctionMeta, instanceID string, args []api.Arg,
					invokeOpt api.InvokeOptions) (string, error) {
					return "123", nil
				}).Reset()
			defer gomonkey.ApplyMethod(reflect.TypeOf(&mockUtils.FakeLibruntimeSdkClient{}),
				"GetAsync", func(_ *mockUtils.FakeLibruntimeSdkClient,
					objectID string, cb api.GetAsyncCallback) {
					cb([]byte("hello"), nil)
				}).Reset()
			res, err := InvokeByInstanceId([]api.Arg{}, "testInstance", "123")
			convey.So(err, convey.ShouldBeNil)
			convey.So(string(res), convey.ShouldEqual, "hello")
		})
	})
}
