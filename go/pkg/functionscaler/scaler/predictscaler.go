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
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/instanceconfig"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	commonUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/metrics"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

const (
	// Millisecond2Second -
	Millisecond2Second = 1000
	// Minute2Second -
	Minute2Second = 60
	// DefaultReqDelay -
	DefaultReqDelay = 3000

	predictEtcdPrefix    = "/arrivalPrediction/"
	dataSetTimeWindowLen = 2
	qpsGroupsNum         = 2
	millisecondToMinute  = 60 * 1000
	longColdStartTime    = 3 * 60 * 1000 * time.Millisecond
)

// PredictScaler will scale based on a predict instance number
type PredictScaler struct {
	logger               api.FormatLogger
	metricsCollector     metrics.Collector
	funcKeyWithRes       string
	scaleUpWindow        time.Duration
	scaleDownWindow      time.Duration
	coldStartTime        time.Duration
	predictGroupWindow   int64 // Grouping window for prediction results
	concurrentNum        int
	pendingInsThdNum     int
	inUseInsThdNum       int // The number of leases used by this function version
	totalInsThdNum       int // The total number of leases that this function version can provide
	remainedInsThdReqNum int
	predictInsNum        int
	predictDownDiff      int
	predictUpDiff        int
	minRsvInsNum         int
	predictScaleUpFlag   bool
	predictScaleDownFlag bool
	predictUpChanFlag    bool
	predictDownChanFlag  bool
	enable               bool
	checkReqNumFunc      func() int
	scaleUpHandler       ScaleUpHandler
	scaleDownHandler     ScaleDownHandler
	// active, Triggered by prediction results
	scaleUpChan   chan struct{}
	scaleDownChan chan struct{}
	// passive, Triggered based on real-time lease application status
	scaleUpTriggerCh   chan struct{}
	scaleDownTriggerCh chan struct{}
	stopCh             chan struct{}
	sync.RWMutex
	sync.Once
}

// NewPredictScaler will create a PredictScaler
func NewPredictScaler(funcKeyWithRes string, metricsCollector metrics.Collector, checkReqNumFunc CheckReqNumFunc,
	scaleUpHandler ScaleUpHandler, scaleDownHandler ScaleDownHandler) InstanceScaler {
	scaleUpWindow := time.Duration(config.GlobalConfig.AutoScaleConfig.SLAQuota) * time.Millisecond
	if scaleUpWindow < minSLATime {
		scaleUpWindow = minSLATime
	}
	scaleDownWindow := time.Duration(config.GlobalConfig.AutoScaleConfig.ScaleDownTime) * time.Millisecond
	if scaleDownWindow < scaleUpWindow {
		scaleDownWindow = scaleUpWindow
	}

	predictScaler := &PredictScaler{
		logger:             log.GetLogger().With(zap.Any("funcKeyWithRes", funcKeyWithRes)),
		metricsCollector:   metricsCollector,
		funcKeyWithRes:     funcKeyWithRes,
		scaleUpWindow:      scaleUpWindow,
		scaleDownWindow:    scaleDownWindow,
		predictGroupWindow: config.GlobalConfig.PredictGroupWindow,
		checkReqNumFunc:    checkReqNumFunc,
		predictInsNum:      0,
		enable:             false,
		scaleUpHandler:     scaleUpHandler,
		scaleDownHandler:   scaleDownHandler,
		scaleUpChan:        make(chan struct{}, 1),
		scaleDownChan:      make(chan struct{}, 1),
		scaleUpTriggerCh:   make(chan struct{}, 1),
		scaleDownTriggerCh: make(chan struct{}, 1),
		stopCh:             make(chan struct{}, 1),
	}
	predictScaler.logger.Infof("create predict scaler")

	go predictScaler.scaleUp()
	go predictScaler.scaleDown()
	return predictScaler
}

