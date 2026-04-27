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

package workermanager

import (
	"errors"
	"sync"

	coordinationv1 "k8s.io/api/coordination/v1"
	"k8s.io/client-go/informers"
	informercdv1 "k8s.io/client-go/informers/coordination/v1"
	"k8s.io/client-go/tools/cache"

	"yuanrong.org/kernel/pkg/common/faas_common/k8sclient"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

var (
	leaseInformer informercdv1.LeaseInformer
	wmAddr        string
	wmAddrMutex   sync.RWMutex
)

func getWmAddr() string {
	wmAddrMutex.RLock()
	ret := wmAddr
	wmAddrMutex.RUnlock()
	return ret
}

func updateWmAddr(addr string) {
	wmAddrMutex.Lock()
	wmAddr = addr
	wmAddrMutex.Unlock()
}

// InitLeaseInformer -
func InitLeaseInformer(stopCh <-chan struct{}) error {

	informerFactory := informers.NewSharedInformerFactoryWithOptions(k8sclient.GetkubeClient().Client,
		0, informers.WithNamespace("default"))
	leaseInformer = informerFactory.Coordination().V1().Leases()

	leaseInformer.Informer().AddEventHandler(generateInformerHandler())

	go leaseInformer.Informer().Run(stopCh)
	if ok := cache.WaitForCacheSync(stopCh, leaseInformer.Informer().HasSynced); !ok {
		return errors.New("failed to sync")
	}
	return nil
}

func generateInformerHandler() cache.FilteringResourceEventHandler {
	return cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			lease, ok := obj.(*coordinationv1.Lease)
			if !ok {
				return false
			}

			if lease.Name == "worker-manager" {
				return true
			}

			return false
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				lease, ok := obj.(*coordinationv1.Lease)
				if !ok {
					return
				}
				if lease.Spec.HolderIdentity == nil {
					return
				}
				updateWmAddr(*lease.Spec.HolderIdentity)
				log.GetLogger().Infof("ip of worker-manger leader is: %s", *lease.Spec.HolderIdentity)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				lease, ok := newObj.(*coordinationv1.Lease)
				if !ok {
					return
				}
				if lease.Spec.HolderIdentity == nil {
					return
				}
				oldAddr := getWmAddr()
				if oldAddr != *lease.Spec.HolderIdentity {
					updateWmAddr(*lease.Spec.HolderIdentity)
					log.GetLogger().Infof("ip of worker-manager leader update to %s", *lease.Spec.HolderIdentity)
				}
			},
			DeleteFunc: func(obj interface{}) {
				log.GetLogger().Errorf("leader of worker-manager lost")
				updateWmAddr("")
			},
		},
	}
}
