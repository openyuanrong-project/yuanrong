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
// To use data system, you should export the data system lib path. Please refer to the Dockerfile of the frontend.
// The lib should copied to home/sn/bin/datasystem/lib. Please refer to
// functioncore/build/common/common_compile.sh and the Dockerfile of the frontend.
// NOTE: To change the version of data system, must revise the version in the common_compile.sh, test.sh and the go.mod
package datasystemclient

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/gin-gonic/gin"
	"github.com/smartystreets/goconvey/convey"
	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/data"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	mockUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
)

type FakeKvClient struct {
	num int
}

func (f *FakeKvClient) KVSet(key string, value []byte, param api.SetParam) api.ErrorInfo {
	switch key {
	case "key1":
		return api.ErrorInfo{Code: 0, Err: nil}
	case "key2":
		return api.ErrorInfo{Code: errOutOfMemory, Err: errors.New("err2")}
	}
	return api.ErrorInfo{Code: 0, Err: nil}
}

func (f *FakeKvClient) KVSetWithoutKey(value []byte, param api.SetParam) (string, api.ErrorInfo) {
	return "", api.ErrorInfo{}
}

func (f *FakeKvClient) KVGet(key string, timeoutms ...uint32) ([]byte, api.ErrorInfo) {
	switch key {
	case "key1":
		return []byte("value1"), api.ErrorInfo{Code: 0, Err: nil}
	case "key2":
		return []byte("value1"), api.ErrorInfo{Code: errOutOfMemory, Err: errors.New("err2")}
	}
	return []byte(""), api.ErrorInfo{}
}

func (f *FakeKvClient) KVGetMulti(keys []string, timeoutms ...uint32) ([][]byte, api.ErrorInfo) {
	switch len(keys) {
	case 3:
		return nil, api.ErrorInfo{Code: errKeyNotFound}
	case 2:
		return [][]byte{[]byte("value1"), []byte("value2")}, api.ErrorInfo{Code: 0, Err: nil}
	case 1:
		return [][]byte{[]byte("value3")}, api.ErrorInfo{Code: errOutOfMemory, Err: errors.New("err2")}
	}
	return [][]byte{}, api.ErrorInfo{}
}

func (f *FakeKvClient) KVQuerySize(keys []string) ([]uint64, api.ErrorInfo) {
	switch len(keys) {
	case 1:
		return []uint64{10}, api.ErrorInfo{}
	case 2:
		return []uint64{10, 10}, api.ErrorInfo{}
	case 3:
		return []uint64{100, 100, 100}, api.ErrorInfo{}
	case 4:
		return []uint64{}, api.ErrorInfo{Code: errOutOfMemory, Err: errors.New("err2")}
	case 5:
		return []uint64{100000, 100000, 100000, 100000, 100000}, api.ErrorInfo{}
	}
	return []uint64{}, api.ErrorInfo{}
}

func (f *FakeKvClient) KVDel(key string) api.ErrorInfo {
	switch key {
	case "key1":
		return api.ErrorInfo{Code: 0, Err: nil}
	case "key2":
		return api.ErrorInfo{Code: errOutOfMemory, Err: errors.New("err2")}
	}
	return api.ErrorInfo{}
}

func (f *FakeKvClient) KVDelMulti(keys []string) ([]string, api.ErrorInfo) {
	switch len(keys) {
	case 2:
		return nil, api.ErrorInfo{Code: 0, Err: nil}
	case 1:
		return []string{"key3"}, api.ErrorInfo{Code: errOutOfMemory, Err: errors.New("err2")}
	}
	return []string{}, api.ErrorInfo{}
}

func (f *FakeKvClient) GenerateKey() string {
	return "1"
}

func (f *FakeKvClient) SetTraceID(traceID string) {
}

func (f *FakeKvClient) DestroyClient() {}

func TestUploadWithKeyRetryKvClient(t *testing.T) {
	cache := &Cache{
		nodeList:   []string{},
		invalidMap: make(map[string]struct{}, 1),
	}
	cache.addNode("127.2.2.101", log.GetLogger())
	cache.addNode("127.2.2.102", log.GetLogger())
	dataSystemCache.Store(noCluster, cache)
	type args struct {
		deviceID     string
		value        []byte
		config       *Config
		param        api.SetParam
		sourceClient DsClientImpl
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		patchesFunc mockUtils.PatchesFunc
	}{
		{"case1 succeed to upload key", args{deviceID: "key1", value: []byte("value1"),
			config:       &Config{TenantID: "tenant1", NodeIP: "127.2.2.101"},
			param:        api.SetParam{WriteMode: api.NoneL2Cache, TTLSecond: 60 * 1000},
			sourceClient: DsClientImpl{kvClient: &FakeKvClient{num: 1}}}, false, func() mockUtils.PatchSlice {
			patches := mockUtils.InitPatchSlice()
			patches.Append(mockUtils.PatchSlice{
				gomonkey.ApplyFunc(NewClient,
					func(tenantID string, nodeIP string) (DsClientImpl, error) {
						return DsClientImpl{kvClient: &FakeKvClient{num: 1}}, nil
					}),
				gomonkey.ApplyFunc((*Cache).healthCheckProcess,
					func(_ *Cache, node string) {
						return
					}),
			})
			return patches
		}},

		{"case2 succeed to upload key when node ip is null", args{deviceID: "key1", value: []byte("value1"),
			config:       &Config{TenantID: "tenant1", NodeIP: ""},
			param:        api.SetParam{WriteMode: api.NoneL2Cache, TTLSecond: 60 * 1000},
			sourceClient: DsClientImpl{kvClient: &FakeKvClient{num: 1}}}, false, func() mockUtils.PatchSlice {
			patches := mockUtils.InitPatchSlice()
			patches.Append(mockUtils.PatchSlice{
				gomonkey.ApplyFunc(NewClient,
					func(tenantID string, nodeIP string) (DsClientImpl, error) {
						return DsClientImpl{kvClient: &FakeKvClient{num: 1}}, nil
					}),
				gomonkey.ApplyFunc((*Cache).healthCheckProcess,
					func(_ *Cache, node string) {
						return
					}),
			})
			return patches
		}},
		{"case3 failed to upload key when node ip and tenantID is null", args{deviceID: "key1", value: []byte("value1"),
			config:       &Config{TenantID: "", NodeIP: ""},
			param:        api.SetParam{WriteMode: api.NoneL2Cache, TTLSecond: 60 * 1000},
			sourceClient: DsClientImpl{kvClient: &FakeKvClient{num: 1}}}, true, func() mockUtils.PatchSlice {
			patches := mockUtils.InitPatchSlice()
			patches.Append(mockUtils.PatchSlice{
				gomonkey.ApplyFunc(NewClient,
					func(tenantID string, nodeIP string) (DsClientImpl, error) {
						return DsClientImpl{}, errors.New("failed to upload")
					}),
				gomonkey.ApplyFunc((*Cache).healthCheckProcess,
					func(_ *Cache, node string) {
						return
					}),
			})
			return patches
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := tt.patchesFunc()
			localClientLibruntime = &mockUtils.FakeLibruntimeSdkClient{}
			if _, err := UploadWithKeyRetry(tt.args.value, tt.args.config, tt.args.param,
				"traceID"); (err != nil) != tt.wantErr {
				t.Errorf("UploadWithKeyRetry() error = %v, wantErr %v", err, tt.wantErr)
			}
			patches.ResetAll()
		})
	}
	clientMap = concurrentMap{mp: make(map[string]*nodeIP2ClientMap)}
	dataSystemCache.Delete(noCluster)
}

