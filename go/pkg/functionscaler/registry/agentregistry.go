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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/k8sclient"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	commonType "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/common/faas_common/urnutils"
	"yuanrong.org/kernel/pkg/common/functioncr"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

const (
	// consistency with 1.x
	defaultMaxRetryTimes = 12

	defaultDoneChSize = 10
)

// AgentRegistry watches agent event of CR
type AgentRegistry struct {
	dynamicClient   dynamic.Interface
	informerFactory dynamicinformer.DynamicSharedInformerFactory
	workQueue       workqueue.RateLimitingInterface

	specLock      sync.RWMutex
	funcSpecs     map[string]*types.FunctionSpecification
	crToFuncSpecs map[string]*types.FunctionSpecification

	lister   cache.GenericLister
	informer cache.SharedInformer
	synced   bool

	crListDoneCh    chan struct{}
	subScriberChans []chan SubEvent
	stopCh          <-chan struct{}
	sync.RWMutex
}

var waitQueue = make(chan crdRawEvent, 10000)

// crdEvent include eventType and obj
type crdEvent struct {
	eventType EventType
	obj       *unstructured.Unstructured
}

type crdRawEvent struct {
	eventType    EventType
	objInterface interface{}
}

type crdSyncEvent struct {
	eventType EventType
	ch        chan struct{}
}

// NewAgentRegistry will create AgentRegistry
func NewAgentRegistry(stopCh <-chan struct{}) *AgentRegistry {
	// prevent component startup exceptions when the YAML file for deployment permissions is not configured
	if os.Getenv(constant.EnableAgentCRDRegistry) == "" {
		return nil
	}
	dynamicClient := k8sclient.GetDynamicClient()
	// Different CR events share the same rate-limiting queue
	workQueue := workqueue.NewNamedRateLimitingQueue(
		workqueue.DefaultControllerRateLimiter(),
		functioncr.CrdEventsQueue,
	)
	agentRegistry := &AgentRegistry{
		dynamicClient:   dynamicClient,
		informerFactory: dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, time.Minute),
		workQueue:       workQueue,
		specLock:        sync.RWMutex{},
		funcSpecs:       make(map[string]*types.FunctionSpecification, utils.DefaultMapSize),
		crToFuncSpecs:   make(map[string]*types.FunctionSpecification, utils.DefaultMapSize),
		crListDoneCh:    make(chan struct{}, defaultDoneChSize),
		subScriberChans: make([]chan SubEvent, 0, 1),
		stopCh:          stopCh,
	}
	return agentRegistry
}

// RunWatcher will start CR watch process
func (ar *AgentRegistry) RunWatcher() {
	if os.Getenv(constant.EnableAgentCRDRegistry) == "" {
		return
	}
	ar.informer = ar.informerFactory.ForResource(functioncr.GetCrdGVR()).Informer()
	ar.lister = ar.informerFactory.ForResource(functioncr.GetCrdGVR()).Lister()
	ar.setupEventHandlers(ar.informer)
	err := ar.initWatch(ar.informer)
	if err != nil {
		log.GetLogger().Errorf("init agent rigistry failed, err: %s", err.Error())
	}
}

// setupEventHandlers setup CRD Event Handlers
func (ar *AgentRegistry) setupEventHandlers(informer cache.SharedInformer) {
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ar.RLock()
			defer ar.RUnlock()
			if ar.synced {
				ar.enqueueEvent(SubEventTypeAdd, obj)
			} else {
				waitQueue <- crdRawEvent{
					eventType:    SubEventTypeAdd,
					objInterface: obj,
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if oldObj.(*unstructured.Unstructured).GetResourceVersion() !=
				newObj.(*unstructured.Unstructured).GetResourceVersion() {
				ar.RLock()
				defer ar.RUnlock()
				if ar.synced {
					ar.enqueueEvent(SubEventTypeUpdate, newObj)
				} else {
					waitQueue <- crdRawEvent{eventType: SubEventTypeUpdate, objInterface: newObj}
				}

			}
		},
		DeleteFunc: func(obj interface{}) {
			ar.RLock()
			defer ar.RUnlock()
			if ar.synced {
				ar.enqueueEvent(SubEventTypeDelete, obj)
			} else {
				waitQueue <- crdRawEvent{eventType: SubEventTypeDelete, objInterface: obj}
			}
		},
	})
}

// enqueueEvent handle crd event enqueue
func (ar *AgentRegistry) enqueueEvent(eventType EventType, objRaw interface{}) {
	if eventType == SubEventTypeSynced {
		ch, ok := objRaw.(chan struct{})
		if !ok {
			return
		}
		ar.workQueue.Add(&crdSyncEvent{eventType: SubEventTypeSynced, ch: ch})
		return
	}
	unstructObj, ok := objRaw.(*unstructured.Unstructured)
	if !ok {
		log.GetLogger().Errorf("failed to assert crd event")
		return
	}
	ar.workQueue.Add(&crdEvent{
		eventType: eventType,
		obj:       unstructObj,
	})
}

