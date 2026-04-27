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

package functionmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"plugin"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	clientv3 "go.etcd.io/etcd/client/v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/informers"
	testing2 "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/k8sclient"
	commonType "yuanrong.org/kernel/pkg/common/faas_common/types"
	mockUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/functionmanager/registry"
	"yuanrong.org/kernel/pkg/functionmanager/state"
	"yuanrong.org/kernel/pkg/functionmanager/types"
)

type KvMock struct{}

func (k *KvMock) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	// TODO implement me
	return nil, nil
}

func (k *KvMock) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	response := &clientv3.GetResponse{}
	response.Count = 10
	return response, nil
}

func (k *KvMock) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	return nil, fmt.Errorf("error")
}

func (k *KvMock) Compact(ctx context.Context, rev int64, opts ...clientv3.CompactOption) (*clientv3.CompactResponse,
	error,
) {
	// TODO implement me
	panic("implement me")
}

func (k *KvMock) Do(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (k *KvMock) Txn(ctx context.Context) clientv3.Txn {
	// TODO implement me
	panic("implement me")
}

func TestNewFaaSManager(t *testing.T) {
	defer gomonkey.ApplyFunc(state.Update, func(value interface{}, tags ...string) {
	}).Reset()
	defer gomonkey.ApplyFunc(registry.StartWatchEvent, func(vpcEventCh chan types.VPCEvent, stopCh chan struct{}, informer informers.GenericInformer) {
		return
	}).Reset()
	listKinds := map[schema.GroupVersionResource]string{
		// Example: MyResource GVK to MyResourceList GVK
		patGVR: "PatList",
	}
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), listKinds)
	defer gomonkey.ApplyFunc(k8sclient.GetDynamicClient, func() dynamic.Interface {
		return fakeClient
	}).Reset()
	type args struct {
		sdkClient api.LibruntimeAPI
		stopCh    chan struct{}
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		patchesFunc mockUtils.PatchesFunc
	}{
		{"case4 succeed to NewFaaSManager", args{}, false, func() mockUtils.PatchSlice {
			patches := mockUtils.InitPatchSlice()
			patches.Append(mockUtils.PatchSlice{
				gomonkey.ApplyFunc(plugin.Open, func(path string) (*plugin.Plugin, error) {
					return &plugin.Plugin{}, nil
				}),
				gomonkey.ApplyFunc((*etcd3.EtcdClient).AttachAZPrefix, func(_ *etcd3.EtcdClient, key string) string { return key }),
			})
			return patches
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := tt.patchesFunc()
			_, err := NewFaaSManagerLibruntime(tt.args.sdkClient, tt.args.stopCh)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFaaSManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			patches.ResetAll()
		})
	}
}

