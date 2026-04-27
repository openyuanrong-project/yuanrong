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

// FaaSFrontendRegistry watche faasfrontend event of etcd
type FaaSFrontendRegistry struct {
	watcher         etcd3.Watcher
	frontendSpecs   map[string]*types.InstanceSpecification
	subscriberChans []chan types.SubEvent
	stopCh          <-chan struct{}
	sync.RWMutex
}

// NewFrontendRegistry will create faaSFrontendRegistry
func NewFrontendRegistry(stopCh <-chan struct{}) *FaaSFrontendRegistry {
	registry := &FaaSFrontendRegistry{
		frontendSpecs: make(map[string]*types.InstanceSpecification, defaultMapSize),
		stopCh:        stopCh,
	}
	return registry
}

// InitWatcher init watcher
func (fr *FaaSFrontendRegistry) InitWatcher() {
	fr.watcher = etcd3.NewEtcdWatcher(constant.InstancePathPrefix,
		fr.watcherFilter, fr.watcherHandler, fr.stopCh, etcd3.GetRouterEtcdClient())
	fr.watcher.StartList()
}

// RunWatcher will start etcd watch process
func (fr *FaaSFrontendRegistry) RunWatcher() {
	go fr.watcher.StartWatch()
}

func (fr *FaaSFrontendRegistry) getFunctionSpec(instanceID string) *types.InstanceSpecification {
	fr.RLock()
	schedulerSpec := fr.frontendSpecs[instanceID]
	fr.RUnlock()
	return schedulerSpec
}

// The etcd key format to be filtered out
// /sn/instance/business/yrk/tenant/0/function/faasfrontend/version/$latest/defaultaz/frontend-xx.xx.xx.xx
func (fr *FaaSFrontendRegistry) watcherFilter(event *etcd3.Event) bool {
	etcdKey := event.Key
	items := strings.Split(etcdKey, constant.KeySeparator)
	if len(items) != constant.ValidEtcdKeyLenForInstance ||
		items[constant.TenantIndexForInstance] != "tenant" || items[constant.FunctionIndexForInstance] != "function" {
		return true
	}
	if !strings.Contains(items[constant.FunctionNameIndexForInstance], constant.FaasFrontendMark) {
		return true
	}
	return false
}

func (fr *FaaSFrontendRegistry) watcherHandler(event *etcd3.Event) {
	log.GetLogger().Infof("handling instance event type %s key %s", event.Type, event.Key)
	if event.Type == etcd3.SYNCED {
		log.GetLogger().Infof("frontend registry ready to receive etcd kv")
		return
	}
	instanceID := utils.ExtractInfoFromEtcdKey(event.Key, constant.InstanceIDIndexForInstance)
	if len(instanceID) == 0 {
		log.GetLogger().Errorf("ignoring invalid etcd key of key %s", event.Key)
		return
	}
	fr.Lock()
	defer fr.Unlock()
	switch event.Type {
	case etcd3.PUT, etcd3.HISTORYUPDATE:
		faasFrontendSpec := utils.GetInstanceSpecFromEtcdValue(event.Value)
		if faasFrontendSpec == nil {
			log.GetLogger().Errorf("ignoring invalid etcd value of key %s", event.Key)
			return
		}
		faasFrontendSpec.InstanceID = instanceID
		targetFrontendSpec, needRecover := fr.createOrUpdateFrontendSpec(faasFrontendSpec)
		fr.publishEvent(types.SubEventTypeUpdate, targetFrontendSpec)
		if needRecover {
			fr.publishEvent(types.SubEventTypeRecover, targetFrontendSpec)
		}
		return
	case etcd3.DELETE, etcd3.HISTORYDELETE:
		specification, exist := fr.frontendSpecs[instanceID]
		if !exist {
			log.GetLogger().Errorf("function faaS frontend %s does not exist in registry", instanceID)
			return
		}
		delete(fr.frontendSpecs, instanceID)
		fr.publishEvent(types.SubEventTypeDelete, specification)
	default:
		log.GetLogger().Errorf("unsupported event: %s", event.Key)
		return
	}
}

func (fr *FaaSFrontendRegistry) createOrUpdateFrontendSpec(
	faasFrontend *types.InstanceSpecification) (*types.InstanceSpecification, bool) {
	specification, exist := fr.frontendSpecs[faasFrontend.InstanceID]
	if !exist {
		funcCtx, cancelFunc := context.WithCancel(context.TODO())
		faasFrontend.FuncCtx = funcCtx
		faasFrontend.CancelFunc = cancelFunc
		fr.frontendSpecs[faasFrontend.InstanceID] = faasFrontend
		specification = faasFrontend
	} else {
		specification.InstanceSpecificationMeta = faasFrontend.InstanceSpecificationMeta
	}
	if utils.CheckNeedRecover(faasFrontend.InstanceSpecificationMeta) {
		return faasFrontend, true
	}
	return specification, false
}

// AddSubscriberChan add chan
func (fr *FaaSFrontendRegistry) AddSubscriberChan(subChan chan types.SubEvent) {
	fr.Lock()
	fr.subscriberChans = append(fr.subscriberChans, subChan)
	fr.Unlock()
}

func (fr *FaaSFrontendRegistry) publishEvent(eventType types.EventType, schedulerSpec *types.InstanceSpecification) {
	for _, subChan := range fr.subscriberChans {
		if subChan != nil {
			subChan <- types.SubEvent{
				EventType: eventType,
				EventKind: types.EventKindFrontend,
				EventMsg:  schedulerSpec,
			}
		}
	}
}
