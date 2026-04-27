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

// Package instancequeue -
package instancequeue

import (
	"container/list"
	"sync"

	"yuanrong.org/kernel/pkg/functionscaler/scaler"
)

// InstanceCreateRequest is the request to create instance
type InstanceCreateRequest struct {
	callback scaler.ScaleUpCallback
}

// InstanceCreateQueue stores instance create requests
type InstanceCreateQueue struct {
	queue      *list.List
	createHead *list.Element
	cancelHead *list.Element
	cond       *sync.Cond
	stopped    bool
}

// NewInstanceCreateQueue creates new InstanceCreateQueue
func NewInstanceCreateQueue() *InstanceCreateQueue {
	return &InstanceCreateQueue{
		queue: list.New(),
		cond:  sync.NewCond(new(sync.Mutex)),
	}
}

func (iq *InstanceCreateQueue) push(request *InstanceCreateRequest) {
	iq.cond.L.Lock()
	newElem := iq.queue.PushBack(request)
	if iq.createHead == nil {
		iq.createHead = newElem
	}
	iq.cancelHead = newElem
	iq.cond.L.Unlock()
	iq.cond.Signal()
}

func (iq *InstanceCreateQueue) getForCreate() *InstanceCreateRequest {
	iq.cond.L.Lock()
	if iq.createHead == nil {
		iq.cond.Wait()
	}
	if iq.stopped {
		iq.cond.L.Unlock()
		return nil
	}
	curElem := iq.createHead
	if curElem == nil {
		iq.cond.L.Unlock()
		return nil
	}
	if curElem == iq.createHead {
		iq.createHead = iq.createHead.Next()
	}
	if curElem == iq.cancelHead {
		iq.cancelHead = iq.cancelHead.Prev()
	}
	iq.queue.Remove(curElem)
	iq.cond.L.Unlock()
	request, ok := curElem.Value.(*InstanceCreateRequest)
	if !ok {
		return nil
	}
	return request
}

func (iq *InstanceCreateQueue) getForCancel() *InstanceCreateRequest {
	iq.cond.L.Lock()
	if iq.stopped {
		iq.cond.L.Unlock()
		return nil
	}
	curElem := iq.cancelHead
	if curElem == nil {
		iq.cond.L.Unlock()
		return nil
	}
	if curElem == iq.createHead {
		iq.createHead = iq.createHead.Next()
	}
	if curElem == iq.cancelHead {
		iq.cancelHead = iq.cancelHead.Prev()
	}
	iq.queue.Remove(curElem)
	iq.cond.L.Unlock()
	request, ok := curElem.Value.(*InstanceCreateRequest)
	if !ok {
		return nil
	}
	return request
}

func (iq *InstanceCreateQueue) getQueLen() int {
	iq.cond.L.Lock()
	queLen := iq.queue.Len()
	iq.cond.L.Unlock()
	return queLen
}

func (iq *InstanceCreateQueue) destroy() {
	iq.cond.L.Lock()
	iq.stopped = true
	iq.cond.L.Unlock()
	iq.cond.Broadcast()
}
