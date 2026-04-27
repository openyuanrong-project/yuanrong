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
	"math"
	"sync"
	"sync/atomic"
	"time"

	"yuanrong.org/kernel/pkg/common/faas_common/instanceconfig"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	commonUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/metrics"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

const (
	// SLA time should not be shorter than cold start time
	minSLATime             = time.Duration(500) * time.Millisecond
	defaultScaleUpInitTime = 100
)

var scaleUpInitTime = time.Duration(defaultScaleUpInitTime) * time.Millisecond

// AutoScaler will scales instance automatically based on calculation upon instance metrics
type AutoScaler struct {
	enable           atomic.Bool
	metricsCollector metrics.Collector
	funcKeyWithRes   string
	scaleUpWindow    time.Duration
	scaleDownWindow  time.Duration
	coldStartTime    time.Duration
	concurrentNum    int
	// create new instance may take longer than scale up window, uses pendingInsThdNum to record these instances to
	// avoid excessive scaling
	pendingInsThdNum int
	inUseInsThdNum   int
	totalInsThdNum   int
	// remainedInsThdReqNum stores this difference between pendingInsThdReqNum and scaleInsThdNum if scaleInsTheNum is
	// less than pendingInsThdReqNum
	remainedInsThdReqNum int
	// autoScaleUpFlag tells if auto scale up process is running
	autoScaleUpFlag bool
	// autoScaleDownFlag tells if auto scale up process is running
	autoScaleDownFlag bool
	checkReqNumFunc   func() int
	scaleUpHandler    ScaleUpHandler
	scaleDownHandler  ScaleDownHandler
	// scaleUpTriggerCh triggers auto scale up process, we will trigger auto scale up process if metrics are collected,
	// before then we will manually create instance for each instance thread request
	scaleUpTriggerCh chan struct{}
	// scaleDownTriggerCh triggers auto scale down process, currently it is triggered as soon as autoScaler is created,
	// consider to control the scale down timing to make scale down process more efficient
	scaleDownTriggerCh chan struct{}
	stopCh             chan struct{}
	sync.RWMutex
}

// NewAutoScaler will create a AutoScaler
func NewAutoScaler(funcKeyWithRes string, metricsCollector metrics.Collector, checkReqNumFunc CheckReqNumFunc,
	scaleUpHandler ScaleUpHandler, scaleDownHandler ScaleDownHandler, autoScaleConfig types.AutoScaleConfig,
) InstanceScaler {
	log.GetLogger().Debugf("autoScaleConfig: %v", autoScaleConfig)

	scaleUpWindow := time.Duration(autoScaleConfig.SLAQuota) * time.Millisecond
	if scaleUpWindow < minSLATime {
		scaleUpWindow = minSLATime
	}
	scaleDownWindow := time.Duration(autoScaleConfig.ScaleDownTime) * time.Millisecond
	if scaleDownWindow < scaleUpWindow {
		scaleDownWindow = scaleUpWindow
	}
	autoScaler := &AutoScaler{
		funcKeyWithRes:     funcKeyWithRes,
		metricsCollector:   metricsCollector,
		scaleUpWindow:      scaleUpWindow,
		scaleDownWindow:    scaleDownWindow,
		checkReqNumFunc:    checkReqNumFunc,
		scaleUpHandler:     scaleUpHandler,
		scaleDownHandler:   scaleDownHandler,
		enable:             atomic.Bool{},
		scaleUpTriggerCh:   make(chan struct{}, 1),
		scaleDownTriggerCh: make(chan struct{}, 1),
		stopCh:             make(chan struct{}),
	}
	if scaleUpInitTime > autoScaler.scaleUpWindow {
		scaleUpInitTime = autoScaler.scaleUpWindow
	}
	go autoScaler.scaleUpLoop()
	// Abandoned before 20250330
	if config.GlobalConfig.Scenario != types.ScenarioWiseCloud {
		go autoScaler.scaleDownLoop()
	}
	return autoScaler
}

// SetEnable will configure the enable of scaler
func (as *AutoScaler) SetEnable(enable bool) {
	as.enable.Store(enable)
}

// TriggerScale will trigger scale
func (as *AutoScaler) TriggerScale() {
	as.Lock()
	if !as.autoScaleUpFlag {
		as.autoScaleUpFlag = true
		as.scaleUpTriggerCh <- struct{}{}
	}
	as.Unlock()
}

