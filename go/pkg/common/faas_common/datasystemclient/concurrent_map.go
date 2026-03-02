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
	"sync"

	"go.uber.org/zap"
	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

// clients - key: node IP; value: data system worker instance, *api.DataSystemClient
type nodeIP2ClientMap struct {
	clientMap map[string]DsClientImpl
	sync.RWMutex
	logger api.FormatLogger
}

// DsClientImpl -
type DsClientImpl struct {
	kvClient api.KvClient
}

func (n *nodeIP2ClientMap) deleteAll() {
	n.Lock()
	defer n.Unlock()
	for _, client := range n.clientMap {
		client.kvClient.DestroyClient()
	}
	n.clientMap = make(map[string]DsClientImpl)
	n.logger.Infof("delete all the client")
}

func (n *nodeIP2ClientMap) delete(nodeIp string) {
	n.Lock()
	defer n.Unlock()
	client, ok := n.clientMap[nodeIp]
	if !ok {
		return
	}
	delete(n.clientMap, nodeIp)
	client.kvClient.DestroyClient()
	n.logger.Infof("delete %s client", nodeIp)
}

func (n *nodeIP2ClientMap) get(nodeIp string) (DsClientImpl, bool) {
	n.RLock()
	defer n.RUnlock()
	client, ok := n.clientMap[nodeIp]
	return client, ok
}

func (n *nodeIP2ClientMap) getRandomOne() (DsClientImpl, bool) {
	n.RLock()
	defer n.RUnlock()

	for _, client := range n.clientMap {
		return client, true
	}
	return DsClientImpl{}, false
}

func (n *nodeIP2ClientMap) add(nodeIp string, client DsClientImpl) {
	n.Lock()
	defer n.Unlock()
	if c, ok := n.clientMap[nodeIp]; ok {
		c.kvClient.DestroyClient()
	}
	n.clientMap[nodeIp] = client
	n.logger.Infof("add %s client", nodeIp)
}

func (n *nodeIP2ClientMap) size() int {
	n.RLock()
	defer n.RUnlock()
	return len(n.clientMap)
}

type concurrentMap struct {
	// clients - key: tenantID; value: map
	mp map[string]*nodeIP2ClientMap
	sync.RWMutex
}

func (m *concurrentMap) get(tenantID string, nodeIP string) (DsClientImpl, bool) {
	m.RLock()
	defer m.RUnlock()
	tenantMap, existed := m.mp[tenantID]
	if !existed {
		return DsClientImpl{}, false
	}
	client, existed := tenantMap.get(nodeIP)
	return client, existed
}

func (m *concurrentMap) getOneClient(tenantID string) api.KvClient {
	m.RLock()
	defer m.RUnlock()
	tenantMap, existed := m.mp[tenantID]
	if !existed {
		return nil
	}
	client, ok := tenantMap.getRandomOne()
	if ok {
		return client.kvClient
	}
	return nil
}

func (m *concurrentMap) getOrCreate(tenantID string, nodeIP string) (DsClientImpl, error) {
	m.Lock()
	defer m.Unlock()
	// double check Before creating a thread, perform the get operation again to check whether other threads have been
	// created. If no, continue to create threads to prevent repeated creation.
	tenantMap, existed := m.mp[tenantID]
	if existed {
		if client, existed := tenantMap.get(nodeIP); existed {
			return client, nil
		}
	} else {
		m.mp[tenantID] = &nodeIP2ClientMap{
			clientMap: make(map[string]DsClientImpl),
			RWMutex:   sync.RWMutex{},
			logger:    log.GetLogger().With(zap.Any("tenantId", tenantID)),
		}
	}
	newClient, err := NewClient(tenantID, nodeIP)
	if err != nil {
		return DsClientImpl{}, err
	}
	m.mp[tenantID].add(nodeIP, newClient)
	return newClient, nil
}

// NewClient -
func NewClient(tenantID string, nodeIP string) (DsClientImpl, error) {
	if localClientLibruntime == nil {
		log.GetLogger().Errorf("local dataSystem client is nil")
		return DsClientImpl{}, errors.New("local dataSystem client is nil")
	}
	credential := localClientLibruntime.GetCredential()
	// create
	var dsClient DsClientImpl
	createConfigLibruntime := api.ConnectArguments{
		Host:      nodeIP,
		Port:      port,
		TimeoutMs: timeoutMs,
		TenantID:  tenantID,
		AccessKey: credential.AccessKey,
		SecretKey: credential.SecretKey,
	}
	newClient, err := localClientLibruntime.CreateClient(createConfigLibruntime)
	if err != nil {
		log.GetLogger().Errorf("failed to create dataSystem client: %s", err.Error())
		return dsClient, err
	}
	dsClient.kvClient = newClient
	log.GetLogger().Infof("create new datasystem client nodeIP is: %s,tenantID: %s", nodeIP, tenantID)
	return dsClient, nil
}

func (m *concurrentMap) deleteNodeIp(nodeIp string, logger api.FormatLogger) {
	m.Lock()
	defer m.Unlock()
	deleteEmptyList := make([]string, 0)
	for tenantId, tenantMap := range m.mp {
		tenantMap.delete(nodeIp)
		if tenantMap.size() == 0 {
			deleteEmptyList = append(deleteEmptyList, tenantId)
		}
	}
	for _, k := range deleteEmptyList {
		delete(m.mp, k)
	}
	logger.Infof("delete nodeIp from clientMap ok")
}

func (m *concurrentMap) deleteClient(tenantID string, nodeIP string) {
	m.Lock()
	defer m.Unlock()
	tenantMap, existed := m.mp[tenantID]
	if !existed {
		return
	}
	tenantMap.delete(nodeIP)
	if tenantMap.size() == 0 {
		delete(m.mp, tenantID)
	}
	log.GetLogger().Infof("delete nodeIp: %s, tenantId: %s from clientMap ok", nodeIP, tenantID)
}

func (m *concurrentMap) deleteTenant(tenantID string) {
	m.Lock()
	defer m.Unlock()
	tenantMap, ok := m.mp[tenantID]
	if ok {
		tenantMap.deleteAll()
	}
	delete(m.mp, tenantID)
	log.GetLogger().Infof("delete tenantId: %s from clientMap ok", tenantID)
}
