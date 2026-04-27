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
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"yuanrong.org/kernel/pkg/common/constants"
	"yuanrong.org/kernel/pkg/common/engine"
	"yuanrong.org/kernel/pkg/common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

type reads struct {
	resp       *clientv3.GetResponse
	modRev     int64
	withPrefix bool
}

const (
	writeOp = iota
	delOp
)

type writes struct {
	value      string
	op         int
	withPrefix bool
}

// Transaction utilities etcd v3 transaction to perform transaction with separated expressions
type Transaction struct {
	etcdClient *etcd3.EtcdClient

	rds map[string]*reads
	wrs map[string]*writes

	ctx    context.Context
	cancel func()
}

// newTransaction creates a new transaction object to records all the method that will be performed in it.
func newTransaction(ctx context.Context, client *etcd3.EtcdClient, timeout time.Duration) *Transaction {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	t := Transaction{etcdClient: client, ctx: ctx, cancel: cancel}
	t.rds = make(map[string]*reads, constants.DefaultMapSize)
	t.wrs = make(map[string]*writes, constants.DefaultMapSize)
	return &t
}

// Put caches a key-value pair, it will be replaced if the same key has been called within the transaction.
func (t *Transaction) Put(key string, value string) {
	log.GetLogger().Debugf("transaction put key: %s, value: %s", key, value)
	t.wrs[key] = &writes{value, writeOp, false}
}

func getRespMaxModRev(resp *clientv3.GetResponse) int64 {
	var rev int64 = 0
	for _, kv := range resp.Kvs {
		if kv.ModRevision > rev {
			rev = kv.ModRevision
		}
	}
	return rev
}

// Get returns value of the key from etcd, and cached it to perform 'If' statement in the transaction.
// The method will return cached value if the same key has been called within the transaction.
func (t *Transaction) Get(key string) (string, error) {
	if v, ok := t.wrs[key]; ok {
		if v.op == writeOp {
			log.GetLogger().Debugf("cached: transaction get key: %s, value: %s", key, v.value)
			return v.value, nil
		}
		log.GetLogger().Debugf("cached: transaction get key: %s, value: %s, key not found", key)
		return "", engine.ErrKeyNotFound
	}
	if v, ok := t.rds[key]; ok {
		if len(v.resp.Kvs) == 1 {
			return string(v.resp.Kvs[0].Value), nil
		}
		return "", engine.ErrKeyNotFound
	}

	etcdCtxInfo := etcd3.CreateEtcdCtxInfo(t.ctx)
	resp, err := t.etcdClient.GetResponse(etcdCtxInfo, key)
	if err != nil {
		log.GetLogger().Errorf("failed to get from etcd, error: %s", err.Error())
		return "", err
	}
	t.rds[key] = &reads{resp, getRespMaxModRev(resp), false}
	if len(resp.Kvs) == 1 {
		log.GetLogger().Debugf("transaction get key: %s, value: %s, modRevision: %d",
			key, string(resp.Kvs[0].Value), resp.Kvs[0].ModRevision)
		return string(resp.Kvs[0].Value), nil
	}
	log.GetLogger().Debugf("transaction get key: %s, value: %s, key not found", key)
	return "", engine.ErrKeyNotFound
}

func (t *Transaction) getRespKVs(resp *clientv3.GetResponse) (keys, values []string) {
	for _, kv := range resp.Kvs {
		key := t.etcdClient.DetachAZPrefix(string(kv.Key))
		keys = append(keys, key)
		values = append(values, string(kv.Value))
		log.GetLogger().Debugf("transaction get prefix key: %s, modRevision: %d", key, kv.ModRevision)
	}
	return
}

// GetPrefix returns values of the key as a prefix from etcd, and cached them to perform 'If' statement in the
// transaction. The method will return cached values if the same key has been called within the transaction.
func (t *Transaction) GetPrefix(key string) (keys, values []string, err error) {
	if v, ok := t.rds[key]; ok {
		keys, values = t.getRespKVs(v.resp)
		log.GetLogger().Debugf("cached: transaction get prefix: %s, resp num: %v", key, len(keys))
		return keys, values, nil
	}

	etcdCtxInfo := etcd3.CreateEtcdCtxInfo(t.ctx)
	resp, err := t.etcdClient.GetResponse(etcdCtxInfo, key, clientv3.WithPrefix())
	if err != nil {
		log.GetLogger().Errorf("failed to get with prefix from etcd, error: %s", err.Error())
		return nil, nil, err
	}
	t.rds[key] = &reads{resp, getRespMaxModRev(resp), true}
	keys, values = t.getRespKVs(resp)
	log.GetLogger().Debugf("transaction get prefix: %s, resp num: %v", key, len(keys))
	return keys, values, nil
}

// Del caches a key, it will be replaced if the same key has been called within the transaction.
func (t *Transaction) Del(key string) {
	log.GetLogger().Debugf("transaction delete key: %s", key)
	t.wrs[key] = &writes{"", delOp, false}
}

// DelPrefix caches a key, it will be replaced if the same key has been called within the transaction.
func (t *Transaction) DelPrefix(key string) {
	log.GetLogger().Debugf("transaction delete prefix: %s", key)
	t.wrs[key] = &writes{"", delOp, true}
}

func (t *Transaction) genCmp() []clientv3.Cmp {
	cs := make([]clientv3.Cmp, 0, len(t.rds))
	for k, v := range t.rds {
		k = t.etcdClient.AttachAZPrefix(k)
		result := "="
		rev := v.modRev
		if v.withPrefix && rev != 0 {
			result = "<"
			rev++
		}

		c := clientv3.Compare(clientv3.ModRevision(k), result, rev)
		if v.withPrefix {
			c = c.WithPrefix()
		}
		cs = append(cs, c)
	}
	return cs
}

