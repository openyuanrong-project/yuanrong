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
	"errors"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type fakeTxn struct {
	response *clientv3.TxnResponse
	err      error
}

func (t *fakeTxn) If(cs ...clientv3.Cmp) clientv3.Txn {
	return t
}

func (t *fakeTxn) Then(ops ...clientv3.Op) clientv3.Txn {
	return t
}

func (t *fakeTxn) Else(ops ...clientv3.Op) clientv3.Txn {
	return t
}

func (t *fakeTxn) Commit() (*clientv3.TxnResponse, error) {
	return t.response, t.err
}

type fakeKv struct {
	txn *fakeTxn
}

func (k *fakeKv) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	return nil, nil
}

func (k *fakeKv) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	return nil, nil
}

func (k *fakeKv) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	return nil, nil
}

func (k *fakeKv) Compact(ctx context.Context, rev int64, opts ...clientv3.CompactOption) (*clientv3.CompactResponse, error) {
	return nil, nil
}

func (k *fakeKv) Do(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
	return clientv3.OpResponse{}, nil
}

func (k *fakeKv) Txn(ctx context.Context) clientv3.Txn {
	return k.txn
}

func buildTxnResponse(success bool, revision int64, kvs1, kvs2 []*mvccpb.KeyValue) *clientv3.TxnResponse {
	responses := []*etcdserverpb.ResponseOp{}
	if kvs1 != nil {
		responses = append(responses, &etcdserverpb.ResponseOp{
			Response: &etcdserverpb.ResponseOp_ResponseRange{
				ResponseRange: &etcdserverpb.RangeResponse{
					Kvs: kvs1,
				},
			},
		})
	}
	if kvs2 != nil {
		responses = append(responses, &etcdserverpb.ResponseOp{
			Response: &etcdserverpb.ResponseOp_ResponseRange{
				ResponseRange: &etcdserverpb.RangeResponse{
					Kvs: kvs2,
				},
			},
		})
	}
	return &clientv3.TxnResponse{
		Succeeded: success,
		Header:    &etcdserverpb.ResponseHeader{Revision: revision},
		Responses: responses,
	}
}

