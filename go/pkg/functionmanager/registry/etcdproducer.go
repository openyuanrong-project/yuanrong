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

// Package registry -
package registry

import (
	"sync/atomic"
	"time"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/functionmanager/types"
)

// EtcdProducer define etcd watcher
type EtcdProducer[T comparable] struct {
	client      *etcd3.EtcdClient
	watchPrefix string
	watchFilter func(event *etcd3.Event) bool

	convertFunc func(eventType types.EventType, event *etcd3.Event) (T, error)
	eventCh     chan T

	synced int32
	logger api.FormatLogger
}

func (e *EtcdProducer[T]) init(stopCh chan struct{}) {
	watcher := etcd3.NewEtcdWatcher(e.watchPrefix, e.watchFilter, func(etcdEvent *etcd3.Event) {
		switch etcdEvent.Type {
		case etcd3.PUT:
			event, err := e.convertFunc(types.SubEventTypeUpdate, etcdEvent)
			if err != nil {
				e.logger.Errorf("convert etcd event %s to vpc event failed %s",
					etcdEvent.Key, err.Error())
				return
			}
			e.eventCh <- event
		case etcd3.DELETE:
			event, err := e.convertFunc(types.SubEventTypeDelete, etcdEvent)
			if err != nil {
				e.logger.Errorf("convert etcd event %s to vpc event failed %s",
					etcdEvent.Key, err.Error())
				return
			}
			e.eventCh <- event
		case etcd3.ERROR:
			e.logger.Warnf("etcd error event: %s", etcdEvent.Value)
		case etcd3.SYNCED:
			atomic.StoreInt32(&e.synced, 1)
		default:
			e.logger.Warnf("unsupported event, key: %s", etcdEvent.Key)
		}
	}, stopCh, e.client)
	watcher.StartWatch()
}

func (e *EtcdProducer[T]) waitForSync(stopCh chan struct{}) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			if atomic.LoadInt32(&e.synced) == 1 {
				e.logger.Info("instance info synced")
				return
			}
		}
	}
}
