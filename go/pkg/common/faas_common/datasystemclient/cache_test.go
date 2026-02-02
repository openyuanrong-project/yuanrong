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
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	mockUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func Test_processDataSystemEvent(t *testing.T) {
	dataSystemCache = sync.Map{}
	defer gomonkey.ApplyFunc(destroy, func() {
		return
	}).Reset()
	convey.Convey("add simple event", t, func() {
		// 添加ready的数据系统节点
		event := &etcd3.Event{
			Type:  etcd3.PUT,
			Key:   "/AZ1/datasystem/cluster/8.8.8.8:31501",
			Value: []byte("1748573798753243935;ready"),
		}

		processDataSystemEvent(event)
		cacheRaw, ok := dataSystemCache.Load("AZ1")
		convey.So(ok, convey.ShouldBeTrue)
		cache, ok := cacheRaw.(*Cache)
		convey.So(ok, convey.ShouldBeTrue)

		convey.So(len(cache.nodeList), convey.ShouldEqual, 1)
		convey.So(cache.nodeList[0], convey.ShouldEqual, "8.8.8.8")
		_, ok = cache.invalidMap["8.8.8.8"]
		convey.So(ok, convey.ShouldBeFalse)

		// 添加状态节点异常的数据系统节点
		for _, status := range []string{"start", "restart", "exiting", "sfafdfafd"} {
			event.Value = []byte("1748573798753243935;" + status)
			processDataSystemEvent(event)
			convey.So(len(cache.nodeList), convey.ShouldEqual, 0)
		}

		// 添加状态节点中etcd value格式异常的数据系统节点
		event.Value = []byte("1748573798753243935ready")
		processDataSystemEvent(event)
		convey.So(len(cache.nodeList), convey.ShouldEqual, 0)
	})
	convey.Convey("simple delete event", t, func() {
		event := &etcd3.Event{
			Type:  etcd3.PUT,
			Key:   "/AZ1/datasystem/cluster/8.8.8.8:31501",
			Value: []byte("1748573798753243935;ready"),
		}

		processDataSystemEvent(event)
		event = &etcd3.Event{
			Type:  etcd3.DELETE,
			Key:   "/AZ1/datasystem/cluster/8.8.8.8:31501",
			Value: []byte("1748573798753243935;ready"),
		}
		processDataSystemEvent(event)
		_, ok := dataSystemCache.Load("AZ1")
		convey.So(ok, convey.ShouldBeFalse)

		event.Type = etcd3.PUT
		processDataSystemEvent(event)

		event.Type = etcd3.DELETE
		event.Key = "/AZ1/datasystem/cluster/8.8.8.9:31501"
		processDataSystemEvent(event)

		cacheRaw, ok := dataSystemCache.Load("AZ1")
		convey.So(ok, convey.ShouldBeTrue)
		cache, ok := cacheRaw.(*Cache)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(len(cache.nodeList), convey.ShouldEqual, 1)

		event.Key = "/AZ1/datasystem/cluster/8.8.8.8:31501"
		processDataSystemEvent(event)
		_, ok = dataSystemCache.Load("AZ1")
		convey.So(ok, convey.ShouldBeFalse)
	})
}

func TestCache_healthCheckProcess(t *testing.T) {
	type fields struct {
		nodeList   []string
		invalidMap map[string]struct{}
		lock       sync.RWMutex
	}
	type args struct {
		node string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		patchesFunc mockUtils.PatchesFunc
	}{
		{"case1 node in invalidMap", fields{
			nodeList:   []string{},
			invalidMap: map[string]struct{}{"8.8.8.8": struct{}{}},
			lock:       sync.RWMutex{},
		}, args{node: "8.8.8.8"}, func() mockUtils.PatchSlice {
			patches := mockUtils.InitPatchSlice()
			patches.Append(mockUtils.PatchSlice{
				gomonkey.ApplyFunc(NewClient, func(tenantID string, nodeIP string) (DsClientImpl, error) { return DsClientImpl{}, nil }),
			})
			return patches
		}},
		{"case2 node not in invalidMap", fields{
			nodeList:   []string{},
			invalidMap: map[string]struct{}{},
			lock:       sync.RWMutex{},
		}, args{node: "8.8.8.8"}, func() mockUtils.PatchSlice {
			patches := mockUtils.InitPatchSlice()
			return patches
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := tt.patchesFunc()
			c := &Cache{
				nodeList:   tt.fields.nodeList,
				invalidMap: tt.fields.invalidMap,
				lock:       tt.fields.lock,
			}
			c.healthCheckProcess(tt.args.node)
			patches.ResetAll()
		})
	}
}

func Test_StartWatch(t *testing.T) {
	etcdClient := &etcd3.EtcdClient{
		Client: &clientv3.Client{},
	}
	watchFlag := false
	defer gomonkey.ApplyFunc(etcd3.GetMetaEtcdClient, func() *etcd3.EtcdClient {
		return etcdClient
	}).Reset()
	defer gomonkey.ApplyFunc((*etcd3.EtcdWatcher).StartWatch, func(ew *etcd3.EtcdWatcher) {
		watchFlag = true
	}).Reset()
	stop := make(chan struct{})
	StartWatch([]string{"/datasystem/cluster/"}, stop)
	close(stop)
	assert.True(t, watchFlag)
}

func Test_ParseDsKey(t *testing.T) {
	key := "/cluster001/datasystem/cluster/127.0.0.1:8080"
	ip, az, err := parseDsKey(key)
	assert.Nil(t, err)
	assert.Equal(t, ip, "127.0.0.1")
	assert.Equal(t, az, "cluster001")

	key = "/datasystem/cluster/127.0.0.1:8080"
	ip, az, err = parseDsKey(key)
	assert.Nil(t, err)
	assert.Equal(t, ip, "127.0.0.1")
	assert.Equal(t, az, noCluster)

	key = "/cluster001/datasystem/cluster"
	ip, az, err = parseDsKey(key)
	assert.NotNil(t, err)
	assert.Equal(t, ip, "")
	assert.Equal(t, az, "")

	key = "/cluster001/datasystem/cluster/127.0.0.1"
	ip, az, err = parseDsKey(key)
	assert.NotNil(t, err)
	assert.Equal(t, ip, "")
	assert.Equal(t, az, "")
}
