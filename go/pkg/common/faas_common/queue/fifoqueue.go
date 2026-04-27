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

// Package queue -
package queue

import (
	"container/list"
)

const (
	defaultMapSize = 16
)

// FifoQueue implements a fifo scheduling queue.
type FifoQueue struct {
	queue         *list.List
	identityFunc  IdentityFunc
	elementRecord map[string]*list.Element
}

// NewFifoQueue return fifo queue
func NewFifoQueue(identityFunc IdentityFunc) *FifoQueue {
	return &FifoQueue{
		queue:         list.New(),
		identityFunc:  identityFunc,
		elementRecord: make(map[string]*list.Element, defaultMapSize),
	}
}

// Front return front item of queue
func (fq *FifoQueue) Front() interface{} {
	if fq.queue.Len() == 0 {
		return nil
	}
	obj := fq.queue.Front().Value
	return obj
}

// Back return rear item of queue
func (fq *FifoQueue) Back() interface{} {
	if fq.queue.Len() == 0 {
		return nil
	}
	obj := fq.queue.Back().Value
	return obj
}

// PopFront pops an object from front
func (fq *FifoQueue) PopFront() interface{} {
	if fq.queue.Len() == 0 {
		return nil
	}
	elem := fq.queue.Front()
	if elem == nil {
		return nil
	}
	obj := elem.Value
	if fq.identityFunc != nil {
		delete(fq.elementRecord, fq.identityFunc(obj))
	}
	fq.queue.Remove(elem)
	return obj
}

// PopBack pops an object from back
func (fq *FifoQueue) PopBack() interface{} {
	if fq.queue.Len() == 0 {
		return nil
	}
	elem := fq.queue.Back()
	if elem == nil {
		return nil
	}
	obj := elem.Value
	if fq.identityFunc != nil {
		delete(fq.elementRecord, fq.identityFunc(obj))
	}
	fq.queue.Remove(elem)
	return obj
}

// PushBack adds an object into queue
func (fq *FifoQueue) PushBack(obj interface{}) error {
	if fq.identityFunc != nil {
		fq.elementRecord[fq.identityFunc(obj)] = fq.queue.PushBack(obj)
	} else {
		fq.queue.PushBack(obj)
	}
	return nil
}

// GetByID gets an object in queue by its ID
func (fq *FifoQueue) GetByID(objID string) interface{} {
	elem, exist := fq.elementRecord[objID]
	if !exist {
		return nil
	}
	return elem.Value
}

// DelByID deletes an object in queue by its ID
func (fq *FifoQueue) DelByID(objID string) error {
	elem, exist := fq.elementRecord[objID]
	if !exist {
		return ErrObjectNotFound
	}
	delete(fq.elementRecord, objID)
	fq.queue.Remove(elem)
	return nil
}

// Len returns length of queue
func (fq *FifoQueue) Len() int {
	return fq.queue.Len()
}

// UpdateObjByID will update an object in queue by its ID and fix the order
func (fq *FifoQueue) UpdateObjByID(objID string, obj interface{}) error {
	return ErrMethodUnsupported
}

// Range iterates item in queue and process item with given function
func (fq *FifoQueue) Range(f func(obj interface{}) bool) bool {
	for item := fq.queue.Front(); item != nil; item = item.Next() {
		if !f(item) {
			break
		}
	}
	return true
}

// SortedRange iterates item in queue and process item with given function in order
func (fq *FifoQueue) SortedRange(f func(obj interface{}) bool) bool {
	for item := fq.queue.Front(); item != nil; item = item.Next() {
		if !f(item) {
			break
		}
	}
	return true
}
