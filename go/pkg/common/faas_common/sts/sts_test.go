/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
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

// Package sts -
package sts

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/sts/raw"
)

func TestDecryptSystemAuthConfig(t *testing.T) {
	convey.Convey("Test DecryptSystemAuthConfig", t, func() {
		convey.Convey("when enableIam is failed", func() {
			auth := raw.Auth{
				EnableIam: "aaa",
			}
			res := DecryptSystemAuthConfig(auth)
			convey.So(res.AccessKey, convey.ShouldBeEmpty)
		})
		convey.Convey("when EnableIam is false", func() {
			auth := raw.Auth{
				EnableIam: "false",
			}
			res := DecryptSystemAuthConfig(auth)
			convey.So(res.EnableIam, convey.ShouldBeEmpty)
		})
		convey.Convey("when ak is empty", func() {
			auth := raw.Auth{
				EnableIam: "true",
			}
			res := DecryptSystemAuthConfig(auth)
			convey.So(res.AccessKey, convey.ShouldBeEmpty)
		})
		convey.Convey("when sk is empty", func() {
			auth := raw.Auth{
				EnableIam: "true",
				AccessKey: "ak",
			}
			res := DecryptSystemAuthConfig(auth)
			convey.So(res.AccessKey, convey.ShouldBeEmpty)
			convey.So(res.SecretKey, convey.ShouldBeEmpty)
		})
		convey.Convey("when dk is empty", func() {
			auth := raw.Auth{
				EnableIam: "true",
				AccessKey: "ak",
				SecretKey: "sk",
			}
			res := DecryptSystemAuthConfig(auth)
			convey.So(res.AccessKey, convey.ShouldBeEmpty)
			convey.So(res.SecretKey, convey.ShouldBeEmpty)
			convey.So(res.DataKey, convey.ShouldBeEmpty)
		})
		convey.Convey("when success", func() {
			auth := raw.Auth{
				EnableIam: "true",
				AccessKey: "ak",
				SecretKey: "sk",
				DataKey:   "dk",
			}
			res := DecryptSystemAuthConfig(auth)
			convey.So(res.AccessKey, convey.ShouldBeEmpty)
			convey.So(res.SecretKey, convey.ShouldBeEmpty)
			convey.So(res.DataKey, convey.ShouldBeEmpty)
		})
	})
}
