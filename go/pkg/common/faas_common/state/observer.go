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

// Package state -
package state

import (
	"context"
	"fmt"

	"go.etcd.io/etcd/client/v3"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
)

// Observer -
type Observer interface {
	Update(value interface{}, tags ...string) // add & update state to datasystem
}

// Queue is used to cache the state processing queue
type Queue struct {
	client *etcd3.EtcdClient
	queue  chan stateData
}

// stateData is a state input parameter structure
type stateData struct {
	data interface{}
	tags []string
}

const (
	maxQueueSize     = 10000
	defaultQueueSize = 1000
)

// NewStateQueue -
func NewStateQueue(size int) *Queue {
	if size > maxQueueSize || size <= 0 {
		size = defaultQueueSize
	}
	client := etcd3.GetRouterEtcdClient()
	if client == nil {
		return nil
	}
	return &Queue{
		queue:  make(chan stateData, size),
		client: client,
	}
}

// SaveState -
func (q *Queue) SaveState(state []byte, key string) error {
	ctx := etcd3.CreateEtcdCtxInfoWithTimeout(context.Background(), etcd3.DurationContextTimeout)
	return q.client.Put(ctx, key, string(state))
}

// GetState - get state from etcd with key
func (q *Queue) GetState(key string) ([]byte, error) {
	ctx := etcd3.CreateEtcdCtxInfoWithTimeout(context.Background(), etcd3.DurationContextTimeout)
	response, err := q.client.GetResponse(ctx, key, clientv3.WithSerializable())
	if err != nil {
		return nil, err
	}
	if len(response.Kvs) == 0 {
		return nil, fmt.Errorf("get empty state from etcd")
	}
	return response.Kvs[0].Value, nil
}

// DeleteState -
func (q *Queue) DeleteState(key string) error {
	ctx := etcd3.CreateEtcdCtxInfoWithTimeout(context.Background(), etcd3.DurationContextTimeout)
	return q.client.Delete(ctx, key, clientv3.WithPrefix())
}

// Push -
func (q *Queue) Push(value interface{}, tags ...string) error {
	select {
	case q.queue <- stateData{
		data: value,
		tags: tags,
	}:
		return nil
	default:
		return fmt.Errorf("state queue is full, can not write data")
	}
}

// Run -
func (q *Queue) Run(handler func(value interface{}, tags ...string)) {
	for state := range q.queue {
		handler(state.data, state.tags...)
	}
}
