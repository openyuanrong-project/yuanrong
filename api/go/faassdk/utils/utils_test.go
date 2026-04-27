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

// Package utils
package utils

import (
	"errors"
	"os"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestUniqueID(t *testing.T) {
	convey.Convey("Test UniqueID", t, func() {
		convey.Convey("UniqueID success", func() {
			convey.So(UniqueID(), convey.ShouldNotBeEmpty)
		})
	})
}

func TestBytesToString(t *testing.T) {
	convey.Convey("Test BytesToString", t, func() {
		convey.Convey("BytesToString success", func() {
			convey.So(BytesToString([]byte("str")), convey.ShouldEqual, "str")
		})
	})
}

func TestGetServerIP(t *testing.T) {
	convey.Convey("Test GetServerIP", t, func() {
		convey.Convey("GetServerIP success", func() {
			ip, err := GetServerIP()
			convey.So(ip, convey.ShouldNotBeEmpty)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestGetLibInfo(t *testing.T) {
	convey.Convey("Test GetLibInfo", t, func() {
		convey.Convey("GetLibInfo when length < minPathPartLength", func() {
			fileName, handlerName := GetLibInfo("", "")
			convey.So(fileName, convey.ShouldBeEmpty)
			convey.So(handlerName, convey.ShouldBeEmpty)
		})
		convey.Convey("GetLibInfo success", func() {
			fileName, handlerName := GetLibInfo("/tmp", "example.init")
			convey.So(fileName, convey.ShouldEqual, "/tmp/example.so")
			convey.So(handlerName, convey.ShouldEqual, "init")
		})
		convey.Convey("GetLibInfo when length > minPathPartLength", func() {
			fileName, handlerName := GetLibInfo("/tmp", "test.example.init")
			convey.So(fileName, convey.ShouldEqual, "/tmp/test/example.so")
			convey.So(handlerName, convey.ShouldEqual, "init")
		})
	})
}

func TestDealEnv(t *testing.T) {
	convey.Convey("Test DealEnv", t, func() {
		convey.Convey("DealEnv when delegateDecrypt  == \"\"", func() {
			err, envMap := DealEnv()
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(envMap), convey.ShouldBeZeroValue)
		})
		os.Setenv("ENV_DELEGATE_DECRYPT", "delegateDecrypt1")
		convey.Convey("DealEnv when json.Unmarshal delegateDecrypt error", func() {
			err, envMap := DealEnv()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(envMap, convey.ShouldBeNil)
		})
		os.Setenv("ENV_DELEGATE_DECRYPT", `{"environment":"env1"}`)
		convey.Convey("DealEnv when json.Unmarshal environment error", func() {
			err, envMap := DealEnv()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(envMap, convey.ShouldBeNil)
		})
		str := `{"environment":"{\"env1\":\"value1\"}","encrypted_user_data":"data1"}`
		os.Setenv("ENV_DELEGATE_DECRYPT", str)
		convey.Convey("DealEnv when json.Unmarshal encrypted_user_data error", func() {
			err, envMap := DealEnv()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(envMap, convey.ShouldBeNil)
		})
		str = `{"environment":"{\"env1\":\"value1\"}","encrypted_user_data":"{\"env2\":\"value2\"}"}`
		os.Setenv("ENV_DELEGATE_DECRYPT", str)
		convey.Convey("DealEnv success", func() {
			err, envMap := DealEnv()
			convey.So(err, convey.ShouldBeNil)
			convey.So(envMap, convey.ShouldNotBeNil)
		})
	})
}

func TestContainsConnRefusedErr(t *testing.T) {
	convey.Convey("Test ContainsConnRefusedErr", t, func() {
		convey.Convey("ContainsConnRefusedErr success", func() {
			flag := ContainsConnRefusedErr(errors.New("connection refused"))
			convey.So(flag, convey.ShouldBeTrue)
		})
	})
}
