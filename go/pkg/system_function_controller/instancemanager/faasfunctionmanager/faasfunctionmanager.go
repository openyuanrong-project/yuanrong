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

// Package faasfunctionmanager manages faasfunction status and instance ID
package faasfunctionmanager

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.etcd.io/etcd/client/v3"
	"k8s.io/apimachinery/pkg/util/wait"

	"yuanrong.org/kernel/runtime/libruntime/api"

	commonconstant "yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/common/uuid"
	"yuanrong.org/kernel/pkg/system_function_controller/config"
	"yuanrong.org/kernel/pkg/system_function_controller/constant"
	"yuanrong.org/kernel/pkg/system_function_controller/types"
	controllerutils "yuanrong.org/kernel/pkg/system_function_controller/utils"
)

var (
	once sync.Once
	// faasFunctionManager is the singaleton of Manager
	functionManager *FunctionManager

	createInstanceBackoff = wait.Backoff{
		Steps:    constant.DefaultCreateRetryTime, // retry times (include first time)
		Duration: constant.DefaultCreateRetryDuration,
		Factor:   constant.DefaultCreateRetryFactor,
		Jitter:   constant.DefaultCreateRetryJitter,
	}
)

// FunctionManager manages faasfunction status and instance ID
type FunctionManager struct {
	ConfigChangeCh        chan *types.ConfigChangeEvent
	instanceCache         map[string]*types.InstanceSpecification
	terminalCache         map[string]*types.InstanceSpecification
	etcdClient            *etcd3.EtcdClient
	sdkClient             api.LibruntimeAPI // add sdkClientLibruntime to adaptor multi runtime
	stopCh                chan struct{}
	recreateInstanceIDCh  chan string
	recreateInstanceIDMap sync.Map
	sync.RWMutex
	count int
}

// NewFaaSFunctionManager supply a singleton function manager
func NewFaaSFunctionManager(libruntimeAPI api.LibruntimeAPI,
	etcdClient *etcd3.EtcdClient, stopCh chan struct{}, size int) *FunctionManager {
	once.Do(func() {
		functionManager = &FunctionManager{
			instanceCache:        make(map[string]*types.InstanceSpecification),
			terminalCache:        map[string]*types.InstanceSpecification{},
			etcdClient:           etcdClient,
			sdkClient:            libruntimeAPI,
			stopCh:               stopCh,
			count:                size,
			ConfigChangeCh:       make(chan *types.ConfigChangeEvent, constant.DefaultChannelSize),
			recreateInstanceIDCh: make(chan string, constant.DefaultChannelSize),
		}
		go functionManager.recreateInstance()
		functionManager.initInstanceCache(etcdClient)
		ctx, cancelFunc := context.WithCancel(context.Background())
		go func() {
			err := functionManager.CreateExpectedInstanceCount(ctx)
			if err != nil {
				log.GetLogger().Errorf("Failed to create expected faasManager instance count, error: %v", err)
			}
		}()
		go functionManager.configChangeProcessor(ctx, cancelFunc)
	})
	return functionManager
}

func (fm *FunctionManager) initInstanceCache(etcdClient *etcd3.EtcdClient) {
	response, err := etcdClient.Client.Get(context.Background(), types.FasSManagerPrefixKey, clientv3.WithPrefix())
	if err != nil {
		log.GetLogger().Errorf("Failed to get faasManager instance: %v", err)
		return
	}
	config.ManagerConfigLock.RLock()
	targetSign := controllerutils.GetManagerConfigSignature(config.GetFaaSManagerConfig())
	config.ManagerConfigLock.RUnlock()
	for _, kv := range response.Kvs {
		meta := &types.InstanceSpecificationMeta{}
		err := json.Unmarshal(kv.Value, meta)
		if err != nil {
			log.GetLogger().Errorf("Failed to unmarshal instance specification: %v", err)
			continue
		}
		if isExceptInstance(meta, targetSign) {
			funcCtx, cancelFunc := context.WithCancel(context.TODO())
			fm.instanceCache[meta.InstanceID] = &types.InstanceSpecification{
				FuncCtx:                   funcCtx,
				CancelFunc:                cancelFunc,
				InstanceID:                meta.InstanceID,
				InstanceSpecificationMeta: *meta,
			}
			log.GetLogger().Infof("find expected manager instance %s add to cache", meta.InstanceID)
		}
	}
}

