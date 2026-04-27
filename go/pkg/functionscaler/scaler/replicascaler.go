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

// Package scaler -
package scaler

import (
	"sync"
	"time"

	"go.uber.org/zap"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/instanceconfig"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/metrics"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

// ReplicaScaler will scale based on a given instance number
type ReplicaScaler struct {
	metricsCollector metrics.Collector
	funcKeyWithRes   string
	concurrentNum    int
	pendingRsvInsNum int
	targetRsvInsNum  int
	currentRsvInsNum int
	createError      error
	enable           bool
	scaleLimit       int
	scaleUpHandler   ScaleUpHandler
	scaleDownHandler ScaleDownHandler
	sync.RWMutex

	logger api.FormatLogger
}

// NewReplicaScaler will create a ReplicaScaler
func NewReplicaScaler(funcKeyWithRes string, metricsCollector metrics.Collector, scaleUpHandler ScaleUpHandler,
	scaleDownHandler ScaleDownHandler) InstanceScaler {
	replicaScaler := &ReplicaScaler{
		metricsCollector: metricsCollector,
		funcKeyWithRes:   funcKeyWithRes,
		targetRsvInsNum:  0,
		enable:           false,
		scaleUpHandler:   scaleUpHandler,
		scaleDownHandler: scaleDownHandler,
		logger:           log.GetLogger().With(zap.Any("funcKey", funcKeyWithRes)),
	}
	replicaScaler.logger.Infof("create replica scaler")
	return replicaScaler
}

// SetEnable will configure the enable of scaler
func (rs *ReplicaScaler) SetEnable(enable bool) {
	rs.Lock()
	if enable == rs.enable {
		rs.Unlock()
		return
	}
	rs.enable = enable
	rs.Unlock()
	if enable {
		rs.handleScale()
	}
}

// TriggerScale will trigger scale
func (rs *ReplicaScaler) TriggerScale() {
	rs.RLock()
	if !rs.enable {
		rs.RUnlock()
		return
	}
	rs.RUnlock()
	rs.handleScale()
}

// CheckScaling will check if scaler is scaling
func (rs *ReplicaScaler) CheckScaling() bool {
	rs.RLock()
	isScaling := rs.currentRsvInsNum != rs.targetRsvInsNum
	rs.RUnlock()
	return isScaling
}

// UpdateCreateMetrics will update create metrics
func (rs *ReplicaScaler) UpdateCreateMetrics(coldStartTime time.Duration) {
}

// HandleInsThdUpdate will update instance thread metrics, totalInsThd increase should be coupled with pendingInsThd
// decrease for better consistency
func (rs *ReplicaScaler) HandleInsThdUpdate(inUseInsThdDiff, totalInsThdDiff int) {
	rs.metricsCollector.UpdateInsThdMetrics(inUseInsThdDiff)
	// replica scaler won't handle scale when only inUseInsThdDiff is non-zero
	if totalInsThdDiff == 0 {
		return
	}
	rs.Lock()
	rs.currentRsvInsNum += totalInsThdDiff / rs.concurrentNum
	rs.Unlock()
	rs.handleScale()
}

// HandleFuncSpecUpdate -
func (rs *ReplicaScaler) HandleFuncSpecUpdate(funcSpec *types.FunctionSpecification) {
	rs.Lock()
	rs.concurrentNum = utils.GetConcurrentNum(funcSpec.InstanceMetaData.ConcurrentNum)
	rs.Unlock()
	rs.logger.Infof("config concurrentNum to %d for replica scaler", rs.concurrentNum)
	// some error may be cleared after function update
	rs.handleScale()
}

// HandleInsConfigUpdate -
func (rs *ReplicaScaler) HandleInsConfigUpdate(insConfig *instanceconfig.Configuration) {
	replicaNum := int(insConfig.InstanceMetaData.MinInstance)
	if replicaNum < 0 {
		replicaNum = 0
	}
	rs.Lock()
	prevTargetRsvInsNum := rs.targetRsvInsNum
	rs.targetRsvInsNum = replicaNum
	rs.logger.Infof("config reserved num from %d to %d for replica scaler", prevTargetRsvInsNum, replicaNum)
	if !rs.enable {
		rs.Unlock()
		return
	}
	rs.Unlock()
	rs.handleScale()
}

// HandleCreateError handles instance create error
func (rs *ReplicaScaler) HandleCreateError(createError error) {
	rs.logger.Infof("handle create error %s", createError)
	if utils.IsUnrecoverableError(createError) {
		if utils.IsNoNeedToRePullError(createError) {
			rs.Lock()
			rs.enable = false
			rs.createError = createError
			rs.Unlock()
			return
		}
	}
	// if unrecoverable error is cleared then trigger scale, if createError stays nil then skip scale
	rs.Lock()
	noNeedScale := rs.createError == nil && createError == nil
	rs.createError = createError
	rs.Unlock()
	if noNeedScale {
		return
	}
	rs.handleScale()
}

// GetExpectInstanceNumber get function minInstance in etcd
func (rs *ReplicaScaler) GetExpectInstanceNumber() int {
	rs.RLock()
	enable := rs.enable
	currentReplicaNum := rs.targetRsvInsNum
	rs.RUnlock()
	if !enable {
		return 0
	}
	return currentReplicaNum
}

// Destroy will destroy scaler
func (rs *ReplicaScaler) Destroy() {
	rs.Lock()
	rs.enable = false
	rs.targetRsvInsNum = 0
	rs.Unlock()
}

func (rs *ReplicaScaler) handlePendingInsNumIncrease(insDiff int) {
	rs.Lock()
	rs.pendingRsvInsNum += insDiff
	rs.Unlock()
}

func (rs *ReplicaScaler) handlePendingInsNumDecrease(insDiff int) {
	rs.Lock()
	rs.pendingRsvInsNum -= insDiff
	rs.Unlock()
}

func (rs *ReplicaScaler) handleScale() {
	rs.RLock()
	enable := rs.enable
	scaleNum := rs.targetRsvInsNum - rs.currentRsvInsNum
	rs.logger.Infof("parameters for handle scale of function targetRsvInsNum %d currentRsvInsNum %d "+
		"pendingRsvInsNum %d scaleNum %d", rs.targetRsvInsNum, rs.currentRsvInsNum, rs.pendingRsvInsNum, scaleNum)
	if scaleNum >= 0 && (!enable || config.GlobalConfig.DisableReplicaScaler) {
		rs.logger.Warnf("replicaScaler disable, targetNum is %d, currentNum is %d, pendingNum is %d",
			rs.targetRsvInsNum, rs.currentRsvInsNum, rs.pendingRsvInsNum)
		rs.RUnlock()
		return
	}
	rs.RUnlock()
	if scaleNum > 0 {
		rs.Lock()
		if rs.pendingRsvInsNum >= 0 {
			scaleNum -= rs.pendingRsvInsNum
			if scaleNum <= 0 {
				rs.logger.Warnf("scaleNum <= 0, no need to scale up, "+
					"targetNum is %d, currentNum is %d, pendingNum is %d",
					rs.targetRsvInsNum, rs.currentRsvInsNum, rs.pendingRsvInsNum)
				rs.Unlock()
				return
			}
		}
		rs.pendingRsvInsNum += scaleNum
		rs.Unlock()
		rs.logger.Infof("calculate scale up instance number is %d", scaleNum)
		rs.scaleUpHandler(scaleNum, rs.handlePendingInsNumDecrease)
	} else if scaleNum < 0 {
		rs.Lock()
		if rs.pendingRsvInsNum <= 0 {
			scaleNum -= rs.pendingRsvInsNum
			if scaleNum >= 0 {
				rs.logger.Warnf("scaleNum >= 0, no need to scale down, "+
					"targetNum is %d, currentNum is %d, pendingNum is %d",
					rs.targetRsvInsNum, rs.currentRsvInsNum, rs.pendingRsvInsNum)
				rs.Unlock()
				return
			}
		}
		rs.pendingRsvInsNum += scaleNum
		rs.Unlock()
		rs.logger.Infof("calculate scale down instance number is %d", -scaleNum)
		rs.scaleDownHandler(-scaleNum, rs.handlePendingInsNumIncrease)
	}
}
