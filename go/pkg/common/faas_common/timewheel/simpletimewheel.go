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

// Package timewheel -
package timewheel

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const (
	minPace           = 2 * time.Millisecond
	minSlotNum        = 1
	notifyChannelSize = 1000
	readyChan         = 10
	readyChanBlock    = 5
)

type timeTrigger struct {
	taskID      string
	times       int
	index       int64
	circle      int64
	circleCount int64
}

// SimpleTimeWheel will trigger task at given interval by given times, it contains a certain number of slots and moves
// from one slot to another with a pace which is also the granularity of time wheel, task interval will be measured
// with a number of slots and recorded in the slot arrays, each slot has a linked list to trigger a series of tasks
// when time wheel moves to this slot
type SimpleTimeWheel struct {
	ticker      *time.Ticker
	pace        time.Duration
	perimeter   int64
	slotNum     int64
	curSlot     atomic.Int64
	pendingTask int
	slots       []*sync.Map
	record      *sync.Map
	notifyCh    chan struct{}
	readyCh     chan *[]string
	stopCh      chan struct{}
}

// NewSimpleTimeWheel will create a SimpleTimeWheel
func NewSimpleTimeWheel(pace time.Duration, slotNum int64) TimeWheel {
	if pace < minPace {
		pace = minPace
	}
	if slotNum < minSlotNum {
		slotNum = minSlotNum
	}
	timeWheel := &SimpleTimeWheel{
		ticker:    time.NewTicker(pace),
		pace:      pace,
		perimeter: slotNum * int64(pace),
		slotNum:   slotNum,
		curSlot:   atomic.Int64{},
		slots:     make([]*sync.Map, slotNum, slotNum),
		record:    new(sync.Map),
		notifyCh:  make(chan struct{}, notifyChannelSize),
		readyCh:   make(chan *[]string, readyChan),
		stopCh:    make(chan struct{}),
	}
	for i := range timeWheel.slots {
		timeWheel.slots[i] = &sync.Map{}
	}
	go timeWheel.run()
	return timeWheel
}

func (gt *SimpleTimeWheel) run() {
	for {
		select {
		case <-gt.ticker.C:
			// 只有这里会修改curSlot，不需要考虑过期
			curSlot := (gt.curSlot.Load() + 1) % int64(len(gt.slots))
			gt.curSlot.Store(curSlot)
			gt.checkAndFireTrigger(curSlot)
		case <-gt.stopCh:
			gt.ticker.Stop()
			return
		}
	}
}

func (gt *SimpleTimeWheel) checkAndFireTrigger(currSlot int64) {
	slot := gt.slots[currSlot]
	var readyList []string
	slot.Range(func(id, object any) bool {
		trigger, ok := object.(*timeTrigger)
		if !ok {
			slot.Delete(id)
			return true
		}
		if trigger.circleCount == trigger.circle {
			trigger.circleCount = 0
			if trigger.times == 0 {
				gt.record.Delete(trigger.taskID)
				slot.Delete(id)
			}
			readyList = append(readyList, trigger.taskID)
			if trigger.times > 0 {
				trigger.times--
			}
		}
		trigger.circleCount++
		return true
	})
	if len(readyList) != 0 {
		select {
		case gt.readyCh <- &readyList:
		case <-gt.stopCh:
			return
		default:
		}
	}
}

// Wait will block until tasks are triggered and returns triggered task list
func (gt *SimpleTimeWheel) Wait() ([]string, error) {
	readyList, ok := <-gt.readyCh
	if !ok {
		return nil, nil
	}
	if len(gt.readyCh) > readyChanBlock {
		return *readyList, fmt.Errorf("ready chan has been block, may lease leak")
	}
	return *readyList, nil
}

// AddTask will add a task which will be triggered periodically over an given interval with given times (-1 means to
// run endlessly), considering that pace has a reasonable size and the logic below won't cost more time than that,
// AddTask won't catch up with the curSlot, so we don't need a mutex. it's also worth noticing that interval can't be
// smaller than the circumference of this time wheel
func (gt *SimpleTimeWheel) AddTask(taskID string, interval time.Duration, times int) error {
	if interval < time.Duration(gt.perimeter) {
		return ErrInvalidTaskInterval
	}
	if _, exist := gt.record.Load(taskID); exist {
		return fmt.Errorf("%s, taskId: %s", ErrTaskAlreadyExist.Error(), taskID)
	}
	curSlot := gt.curSlot.Load()
	circle := (int64(interval)/int64(gt.pace) + curSlot + 1) / gt.slotNum
	circleCount := int64(1)
	index := (int64(interval)/int64(gt.pace) + curSlot + 1) % gt.slotNum
	if index > curSlot {
		circleCount--
	}
	trigger := &timeTrigger{
		taskID:      taskID,
		times:       times,
		index:       index,
		circle:      circle,
		circleCount: circleCount,
	}
	gt.slots[index].Store(trigger.taskID, trigger)
	gt.record.Store(taskID, trigger)
	return nil
}

// DelTask will delete a task in SimpleTimeWheel and remove its trigger
func (gt *SimpleTimeWheel) DelTask(taskID string) error {
	object, exist := gt.record.Load(taskID)
	if !exist {
		return nil
	}
	gt.record.Delete(taskID)
	trigger, ok := object.(*timeTrigger)
	if !ok {
		return errors.New("not a timeTrigger type")
	}
	gt.slots[trigger.index].Delete(taskID)
	return nil
}

// Stop will stop time wheel
func (gt *SimpleTimeWheel) Stop() {
	close(gt.stopCh)
	close(gt.readyCh)
}

// UpdateTask del and add a task
func (gt *SimpleTimeWheel) UpdateTask(taskID string, interval time.Duration, times int) error {
	object, exist := gt.record.Load(taskID)
	if !exist {
		return gt.AddTask(taskID, interval, times)
	}
	trigger, ok := object.(*timeTrigger)
	if !ok {
		return errors.New("not a timeTrigger type")
	}
	gt.slots[trigger.index].Delete(trigger.taskID)

	curSlot := gt.curSlot.Load()
	circle := (int64(interval)/int64(gt.pace) + curSlot + 1) / gt.slotNum
	circleCount := int64(1)
	index := (int64(interval)/int64(gt.pace) + curSlot + 1) % gt.slotNum
	if index > curSlot {
		circleCount--
	}
	trigger.times = times
	trigger.circle = circle
	trigger.circleCount = circleCount
	trigger.index = index
	gt.slots[index].Store(trigger.taskID, trigger)
	return nil
}