// CheckScaling will check if scaler is scaling
func (as *AutoScaler) CheckScaling() bool {
	isScaling := false
	as.RLock()
	isScaling = as.autoScaleUpFlag || as.autoScaleDownFlag
	as.RUnlock()
	return isScaling
}

// GetExpectInstanceNumber - number of pending and running instance
func (as *AutoScaler) GetExpectInstanceNumber() int {
	as.RLock()
	expectNum := (as.totalInsThdNum + as.pendingInsThdNum) / as.concurrentNum
	as.RUnlock()
	return expectNum
}

// UpdateCreateMetrics will update create metrics
func (as *AutoScaler) UpdateCreateMetrics(coldStartTime time.Duration) {
	as.Lock()
	if coldStartTime > as.coldStartTime {
		as.coldStartTime = coldStartTime
		log.GetLogger().Infof("cold start time for function %s is updated to %d ms", as.funcKeyWithRes,
			coldStartTime.Milliseconds())
	}
	as.Unlock()
}

// HandleInsThdUpdate will update instance thread metrics, totalInsThd increase should be coupled with pendingInsThd
// decrease for better consistency
func (as *AutoScaler) HandleInsThdUpdate(inUseInsThdDiff, totalInsThdDiff int) {
	as.metricsCollector.UpdateInsThdMetrics(inUseInsThdDiff)
	as.Lock()
	as.inUseInsThdNum += inUseInsThdDiff
	as.totalInsThdNum += totalInsThdDiff
	// trigger scale down process if curRsvInsThdNum != 0
	if (as.totalInsThdNum - as.inUseInsThdNum) >= as.concurrentNum {
		select {
		case as.scaleDownTriggerCh <- struct{}{}:
		default:
			log.GetLogger().Warnf("scale down channel blocks for function %s", as.funcKeyWithRes)
		}
	}
	as.Unlock()
}

// HandleFuncSpecUpdate -
func (as *AutoScaler) HandleFuncSpecUpdate(funcSpec *types.FunctionSpecification) {
	as.Lock()
	as.concurrentNum = utils.GetConcurrentNum(funcSpec.InstanceMetaData.ConcurrentNum)
	as.Unlock()
	log.GetLogger().Infof("config concurrentNum to %d for auto scaler %s", as.concurrentNum, as.funcKeyWithRes)
}

// HandleInsConfigUpdate -
func (as *AutoScaler) HandleInsConfigUpdate(insConfig *instanceconfig.Configuration) {
}

// HandleCreateError handles instance create error
func (as *AutoScaler) HandleCreateError(createError error) {
}

// Destroy will destroy scaler
func (as *AutoScaler) Destroy() {
	commonUtils.SafeCloseChannel(as.stopCh)
}

func (as *AutoScaler) scaleUpLoop() {
	// unlike scale down, scale up process should start right away, so we define scale up channel this way to manually
	// trigger first round's scale up before ticker fires
	scaleUpChan := make(chan struct{}, 1)
	tickerChan := make(<-chan time.Time, 1)
	var ticker *time.Ticker
	for {
		select {
		case _, ok := <-scaleUpChan:
			if !ok {
				log.GetLogger().Warnf("scale up channel is closed")
				return
			}
			insThdReq := as.checkReqNumFunc()
			if insThdReq == 0 {
				as.pauseScale(ticker)
				continue
			}
			as.scaleUpInstances()
		case _, ok := <-as.scaleUpTriggerCh:
			if !ok {
				log.GetLogger().Warnf("trigger channel is closed")
				return
			}
			// let requests come in for certain time to calculate a more reasonable scale up number
			time.Sleep(scaleUpInitTime)
			scaleUpChan <- struct{}{}
			ticker = time.NewTicker(as.scaleUpWindow)
			tickerChan = ticker.C
			log.GetLogger().Infof("scale up loop for function %s is running", as.funcKeyWithRes)
		case <-tickerChan:
			scaleUpChan <- struct{}{}
		case <-as.stopCh:
			log.GetLogger().Warnf("stop scale up loop for function %s now", as.funcKeyWithRes)
			return
		}
	}
}

func (as *AutoScaler) pauseScale(ticker *time.Ticker) {
	if ticker != nil {
		ticker.Stop()
	}
	as.Lock()
	as.autoScaleUpFlag = false
	as.Unlock()
	log.GetLogger().Infof("scale up loop for function %s is paused", as.funcKeyWithRes)
}

