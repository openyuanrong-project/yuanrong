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

// Package loadbalance provides consistent hash algorithm
package loadbalance

import (
	"github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestSimpleCHGeneric_Next(t *testing.T) {
	convey.Convey("SimpleCHGeneric next", t, func() {
		generic := NewSimpleCHGeneric()

		generic.Add("node1", 0)
		generic.Add("node2", 0)
		next1 := generic.Next("function1", false)
		next2 := generic.Next("function1", false)
		convey.So(next1, convey.ShouldResemble, next2)
	})
}

func TestCHGeneric_Previous(t *testing.T) {
	convey.Convey("CHGeneric previous", t, func() {
		generic := NewCHGeneric()
		generic.Add("node1", 0)
		generic.Add("node2", 0)
		generic.Add("node3", 0)

		previous := generic.Previous("node2", false)
		convey.So(previous, convey.ShouldEqual, "node1")

		previous = generic.Previous("node2", true)
		convey.So(previous, convey.ShouldEqual, "node3")
	})
}

func TestLimiterCHGeneric_DeleteBalancer(t *testing.T) {
	convey.Convey("LimiterCHGeneric_DeleteBalancer", t, func() {
		generic := NewLimiterCHGeneric(1 * time.Second)
		generic.Add("node1", 0)
		generic.Add("node2", 0)
		generic.Add("node3", 0)

		next1 := generic.Next("function1", false)
		convey.So(next1, convey.ShouldEqual, "node2")
		next2 := generic.Next("function2", false)
		convey.So(next2, convey.ShouldEqual, "node3")

		_, ok := generic.limiter["function1"]
		_, exist := generic.anchorPoint["function1"]
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(exist, convey.ShouldBeTrue)

		generic.DeleteBalancer("function1")
		_, ok = generic.limiter["function1"]
		_, exist = generic.anchorPoint["function1"]
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(exist, convey.ShouldBeFalse)
	})
}
