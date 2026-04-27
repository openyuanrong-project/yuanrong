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

// Package etcd3 -
package etcd3

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.etcd.io/etcd/client/v3"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

const (
	defaultEventChanSize = 1000
	// DurationContextTimeout default context duration timeout
	DurationContextTimeout = 5 * time.Second
)

var (
	// keepConnAliveTTL -
	keepConnAliveTTL = 10 * time.Second
)

// EtcdCtxInfo etcd context info
type EtcdCtxInfo struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}

// Watcher defines watcher of registry
type Watcher interface {
	StartWatch()
	StartList()
	EtcdHistory(revision int64)
}

// EtcdClientInterface is the interface of ETCD client
type EtcdClientInterface interface {
	GetResponse(ctxInfo EtcdCtxInfo, etcdKey string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
	Put(ctxInfo EtcdCtxInfo, etcdKey string, value string, opts ...clientv3.OpOption) error
	Delete(ctxInfo EtcdCtxInfo, etcdKey string, opts ...clientv3.OpOption) error
}

// NewEtcdWatcher create a EtcdWatcher object
func NewEtcdWatcher(prefix string, filter EtcdWatcherFilter, handler EtcdWatcherHandler, stopCh <-chan struct{},
	etcdClient *EtcdClient) *EtcdWatcher {
	ew := &EtcdWatcher{
		watcher:      etcdClient,
		ResultChan:   make(chan *Event, defaultEventChanSize),
		CacheChan:    make(chan *Event, defaultEventChanSize),
		filter:       filter,
		handler:      handler,
		key:          etcdClient.AttachAZPrefix(prefix),
		resultChanWG: &sync.WaitGroup{},
		configCh:     make(chan struct{}, 1),
		stopCh:       stopCh,
	}
	if etcdClient != nil {
		ew.etcdType = etcdClient.GetEtcdType()
	}
	ew.resultChanWG.Add(1)
	go ew.processEventLoop()
	return ew
}

// etcdList get current events in etcd and handle these events
func (ew *EtcdWatcher) etcdList(handler func(*clientv3.GetResponse)) error {
	opts := []clientv3.OpOption{clientv3.WithPrefix()}
	response, err := ew.watcher.Client.KV.Get(context.TODO(), ew.key, opts...)
	if err != nil {
		log.GetLogger().Errorf("failed to get value from etcd, key: %s, err: %s", ew.key, err.Error())
		return err
	}
	ew.initialRev = response.Header.Revision
	handler(response)
	return nil
}

// EtcdHistory find if delete event happened while recovering
func (ew *EtcdWatcher) EtcdHistory(revision int64) {
	if revision == 0 || revision >= ew.initialRev {
		return
	}
	log.GetLogger().Debugf("start to find key %s history event", ew.key)
	watchOption := []clientv3.OpOption{clientv3.WithPrefix(), clientv3.WithPrevKV(), clientv3.WithRev(revision),
		clientv3.WithProgressNotify()}
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	watchChan := clientv3.NewWatcher(ew.watcher.Client).Watch(ctx, ew.key, watchOption...)
	if watchChan == nil {
		log.GetLogger().Errorf("failed to watch %s, watch channel is empty", ew.key)
		return
	}
	events, ok := <-watchChan
	if !ok {
		log.GetLogger().Warnf("the channel received the result may be closed")
		return
	}
	for _, event := range events.Events {
		ew.sendEvent(parseHistoryEvent(event, ew.etcdType))
	}
}

// StartWatch start watch etcd event
func (ew *EtcdWatcher) StartWatch() {
	go ew.recoverWatch()
	if !ew.watcher.etcdStatusNow {
		log.GetLogger().Warnf("no connection with etcd.")
		return
	}
	go ew.run()
}

// recoverWatch recover watch etcd event when etcd reconnected
func (ew *EtcdWatcher) recoverWatch() {
loop:
	for {
		if ew.watcher.cond == nil {
			log.GetLogger().Warnf("etcd client condition lock is not initialized")
			return
		}
		ew.watcher.cond.L.Lock()
		ew.watcher.cond.Wait()
		ew.watcher.cond.L.Unlock()
		select {
		case <-ew.stopCh:
			break loop
		default:
		}
		go ew.run()
	}
	ew.resultChanWG.Wait()
	close(ew.ResultChan)
}

func (ew *EtcdWatcher) run() {
	log.GetLogger().Infof("start to watch etcd prefix %s", ew.key)
	if ew.watcher.Client == nil {
		log.GetLogger().Errorf("failed to watch %s, etcd client is nil", ew.key)
		return
	}
	if ew.cacheConfig.EnableCache {
		go ew.processETCDCache()
	}
	ew.StartList()
	watchChan, cancel, err := createWatchChan(ew)
	defer cancel()
	if err != nil || watchChan == nil {
		return
	}
	for {
		select {
		case events, ok := <-watchChan:
			if !ok {
				cancel()
				log.GetLogger().Warnf("the channel received the result may be closed")
				watchChan, cancel, err = createWatchChan(ew)
				if err != nil {
					return
				}
				continue
			}
			if events.Err() != nil {
				log.GetLogger().Errorf("etcd receive err events, err:%s", events.Err().Error())
			}
			if ew.historyRev > 0 && ew.historyRev < ew.initialRev {
				ew.EtcdHistory(ew.historyRev)
			}
			for _, event := range events.Events {
				e := parseEvent(event, ew.etcdType)
				ew.initialRev = e.Rev
				ew.historyRev = ew.initialRev
				ew.sendEvent(e)
			}
		case <-ew.stopCh:
			log.GetLogger().Infof("stop watching etcd prefix %s", ew.key)
			return
		case <-ew.watcher.clientExitCh:
			log.GetLogger().Errorf("lost %s etcd client", ew.watcher.etcdType)
			return
		}
	}
}

func createWatchChan(ew *EtcdWatcher) (clientv3.WatchChan, context.CancelFunc, error) {
	watchOption := []clientv3.OpOption{clientv3.WithPrefix(), clientv3.WithPrevKV(),
		clientv3.WithRev(ew.initialRev), clientv3.WithProgressNotify()}
	ctx, cancelFunc := context.WithCancel(context.Background())
	if err := ew.etcdList(func(_ *clientv3.GetResponse) {}); err != nil {
		log.GetLogger().Errorf("failed to etcdList, err: %s", err.Error())
		return nil, cancelFunc, err
	}
	watchChan := clientv3.NewWatcher(ew.watcher.Client).Watch(ctx, ew.key, watchOption...)
	if watchChan == nil {
		log.GetLogger().Errorf("failed to watch %s, watch channel is empty", ew.key)
		return nil, cancelFunc, fmt.Errorf("failed to watch %s, watch channel is empty", ew.key)
	}
	return watchChan, cancelFunc, nil
}

// StartList performs a ETCD List and send corresponding events, revision will be set after list
func (ew *EtcdWatcher) StartList() {
	if ew.initialRev == 0 {
		var restoreErr error
		if ew.cacheConfig.EnableCache {
			restoreErr = ew.restoreCacheFromFile()
		}
		if !ew.cacheConfig.EnableCache || restoreErr != nil {
			if err := ew.etcdList(func(response *clientv3.GetResponse) {
				for _, event := range response.Kvs {
					ew.sendEvent(parseKV(event, ew.etcdType))
				}
			}); err != nil {
				log.GetLogger().Errorf("failed to sync with latest state, error: %s", err.Error())
			}
		}
		// notice watcher, ready to watch etcd kv
		ew.sendEvent(syncedEvent())
		ew.historyRev = ew.initialRev
	}
}

// processEventLoop receive etcd event and process
func (ew *EtcdWatcher) processEventLoop() {
	defer ew.resultChanWG.Done()
	for {
		select {
		case event, ok := <-ew.ResultChan:
			if !ok {
				log.GetLogger().Warnf("event channel is closed, stop processing event")
				return
			}
			if event.Type == SYNCED || !ew.filter(event) {
				ew.handler(event)
			}
		case <-ew.stopCh:
			log.GetLogger().Warnf("stop processing etcd event loop")
			return
		}
	}
}

func (ew *EtcdWatcher) sendEvent(e *Event) {
	if len(ew.ResultChan) == defaultEventChanSize {
		log.GetLogger().Warnf("Fast watcher, slow processing. Number of buffered events: %d."+
			"Probably caused by slow decoding, user not receiving fast, or other processing logic",
			defaultEventChanSize)
	}
	if ew.watcher != nil {
		e.Key = ew.watcher.DetachAZPrefix(e.Key)
	}
	select {
	case ew.ResultChan <- e:
	case <-ew.stopCh:
		log.GetLogger().Warnf("etcd watcher chan closed")
	}
	if ew.cacheConfig.EnableCache && (e.Type == PUT || e.Type == DELETE) {
		select {
		case ew.CacheChan <- e:
		case <-ew.stopCh:
			log.GetLogger().Warnf("etcd watcher chan closed")
		}
	}
}

// GetResponse get etcd value and return pointer of GetResponse struct
func (e *EtcdClient) GetResponse(ctxInfo EtcdCtxInfo, etcdKey string,
	opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	etcdKey = e.AttachAZPrefix(etcdKey)
	ctx, cancel := ctxInfo.Ctx, ctxInfo.Cancel
	defer cancel()

	kv := clientv3.NewKV(e.Client)
	getResp, err := kv.Get(ctx, etcdKey, opts...)

	return getResp, err
}

// Put put context key and value
func (e *EtcdClient) Put(ctxInfo EtcdCtxInfo, etcdKey string, value string, opts ...clientv3.OpOption) error {
	etcdKey = e.AttachAZPrefix(etcdKey)
	ctx, cancel := ctxInfo.Ctx, ctxInfo.Cancel
	defer cancel()

	kv := clientv3.NewKV(e.Client)
	_, err := kv.Put(ctx, etcdKey, value, opts...)
	return err
}

// Delete delete key
func (e *EtcdClient) Delete(ctxInfo EtcdCtxInfo, etcdKey string, opts ...clientv3.OpOption) error {
	etcdKey = e.AttachAZPrefix(etcdKey)
	ctx, cancel := ctxInfo.Ctx, ctxInfo.Cancel
	defer cancel()
	kv := clientv3.NewKV(e.Client)
	_, err := kv.Delete(ctx, etcdKey, opts...)
	return err
}

// Get gets from etcd
func (e *EtcdClient) Get(ctxInfo EtcdCtxInfo, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	key = e.AttachAZPrefix(key)
	ctx, cancel := ctxInfo.Ctx, ctxInfo.Cancel
	defer cancel()
	kv := clientv3.NewKV(e.Client)
	response, err := kv.Get(ctx, key, opts...)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// GetValues return list of object for key
func (e *EtcdClient) GetValues(ctxInfo EtcdCtxInfo, key string, opts ...clientv3.OpOption) ([]string, error) {
	key = e.AttachAZPrefix(key)
	ctx, cancel := ctxInfo.Ctx, ctxInfo.Cancel
	defer cancel()

	kv := clientv3.NewKV(e.Client)
	response, err := kv.Get(ctx, key, opts...)
	if err != nil {
		return nil, err
	}
	values := make([]string, len(response.Kvs))

	for index, v := range response.Kvs {
		values[index] = string(v.Value)
	}
	return values, err
}

// CreateEtcdCtxInfoWithTimeout create a context with timeout, default timeout is DurationContextTimeout
func CreateEtcdCtxInfoWithTimeout(ctx context.Context, duration time.Duration) EtcdCtxInfo {
	ctx, cancel := context.WithTimeout(ctx, duration)
	return EtcdCtxInfo{
		Ctx:    ctx,
		Cancel: cancel,
	}
}
