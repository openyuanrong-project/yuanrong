/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
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

// Package registry -
package registry

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	clientv3 "go.etcd.io/etcd/client/v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/dynamic/fake"
	testing2 "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/functionmanager/types"
	"yuanrong.org/kernel/pkg/functionmanager/utils"
)

var patGVR = schema.GroupVersionResource{
	Group:    "patservice.cap.io",
	Version:  "v1",
	Resource: "pats",
}

type KvMock struct {
}

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
	error) {
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

func TestInstanceFilter(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{
			name:     "invalid length",
			key:      "/instance/tenant/function",
			expected: true,
		},
		{
			name:     "invalid instance part",
			key:      "/sn/invalid/business/yrk/tenant/123/function/invalidExecutorJava21/version/$latest/defaultaz/123/123",
			expected: true,
		},
		{
			name:     "invalid tenant part",
			key:      "/sn/instance/business/yrk/invalid/123/function/invalidExecutorJava21/version/$latest/defaultaz/123/123",
			expected: true,
		},
		{
			name:     "invalid function part",
			key:      "/sn/instance/business/yrk/tenant/123/invalid/invalidExecutorJava21/version/$latest/defaultaz/123/123",
			expected: true,
		},
		{
			name:     "invalid executor part (faasExecutor)",
			key:      "/sn/instance/business/yrk/tenant/123/function/invalidExecutorJava21/version/$latest/defaultaz/123/123",
			expected: true,
		},
		{
			name:     "valid executor part (faasExecutor)",
			key:      "/sn/instance/business/yrk/tenant/123/function/0-system-faasExecutorJava21/version/$latest/defaultaz/123/123",
			expected: false,
		},
		{
			name:     "valid executor part (serveExecutor)",
			key:      "/sn/instance/business/yrk/tenant/123/function/0-system-serveExecutorJava21/version/$latest/defaultaz/123/123",
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &etcd3.Event{Key: tt.key}
			result := instanceFilter(event)
			if result != tt.expected {
				t.Errorf("expected %v, got %v for key %s", tt.expected, result, tt.key)
			}
		})
	}
}

