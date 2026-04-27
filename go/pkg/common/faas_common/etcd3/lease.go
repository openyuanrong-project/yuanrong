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
	"go.etcd.io/etcd/client/v3"
)

// Grant -
func (e *EtcdClient) Grant(ctxInfo EtcdCtxInfo, ttl int64) (clientv3.LeaseID, error) {
	ctx, cancel := ctxInfo.Ctx, ctxInfo.Cancel
	e.rwMutex.RLock()
	resp, err := e.Client.Grant(ctx, ttl)
	e.rwMutex.RUnlock()
	cancel()
	if err != nil {
		return 0, err
	}
	return resp.ID, nil
}

// KeepAliveOnce -
func (e *EtcdClient) KeepAliveOnce(ctxInfo EtcdCtxInfo, leaseID clientv3.LeaseID) error {
	ctx, cancel := ctxInfo.Ctx, ctxInfo.Cancel
	e.rwMutex.RLock()
	_, err := e.Client.KeepAliveOnce(ctx, leaseID)
	e.rwMutex.RUnlock()
	cancel()
	return err
}

// Revoke -
func (e *EtcdClient) Revoke(ctxInfo EtcdCtxInfo, leaseID clientv3.LeaseID) error {
	ctx, cancel := ctxInfo.Ctx, ctxInfo.Cancel
	e.rwMutex.RLock()
	_, err := e.Client.Revoke(ctx, leaseID)
	e.rwMutex.RUnlock()
	cancel()
	return err
}