func TestUploadWithKeyRetry(t *testing.T) {
	cache := &Cache{
		nodeList:   []string{},
		invalidMap: make(map[string]struct{}, 1),
	}
	cache.addNode("127.2.2.101", log.GetLogger())
	cache.addNode("127.2.2.102", log.GetLogger())
	dataSystemCache.Store(noCluster, cache)
	type args struct {
		deviceID              string
		value                 []byte
		config                *Config
		param                 api.SetParam
		localClientLibruntime api.LibruntimeAPI
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		patchesFunc mockUtils.PatchesFunc
	}{
		{"case1 succeed to upload key", args{deviceID: "key1", value: []byte("value1"),
			config:                &Config{TenantID: "tenant1", NodeIP: "127.2.2.101"},
			param:                 api.SetParam{WriteMode: api.NoneL2Cache, TTLSecond: 60 * 1000},
			localClientLibruntime: &mockUtils.FakeLibruntimeSdkClient{}},
			false,
			func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					gomonkey.ApplyFunc(NewClient,
						func(tenantID string, nodeIP string) (DsClientImpl, error) {
							return DsClientImpl{kvClient: &FakeKvClient{}}, nil
						}),
					gomonkey.ApplyFunc((*Cache).healthCheckProcess,
						func(_ *Cache, node string) {
							return
						}),
				})
				return patches
			}},

		{"case2 succeed to upload key when node ip is null", args{deviceID: "key1", value: []byte("value1"),
			config:                &Config{TenantID: "tenant1", NodeIP: ""},
			param:                 api.SetParam{WriteMode: api.NoneL2Cache, TTLSecond: 60 * 1000},
			localClientLibruntime: &mockUtils.FakeLibruntimeSdkClient{}},
			false, func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					gomonkey.ApplyFunc(NewClient,
						func(tenantID string, nodeIP string) (DsClientImpl, error) {
							return DsClientImpl{kvClient: &FakeKvClient{}}, nil
						}),
					gomonkey.ApplyFunc((*Cache).healthCheckProcess,
						func(_ *Cache, node string) {
							return
						}),
				})
				return patches
			}},
		{"case3 failed to upload key when node ip and tenantID is null", args{deviceID: "key1", value: []byte("value1"),
			config:                &Config{TenantID: "", NodeIP: ""},
			param:                 api.SetParam{WriteMode: api.NoneL2Cache, TTLSecond: 60 * 1000},
			localClientLibruntime: &mockUtils.FakeLibruntimeSdkClient{}}, true, func() mockUtils.PatchSlice {
			patches := mockUtils.InitPatchSlice()
			patches.Append(mockUtils.PatchSlice{
				gomonkey.ApplyFunc(NewClient,
					func(tenantID string, nodeIP string) (DsClientImpl, error) {
						return DsClientImpl{}, errors.New("failed to upload")
					}),
				gomonkey.ApplyFunc((*Cache).healthCheckProcess,
					func(_ *Cache, node string) {
						return
					}),
			})
			return patches
		}},
		{"case4 failed to set tenant id", args{deviceID: "key1", value: []byte("value1"),
			config:                &Config{TenantID: "tenant1", NodeIP: "127.2.2.101"},
			param:                 api.SetParam{WriteMode: api.NoneL2Cache, TTLSecond: 60 * 1000},
			localClientLibruntime: &mockUtils.FakeLibruntimeSdkClient{}},
			true,
			func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					gomonkey.ApplyFunc(NewClient,
						func(tenantID string, nodeIP string) (DsClientImpl, error) {
							return DsClientImpl{kvClient: &FakeKvClient{}}, nil
						}),
					gomonkey.ApplyFunc((*Cache).healthCheckProcess,
						func(_ *Cache, node string) {
							return
						}),
					gomonkey.ApplyFunc((*mockUtils.FakeLibruntimeSdkClient).SetTenantID,
						func(_ *mockUtils.FakeLibruntimeSdkClient, tenantID string) error {
							return errors.New("set tenant failed")
						}),
				})
				return patches
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := tt.patchesFunc()
			localClientLibruntime = tt.args.localClientLibruntime
			tt.args.config.KeyPrefix = tt.args.deviceID
			if _, err := UploadWithKeyRetry(tt.args.value, tt.args.config, tt.args.param,
				"traceID"); (err != nil) != tt.wantErr {
				t.Errorf("UploadWithKeyRetry() error = %v, wantErr %v", err, tt.wantErr)
			}
			patches.ResetAll()
		})
	}
	clientMap = concurrentMap{mp: make(map[string]*nodeIP2ClientMap)}
	dataSystemCache.Delete(noCluster)
}

