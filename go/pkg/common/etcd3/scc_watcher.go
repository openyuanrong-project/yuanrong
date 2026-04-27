//go:build cryptoapi
// +build cryptoapi

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
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/net/context"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

// NewEtcdWatcher new a etcd watcher
func NewEtcdWatcher(config EtcdConfig) (*EtcdClient, error) {
	cfg, err := config.getEtcdAuthType().getEtcdConfig()
	if err != nil {
		return nil, err
	}

	cfg.DialTimeout = dialTimeout
	cfg.DialKeepAliveTime = keepaliveTime
	cfg.DialKeepAliveTimeout = keepaliveTimeout

	cfg.Endpoints = config.Servers
	client, err := clientv3.New(*cfg)
	if err != nil {
		return nil, err
	}
	stopCh := make(chan struct{})
	go config.getEtcdAuthType().renewToken(client, stopCh)

	// fetch registered grpc-proxy endpoints
	if err = client.Sync(context.Background()); err != nil {
		log.GetLogger().Warnf("Sync endpoints: %s", err.Error())
	}
	log.GetLogger().Infof("Etcd discovered endpoints: %v", client.Endpoints())
	return &EtcdClient{Client: client, AZPrefix: config.AZPrefix}, nil
}
