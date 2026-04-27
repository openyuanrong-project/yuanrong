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

package etcd3

import (
	"errors"
	"reflect"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/client/v3"
)

func Test_syncedEvent(t *testing.T) {
	convey.Convey("syncedEvent", t, func() {
		event := syncedEvent()
		convey.So(event.Type, convey.ShouldEqual, SYNCED)
	})
}

func Test_parseKV(t *testing.T) {
	convey.Convey("parseKV", t, func() {
		kv := parseKV(&mvccpb.KeyValue{Key: []byte("key1"), Value: []byte("value1")}, Router)
		convey.So(kv.Key, convey.ShouldEqual, "key1")
		convey.So(string(kv.Value), convey.ShouldEqual, "value1")
		convey.So(kv.Type, convey.ShouldEqual, PUT)
	})
}

func Test_parseEvent(t *testing.T) {
	convey.Convey("parseEvent", t, func() {
		event := parseEvent(&clientv3.Event{
			Type:   DELETE,
			Kv:     &mvccpb.KeyValue{Key: []byte("key1"), Value: []byte("value1")},
			PrevKv: &mvccpb.KeyValue{Key: []byte("key2"), Value: []byte("value2")},
		}, Router)
		convey.So(event.Type, convey.ShouldEqual, DELETE)
		convey.So(event.Key, convey.ShouldEqual, "key1")
		convey.So(string(event.Value), convey.ShouldEqual, "value1")
		convey.So(string(event.PrevValue), convey.ShouldEqual, "value2")
	})
}

func Test_parseErr(t *testing.T) {
	convey.Convey("parseErr", t, func() {
		err := parseErr(errors.New("parseErr"), Router)
		convey.So(err.Type, convey.ShouldEqual, ERROR)
		convey.So(string(err.Value), convey.ShouldEqual, "parseErr")
	})
}

func Test_parseHistoryEvent(t *testing.T) {
	type args struct {
		e *clientv3.Event
	}
	tests := []struct {
		name     string
		args     args
		wantType int
	}{
		{"case1", args{e: &clientv3.Event{
			Type:   DELETE,
			Kv:     &mvccpb.KeyValue{Key: []byte("key1"), Value: []byte("value1")},
			PrevKv: &mvccpb.KeyValue{Key: []byte("key2"), Value: []byte("value2")},
		}}, HISTORYDELETE},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseHistoryEvent(tt.args.e, Router); !reflect.DeepEqual(got.Type, tt.wantType) {
				t.Errorf("parseHistoryEvent() = %v, want %v", got.Type, tt.wantType)
			}
		})
	}
}