func isExceptInstance(meta *types.InstanceSpecificationMeta, targetSign string) bool {
	if len(meta.Args) == 0 {
		log.GetLogger().Errorf("manager ins args is empty, %v", meta)
		return false
	}
	v := meta.Args[0]["value"]
	s, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		log.GetLogger().Errorf("manager ins failed to decode args: %v", err)
		return false
	}
	cfg := &types.ManagerConfig{}
	err = json.Unmarshal(s, cfg)
	if err != nil && len(s) > commonconstant.LibruntimeHeaderSize {
		// args in libruntime create request with 16 bytes header.
		// except libruntime, if other modules use this field, should try yo delete the header
		err = json.Unmarshal(s[commonconstant.LibruntimeHeaderSize:], cfg)
		if err != nil {
			log.GetLogger().Errorf("Failed to unmarshal faasManager config: %v, value: %s", err, s)
			return false
		}
	}
	oldSign := controllerutils.GetManagerConfigSignature(cfg)
	if oldSign == "" {
		log.GetLogger().Errorf("old sign is empty, insID:%s", meta.InstanceID)
		return false
	}
	log.GetLogger().Infof("manager(%s) sign: %s, expect sign: %s", meta.InstanceID, oldSign, targetSign)
	return strings.Compare(oldSign, targetSign) == 0
}

// GetInstanceCountFromEtcd get current instance count from etcd
func (fm *FunctionManager) GetInstanceCountFromEtcd() map[string]struct{} {
	resp, err := fm.etcdClient.Client.Get(context.TODO(), types.FasSManagerPrefixKey, clientv3.WithPrefix())
	if err != nil {
		log.GetLogger().Errorf("failed to search etcd key, prefixKey=%s, err=%s", types.FasSManagerPrefixKey,
			err.Error())
		return nil
	}
	instanceIDs := make(map[string]struct{}, resp.Count)
	for _, kv := range resp.Kvs {
		instanceID := controllerutils.ExtractInfoFromEtcdKey(string(kv.Key), commonconstant.InstanceIDIndexForInstance)
		if instanceID != "" {
			instanceIDs[instanceID] = struct{}{}
		}
	}
	log.GetLogger().Infof("get etcd function instance count=%d, %+v", resp.Count, instanceIDs)
	return instanceIDs
}

// CreateExpectedInstanceCount create expected faasManager instance count
func (fm *FunctionManager) CreateExpectedInstanceCount(ctx context.Context) error {
	// 此方法在启动的时候调用一次或者在配置更新的时候调用, 目的是将manager实例数量补充至设置的实例数
	// 不需要删除多余实例, 多余实例会在 HandleInstanceUpdate 中删除
	// manager的 instanceID不需要保持一致
	fm.RLock()
	currentCount := len(fm.instanceCache)
	expectedCount := fm.count - currentCount
	fm.RUnlock()
	return fm.CreateMultiInstances(ctx, expectedCount)
}

// CreateMultiInstances create multi instances
func (fm *FunctionManager) CreateMultiInstances(ctx context.Context, count int) error {
	if count <= 0 {
		log.GetLogger().Infof("no need to create manager instance, kill %d instances instead.", -count)
		return fm.KillExceptInstance(-count)
	}
	log.GetLogger().Infof("need to create %d faas manager instances", count)

	functionArgs, params, err := genFunctionConfig()
	if err != nil {
		return err
	}
	group := &sync.WaitGroup{}
	var createErr error
	for i := 0; i < count; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if err = fm.createOrRetry(ctx, functionArgs, *params,
				config.GetFaaSControllerConfig().EnableRetry); err != nil {
				createErr = err
			}
		}()
	}
	group.Wait()
	if createErr != nil {
		return createErr
	}
	log.GetLogger().Infof("succeed to create %d faas manager instances", count)
	return nil
}

func genFunctionConfig() ([]api.Arg, *commonTypes.ExtraParams, error) {
	config.ManagerConfigLock.RLock()
	managerConfig := config.GetFaaSManagerConfig()
	extraParams, err := createExtraParams(managerConfig)
	if err != nil {
		config.ManagerConfigLock.RUnlock()
		log.GetLogger().Errorf("failed to prepare faasManager createExtraParams, err:%s", err.Error())
		return nil, nil, err
	}
	managerConf, err := json.Marshal(managerConfig)
	config.ManagerConfigLock.RUnlock()
	if err != nil {
		log.GetLogger().Errorf("faasManager config json marshal failed, err:%s", err.Error())
		return nil, nil, err
	}
	args := []api.Arg{
		{
			Type: api.Value,
			Data: managerConf,
		},
	}
	return args, extraParams, nil
}

