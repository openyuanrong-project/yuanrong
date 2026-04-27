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
	"time"

	"go.etcd.io/etcd/client/v3"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

const (
	instanceEtcdKeyTTL     = 30
	defaultRefreshInterval = 15 * time.Second
)

var (
	refreshInterval = defaultRefreshInterval
)

// EtcdRegister - register to specified ETCD
type EtcdRegister struct {
	EtcdClient  *EtcdClient
	InstanceKey string
	Value       string
	leaseID     clientv3.LeaseID
	StopCh      <-chan struct{}
}

// Register - register instance to meta etcd or router etcd
func (r *EtcdRegister) Register() error {
	if r.EtcdClient != GetMetaEtcdClient() && r.EtcdClient != GetRouterEtcdClient() {
		log.GetLogger().Errorf("etcdClient is not meta or route etcd")
		return errors.New("etcdClient is not meta or route etcd")
	}
	var err error
	err = r.putInstanceInfoToEtcd()
	if err != nil {
		log.GetLogger().Errorf("failed to register instance to %s etcd when start, error:%s",
			r.EtcdClient.GetEtcdType(), err.Error())
		return err
	}
	go r.startRefreshLeaseJob()
	return nil
}

func (r *EtcdRegister) startRefreshLeaseJob() {
	if r.StopCh == nil {
		log.GetLogger().Errorf("StopCh is nil, lease in %s etcd will not be refreshed",
			r.EtcdClient.GetEtcdType())
		return
	}
	refreshTicker := time.NewTicker(refreshInterval)
	defer refreshTicker.Stop()
	for {
		select {
		case <-refreshTicker.C:
			r.refreshLease()
		case <-r.StopCh:
			log.GetLogger().Warnf("stopping refresh lease job")
			refreshTicker.Stop()
			r.stopLease()
			return
		}
	}
}

func (r *EtcdRegister) stopLease() {
	revokeCtx := CreateEtcdCtxInfoWithTimeout(context.Background(), DurationContextTimeout)
	err := r.EtcdClient.Revoke(revokeCtx, r.leaseID)
	if err != nil {
		log.GetLogger().Warnf("revoke lease in %s etcd failed, err:%s",
			r.EtcdClient.GetEtcdType(), err.Error())
	}
	ctx := CreateEtcdCtxInfoWithTimeout(context.Background(), DurationContextTimeout)
	err = r.EtcdClient.Delete(ctx, r.InstanceKey)
	if err != nil {
		log.GetLogger().Errorf("delete key: %s,from %s etcd failed, err:%s",
			r.InstanceKey, r.EtcdClient.GetEtcdType(), err.Error())
	}
}

func (r *EtcdRegister) refreshLease() {
	if !r.isKeyExist() {
		if err := r.putInstanceInfoToEtcd(); err != nil {
			return
		}
	}
	keepAliveOnceCtx := CreateEtcdCtxInfoWithTimeout(context.Background(), DurationContextTimeout)
	err := r.EtcdClient.KeepAliveOnce(keepAliveOnceCtx, r.leaseID)
	if err != nil {
		log.GetLogger().Errorf("unable to refresh lease in %s etcd:%s",
			r.EtcdClient.GetEtcdType(), err.Error())
	}
}

func (r *EtcdRegister) isKeyExist() bool {
	ctx := CreateEtcdCtxInfoWithTimeout(context.Background(), DurationContextTimeout)
	resp, err := r.EtcdClient.GetResponse(ctx, r.InstanceKey,
		clientv3.WithKeysOnly(), clientv3.WithSerializable())
	if err != nil {
		log.GetLogger().Errorf("failed to get new key:%s from %s etcd, err:%s",
			r.InstanceKey, r.EtcdClient.GetEtcdType(), err.Error())
		return false
	}
	return len(resp.Kvs) > 0
}

func (r *EtcdRegister) putInstanceInfoToEtcd() error {
	grantCtx := CreateEtcdCtxInfoWithTimeout(context.Background(), DurationContextTimeout)
	id, err := r.EtcdClient.Grant(grantCtx, instanceEtcdKeyTTL)
	if err != nil {
		log.GetLogger().Errorf("failed to grant instance lease in %s etcd: %s", r.EtcdClient.GetEtcdType(),
			err.Error())
		return err
	}

	ctx := CreateEtcdCtxInfoWithTimeout(context.Background(), DurationContextTimeout)
	err = r.EtcdClient.Put(ctx, r.InstanceKey, r.Value, clientv3.WithLease(id))
	if err != nil {
		log.GetLogger().Errorf("unable to put new key:%s to %s etcd, err:%s",
			r.InstanceKey, r.EtcdClient.GetEtcdType(), err.Error())
		return err
	}
	r.leaseID = id
	log.GetLogger().Infof("register instance key:%s, value:%s to %s etcd successfully!",
		r.InstanceKey, r.Value, r.EtcdClient.GetEtcdType())
	return nil
}
