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

// Package registry -
package registry

import (
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/rest"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	commonType "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/common/functioncr"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

func TestNewAgentRegistry_BasicInitialization(t *testing.T) {
	defer gomonkey.ApplyFunc(rest.InClusterConfig, func() (*rest.Config, error) {
		return &rest.Config{}, nil
	}).Reset()
	os.Setenv(constant.EnableAgentCRDRegistry, "true")
	stopCh := make(chan struct{})
	registry := NewAgentRegistry(stopCh)
	assert.NotNil(t, registry.dynamicClient, "dynamic client initialization failed")
	assert.NotNil(t, registry.workQueue, "work queue not initialized")

	os.Setenv(constant.EnableAgentCRDRegistry, "")
}

// createTestCR 创建一个测试用的自定义资源
func createTestCR(name, functionUrn string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "yr.cap.io/v1",
			"kind":       "YRTask",
			"metadata": map[string]interface{}{
				"name":              name,
				"namespace":         "default",
				"uid":               "1234-5678-90ab-cdef-" + name,
				"resourceVersion":   "1",
				"creationTimestamp": metav1.Now().Format(time.RFC3339),
			},
			"spec": map[string]interface{}{
				"funcMetaData": map[string]interface{}{
					"functionURN": functionUrn,
				},
				"extendedMetaData": map[string]interface{}{
					"customContainerConfig": map[string]interface{}{
						"image": "nginx:latest",
					},
				},
			},
		},
	}
}

var (
	// AgentList 资源的 GVK
	agentListGVK = schema.GroupVersionKind{
		Group:   "agentrun.cap.io",
		Version: "v1",
		Kind:    "YRTaskList",
	}
)

// TestControllerWithPrePopulatedData 验证在 Informer 启动前预填充的数据
func TestControllerWithPrePopulatedData(t *testing.T) {
	// 1️⃣ 创建 fake client
	scheme := runtime.NewScheme()
	// 添加必要的类型到 scheme
	if err := appsv1.AddToScheme(scheme); err != nil {
		t.Fatalf("Failed to add appsv1 to scheme: %v", err)
	}
	listResourceMap := map[schema.GroupVersionResource]string{
		functioncr.GetCrdGVR(): agentListGVK.Kind,
	}
	//fakeClient := dynamicfake.NewSimpleDynamicClient(scheme)
	fakeClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, listResourceMap,
		createTestCR("init-cr-1", "func1"), createTestCR("init-cr-2", "func2"))
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(fakeClient, 0, corev1.NamespaceAll, nil)

	stopCh := make(chan struct{})
	defer close(stopCh)
	os.Setenv(constant.EnableAgentCRDRegistry, "on")
	agentRegistry := NewAgentRegistry(stopCh)
	agentRegistry.dynamicClient = fakeClient
	agentRegistry.informerFactory = factory
	ch := make(chan SubEvent, 1000)
	agentRegistry.addSubscriberChan(ch)
	go func() {
		for {
			select {
			case event, ok := <-ch:
				if !ok {
					t.Fatalf("ch event error")
				}
				if event.EventType == SubEventTypeSynced {
					listCh := event.EventMsg.(chan struct{})
					listCh <- struct{}{}
					t.Logf("synced event comming")
				} else {
					t.Logf("eventType: %s, eventValue: %+v", event.EventType, event.EventMsg)
				}
			case <-stopCh:
				return
			}
		}
	}()
	agentRegistry.RunWatcher()
}

func TestAgentRegistry_AddFuncSpec(t *testing.T) {
	convey.Convey("Given an AgentRegistry instance", t, func() {
		ar := &AgentRegistry{
			specLock:      sync.RWMutex{},
			crToFuncSpecs: make(map[string]*types.FunctionSpecification),
			funcSpecs:     make(map[string]*types.FunctionSpecification),
		}

		// Create a FuncSpec to use in tests
		spec := &types.FunctionSpecification{FuncKey: "test-func"}

		convey.Convey("When agentRunEventInfo or FuncSpec is nil", func() {
			// Case 1: agentRunEventInfo is nil
			ar.addFuncSpec(nil, "default")
			convey.So(len(ar.crToFuncSpecs), convey.ShouldEqual, 0)
			convey.So(len(ar.funcSpecs), convey.ShouldEqual, 0)

			// Case 2: agentRunEventInfo.FuncSpec is nil
			eventInfo := &types.AgentEventInfo{CrKey: "default:test-cr"}
			ar.addFuncSpec(eventInfo, "default")
			convey.So(len(ar.crToFuncSpecs), convey.ShouldEqual, 0)
			convey.So(len(ar.funcSpecs), convey.ShouldEqual, 0)
		})

		convey.Convey("When agentRunEventInfo is valid", func() {
			eventInfo := &types.AgentEventInfo{
				CrKey:    "test-namespace:test-cr-123",
				FuncSpec: spec,
			}
			namespace := "test-namespace"

			ar.addFuncSpec(eventInfo, namespace)

			convey.So(ar.crToFuncSpecs, convey.ShouldContainKey, eventInfo.CrKey)
			convey.So(ar.funcSpecs, convey.ShouldContainKey, spec.FuncKey)
			convey.So(ar.crToFuncSpecs[eventInfo.CrKey].NameSpace, convey.ShouldEqual, namespace)
			convey.So(ar.funcSpecs[spec.FuncKey].NameSpace, convey.ShouldEqual, namespace)
		})
	})
}

