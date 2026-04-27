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

// Package signalmanager -
package signalmanager

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/aliasroute"
	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/uuid"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

// 测试SignalManager初始化
func TestGetSignalManager(t *testing.T) {
	convey.Convey("TestGetSignalManager", t, func() {
		// 测试SignalManager单例
		sm1 := GetSignalManager()
		sm2 := GetSignalManager()
		convey.So(sm1 == sm2, convey.ShouldBeTrue)
		convey.So(sm1.Logger, convey.ShouldNotBeNil)
		convey.So(sm2.instances, convey.ShouldNotBeNil)
	})
}

func TestSignalManager_SetKillFunc(t *testing.T) {
	convey.Convey("TestGetSignalManager", t, func() {
		// 测试SignalManager单例
		GetSignalManager().SetKillFunc(func(_ string, value int, _ []byte) error {
			if value == 0 {
				return nil
			}
			return fmt.Errorf("%d", value)
		})
		sm := GetSignalManager()

		convey.So(sm == nil, convey.ShouldBeFalse)
		convey.So(sm.killFunc == nil, convey.ShouldBeFalse)
		convey.So(sm.killFunc("", 0, nil), convey.ShouldBeNil)
		convey.So(sm.killFunc("", 199, nil), convey.ShouldResemble, fmt.Errorf("199"))
	})
}

func TestSignalManager_SignalInstance(t *testing.T) {
	convey.Convey("SignalInstance", t, func() {
		GetSignalManager().instances = make(map[string]*signalInstance)
		result := make(map[string]struct{})
		var lock sync.RWMutex
		defer gomonkey.ApplyPrivateMethod(reflect.TypeOf(&signalProcessor{}), "signalInstance", func(sp *signalProcessor, _ string) {
			lock.Lock()
			result[fmt.Sprintf("%s_%d", sp.InstanceId, sp.SignalNo)] = struct{}{}
			lock.Unlock()
		}).Reset()
		convey.Convey("SignalInstance simple", func() {
			GetSignalManager().SignalInstance(&types.Instance{InstanceID: "signal0", FuncKey: "default/hello/$latest"}, 0)
			GetSignalManager().SignalInstance(&types.Instance{InstanceID: "signal1", FuncKey: "default/hello/$latest"}, 1)
			convey.So(len(GetSignalManager().instances) == 0, convey.ShouldBeTrue)
		})
		convey.Convey("SignalInstance complex", func() {
			GetSignalManager().SignalInstance(&types.Instance{InstanceID: "signal2"}, constant.KillSignalAliasUpdate)
			convey.So(len(GetSignalManager().instances), convey.ShouldEqual, 0)

			GetSignalManager().SignalInstance(&types.Instance{InstanceID: "signal2", FuncKey: "default/hello/$latest"}, constant.KillSignalAliasUpdate)
			GetSignalManager().SignalInstance(&types.Instance{InstanceID: "signal2", FuncKey: "default/hello/$latest"}, constant.KillSignalFaaSSchedulerUpdate)
			GetSignalManager().SignalInstance(&types.Instance{InstanceID: "signal3", FuncKey: "default/hello/$latest"}, constant.KillSignalAliasUpdate)
			GetSignalManager().SignalInstance(&types.Instance{InstanceID: "signal3", FuncKey: "default/hello/$latest"}, constant.KillSignalFaaSSchedulerUpdate)
			GetSignalManager().SignalInstance(&types.Instance{InstanceID: "signal4", FuncKey: "default/hello/$latest"}, constant.KillSignalAliasUpdate)
			GetSignalManager().SignalInstance(&types.Instance{InstanceID: "signal4", FuncKey: "default/hello/$latest"}, constant.KillSignalAliasUpdate)
			time.Sleep(500 * time.Millisecond)
			convey.So(len(GetSignalManager().instances), convey.ShouldEqual, 3)
			processorStrArr := []string{"signal2_64", "signal2_72", "signal3_64", "signal3_72", "signal4_64"}
			fmt.Printf("result is %v\n", result)
			for _, s := range processorStrArr {
				_, ok := result[s]
				convey.So(ok, convey.ShouldBeTrue)
			}
			convey.So(len(result), convey.ShouldEqual, 5)
		})
	})
}