func (ps *PredictScaler) startPredictRegistry() {
	metaEtcdClient := etcd3.GetMetaEtcdClient()
	if metaEtcdClient == nil {
		ps.logger.Errorf("failed to get meta etcd")
		return
	}
	watcher := etcd3.NewEtcdWatcher(
		predictEtcdPrefix+os.Getenv(constant.ClusterName),
		func(event *etcd3.Event) bool {
			return false
		},
		ps.watcherHandler,
		ps.stopCh,
		metaEtcdClient)
	watcher.StartWatch()
}

func (ps *PredictScaler) watcherHandler(event *etcd3.Event) {
	ps.logger.Infof("handling predict event type %s key %s", event.Type, event.Key)
	if event.Type == etcd3.SYNCED {
		ps.logger.Infof("predict registry ready to receive etcd kv")
		return
	}
	switch event.Type {
	case etcd3.PUT:
		predictBytes := event.Value
		predictResult := getPredictFromEtcdValue(predictBytes)
		if predictResult == nil {
			return
		}
		if !checkPredictResultValid(predictResult, ps.coldStartTime) {
			ps.logger.Warnf("predict result is invalid, dataSetTimeWindow is in [%d, %d]",
				predictResult.DataSetTimeWindow[0], predictResult.DataSetTimeWindow[1])
			return
		}
		filterPredictFunction(predictResult)
		for funcKey, QPSNums := range predictResult.QPSResult {
			predictQPSGroups := &types.PredictQPSGroups{FuncKey: funcKey}
			predictQPSGroups.QPSGroups = groupQPSNumByPredictWindow(QPSNums, config.GlobalConfig.PredictGroupWindow)
			ps.HandlePredictUpdate(predictQPSGroups)
		}
	default:
	}
}

// getPredictFromEtcdValue parse the PredictResult from etcd value
func getPredictFromEtcdValue(etcdValue []byte) *types.PredictResult {
	if len(etcdValue) == 0 {
		return nil
	}
	predictResult := &types.PredictResult{}
	if err := json.Unmarshal(etcdValue, predictResult); err != nil {
		log.GetLogger().Errorf("failed to unmarshal etcd value to PredictResult, err: %s", err.Error())
		return nil
	}
	return predictResult
}

func checkPredictResultValid(predictResult *types.PredictResult, startTime time.Duration) bool {
	if len(predictResult.DataSetTimeWindow) != dataSetTimeWindowLen {
		return false
	}

	// Limitation: The instance cold start time must be less than the expansion and contraction window
	if config.GlobalConfig.PredictGroupWindow < startTime.Milliseconds() {
		return false
	}

	currentTimeStamp := time.Now().UnixMilli()
	// If there is no time to scaleUp for the next predictGroupWindow when faasScheduler is started,
	// the prediction results at this time are ignored.
	// valid :  dataSetTimeWindow[0] < currentTimeStamp < dataSetTimeWindow[0] + PredictGroupWindow - startTime
	if currentTimeStamp < predictResult.DataSetTimeWindow[0] || currentTimeStamp >
		predictResult.DataSetTimeWindow[0]+config.GlobalConfig.PredictGroupWindow-startTime.Milliseconds() {
		return false
	}

	return true
}

// before, QpsResult is kv of functionURN and QpsNum
// after, QpsResult is kv of funcKey and QpsNum
func filterPredictFunction(predictResult *types.PredictResult) {
	filterMap := make(map[string][]float64, len(predictResult.QPSResult))
	for functionURN, predictNum := range predictResult.QPSResult {
		if selfregister.GlobalSchedulerProxy.IsFuncOwner(functionURN) {
			filterMap[functionURN] = predictNum
		}
	}
	predictResult.QPSResult = filterMap
}

