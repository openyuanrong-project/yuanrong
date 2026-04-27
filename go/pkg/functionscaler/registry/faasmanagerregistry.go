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
	"strings"
	"sync"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/instance"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

// FaaSManagerRegistry watches faas manager instance event of etcd
type FaaSManagerRegistry struct {
	watcher         etcd3.Watcher
	subscriberChans []chan SubEvent
	stopCh          <-chan struct{}
	sync.RWMutex
}

// NewFaaSManagerRegistry will create FaaSManagerRegistry
func NewFaaSManagerRegistry(stopCh <-chan struct{}) *FaaSManagerRegistry {
	faasManagerRegistry := &FaaSManagerRegistry{
		stopCh: stopCh,
	}
	return faasManagerRegistry
}

func (fr *FaaSManagerRegistry) initWatcher(etcdClient *etcd3.EtcdClient) {
	fr.watcher = etcd3.NewEtcdWatcher(
		instanceEtcdPrefix,
		fr.watcherFilter,
		fr.watcherHandler,
		fr.stopCh,
		etcdClient)
	fr.watcher.StartList()
}

// RunWatcher will start etcd watch process for instance event
func (fr *FaaSManagerRegistry) RunWatcher() {
	go fr.watcher.StartWatch()
}

// watcherFilter will filter instance event from etcd event
func (fr *FaaSManagerRegistry) watcherFilter(event *etcd3.Event) bool {
	items := strings.Split(event.Key, keySeparator)
	if len(items) != validEtcdKeyLenForInstance {
		return true
	}
	if items[instanceKeyIndex] != "instance" || items[tenantKeyIndex] != "tenant" ||
		items[functionKeyIndex] != "function" {
		return true
	}
	// also check tenantID
	return !utils.IsFaaSManager(items[functionKeyIndex+1])
}

// watcherHandler will handle instance event from etcd
func (fr *FaaSManagerRegistry) watcherHandler(event *etcd3.Event) {
	log.GetLogger().Infof("handling instance event type %s key %s", event.Type, event.Key)
	if event.Type == etcd3.SYNCED {
		log.GetLogger().Infof("manager registry ready to receive etcd kv")
		return
	}
	instanceID := instance.GetInstanceIDFromEtcdKey(event.Key)
	if len(instanceID) == 0 {
		log.GetLogger().Warnf("ignoring invalid etcd key of key %s", event.Key)
		return
	}
	fr.Lock()
	defer fr.Unlock()
	switch event.Type {
	case etcd3.PUT:
		insSpec := instance.GetInsSpecFromEtcdValue(event.Key, event.Value)
		if insSpec == nil {
			log.GetLogger().Warnf("ignoring invalid etcd value of key %s", event.Key)
			return
		}
		insSpec.InstanceID = instanceID
		fr.publishEvent(SubEventTypeUpdate, insSpec)
	case etcd3.DELETE:
		insSpec := instance.GetInsSpecFromEtcdValue(event.Key, event.PrevValue)
		if insSpec == nil {
			log.GetLogger().Warnf("ignoring invalid etcd value of key %s", event.Key)
			return
		}
		insSpec.InstanceID = instanceID
		fr.publishEvent(SubEventTypeDelete, insSpec)
	default:
		log.GetLogger().Warnf("unsupported event: %#v", event)
	}
}

// addSubscriberChan will add channel, subscribed by FaaSScheduler
func (fr *FaaSManagerRegistry) addSubscriberChan(subChan chan SubEvent) {
	fr.Lock()
	fr.subscriberChans = append(fr.subscriberChans, subChan)
	fr.Unlock()
}

// publishEvent will publish instance event via channel
func (fr *FaaSManagerRegistry) publishEvent(eventType EventType, insSpec *types.InstanceSpecification) {
	for _, subChan := range fr.subscriberChans {
		if subChan != nil {
			subChan <- SubEvent{
				EventType: eventType,
				EventMsg:  insSpec,
			}
		}
	}
}