func TestTryLock(t *testing.T) {
	convey.Convey("test TryLock", t, func() {
		ft := &fakeTxn{}
		stopCh := make(chan struct{})
		lock := &EtcdLocker{EtcdClient: &EtcdClient{Client: &clientv3.Client{KV: &fakeKv{txn: ft}}}, LeaseTTL: 10,
			StopCh: stopCh}
		patches := []*gomonkey.Patches{
			gomonkey.ApplyFunc((*EtcdClient).Grant, func(_ *EtcdClient, ctxInfo EtcdCtxInfo, ttl int64) (clientv3.LeaseID,
				error) {
				return 123, nil
			}),
			gomonkey.ApplyFunc((*EtcdClient).Revoke, func(_ *EtcdClient, ctxInfo EtcdCtxInfo, leaseID clientv3.LeaseID) error {
				return nil
			}),
			gomonkey.ApplyFunc((*EtcdClient).Delete, func(_ *EtcdClient, ctxInfo EtcdCtxInfo, etcdKey string,
				opts ...clientv3.OpOption) error {
				return nil
			}),
		}
		defer func() {
			close(stopCh)
			time.Sleep(100 * time.Millisecond)
			for _, p := range patches {
				p.Reset()
			}
		}()
		convey.Convey("got and locked", func() {
			patch1 := gomonkey.ApplyFunc((*EtcdClient).Get, func(_ *EtcdClient, ctxInfo EtcdCtxInfo, key string,
				opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
				return &clientv3.GetResponse{Kvs: []*mvccpb.KeyValue{{Key: []byte("/test/key1")}}}, nil
			})
			defer patch1.Reset()
			ft.response = buildTxnResponse(true, 123, []*mvccpb.KeyValue{}, []*mvccpb.KeyValue{
				{
					Key:            []byte("/test/key1"),
					CreateRevision: 100,
				},
			})
			ft.err = nil
			err := lock.TryLockWithPrefix("/test", func(k, v []byte) bool { return false })
			convey.So(err, convey.ShouldBeNil)
			key := lock.GetLockedKey()
			convey.So(key, convey.ShouldEqual, "/test/key1")
		})
		convey.Convey("got error", func() {
			patch1 := gomonkey.ApplyFunc((*EtcdClient).Get, func(_ *EtcdClient, ctxInfo EtcdCtxInfo, key string,
				opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
				return nil, errors.New("some error")
			})
			defer patch1.Reset()
			err := lock.TryLockWithPrefix("/test", func(k, v []byte) bool { return false })
			convey.So(err.Error(), convey.ShouldEqual, "some error")
		})
		convey.Convey("lock key lost", func() {
			patch1 := gomonkey.ApplyFunc((*EtcdClient).Get, func(_ *EtcdClient, ctxInfo EtcdCtxInfo, key string,
				opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
				return &clientv3.GetResponse{Kvs: []*mvccpb.KeyValue{{Key: []byte("/test/key1")}}}, nil
			})
			defer patch1.Reset()
			ft.response = buildTxnResponse(true, 123, []*mvccpb.KeyValue{}, nil)
			ft.err = nil
			err := lock.TryLockWithPrefix("/test", func(k, v []byte) bool { return false })
			convey.So(err, convey.ShouldNotBeNil)
			ft.response = buildTxnResponse(true, 123, []*mvccpb.KeyValue{}, []*mvccpb.KeyValue{
				{
					Key:            []byte("/test/key1/123"),
					CreateRevision: 100,
				},
			})
			ft.err = nil
			err = lock.TryLockWithPrefix("/test", func(k, v []byte) bool { return false })
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("lock key locked by others", func() {
			patch1 := gomonkey.ApplyFunc((*EtcdClient).Get, func(_ *EtcdClient, ctxInfo EtcdCtxInfo, key string,
				opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
				return &clientv3.GetResponse{Kvs: []*mvccpb.KeyValue{{Key: []byte("/test/key1")}}}, nil
			})
			defer patch1.Reset()
			ft.response = buildTxnResponse(true, 123, []*mvccpb.KeyValue{}, []*mvccpb.KeyValue{
				{
					Key:            []byte("/test/key1"),
					CreateRevision: 100,
				},
				{
					Key:            []byte("/test/key1/xxx"),
					CreateRevision: 101,
				},
			})
			ft.err = nil
			err := lock.TryLockWithPrefix("/test", func(k, v []byte) bool { return false })
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("lock callback", func() {
			patch1 := gomonkey.ApplyFunc((*EtcdClient).Get, func(_ *EtcdClient, ctxInfo EtcdCtxInfo, key string,
				opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
				return &clientv3.GetResponse{Kvs: []*mvccpb.KeyValue{{Key: []byte("/test/key1")}}}, nil
			})
			defer patch1.Reset()
			ft.response = buildTxnResponse(true, 123, []*mvccpb.KeyValue{}, []*mvccpb.KeyValue{
				{
					Key:            []byte("/test/key1"),
					CreateRevision: 100,
				},
			})
			ft.err = nil
			lock.LockCallback = func(l *EtcdLocker) error { return errors.New("some error") }
			err := lock.TryLockWithPrefix("/test", func(k, v []byte) bool { return false })
			convey.So(err.Error(), convey.ShouldEqual, "some error")
		})
	})
}

func TestUnlock(t *testing.T) {
	stopCh := make(chan struct{})
	lock := &EtcdLocker{EtcdClient: &EtcdClient{}, StopCh: stopCh}
	convey.Convey("test Unlock", t, func() {
		convey.Convey("unlock ok", func() {
			patch := gomonkey.ApplyFunc((*EtcdClient).Delete, func(_ *EtcdClient, ctxInfo EtcdCtxInfo, etcdKey string,
				opts ...clientv3.OpOption) error {
				return nil
			})
			defer patch.Reset()
			err := lock.Unlock()
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("unlock error", func() {
			patch := gomonkey.ApplyFunc((*EtcdClient).Delete, func(_ *EtcdClient, ctxInfo EtcdCtxInfo, etcdKey string,
				opts ...clientv3.OpOption) error {
				return errors.New("some error")
			})
			defer patch.Reset()
			err := lock.Unlock()
			convey.So(err.Error(), convey.ShouldEqual, "some error")
		})
		convey.Convey("unlock callback", func() {
			patch := gomonkey.ApplyFunc((*EtcdClient).Delete, func(_ *EtcdClient, ctxInfo EtcdCtxInfo, etcdKey string,
				opts ...clientv3.OpOption) error {
				return nil
			})
			defer patch.Reset()
			lock.UnlockCallback = func(l *EtcdLocker) error { return errors.New("some error") }
			err := lock.Unlock()
			convey.So(err.Error(), convey.ShouldEqual, "some error")
		})
	})
}

func TestLockKeeperLoop(t *testing.T) {
	convey.Convey("test lockKeeperLoop", t, func() {
		stopCh := make(chan struct{})
		lock := &EtcdLocker{EtcdClient: &EtcdClient{}, LeaseTTL: 0, StopCh: stopCh}
		getResp := &clientv3.GetResponse{}
		getErr := error(nil)
		patches := []*gomonkey.Patches{
			gomonkey.ApplyFunc((*EtcdClient).Get, func(_ *EtcdClient, ctxInfo EtcdCtxInfo, key string,
				opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
				return getResp, getErr
			}),
			gomonkey.ApplyFunc((*EtcdLocker).Unlock, func(_ *EtcdLocker) error {
				return nil
			}),
			gomonkey.ApplyGlobalVar(&lockFailCountLimit, 0),
		}
		defer func() {
			for _, p := range patches {
				p.Reset()
			}
		}()
		convey.Convey("ticker case 1", func() {
			ticker := time.NewTicker(100 * time.Millisecond)
			patch1 := gomonkey.ApplyFunc(time.NewTicker, func(d time.Duration) *time.Ticker {
				return ticker
			})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyFunc((*EtcdLocker).tryLock, func(_ *EtcdLocker, key string) error {
				return ErrNoKeyCanBeFound
			})
			defer patch2.Reset()
			getErr = errors.New("get key error")
			lock.leaseID = 123
			called := false
			lock.FailCallback = func() { called = true }
			lock.lockKeeperLoop()
			convey.So(called, convey.ShouldBeTrue)
		})
		convey.Convey("ticker case 2", func() {
			ticker := time.NewTicker(100 * time.Millisecond)
			patch1 := gomonkey.ApplyFunc(time.NewTicker, func(d time.Duration) *time.Ticker {
				return ticker
			})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyFunc((*EtcdClient).KeepAliveOnce, func(_ *EtcdClient, ctxInfo EtcdCtxInfo,
				leaseID clientv3.LeaseID) error {
				return errors.New("context deadline exceeded")
			})
			defer patch2.Reset()
			patch3 := gomonkey.ApplyFunc((*EtcdLocker).tryLock, func(_ *EtcdLocker, key string) error {
				return ErrNoKeyCanBeFound
			})
			defer patch3.Reset()
			getErr = nil
			getResp.Kvs = []*mvccpb.KeyValue{{}}
			lock.leaseID = 123
			called := false
			lock.FailCallback = func() { called = true }
			lock.lockKeeperLoop()
			convey.So(called, convey.ShouldBeTrue)
		})
		convey.Convey("other case", func() {
			called := false
			patch1 := gomonkey.ApplyFunc((*EtcdLocker).Unlock, func(_ *EtcdLocker) error {
				called = true
				return nil
			})
			defer patch1.Reset()
			ticker := time.NewTicker(100 * time.Millisecond)
			patch2 := gomonkey.ApplyFunc(time.NewTicker, func(d time.Duration) *time.Ticker {
				return ticker
			})
			defer patch2.Reset()
			unlockCh := make(chan struct{})
			lock.unlockCh = unlockCh
			close(unlockCh)
			lock.lockKeeperLoop()
			unlockCh = make(chan struct{})
			lock.unlockCh = unlockCh
			close(stopCh)
			lock.lockKeeperLoop()
			convey.So(called, convey.ShouldBeTrue)
		})
	})
}
