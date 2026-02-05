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
	"math/bits"
	"math/rand"
	"time"
)

const (
	defaultQueueLength = 20
)

// Item is element stored in heap
type Item struct {
	ObjID string
	// Obj should be a pointer, otherwise UpdateObjByID will fail
	Obj      interface{}
	Priority int
}

// PriorityFunc returns priority of an object
type PriorityFunc func(interface{}) (int, error)

// UpdateObjFunc updates object inside queue
type UpdateObjFunc func(interface{}) error

// PriorityQueue is a two-ended priority queue which keeps item with max priority at front and item with min priority
// at rear using DeHeap
type PriorityQueue struct {
	deHeap       *DeHeap
	identityFunc IdentityFunc
	priorityFunc PriorityFunc
}

// NewPriorityQueue creates priority queue
func NewPriorityQueue(idFunc IdentityFunc, priorityFunc PriorityFunc) *PriorityQueue {
	return &PriorityQueue{
		deHeap:       NewDeHeap(),
		identityFunc: idFunc,
		priorityFunc: priorityFunc,
	}
}

// Front returns the item with max priority
func (pq *PriorityQueue) Front() interface{} {
	if item, ok := pq.deHeap.GetMax().(*Item); ok {
		return item.Obj
	}
	return nil
}

// Range iterates item in queue and process item with given function
func (pq *PriorityQueue) Range(f func(obj interface{}) bool) bool {
	for _, item := range pq.deHeap.items {
		if !f(item.Obj) {
			return false
		}
	}
	return true
}

// SortedRange iterates item in queue and process item with given function in order
func (pq *PriorityQueue) SortedRange(f func(obj interface{}) bool) bool {
	tmpHeap := pq.deHeap.Copy()
	for {
		item, ok := tmpHeap.PopMax().(*Item)
		if !ok {
			return true
		}
		if !f(item.Obj) {
			return false
		}
	}
}

// Back returns the item with min priority
func (pq *PriorityQueue) Back() interface{} {
	if item, ok := pq.deHeap.GetMin().(*Item); ok {
		return item.Obj
	}
	return nil
}

// PopFront pops the item with max priority
func (pq *PriorityQueue) PopFront() interface{} {
	if item, ok := pq.deHeap.PopMax().(*Item); ok {
		return item.Obj
	}
	return nil
}

// PopBack pops the item with min priority
func (pq *PriorityQueue) PopBack() interface{} {
	if item, ok := pq.deHeap.PopMin().(*Item); ok {
		return item.Obj
	}
	return nil
}

// PushBack adds an object into queue
func (pq *PriorityQueue) PushBack(obj interface{}) error {
	priority, err := pq.priorityFunc(obj)
	if err != nil {
		return err
	}
	pq.deHeap.Push(&Item{ObjID: pq.identityFunc(obj), Obj: obj, Priority: priority})
	return nil
}

// GetByID gets an object in queue by its ID
func (pq *PriorityQueue) GetByID(objID string) interface{} {
	index, item := pq.getIndexAndItemByObjID(objID)
	if index == keyNotFoundIndex {
		return nil
	}
	return item.Obj
}

// DelByID deletes an object in queue by its ID
func (pq *PriorityQueue) DelByID(objID string) error {
	index, _ := pq.getIndexAndItemByObjID(objID)
	if index != keyNotFoundIndex {
		pq.deHeap.Remove(index)
		return nil
	}
	return ErrObjectNotFound
}

// Len returns length of queue
func (pq *PriorityQueue) Len() int {
	return pq.deHeap.Len()
}

// UpdateObjByID will update an object in queue by its ID and fix the order
func (pq *PriorityQueue) UpdateObjByID(objID string, obj interface{}) error {
	var err error
	index, item := pq.getIndexAndItemByObjID(objID)
	if index == keyNotFoundIndex {
		return ErrObjectNotFound
	}
	item.Obj = obj
	// update this object's priority and fix the heap
	if item.Priority, err = pq.priorityFunc(obj); err != nil {
		return err
	}
	pq.deHeap.Fix(index)
	return nil
}

// UpdatePriorityFunc -
func (pq *PriorityQueue) UpdatePriorityFunc(priorityFunc PriorityFunc) {
	pq.priorityFunc = priorityFunc
}

func (pq *PriorityQueue) getIndexAndItemByObjID(objID string) (int, *Item) {
	for i := 0; i < pq.deHeap.Len(); i++ {
		if pq.deHeap.items[i].ObjID == objID {
			return i, pq.deHeap.items[i]
		}
	}
	return keyNotFoundIndex, nil
}

// DeHeap is a max-min heap which stores items in max and min levels, root contains the item with max value of all
// levels and one of root's children contains the item with min value of all levels
type DeHeap struct {
	items []*Item
	count int
}

// NewDeHeap creates a DeHeap
func NewDeHeap() *DeHeap {
	rand.Seed(time.Now().UnixNano())
	return &DeHeap{
		items: make([]*Item, 0, defaultQueueLength),
	}
}

// Copy creates a shallow copy of DeHeap
func (dh *DeHeap) Copy() *DeHeap {
	copyItems := make([]*Item, len(dh.items))
	copy(copyItems, dh.items)
	return &DeHeap{
		items: copyItems,
		count: dh.count,
	}
}