func TestManager_ProcessSchedulerRequest(t *testing.T) {
	defer gomonkey.ApplyFunc(state.Update, func(value interface{}, tags ...string) {
	}).Reset()
	type fields struct {
		patKeyList        map[string]map[string]struct{}
		vpcMutex          sync.Mutex
		remoteClientLease map[string]*leaseTimer
		remoteClientList  map[string]struct{}
		clientMutex       sync.Mutex
	}
	type args struct {
		args []*api.Arg
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		want        int
		patchesFunc mockUtils.PatchesFunc
	}{
		{
			"case7 failed to requestOpNewLease",
			fields{
				patKeyList: make(map[string]map[string]struct{}, 1),
				vpcMutex:   sync.Mutex{}, remoteClientLease: map[string]*leaseTimer{"test-client-ID": nil},
				remoteClientList: map[string]struct{}{}, clientMutex: sync.Mutex{},
			},
			args{args: []*api.Arg{{Data: []byte(requestOpNewLease)}, {Data: []byte("")}}},
			constant.UnsupportedOperationErrorCode, func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches = append(patches, gomonkey.ApplyFunc((*etcd3.EtcdClient).AttachAZPrefix, func(_ *etcd3.EtcdClient, key string) string { return key }))
				return patches
			},
		},
		{
			"case8 success to requestOpNewLease",
			fields{
				patKeyList: make(map[string]map[string]struct{}, 1),
				vpcMutex:   sync.Mutex{}, remoteClientLease: make(map[string]*leaseTimer, 1),
				remoteClientList: map[string]struct{}{}, clientMutex: sync.Mutex{},
			},
			args{args: []*api.Arg{
				{Data: []byte(requestOpNewLease)},
				{Data: []byte("test-client-ID")},
				{Data: []byte("")},
			}},
			constant.InsReqSuccessCode, func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches = append(patches, gomonkey.ApplyFunc((*etcd3.EtcdClient).AttachAZPrefix, func(_ *etcd3.EtcdClient, key string) string { return key }))
				return patches
			},
		},
		{
			"case9 failed to requestOpKeepAlive",
			fields{
				patKeyList: make(map[string]map[string]struct{}, 1),
				vpcMutex:   sync.Mutex{}, remoteClientLease: make(map[string]*leaseTimer, 1),
				remoteClientList: map[string]struct{}{}, clientMutex: sync.Mutex{},
			},
			args{args: []*api.Arg{
				{Data: []byte(requestOpKeepAlive)},
				{Data: []byte("test-client-ID")},
				{Data: []byte("")},
			}},
			constant.UnsupportedOperationErrorCode, func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches = append(patches, gomonkey.ApplyFunc((*etcd3.EtcdClient).AttachAZPrefix, func(_ *etcd3.EtcdClient, key string) string { return key }))
				return patches
			},
		},
		{
			"case10 success to requestOpKeepAlive",
			fields{
				patKeyList: make(map[string]map[string]struct{}, 1),
				vpcMutex:   sync.Mutex{}, remoteClientLease: map[string]*leaseTimer{"test-client-ID": newLeaseTimer(1000)},
				remoteClientList: map[string]struct{}{}, clientMutex: sync.Mutex{},
			},
			args{args: []*api.Arg{
				{Data: []byte(requestOpKeepAlive)},
				{Data: []byte("test-client-ID")},
				{Data: []byte("")},
			}},
			constant.InsReqSuccessCode, func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches = append(patches, gomonkey.ApplyFunc((*etcd3.EtcdClient).AttachAZPrefix, func(_ *etcd3.EtcdClient, key string) string { return key }))
				return patches
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := tt.patchesFunc()
			fm := &Manager{
				remoteClientLease: tt.fields.remoteClientLease,
				remoteClientList:  tt.fields.remoteClientList,
				clientMutex:       tt.fields.clientMutex,
				queue: &queue{
					cond:       sync.NewCond(&sync.RWMutex{}),
					dirty:      map[interface{}]struct{}{},
					processing: map[interface{}]struct{}{},
				},
				leaseRenewMinute: 5,
			}
			go fm.saveStateLoop()
			libruntimeClient = &mockUtils.FakeLibruntimeSdkClient{}
			if got := fm.ProcessSchedulerRequest(tt.args.args, ""); !reflect.DeepEqual(got.Code, tt.want) {
				t.Errorf("ProcessSchedulerRequest() = %v, want %v", got, tt.want)
			}
			patches.ResetAll()
			libruntimeClient = nil
		})
	}
}

func TestManager_ProcessSchedulerRequestLibruntime(t *testing.T) {
	defer gomonkey.ApplyFunc(state.Update, func(value interface{}, tags ...string) {
	}).Reset()
	type fields struct {
		patKeyList        map[string]map[string]struct{}
		vpcMutex          sync.Mutex
		remoteClientLease map[string]*leaseTimer
		remoteClientList  map[string]struct{}
		clientMutex       sync.Mutex
	}
	type args struct {
		args []api.Arg
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		want        int
		patchesFunc mockUtils.PatchesFunc
	}{
		{
			"case1 failed to requestOpNewLease",
			fields{
				patKeyList: make(map[string]map[string]struct{}, 1),
				vpcMutex:   sync.Mutex{}, remoteClientLease: map[string]*leaseTimer{"test-client-ID": nil},
				remoteClientList: map[string]struct{}{}, clientMutex: sync.Mutex{},
			},
			args{args: []api.Arg{{}, {Data: []byte(requestOpNewLease)}, {Data: []byte("")}}},
			constant.UnsupportedOperationErrorCode, func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{})
				return patches
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := tt.patchesFunc()
			fm := &Manager{
				remoteClientLease: tt.fields.remoteClientLease,
				remoteClientList:  tt.fields.remoteClientList,
				clientMutex:       tt.fields.clientMutex,
				queue: &queue{
					cond:       sync.NewCond(&sync.RWMutex{}),
					dirty:      map[interface{}]struct{}{},
					processing: map[interface{}]struct{}{},
				},
				leaseRenewMinute: 5,
			}
			go fm.saveStateLoop()
			libruntimeClient = &mockUtils.FakeLibruntimeSdkClient{}
			if got := fm.ProcessSchedulerRequestLibruntime(tt.args.args, ""); !reflect.DeepEqual(got.Code, tt.want) {
				t.Errorf("ProcessSchedulerRequest() = %v, want %v", got, tt.want)
			}
			patches.ResetAll()
			libruntimeClient = nil
		})
	}
}

