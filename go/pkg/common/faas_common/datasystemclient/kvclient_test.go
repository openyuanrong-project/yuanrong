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

// Package datasystemclient is data system client used for communicating with data system worker.
package datasystemclient

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	mockUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
)

func localInit() {
	cache := &Cache{
		nodeList:   []string{},
		invalidMap: make(map[string]struct{}, 1),
	}
	localClientLibruntime = &mockUtils.FakeLibruntimeSdkClient{}
	cache.addNode("127.2.2.101", log.GetLogger())
	cache.addNode("127.2.2.102", log.GetLogger())
	dataSystemCache.Store(noCluster, cache)
}

func localRecover() {
	dataSystemCache.Delete(noCluster)
}

func TestKVDelWithRetry(t *testing.T) {
	defer gomonkey.ApplyFunc(NewClient,
		func(tenantID string, nodeIP string) (DsClientImpl, error) {
			return DsClientImpl{kvClient: &FakeKvClient{num: 1}}, nil
		}).Reset()
	defer gomonkey.ApplyFunc((*Cache).healthCheckProcess,
		func(_ *Cache, node string) {
			return
		})
	localInit()
	defer localRecover()
	type args struct {
		key     string
		option  *Option
		traceID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "del success",
			args: args{
				key: "aaaa",
				option: &Option{
					TenantID:  "tenant1",
					NodeIP:    "127.2.2.101",
					Cluster:   noCluster,
					WriteMode: 0,
					TTLSecond: 60,
				},
				traceID: "aaaaa",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				if err != nil {
					t.Errorf("del failed, err %s", err.Error())
					return false
				}
				return true
			},
		},
		{
			name: "del failed",
			args: args{
				key: "key2",
				option: &Option{
					TenantID:  "tenant1",
					NodeIP:    "127.2.2.101",
					Cluster:   noCluster,
					WriteMode: 0,
					TTLSecond: 60,
				},
				traceID: "aaaaa",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Equal(t, "no data system node is available", err.Error())
			},
		},
		{
			name: "find other node success",
			args: args{
				key: "aaaa",
				option: &Option{
					TenantID:  "tenant1",
					NodeIP:    "127.2.2.1",
					Cluster:   noCluster,
					WriteMode: 0,
					TTLSecond: 60,
				},
				traceID: "aaaaa",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				if err != nil {
					t.Errorf("del failed, err %s", err.Error())
					return false
				}
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, KVDelWithRetry(tt.args.key, tt.args.option, tt.args.traceID), fmt.Sprintf("KVDelWithRetry(%v, %v, %v)", tt.args.key, tt.args.option, tt.args.traceID))
		})
	}
}

func TestKVGetWithRetry(t *testing.T) {
	defer gomonkey.ApplyFunc(NewClient,
		func(tenantID string, nodeIP string) (DsClientImpl, error) {
			return DsClientImpl{kvClient: &FakeKvClient{num: 1}}, nil
		}).Reset()
	defer gomonkey.ApplyFunc((*Cache).healthCheckProcess,
		func(_ *Cache, node string) {
			return
		})
	localInit()
	defer localRecover()
	type args struct {
		key     string
		option  *Option
		traceID string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "get success",
			args: args{
				key: "key1",
				option: &Option{
					TenantID:  "tenant1",
					NodeIP:    "127.2.2.101",
					Cluster:   noCluster,
					WriteMode: 0,
					TTLSecond: 60,
				},
				traceID: "aaaaa",
			},
			want: []byte("value1"),
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				if err != nil {
					t.Errorf("get failed, err %s", err.Error())
					return false
				}
				return true
			},
		},
		{
			name: "get failed",
			args: args{
				key: "key2",
				option: &Option{
					TenantID:  "tenant1",
					NodeIP:    "127.2.2.101",
					Cluster:   noCluster,
					WriteMode: 0,
					TTLSecond: 60,
				},
				traceID: "aaaaa",
			},
			want: nil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Equal(t, "no data system node is available", err.Error())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := KVGetWithRetry(tt.args.key, tt.args.option, tt.args.traceID)
			if !tt.wantErr(t, err, fmt.Sprintf("KVGetWithRetry(%v, %v, %v)", tt.args.key, tt.args.option, tt.args.traceID)) {
				return
			}
			assert.Equalf(t, tt.want, got, "KVGetWithRetry(%v, %v, %v)", tt.args.key, tt.args.option, tt.args.traceID)
		})
	}
}

