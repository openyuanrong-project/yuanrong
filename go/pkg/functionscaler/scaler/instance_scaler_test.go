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
	"errors"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/instanceconfig"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	"yuanrong.org/kernel/pkg/common/faas_common/statuscode"
	commontypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/metrics"
	"yuanrong.org/kernel/pkg/functionscaler/registry"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

func TestMiscellaneous(t *testing.T) {
	config.GlobalConfig = types.Configuration{}
	scaler := NewAutoScaler("test", &metrics.BucketCollector{}, func() int { return 1 },
		func(i int, cb ScaleUpCallback) {}, func(i int, cb ScaleDownCallback) {})
	as := scaler.(*AutoScaler)
	as.HandleFuncSpecUpdate(&types.FunctionSpecification{InstanceMetaData: commontypes.InstanceMetaData{ConcurrentNum: 2}})
	as.pendingInsThdNum = 6
	convey.Convey("CheckScaling", t, func() {
		res := as.CheckScaling()
		convey.So(res, convey.ShouldBeFalse)
	})
	convey.Convey("HandleCreateError", t, func() {
		as.HandleCreateError(nil)
		convey.So(as.pendingInsThdNum, convey.ShouldEqual, 6)
	})
}

func TestReplicaScaler(t *testing.T) {
	scaler := NewReplicaScaler("test", &metrics.BucketCollector{}, func(i int, cb ScaleUpCallback) {},
		func(i int, cb ScaleDownCallback) {})
	rs := scaler.(*ReplicaScaler)
	rs.HandleFuncSpecUpdate(&types.FunctionSpecification{InstanceMetaData: commontypes.InstanceMetaData{ConcurrentNum: 1}})
	convey.Convey("CheckScaling", t, func() {
		res := rs.CheckScaling()
		convey.So(res, convey.ShouldBeFalse)
	})
}

func TestScaleUpInstances(t *testing.T) {
	me := metrics.NewBucketMetricsCollector("funcKey123", "resource300")
	as := &AutoScaler{
		metricsCollector: me,
		autoScaleUpFlag:  true,
		concurrentNum:    2,
		pendingInsThdNum: 6,
		scaleUpWindow:    1 * time.Second,
		coldStartTime:    1 * time.Second,
		scaleUpHandler:   func(_ int, _ ScaleUpCallback) {},
	}
	checkNum := func() int {
		return as.pendingInsThdNum + 2
	}
	as.checkReqNumFunc = checkNum
	defer gomonkey.ApplyMethod(reflect.TypeOf(&metrics.BucketCollector{}), "InvokeMetricsCollected",
		func(_ *metrics.BucketCollector) bool {
			return true
		}).Reset()

	convey.Convey("insThdProcNumPS is 0", t, func() {
		defer gomonkey.ApplyMethod(reflect.TypeOf(&metrics.BucketCollector{}), "GetCalculatedInvokeMetrics",
			func(_ *metrics.BucketCollector) (float64, float64, float64) {
				return 0, 0, 0
			}).Reset()
		as.scaleUpInstances()
		convey.So(as.pendingInsThdNum, convey.ShouldEqual, 6)
		convey.So(as.remainedInsThdReqNum, convey.ShouldEqual, 0)
	})
	convey.Convey("procWindow = 0", t, func() {
		defer gomonkey.ApplyMethod(reflect.TypeOf(&metrics.BucketCollector{}), "GetCalculatedInvokeMetrics",
			func(_ *metrics.BucketCollector) (float64, float64, float64) {
				return 0, 1, 0
			}).Reset()
		as.scaleUpInstances()
		convey.So(as.pendingInsThdNum, convey.ShouldEqual, 8)
		convey.So(as.remainedInsThdReqNum, convey.ShouldEqual, 6)
	})
	convey.Convey("procWindow > 0", t, func() {
		defer gomonkey.ApplyMethod(reflect.TypeOf(&metrics.BucketCollector{}), "GetCalculatedInvokeMetrics",
			func(_ *metrics.BucketCollector) (float64, float64, float64) {
				return 0, 1, 0
			}).Reset()
		as.scaleUpWindow = 2 * time.Second
		as.remainedInsThdReqNum = 1
		as.scaleUpInstances()
		convey.So(as.pendingInsThdNum, convey.ShouldEqual, 10)
		convey.So(as.remainedInsThdReqNum, convey.ShouldEqual, 8)
	})
	convey.Convey("InvokeMetricsCollected is false", t, func() {
		defer gomonkey.ApplyMethod(reflect.TypeOf(&metrics.BucketCollector{}), "InvokeMetricsCollected",
			func(_ *metrics.BucketCollector) bool {
				return false
			}).Reset()
		as.scaleUpInstances()
		convey.So(as.pendingInsThdNum, convey.ShouldEqual, 12)
	})
}

