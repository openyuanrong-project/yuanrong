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

// Package functionmanager -
package functionmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/k8sclient"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	commonType "yuanrong.org/kernel/pkg/common/faas_common/types"
	commonUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/functionmanager/config"
	"yuanrong.org/kernel/pkg/functionmanager/registry"
	"yuanrong.org/kernel/pkg/functionmanager/state"
	"yuanrong.org/kernel/pkg/functionmanager/types"
	"yuanrong.org/kernel/pkg/functionmanager/utils"
	"yuanrong.org/kernel/pkg/functionmanager/vpcmanager"
)

// Manager manages functions for faas pattern
type Manager struct {
	vpcManager        *vpcmanager.VPCManager
	patLister         cache.GenericLister
	remoteClientLease map[string]*leaseTimer
	remoteClientList  map[string]struct{}
	clientMutex       sync.Mutex
	stopCh            chan struct{}
	queue             *queue
	leaseRenewMinute  time.Duration
}

type leaseTimer struct {
	Timer  *time.Timer
	Cancel context.CancelFunc
	Ctx    context.Context
}

// RequestOperation defines request operations
type RequestOperation string

var (
	// requestOpCreate stands for create pat-service operation
	requestOpCreate RequestOperation = "CreatePATService"
	// requestOpUnknown stands for unknown operation
	requestOpUnknown RequestOperation = "Unknown"
	// requestOpNewLease stands for create lease operation
	requestOpNewLease RequestOperation = constant.NewLease
	// requestOpKeepAlive stands for keep-alive lease operation
	requestOpKeepAlive RequestOperation = constant.KeepAlive
	// requestOpDelLease stands for delete lease operation
	requestOpDelLease RequestOperation = constant.DelLease
	libruntimeClient  api.LibruntimeAPI

	defaultCreatePatPodTimeout = time.Minute

	patGVR = schema.GroupVersionResource{
		Group:    "patservice.cap.io",
		Version:  "v1",
		Resource: "pats",
	}
)

const (
	validArgsNum                     = 3
	validArgsNumLibruntime           = 4
	faasManagerOpIndex               = 0
	faasManagerOpDataIndex           = 1
	faasManagerOpIndexLibruntime     = 1
	faasManagerOpDataIndexLibruntime = 2

	leaseEtcdKeyLen = 3

	defaultLeaseRenewMinute = 5

	defaultDeletePatAfterPatPodEmpty = 10 * time.Minute
)

// NewFaaSManagerLibruntime will create a new faas functions manager
func NewFaaSManagerLibruntime(libruntimeAPI api.LibruntimeAPI, stopCh chan struct{}) (*Manager, error) {
	libruntimeClient = libruntimeAPI
	return MakeFaasManager(stopCh)
}

// MakeFaasManager will create a new faas functions manager
func MakeFaasManager(stopCh chan struct{}) (*Manager, error) {
	cfg := config.GetConfig()
	leaseRenewMinute := cfg.LeaseRenewMinute
	if leaseRenewMinute == 0 {
		leaseRenewMinute = defaultLeaseRenewMinute
	}
	factory := dynamicinformer.NewDynamicSharedInformerFactory(k8sclient.GetDynamicClient(), time.Minute)
	informer := factory.ForResource(patGVR)
	faaSManager := &Manager{
		vpcManager:        vpcmanager.MakeVPCManager(informer, defaultDeletePatAfterPatPodEmpty, stopCh),
		patLister:         informer.Lister(),
		remoteClientLease: make(map[string]*leaseTimer),
		remoteClientList:  map[string]struct{}{},
		clientMutex:       sync.Mutex{},
		stopCh:            stopCh,
		queue: &queue{
			cond:       sync.NewCond(&sync.RWMutex{}),
			dirty:      map[interface{}]struct{}{},
			processing: map[interface{}]struct{}{},
		},
		leaseRenewMinute: time.Duration(cfg.LeaseRenewMinute),
	}
	go faaSManager.saveStateLoop()
	if cfg.EnableVPCManage {
		go faaSManager.vpcManager.Run()
		go informer.Informer().Run(stopCh)
		registry.StartWatchEvent(faaSManager.vpcManager.EventCh, stopCh, informer)
	}
	return faaSManager, nil
}

