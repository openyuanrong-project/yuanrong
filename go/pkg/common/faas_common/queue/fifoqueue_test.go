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

// Package queue -
package queue

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestSetupLogger(t *testing.T) {
	fq := NewFifoQueue(nil)
	convey.Convey("front/back is nil", t, func() {
		res := fq.Front()
		convey.So(res, convey.ShouldBeNil)
		res = fq.Back()
		convey.So(res, convey.ShouldBeNil)
	})
	convey.Convey("do support by id", t, func() {
		res := fq.GetByID("test")
		convey.So(res, convey.ShouldBeNil)
		err := fq.UpdateObjByID("test", "test")
		convey.So(err, convey.ShouldEqual, ErrMethodUnsupported)
		err = fq.DelByID("test")
		convey.So(err, convey.ShouldEqual, ErrObjectNotFound)
	})
	convey.Convey("pushback one ele", t, func() {
		res := fq.PushBack("obj1")
		convey.So(res, convey.ShouldBeNil)
		front := fq.Front()
		back := fq.Back()
		convey.So(front, convey.ShouldEqual, back)
		len := fq.Len()
		convey.So(len, convey.ShouldEqual, 1)
	})
	convey.Convey("pushback other ele", t, func() {
		res := fq.PushBack("obj2")
		convey.So(res, convey.ShouldBeNil)
		len := fq.Len()
		convey.So(len, convey.ShouldEqual, 2)

		front := fq.PopFront()
		convey.So(front, convey.ShouldEqual, "obj1")
		back := fq.PopBack()
		convey.So(back, convey.ShouldEqual, "obj2")
	})
}
