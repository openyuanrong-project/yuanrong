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

// Package datasystemclient is data system client used for communicating with data system worker.
// To use data system, you should export the data system lib path. Please refer to the Dockerfile of the frontend.
// The lib should copied to home/sn/bin/datasystem/lib. Please refer to
// functioncore/build/common/common_compile.sh and the Dockerfile of the frontend.
// NOTE: To change the version of data system, must revise the version in the common_compile.sh, test.sh and the go.mod
package datasystemclient

import (
	"errors"
	"math/rand"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
)

const (
	dataSystemKeyWithAZLen    = 5
	dataSystemKeyWithoutAZLen = 4
	dataSystemEndpointsLen    = 2
	noCluster                 = "noCluster"
)

var dataSystemCache = sync.Map{}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Cache -
type Cache struct {
	nodeList     []string
	invalidMap   map[string]struct{}
	lastUsedNode string
	lock         sync.RWMutex
}

func (c *Cache) addNode(node string, logger api.FormatLogger) {
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, v := range c.nodeList {
		if node == v {
			logger.Warnf("the node is already existed, no need add")
			return
		}
	}
	c.nodeList = append(c.nodeList, node)
	logger.Infof("add dataSystem node successfully")
}

func (c *Cache) deleteNode(node string, logger api.FormatLogger) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if !c.delete(node) {
		logger.Warnf("the node is not exist")
		return
	}
	// delete invalid node
	delete(c.invalidMap, node)
	logger.Infof("delete dataSystem node from cache successfully")
}

func (c *Cache) isEmpty() bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	return len(c.nodeList) == 0 && len(c.invalidMap) == 0
}

// must be used in a method with a lock
func (c *Cache) delete(node string) bool {
	l := len(c.nodeList)
	for i, v := range c.nodeList {
		if node == v {
			// no need for order
			// replace the last digit to the index need to delete, and then delete the last digit
			c.nodeList[i] = c.nodeList[l-1]
			c.nodeList = c.nodeList[:l-1]
			return true
		}
	}
	return false
}

func (c *Cache) ifNodeExist(node string) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	for _, v := range c.nodeList {
		if v == node {
			return true
		}
	}
	return false
}

func (c *Cache) getRandomNode() (string, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if len(c.nodeList) == 0 {
		log.GetLogger().Warnf("no data system node is available")
		return "", errors.New("no data system node is available")
	}
	node := c.nodeList[rand.Intn(len(c.nodeList))]
	return node, nil
}

func (c *Cache) getLastUsedNodeWithInvalidNode(invalidNodes []string) (string, error) {
	if c.lastUsedNode != "" && !utils.IsStringInArray(c.lastUsedNode, invalidNodes) {
		return c.lastUsedNode, nil
	}
	node, err := c.getRandomNodeWithInvalidNode(invalidNodes)
	if err == nil {
		c.lastUsedNode = node
	}
	return node, err
}

func (c *Cache) getRandomNodeWithInvalidNode(invalidNodes []string) (string, error) {
	if len(invalidNodes) == 0 {
		return c.getRandomNode()
	}
	c.lock.RLock()
	defer c.lock.RUnlock()
	var nodeList []string
	for _, node := range c.nodeList {
		if !utils.IsStringInArray(node, invalidNodes) {
			nodeList = append(nodeList, node)
		}
	}
	if len(nodeList) == 0 {
		log.GetLogger().Warnf("no data system node is available")
		return "", errors.New("no data system node is available")
	}
	node := nodeList[rand.Intn(len(nodeList))]
	return node, nil
}

func (c *Cache) invalidateNode(node string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if !c.delete(node) {
		log.GetLogger().Warnf("invalid node is already deleted")
		return
	}
	c.invalidMap[node] = struct{}{}
	// the number of failed nodes is not too large, and too many coroutines are not started
	// if used a single-process to traversal map, high-latency operations need add read lock
	go c.healthCheckProcess(node)
}

func (c *Cache) checkInvalidNode(node string) bool {
	c.lock.RLock()
	if _, ok := c.invalidMap[node]; !ok {
		c.lock.RUnlock()
		return false
	}
	c.lock.RUnlock()
	return true
}

func (c *Cache) restoreInvalidNode(node string) {
	c.lock.Lock()
	delete(c.invalidMap, node)
	c.nodeList = append(c.nodeList, node)
	c.lock.Unlock()
}

func (c *Cache) healthCheckProcess(node string) {
	log.GetLogger().Infof("start the health check process, nodeIP: %s", node)
	trigger := time.NewTicker(5 * time.Second) // healthCheck interval
	defer trigger.Stop()
	for {
		<-trigger.C
		if !c.checkInvalidNode(node) {
			log.GetLogger().Warnf("invalid node has been deleted, stop health check process, node: %s", node)
			return
		}

		if !healthCheck(node) {
			continue
		}
		c.restoreInvalidNode(node)
		return
	}
}

func healthCheck(ip string) bool {
	_, err := NewClient("", ip)
	if err != nil {
		return false
	}
	return true
}