// RecoverData -
func (m *Manager) RecoverData() {
	rcm := map[string]struct{}{}
	err := commonUtils.DeepCopyObj(state.GetState().RemoteClientList, &rcm)
	if err != nil {
		log.GetLogger().Errorf("recovered remoteClientList error: %v", err)
		return
	}
	log.GetLogger().Infof("now recover faaSManager remoteClientList")
	m.clientMutex.Lock()
	for client := range rcm {
		lease := newLeaseTimer(m.leaseRenewMinute * time.Minute)
		m.remoteClientLease[client] = lease
		m.remoteClientList[client] = struct{}{}
		go startLeaseTimeOutWatcher(lease.Ctx, m, lease.Timer, client, client)
	}
	log.GetLogger().Infof("recovered remoteClientList: %v", m.remoteClientList)
	m.clientMutex.Unlock()
}

func newLeaseTimer(timeout time.Duration) *leaseTimer {
	timer := time.NewTimer(timeout)
	ctx, cancel := context.WithCancel(context.Background())
	return &leaseTimer{
		Timer:  timer,
		Cancel: cancel,
		Ctx:    ctx,
	}
}

func startLeaseTimeOutWatcher(ctx context.Context, fm *Manager, timer *time.Timer, clientID string, traceID string) {
	select {
	case <-timer.C:
		log.GetLogger().Infof("lease timeout, traceID: %s, remoteClientID: %s", traceID, clientID)
		if fm != nil {
			fm.clearLease(clientID, traceID)
		}
	case <-ctx.Done():
		log.GetLogger().Infof("lease stopped before timeout, traceID: %s, remoteClientID: %s", traceID,
			clientID)
		timer.Stop()
	}
}

// HandlerRequest -
func (m *Manager) HandlerRequest(requestOp RequestOperation, requestData []byte,
	traceID string,
) *commonType.CallHandlerResponse {
	switch requestOp {
	case requestOpCreate:
		return m.handleRequestOpCreate(requestData, traceID)
	case requestOpNewLease:
		return m.handleNewLease(requestData, traceID, 0)
	case requestOpKeepAlive:
		return m.handleKeepAlive(requestData, traceID)
	case requestOpDelLease:
		return m.handleDelLease(requestData, traceID)
	default:
		log.GetLogger().Warnf("unknown request operation %s, traceID %s", requestOp, traceID)
	}
	return utils.GenerateErrorResponse(constant.UnsupportedOperationErrorCode,
		constant.UnsupportedOperationErrorMessage)
}

// ProcessSchedulerRequestLibruntime will handle create, report and delete of instance
func (m *Manager) ProcessSchedulerRequestLibruntime(args []api.Arg, traceID string) *commonType.CallHandlerResponse {
	requestOp, requestData := parseRequestOperationLibruntime(args)
	return m.HandlerRequest(requestOp, requestData, traceID)
}

// ProcessSchedulerRequest will handle create, report and delete of instance
func (m *Manager) ProcessSchedulerRequest(args []*api.Arg, traceID string) *commonType.CallHandlerResponse {
	requestOp, requestData := parseRequestOperation(args)
	return m.HandlerRequest(requestOp, requestData, traceID)
}

func (m *Manager) handleRequestOpCreate(requestData []byte, traceID string) *commonType.CallHandlerResponse {
	ctx, cancel := context.WithTimeout(context.Background(), defaultCreatePatPodTimeout)
	defer cancel()
	patRequest := commonType.PATServiceRequest{}
	logger := log.GetLogger().With(zap.Any("traceID", traceID))
	logger.Debugf("handle pat create request, %s", string(requestData))
	err := json.Unmarshal(requestData, &patRequest)
	if err != nil {
		logger.Errorf("pat info unmarshal failed %s", err.Error())
		return utils.GenerateErrorResponse(constant.AcquireVPCPatInfoErrorCode, "requestData format error")
	}
	runningPatPods, err := m.getRunningPatPod(ctx, patRequest)
	if err != nil && k8serror.IsNotFound(err) {
		_, err = m.createPat(ctx, patRequest)
		if err != nil {
			logger.Errorf("create pat failed, err %s", err.Error())
			return utils.GenerateErrorResponse(constant.AcquireVPCPatInfoErrorCode, "pat cr create failed")
		}
	}
	if len(runningPatPods) == 0 {
		runningPatPods, err = m.getRunningPatPodWithRetry(ctx, patRequest, logger)
	}
	if err != nil {
		logger.Errorf("query patpod failed, err %s", err.Error())
		return utils.GenerateErrorResponse(constant.AcquireVPCPatInfoErrorCode, "query pat pod info failed")
	}
	patInfoByte, err := json.Marshal(commonType.PatResponseInfo{PatPods: runningPatPods})
	if err != nil {
		logger.Warnf("unmarshal pat pod info failed, err", err.Error())
		return utils.GenerateErrorResponse(constant.AcquireVPCPatInfoErrorCode, "unmarshal pat pod info failed")
	}
	logger.Infof("create pat success, %s", string(patInfoByte))
	return utils.GenerateSuccessResponse(constant.InsReqSuccessCode, string(patInfoByte))
}

