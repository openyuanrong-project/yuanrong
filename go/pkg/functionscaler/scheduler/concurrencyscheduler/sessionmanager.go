/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
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

// Package concurrencyscheduler -
package concurrencyscheduler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/datasystemclient"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/common/uuid"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

// SessionInDS -
type SessionInDS struct {
	SchedulerID                       string
	commonTypes.InstanceSessionConfig `json:",inline"`
}

type sessionRecord struct {
	ctx            context.Context
	timer          *time.Timer
	availThdMap    map[string]struct{}
	allocThdMap    map[string]struct{}
	overAcqThdMap  map[string]struct{}
	expiring       atomic.Value
	ttl            time.Duration
	concurrency    int
	sessionID      string
	expireCancelCh chan struct{}
	expireCh       chan struct{}
	cancelFunc     func()

	insElem *instanceElement
}

func (s *sessionRecord) PutThreadToAvailThdMap(threadID string) error {
	if _, ok := s.allocThdMap[threadID]; !ok {
		return fmt.Errorf("thread %s doesn't belong to session %s for function", threadID, s.sessionID)
	}
	s.availThdMap[threadID] = struct{}{}
	return nil
}

func (s *sessionRecord) PutThreadToAllocThdMap(threadID string) {
	s.allocThdMap[threadID] = struct{}{}
}

func (s *sessionRecord) GetThreadFromAvailThdMap() string {
	var (
		threadID string
	)
	for key := range s.availThdMap {
		threadID = key
		break
	}
	delete(s.availThdMap, threadID)
	return threadID
}

type sessionManager struct {
	currentSchedulerID  string
	sessionMap          map[string]*sessionRecord
	funcKeyWithLabel    string
	recordSaveTrigger   chan struct{}
	recordDeleteTrigger chan struct{}
	currentNode         string
	instanceType        types.InstanceType
	ctx                 context.Context
	cancel              func()
	isFuncOwner         bool
	*sync.RWMutex
}

func makeSessionManager(funcKeyWithLabel string, currentNode string, instanceType types.InstanceType) *sessionManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &sessionManager{
		funcKeyWithLabel:    funcKeyWithLabel,
		currentSchedulerID:  os.Getenv("POD_IP"),
		sessionMap:          make(map[string]*sessionRecord, utils.DefaultMapSize),
		recordSaveTrigger:   make(chan struct{}, recordTriggerChanLen),
		recordDeleteTrigger: make(chan struct{}, recordTriggerChanLen),
		currentNode:         currentNode,
		instanceType:        instanceType,
		ctx:                 ctx,
		cancel:              cancel,
		RWMutex:             &sync.RWMutex{},
	}
}

func (sm *sessionManager) setFuncOwner(isFuncOwner bool) {
	sm.Lock()
	defer sm.Unlock()
	sm.isFuncOwner = isFuncOwner
}

func (sm *sessionManager) getSession(sessionID string) (*sessionRecord, bool) {
	sm.RLock()
	defer sm.RUnlock()
	record, ok := sm.sessionMap[sessionID]
	// if inselem is nil, this is invalid session
	if ok && record.insElem == nil {
		return nil, false
	}
	return record, ok
}

func (sm *sessionManager) addSession(sessionID string, sessionRecord *sessionRecord) {
	sm.Lock()
	sm.sessionMap[sessionID] = sessionRecord
	sm.Unlock()
	sm.triggerSaveSessionRecord()
}

func (sm *sessionManager) delSession(sessionID string) {
	isEmpty := false
	sm.Lock()
	_, exist := sm.sessionMap[sessionID]
	delete(sm.sessionMap, sessionID)
	if len(sm.sessionMap) == 0 {
		isEmpty = true
	}
	sm.Unlock()
	if isEmpty {
		sm.triggerDeleteSessionRecord()
		return
	}
	if exist {
		sm.triggerSaveSessionRecord()
		return
	}
}

func (sm *sessionManager) stopAndClean() {
	sm.cancel()
	if sm.isFuncOwner && config.GlobalConfig.EnableSessionRecover {
		sm.deleteSessionRecordToDataSystem()
	}
}

func (sm *sessionManager) loadSessionFromDataSystem() map[string][]SessionInDS {
	sessionDSCache := map[string][]SessionInDS{}
	resp, err := datasystemclient.KVGetWithRetry(string(sm.instanceType)+sm.funcKeyWithLabel,
		&datasystemclient.Option{
			TenantID: "0",
			NodeIP:   sm.currentNode,
			Cluster:  config.GlobalConfig.DataSystemConfig.CurrentCluster,
		}, uuid.New().String())
	if err != nil || resp == nil {
		log.GetLogger().Warnf("recover sessionInstance from dataSystem failed, err is %s", err.Error())
		return sessionDSCache
	}
	err = json.Unmarshal(resp, &sessionDSCache)
	if err != nil {
		log.GetLogger().Errorf("recover session failed, unmarshal sessCache failed, err is %s", err.Error())
		return sessionDSCache
	}
	return sessionDSCache
}

