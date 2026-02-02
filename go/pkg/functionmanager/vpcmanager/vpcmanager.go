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
	"encoding/json"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/k8sclient"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/functionmanager/types"
	"yuanrong.org/kernel/pkg/functionmanager/utils"
	"yuanrong.org/kernel/runtime/libruntime/api"
)

var patGVR = schema.GroupVersionResource{
	Group:    "patservice.cap.io",
	Version:  "v1",
	Resource: "pats",
}

var instanceStatusOfDeleted = map[constant.InstanceStatus]struct{}{
	constant.KernelInstanceStatusExited:         {},
	constant.KernelInstanceStatusFailed:         {},
	constant.KernelInstanceStatusFatal:          {},
	constant.KernelInstanceStatusScheduleFailed: {},
	constant.KernelInstanceStatusEvicted:        {},
}

// VPCManager manage vpc info
type VPCManager struct {
	EventCh chan types.VPCEvent

	deleteAfterDuration time.Duration

	patPodOfInstance map[k8stypes.NamespacedName]map[string]struct{} // key: patPodName  value: instanceId
	patMu            sync.RWMutex
	gvrLister        cache.GenericLister
	gvrClient        dynamic.NamespaceableResourceInterface
	waitDeleteQueue  workqueue.TypedDelayingInterface[k8stypes.NamespacedName]

	stopCh chan struct{}
	logger api.FormatLogger
}

// MakeVPCManager create vpc manager
func MakeVPCManager(informer informers.GenericInformer, deleteAfterDuration time.Duration, stopCh chan struct{}) *VPCManager {
	return &VPCManager{
		EventCh:             make(chan types.VPCEvent, 100),
		deleteAfterDuration: deleteAfterDuration,
		patPodOfInstance:    make(map[k8stypes.NamespacedName]map[string]struct{}),
		patMu:               sync.RWMutex{},
		gvrLister:           informer.Lister(),
		gvrClient:           k8sclient.GetDynamicClient().Resource(patGVR),
		waitDeleteQueue:     workqueue.TypedNewDelayingQueue[k8stypes.NamespacedName](),
		stopCh:              stopCh,
		logger:              log.GetLogger().With(zap.String("model", "vpcmanager")),
	}
}

// Run start vpc manager
func (v *VPCManager) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go v.processDeleteQueue(ctx)
	go v.reconcilePat(ctx)
	for {
		select {
		case e, ok := <-v.EventCh:
			if !ok {
				v.logger.Warnf("event chan closed, return")
				return
			}
			if e.InsInfo != nil {
				v.handleInstanceEvent(e)
			}
			if e.PatInfo != nil {
				v.handlePatCREvent(ctx, e)
			}
		case <-v.stopCh:
			v.waitDeleteQueue.ShutDown()
			v.logger.Info("vpc manager stop, return")
			return
		}
	}
}

func (v *VPCManager) handleInstanceEvent(event types.VPCEvent) {
	ins := event.InsInfo
	if ins == nil {
		return
	}
	logger := v.logger.With(zap.String("instance", ins.InstanceID))
	if ins.CreateOptions[types.NetWorkDelegateKey] == "" {
		logger.Debugf("skip ins event")
		return
	}
	networkConfig := &types.NetWorkConfig{}
	err := json.Unmarshal([]byte(ins.CreateOptions[types.NetWorkDelegateKey]), networkConfig)
	if err != nil {
		logger.Errorf("DELEGATE_NETWORK_CONFIG %s format error, %s",
			ins.CreateOptions["DELEGATE_NETWORK_CONFIG"], err.Error())
		return
	}
	v.patMu.Lock()
	defer v.patMu.Unlock()
	if event.EventType == types.SubEventTypeDelete || v.isDeadInstance(constant.InstanceStatus(ins.InstanceStatus.Code)) {
		logger.Infof("process instance delete event")
		for _, patPod := range networkConfig.PatInstances {
			delete(v.patPodOfInstance[patPod], ins.InstanceID)
			if len(v.patPodOfInstance[patPod]) == 0 {
				delete(v.patPodOfInstance, patPod)
				logger.Infof("%s has no instance", patPod.String())
			}
		}
		return
	}
	for _, patPod := range networkConfig.PatInstances {
		if _, ok := v.patPodOfInstance[patPod]; !ok {
			v.patPodOfInstance[patPod] = make(map[string]struct{})
		}
		v.patPodOfInstance[patPod][ins.InstanceID] = struct{}{}
		logger.Infof("%s add instance %s", patPod.String(), ins.InstanceID)
	}
}

func (v *VPCManager) handlePatCREvent(ctx context.Context, event types.VPCEvent) {
	if event.PatInfo == nil {
		return
	}
	logger := v.logger.With(zap.String("patCr", event.PatInfo.Name))
	logger.Infof("handle pat cr event")
	obj, err := v.gvrLister.ByNamespace(event.PatInfo.Namespace).Get(event.PatInfo.Name)
	if k8serror.IsNotFound(err) {
		logger.Infof("pat cr has been deleted")
		v.patMu.RLock()
		defer v.patMu.RUnlock()
		for _, pod := range event.PatInfo.Status.PatPods {
			patPod := k8stypes.NamespacedName{
				Namespace: event.PatInfo.Namespace,
				Name:      pod.PatPodName,
			}
			if len(v.patPodOfInstance[patPod]) != 0 {
				logger.Errorf("pat cr delete, but pat pod is still using by %v", v.patPodOfInstance[patPod])
			}
		}
		return
	}
	if err != nil {
		logger.Errorf("failed to get pat info, err %s, skip", err.Error())
		return
	}
	unstructedPat, ok := obj.(*unstructured.Unstructured)
	if !ok {
		logger.Errorf("pat info type error")
		return
	}
	patInfo, err := utils.UnstructuredToPat(unstructedPat)
	if err != nil {
		logger.Errorf("convert pat failed, %s", err.Error())
		return
	}
	if patInfo.Spec.RequireCount == 0 && len(patInfo.Status.PatPods) == 0 {
		logger.Infof("pat cr is empty, wait to delete")
		v.waitDeleteQueue.AddAfter(k8stypes.NamespacedName{
			Namespace: patInfo.Namespace,
			Name:      patInfo.Name,
		}, v.deleteAfterDuration)
		return
	}
	v.waitDeleteQueue.Done(k8stypes.NamespacedName{
		Namespace: patInfo.Namespace,
		Name:      patInfo.Name,
	})
	err = v.updatePatIdlePod(ctx, patInfo)
	if err != nil {
		logger.Errorf("update pat idle pod failed, err %s", err.Error())
	}
}

