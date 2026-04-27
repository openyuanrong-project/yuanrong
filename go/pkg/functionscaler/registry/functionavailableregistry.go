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
	"encoding/json"
	"strings"
	"sync"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

// FunctionAvailableRegistry watches instance event of etcd
type FunctionAvailableRegistry struct {
	watcher               etcd3.Watcher
	FuncAvailableClusters map[string][]string // function available clusters
	stopCh                <-chan struct{}
	sync.RWMutex
}

// NewFunctionAvailableRegistry will create FunctionAvailableRegistry
func NewFunctionAvailableRegistry(stopCh <-chan struct{}) *FunctionAvailableRegistry {
	functionAvailableRegistry := &FunctionAvailableRegistry{
		FuncAvailableClusters: make(map[string][]string, utils.DefaultMapSize),
		stopCh:                stopCh,
	}
	return functionAvailableRegistry
}

func (cr *FunctionAvailableRegistry) initWatcher(etcdClient *etcd3.EtcdClient) {
	cr.watcher = etcd3.NewEtcdWatcher(
		constant.FunctionAvailClusterPrefix,
		cr.watcherFilter,
		cr.watcherHandler,
		cr.stopCh,
		etcdClient)
	cr.watcher.StartList()
}

// RunWatcher will start etcd watch process for function available clusters event
func (cr *FunctionAvailableRegistry) RunWatcher() {
	go cr.watcher.StartWatch()
}

// GeClusters -
func (cr *FunctionAvailableRegistry) GeClusters(cluster string) []string {
	cr.RLock()
	clusters := cr.FuncAvailableClusters[cluster]
	cr.RUnlock()
	return clusters
}

// watcherFilter will filter function available clusters event from etcd event
func (cr *FunctionAvailableRegistry) watcherFilter(event *etcd3.Event) bool {
	items := strings.Split(event.Key, keySeparator)
	if len(items) != validEtcdKeyLenForCluster {
		return true
	}
	return false
}

// watcherHandler will handle function available clusters event from etcd
func (cr *FunctionAvailableRegistry) watcherHandler(event *etcd3.Event) {
	log.GetLogger().Infof("handling function available clusters event type %s key %s", event.Type, event.Key)
	clusters, funcURN := GetFuncAvailableClusterFromEtcd(event)
	if clusters == nil {
		log.GetLogger().Warnf("ignoring invalid etcd value of key %s", event.Key)
		return
	}
	switch event.Type {
	case etcd3.PUT:
		cr.Lock()
		cr.FuncAvailableClusters[funcURN] = clusters
		cr.Unlock()
	case etcd3.DELETE:
		cr.Lock()
		delete(cr.FuncAvailableClusters, funcURN)
		cr.Unlock()
	case etcd3.ERROR:
		log.GetLogger().Warnf("etcd error event: %s", event.Value)
	default:
		log.GetLogger().Warnf("unsupported event, key: %s", event.Key)
	}
}

// GetFuncAvailableClusterFromEtcd get function available clusters from etcd
func GetFuncAvailableClusterFromEtcd(ev *etcd3.Event) ([]string, string) {
	clusters := make([]string, 0)
	err := json.Unmarshal(ev.Value, &clusters)
	if err != nil {
		log.GetLogger().Errorf("failed to unmarshal function cluster update event: %s, reason: %s", ev.Value, err)
		return nil, ""
	}

	key := ev.Key
	parts := strings.Split(key, constant.ETCDEventKeySeparator)
	if len(parts) != validEtcdKeyLenForCluster {
		log.GetLogger().Warnf("invalied function cluster key:%s, ignore it", key)
		return nil, ""
	}

	return clusters, parts[functionClusterKeyIdx]
}
