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

// Package faasscheduler -
package faasscheduler

import (
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/runtime/faassdk/common/loadbalance"
	"yuanrong.org/kernel/runtime/libruntime/api"
)

func Test_schedulerProxy_Get(t *testing.T) {
	convey.Convey("Get", t, func() {
		convey.Convey("assert failed", func() {
			patches := []*gomonkey.Patches{
				gomonkey.ApplyMethod(reflect.TypeOf(&loadbalance.ConcurrentCHGeneric{}), "Next",
					func(_ *loadbalance.ConcurrentCHGeneric, name string, move bool) interface{} {
						return 123
					}),
			}
			defer func() {
				for _, patch := range patches {
					time.Sleep(100 * time.Millisecond)
					patch.Reset()
				}
			}()
			_, err := Proxy.Get("functionKey")
			convey.So(err, convey.ShouldBeError)
		})

		convey.Convey("no avaiable faas scheduler was found", func() {
			patches := []*gomonkey.Patches{
				gomonkey.ApplyMethod(reflect.TypeOf(&loadbalance.ConcurrentCHGeneric{}), "Next",
					func(_ *loadbalance.ConcurrentCHGeneric, name string, move bool) interface{} {
						return ""
					}),
			}
			defer func() {
				for _, patch := range patches {
					time.Sleep(100 * time.Millisecond)
					patch.Reset()
				}
			}()
			_, err := Proxy.Get("functionKey")
			convey.So(err, convey.ShouldBeError)
		})

		convey.Convey("failed to get the faas scheduler named", func() {
			patches := []*gomonkey.Patches{
				gomonkey.ApplyMethod(reflect.TypeOf(&loadbalance.ConcurrentCHGeneric{}), "Next",
					func(_ *loadbalance.ConcurrentCHGeneric, name string, move bool) interface{} {
						return "faaSScheduler"
					}),
			}
			defer func() {
				for _, patch := range patches {
					time.Sleep(100 * time.Millisecond)
					patch.Reset()
				}
			}()
			_, err := Proxy.Get("functionKey")
			convey.So(err, convey.ShouldBeError)
		})

		convey.Convey("success", func() {
			patches := []*gomonkey.Patches{
				gomonkey.ApplyMethod(reflect.TypeOf(&loadbalance.ConcurrentCHGeneric{}), "Next",
					func(_ *loadbalance.ConcurrentCHGeneric, name string, move bool) interface{} {
						return "instance1"
					}),
			}
			defer func() {
				for _, patch := range patches {
					time.Sleep(100 * time.Millisecond)
					patch.Reset()
				}
			}()
			Proxy.Add("instance1")
			instanceID, err := Proxy.Get("functionKey")
			convey.So(err, convey.ShouldBeNil)
			convey.So(instanceID, convey.ShouldEqual, "instance1")

			Proxy.Remove("instance1")
			_, err2 := Proxy.Get("functionKey")
			convey.So(err2.Error(), convey.ShouldEqual, "no avaiable faas scheduler was found")
		})
	})
}

func TestParseSchedulerData(t *testing.T) {
	convey.Convey("TestParseSchedulerData", t, func() {
		convey.Convey("success", func() {
			err := ParseSchedulerData(api.Arg{
				Type: api.Value,
				Data: []byte(`{"schedulerFuncKey": "faasscheudler", "schedulerIDList": ["123"]}`),
			})
			convey.So(err, convey.ShouldBeNil)
			key := Proxy.GetSchedulerFuncKey()
			convey.So(key, convey.ShouldEqual, "faasscheudler")
			Proxy.Remove("faasscheudler")
		})
	})
}

func TestSetStain(t *testing.T) {
	convey.Convey("", t, func() {
		Proxy.Add("faasscheduler")
		Proxy.SetStain("function", "faasscheduler")
		Proxy.Remove("faasscheduler")
	})
}
