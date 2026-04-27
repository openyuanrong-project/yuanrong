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

// Package yr for actor
package yr

import (
	"bytes"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/vmihailenco/msgpack"
)

func TestNewObjectRef(t *testing.T) {
	convey.Convey(
		"Test NewObjectRef", t, func() {
			convey.Convey(
				"NewObjectRef success", func() {
					objRef := NewObjectRef()
					convey.So(objRef, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestDecodeMsgpack(t *testing.T) {
	convey.Convey(
		"Test DecodeMsgpack", t, func() {
			convey.Convey(
				"DecodeMsgpack success", func() {
					objRef := NewObjectRef()
					r := bytes.NewReader([]byte{})
					err := objRef.DecodeMsgpack(msgpack.NewDecoder(r))
					convey.So(err, convey.ShouldNotBeNil)
				},
			)
		},
	)
}
