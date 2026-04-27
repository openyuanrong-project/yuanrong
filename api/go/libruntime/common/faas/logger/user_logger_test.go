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

// Package logger for printing go runtime logger
package logger

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestSetupUserLogger(t *testing.T) {
	convey.Convey(
		"Test SetupUserLogger", t, func() {
			convey.Convey(
				"SetupUserLogger success", func() {
					err := SetupUserLogger("")
					convey.So(err, convey.ShouldBeNil)
					err = SetupUserLogger("1")
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestUserFunctionLogger(t *testing.T) {
	convey.Convey(
		"Test UserFunctionLogger ", t, func() {
			u := GetUserLogger()
			convey.Convey(
				"UserFunctionLogger success", func() {
					convey.So(u, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"Infof success", func() {
					convey.So(func() {
						u.Infof("format", "params")
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"Debugf success", func() {
					convey.So(func() {
						u.Debugf("format", "params")
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"Warnf success", func() {
					convey.So(func() {
						u.Warnf("format", "params")
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"Errorf success", func() {
					convey.So(func() {
						u.Errorf("format", "params")
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}