func (t *Transaction) genOp() []clientv3.Op {
	ops := make([]clientv3.Op, 0, len(t.wrs))
	for k, v := range t.wrs {
		k = t.etcdClient.AttachAZPrefix(k)
		var op clientv3.Op
		if v.op == writeOp {
			op = clientv3.OpPut(k, v.value)
		} else {
			if v.withPrefix {
				op = clientv3.OpDelete(k, clientv3.WithPrefix())
			} else {
				op = clientv3.OpDelete(k)
			}
		}
		ops = append(ops, op)
	}
	return ops
}

// Commit do the 'If' statement with all Get and GetPrefix methods,
// and do the 'Then' statement with all Put, Del, DelPrefix methods.
func (t *Transaction) Commit() error {
	resp, err := t.etcdClient.Client.KV.Txn(t.ctx).If(t.genCmp()...).Then(t.genOp()...).Commit()
	if err != nil {
		log.GetLogger().Errorf("failed to commit transaction, error: %s", err.Error())
		return err
	}
	if !resp.Succeeded {
		log.GetLogger().Errorf("transaction commit not succeeded")
		t.printError()
		return engine.ErrTransaction
	}
	log.GetLogger().Debugf("transaction get revision, revision: %d", resp.Header.Revision)
	return nil
}

// Cancel undoes the commit.
func (t *Transaction) Cancel() {
	t.cancel()
}

func (t *Transaction) printError() {
	for k, v := range t.rds {
		rev := v.modRev
		if v.withPrefix {
			if t.rereadPrefix(k, rev) != nil {
				continue
			}
		} else {
			if t.rereadKey(k, rev) != nil {
				continue
			}
		}
	}
}

func (t *Transaction) rereadPrefix(k string, revision int64) error {
	etcdCtxInfo := etcd3.CreateEtcdCtxInfo(t.ctx)
	resp, err := t.etcdClient.GetResponse(etcdCtxInfo, k, clientv3.WithPrefix())
	if err != nil {
		log.GetLogger().Errorf("failed to reread data with prefix from etcd, error: %s", err.Error())
		return err
	}
	t.printPrefixValue(resp, k, revision)
	return nil
}

func (t *Transaction) printPrefixValue(resp *clientv3.GetResponse, k string, revision int64) {
	for index, kv := range resp.Kvs {
		if kv.ModRevision > revision {
			log.GetLogger().Errorf("reread etcd data by prefix, Key : %s", string(resp.Kvs[index].Key))
			log.GetLogger().Errorf("reread etcd data by prefix, Value: %s", string(resp.Kvs[index].Value))
			log.GetLogger().Errorf("reread etcd data by prefix, CreateRevision: %d, ModRevision: %d, Version: %d",
				resp.Kvs[index].CreateRevision, resp.Kvs[index].ModRevision, resp.Kvs[index].Version)
			t.printCachedPrefixValue(k, string(kv.Key))
		}
	}
}

func (t *Transaction) printCachedPrefixValue(prefix string, key string) {
	if v, ok := t.rds[prefix]; ok {
		for i, value := range v.resp.Kvs {
			if string(value.Key) == string(key) {
				log.GetLogger().Errorf("get cached data by prefix, Key : %s", string(v.resp.Kvs[i].Key))
				log.GetLogger().Errorf("get cached data by prefix, Value: %s", string(v.resp.Kvs[i].Value))
				log.GetLogger().Errorf("get cached data by prefix, CreateRevision: %d, ModRevision: %d, Version: %d",
					v.resp.Kvs[i].CreateRevision, v.resp.Kvs[i].ModRevision, v.resp.Kvs[i].Version)
			}
		}
	} else {
		log.GetLogger().Errorf("invalid prefix, prefix : %s", prefix)
	}
}

func (t *Transaction) rereadKey(k string, revision int64) error {
	etcdCtxInfo := etcd3.CreateEtcdCtxInfo(t.ctx)
	resp, err := t.etcdClient.GetResponse(etcdCtxInfo, k)
	if err != nil {
		log.GetLogger().Errorf("invalid key, k: %s, error: %s", k, err.Error())
		return err
	}
	t.printKeyValue(resp, k, revision)
	return nil
}

func (t *Transaction) printKeyValue(resp *clientv3.GetResponse, k string, revision int64) {
	for index, kv := range resp.Kvs {
		if kv.ModRevision != revision {
			log.GetLogger().Errorf("reread etcd data, Key : %s", string(resp.Kvs[index].Key))
			log.GetLogger().Errorf("reread etcd data, Value: %s", string(resp.Kvs[index].Value))
			log.GetLogger().Errorf("reread etcd data, CreateRevision: %d, ModRevision: %d, Version: %d",
				resp.Kvs[index].CreateRevision, resp.Kvs[index].ModRevision, resp.Kvs[index].Version)
			t.printCachedKeyValue(k)
		}
	}
}

func (t *Transaction) printCachedKeyValue(k string) {
	if v, ok := t.rds[k]; ok {
		if len(v.resp.Kvs) == 1 {
			log.GetLogger().Errorf("get cached data, Key : %s", string(v.resp.Kvs[0].Key))
			log.GetLogger().Errorf("get cached data, Value: %s", string(v.resp.Kvs[0].Value))
			log.GetLogger().Errorf("get cached data, CreateRevision: %d, ModRevision: %d, Version: %d",
				v.resp.Kvs[0].CreateRevision, v.resp.Kvs[0].ModRevision, v.resp.Kvs[0].Version)
		}
	} else {
		log.GetLogger().Errorf("invalid key, k: %s", k)
	}
}