// initWatch start crd Controller
func (ar *AgentRegistry) initWatch(informer cache.SharedInformer) error {
	ctx, _ := context.WithCancel(context.Background())
	go ar.informerFactory.Start(ctx.Done())
	go ar.processQueue()
	if !cache.WaitForCacheSync(ctx.Done(), informer.HasSynced) {
		log.GetLogger().Warnf("failed to sync crd cache")
		return fmt.Errorf("wait for cache sync failed")
	}
	ar.Lock()
	defer ar.Unlock()
	ar.synced = true
	crRaws, err := ar.lister.List(labels.Everything())
	if err != nil {
		return nil
	}
	for _, crRaw := range crRaws {
		cr, ok := crRaw.(*unstructured.Unstructured)
		if ok {
			ar.enqueueEvent(SubEventTypeAdd, cr)
		}
	}

	waitQueueEventLenth := len(waitQueue)
	for i := 0; i < waitQueueEventLenth; i++ {
		event := <-waitQueue
		ar.enqueueEvent(event.eventType, event.objInterface)
	}
	ar.enqueueEvent(SubEventTypeSynced, ar.crListDoneCh)
	lenth := len(ar.subScriberChans)

	crListDoneChCount := 0
	for {
		if crListDoneChCount == lenth {
			break
		}
		select {
		case _, ok := <-ar.crListDoneCh:
			if !ok {
				return fmt.Errorf("listen to list done ch failed")
			}
			crListDoneChCount++
		}
	}
	return nil
}

// processQueue process crd event Queue
func (ar *AgentRegistry) processQueue() {
	for {
		item, shutdown := ar.workQueue.Get()
		if shutdown {
			return
		}
		syncEvent, ok := item.(*crdSyncEvent)
		if ok {
			log.GetLogger().Infof("recv cr sync event")
			ar.publishEvent(SubEventTypeSynced, syncEvent.ch)
			ar.workQueue.Forget(item)
			ar.workQueue.Done(item)
			continue
		}
		event, ok := item.(*crdEvent)
		if !ok {
			log.GetLogger().Warnf("invalid crd event")
			ar.workQueue.Forget(item)
			ar.workQueue.Done(item)
			continue
		}
		if err := ar.processEvent(event); err != nil {
			// Limited number of retries
			if ar.workQueue.NumRequeues(item) < defaultMaxRetryTimes {
				log.GetLogger().Warnf("process crd event error: %s, retry", err.Error())
				ar.workQueue.AddRateLimited(item)
			}
		} else {
			ar.workQueue.Forget(item)
		}
		ar.workQueue.Done(item)
	}
}

func (ar *AgentRegistry) addSubscriberChan(subChan chan SubEvent) {
	ar.Lock()
	ar.subScriberChans = append(ar.subScriberChans, subChan)
	ar.Unlock()
}

func (ar *AgentRegistry) publishEvent(eventType EventType, info interface{}) {
	for _, subChan := range ar.subScriberChans {
		if subChan != nil {
			subChan <- SubEvent{
				EventType: eventType,
				EventMsg:  info,
			}
		}
	}
}

func (ar *AgentRegistry) getFuncSpec(funcKey string) *types.FunctionSpecification {
	ar.specLock.RLock()
	defer ar.specLock.RUnlock()
	return ar.funcSpecs[funcKey]
}

func (ar *AgentRegistry) getFuncSpecFromCrName(crKey string) *types.FunctionSpecification {
	ar.specLock.RLock()
	defer ar.specLock.RUnlock()
	return ar.crToFuncSpecs[crKey]
}

func (ar *AgentRegistry) addFuncSpec(agentRunEventInfo *types.AgentEventInfo, namespace string) {
	ar.specLock.Lock()
	defer ar.specLock.Unlock()
	if agentRunEventInfo == nil || agentRunEventInfo.FuncSpec == nil {
		return
	}
	if namespace == "" {
		namespace = constant.DefaultNameSpace
	}
	agentRunEventInfo.FuncSpec.NameSpace = namespace
	ar.crToFuncSpecs[agentRunEventInfo.CrKey] = agentRunEventInfo.FuncSpec
	ar.funcSpecs[agentRunEventInfo.FuncSpec.FuncKey] = agentRunEventInfo.FuncSpec
}

func (ar *AgentRegistry) deleteFuncSpec(crKey string) {
	ar.specLock.Lock()
	defer ar.specLock.Unlock()
	funcSpec, ok := ar.crToFuncSpecs[crKey]
	delete(ar.crToFuncSpecs, crKey)
	if !ok {
		return
	}
	delete(ar.funcSpecs, funcSpec.FuncKey)
}

