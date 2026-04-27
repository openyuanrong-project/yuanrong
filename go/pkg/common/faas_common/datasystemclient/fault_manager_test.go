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

package datasystemclient

import (
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

func TestIsStatusReady(t *testing.T) {
	convey.Convey("test *LocalDataSystemStatusCache IsStatusReady()", t, func() {
		var dataSystemStatusCache LocalDataSystemStatusCache
		convey.Convey("test dataSystem status is ready", func() {
			dataSystemStatusCache.status = dataSystemStatusReady
			result := dataSystemStatusCache.IsStatusReady()
			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("test dataSystem status is not ready", func() {
			dataSystemStatusCache.status = dataSystemStatusExiting
			result := dataSystemStatusCache.IsStatusReady()
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestLocalDataSystemStatusCacheGetLocalDataSystemStatus(t *testing.T) {
	convey.Convey("test *LocalDataSystemStatusCache GetLocalDataSystemStatus", t, func() {
		convey.Convey("test get dataSystem status", func() {
			var dataSystemStatusCache LocalDataSystemStatusCache
			dataSystemStatusCache.status = dataSystemStatusReady
			convey.So(dataSystemStatusCache.GetLocalDataSystemStatus(), convey.ShouldEqual, dataSystemStatusReady)
		})
	})
}

func TestLocalDataSystemStatusCacheSetLocalDataSystemStatus(t *testing.T) {
	convey.Convey("test *LocalDataSystemStatusCache SetLocalDataSystemStatus", t, func() {
		var dataSystemStatusCache LocalDataSystemStatusCache
		convey.Convey("test set dataSystem status, when NODE_IP is empty", func() {
			dataSystemStatusCache.SetLocalDataSystemStatus("", dataSystemStatusReady)
			convey.So(dataSystemStatusCache.GetLocalDataSystemStatus(), convey.ShouldEqual, "")
		})

		convey.Convey("test set dataSystem status, when NODE_IP is not equal dataSystemStatusCache", func() {
			defer gomonkey.ApplyFunc(os.Getenv, func(key string) string {
				return "0.0.0.0"
			}).Reset()
			dataSystemStatusCache.SetLocalDataSystemStatus("0.0.0.1", dataSystemStatusReady)
			convey.So(dataSystemStatusCache.GetLocalDataSystemStatus(), convey.ShouldNotEqual, dataSystemStatusReady)
		})

		convey.Convey("test set dataSystem status, when NODE_IP is equal dataSystemStatusCache", func() {
			defer gomonkey.ApplyFunc(os.Getenv, func(key string) string {
				return "0.0.0.0"
			}).Reset()
			dataSystemStatusCache.SetLocalDataSystemStatus("0.0.0.0", dataSystemStatusReady)
			convey.So(dataSystemStatusCache.GetLocalDataSystemStatus(), convey.ShouldEqual, dataSystemStatusReady)
		})
	})
}

func TestIsLocalDataSystemStatusReady(t *testing.T) {
	convey.Convey("test IsLocalDataSystemStatusReady", t, func() {
		original := localDataSystemStatusCache.status
		defer func() {
			localDataSystemStatusCache.status = original
		}()
		convey.Convey("local dataSystem status is ready", func() {
			localDataSystemStatusCache.status = dataSystemStatusReady
			result := IsLocalDataSystemStatusReady()
			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("local dataSystem status is not ready", func() {
			localDataSystemStatusCache.status = dataSystemStatusExiting
			result := IsLocalDataSystemStatusReady()
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestSetStreamEnable(t *testing.T) {
	convey.Convey("test SetStreamEnable", t, func() {
		convey.Convey("test set streamEnable", func() {
			SetStreamEnable(false)
			convey.So(streamEnable.Load(), convey.ShouldBeFalse)
		})
	})
}

func TestIsShutdownFronted(t *testing.T) {
	convey.Convey("test is shout down frontend", t, func() {
		originalStreamEnable := streamEnable.Load()
		defer func() {
			streamEnable.Store(originalStreamEnable)
		}()
		streamEnable.Store(true)

		convey.Convey("when streamEnable is false, skip shutdown", func() {
			streamEnable.Store(false)
			result := isShutdownFronted()
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("when dataSystem status is ready, skip shutdown", func() {
			defer gomonkey.ApplyMethodFunc(&LocalDataSystemStatusCache{}, "GetLocalDataSystemStatus", func() string {
				return dataSystemStatusReady
			}).Reset()
			result := isShutdownFronted()
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("when dataSystem status is exiting, skip shutdown", func() {
			defer gomonkey.ApplyMethodFunc(&LocalDataSystemStatusCache{}, "GetLocalDataSystemStatus", func() string {
				return dataSystemStatusExiting
			}).Reset()
			result := isShutdownFronted()
			convey.So(result, convey.ShouldBeTrue)
		})
	})
}

func TestDestroy(t *testing.T) {
	convey.Convey("test destroy frontend, when watch dataSystem", t, func() {
		convey.Convey("destroy success", func() {
			defer gomonkey.ApplyMethodFunc(&os.Process{}, "Signal", func(sig os.Signal) error {
				return nil
			}).Reset()
			destroy()
			convey.So("", convey.ShouldBeEmpty)
		})
	})
}