func (sm *sessionManager) loadSessionWithFilter(filter func(SessionInDS) bool) map[string][]SessionInDS {
	sessionDSCache := sm.loadSessionFromDataSystem()
	for insId, originSessionInfo := range sessionDSCache {
		var sessionInfos []SessionInDS
		for _, info := range originSessionInfo {
			if filter(info) {
				sessionInfos = append(sessionInfos, info)
			}
		}
		if len(sessionInfos) == 0 {
			delete(sessionDSCache, insId)
		} else {
			sessionDSCache[insId] = sessionInfos
		}
	}
	return sessionDSCache
}

func (sm *sessionManager) triggerSaveSessionRecord() {
	if !sm.isFuncOwner {
		return
	}
	select {
	case <-sm.ctx.Done():
	case sm.recordSaveTrigger <- struct{}{}:
	default:
	}
}

func (sm *sessionManager) triggerDeleteSessionRecord() {
	if !sm.isFuncOwner {
		return
	}
	select {
	case <-sm.ctx.Done():
	case sm.recordDeleteTrigger <- struct{}{}:
	default:
	}
}

func (sm *sessionManager) saveOrDeleteSessionRecordLoop() {
	for {
		select {
		case _, ok := <-sm.recordSaveTrigger:
			if !ok {
				continue
			}
			sm.saveSessionRecordToDataSystem()
		case _, ok := <-sm.recordDeleteTrigger:
			if !ok {
				continue
			}
			sm.deleteSessionRecordToDataSystem()
		case <-sm.ctx.Done():
			return
		}
	}
}

func (sm *sessionManager) saveSessionRecordToDataSystem() {
	sm.RLock()
	sessionCache := map[string][]SessionInDS{}
	for _, record := range sm.sessionMap {
		if record.insElem == nil {
			continue
		}
		sessionCache[record.insElem.instance.InstanceID] = append(sessionCache[record.insElem.instance.InstanceID],
			SessionInDS{
				SchedulerID: sm.currentSchedulerID,
				InstanceSessionConfig: commonTypes.InstanceSessionConfig{
					SessionID:   record.sessionID,
					SessionTTL:  int(record.ttl.Seconds()),
					Concurrency: record.concurrency,
				},
			},
		)
	}
	sm.RUnlock()
	if sm.isGrayStatus() {
		sessionDSCache := sm.loadSessionWithFilter(func(ds SessionInDS) bool {
			return ds.SchedulerID != sm.currentSchedulerID
		})
		for insId, sessionInfo := range sessionCache {
			sessionDSCache[insId] = sessionInfo
		}
		sessionCache = sessionDSCache
	}
	jsonByte, err := json.Marshal(sessionCache)
	if err != nil {
		log.GetLogger().Errorf("save session failed, marshal sessCache failed, err is %s", err.Error())
		return
	}
	err = datasystemclient.KVPutWithRetry(string(sm.instanceType)+sm.funcKeyWithLabel, jsonByte,
		&datasystemclient.Option{
			TenantID:  "0",
			NodeIP:    sm.currentNode,
			Cluster:   config.GlobalConfig.DataSystemConfig.CurrentCluster,
			WriteMode: api.WriteModeEnum(config.GlobalConfig.DataSystemConfig.UploadWriteMode),
			TTLSecond: config.GlobalConfig.DataSystemConfig.UploadTTLSec,
		}, uuid.New().String())
	if err != nil {
		log.GetLogger().Errorf("save session failed, put sessCache to datasystem failed, err is %s", err.Error())
		return
	}
}

func (sm *sessionManager) deleteSessionRecordToDataSystem() {
	var err error
	if sm.isGrayStatus() {
		sm.Lock()
		sm.sessionMap = map[string]*sessionRecord{}
		sm.Unlock()
		sm.saveSessionRecordToDataSystem()
		return
	}
	err = datasystemclient.KVDelWithRetry(string(sm.instanceType)+sm.funcKeyWithLabel, &datasystemclient.Option{
		TenantID: "0",
		NodeIP:   sm.currentNode,
		Cluster:  config.GlobalConfig.DataSystemConfig.CurrentCluster,
	}, uuid.New().String())
	if err != nil {
		log.GetLogger().Errorf("delete session failed, delete sessCache to datasystem failed, err is %s", err.Error())
		return
	}
}

func (sm *sessionManager) queryInsBySessionFromDS(sessionId string) string {
	sessionCache := sm.loadSessionFromDataSystem()
	for insId, sessionInDS := range sessionCache {
		for _, sessionInfo := range sessionInDS {
			if sessionId == sessionInfo.SessionID {
				return insId
			}
		}
	}
	return ""
}

func (sm *sessionManager) isGrayStatus() bool {
	return selfregister.IsRollingOut
}

func makeSessionCacheKey(funcName, funcKeyWithRes string) string {
	hash := sha256.Sum256([]byte(funcKeyWithRes))
	hashStr := hex.EncodeToString(hash[:])[:16] // hashStr len is 16
	return fmt.Sprintf("sessioncache-%s-%s", funcName, hashStr)
}