func createExtraParams(conf *types.ManagerConfig) (*commonTypes.ExtraParams, error) {
	extraParams := &commonTypes.ExtraParams{}
	extraParams.Resources = utils.GenerateResourcesMap(conf.CPU, conf.Memory)
	extraParams.CustomExtensions = utils.CreateCustomExtensions(extraParams.CustomExtensions,
		utils.MonopolyPolicyValue)
	extraParams.ScheduleAffinities = utils.CreatePodAffinity(constant.SystemFuncName, constant.FuncNameFaasmanager,
		api.PreferredAntiAffinity)
	utils.AddNodeSelector(conf.NodeSelector, extraParams)
	createOpts, err := prepareCreateOptions(conf)
	extraParams.CreateOpt = createOpts
	extraParams.Label = []string{constant.FuncNameFaasmanager}
	return extraParams, err
}

func (fm *FunctionManager) createOrRetry(ctx context.Context, args []api.Arg,
	extraParams commonTypes.ExtraParams, enableRetry bool) error {
	// 只有首次拉起/扩容时insID为空,需要在拉起失败时防止进入失败回调中的重试逻辑
	if extraParams.DesignatedInstanceID == "" {
		instanceID := uuid.New().String()
		fm.recreateInstanceIDMap.Store(instanceID, nil)
		extraParams.DesignatedInstanceID = instanceID
	}
	defer fm.recreateInstanceIDMap.Delete(extraParams.DesignatedInstanceID)
	if enableRetry {
		err := fm.CreateWithRetry(ctx, args, extraParams)
		if err != nil {
			return err
		}
	} else {
		instanceID := fm.CreateInstance(ctx, types.FasSManagerFunctionKey, args, &extraParams)
		if instanceID == "" {
			log.GetLogger().Errorf("failed to create function manager instance")
			return errors.New("failed to create function manager instance")
		}
	}
	return nil
}

// CreateWithRetry -
func (fm *FunctionManager) CreateWithRetry(ctx context.Context, args []api.Arg,
	extraParams commonTypes.ExtraParams) error {
	err := wait.ExponentialBackoffWithContext(ctx, createInstanceBackoff, func(context.Context) (bool, error) {
		instanceID := fm.CreateInstance(ctx, types.FasSManagerFunctionKey, args, &extraParams)
		if instanceID == "" {
			log.GetLogger().Errorf("failed to create funcManager instance")
			return false, nil
		}
		if instanceID == "cancelled" {
			return true, fmt.Errorf("create has been cancelled")
		}
		return true, nil
	})
	return err
}

