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
	"sync"

	"go.uber.org/zap"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/instanceconfig"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/functionscaler/config"
)

// InstanceConfigRegistry watches /instances event of etcd
type InstanceConfigRegistry struct {
	watcher         etcd3.Watcher
	subscriberChans []chan SubEvent
	listDoneCh      chan struct{}
	stopCh          <-chan struct{}
	sync.RWMutex
}

// NewInstanceConfigRegistry will create InstanceConfigRegistry
func NewInstanceConfigRegistry(stopCh <-chan struct{}) *InstanceConfigRegistry {
	instanceConfigRegistry := &InstanceConfigRegistry{
		listDoneCh: make(chan struct{}, 1),
		stopCh:     stopCh,
	}
	return instanceConfigRegistry
}

func (ifr *InstanceConfigRegistry) initWatcher(etcdClient *etcd3.EtcdClient) {
	ifr.watcher = etcd3.NewEtcdWatcher(
		instanceconfig.InsConfigEtcdPrefix,
		instanceconfig.GetWatcherFilter(config.GlobalConfig.ClusterID),
		ifr.watcherHandler,
		ifr.stopCh,
		etcdClient)
	ifr.watcher.StartList()
	ifr.WaitForETCDList()
}

// WaitForETCDList -
func (ifr *InstanceConfigRegistry) WaitForETCDList() {
	log.GetLogger().Infof("start to wait instance config ETCD list")
	select {
	case <-ifr.listDoneCh:
		log.GetLogger().Infof("receive list done, stop waiting ETCD list")
	case <-ifr.stopCh:
		log.GetLogger().Warnf("registry is stopped, stop waiting ETCD list")
	}
}

// RunWatcher will start etcd watch process for instance event
func (ifr *InstanceConfigRegistry) RunWatcher() {
	go ifr.watcher.StartWatch()
}

func parseInstanceConfig(event *etcd3.Event) (*instanceconfig.Configuration, error) {
	value := event.Value
	if event.Type == etcd3.DELETE || event.Type == etcd3.HISTORYDELETE {
		value = event.PrevValue
	}
	return instanceconfig.ParseInstanceConfigFromEtcdEvent(event.Key, value)
}

// watcherHandler will handle instance event from etcd
func (ifr *InstanceConfigRegistry) watcherHandler(event *etcd3.Event) {
	logger := log.GetLogger().With(zap.Any("eventType", event.Type), zap.Any("eventKey", event.Key))
	logger.Infof("handling instances info")
	if event.Type == etcd3.SYNCED {
		logger.Infof("received instance config synced event")
		ifr.listDoneCh <- struct{}{}
		ifr.publishEvent(SubEventTypeSynced, &instanceconfig.Configuration{})
		return
	}

	insSpec, err := parseInstanceConfig(event)
	if err != nil {
		log.GetLogger().Warnf("ParseInstanceConfigFromEtcdEvent failed, eventvalue is %s, err: %s",
			string(event.Value), err.Error())
		return
	}

	switch event.Type {
	case etcd3.PUT, etcd3.HISTORYUPDATE:
		ifr.publishEvent(SubEventTypeUpdate, insSpec)
	case etcd3.DELETE, etcd3.HISTORYDELETE:
		ifr.publishEvent(SubEventTypeDelete, insSpec)
	default:
		logger.Warnf("unsupported event")
	}
}

// addSubscriberChan will add channel, subscribed by FaaSScheduler
func (ifr *InstanceConfigRegistry) addSubscriberChan(subChan chan SubEvent) {
	ifr.Lock()
	ifr.subscriberChans = append(ifr.subscriberChans, subChan)
	ifr.Unlock()
}

// publishEvent will publish instance event via channel
func (ifr *InstanceConfigRegistry) publishEvent(eventType EventType, insConfig *instanceconfig.Configuration) {
	for _, subChan := range ifr.subscriberChans {
		if subChan != nil {
			subChan <- SubEvent{
				EventType: eventType,
				EventMsg:  insConfig,
			}
		}
	}
}
