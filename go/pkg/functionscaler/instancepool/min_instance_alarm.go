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

// Package instancepool -
package instancepool

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/alarm"
	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

var (
	checkInterval              time.Duration
	defaultStartInterval       = 15
	defaultMinInsCheckInterval = 15
	insufficientNumMap         = make(map[int64]int64)
)

const uint64Base = 10

func (pm *PoolManager) checkMinInsRegularly(stopCh <-chan struct{}) {
	startIntervalNum := config.GlobalConfig.AlarmConfig.MinInsStartInterval
	if startIntervalNum == 0 {
		startIntervalNum = defaultStartInterval
	}
	startInterval := time.Duration(startIntervalNum) * time.Minute
	checkIntervalNum := config.GlobalConfig.AlarmConfig.MinInsCheckInterval
	if checkIntervalNum == 0 {
		checkIntervalNum = defaultMinInsCheckInterval
	}
	checkInterval = time.Duration(checkIntervalNum) * time.Minute

	if stopCh == nil {
		log.GetLogger().Errorf("stopCh is nil")
		return
	}
	clusterID := config.GlobalConfig.ClusterID
	if clusterID == "" {
		log.GetLogger().Warnf("failed to get cluster ID")
	}

	log.GetLogger().Infof("start to check min instance with after %v", startInterval)
	log.GetLogger().Infof("start to check min instance with interval %v", checkInterval)

	time.AfterFunc(startInterval, func() {
		ticker := time.NewTicker(checkInterval)
		for {
			select {
			case <-stopCh:
				log.GetLogger().Infof("stop check min instance")
				ticker.Stop()
				return
			case <-ticker.C:
				pm.judgeAndReport(clusterID)
			}
		}
	})
}

func (pm *PoolManager) judgeAndReport(clusterID string) {
	insufficientNum := pm.checkMinScale(clusterID)
	if insufficientNum > 0 {
		errInfo := fmt.Sprintf("Insufficient number of reserved instances: %d, cluster: %s",
			insufficientNum, clusterID)
		reportInsufficientAlarm(errInfo, insufficientNum)
		log.GetLogger().Errorf(errInfo)
		return
	} else if insufficientNum == 0 {
		log.GetLogger().Infof("clear ins Insufficient alarm")
		clearInsufficientAlarm(clusterID)
		return
	}
	log.GetLogger().Warnf("insufficientNum is under zero: %d", insufficientNum)
}

func (pm *PoolManager) checkMinScale(clusterID string) int64 {
	logger := log.GetLogger().With(zap.Any("clusterID", clusterID))
	logger.Infof("start to check min scale")
	res := int64(0)
	pm.RLock()
	defer pm.RUnlock()
	for funcKey, pool := range pm.instancePool {
		funcLogger := logger.With(zap.Any("funcKey", funcKey))
		gi, ok := pool.(*GenericInstancePool)
		if !ok {
			logger.Warnf("is not a generic instance pool, skip")
			continue
		}
		if gi.FuncSpec.InstanceMetaData.ScalePolicy == types.InstanceScalePolicyStaticFunction {
			continue
		}
		gi.RLock()
		for label, _ := range gi.insConfig {
			res += pm.checkMinScaleForLabel(gi, label, funcLogger)
		}
		gi.RUnlock()
	}
	logger.Infof("finish to check min scale")
	return res
}

func (pm *PoolManager) checkMinScaleForLabel(gi *GenericInstancePool, resKey resspeckey.ResSpecKey,
	logger api.FormatLogger) int64 {
	res := int64(0)
	if gi.reservedInstanceQueue[resKey] == nil || gi.insConfig[resKey] == nil {
		logger.Warnf("resource %+v in this faas scheduler does not have reserved instance queues", resKey)
		return res
	}
	if gi.insConfig[resKey].InstanceMetaData.MinInstance == 0 {
		logger.Infof("funcKey: %s reserved ins expectNum is 0", gi.FuncSpec.FuncKey)
		return res
	}
	if time.Now().Sub(gi.minScaleUpdatedTime) > checkInterval {
		actualNum := gi.reservedInstanceQueue[resKey].GetInstanceNumber(true)
		expectNum := int(gi.insConfig[resKey].InstanceMetaData.MinInstance)
		logger.Infof("check min instance, currentNum: %d, expectNum: %d", actualNum, expectNum)
		if actualNum < expectNum {
			res += int64(expectNum) - int64(actualNum)
			logger.Errorf("actual reserved instance num %d, configured minScale num %d",
				actualNum, expectNum)
			gi.minScaleAlarmSign[resKey.InvokeLabel] = true
		} else if gi.minScaleAlarmSign[resKey.InvokeLabel] == true {
			logger.Warnf("the number of reserved instances of reaches the set value %d", expectNum)
			gi.minScaleAlarmSign[resKey.InvokeLabel] = false
		}
	}
	return res
}

func reportInsufficientAlarm(errMsg string, insufficientNum int64) {
	alarmInfo := &alarm.LogAlarmInfo{
		AlarmID:    alarm.InsufficientMinInstance00001,
		AlarmName:  "InsufficientMinInstance",
		AlarmLevel: alarm.Level2,
	}
	alarmDetail := &alarm.Detail{
		SourceTag: os.Getenv(constant.PodNameEnvKey) + "|" + os.Getenv(constant.PodIPEnvKey) +
			"|" + os.Getenv(constant.ClusterName) + "|InsufficientMinInstance" +
			strconv.FormatInt(insufficientNum, uint64Base),
		OpType:         alarm.GenerateAlarmLog,
		Details:        errMsg,
		StartTimestamp: int(time.Now().Unix()),
		EndTimestamp:   0,
	}

	if _, ok := insufficientNumMap[insufficientNum]; !ok {
		insufficientNumMap[insufficientNum] = 1
	}
	alarm.ReportOrClearAlarm(alarmInfo, alarmDetail)
}

func clearInsufficientAlarm(clusterID string) {
	alarmInfo := &alarm.LogAlarmInfo{
		AlarmID:    alarm.InsufficientMinInstance00001,
		AlarmName:  "InsufficientMinInstance",
		AlarmLevel: alarm.Level2,
	}
	alarmDetail := &alarm.Detail{
		OpType:         alarm.ClearAlarmLog,
		Details:        fmt.Sprintf("The number of reserved instances is normal, cluster: %s", clusterID),
		StartTimestamp: 0,
		EndTimestamp:   int(time.Now().Unix()),
	}

	for k := range insufficientNumMap {
		alarmDetail.SourceTag = os.Getenv(constant.PodNameEnvKey) + "|" + os.Getenv(constant.PodIPEnvKey) +
			"|" + os.Getenv(constant.ClusterName) + "|InsufficientMinInstance" + strconv.FormatInt(k, uint64Base)
		alarm.ReportOrClearAlarm(alarmInfo, alarmDetail)
		delete(insufficientNumMap, k)
	}
}
