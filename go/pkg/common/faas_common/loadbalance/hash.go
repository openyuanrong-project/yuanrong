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
	"hash/crc32"
	"sort"
	"sync"
	"time"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

const (
	// MaxInstanceSize is the max instance size be stored in hash ring
	MaxInstanceSize = 100
	defaultMapSize  = 100
)

type uint32Slice []uint32

// Len returns the size
func (u uint32Slice) Len() int {
	return len(u)
}

// Swap will swap two elements
func (u uint32Slice) Swap(i, j int) {
	if i < 0 || i >= len(u) || j < 0 || j >= len(u) {
		return
	}
	u[i], u[j] = u[j], u[i]
}

// Less returns true if i less than j
func (u uint32Slice) Less(i, j int) bool {
	if i < 0 || i >= len(u) || j < 0 || j >= len(u) {
		return false
	}
	return u[i] < u[j]
}

// SimpleCHGeneric is the simple generic consistent hash
type SimpleCHGeneric struct {
	instanceMap map[uint32]string
	hashPool    uint32Slice
	insMutex    sync.RWMutex
}

// NewSimpleCHGeneric creates generic consistent hash
func NewSimpleCHGeneric() *SimpleCHGeneric {
	return &SimpleCHGeneric{
		hashPool:    make([]uint32, 0, MaxInstanceSize),
		instanceMap: make(map[uint32]string, defaultMapSize),
	}
}
func (c *SimpleCHGeneric) getNextHashKey(hashKey uint32) uint32 {
	// need to be called with insMutex locked
	if len(c.hashPool) == 0 {
		return 0
	}
	nextHashKey := c.hashPool[0]
	for _, v := range c.hashPool {
		if v > hashKey {
			nextHashKey = v
			break
		}
	}
	return nextHashKey
}

// Next returns the next scheduled node of a function, move is disable
func (c *SimpleCHGeneric) Next(name string, move bool) interface{} {
	// need to be called in a thread safe context
	hashKey := getHashKeyCRC32([]byte(name))
	c.insMutex.RLock()
	instanceHash := c.getNextHashKey(hashKey)
	instanceKey, exist := c.instanceMap[instanceHash]
	c.insMutex.RUnlock()
	// check if node still exists, no maxReqCount limitation
	if !exist {
		return ""
	}
	return instanceKey
}

// Add will add a node into hash ring
func (c *SimpleCHGeneric) Add(node interface{}, weight int) {
	c.insMutex.Lock()
	defer c.insMutex.Unlock()
	name, ok := node.(string)
	if !ok {
		log.GetLogger().Errorf("unable to convert %T to string", node)
		return
	}
	hashKey := getHashKeyCRC32([]byte(name))
	_, exist := c.instanceMap[hashKey]
	if exist {
		return
	}
	c.instanceMap[hashKey] = name
	c.hashPool = append(c.hashPool, hashKey)
	sort.Sort(c.hashPool)
	log.GetLogger().Debugf("add node %s, hashKey %d to hash ring, hashPool is %v", name, hashKey, c.hashPool)
}

// Remove will remove a node from hash ring
func (c *SimpleCHGeneric) Remove(node interface{}) {
	name, assertOK := node.(string)
	if !assertOK {
		log.GetLogger().Errorf("unable to convert %T to string", node)
		return
	}
	hashKey := getHashKeyCRC32([]byte(name))
	c.insMutex.Lock()
	delete(c.instanceMap, hashKey)
	for i, hash := range c.hashPool {
		if hash == hashKey {
			copy(c.hashPool[i:], c.hashPool[i+1:])
			c.hashPool[len(c.hashPool)-1] = 0
			c.hashPool = c.hashPool[:len(c.hashPool)-1]
			break
		}
	}
	log.GetLogger().Infof("delete node %s from hash ring", name)
	c.insMutex.Unlock()

}

// RemoveAll will remove all nodes from hash ring
func (c *SimpleCHGeneric) RemoveAll() {
	c.insMutex.Lock()
	c.hashPool = make([]uint32, 0, MaxInstanceSize)
	c.instanceMap = make(map[uint32]string, defaultMapSize)
	c.insMutex.Unlock()
	return
}

// Reset will clean all anchor infos
func (c *SimpleCHGeneric) Reset() {
	return
}

// DeleteBalancer -
func (c *SimpleCHGeneric) DeleteBalancer(name string) {

}

type anchorInfo struct {
	instanceHash uint32
	instanceKey  string
}

// CHGeneric is the generic consistent hash
type CHGeneric struct {
	anchorPoint map[string]*anchorInfo
	instanceMap map[uint32]string
	hashPool    uint32Slice
	insMutex    sync.RWMutex
	anchorMutex sync.Mutex
}