func TestDownloadArrayRetryKvClient(t *testing.T) {
	cache := &Cache{
		nodeList:   []string{},
		invalidMap: make(map[string]struct{}, 1),
	}
	cache.addNode("127.2.2.101", log.GetLogger())
	cache.addNode("127.2.2.102", log.GetLogger())
	dataSystemCache.Store(noCluster, cache)
	type args struct {
		keys         []string
		config       *Config
		sourceClient DsClientImpl
	}
	tests := []struct {
		name        string
		args        args
		want        [][]byte
		wantErr     bool
		patchesFunc mockUtils.PatchesFunc
	}{
		{"case1 succeed to download array", args{keys: []string{"key1", "key2"},
			config:       &Config{TenantID: "tenant1", NodeIP: "127.2.2.101"},
			sourceClient: DsClientImpl{kvClient: &FakeKvClient{1}}}, [][]byte{[]byte("value1"), []byte("value2")},
			false,
			func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					gomonkey.ApplyFunc(NewClient,
						func(tenantID string, nodeIP string) (DsClientImpl, error) {
							return DsClientImpl{kvClient: &FakeKvClient{1}}, nil
						}),
					gomonkey.ApplyFunc((*Cache).healthCheckProcess,
						func(_ *Cache, node string) {
							return
						}),
				})
				return patches
			}},
		{"case2 failed to query size", args{keys: []string{"key1", "key2", "key3", "key4"},
			config:       &Config{TenantID: "tenant1", NodeIP: "127.2.2.101"},
			sourceClient: DsClientImpl{kvClient: &FakeKvClient{1}}}, nil, true,
			func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					gomonkey.ApplyFunc(NewClient,
						func(tenantID string, nodeIP string) (DsClientImpl, error) {
							return DsClientImpl{kvClient: &FakeKvClient{1}}, nil
						}),
					gomonkey.ApplyFunc((*Cache).healthCheckProcess,
						func(_ *Cache, node string) {
							return
						}),
				})
				return patches
			}},
		{"case3  failed to download array node ip and tenantID is null", args{keys: []string{"key3"},
			config:       &Config{TenantID: "", NodeIP: ""},
			sourceClient: DsClientImpl{kvClient: &FakeKvClient{1}}}, nil, true, func() mockUtils.PatchSlice {
			patches := mockUtils.InitPatchSlice()
			patches.Append(mockUtils.PatchSlice{
				gomonkey.ApplyFunc(NewClient,
					func(tenantID string, nodeIP string) (DsClientImpl, error) {
						return DsClientImpl{}, errors.New("failed to download")
					}),
				gomonkey.ApplyFunc((*Cache).healthCheckProcess,
					func(_ *Cache, node string) {
						return
					}),
			})
			return patches
		}},
		{"case4 errKeyNotFound", args{keys: []string{"key1", "key2", "key3"},
			config:       &Config{TenantID: "tenant1", NodeIP: "127.2.2.101"},
			sourceClient: DsClientImpl{kvClient: &FakeKvClient{1}}}, nil, true, func() mockUtils.PatchSlice {
			patches := mockUtils.InitPatchSlice()
			patches.Append(mockUtils.PatchSlice{
				gomonkey.ApplyFunc(NewClient,
					func(tenantID string, nodeIP string) (DsClientImpl, error) {
						return DsClientImpl{}, errors.New("failed to download")
					}),
				gomonkey.ApplyFunc((*Cache).healthCheckProcess,
					func(_ *Cache, node string) {
						return
					}),
			})
			return patches
		}},
		{"case5 failed to query in limit", args{keys: []string{"key1", "key2", "key3", "key4", "key5"},
			config:       &Config{TenantID: "tenant1", NodeIP: "127.2.2.101", Limit: 2},
			sourceClient: DsClientImpl{kvClient: &FakeKvClient{1}}}, nil, true,
			func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					gomonkey.ApplyFunc(NewClient,
						func(tenantID string, nodeIP string) (DsClientImpl, error) {
							return DsClientImpl{kvClient: &FakeKvClient{1}}, nil
						}),
					gomonkey.ApplyFunc((*Cache).healthCheckProcess,
						func(_ *Cache, node string) {
							return
						}),
				})
				return patches
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := tt.patchesFunc()
			localClientLibruntime = &mockUtils.FakeLibruntimeSdkClient{}
			got, err := DownloadArrayRetry(tt.args.keys, tt.args.config, "traceID")
			if (err != nil) != tt.wantErr {
				t.Errorf("DownloadArrayRetry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DownloadArrayRetry() got = %v, want %v", got, tt.want)
			}
			patches.ResetAll()
		})
	}
	clientMap = concurrentMap{mp: make(map[string]*nodeIP2ClientMap)}
	dataSystemCache.Delete(noCluster)
}

