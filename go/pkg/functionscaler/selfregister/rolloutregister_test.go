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

// Package selfregister contains service route logic
package selfregister

import (
	"errors"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/rollout"
)

func TestRegisterRolloutToEtcd(t *testing.T) {
	config.GlobalConfig.EnableRollout = true
	defer func() {
		config.GlobalConfig.EnableRollout = false
	}()
	stopCh := make(chan struct{})
	var regErr error
	defer gomonkey.ApplyFunc((*etcd3.EtcdRegister).Register, func() error {
		return regErr
	}).Reset()
	convey.Convey("Test RegisterRolloutToEtcd", t, func() {
		regErr = errors.New("some error")
		err := RegisterRolloutToEtcd(stopCh)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestContendRolloutInEtcd(t *testing.T) {
	maxContendTime = 1
	config.GlobalConfig.EnableRollout = true
	defer func() {
		config.GlobalConfig.EnableRollout = false
	}()
	var (
		getResponse *clientv3.GetResponse
		getError    error
		invokeRes   []byte
		invokeErr   error
		lockErr     error
	)
	patches := []*gomonkey.Patches{
		gomonkey.ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
			return &etcd3.EtcdClient{}
		}),
		gomonkey.ApplyFunc((*etcd3.EtcdClient).Get, func(_ *etcd3.EtcdClient, ctxInfo etcd3.EtcdCtxInfo,
			key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
			return getResponse, getError
		}),
		gomonkey.ApplyFunc((*etcd3.EtcdClient).Put, func(_ *etcd3.EtcdClient, ctxInfo etcd3.EtcdCtxInfo,
			etcdKey string, value string, opts ...clientv3.OpOption) error {
			return nil
		}),
		gomonkey.ApplyFunc((*etcd3.EtcdClient).Delete, func(_ *etcd3.EtcdClient, ctxInfo etcd3.EtcdCtxInfo,
			etcdKey string, opts ...clientv3.OpOption) error {
			return nil
		}),
		gomonkey.ApplyFunc((*etcd3.EtcdLocker).TryLock, func(_ *etcd3.EtcdLocker, key string) error {
			return lockErr
		}),
		gomonkey.ApplyFunc(rollout.InvokeByInstanceId, func(args []api.Arg, instanceID string, traceID string) ([]byte,
			error) {
			return invokeRes, invokeErr
		}),
	}
	defer func() {
		for _, p := range patches {
			p.Reset()
		}
	}()
	convey.Convey("Test RegisterRolloutToEtcd", t, func() {
		convey.Convey("register success", func() {
			SelfInstanceID = "instance1"
			selfInstanceSpec = &types.InstanceSpecification{}
			invokeRes = []byte(`{"registerKey": "/sn/faas-scheduler/instances/cluster001/node001/bj-pod1id"}`)
			invokeErr = nil
			getResponse = &clientv3.GetResponse{
				Kvs: []*mvccpb.KeyValue{
					{Key: []byte("/sn/faas-scheduler/rollout/cluster1/node1/aaa"), Value: []byte("invalid value")},
					{Key: []byte("/sn/faas-scheduler/rollout/cluster1/node1/bbb")},
				},
			}
			stopCh := make(chan struct{})
			err := ContendRolloutInEtcd(stopCh)
			convey.So(err, convey.ShouldBeNil)
			close(stopCh)
			time.Sleep(200 * time.Millisecond)
		})
		convey.Convey("putInsSpecForRolloutKey fail", func() {
			SelfInstanceID = ""
			selfInstanceSpec = nil
			getResponse = &clientv3.GetResponse{
				Kvs: []*mvccpb.KeyValue{
					{Key: []byte("/sn/faas-scheduler/rollout/cluster1/node1/aaa"), Value: []byte("invalid value")},
					{Key: []byte("/sn/faas-scheduler/rollout/cluster1/node1/bbb")},
				},
			}
			stopCh := make(chan struct{})
			lockErr = errors.New("some error")
			err := ContendRolloutInEtcd(stopCh)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestReplaceRolloutSubject(t *testing.T) {
	config.GlobalConfig.EnableRollout = true
	defer func() {
		config.GlobalConfig.EnableRollout = false
	}()
	rolloutLocker = &etcd3.EtcdLocker{}
	selfLocker = &etcd3.EtcdLocker{LockedKey: "/sn/faas-scheduler/rollout/cluster1/node1/aaa"}
	var (
		lockErr   error
		unlockErr error
		regErr    error
	)
	patches := []*gomonkey.Patches{
		gomonkey.ApplyFunc((*etcd3.EtcdLocker).TryLock, func(_ *etcd3.EtcdLocker, key string) error {
			return lockErr
		}),
		gomonkey.ApplyFunc((*etcd3.EtcdLocker).Unlock, func(_ *etcd3.EtcdLocker) error {
			return unlockErr
		}),
		gomonkey.ApplyFunc((*etcd3.EtcdRegister).Register, func() error {
			return regErr
		}),
		gomonkey.ApplyFunc((*etcd3.EtcdClient).Delete, func(_ *etcd3.EtcdClient, ctxInfo etcd3.EtcdCtxInfo,
			etcdKey string, opts ...clientv3.OpOption) error {
			return unlockErr
		}),
		gomonkey.ApplyFunc(contendInstanceInEtcd, func(stopCh <-chan struct{}) error {
			return lockErr
		}),
	}
	defer func() {
		for _, p := range patches {
			p.Reset()
		}
	}()
	convey.Convey("Test ReplaceRolloutSubject", t, func() {
		rollout.GetGlobalRolloutHandler().CurrentRatio = 100
		convey.Convey("replace success", func() {
			IsRollingOut = true
			IsRolloutObject = true
			stopCh := make(chan struct{})
			ReplaceRolloutSubject(stopCh)
			convey.So(IsRollingOut, convey.ShouldBeFalse)
			convey.So(IsRolloutObject, convey.ShouldBeFalse)
		})
		convey.Convey("replace fail", func() {
			IsRollingOut = true
			IsRolloutObject = true
			stopCh := make(chan struct{})
			lockErr = errors.New("some error")
			ReplaceRolloutSubject(stopCh)
			convey.So(IsRollingOut, convey.ShouldBeTrue)
			convey.So(IsRolloutObject, convey.ShouldBeTrue)
			lockErr = nil
			unlockErr = errors.New("some error")
			ReplaceRolloutSubject(stopCh)
			convey.So(IsRollingOut, convey.ShouldBeFalse)
			convey.So(IsRolloutObject, convey.ShouldBeFalse)
			lockErr = nil
			unlockErr = nil
			regErr = errors.New("some error")
			ReplaceRolloutSubject(stopCh)
			convey.So(IsRollingOut, convey.ShouldBeFalse)
			convey.So(IsRolloutObject, convey.ShouldBeFalse)
		})
	})
}

func TestPutInsSpecForRolloutKey(t *testing.T) {
	var (
		putErr     error
		lockedKey  string
		rolloutRes *types.RolloutResponse
		rolloutErr error
	)
	rolloutRes = &types.RolloutResponse{}
	patches := []*gomonkey.Patches{
		gomonkey.ApplyFunc((*etcd3.EtcdClient).Put, func(_ *etcd3.EtcdClient, ctxInfo etcd3.EtcdCtxInfo,
			etcdKey string, value string, opts ...clientv3.OpOption) error {
			return putErr
		}),
		gomonkey.ApplyFunc((*etcd3.EtcdLocker).GetLockedKey, func(_ *etcd3.EtcdLocker) string {
			return lockedKey
		}),
		gomonkey.ApplyFunc((*rollout.RFHandler).SendRolloutRequest, func(_ *rollout.RFHandler, selfInsID,
			targetInsID string) (*types.RolloutResponse, error) {
			return rolloutRes, rolloutErr
		}),
	}
	defer func() {
		for _, p := range patches {
			p.Reset()
		}
	}()
	convey.Convey("Test putInsSpecForRolloutKey", t, func() {
		locker := &etcd3.EtcdLocker{EtcdClient: &etcd3.EtcdClient{}}
		putErr = errors.New("some error")
		rolloutErr = errors.New("some error")
		err := putInsSpecForRolloutKey(locker)
		convey.So(err, convey.ShouldNotBeNil)
		lockedKey = "testKey"
		err = putInsSpecForRolloutKey(locker)
		convey.So(err, convey.ShouldNotBeNil)
		SelfInstanceID = "testInstanceID"
		err = putInsSpecForRolloutKey(locker)
		convey.So(err, convey.ShouldNotBeNil)
		selfInstanceSpec = &types.InstanceSpecification{}
		err = putInsSpecForRolloutKey(locker)
		convey.So(err, convey.ShouldNotBeNil)
		// test processRolloutRequest start
		lockedKey = "/sn/faas-scheduler/rollout/cluster1/node1/aaa"
		err = putInsSpecForRolloutKey(locker)
		convey.So(err, convey.ShouldNotBeNil)
		rolloutErr = nil
		err = putInsSpecForRolloutKey(locker)
		convey.So(err, convey.ShouldNotBeNil)
		rolloutRes.RegisterKey = "/sn/faas-scheduler/instance/cluster1/node1/aaa"
		err = putInsSpecForRolloutKey(locker)
		convey.So(err, convey.ShouldNotBeNil)
		// test processRolloutRequest end
		putErr = nil
		err = putInsSpecForRolloutKey(locker)
		convey.So(err, convey.ShouldBeNil)
	})
}

func Test_delInsSpecForRolloutKey(t *testing.T) {
	convey.Convey("Given delInsSpecForRolloutKey function", t, func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		isKeyExistFunc := func(client interface{}, key string) (bool, error) { return true, nil }
		processEtcdPutFunc := func(client interface{}, key, val string) error { return nil }
		patches.ApplyFunc(isKeyExist, isKeyExistFunc)
		patches.ApplyFunc(processEtcdPut, processEtcdPutFunc)

		convey.Convey("When lockedKey is empty", func() {
			locker := &etcd3.EtcdLocker{LockedKey: ""}
			err := delInsSpecForRolloutKey(locker)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "locked key is empty")
		})

		convey.Convey("When key does not exist in etcd", func() {
			isKeyExistFunc = func(client interface{}, key string) (bool, error) { return false, nil }
			patches.ApplyFunc(isKeyExist, isKeyExistFunc)

			locker := &etcd3.EtcdLocker{LockedKey: "test-key"}
			err := delInsSpecForRolloutKey(locker)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "key not exist")
		})

		convey.Convey("When isKeyExist returns an error", func() {
			isKeyExistFunc = func(client interface{}, key string) (bool, error) { return false, errors.New("etcd error") }
			patches.ApplyFunc(isKeyExist, isKeyExistFunc)

			locker := &etcd3.EtcdLocker{LockedKey: "test-key"}
			err := delInsSpecForRolloutKey(locker)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "key not exist or get error")
		})

		convey.Convey("When processEtcdPut fails", func() {
			processEtcdPutFunc = func(client interface{}, key, val string) error { return errors.New("put error") }
			patches.ApplyFunc(processEtcdPut, processEtcdPutFunc)

			locker := &etcd3.EtcdLocker{LockedKey: "test-key"}
			err := delInsSpecForRolloutKey(locker)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "put error")
		})
	})
}

func Test_unsetRolloutRegister(t *testing.T) {
	convey.Convey("Given unsetRolloutRegister function", t, func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		RolloutRegisterKey = "my-key"
		unsetRolloutRegister()
		convey.So(IsRolloutObject, convey.ShouldBeFalse)
		convey.So(RolloutRegisterKey, convey.ShouldBeEmpty)
	})
}