// CreateInstance create an instance of system function, faaS function
func (fm *FunctionManager) CreateInstance(ctx context.Context, function string, args []api.Arg,
	extraParams *commonTypes.ExtraParams) string {
	instanceID := extraParams.DesignatedInstanceID
	log.GetLogger().Infof("start to create funcManager instance(id=%s)",
		extraParams.DesignatedInstanceID)
	funcMeta := api.FunctionMeta{Name: &instanceID, FuncID: function, Api: api.PosixApi}
	invokeOpts := api.InvokeOptions{
		Cpu:                int(extraParams.Resources[controllerutils.ResourcesCPU]),
		Memory:             int(extraParams.Resources[controllerutils.ResourcesMemory]),
		ScheduleAffinities: extraParams.ScheduleAffinities,
		CustomExtensions:   extraParams.CustomExtensions,
		CreateOpt:          extraParams.CreateOpt,
		Labels:             extraParams.Label,
		Timeout:            150,
	}
	createCh := make(chan api.ErrorInfo, 1)
	go func() {
		_, createErr := fm.sdkClient.CreateInstance(funcMeta, args, invokeOpts)
		if createErr != nil {
			if errorInfo, ok := createErr.(api.ErrorInfo); ok {
				createCh <- errorInfo
			} else {
				createCh <- api.ErrorInfo{Code: commonconstant.KernelInnerSystemErrCode, Err: createErr}
			}
			return
		}
		createCh <- api.ErrorInfo{Code: api.Ok}
	}()
	if result := fm.waitForInstanceCreation(ctx, createCh, instanceID); result != instanceID {
		return result
	}
	log.GetLogger().Infof("succeed to create manager instance(id=%s)", instanceID)
	fm.addInstance(instanceID)
	return instanceID
}
func (fm *FunctionManager) waitForInstanceCreation(ctx context.Context, createCh <-chan api.ErrorInfo,
	instanceID string) string {
	timer := time.NewTimer(types.CreatedTimeout)
	select {
	case err, ok := <-createCh:
		defer timer.Stop()
		if !ok {
			log.GetLogger().Errorf("result channel of manager instance request is closed")
			return ""
		}
		if !err.IsOk() {
			log.GetLogger().Errorf("failed to bring up manager instance(id=%s), err: %v",
				instanceID, err.Error())
			fm.clearInstanceAfterError(instanceID)
			return ""
		}
		return instanceID
	case <-timer.C:
		log.GetLogger().Errorf("time out waiting for instance creation")
		fm.clearInstanceAfterError(instanceID)
		return ""
	case <-ctx.Done():
		log.GetLogger().Errorf("create instance has been cancelled")
		fm.clearInstanceAfterError(instanceID)
		return "cancelled"
	}
}
func (fm *FunctionManager) recreateInstance() {
	for {
		select {
		case <-fm.stopCh:
			return
		case instanceID, ok := <-fm.recreateInstanceIDCh:
			if !ok {
				log.GetLogger().Warnf("recreateInstanceIDCh is closed")
				return
			}
			fm.RLock()
			if _, exist := fm.instanceCache[instanceID]; exist || len(fm.instanceCache) >= fm.count {
				log.GetLogger().Infof("current manager num is %s, no need to recreate instance:%s",
					len(fm.instanceCache), instanceID)
				fm.RUnlock()
				break
			}
			fm.RUnlock()
			ctx, cancel := context.WithCancel(context.Background())
			_, loaded := fm.recreateInstanceIDMap.LoadOrStore(instanceID, cancel)
			if loaded {
				log.GetLogger().Warnf("instance[%s] is recreating", instanceID)
				break
			}
			args, extraParams, err := genFunctionConfig()
			if err != nil {
				break
			}
			extraParams.DesignatedInstanceID = instanceID
			if err != nil {
				log.GetLogger().Errorf("failed to prepare createExtraParams, err:%s", err.Error())
				break
			}
			go func() {
				time.Sleep(constant.RecreateSleepTime)
				log.GetLogger().Infof("start to recover faaSManager instance: %s", instanceID)
				if err = fm.createOrRetry(ctx, args, *extraParams,
					config.GetFaaSControllerConfig().EnableRetry); err != nil {
					log.GetLogger().Errorf("failed to recreate manager instance: %s", instanceID)
				}
			}()
		}
	}
}

func (fm *FunctionManager) clearInstanceAfterError(instanceID string) {
	if err := fm.sdkClient.Kill(instanceID, types.KillSignalVal, []byte{}); err != nil {
		log.GetLogger().Errorf("failed to kill manager instance: %s", instanceID)
	}
}

func (fm *FunctionManager) addInstance(instanceID string) {
	fm.Lock()
	defer fm.Unlock()
	_, exist := fm.instanceCache[instanceID]
	if exist {
		log.GetLogger().Warnf("the manager instance(id=%s) already exist", instanceID)
		return
	}
	log.GetLogger().Infof("add instance(id=%s) to local cache", instanceID)
	fm.instanceCache[instanceID] = &types.InstanceSpecification{InstanceID: instanceID}
}

// SyncCreateInstance -
func (fm *FunctionManager) SyncCreateInstance(ctx context.Context) error {
	log.GetLogger().Infof("start to sync create funcManager instance")
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	args, extraParams, err := genFunctionConfig()
	if err != nil {
		return err
	}
	err = fm.createOrRetry(ctx, args, *extraParams, config.GetFaaSControllerConfig().EnableRetry)
	if err != nil {
		return err
	}
	return nil
}

// GetInstanceCache supply local instance cache
func (fm *FunctionManager) GetInstanceCache() map[string]*types.InstanceSpecification {
	return fm.instanceCache
}

