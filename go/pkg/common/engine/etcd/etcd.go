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
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"yuanrong.org/kernel/pkg/common/engine"
	commonetcd "yuanrong.org/kernel/pkg/common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

// Config is the configuration with etcd
type Config struct {
	Timeout            time.Duration
	TransactionTimeout time.Duration
}

const (
	// DefaultTimeout is the default etcd execution timeout
	DefaultTimeout = 40 * time.Second
	// DefaultTransactionTimeout is the default etcd transaction timeout
	DefaultTransactionTimeout = 40 * time.Second
)

// DefaultConfig is default etcd engine configuration
var DefaultConfig = Config{
	Timeout:            DefaultTimeout,
	TransactionTimeout: DefaultTransactionTimeout,
}

type etcdE struct {
	cli *commonetcd.EtcdClient
	cfg Config
}

// NewEtcdEngine creates a new etcd engine
func NewEtcdEngine(cli *commonetcd.EtcdClient, cfg Config) engine.Engine {
	return &etcdE{cli: cli, cfg: cfg}
}

// Get implements engine.Engine
func (e *etcdE) Get(ctx context.Context, etcdKey string) (string, error) {
	etcdCtxInfo := commonetcd.CreateEtcdCtxInfoWithTimeout(ctx, e.cfg.Timeout)
	values, err := e.cli.GetValues(etcdCtxInfo, etcdKey)
	if err != nil {
		return "", err
	}
	if len(values) == 0 {
		log.GetLogger().Debugf("get key: %s, key not found", etcdKey)
		return "", engine.ErrKeyNotFound
	}

	log.GetLogger().Debugf("get key: %s", etcdKey)
	return values[0], nil
}

// Count implements engine.Engine
func (e *etcdE) Count(ctx context.Context, prefix string) (int64, error) {
	etcdCtxInfo := commonetcd.CreateEtcdCtxInfoWithTimeout(ctx, e.cfg.Timeout)
	resp, err := e.cli.GetResponse(etcdCtxInfo, prefix, clientv3.WithPrefix(), clientv3.WithCountOnly())
	if err != nil {
		return 0, err
	}
	log.GetLogger().Debugf("count prefix: %s, count: %s", prefix, resp.Count)
	return resp.Count, nil
}

func (e *etcdE) firstInRange(ctx context.Context, prefix string, last bool) (string, string, error) {
	etcdCtxInfo := commonetcd.CreateEtcdCtxInfoWithTimeout(ctx, e.cfg.Timeout)
	var opts []clientv3.OpOption
	if last {
		opts = clientv3.WithLastKey()
	} else {
		opts = clientv3.WithFirstKey()
	}

	resp, err := e.cli.GetResponse(etcdCtxInfo, prefix, opts...)
	if err != nil {
		return "", "", err
	}
	if len(resp.Kvs) == 0 {
		if last {
			log.GetLogger().Debugf("last in range prefix: %s, key not found", prefix)
		} else {
			log.GetLogger().Debugf("first in range prefix: %s, key not found", prefix)
		}
		return "", "", engine.ErrKeyNotFound
	}

	if last {
		log.GetLogger().Debugf("last in range prefix: %s, key: %s", prefix, string(resp.Kvs[0].Key))
	} else {
		log.GetLogger().Debugf("first in range prefix: %s, key: %s", prefix, string(resp.Kvs[0].Key))
	}
	return e.cli.DetachAZPrefix(string(resp.Kvs[0].Key)), string(resp.Kvs[0].Value), nil
}

// FirstInRange implements engine.Engine
func (e *etcdE) FirstInRange(ctx context.Context, prefix string) (string, string, error) {
	return e.firstInRange(ctx, prefix, false)
}

// LastInRange implements engine.Engine
func (e *etcdE) LastInRange(ctx context.Context, prefix string) (string, string, error) {
	return e.firstInRange(ctx, prefix, true)
}

// PrepareStream implements engine.Engine
func (e *etcdE) PrepareStream(
	ctx context.Context, prefix string, decode engine.DecodeHandleFunc, by engine.SortBy,
) engine.PrepareStmt {
	sortOp, err := e.genSortOpOption(by)
	if err != nil {
		return &etcdStmt{
			fn: func() ([]interface{}, error) {
				return nil, err
			},
		}
	}

	fn := func() ([]interface{}, error) {
		etcdCtxInfo := commonetcd.CreateEtcdCtxInfoWithTimeout(ctx, e.cfg.Timeout)
		resp, err := e.cli.GetResponse(etcdCtxInfo, prefix, clientv3.WithPrefix(), sortOp)
		if err != nil {
			return nil, err
		}

		log.GetLogger().Debugf("stream prefix: %s, resp num: %v", prefix, len(resp.Kvs))

		var res []interface{}
		for _, kv := range resp.Kvs {
			key := e.cli.DetachAZPrefix(string(kv.Key))
			i, err := decode(key, string(kv.Value))
			if err != nil {
				return nil, err
			}
			res = append(res, i)
		}
		return res, nil
	}

	return &etcdStmt{
		fn: fn,
	}
}

func (e *etcdE) genSortOpOption(by engine.SortBy) (clientv3.OpOption, error) {
	var (
		order  clientv3.SortOrder
		target clientv3.SortTarget
	)
	switch by.Order {
	case engine.Ascend:
		order = clientv3.SortAscend
	case engine.Descend:
		order = clientv3.SortDescend
	default:
		return nil, fmt.Errorf("invalid sort order: %v", by.Order)
	}
	switch by.Target {
	case engine.SortName:
		target = clientv3.SortByKey
	case engine.SortCreate:
		target = clientv3.SortByCreateRevision
	case engine.SortModify:
		target = clientv3.SortByModRevision
	default:
		return nil, fmt.Errorf("invalid sort target: %v", by.Target)
	}
	sortOp := clientv3.WithSort(target, order)
	return sortOp, nil
}

// Put implements engine.Engine
func (e *etcdE) Put(ctx context.Context, etcdKey string, value string) error {
	etcdCtxInfo := commonetcd.CreateEtcdCtxInfoWithTimeout(ctx, e.cfg.Timeout)
	err := e.cli.Put(etcdCtxInfo, etcdKey, value)
	if err != nil {
		log.GetLogger().Debugf("put key: %s", etcdKey)
	}
	return err
}

// Delete implements engine.Engine
func (e *etcdE) Delete(ctx context.Context, etcdKey string) error {
	etcdCtxInfo := commonetcd.CreateEtcdCtxInfoWithTimeout(ctx, e.cfg.Timeout)
	err := e.cli.Delete(etcdCtxInfo, etcdKey)
	if err != nil {
		log.GetLogger().Debugf("delete etcd key: %s", etcdKey)
	}
	return err
}

// BeginTx implements engine.Engine
func (e *etcdE) BeginTx(ctx context.Context) engine.Transaction {
	return newTransaction(ctx, e.cli, e.cfg.TransactionTimeout)
}

// Close implements engine.Engine
func (e *etcdE) Close() error {
	return e.cli.Client.Close()
}