// NewCHGeneric creates generic consistent hash
func NewCHGeneric() *CHGeneric {
	return &CHGeneric{
		hashPool:    make([]uint32, 0, MaxInstanceSize),
		instanceMap: make(map[uint32]string, defaultMapSize),
		anchorPoint: make(map[string]*anchorInfo, defaultMapSize),
	}
}

// Next returns the next scheduled node of a function
func (c *CHGeneric) Next(name string, move bool) interface{} {
	c.anchorMutex.Lock()
	anchor, exist := c.anchorPoint[name]
	if !exist {
		anchor = c.addAnchorPoint(name)
		c.anchorMutex.Unlock()
		return anchor.instanceKey
	}
	if move {
		c.moveAnchorPoint(name, anchor.instanceHash)
	}
	c.insMutex.RLock()
	_, exist = c.instanceMap[anchor.instanceHash]
	c.insMutex.RUnlock()
	// check if node still exists, no maxReqCount limitation
	if !exist {
		c.moveAnchorPoint(name, anchor.instanceHash)
	}
	c.anchorMutex.Unlock()
	return anchor.instanceKey
}

// Previous - returns the previous scheduled node of a function
func (c *CHGeneric) Previous(name string, move bool) interface{} {
	previous := c.getPreviousHashKey(getHashKeyCRC32([]byte(name)))
	if move {
		previous = c.getPreviousHashKey(previous)
	}
	c.insMutex.RLock()
	_, exist := c.instanceMap[previous]
	c.insMutex.RUnlock()
	if !exist {
		previous = c.getPreviousHashKey(previous)
	}
	return c.instanceMap[previous]
}

// Add will add a node into hash ring
func (c *CHGeneric) Add(node interface{}, weight int) {
	c.insMutex.Lock()
	defer c.insMutex.Unlock()
	name, ok := node.(string)
	if !ok {
		log.GetLogger().Errorf("unable to convert %T to string", node)
		return
	}
	hashKey := getHashKeyCRC32([]byte(name))
	_, exist := c.instanceMap[hashKey]
	if exist {
		return
	}
	c.instanceMap[hashKey] = name
	c.hashPool = append(c.hashPool, hashKey)
	sort.Sort(c.hashPool)
	log.GetLogger().Infof("add node %s, hashKey %d to hash ring, hashPool is %v", name, hashKey, c.hashPool)
}

// Remove will remove a node from hash ring
func (c *CHGeneric) Remove(node interface{}) {
	name, assertOK := node.(string)
	if !assertOK {
		log.GetLogger().Errorf("unable to convert %T to string", node)
		return
	}
	hashKey := getHashKeyCRC32([]byte(name))
	c.insMutex.Lock()
	delete(c.instanceMap, hashKey)
	for i, hash := range c.hashPool {
		if hash == hashKey {
			copy(c.hashPool[i:], c.hashPool[i+1:])
			c.hashPool[len(c.hashPool)-1] = 0
			c.hashPool = c.hashPool[:len(c.hashPool)-1]
			break
		}
	}
	log.GetLogger().Infof("delete node %s from hash ring", name)
	c.insMutex.Unlock()

}

// RemoveAll will remove all nodes from hash ring
func (c *CHGeneric) RemoveAll() {
	c.insMutex.Lock()
	c.hashPool = make([]uint32, 0, MaxInstanceSize)
	c.instanceMap = make(map[uint32]string, defaultMapSize)
	c.insMutex.Unlock()
	return
}

// Reset will clean all anchor infos
func (c *CHGeneric) Reset() {
	c.anchorMutex.Lock()
	c.anchorPoint = make(map[string]*anchorInfo, defaultMapSize)
	c.anchorMutex.Unlock()
	log.GetLogger().Infof("reset hash ring anchorPoint")
	return
}

// DeleteBalancer -
func (c *CHGeneric) DeleteBalancer(name string) {
	c.anchorMutex.Lock()
	defer c.anchorMutex.Unlock()
	delete(c.anchorPoint, name)
}

func (c *CHGeneric) addAnchorPoint(name string) *anchorInfo {
	// need to be called in a thread safe context
	hashKey := getHashKeyCRC32([]byte(name))
	c.insMutex.RLock()
	instanceHash := c.getNextHashKey(hashKey)
	c.insMutex.RUnlock()
	newAnchor := &anchorInfo{
		instanceHash: instanceHash,
		instanceKey:  c.instanceMap[instanceHash],
	}
	c.anchorPoint[name] = newAnchor
	log.GetLogger().Debugf("name %s hashKey %d", name, hashKey)
	return newAnchor
}

func (c *CHGeneric) moveAnchorPoint(name string, curHash uint32) {
	c.insMutex.Lock()
	instanceHash := c.getNextHashKey(curHash)
	c.anchorPoint[name].instanceHash = instanceHash
	c.anchorPoint[name].instanceKey = c.instanceMap[instanceHash]
	c.insMutex.Unlock()
}

