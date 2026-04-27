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

// Package context
package context

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestProvider(t *testing.T) {
	convey.Convey(
		"Test Provider", t, func() {
			p := &Provider{
				CtxEnv:      &Env{},
				CtxHTTPHead: &InvokeContext{},
			}
			convey.Convey(
				"GetRemainingTimeInMilliSeconds success", func() {
					i := p.GetRemainingTimeInMilliSeconds()
					convey.So(i, convey.ShouldBeZeroValue)
				},
			)
			convey.Convey(
				"GetFunctionName success", func() {
					str := p.GetFunctionName()
					convey.So(str, convey.ShouldBeEmpty)
				},
			)
			convey.Convey(
				"GetRunningTimeInSeconds success", func() {
					i := p.GetRunningTimeInSeconds()
					convey.So(i, convey.ShouldBeZeroValue)
				},
			)
			convey.Convey(
				"GetVersion success", func() {
					str := p.GetVersion()
					convey.So(str, convey.ShouldBeEmpty)
				},
			)
			convey.Convey(
				"GetMemorySize success", func() {
					i := p.GetMemorySize()
					convey.So(i, convey.ShouldBeZeroValue)
				},
			)
			convey.Convey(
				"GetCPUNumber success", func() {
					i := p.GetCPUNumber()
					convey.So(i, convey.ShouldBeZeroValue)
				},
			)
			convey.Convey(
				"GetUserData success", func() {
					str := p.GetUserData("key")
					convey.So(str, convey.ShouldBeEmpty)
				},
			)
			convey.Convey(
				"GetLogger success", func() {
					l := p.GetLogger()
					convey.So(l, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"GetProjectID success", func() {
					str := p.GetProjectID()
					convey.So(str, convey.ShouldBeEmpty)
				},
			)
			convey.Convey(
				"GetPackage success", func() {
					str := p.GetPackage()
					convey.So(str, convey.ShouldBeEmpty)
				},
			)
			convey.Convey(
				"GetAccessKey success", func() {
					str := p.GetAccessKey()
					convey.So(str, convey.ShouldBeEmpty)
				},
			)
			convey.Convey(
				"GetSecretKey success", func() {
					str := p.GetSecretKey()
					convey.So(str, convey.ShouldBeEmpty)
				},
			)
			convey.Convey(
				"GetToken success", func() {
					str := p.GetToken()
					convey.So(str, convey.ShouldBeEmpty)
				},
			)
			convey.Convey(
				"GetRequestID success", func() {
					str := p.GetRequestID()
					convey.So(str, convey.ShouldBeEmpty)
				},
			)
			convey.Convey(
				"GetState success", func() {
					str := p.GetState()
					convey.So(str, convey.ShouldBeEmpty)
				},
			)
			convey.Convey(
				"SetState success", func() {
					err := p.SetState("state")
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"GetInvokeProperty success", func() {
					str := p.GetInvokeProperty()
					convey.So(str, convey.ShouldBeEmpty)
				},
			)
			convey.Convey(
				"GetTraceID success", func() {
					str := p.GetTraceID()
					convey.So(str, convey.ShouldBeEmpty)
				},
			)
			convey.Convey(
				"GetInvokeID success", func() {
					str := p.GetInvokeID()
					convey.So(str, convey.ShouldBeEmpty)
				},
			)
			convey.Convey(
				"GetAlias success", func() {
					str := p.GetAlias()
					convey.So(str, convey.ShouldBeEmpty)
				},
			)
			convey.Convey(
				"GetSecurityToken success", func() {
					str := p.GetSecurityToken()
					convey.So(str, convey.ShouldBeEmpty)
				},
			)
		},
	)
}
