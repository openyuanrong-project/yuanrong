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

// Package vpcmanager -
package vpcmanager

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/dynamic/fake"
	testing2 "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/k8sclient"
	types2 "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionmanager/types"
	"yuanrong.org/kernel/pkg/functionmanager/utils"
)

func TestWaitDelete(t *testing.T) {
	listKinds := map[schema.GroupVersionResource]string{
		// Example: MyResource GVK to MyResourceList GVK
		patGVR: "PatList",
	}
	initialObjects := []runtime.Object{}
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), listKinds, initialObjects...)
	defer gomonkey.ApplyFunc(k8sclient.GetDynamicClient, func() dynamic.Interface {
		return fakeClient
	}).Reset()
	factory := dynamicinformer.NewDynamicSharedInformerFactory(fakeClient, time.Minute)
	informer := factory.ForResource(patGVR)
	stopCh := make(chan struct{})
	go informer.Informer().Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, informer.Informer().HasSynced) {
	}
	vpcManager := MakeVPCManager(informer, 5*time.Millisecond, stopCh)
	go vpcManager.Run()
	vpcEvent := types.VPCEvent{
		EventType: "update",
		InsInfo:   nil,
		PatInfo: &types.Pat{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pat-simple-1",
				Namespace: "default",
			},
			Status: types.PatStatus{PatPods: []types.PatPodStatus{
				{PatPodName: "pod1"}, {PatPodName: "pod2"},
			}},
		},
	}
	emptyResult := &unstructured.Unstructured{}
	convey.Convey("handle pat empty, delete pat", t, func() {
		fakeClient.Invokes(testing2.NewCreateAction(patGVR, "default", &unstructured.Unstructured{Object: map[string]interface{}{
			"kind":       "Pat",
			"apiVersion": "patservice.cap.io/v1",
			"metadata": map[string]interface{}{
				"name":      "pat-simple-1",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"require_count": int64(0),
			},
		}}), emptyResult)
		vpcManager.EventCh <- vpcEvent
		time.Sleep(2 * time.Millisecond)
		fakeClient.Invokes(testing2.NewUpdateAction(patGVR, "default", &unstructured.Unstructured{Object: map[string]interface{}{
			"kind":       "Pat",
			"apiVersion": "patservice.cap.io/v1",
			"metadata": map[string]interface{}{
				"name":      "pat-simple-1",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"require_count": int64(2),
			},
		}}), emptyResult)
		time.Sleep(5 * time.Millisecond)
		fakeClient.Invokes(testing2.NewUpdateAction(patGVR, "default", &unstructured.Unstructured{Object: map[string]interface{}{
			"kind":       "Pat",
			"apiVersion": "patservice.cap.io/v1",
			"metadata": map[string]interface{}{
				"name":      "pat-simple-1",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"require_count": int64(0),
			},
		}}), emptyResult)
		time.Sleep(5 * time.Millisecond)
		vpcManager.EventCh <- vpcEvent
		time.Sleep(10 * time.Millisecond)
		_, err := informer.Lister().ByNamespace("default").Get("pat-simple-1")
		convey.So(k8serrors.IsNotFound(err), convey.ShouldEqual, true)
	})
	close(stopCh)
}

