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

	"yuanrong.org/kernel/pkg/common/faas_common/aliasroute"
	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

// AliasRegistry watches instance event of etcd
type AliasRegistry struct {
	watcher         etcd3.Watcher
	subscriberChans []chan SubEvent
	stopCh          <-chan struct{}
	sync.RWMutex
}

// NewAliasRegistry will create InstanceRegistry
func NewAliasRegistry(stopCh <-chan struct{}) *AliasRegistry {
	aliasRegistry := &AliasRegistry{
		stopCh: stopCh,
	}
	return aliasRegistry
}

func (ar *AliasRegistry) initWatcher(etcdClient *etcd3.EtcdClient) {
	ar.watcher = etcd3.NewEtcdWatcher(
		constant.AliasPrefix,
		ar.watcherFilter,
		ar.watcherHandler,
		ar.stopCh,
		etcdClient)
	ar.watcher.StartList()
}

// RunWatcher will start etcd watch process for instance event
func (ar *AliasRegistry) RunWatcher() {
	go ar.watcher.StartWatch()
}

// watcherFilter will filter alias event from etcd event
func (ar *AliasRegistry) watcherFilter(event *etcd3.Event) bool {
	items := strings.Split(event.Key, keySeparator)
	if len(items) != validEtcdKeyLenForAlias {
		return true
	}
	if items[aliasKeyIndex] != "aliases" || items[tenantKeyIndex] != "tenant" ||
		items[functionKeyIndex] != "function" {
		return true
	}
	return false
}

// watcherHandler will handle instance event from etcd
func (ar *AliasRegistry) watcherHandler(event *etcd3.Event) {
	log.GetLogger().Infof("handling alias event type %s key %s", event.Type, event.Key)
	switch event.Type {
	case etcd3.PUT:
		aliasURN, err := aliasroute.ProcessUpdate(event)
		if err != nil {
			return
		}
		ar.publishEvent(SubEventTypeUpdate, aliasURN)
	case etcd3.DELETE:
		aliasURN := aliasroute.ProcessDelete(event)
		ar.publishEvent(SubEventTypeDelete, aliasURN)
	case etcd3.ERROR:
		log.GetLogger().Warnf("etcd error event: %s", event.Value)
	default:
		log.GetLogger().Warnf("unsupported event, key: %s", event.Key)
	}
}

// addSubscriberChan will add channel, subscribed by FaaSScheduler
func (ar *AliasRegistry) addSubscriberChan(subChan chan SubEvent) {
	ar.Lock()
	ar.subscriberChans = append(ar.subscriberChans, subChan)
	ar.Unlock()
}

// publishEvent will publish instance event via channel
func (ar *AliasRegistry) publishEvent(eventType EventType, aliasUrn string) {
	for _, subChan := range ar.subscriberChans {
		if subChan != nil {
			subChan <- SubEvent{
				EventType: eventType,
				EventMsg:  aliasUrn,
			}
		}
	}
}
