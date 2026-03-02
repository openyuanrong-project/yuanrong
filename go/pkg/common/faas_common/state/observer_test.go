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
	"testing"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
)

func TestNewStateQueue(t *testing.T) {
	defer gomonkey.ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
		return &etcd3.EtcdClient{}
	}).Reset()
	convey.Convey("get queue", t, func() {
		q := NewStateQueue(10)
		q.queue <- stateData{}
		convey.So(len(q.queue), convey.ShouldEqual, 1)
	})
	convey.Convey("get queue", t, func() {
		q := NewStateQueue(-1)
		q.queue <- stateData{}
		convey.So(len(q.queue), convey.ShouldEqual, 1)
	})
}

func TestStateOperation(t *testing.T) {
	defer gomonkey.ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
		return &etcd3.EtcdClient{}
	}).Reset()
	q := NewStateQueue(10)
	convey.Convey("save state", t, func() {
		defer gomonkey.ApplyFunc((*etcd3.EtcdClient).Put, func(_ *etcd3.EtcdClient, ctxInfo etcd3.EtcdCtxInfo,
			etcdKey string, value string, opts ...clientv3.OpOption) error {
			return nil
		}).Reset()
		err := q.SaveState(nil, "testKey")
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("get state", t, func() {
		defer gomonkey.ApplyFunc((*etcd3.EtcdClient).GetResponse, func(_ *etcd3.EtcdClient, ctxInfo etcd3.EtcdCtxInfo,
			etcdKey string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
			return &clientv3.GetResponse{Kvs: []*mvccpb.KeyValue{{Key: nil, Value: nil}}}, nil
		}).Reset()
		_, err := q.GetState("testKey")
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("get state", t, func() {
		err := q.Push("someData", "someKey")
		convey.So(err, convey.ShouldBeNil)
	})
}
