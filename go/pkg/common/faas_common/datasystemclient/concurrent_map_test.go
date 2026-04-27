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
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/types"
	mockUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
)

func TestNewKvClient(t *testing.T) {
	dsConfig := &types.DataSystemConfig{
		TimeoutMs: 60000,
		Clusters:  []string{"AZ1"},
	}
	lease := gomonkey.ApplyFunc(StartWatch,
		func(dataSystemKeyPrefixList []string, stopCh <-chan struct{}) { return })
	InitDataSystemLibruntime(dsConfig, &mockUtils.FakeLibruntimeSdkClient{}, make(chan struct{}))
	type args struct {
		tenantID string
		nodeIP   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"case1 succeed to new a client", args{
			tenantID: "t1",
			nodeIP:   "127.0.0.2",
		}, false}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.args.tenantID, tt.args.nodeIP)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
	time.Sleep(100 * time.Millisecond)
	lease.Reset()
}

func TestNodeIP2ClientMap(t *testing.T) {
	convey.Convey("simple", t, func() {
		n := &nodeIP2ClientMap{
			clientMap: make(map[string]DsClientImpl),
			RWMutex:   sync.RWMutex{},
			logger:    log.GetLogger(),
		}

		n.add("1.1.1.1", DsClientImpl{kvClient: &FakeKvClient{}})
		convey.So(n.size(), convey.ShouldEqual, 1)

		n.add("1.1.1.1", DsClientImpl{kvClient: &FakeKvClient{}})
		convey.So(n.size(), convey.ShouldEqual, 1)

		_, ok := n.get("1.1.1.1")
		convey.So(ok, convey.ShouldBeTrue)

		_, ok = n.get("2.2.2.2")
		convey.So(ok, convey.ShouldBeFalse)

		n.delete("1.1.1.1")
		convey.So(n.size(), convey.ShouldEqual, 0)

		n.add("1.1.1.1", DsClientImpl{kvClient: &FakeKvClient{}})
		n.add("2.1.1.1", DsClientImpl{kvClient: &FakeKvClient{}})
		n.add("3.1.1.1", DsClientImpl{kvClient: &FakeKvClient{}})
		n.add("4.1.1.1", DsClientImpl{kvClient: &FakeKvClient{}})
		_, ok = n.getRandomOne()
		convey.So(ok, convey.ShouldBeTrue)

		n.deleteAll()
		convey.So(n.size(), convey.ShouldEqual, 0)
	})
}
