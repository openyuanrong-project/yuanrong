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

// Package etcd3 -
package etcd3

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/client/v3"
)

type mockLease struct {
}

func (m mockLease) Grant(ctx context.Context, ttl int64) (*clientv3.LeaseGrantResponse, error) {
	return &clientv3.LeaseGrantResponse{ID: 1}, nil
}

func (m mockLease) Revoke(ctx context.Context, id clientv3.LeaseID) (*clientv3.LeaseRevokeResponse, error) {
	return nil, nil
}

func (m mockLease) TimeToLive(ctx context.Context, id clientv3.LeaseID, opts ...clientv3.LeaseOption) (*clientv3.LeaseTimeToLiveResponse, error) {
	return nil, nil
}

func (m mockLease) Leases(ctx context.Context) (*clientv3.LeaseLeasesResponse, error) {
	return nil, nil
}

func (m mockLease) KeepAlive(ctx context.Context, id clientv3.LeaseID) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	return nil, nil
}

func (m mockLease) KeepAliveOnce(ctx context.Context, id clientv3.LeaseID) (*clientv3.LeaseKeepAliveResponse, error) {
	return nil, nil
}

func (m mockLease) Close() error {
	panic("implement me")
}

type mockKV struct {
	put    uint32
	get    uint32
	delete uint32
	do     uint32
}

func (fk *mockKV) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	atomic.AddUint32(&fk.put, 1)
	return &clientv3.PutResponse{}, nil
}

func (fk *mockKV) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	atomic.AddUint32(&fk.get, 1)
	return &clientv3.GetResponse{Kvs: []*mvccpb.KeyValue{{}}}, nil
}

func (fk *mockKV) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	atomic.AddUint32(&fk.delete, 1)
	return &clientv3.DeleteResponse{}, nil
}

func (mockKV) Compact(ctx context.Context, rev int64, opts ...clientv3.CompactOption) (*clientv3.CompactResponse, error) {
	return &clientv3.CompactResponse{}, nil
}

func (fk *mockKV) Do(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
	atomic.AddUint32(&fk.do, 1)
	return clientv3.OpResponse{}, nil
}

func (mockKV) Txn(ctx context.Context) clientv3.Txn {
	return nil
}

