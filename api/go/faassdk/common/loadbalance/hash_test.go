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

// Package loadbalance provides consistent hash alogrithm
package loadbalance

import (
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

func TestUint32Slice(t *testing.T) {
	convey.Convey(
		"Test uint32Slice", t, func() {
			var u uint32Slice = []uint32{1, 2}
			convey.Convey(
				"Swap success", func() {
					convey.So(func() {
						u.Swap(-1, 0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"Less success", func() {
					flag := u.Less(-1, 0)
					convey.So(flag, convey.ShouldBeFalse)
				},
			)
		},
	)
}

func TestRemove(t *testing.T) {
	convey.Convey(
		"Test Remove", t, func() {
			convey.Convey(
				"Remove success", func() {
					c := &CHGeneric{}
					convey.So(func() {
						c.Remove(1.0)
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestConcurrentCHGenericNext(t *testing.T) {
	convey.Convey(
		"Test ConcurrentCHGenericNext", t, func() {
			convey.Convey(
				"ConcurrentCHGenericNext success", func() {
					c := NewConcurrentCHGeneric(0)
					c.counter["name"] = &concurrentCounter{
						count: 1,
					}
					res := c.Next("name", false)
					convey.So(res, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestLimiterCHGenericRemoveAll(t *testing.T) {
	convey.Convey(
		"Test LimiterCHGenericRemoveAll", t, func() {
			convey.Convey(
				"LimiterCHGenericRemoveAll success", func() {
					c := NewLimiterCHGeneric(time.Second)
					convey.So(c.RemoveAll, convey.ShouldNotPanic)
				},
			)
		},
	)
}
