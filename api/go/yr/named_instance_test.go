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

func TestNewNamedInstance(t *testing.T) {
	convey.Convey(
		"Test NamedInstance", t, func() {
			instance := NewNamedInstance("")
			convey.Convey(
				"NewNamedInstance success", func() {
					convey.So(instance, convey.ShouldNotBeNil)
				},
			)
			buf := new(bytes.Buffer)
			convey.Convey(
				"EncodeMsgpack success", func() {
					enc := msgpack.NewEncoder(buf)
					err := instance.EncodeMsgpack(enc)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"DecodeMsgpack success", func() {
					dec := msgpack.NewDecoder(buf)
					err := instance.DecodeMsgpack(dec)
					convey.So(err.Error(), convey.ShouldEqual, "EOF")
				},
			)
			convey.Convey(
				"Function success", func() {
					ifh := instance.Function((*CounterTest).Add)
					convey.So(ifh, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"Terminate success", func() {
					err := instance.Terminate()
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}
