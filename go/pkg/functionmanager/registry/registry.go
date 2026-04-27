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
	"encoding/json"
	"strings"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/informers"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	commonType "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionmanager/types"
	"yuanrong.org/kernel/pkg/functionmanager/utils"
)

const (
	validEtcdKeyLenForInstance = 14
	instanceKeyIndex           = 2
	tenantKeyIndex             = 5
	functionKeyIndex           = 7
	executorKeyIndex           = 8
)

// StartWatchEvent start watch k8s and etcd event
func StartWatchEvent(vpcEventCh chan types.VPCEvent, stopCh chan struct{}, informer informers.GenericInformer) {
	instanceProducer := EtcdProducer[types.VPCEvent]{
		client:      etcd3.GetRouterEtcdClient(),
		watchPrefix: "/sn/instance",
		watchFilter: instanceFilter,
		convertFunc: func(eventType types.EventType, event *etcd3.Event) (types.VPCEvent, error) {
			insSpec := &commonType.InstanceSpecification{}
			if len(event.Value) != 0 {
				err := json.Unmarshal(event.Value, insSpec)
				if err != nil {
					return types.VPCEvent{EventType: eventType}, err
				}
			} else if len(event.PrevValue) != 0 {
				err := json.Unmarshal(event.PrevValue, insSpec)
				if err != nil {
					return types.VPCEvent{EventType: eventType}, err
				}
			}
			return types.VPCEvent{
				EventType: eventType,
				InsInfo:   insSpec,
			}, nil
		},
		eventCh: vpcEventCh,
		logger:  log.GetLogger().With(zap.Any("model", "instanceProducer")),
	}
	instanceProducer.init(stopCh)
	instanceProducer.waitForSync(stopCh)
	patCrProducer := K8sProducer[types.VPCEvent]{
		gvrInformer: informer.Informer(),
		convertFunc: func(eventType types.EventType, unstructured *unstructured.Unstructured) (types.VPCEvent, error) {
			pat, err := utils.UnstructuredToPat(unstructured)
			if err != nil {
				return types.VPCEvent{}, err
			}
			return types.VPCEvent{
				EventType: eventType,
				PatInfo:   pat,
			}, nil
		},
		eventCh: vpcEventCh,
		logger:  log.GetLogger().With(zap.Any("model", "patCrProducer")),
	}
	patCrProducer.init(stopCh)
	patCrProducer.waitForSync(stopCh)
}

func instanceFilter(event *etcd3.Event) bool {
	items := strings.Split(event.Key, constant.ETCDEventKeySeparator)
	if len(items) != validEtcdKeyLenForInstance {
		return true
	}
	if items[instanceKeyIndex] != "instance" || items[tenantKeyIndex] != "tenant" ||
		items[functionKeyIndex] != "function" ||
		!strings.HasPrefix(items[executorKeyIndex], "0-system-faasExecutor") &&
			!strings.HasPrefix(items[executorKeyIndex], "0-system-serveExecutor") {
		return true
	}
	return false
}
