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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeHeap(t *testing.T) {
	items := []*Item{
		{
			Obj:      "apple",
			Priority: 16,
		},
		{
			Obj:      "banana",
			Priority: 15,
		},
		{
			Obj:      "berry",
			Priority: 17,
		},
		{
			Obj:      "cherry",
			Priority: 14,
		},
		{
			Obj:      "grape",
			Priority: 18,
		},
		{
			Obj:      "lemon",
			Priority: 13,
		},
		{
			Obj:      "haw",
			Priority: 12,
		},
		{
			Obj:      "mango",
			Priority: 19,
		},
		{
			Obj:      "orange",
			Priority: 20,
		},
		{
			Obj:      "watermelon",
			Priority: 11,
		},
	}
	dh := NewDeHeap()
	for _, item := range items {
		dh.Push(item)
	}
	popMax1 := dh.PopMax().(*Item).Obj
	popMin1 := dh.PopMin().(*Item).Obj
	assert.Equal(t, "orange", popMax1.(string))
	assert.Equal(t, "watermelon", popMin1.(string))
	getMax1 := dh.GetMax().(*Item).Obj
	getMin1 := dh.GetMin().(*Item).Obj
	assert.Equal(t, "mango", getMax1.(string))
	assert.Equal(t, "haw", getMin1.(string))
	dh.items[3].Priority = 11
	dh.Fix(3)
	getMin2 := dh.GetMin().(*Item).Obj
	assert.Equal(t, "grape", getMin2.(string))
	dh.items[1].Priority = 20
	dh.Fix(1)
	getMax2 := dh.GetMax().(*Item).Obj
	assert.Equal(t, "grape", getMax2.(string))
	dh.Remove(2)
	getMin3 := dh.GetMin().(*Item).Obj
	assert.Equal(t, "lemon", getMin3.(string))
}

func TestPriorityQueue(t *testing.T) {
	type testItem struct {
		id       string
		priority int
	}
	identityFunc := func(obj interface{}) string {
		if item, ok := obj.(*testItem); ok {
			return item.id
		}
		return ""
	}
	priorityFunc := func(obj interface{}) (int, error) {
		if item, ok := obj.(*testItem); ok {
			return item.priority, nil
		}
		return -1, fmt.Errorf("failed to get priority")
	}
	item1 := &testItem{id: "1", priority: 50}
	item2 := &testItem{id: "2", priority: 51}
	item3 := &testItem{id: "3", priority: 51}
	item4 := &testItem{id: "4", priority: 51}
	item5 := &testItem{id: "5", priority: 60}
	queue := NewPriorityQueue(identityFunc, priorityFunc)
	frontItem1 := queue.Front()
	backItem1 := queue.Back()
	assert.Equal(t, nil, frontItem1)
	assert.Equal(t, nil, backItem1)

	popBack1 := queue.PopBack()
	assert.Equal(t, nil, popBack1)
	popFront1 := queue.PopFront()
	assert.Equal(t, nil, popFront1)

	queue.PushBack(item1)
	frontItem2 := queue.Front().(*testItem)
	backItem2 := queue.Back().(*testItem)
	assert.Equal(t, "1", frontItem2.id)
	assert.Equal(t, "1", backItem2.id)

	queue.PushBack(item2)
	queue.PushBack(item3)
	frontItem3 := queue.Front().(*testItem)
	backItem3 := queue.Back().(*testItem)
	assert.Equal(t, 51, frontItem3.priority)
	assert.Equal(t, 50, backItem3.priority)

	queue.PushBack(item4)
	queue.PushBack(item5)
	frontItem4 := queue.Front().(*testItem)
	backItem4 := queue.Back().(*testItem)
	assert.Equal(t, 60, frontItem4.priority)
	assert.Equal(t, 50, backItem4.priority)

	item2.priority = 40
	queue.UpdateObjByID("2", item2)
	backItem5 := queue.Back().(*testItem)
	assert.Equal(t, "2", backItem5.id)
	item3.priority = 70
	queue.UpdateObjByID("3", item3)
	frontItem5 := queue.Front().(*testItem)
	assert.Equal(t, "3", frontItem5.id)

	queue.DelByID("4")
	frontItem6 := queue.Front().(*testItem)
	backItem6 := queue.Back().(*testItem)
	assert.Equal(t, "3", frontItem6.id)
	assert.Equal(t, "2", backItem6.id)

	item3.priority = 40
	queue.UpdateObjByID("3", item3)
	frontItem7 := queue.Front().(*testItem)
	assert.Equal(t, "5", frontItem7.id)

	item2.priority = 30
	queue.UpdateObjByID("2", item2)
	backItem7 := queue.Back().(*testItem)
	assert.Equal(t, "2", backItem7.id)

	getByID1 := queue.GetByID("qwe")
	assert.Equal(t, getByID1, nil)

	getByID2 := queue.GetByID("2").(*testItem)
	assert.Equal(t, "2", getByID2.id)

	popBack2 := queue.PopBack().(*testItem)
	assert.Equal(t, "2", popBack2.id)
	popFront2 := queue.PopFront().(*testItem)
	assert.Equal(t, "5", popFront2.id)

	length := queue.Len()
	assert.Equal(t, 2, length)
}