func (m *Manager) getRunningPatPodWithRetry(ctx context.Context, patRequest commonType.PATServiceRequest,
	logger api.FormatLogger,
) ([]commonType.NATConfigure, error) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			logger.Warnf("query pat pod timeout")
			return nil, fmt.Errorf("query pat pod timeout")
		case <-ticker.C:
			runningPatPod, err := m.getRunningPatPod(ctx, patRequest)
			if err != nil {
				logger.Warnf("query pat pod info failed, err: %s", err.Error())
				continue
			}
			if len(runningPatPod) > 0 {
				return runningPatPod, nil
			}
		}
	}
}

func (m *Manager) getRunningPatPod(ctx context.Context,
	patRequest commonType.PATServiceRequest,
) ([]commonType.NATConfigure, error) {
	var runningPatPods []commonType.NATConfigure
	obj, err := m.patLister.ByNamespace(patRequest.Namespace).
		Get(utils.GetPatName(patRequest.SubnetID, patRequest.SecurityGroups))
	if err != nil {
		return nil, err
	}
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("failed to get pat info, type error")
	}
	patInfo, err := utils.UnstructuredToPat(u)
	if err != nil {
		return nil, err
	}
	for _, pod := range patInfo.Status.PatPods {
		if pod.Status == types.PatPodRunningStatus {
			runningPatPods = append(runningPatPods, commonType.NATConfigure{
				Namespace:      u.GetNamespace(),
				PatContainerIP: pod.PatContainerIp,
				PatVMIP:        pod.PatVmIp,
				PatPortIP:      pod.PatPortIp,
				PatMacAddr:     pod.PatMacAddr,
				PatGateway:     pod.PatGateway,
				PatPodName:     pod.PatPodName,
				TenantCidr:     pod.TenantCidr,
				SubMetaDigest:  pod.SubMetaDigest,
			})
		}
	}
	if len(runningPatPods) > 0 {
		return runningPatPods, nil
	}
	return nil, fmt.Errorf("pat pod is empty")
}

func (m *Manager) createPat(ctx context.Context,
	patRequest commonType.PATServiceRequest,
) (*unstructured.Unstructured, error) {
	pat := &types.Pat{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pat",
			APIVersion: "patservice.cap.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        utils.GetPatName(patRequest.SubnetID, patRequest.SecurityGroups),
			Namespace:   patRequest.Namespace,
			Annotations: nil,
		},
		Spec: types.PatSpec{
			KeepAliveTimeout: 60, // default pat pod live time after idle
			DelegateRole:     patRequest.Xrole,
			RequireCount:     2, // default pat pod replica
			EnvironmentId:    patRequest.Namespace,
			Vpc: types.VPC{
				DomainId:       patRequest.DomainID,
				ProjectId:      patRequest.ProjectID,
				VpcId:          patRequest.VpcID,
				SubnetId:       patRequest.SubnetID,
				TenantCidr:     patRequest.TenantCidr,
				HostVmCidr:     patRequest.HostVMCidr,
				Gateway:        patRequest.Gateway,
				SecurityGroups: patRequest.SecurityGroups,
			},
		},
		Status: types.PatStatus{},
	}
	if pat.Spec.Vpc.SecurityGroups == nil {
		pat.Spec.Vpc.SecurityGroups = []string{}
	}
	patInfo, err := utils.PatToUnstructured(pat)
	patInfo, err = k8sclient.GetDynamicClient().Resource(patGVR).
		Namespace(patRequest.Namespace).Create(ctx, patInfo, metav1.CreateOptions{})
	if err != nil && !k8serror.IsAlreadyExists(err) {
		return nil, err
	}
	return patInfo, nil
}

func (m *Manager) handleNewLease(requestData []byte, traceID string, timeout int64) *commonType.CallHandlerResponse {
	remoteClientID := string(requestData)
	log.GetLogger().Infof("receive new lease from faas-frontend, traceID: %s", traceID)
	m.clientMutex.Lock()
	if _, value := m.remoteClientLease[remoteClientID]; value {
		m.clientMutex.Unlock()
		log.GetLogger().Infof("lease already existed, traceID: %s", traceID)
		return &commonType.CallHandlerResponse{Code: constant.InsReqSuccessCode, Message: "lease existed"}
	}
	t := time.Minute * m.leaseRenewMinute
	if timeout > 0 {
		t = time.Duration(timeout) * time.Second
	}
	lease := newLeaseTimer(t)
	m.remoteClientLease[remoteClientID] = lease
	m.remoteClientList[remoteClientID] = struct{}{}
	m.clientMutex.Unlock()
	m.saveStateData(traceID)

	go startLeaseTimeOutWatcher(lease.Ctx, m, lease.Timer, remoteClientID, traceID)

	log.GetLogger().Infof("succeed to create lease, traceID: %s", traceID)
	return &commonType.CallHandlerResponse{Code: constant.InsReqSuccessCode, Message: "Succeed to create lease"}
}

