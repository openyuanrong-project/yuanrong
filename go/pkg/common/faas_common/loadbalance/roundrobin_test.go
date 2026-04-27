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

package loadbalance

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestNext(t *testing.T) {
	convey.Convey("node length is 0", t, func() {
		node := []*WeightNginx{}
		wnginx := WNGINX{node}

		res := wnginx.Next("", true)
		convey.So(res, convey.ShouldBeNil)
	})
	convey.Convey("node length is 1", t, func() {
		node := []*WeightNginx{
			{"Node1", 30, 10, 20},
		}
		wnginx := WNGINX{node}

		res := wnginx.Next("", true)
		convey.So(res, convey.ShouldNotBeNil)
	})
	convey.Convey("node length > 1", t, func() {
		node := []*WeightNginx{
			{"Node1", 30, 10, 20},
			{"Node2", 30, 60, 20},
		}
		wnginx := WNGINX{node}

		res := wnginx.Next("", true)
		resStr, ok := res.(string)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(resStr, convey.ShouldEqual, "Node2")
	})

	convey.Convey("remove", t, func() {
		node := []*WeightNginx{
			{"Node1", 30, 10, 20},
		}
		wnginx := WNGINX{node}
		wnginx.Add("Node2", 60)
		res := wnginx.Next("", true)
		resStr, ok := res.(string)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(resStr, convey.ShouldEqual, "Node2")

		wnginx.Remove("Node2")
		res = wnginx.Next("", true)
		resStr, ok = res.(string)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(resStr, convey.ShouldEqual, "Node1")
	})

	convey.Convey("remove", t, func() {
		node := []*WeightNginx{
			{"Node1", 30, 10, 20},
			{"Node2", 30, 60, 20},
		}
		wnginx := WNGINX{node}
		wnginx.RemoveAll()
		convey.So(len(wnginx.nodes), convey.ShouldEqual, 0)
	})
}

func TestReset(t *testing.T) {
	convey.Convey("Reset success", t, func() {
		weightNginx := &WeightNginx{"Node1", 30, 10, 20}
		var node []*WeightNginx
		node = append(node, weightNginx)
		wnginx := WNGINX{node}

		wnginx.Reset()
		convey.So(weightNginx.EffectiveWeight, convey.ShouldEqual, weightNginx.Weight)
	})

}
