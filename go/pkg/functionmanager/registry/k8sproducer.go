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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/cache"
	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/functionmanager/types"
)

// K8sProducer define k8s watcher
type K8sProducer[T comparable] struct {
	gvrInformer cache.SharedIndexInformer

	convertFunc func(eventType types.EventType, unstructured *unstructured.Unstructured) (T, error)
	eventCh     chan T

	logger api.FormatLogger
}

func (p *K8sProducer[T]) init(stopCh chan struct{}) {
	_, err := p.gvrInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			p.putEventToChan(types.SubEventTypeUpdate, obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			p.putEventToChan(types.SubEventTypeUpdate, newObj)
		},
		DeleteFunc: func(obj interface{}) {
			p.putEventToChan(types.SubEventTypeDelete, obj)
		},
	})
	if err != nil {
		p.logger.Errorf("add pat event handler failed, err %s", err.Error())
		return
	}
	p.logger.Info("start to watch pat")
}

func (p *K8sProducer[T]) waitForSync(stopCh chan struct{}) {
	if !cache.WaitForCacheSync(stopCh, p.gvrInformer.HasSynced) {
		p.logger.Errorf("failed to sync cache")
		return
	}
	p.logger.Info("pat info synced")
}

func (p *K8sProducer[T]) putEventToChan(eventType types.EventType, obj interface{}) {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return
	}
	event, err := p.convertFunc(eventType, u)
	if err != nil {
		p.logger.Errorf("convert unstructured %s %s to event failed %s",
			u.GetNamespace(), u.GetName(), err.Error())
		return
	}
	p.eventCh <- event
}