func Test_handleRequestOpCreate(t *testing.T) {
	convey.Convey("ProcessSchedulerRequest-handleRequestOpCreate", t, func() {
		listKinds := map[schema.GroupVersionResource]string{
			// Example: MyResource GVK to MyResourceList GVK
			patGVR: "PatList",
		}
		fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), listKinds)
		defer gomonkey.ApplyFunc(k8sclient.GetDynamicClient, func() dynamic.Interface {
			return fakeClient
		}).Reset()
		factory := dynamicinformer.NewDynamicSharedInformerFactory(fakeClient, time.Minute)
		informer := factory.ForResource(patGVR)
		stopCh := make(chan struct{})
		go informer.Informer().Run(stopCh)
		if !cache.WaitForCacheSync(stopCh, informer.Informer().HasSynced) {
		}
		m := &Manager{
			leaseRenewMinute: 5,
			patLister:        informer.Lister(),
		}
		emptyArgs := []*api.Arg{
			{
				Type: 0,
				Data: []byte(requestOpCreate),
			},
			{
				Type: 0,
				Data: []byte{},
			},
			{
				Type: 0,
				Data: []byte("trace-id"),
			},
		}
		convey.Convey("failed to create vpc pat pod, error", func() {
			response := m.ProcessSchedulerRequest(emptyArgs, "")
			convey.So(response.Code, convey.ShouldEqual, 6040)
		})
		normalArgs := []*api.Arg{
			{
				Type: 0,
				Data: []byte(requestOpCreate),
			},
			{
				Type: 0,
				Data: []byte("{\"namespace\": \"default\",\"domainID\": \"\",\"projectID\": \"\",\"environmentID\": \"\",\"vpcID\": \"\",\"subnetID\": \"\",\"tenantCidr\": \"\",\"hostVMCidr\": \"\",\"gateway\": \"\",\"securityGroup\": [],\"xrole\": \"\",\"IPV6Enable\": false}"),
			},
			{
				Type: 0,
				Data: []byte("trace-id"),
			},
		}

		convey.Convey("wait pat pod timeout", func() {
			defaultCreatePatPodTimeout = 1500 * time.Millisecond
			response := m.ProcessSchedulerRequest(normalArgs, "")
			convey.So(response.Code, convey.ShouldEqual, 6040)
		})
		convey.Convey("pat pod exist", func() {
			emptyResult := &unstructured.Unstructured{}
			fakeClient.Invokes(testing2.NewCreateAction(patGVR, "default", &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "patservice.cap.io/v1",
					"kind":       "Pat",
					"metadata": map[string]interface{}{
						"name":      "pat--e3b0c442",
						"namespace": "default",
					},
					"spec": map[string]interface{}{
						"delegate_role":  "fgs_admin",
						"environment_id": "sdfase",
						"require_count":  int64(2),
						"vpc": map[string]interface{}{
							"domain_id":    "xxxxx",
							"gateway":      "182.20.0.1",
							"host_vm_cidr": "10.1.0.0/18",
							"project_id":   "xxxxxx",
							"subnet_id":    "subnet",
							"tenant_cidr":  "182.20.0.0/14",
							"vpc_id":       "vpc",
						},
					},
					"status": map[string]interface{}{
						"pat_pods": []interface{}{
							map[string]interface{}{
								"status":       "Active",
								"pat_pod_name": "pod1",
							},
						},
					},
				},
			}), emptyResult)
			time.Sleep(100 * time.Millisecond)
			defaultCreatePatPodTimeout = 1500 * time.Millisecond
			response := m.ProcessSchedulerRequest(normalArgs, "")
			convey.So(response.Code, convey.ShouldEqual, 6030)
		})
	})
}