func Test_getDataSystemKey(t *testing.T) {
	convey.Convey("getDataSystemKey ok", t, func() {
		config := &Config{}
		config.NoNeedGenKey = true
		config.KeyPrefix = "/dt"
		key, genKey, err := getDataSystemKey(config, DsClientImpl{}, "aaaa")
		convey.So(key, convey.ShouldNotBeNil)
		convey.So(genKey, convey.ShouldEqual, "")
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("getDataSystemKey ok 1", t, func() {
		p := gomonkey.ApplyFunc(getClient, func(cfg *Config, _ string) (DsClientImpl, bool, error) {
			return DsClientImpl{kvClient: &FakeKvClient{num: 1}}, false, nil
		})
		defer p.Reset()
		config := &Config{}
		config.NoNeedGenKey = false
		config.KeyPrefix = "aaa"
		dsClient, _, _ := getClient(config, "")
		_, _, err := getDataSystemKey(config, dsClient, "aaaa")
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("getDataSystemKey ok 2", t, func() {
		p := gomonkey.ApplyFunc(getClient, func(cfg *Config) (DsClientImpl, bool, error) {
			return DsClientImpl{kvClient: &FakeKvClient{num: 1}}, false, nil
		})
		defer p.Reset()
		config := &Config{}
		config.NoNeedGenKey = false
		config.KeyPrefix = "aaa"
		config.useLastUsedNode = true
		dsClient, _, _ := getClient(config, "")
		_, _, err := getDataSystemKey(config, dsClient, "aaaa")
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestUploadWithoutKeyRetry(t *testing.T) {
	cache := &Cache{
		nodeList:   []string{},
		invalidMap: make(map[string]struct{}, 1),
	}
	cache.addNode("127.2.2.101", log.GetLogger())
	cache.addNode("127.2.2.102", log.GetLogger())
	dataSystemCache.Store(noCluster, cache)
	type args struct {
		value                 []byte
		config                *Config
		param                 api.SetParam
		localClientLibruntime api.LibruntimeAPI
	}
	tests := []struct {
		name        string
		args        args
		want        string
		wantErr     bool
		patchesFunc mockUtils.PatchesFunc
	}{
		{"case1 succeed to upload without key", args{value: []byte("value1"),
			config:                &Config{TenantID: "tenant1", NodeIP: "127.2.2.101"},
			param:                 api.SetParam{WriteMode: api.NoneL2Cache, TTLSecond: 60 * 1000},
			localClientLibruntime: &mockUtils.FakeLibruntimeSdkClient{}},
			"",
			false,
			func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					gomonkey.ApplyFunc(NewClient,
						func(tenantID string, nodeIP string) (DsClientImpl, error) {
							return DsClientImpl{kvClient: &FakeKvClient{}}, nil
						}),
					gomonkey.ApplyFunc((*Cache).healthCheckProcess,
						func(_ *Cache, node string) {
							return
						}),
				})
				return patches
			}},
		{"case2 failed to upload without key node ip and tenantID is null", args{value: []byte("value2"),
			config:                &Config{TenantID: "", NodeIP: ""},
			param:                 api.SetParam{WriteMode: api.NoneL2Cache, TTLSecond: 60 * 1000},
			localClientLibruntime: &mockUtils.FakeLibruntimeSdkClient{}}, "", true, func() mockUtils.PatchSlice {
			patches := mockUtils.InitPatchSlice()
			patches.Append(mockUtils.PatchSlice{
				gomonkey.ApplyFunc(NewClient,
					func(tenantID string, nodeIP string) (DsClientImpl, error) {
						return DsClientImpl{}, errors.New("failed to upload")
					}),
				gomonkey.ApplyFunc((*Cache).healthCheckProcess,
					func(_ *Cache, node string) {
						return
					}),
			})
			return patches
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := tt.patchesFunc()
			localClientLibruntime = tt.args.localClientLibruntime
			got, err := UploadWithoutKeyRetry(tt.args.value, tt.args.config, tt.args.param, "traceID")
			if (err != nil) != tt.wantErr {
				t.Errorf("UploadWithoutKeyRetry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UploadWithoutKeyRetry() got = %v, want %v", got, tt.want)
			}
			patches.ResetAll()
		})
	}
	clientMap = concurrentMap{mp: make(map[string]*nodeIP2ClientMap)}
	dataSystemCache.Delete(noCluster)
}

func TestDeleteArrayRetryKvClient(t *testing.T) {
	cache := &Cache{
		nodeList:   []string{},
		invalidMap: make(map[string]struct{}, 1),
	}
	cache.addNode("127.2.2.101", log.GetLogger())
	cache.addNode("127.2.2.102", log.GetLogger())
	dataSystemCache.Store(noCluster, cache)
	type args struct {
		keys         []string
		config       *Config
		sourceClient DsClientImpl
	}
	tests := []struct {
		name        string
		args        args
		want        []string
		wantErr     bool
		patchesFunc mockUtils.PatchesFunc
	}{
		{"case1 succeed to download array", args{keys: []string{"key1", "key2"},
			config:       &Config{TenantID: "tenant1", NodeIP: "127.2.2.101"},
			sourceClient: DsClientImpl{kvClient: &FakeKvClient{1}}}, nil, false, func() mockUtils.PatchSlice {
			patches := mockUtils.InitPatchSlice()
			patches.Append(mockUtils.PatchSlice{
				gomonkey.ApplyFunc(NewClient,
					func(tenantID string, nodeIP string) (DsClientImpl, error) {
						return DsClientImpl{kvClient: &FakeKvClient{num: 1}}, nil
					}),
				gomonkey.ApplyFunc((*Cache).healthCheckProcess,
					func(_ *Cache, node string) {
						return
					}),
			})
			return patches
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := tt.patchesFunc()
			got, err := DeleteArrayRetry(tt.args.keys, tt.args.config, "traceID")
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteArrayRetry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeleteArrayRetry() got = %v, want %v", got, tt.want)
			}
			patches.ResetAll()
		})
	}
	clientMap = concurrentMap{mp: make(map[string]*nodeIP2ClientMap)}
	dataSystemCache.Delete(noCluster)
}

func TestDeleteArrayRetry(t *testing.T) {
	cache := &Cache{
		nodeList:   []string{},
		invalidMap: make(map[string]struct{}, 1),
	}
	cache.addNode("127.2.2.101", log.GetLogger())
	cache.addNode("127.2.2.102", log.GetLogger())
	dataSystemCache.Store(noCluster, cache)
	type args struct {
		keys         []string
		config       *Config
		sourceClient DsClientImpl
	}
	tests := []struct {
		name        string
		args        args
		want        []string
		wantErr     bool
		patchesFunc mockUtils.PatchesFunc
	}{
		{"case1 succeed to download array", args{keys: []string{"key1", "key2"},
			config:       &Config{TenantID: "tenant1", NodeIP: "127.2.2.101"},
			sourceClient: DsClientImpl{kvClient: &FakeKvClient{}}}, nil, false, func() mockUtils.PatchSlice {
			patches := mockUtils.InitPatchSlice()
			patches.Append(mockUtils.PatchSlice{
				gomonkey.ApplyFunc(NewClient,
					func(tenantID string, nodeIP string) (DsClientImpl, error) {
						return DsClientImpl{kvClient: &FakeKvClient{num: 1}}, nil
					}),
				gomonkey.ApplyFunc((*Cache).healthCheckProcess,
					func(_ *Cache, node string) {
						return
					}),
			})
			return patches
		}},
		{"case2  failed to download array after retry", args{keys: []string{"key3"},
			config:       &Config{TenantID: "tenant1", NodeIP: "127.2.2.102"},
			sourceClient: DsClientImpl{kvClient: &FakeKvClient{}}}, []string{"key3"}, true, func() mockUtils.PatchSlice {
			patches := mockUtils.InitPatchSlice()
			patches.Append(mockUtils.PatchSlice{
				gomonkey.ApplyFunc(NewClient,
					func(tenantID string, nodeIP string) (DsClientImpl, error) {
						return DsClientImpl{}, errors.New("failed to delete")
					}),
				gomonkey.ApplyFunc((*Cache).healthCheckProcess,
					func(_ *Cache, node string) {
						return
					}),
			})
			return patches
		}},
		{"case3  failed to download node ip and tenantID is null", args{keys: []string{"key3"},
			config:       &Config{TenantID: "", NodeIP: ""},
			sourceClient: DsClientImpl{kvClient: &FakeKvClient{}}}, []string{"key3"}, true, func() mockUtils.PatchSlice {
			patches := mockUtils.InitPatchSlice()
			patches.Append(mockUtils.PatchSlice{
				gomonkey.ApplyFunc(NewClient,
					func(tenantID string, nodeIP string) (DsClientImpl, error) {
						return DsClientImpl{}, errors.New("failed to delete")
					}),
				gomonkey.ApplyFunc((*Cache).healthCheckProcess,
					func(_ *Cache, node string) {
						return
					}),
			})
			return patches
		}},
		{
			"case4 len key is 0", args{keys: []string{},
				config:       &Config{TenantID: "", NodeIP: ""},
				sourceClient: DsClientImpl{kvClient: &FakeKvClient{}}}, []string{}, true, func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					gomonkey.ApplyFunc(NewClient,
						func(tenantID string, nodeIP string) (DsClientImpl, error) {
							return DsClientImpl{}, errors.New("failed to delete")
						}),
					gomonkey.ApplyFunc((*Cache).healthCheckProcess,
						func(_ *Cache, node string) {
							return
						}),
				})
				return patches
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := tt.patchesFunc()
			got, err := DeleteArrayRetry(tt.args.keys, tt.args.config, "traceID")
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteArrayRetry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeleteArrayRetry() got = %v, want %v", got, tt.want)
			}
			patches.ResetAll()
		})
	}
	clientMap = concurrentMap{mp: make(map[string]*nodeIP2ClientMap)}
	dataSystemCache.Delete(noCluster)
}