func TestSignalManager_RemoveInstance(t *testing.T) {
	convey.Convey("RemoveInstance", t, func() {
		GetSignalManager().instances = make(map[string]*signalInstance)
		result := make(map[string]struct{})
		wg := sync.WaitGroup{}
		lock := sync.Mutex{}
		defer gomonkey.ApplyPrivateMethod(reflect.TypeOf(&signalProcessor{}), "signalInstance", func(sp *signalProcessor, _ string) {
			wg.Add(1)
			for {
				select {
				case <-sp.StopChan:
					lock.Lock()
					result[fmt.Sprintf("%s_%d", sp.InstanceId, sp.SignalNo)] = struct{}{}
					lock.Unlock()
					wg.Done()
					return
				}
			}
		}).Reset()
		convey.Convey("RemoveInstance complex", func() {
			GetSignalManager().SignalInstance(&types.Instance{InstanceID: "signal2", FuncKey: "default/hello/$latest"}, constant.KillSignalAliasUpdate)
			GetSignalManager().SignalInstance(&types.Instance{InstanceID: "signal2", FuncKey: "default/hello/$latest"}, constant.KillSignalFaaSSchedulerUpdate)
			GetSignalManager().SignalInstance(&types.Instance{InstanceID: "signal3", FuncKey: "default/hello/$latest"}, constant.KillSignalAliasUpdate)
			GetSignalManager().SignalInstance(&types.Instance{InstanceID: "signal3", FuncKey: "default/hello/$latest"}, constant.KillSignalFaaSSchedulerUpdate)
			GetSignalManager().SignalInstance(&types.Instance{InstanceID: "signal4", FuncKey: "default/hello/$latest"}, constant.KillSignalAliasUpdate)
			GetSignalManager().SignalInstance(&types.Instance{InstanceID: "signal4", FuncKey: "default/hello/$latest"}, constant.KillSignalAliasUpdate)

			GetSignalManager().RemoveInstance("signal2")
			convey.So(len(GetSignalManager().instances), convey.ShouldEqual, 2)
			time.Sleep(500 * time.Millisecond)
			var ok bool
			_, ok = result["signal2_64"]
			convey.So(ok, convey.ShouldBeTrue)
			_, ok = result["signal2_72"]
			convey.So(ok, convey.ShouldBeTrue)
			GetSignalManager().RemoveInstance("signal3")
			GetSignalManager().RemoveInstance("signal4")
			wg.Wait()
			convey.So(len(GetSignalManager().instances), convey.ShouldEqual, 0)
			convey.So(len(result), convey.ShouldEqual, 5)
		})
	})
}

func TestSignalProcessor_signalInstance(t *testing.T) {
	convey.Convey("signalInstance", t, func() {
		sp := &signalProcessor{
			InstanceId: "mock",
			HasSignal:  false,
			IsRunning:  false,
			StopChan:   make(chan struct{}),
			SignalNo:   64,
			getDataFunc: func() ([]byte, error) {
				return []byte("aaa"), nil
			},
			RWMutex: sync.RWMutex{},
			Logger:  log.GetLogger(),
		}

		result := []time.Duration{}
		defer gomonkey.ApplyFunc(time.Sleep, func(duration time.Duration) {
			result = append(result, duration)
		}).Reset()

		failTimes := 0
		totalFailTimes := 100
		sp.killFunc = func(string, []byte) error {
			if failTimes < totalFailTimes {
				failTimes++
				return fmt.Errorf("%d times failed", failTimes)
			}
			return nil
		}
		sp.HasSignal = true
		sp.signalInstance("111")
		convey.So(len(result), convey.ShouldEqual, 100)
		convey.So(sp.HasSignal, convey.ShouldEqual, false)
		convey.So(sp.IsRunning, convey.ShouldEqual, false)

		sleeptTime := 100 * time.Millisecond
		for i := 0; i < 99; i++ {
			convey.So(result[i], convey.ShouldEqual, sleeptTime)
			sleeptTime *= 2
			if sleeptTime > 5*time.Minute {
				sleeptTime = 5 * time.Minute
			}
		}
		failTimes = 0
		sp.getDataFunc = func() ([]byte, error) {
			return nil, fmt.Errorf("error")
		}

		sp.signalInstance("222")
		convey.So(failTimes, convey.ShouldEqual, 0)
	})
}

func TestSignalManager_killFunc_return_instanceNotExist(t *testing.T) {
	convey.Convey("killFunc_return_instanceNotExist", t, func() {
		flag := false
		GetSignalManager().SetKillFunc(func(string, int, []byte) error {
			flag = true
			return fmt.Errorf("instance not found, the instance may have been killed")
		})

		GetSignalManager().SignalInstance(&types.Instance{
			InstanceID:   "1111",
			InstanceName: "1111",
			FuncKey:      "default/hello/$latest",
		}, constant.KillSignalAliasUpdate)
		time.Sleep(100 * time.Millisecond)
		_, ok := GetSignalManager().instances["1111"]
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(flag, convey.ShouldBeTrue)
	})
}

