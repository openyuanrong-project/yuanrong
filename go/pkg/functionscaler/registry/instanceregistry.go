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

	clientv3 "go.etcd.io/etcd/client/v3"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/instance"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	commonUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
	"yuanrong.org/kernel/pkg/functionscaler/state"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

// InstanceRegistry watches instance event of etcd
type InstanceRegistry struct {
	watcher               etcd3.Watcher
	fgWatcher             etcd3.Watcher
	InstanceIDMap         map[string]*types.Instance
	functionInstanceIDMap map[string]map[string]*commonTypes.InstanceSpecification
	subscriberChans       []chan SubEvent
	listDoneCh            chan struct{}
	fgListDoneCh          chan struct{}
	stopCh                <-chan struct{}
	once                  sync.Once
	sync.RWMutex
}

type getInstanceSpecFunc func(etcdKey string, etcdValue []byte) *commonTypes.InstanceSpecification

// NewInstanceRegistry will create InstanceRegistry
func NewInstanceRegistry(stopCh <-chan struct{}) *InstanceRegistry {
	instanceRegistry := &InstanceRegistry{
		InstanceIDMap:         make(map[string]*types.Instance, utils.DefaultMapSize),
		functionInstanceIDMap: make(map[string]map[string]*commonTypes.InstanceSpecification, utils.DefaultMapSize),
		listDoneCh:            make(chan struct{}, 1),
		fgListDoneCh:          make(chan struct{}, 1),
		stopCh:                stopCh,
	}
	return instanceRegistry
}

// GetInstance -
func (ir *InstanceRegistry) GetInstance(instanceID string) *types.Instance {
	ir.RLock()
	instance := ir.InstanceIDMap[instanceID]
	ir.RUnlock()
	return instance
}

// GetFunctionInstanceIDMap -
func (ir *InstanceRegistry) GetFunctionInstanceIDMap() map[string]map[string]*commonTypes.InstanceSpecification {
	ir.Lock()
	defer ir.Unlock()
	var idMap map[string]map[string]*commonTypes.InstanceSpecification
	commonUtils.DeepCopyObj(ir.functionInstanceIDMap, &idMap)
	return idMap
}

func (ir *InstanceRegistry) initWatcher(etcdClient *etcd3.EtcdClient) {
	ir.watcher = etcd3.NewEtcdWatcher(
		instanceEtcdPrefix,
		ir.watcherFilter,
		ir.watcherHandler,
		ir.stopCh,
		etcdClient)
	ir.watcher.StartList()
	ir.fgWatcher = etcd3.NewEtcdWatcher(
		workersEtcdPrefix,
		ir.watcherFGFilter,
		ir.watcherFGHandler,
		ir.stopCh,
		etcdClient)
	ir.fgWatcher.StartList()
	ir.WaitForETCDList()
}

// WaitForETCDList while recovering, must get all instance including running and creating
func (ir *InstanceRegistry) WaitForETCDList() {
	log.GetLogger().Infof("start to wait instance ETCD list")
	select {
	case <-ir.listDoneCh:
		log.GetLogger().Infof("receive list done, stop waiting ETCD list")
	case <-ir.fgListDoneCh:
		log.GetLogger().Infof("receive fg list done, stop waiting ETCD list")
	case <-ir.stopCh:
		log.GetLogger().Warnf("registry is stopped, stop waiting ETCD list")
	}
}

// RunWatcher will start etcd watch process for instance event
func (ir *InstanceRegistry) RunWatcher() {
	go ir.watcher.StartWatch()
	go ir.fgWatcher.StartWatch()
}

// watcherFilter will filter instance event from etcd event
func (ir *InstanceRegistry) watcherFilter(event *etcd3.Event) bool {
	items := strings.Split(event.Key, keySeparator)
	if len(items) != validEtcdKeyLenForInstance {
		return true
	}
	if items[instanceKeyIndex] != "instance" || items[tenantKeyIndex] != "tenant" ||
		items[functionKeyIndex] != "function" ||
		!strings.HasPrefix(items[executorKeyIndex], "0-system-faasExecutor") &&
			!strings.HasPrefix(items[executorKeyIndex], "0-system-serveExecutor") {
		return true
	}
	return false
}

// watcherHandler will handle instance event from etcd
func (ir *InstanceRegistry) watcherHandler(event *etcd3.Event) {
	log.GetLogger().Infof("handling instance event type %s key %s", event.Type, event.Key)
	if event.Type == etcd3.SYNCED {
		log.GetLogger().Infof("received instance synced event")
		ir.listDoneCh <- struct{}{}
		ir.publishEvent(SubEventTypeSynced, &commonTypes.InstanceSpecification{})
		return
	}
	instanceID := instance.GetInstanceIDFromEtcdKey(event.Key)
	if len(instanceID) == 0 {
		log.GetLogger().Warnf("ignoring invalid etcd key of key %s", event.Key)
		return
	}
	ir.handleEtcdEvent(event, instanceID, instance.GetInsSpecFromEtcdValue)
}