// len(QpsNums) should be 30, unit is minute
// In the future, the prediction instance strategy can be made configurable, such as average value and extreme value.
func groupQPSNumByPredictWindow(QPSNums []float64, predictWindow int64) []float64 {
	if predictWindow == 0 {
		return []float64{}
	}
	step := int(predictWindow / millisecondToMinute)
	var groupQPS []float64
	for i := 0; i < len(QPSNums); i += step {
		var totalNum float64
		for j := i + 1; j < i+step && j < len(QPSNums); j++ {
			totalNum += QPSNums[j]
		}
		avg := totalNum / float64(step*1.0)
		groupQPS = append(groupQPS, avg)
	}
	return groupQPS
}

// SetEnable will configure the enable of scaler
func (ps *PredictScaler) SetEnable(enable bool) {
}

// TriggerScale will trigger scale
func (ps *PredictScaler) TriggerScale() {
	ps.Lock()
	// 实时租约触发实例扩容
	if !ps.predictScaleUpFlag {
		ps.predictScaleUpFlag = true
		ps.scaleUpTriggerCh <- struct{}{}
	}
	ps.Unlock()
}

// CheckScaling will check if scaler is scaling
func (ps *PredictScaler) CheckScaling() bool {
	isScaling := false
	ps.RLock()
	isScaling = ps.predictScaleUpFlag || ps.predictScaleDownFlag
	ps.RUnlock()
	return isScaling
}

// UpdateCreateMetrics will update create metrics
func (ps *PredictScaler) UpdateCreateMetrics(coldStartTime time.Duration) {
}

// HandleInsThdUpdate will update instance thread metrics, totalInsThd increase should be coupled with pendingInsThd
// decrease for better consistency
func (ps *PredictScaler) HandleInsThdUpdate(inUseInsThdDiff, totalInsThdDiff int) {
	ps.metricsCollector.UpdateInsThdMetrics(inUseInsThdDiff)
	ps.Lock()
	ps.inUseInsThdNum += inUseInsThdDiff
	ps.totalInsThdNum += totalInsThdDiff
	if (ps.totalInsThdNum - ps.inUseInsThdNum) >= ps.concurrentNum {
		select {
		case ps.scaleDownTriggerCh <- struct{}{}:
		default:
			ps.logger.Warnf("scale down channel blocks")
		}
	}
	ps.Unlock()
}

// HandleFuncSpecUpdate -
func (ps *PredictScaler) HandleFuncSpecUpdate(funcSpec *types.FunctionSpecification) {
	ps.Lock()
	ps.concurrentNum = utils.GetConcurrentNum(funcSpec.InstanceMetaData.ConcurrentNum)
	ps.coldStartTime = time.Duration(funcSpec.ExtendedMetaData.Initializer.Timeout) * time.Second
	ps.Unlock()
	resSpec := resspeckey.ConvertResourceMetaDataToResSpec(funcSpec.ResourceMetaData)
	ps.funcKeyWithRes = fmt.Sprintf("%s-%s", funcSpec.FuncKey, resSpec.String())
	ps.logger = log.GetLogger().With(zap.Any("funcKeyWithRes", ps.funcKeyWithRes))
	ps.logger.Infof("config concurrentNum to %d for predict scaler, coldStartTime is %f s",
		ps.concurrentNum, ps.coldStartTime.Seconds())
}

// HandleInsConfigUpdate -
func (ps *PredictScaler) HandleInsConfigUpdate(insConfig *instanceconfig.Configuration) {
	replicaNum := int(insConfig.InstanceMetaData.MinInstance)
	if replicaNum < 0 {
		replicaNum = 0
	}
	ps.Lock()
	prevMinRsvInsNum := ps.minRsvInsNum
	ps.minRsvInsNum = replicaNum
	ps.Unlock()
	ps.logger.Infof("config reserved num from %d to %d for predict scaler", prevMinRsvInsNum, replicaNum)
	ps.Once.Do(func() {
		go ps.startPredictRegistry()
	})
}

// HandleCreateError handles instance create error
func (ps *PredictScaler) HandleCreateError(createError error) {
}