// Len returns the number of deHeap in heap
func (dh *DeHeap) Len() int { return len(dh.items) }

// Compare is used to compare two items in heap
func (dh *DeHeap) Compare(i, j int) bool {
	if i >= len(dh.items) || j >= len(dh.items) {
		return false
	}
	return dh.items[i].Priority > dh.items[j].Priority
}

// Swap swaps two items in heap
func (dh *DeHeap) Swap(i, j int) {
	if i >= len(dh.items) || j >= len(dh.items) {
		return
	}
	dh.items[i], dh.items[j] = dh.items[j], dh.items[i]
}

// Push pushes an item to heap
func (dh *DeHeap) Push(x interface{}) {
	item, ok := x.(*Item)
	if !ok {
		return
	}
	dh.items = append(dh.items, item)
	dh.shiftUp(dh.Len() - 1)
}

// Fix fixes heap's order
func (dh *DeHeap) Fix(i int) {
	if j := dh.shiftDown(i); j > 0 {
		dh.shiftUp(j)
	}
}

// Remove removes an item from heap
func (dh *DeHeap) Remove(i int) {
	n := dh.Len() - 1
	if i > n {
		return
	}
	dh.Swap(i, n)
	dh.items[n] = nil
	dh.items = dh.items[0:n]
	dh.shiftDown(i)
}

// GetMax returns the item with max value
func (dh *DeHeap) GetMax() interface{} {
	if dh.Len() < 1 {
		return nil
	}
	return dh.items[0]
}

// GetMin returns the item with min value
func (dh *DeHeap) GetMin() interface{} {
	n := dh.Len() - 1
	if n < 0 {
		return nil
	}
	lChd := 1
	if lChd > n {
		return dh.items[0]
	}
	rChd := 2
	min := lChd
	if rChd <= n && dh.Compare(lChd, rChd) {
		min = rChd
	}
	return dh.items[min]
}

// PopMax pops item with max value
func (dh *DeHeap) PopMax() interface{} {
	n := dh.Len() - 1
	if n < 0 {
		return nil
	}
	item := dh.items[0]
	dh.Swap(0, n)
	dh.items[n] = nil
	dh.items = dh.items[0:n]
	dh.shiftDown(0)
	return item
}

// PopMin pops item with min value
func (dh *DeHeap) PopMin() interface{} {
	n := dh.Len() - 1
	if n < 0 {
		return nil
	}
	lc := leftChild(0)
	rc := rightChild(0)
	if lc > n {
		item := dh.items[0]
		dh.items[0] = nil
		dh.items = dh.items[0:n]
		return item
	}
	t := lc
	if rc <= n && dh.Compare(lc, rc) {
		t = rc
	}
	if t >= len(dh.items) || n >= len(dh.items) {
		return nil
	}
	item := dh.items[t]
	dh.Swap(t, n)
	dh.items[n] = nil
	dh.items = dh.items[0:n]
	dh.shiftDown(t)
	return item
}

func (dh *DeHeap) shiftUp(i int) int {
	if i < 0 {
		return i
	}
	isMax := isMaxLevel(i)
	p := parent(i)
	if p >= 0 {
		if dh.Compare(p, i) == isMax {
			dh.Swap(p, i)
			i = p
			isMax = !isMax
		}
	}
	for g := grandparent(i); g >= 0; g = grandparent(i) {
		if dh.Compare(g, i) == isMax {
			break
		}
		dh.Swap(g, i)
		i = g
	}
	return i
}

func (dh *DeHeap) shiftDown(i int) int {
	if i < 0 {
		return i
	}
	n := dh.Len()
	for i < n {
		isMax := isMaxLevel(i)
		t := i
		// check i's children
		lc, rc := leftChild(i), rightChild(i)
		// no need to go further if lc reaches n but should handle rc reaches n and lc doesn't
		if lc >= n {
			break
		}
		if dh.Compare(lc, t) == isMax {
			t = lc
		}
		if rc < n && dh.Compare(rc, t) == isMax {
			t = rc
		}
		// check i's grandchildren
		for gc := leftChild(lc); gc < n && gc <= rightChild(rc); gc++ {
			if dh.Compare(gc, t) == isMax {
				t = gc
			}
		}
		if t == i {
			break
		}
		dh.Swap(i, t)
		i = t
		// t is i's children, which means i has no conflict with its grandchildren who stand in the same max/min level
		// with i, no need to go further
		if t == lc || t == rc {
			break
		}
		// t is i's grandchildren, need to check if t has conflict with t's parent
		p := parent(t)
		if dh.Compare(p, t) == isMax {
			dh.Swap(p, t)
			i = p
		}
	}
	return i
}

func isMaxLevel(i int) bool {
	level := bits.Len(uint(i)+1) - 1
	return level%2 == 0 // whether the given integer i is at the maximum level
}

func parent(i int) int {
	return (i - 1) / 2 // find the parent node's index
}

func grandparent(i int) int {
	return ((i + 1) / 4) - 1 // find the grandparent node's index
}

func leftChild(i int) int {
	return i*2 + 1 // find the leftChild node's index
}

func rightChild(i int) int {
	return i*2 + 2 // find the rightChild node's index
}