func TestDeleteClient(t *testing.T) {
	cm := concurrentMap{mp: make(map[string]*nodeIP2ClientMap)}
	var sourceClient = &FakeKvClient{}
	cm.mp["tenant1"] = &nodeIP2ClientMap{
		clientMap: map[string]DsClientImpl{
			"127.2.2.101": {kvClient: sourceClient},
			"127.2.2.102": {kvClient: sourceClient},
		},
		RWMutex: sync.RWMutex{},
		logger:  log.GetLogger(),
	}
	convey.Convey("delete tenantID & nodeIP", t, func() {
		cm.deleteClient("tenant1", "127.2.2.101")
		client, ok := cm.get("tenant1", "127.2.2.101")
		convey.So(client.kvClient, convey.ShouldBeNil)
		convey.So(ok, convey.ShouldBeFalse)
	})
	convey.Convey("delete tenantID", t, func() {
		cm.deleteTenant("tenant1")
		client := cm.getOneClient("tenant1")
		convey.So(client, convey.ShouldBeNil)
	})
	convey.Convey("delete nodeIp", t, func() {
		cm.mp["tenant1"] = &nodeIP2ClientMap{
			clientMap: map[string]DsClientImpl{
				"127.2.2.101": {kvClient: sourceClient},
				"127.2.2.102": {kvClient: sourceClient},
			},
			RWMutex: sync.RWMutex{},
			logger:  log.GetLogger(),
		}
		cm.mp["tenant2"] = &nodeIP2ClientMap{
			clientMap: map[string]DsClientImpl{
				"127.2.2.101": {kvClient: sourceClient},
				"127.2.2.102": {kvClient: sourceClient},
			},
			RWMutex: sync.RWMutex{},
			logger:  log.GetLogger(),
		}
		cm.mp["tenant3"] = &nodeIP2ClientMap{
			clientMap: map[string]DsClientImpl{
				"127.2.2.101": {kvClient: sourceClient},
			},
			RWMutex: sync.RWMutex{},
			logger:  log.GetLogger(),
		}
		cm.deleteNodeIp("127.2.2.101", log.GetLogger())
		convey.So(cm.mp["tenant1"].size(), convey.ShouldEqual, 1)
		convey.So(cm.mp["tenant2"].size(), convey.ShouldEqual, 1)
		_, ok := cm.mp["tenant3"]
		convey.So(ok, convey.ShouldEqual, false)
	})
}