func TestHandlePatCREvent(t *testing.T) {
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
			"status": map[string]interface{}{
				"pat_pods": []interface{}{
					map[string]interface{}{
						"pat_pod_name": "pod1",
						"status":       "Active",
					},
				},
			},
		}},
	}
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), listKinds, initialObjects...)
	defer gomonkey.ApplyFunc(k8sclient.GetDynamicClient, func() dynamic.Interface {
		return fakeClient
	}).Reset()
	factory := dynamicinformer.NewDynamicSharedInformerFactory(fakeClient, time.Minute)
	informer := factory.ForResource(patGVR)
	stopCh := make(chan struct{})
	go informer.Informer().Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, informer.Informer().HasSynced) {
	}
	vpcManager := MakeVPCManager(informer, 5*time.Millisecond, stopCh)
	go vpcManager.Run()
	vpcEvent := types.VPCEvent{
		EventType: "update",
		InsInfo:   nil,
		PatInfo: &types.Pat{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pat-simple",
				Namespace: "default",
			},
			Status: types.PatStatus{PatPods: []types.PatPodStatus{
				{PatPodName: "pod1"}, {PatPodName: "pod2"},
			}},
		},
	}
	convey.Convey("normal handle pat, update idle", t, func() {
		vpcManager.EventCh <- vpcEvent
		time.Sleep(5 * time.Millisecond)
		obj, err := informer.Lister().ByNamespace("default").Get("pat-simple")
		convey.So(err, convey.ShouldBeNil)
		unstruct, ok := obj.(*unstructured.Unstructured)
		convey.So(ok, convey.ShouldEqual, true)
		convey.So(unstruct.GetAnnotations()[types.PatIdlePodAnnotationKey], convey.ShouldEqual, "")
		emptyResult := &unstructured.Unstructured{}
		fakeClient.Invokes(testing2.NewUpdateAction(patGVR, "default", &unstructured.Unstructured{Object: map[string]interface{}{
			"kind":       "Pat",
			"apiVersion": "patservice.cap.io/v1",
			"metadata": map[string]interface{}{
				"name":      "pat-simple",
				"namespace": "default",
				"annotations": map[string]interface{}{
					"patservice.cap.io/idle-pods": "pod3,pod5",
				},
			},
			"spec": map[string]interface{}{
				"require_count": int64(2),
			},
			"status": map[string]interface{}{
				"pat_pods": []interface{}{
					map[string]interface{}{
						"pat_pod_name": "pod1",
						"status":       "Active",
					},
					map[string]interface{}{
						"pat_pod_name": "pod2",
						"status":       "Active",
					},
				},
			},
		}}), emptyResult)
		time.Sleep(5 * time.Millisecond)
		vpcManager.EventCh <- vpcEvent
		time.Sleep(5 * time.Millisecond)
		obj, err = informer.Lister().ByNamespace("default").Get("pat-simple")
		convey.So(err, convey.ShouldBeNil)
		unstruct, ok = obj.(*unstructured.Unstructured)
		convey.So(ok, convey.ShouldEqual, true)
		convey.So(unstruct.GetAnnotations()[types.PatIdlePodAnnotationKey], convey.ShouldEqual, "pod1,pod2")

		ins := types2.InstanceSpecification{
			InstanceID:    "c84dd90e-98ad-4371-8000-00000000a2cb",
			CreateOptions: map[string]string{"DELEGATE_NETWORK_CONFIG": "{\"patInstances\":[{\"namespace\":\"default\",\"name\":\"pod1\"},{\"namespace\":\"default\",\"name\":\"pod2\"}]}"},
			InstanceStatus: types2.InstanceStatus{
				Code: int32(constant.KernelInstanceStatusRunning),
			},
		}
		vpcManager.EventCh <- types.VPCEvent{
			EventType: "update",
			InsInfo:   &ins,
		}
		ins2 := types2.InstanceSpecification{
			InstanceID: "c84dd90e-98ad-4371-8000-00000000a2cb",
			InstanceStatus: types2.InstanceStatus{
				Code: int32(constant.KernelInstanceStatusRunning),
			},
		}
		vpcManager.EventCh <- types.VPCEvent{
			EventType: "update",
			InsInfo:   &ins2,
		}
		ins3 := types2.InstanceSpecification{
			InstanceID:    "c84dd90e-98ad-4371-8000-00000000a2cb",
			CreateOptions: map[string]string{"DELEGATE_NETWORK_CONFIG": "{"},
			InstanceStatus: types2.InstanceStatus{
				Code: int32(constant.KernelInstanceStatusRunning),
			},
		}
		vpcManager.EventCh <- types.VPCEvent{
			EventType: "update",
			InsInfo:   &ins3,
		}
		vpcManager.EventCh <- vpcEvent
		time.Sleep(5 * time.Millisecond)
		vpcManager.EventCh <- vpcEvent
		obj, err = informer.Lister().ByNamespace("default").Get("pat-simple")
		convey.So(err, convey.ShouldBeNil)
		unstruct, ok = obj.(*unstructured.Unstructured)
		convey.So(ok, convey.ShouldEqual, true)
		convey.So(unstruct.GetAnnotations()[types.PatIdlePodAnnotationKey], convey.ShouldEqual, "")

		ins = types2.InstanceSpecification{
			InstanceID:    "c84dd90e-98ad-4371-8000-00000000a2cb",
			CreateOptions: map[string]string{"DELEGATE_NETWORK_CONFIG": "{\"patInstances\":[{\"namespace\":\"default\",\"name\":\"pod1\"},{\"namespace\":\"default\",\"name\":\"pod2\"}]}"},
			InstanceStatus: types2.InstanceStatus{
				Code: int32(constant.KernelInstanceStatusEvicted),
			},
		}
		vpcManager.EventCh <- types.VPCEvent{
			EventType: "update",
			InsInfo:   &ins,
		}
		vpcManager.EventCh <- vpcEvent
		time.Sleep(10 * time.Millisecond)
		obj, err = informer.Lister().ByNamespace("default").Get("pat-simple")
		convey.So(err, convey.ShouldBeNil)
		unstruct, ok = obj.(*unstructured.Unstructured)
		convey.So(ok, convey.ShouldEqual, true)
		convey.So(unstruct.GetAnnotations()[types.PatIdlePodAnnotationKey], convey.ShouldEqual, "pod1,pod2")
		convertCnt := 0
		defer gomonkey.ApplyFunc(utils.UnstructuredToPat, func(unstruct *unstructured.Unstructured) (*types.Pat, error) {
			convertCnt++
			return nil, nil
		})
		fakeClient.Invokes(testing2.NewDeleteAction(patGVR, "default", "pat-simple"), emptyResult)
		vpcManager.EventCh <- vpcEvent
		time.Sleep(5 * time.Millisecond)
		convey.So(convertCnt, convey.ShouldEqual, 0)
	})
	close(stopCh)
}