func TestSignalManager_complex(t *testing.T) {
	GetSignalManager().instances = make(map[string]*signalInstance)
	defer gomonkey.ApplyFunc(PrepareSchedulerArg, func() ([]byte, error) {
		return nil, nil
	}).Reset()

	defer gomonkey.ApplyFunc(aliasroute.MarshalTenantAliasList, func(string) ([]byte, error) {
		return nil, nil
	}).Reset()

	var okFlag atomic.Bool

	var killFuncExecuted atomic.Bool
	var blockFlag atomic.Bool
	killFunc := func(string, int, []byte) error {
		killFuncExecuted.Store(true)
		if okFlag.Load() {
			return nil
		}
		for blockFlag.Load() {

		}
		return fmt.Errorf("")
	}

	uuidCount := 0
	defer gomonkey.ApplyMethod(reflect.TypeOf(uuid.New()), "String", func(_ uuid.RandomUUID) string {
		uuidCount++
		return ""
	}).Reset()
	defer gomonkey.ApplyFunc(time.Sleep, func(duration time.Duration) {

	}).Reset()

	convey.Convey("concurrency processor", t, func() {
		GetSignalManager().SetKillFunc(killFunc)
		killFuncExecuted.Store(false)

		okFlag.Store(false)
		GetSignalManager().SignalInstance(&types.Instance{InstanceID: "signal0", FuncKey: "default/hello/$latest"}, constant.KillSignalAliasUpdate)

		convey.So(GetSignalManager().instances["signal0"], convey.ShouldNotBeNil)
		sp := GetSignalManager().instances["signal0"].signalProcessors[constant.KillSignalAliasUpdate]
		convey.So(sp, convey.ShouldNotBeNil)

		for !killFuncExecuted.Load() {
		}
		blockFlag.Store(true)
		fmt.Printf("executed is %v\n", killFuncExecuted.Load())
		sp.Lock()
		convey.So(sp.IsRunning, convey.ShouldEqual, true)
		convey.So(sp.HasSignal, convey.ShouldEqual, false)
		convey.So(uuidCount, convey.ShouldEqual, 1)

		killFuncExecuted.Store(false)
		blockFlag.Store(false)
		sp.Unlock()

		GetSignalManager().SignalInstance(&types.Instance{InstanceID: "signal0", FuncKey: "default/hello/$latest"}, constant.KillSignalAliasUpdate)
		blockFlag.Store(true)
		<-time.After(200 * time.Millisecond)
		sp.Lock()
		convey.So(sp.IsRunning, convey.ShouldEqual, true)
		convey.So(sp.HasSignal, convey.ShouldEqual, false)
		convey.So(uuidCount, convey.ShouldEqual, 1)
		sp.Unlock()
		blockFlag.Store(false)
		okFlag.Store(true)
		<-time.After(200 * time.Millisecond)
		convey.So(sp.IsRunning, convey.ShouldEqual, false)
		convey.So(sp.HasSignal, convey.ShouldEqual, false)
	})

	convey.Convey("stop chan delete", t, func() {
		okFlag.Store(false)
		blockFlag.Store(false)
		GetSignalManager().SignalInstance(&types.Instance{InstanceID: "signal0", FuncKey: "default/hello/$latest"}, constant.KillSignalAliasUpdate)

		convey.So(GetSignalManager().instances["signal0"], convey.ShouldNotBeNil)
		sp := GetSignalManager().instances["signal0"].signalProcessors[constant.KillSignalAliasUpdate]
		convey.So(sp, convey.ShouldNotBeNil)
		<-time.After(100 * time.Millisecond)
		blockFlag.Store(true)
		sp.Lock()
		convey.So(sp.IsRunning, convey.ShouldEqual, true)
		convey.So(sp.HasSignal, convey.ShouldEqual, false)
		sp.Unlock()

		blockFlag.Store(false)
		GetSignalManager().RemoveInstance("signal0")
		<-time.After(100 * time.Millisecond)
		convey.So(sp.IsRunning, convey.ShouldEqual, false)
		convey.So(sp.HasSignal, convey.ShouldEqual, false)
	})
}
