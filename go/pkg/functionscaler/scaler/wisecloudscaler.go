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
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/instanceconfig"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	"yuanrong.org/kernel/pkg/common/faas_common/statuscode"
	commonUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/common/faas_common/wisecloudtool"
	wisecloudTypes "yuanrong.org/kernel/pkg/common/faas_common/wisecloudtool/types"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

const (
	errorExpireTime = 10 * time.Second // 失败后等10s重置unrecovererr
)

var (
	localColdStarter     *wisecloudtool.PodOperator
	localColdStarterOnce sync.Once
)

func newColdStarter(logger api.FormatLogger) *wisecloudtool.PodOperator {
	localColdStarterOnce.Do(func() {
		localColdStarter = wisecloudtool.NewColdStarter(&config.GlobalConfig.ServiceAccountJwt, logger)
	})
	return localColdStarter
}

// WiseCloudScaler will scales instance automatically based on calculation upon instance metrics
type WiseCloudScaler struct {
	funcKeyWithRes  string
	nuwaRuntimeInfo *wisecloudTypes.NuwaRuntimeInfo
	resSpec         resspeckey.ResSpecKey
	logger          api.FormatLogger

	podOperator *wisecloudtool.PodOperator

	CreateCallback   func(err error)
	totalInsThdNum   int
	concurrentNum    int
	isReserve        bool
	enableScale      bool
	coldStartTrigger chan struct{}
	stopCh           chan struct{}
	sync.RWMutex
}

// NewWiseCloudScaler will create a WiseCloudScaler
func NewWiseCloudScaler(funcKeyWithRes string, resSpec resspeckey.ResSpecKey,
	isReserve bool, createCallback func(err error)) InstanceScaler {
	stopCh := make(chan struct{})
	wiseCloudScaler := &WiseCloudScaler{
		funcKeyWithRes:   funcKeyWithRes,
		resSpec:          resSpec,
		podOperator:      newColdStarter(log.GetLogger().With(zap.Any("funcKey", funcKeyWithRes))),
		logger:           log.GetLogger().With(zap.Any("funcKey", funcKeyWithRes)),
		CreateCallback:   createCallback,
		totalInsThdNum:   0,
		concurrentNum:    1,
		isReserve:        isReserve,
		enableScale:      false,
		coldStartTrigger: make(chan struct{}, 1),
		stopCh:           stopCh,
	}
	if isReserve {
		go wiseCloudScaler.coldStartLoop()
	}
	return wiseCloudScaler
}

// SetEnable will configure the enable of scaler
func (as *WiseCloudScaler) SetEnable(enable bool) {
	if as.isReserve {
		as.logger.Infof("set enable, from %v to %v", as.enableScale, enable)
		as.enableScale = enable
	}
}

// DelNuwaPod will send a req to erase runtime pod
func (as *WiseCloudScaler) DelNuwaPod(ins *types.Instance) error {
	if ins == nil {
		return fmt.Errorf("ins is nil, skip")
	}
	var err error
	for {
		select {
		case <-as.stopCh:
			as.logger.Warnf("stop del nuwa pod %s loop, now", ins.PodID)
			return err
		default:
			if as.nuwaRuntimeInfo == nil {
				as.logger.Errorf("failed del nuwa pod %s, nuwaRuntimeInfo empty", ins.PodID)
				time.Sleep(time.Second)
				continue
			}
			err = as.podOperator.DelPod(as.nuwaRuntimeInfo, ins.PodDeploymentName, ins.PodID)
			if err != nil {
				as.logger.Errorf("failed del nuwa pod %s, %s", ins.PodID, err.Error())
			}
			return err
		}
	}

}

// TriggerScale will trigger scale
func (as *WiseCloudScaler) TriggerScale() {
	as.logger.Infof("trigger scale, enablescale: %v, totalInsThdNum: %v", as.enableScale, as.totalInsThdNum)
	if as.enableScale && as.totalInsThdNum == 0 {
		select {
		case as.coldStartTrigger <- struct{}{}:
		default:
			as.logger.Debugf("scale has been trigger, skip")
		}
	}
}

// CheckScaling will check if scaler is scaling
func (as *WiseCloudScaler) CheckScaling() bool {
	return false
}

// GetExpectInstanceNumber - number of pending and running instance
func (as *WiseCloudScaler) GetExpectInstanceNumber() int {
	as.RLock()
	defer as.RUnlock()
	expectNum := as.totalInsThdNum / as.concurrentNum
	return expectNum
}

// UpdateCreateMetrics will update create metrics
func (as *WiseCloudScaler) UpdateCreateMetrics(coldStartTime time.Duration) {
}

// HandleInsThdUpdate will update instance thread metrics, totalInsThd increase should be coupled with pendingInsThd
// decrease for better consistency
func (as *WiseCloudScaler) HandleInsThdUpdate(inUseInsThdDiff, totalInsThdDiff int) {
	as.Lock()
	defer as.Unlock()
	as.totalInsThdNum += totalInsThdDiff
}

// HandleFuncSpecUpdate -
func (as *WiseCloudScaler) HandleFuncSpecUpdate(funcSpec *types.FunctionSpecification) {
	as.Lock()
	defer as.Unlock()
	as.concurrentNum = utils.GetConcurrentNum(funcSpec.InstanceMetaData.ConcurrentNum)
}

// HandleInsConfigUpdate -
func (as *WiseCloudScaler) HandleInsConfigUpdate(insConfig *instanceconfig.Configuration) {
	if insConfig == nil || insConfig.NuwaRuntimeInfo.WisecloudRuntimeId == "" {
		return
	}
	as.nuwaRuntimeInfo = &insConfig.NuwaRuntimeInfo
}

// HandleCreateError handles instance create error
func (as *WiseCloudScaler) HandleCreateError(createError error) {
}

// Destroy will destroy scaler
func (as *WiseCloudScaler) Destroy() {
	commonUtils.SafeCloseChannel(as.stopCh)
}

func (as *WiseCloudScaler) coldStartLoop() {
	for {
		select {
		case _, ok := <-as.coldStartTrigger:
			if !ok {
				as.logger.Warnf("trigger channel is closed")
				return
			}
			if as.nuwaRuntimeInfo == nil {
				as.logger.Warnf("nuwa runtime info is empty, skip")
				continue
			}
			if as.totalInsThdNum != 0 {
				continue
			}
			err := as.podOperator.ColdStart(as.funcKeyWithRes, as.resSpec, as.nuwaRuntimeInfo)
			if err != nil {
				as.logger.Errorf("cold start failed, err %s", err.Error())
				as.CreateCallback(snerror.New(statuscode.WiseCloudNuwaColdStartErrCode, "cold start failed, err: "+err.Error()))
				go func() {
					time.Sleep(errorExpireTime)
					as.CreateCallback(nil)
				}()
				continue
			}
			as.CreateCallback(nil)
			as.logger.Infof("cold start succeed")
		case <-as.stopCh:
			as.logger.Warnf("stop cold create loop for function %s now", as.funcKeyWithRes)
			return
		}
	}
}
