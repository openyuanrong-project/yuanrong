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

// Package etcd3 implements crud and watch operations based etcd clientv3
package etcd3

import (
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	// PUT event
	PUT = iota
	// DELETE event
	DELETE
	// ERROR unexpected event
	ERROR
	// SYNCED synced event
	SYNCED
)

// Event of databases
type Event struct {
	Type      int
	Key       string
	Value     []byte
	PrevValue []byte
	Rev       int64
}

// parseKV converts a KeyValue retrieved from an initial sync() listing to a synthetic isCreated event.
func parseKV(kv *mvccpb.KeyValue) *Event {
	return &Event{
		Type:      PUT,
		Key:       string(kv.Key),
		Value:     kv.Value,
		PrevValue: nil,
		Rev:       kv.ModRevision,
	}
}

func parseEvent(e *clientv3.Event) *Event {
	eType := 0
	if e.Type == clientv3.EventTypeDelete {
		eType = DELETE
	}
	ret := &Event{
		Type:  eType,
		Key:   string(e.Kv.Key),
		Value: e.Kv.Value,
		Rev:   e.Kv.ModRevision,
	}
	if e.PrevKv != nil {
		ret.PrevValue = e.PrevKv.Value
	}
	return ret
}

func parseErr(err error) *Event {
	return &Event{Type: ERROR, Value: []byte(err.Error())}
}

func parseSync(t time.Time) *Event {
	return &Event{Type: SYNCED, Value: []byte(t.String())}
}
