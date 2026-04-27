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

// Package etcd3 event
package etcd3

import (
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/client/v3"
)

const (
	// PUT event
	PUT = iota
	// DELETE event
	DELETE
	// HISTORYDELETE event
	HISTORYDELETE
	// HISTORYUPDATE event
	HISTORYUPDATE
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
	ETCDType  string
}

// only type can be used
// notice watcher, ready to watch etcd kv.
func syncedEvent() *Event {
	return &Event{
		Type:      SYNCED,
		Key:       "",
		Value:     nil,
		PrevValue: nil,
		Rev:       0,
		ETCDType:  "",
	}
}

// parseKV converts a KeyValue retrieved from an initial sync() listing to a synthetic isCreated event.
func parseKV(kv *mvccpb.KeyValue, etcdType string) *Event {
	return &Event{
		Type:      PUT,
		Key:       string(kv.Key),
		Value:     kv.Value,
		PrevValue: nil,
		Rev:       kv.ModRevision,
		ETCDType:  etcdType,
	}
}

func parseEvent(e *clientv3.Event, etcdType string) *Event {
	eType := PUT
	if e.Type == clientv3.EventTypeDelete {
		eType = DELETE
	}
	ret := &Event{
		Type:     eType,
		Key:      string(e.Kv.Key),
		Value:    e.Kv.Value,
		Rev:      e.Kv.ModRevision,
		ETCDType: etcdType,
	}
	if e.PrevKv != nil {
		ret.PrevValue = e.PrevKv.Value
	}
	return ret
}

func parseHistoryEvent(e *clientv3.Event, etcdType string) *Event {
	event := parseEvent(e, etcdType)
	if event.Type == DELETE {
		event.Type = HISTORYDELETE
	}
	if event.Type == PUT {
		event.Type = HISTORYUPDATE
	}
	return event
}

func parseErr(err error, source string) *Event {
	return &Event{Type: ERROR, Value: []byte(err.Error()), ETCDType: source}
}
