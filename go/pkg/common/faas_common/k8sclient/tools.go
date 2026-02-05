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

// Package k8sclient include some k8s Client operation
package k8sclient

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

// KubeClient -
type KubeClient struct {
	Client kubernetes.Interface
}

var (
	// KubeClientSet -
	KubeClientSet  *KubeClient
	kubeClientOnce sync.Once

	// dynamicClient -
	dynamicClient     dynamic.Interface
	dynamicClientOnce sync.Once
)

// GetkubeClient is used to obtain a K8S Client
func GetkubeClient() *KubeClient {
	kubeClientOnce.Do(func() {
		// create Kubernetes config
		config, err := rest.InClusterConfig()
		if err != nil {
			log.GetLogger().Errorf("Failed to create Kubernetes config: %v", err)
			return
		}

		// create Kubernetes Client
		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			log.GetLogger().Errorf("Failed to create Kubernetes Client: %v", err)
			return
		}

		KubeClientSet = &KubeClient{
			Client: client,
		}
	})
	return KubeClientSet
}

// GetDynamicClient Get Dynamic Client
func GetDynamicClient() dynamic.Interface {
	dynamicClientOnce.Do(func() {
		// create Kubernetes config
		config, err := rest.InClusterConfig()
		if err != nil {
			log.GetLogger().Errorf("Failed to create Kubernetes config: %s", err.Error())
			return
		}
		// create dynamic client
		dynamicClient, err = dynamic.NewForConfig(config)
		if err != nil {
			log.GetLogger().Errorf("failed to create dynamic client: %s", err.Error())
			return
		}
	})
	return dynamicClient
}

// DeleteK8sService -
func (kc *KubeClient) DeleteK8sService(namespace string, serviceName string) error {
	if kc == nil {
		return fmt.Errorf("kubeclient is nil")
	}
	err := kc.Client.CoreV1().Services(namespace).Delete(context.TODO(), serviceName, metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.GetLogger().Infof("Service %s in namespace %s not found", serviceName, namespace)
			return nil
		}
		return err
	}
	log.GetLogger().Infof("Service %s in namespace %s deleted", serviceName, namespace)
	return nil
}

// CreateK8sService -
func (kc *KubeClient) CreateK8sService(service *v1.Service) error {
	if kc == nil {
		return fmt.Errorf("kubeclient is nil")
	}

	// delete service
	if err := kc.DeleteK8sService(service.Namespace, service.Name); err != nil {
		return err
	}
	// create Service
	result, err := kc.Client.CoreV1().Services(service.Namespace).Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create Service: %s", err.Error())
	}

	log.GetLogger().Infof("created Service %q with IP %q", result.GetObjectMeta().GetName(), result.Spec.ClusterIP)
	return nil
}

// CreateK8sConfigMap -
func (kc *KubeClient) CreateK8sConfigMap(configMap *v1.ConfigMap) error {
	if kc == nil {
		return fmt.Errorf("kubeclient is nil")
	}

	// delete configMap
	if err := kc.DeleteK8sConfigMap(configMap.Namespace, configMap.Name); err != nil {
		return err
	}
	// create configMap
	result, err := kc.Client.CoreV1().ConfigMaps(configMap.Namespace).Create(context.TODO(),
		configMap, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create ConfigMap: %s", err.Error())
	}

	log.GetLogger().Infof("created ConfigMap: %s", result.GetObjectMeta().GetName())
	return nil
}

// DeleteK8sConfigMap -
func (kc *KubeClient) DeleteK8sConfigMap(namespace string, configMapName string) error {
	if kc == nil {
		return fmt.Errorf("kubeclient is nil")
	}

	// delete configMap
	err := kc.Client.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), configMapName, metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.GetLogger().Infof("configMap %s in namespace %s not found", configMapName, namespace)
			return nil
		}
		return err
	}
	log.GetLogger().Infof("configMap %s in namespace %s deleted", configMapName, namespace)
	return nil
}

// UpdateK8sConfigMap -
func (kc *KubeClient) UpdateK8sConfigMap(configMap *v1.ConfigMap) error {
	if kc == nil {
		return fmt.Errorf("kubeclient is nil")
	}
	_, err := kc.Client.CoreV1().ConfigMaps(configMap.Namespace).Get(context.TODO(), configMap.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.GetLogger().Infof("configMap %s in namespace %s not found", configMap.Name, configMap.Namespace)
		}
		return err
	}
	_, err = kc.Client.CoreV1().ConfigMaps(configMap.Namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
	if err != nil {
		log.GetLogger().Errorf("update configmap failed, error is %s", err.Error())
		return err
	}
	log.GetLogger().Infof("configMap %s in namespace %s updated", configMap.Name, configMap.Namespace)
	return nil
}

