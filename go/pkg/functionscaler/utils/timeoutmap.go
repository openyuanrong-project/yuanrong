/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
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

// Package utils -
package utils

import (
	"sync"
	"time"
)

// TimeoutMap define map, key will be deleted after timeout
type TimeoutMap struct {
	mu sync.RWMutex
	m  map[string]*item
}

type item struct {
	value      interface{}
	expireTime time.Time
}

// NewTimeoutMap create timeoutMap
func NewTimeoutMap(cleanupInterval time.Duration) *TimeoutMap {
	tm := &TimeoutMap{
		m: make(map[string]*item),
	}

	go tm.startCleanup(cleanupInterval)
	return tm
}

// Set add key to map
func (tm *TimeoutMap) Set(key string, value interface{}, ttl time.Duration) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.m[key] = &item{
		value:      value,
		expireTime: time.Now().Add(ttl),
	}
}

// Get from map
func (tm *TimeoutMap) Get(key string) (interface{}, bool) {
	tm.mu.RLock()
	item, exists := tm.m[key]
	tm.mu.RUnlock()

	if !exists {
		return nil, false
	}

	if time.Now().After(item.expireTime) {
		tm.mu.Lock()
		delete(tm.m, key)
		tm.mu.Unlock()
		return nil, false
	}

	return item.value, true
}

// Delete from map
func (tm *TimeoutMap) Delete(key string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	delete(tm.m, key)
}

func (tm *TimeoutMap) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		tm.cleanup()
	}
}

func (tm *TimeoutMap) cleanup() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	now := time.Now()
	for key, item := range tm.m {
		if now.After(item.expireTime) {
			delete(tm.m, key)
		}
	}
}