func TestKVPutWithRetry(t *testing.T) {
	defer gomonkey.ApplyFunc(NewClient,
		func(tenantID string, nodeIP string) (DsClientImpl, error) {
			return DsClientImpl{kvClient: &FakeKvClient{num: 1}}, nil
		}).Reset()
	defer gomonkey.ApplyFunc((*Cache).healthCheckProcess,
		func(_ *Cache, node string) {
			return
		})
	localInit()
	defer localRecover()
	type args struct {
		key     string
		value   []byte
		option  *Option
		traceID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "put success",
			args: args{
				key:   "key1",
				value: []byte("value1"),
				option: &Option{
					TenantID:  "tenant1",
					NodeIP:    "127.2.2.101",
					Cluster:   noCluster,
					WriteMode: 0,
					TTLSecond: 60,
				},
				traceID: "aaaaa",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				if err != nil {
					t.Errorf("del failed, err %s", err.Error())
					return false
				}
				return true
			},
		},
		{
			name: "put failed",
			args: args{
				key:   "key2",
				value: []byte("value1"),
				option: &Option{
					TenantID:  "tenant1",
					NodeIP:    "127.2.2.101",
					Cluster:   noCluster,
					WriteMode: 0,
					TTLSecond: 60,
				},
				traceID: "aaaaa",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Equal(t, "no data system node is available", err.Error())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, KVPutWithRetry(tt.args.key, tt.args.value, tt.args.option, tt.args.traceID), fmt.Sprintf("KVPutWithRetry(%v, %v, %v, %v)", tt.args.key, tt.args.value, tt.args.option, tt.args.traceID))
		})
	}
}

func Test_kvDel(t *testing.T) {
	type args struct {
		key     string
		config  *Config
		traceID string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := kvDel(tt.args.key, tt.args.config, tt.args.traceID)
			if !tt.wantErr(t, err, fmt.Sprintf("kvDel(%v, %v, %v)", tt.args.key, tt.args.config, tt.args.traceID)) {
				return
			}
			assert.Equalf(t, tt.want, got, "kvDel(%v, %v, %v)", tt.args.key, tt.args.config, tt.args.traceID)
		})
	}
}

func Test_kvGet(t *testing.T) {
	type args struct {
		key     string
		config  *Config
		traceID string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		want1   bool
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := kvGet(tt.args.key, tt.args.config, tt.args.traceID)
			if !tt.wantErr(t, err, fmt.Sprintf("kvGet(%v, %v, %v)", tt.args.key, tt.args.config, tt.args.traceID)) {
				return
			}
			assert.Equalf(t, tt.want, got, "kvGet(%v, %v, %v)", tt.args.key, tt.args.config, tt.args.traceID)
			assert.Equalf(t, tt.want1, got1, "kvGet(%v, %v, %v)", tt.args.key, tt.args.config, tt.args.traceID)
		})
	}
}

func Test_kvPut(t *testing.T) {
	type args struct {
		value   []byte
		param   api.SetParam
		config  *Config
		traceID string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := kvPut(tt.args.value, tt.args.param, tt.args.config, tt.args.traceID)
			if !tt.wantErr(t, err, fmt.Sprintf("kvPut(%v, %v, %v, %v)", tt.args.value, tt.args.param, tt.args.config, tt.args.traceID)) {
				return
			}
			assert.Equalf(t, tt.want, got, "kvPut(%v, %v, %v, %v)", tt.args.value, tt.args.param, tt.args.config, tt.args.traceID)
		})
	}
}