func TestScaleDownInstances(t *testing.T) {
	me := metrics.NewBucketMetricsCollector("funcKey123", "resource300")
	res := 0
	as := &AutoScaler{
		metricsCollector: me,
		autoScaleUpFlag:  true,
		inUseInsThdNum:   2,
		totalInsThdNum:   6,
		concurrentNum:    2,
		pendingInsThdNum: 0,
		scaleDownWindow:  1 * time.Second,
		coldStartTime:    1 * time.Second,
		scaleUpHandler:   func(_ int, cb ScaleUpCallback) {},
		scaleDownHandler: func(input int, cb ScaleDownCallback) { res = input },
	}
	checkNum := func() int {
		return as.pendingInsThdNum + 2
	}
	as.checkReqNumFunc = checkNum

	defer gomonkey.ApplyMethod(reflect.TypeOf(&metrics.BucketCollector{}), "InvokeMetricsCollected",
		func(_ *metrics.BucketCollector) bool {
			return false
		}).Reset()
	convey.Convey("pendingThsThdNum negative and little", t, func() {
		as.pendingInsThdNum = -2
		as.scaleDownInstances()
		convey.So(res, convey.ShouldEqual, 1)
	})
	convey.Convey("pendingThsThdNum negative and big", t, func() {
		as.pendingInsThdNum = -6
		as.scaleDownInstances()
		convey.So(res, convey.ShouldEqual, 0)
	})
	convey.Convey("pendingThsThdNum positive", t, func() {
		as.pendingInsThdNum = 2
		as.scaleDownInstances()
		convey.So(res, convey.ShouldEqual, 3)
	})
	defer gomonkey.ApplyMethod(reflect.TypeOf(&metrics.BucketCollector{}), "InvokeMetricsCollected",
		func(_ *metrics.BucketCollector) bool {
			return true
		}).Reset()
	convey.Convey("insThdProcNumPS is 0", t, func() {
		res = 0
		defer gomonkey.ApplyMethod(reflect.TypeOf(&metrics.BucketCollector{}), "GetCalculatedInvokeMetrics",
			func(_ *metrics.BucketCollector) (float64, float64, float64) {
				return 0, 0, 0
			}).Reset()
		as.scaleDownInstances()
		convey.So(res, convey.ShouldEqual, 0)
	})
	convey.Convey("insThdProcNumPS is not 0", t, func() {
		res = 0
		defer gomonkey.ApplyMethod(reflect.TypeOf(&metrics.BucketCollector{}), "GetCalculatedInvokeMetrics",
			func(_ *metrics.BucketCollector) (float64, float64, float64) {
				return 0, 10, 10
			}).Reset()
		as.scaleDownInstances()
		convey.So(res, convey.ShouldEqual, 0)
	})
}

func TestAutoScalerProcess(t *testing.T) {
	me := metrics.NewBucketMetricsCollector("funcKey123", "resource300")
	res := 0
	as := &AutoScaler{
		metricsCollector: me,
		autoScaleUpFlag:  true,
		concurrentNum:    2,
		pendingInsThdNum: 6,
		scaleUpWindow:    1 * time.Second,
		coldStartTime:    1 * time.Second,
		scaleUpHandler:   func(_ int, cb ScaleUpCallback) {},
		scaleDownHandler: func(input int, cb ScaleDownCallback) { res = input },
	}

	convey.Convey("HandleInsThdUpdate", t, func() {
		as.inUseInsThdNum = 4
		as.totalInsThdNum = 6
		as.pendingInsThdNum = 0
		as.HandleInsThdUpdate(0, 1)
		expectNum := as.GetExpectInstanceNumber()
		convey.So(expectNum, convey.ShouldEqual, 3)
		convey.So(res, convey.ShouldEqual, 0)
	})
}

