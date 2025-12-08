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

package logmanager

import (
	"strings"
	"sync"

	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/logservice"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/dashboard/etcdcache"
)

const userStdStreamPrefix = "/log/runtime/std/"

type manager struct {
	Collectors   map[string]collectorClient
	pendingItems chan *logservice.LogItem
	LogDB

	mtx sync.RWMutex // TO protect Collectors
}

var (
	managerSingleton *manager
	logManagerOnce   sync.Once
)

func init() {
	logManagerOnce.Do(func() {
		managerSingleton = &manager{
			Collectors: map[string]collectorClient{},
			LogDB:      newGeneralLogDBImpl(),
		}
		etcdcache.InstanceCache.RegisterInstanceStartHandler(managerSingleton.OnInstanceStart)
		etcdcache.InstanceCache.RegisterInstanceExitHandler(managerSingleton.OnInstanceExit)
	})
}

// RegisterLogCollector -
func (m *manager) RegisterLogCollector(collectorInfo collectorClientInfo) error {
	client := collectorClient{
		collectorClientInfo: collectorInfo,
	}
	err := client.Connect()
	if err != nil {
		return err
	}
	go client.Healthcheck(func() {
		// unregister self if shutdown
		m.UnregisterLogCollector(collectorInfo.ID)
	})
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.Collectors[collectorInfo.ID] = client
	return nil
}

// UnregisterLogCollector -
func (m *manager) UnregisterLogCollector(id string) {
	m.mtx.RLock()
	c, ok := m.Collectors[id]
	m.mtx.RUnlock()
	if !ok {
		return
	}

	m.mtx.Lock()
	delete(m.Collectors, id)
	m.mtx.Unlock()

	if err := c.grpcConn.Close(); err != nil {
		log.GetLogger().Warnf("failed to close connection to collector %s at %s: %s", id, c.Address, err.Error())
	}
}

// GetCollector -
func (m *manager) GetCollector(id string) *collectorClient {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	if c, ok := m.Collectors[id]; ok {
		return &c
	}
	return nil
}

// OnInstanceStart handles event when instance is running
func (m *manager) OnInstanceStart(instance *types.InstanceSpecification) {
	log.GetLogger().Infof("running instance %s started cb", instance.InstanceID)
	if isDriverInstance(instance) {
		// do nothing to driver
		log.GetLogger().Debugf("skip driver instance event of %s", instance.InstanceID)
		return
	}
	// then try to fulfill the log entry
	m.fulfillLogEntryInLogDB(instance)
}

// ReportLogItem handles the report request
func (m *manager) ReportLogItem(item *logservice.LogItem) {
	// match and check the component id (runtime id)
	m.LogDB.Put(NewLogEntry(item), putOptionIfExistsNoop)
	if item.RuntimeID != "" {
		// if not empty, this is a runtime, try match the runtime and try to fulfill the log entry
		m.fulfillLogEntryInLogDB(etcdcache.InstanceCache.GetByRuntimeID(item.RuntimeID))
	}
}

func (m *manager) fulfillLogEntryInLogDB(instance *types.InstanceSpecification) {
	if instance == nil {
		log.GetLogger().Debugf("try fulfill a log entry with nil instance ptr")
		return
	}
	result := m.LogDB.Query(logDBQuery{RuntimeID: instance.RuntimeID})
	result.Range(func(entry *LogEntry) {
		if entry.InstanceID == instance.InstanceID {
			log.GetLogger().Debugf("entry of instance(%s) with filename(%s) on collector(%s) is already been set, "+
				"no update need", instance.InstanceID, entry.Filename, entry.CollectorID)
			return
		}
		entry.InstanceID = instance.InstanceID
		entry.JobID = instance.JobID
		// actually performs an in-place modification, but still put it to avoid some unexpected problem
		m.LogDB.Put(entry, putOptionIfExistsReplace)
	})
}

// OnInstanceExit -
func (m *manager) OnInstanceExit(instance *types.InstanceSpecification) {
	if !isDriverInstance(instance) {
		return
	}
}

func isDriverInstance(instance *types.InstanceSpecification) bool {
	return strings.HasPrefix(instance.InstanceID, "driver") && instance.ParentID == ""
}

// RemoveLogItem handles the remove log request
func (m *manager) RemoveLogItem(item *logservice.LogItem) {
	m.LogDB.Remove(NewLogEntry(item))
}