func TestPriorityQueueUpdateFrontInSequence(t *testing.T) {
	type testItem struct {
		id       string
		priority int
	}
	identityFunc := func(obj interface{}) string {
		if item, ok := obj.(*testItem); ok {
			return item.id
		}
		return ""
	}
	priorityFunc := func(obj interface{}) (int, error) {
		if item, ok := obj.(*testItem); ok {
			return item.priority, nil
		}
		return -1, fmt.Errorf("failed to get priority")
	}
	items := []*testItem{
		&testItem{id: "1", priority: 2},
		&testItem{id: "2", priority: 2},
		&testItem{id: "3", priority: 2},
		&testItem{id: "4", priority: 2},
	}
	queue := NewPriorityQueue(identityFunc, priorityFunc)
	for _, item := range items {
		queue.PushBack(item)
	}
	for {
		front := queue.Front().(*testItem)
		if front.priority == 0 {
			break
		}
		front.priority -= 1
		queue.UpdateObjByID(front.id, front)
	}
	queue.Range(func(obj interface{}) bool {
		item := obj.(*testItem)
		if item.priority != 0 {
			t.Errorf("item %s priority %d should be 0", item.id, item.priority)
		}
		return true
	})
}

func TestPriorityQueue_SortedRange(t *testing.T) {
	type testItem struct {
		id       string
		priority int
	}
	identityFunc := func(obj interface{}) string {
		if item, ok := obj.(*testItem); ok {
			return item.id
		}
		return ""
	}
	priorityFunc := func(obj interface{}) (int, error) {
		if item, ok := obj.(*testItem); ok {
			return item.priority, nil
		}
		return -1, fmt.Errorf("failed to get priority")
	}
	items := []*testItem{
		&testItem{id: "1", priority: 1},
		&testItem{id: "2", priority: 2},
		&testItem{id: "3", priority: 3},
		&testItem{id: "4", priority: 4},
	}
	queue := NewPriorityQueue(identityFunc, priorityFunc)
	for _, item := range items {
		queue.PushBack(item)
	}
	rangeItems := make([]*testItem, 0, 4)
	queue.SortedRange(func(obj interface{}) bool {
		item := obj.(*testItem)
		rangeItems = append(rangeItems, item)
		return true
	})
	for i, item := range rangeItems {
		if i+item.priority != 4 {
			t.Errorf("range item %+v in wrong order %d\n", item, i)
		}
	}
	front := queue.PopFront().(*testItem)
	back := queue.PopBack().(*testItem)
	assert.Equal(t, 4, front.priority)
	assert.Equal(t, 1, back.priority)
}