// KillInstance kill an instance of system function, faaS function
func (fm *FunctionManager) KillInstance(instanceID string) error {
	log.GetLogger().Infof("start to kill funcManager instance %s", instanceID)
	return wait.ExponentialBackoffWithContext(
		context.Background(), createInstanceBackoff, func(context.Context) (bool, error) {
			var err error
			err = fm.sdkClient.Kill(instanceID, types.KillSignalVal, []byte{})
			if err != nil && !strings.Contains(err.Error(), "instance not found") {
				log.GetLogger().Warnf("failed to kill funcManager instanceID: %s, err: %s",
					instanceID, err.Error())
				return false, nil
			}
			return true, nil
		})
}

// SyncKillAllInstance kill all instances of system function, faaS manager
func (fm *FunctionManager) SyncKillAllInstance() {
	var wg sync.WaitGroup
	var deletedInstance []string
	fm.Lock()
	defer fm.Unlock()
	for instanceID := range fm.instanceCache {
		wg.Add(1)
		go func(instanceID string) {
			defer wg.Done()
			if err := fm.sdkClient.Kill(instanceID, types.SyncKillSignalVal, []byte{}); err != nil {
				log.GetLogger().Errorf("failed to kill manager instance(id=%s), err:%s", instanceID, err.Error())
				return
			}
			deletedInstance = append(deletedInstance, instanceID)
			log.GetLogger().Infof("success to kill manager instance(id=%s)", instanceID)
		}(instanceID)
	}
	wg.Wait()
	for _, instanceID := range deletedInstance {
		log.GetLogger().Infof("delete manager instance(id=%s) from local cache", instanceID)
		delete(fm.instanceCache, instanceID)
		fm.count--
	}
}

// KillExceptInstance -
func (fm *FunctionManager) KillExceptInstance(count int) error {
	if len(fm.instanceCache) < count {
		return nil
	}
	for instanceID := range fm.instanceCache {
		if count <= 0 {
			return nil
		}
		if err := fm.KillInstance(instanceID); err != nil {
			log.GetLogger().Errorf("kill manager instance:%s, error:%s", instanceID, err.Error())
			return err
		}
		count--
	}
	return nil
}

// RecoverInstance recover a faaS manager instance when faults occur
func (fm *FunctionManager) RecoverInstance(info *types.InstanceSpecification) {
	err := fm.KillInstance(info.InstanceID)
	if err != nil {
		log.GetLogger().Warnf("failed to kill instanceID: %s, err: %s", info.InstanceID, err.Error())
	}
}

// HandleInstanceUpdate handle the etcd PUT event
func (fm *FunctionManager) HandleInstanceUpdate(instanceSpec *types.InstanceSpecification) {
	log.GetLogger().Infof("handling funcManager instance %s update", instanceSpec.InstanceID)
	if instanceSpec.InstanceSpecificationMeta.InstanceStatus.Code == int(commonconstant.KernelInstanceStatusExiting) {
		log.GetLogger().Infof("funcManager instance %s is exiting,no need to update", instanceSpec.InstanceID)
		return
	}
	config.ManagerConfigLock.RLock()
	signature := controllerutils.GetManagerConfigSignature(config.GetFaaSManagerConfig())
	config.ManagerConfigLock.RUnlock()

	if isExceptInstance(&instanceSpec.InstanceSpecificationMeta, signature) {
		fm.Lock()
		currentNum := len(fm.instanceCache)
		_, exist := fm.instanceCache[instanceSpec.InstanceID]
		if currentNum > fm.count || (currentNum == fm.count && !exist) {
			log.GetLogger().Infof("current funcManager num is %s, kill the new instance %s",
				currentNum, instanceSpec.InstanceID)
			delete(fm.instanceCache, instanceSpec.InstanceID)
			fm.Unlock()
			if err := fm.sdkClient.Kill(instanceSpec.InstanceID, types.KillSignalVal, []byte{}); err != nil {
				log.GetLogger().Errorf("failed to kill instance %s error:%s", instanceSpec.InstanceID,
					err.Error())
			}
			return
		}
		// add instance to cache if not exist, otherwise update the instance
		if !exist {
			fm.instanceCache[instanceSpec.InstanceID] = instanceSpec
			log.GetLogger().Infof("add funcManager instance %s to cache", instanceSpec.InstanceID)
			fm.Unlock()
			return
		}
		fm.instanceCache[instanceSpec.InstanceID].InstanceSpecificationMeta = instanceSpec.InstanceSpecificationMeta
		log.GetLogger().Infof("funcManager instance %s is updated, refresh instance cache",
			instanceSpec.InstanceID)
		fm.Unlock()
		return
	}
	fm.RLock()
	_, exist := fm.terminalCache[instanceSpec.InstanceID]
	fm.RUnlock()
	if !exist {
		log.GetLogger().Infof("funcManager instance %s is not expected, start to delete",
			instanceSpec.InstanceID)
		if err := fm.KillInstance(instanceSpec.InstanceID); err != nil {
			log.GetLogger().Errorf("failed to kill funcManager instance %s error:%s",
				instanceSpec.InstanceID, err.Error())
		}
	}
}

