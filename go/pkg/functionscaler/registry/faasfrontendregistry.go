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

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

// FaaSFrontendRegistry watches frontend instance info event of etcd
type FaaSFrontendRegistry struct {
	watcher          etcd3.Watcher
	ClusterFrontends map[string][]string // cluster frontend instance
	stopCh           <-chan struct{}
	sync.RWMutex
}

// NewFaaSFrontendRegistry will create FaaSFrontendRegistry
func NewFaaSFrontendRegistry(stopCh <-chan struct{}) *FaaSFrontendRegistry {
	faaSFrontendRegistry := &FaaSFrontendRegistry{
		ClusterFrontends: make(map[string][]string, utils.DefaultMapSize),
		stopCh:           stopCh,
	}
	return faaSFrontendRegistry
}

func (ffr *FaaSFrontendRegistry) initWatcher(etcdClient *etcd3.EtcdClient) {
	ffr.watcher = etcd3.NewEtcdWatcher(
		constant.FrontendInstancePrefix,
		ffr.watcherFilter,
		ffr.watcherHandler,
		ffr.stopCh,
		etcdClient)
	ffr.watcher.StartList()
}

// RunWatcher will start etcd watch process for frontend instance info event
func (ffr *FaaSFrontendRegistry) RunWatcher() {
	go ffr.watcher.StartWatch()
}

// GetFrontends -
func (ffr *FaaSFrontendRegistry) GetFrontends(cluster string) []string {
	ffr.RLock()
	frontends := ffr.ClusterFrontends[cluster]
	ffr.RUnlock()
	return frontends
}

// watcherFilter will filter frontend instance info event from etcd event
func (ffr *FaaSFrontendRegistry) watcherFilter(event *etcd3.Event) bool {
	items := strings.Split(event.Key, keySeparator)
	if len(items) != validEtcdKeyLenForFrontend {
		return true
	}
	return false
}

// watcherHandler will handle frontend instance info event from etcd
func (ffr *FaaSFrontendRegistry) watcherHandler(event *etcd3.Event) {
	log.GetLogger().Infof("handling alias event type %s key %s", event.Type, event.Key)

	switch event.Type {
	case etcd3.PUT:
		ffr.updateFrontendInstances(event)
	case etcd3.DELETE:
		ffr.deleteFrontendInstances(event)
	case etcd3.ERROR:
		log.GetLogger().Warnf("etcd error event: %s", event.Value)
	default:
		log.GetLogger().Warnf("unsupported event, key: %s", event.Key)
	}
}

func (ffr *FaaSFrontendRegistry) updateFrontendInstances(ev *etcd3.Event) {
	key := ev.Key
	parts := strings.Split(key, keySeparator)
	if len(parts) != validEtcdKeyLenForFrontend {
		log.GetLogger().Warnf("invalied cluster frontend key:%s, ignore it", key)
		return
	}
	status := string(ev.Value)
	if status != "active" {
		ffr.Lock()
		cluster := parts[clusterFrontendClusterIndex]
		ips := ffr.ClusterFrontends[cluster]
		deleteIndex := -1
		for i, val := range ips {
			if val == parts[clusterFrontendIPIndex] {
				deleteIndex = i
			}
		}
		if deleteIndex != -1 {
			ips = append(ips[:deleteIndex], ips[deleteIndex+1:]...)
		}
		if len(ips) == 0 {
			delete(ffr.ClusterFrontends, cluster)
		} else {
			ffr.ClusterFrontends[cluster] = ips
		}
		ffr.Unlock()
		return
	}
	ffr.Lock()
	cluster := parts[clusterFrontendClusterIndex]
	ips := ffr.ClusterFrontends[cluster]
	if ips == nil {
		ips = make([]string, 0)
		ffr.ClusterFrontends[cluster] = ips
	}
	ffr.ClusterFrontends[cluster] = append(ffr.ClusterFrontends[cluster], parts[clusterFrontendIPIndex])
	ffr.Unlock()
}

func (ffr *FaaSFrontendRegistry) deleteFrontendInstances(ev *etcd3.Event) {
	key := ev.Key
	parts := strings.Split(key, keySeparator)
	if len(parts) != validEtcdKeyLenForFrontend {
		log.GetLogger().Warnf("invalied cluster frontend key:%s, ignore it", key)
		return
	}
	ffr.Lock()
	cluster := parts[clusterFrontendClusterIndex]
	newIPs := make([]string, 0)
	ips := ffr.ClusterFrontends[cluster]
	for _, val := range ips {
		if val != parts[clusterFrontendIPIndex] {
			newIPs = append(newIPs, val)
		}
	}
	if len(newIPs) == 0 {
		delete(ffr.ClusterFrontends, cluster)
	} else {
		ffr.ClusterFrontends[cluster] = newIPs
	}
	ffr.Unlock()
}
