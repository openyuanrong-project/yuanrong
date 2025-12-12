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

// Package k8sclient include some k8s client operation
package k8sclient

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	k8testing "k8s.io/client-go/testing"
)

var inClusterConfigFunc = rest.InClusterConfig

type mockK8sClient struct {
	createConfigError    error
	createClientError    error
	expectedConfigCalled bool
}

func TestGetkubeClient(t *testing.T) {
	defer gomonkey.ApplyFunc(rest.InClusterConfig, func() (*rest.Config, error) {
		return &rest.Config{}, nil
	}).Reset()
	convey.Convey("get client success", t, func() {
		defer gomonkey.ApplyFunc(kubernetes.NewForConfig, func(c *rest.Config) (*kubernetes.Clientset, error) {
			return &kubernetes.Clientset{}, nil
		}).Reset()
		client := GetkubeClient()
		convey.So(client, convey.ShouldNotBeNil)
	})
	KubeClientSet = nil
	kubeClientOnce = sync.Once{}
	convey.Convey("get client error", t, func() {
		defer gomonkey.ApplyFunc(kubernetes.NewForConfig, func(c *rest.Config) (*kubernetes.Clientset, error) {
			return nil, fmt.Errorf("get client error")
		}).Reset()
		client := GetkubeClient()
		convey.So(client, convey.ShouldBeNil)
	})
	kubeClientOnce = sync.Once{}
	convey.Convey("get cfg error", t, func() {
		defer gomonkey.ApplyFunc(rest.InClusterConfig, func() (*rest.Config, error) {
			return nil, fmt.Errorf("get cfg error")
		}).Reset()
		client := GetkubeClient()
		convey.So(client, convey.ShouldBeNil)
	})
}

