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
package tokentosecret

import (
	"fmt"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/runtime/libruntime/common"
)

func TestSecretMgr(t *testing.T) {
	convey.Convey("Test token Mgr: GetToken", t, func() {
		ch := make(chan struct{})
		GetSecretMgr().SetAuthContext(&common.AuthContext{
			ServerAuthEnable:   false,
			RootCertData:       nil,
			ModuleCertData:     nil,
			ModuleKeyData:      nil,
			Token:              "fakeToken",
			Salt:               "134134134314134",
			EnableServerMode:   false,
			ServerNameOverride: "",
		})
		tk := GetSecretMgr().token
		convey.So(tk, convey.ShouldEqual, "fakeToken")
		close(ch)
	})
}

func TestSecretMgr_UpdateToken_GetSk(t *testing.T) {
	convey.Convey("Test SecretMgr TestSecretMgr_UpdateToken_GetSk", t, func() {
		timeStamp := time.Now().Second() + 5 // 5秒后过期
		authContext := &common.AuthContext{
			Token: fmt.Sprintf("%d_%s", timeStamp, "token0"),
			Salt:  "134134134",
		}

		GetSecretMgr().SetAuthContext(authContext)
		//convey.So(GetSecretMgr().expireTime, convey.ShouldEqual, timeStamp)
		convey.So(GetSecretMgr().token, convey.ShouldEqual, fmt.Sprintf("%d_%s", timeStamp, "token0"))
		convey.So(GetSecretMgr().salt, convey.ShouldEqual, "134134134")
		convey.So(GetSecretMgr().sk, convey.ShouldNotBeNil)
		convey.So(GetSecretMgr().GetSk(), convey.ShouldNotBeNil)
	})
}