// GetExpectInstanceNumber get function minInstance in etcd
func (ps *PredictScaler) GetExpectInstanceNumber() int {
	ps.RLock()
	expectNum := (ps.totalInsThdNum + ps.pendingInsThdNum) / ps.concurrentNum
	ps.RUnlock()
	return expectNum
}

// Destroy will destroy scaler
func (ps *PredictScaler) Destroy() {
	commonUtils.SafeCloseChannel(ps.stopCh)
}

// HandlePredictUpdate Use prediction results
func (ps *PredictScaler) HandlePredictUpdate(prg *types.PredictQPSGroups) {
	if len(prg.QPSGroups) != qpsGroupsNum || !strings.HasPrefix(ps.funcKeyWithRes, prg.FuncKey) {
		ps.logger.Errorf("prg.QPSGroups is invalid")
		return
	}
	var avgProcTime float64
	if !ps.metricsCollector.InvokeMetricsCollected() {
		// if no metrics is ever collected
		avgProcTime = DefaultReqDelay
		ps.logger.Info("if no metrics is ever collected")
	} else {
		avgProcTime, _, _ = ps.metricsCollector.GetCalculatedInvokeMetrics()
		ps.logger.Infof("parameters avgProcTime %f ms", avgProcTime)
	}
	predictInsNum0 := int(math.Ceil(prg.QPSGroups[0] * avgProcTime /
		(float64(ps.concurrentNum) * Minute2Second * Millisecond2Second)))
	predictInsNum1 := int(math.Ceil(prg.QPSGroups[1] * avgProcTime /
		(float64(ps.concurrentNum) * Minute2Second * Millisecond2Second)))
	expectNum := ps.GetExpectInstanceNumber() // managed by predictScaler
	currentInsNum := expectNum + ps.minRsvInsNum
	ps.logger.Infof("forecast results: predictInsNum0 %d predictInsNum1 %d expectNum %d"+
		" ps.minRsvInsNum %d Current number of instances", predictInsNum0, predictInsNum1, expectNum, ps.minRsvInsNum)
	ps.logger.Infof("current number of instances %d", currentInsNum)

	ps.predictDownDiff = 0
	ps.predictUpDiff = 0
	ps.predictInsNum = int(math.Max(float64(predictInsNum0), float64(predictInsNum1)))
	predictInsNum0 = int(math.Max(float64(predictInsNum0), float64(ps.minRsvInsNum)))

	if currentInsNum > predictInsNum0 {
		ps.predictDownChanFlag = true
		ps.predictDownDiff = currentInsNum - predictInsNum0
		ps.scaleDownChan <- struct{}{}
	}

	// 为下个窗口期提前扩容，下个窗口期提前一个冷启时间开始扩容
	sleepTime := CalSleepTime(time.Now().UnixMilli(), ps.coldStartTime)
	time.Sleep(sleepTime)
	currentInsNum = ps.GetExpectInstanceNumber() + ps.minRsvInsNum
	if currentInsNum < predictInsNum1 {
		ps.predictUpDiff = predictInsNum1 - currentInsNum
		ps.scaleUpChan <- struct{}{}
	}
}

// CalSleepTime -
func CalSleepTime(currentTimeStamp int64, coldStartTime time.Duration) time.Duration {
	return time.Duration(config.GlobalConfig.PredictGroupWindow)*time.Millisecond - coldStartTime -
		time.Duration(currentTimeStamp%config.GlobalConfig.PredictGroupWindow)*time.Millisecond
}

func (ps *PredictScaler) getScaleUpInstancesNum() int {
	pendingInsThdReqNum := float64(ps.checkReqNumFunc())
	// fire at will if no metrics is ever collected for this function
	ps.Lock()
	scaleInsThdNum := math.Max(pendingInsThdReqNum-math.Max(float64(ps.pendingInsThdNum), 0), 0)
	scaleInsNum := int(math.Ceil(scaleInsThdNum / float64(ps.concurrentNum)))
	ps.Unlock()
	ps.logger.Infof("calculated scale up instance number is %d", scaleInsNum)
	if scaleInsNum > 0 {
		return scaleInsNum
	}
	return 0
}