func Test_handleLeaseTimeout(t *testing.T) {
	convey.Convey("handleLeaseTimeout", t, func() {
		m := &Manager{
			remoteClientLease: map[string]*leaseTimer{},
			remoteClientList:  map[string]struct{}{},
			clientMutex:       sync.Mutex{},
			queue: &queue{
				cond:       sync.NewCond(&sync.RWMutex{}),
				dirty:      map[interface{}]struct{}{},
				processing: map[interface{}]struct{}{},
			},
			leaseRenewMinute: 5,
		}
		go m.saveStateLoop()
		args := []*api.Arg{
			{
				Type: 0,
				Data: []byte(requestOpNewLease),
			},
			{
				Type: 0,
				Data: []byte("test-client-ID"),
			},
			{
				Type: 0,
				Data: []byte("trace-id"),
			},
		}
		convey.Convey("handle client lease timeout", func() {
			defer gomonkey.ApplyFunc(state.Update, func(value interface{}, tags ...string) {}).Reset()
			libruntimeClient = &mockUtils.FakeLibruntimeSdkClient{}
			request := m.ProcessSchedulerRequest(args, "test-trace-ID")
			convey.So(request.Message, convey.ShouldEqual, "Succeed to create lease")
			convey.So(len(m.remoteClientList), convey.ShouldEqual, 1)
			convey.So(m.remoteClientLease["test-client-ID"], convey.ShouldNotBeNil)
			m.remoteClientLease["test-client-ID"].Timer.Reset(1 * time.Second)
			time.Sleep(2 * time.Second)
			convey.So(len(m.remoteClientList), convey.ShouldEqual, 0)
			convey.So(len(m.remoteClientLease), convey.ShouldEqual, 0)
			libruntimeClient = nil
		})
	})
}

func Test_handleDelLease(t *testing.T) {
	convey.Convey("handleDelLease", t, func() {
		m := &Manager{
			remoteClientLease: map[string]*leaseTimer{"test-client-ID": newLeaseTimer(1000)},
			remoteClientList:  map[string]struct{}{"test-client-ID": {}},
			clientMutex:       sync.Mutex{},
			queue: &queue{
				cond:       sync.NewCond(&sync.RWMutex{}),
				dirty:      map[interface{}]struct{}{},
				processing: map[interface{}]struct{}{},
			},
			leaseRenewMinute: 5,
		}
		go m.saveStateLoop()
		args := []*api.Arg{
			{
				Type: 0,
				Data: []byte(requestOpDelLease),
			},
			{
				Type: 0,
				Data: []byte("test-client-ID"),
			},
			{
				Type: 0,
				Data: []byte("trace-id"),
			},
		}
		libruntimeClient = &mockUtils.FakeLibruntimeSdkClient{}
		convey.Convey("delete no exited lease", func() {
			args[1].Data = []byte("test-client-ID-error")
			defer gomonkey.ApplyFunc(state.Update, func(value interface{}, tags ...string) {}).Reset()
			request := m.ProcessSchedulerRequest(args, "test-trace-ID")
			convey.So(request.Message, convey.ShouldEqual, "Succeed to delete lease")
			convey.So(len(m.remoteClientList), convey.ShouldEqual, 1)
			convey.So(len(m.remoteClientLease), convey.ShouldEqual, 1)
		})
		convey.Convey("success to delete lease", func() {
			defer gomonkey.ApplyFunc(state.Update, func(value interface{}, tags ...string) {}).Reset()
			request := m.ProcessSchedulerRequest(args, "test-trace-ID")
			convey.So(request.Message, convey.ShouldEqual, "Succeed to delete lease")
			convey.So(len(m.remoteClientList), convey.ShouldEqual, 0)
			convey.So(len(m.remoteClientLease), convey.ShouldEqual, 0)
		})
		libruntimeClient = nil
	})
}

