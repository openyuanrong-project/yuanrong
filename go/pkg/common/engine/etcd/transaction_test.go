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

package etcd

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey"
	. "github.com/smartystreets/goconvey/convey"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"

	"yuanrong.org/kernel/pkg/common/etcd3"
)

// 给一个prefix，从rds里找出对应的reads
func TestPrintCachedPrefixValue(t *testing.T) {
	kv1 := &mvccpb.KeyValue{
		Key:         []byte("mock-key"),
		Value:       []byte("mock-key"),
		ModRevision: 1,
		Version:     1,
	}
	tr := &Transaction{
		rds: map[string]*reads{
			"mock-prefix": {resp: &clientv3.GetResponse{
				Kvs: []*mvccpb.KeyValue{kv1},
			}},
		},
	}
	Convey("Test PrintCachedPrefixValue", t, func() {
		tr.printCachedPrefixValue("mock-prefix", "mock-key")
		tr.printCachedPrefixValue("mock-prefix00", "mock-key")
	})
}

func TestPrintPrefixValue(t *testing.T) {
	Convey("Test printPrefixValue", t, func() {
		kv1 := &mvccpb.KeyValue{
			Key:            []byte("mock-key"),
			Value:          []byte("mock-key"),
			CreateRevision: 1,
			ModRevision:    1,
			Version:        1,
		}
		resp := &clientv3.GetResponse{
			Kvs: []*mvccpb.KeyValue{kv1},
		}
		tr := &Transaction{
			rds: map[string]*reads{
				"mock-prefix": {resp: resp},
			},
		}
		tr.printPrefixValue(resp, "mock-prefix", 0)
	})
}

func TestRereadPrefix(t *testing.T) {
	kv1 := &mvccpb.KeyValue{
		Key:   []byte("mock-key"),
		Value: []byte("mock-key"),
	}
	resp := &clientv3.GetResponse{
		Kvs: []*mvccpb.KeyValue{kv1},
	}
	tr := &Transaction{
		rds: map[string]*reads{
			"mock-prefix": {resp: resp},
		},
		ctx:        context.TODO(),
		etcdClient: &etcd3.EtcdClient{},
	}
	Convey("Test RereadPrefix", t, func() {
		Convey("with err", func() {
			patch := gomonkey.ApplyMethod(reflect.TypeOf(&etcd3.EtcdClient{}), "GetResponse",
				func(w *etcd3.EtcdClient, ctxInfo etcd3.EtcdCtxInfo, etcdKey string,
					opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
					return nil, errors.New("mock err")
				})
			defer patch.Reset()
			err := tr.rereadPrefix("mock-prefix", 0)
			So(err, ShouldNotBeNil)
		})
		Convey("without err", func() {
			patch := gomonkey.ApplyMethod(reflect.TypeOf(&etcd3.EtcdClient{}), "GetResponse",
				func(w *etcd3.EtcdClient, ctxInfo etcd3.EtcdCtxInfo, etcdKey string,
					opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
					return resp, nil
				})
			defer patch.Reset()
			err := tr.rereadPrefix("mock-prefix", 0)
			So(err, ShouldBeNil)
		})
	})
}

func TestRereadKey(t *testing.T) {
	kv1 := &mvccpb.KeyValue{
		Key:            []byte("mock-key"),
		Value:          []byte("mock-key"),
		CreateRevision: 1,
		ModRevision:    1,
		Version:        1,
	}
	resp := &clientv3.GetResponse{
		Kvs: []*mvccpb.KeyValue{kv1},
	}
	tr := &Transaction{
		rds: map[string]*reads{
			"mock-prefix": {resp: resp},
		},
		ctx:        context.TODO(),
		etcdClient: &etcd3.EtcdClient{},
	}
	Convey("Test rereadKey", t, func() {
		Convey("with err", func() {
			patch := gomonkey.ApplyMethod(reflect.TypeOf(&etcd3.EtcdClient{}), "GetResponse",
				func(w *etcd3.EtcdClient, ctxInfo etcd3.EtcdCtxInfo, etcdKey string,
					opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
					return nil, errors.New("mock err")
				})
			defer patch.Reset()
			err := tr.rereadKey("mock-prefix", 0)
			So(err, ShouldNotBeNil)
		})
		Convey("without err", func() {
			patch := gomonkey.ApplyMethod(reflect.TypeOf(&etcd3.EtcdClient{}), "GetResponse",
				func(w *etcd3.EtcdClient, ctxInfo etcd3.EtcdCtxInfo, etcdKey string,
					opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
					return resp, nil
				})
			defer patch.Reset()
			err := tr.rereadKey("mock-prefix", 0)
			So(err, ShouldBeNil)
		})
		Convey("with printCachedKeyValue err", func() {
			patch := gomonkey.ApplyMethod(reflect.TypeOf(&etcd3.EtcdClient{}), "GetResponse",
				func(w *etcd3.EtcdClient, ctxInfo etcd3.EtcdCtxInfo, etcdKey string,
					opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
					return resp, nil
				})
			defer patch.Reset()
			delete(tr.rds, "mock-prefix")
			err := tr.rereadKey("mock-prefix", 0)
			So(err, ShouldBeNil)
		})
	})
}

func TestPrintError(t *testing.T) {
	kv1 := &mvccpb.KeyValue{Key: []byte("mock-key"), Value: []byte("mock-key"), CreateRevision: 1,
		ModRevision: 1, Version: 1,
	}
	kv2 := &mvccpb.KeyValue{Key: []byte("mock-key2"), Value: []byte("mock-key2"), CreateRevision: 1,
		ModRevision: 1, Version: 1,
	}
	resp := &clientv3.GetResponse{Kvs: []*mvccpb.KeyValue{kv1, kv2}}
	tr := &Transaction{
		rds: map[string]*reads{
			"mock-prefix":  {resp: resp, withPrefix: true},
			"mock-prefix2": {resp: resp, withPrefix: false}},
		ctx:        context.TODO(),
		etcdClient: &etcd3.EtcdClient{},
	}
	Convey("Test printError", t, func() {
		patch := gomonkey.ApplyMethod(reflect.TypeOf(&etcd3.EtcdClient{}), "GetResponse",
			func(w *etcd3.EtcdClient, ctxInfo etcd3.EtcdCtxInfo, etcdKey string,
				opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
				return resp, nil
			})
		defer patch.Reset()
		tr.printError()
	})
}
