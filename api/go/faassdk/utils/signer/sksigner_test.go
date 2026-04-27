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

package signer

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestSign(t *testing.T) {
	convey.Convey("Sign", t, func() {
		value := Sign([]byte("abcdefgh"), []byte("afdhkajhfkjasdhfkjasdhfkjasdfh"))
		convey.So(value, convey.ShouldEqual, []byte{233, 43, 186, 45, 115, 41, 252, 5, 25, 20, 102, 213, 47, 18, 30, 161, 100, 248, 235, 212, 108, 19, 103, 58, 242, 244, 24, 54, 20, 176, 247, 230})
	})
}

func TestEncodeHex(t *testing.T) {
	convey.Convey("EncodeHex", t, func() {
		tests := []struct {
			Input  string
			Expect string
		}{{
			"afasdfadsfdasf",
			"6166617364666164736664617366",
		}, {
			"1234567890-=",
			"313233343536373839302d3d",
		}, {
			"1234567890-\n\t\r   \t",
			"313233343536373839302d0a090d20202009",
		},
		}
		for _, tt := range tests {
			convey.So(EncodeHex([]byte(tt.Input)), convey.ShouldEqual, tt.Expect)
		}
	})
}

func TestBuildAuthorization(t *testing.T) {
	convey.Convey("BuildAuthorization", t, func() {
		tests := []struct {
			Ak     string
			Ts     string
			Sign   string
			Expect string
		}{
			{
				Ak:     "tenatnId",
				Ts:     "123456789",
				Sign:   "adfasfasdfadsfadf",
				Expect: "SDK-HMAC-SHA256 accessId=tenatnId,timestamp=123456789,signature=adfasfasdfadsfadf",
			}, {
				Ak:     "",
				Ts:     "123456789",
				Sign:   "qerqreqwrqewr",
				Expect: "SDK-HMAC-SHA256 accessId=,timestamp=123456789,signature=qerqreqwrqewr",
			}, {
				Ak:     "tenatnId",
				Ts:     "",
				Sign:   "zvvzcvzcvc",
				Expect: "SDK-HMAC-SHA256 accessId=tenatnId,timestamp=,signature=zvvzcvzcvc",
			},
			{
				Ak:     "tenantId",
				Ts:     "",
				Sign:   "",
				Expect: "SDK-HMAC-SHA256 accessId=tenantId,timestamp=,signature=",
			},
		}

		for _, tt := range tests {
			convey.So(BuildAuthorization(tt.Ak, tt.Ts, tt.Sign), convey.ShouldEqual, tt.Expect)
		}
	})
}
