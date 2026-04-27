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

// Package faasscheduler -
package faasscheduler

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"yuanrong.org/kernel/runtime/faassdk/common/loadbalance"
	"yuanrong.org/kernel/runtime/libruntime/api"
	"yuanrong.org/kernel/runtime/libruntime/common/faas/logger"
)

const (
	// hashRingSize the concurrent hash ring length
	hashRingSize = 100
	limiterTime  = 1 * time.Millisecond
)

// SchedulerInfo is scheduler info
type SchedulerInfo struct {
	SchedulerFuncKey string   `json:"schedulerFuncKey"`
	SchedulerIDList  []string `json:"schedulerIDList"`
}

// Proxy is the singleton proxy
var Proxy *schedulerProxy

func init() {
	Proxy = newSchedulerProxy(
		loadbalance.NewLimiterCHGeneric(limiterTime),
	)
}

// schedulerProxy is used to get instances from FaaSScheduler via a grpc stream
type schedulerProxy struct {
	// used to select a FaaSScheduler by the func info Concurrent Consistent Hash
	loadBalance      loadbalance.LBInterface
	schedulerFuncKey string
}

// Add an FaaSScheduler
func (im *schedulerProxy) Add(faaSSchedulerID string) {
	im.loadBalance.Add(faaSSchedulerID, 0)
}

// Remove a FaaSScheduler
func (im *schedulerProxy) Remove(faaSSchedulerID string) {
	im.loadBalance.Remove(faaSSchedulerID)
}

// Get an instance for this request
func (im *schedulerProxy) Get(funcKey string) (string, error) {
	logger.GetLogger().Infof("Getting instance from scheduler for funcKey: %s", funcKey)
	// select one FaaSScheduler by the func key
	next := im.loadBalance.Next(funcKey, false)
	faaSSchedulerID, ok := next.(string)
	if !ok {
		return "", fmt.Errorf("failed to parse the result of loadbanlance: %+v", next)
	}
	if strings.TrimSpace(faaSSchedulerID) == "" {
		return "", fmt.Errorf("no avaiable faas scheduler was found")
	}
	return faaSSchedulerID, nil
}

// GetSchedulerFuncKey -
func (im *schedulerProxy) GetSchedulerFuncKey() string {
	return im.schedulerFuncKey
}

// newSchedulerProxy return an instance pool which get the instance from the remote FaaSScheduler
func newSchedulerProxy(lb loadbalance.LBInterface) *schedulerProxy {
	return &schedulerProxy{
		loadBalance: lb,
	}
}

// ParseSchedulerData -
func ParseSchedulerData(args api.Arg) error {
	schedulerInfo := &SchedulerInfo{}
	err := json.Unmarshal(args.Data, schedulerInfo)
	if err != nil {
		return err
	}
	for _, schedulerID := range schedulerInfo.SchedulerIDList {
		Proxy.Add(schedulerID)
	}
	Proxy.schedulerFuncKey = schedulerInfo.SchedulerFuncKey
	return nil
}

// SetStain -
func (im *schedulerProxy) SetStain(funcKey, instanceID string) {
	if v, ok := im.loadBalance.(*loadbalance.LimiterCHGeneric); ok {
		v.SetStain(funcKey, instanceID)
	}
}