func TestManager_handleLeaseEvent(t *testing.T) {
	libruntimeClient = &mockUtils.FakeLibruntimeSdkClient{}
	convey.Convey("test handleLeaseEvent", t, func() {
		convey.Convey("baseline", func() {
			p := gomonkey.ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
				return &etcd3.EtcdClient{Client: &clientv3.Client{KV: &KvMock{}}}
			})
			defer p.Reset()
			ch := make(chan struct{})
			fm := &Manager{
				remoteClientLease: make(map[string]*leaseTimer),
				remoteClientList:  map[string]struct{}{},
				clientMutex:       sync.Mutex{},
				stopCh:            ch,
				queue: &queue{
					cond:       sync.NewCond(&sync.RWMutex{}),
					dirty:      map[interface{}]struct{}{},
					processing: map[interface{}]struct{}{},
				},
				leaseRenewMinute: 5,
			}
			go fm.saveStateLoop()
			e := &commonType.LeaseEvent{
				Type:           constant.NewLease,
				RemoteClientID: "123456",
				Timestamp:      time.Now().Unix(),
				TraceID:        "abc",
			}
			marshal, _ := json.Marshal(e)
			event := &etcd3.Event{
				Type:      0,
				Key:       "/sn/lease/123456",
				Value:     marshal,
				PrevValue: nil,
				Rev:       0,
			}
			fm.handleLeaseEvent(event)
			convey.So(fm.remoteClientLease["123456"], convey.ShouldNotBeNil)

			e = &commonType.LeaseEvent{
				Type:           constant.DelLease,
				RemoteClientID: "123456",
				Timestamp:      time.Now().Unix(),
				TraceID:        "abcd",
			}
			marshal, _ = json.Marshal(e)
			event = &etcd3.Event{
				Type:      0,
				Key:       "/sn/lease/123456",
				Value:     marshal,
				PrevValue: nil,
				Rev:       0,
			}
			fm.handleLeaseEvent(event)
			convey.So(fm.remoteClientLease["123456"], convey.ShouldBeNil)

			e = &commonType.LeaseEvent{
				Type:           constant.KeepAlive,
				RemoteClientID: "123456",
				Timestamp:      time.Now().Unix(),
				TraceID:        "abcde",
			}
			marshal, _ = json.Marshal(e)
			event = &etcd3.Event{
				Type:      0,
				Key:       "/sn/lease/123456",
				Value:     marshal,
				PrevValue: nil,
				Rev:       0,
			}
			fm.handleLeaseEvent(event)
			time.Sleep(100 * time.Millisecond)
			convey.So(fm.remoteClientLease["123456"], convey.ShouldNotBeNil)

			e = &commonType.LeaseEvent{
				Type:           constant.KeepAlive,
				RemoteClientID: "123456",
				Timestamp:      time.Now().Unix(),
				TraceID:        "abcdef",
			}
			marshal, _ = json.Marshal(e)
			event = &etcd3.Event{
				Type:      0,
				Key:       "/sn/lease/123456",
				Value:     marshal,
				PrevValue: nil,
				Rev:       0,
			}
			fm.handleLeaseEvent(event)
			convey.So(fm.remoteClientLease["123456"], convey.ShouldNotBeNil)
		})

		convey.Convey("event is nil", func() {
			ch := make(chan struct{})
			fm := &Manager{
				remoteClientLease: make(map[string]*leaseTimer),
				remoteClientList:  map[string]struct{}{},
				clientMutex:       sync.Mutex{},
				stopCh:            ch,
				queue: &queue{
					cond:       sync.NewCond(&sync.RWMutex{}),
					dirty:      map[interface{}]struct{}{},
					processing: map[interface{}]struct{}{},
				},
				leaseRenewMinute: 5,
			}
			go fm.saveStateLoop()
			fm.handleLeaseEvent(nil)
			convey.So(len(fm.remoteClientLease), convey.ShouldEqual, 0)
		})

		convey.Convey("unmarshal failed", func() {
			ch := make(chan struct{})
			fm := &Manager{
				remoteClientLease: make(map[string]*leaseTimer),
				remoteClientList:  map[string]struct{}{},
				clientMutex:       sync.Mutex{},
				stopCh:            ch,
				queue: &queue{
					cond:       sync.NewCond(&sync.RWMutex{}),
					dirty:      map[interface{}]struct{}{},
					processing: map[interface{}]struct{}{},
				},
				leaseRenewMinute: 5,
			}
			go fm.saveStateLoop()
			event := &etcd3.Event{
				Type:      0,
				Key:       "/sn/lease/123456",
				Value:     []byte("aaa"),
				PrevValue: nil,
				Rev:       0,
			}
			fm.handleLeaseEvent(event)
			convey.So(len(fm.remoteClientLease), convey.ShouldEqual, 0)
		})

		convey.Convey("error type", func() {
			ch := make(chan struct{})
			fm := &Manager{
				remoteClientLease: make(map[string]*leaseTimer),
				remoteClientList:  map[string]struct{}{},
				clientMutex:       sync.Mutex{},
				stopCh:            ch,
				queue: &queue{
					cond:       sync.NewCond(&sync.RWMutex{}),
					dirty:      map[interface{}]struct{}{},
					processing: map[interface{}]struct{}{},
				},
				leaseRenewMinute: 5,
			}
			go fm.saveStateLoop()
			e := &commonType.LeaseEvent{
				Type:           "aaa",
				RemoteClientID: "123456",
				Timestamp:      time.Now().Unix(),
				TraceID:        "abcdef",
			}
			marshal, _ := json.Marshal(e)
			event := &etcd3.Event{
				Type:      0,
				Key:       "/sn/lease/123456",
				Value:     marshal,
				PrevValue: nil,
				Rev:       0,
			}
			fm.handleLeaseEvent(event)
			convey.So(len(fm.remoteClientLease), convey.ShouldEqual, 0)
		})
	})
	libruntimeClient = nil
}