func TestRegisterInstance_PutInstanceToEtcd(t *testing.T) {
	convey.Convey("test put instance info", t, func() {
		patch := gomonkey.ApplyFunc(GetMetaEtcdClient, func() *EtcdClient {
			return &EtcdClient{Client: &clientv3.Client{KV: &mockKV{}, Lease: &mockLease{}}}
		})

		defer func() {
			patch.Reset()
		}()
		register := &EtcdRegister{
			EtcdClient:  GetMetaEtcdClient(),
			InstanceKey: "/sn/frontend/instances/CLUSTER_ID/HOST_IP/POD_NAME",
			Value:       "active",
		}
		convey.Convey("lease id not exist", func() {
			var keyInput string
			defer gomonkey.ApplyMethod(reflect.TypeOf(&EtcdClient{}), "Put", func(_ *EtcdClient,
				ctxInfo EtcdCtxInfo, key string, value string, opts ...clientv3.OpOption) error {
				keyInput = key
				return nil
			}).Reset()
			defer gomonkey.ApplyFunc(os.Getenv, func(key string) string {
				return key
			}).Reset()
			err := register.putInstanceInfoToEtcd()
			convey.So(err, convey.ShouldBeNil)
			convey.So(keyInput, convey.ShouldEqual, "/sn/frontend/instances/CLUSTER_ID/HOST_IP/POD_NAME")
		})

		convey.Convey("Grant error", func() {
			defer gomonkey.ApplyMethod(reflect.TypeOf(&EtcdClient{}), "Grant", func(_ *EtcdClient,
				ctxInfo EtcdCtxInfo, ttl int64) (clientv3.LeaseID, error) {
				return 111, fmt.Errorf("grant failed")
			}).Reset()
			defer gomonkey.ApplyFunc(os.Getenv, func(key string) string {
				return key
			}).Reset()
			err := register.putInstanceInfoToEtcd()
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("put error", func() {
			var keyInput string
			defer gomonkey.ApplyMethod(reflect.TypeOf(&EtcdClient{}), "Put", func(_ *EtcdClient,
				ctxInfo EtcdCtxInfo, key string, value string, opts ...clientv3.OpOption) error {
				return fmt.Errorf("put failed")
			}).Reset()
			defer gomonkey.ApplyFunc(os.Getenv, func(key string) string {
				return key
			}).Reset()
			err := register.putInstanceInfoToEtcd()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(keyInput, convey.ShouldBeBlank)
			convey.So(err.Error(), convey.ShouldEqual, "put failed")
		})
	})
}

func Test_registerInstance(t *testing.T) {
	etcdClient := &EtcdClient{Client: &clientv3.Client{Lease: &mockLease{}}}
	patch := gomonkey.ApplyFunc(GetMetaEtcdClient, func() *EtcdClient {
		return etcdClient
	})
	patch.ApplyMethod(reflect.TypeOf(&EtcdClient{}), "Put", func(_ *EtcdClient,
		ctxInfo EtcdCtxInfo, key string, value string, opts ...clientv3.OpOption) error {
		return errors.New("put etcd error")
	})
	defer func() {
		patch.Reset()
	}()

	register := &EtcdRegister{
		EtcdClient:  GetMetaEtcdClient(),
		InstanceKey: "/sn/frontend/instances/CLUSTER_ID/HOST_IP/POD_NAME",
		Value:       "active",
	}
	err := register.Register()
	assert.NotNil(t, err)
}

func Test_isKeyExist(t *testing.T) {
	convey.Convey("Test isKeyExist", t, func() {
		patch := gomonkey.ApplyFunc(GetMetaEtcdClient, func() *EtcdClient {
			return &EtcdClient{
				Client: &clientv3.Client{},
			}
		})
		defer patch.Reset()
		register := &EtcdRegister{
			EtcdClient:  GetMetaEtcdClient(),
			InstanceKey: "/sn/frontend/instances/CLUSTER_ID/HOST_IP/POD_NAME",
			Value:       "active",
		}
		convey.Convey("get etcd key return empty", func() {
			patch := gomonkey.ApplyMethod(reflect.TypeOf(&EtcdClient{}), "GetResponse", func(_ *EtcdClient,
				ctxInfo EtcdCtxInfo, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
				return &clientv3.GetResponse{Kvs: []*mvccpb.KeyValue{}}, nil
			})
			defer patch.Reset()
			existed := register.isKeyExist()
			convey.So(existed, convey.ShouldBeFalse)
		})

		convey.Convey("succeed to get etcd key", func() {
			patch := gomonkey.ApplyMethod(reflect.TypeOf(&EtcdClient{}), "GetResponse", func(_ *EtcdClient,
				ctxInfo EtcdCtxInfo, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
				return &clientv3.GetResponse{Kvs: []*mvccpb.KeyValue{{Key: []byte("key")}}}, nil
			})
			defer patch.Reset()
			existed := register.isKeyExist()
			convey.So(existed, convey.ShouldBeTrue)
		})

		convey.Convey("failed to get etcd key", func() {
			patch := gomonkey.ApplyMethod(reflect.TypeOf(&EtcdClient{}), "GetResponse", func(_ *EtcdClient,
				ctxInfo EtcdCtxInfo, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
				return nil, errors.New("failed")
			})
			defer patch.Reset()
			existed := register.isKeyExist()
			convey.So(existed, convey.ShouldBeFalse)
		})
	})
}

func Test_startRefreshLeaseJob(t *testing.T) {
	kv := &mockKV{}
	patches := gomonkey.NewPatches()
	patches.ApplyFunc((*EtcdClient).Put, func(_ *EtcdClient, ctxInfo EtcdCtxInfo, key string, value string, opts ...clientv3.OpOption) error {
		return nil
	})
	patches.ApplyFunc((*clientv3.Client).Ctx, func(_ *clientv3.Client) context.Context { return context.TODO() })
	patches.ApplyFunc((*clientv3.Client).Close, func(_ *clientv3.Client) error { return nil })
	patches.ApplyFunc(clientv3.NewKV, func(c *clientv3.Client) clientv3.KV {
		return kv
	})
	patches.ApplyFunc(GetMetaEtcdClient, func() *EtcdClient {
		return &EtcdClient{Client: &clientv3.Client{KV: kv, Lease: &mockLease{}}}
	})
	defer func() {
		patches.Reset()
	}()
	refreshInterval = 1 * time.Millisecond

	register := &EtcdRegister{
		EtcdClient:  GetMetaEtcdClient(),
		InstanceKey: "/sn/frontend/instances/CLUSTER_ID/HOST_IP/POD_NAME",
		Value:       "active",
	}

	// stop chan is nil, will not trigger refresh
	go register.startRefreshLeaseJob()
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, uint32(0), atomic.LoadUint32(&kv.get))
	assert.Equal(t, uint32(0), atomic.LoadUint32(&kv.put))
	assert.Equal(t, uint32(0), atomic.LoadUint32(&kv.do))

	stopCh := make(chan struct{})
	register.StopCh = stopCh
	go register.startRefreshLeaseJob()
	time.Sleep(100 * time.Millisecond)
	assert.NotEqual(t, uint32(0), atomic.LoadUint32(&kv.get))
	close(stopCh)
	time.Sleep(1 * time.Second)
}