func processDataSystemEvent(event *etcd3.Event) {
	logger := log.GetLogger().With(zap.Any("etcdType", event.Type), zap.Any("revisionId", event.Rev))
	logger.Infof("process dataSystem etcd event type")
	switch event.Type {
	case etcd3.PUT:
		err := processAddEvent(event, logger)
		if err != nil {
			logger.Warnf("process data system put event error: %s", err.Error())
		}
	case etcd3.DELETE:
		err := processDeleteEvent(event, logger)
		if err != nil {
			logger.Warnf("process data system delete event error: %s", err.Error())
		}
	case etcd3.ERROR:
		logger.Warnf("etcd error event: %s", event.Value)
	default:
		logger.Warnf("unsupported event: %s", event.Value)
	}
}

func processAddEvent(event *etcd3.Event, logger api.FormatLogger) error {
	logger = logger.With(zap.Any("etcdValue", string(event.Value)))
	ip, az, err := parseDsKey(event.Key)
	if err != nil {
		logger.Errorf("failed to parse dataSystem Key, err: %s", err.Error())
		return err
	}
	cacheData, _ := dataSystemCache.LoadOrStore(az, &Cache{
		nodeList:   []string{},
		invalidMap: make(map[string]struct{}, 1),
	})
	cache, ok := cacheData.(*Cache)
	if !ok {
		return errors.New("dataSystem load from cache is invalid")
	}

	_, status, err := parseDsValue(string(event.Value))
	if err != nil {
		log.GetLogger().Warnf("failed to parse dataSystemValue, err: %s", err.Error())
	}
	localDataSystemStatusCache.SetLocalDataSystemStatus(ip, status)
	readyStatus := map[string]struct{}{
		dataSystemStatusReady: struct{}{}, // only ready status can add
	}
	_, ready := readyStatus[status]
	if err != nil || !ready {
		cache.deleteNode(ip, logger)
		clientMap.deleteNodeIp(ip, logger)
		logger.Warnf("but node is not ready, don't add")
		if isShutdownFronted() {
			destroy()
		}
		return nil
	}

	cache.addNode(ip, logger)
	logger.Warnf("add node to cache")
	return nil
}

func processDeleteEvent(event *etcd3.Event, logger api.FormatLogger) error {
	ip, az, err := parseDsKey(event.Key)
	if err != nil {
		logger.Errorf("failed to parse dataSystem Key, err: %s", err.Error())
		return err
	}
	cacheData, ok := dataSystemCache.Load(az)
	if !ok {
		logger.Warnf("no datasystem node in az %s,no need to delete", az)
		return nil
	}
	cache, ok := cacheData.(*Cache)
	if !ok {
		return errors.New("dataSystem cache is invalid")
	}
	localDataSystemStatusCache.SetLocalDataSystemStatus(ip, "")
	cache.deleteNode(ip, logger)
	if cache.isEmpty() {
		dataSystemCache.Delete(az)
	}
	clientMap.deleteNodeIp(ip, logger)
	if isShutdownFronted() {
		destroy()
	}
	return nil
}

// get ip and az form dataSystem key
// dataSystem key format: /[AZ]/datasystem/cluster/[ip:port]
func parseDsKey(key string) (string, string, error) {
	keys := strings.Split(key, "/")
	if len(keys) != dataSystemKeyWithAZLen && len(keys) != dataSystemKeyWithoutAZLen { // length of dataSystem key
		return "", "", errors.New("invalid length of dataSystem key")
	}
	az := noCluster
	endpoints := strings.Split(keys[len(keys)-1], ":") // index of endpoints in dataSystem key
	if len(endpoints) != dataSystemEndpointsLen {      // length of endpoints key
		return "", "", errors.New("invalid length of endpoints in dataSystem key")
	}
	if len(keys) == dataSystemKeyWithAZLen {
		az = keys[1]
	}
	return endpoints[0], az, nil
}

// get timestamp and status form dataSystem value
// dataSystem value format: 1748573798753243935;ready
func parseDsValue(value string) (string, string, error) {
	splits := strings.Split(value, ";")
	if len(splits) != 2 { // magic number
		return "", "", errors.New("invalid format of dataSystem key")
	}
	timeStamp, status := splits[0], splits[1] // magic number
	return timeStamp, status, nil
}

// dataSystemKeyFilter no need filter
func dataSystemKeyFilter(event *etcd3.Event) bool {
	return false
}

// StartWatch -
func StartWatch(dataSystemKeyPrefixList []string, stopCh <-chan struct{}) {
	etcdClient := etcd3.GetDataSystemEtcdClient()
	if etcdClient == nil {
		etcdClient = etcd3.GetMetaEtcdClient()
		log.GetLogger().Infof("watch dataSystem from meta etcd")
	}
	for _, dataSystemKeyPrefix := range dataSystemKeyPrefixList {
		watcher := etcd3.NewEtcdWatcher(dataSystemKeyPrefix, dataSystemKeyFilter,
			processDataSystemEvent, stopCh, etcdClient)
		watcher.StartWatch()
	}
}