func (c *CHGeneric) getNextHashKey(hashKey uint32) uint32 {
	// need to be called with insMutex locked
	if len(c.hashPool) == 0 {
		return 0
	}
	nextHashKey := c.hashPool[0]
	for _, v := range c.hashPool {
		if v > hashKey {
			nextHashKey = v
			break
		}
	}
	return nextHashKey
}

func (c *CHGeneric) getPreviousHashKey(hashKey uint32) uint32 {
	// need to be called with insMutex locked
	if len(c.hashPool) == 0 {
		return 0
	}
	hashLen := len(c.hashPool)
	previousHashKey := c.hashPool[hashLen-1]
	for i := hashLen - 1; i >= 0; i-- {
		if c.hashPool[i] < hashKey {
			previousHashKey = c.hashPool[i]
			break
		}
	}
	return previousHashKey
}

func getHashKeyCRC32(key []byte) uint32 {
	return crc32.ChecksumIEEE(key)
}

// NewLimiterCHGeneric return limiterCHGeneric with given concurrency
func NewLimiterCHGeneric(limiterTime time.Duration) *LimiterCHGeneric {
	return &LimiterCHGeneric{
		CHGeneric:   NewCHGeneric(),
		limiterTime: limiterTime,
		limiter:     make(map[string]*concurrentLimiter, constant.DefaultMapSize),
	}
}

type concurrentLimiter struct {
	head *limiterNode
}

type limiterNode struct {
	instanceKey interface{}
	lastTime    time.Time
	next        *limiterNode
}

// LimiterCHGeneric is limiter balanced
type LimiterCHGeneric struct {
	*CHGeneric
	limiter      map[string]*concurrentLimiter
	nodeCount    int
	limiterMutex sync.Mutex
	limiterTime  time.Duration
}

// Next returns the next scheduled node
func (c *LimiterCHGeneric) Next(name string, move bool) interface{} {
	c.limiterMutex.Lock()
	defer c.limiterMutex.Unlock()
	if _, ok := c.limiter[name]; !ok {
		c.limiter[name] = &concurrentLimiter{
			head: &limiterNode{},
		}
	}

	moveFlag := move
label:
	for exitFlag := 0; exitFlag <= c.nodeCount; exitFlag++ {
		instanceKey := c.CHGeneric.Next(name, moveFlag)
		h := c.limiter[name].head
		n := h.next
		for ; n != nil; n = n.next {
			if n.instanceKey == instanceKey && !n.lastTime.IsZero() && time.Now().Sub(n.lastTime) < c.limiterTime {
				moveFlag = true
				continue label
			}
			if n.instanceKey == instanceKey && (n.lastTime.IsZero() || time.Now().Sub(n.lastTime) >= c.limiterTime) {
				break
			}
		}
		if n == nil {
			h.next = &limiterNode{
				instanceKey: instanceKey,
				next:        h.next,
			}
		}
		return instanceKey
	}
	return nil
}

// Previous - returns the previous scheduled node of a function
func (c *LimiterCHGeneric) Previous(name string, move bool) interface{} {
	return nil
}

// Add a node to hash ring
func (c *LimiterCHGeneric) Add(node interface{}, weight int) {
	c.limiterMutex.Lock()
	defer c.limiterMutex.Unlock()
	c.nodeCount++
	c.CHGeneric.Add(node, weight)
}

// Remove a node from hash ring
func (c *LimiterCHGeneric) Remove(node interface{}) {
	c.limiterMutex.Lock()
	defer c.limiterMutex.Unlock()
	c.nodeCount--
	c.CHGeneric.Remove(node)
}

// RemoveAll remove all nodes from hash ring
func (c *LimiterCHGeneric) RemoveAll() {
	c.limiterMutex.Lock()
	defer c.limiterMutex.Unlock()
	c.nodeCount = 0
	c.CHGeneric.RemoveAll()
}

// Reset clean all anchor infos and counters
func (c *LimiterCHGeneric) Reset() {
	c.limiterMutex.Lock()
	defer c.limiterMutex.Unlock()
	c.limiter = make(map[string]*concurrentLimiter, constant.DefaultMapSize)
	c.CHGeneric.Reset()
}

// DeleteBalancer -
func (c *LimiterCHGeneric) DeleteBalancer(name string) {
	c.limiterMutex.Lock()
	defer c.limiterMutex.Unlock()
	c.CHGeneric.DeleteBalancer(name)
	delete(c.limiter, name)
}

// SetStain give the specified function, specify the node to set the stain
func (c *LimiterCHGeneric) SetStain(function string, node interface{}) {
	c.limiterMutex.Lock()
	defer c.limiterMutex.Unlock()
	if _, ok := c.limiter[function]; !ok {
		return
	}
	n := c.limiter[function].head
	for ; n != nil; n = n.next {
		if n.instanceKey == node {
			n.lastTime = time.Now()
			return
		}
	}
}
