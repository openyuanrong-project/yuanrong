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
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/net/context"

	"yuanrong.org/kernel/pkg/common/crypto"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

const (
	// We have set a buffer in order to reduce times of context switches.
	incomingBufSize = 2000
	outgoingBufSize = 2000
)

// EtcdClientInterface is the interface of ETCD client
type EtcdClientInterface interface {
	GetResponse(ctxInfo EtcdCtxInfo, etcdKey string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
	Put(ctxInfo EtcdCtxInfo, etcdKey string, value string, opts ...clientv3.OpOption) error
	Delete(ctxInfo EtcdCtxInfo, etcdKey string, opts ...clientv3.OpOption) error
}

// EtcdClient etcd client struct
type EtcdClient struct {
	Client   *clientv3.Client
	AZPrefix string
}

// EtcdWatchChan implements watch.Interface.
type EtcdWatchChan struct {
	incomingEventChan chan *Event
	ResultChan        chan *Event
	errChan           chan error
	watcher           *EtcdClient
	key               string
	initialRev        int64
	recursive         bool
	ctx               context.Context
	cancel            context.CancelFunc
}

// EtcdCtxInfo etcd context info
type EtcdCtxInfo struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}

const (
	// keepaliveTime is the time after which client pings the server to see if
	// transport is alive.
	keepaliveTime = 30 * time.Second

	// keepaliveTimeout is the time that the client waits for a response for the
	// keep-alive attempt.
	keepaliveTimeout = 10 * time.Second

	// dialTimeout is the timeout for establishing a connection.
	// 20 seconds as times should be set shorter than that will cause TLS connections to fail
	dialTimeout = 20 * time.Second

	// tokenRenewTTL the default TTL of etcd simple token is 300s, so the renew TTL should be smaller than 300s
	tokenRenewTTL = 30 * time.Second

	// renewKey etcd server will renew simple token TTL if token is not expired.
	// so use an random key path for querying in order to renew token.
	renewKey = "/keyforrenew"

	// DurationContextTimeout etcd request timeout, default context duration timeout
	DurationContextTimeout = 5 * time.Second
)

// AttachAZPrefix -
func (w *EtcdClient) AttachAZPrefix(key string) string {
	if len(w.AZPrefix) != 0 {
		return fmt.Sprintf("/%s%s", w.AZPrefix, key)
	}
	return key
}

// DetachAZPrefix -
func (w *EtcdClient) DetachAZPrefix(key string) string {
	if len(w.AZPrefix) != 0 {
		return strings.TrimPrefix(key, fmt.Sprintf("/%s", w.AZPrefix))
	}
	return key
}

// GetResponse get etcd value and return pointer of GetResponse struct
func (w *EtcdClient) GetResponse(ctxInfo EtcdCtxInfo, etcdKey string,
	opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	etcdKey = w.AttachAZPrefix(etcdKey)
	ctx, cancel := ctxInfo.Ctx, ctxInfo.Cancel
	defer cancel()

	kv := clientv3.NewKV(w.Client)
	leaderCtx := clientv3.WithRequireLeader(ctx)
	getResp, err := kv.Get(leaderCtx, etcdKey, opts...)

	return getResp, err
}

// Put put context key and value
func (w *EtcdClient) Put(ctxInfo EtcdCtxInfo, etcdKey string, value string, opts ...clientv3.OpOption) error {
	etcdKey = w.AttachAZPrefix(etcdKey)
	ctx, cancel := ctxInfo.Ctx, ctxInfo.Cancel
	defer cancel()

	kv := clientv3.NewKV(w.Client)
	leaderCtx := clientv3.WithRequireLeader(ctx)
	_, err := kv.Put(leaderCtx, etcdKey, value, opts...)
	return err
}

// Delete delete key
func (w *EtcdClient) Delete(ctxInfo EtcdCtxInfo, etcdKey string, opts ...clientv3.OpOption) error {
	etcdKey = w.AttachAZPrefix(etcdKey)
	ctx, cancel := ctxInfo.Ctx, ctxInfo.Cancel
	defer cancel()

	kv := clientv3.NewKV(w.Client)
	leaderCtx := clientv3.WithRequireLeader(ctx)
	_, err := kv.Delete(leaderCtx, etcdKey, opts...)
	return err
}

// TxnPut transaction put operation with if key existed put failed
func (w *EtcdClient) TxnPut(ctxInfo EtcdCtxInfo, etcdKey string, value string) error {
	etcdKey = w.AttachAZPrefix(etcdKey)
	ctx, cancel := ctxInfo.Ctx, ctxInfo.Cancel
	defer cancel()
	kv := clientv3.NewKV(w.Client)
	leaderCtx := clientv3.WithRequireLeader(ctx)
	txnRsp, err := kv.Txn(leaderCtx).
		If(clientv3.Compare(clientv3.CreateRevision(etcdKey), "=", 0)).
		Then(clientv3.OpPut(etcdKey, value)).
		Else(clientv3.OpGet(etcdKey)).Commit()
	if err != nil {
		return err
	}
	if !txnRsp.Succeeded {
		log.GetLogger().Warnf("the key has already exist: %s", etcdKey)
		return fmt.Errorf("duplicated key")
	}
	return nil
}

// GetValues return list of object for value
func (w *EtcdClient) GetValues(ctxInfo EtcdCtxInfo, etcdKey string, opts ...clientv3.OpOption) ([]string, error) {
	etcdKey = w.AttachAZPrefix(etcdKey)
	ctx, cancel := ctxInfo.Ctx, ctxInfo.Cancel
	defer cancel()

	kv := clientv3.NewKV(w.Client)
	leaderCtx := clientv3.WithRequireLeader(ctx)
	response, err := kv.Get(leaderCtx, etcdKey, opts...)

	if err != nil {
		return nil, err
	}
	values := make([]string, len(response.Kvs), len(response.Kvs))

	for index, v := range response.Kvs {
		values[index] = string(v.Value)
	}
	return values, err
}

// CreateEtcdCtxInfo return context with cancle function
func CreateEtcdCtxInfo(ctx context.Context) EtcdCtxInfo {
	ctx, cancel := context.WithCancel(ctx)
	leaderCtx := clientv3.WithRequireLeader(ctx)
	return EtcdCtxInfo{leaderCtx, cancel}
}

// CreateEtcdCtxInfoWithTimeout create a context with timeout, default timeout is DurationContextTimeout
func CreateEtcdCtxInfoWithTimeout(ctx context.Context, duration time.Duration) EtcdCtxInfo {
	ctx, cancel := context.WithTimeout(ctx, duration)
	leaderCtx := clientv3.WithRequireLeader(ctx)
	return EtcdCtxInfo{leaderCtx, cancel}
}

// DecryptEtcdPassword decrypts the password of etcd
func DecryptEtcdPassword(plainPwd, workKey []byte) ([]byte, error) {
	var (
		etcdPwd []byte
		err     error
	)
	etcdPwd, err = crypto.DecryptByte(plainPwd, workKey)
	if err != nil {
		return nil, err
	}
	return etcdPwd, nil
}
