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

// Package rollout -
package rollout

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/loadbalance"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
)

const (
	defaultAllocRecordSyncChanSize = 1000
)

var (
	globalRolloutHandler = &RFHandler{allocRecordSyncCh: make(chan map[string][]string, defaultAllocRecordSyncChanSize)}
	rolloutSdkClient     api.LibruntimeAPI
)

// GetGlobalRolloutHandler -
func GetGlobalRolloutHandler() *RFHandler {
	return globalRolloutHandler
}

// SetRolloutSdkClient -
func SetRolloutSdkClient(sdkClient api.LibruntimeAPI) {
	rolloutSdkClient = sdkClient
}

// RolloutRatio -
type RolloutRatio struct {
	CurrentVersion string `json:"CurrentVersion"`
	RolloutRatio   string `json:"rolloutRatio"`
}

// RFHandler -
type RFHandler struct {
	LoadBalance       loadbalance.WNGINX
	allocRecordSyncCh chan map[string][]string
	IsGaryUpdating    bool
	ForwardInstance   string
	CurrentVersion    string
	CurrentRatio      int
	sync.RWMutex
}

const (
	forward    = "forward"
	notForward = "noForward"
	maxRatio   = 100
)

// UpdateForwardInstance -
func (rf *RFHandler) UpdateForwardInstance(instanceID string) {
	rf.Lock()
	defer rf.Unlock()
	log.GetLogger().Infof("update forward instance to %s", instanceID)
	rf.ForwardInstance = instanceID
}

// ShouldForwardRequest -
func (rf *RFHandler) ShouldForwardRequest() bool {
	rf.Lock()
	defer rf.Unlock()
	if len(rf.ForwardInstance) == 0 {
		return false
	}
	node := rf.LoadBalance.Next("", true)
	nodeStr, ok := node.(string)
	if !ok {
		return false
	}
	if nodeStr == forward {
		return true
	}
	return false
}

// GetCurrentRatio -
func (rf *RFHandler) GetCurrentRatio() int {
	rf.RLock()
	ratio := rf.CurrentRatio
	rf.RUnlock()
	return ratio
}

// ProcessRatioUpdate -
func (rf *RFHandler) ProcessRatioUpdate(ratioData []byte) error {
	rf.Lock()
	defer rf.Unlock()
	rolloutRatio := &RolloutRatio{}
	err := json.Unmarshal(ratioData, rolloutRatio)
	if err != nil {
		log.GetLogger().Errorf("failed to process ratio update, unmarshal error %s", err.Error())
		return err
	}
	ratio, err := strconv.Atoi(strings.TrimSuffix(rolloutRatio.RolloutRatio, "%"))
	if err != nil {
		log.GetLogger().Errorf("failed to process ratio update, ratio parse error %s", err.Error())
		return err
	}
	if ratio > maxRatio {
		log.GetLogger().Errorf("failed to process ratio update, ratio %s is invalid", ratio)
		return errors.New("rolloutRatio larger than 100%")
	}
	rf.CurrentVersion = rolloutRatio.CurrentVersion
	grayLoadBalance := loadbalance.WNGINX{}
	grayLoadBalance.Add(forward, ratio)
	grayLoadBalance.Add(notForward, maxRatio-ratio)
	rf.LoadBalance = grayLoadBalance
	rf.CurrentRatio = ratio
	log.GetLogger().Infof("succeed to update rollout ratio to %d%", ratio)
	return nil
}

// ProcessRatioDelete -
func (rf *RFHandler) ProcessRatioDelete() {
	rf.Lock()
	defer rf.Unlock()
	grayLoadBalance := loadbalance.WNGINX{}
	rf.LoadBalance = grayLoadBalance
	log.GetLogger().Infof("succeed to delete rollout ratio")
}

// GetAllocRecordSyncChan -
func (rf *RFHandler) GetAllocRecordSyncChan() chan map[string][]string {
	return rf.allocRecordSyncCh
}

// ProcessAllocRecordSync -
func (rf *RFHandler) ProcessAllocRecordSync(selfInsID, targetInsID string) {
	log.GetLogger().Infof("start to process allocation record synchronize")
	rsp, err := rf.SendRolloutRequest(selfInsID, targetInsID)
	if err != nil {
		log.GetLogger().Errorf("failed to sync alloc record from instance %s error %s", targetInsID, err.Error())
		return
	}
	rf.allocRecordSyncCh <- rsp.AllocRecord
	log.GetLogger().Infof("succeed to process allocation record synchronize")
}

// SendRolloutRequest -
func (rf *RFHandler) SendRolloutRequest(selfInsID, targetInsID string) (*commonTypes.RolloutResponse, error) {
	log.GetLogger().Infof("start to send rollout request from %s to %s", selfInsID, targetInsID)
	rolloutArg := api.Arg{
		Type: api.Value,
		Data: []byte(fmt.Sprintf("rollout#%s", selfInsID)),
	}
	rspData, err := InvokeByInstanceId([]api.Arg{rolloutArg}, targetInsID, "")
	if err != nil {
		log.GetLogger().Errorf("failed to send rollout request to %s error %s", targetInsID, err.Error())
		return nil, err
	}
	rolloutResp := &commonTypes.RolloutResponse{}
	err = json.Unmarshal(rspData, rolloutResp)
	if err != nil {
		log.GetLogger().Errorf("failed to unmarshal rollout response error %s", err.Error())
		return nil, err
	}
	log.GetLogger().Infof("succeed to send rollout request from %s to %s", selfInsID, targetInsID)
	return rolloutResp, nil
}

// InvokeByInstanceId -
func InvokeByInstanceId(args []api.Arg, instanceID string, traceID string) ([]byte, error) {
	wait := make(chan struct{}, 1)
	var (
		res    []byte
		resErr error
	)
	invokeStart := time.Now()
	funcMeta := api.FunctionMeta{FuncID: constant.FaasSchedulerName, Api: api.PosixApi}
	invokeOpts := api.InvokeOptions{TraceID: traceID}
	objID, invokeErr := rolloutSdkClient.InvokeByInstanceId(funcMeta, instanceID, args, invokeOpts)
	if invokeErr != nil {
		log.GetLogger().Errorf("failed to invoke by id %s, traceID: %s, function: %s, error: %s",
			instanceID, traceID, constant.FaasSchedulerName, invokeErr.Error())
		return nil, invokeErr
	}
	rolloutSdkClient.GetAsync(objID, func(result []byte, err error) {
		res = result
		resErr = err
		wait <- struct{}{}
		if _, err := rolloutSdkClient.GDecreaseRef([]string{objID}); err != nil {
			log.GetLogger().Warnf("GDecreaseRef objID %s failed, %s", objID, err.Error())
		}
	})
	<-wait
	log.GetLogger().Infof("success invoke instance: %s, resErr %v, totalTime: %f", instanceID, resErr,
		time.Since(invokeStart).Seconds())
	return res, resErr
}