func TestStartWatchEvent(t *testing.T) {
	defer gomonkey.ApplyFunc((*etcd3.EtcdClient).AttachAZPrefix, func(_ *etcd3.EtcdClient, str string) string {
		return str
	}).Reset()
	defer gomonkey.ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
		return &etcd3.EtcdClient{Client: &clientv3.Client{KV: &KvMock{}}}
	}).Reset()
	convey.Convey("test etcd event", t, func() {
		resultCh := make(chan *etcd3.Event)
		defer gomonkey.ApplyFunc((*etcd3.EtcdWatcher).StartWatch, func(ew *etcd3.EtcdWatcher) {
			ew.ResultChan = resultCh
		}).Reset()

		go func() {
			resultCh <- &etcd3.Event{
				Type:  5,
				Key:   "",
				Value: nil,
				Rev:   0,
			}
			resultCh <- &etcd3.Event{
				Type:      0,
				Key:       "/sn/instance/business/yrk/tenant/123/function/0-system-faasExecutorJava21/version/$latest/defaultaz/123/123",
				Value:     nil,
				PrevValue: nil,
				Rev:       0,
			}
			resultCh <- &etcd3.Event{
				Type:      0,
				Key:       "/sn/instance/business/yrk/tenant/123/function/0-system-faasExecutorJava21/version/$latest/defaultaz/123/123",
				Value:     []byte("{"),
				PrevValue: nil,
				Rev:       0,
			}
			resultCh <- &etcd3.Event{
				Type:      0,
				Key:       "/sn/instance/business/yrk/tenant/123/function/0-system-faasExecutorJava21/version/$latest/defaultaz/123/123",
				Value:     []byte("{\"instanceID\": \"103a8efa-9bef-4900-8000-000000db00e8\",\"requestID\": \"e107f994a8f5b57000\",\"functionProxyID\": \"dggphis131877\",\"function\": \"12345678901234561234567890123456/0-system-faasExecutorGo1.x/$latest\",\"resources\": {\"resources\": {},\"scheduleOption\": {},\"createOptions\": {},\"labels\": [\"faas\"],\"instanceStatus\": {\"code\": 3,\"msg\": \"running\"},\"jobID\": \"job-ime-26007305\",\"parentID\": \"scheduler-faas-scheduler-6b68469c4f-f6dw6\",\"parentFunctionProxyAID\": \"\",\"storageType\": \"local\",\"scheduleTimes\": 1,\"deployTimes\": 1,\"args\": [],\"version\": \"1\",\"detached\": true,\"gracefulShutdownTime\": \"900\",\"extensions\": {}}}"),
				PrevValue: nil,
				Rev:       0,
			}
			resultCh <- &etcd3.Event{
				Type:      1,
				Key:       "/sn/instance/business/yrk/tenant/123/function/0-system-faasExecutorJava21/version/$latest/defaultaz/123/123",
				Value:     nil,
				PrevValue: []byte("{"),
				Rev:       0,
			}
			resultCh <- &etcd3.Event{
				Type:      1,
				Key:       "/sn/instance/business/yrk/tenant/123/function/0-system-faasExecutorJava21/version/$latest/defaultaz/123/123",
				Value:     nil,
				PrevValue: []byte("{\"instanceID\": \"103a8efa-9bef-4900-8000-000000db00e8\",\"requestID\": \"e107f994a8f5b57000\",\"functionProxyID\": \"dggphis131877\",\"function\": \"12345678901234561234567890123456/0-system-faasExecutorGo1.x/$latest\",\"resources\": {\"resources\": {},\"scheduleOption\": {},\"createOptions\": {},\"labels\": [\"faas\"],\"instanceStatus\": {\"code\": 3,\"msg\": \"running\"},\"jobID\": \"job-ime-26007305\",\"parentID\": \"scheduler-faas-scheduler-6b68469c4f-f6dw6\",\"parentFunctionProxyAID\": \"\",\"storageType\": \"local\",\"scheduleTimes\": 1,\"deployTimes\": 1,\"args\": [],\"version\": \"1\",\"detached\": true,\"gracefulShutdownTime\": \"900\",\"extensions\": {}}}"),
				Rev:       0,
			}
		}()
		listKinds := map[schema.GroupVersionResource]string{
			// Example: MyResource GVK to MyResourceList GVK
			patGVR: "PatList",
		}
		factory := dynamicinformer.NewDynamicSharedInformerFactory(fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), listKinds), time.Minute)
		informer := factory.ForResource(patGVR)
		eventCh := make(chan types.VPCEvent)
		stopCh := make(chan struct{})
		go informer.Informer().Run(stopCh)
		StartWatchEvent(eventCh, stopCh, informer)
		e1 := <-eventCh
		convey.So(e1.EventType, convey.ShouldEqual, "update")
		convey.So(e1.InsInfo.InstanceID, convey.ShouldEqual, "")
		e2 := <-eventCh
		convey.So(e2.EventType, convey.ShouldEqual, "update")
		convey.So(e2.InsInfo.InstanceID, convey.ShouldEqual, "103a8efa-9bef-4900-8000-000000db00e8")
		e3 := <-eventCh
		convey.So(e3.EventType, convey.ShouldEqual, "delete")
		convey.So(e3.InsInfo.InstanceID, convey.ShouldEqual, "103a8efa-9bef-4900-8000-000000db00e8")
		close(stopCh)
	})
	convey.Convey("test k8s event", t, func() {
		defer gomonkey.ApplyFunc((*etcd3.EtcdWatcher).StartWatch, func(ew *etcd3.EtcdWatcher) {
			ew.ResultChan <- &etcd3.Event{
				Type:  5,
				Key:   "",
				Value: nil,
				Rev:   0,
			}
		}).Reset()
		listKinds := map[schema.GroupVersionResource]string{
			// Example: MyResource GVK to MyResourceList GVK
			patGVR: "PatList",
		}
		initialObjects := []runtime.Object{
			&unstructured.Unstructured{Object: map[string]interface{}{
				"kind":       "Pat",
				"apiVersion": "patservice.cap.io/v1",
				"metadata": map[string]interface{}{
					"name":      "pat-simple",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"require_count": int64(2),
				},
			}},
		}
		fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), listKinds, initialObjects...)
		factory := dynamicinformer.NewDynamicSharedInformerFactory(fakeClient, time.Minute)
		informer := factory.ForResource(patGVR)
		eventCh := make(chan types.VPCEvent)
		stopCh := make(chan struct{})
		go informer.Informer().Run(stopCh)
		StartWatchEvent(eventCh, stopCh, informer)
		e1 := <-eventCh
		convey.So(e1.EventType, convey.ShouldEqual, "update")
		convey.So(e1.PatInfo.Spec.RequireCount, convey.ShouldEqual, 2)
		convey.So(e1.PatInfo.Name, convey.ShouldEqual, "pat-simple")
		emptyResult := &unstructured.Unstructured{}
		fakeClient.Invokes(testing2.NewUpdateAction(patGVR, "default", &unstructured.Unstructured{Object: map[string]interface{}{
			"kind":       "Pat",
			"apiVersion": "patservice.cap.io/v1",
			"metadata": map[string]interface{}{
				"name":      "pat-simple",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"require_count": int64(1),
			},
		}}), emptyResult)
		e2 := <-eventCh
		convey.So(e2.EventType, convey.ShouldEqual, "update")
		convey.So(e2.PatInfo.Spec.RequireCount, convey.ShouldEqual, 1)
		convey.So(e2.PatInfo.Name, convey.ShouldEqual, "pat-simple")
		fakeClient.Invokes(testing2.NewDeleteAction(patGVR, "default", "pat-simple"), emptyResult)
		e3 := <-eventCh
		convey.So(e3.EventType, convey.ShouldEqual, "delete")
		convey.So(e3.PatInfo.Spec.RequireCount, convey.ShouldEqual, 1)
		convey.So(e3.PatInfo.Name, convey.ShouldEqual, "pat-simple")
		failedCnt := 0
		defer gomonkey.ApplyFunc(utils.UnstructuredToPat, func(unstruct *unstructured.Unstructured) (*types.Pat, error) {
			failedCnt++
			return nil, errors.New("convert failed")
		}).Reset()
		fakeClient.Invokes(testing2.NewCreateAction(patGVR, "default", &unstructured.Unstructured{Object: map[string]interface{}{
			"kind":       "Pat",
			"apiVersion": "patservice.cap.io/v1",
			"metadata": map[string]interface{}{
				"name":      "pat-simple",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"require_count": int64(1),
			},
		}}), emptyResult)
		fakeClient.Invokes(testing2.NewUpdateAction(patGVR, "default", &unstructured.Unstructured{Object: map[string]interface{}{
			"kind":       "Pat",
			"apiVersion": "patservice.cap.io/v1",
			"metadata": map[string]interface{}{
				"name":      "pat-simple",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"require_count": int64(1),
			},
		}}), emptyResult)
		fakeClient.Invokes(testing2.NewDeleteAction(patGVR, "default", "pat-simple"), emptyResult)
		time.Sleep(10 * time.Millisecond)
		convey.So(failedCnt, convey.ShouldEqual, 3)
		close(stopCh)
	})
	convey.Convey("test sync failed", t, func() {
		defer gomonkey.ApplyFunc((*etcd3.EtcdWatcher).StartWatch, func(ew *etcd3.EtcdWatcher) {
			ew.ResultChan <- &etcd3.Event{
				Type:  5,
				Key:   "",
				Value: nil,
				Rev:   0,
			}
		}).Reset()
		waitSync := 0
		defer gomonkey.ApplyFunc(cache.WaitForCacheSync, func(stopCh <-chan struct{}, cacheSyncs ...cache.InformerSynced) bool {
			waitSync++
			return false
		}).Reset()
		listKinds := map[schema.GroupVersionResource]string{
			// Example: MyResource GVK to MyResourceList GVK
			patGVR: "PatList",
		}
		initialObjects := []runtime.Object{
			&unstructured.Unstructured{Object: map[string]interface{}{
				"kind":       "Pat",
				"apiVersion": "patservice.cap.io/v1",
				"metadata": map[string]interface{}{
					"name":      "pat-simple",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"require_count": int64(2),
				},
			}},
		}
		fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), listKinds, initialObjects...)
		factory := dynamicinformer.NewDynamicSharedInformerFactory(fakeClient, time.Minute)
		informer := factory.ForResource(patGVR)
		eventCh := make(chan types.VPCEvent)
		stopCh := make(chan struct{})
		go informer.Informer().Run(stopCh)
		StartWatchEvent(eventCh, stopCh, informer)
		time.Sleep(10 * time.Millisecond)
		convey.So(waitSync, convey.ShouldEqual, 1)
	})
}
