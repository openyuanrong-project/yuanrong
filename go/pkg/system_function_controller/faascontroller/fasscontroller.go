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

// Package faascontroller -
package faascontroller

import (
	"encoding/base64"
	"encoding/json"
	"sync"

	"go.uber.org/zap"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/uuid"
	"yuanrong.org/kernel/pkg/system_function_controller/config"
	"yuanrong.org/kernel/pkg/system_function_controller/instancemanager"
	"yuanrong.org/kernel/pkg/system_function_controller/instancemanager/faasfrontendmanager"
	"yuanrong.org/kernel/pkg/system_function_controller/instancemanager/faasfunctionmanager"
	"yuanrong.org/kernel/pkg/system_function_controller/instancemanager/faasschedulermanager"
	"yuanrong.org/kernel/pkg/system_function_controller/registry"
	"yuanrong.org/kernel/pkg/system_function_controller/service"
	"yuanrong.org/kernel/pkg/system_function_controller/types"
)

// FaaSController define the controller that manages the faas scheduler instances
type FaaSController struct {
	instanceManager *instancemanager.InstanceManager
	sdkClient       api.LibruntimeAPI
	frontendOnce    sync.Once
	schedulerOnce   sync.Once
	managerOnce     sync.Once
	funcCh          chan types.SubEvent
	stopCh          chan struct{}
	allocRecord     sync.Map
	sync.RWMutex
}

const (
	defaultChanSize = 100
)

// NewFaaSControllerLibruntime will create a new scheduler instance manager by new sdk of multi libruntime
func NewFaaSControllerLibruntime(libruntimeAPI api.LibruntimeAPI, stopCh chan struct{}) (*FaaSController, error) {
	faaSController := &FaaSController{
		instanceManager: &instancemanager.InstanceManager{},
		sdkClient:       libruntimeAPI,
		funcCh:          make(chan types.SubEvent, defaultChanSize),
		allocRecord:     sync.Map{},
		RWMutex:         sync.RWMutex{},
		stopCh:          stopCh,
	}
	go faaSController.processFunctionSubscription()
	return faaSController, nil
}

// NewFaaSController will create a new scheduler instance manager
func NewFaaSController(sdkClient api.LibruntimeAPI, stopCh chan struct{}) (*FaaSController, error) {
	faaSController := &FaaSController{
		instanceManager: &instancemanager.InstanceManager{},
		sdkClient:       sdkClient,
		funcCh:          make(chan types.SubEvent, defaultChanSize),
		allocRecord:     sync.Map{},
		RWMutex:         sync.RWMutex{},
		stopCh:          stopCh,
	}
	if err := service.CreateFrontendService(); err != nil {
		log.GetLogger().Errorf("failed to create service, reason: %s", err.Error())
	}
	go faaSController.processFunctionSubscription()
	return faaSController, nil
}

func (fc *FaaSController) processFunctionSubscription() {
	for {
		select {
		case event, ok := <-fc.funcCh:
			if !ok {
				log.GetLogger().Warnf("function channel is closed")
				return
			}
			functionSpec, ok := event.EventMsg.(*types.InstanceSpecification)
			if !ok {
				log.GetLogger().Warnf("event message doesn't contain instance specification")
				continue
			}
			if event.EventType == types.SubEventTypeUpdate {
				log.GetLogger().Debugf("receive update event. instanceID=%s", functionSpec.InstanceID)
				go fc.instanceManager.HandleEventUpdate(functionSpec, event.EventKind)
			}
			if event.EventType == types.SubEventTypeDelete {
				log.GetLogger().Debugf("receive delete event. instanceID=%s", functionSpec.InstanceID)
				go fc.instanceManager.HandleEventDelete(functionSpec, event.EventKind)
			}
			if event.EventType == types.SubEventTypeRecover {
				log.GetLogger().Debugf("receive recover event. instanceID=%s", functionSpec.InstanceID)
				go fc.instanceManager.HandleEventRecover(functionSpec, event.EventKind)
			}
		}
	}
}

