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

// Package registry -
package registry

import (
	"context"
	"strings"
	"sync"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/system_function_controller/types"
	"yuanrong.org/kernel/pkg/system_function_controller/utils"
)

// FaaSSchedulerRegistry watches scheduler event of etcd
type FaaSSchedulerRegistry struct {
	schedulerSpecs  map[string]*types.InstanceSpecification
	watcher         etcd3.Watcher
	subscriberChans []chan types.SubEvent
	stopCh          <-chan struct{}
	sync.RWMutex
}

// NewSchedulerRegistry will create faaSSchedulerRegistry
func NewSchedulerRegistry(stopCh <-chan struct{}) *FaaSSchedulerRegistry {
	schedulerRegistry := &FaaSSchedulerRegistry{
		schedulerSpecs: make(map[string]*types.InstanceSpecification, defaultMapSize),
		stopCh:         stopCh,
	}
	return schedulerRegistry
}

// InitWatcher init watcher
func (sr *FaaSSchedulerRegistry) InitWatcher() {
	sr.watcher = etcd3.NewEtcdWatcher(constant.InstancePathPrefix,
		sr.watcherFilter, sr.watcherHandler, sr.stopCh, etcd3.GetRouterEtcdClient())
	sr.watcher.StartList()
}

// RunWatcher will start etcd watch process
func (sr *FaaSSchedulerRegistry) RunWatcher() {
	go sr.watcher.StartWatch()
}

func (sr *FaaSSchedulerRegistry) getFunctionSpec(instanceID string) *types.InstanceSpecification {
	sr.RLock()
	schedulerSpec := sr.schedulerSpecs[instanceID]
	sr.RUnlock()
	return schedulerSpec
}

func (sr *FaaSSchedulerRegistry) watcherFilter(event *etcd3.Event) bool {
	log.GetLogger().Infof("watcherFilter get etcdKey=%s", event.Key)
	items := strings.Split(event.Key, constant.KeySeparator)
	if len(items) != constant.ValidEtcdKeyLenForInstance {
		return true
	}
	if items[constant.FunctionsIndexForInstance] != "instance" || items[constant.TenantIndexForInstance] != "tenant" ||
		items[constant.FunctionIndexForInstance] != "function" {
		return true
	}
	if !strings.Contains(items[constant.FunctionNameIndexForInstance], constant.FaasSchedulerMark) {
		return true
	}
	return false
}

func (sr *FaaSSchedulerRegistry) watcherHandler(event *etcd3.Event) {
	log.GetLogger().Infof("handling function event type %s key %s", event.Type, event.Key)
	instanceID := utils.ExtractInstanceIDFromEtcdKey(event.Key)
	if event.Type == etcd3.SYNCED {
		log.GetLogger().Infof("scheduler registry ready to receive etcd kv")
		return
	}
	if len(instanceID) == 0 {
		log.GetLogger().Warnf("ignore invalid etcd key of key %s", event.Key)
		return
	}
	sr.Lock()
	defer sr.Unlock()
	switch event.Type {
	case etcd3.PUT, etcd3.HISTORYUPDATE:
		targetSchedulerSpec := utils.GetInstanceSpecFromEtcdValue(event.Value)
		if targetSchedulerSpec == nil {
			log.GetLogger().Errorf("ignoring invalid etcd value of key %s", event.Key)
			return
		}
		targetSchedulerSpec.InstanceID = instanceID
		schedulerSpec, needRecover := sr.createOrUpdateSchedulerSpec(targetSchedulerSpec)
		sr.publishEvent(types.SubEventTypeUpdate, schedulerSpec)
		if needRecover {
			sr.publishEvent(types.SubEventTypeRecover, schedulerSpec)
		}
	case etcd3.DELETE, etcd3.HISTORYDELETE:
		schedulerSpec, exist := sr.schedulerSpecs[instanceID]
		if !exist {
			log.GetLogger().Errorf("function faaS scheduler %s doesn't exist in registry", instanceID)
			return
		}
		delete(sr.schedulerSpecs, instanceID)
		sr.publishEvent(types.SubEventTypeDelete, schedulerSpec)
	default:
		log.GetLogger().Errorf("unsupported event: %s", event.Key)
		return
	}
}

func (sr *FaaSSchedulerRegistry) createOrUpdateSchedulerSpec(
	targetSchedulerSpec *types.InstanceSpecification) (*types.InstanceSpecification, bool) {
	schedulerSpec, exist := sr.schedulerSpecs[targetSchedulerSpec.InstanceID]
	if !exist {
		funcCtx, cancelFunc := context.WithCancel(context.TODO())
		targetSchedulerSpec.FuncCtx = funcCtx
		targetSchedulerSpec.CancelFunc = cancelFunc
		sr.schedulerSpecs[targetSchedulerSpec.InstanceID] = targetSchedulerSpec
		schedulerSpec = targetSchedulerSpec
	} else {
		schedulerSpec.InstanceSpecificationMeta = targetSchedulerSpec.InstanceSpecificationMeta
	}
	if utils.CheckNeedRecover(targetSchedulerSpec.InstanceSpecificationMeta) {
		return targetSchedulerSpec, true
	}
	return schedulerSpec, false
}

// AddSubscriberChan add chan
func (sr *FaaSSchedulerRegistry) AddSubscriberChan(subChan chan types.SubEvent) {
	sr.Lock()
	sr.subscriberChans = append(sr.subscriberChans, subChan)
	sr.Unlock()
}

func (sr *FaaSSchedulerRegistry) publishEvent(eventType types.EventType, schedulerSpec *types.InstanceSpecification) {
	for _, subChan := range sr.subscriberChans {
		if subChan != nil {
			subChan <- types.SubEvent{
				EventType: eventType,
				EventKind: types.EventKindScheduler,
				EventMsg:  schedulerSpec,
			}
		}
	}
}
