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

package registry

import (
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/rollout"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
)

func TestWatcherFilter(t *testing.T) {
	config.GlobalConfig.ClusterID = "cluster1"
	registry := RolloutRegistry{}
	event := &etcd3.Event{
		Key: "/sn/faas-scheduler/rolloutConfig/cluster1",
	}
	ignore := registry.watcherFilterForConfig(event)
	assert.False(t, ignore)
	event = &etcd3.Event{
		Key: "/sn/faas-scheduler/rolloutConfig",
	}
	ignore = registry.watcherFilterForConfig(event)
	assert.True(t, ignore)
	event = &etcd3.Event{
		Key: "/sn/faas-scheduler/rolloutConfig/cluster2",
	}
	ignore = registry.watcherFilterForConfig(event)
	assert.True(t, ignore)
	selfregister.SelfInstanceID = "instance1"
	event = &etcd3.Event{
		Key: "/sn/faas-scheduler/rollout/cluster1/node1",
	}
	ignore = registry.watcherFilterForRollout(event)
	assert.True(t, ignore)
	event = &etcd3.Event{
		Key: "/sn/faas-scheduler/rollout/cluster1/node1/instance2",
	}
	ignore = registry.watcherFilterForRollout(event)
	assert.True(t, ignore)
	event = &etcd3.Event{
		Key: "/sn/faas-scheduler/rollout/cluster1/node1/instance1",
	}
	ignore = registry.watcherFilterForRollout(event)
	assert.False(t, ignore)
	selfregister.SelfInstanceID = ""
}

func TestInitWatch(t *testing.T) {
	config.GlobalConfig.EnableRollout = true
	defer func() {
		config.GlobalConfig.EnableRollout = false
	}()
	stopCh := make(chan struct{})
	once := sync.Once{}
	rr := NewRolloutRegistry(stopCh)
	listCalled := 0
	watchCalled := 0
	defer gomonkey.ApplyFunc((*etcd3.EtcdWatcher).StartList, func(_ *etcd3.EtcdWatcher) {
		once.Do(func() {
			rr.configDone <- struct{}{}
		})
		listCalled++
	}).Reset()
	defer gomonkey.ApplyFunc((*etcd3.EtcdWatcher).StartWatch, func(_ *etcd3.EtcdWatcher) {
		watchCalled++
	}).Reset()
	convey.Convey("test init watch", t, func() {
		rr.initWatcher(&etcd3.EtcdClient{})
		convey.So(listCalled, convey.ShouldEqual, 2)
		rr.RunWatcher()
		time.Sleep(100 * time.Millisecond)
		convey.So(watchCalled, convey.ShouldEqual, 2)
	})
}

func TestWatchHandlerForConfig(t *testing.T) {
	config.GlobalConfig.EnableRollout = true
	selfregister.IsRolloutObject = true
	defer func() {
		config.GlobalConfig.EnableRollout = false
		selfregister.IsRolloutObject = false
	}()
	stopCh := make(chan struct{})
	rr := NewRolloutRegistry(stopCh)
	event := &etcd3.Event{
		Rev: 1,
	}
	subChan := make(chan SubEvent, 1)
	rr.addSubscriberChan(subChan)
	processCalled := 0
	defer gomonkey.ApplyFunc((*rollout.RFHandler).ProcessAllocRecordSync, func(_ *rollout.RFHandler, selfInsID,
		targetInsID string) {
		processCalled++
	}).Reset()
	convey.Convey("test watch process", t, func() {
		event.Type = etcd3.PUT
		event.Key = "/sn/faas-scheduler/rolloutConfig/cluster1"
		rr.watcherHandlerForConfig(event)
		convey.So(len(subChan), convey.ShouldEqual, 0)
		event.Value = []byte(`{"rolloutRatio":"100%"}`)
		rr.watcherHandlerForConfig(event)
		convey.So(len(subChan), convey.ShouldEqual, 1)
		e := <-subChan
		ratio := e.EventMsg.(int)
		convey.So(ratio, convey.ShouldEqual, 100)
		convey.So(processCalled, convey.ShouldEqual, 1)
		event.Type = etcd3.DELETE
		rr.watcherHandlerForConfig(event)
		convey.So(len(subChan), convey.ShouldEqual, 1)
		e = <-subChan
		ratio = e.EventMsg.(int)
		convey.So(ratio, convey.ShouldEqual, 0)
	})
}

func TestWatchHandlerForRollout(t *testing.T) {
	config.GlobalConfig.EnableRollout = true
	defer func() {
		config.GlobalConfig.EnableRollout = false
	}()
	stopCh := make(chan struct{})
	rr := NewRolloutRegistry(stopCh)
	event := &etcd3.Event{
		Rev: 1,
	}
	convey.Convey("test watch process", t, func() {
		event.Type = etcd3.PUT
		event.Key = "/sn/faas-scheduler/rollout/cluster1/node1/instance1"
		rr.watcherHandlerForRollout(event)
		convey.So(len(rollout.GetGlobalRolloutHandler().ForwardInstance), convey.ShouldEqual, 0)
		event.Value = []byte(`{"instanceID":"aaa"}`)
		rr.watcherHandlerForRollout(event)
		convey.So(rollout.GetGlobalRolloutHandler().ForwardInstance, convey.ShouldEqual, "aaa")
		event.Type = etcd3.DELETE
		rr.watcherHandlerForRollout(event)
		convey.So(len(rollout.GetGlobalRolloutHandler().ForwardInstance), convey.ShouldEqual, 0)
	})
}
