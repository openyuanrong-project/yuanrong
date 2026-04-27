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
	"context"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"go.etcd.io/etcd/client/v3"

	"yuanrong.org/kernel/pkg/common/faas_common/utils"
)

// TestEtcdClient_Grant -
func TestEtcdClient_Grant(t *testing.T) {
	convey.Convey("test: grant", t, func() {
		client := &EtcdClient{
			Client: &clientv3.Client{
				Lease: utils.FakeEtcdLease{},
			},
		}
		id, err := client.Grant(CreateEtcdCtxInfoWithTimeout(context.Background(), 100*time.Millisecond), 10)
		convey.So(id, convey.ShouldEqual, 1)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestEtcdClient_KeepAliveOnce -
func TestEtcdClient_KeepAliveOnce(t *testing.T) {
	convey.Convey("test: keepAliveOnce", t, func() {
		client := &EtcdClient{
			Client: &clientv3.Client{
				Lease: utils.FakeEtcdLease{},
			},
		}
		err := client.KeepAliveOnce(CreateEtcdCtxInfoWithTimeout(context.Background(), 100*time.Millisecond), 1)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestEtcdClient_KeepAliveOnce -
func TestEtcdClient_Revoke(t *testing.T) {
	convey.Convey("test: revoke", t, func() {
		client := &EtcdClient{
			Client: &clientv3.Client{
				Lease: utils.FakeEtcdLease{},
			},
		}
		err := client.Revoke(CreateEtcdCtxInfoWithTimeout(context.Background(), 100*time.Millisecond), 1)
		convey.So(err, convey.ShouldBeNil)
	})
}