func (m *Manager) clearLease(clientID string, traceID string) {
	m.clientMutex.Lock()
	lease, ok := m.remoteClientLease[clientID]
	if !ok {
		log.GetLogger().Warnf("client id not existed, %s %s ", clientID, traceID)
		m.clientMutex.Unlock()
		killInstanceOuter(clientID, traceID)
		return
	}
	lease.Timer.Stop()
	lease.Cancel()
	delete(m.remoteClientLease, clientID)
	delete(m.remoteClientList, clientID)
	m.clientMutex.Unlock()
	m.saveStateData(traceID)
	killInstanceOuter(clientID, traceID)
}

type stateItem struct {
	ch      chan error
	traceID string
}

// String -
func (s *stateItem) String() string {
	return s.traceID
}

// Done -
func (s *stateItem) Done(err error) {
	s.ch <- err
}

func (m *Manager) saveStateData(traceID string) {
	i := &stateItem{
		ch:      make(chan error, 1),
		traceID: traceID,
	}
	if !m.queue.add(i) {
		log.GetLogger().Warnf("add save state req to queue failed, traceID: %s", traceID)
		return
	}
	err := <-i.ch
	if err != nil {
		log.GetLogger().Warnf("save state data failed, err:%v, traceID:%s", err, traceID)
	}
}

func (m *Manager) saveStateLoop() {
	for {
		all, shutdown := m.queue.getAll()
		if shutdown {
			return
		}
		m.clientMutex.Lock()
		managerState := state.GetState()
		state.ManagerStateLock.Lock()
		managerState.RemoteClientList = m.remoteClientList
		marshal, err := json.Marshal(managerState)
		state.ManagerStateLock.Unlock()
		if err != nil {
			m.queue.doneAll(all)
			for _, it := range all {
				it.Done(err)
			}
			continue
		}
		m.clientMutex.Unlock()
		client := etcd3.GetRouterEtcdClient()
		if client != nil {
			ctx, cancel := context.WithTimeout(context.Background(), etcd3.DurationContextTimeout)
			_, err = client.Client.Put(ctx, "/faas/state/recover/faasmanager", string(marshal))
			cancel()
		} else {
			err = fmt.Errorf("router etcd client is nil")
		}
		m.queue.doneAll(all)
		for _, it := range all {
			it.Done(err)
		}
	}
}

func killInstanceOuter(clientID string, traceID string) {
	log.GetLogger().Infof("start to kill instance outer, clientID: %s, traceID:%s", clientID, traceID)
	if err := libruntimeClient.Kill(clientID, types.KillSignalVal, []byte{}); err != nil {
		log.GetLogger().Warnf("failed to clean instances when delete lease, traceID: %s, "+
			"remoteClientID:%s, status:%s", traceID, clientID, err.Error())
	}
	// libruntimeClient.SetTraceID(traceID)
	// if err := libruntimeClient.ReleaseGRefs(clientID); err != nil {
	// 	log.GetLogger().Warnf("failed to release refs when delete lease, traceID: %s, "+
	// 		"remoteClientID:%s, status:%s", traceID, clientID, err.Error())
	// }
}

func (m *Manager) handleKeepAlive(requestData []byte, traceID string) *commonType.CallHandlerResponse {
	remoteClientID := string(requestData)
	log.GetLogger().Infof("receive keep-alive lease from faas-frontend, traceID: %s", traceID)
	m.clientMutex.Lock()
	if _, value := m.remoteClientLease[remoteClientID]; value {
		lease := m.remoteClientLease[remoteClientID]
		lease.Timer.Reset(m.leaseRenewMinute * time.Minute)
		m.clientMutex.Unlock()
		log.GetLogger().Infof("succeed to renew lease, traceID: %s", traceID)
		return &commonType.CallHandlerResponse{
			Code:    constant.InsReqSuccessCode,
			Message: "Succeed to renew lease",
		}
	}
	m.clientMutex.Unlock()
	log.GetLogger().Infof("failed to renew unknown lease, traceID: %s", traceID)
	return &commonType.CallHandlerResponse{
		Code:    constant.UnsupportedOperationErrorCode,
		Message: fmt.Sprintf("%s remote client id not exist", remoteClientID),
	}
}

