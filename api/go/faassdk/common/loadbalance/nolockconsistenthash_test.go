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

	"github.com/smartystreets/goconvey/convey"
)

const (
	nodeKey        = "faas-scheduler-6b758c8b74-5zdwv"
	funcKeyWithRes = "7e186a/0@base@testresourcepython36768/latest/300-128"
)

var (
	node1 = &Node{
		Key: nodeKey,
	}

	node2 = &Node{
		Key: nodeKey + "1",
	}

	node3 = &Node{
		Key: nodeKey + "2",
	}
)

type mockRealNode struct {
	state bool
}

func (node *mockRealNode) IsEnable() bool {
	return node.state
}

func TestStatefulConsistent(t *testing.T) {
	convey.Convey("TestStatefulConsistentHashWithOneNode", t, func() {
		lb := CreateNoLockLB()
		outNode := lb.Next(funcKeyWithRes)
		convey.So(outNode, convey.ShouldBeNil)

		lb.Add(node1)
		lb.Add(node1)

		outNode = lb.Next(funcKeyWithRes)
		convey.So(outNode, convey.ShouldNotBeNil)
		convey.So(outNode.Key, convey.ShouldEqual, node1.Key)

		outNode = lb.Delete(nodeKey)
		convey.So(outNode, convey.ShouldNotBeNil)
		convey.So(outNode.Key, convey.ShouldEqual, node1.Key)

		outNode = lb.Delete(nodeKey)
		convey.So(outNode, convey.ShouldBeNil)

		outNode = lb.Next(funcKeyWithRes)
		convey.So(outNode, convey.ShouldBeNil)

		lb.Add(node2)
		lb.Add(node3)
		lb.Add(node1)

		outNode = lb.Next(funcKeyWithRes)
		convey.So(outNode, convey.ShouldNotBeNil)
		convey.So(outNode.Key, convey.ShouldEqual, node2.Key)

	})
}

func BenchmarkStatefulConsistentHashWithThreeNode(b *testing.B) {
	lb := CreateNoLockLB()

	lb.Add(node1)
	lb.Add(node2)
	lb.Add(node3)

	for i := 0; i < b.N; i++ {
		lb.Next(funcKeyWithRes + "3")
	}
}

func TestNodeSlice(t *testing.T) {
	convey.Convey(
		"Test nodeSlice", t, func() {
			var s nodeSlice = []*Node{}
			convey.Convey(
				"Swap success", func() {
					convey.So(func() {
						s.Swap(-1, 0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"Less success", func() {
					flag := s.Less(-1, 0)
					convey.So(flag, convey.ShouldBeFalse)
				},
			)
		},
	)
}