// GetK8sConfigMap -
func (kc *KubeClient) GetK8sConfigMap(namespace string, configMapName string) (*v1.ConfigMap, error) {
	if kc == nil {
		return nil, fmt.Errorf("kubeclient is nil")
	}
	configmap, err := kc.Client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.GetLogger().Infof("configMap %s in namespace %s not found", configMapName, namespace)
		}
		return nil, err
	}
	log.GetLogger().Infof("Get configMap %s in namespace %s updated", configMapName, namespace)
	return configmap, nil
}

// GetK8sSecret -
func (kc *KubeClient) GetK8sSecret(namespace string, secretName string) (*v1.Secret, error) {
	if kc == nil {
		return nil, fmt.Errorf("kubeclient is nil")
	}
	ctx := context.TODO()
	secret, err := kc.Client.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.GetLogger().Infof("secret %s not found", secretName)
			return nil, err
		}
		log.GetLogger().Errorf("secret %s get failed, err is: %s", secretName, err)
		return nil, err
	}
	log.GetLogger().Errorf("secret %s already exists, no need create.", secretName)
	return secret, nil
}

// CreateK8sSecret -
func (kc *KubeClient) CreateK8sSecret(namespace string, s *v1.Secret) (*v1.Secret, error) {
	if kc == nil {
		return nil, fmt.Errorf("kubeclient is nil")
	}
	ctx := context.TODO()
	secret, err := kc.Client.CoreV1().Secrets(namespace).Create(ctx, s, metav1.CreateOptions{})
	if err != nil {
		log.GetLogger().Errorf("k8s failed to create secret: %s, secretName: %s", err.Error(), s.Name)
		return nil, err
	}
	log.GetLogger().Infof("secret %s in namespace %s created", secret.Name, namespace)

	return secret, nil
}

// UpdateK8sSecret -
func (kc *KubeClient) UpdateK8sSecret(namespace string, s *v1.Secret) (*v1.Secret, error) {
	if kc == nil {
		return nil, fmt.Errorf("kubeclient is nil")
	}
	ctx := context.TODO()
	secret, err := kc.Client.CoreV1().Secrets(namespace).Update(ctx, s, metav1.UpdateOptions{})
	if err != nil {
		log.GetLogger().Errorf("k8s failed to update secret: %s, secretName: %s", err.Error(), s.Name)
		return nil, err
	}
	log.GetLogger().Infof("secret %s in namespace %s updated", secret.Name, namespace)

	return secret, nil
}

// DeleteK8sSecret -
func (kc *KubeClient) DeleteK8sSecret(namespace string, secretName string) error {
	if kc == nil {
		return fmt.Errorf("kubeclient is nil")
	}
	ctx := context.TODO()
	err := kc.Client.CoreV1().Secrets(namespace).Delete(ctx, secretName, metav1.DeleteOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			log.GetLogger().Infof("secret %s in namespace %s not found", secretName, namespace)
			return nil
		}
		log.GetLogger().Errorf("k8s failed to delete secret: %s, secretName: %s", err.Error(), secretName)
		return err
	}
	log.GetLogger().Infof("secret %s in namespace %s deleted successfully", secretName, namespace)

	return nil
}

// CreateOrUpdateConfigMap -
func (kc *KubeClient) CreateOrUpdateConfigMap(c *v1.ConfigMap) error {
	if kc == nil {
		return fmt.Errorf("kubeclient is nil")
	}
	ctx := context.TODO()
	oldConfig, getErr := kc.Client.CoreV1().ConfigMaps(c.Namespace).Get(ctx, c.Name, metav1.GetOptions{})
	if getErr != nil && errors.IsNotFound(getErr) {
		log.GetLogger().Infof("Creating a new Configmap, Configmap.Name: %s", c.Name)
		_, createErr := kc.Client.CoreV1().ConfigMaps(c.Namespace).Create(ctx, c, metav1.CreateOptions{})
		if createErr != nil {
			log.GetLogger().Errorf("k8s failed to create configmap: %s, traceID: %s",
				createErr.Error(), "TraceID")
			return createErr
		}
		return nil
	}
	if getErr != nil {
		log.GetLogger().Errorf("failed to get configmap: %s, err:%v", c.Name, getErr.Error())
		return getErr
	}

	if !reflect.DeepEqual(oldConfig, c) {
		_, updateErr := kc.Client.CoreV1().ConfigMaps(c.Namespace).Update(ctx, c, metav1.UpdateOptions{})
		if updateErr != nil {
			log.GetLogger().Errorf("k8s failed to update configmap: %s, traceID: %s", updateErr.Error(),
				"TraceID")
			return updateErr
		}
	}
	return nil
}
