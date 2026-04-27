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

// Package common for token
package common

import (
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

func TestTokenMgr(t *testing.T) {
	convey.Convey("Test token Mgr: GetToken", t, func() {
		ch := make(chan struct{})
		t := GetTokenMgr()
		t.callback = func(ctx *AuthContext) {}
		t.UpdateToken(&AuthContext{
			ServerAuthEnable:   false,
			RootCertData:       nil,
			ModuleCertData:     nil,
			ModuleKeyData:      nil,
			Token:              "fakeToken",
			Salt:               "134134134314134",
			EnableServerMode:   false,
			ServerNameOverride: "",
		})
		tk := GetTokenMgr().GetToken()
		convey.So(tk, convey.ShouldEqual, "fakeToken")
		GetTokenMgr().UpdateToken(nil)
		close(ch)
	})
}

func TestSetCallback(t *testing.T) {
	convey.Convey(
		"Test SetCallback", t, func() {
			convey.Convey(
				"SetCallback success", func() {
					f := func(ctx *AuthContext) {}
					convey.So(func() {
						GetTokenMgr().SetCallback(f)
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestUpdateTokenLoop(t *testing.T) {
	convey.Convey(
		"Test UpdateTokenLoop", t, func() {
			convey.Convey(
				"UpdateTokenLoop success", func() {
					ch := WaitForSignal()
					convey.So(func() {
						go GetTokenMgr().UpdateTokenLoop(ch)
						time.Sleep(300 * time.Millisecond)
						close(stopCh)
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}
