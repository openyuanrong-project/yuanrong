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

// Package serviceaccount sign http request by jwttoken
package serviceaccount

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestCipherSuitesFromName(t *testing.T) {
	convey.Convey("Test cipherSuitesFromName", t, func() {
		convey.Convey("success", func() {
			cipherSuitesArr := []string{"TLS_AES_128_GCM_SHA256", "TLS_AES_256_GCM_SHA384"}
			tlsSuite := cipherSuitesID(cipherSuitesFromName(cipherSuitesArr))
			convey.So(len(tlsSuite), convey.ShouldEqual, 2)
		})
	})
}
