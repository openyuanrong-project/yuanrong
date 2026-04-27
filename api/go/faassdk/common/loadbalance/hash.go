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

	"yuanrong.org/kernel/runtime/faassdk/common/constants"
	"yuanrong.org/kernel/runtime/libruntime/common/faas/logger"
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
	c.insMutex.RLock()
	if move {
		c.moveAnchorPoint(name, anchor.instanceHash)
	}
	_, exist = c.instanceMap[anchor.instanceHash]
	c.insMutex.RUnlock()
	// check if node still exists, no maxReqCount limitation
	if !exist {
		c.moveAnchorPoint(name, anchor.instanceHash)
	}
	c.anchorMutex.Unlock()
	return anchor.instanceKey
}

// Add will add a node into hash ring
func (c *CHGeneric) Add(node interface{}, weight int) {
	c.insMutex.Lock()
	defer c.insMutex.Unlock()
	name, ok := node.(string)
	if !ok {
		logger.GetLogger().Errorf("unable to convert %T to string", node)
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
	logger.GetLogger().Infof("add node %s to hash ring", name)
}

// Remove will remove a node from hash ring
func (c *CHGeneric) Remove(node interface{}) {
	name, ok := node.(string)
	if !ok {
		logger.GetLogger().Errorf("unable to convert %T to string", node)
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
	logger.GetLogger().Infof("delete node %s from hash ring", name)
	c.insMutex.Unlock()
	return
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
	return
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
	return newAnchor
}

func (c *CHGeneric) moveAnchorPoint(name string, curHash uint32) {
	c.insMutex.RLock()
	instanceHash := c.getNextHashKey(curHash)
	c.anchorPoint[name].instanceHash = instanceHash
	c.anchorPoint[name].instanceKey = c.instanceMap[instanceHash]
	c.insMutex.RUnlock()
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

func getHashKeyCRC32(key []byte) uint32 {
	return crc32.ChecksumIEEE(key)
}

// NewConcurrentCHGeneric return ConcurrentCHGeneric with given concurrency
func NewConcurrentCHGeneric(concurrency int) *ConcurrentCHGeneric {
	return &ConcurrentCHGeneric{
		CHGeneric:   NewCHGeneric(),
		concurrency: concurrency,
		counter:     make(map[string]*concurrentCounter, constants.DefaultMapSize),
	}
}

type concurrentCounter struct {
	count int
	last  time.Time
}

// ConcurrentCHGeneric is concurrency balanced
type ConcurrentCHGeneric struct {
	*CHGeneric
	counter     map[string]*concurrentCounter
	countMutex  sync.Mutex
	concurrency int
}

// Next returns the next scheduled node
func (c *ConcurrentCHGeneric) Next(name string, move bool) interface{} {
	c.countMutex.Lock()
	defer c.countMutex.Unlock()
	l, ok := c.counter[name]
	if !ok {
		c.counter[name] = &concurrentCounter{
			last: time.Now(),
		}
		return c.CHGeneric.Next(name, move)
	}
	l.count++
	if l.count >= c.concurrency {
		now := time.Now()
		l.count = 0
		if now.Sub(l.last) < 1*time.Second {
			move = true
		}
		l.last = now
	}
	return c.CHGeneric.Next(name, move)
}

// Add a node to hash ring
func (c *ConcurrentCHGeneric) Add(node interface{}, weight int) {
	c.CHGeneric.Add(node, weight)
}

// Remove a node from hash ring
func (c *ConcurrentCHGeneric) Remove(node interface{}) {
	c.countMutex.Lock()
	defer c.countMutex.Unlock()
	delete(c.counter, node.(string))
	c.CHGeneric.Remove(node)
}

// RemoveAll remove all nodes from hash ring
func (c *ConcurrentCHGeneric) RemoveAll() {
	c.countMutex.Lock()
	defer c.countMutex.Unlock()
	c.counter = make(map[string]*concurrentCounter, constants.DefaultMapSize)
	c.CHGeneric.RemoveAll()
}

// Reset clean all anchor infos and counters
func (c *ConcurrentCHGeneric) Reset() {
	c.countMutex.Lock()
	defer c.countMutex.Unlock()
	c.counter = make(map[string]*concurrentCounter, constants.DefaultMapSize)
	c.CHGeneric.Reset()
}

// NewLimiterCHGeneric return limiterCHGeneric with given concurrency
func NewLimiterCHGeneric(limiterTime time.Duration) *LimiterCHGeneric {
	return &LimiterCHGeneric{
		CHGeneric:   NewCHGeneric(),
		limiterTime: limiterTime,
		limiter:     make(map[string]*concurrentLimiter, constants.DefaultMapSize),
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

// Add a node to hash ring
func (c *LimiterCHGeneric) Add(node interface{}, weight int) {
	c.nodeCount++
	c.CHGeneric.Add(node, weight)
}

// Remove a node from hash ring
func (c *LimiterCHGeneric) Remove(node interface{}) {
	c.nodeCount--
	c.CHGeneric.Remove(node)
}

// RemoveAll remove all nodes from hash ring
func (c *LimiterCHGeneric) RemoveAll() {
	c.nodeCount = 0
	c.CHGeneric.RemoveAll()
}

// Reset clean all anchor infos and counters
func (c *LimiterCHGeneric) Reset() {
	c.nodeCount = 0
	c.CHGeneric.Reset()
}

// SetStain give the specified function, specify the node to set the stain
func (c *LimiterCHGeneric) SetStain(function string, node interface{}) {
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
