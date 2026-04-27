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
package etcd3

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/util/wait"

	"yuanrong.org/kernel/pkg/common/crypto"
)

type EtcdTestSuite struct {
	suite.Suite
	defaultEtcdCtx EtcdCtxInfo
	etcdClient     *EtcdClient
}

func (es *EtcdTestSuite) setupEtcdClient() {
	var err error
	patches := gomonkey.NewPatches()
	cli := clientv3.Client{}
	patches.ApplyFunc(clientv3.New, func(_ clientv3.Config) (*clientv3.Client, error) {
		return &cli, nil
	})
	patches.ApplyMethod(reflect.TypeOf(&cli), "Sync", func(_ *clientv3.Client, _ context.Context) error {
		return nil
	})
	patches.ApplyMethod(reflect.TypeOf(&cli), "Endpoints", func(_ *clientv3.Client) []string {
		return []string{"localhost:0"}
	})
	defer patches.Reset()
	// create a new etcd watcher for the following tests
	auth := EtcdConfig{
		Servers:   []string{"localhost:0"},
		User:      "",
		Passwd:    "",
		SslEnable: false,
	}
	es.etcdClient, err = NewEtcdWatcher(auth)
	if err != nil {
		err = fmt.Errorf("failed to create etcd watcher; err: %v", err)
	}
}

func (es *EtcdTestSuite) SetupSuite() {
	es.defaultEtcdCtx = CreateEtcdCtxInfo(context.Background())
	es.setupEtcdClient()
}
func (es *EtcdTestSuite) TearDownSuite() {
	es.etcdClient = nil
}

func (es *EtcdTestSuite) TestNewEtcdWatcher() {
	var (
		serverList []string
		auth       EtcdConfig
		err        error
	)

	patches := gomonkey.NewPatches()
	cli := clientv3.Client{}
	patches.ApplyFunc(clientv3.New, func(_ clientv3.Config) (*clientv3.Client, error) {
		return &cli, nil
	})
	patches.ApplyMethod(reflect.TypeOf(&cli), "Sync", func(_ *clientv3.Client, _ context.Context) error {
		return nil
	})
	patches.ApplyMethod(reflect.TypeOf(&cli), "Endpoints", func(_ *clientv3.Client) []string {
		return []string{"localhost:0"}
	})
	patches.ApplyFunc(crypto.Decrypt, func(cipherText []byte, secret []byte) (string, error) {
		return "key", nil
	})
	patches.ApplyFunc(wait.Until, func(f func(), period time.Duration, stopCh <-chan struct{}) {
		return
	})
	serverList = []string{"localhost:0"}
	auth = EtcdConfig{
		Servers:   serverList,
		User:      "",
		Passwd:    "",
		SslEnable: false,
		CaFile:    "",
		CertFile:  "",
		KeyFile:   "",
	}
	_, err = NewEtcdWatcher(auth)
	assert.Nil(es.T(), err)

	serverList = []string{"localhost:0"}
	auth = EtcdConfig{
		Servers:   serverList,
		User:      "",
		Passwd:    "",
		SslEnable: true,
		CaFile:    "xxx.ca",
		CertFile:  "xxx.cert",
		KeyFile:   "xxx.key",
	}
	_, err = NewEtcdWatcher(auth)
	assert.NotNil(es.T(), err)

	serverList = []string{"localhost:0"}
	auth = EtcdConfig{
		Servers:   serverList,
		User:      "user",
		Passwd:    "",
		SslEnable: false,
	}
	_, err = NewEtcdWatcher(auth)
	assert.Nil(es.T(), err)
	patches.Reset()
}

func (es *EtcdTestSuite) TestCRUD() {
	cli := es.etcdClient
	ctxInfo := es.defaultEtcdCtx
	etcdKey := "test_key"
	etcdValue := "test_value"

	patch := gomonkey.ApplyFunc(clientv3.NewKV, func(*clientv3.Client) clientv3.KV {
		return KV{}
	})
	err := cli.Put(ctxInfo, etcdKey, etcdValue)
	assert.Nil(es.T(), err)

	_, err = cli.GetValues(ctxInfo, etcdKey)
	assert.Nil(es.T(), err)

	_, err = cli.GetResponse(ctxInfo, etcdKey)
	assert.Nil(es.T(), err)

	err = cli.Delete(ctxInfo, etcdKey)
	assert.Nil(es.T(), err)

	patch.Reset()
}

func (es *EtcdTestSuite) TestCreateEtcdCtxInfoWithTimeout() {
	ctxInfo := es.defaultEtcdCtx.Ctx
	etcdCtxInfo := CreateEtcdCtxInfoWithTimeout(ctxInfo, 1)
	assert.NotNil(es.T(), etcdCtxInfo)
}

func TestEtcdTestSuite(t *testing.T) {
	suite.Run(t, new(EtcdTestSuite))
}

func TestDecryptEtcdPassword(t *testing.T) {
	DecryptEtcdPassword([]byte{}, []byte{})
}

type KV struct {
}

func (k KV) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	return &clientv3.GetResponse{}, nil
}

func (k KV) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	return &clientv3.PutResponse{}, nil
}

func (k KV) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	return &clientv3.DeleteResponse{}, nil
}

func (k KV) Compact(ctx context.Context, rev int64, opts ...clientv3.CompactOption) (*clientv3.CompactResponse, error) {
	return &clientv3.CompactResponse{}, nil
}

func (k KV) Do(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
	return clientv3.OpResponse{}, nil
}

func (k KV) Txn(ctx context.Context) clientv3.Txn {
	return nil
}