func Test_uploadWithKeyKvClient(t *testing.T) {
	convey.Convey("uploadWithKey test", t, func() {
		convey.Convey("test encrypt", func() {
			p := gomonkey.ApplyFunc(getClient, func(cfg *Config, _ string) (DsClientImpl, bool, error) {
				return DsClientImpl{kvClient: &FakeKvClient{num: 1}}, false, nil
			})
			defer p.Reset()
			key, b, err := uploadWithKey([]byte("value"), &Config{NeedEncrypt: true, KeyPrefix: "aaa"}, api.SetParam{},
				"")
			convey.So(key, convey.ShouldEqual, "1")
			convey.So(b, convey.ShouldBeFalse)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("test set tenantID fail", func() {
			patches := []*gomonkey.Patches{
				gomonkey.ApplyFunc(getClient, func(cfg *Config, _ string) (DsClientImpl, bool, error) {
					return DsClientImpl{kvClient: &FakeKvClient{num: 1}}, false, nil
				}),
				gomonkey.ApplyFunc((*mockUtils.FakeLibruntimeSdkClient).SetTenantID,
					func(_ *mockUtils.FakeLibruntimeSdkClient, tenantID string) error {
						return errors.New("set tenant failed")
					}),
			}
			for _, patch := range patches {
				patch.Reset()
			}

			key, b, err := uploadWithKey([]byte("value"), &Config{NeedEncrypt: false}, api.SetParam{}, "")
			convey.So(key, convey.ShouldEqual, "")
			convey.So(b, convey.ShouldBeFalse)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func Test_uploadWithKey(t *testing.T) {
	type args struct {
		deviceID string
		value    []byte
		config   *Config
		param    api.SetParam
		traceID  string
	}
	tests := []struct {
		name        string
		args        args
		want1       bool
		wantErr     bool
		patchesFunc mockUtils.PatchesFunc
	}{
		{"case1 dsClient is nil", args{config: &Config{}}, true, true, func() mockUtils.PatchSlice {
			patches := mockUtils.InitPatchSlice()
			patches.Append(mockUtils.PatchSlice{
				gomonkey.ApplyFunc(getClient,
					func(cfg *Config, _ string) (DsClientImpl, bool, error) { return DsClientImpl{}, true, errors.New("e") }),
			})
			return patches
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := tt.patchesFunc()
			_, got1, err := uploadWithKey(tt.args.value, tt.args.config, tt.args.param, tt.args.traceID)
			if (err != nil) != tt.wantErr {
				t.Errorf("uploadWithKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got1 != tt.want1 {
				t.Errorf("uploadWithKey() got1 = %v, want %v", got1, tt.want1)
			}
			patches.ResetAll()
		})
	}
}

type invokerLibruntimeMock struct {
	setTenantIDSuccessfully bool
}

func (c *invokerLibruntimeMock) CreateInstance(funcMeta api.FunctionMeta, args []api.Arg,
	invokeOpt api.InvokeOptions) (instanceID string, err error) {
	return "", nil
}

func (c *invokerLibruntimeMock) InvokeByInstanceId(funcMeta api.FunctionMeta, instanceID string, args []api.Arg,
	invokeOpt api.InvokeOptions) (returnObjectID string, err error) {
	return "", nil
}

func (c *invokerLibruntimeMock) InvokeByFunctionName(funcMeta api.FunctionMeta, args []api.Arg,
	invokeOpt api.InvokeOptions) (returnObjectID string, err error) {
	return "", nil
}

func (c *invokerLibruntimeMock) AcquireInstance(state string, funcMeta api.FunctionMeta, acquireOpt api.InvokeOptions) (api.InstanceAllocation, error) {
	return api.InstanceAllocation{}, nil
}

func (c *invokerLibruntimeMock) ReleaseInstance(allocation api.InstanceAllocation, stateID string, abnormal bool, option api.InvokeOptions) {

}

func (c *invokerLibruntimeMock) Kill(instanceID string, signal int, payload []byte) (err error) {
	return nil
}

func (c *invokerLibruntimeMock) CreateInstanceRaw(createReqRaw []byte) (createRespRaw []byte, err error) {
	return []byte{}, nil
}

func (c *invokerLibruntimeMock) InvokeByInstanceIdRaw(invokeReqRaw []byte) (resultRaw []byte, err error) {
	return []byte{}, nil
}

func (f *invokerLibruntimeMock) KillRaw(killReqRaw []byte) (killRespRaw []byte, err error) {
	return []byte{}, nil
}

func (f *invokerLibruntimeMock) SaveState(state []byte) (stateID string, err error) {
	return "", nil
}

func (f *invokerLibruntimeMock) LoadState(checkpointID string) (state []byte, err error) {
	return []byte{}, nil
}

func (f *invokerLibruntimeMock) Exit(code int, message string) {
	return
}

func (f *invokerLibruntimeMock) Finalize() {
	return
}

func (f *invokerLibruntimeMock) KVSet(key string, value []byte, param api.SetParam) (err error) {
	return nil
}

func (f *invokerLibruntimeMock) KVSetWithoutKey(value []byte, param api.SetParam) (key string, err error) {
	return "", nil
}

func (f *invokerLibruntimeMock) KVMSetTx(keys []string, values [][]byte, param api.MSetParam) error {
	return nil
}

func (f *invokerLibruntimeMock) KVGet(key string, timeoutms uint) (value []byte, err error) {
	return []byte{}, nil
}

func (f *invokerLibruntimeMock) KVGetMulti(keys []string, timeoutms uint) (values [][]byte, err error) {
	return [][]byte{}, nil
}

func (f *invokerLibruntimeMock) KVDel(key string) (err error) {
	return nil
}

func (f *invokerLibruntimeMock) KVDelMulti(keys []string) (failedKeys []string, err error) {
	return []string{}, nil
}

func (f *invokerLibruntimeMock) QueryGlobalProducersNum(streamName string) (uint64, error) {
	return 0, nil
}

func (f *invokerLibruntimeMock) QueryGlobalConsumersNum(streamName string) (uint64, error) {
	return 0, nil
}

func (f *invokerLibruntimeMock) Wait(objectIDs []string, waitNum uint64, timeoutMs int) (readyIDs, unReadyIDs []string, errors map[string]error) {
	return []string{}, []string{}, make(map[string]error)
}

func (f *invokerLibruntimeMock) SetTraceID(traceID string) {
	return
}

func (f *invokerLibruntimeMock) SetTenantID(tenantID string) error {
	if f.setTenantIDSuccessfully {
		return nil
	}
	return api.ErrorInfo{Code: 1001, Err: errors.New("failed to set tenant id")}
}

func (f *invokerLibruntimeMock) Put(objectID string, value []byte, param api.PutParam,
	nestedObjectIDs ...string) (err error) {
	return nil
}

func (f *invokerLibruntimeMock) Get(objectIDs []string, timeoutMs int) (data [][]byte, err error) {
	return [][]byte{}, nil
}

func (f *invokerLibruntimeMock) GIncreaseRef(objectIDs []string, remoteClientID ...string) (failedIDs []string, err error) {
	return []string{}, nil
}

func (f *invokerLibruntimeMock) GDecreaseRef(objectIDs []string, remoteClientID ...string) (failedIDs []string, err error) {
	return []string{}, nil
}

func (f *invokerLibruntimeMock) GetAsync(objectID string, cb api.GetAsyncCallback) {
	return
}

func (f *invokerLibruntimeMock) GetFormatLogger() api.FormatLogger {
	return nil
}

func (c *invokerLibruntimeMock) CreateProducer(streamName string, producerConf api.ProducerConf) (producer api.StreamProducer, err error) {
	return producer, err
}

func (c *invokerLibruntimeMock) Subscribe(streamName string, config api.SubscriptionConfig) (consumer api.StreamConsumer, err error) {
	return consumer, nil
}

func (c *invokerLibruntimeMock) DeleteStream(streamName string) (err error) {
	return nil
}

func (c *invokerLibruntimeMock) CreateClient(config api.ConnectArguments) (api.KvClient, error) {
	return &FakeKvClient{1}, nil
}

func (l *invokerLibruntimeMock) GIncreaseRefRaw(objectIDs []string, remoteClientID ...string) ([]string, error) {
	return []string{}, nil
}

func (l *invokerLibruntimeMock) PutRaw(objectID string, value []byte, param api.PutParam, nestedObjectIDs ...string) error {
	return nil
}

func (l *invokerLibruntimeMock) GetRaw(objectIDs []string, timeoutMs int) ([][]byte, error) {
	return [][]byte{}, nil
}

func (l *invokerLibruntimeMock) GDecreaseRefRaw(objectIDs []string, remoteClientID ...string) ([]string, error) {
	return []string{}, nil
}
func (l *invokerLibruntimeMock) GetCredential() api.Credential {
	return api.Credential{}
}

func (l *invokerLibruntimeMock) UpdateSchdulerInfo(schedulerName string, schedulerId string, option string) {
	return
}

func (f *invokerLibruntimeMock) IsHealth() bool {
	return true
}

func (f *invokerLibruntimeMock) IsDsHealth() bool {
	return true
}

func TestKVGet(t *testing.T) {
	convey.Convey("TestKVGetLibruntime", t, func() {
		keyNotFound := 1
		mock := &invokerLibruntimeMock{setTenantIDSuccessfully: true}
		kvStore := map[string][]byte{}
		testKey := "test_key"
		testValue := []byte{'1', '1', '1'}
		patches := []*gomonkey.Patches{
			gomonkey.ApplyMethod(reflect.TypeOf(mock), "KVGetMulti", func(_ *invokerLibruntimeMock, keys []string,
				timeoutms uint) ([][]byte, error) {
				values := [][]byte{}
				for _, key := range keys {
					if value, ok := kvStore[key]; ok {
						values = append(values, value)
						continue
					}
					return values, api.ErrorInfo{Code: keyNotFound, Err: fmt.Errorf("key %s not found", key)}
				}
				return values, nil
			}),
			gomonkey.ApplyMethod(reflect.TypeOf(mock), "KVSet", func(_ *invokerLibruntimeMock, key string,
				value []byte, param api.SetParam) error {
				kvStore[key] = value
				return nil
			}),
		}
		defer func() {
			for _, patch := range patches {
				patch.Reset()
			}
		}()

		setLocalClient(mock)
		setReq := data.KvSetRequest{
			Key:   testKey,
			Value: testValue,
		}
		err := Set(&setReq, &Config{}, "")
		convey.So(err.Code, convey.ShouldEqual, 0)
		convey.So(err.Err, convey.ShouldBeNil)

		getReq := data.KvGetRequest{
			Keys: []string{testKey},
		}
		values, status := Get(&getReq, &Config{}, "")
		convey.So(status.Code, convey.ShouldEqual, 0)
		convey.So(status.Err, convey.ShouldBeNil)
		convey.So(string(values[0]), convey.ShouldEqual, string(testValue))
	})
}

func TestObjPut(t *testing.T) {
	convey.Convey("ObjPut", t, func() {
		convey.Convey("dsclient is nil", func() {
			localClientLibruntime = nil
			put := ObjPut(&data.PutRequest{}, &Config{}, "test-trace-ID")
			convey.So(put.Code, convey.ShouldEqual, errRPCUnavailable)
		})
		convey.Convey("success", func() {
			mock := &invokerLibruntimeMock{setTenantIDSuccessfully: true}
			setLocalClient(mock)
			put := ObjPut(&data.PutRequest{
				ObjectId:        "",
				ObjectData:      nil,
				WriteMode:       int32(os.O_WRONLY),
				ConsistencyType: 1,
				NestedObjectIds: nil,
			}, &Config{}, "test-trace-ID")
			convey.So(put.Err, convey.ShouldBeNil)
		})

		convey.Convey("put failed", func() {
			failedMock := &invokerLibruntimeMock{setTenantIDSuccessfully: false}
			setLocalClient(failedMock)
			put := ObjPut(&data.PutRequest{}, &Config{}, "test-trace-ID")
			convey.So(put.Code, convey.ShouldEqual, 1001)
		})
	})
}

func TestObjGet(t *testing.T) {
	convey.Convey("ObjGet", t, func() {
		convey.Convey("dsclient is nil", func() {
			localClientLibruntime = nil
			_, put := ObjGet(&data.GetRequest{}, &Config{}, "test-trace-ID")
			convey.So(put.Code, convey.ShouldEqual, errRPCUnavailable)
		})
		convey.Convey("success", func() {
			localClientLibruntime = &invokerLibruntimeMock{setTenantIDSuccessfully: true}
			values, put := ObjGet(&data.GetRequest{
				ObjectIds: nil,
				TimeoutMs: 0,
			}, &Config{}, "test-trace-ID")
			convey.So(put.Err, convey.ShouldBeNil)
			convey.So(len(values), convey.ShouldEqual, 0)
		})
	})
}

func TestGIncreaseRef(t *testing.T) {
	convey.Convey("GIncreaseRef", t, func() {
		convey.Convey("dsclient is nil", func() {
			localClientLibruntime = nil
			_, put := GIncreaseRef(&data.IncreaseRefRequest{}, &Config{}, "test-trace-ID")
			convey.So(put.Code, convey.ShouldEqual, errRPCUnavailable)
		})
		convey.Convey("success", func() {
			localClientLibruntime = &invokerLibruntimeMock{setTenantIDSuccessfully: true}
			_, status := GIncreaseRef(&data.IncreaseRefRequest{}, &Config{}, "test-trace-ID")
			convey.So(status.Code, convey.ShouldEqual, 0)
		})
		convey.Convey("failed", func() {
			localClientLibruntime = &invokerLibruntimeMock{setTenantIDSuccessfully: false}
			_, status := GIncreaseRef(&data.IncreaseRefRequest{}, &Config{}, "test-trace-ID")
			convey.So(status.Code, convey.ShouldNotEqual, 0)
		})
	})
}

func TestGDecreaseRef(t *testing.T) {
	convey.Convey("GDecreaseRef", t, func() {
		convey.Convey("dsclient is nil", func() {
			localClientLibruntime = nil
			_, put := GDecreaseRef(&data.DecreaseRefRequest{}, &Config{}, "test-trace-ID")
			convey.So(put.Code, convey.ShouldEqual, errRPCUnavailable)
		})
		convey.Convey("success", func() {
			localClientLibruntime = &invokerLibruntimeMock{setTenantIDSuccessfully: true}
			_, status := GDecreaseRef(&data.DecreaseRefRequest{}, &Config{}, "test-trace-ID")
			convey.So(status.Code, convey.ShouldEqual, 0)
		})
		convey.Convey("failed", func() {
			localClientLibruntime = &invokerLibruntimeMock{setTenantIDSuccessfully: false}
			_, status := GDecreaseRef(&data.DecreaseRefRequest{}, &Config{}, "test-trace-ID")
			convey.So(status.Code, convey.ShouldNotEqual, 0)
		})
	})
}

func TestSet(t *testing.T) {
	convey.Convey("Set", t, func() {
		convey.Convey("dsclient is nil", func() {
			localClientLibruntime = nil
			put := Set(&data.KvSetRequest{}, &Config{}, "test-trace-ID")
			convey.So(put.Code, convey.ShouldEqual, errRPCUnavailable)
		})
		convey.Convey("success", func() {
			localClientLibruntime = &invokerLibruntimeMock{setTenantIDSuccessfully: true}
			status := Set(&data.KvSetRequest{}, &Config{}, "test-trace-ID")
			convey.So(status.Code, convey.ShouldEqual, 0)
		})
		convey.Convey("failed", func() {
			localClientLibruntime = &invokerLibruntimeMock{setTenantIDSuccessfully: false}
			status := Set(&data.KvSetRequest{}, &Config{}, "test-trace-ID")
			convey.So(status.Code, convey.ShouldNotEqual, 0)
		})
	})
}

func TestMSetTx(t *testing.T) {
	convey.Convey("MSetTx", t, func() {
		convey.Convey("dsclient is nil", func() {
			localClientLibruntime = nil
			put := MSetTx(&data.KvMSetTxRequest{}, &Config{}, "test-trace-ID")
			convey.So(put.Code, convey.ShouldEqual, errRPCUnavailable)
		})
		convey.Convey("success", func() {
			localClientLibruntime = &invokerLibruntimeMock{setTenantIDSuccessfully: true}
			status := MSetTx(&data.KvMSetTxRequest{}, &Config{}, "test-trace-ID")
			convey.So(status.Code, convey.ShouldEqual, 0)
		})
		convey.Convey("failed", func() {
			localClientLibruntime = &invokerLibruntimeMock{setTenantIDSuccessfully: false}
			status := MSetTx(&data.KvMSetTxRequest{}, &Config{}, "test-trace-ID")
			convey.So(status.Code, convey.ShouldNotEqual, 0)
		})
	})
}

func TestGet(t *testing.T) {
	convey.Convey("Get", t, func() {
		convey.Convey("dsclient is nil", func() {
			localClientLibruntime = nil
			_, put := Get(&data.KvGetRequest{}, &Config{}, "test-trace-ID")
			convey.So(put.Code, convey.ShouldEqual, errRPCUnavailable)
		})
		convey.Convey("success", func() {
			localClientLibruntime = &invokerLibruntimeMock{setTenantIDSuccessfully: true}
			_, status := Get(&data.KvGetRequest{}, &Config{}, "test-trace-ID")
			convey.So(status.Code, convey.ShouldEqual, 0)
		})
		convey.Convey("failed", func() {
			localClientLibruntime = &invokerLibruntimeMock{setTenantIDSuccessfully: false}
			_, status := Get(&data.KvGetRequest{}, &Config{}, "test-trace-ID")
			convey.So(status.Code, convey.ShouldNotEqual, 0)
		})
	})
}

func TestDel(t *testing.T) {
	convey.Convey("Del", t, func() {
		convey.Convey("dsclient is nil", func() {
			localClientLibruntime = nil
			_, put := Del(&data.KvDelRequest{}, &Config{}, "test-trace-ID")
			convey.So(put.Code, convey.ShouldEqual, errRPCUnavailable)
		})

		convey.Convey("success", func() {
			localClientLibruntime = &invokerLibruntimeMock{setTenantIDSuccessfully: true}
			_, status := Del(&data.KvDelRequest{}, &Config{}, "test-trace-ID")
			convey.So(status.Code, convey.ShouldEqual, 0)
		})

		convey.Convey("failed", func() {
			localClientLibruntime = &invokerLibruntimeMock{setTenantIDSuccessfully: false}
			_, status := Del(&data.KvDelRequest{}, &Config{}, "test-trace-ID")
			convey.So(status.Code, convey.ShouldNotEqual, 0)
		})
	})

}

func Test_downloadArray(t *testing.T) {
	convey.Convey("download array test", t, func() {
		localClientLibruntime = &mockUtils.FakeLibruntimeSdkClient{}
		convey.Convey("decrypt failed", func() {
			p := gomonkey.ApplyFunc(getClient, func(cfg *Config, _ string) (DsClientImpl, bool, error) {
				return DsClientImpl{kvClient: &FakeKvClient{}}, false, nil
			})
			defer p.Reset()
			key, b, err := downloadArray([]string{"aaa", "bbb"}, &Config{NeedEncrypt: true, TenantID: "aaaa"}, "")
			convey.So(key, convey.ShouldBeNil)
			convey.So(b, convey.ShouldBeFalse)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("decrypt failed with tenantID dataKey", func() {
			p := gomonkey.ApplyFunc(getClient, func(cfg *Config, _ string) (DsClientImpl, bool, error) {
				return DsClientImpl{kvClient: &FakeKvClient{}}, false, nil
			})
			defer p.Reset()
			key, b, err := downloadArray([]string{"aaa", "bbb"}, &Config{NeedEncrypt: true, TenantID: "aaaa", DataKey: []byte("test")}, "")
			convey.So(key, convey.ShouldBeNil)
			convey.So(b, convey.ShouldBeFalse)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("decrypt ok", func() {
			p := gomonkey.ApplyFunc(getClient, func(cfg *Config, _ string) (DsClientImpl, bool, error) {
				return DsClientImpl{kvClient: &FakeKvClient{}}, false, nil
			})
			defer p.Reset()
			p1 := gomonkey.ApplyFunc(decryptData, func(cfg *Config, data []byte) ([]byte, error) {
				return []byte("hello"), nil
			})
			defer p1.Reset()
			key, b, err := downloadArray([]string{"aaa", "bbb"}, &Config{NeedEncrypt: true, TenantID: "aaaa", Limit: 1000000, DataKey: []byte("test")}, "")
			convey.So(key, convey.ShouldNotBeNil)
			convey.So(b, convey.ShouldBeFalse)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("retry", func() {
			kvclient := &FakeKvClient{}
			p := gomonkey.ApplyFunc(getClient, func(cfg *Config, _ string) (DsClientImpl, bool, error) {
				return DsClientImpl{kvClient: kvclient}, false, nil
			})
			defer p.Reset()
			p1 := gomonkey.ApplyFunc(decryptData, func(cfg *Config, data []byte) ([]byte, error) {
				return []byte("hello"), nil
			})
			defer p1.Reset()

			errCode := 0
			p2 := gomonkey.ApplyMethod(reflect.TypeOf(kvclient), "KVQuerySize", func(_ *FakeKvClient, _ []string) ([]uint64, api.ErrorInfo) {
				return nil, api.ErrorInfo{Code: errCode, Err: fmt.Errorf("do KVQuerySize func failed")}
			})
			defer p2.Reset()
			/*
				errOutOfMemory      = 6
				errDsWorkerNotReady = 8
				errTryAgain      = 19
				errRPCCancelled     = 1000
				errRPCUnavailable   = 1002
				errAsyncQueueFull   = 2003

				errDsClientNil = 11001
			*/
			testCodeMap := []int{6, 8, 19, 1000, 1002, 2003, 11001}
			for _, code := range testCodeMap {
				errCode = code
				_, b, err := downloadArray([]string{"aaa", "bbb"}, &Config{NeedEncrypt: true, TenantID: "aaaa", Limit: 1000000, DataKey: []byte("test")}, "")
				convey.So(b, convey.ShouldBeTrue)
				convey.So(err, convey.ShouldNotBeNil)
			}

		})
	})
}

func Test_SubscribeStream(t *testing.T) {
	convey.Convey("test SubscribeStream simple failed", t, func() {
		localClientLibruntime = &invokerLibruntimeMock{setTenantIDSuccessfully: true}
		p := gomonkey.ApplyMethod(reflect.TypeOf(localClientLibruntime), "Subscribe",
			func(_ *invokerLibruntimeMock, streamName string,
				config api.SubscriptionConfig) (consumer api.StreamConsumer, err error) {
				return nil, fmt.Errorf("just for test")
			})
		defer p.Reset()
		errInfo := SubscribeStream(SubscribeParam{
			StreamName:       "",
			TimeoutMs:        0,
			ExpectReceiveNum: 2,
			TraceId:          "traceId",
		}, &GinCtxAdapter{&gin.Context{Request: &http.Request{}}})
		convey.So(errInfo.Error(), convey.ShouldEqual, "just for test")
	})
}

type MockCloseNotifier struct {
	flushFlag bool
}

// CloseNotify
func (m *MockCloseNotifier) CloseNotify() <-chan bool {
	notify := make(chan bool, 1)
	return notify
}
func (m *MockCloseNotifier) Flush() {
	m.flushFlag = true
}

// MockResponseWriter
type MockResponseWriter struct {
	http.ResponseWriter
	*MockCloseNotifier
}

func Test_receiveStream(t *testing.T) {
	convey.Convey("test SubscribeStream simple failed", t, func() {
		var testData uint8 = 10
		count := 0
		mockConsumer := &mockUtils.FakeStreamConsumer{}
		p := gomonkey.ApplyMethod(reflect.TypeOf(mockConsumer), "ReceiveExpectNum",
			func(_ *mockUtils.FakeStreamConsumer, expectNum uint32, timeoutMs uint32) ([]api.Element, error) {
				if count == 0 {
					count = count + 1
					return []api.Element{{
						Ptr:  &testData,
						Size: 8,
					}}, nil
				} else {
					return []api.Element{}, errors.New("Producer not found")
				}
			})
		defer p.Reset()
		q := gomonkey.ApplyMethod(reflect.TypeOf(&GinCtxAdapter{}), "Done",
			func(_ *GinCtxAdapter) <-chan struct{} {
				done := make(<-chan struct{}, 2)
				return done
			})
		defer q.Reset()
		streamName := "test-Stream"
		timeoutMs := 100
		expectReceiveNum := 2
		consumer := &mockUtils.FakeStreamConsumer{}
		rw := MockResponseWriter{
			ResponseWriter:    httptest.NewRecorder(),
			MockCloseNotifier: &MockCloseNotifier{},
		}
		ctx, _ := gin.CreateTestContext(rw)
		ginCtx := &GinCtxAdapter{
			Context: ctx,
		}
		var called int32
		callback := func() {
			atomic.AddInt32(&called, 1)
		}

		receiveStream(SubscribeParam{
			StreamName:       streamName,
			TimeoutMs:        uint32(timeoutMs),
			ExpectReceiveNum: int32(expectReceiveNum),
			Callback:         callback,
		}, consumer, ginCtx)
		convey.So(rw.flushFlag, convey.ShouldBeTrue)
		convey.So(called, convey.ShouldEqual, 1)
	})
}
