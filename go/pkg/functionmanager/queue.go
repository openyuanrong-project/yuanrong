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

// Package faasmanager -
package functionmanager

import (
	"sync"
)

// item item
type item interface {
	String() string
	Done(err error)
}

type set map[interface{}]struct{}

type queue struct {
	cond *sync.Cond

	// Waiting for the execution of the set
	dirty set

	// Executing set
	processing set

	queue []item

	shutDownFlag bool
}

func (s set) has(item string) bool {
	_, exists := s[item]
	return exists
}

func (s set) insert(item string) {
	s[item] = struct{}{}
}

func (s set) delete(item string) {
	delete(s, item)
}

func (q *queue) add(item item) bool {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	if q.shutDownFlag {
		return false
	}
	if q.dirty.has(item.String()) {
		return false
	}
	q.dirty.insert(item.String())
	if q.processing.has(item.String()) {
		return false
	}
	q.queue = append(q.queue, item)
	q.cond.Signal()
	return true
}

func (q *queue) len() int {
	q.cond.L.Lock()
	l := len(q.queue)
	q.cond.L.Unlock()
	return l
}

func (q *queue) getAll() ([]item, bool) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	for len(q.queue) == 0 && !q.shutDownFlag {
		q.cond.Wait()
	}
	if len(q.queue) == 0 {
		// We must be shutting down.
		return nil, true
	}
	items := q.queue
	q.queue = []item{}
	for _, v := range items {
		q.processing.insert(v.String())
		q.dirty.delete(v.String())
	}
	return items, false
}

func (q *queue) shutDown() {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	q.shutDownFlag = true
	q.cond.Broadcast()
}

func (q *queue) doneAll(items []item) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	for _, v := range items {
		q.processing.delete(v.String())
		if q.dirty.has(v.String()) {
			q.queue = append(q.queue, v)
		}
	}
	q.cond.Signal()
}