func TestDeleteK8sService(t *testing.T) {
	client := fake.NewSimpleClientset()
	KubeClientSet = &KubeClient{client}
	namespace := "default"
	serviceName := "frontend"
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"app": "frontend",
			},
			Ports: []v1.ServicePort{
				{
					Name:     "http",
					Protocol: "TCP",
					Port:     8888,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 32104,
					},
					NodePort: 31222,
				},
			},
			Type: v1.ServiceTypeNodePort,
		},
	}
	client.PrependReactor("delete", "services", func(action k8testing.Action) (handled bool, ret runtime.Object, err error) {
		deleteAction := action.(k8testing.DeleteAction)
		if deleteAction.GetName() == service.Name && deleteAction.GetNamespace() == service.Namespace {
			return true, service, nil
		}
		return true, nil, fmt.Errorf("Not found")
	})

	convey.Convey("delete service success", t, func() {
		err := KubeClientSet.DeleteK8sService(namespace, serviceName)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("delete service not found", t, func() {
		err := KubeClientSet.DeleteK8sService(namespace, "error service name")
		convey.So(err.Error(), convey.ShouldContainSubstring, "Not found")
	})
	convey.Convey("delete service not found", t, func() {
		client.PrependReactor("delete", "services", func(action k8testing.Action) (handled bool, ret runtime.Object, err error) {
			deleteAction := action.(k8testing.DeleteAction)
			if deleteAction.GetName() == service.Name && deleteAction.GetNamespace() == service.Namespace {
				return true, service, fmt.Errorf("delete error")
			}
			return false, nil, nil
		})

		err := KubeClientSet.DeleteK8sService(namespace, serviceName)
		convey.So(err.Error(), convey.ShouldContainSubstring, "delete error")
	})
	convey.Convey("KubeClientSet is nil", t, func() {
		KubeClientSet = nil
		err := KubeClientSet.DeleteK8sService(namespace, serviceName)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestCreateK8sService(t *testing.T) {
	client := fake.NewSimpleClientset()
	KubeClientSet = &KubeClient{client}
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "frontend",
			Namespace: "default",
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"app": "frontend",
			},
			Ports: []v1.ServicePort{
				{
					Name:     "http",
					Protocol: "TCP",
					Port:     8888,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 32104,
					},
					NodePort: 31222,
				},
			},
			Type: v1.ServiceTypeNodePort,
		},
	}

	convey.Convey("create service success", t, func() {
		err := KubeClientSet.CreateK8sService(service)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("create service error", t, func() {
		client.PrependReactor("create", "services", func(action k8testing.Action) (handled bool, ret runtime.Object, err error) {
			createAction := action.(k8testing.CreateAction)
			if createAction.GetObject().(*v1.Service).Name == service.Name && createAction.GetNamespace() == service.Namespace {
				return true, service, fmt.Errorf("failed to create service")
			}
			return false, nil, nil
		})
		err := KubeClientSet.CreateK8sService(service)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("KubeClientSet is nil", t, func() {
		KubeClientSet = nil
		err := KubeClientSet.CreateK8sService(service)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestCreateK8sConfigMap(t *testing.T) {
	client := fake.NewSimpleClientset()
	KubeClientSet = &KubeClient{client}
	configmap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_configmap",
			Namespace: "default",
		},
		Data: map[string]string{"key": "value"},
	}

	convey.Convey("create configmap success", t, func() {
		err := KubeClientSet.CreateK8sConfigMap(configmap)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("create configmap error", t, func() {
		client.PrependReactor("create", "configmaps", func(action k8testing.Action) (handled bool, ret runtime.Object, err error) {
			createAction := action.(k8testing.CreateAction)
			if createAction.GetObject().(*v1.ConfigMap).Name == configmap.Name && createAction.GetNamespace() == configmap.Namespace {
				return true, configmap, fmt.Errorf("failed to create configmap")
			}
			return false, nil, nil
		})
		err := KubeClientSet.CreateK8sConfigMap(configmap)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("KubeClientSet is nil", t, func() {
		KubeClientSet = nil
		err := KubeClientSet.CreateK8sConfigMap(configmap)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestDeleteK8sConfigMap(t *testing.T) {
	client := fake.NewSimpleClientset()
	KubeClientSet = &KubeClient{client}
	namespace := "default"
	configmapName := "test_configmap"
	configmap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configmapName,
			Namespace: namespace,
		},
		Data: map[string]string{"key": "value"},
	}

	convey.Convey("delete configmap success", t, func() {
		err := KubeClientSet.DeleteK8sConfigMap(namespace, configmapName)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("delete configmap error", t, func() {
		client.PrependReactor("delete", "configmaps", func(action k8testing.Action) (handled bool, ret runtime.Object, err error) {
			deleteAction := action.(k8testing.DeleteAction)
			if deleteAction.GetName() == configmap.Name && deleteAction.GetNamespace() == configmap.Namespace {
				return true, configmap, fmt.Errorf("failed to delete service")
			}
			return false, nil, nil
		})
		err := KubeClientSet.CreateK8sConfigMap(configmap)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("KubeClientSet is nil", t, func() {
		KubeClientSet = nil
		err := KubeClientSet.DeleteK8sConfigMap(namespace, configmapName)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestUpdateK8sConfigMap(t *testing.T) {
	client := fake.NewSimpleClientset()
	KubeClientSet = &KubeClient{client}
	namespace := "default"
	configmapName := "test_configmap"
	configmap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configmapName,
			Namespace: namespace,
		},
		Data: map[string]string{"key": "value"},
	}
	_, err := client.CoreV1().ConfigMaps(namespace).Create(context.TODO(), configmap, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create configmap: %v", err)
	}
	convey.Convey("update configmap success", t, func() {
		err := KubeClientSet.UpdateK8sConfigMap(configmap)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("update configmap error", t, func() {
		client.PrependReactor("update", "configmaps", func(action k8testing.Action) (handled bool, ret runtime.Object, err error) {
			updateAction := action.(k8testing.UpdateAction)
			if updateAction.GetObject().(*v1.ConfigMap).Name == configmap.Name && updateAction.GetNamespace() == configmap.Namespace {
				return true, configmap, fmt.Errorf("failed to update configmap")
			}
			return false, nil, nil
		})
		err := KubeClientSet.UpdateK8sConfigMap(configmap)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("KubeClientSet is nil", t, func() {
		KubeClientSet = nil
		err := KubeClientSet.UpdateK8sConfigMap(configmap)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestGetK8sConfigMap(t *testing.T) {
	client := fake.NewSimpleClientset()
	KubeClientSet = &KubeClient{client}
	namespace := "default"
	configmapName := "test_configmap"
	configmap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configmapName,
			Namespace: namespace,
		},
		Data: map[string]string{"key": "value"},
	}
	_, err := client.CoreV1().ConfigMaps(namespace).Create(context.TODO(), configmap, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create configmap: %v", err)
	}
	convey.Convey("get configmap success", t, func() {
		_, err := KubeClientSet.GetK8sConfigMap(namespace, namespace)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("KubeClientSet is nil", t, func() {
		KubeClientSet = nil
		err := KubeClientSet.UpdateK8sConfigMap(configmap)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestKubeClient_GetK8sSecret(t *testing.T) {
	client := fake.NewSimpleClientset()
	KubeClientSet = &KubeClient{client}
	namespace := "default"
	secretName := "test_secret"
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Data: map[string][]byte{"key": []byte("value")},
	}
	_, err := client.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create secret: %v", err)
	}
	convey.Convey("get secret success", t, func() {
		_, err := KubeClientSet.GetK8sSecret(namespace, namespace)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("KubeClientSet is nil", t, func() {
		KubeClientSet = nil
		_, err := KubeClientSet.GetK8sSecret(namespace, secretName)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestKubeClient_CreateK8sSecret(t *testing.T) {
	client := fake.NewSimpleClientset()
	KubeClientSet = &KubeClient{client}
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_secret",
			Namespace: "default",
		},
		Data: map[string][]byte{"key": []byte("value")},
	}

	convey.Convey("create secret success", t, func() {
		_, err := KubeClientSet.CreateK8sSecret("default", secret)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("create secret error", t, func() {
		client.PrependReactor("create", "secrets", func(action k8testing.Action) (handled bool, ret runtime.Object, err error) {
			createAction := action.(k8testing.CreateAction)
			if createAction.GetObject().(*v1.Secret).Name == secret.Name && createAction.GetNamespace() == secret.Namespace {
				return true, secret, fmt.Errorf("failed to create secret")
			}
			return false, nil, nil
		})
		_, err := KubeClientSet.CreateK8sSecret("default", secret)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("KubeClientSet is nil", t, func() {
		KubeClientSet = nil
		_, err := KubeClientSet.CreateK8sSecret("default", secret)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestKubeClient_UpdateK8sSecret(t *testing.T) {
	client := fake.NewSimpleClientset()
	KubeClientSet = &KubeClient{client}
	namespace := "default"
	secretName := "test_secret"
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Data: map[string][]byte{"key": []byte("value")},
	}
	_, err := client.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create secret: %v", err)
	}
	convey.Convey("update secret success", t, func() {
		_, err := KubeClientSet.UpdateK8sSecret(namespace, secret)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("update secret error", t, func() {
		client.PrependReactor("update", "secrets", func(action k8testing.Action) (handled bool, ret runtime.Object, err error) {
			updateAction := action.(k8testing.UpdateAction)
			if updateAction.GetObject().(*v1.Secret).Name == secret.Name && updateAction.GetNamespace() == secret.Namespace {
				return true, secret, fmt.Errorf("failed to update secret")
			}
			return false, nil, nil
		})
		_, err := KubeClientSet.UpdateK8sSecret(namespace, secret)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("KubeClientSet is nil", t, func() {
		KubeClientSet = nil
		_, err := KubeClientSet.UpdateK8sSecret(namespace, secret)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestKubeClient_DeleteK8sSecret(t *testing.T) {
	client := fake.NewSimpleClientset()
	KubeClientSet = &KubeClient{client}
	namespace := "default"
	secretName := "test_secret"
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Data: map[string][]byte{"key": []byte("value")},
	}

	convey.Convey("delete secret success", t, func() {
		err := KubeClientSet.DeleteK8sSecret(namespace, secretName)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("delete secret error", t, func() {
		client.PrependReactor("delete", "secrets", func(action k8testing.Action) (handled bool, ret runtime.Object, err error) {
			deleteAction := action.(k8testing.DeleteAction)
			if deleteAction.GetName() == secret.Name && deleteAction.GetNamespace() == secret.Namespace {
				return true, secret, fmt.Errorf("failed to delete service")
			}
			return false, nil, nil
		})
		err := KubeClientSet.DeleteK8sSecret(namespace, secretName)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("KubeClientSet is nil", t, func() {
		KubeClientSet = nil
		err := KubeClientSet.DeleteK8sSecret(namespace, secretName)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestKubeClient_CreateOrUpdateConfigMap(t *testing.T) {
	client := fake.NewSimpleClientset()
	KubeClientSet = &KubeClient{client}
	configmap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_configmap",
			Namespace: "default",
		},
		Data: map[string]string{"key": "value"},
	}
	configmap2 := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_configmap",
			Namespace: "default",
		},
		Data: map[string]string{"key": "value1"},
	}
	_, err := client.CoreV1().ConfigMaps("default").Create(context.TODO(), configmap2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create configmap: %v", err)
	}

	convey.Convey("create configmap success", t, func() {
		err := KubeClientSet.CreateOrUpdateConfigMap(configmap)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("KubeClientSet is nil", t, func() {
		KubeClientSet = nil
		err := KubeClientSet.CreateOrUpdateConfigMap(configmap)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func (m *mockK8sClient) InClusterConfig() (*rest.Config, error) {
	m.expectedConfigCalled = true
	return &rest.Config{}, m.createConfigError
}

func (m *mockK8sClient) NewForConfig(_ *rest.Config) (dynamic.Interface, error) {
	return nil, m.createClientError
}

func TestNewDynamicClient_Singleton(t *testing.T) {
	dynamicClient = nil
	dynamicClientOnce = sync.Once{}
	mock := &mockK8sClient{}
	oldInClusterConfig := inClusterConfigFunc
	inClusterConfigFunc = mock.InClusterConfig
	defer func() { inClusterConfigFunc = oldInClusterConfig }()
	client1 := NewDynamicClient()
	client2 := NewDynamicClient()
	if client1 != client2 {
		t.Error("Expected singleton instance, got different clients")
	}
}
