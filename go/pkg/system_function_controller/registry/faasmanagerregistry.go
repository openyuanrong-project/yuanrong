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

// FaaSManagerRegistry watche faasmanager event of etcd
type FaaSManagerRegistry struct {
	watcher         etcd3.Watcher
	managerSpecs    map[string]*types.InstanceSpecification
	subscriberChans []chan types.SubEvent
	stopCh          <-chan struct{}
	sync.RWMutex
}

// NewManagerRegistry will create faaSManagerRegistry
func NewManagerRegistry(stopCh <-chan struct{}) *FaaSManagerRegistry {
	registry := &FaaSManagerRegistry{
		managerSpecs: make(map[string]*types.InstanceSpecification, defaultMapSize),
		stopCh:       stopCh,
	}
	return registry
}

// InitWatcher init watcher
func (fm *FaaSManagerRegistry) InitWatcher() {
	fm.watcher = etcd3.NewEtcdWatcher(constant.InstancePathPrefix,
		fm.watcherFilter, fm.watcherHandler, fm.stopCh, etcd3.GetRouterEtcdClient())
	fm.watcher.StartList()
}

// RunWatcher will start etcd watch process
func (fm *FaaSManagerRegistry) RunWatcher() {
	go fm.watcher.StartWatch()
}

func (fm *FaaSManagerRegistry) getFunctionSpec(instanceID string) *types.InstanceSpecification {
	fm.RLock()
	schedulerSpec := fm.managerSpecs[instanceID]
	fm.RUnlock()
	return schedulerSpec
}

// The etcd key format to be filtered out
func (fm *FaaSManagerRegistry) watcherFilter(event *etcd3.Event) bool {
	etcdKey := event.Key
	log.GetLogger().Infof("watcherFilter get etcdKey=%s", etcdKey)
	items := strings.Split(etcdKey, constant.KeySeparator)
	if len(items) != constant.ValidEtcdKeyLenForInstance ||
		items[constant.TenantIndexForInstance] != "tenant" || items[constant.FunctionIndexForInstance] != "function" ||
		!strings.Contains(etcdKey, "0-system-faasmanager") {
		log.GetLogger().Warnf("invalid faaS manager instance key: %s", etcdKey)
		return true
	}
	return false
}

func (fm *FaaSManagerRegistry) watcherHandler(event *etcd3.Event) {
	log.GetLogger().Infof("handling instance event type %s key %s", event.Type, event.Key)
	if event.Type == etcd3.SYNCED {
		log.GetLogger().Infof("manager registry ready to receive etcd kv")
		return
	}
	instanceID := utils.ExtractInfoFromEtcdKey(event.Key, constant.InstanceIDIndexForInstance)
	if len(instanceID) == 0 {
		log.GetLogger().Errorf("ignoring invalid etcd key of key %s", event.Key)
		return
	}
	fm.Lock()
	defer fm.Unlock()
	switch event.Type {
	case etcd3.PUT, etcd3.HISTORYUPDATE:
		faasManagerSpec := utils.GetInstanceSpecFromEtcdValue(event.Value)
		if faasManagerSpec == nil {
			log.GetLogger().Errorf("ignoring invalid etcd value of key %s", event.Key)
			return
		}
		faasManagerSpec.InstanceID = instanceID
		targetFrontendSpec, needRecover := fm.createOrUpdateManagerSpec(faasManagerSpec)
		if needRecover {
			fm.publishEvent(types.SubEventTypeRecover, targetFrontendSpec)
		} else {
			fm.publishEvent(types.SubEventTypeUpdate, targetFrontendSpec)
		}
		return
	case etcd3.DELETE, etcd3.HISTORYDELETE:
		specification, exist := fm.managerSpecs[instanceID]
		if !exist {
			log.GetLogger().Errorf("function faaS manager %s does not exist in registry", instanceID)
			return
		}
		delete(fm.managerSpecs, instanceID)
		fm.publishEvent(types.SubEventTypeDelete, specification)
	default:
		log.GetLogger().Errorf("unsupported event: %s", event.Key)
		return
	}
}

func (fm *FaaSManagerRegistry) createOrUpdateManagerSpec(
	faasManager *types.InstanceSpecification) (*types.InstanceSpecification, bool) {
	specification, exist := fm.managerSpecs[faasManager.InstanceID]
	if !exist {
		funcCtx, cancelFunc := context.WithCancel(context.TODO())
		faasManager.FuncCtx = funcCtx
		faasManager.CancelFunc = cancelFunc
		fm.managerSpecs[faasManager.InstanceID] = faasManager
		specification = faasManager
	} else {
		specification.InstanceSpecificationMeta = faasManager.InstanceSpecificationMeta
	}
	if utils.CheckNeedRecover(faasManager.InstanceSpecificationMeta) {
		return faasManager, true
	}
	return specification, false
}

// AddSubscriberChan add chan
func (fm *FaaSManagerRegistry) AddSubscriberChan(subChan chan types.SubEvent) {
	fm.Lock()
	fm.subscriberChans = append(fm.subscriberChans, subChan)
	fm.Unlock()
}

func (fm *FaaSManagerRegistry) publishEvent(eventType types.EventType, schedulerSpec *types.InstanceSpecification) {
	for _, subChan := range fm.subscriberChans {
		if subChan != nil {
			subChan <- types.SubEvent{
				EventType: eventType,
				EventKind: types.EventKindManager,
				EventMsg:  schedulerSpec,
			}
		}
	}
}