// FrontendSignalHandler frontend signal handler
func (fc *FaaSController) FrontendSignalHandler(data []byte) error {
	traceID := uuid.New().String()
	logger := log.GetLogger().With(zap.Any("traceID", traceID))
	if len(data) == 0 {
		logger.Warnf("config is empty, exit frontend hot update")
		return nil
	}
	cfgStr, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		logger.Errorf("failed to decode frontend config bytes: %v, config: %s", err, data)
		return err
	}
	log.GetLogger().Infof("frontend config string is: %s", cfgStr)
	frontendConfig := &types.FrontendConfig{}
	err = json.Unmarshal(cfgStr, frontendConfig)
	if err != nil {
		logger.Errorf("failed to parse frontend config: %v, config: %s", err, cfgStr)
		return err
	}
	etcdConfig, err := config.DecryptEtcdConfig(frontendConfig.MetaEtcd)
	if err != nil {
		return err
	}
	frontendConfig.MetaEtcd = *etcdConfig
	fc.frontendOnce.Do(func() {
		config.UpdateFrontendConfig(frontendConfig)
		frontendManager := faasfrontendmanager.NewFaaSFrontendManager(fc.sdkClient,
			etcd3.GetRouterEtcdClient(), fc.stopCh, frontendConfig.InstanceNum, frontendConfig.DynamicPoolEnable)
		fc.instanceManager.FrontendManager = frontendManager
		frontendRegistry := registry.NewFrontendRegistry(fc.stopCh)
		registry.GlobalRegistry.AddFunctionRegistry(frontendRegistry, types.EventKindFrontend)
		frontendRegistry.AddSubscriberChan(fc.funcCh)
		frontendRegistry.InitWatcher()
		frontendRegistry.RunWatcher()
	})
	cfgEvent := &types.ConfigChangeEvent{
		FrontendCfg: frontendConfig,
		TraceID:     traceID,
	}
	cfgEvent.Add(1)
	fc.instanceManager.FrontendManager.ConfigChangeCh <- cfgEvent
	cfgEvent.Wait()
	log.GetLogger().Infof("frontend signal handler deal complete, err: %v", cfgEvent.Error)
	return cfgEvent.Error
}

// SchedulerSignalHandler scheduler signal handler
func (fc *FaaSController) SchedulerSignalHandler(data []byte) error {
	traceID := uuid.New().String()
	logger := log.GetLogger().With(zap.Any("traceID", traceID))
	if len(data) == 0 {
		logger.Warnf("config is empty, exit scheduler hot update")
		return nil
	}
	cfgStr, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		logger.Errorf("failed to decode scheduler config bytes: %v, config: %s", err, data)
		return err
	}
	log.GetLogger().Infof("scheduler config string is: %s", cfgStr)
	schedulerConfig := &types.SchedulerConfig{}
	err = json.Unmarshal(cfgStr, schedulerConfig)
	if err != nil {
		logger.Errorf("failed to parse scheduler config: %v, config: %s", err, cfgStr)
		return err
	}
	etcdConfig, err := config.DecryptEtcdConfig(schedulerConfig.MetaETCDConfig)
	if err != nil {
		return err
	}
	schedulerConfig.MetaETCDConfig = *etcdConfig
	fc.schedulerOnce.Do(func() {
		config.UpdateSchedulerConfig(schedulerConfig)
		// init default scheduler manager and other scheduler managers for exclusivity tenant
		fc.initSchedulerManagers(schedulerConfig)
		schedulerRegistry := registry.NewSchedulerRegistry(fc.stopCh)
		registry.GlobalRegistry.AddFunctionRegistry(schedulerRegistry, types.EventKindScheduler)
		schedulerRegistry.AddSubscriberChan(fc.funcCh)
		schedulerRegistry.InitWatcher()
		schedulerRegistry.RunWatcher()
	})
	cfgEvent := &types.ConfigChangeEvent{
		SchedulerCfg: schedulerConfig,
		TraceID:      traceID,
	}
	cfgEvent.Add(1)
	fc.instanceManager.CommonSchedulerManager.ConfigChangeCh <- cfgEvent
	for tenantID := range fc.instanceManager.ExclusivitySchedulerManagers {
		cfgEvent.Add(1)
		if fc.instanceManager.ExclusivitySchedulerManagers[tenantID] != nil {
			fc.instanceManager.ExclusivitySchedulerManagers[tenantID].ConfigChangeCh <- cfgEvent
		}
	}
	cfgEvent.Wait()
	log.GetLogger().Infof("scheduler signal handler deal complete, err: %v", cfgEvent.Error)
	return cfgEvent.Error
}

