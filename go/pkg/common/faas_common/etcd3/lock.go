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
	"fmt"
	"sync/atomic"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
)

const (
	defaultRequestTimeout = 30 * time.Second
	refreshAheadTime      = 1 * time.Second
	lockedKeyHoldIndex    = 1
)

var (
	// ErrEtcdResponseInvalid -
	ErrEtcdResponseInvalid = errors.New("etcd response is invalid")
	// ErrNoKeyCanBeFound -
	ErrNoKeyCanBeFound = errors.New("no etcd key can be found")
	// ErrNoKeyCanBeLocked -
	ErrNoKeyCanBeLocked = errors.New("no etcd key can be locked")
	lockFailCountLimit  = 10
)

// EtcdLocker -
type EtcdLocker struct {
	EtcdClient     *EtcdClient
	acquiredLock   *concurrency.Mutex
	LockedKey      string
	holderKey      string
	LeaseTTL       int
	leaseID        clientv3.LeaseID
	locked         atomic.Uint32
	LockCallback   func(locker *EtcdLocker) error
	UnlockCallback func(locker *EtcdLocker) error
	FailCallback   func()
	unlockCh       chan struct{}
	StopCh         <-chan struct{}
}

// GetLockedKey -
func (l *EtcdLocker) GetLockedKey() string {
	return l.LockedKey
}

// TryLockWithPrefix will get all identities(instanceID) distributed from control plane and try to lock one
func (l *EtcdLocker) TryLockWithPrefix(prefix string, filter func(k, v []byte) bool) error {
	resp, err := l.EtcdClient.Get(CreateEtcdCtxInfoWithTimeout(context.TODO(), defaultRequestTimeout), prefix,
		clientv3.WithPrefix())
	if err != nil {
		log.GetLogger().Errorf("failed to get prefix %s from etcd error %s", prefix, err.Error())
		return err
	}
	if len(resp.Kvs) == 0 {
		log.GetLogger().Warnf("no etcd key is found for prefix %s", prefix)
		return ErrNoKeyCanBeLocked
	}
	var (
		locked     bool
		tryLockErr error
	)
	for _, kv := range resp.Kvs {
		if filter(kv.Key, kv.Value) {
			tryLockErr = ErrNoKeyCanBeLocked
			continue
		}
		tryLockErr = l.TryLock(string(kv.Key))
		if tryLockErr == nil {
			locked = true
			break
		}
	}
	if !locked {
		if tryLockErr != nil {
			return tryLockErr
		} else {
			return ErrNoKeyCanBeLocked
		}
	}
	return nil
}

// TryLock -
func (l *EtcdLocker) TryLock(key string) error {
	if err := l.tryLock(key); err != nil {
		return err
	}
	go l.lockKeeperLoop()
	return nil
}

func (l *EtcdLocker) tryLock(key string) error {
	grtCtx := CreateEtcdCtxInfoWithTimeout(context.Background(), DurationContextTimeout)
	if l.leaseID == clientv3.NoLease {
		leaseID, err := l.EtcdClient.Grant(grtCtx, int64(l.LeaseTTL))
		if err != nil {
			log.GetLogger().Errorf("failed to grant lease for key in %s etcd error %s", l.EtcdClient.GetEtcdType(),
				err.Error())
			return err
		}
		l.leaseID = leaseID
	}
	l.holderKey = fmt.Sprintf("%s/%x", key, l.leaseID)
	log.GetLogger().Infof("generate holderKey %s", l.holderKey)
	var lockErr error
	defer func() {
		if lockErr != nil {
			log.GetLogger().Errorf("failed to lock key %s, delete holder key %s", key, l.holderKey)
			rvkCtx := CreateEtcdCtxInfoWithTimeout(context.Background(), DurationContextTimeout)
			if err := l.EtcdClient.Revoke(rvkCtx, l.leaseID); err != nil {
				log.GetLogger().Errorf("failed to revoke lease %d error %d", l.leaseID, err.Error())
			}
			l.leaseID = clientv3.NoLease
			delCtx := CreateEtcdCtxInfoWithTimeout(context.Background(), DurationContextTimeout)
			if err := l.EtcdClient.Delete(delCtx, l.holderKey); err != nil {
				log.GetLogger().Errorf("failed to delete holder key %s error %d", l.holderKey, err.Error())
			}
		}
	}()
	cmp := clientv3.Compare(clientv3.LeaseValue(l.holderKey), "=", clientv3.NoLease)
	put := clientv3.OpPut(l.holderKey, "", clientv3.WithLease(l.leaseID))
	get := clientv3.OpGet(l.holderKey)
	// key is already been put, we want to get the minimum holder key so use WithLimit(2)
	getKeyHolder := clientv3.OpGet(key, []clientv3.OpOption{clientv3.WithPrefix(), clientv3.WithSort(
		clientv3.SortByCreateRevision, clientv3.SortAscend), clientv3.WithLimit(lockedKeyHoldIndex + 1)}...)
	txnCtx := CreateEtcdCtxInfoWithTimeout(context.Background(), DurationContextTimeout)
	var resp *clientv3.TxnResponse
	resp, lockErr = l.EtcdClient.Client.Txn(txnCtx.Ctx).If(cmp).Then(put, getKeyHolder).Else(get, getKeyHolder).Commit()
	if lockErr != nil {
		log.GetLogger().Errorf("failed to lock key %s, transaction error %s", key, lockErr.Error())
		return lockErr
	}
	if len(resp.Responses) != lockedKeyHoldIndex+1 {
		log.GetLogger().Errorf("failed to lock key %s, transaction response size %s is invalid", key,
			len(resp.Responses))
		lockErr = ErrEtcdResponseInvalid
		return lockErr
	}
	var myRevision int64
	if resp.Succeeded {
		myRevision = resp.Header.Revision
	} else {
		if len(resp.Responses[0].GetResponseRange().Kvs) == 0 {
			log.GetLogger().Errorf("failed to lock key %s, transaction response[0] kvs size is 0", key)
			lockErr = ErrEtcdResponseInvalid
			return lockErr
		}
		myRevision = resp.Responses[0].GetResponseRange().Kvs[0].CreateRevision
	}
	log.GetLogger().Infof("get holderKey %s my revision %d", l.holderKey, myRevision)
	// resp.Responses[1] contains info got from getKeyHolder, ideally looks like [originKey, holderKey] after sorting,
	// because originKey is put by control plane and has lower revision than any holderKey attached with a lease
	holderKvs := resp.Responses[1].GetResponseRange().Kvs
	// holderKvs[0] is not the originKey means originKey is deleted
	if len(holderKvs) == 0 || string(holderKvs[0].Key) != key {
		log.GetLogger().Warnf("failed to find key %s, key may be deleted", l.holderKey)
		lockErr = ErrNoKeyCanBeFound
		return lockErr
	}
	// holderKvs[1] has different revision from myRevision means other one has locked this key before me
	if len(holderKvs) > 1 && holderKvs[1].CreateRevision != myRevision {
		log.GetLogger().Warnf("failed to lock key %s, key already locked, holder revision %d", l.holderKey,
			holderKvs[1].CreateRevision)
		lockErr = ErrNoKeyCanBeLocked
		return lockErr
	}
	l.LockedKey = key
	l.unlockCh = make(chan struct{})
	if l.LockCallback != nil {
		if lockErr = l.LockCallback(l); lockErr != nil {
			log.GetLogger().Warnf("failed to process lock callback of %s error %s", key, lockErr.Error())
			return lockErr
		}
	}
	log.GetLogger().Infof("succeed to lock key %s", key)
	return nil
}