func (ps *PredictScaler) getScaleDownInstancesNum() int {
	if !ps.metricsCollector.InvokeMetricsCollected() {
		// try to scale down instance even if no metrics is ever collected
		return 1
	}
	avgProcTime, insThdProcNumPS, insThdReqNumPS := ps.metricsCollector.GetCalculatedInvokeMetrics()
	ps.logger.Infof("parameters for calculating scale down avgProcTime %f insThdProcNumPS %f "+
		"insThdReqNumPS %f totalInsThdNum %d pendingInsThdNum %d", avgProcTime, insThdProcNumPS,
		insThdReqNumPS, ps.totalInsThdNum, ps.pendingInsThdNum)
	if insThdProcNumPS == 0 {
		ps.logger.Errorf("invalid value for insThdProcNumPS")
		return -1
	}
	ps.Lock()
	defer ps.Unlock()
	procNumExcess := insThdProcNumPS*float64(ps.totalInsThdNum-ps.inUseInsThdNum)*ps.scaleDownWindow.Seconds() -
		insThdReqNumPS*ps.scaleDownWindow.Seconds() + math.Min(float64(ps.pendingInsThdNum), 0)
	scaleInsThdNum := math.Ceil(procNumExcess / insThdProcNumPS / ps.scaleDownWindow.Seconds())
	scaleInsNum := int(math.Floor(scaleInsThdNum / float64(ps.concurrentNum)))
	ps.logger.Infof("calculated scale down instance number is %d", scaleInsNum)
	if scaleInsNum > 0 {
		return scaleInsNum
	}
	return scaleInsNum
}

func (ps *PredictScaler) handlePendingInsNumIncrease(insDiff int) {
	ps.Lock()
	ps.pendingInsThdNum += insDiff * ps.concurrentNum
	ps.predictDownChanFlag = false
	ps.Unlock()
}

func (ps *PredictScaler) handlePendingInsNumDecrease(insDiff int) {
	ps.Lock()
	ps.pendingInsThdNum -= insDiff * ps.concurrentNum
	ps.predictUpChanFlag = false
	ps.Unlock()
}

func (ps *PredictScaler) scaleUp() {
	scaleUpChan := make(chan struct{}, 1)
	tickerChan := make(<-chan time.Time, 1)
	var ticker *time.Ticker
	for {
		select {
		case _, ok := <-ps.scaleUpChan:
			if !ok {
				ps.logger.Warnf("trigger channel is closed")
				return
			}
			ps.logger.Infof("ps.predictUpDiff %d ", ps.predictUpDiff)
			ps.Lock()
			ps.pendingInsThdNum += ps.predictUpDiff * ps.concurrentNum
			ps.Unlock()
			ps.scaleUpHandler(ps.predictUpDiff, ps.handlePendingInsNumDecrease)
		case _, ok := <-scaleUpChan:
			if !ok {
				ps.logger.Warnf("scale up channel is closed")
				return
			}
			insThdReq := ps.checkReqNumFunc()
			if insThdReq == 0 {
				ps.stopScale(ticker)
				continue
			}
			ps.getScaleNum()
		case _, ok := <-ps.scaleUpTriggerCh:
			ps.logger.Info("receive scaleUpTriggerCh")
			if !ok {
				ps.logger.Warnf("trigger channel is closed")
				return
			}
			// Functions with long cold start times do not use passive scaleDown
			if ps.coldStartTime > longColdStartTime {
				continue
			}
			if ps.predictUpChanFlag || ps.predictDownChanFlag {
				ps.logger.Warnf("It is handling predictive scaling, ignoring current signals")
				continue
			}
			// let requests come in for certain time to calculate a more reasonable scale up number
			time.Sleep(ps.scaleUpWindow)
			scaleUpChan <- struct{}{}
			ticker = time.NewTicker(ps.scaleUpWindow)
			tickerChan = ticker.C
			ps.logger.Infof("scale up loop is running")
		case <-tickerChan:
			scaleUpChan <- struct{}{}
		case <-ps.stopCh:
			ps.logger.Warnf("stop scale up loop now")
			return
		}
	}
}

