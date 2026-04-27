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
	"errors"
	"sort"
)

// Node -
type Node struct {
	Obj  interface{}
	Key  string
	hash uint32
}

// NoLockLoadBalance -
type NoLockLoadBalance interface {
	Add(node *Node) error
	Next(key string) *Node
	Delete(nodeKey string) *Node
}

// CreateNoLockLB -
func CreateNoLockLB() NoLockLoadBalance {
	return &ConsistentHash{
		nodes: make([]*Node, 0),
		cache: createHashCache(),
	}
}

type nodeSlice []*Node

// Len returns the size
func (s nodeSlice) Len() int {
	return len(s)
}

// Swap will swap two elements
func (s nodeSlice) Swap(i, j int) {
	if i < 0 || i >= len(s) || j < 0 || j >= len(s) {
		return
	}
	s[i], s[j] = s[j], s[i]
}

// Less returns true if i less than j
func (s nodeSlice) Less(i, j int) bool {
	if i < 0 || i >= len(s) || j < 0 || j >= len(s) {
		return false
	}
	return s[i].hash < s[j].hash
}

// ConsistentHash -
type ConsistentHash struct {
	cache *hashCache
	nodes nodeSlice
}

// Add -
func (c *ConsistentHash) Add(newNode *Node) error {
	newNode.hash = getHashKeyCRC32([]byte(newNode.Key))
	for _, node := range c.nodes {
		if node.Key == newNode.Key {
			return errors.New("node already exist")
		}
		if node.hash == newNode.hash {
			return errors.New("node hash already exist")
		}
	}

	c.nodes = append(c.nodes, newNode)
	sort.Sort(c.nodes)
	return nil
}

// Next -
func (c *ConsistentHash) Next(key string) *Node {
	if len(c.nodes) == 0 {
		return nil
	}

	keyHash := c.cache.getHash(key)
	index := c.search(keyHash)
	return c.nodes[index]
}

func (c *ConsistentHash) search(keyHash uint32) int {
	f := func(x int) bool {
		if x >= len(c.nodes) {
			return false
		}
		return c.nodes[x].hash > keyHash
	}
	index := sort.Search(len(c.nodes), f)
	if index >= len(c.nodes) {
		return 0
	}
	return index
}

// Delete -
func (c *ConsistentHash) Delete(nodeKey string) *Node {
	for i, node := range c.nodes {
		if node.Key == nodeKey {
			c.nodes = append(c.nodes[:i], c.nodes[i+1:]...)
			return node
		}
	}
	return nil
}