func TestHandleCreateError(t *testing.T) {
	res := 0
	rs := &ReplicaScaler{
		metricsCollector: metrics.NewBucketMetricsCollector("", ""),
		scaleUpHandler:   func(in int, cb ScaleUpCallback) { res = in },
		enable:           true,
		pendingRsvInsNum: 2,
		concurrentNum:    1,
		currentRsvInsNum: 1,
		targetRsvInsNum:  4,
		logger:           log.GetLogger(),
	}

	convey.Convey("HandleCreateError", t, func() {
		rs.HandleCreateError(snerror.New(statuscode.StsConfigErrCode, "config sts error"))
		convey.So(res, convey.ShouldEqual, 0)
		rs.enable = true
		rs.HandleCreateError(snerror.New(statuscode.UserFuncEntryNotFoundErrCode, "entry not found"))
		convey.So(res, convey.ShouldEqual, 1)
		rs.HandleCreateError(errors.New("recoverable error"))
		convey.So(res, convey.ShouldEqual, 1)
	})
}

func TestReplicaScalerProcess(t *testing.T) {
	rs := &ReplicaScaler{
		metricsCollector: metrics.NewBucketMetricsCollector("", ""),
		targetRsvInsNum:  1,
		concurrentNum:    1,
		scaleUpHandler: func(i int, cb ScaleUpCallback) {
		},
		scaleDownHandler: func(i int, cb ScaleDownCallback) {
		},
		logger: log.GetLogger(),
	}
	convey.Convey("HandleInsThdUpdate", t, func() {
		rs.HandleInsThdUpdate(0, 1)
		convey.So(rs.currentRsvInsNum, convey.ShouldEqual, 1)
	})
	convey.Convey("HandleInsConfigUpdate", t, func() {
		rs.HandleInsConfigUpdate(&instanceconfig.Configuration{InstanceMetaData: commontypes.InstanceMetaData{MinInstance: 2}})
		convey.So(rs.targetRsvInsNum, convey.ShouldEqual, 2)
	})
	convey.Convey("HandleFuncSpecUpdate", t, func() {
		rs.HandleFuncSpecUpdate(&types.FunctionSpecification{InstanceMetaData: commontypes.InstanceMetaData{ConcurrentNum: 2}})
		convey.So(rs.concurrentNum, convey.ShouldEqual, 2)
	})
}

func TestTriggerScale(t *testing.T) {
	convey.Convey("TriggerScale", t, func() {
		var scaleDownNum int
		replicaScaler := NewReplicaScaler("testfunc-cpu-500-mem-500", metrics.NewBucketMetricsCollector("", ""),
			func(i int, callback ScaleUpCallback) {}, func(i int, callback ScaleDownCallback) {
				scaleDownNum = i
				callback(1)
			})
		replicaScaler.SetEnable(false)
		replicaScaler.HandleFuncSpecUpdate(&types.FunctionSpecification{InstanceMetaData: commontypes.InstanceMetaData{ConcurrentNum: 1}})
		replicaScaler.HandleInsThdUpdate(0, 2)
		replicaScaler.HandleInsConfigUpdate(&instanceconfig.Configuration{InstanceMetaData: commontypes.InstanceMetaData{
			MaxInstance:   1,
			MinInstance:   1,
			ConcurrentNum: 1,
		}})
		replicaScaler.SetEnable(true)
		convey.So(scaleDownNum, convey.ShouldEqual, 1)
	})
}

func TestCalSleepTime(t *testing.T) {
	config.GlobalConfig.PredictGroupWindow = 15 * 60 * 1000
	convey.Convey("CalSleepTime", t, func() {
		// 2024-05-08 21:33:00 --> 1715175180000
		t1 := 1715175180000 % config.GlobalConfig.PredictGroupWindow
		convey.So(t1, convey.ShouldEqual, 3000*60)

		time := CalSleepTime(1715175180000, time.Duration(10)*time.Minute)
		convey.So(time.Milliseconds(), convey.ShouldEqual, 2000*60)
	})
}

