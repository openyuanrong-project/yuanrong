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

// Package scaler -
package scaler

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/instanceconfig"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
)

func TestReplicaScaler_CreateAndDestroy(t *testing.T) {
	convey.Convey("create and destroy", t, func() {
		s := NewReplicaScaler("testFunc", nil, nil, nil)
		rs, ok := s.(*ReplicaScaler)
		convey.So(ok, convey.ShouldEqual, true)
		rs.Destroy()
		convey.So(rs.enable, convey.ShouldEqual, false)
		convey.So(rs.targetRsvInsNum, convey.ShouldEqual, 0)
	})
}

func TestReplicaScaler_TriggerScale(t *testing.T) {
	rs := &ReplicaScaler{}
	convey.Convey("trigger scale", t, func() {
		callCount := 0
		p := gomonkey.ApplyFunc((*ReplicaScaler).handleScale, func() {
			callCount++
		})
		rs.TriggerScale()
		convey.So(callCount, convey.ShouldEqual, 0)
		rs.SetEnable(true)
		rs.TriggerScale()
		convey.So(callCount, convey.ShouldEqual, 2)
		p.Reset()
	})
}

func TestReplicaScaler_GetExpectInstanceNumber(t *testing.T) {
	rs := &ReplicaScaler{
		concurrentNum: 100,
		logger:        log.GetLogger(),
	}
	rs.scaleUpHandler = func(i int, callback ScaleUpCallback) {}
	rs.SetEnable(true)
	convey.Convey("increase", t, func() {
		rs.HandleInsConfigUpdate(&instanceconfig.Configuration{
			InstanceMetaData: commonTypes.InstanceMetaData{
				MinInstance: 1,
			},
		})
		convey.So(rs.GetExpectInstanceNumber(), convey.ShouldEqual, 1)
	})
}

func TestReplicaScaler_pendingInsNumOperation(t *testing.T) {
	rs := &ReplicaScaler{
		concurrentNum: 100,
		logger:        log.GetLogger(),
	}
	convey.Convey("increase", t, func() {
		rs.handlePendingInsNumIncrease(1)
		convey.So(rs.pendingRsvInsNum, convey.ShouldEqual, 1)
	})
	convey.Convey("decrease", t, func() {
		rs.handlePendingInsNumDecrease(1)
		convey.So(rs.pendingRsvInsNum, convey.ShouldEqual, 0)
	})
}