// HandleInstanceDelete handle the etcd DELETE event
func (fm *FunctionManager) HandleInstanceDelete(instanceSpec *types.InstanceSpecification) {
	log.GetLogger().Infof("handling funcManager instance %s delete", instanceSpec.InstanceID)
	config.ManagerConfigLock.RLock()
	signature := controllerutils.GetManagerConfigSignature(config.GetFaaSManagerConfig())
	config.ManagerConfigLock.RUnlock()
	fm.Lock()
	delete(fm.instanceCache, instanceSpec.InstanceID)
	fm.Unlock()
	if isExceptInstance(&instanceSpec.InstanceSpecificationMeta, signature) {
		fm.RLock()
		if len(fm.instanceCache) < fm.count {
			log.GetLogger().Infof("current funcManager instance num is %d, need to recreate instance: %s",
				len(fm.instanceCache), instanceSpec.InstanceID)
			fm.RUnlock()
			fm.recreateInstanceIDCh <- instanceSpec.InstanceID
			return
		}
		fm.RUnlock()
	}
	cancel, exist := fm.recreateInstanceIDMap.Load(instanceSpec.InstanceID)
	if exist {
		if cancelFunc, ok := cancel.(context.CancelFunc); ok {
			cancelFunc()
			log.GetLogger().Infof("funcManager instance %s bring up has been canceled", instanceSpec.InstanceID)
			return
		}
		log.GetLogger().Errorf("get cancel func failed from instanceIDMap, instanceID:%s",
			instanceSpec.InstanceID)
	}
}

func (fm *FunctionManager) killAllTerminalInstance() {
	fm.Lock()
	insMap := fm.terminalCache
	fm.terminalCache = map[string]*types.InstanceSpecification{}
	fm.Unlock()
	for _, ins := range insMap {
		insID := ins.InstanceID
		go func() {
			err := fm.KillInstance(insID)
			if err != nil {
				log.GetLogger().Errorf("Failed to kill funcManager instance %v, err: %v", insID, err)
			}
		}()
	}
}

// RollingUpdate -
func (fm *FunctionManager) RollingUpdate(ctx context.Context, event *types.ConfigChangeEvent) {
	// 1. 更新 预期实例个数
	// 2. 把不符合预期的instanceCache -> terminalCache
	// 3. 从terminalCache随机删除一个实例 同步
	// 4. 创建新实例 同步
	// 5. terminalCache为空时将实例数调谐为预期实例数(同步), instanceCache到达预期数量时清空terminalCache(异步)
	newSign := controllerutils.GetManagerConfigSignature(event.ManagerCfg)
	fm.Lock()
	fm.count = event.ManagerCfg.ManagerInstanceNum
	for _, ins := range fm.instanceCache {
		if !isExceptInstance(&ins.InstanceSpecificationMeta, newSign) {
			fm.terminalCache[ins.InstanceID] = ins
			delete(fm.instanceCache, ins.InstanceID)
		}
	}
	fm.Unlock()
	for {
		select {
		case <-ctx.Done():
			event.Error = fmt.Errorf("rolling update has stopped")
			event.Done()
			return
		default:
		}
		fm.RLock()
		if len(fm.instanceCache) == fm.count {
			fm.RUnlock()
			log.GetLogger().Infof("faasManager instance count arrive at expectation:%d,"+
				" delete all terminating instance", fm.count)
			go fm.killAllTerminalInstance()
			event.Done()
			return
		}
		if len(fm.terminalCache) == 0 {
			fm.RUnlock()
			log.GetLogger().Infof("no faasManager instance need to terminate, pull up missing instance count:%d",
				fm.count-len(fm.instanceCache))
			err := fm.CreateExpectedInstanceCount(ctx)
			if err != nil {
				event.Error = err
			}
			event.Done()
			return
		}
		insID := ""
		for _, ins := range fm.terminalCache {
			insID = ins.InstanceID
			break
		}
		fm.RUnlock()
		log.GetLogger().Infof("start to terminate instance:%d", insID)
		var err error
		if err = fm.sdkClient.Kill(insID, types.SyncKillSignalVal, []byte{}); err != nil {
			log.GetLogger().Errorf("failed to kill faasManager instance(id=%s), err:%v", insID, err)
		}
		fm.Lock()
		delete(fm.terminalCache, insID)
		fm.Unlock()
		time.Sleep(constant.RecreateSleepTime)
		err = fm.SyncCreateInstance(ctx)
		if err != nil {
			event.Error = err
			event.Done()
			return
		}
	}
}