func TestProcessAllRunningPat(t *testing.T) {
	listKinds := map[schema.GroupVersionResource]string{
		// Example: MyResource GVK to MyResourceList GVK
		patGVR: "PatList",
	}
	initialObjects := []runtime.Object{
		&unstructured.Unstructured{Object: map[string]interface{}{
			"kind":       "Pat",
			"apiVersion": "patservice.cap.io/v1",
			"metadata": map[string]interface{}{
				"name":      "pat-simple-2",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"require_count": int64(1),
			},
			"status": map[string]interface{}{
				"pat_pods": []interface{}{
					map[string]interface{}{
						"pat_pod_name": "pod1",
						"status":       "Active",
					},
				},
			},
		}},
	}
	fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), listKinds, initialObjects...)
	defer gomonkey.ApplyFunc(k8sclient.GetDynamicClient, func() dynamic.Interface {
		return fakeClient
	}).Reset()
	factory := dynamicinformer.NewDynamicSharedInformerFactory(fakeClient, time.Minute)
	informer := factory.ForResource(patGVR)
	stopCh := make(chan struct{})
	go informer.Informer().Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, informer.Informer().HasSynced) {
	}
	vpcManager2 := MakeVPCManager(informer, 5*time.Minute, stopCh)
	ctx := context.Background()
	convey.Convey("handle pat info, update pat idle", t, func() {
		failedCnt := 0
		defer gomonkey.ApplyFunc((*VPCManager).updatePatIdlePod, func(_ *VPCManager, ctx context.Context, pat *types.Pat) error {
			failedCnt++
			return errors.New("aaaaa")
		}).Reset()
		vpcManager2.processAllRunningPat(ctx)
		convey.So(failedCnt, convey.ShouldEqual, 1)
		defer gomonkey.ApplyFunc(utils.UnstructuredToPat, func(unstruct *unstructured.Unstructured) (*types.Pat, error) {
			failedCnt++
			return nil, errors.New("failed convert")
		}).Reset()
		vpcManager2.processAllRunningPat(ctx)
		convey.So(failedCnt, convey.ShouldEqual, 2)
	})
	close(stopCh)
}