func (ir *InstanceRegistry) handleEtcdEvent(event *etcd3.Event, instanceID string,
	getInstanceSpec getInstanceSpecFunc) {
	ir.Lock()
	defer ir.Unlock()
	switch event.Type {
	case etcd3.PUT, etcd3.HISTORYUPDATE:
		insSpec := getInstanceSpec(event.Key, event.Value)
		if insSpec == nil {
			log.GetLogger().Warnf("ignoring invalid etcd value of key %s", event.Key)
			return
		}
		ir.InstanceIDMap[insSpec.InstanceID] = utils.BuildInstanceFromInsSpec(insSpec, nil)
		functionKey := insSpec.CreateOptions[types.FunctionKeyNote]
		if functionKey == "" {
			log.GetLogger().Warnf("ignoring invalid instance meta data, function is empty")
			return
		}
		if _, ok := ir.functionInstanceIDMap[functionKey]; !ok {
			ir.functionInstanceIDMap[functionKey] = make(map[string]*commonTypes.InstanceSpecification,
				utils.DefaultMapSize)
		}
		ir.functionInstanceIDMap[functionKey][instanceID] = insSpec
		if insSpec.CreateOptions[types.SchedulerIDNote] != selfregister.SelfInstanceID {
			log.GetLogger().Warnf(
				"carefully, instance[%s][%s] dose not created by  this faaSScheduler[%s]", instanceID,
				insSpec.CreateOptions[types.SchedulerIDNote], selfregister.SelfInstanceID)
		}
		insSpec.InstanceID = instanceID
		if event.Type == etcd3.HISTORYUPDATE {
			ir.publishEvent(SubEventTypeUpdate, insSpec)
			return
		}
		log.GetLogger().Infof("put instance event, instanceId: %s, instanceStatus: %v", instanceID,
			insSpec.InstanceStatus)
		ir.publishEvent(SubEventTypeUpdate, insSpec)
		state.Update(int32(event.Rev))
	case etcd3.DELETE, etcd3.HISTORYDELETE:
		insSpec := getInstanceSpec(event.Key, event.PrevValue)
		if insSpec == nil {
			log.GetLogger().Warnf("ignoring invalid etcd value of key %s", event.Key)
			return
		}
		delete(ir.InstanceIDMap, insSpec.InstanceID)
		functionKey := insSpec.CreateOptions[types.FunctionKeyNote]
		if functionKey == "" {
			log.GetLogger().Warnf("ignoring invalid instance meta data, function is empty")
			return
		}
		if instanceIDMap, ok := ir.functionInstanceIDMap[functionKey]; ok {
			delete(instanceIDMap, instanceID)
			if len(instanceIDMap) == 0 {
				delete(ir.functionInstanceIDMap, functionKey)
			}
		} else {
			log.GetLogger().Warnf("no instances of function %s exist", functionKey)
			return
		}
		if insSpec.CreateOptions[types.SchedulerIDNote] != selfregister.SelfInstanceID {
			log.GetLogger().Warnf(
				"carefully, instance[%s] dose not created by this faaSScheduler[%s]", instanceID,
				selfregister.SelfInstanceID)
		}
		insSpec.InstanceID = instanceID
		ir.publishEvent(SubEventTypeDelete, insSpec)
		state.Update(int32(event.Rev))
	default:
		log.GetLogger().Warnf("unsupported event: %#v", event)
	}
}

// addSubscriberChan will add channel, subscribed by FaaSScheduler
func (ir *InstanceRegistry) addSubscriberChan(subChan chan SubEvent) {
	ir.Lock()
	ir.subscriberChans = append(ir.subscriberChans, subChan)
	ir.Unlock()
}

// publishEvent will publish instance event via channel
func (ir *InstanceRegistry) publishEvent(eventType EventType, insSpec *commonTypes.InstanceSpecification) {
	for _, subChan := range ir.subscriberChans {
		if subChan != nil {
			subChan <- SubEvent{
				EventType: eventType,
				EventMsg:  insSpec,
			}
		}
	}
}

// EtcdList -
func (ir *InstanceRegistry) EtcdList() []*commonTypes.InstanceSpecification {
	client := etcd3.GetRouterEtcdClient()
	if client == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), etcd3.DurationContextTimeout)
	etcdCtx := etcd3.EtcdCtxInfo{
		Ctx:    ctx,
		Cancel: cancel,
	}
	res, err := client.Get(etcdCtx, instanceEtcdPrefix, clientv3.WithPrefix())
	if err != nil {
		log.GetLogger().Errorf("get function meta failed, error: %v", err)
		return nil
	}
	var result []*commonTypes.InstanceSpecification
	for _, kv := range res.Kvs {
		e := &etcd3.Event{
			Key:   string(kv.Key),
			Value: kv.Value,
		}
		if ir.watcherFilter(e) {
			continue
		}
		instanceID := instance.GetInstanceIDFromEtcdKey(e.Key)
		if len(instanceID) == 0 {
			log.GetLogger().Warnf("ignoring invalid etcd key of key %s", e.Key)
			continue
		}
		insSpec := instance.GetInsSpecFromEtcdValue(e.Key, e.Value)
		if insSpec == nil {
			log.GetLogger().Warnf("ignoring invalid etcd value of key %s", e.Key)
			continue
		}
		functionKey := insSpec.CreateOptions[types.FunctionKeyNote]
		if functionKey == "" {
			log.GetLogger().Warnf("ignoring invalid instance meta data, function is empty")
			continue
		}
		insSpec.InstanceID = instanceID
		result = append(result, insSpec)
	}
	return result
}
