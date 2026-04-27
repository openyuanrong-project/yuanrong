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

// Package stateinstance -
package stateinstance

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	"yuanrong.org/kernel/pkg/common/faas_common/statuscode"
	"yuanrong.org/kernel/pkg/common/uuid"
)

// DeleteStateInstance -
type DeleteStateInstance func(stateID string, instanceID string)

// Lease one Leaser has multiple leases
type Lease struct {
	ID    int
	timer *time.Timer // retain lease interval
}

// Leaser one state instance has one Leaser
type Leaser struct {
	Leases              map[int]*Lease
	mu                  sync.Mutex
	nextID              int
	maxLeases           int
	timer               *time.Timer // Scale-down Interval Without lease
	deleteStateInstance DeleteStateInstance
	stateID             string
	instanceID          string
	scaleDownWindow     time.Duration
}

// NewLeaser new leaser for the state instance
func NewLeaser(maxLeases int, deleteStateInstance DeleteStateInstance, stateID string, instanceID string,
	scaleDownWindow time.Duration) *Leaser {
	return &Leaser{
		Leases:              make(map[int]*Lease),
		maxLeases:           maxLeases,
		deleteStateInstance: deleteStateInstance,
		stateID:             stateID,
		instanceID:          instanceID,
		scaleDownWindow:     scaleDownWindow,
	}
}

// Recover when the leaser is recovering, start timer for deleting the instance.
// because the timer is set when aquire a lease normally but no aquiring existing.
func (l *Leaser) Recover() {
	l.mu.Lock()
	defer l.mu.Unlock()
	log.GetLogger().Infof("recover leaser timer")
	if l.timer != nil {
		l.timer.Stop()
	}
	l.timer = time.AfterFunc(l.scaleDownWindow, func() {
		log.GetLogger().Warnf("No lease for %v, scale down, stateKey is %s", l.scaleDownWindow, l.stateID)
		l.deleteStateInstance(l.stateID, l.instanceID)
	})
}

// Terminate when state route is delete, the leaser should be terminate
func (l *Leaser) Terminate() {
	l.mu.Lock()
	defer l.mu.Unlock()

	log.GetLogger().Infof("terminate leaser %s", l.stateID)
	if l.timer != nil {
		l.timer.Stop()
		l.timer = nil
	}
}

// AcquireLease acquire lease for instance
func (l *Leaser) AcquireLease(leaseInterval time.Duration) (*Lease, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.Leases) >= l.maxLeases {
		return nil, snerror.New(statuscode.StateInstanceNoLease, statuscode.StateInstanceNoLeaseMsg)
	}

	lease := &Lease{ID: l.nextID}
	lease.timer = time.AfterFunc(leaseInterval, func() {
		l.ReleaseLease(lease.ID)
	})

	l.Leases[lease.ID] = lease
	l.nextID++
	if l.timer != nil {
		l.timer.Stop()
		l.timer = nil
	} else {
		log.GetLogger().Infof("timer is nil")
	}
	return lease, nil
}

// ReleaseLease lease expiration or call release lease
func (l *Leaser) ReleaseLease(id int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	logger := log.GetLogger().With(zap.Any("traceID", uuid.New()))

	lease, exists := l.Leases[id]
	if exists {
		logger.Warnf("Release Lease id %d, stateKey is %s", id, l.stateID)
		lease.timer.Stop()
		delete(l.Leases, id)
	}

	if len(l.Leases) == 0 {
		log.GetLogger().Debugf("set timer when lease len = 0")
		if l.timer != nil {
			l.timer.Stop()
		}
		l.timer = time.AfterFunc(l.scaleDownWindow, func() {
			logger.Warnf("No lease for %v, scale down, stateKey is %s", l.scaleDownWindow, l.stateID)
			l.deleteStateInstance(l.stateID, l.instanceID)
		})
	}
}

// RetainLease renew within the lease interval
func (l *Leaser) RetainLease(id int, leaseInterval time.Duration) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	lease, exists := l.Leases[id]
	if !exists {
		return fmt.Errorf("lease%d not found", id)
	}

	lease.timer.Stop()
	lease.timer = time.AfterFunc(leaseInterval, func() {
		l.ReleaseLease(id)
	})

	return nil
}