func TestAgentRegistry_DeleteFuncSpec(t *testing.T) {
	convey.Convey("Given an AgentRegistry with pre-populated data", t, func() {
		ar := &AgentRegistry{
			specLock:      sync.RWMutex{},
			crToFuncSpecs: make(map[string]*types.FunctionSpecification),
			funcSpecs:     make(map[string]*types.FunctionSpecification),
		}

		crKey := "default:test-cr-for-delete"
		funcKey := "test-func-for-delete"
		spec := &types.FunctionSpecification{FuncKey: funcKey}

		ar.crToFuncSpecs[crKey] = spec
		ar.funcSpecs[funcKey] = spec

		convey.So(len(ar.crToFuncSpecs), convey.ShouldEqual, 1)
		convey.So(len(ar.funcSpecs), convey.ShouldEqual, 1)

		convey.Convey("When deleting an existing CR name", func() {
			ar.deleteFuncSpec(crKey)

			convey.So(len(ar.crToFuncSpecs), convey.ShouldEqual, 0)
			convey.So(len(ar.funcSpecs), convey.ShouldEqual, 0)
		})

		convey.Convey("When trying to delete a non-existent CR name", func() {
			nonExistentCrName := "non-existent-cr"
			// Pre-populate with valid data again for this specific test run
			ar.crToFuncSpecs[crKey] = spec
			ar.funcSpecs[funcKey] = spec

			ar.deleteFuncSpec(nonExistentCrName)

			// Nothing should be deleted
			convey.So(len(ar.crToFuncSpecs), convey.ShouldEqual, 1)
			convey.So(len(ar.funcSpecs), convey.ShouldEqual, 1)
			convey.So(ar.crToFuncSpecs[crKey], convey.ShouldEqual, spec)
			convey.So(ar.funcSpecs[funcKey], convey.ShouldEqual, spec)
		})
	})
}

func TestBuildAgentCRInfo(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]interface{}
		wantErr bool
	}{
		{
			name: "invalid crd without spec",
			input: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "test-cr",
				},
			},
			wantErr: true,
		},
		{
			name: "valid crd with spec",
			input: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "test-cr",
				},
				"spec": map[string]interface{}{
					"funcMetaData": map[string]interface{}{
						"functionVersionURN": "sn:cn:yrk:0aae5ae9ea9a42c3adc43dd368efb29b:function:0@default@callee:latest",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid crd with status",
			input: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "test-cr",
				},
				"spec": map[string]interface{}{
					"funcMetaData": map[string]interface{}{
						"functionVersionURN": "sn:cn:yrk:0aae5ae9ea9a42c3adc43dd368efb29b:function:0@default@caller:latest",
					},
				},
				"status": map[string]interface{}{
					"ready": true,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &unstructured.Unstructured{Object: tt.input}
			ar := &AgentRegistry{}
			_, err := ar.buildAgentCRInfo(obj, "default:test-cr")
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestSetDefaultValue(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]interface{}
		want  *commonType.FunctionMetaInfo
	}{
		{
			name: "set default timeout",
			input: map[string]interface{}{
				"extendedMetaData": map[string]interface{}{
					"initializer": map[string]interface{}{
						"timeout": 0,
					},
				},
			},
		},
		{
			name: "set default concurrent num",
			input: map[string]interface{}{
				"instanceMetaData": map[string]interface{}{
					"concurrentNum": 0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.input)
			var info commonType.FunctionMetaInfo
			json.Unmarshal(jsonData, &info)
			setDefaultValue(&info)
			if info.InstanceMetaData.ConcurrentNum == 0 || info.ExtendedMetaData.Initializer.Timeout == 0 {
				t.Errorf("not expected")
			}
		})
	}
}