func TestHandlePredictUpdate(t *testing.T) {
	config.GlobalConfig = types.Configuration{PredictGroupWindow: 15 * 60 * 1000}
	pScaler := NewPredictScaler("test", &metrics.BucketCollector{}, func() int { return 1 },
		func(i int, cb ScaleUpCallback) {}, func(i int, cb ScaleDownCallback) {})
	ps := pScaler.(*PredictScaler)
	ps.coldStartTime = 15 * 60 * 1000
	result := []float64{30.0, 60.0} // [2,3]
	gomonkey.ApplyFunc(CalSleepTime,
		func(currentTimeStamp int64, coldStartTime time.Duration) time.Duration {
			return time.Duration(1)
		})
	// currentInsNum is 6
	ps.totalInsThdNum = 4
	ps.pendingInsThdNum = 1
	ps.concurrentNum = 1
	ps.minRsvInsNum = 1

	convey.Convey("HandlePredictUpdate", t, func() {
		ps.HandlePredictUpdate(&types.PredictQPSGroups{
			QPSGroups: result,
		})
		ps.scaleDownChan <- struct{}{}
		convey.So(ps.predictDownDiff, convey.ShouldEqual, 4)
		// convey.So(ps.predictUpDiff, convey.ShouldEqual, 1)
	})

	convey.Convey("CheckScaling", t, func() {
		res := ps.CheckScaling()
		convey.So(res, convey.ShouldBeFalse)
	})

	// currentInsNum is 1
	ps.totalInsThdNum = 0
	ps.pendingInsThdNum = 0
	ps.concurrentNum = 1
	ps.minRsvInsNum = 1
	convey.Convey("HandlePredictUpdate", t, func() {
		ps.HandlePredictUpdate(&types.PredictQPSGroups{
			QPSGroups: result,
		})
		ps.scaleUpChan <- struct{}{}
		convey.So(ps.predictDownDiff, convey.ShouldEqual, 0)
		// convey.So(ps.predictUpDiff, convey.ShouldEqual, 2)
	})

	convey.Convey("HandleInsConfigUpdate", t, func() {
		ps.HandleInsConfigUpdate(&instanceconfig.Configuration{InstanceMetaData: commontypes.InstanceMetaData{MinInstance: 2}})
		convey.So(ps.minRsvInsNum, convey.ShouldEqual, 2)
	})

	convey.Convey("HandleFuncSpecUpdate", t, func() {
		ps.HandleFuncSpecUpdate(&types.FunctionSpecification{InstanceMetaData: commontypes.InstanceMetaData{ConcurrentNum: 2}})
		convey.So(ps.concurrentNum, convey.ShouldEqual, 2)
	})
}