func TestUpdatePatIdlePod(t *testing.T) {
	convey.Convey("TestUpdatePatIdlePod_with_multiple_namespace", t, func() {
		initialObjects := []runtime.Object{
			&unstructured.Unstructured{Object: map[string]interface{}{
				"kind":       "Pat",
				"apiVersion": "patservice.cap.io/v1",
				"metadata": map[string]interface{}{
					"name":      "pats.patservice.cap.io",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"require_count": int64(1),
				},
				"status": map[string]interface{}{
					"pat_pods": []interface{}{
						map[string]interface{}{
							"pat_pod_name": "pat-xxx1",
							"status":       "Active",
						},
					},
				},
			}},
		}
		listKinds := map[schema.GroupVersionResource]string{
			patGVR: "PatList",
		}
		fakeClient := fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), listKinds, initialObjects...)

		p := []*gomonkey.Patches{
			gomonkey.ApplyFunc(k8sclient.GetDynamicClient, func() dynamic.Interface {
				return fakeClient
			}),
		}
		defer func() {
			for _, patche := range p {
				patche.Reset()
			}
		}()
		factory := dynamicinformer.NewDynamicSharedInformerFactory(fakeClient, time.Minute)
		informer := factory.ForResource(patGVR)
		stopCh := make(chan struct{})
		vpcManager := MakeVPCManager(informer, 5*time.Minute, stopCh)

		pat1 := k8stypes.NamespacedName{Namespace: "919afa3d-5dc8-4f13-b2e7-eee79d51e2cd", Name: "cff-pat-pod-abc111"}
		instanceMap1 := make(map[string]struct{})
		instanceMap1["custom-function-agent-00000000007c-500m-500mi-4d109db5000000000"] = struct{}{}
		vpcManager.patPodOfInstance[pat1] = instanceMap1

		pat2 := k8stypes.NamespacedName{Namespace: "919afa3d-5dc8-4f13-b2e7-eee79d51e2cd", Name: "cff-pat-pod-abc222"}
		instanceMap2 := make(map[string]struct{})
		instanceMap2["custom-function-agent-00000000007c-500m-500mi-4d109db5000000000"] = struct{}{}
		vpcManager.patPodOfInstance[pat2] = instanceMap2

		defaultNamespacePat := &types.Pat{
			Status: types.PatStatus{
				PatPods: []types.PatPodStatus{
					{PatPodName: "cff-pat-pod-abc111"},
					{PatPodName: "cff-pat-pod-abc222"},
				},
			},
		}
		defaultNamespacePat.Annotations = make(map[string]string)
		defaultNamespacePat.Namespace = "default"
		defaultNamespacePat.Name = "pat-xxx1"

		otherNamespacePat := &types.Pat{
			Status: types.PatStatus{
				PatPods: []types.PatPodStatus{
					{PatPodName: "cff-pat-pod-abc111"},
					{PatPodName: "cff-pat-pod-abc222"},
				},
			},
		}
		otherNamespacePat.Annotations = make(map[string]string)
		otherNamespacePat.Namespace = "919afa3d-5dc8-4f13-b2e7-eee79d51e2cd"
		otherNamespacePat.Name = "pat-xxx2"

		vpcManager.updatePatIdlePod(context.Background(), defaultNamespacePat)
		convey.ShouldEqual(t, "cff-pat-pod-abc111,cff-pat-pod-abc222", defaultNamespacePat.Annotations[types.PatIdlePodAnnotationKey])
		convey.ShouldEqual(t, "", otherNamespacePat.Annotations[types.PatIdlePodAnnotationKey])

		vpcManager.updatePatIdlePod(context.Background(), defaultNamespacePat)
		convey.ShouldEqual(t, "cff-pat-pod-abc111,cff-pat-pod-abc222", defaultNamespacePat.Annotations[types.PatIdlePodAnnotationKey])
		convey.ShouldEqual(t, "", otherNamespacePat.Annotations[types.PatIdlePodAnnotationKey])

		vpcManager.updatePatIdlePod(context.Background(), otherNamespacePat)
		convey.ShouldEqual(t, "cff-pat-pod-abc111,cff-pat-pod-abc222", defaultNamespacePat.Annotations[types.PatIdlePodAnnotationKey])
		convey.ShouldEqual(t, "", otherNamespacePat.Annotations[types.PatIdlePodAnnotationKey])
	})
}
