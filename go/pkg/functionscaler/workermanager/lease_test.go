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
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/coordination/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"yuanrong.org/kernel/pkg/common/faas_common/k8sclient"
)

func Test_initLeaseInformer(t *testing.T) {
	addr1 := "10.0.0.1:58866"
	leases := []*v1.Lease{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "state-manager"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "worker-manager"},
			Spec:       v1.LeaseSpec{HolderIdentity: &addr1},
		},
	}

	fakeClient := fake.NewSimpleClientset(leases[0], leases[1])
	fakeInformer := informers.NewSharedInformerFactory(fakeClient, 0)

	patches := []*gomonkey.Patches{
		gomonkey.ApplyFunc(k8sclient.GetkubeClient, func() *k8sclient.KubeClient {
			return &k8sclient.KubeClient{
				Client: &kubernetes.Clientset{},
			}
		}),
		gomonkey.ApplyFunc(informers.NewSharedInformerFactoryWithOptions, func(client kubernetes.Interface, defaultResync time.Duration, options ...informers.SharedInformerOption) informers.SharedInformerFactory {
			return fakeInformer
		}),
	}
	defer func() {
		for i := range patches {
			patches[i].Reset()
		}
	}()

	stopCh := make(chan struct{})
	err := InitLeaseInformer(stopCh)
	assert.Nil(t, err)
	close(stopCh)
}

func Test_generateInformerHandler(t *testing.T) {
	addr1 := "10.0.0.1:58866"
	addr2 := "10.0.0.2:58866"
	addr3 := "10.0.0.3:58866"
	emptyLease := &v1.Lease{
		ObjectMeta: metav1.ObjectMeta{Name: "worker-manager"},
	}
	lease1 := &v1.Lease{
		ObjectMeta: metav1.ObjectMeta{Name: "worker-manager"},
		Spec:       v1.LeaseSpec{HolderIdentity: &addr1},
	}
	lease2 := &v1.Lease{
		ObjectMeta: metav1.ObjectMeta{Name: "worker-manager"},
		Spec:       v1.LeaseSpec{HolderIdentity: &addr2},
	}

	stateLease := &v1.Lease{
		ObjectMeta: metav1.ObjectMeta{Name: "state-manager"},
		Spec:       v1.LeaseSpec{HolderIdentity: &addr3},
	}

	updateWmAddr("")
	handlers := generateInformerHandler()
	handlers.OnAdd(*lease1, true)
	assert.Equal(t, "", wmAddr)
	handlers.OnAdd(emptyLease, true)
	assert.Equal(t, "", wmAddr)
	handlers.OnAdd(lease1, true)
	assert.Equal(t, addr1, wmAddr)
	handlers.OnAdd(stateLease, true)
	assert.Equal(t, addr1, wmAddr)

	filter := handlers.FilterFunc
	handlers.FilterFunc = func(obj interface{}) bool {
		return true
	}
	handlers.OnUpdate(lease1, *lease1)
	assert.Equal(t, addr1, wmAddr)
	handlers.OnUpdate(lease1, emptyLease)
	assert.Equal(t, addr1, wmAddr)
	handlers.OnUpdate(lease1, lease2)
	assert.Equal(t, addr2, wmAddr)

	handlers.FilterFunc = filter
	handlers.OnDelete(stateLease)
	assert.Equal(t, addr2, wmAddr)
	handlers.OnDelete(lease2)
	assert.Equal(t, "", wmAddr)
}
