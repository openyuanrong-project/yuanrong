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
	"fmt"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/smartystreets/goconvey/convey"
	"go.etcd.io/etcd/api/v3/mvccpb"
)

func TestParseKV(t *testing.T) {
	kv := &mvccpb.KeyValue{
		Key:         []byte("/sn/workeragent/abc"),
		Value:       []byte("value_abc"),
		ModRevision: 1,
	}
	res := parseKV(kv)
	if res == nil {
		t.Errorf("failed to parse kv")
	}
}

func TestParseEvent(t *testing.T) {
	e := &clientv3.Event{
		Type: clientv3.EventTypeDelete,
		Kv: &mvccpb.KeyValue{
			Key:         []byte("/sn/workeragent/abc"),
			Value:       []byte("value_abc"),
			ModRevision: 1,
		},
	}

	res := parseEvent(e)
	if res == nil {
		t.Errorf("failed to parse event")
	}
	e = &clientv3.Event{
		Type: clientv3.EventTypePut,
		Kv: &mvccpb.KeyValue{
			Key:         []byte("/sn/workeragent/abc"),
			Value:       []byte("value_abc"),
			ModRevision: 1,
		},
		PrevKv: &mvccpb.KeyValue{
			Key:         []byte("/sn/workeragent/def"),
			Value:       []byte("value_def"),
			ModRevision: 1,
		},
	}
	res = parseEvent(e)
	convey.ShouldNotBeNil(res)
}
func TestParseErr(t *testing.T) {
	res := parseErr(fmt.Errorf("test"))
	convey.ShouldNotBeNil(res)
}

func TestParseSync(t *testing.T) {
	res := parseSync(time.Time{})
	convey.ShouldNotBeNil(res)
}
