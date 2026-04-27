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

// Package etcdcache caches all etcd events listened from etcd
package etcdcache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	clientv3 "go.etcd.io/etcd/client/v3"

	"yuanrong.org/kernel/pkg/common/etcdkey"
	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/types"
)

var (
	// InstanceCache is the cache struct instance for function system instances
	InstanceCache instanceCache
)

func init() {
	InstanceCache = instanceCache{
		id2Instance:        map[string]*types.InstanceSpecification{},
		jobID2Instances:    map[string]map[string]*types.InstanceSpecification{},
		runtimeID2Instance: map[string]*types.InstanceSpecification{},
	}
}

type instanceCache struct {
	// instanceID => instanceInfo
	id2Instance        map[string]*types.InstanceSpecification
	jobID2Instances    map[string]map[string]*types.InstanceSpecification
	runtimeID2Instance map[string]*types.InstanceSpecification
	iMtx               sync.RWMutex

	instanceExitHandler  func(instance *types.InstanceSpecification)
	instanceStartHandler func(instance *types.InstanceSpecification)
}

// Put an instance
func (c *instanceCache) Put(instance *types.InstanceSpecification) {
	log.GetLogger().Infof("Put instance: %s %#v with runtime id %s", instance.InstanceID,
		instance.InstanceStatus.Code, instance.RuntimeID)
	c.iMtx.Lock()
	defer c.iMtx.Unlock()
	c.id2Instance[instance.InstanceID] = instance
	if _, ok := c.jobID2Instances[instance.JobID]; ok {
		c.jobID2Instances[instance.JobID][instance.InstanceID] = instance
	} else {
		c.jobID2Instances[instance.JobID] = map[string]*types.InstanceSpecification{instance.InstanceID: instance}
	}
	c.runtimeID2Instance[instance.RuntimeID] = instance

	if c.instanceStartHandler != nil && instance.InstanceStatus.Code == int32(constant.KernelInstanceStatusRunning) {
		log.GetLogger().Infof("instance %s started", instance.InstanceID)
		go c.instanceStartHandler(instance)
	}
}

// Remove an instance
func (c *instanceCache) Remove(instanceID string) {
	c.iMtx.Lock()
	defer c.iMtx.Unlock()
	inst, ok := c.id2Instance[instanceID]
	if ok {
		delete(c.jobID2Instances, inst.JobID)
		delete(c.runtimeID2Instance, inst.RuntimeID) // Assume no restart... or it will be left here
	}
	delete(c.id2Instance, instanceID)
	if c.instanceExitHandler != nil && inst != nil {
		go c.instanceExitHandler(inst)
	}
}

// RegisterInstanceExitHandler an instance
func (c *instanceCache) RegisterInstanceExitHandler(handler func(instance *types.InstanceSpecification)) {
	c.iMtx.Lock()
	defer c.iMtx.Unlock()
	if c.instanceExitHandler == nil { // in case someone re-register the handler
		log.GetLogger().Infof("instance exit handler registered")
		c.instanceExitHandler = handler
	}
}

// RegisterInstanceStartHandler an instance
func (c *instanceCache) RegisterInstanceStartHandler(handler func(instance *types.InstanceSpecification)) {
	c.iMtx.Lock()
	defer c.iMtx.Unlock()
	if c.instanceStartHandler == nil { // in case someone re-register the handler
		log.GetLogger().Infof("instance start handler registered")
		c.instanceStartHandler = handler
	}
}

// Get an instance by instanceID
func (c *instanceCache) Get(instanceID string) *types.InstanceSpecification {
	c.iMtx.RLock()
	defer c.iMtx.RUnlock()
	log.GetLogger().Infof("Get an instance by id: %s", instanceID)
	return c.id2Instance[instanceID]
}

// GetByJobID a map of instance
func (c *instanceCache) GetByJobID(jobID string) map[string]*types.InstanceSpecification {
	c.iMtx.RLock()
	defer c.iMtx.RUnlock()
	return c.jobID2Instances[jobID]
}

// GetByParentID a map of instance
func (c *instanceCache) GetByParentID(ParentInstanceID string) map[string]*types.InstanceSpecification {
	c.iMtx.RLock()
	defer c.iMtx.RUnlock()
	for _, inst := range c.id2Instance {
		if inst.ParentID == ParentInstanceID {
			return c.jobID2Instances[inst.JobID]
		}
	}
	return map[string]*types.InstanceSpecification{}
}

// GetByRuntimeID a instance
func (c *instanceCache) GetByRuntimeID(runtimeID string) *types.InstanceSpecification {
	c.iMtx.RLock()
	defer c.iMtx.RUnlock()
	return c.runtimeID2Instance[runtimeID]
}

// String -
func (c *instanceCache) String() string {
	c.iMtx.RLock()
	defer c.iMtx.RUnlock()
	return fmt.Sprintf("cache:{%#v}", c.id2Instance)
}

// StartWatchInstance to watch the instance faas schedulers by the etcd
func StartWatchInstance(stopCh <-chan struct{}) {
	etcdClient := etcd3.GetRouterEtcdClient()
	// no filter, always return false
	watcher := etcd3.NewEtcdWatcher(constant.InstancePathPrefix, func(_ *etcd3.Event) bool { return false },
		instanceHandler, stopCh, etcdClient)
	watcher.StartWatch()
}

// SyncAllInstances will get all instances from etcd
func SyncAllInstances() error {
	etcdClient := etcd3.GetRouterEtcdClient()
	log.GetLogger().Infof("etcdclient: %v", etcdClient)
	getResponse, err := etcdClient.Get(
		etcd3.CreateEtcdCtxInfoWithTimeout(context.Background(), etcd3.DurationContextTimeout),
		constant.InstancePathPrefix,
		clientv3.WithPrefix())
	if err != nil {
		log.GetLogger().Errorf("failed to get all instance info, %s", err.Error())
		return err
	}
	for _, kv := range getResponse.Kvs {
		instance := &types.InstanceSpecification{}
		err := json.Unmarshal(kv.Value, instance)
		if err != nil {
			log.GetLogger().Warnf("failed to unmarshal synced kv's value (key: %s)", kv.Key)
			return err
		}
		InstanceCache.Put(instance)
	}
	return nil
}

func instanceHandler(event *etcd3.Event) {
	log.GetLogger().Debugf("handling instance event type %d, key:%s", event.Type, event.Key)

	switch event.Type {
	case etcd3.PUT:
		instance := &types.InstanceSpecification{}
		err := json.Unmarshal(event.Value, instance)
		if err != nil {
			log.GetLogger().Warnf("failed to unmarshal watched event's value (key: %s)", event.Key)
			return
		}
		InstanceCache.Put(instance)
	case etcd3.DELETE:
		etcdInstanceKey := etcdkey.FunctionInstanceKey{}
		err := etcdInstanceKey.ParseFrom(event.Key)
		if err != nil {
			log.GetLogger().Warnf("failed to unmarshal watched event's key: %s, err: %s", event.Key, err)
			return
		}
		InstanceCache.Remove(etcdInstanceKey.InstanceID)
	default:
		log.GetLogger().Debugf("unsupported event: %#v", event)
	}
}
