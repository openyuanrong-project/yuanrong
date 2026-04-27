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

// Package etcd3 type
package etcd3

import (
	"sync"
	"time"

	"go.etcd.io/etcd/client/v3"
)

// EtcdWatcherFilter defines watch filter of etcd
type EtcdWatcherFilter func(*Event) bool

// EtcdWatcherHandler defines watch handler of etcd
type EtcdWatcherHandler func(*Event)

// EtcdClient wrapper etcd client
type EtcdClient struct {
	Client    *clientv3.Client
	config    *EtcdConfig
	etcdTimer *time.Timer
	rwMutex   sync.RWMutex
	cond      *sync.Cond
	// notify goroutine keepConnAlive exit
	stopCh       <-chan struct{}
	clientExitCh chan struct{}
	exitOnce     sync.Once
	etcdType     string
	// router etcd status lost contact after defaultEtcdLostContactTime, true is healthy, false is unhealthy
	etcdStatusAfterLostContact bool
	etcdStatusNow              bool
	isAlarmEnable              bool
	abnormalContinuouslyTimes  int
}

// EtcdWatcher -
type EtcdWatcher struct {
	filter        EtcdWatcherFilter
	handler       EtcdWatcherHandler
	cacheConfig   EtcdCacheConfig
	watcher       *EtcdClient
	ResultChan    chan *Event
	CacheChan     chan *Event
	resultChanWG  *sync.WaitGroup
	configCh      chan struct{}
	stopCh        <-chan struct{}
	key           string
	etcdType      string
	initialRev    int64
	historyRev    int64
	cacheFlushing bool
	sync.Mutex
}

// EtcdInitParam -
type EtcdInitParam struct {
	metaEtcdConfig       *EtcdConfig
	routeEtcdConfig      *EtcdConfig
	CAEMetaEtcdConfig    *EtcdConfig
	DataSystemEtcdConfig *EtcdConfig
	stopCh               <-chan struct{}
	enableAlarm          bool
}

// EtcdConfig the info to get function instance
type EtcdConfig struct {
	Servers        []string `json:"servers" valid:"optional"`
	AZPrefix       string   `json:"azPrefix" valid:"optional"`
	User           string   `json:"user" valid:"optional"`
	Password       string   `json:"password" valid:"optional"`
	SslEnable      bool     `json:"sslEnable" valid:"optional"`
	AuthType       string   `json:"authType" valid:"optional"`
	UseSecret      bool     `json:"useSecret" valid:"optional"`
	SecretName     string   `json:"secretName" valid:"optional"`
	LimitRate      int      `json:"limitRate,omitempty" valid:"optional"`
	LimitBurst     int      `json:"limitBurst,omitempty" valid:"optional"`
	LimitTimeout   int      `json:"limitTimeout,omitempty" valid:"optional"`
	CaFile         string   `json:"cafile,omitempty" valid:",optional"`
	CertFile       string   `json:"certfile,omitempty" valid:",optional"`
	KeyFile        string   `json:"keyfile,omitempty" valid:",optional"`
	PassphraseFile string   `json:"passphraseFile,omitempty" valid:",optional"`
}

// EtcdCacheConfig -
type EtcdCacheConfig struct {
	EnableCache    bool   `json:"enableCache"`
	PersistPath    string `json:"persistPath"`
	FlushInterval  int    `json:"flushInterval"`
	FlushThreshold int    `json:"flushThreshold"`
	MetaFilePath   string
	DataFilePath   string
	BackupFilePath string
}