func (as *AutoScaler) scaleDownLoop() {
	scaleDownChan := make(<-chan time.Time, 1)
	var timer *time.Timer
	for {
		select {
		case <-scaleDownChan:
			as.scaleDownInstances()
			if timer != nil {
				timer.Stop()
			}
			as.Lock()
			as.autoScaleDownFlag = false
			as.Unlock()
			log.GetLogger().Infof("scale down loop for function %s is paused", as.funcKeyWithRes)
		case _, ok := <-as.scaleDownTriggerCh:
			if !ok {
				log.GetLogger().Warnf("trigger channel is closed")
				return
			}
			as.Lock()
			as.autoScaleDownFlag = true
			as.Unlock()
			if timer == nil {
				timer = time.NewTimer(as.scaleDownWindow)
			} else {
				select {
				case <-timer.C:
				default:
				}
				timer.Reset(as.scaleDownWindow)
			}
			scaleDownChan = timer.C
			log.GetLogger().Infof("scale down loop for function %s is running", as.funcKeyWithRes)
		case <-as.stopCh:
			log.GetLogger().Warnf("stop scale down loop for function %s now", as.funcKeyWithRes)
			return
		}
	}
}

func (as *AutoScaler) handlePendingInsNumIncrease(insDiff int) {
	as.Lock()
	as.pendingInsThdNum += insDiff * as.concurrentNum
	as.Unlock()
}

func (as *AutoScaler) handlePendingInsNumDecrease(insDiff int) {
	as.Lock()
	as.pendingInsThdNum -= insDiff * as.concurrentNum
	as.Unlock()
}

func (as *AutoScaler) scaleUpInstances() {
	as.Lock()
	defer as.Unlock()
	pendingInsThdReqNum := float64(as.checkReqNumFunc())
	if !as.metricsCollector.InvokeMetricsCollected() {
		// fire at will if no metrics is ever collected for this function
		scaleInsThdNum := pendingInsThdReqNum
		// when request triggers scale up, there must be no availInsThdNum, no need to calculate with availInsThdNum,
		// be aware that scaleInsThdNum is unsigned but pendingInsThdNum is signed
		scaleInsThdNum = math.Max(scaleInsThdNum-float64(as.pendingInsThdNum), 0)
		scaleInsNum := int(math.Ceil(scaleInsThdNum / float64(as.concurrentNum)))
		as.pendingInsThdNum += scaleInsNum * as.concurrentNum
		log.GetLogger().Infof("calculated scale up instance number for function %s is %d", as.funcKeyWithRes,
			scaleInsNum)
		as.scaleUpHandler(scaleInsNum, as.handlePendingInsNumDecrease)
		return
	}
	avgProcTime, insThdProcNumPS, insThdReqNumPS := as.metricsCollector.GetCalculatedInvokeMetrics()
	log.GetLogger().Infof("parameters for calculating scale up of function %s avgProcTime %f insThdProcNumPS %f "+
		"insThdReqNumPS %f pendingInsThdReqNum %f pendingInsThdNum %d", as.funcKeyWithRes,
		avgProcTime, insThdProcNumPS, insThdReqNumPS, pendingInsThdReqNum, as.pendingInsThdNum)
	if insThdProcNumPS == 0 {
		log.GetLogger().Errorf("invalid value for insThdProcNumPS")
		return
	}
	// when request triggers scale up, there must be no availInsThdNum, no need to calculate with availInsThdNum
	procNumToDo := pendingInsThdReqNum + insThdReqNumPS*as.scaleUpWindow.Seconds()
	procCapAvail := insThdProcNumPS * float64(as.inUseInsThdNum) * math.Max(as.scaleUpWindow.Seconds()-avgProcTime, 0)
	procNumToDo = math.Max(procNumToDo-procCapAvail, 0)
	procWindow := as.scaleUpWindow.Seconds() - as.coldStartTime.Seconds()
	// handle the exception case that cold start takes longer than scale up window, set scaleInsThdNum to difference
	// between procNumToDo and pendingInsThdNum to avoid scale up 0
	scaleInsThdNum := float64(0)
	if procWindow > 0 {
		scaleInsThdNum = math.Max(math.Ceil(procNumToDo/insThdProcNumPS/procWindow), 0)
	} else {
		scaleInsThdNum = math.Max(math.Ceil(procNumToDo), 0)
	}
	// try to scale less than pendingInsThdReqNum to take most advantage of instance reuse, considering these cases:
	// 1. insThdProcNumPS is relatively large and one instance thread can process several requests during this scaling
	//    window
	// 2. in-used instance thread may be released to faas scheduler during autoscaling which can be reused
	scaleInsThdNum = math.Min(scaleInsThdNum, pendingInsThdReqNum)
	// check remainedInsThdReqNum from last round, if it's less than pendingInsThdReqNum (otherwise some requests in
	// remainedInsThdReqNum must be fulfilled and we just need to calculate with pendingInsThdReqNum in previous step),
	// calibrate the scaleInsNum to avoid these remained instance thread requests being remained again
	if float64(as.remainedInsThdReqNum) < pendingInsThdReqNum {
		scaleInsThdNum = math.Max(scaleInsThdNum, float64(as.remainedInsThdReqNum))
	}
	// calibrate scaleInsThdNum with pendingInsThdNum to avoid excessive scaling, be aware that scaleInsThdNum is
	// unsigned but pendingInsThdNum is signed
	scaleInsThdNum = math.Max(scaleInsThdNum-float64(as.pendingInsThdNum), 0)
	// scaleInsThdNum may be smaller than pendingInsThdReqNum if insThdProcNumPS is relatively large, in this case
	// remainedInsThdReqNum will store this difference between pendingInsThdReqNum and scaleInsThdNum to tell next
	// round's scaleUpInstances there may be several insThdReqs unfulfilled which it should be aware
	as.remainedInsThdReqNum = int(math.Max(pendingInsThdReqNum-scaleInsThdNum, 0))
	scaleInsNum := int(math.Ceil(scaleInsThdNum / float64(as.concurrentNum)))
	log.GetLogger().Infof("calculated scale up instance number for function %s is %d", as.funcKeyWithRes, scaleInsNum)
	if scaleInsNum > 0 {
		as.pendingInsThdNum += scaleInsNum * as.concurrentNum
		as.scaleUpHandler(scaleInsNum, as.handlePendingInsNumDecrease)
	}
}