func (ps *PredictScaler) getScaleNum() {
	scaleInsNum := ps.getScaleUpInstancesNum()
	ps.logger.Infof("scaleUpTriggerCh scaleInsNum : %d", scaleInsNum)
	if scaleInsNum > 0 {
		ps.Lock()
		ps.pendingInsThdNum += scaleInsNum * ps.concurrentNum
		ps.Unlock()
		ps.scaleUpHandler(scaleInsNum, ps.handlePendingInsNumDecrease)
	}
}

func (ps *PredictScaler) stopScale(ticker *time.Ticker) {
	if ticker != nil {
		ticker.Stop()
	}
	ps.Lock()
	ps.predictScaleUpFlag = false
	ps.Unlock()
	ps.logger.Infof("scale up loop is paused")
}

func (ps *PredictScaler) scaleDown() {
	scaleDownChan := make(<-chan time.Time, 1)
	var timer *time.Timer
	for {
		select {
		case _, ok := <-ps.scaleDownChan:
			if !ok {
				ps.logger.Warnf("trigger channel is closed")
				return
			}
			ps.logger.Infof("ps.predictDownDiff %d ", ps.predictDownDiff)

			ps.Lock()
			ps.pendingInsThdNum -= ps.predictDownDiff * ps.concurrentNum
			ps.Unlock()
			ps.scaleDownHandler(ps.predictDownDiff, ps.handlePendingInsNumIncrease)
		case <-scaleDownChan:
			scaleInsNum := ps.getScaleDownInstancesNum()
			ps.logger.Infof("scaleDownTriggerCh scaleInsNum: %d", scaleInsNum)
			if scaleInsNum > 0 {
				ps.Lock()
				// Scale down based on the predicted number of instances as the bottom line
				currentAvailableInsNum := ps.totalInsThdNum/ps.concurrentNum + ps.minRsvInsNum
				scaleInsNum = int(math.Min(float64(scaleInsNum),
					math.Max(float64(currentAvailableInsNum-ps.predictInsNum), 0)))
				ps.logger.Infof("scaleDownTriggerCh new scaleInsNum: %d", scaleInsNum)
				ps.pendingInsThdNum -= scaleInsNum * ps.concurrentNum
				ps.Unlock()
				ps.scaleDownHandler(scaleInsNum, ps.handlePendingInsNumIncrease)
			}
			if timer != nil {
				timer.Stop()
			}
			ps.Lock()
			ps.predictScaleDownFlag = false
			ps.Unlock()
			ps.logger.Infof("scale down loop is paused")
		case _, ok := <-ps.scaleDownTriggerCh:
			if !ok {
				ps.logger.Warnf("trigger channel is closed")
				return
			}
			// Functions with long cold start times do not use passive scaleUp
			if ps.coldStartTime > longColdStartTime {
				continue
			}
			if ps.predictUpChanFlag || ps.predictDownChanFlag {
				ps.logger.Warnf("handling predictive scaling, ignoring current signals")
				continue
			}
			ps.Lock()
			ps.predictScaleDownFlag = true
			ps.Unlock()
			if timer == nil {
				timer = time.NewTimer(ps.scaleDownWindow)
			} else {
				select {
				case <-timer.C:
				default:
				}
				timer.Reset(ps.scaleDownWindow)
			}
			scaleDownChan = timer.C
			ps.logger.Infof("scale down loop is running", ps.funcKeyWithRes)
		case <-ps.stopCh:
			ps.logger.Warnf("stop scale down loop now", ps.funcKeyWithRes)
			return
		}
	}
}
