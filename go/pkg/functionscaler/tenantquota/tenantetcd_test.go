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

// Package tenantquota -
package tenantquota

import (
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

func TestGetTenantInsInfoFromETCD(t *testing.T) {
	info := types.TenantInsInfo{
		OnDemandInsNum: 1,
		ReversedInsNum: 2,
	}
	bytes, _ := json.Marshal(info)

	convey.Convey("Test getTenantInsInfoFromETCD", t, func() {
		convey.Convey("value got from etcd is empty", func() {
			defer gomonkey.ApplyFunc(etcd3.GetValueFromEtcdWithRetry,
				func(key string, etcdClient *etcd3.EtcdClient) ([]byte, error) {
					return nil, nil
				}).Reset()
			tenantInsInfo := getTenantInsInfoFromETCD("test")
			convey.So(tenantInsInfo.ReversedInsNum, convey.ShouldEqual, 0)
			convey.So(tenantInsInfo.OnDemandInsNum, convey.ShouldEqual, 0)
		})

		convey.Convey("value got from etcd is valid", func() {
			defer gomonkey.ApplyFunc(etcd3.GetValueFromEtcdWithRetry,
				func(key string, etcdClient *etcd3.EtcdClient) ([]byte, error) {
					return bytes, nil
				}).Reset()
			tenantInsInfo := getTenantInsInfoFromETCD("test")
			convey.So(tenantInsInfo.ReversedInsNum, convey.ShouldEqual, 2)
			convey.So(tenantInsInfo.OnDemandInsNum, convey.ShouldEqual, 1)
		})

		convey.Convey("value got from etcd err", func() {
			defer gomonkey.ApplyFunc(etcd3.GetValueFromEtcdWithRetry,
				func(key string, etcdClient *etcd3.EtcdClient) ([]byte, error) {
					return nil, errors.New("fail")
				}).Reset()
			tenantInsInfo := getTenantInsInfoFromETCD("test")
			convey.So(tenantInsInfo.ReversedInsNum, convey.ShouldEqual, 0)
			convey.So(tenantInsInfo.OnDemandInsNum, convey.ShouldEqual, 0)
		})
	})
}

func TestUpdateTenantInstance(t *testing.T) {
	convey.Convey("Test updateTenantInstance", t, func() {
		convey.Convey("normal process", func() {
			defer gomonkey.ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
				return &etcd3.EtcdClient{}
			}).Reset()
			defer gomonkey.ApplyFunc((*etcd3.EtcdClient).Put,
				func(_ *etcd3.EtcdClient, ctxInfo etcd3.EtcdCtxInfo, key string, value string,
					opts ...clientv3.OpOption) error {
					return nil
				}).Reset()
			updateTenantInstance("test", 1, 2)
		})
		convey.Convey("put err", func() {
			defer gomonkey.ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
				return &etcd3.EtcdClient{}
			}).Reset()
			defer gomonkey.ApplyFunc((*etcd3.EtcdClient).Put,
				func(_ *etcd3.EtcdClient, ctxInfo etcd3.EtcdCtxInfo, key string, value string,
					opts ...clientv3.OpOption) error {
					return errors.New("fail")
				}).Reset()
			updateTenantInstance("test", 1, 2)
		})
	})
}

func TestAddOrDelTenantInstanceNum(t *testing.T) {
	convey.Convey("Test AddOrDelTenantInstanceNum", t, func() {
		convey.Convey("new session err", func() {
			defer gomonkey.ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
				return &etcd3.EtcdClient{}
			}).Reset()
			defer gomonkey.ApplyFunc(concurrency.NewSession,
				func(client *clientv3.Client, opts ...concurrency.SessionOption) (*concurrency.Session, error) {
					return nil, errors.New("fail")
				}).Reset()
			reachMaxOnDemandInsNum, reachMaxReversedInsNum := IncreaseTenantInstanceNum("test", true)
			convey.So(reachMaxOnDemandInsNum, convey.ShouldEqual, false)
			convey.So(reachMaxReversedInsNum, convey.ShouldEqual, false)
			DecreaseTenantInstance("test", true)

			reachMaxOnDemandInsNum, reachMaxReversedInsNum = IncreaseTenantInstanceNum("test", false)
			convey.So(reachMaxOnDemandInsNum, convey.ShouldEqual, false)
			convey.So(reachMaxReversedInsNum, convey.ShouldEqual, false)
			DecreaseTenantInstance("test", false)
		})
		convey.Convey("new session success", func() {
			info1 := types.TenantInsInfo{
				OnDemandInsNum: 1,
				ReversedInsNum: 2,
			}
			bytes1, _ := json.Marshal(info1)
			defer gomonkey.ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
				return &etcd3.EtcdClient{}
			}).Reset()
			defer gomonkey.ApplyFunc(concurrency.NewSession,
				func(client *clientv3.Client, opts ...concurrency.SessionOption) (*concurrency.Session, error) {
					return nil, nil
				}).Reset()
			defer gomonkey.ApplyFunc(concurrency.NewLocker, func(s *concurrency.Session, pfx string) sync.Locker {
				return &sync.RWMutex{}
			}).Reset()
			defer gomonkey.ApplyFunc((*concurrency.Session).Close, func(_ *concurrency.Session) error {
				return nil
			}).Reset()
			defer gomonkey.ApplyFunc((*etcd3.EtcdClient).Put,
				func(_ *etcd3.EtcdClient, ctxInfo etcd3.EtcdCtxInfo, key string, value string,
					opts ...clientv3.OpOption) error {
					return nil
				}).Reset()
			defer gomonkey.ApplyFunc(etcd3.GetValueFromEtcdWithRetry,
				func(key string, etcdClient *etcd3.EtcdClient) ([]byte, error) {
					return bytes1, nil
				}).Reset()
			reachMaxOnDemandInsNum, reachMaxReversedInsNum := IncreaseTenantInstanceNum("test", true)
			convey.So(reachMaxOnDemandInsNum, convey.ShouldEqual, false)
			convey.So(reachMaxReversedInsNum, convey.ShouldEqual, false)
			DecreaseTenantInstance("test", true)
			reachMaxOnDemandInsNum, reachMaxReversedInsNum = IncreaseTenantInstanceNum("test", false)
			convey.So(reachMaxOnDemandInsNum, convey.ShouldEqual, false)
			convey.So(reachMaxReversedInsNum, convey.ShouldEqual, false)
			DecreaseTenantInstance("test", false)

			info2 := types.TenantInsInfo{
				OnDemandInsNum: 1000,
				ReversedInsNum: 1000,
			}
			bytes2, _ := json.Marshal(info2)
			defer gomonkey.ApplyFunc(etcd3.GetValueFromEtcdWithRetry,
				func(key string, etcdClient *etcd3.EtcdClient) ([]byte, error) {
					return bytes2, nil
				}).Reset()
			reachMaxOnDemandInsNum, reachMaxReversedInsNum = IncreaseTenantInstanceNum("test", false)
			convey.So(reachMaxOnDemandInsNum, convey.ShouldEqual, true)
			convey.So(reachMaxReversedInsNum, convey.ShouldEqual, false)
		})
	})
}