func (as *AutoScaler) scaleDownInstances() {
	if !as.metricsCollector.InvokeMetricsCollected() {
		as.Lock()
		scaleInsThdNum := math.Max(float64(as.totalInsThdNum-as.inUseInsThdNum), 0)
		// be aware that scaleInsThdNum is unsigned but pendingInsThdNum is signed
		scaleInsThdNum = -math.Min(-scaleInsThdNum-float64(as.pendingInsThdNum), 0)
		scaleInsNum := int(math.Floor(scaleInsThdNum / float64(as.concurrentNum)))
		as.pendingInsThdNum -= scaleInsNum * as.concurrentNum
		as.Unlock()
		as.scaleDownHandler(scaleInsNum, as.handlePendingInsNumIncrease)
		return
	}
	avgProcTime, insThdProcNumPS, insThdReqNumPS := as.metricsCollector.GetCalculatedInvokeMetrics()
	log.GetLogger().Infof("parameters for calculating scale down of function %s avgProcTime %f insThdProcNumPS %f "+
		"insThdReqNumPS %f totalInsThdNum %d pendingInsThdNum %d", as.funcKeyWithRes, avgProcTime, insThdProcNumPS,
		insThdReqNumPS, as.totalInsThdNum, as.pendingInsThdNum)
	if insThdProcNumPS == 0 {
		log.GetLogger().Errorf("invalid value for insThdProcNumPS")
		return
	}
	as.Lock()
	defer as.Unlock()
	procNumToDo := insThdReqNumPS * as.scaleDownWindow.Seconds()
	procCapAvail := insThdProcNumPS * (math.Max(float64(as.totalInsThdNum-as.inUseInsThdNum), 0)*
		as.scaleDownWindow.Seconds() + float64(as.inUseInsThdNum)*math.Max(as.scaleUpWindow.Seconds()-avgProcTime, 0))
	procCapExcess := math.Max(procCapAvail-procNumToDo, 0)
	scaleInsThdNum := math.Ceil(procCapExcess / insThdProcNumPS / as.scaleDownWindow.Seconds())
	// be aware that scaleInsThdNum is unsigned but pendingInsThdNum is signed
	scaleInsThdNum = -math.Min(-scaleInsThdNum-float64(as.pendingInsThdNum), 0)
	scaleInsNum := int(math.Floor(scaleInsThdNum / float64(as.concurrentNum)))
	log.GetLogger().Infof("calculated scale down instance number for function %s is %d", as.funcKeyWithRes, scaleInsNum)
	if scaleInsNum > 0 {
		as.pendingInsThdNum -= scaleInsNum * as.concurrentNum
		as.scaleDownHandler(scaleInsNum, as.handlePendingInsNumIncrease)
	}
}