func (m *Manager) handleDelLease(requestData []byte, traceID string) *commonType.CallHandlerResponse {
	remoteClientID := string(requestData)
	log.GetLogger().Infof("receive delete lease from faas-frontend, traceID: %s", traceID)
	m.clearLease(remoteClientID, traceID)

	return &commonType.CallHandlerResponse{
		Code:    constant.InsReqSuccessCode,
		Message: "Succeed to delete lease",
	}
}

func parseRequestOperationLibruntime(args []api.Arg) (RequestOperation, []byte) {
	requestOp := requestOpUnknown
	if len(args) != validArgsNumLibruntime {
		log.GetLogger().Errorf("invalid argument number")
		return requestOp, []byte{}
	}
	requestOp = RequestOperation(args[faasManagerOpIndexLibruntime].Data)
	requestData := args[faasManagerOpDataIndexLibruntime].Data
	return requestOp, requestData
}

func parseRequestOperation(args []*api.Arg) (RequestOperation, []byte) {
	requestOp := requestOpUnknown
	if len(args) != validArgsNum {
		log.GetLogger().Errorf("invalid argument number")
		return requestOp, []byte{}
	}
	requestOp = RequestOperation(args[faasManagerOpIndex].Data)
	requestData := args[faasManagerOpDataIndex].Data
	return requestOp, requestData
}

// WatchLeaseEvent -
func (m *Manager) WatchLeaseEvent() {
	etcdClient := etcd3.GetRouterEtcdClient()
	watcher := etcd3.NewEtcdWatcher(constant.LeasePrefix, func(event *etcd3.Event) bool {
		etcdKey := event.Key
		keyParts := strings.Split(etcdKey, constant.ETCDEventKeySeparator)
		return len(keyParts) == leaseEtcdKeyLen
	}, func(event *etcd3.Event) {
		log.GetLogger().Infof("handling lease event type %d, key:%s, remoteClientLease in use len: %d", event.Type,
			event.Key, len(m.remoteClientLease))
		switch event.Type {
		case etcd3.PUT:
			m.handleLeaseEvent(event)
			go func() {
				_, err := etcdClient.Client.Delete(context.Background(), event.Key)
				if err != nil {
					log.GetLogger().Errorf("delete lease event failed, key: %s, err: %v", event.Key, err)
				}
			}()
		case etcd3.DELETE:
		case etcd3.ERROR:
			log.GetLogger().Warnf("etcd error event: %s", event.Value)
		default:
			log.GetLogger().Warnf("unsupported event, key: %s", event.Key)
		}
	}, m.stopCh, etcdClient)
	watcher.StartWatch()
	select {
	case <-m.stopCh:
		m.queue.shutDown()
		return
	}
}

func (m *Manager) handleLeaseEvent(event *etcd3.Event) {
	if event == nil || len(event.Value) == 0 {
		log.GetLogger().Warnf("event is nil or value is empty")
		return
	}
	log.GetLogger().Infof("handling lease put event key:%s", event.Key)
	e := &commonType.LeaseEvent{}
	err := json.Unmarshal(event.Value, e)
	if err != nil {
		log.GetLogger().Errorf("error unmarshalling lease event: %s, err: %v", string(event.Value), err)
		return
	}
	switch e.Type {
	case constant.NewLease:
		m.handleNewLease([]byte(e.RemoteClientID), e.TraceID, 0)
	case constant.DelLease:
		m.handleDelLease([]byte(e.RemoteClientID), e.TraceID)
	case constant.KeepAlive:
		// 判断是否是过期心跳
		timeout := int64((m.leaseRenewMinute * time.Minute).Seconds()) - (time.Now().Unix() - e.Timestamp)
		if timeout <= 0 {
			log.GetLogger().Infof("lease is expired, traceID: %s", e.TraceID)
			m.handleDelLease([]byte(e.RemoteClientID), e.TraceID)
			return
		}
		m.clientMutex.Lock()
		if lease, value := m.remoteClientLease[e.RemoteClientID]; value {
			lease.Timer.Reset(time.Duration(timeout) * time.Second)
			m.remoteClientLease[e.RemoteClientID] = lease
			m.clientMutex.Unlock()
			log.GetLogger().Infof("succeed to renew lease, traceID: %s", e.TraceID)
			return
		}
		m.clientMutex.Unlock()
		m.handleNewLease([]byte(e.RemoteClientID), e.TraceID, timeout)
	default:
		log.GetLogger().Errorf("unexpected lease type: %s, traceId: %s", e.Type, e.TraceID)
	}
}
