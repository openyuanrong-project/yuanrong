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
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

type CounterTest struct {
	Count int
}

func (c *CounterTest) Add(x int) int {
	c.Count += x
	return c.Count
}

func NewCounter(init int) *CounterTest {
	return &CounterTest{Count: init}
}

func TestInstanceCreator(t *testing.T) {
	convey.Convey(
		"Test InstanceCreator", t, func() {
			convey.Convey(
				"InstanceCreator success when !IsFunction", func() {
					convey.So(func() {
						Instance("")
					}, convey.ShouldPanic)
				},
			)
			creator := Instance(NewCounter(1).Add)
			convey.Convey(
				"InstanceCreator success", func() {
					convey.So(creator, convey.ShouldNotBeNil)
				},
			)
			opts := new(InvokeOptions)
			newCreator := creator.Options(opts)
			convey.Convey(
				"Options success", func() {
					convey.So(newCreator, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"Invoke success", func() {
					convey.So(func() {
						newCreator.Invoke(2, 3)
					}, convey.ShouldPanic)
				},
			)
		},
	)
}
