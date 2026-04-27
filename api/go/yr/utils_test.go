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
	"math"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/vmihailenco/msgpack/v5"
)

func PlusOne() int {
	return 1
}

type T struct{}

func (t *T) A() {}
func (t T) B()  {}

func TestIsFunction(t *testing.T) {
	if !IsFunction(PlusOne) {
		t.Errorf("PlusOne is a function, judge failed.")
	}

	if !IsFunction((*T).A) {
		t.Errorf("(*T).A is a function, judge failed.")
	}

	if !IsFunction((*T).B) {
		t.Errorf("(*T).B is a function, judge failed.")
	}

	if !IsFunction(T.B) {
		t.Errorf("T.B is a function, judge failed.")
	}

	a := func() {}
	if !IsFunction(a) {
		t.Errorf("a is a function, judge failed.")
	}

	b := 1
	if IsFunction(b) {
		t.Errorf("b is not a function, judge failed.")
	}
}

func TestPackInvokeArg(t *testing.T) {
	convey.Convey(
		"Test DeleteStream", t, func() {
			convey.Convey(
				"arg type: yr.ObjectRef success", func() {
					res, err := PackInvokeArg(ObjectRef{})
					convey.So(res, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)

			convey.Convey(
				"arg type: *yr.ObjectRef success", func() {
					objRef := ObjectRef{}
					res, err := PackInvokeArg(&objRef)
					convey.So(res, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)

			convey.Convey(
				"arg type: []yr.ObjectRef success", func() {
					res, err := PackInvokeArg([]ObjectRef{})
					convey.So(res, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)

			convey.Convey(
				"arg type: []*yr.ObjectRef success", func() {
					res, err := PackInvokeArg([]*ObjectRef{})
					convey.So(res, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)

			convey.Convey(
				"arg type: else success", func() {
					d := 122.0
					res, err := PackInvokeArg(d)
					convey.So(err, convey.ShouldBeNil)
					var f float64
					err = msgpack.Unmarshal(res.Data, &f)
					if err != nil {
						t.Errorf("msgpack.Unmarshal err: %s", err)
						return
					}
					convey.So(math.Abs(d-f), convey.ShouldBeLessThanOrEqualTo, 1e-9)
				},
			)
		},
	)
}

func TestPackInvokeArgs(t *testing.T) {
	convey.Convey(
		"Test PackInvokeArgs", t, func() {
			convey.Convey(
				"PackInvokeArgs success", func() {
					objRef := ObjectRef{}
					args := []any{objRef, &objRef, []ObjectRef{}, []*ObjectRef{}}
					res, err := PackInvokeArgs(args)
					convey.So(len(res), convey.ShouldNotBeZeroValue)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}