func TestManager_WatchLeaseEvent(t *testing.T) {
	libruntimeClient = &mockUtils.FakeLibruntimeSdkClient{}
	convey.Convey("test watch lease event", t, func() {
		convey.Convey("baseline", func() {
			p := gomonkey.ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
				return &etcd3.EtcdClient{Client: &clientv3.Client{KV: &KvMock{}}}
			})
			defer p.Reset()
			p2 := gomonkey.ApplyFunc((*etcd3.EtcdWatcher).StartWatch, func(ew *etcd3.EtcdWatcher) {
				ew.ResultChan <- &etcd3.Event{
					Type:      0,
					Key:       "/sn/lease/123456",
					Value:     nil,
					PrevValue: nil,
					Rev:       0,
				}
			})
			defer p2.Reset()
			p4 := gomonkey.ApplyFunc((*etcd3.EtcdClient).AttachAZPrefix, func(_ *etcd3.EtcdClient, str string) string {
				return str
			})
			defer p4.Reset()
			flag := false
			p3 := gomonkey.ApplyFunc((*Manager).handleLeaseEvent, func(_ *Manager) {
				flag = true
			})
			defer p3.Reset()
			ch := make(chan struct{})
			fm := &Manager{
				remoteClientLease: make(map[string]*leaseTimer),
				remoteClientList:  map[string]struct{}{},
				clientMutex:       sync.Mutex{},
				stopCh:            ch,
				queue: &queue{
					cond:       sync.NewCond(&sync.RWMutex{}),
					dirty:      map[interface{}]struct{}{},
					processing: map[interface{}]struct{}{},
				},
				leaseRenewMinute: 5,
			}
			go fm.saveStateLoop()
			go fm.WatchLeaseEvent()
			time.Sleep(100 * time.Millisecond)
			convey.So(flag, convey.ShouldBeTrue)
			ch <- struct{}{}
		})
	})
	libruntimeClient = nil
}

func TestManager_RecoverData(t *testing.T) {
	libruntimeClient = &mockUtils.FakeLibruntimeSdkClient{}
	convey.Convey("test recover data", t, func() {
		convey.Convey("baseline", func() {
			p := gomonkey.ApplyFunc(state.GetState, func() *state.ManagerState {
				return &state.ManagerState{
					PatKeyList: map[string]map[string]struct{}{
						"a": {"a": struct{}{}},
					},
					RemoteClientList: map[string]struct{}{"aaa": {}, "bbb": {}},
				}
			})
			defer p.Reset()
			fm := &Manager{
				remoteClientLease: make(map[string]*leaseTimer),
				remoteClientList:  map[string]struct{}{},
				clientMutex:       sync.Mutex{},
				stopCh:            nil,
				queue: &queue{
					cond:       sync.NewCond(&sync.RWMutex{}),
					dirty:      map[interface{}]struct{}{},
					processing: map[interface{}]struct{}{},
				},
				leaseRenewMinute: 5,
			}
			go fm.saveStateLoop()
			fm.RecoverData()
			convey.So(len(fm.remoteClientLease), convey.ShouldEqual, 2)
		})
	})
	libruntimeClient = nil
}