func TestPredictScalerProcess(t *testing.T) {
	me := metrics.NewBucketMetricsCollector("funcKey123", "resource300")
	scaleInsNum := 0
	res := 0
	ps := &PredictScaler{
		logger:             log.GetLogger().With(zap.Any("funcKeyWithRes", "funcKey123")),
		metricsCollector:   me,
		predictScaleUpFlag: true,
		concurrentNum:      2,
		pendingInsThdNum:   6,
		scaleUpWindow:      1 * time.Second,
		scaleDownWindow:    1 * time.Second,
		coldStartTime:      1 * time.Second,
		scaleUpHandler:     func(_ int, cb ScaleUpCallback) {},
		scaleDownHandler:   func(input int, cb ScaleDownCallback) { res = input },
		scaleUpChan:        make(chan struct{}, 1),
		scaleDownChan:      make(chan struct{}, 1),
		scaleUpTriggerCh:   make(chan struct{}, 1),
		scaleDownTriggerCh: make(chan struct{}, 1),
		stopCh:             make(chan struct{}, 1),
	}
	checkNum := func() int {
		return ps.pendingInsThdNum + 2
	}
	ps.checkReqNumFunc = checkNum

	convey.Convey("InvokeMetricsCollected is true", t, func() {
		defer gomonkey.ApplyMethod(reflect.TypeOf(&metrics.BucketCollector{}), "InvokeMetricsCollected",
			func(_ *metrics.BucketCollector) bool {
				return true
			}).Reset()
		defer gomonkey.ApplyMethod(reflect.TypeOf(&metrics.BucketCollector{}), "GetCalculatedInvokeMetrics",
			func(_ *metrics.BucketCollector) (float64, float64, float64) {
				return 0, 0, 0
			}).Reset()
		scaleInsNum = ps.getScaleDownInstancesNum()
		convey.So(scaleInsNum, convey.ShouldEqual, -1)
		convey.So(res, convey.ShouldEqual, 0)

		defer gomonkey.ApplyMethod(reflect.TypeOf(&metrics.BucketCollector{}), "GetCalculatedInvokeMetrics",
			func(_ *metrics.BucketCollector) (float64, float64, float64) {
				return 3, 1, 0
			}).Reset()
		ps.totalInsThdNum = 4
		ps.inUseInsThdNum = 1
		ps.scaleDownWindow = 60000 * time.Millisecond
		ps.pendingInsThdNum = -1
		scaleInsNum = ps.getScaleDownInstancesNum()
		convey.So(scaleInsNum, convey.ShouldEqual, 1)
	})

	convey.Convey("getScaleUpInstancesNum", t, func() {
		scaleInsNum = ps.getScaleUpInstancesNum()
		convey.So(scaleInsNum, convey.ShouldEqual, 1)
		convey.So(res, convey.ShouldEqual, 0)
	})

	convey.Convey("calculate insThdNum", t, func() {
		ps.inUseInsThdNum = 1
		ps.totalInsThdNum = 4
		ps.concurrentNum = 1
		ps.HandleInsThdUpdate(1, 0)
		ps.TriggerScale()

		ps.pendingInsThdNum = 2
		ps.predictDownDiff = 1
		expectNum := ps.GetExpectInstanceNumber()
		convey.So(expectNum, convey.ShouldEqual, 6)
		ps.handlePendingInsNumIncrease(1)
		convey.So(ps.pendingInsThdNum, convey.ShouldEqual, 3)
		ps.handlePendingInsNumDecrease(1)
		convey.So(ps.pendingInsThdNum, convey.ShouldEqual, 2)
	})

	convey.Convey("scaleUp", t, func() {
		go ps.scaleUp()
		ps.scaleUpTriggerCh <- struct{}{}
		convey.So(res, convey.ShouldEqual, 0)
		time.Sleep(2 * time.Second)
		close(ps.stopCh)
	})

	convey.Convey("scaleDown", t, func() {
		ps.stopCh = make(chan struct{}, 1)
		go ps.scaleDown()
		ps.scaleDownTriggerCh <- struct{}{}
		convey.So(res, convey.ShouldEqual, 0)
		time.Sleep(2 * time.Second)
		close(ps.stopCh)
	})
}