func prepareCreateOptions(conf *types.ManagerConfig) (map[string]string, error) {
	podLabels := map[string]string{
		constant.SystemFuncName: constant.FuncNameFaasmanager,
	}
	labels, err := json.Marshal(podLabels)
	if err != nil {
		return nil, fmt.Errorf("pod labels json marshal failed, err:%s", err.Error())
	}
	delegateRuntime, err := json.Marshal(map[string]interface{}{
		"image": conf.Image,
	})
	if err != nil {
		return nil, err
	}
	createOpts := map[string]string{
		commonconstant.DelegatePodLabels:          string(labels),
		commonconstant.DelegateRuntimeManagerTag:  string(delegateRuntime),
		commonconstant.InstanceLifeCycle:          commonconstant.InstanceLifeCycleDetached,
		commonconstant.DelegateNodeAffinity:       conf.NodeAffinity,
		commonconstant.DelegateNodeAffinityPolicy: conf.NodeAffinityPolicy,
	}
	if config.GetFaaSManagerConfig().Affinity != "" {
		createOpts[commonconstant.DelegateAffinity] = config.GetFaaSManagerConfig().Affinity
	}
	return createOpts, nil
}

func (fm *FunctionManager) configChangeProcessor(ctx context.Context, cancel context.CancelFunc) {
	if ctx == nil || cancel == nil || fm.ConfigChangeCh == nil {
		return
	}
	for {
		select {
		case cfgEvent, ok := <-fm.ConfigChangeCh:
			if !ok {
				cancel()
				return
			}
			managerConfig, insNum := fm.ConfigDiff(cfgEvent)
			if managerConfig != nil || insNum != -1 {
				log.GetLogger().Infof("manager config or instance num is changed," +
					" need to update manager instance")
				cancel()
				ctx, cancel = context.WithCancel(context.Background())
				config.UpdateManagerConfig(cfgEvent.ManagerCfg)
				cfgEvent.Add(1)
				go fm.RollingUpdate(ctx, cfgEvent)
			} else {
				log.GetLogger().Infof("manager config is same as current, no need to update")
			}
			cfgEvent.Done()
		}
	}
}

// ConfigDiff config diff
func (fm *FunctionManager) ConfigDiff(event *types.ConfigChangeEvent) (*types.ManagerConfig, int) {
	newSign := controllerutils.GetManagerConfigSignature(event.ManagerCfg)
	config.ManagerConfigLock.RLock()
	managerOldCfg := config.GetFaaSManagerConfig()
	config.ManagerConfigLock.RUnlock()
	if strings.Compare(newSign,
		controllerutils.GetManagerConfigSignature(managerOldCfg)) == 0 {
		if event.ManagerCfg.ManagerInstanceNum != managerOldCfg.ManagerInstanceNum {
			config.ManagerConfigLock.Lock()
			managerOldCfg.ManagerInstanceNum = event.ManagerCfg.ManagerInstanceNum
			config.ManagerConfigLock.Unlock()
			return nil, event.ManagerCfg.ManagerInstanceNum
		}
		return nil, -1
	}
	return event.ManagerCfg, event.ManagerCfg.ManagerInstanceNum
}