// ManagerSignalHandler manager signal handler
func (fc *FaaSController) ManagerSignalHandler(data []byte) error {
	traceID := uuid.New().String()
	logger := log.GetLogger().With(zap.Any("traceID", traceID))
	if len(data) == 0 {
		logger.Warnf("config is empty, exit funcManager hot update")
		return nil
	}
	cfgStr, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		logger.Errorf("failed to decode funcManager config bytes: %v, config: %s", err, data)
		return err
	}
	log.GetLogger().Infof("funcManager config string is: %s", cfgStr)
	managerConfig := &types.ManagerConfig{}
	err = json.Unmarshal(cfgStr, managerConfig)
	if err != nil {
		logger.Errorf("failed to parse funcManager config: %v, config: %s", err, cfgStr)
		return err
	}
	etcdConfig, err := config.DecryptEtcdConfig(managerConfig.MetaEtcd)
	if err != nil {
		return err
	}
	managerConfig.MetaEtcd = *etcdConfig
	fc.managerOnce.Do(func() {
		config.UpdateManagerConfig(managerConfig)
		funcManager := faasfunctionmanager.NewFaaSFunctionManager(fc.sdkClient,
			etcd3.GetRouterEtcdClient(),
			fc.stopCh, managerConfig.ManagerInstanceNum)
		fc.instanceManager.FunctionManager = funcManager
		managerRegistry := registry.NewManagerRegistry(fc.stopCh)
		registry.GlobalRegistry.AddFunctionRegistry(managerRegistry, types.EventKindManager)
		managerRegistry.AddSubscriberChan(fc.funcCh)
		managerRegistry.InitWatcher()
		managerRegistry.RunWatcher()
	})
	cfgEvent := &types.ConfigChangeEvent{
		ManagerCfg: managerConfig,
		TraceID:    traceID,
	}
	cfgEvent.Add(1)
	fc.instanceManager.FunctionManager.ConfigChangeCh <- cfgEvent
	cfgEvent.Wait()
	log.GetLogger().Infof("funcManager signal handler deal complete, err: %v", cfgEvent.Error)
	return cfgEvent.Error
}

func (fc *FaaSController) initSchedulerManagers(schedulerConfig *types.SchedulerConfig) {
	schedulerManager := faasschedulermanager.NewFaaSSchedulerManager(fc.sdkClient, etcd3.GetRouterEtcdClient(),
		fc.stopCh, schedulerConfig.SchedulerNum, "")
	fc.instanceManager.CommonSchedulerManager = schedulerManager
	fc.instanceManager.ExclusivitySchedulerManagers = make(map[string]*faasschedulermanager.SchedulerManager,
		len(config.GetFaaSControllerConfig().SchedulerExclusivity))
	for _, tenantID := range config.GetFaaSControllerConfig().SchedulerExclusivity {
		log.GetLogger().Infof("begin to new scheduler manager for tenant %s", tenantID)
		fc.instanceManager.ExclusivitySchedulerManagers[tenantID] = faasschedulermanager.NewFaaSSchedulerManager(
			fc.sdkClient, etcd3.GetRouterEtcdClient(), fc.stopCh, 1, tenantID)
	}
}