// processEvent process cr add update delete Event
func (ar *AgentRegistry) processEvent(event *crdEvent) error {
	crName := event.obj.GetName()
	namespace := event.obj.GetNamespace()
	crKey := namespace + utils.CrKeySep + crName
	logger := log.GetLogger().With(zap.Any("namespace", namespace), zap.Any("crName", crName))
	objRaw, err := ar.lister.ByNamespace(namespace).Get(crName)
	if err != nil && !k8serrors.IsNotFound(err) {
		log.GetLogger().Warnf("failed to list cr %s in namespace %s "+
			"and will ignore this error: %s", crName, namespace, err.Error())
		return nil
	}
	eventType := SubEventTypeUpdate
	if err != nil && k8serrors.IsNotFound(err) {
		eventType = SubEventTypeDelete
	}

	obj, ok := objRaw.(*unstructured.Unstructured)
	if !ok {
		return nil
	}

	switch eventType {
	case SubEventTypeUpdate:
		logger.Infof("recv agent update event")
		defer logger.Infof("publish update event to ch over")
		agentRunInfo, err := ar.buildAgentCRInfo(obj, crKey)
		if err != nil {
			return err
		}
		ar.addFuncSpec(agentRunInfo, namespace)
		ar.publishEvent(SubEventTypeUpdate, agentRunInfo)
	case SubEventTypeDelete:
		logger.Infof("recv agent delete event")
		defer logger.Infof("publish delete event to ch over")
		agentRunInfo := &types.AgentEventInfo{
			CrKey: crKey,
		}
		funcSpec := ar.getFuncSpecFromCrName(crKey)
		if funcSpec == nil {
			logger.Infof("no funcSpec")
			return nil
		}
		funcSpec.CancelFunc()
		ar.deleteFuncSpec(crKey)
		ar.publishEvent(SubEventTypeDelete, agentRunInfo)
	default:
		logger.Warnf("invalid event type")
	}
	return nil
}

func (ar *AgentRegistry) buildAgentCRInfo(obj *unstructured.Unstructured, crKey string) (*types.AgentEventInfo, error) {
	logger := log.GetLogger().With(zap.Any("crKey", crKey))
	specBytes, err := functioncr.GetFunctionSpecData(obj)
	if err != nil {
		logger.Errorf("failed to get func spec data: %s", err.Error())
		return nil, err
	}

	var info = &commonType.FunctionMetaInfo{}
	if err := json.Unmarshal(specBytes, info); err != nil {
		logger.Errorf("failed to convert crd spec: %s", err.Error())
		return nil, err
	}
	logger = logger.With(zap.Any("urn", info.FuncMetaData.FunctionVersionURN))
	functionUrnInfo, err := urnutils.GetFunctionInfo(info.FuncMetaData.FunctionVersionURN)
	if err != nil {
		logger.Errorf("failed to convert urn, err: %s", err.Error())
		return nil, err
	}
	info.FuncMetaData.TenantID = functionUrnInfo.TenantID
	info.FuncMetaData.FuncName = functionUrnInfo.FuncName
	info.FuncMetaData.Version = functionUrnInfo.FuncVersion
	info.FuncMetaData.BusinessID = functionUrnInfo.BusinessID
	info.FuncMetaData.Service = urnutils.GetServiceNameFromFullName(functionUrnInfo.FuncName)

	setDefaultValue(info)

	funcKey := urnutils.CombineFunctionKey(info.FuncMetaData.TenantID, info.FuncMetaData.FuncName,
		info.FuncMetaData.Version)
	funcSpec := createOrUpdateFuncSpec(ar.getFuncSpec(funcKey), funcKey, info)
	funcSpec.MetaFromCR = true // 后续修正

	var status = &types.AgentCRStatus{}
	statusRaww, ok := obj.UnstructuredContent()["status"]
	if ok {
		statusRaw, ok := statusRaww.(map[string]interface{})
		if !ok {
			status = nil
		} else if err := runtime.DefaultUnstructuredConverter.FromUnstructured(statusRaw, status); err != nil {
			logger.Errorf("failed to convert crd status: %s", err.Error())
			return nil, err
		}
	} else {
		status = nil
	}

	return &types.AgentEventInfo{
		CrKey:    crKey,
		FuncSpec: funcSpec,
		Status:   status,
	}, nil
}

func setDefaultValue(info *commonType.FunctionMetaInfo) {
	if info == nil {
		return
	}
	if info.ExtendedMetaData.Initializer.Timeout == 0 {
		info.ExtendedMetaData.Initializer.Timeout = 300 // default initializer timeout
	}
	if info.InstanceMetaData.ConcurrentNum == 0 {
		info.InstanceMetaData.ConcurrentNum = 100 // default concurrent num
	}
}