// Unlock -
func (l *EtcdLocker) Unlock() error {
	delCtx := CreateEtcdCtxInfoWithTimeout(context.Background(), DurationContextTimeout)
	err := l.EtcdClient.Delete(delCtx, l.holderKey)
	if err != nil {
		log.GetLogger().Errorf("failed to unlock key %s , delete holder %s error %s", l.LockedKey, l.holderKey,
			err.Error())
	}
	if l.UnlockCallback != nil {
		if err = l.UnlockCallback(l); err != nil {
			log.GetLogger().Errorf("failed to process unlock callback of %s error %s", l.LockedKey, err.Error())
		}
	}
	l.LockedKey = ""
	l.holderKey = ""
	l.leaseID = clientv3.NoLease
	utils.SafeCloseChannel(l.unlockCh)
	return err
}

func (l *EtcdLocker) lockKeeperLoop() {
	leaseTicker := time.NewTicker(time.Duration(l.LeaseTTL)*time.Second - refreshAheadTime)
	defer leaseTicker.Stop()
	failCount := 0
	for {
		select {
		case _, ok := <-l.unlockCh:
			if !ok {
				log.GetLogger().Warnf("unlock channel triggers for etcd lock of key %s", l.LockedKey)
			}
			return
		case _, ok := <-l.StopCh:
			if !ok {
				log.GetLogger().Warnf("stop channel triggers for etcd lock of key %s", l.LockedKey)
			}
			l.Unlock()
			return
		case <-leaseTicker.C:
			if l.leaseID == clientv3.NoLease {
				// wait for multiple leaseTTL time to make sure lease is expired at server side
				time.Sleep(time.Duration(l.LeaseTTL) * time.Second)
				if err := l.tryLock(l.LockedKey); err == ErrNoKeyCanBeFound || err == ErrNoKeyCanBeLocked {
					log.GetLogger().Errorf("cannot keep lock key %s, lock fail count %d error %s", l.LockedKey,
						failCount, err)
					if failCount >= lockFailCountLimit {
						l.FailCallback()
						return
					}
					failCount++
				} else {
					failCount = 0
				}
			} else {
				getCtx := CreateEtcdCtxInfoWithTimeout(context.Background(), DurationContextTimeout)
				resp, err := l.EtcdClient.Get(getCtx, l.LockedKey)
				if err != nil {
					log.GetLogger().Errorf("unable to get locked key %s in %s etcd error %s", l.LockedKey,
						l.EtcdClient.GetEtcdType(), err.Error())
					l.leaseID = clientv3.NoLease
					continue
				}
				if len(resp.Kvs) == 0 {
					log.GetLogger().Warnf("locked key %s is deleted in %s etcd unlock now", l.LockedKey,
						l.EtcdClient.GetEtcdType())
					l.Unlock()
					l.FailCallback()
					return
				}
				keepAliveOnceCtx := CreateEtcdCtxInfoWithTimeout(context.Background(), DurationContextTimeout)
				err = l.EtcdClient.KeepAliveOnce(keepAliveOnceCtx, l.leaseID)
				if err != nil {
					log.GetLogger().Errorf("unable to refresh lease in %s etcd error %s", l.EtcdClient.GetEtcdType(),
						err.Error())
					l.leaseID = clientv3.NoLease
				}
			}
		}
	}
}
