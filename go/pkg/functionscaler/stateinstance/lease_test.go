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

package stateinstance

import (
	"sync"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

func TestNewLeaser(t *testing.T) {
	convey.Convey("test NewLeaser", t, func() {
		leaser := NewLeaser(10, nil, "testStateID", "testInstanceID", 10*time.Second)
		convey.So(leaser, convey.ShouldNotBeNil)
		leaser.Recover()
	})
}

func TestReleaseLease(t *testing.T) {
	convey.Convey("test ReleaseLease", t, func() {
		timer1 := time.NewTimer(1 * time.Second)
		leaser := &Leaser{
			Leases: map[int]*Lease{1: {
				timer: timer1,
			}, 2: {}},
			mu:      sync.Mutex{},
			timer:   &time.Timer{},
			stateID: "id",
		}
		leaser.ReleaseLease(1)
		convey.So(len(leaser.Leases), convey.ShouldEqual, 1)
	})
}

func TestRetainLease(t *testing.T) {
	convey.Convey("test RetainLease", t, func() {
		timer1 := time.NewTimer(1 * time.Second)
		leaser := &Leaser{
			Leases: map[int]*Lease{1: {
				timer: timer1,
			}, 2: {}},
			mu:      sync.Mutex{},
			stateID: "id",
		}
		err := leaser.RetainLease(1, 1*time.Millisecond)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestAcquireLease(t *testing.T) {
	convey.Convey("test AcquireLease", t, func() {
		convey.Convey("max leases", func() {
			leaser := &Leaser{
				Leases:    map[int]*Lease{1: {}, 2: {}},
				mu:        sync.Mutex{},
				timer:     &time.Timer{},
				stateID:   "id",
				maxLeases: 1,
			}
			_, err := leaser.AcquireLease(1 * time.Millisecond)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("timer not nil", func() {
			timer1 := time.NewTimer(1 * time.Second)
			leaser := &Leaser{
				Leases:    map[int]*Lease{1: {}, 2: {}},
				mu:        sync.Mutex{},
				timer:     timer1,
				stateID:   "id",
				maxLeases: 3,
				nextID:    2,
			}
			lease, err := leaser.AcquireLease(1 * time.Millisecond)
			convey.So(lease, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("max leases timer is nil", func() {
			timer1 := time.NewTimer(1 * time.Second)
			leaser := &Leaser{
				Leases:    map[int]*Lease{1: {}, 2: {}},
				mu:        sync.Mutex{},
				stateID:   "id",
				maxLeases: 3,
				nextID:    2,
				timer:     timer1,
			}
			lease, err := leaser.AcquireLease(1 * time.Millisecond)
			convey.So(lease, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestTerminate(t *testing.T) {
	convey.Convey("test Terminate", t, func() {
		timer1 := time.NewTimer(1 * time.Second)
		leaser := &Leaser{
			Leases:  map[int]*Lease{1: {}, 2: {}},
			mu:      sync.Mutex{},
			timer:   timer1,
			stateID: "id",
		}
		leaser.Terminate()
		convey.So(len(leaser.Leases), convey.ShouldEqual, 2)
	})
}