func (v *VPCManager) isDeadInstance(statusCode constant.InstanceStatus) bool {
	if _, ok := instanceStatusOfDeleted[statusCode]; ok {
		return true
	}
	return false
}

func (v *VPCManager) reconcilePat(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			v.processAllRunningPat(ctx)
		}
	}
}

func (v *VPCManager) processAllRunningPat(ctx context.Context) {
	patList, err := v.gvrLister.List(labels.Everything())
	if err != nil {
		v.logger.Errorf("list pat info failed, continue, err %s", err.Error())
		return
	}
	v.patMu.RLock()
	defer v.patMu.RUnlock()
	for _, object := range patList {
		u, ok := object.(*unstructured.Unstructured)
		if !ok {
			v.logger.Infof("convert type error")
			continue
		}
		patInfo, err := utils.UnstructuredToPat(u)
		if err != nil {
			v.logger.Errorf("conver to pat error %s", err.Error())
			continue
		}
		err = v.updatePatIdlePod(ctx, patInfo)
		if err != nil {
			v.logger.Errorf("update pat idle pod failed, err %s", err.Error())
		}
	}
}

func (v *VPCManager) processDeleteQueue(ctx context.Context) {
	for {
		pat, shutdown := v.waitDeleteQueue.Get()
		if shutdown {
			return
		}
		v.logger.Infof("start to delete pat %s", pat.String())
		canDel, err := v.isPatNeedDel(ctx, pat)
		if err != nil {
			v.waitDeleteQueue.AddAfter(pat, time.Minute)
		}
		if !canDel {
			v.logger.Infof("no need delete %s, err %v", pat.String(), err)
			v.waitDeleteQueue.Done(pat)
			continue
		}
		err = v.gvrClient.Namespace(pat.Namespace).Delete(ctx, pat.Name, metav1.DeleteOptions{})
		if err != nil && !k8serror.IsNotFound(err) {
			v.logger.Errorf("delete pat %s failed, retry after 1 min, %s", pat.String(), err.Error())
			v.waitDeleteQueue.AddAfter(pat, time.Minute)
		}
		v.logger.Infof("succeed delete %s", pat.String())
		v.waitDeleteQueue.Done(pat)
	}
}

func (v *VPCManager) isPatNeedDel(ctx context.Context, pat k8stypes.NamespacedName) (bool, error) {
	unstructedPat, err := v.gvrClient.Namespace(pat.Namespace).Get(ctx, pat.Name, metav1.GetOptions{})
	if k8serror.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	patInfo, err := utils.UnstructuredToPat(unstructedPat)
	if err != nil {
		return false, err
	}
	if patInfo.Spec.RequireCount != 0 || len(patInfo.Status.PatPods) != 0 {
		return false, nil
	}
	return true, nil
}

func (v *VPCManager) updatePatIdlePod(ctx context.Context, pat *types.Pat) error {
	if pat == nil || len(pat.Status.PatPods) == 0 {
		return nil
	}
	if pat.Annotations == nil {
		pat.Annotations = map[string]string{}
	}
	var originIdlePodList []string
	if idlePodStr, ok := pat.Annotations[types.PatIdlePodAnnotationKey]; ok && len(idlePodStr) > 0 {
		originIdlePodList = strings.Split(idlePodStr, ",")
	}
	runningPatPod := 0
	for _, pod := range pat.Status.PatPods {
		if pod.Status == types.PatPodRunningStatus {
			runningPatPod++
		}
	}
	if int(pat.Spec.RequireCount) > runningPatPod {
		v.logger.Infof("pat pod has not ready, skip")
		return nil
	}
	var idlePodList []string
	requireCount := 0
	for _, pod := range pat.Status.PatPods {
		patPod := k8stypes.NamespacedName{
			Namespace: pat.Namespace,
			Name:      pod.PatPodName,
		}
		if len(v.patPodOfInstance[patPod]) == 0 {
			idlePodList = append(idlePodList, pod.PatPodName)
		} else {
			requireCount++
		}
	}
	sort.Strings(idlePodList)
	if slices.Equal(idlePodList, originIdlePodList) {
		return nil
	}
	pat.Annotations[types.PatIdlePodAnnotationKey] = strings.Join(idlePodList, ",")
	pat.Spec.RequireCount = int64(requireCount)
	targetPat, err := utils.PatToUnstructured(pat)
	if err != nil {
		v.logger.Errorf("convert pat failed %s", err.Error())
		return err
	}
	v.logger.Infof("update pat idle list from %v to %v", originIdlePodList, idlePodList)
	_, err = v.gvrClient.Namespace(pat.Namespace).Update(ctx, targetPat, metav1.UpdateOptions{})
	return err
}