func TestStartPredictRegistry(t *testing.T) {
	convey.Convey("StartPredictRegistry", t, func() {

		registry.GlobalRegistry = &registry.Registry{FaaSSchedulerRegistry: registry.NewFaasSchedulerRegistry(make(chan struct{}))}
		selfregister.GlobalSchedulerProxy.Add(&commontypes.InstanceInfo{
			FunctionName: "faasscheduler",
			InstanceName: "abcdefg",
		}, "")
		selfregister.SelfInstanceID = "abcdefg"
		os.Setenv(constant.ClusterNameEnvKey, "localAZ")
		os.Setenv("INSTANCE_ID", "abcdefg")
		defer func() {
			os.Unsetenv(constant.ClusterNameEnvKey)
			os.Unsetenv("INSTANCE_ID")
		}()
		config.GlobalConfig.PredictGroupWindow = 15 * 60 * 1000
		dsw0 := time.Now().Add(-1 * time.Minute)
		dsw1 := time.Now().Add(30 * time.Minute)
		dataSetTimeWindow := []int64{dsw0.UnixMilli(), dsw1.UnixMilli()}
		qpsResult := map[string][]float64{
			"244177614494574475/0@default@primary_secondary/latest": {0.0, 0.5, 1.0, 1.5, 2.0, 2.5, 3.0, 3.5, 4.0,
				4.5, 5.5, 6.0, 6.5, 7.0, 7.5, 8.0, 8.5,
				9.0, 9.5, 10.0, 10.5, 11.0, 11.5, 12.0,
				12.5, 13.0, 13.5, 14.0, 14.5, 15.0},
			"244177614494574475/0@default@primary_secondary-slv0/latest": {0.0, 0.5, 1.0, 1.5, 2.0, 2.5, 3.0, 3.5, 4.0,
				4.5, 5.5, 6.0, 6.5, 7.0, 7.5, 8.0, 8.5,
				9.0, 9.5, 10.0, 10.5, 11.0, 11.5, 12.0,
				12.5, 13.0, 13.5, 14.0},
		}
		p := &types.PredictResult{
			DataSetTimeWindow: dataSetTimeWindow,
			QPSResult:         qpsResult,
			IsValid:           false,
		}
		etcdClient := &etcd3.EtcdClient{}
		var (
			patches                 []*gomonkey.Patches
			mockHandlePredictUpdate func(prg *types.PredictQPSGroups)
		)
		patches = append(patches,
			gomonkey.ApplyFunc(etcd3.GetMetaEtcdClient, func() *etcd3.EtcdClient {
				return etcdClient
			}),
			gomonkey.ApplyMethod(reflect.TypeOf(&etcd3.EtcdWatcher{}), "StartWatch", func(ew *etcd3.EtcdWatcher) {
				bytes, _ := json.Marshal(p)
				ew.ResultChan <- &etcd3.Event{
					Type:      0,
					Key:       "/arrivalPrediction/localAZ",
					Value:     bytes,
					PrevValue: nil,
					Rev:       0,
				}
				p.DataSetTimeWindow = []int64{time.Now().Add(-40 * time.Minute).UnixMilli(),
					time.Now().Add(-10 * time.Minute).UnixMilli()}
				bytes, _ = json.Marshal(p)
				ew.ResultChan <- &etcd3.Event{
					Type:      0,
					Key:       "/arrivalPrediction/localAZ",
					Value:     bytes,
					PrevValue: nil,
					Rev:       0,
				}
			}),
			gomonkey.ApplyMethod(reflect.TypeOf(&PredictScaler{}), "HandlePredictUpdate",
				func(ps *PredictScaler, prg *types.PredictQPSGroups) {
					mockHandlePredictUpdate(prg)
				}),
		)
		defer func() {
			for _, patch := range patches {
				patch.Reset()
				time.Sleep(10 * time.Millisecond)
			}
		}()

		predictScaler := NewPredictScaler("test-function-res", nil, func() int {
			return 0
		}, func(i int, callback ScaleUpCallback) {

		}, func(i int, callback ScaleDownCallback) {

		}).(*PredictScaler)
		predictScaler.HandleInsConfigUpdate(&instanceconfig.Configuration{InstanceMetaData: commontypes.InstanceMetaData{MinInstance: 0}})
		var predictGroupNum int
		var wg sync.WaitGroup
		var (
			result_primary_secondary      []float64
			result_primary_secondary_slv0 []float64
		)
		wg.Add(2)
		mockHandlePredictUpdate = func(prg *types.PredictQPSGroups) {
			predictGroupNum++
			if prg.FuncKey == "244177614494574475/0@default@primary_secondary/latest" {
				result_primary_secondary = prg.QPSGroups

			}
			if prg.FuncKey == "244177614494574475/0@default@primary_secondary-slv0/latest" {
				result_primary_secondary_slv0 = prg.QPSGroups
			}
			wg.Done()
		}

		wg.Wait()
		convey.So(predictGroupNum, convey.ShouldEqual, 2)
		convey.So(len(result_primary_secondary), convey.ShouldEqual, 2)
		convey.So(result_primary_secondary[0], convey.ShouldEqual, 3.6666666666666665)
		convey.So(result_primary_secondary[1], convey.ShouldEqual, 10.966666666666667)

		convey.So(len(result_primary_secondary_slv0), convey.ShouldEqual, 2)
		convey.So(result_primary_secondary_slv0[0], convey.ShouldEqual, 3.6666666666666665)
		convey.So(result_primary_secondary_slv0[1], convey.ShouldEqual, 9)
		predictScaler.Destroy()
	})
}
