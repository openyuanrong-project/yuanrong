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

package selfregister

import (
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/loadbalance"
	"yuanrong.org/kernel/pkg/common/faas_common/types"
)

func Test_schedulerProxy_DealFilter(t *testing.T) {
	convey.Convey("test_deal_filter", t, func() {
		convey.Convey("base", func() {

			proxy := NewSchedulerProxy(loadbalance.NewCHGeneric())

			proxy.Add(&types.InstanceInfo{
				InstanceName: "aa2794fb-dc9e-420d-ae54-bedfa3577930",
			}, "")

			proxy.Add(&types.InstanceInfo{
				InstanceName: "7d3f736e-b2b0-4b7e-bc8d-3a390ec0ed31",
			}, "")

			proxy.Add(&types.InstanceInfo{
				InstanceName: "d06832bc-8c02-4589-9c37-edae4109302d",
			}, "")

			SelfInstanceID = "aa2794fb-dc9e-420d-ae54-bedfa3577930"

			flag := proxy.IsFuncOwner("244177614494719500/0@default@testcustom001/latest")

			convey.So(flag, convey.ShouldBeFalse)

			proxy.Remove(&types.InstanceInfo{
				InstanceName: "d06832bc-8c02-4589-9c37-edae4109302d",
			})

			flag = proxy.IsFuncOwner("244177614494719500/0@default@testcustom001/latest")

			convey.So(flag, convey.ShouldBeFalse)

			proxy.Remove(&types.InstanceInfo{
				InstanceName: "7d3f736e-b2b0-4b7e-bc8d-3a390ec0ed31",
			})

			flag = proxy.IsFuncOwner("244177614494719500/0@default@testcustom001/latest")

			convey.So(flag, convey.ShouldBeTrue)

			proxy.Add(&types.InstanceInfo{
				InstanceName: "d06832bc-8c02-4589-9c37-edae4109302d",
			}, "")

			flag = proxy.IsFuncOwner("244177614494719500/0@default@testcustom001/latest")

			convey.So(flag, convey.ShouldBeTrue)

			proxy.Add(&types.InstanceInfo{
				InstanceName: "7d3f736e-b2b0-4b7e-bc8d-3a390ec0ed31",
			}, "")

			proxy.Reset()

			flag = proxy.IsFuncOwner("244177614494719500/0@default@testcustom001/latest")

			convey.So(flag, convey.ShouldBeFalse)
		})
	})
}

func TestDealFilter(t *testing.T) {
	proxy := NewSchedulerProxy(loadbalance.NewSimpleCHGeneric())
	convey.Convey("start failed", t, func() {
		res := proxy.IsFuncOwner("mock-funcKey")
		convey.So(res, convey.ShouldBeFalse)
	})
	proxy.Add(&types.InstanceInfo{
		InstanceName: "scheduler-001",
	}, "")
	convey.Convey("start failed", t, func() {
		res := proxy.IsFuncOwner("mock-funcKey")
		convey.So(res, convey.ShouldBeFalse)
	})
	SelfInstanceID = "scheduler-001"
	convey.Convey("start success", t, func() {
		res := proxy.IsFuncOwner("mock-funcKey")
		convey.So(res, convey.ShouldBeTrue)
	})
	proxy.FaaSSchedulers.Delete("scheduler-001")
	convey.Convey("start failed", t, func() {
		res := proxy.IsFuncOwner("mock-funcKey")
		convey.So(res, convey.ShouldBeFalse)
	})
}

func TestContains(t *testing.T) {
	proxy := NewSchedulerProxy(loadbalance.NewSimpleCHGeneric())
	convey.Convey("not contains", t, func() {
		res := proxy.Contains("instance1")
		convey.So(res, convey.ShouldBeFalse)
	})
}

func Test(t *testing.T) {
	proxy := NewSchedulerProxy(loadbalance.NewSimpleCHGeneric())
	callTime := 0
	defer gomonkey.ApplyFunc(time.Sleep, func(d time.Duration) {
		callTime++
	}).Reset()
	convey.Convey("wait for hash", t, func() {
		proxy.WaitForHash(0)
		convey.So(callTime, convey.ShouldEqual, 0)
		proxy.FaaSSchedulers.Store("instance1", nil)
		proxy.WaitForHash(1)
		convey.So(callTime, convey.ShouldEqual, 0)
	})
}
