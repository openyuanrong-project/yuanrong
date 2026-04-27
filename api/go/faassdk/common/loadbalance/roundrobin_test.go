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
	"errors"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
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

func TestAdd(t *testing.T) {
	convey.Convey("add", t, func() {
		weightLvs := &WeightLvs{"Node1", 10}
		var node []*WeightLvs
		node = append(node, weightLvs)
		wlvs := &WLVS{node, 10, 0, 10, 10}
		wlvs.Add("Node2", 20)
		convey.So(wlvs.gcd, convey.ShouldNotEqual, 20)
		convey.So(wlvs.maxW, convey.ShouldEqual, 20)

		wlvs.gcd = 0
		wlvs.Add("Node2", 20)
		convey.So(wlvs.gcd, convey.ShouldEqual, 20)
		convey.So(wlvs.maxW, convey.ShouldEqual, 20)
	})
}

func TestRemoveAll(t *testing.T) {
	weightLvs := &WeightLvs{"Node1", 10}
	var node []*WeightLvs
	node = append(node, weightLvs)
	wlvs := &WLVS{node, 10, 0, 10, 10}
	wlvs.RemoveAll()
	if wlvs.maxW != 0 || wlvs.gcd != 0 || wlvs.cw != 0 || wlvs.i != -1 {
		err := errors.New("Test_RemoveAll error")
		t.Fatal(err)
	}
}

func TestWLVSReset(t *testing.T) {
	weightLvs := &WeightLvs{"Node1", 10}
	var node []*WeightLvs
	node = append(node, weightLvs)
	wlvs := &WLVS{node, 10, 0, 10, 10}
	wlvs.Reset()
	if wlvs.cw != 0 || wlvs.i != -1 {
		err := errors.New("Test_WLVSReset error")
		t.Fatal(err)
	}

}

func TestWLVSNext(t *testing.T) {
	convey.Convey("node length is 0", t, func() {
		var node []*WeightLvs
		wlvs := &WLVS{node, 0, 0, 10, 10}

		res := wlvs.Next("", true)
		convey.So(res, convey.ShouldBeNil)
	})

	convey.Convey("add", t, func() {
		node := []*WeightLvs{
			{"Node1", 10},
			{"Node2", 20},
		}
		wlvs := &WLVS{node, 0, 0, 10, 10}

		res := wlvs.Next("", true)
		convey.So(res, convey.ShouldNotBeNil)

		wlvss := &WLVS{node, 1, 0, 10, 10}
		res = wlvss.Next("", true)
		resStr, ok := res.(string)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(resStr, convey.ShouldEqual, "Node2")
	})

	convey.Convey("remove all", t, func() {
		node := []*WeightLvs{
			{"Node1", 10},
		}
		wlvss := &WLVS{node, 1, 0, 10, 10}
		res := wlvss.Next("", true)
		resStr, ok := res.(string)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(resStr, convey.ShouldEqual, "Node1")

		wlvss.RemoveAll()
		res = wlvss.Next("", true)
		resStr, ok = res.(string)
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(resStr, convey.ShouldEqual, "")
	})
	convey.Convey("remove", t, func() {
		node := []*WeightLvs{
			{"Node1", 10},
		}
		wlvss := &WLVS{node, 1, 0, 10, 10}
		res := wlvss.Next("", true)
		resStr, ok := res.(string)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(resStr, convey.ShouldEqual, "Node1")

		wlvss.Remove(node[0].Node)
		res = wlvss.Next("", true)
		resStr, ok = res.(string)
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(resStr, convey.ShouldEqual, "")
	})
}

func Test_gcd(t *testing.T) {
	cases := []struct {
		x    int
		y    int
		want int
	}{
		{9, 0, 9},
		{0, 8, 8},
		{1, 1, 1},
		{2, 9, 1},
		{12, 18, 6},
		{60, 140, 20},
		{3, 5, 1},
		{111, 111, 111},
	}
	for _, tt := range cases {
		assert.Equal(t, tt.want, gcd(tt.x, tt.y))
	}
}

func Test_WNGINX(t *testing.T) {
	w := &WNGINX{}
	w.Done("node")
	request := w.NextWithRequest(&Request{}, false)
	assert.Equal(t, request, nil)
	w.SetConcurrency(0)
	w.Start()
	w.Stop()
	lock := w.NoLock()
	assert.Equal(t, lock, false)
}

func Test_WLVS(t *testing.T) {
	w := &WLVS{}
	w.Done("node")
	request := w.NextWithRequest(&Request{}, false)
	assert.Equal(t, request, nil)
	w.SetConcurrency(0)
	w.Start()
	w.Stop()
	lock := w.NoLock()
	assert.Equal(t, lock, false)
}

func TestNextWeightedNode(t *testing.T) {
	convey.Convey(
		"Test nextWeightedNode", t, func() {
			convey.Convey(
				"nextWeightedNode success", func() {
					wn := nextWeightedNode([]*WeightNginx{})
					convey.So(wn, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestUpdateCwAnsIsReturn(t *testing.T) {
	convey.Convey(
		"Test updateCwAnsIsReturn", t, func() {
			convey.Convey(
				"updateCwAnsIsReturn success", func() {
					weightLvs := &WeightLvs{"Node1", 10}
					var node []*WeightLvs
					node = append(node, weightLvs)
					wlvs := &WLVS{node, 10, 0, 10, 10}

					flag := wlvs.updateCwAnsIsReturn()
					convey.So(flag, convey.ShouldBeTrue)
				},
			)
		},
	)
}

func TestRandomRR(t *testing.T) {
	convey.Convey(
		"Test RandomRR", t, func() {
			convey.Convey(
				"RandomRR success", func() {
					convey.So(func() {
						r := NewRandomRR()
						r.Reset()
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}
