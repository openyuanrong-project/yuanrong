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

// Package faasschedulermanager -
package faasschedulermanager

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.etcd.io/etcd/client/v3"
	"k8s.io/apimachinery/pkg/util/wait"

	"yuanrong.org/kernel/runtime/libruntime/api"

	commonconstant "yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/sts"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/common/uuid"
	"yuanrong.org/kernel/pkg/system_function_controller/config"
	"yuanrong.org/kernel/pkg/system_function_controller/constant"
	"yuanrong.org/kernel/pkg/system_function_controller/state"
	"yuanrong.org/kernel/pkg/system_function_controller/types"
	controllerutils "yuanrong.org/kernel/pkg/system_function_controller/utils"
)

var (
	createInstanceBackoff = wait.Backoff{
		Steps:    constant.DefaultCreateRetryTime, // retry times (include first time)
		Duration: constant.DefaultCreateRetryDuration,
		Factor:   constant.DefaultCreateRetryFactor,
		Jitter:   constant.DefaultCreateRetryJitter,
	}
)

// SchedulerManager manages faaS scheduler specification
type SchedulerManager struct {
	instanceCache         map[string]*types.InstanceSpecification
	terminalCache         map[string]*types.InstanceSpecification
	etcdClient            *etcd3.EtcdClient
	sdkClient             api.LibruntimeAPI
	ConfigChangeCh        chan *types.ConfigChangeEvent
	stopCh                chan struct{}
	recreateInstanceIDMap sync.Map
	sync.RWMutex
	count    int
	tenantID string
}

// NewFaaSSchedulerManager create a scheduler instance manager
func NewFaaSSchedulerManager(libruntimeAPI api.LibruntimeAPI,
	etcdClient *etcd3.EtcdClient, stopCh chan struct{}, size int, tenantID string) *SchedulerManager {
	schedulerManager := &SchedulerManager{
		instanceCache:  map[string]*types.InstanceSpecification{},
		terminalCache:  map[string]*types.InstanceSpecification{},
		etcdClient:     etcdClient,
		sdkClient:      libruntimeAPI,
		count:          size,
		tenantID:       tenantID,
		stopCh:         stopCh,
		ConfigChangeCh: make(chan *types.ConfigChangeEvent, constant.DefaultChannelSize),
	}
	schedulerManager.initInstanceCache(etcdClient)
	ctx, cancelFunc := context.WithCancel(context.Background())
	go schedulerManager.configChangeProcessor(ctx, cancelFunc)
	go func() {
		err := schedulerManager.CreateExpectedInstanceCount(ctx)
		if err != nil {
			log.GetLogger().Errorf("Failed to create expected scheduler instance count, error: %v", err)
			return
		}
		log.GetLogger().Infof("succeed to create expected scheduler instance count %d for tenantID %s",
			size, tenantID)
	}()
	return schedulerManager
}

func (s *SchedulerManager) initInstanceCache(etcdClient *etcd3.EtcdClient) {
	response, err := etcdClient.Client.Get(context.Background(), types.FaaSSchedulerPrefixKey, clientv3.WithPrefix())
	if err != nil {
		log.GetLogger().Errorf("Failed to get scheduler instance: %v", err)
		return
	}
	config.SchedulerConfigLock.RLock()
	targetSign := controllerutils.GetSchedulerConfigSignature(config.GetFaaSSchedulerConfig())
	config.SchedulerConfigLock.RUnlock()
	for _, kv := range response.Kvs {
		meta := &types.InstanceSpecificationMeta{}
		err := json.Unmarshal(kv.Value, meta)
		if err != nil {
			log.GetLogger().Errorf("Failed to unmarshal instance specification: %v", err)
			continue
		}
		if isExpectInstance(meta, targetSign) {
			tenantID := ""
			if meta.CreateOptions != nil {
				tenantID = meta.CreateOptions[constant.SchedulerExclusivity]
			}
			if s.tenantID == tenantID {
				funcCtx, cancelFunc := context.WithCancel(context.TODO())
				s.instanceCache[meta.InstanceID] = &types.InstanceSpecification{
					FuncCtx:                   funcCtx,
					CancelFunc:                cancelFunc,
					InstanceID:                meta.InstanceID,
					InstanceSpecificationMeta: *meta,
				}
				log.GetLogger().Infof("find expected scheduler instance %s add to cache for tenantID %s",
					meta.InstanceID, tenantID)
			}
		}
		state.Update(meta.InstanceID, types.StateUpdate, types.FaasSchedulerInstanceState+s.tenantID)
	}
}

func isExpectInstance(meta *types.InstanceSpecificationMeta, targetSign string) bool {
	if len(meta.Args) == 0 {
		log.GetLogger().Errorf("args is empty, %v", meta)
		return false
	}
	v := meta.Args[0]["value"]
	s, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		log.GetLogger().Errorf("Failed to decode args: %v", err)
		return false
	}
	cfg := &types.SchedulerConfig{}
	err = json.Unmarshal(s, cfg)
	if err != nil && len(s) > commonconstant.LibruntimeHeaderSize {
		// args in libruntime create request with 16 bytes header.
		// except libruntime, if other modules use this field, should try yo delete the header
		err = json.Unmarshal(s[commonconstant.LibruntimeHeaderSize:], cfg)
		if err != nil {
			log.GetLogger().Errorf("Failed to unmarshal scheduler config: %v, value: %s", err, s)
			return false
		}
	}
	oldSign := controllerutils.GetSchedulerConfigSignature(cfg)
	if oldSign == "" {
		log.GetLogger().Errorf("old sign is empty, insID:%s", meta.InstanceID)
		return false
	}
	log.GetLogger().Infof("scheduler(%s) sign: %s, expect sign: %s", meta.InstanceID, oldSign, targetSign)
	return strings.Compare(oldSign, targetSign) == 0
}

// CreateExpectedInstanceCount create expected scheduler instance count
func (s *SchedulerManager) CreateExpectedInstanceCount(ctx context.Context) error {
	// 此方法在启动的时候调用一次, 目的是将scheduler实例数量补充至设置的实例数
	// 不需要删除多余实例, 多余实例会在 HandleInstanceUpdate 中删除
	// 因为需要保证拉起的scheduler的instanceID不变, 该函数需要同步FaasInstance缓存
	s.Lock()
	currentCount := len(s.instanceCache)
	expectedCount := s.count - currentCount
	recoverIns := state.GetState().FaasInstance[types.FaasSchedulerInstanceState+s.tenantID]

	// 这里将已有实例全部加到FaasInstance中, 保证FaasInstance的缓存包含已有实例(防止有些极端情况下FaasInstance缓存丢失)
	for _, meta := range s.instanceCache {
		_, exist := recoverIns[meta.InstanceID]
		if !exist {
			state.Update(meta.InstanceID, types.StateUpdate, types.FaasSchedulerInstanceState+s.tenantID)
		}
	}

	var insList []string
	// 同步FaasInstance缓存
	// 如果已有实例数没有达到设置的值, 拉起新实例的instanceID需要从缓存中取出, 如果缓存中没有则新建
	// 如果已有实例数已经达到了设置的值, 将缓存中多余的instanceID删除
	for instanceID := range recoverIns {
		if _, exist := s.instanceCache[instanceID]; !exist {
			if expectedCount <= 0 {
				state.Update(instanceID, types.StateDelete, types.FaasSchedulerInstanceState+s.tenantID)
				continue
			}
			insList = append(insList, instanceID)
			expectedCount--
		}
	}
	s.Unlock()
	wg := sync.WaitGroup{}
	for _, insID := range insList {
		go func(instanceID string) {
			log.GetLogger().Infof("instance %s not exist, need recover for %s", instanceID, s.tenantID)
			wg.Add(1)
			time.Sleep(constant.RecreateSleepTime)
			err := s.SyncCreateInstanceByID(ctx, instanceID)
			wg.Done()
			if err != nil {
				return
			}
		}(insID)
	}
	err := s.CreateMultiInstances(ctx, expectedCount)
	wg.Wait()
	return err
}

// CreateMultiInstances create multi instances
func (s *SchedulerManager) CreateMultiInstances(ctx context.Context, count int) error {
	if count <= 0 {
		log.GetLogger().Infof("no need to create scheduler instance, kill %d instances instead.", -count)
		return s.KillExceptInstance(-count)
	}
	log.GetLogger().Infof("need to create %d faas scheduler instances", count)

	args, extraParams, err := genFunctionConfig(s.tenantID)
	if err != nil {
		return err
	}
	group := &sync.WaitGroup{}
	var createErr error
	for i := 0; i < count; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if err = s.createOrRetry(ctx, args, *extraParams,
				config.GetFaaSControllerConfig().EnableRetry); err != nil {
				createErr = err
			}
		}()
	}
	group.Wait()
	if createErr != nil {
		return createErr
	}
	log.GetLogger().Infof("succeed to create %d faaS scheduler instances", count)
	return nil
}

func createExtraParams(conf *types.SchedulerConfig, tenantID string) (*commonTypes.ExtraParams, error) {
	config.SchedulerConfigLock.RLock()
	defer config.SchedulerConfigLock.RUnlock()
	extraParams := &commonTypes.ExtraParams{}
	extraParams.Resources = utils.GenerateResourcesMap(conf.CPU, conf.Memory)
	extraParams.CustomExtensions = utils.CreateCustomExtensions(extraParams.CustomExtensions,
		utils.MonopolyPolicyValue)
	extraParams.ScheduleAffinities = utils.CreatePodAffinity(constant.SystemFuncName, constant.FuncNameFaasscheduler,
		api.PreferredAntiAffinity)
	utils.AddNodeSelector(conf.NodeSelector, extraParams)
	encryptMap := map[string]string{config.MetaEtcdPwdKey: conf.MetaETCDConfig.Password}
	encryptData, err := json.Marshal(encryptMap)
	if err != nil {
		log.GetLogger().Errorf("encryptData json marshal failed, err:%s", err.Error())
		return nil, err
	}
	createOptions := make(map[string]string)
	createOptions = utils.CreateCreateOptions(createOptions, commonconstant.EnvDelegateEncrypt, string(encryptData))

	delegateRuntime, err := json.Marshal(map[string]interface{}{
		"image": conf.Image,
	})
	if err != nil {
		return nil, err
	}
	createOptions[commonconstant.DelegateRuntimeManagerTag] = string(delegateRuntime)
	createOptions[constant.InitCallTimeoutKey] = strconv.Itoa(int(types.CreatedTimeout.Seconds()))
	createOptions[constant.ConcurrencyKey] = strconv.Itoa(constant.MaxConcurrency)
	createOptions[commonconstant.InstanceLifeCycle] = commonconstant.InstanceLifeCycleDetached
	createOptions[constant.SchedulerExclusivity] = tenantID
	createOptions[commonconstant.DelegateNodeAffinity] = conf.NodeAffinity
	createOptions[commonconstant.DelegateNodeAffinityPolicy] = conf.NodeAffinityPolicy
	if conf.Affinity != "" {
		createOptions[commonconstant.DelegateAffinity] = conf.Affinity
	}
	createOptions[commonconstant.DelegateEnvVar] = fmt.Sprintf(`{"%s":"%s"}`,
		commonconstant.FaaSSchedulerExclusivityEnvKey, tenantID)
	makePodLabel(createOptions, tenantID)
	if config.GetFaaSControllerConfig().RawStsConfig.StsEnable {
		secretVolumeMounts, err := sts.GenerateSecretVolumeMounts(sts.FaaSSchedulerName, utils.NewVolumeBuilder())
		if err != nil {
			return nil, err
		}
		createOptions = utils.CreateCreateOptions(createOptions, commonconstant.DelegateVolumeMountKey,
			string(secretVolumeMounts))
	}
	if config.GetFaaSSchedulerConfig().ConcurrentNum == 0 {
		createOptions[constant.ConcurrentNumKey] = strconv.Itoa(constant.DefaultConcurrentNum)
	} else {
		createOptions[constant.ConcurrentNumKey] = strconv.Itoa(config.GetFaaSSchedulerConfig().ConcurrentNum)
	}
	err = prepareVolumesAndMounts(conf, createOptions)
	if err != nil {
		return nil, err
	}

	extraParams.CreateOpt = createOptions
	extraParams.Label = []string{constant.FuncNameFaasscheduler}
	return extraParams, nil
}

func prepareVolumesAndMounts(conf *types.SchedulerConfig, createOptions map[string]string) error {
	if conf == nil || createOptions == nil {
		return fmt.Errorf("parameter of prepareVolumesAndMounts is nil")
	}
	var delegateVolumes string
	var delegateVolumesMounts string
	var delegateInitVolumesMounts string
	builder := utils.NewVolumeBuilder()

	if conf.RawStsConfig.StsEnable {
		delegateVolumesMountsData, err := sts.GenerateSecretVolumeMounts(sts.FaaSSchedulerName, builder)
		if err != nil {
			return err
		}
		delegateVolumesMounts = string(delegateVolumesMountsData)
	}

	delegateInitVolumesMounts = delegateVolumesMounts
	log.GetLogger().Debugf("delegateVolumes: %s, delegateVolumesMounts: %s, delegateInitVolumesMounts: %s",
		delegateVolumes, delegateVolumesMounts, delegateInitVolumesMounts)
	if delegateVolumes != "" {
		createOptions[commonconstant.DelegateVolumesKey] = delegateVolumes
	}
	if delegateVolumesMounts != "" {
		createOptions[commonconstant.DelegateVolumeMountKey] = delegateVolumesMounts
	}
	if delegateInitVolumesMounts != "" {
		createOptions[commonconstant.DelegateInitVolumeMountKey] = delegateInitVolumesMounts
	}
	return nil
}

func (s *SchedulerManager) createOrRetry(ctx context.Context, args []api.Arg, extraParams commonTypes.ExtraParams,
	enableRetry bool) error {
	// 只有首次拉起/扩容时insID为空,需要在拉起失败时防止进入失败回调中的重试逻辑
	if extraParams.DesignatedInstanceID == "" {
		instanceID := uuid.New().String()
		s.recreateInstanceIDMap.Store(instanceID, nil)
		extraParams.DesignatedInstanceID = instanceID
		defer s.recreateInstanceIDMap.Delete(instanceID)
	}
	if enableRetry {
		err := s.CreateWithRetry(ctx, args, &extraParams)
		if err != nil {
			return err
		}
	} else {
		instanceID := s.CreateInstance(ctx, types.FaaSSchedulerFunctionKey, args, &extraParams)
		if instanceID == "" {
			log.GetLogger().Errorf("failed to create scheduler manager instance")
			return errors.New("failed to create scheduler manager instance")
		}
	}
	return nil
}

// CreateWithRetry -
func (s *SchedulerManager) CreateWithRetry(ctx context.Context, args []api.Arg,
	extraParams *commonTypes.ExtraParams) error {
	err := wait.ExponentialBackoffWithContext(ctx, createInstanceBackoff, func(context.Context) (done bool, err error) {
		instanceID := s.CreateInstance(ctx, types.FaaSSchedulerFunctionKey, args, extraParams)
		if instanceID == "" {
			log.GetLogger().Errorf("failed to create scheduler instance")
			return false, nil
		}
		if instanceID == "cancelled" {
			return true, fmt.Errorf("create has been cancelled")
		}
		return true, nil
	})
	return err
}

func makePodLabel(createOptions map[string]string, tenantID string) {
	podLabels := map[string]string{
		constant.SystemFuncName:       constant.FuncNameFaasscheduler,
		constant.SchedulerExclusivity: tenantID,
	}
	labels, err := json.Marshal(podLabels)
	if err != nil {
		log.GetLogger().Warnf("faasscheduler label json marshal error : %s", err.Error())
		return
	}
	if createOptions == nil {
		return
	}
	createOptions[commonconstant.DelegatePodLabels] = string(labels)
}

// CreateInstance create scheduler instance
func (s *SchedulerManager) CreateInstance(ctx context.Context, function string, args []api.Arg,
	extraParams *commonTypes.ExtraParams) string {
	instanceID := extraParams.DesignatedInstanceID
	funcMeta := api.FunctionMeta{FuncID: function, Api: api.PosixApi, Name: &instanceID}
	log.GetLogger().Infof("start to create scheduler instance(FuncID=%s)", funcMeta.FuncID)
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
		_, createErr := s.sdkClient.CreateInstance(funcMeta, args, invokeOpts)
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
	timer := time.NewTimer(types.CreatedTimeout)
	select {
	case notifyErr, ok := <-createCh:
		defer timer.Stop()
		if !ok {
			log.GetLogger().Errorf("result channel of scheduler instance request is closed")
			return ""
		}
		if !notifyErr.IsOk() {
			log.GetLogger().Errorf("failed to bring up scheduler instance(id=%s), code:%d, err:%s", instanceID,
				notifyErr.Code, notifyErr.Error())
			s.clearInstanceAfterError(instanceID)
			return ""
		}
	case <-timer.C:
		log.GetLogger().Errorf("time out waiting for instance(id=%s) creation", instanceID)
		s.clearInstanceAfterError(instanceID)
		return ""
	case <-ctx.Done():
		log.GetLogger().Errorf("instance(id=%s) creation has been cancelled", instanceID)
		s.clearInstanceAfterError(instanceID)
		return "cancelled"
	}
	s.addInstance(instanceID)
	log.GetLogger().Infof("succeed to create scheduler instance(id=%s)", instanceID)
	return instanceID
}

func genFunctionConfig(tenantID string) ([]api.Arg, *commonTypes.ExtraParams, error) {
	config.SchedulerConfigLock.RLock()
	schedulerConfig := config.GetFaaSSchedulerConfig()
	extraParams, err := createExtraParams(schedulerConfig, tenantID)
	if err != nil {
		config.SchedulerConfigLock.RUnlock()
		log.GetLogger().Errorf("failed to prepare faasscheduler createExtraParams, err:%s", err.Error())
		return nil, nil, err
	}
	schedulerConf, err := json.Marshal(schedulerConfig)
	config.SchedulerConfigLock.RUnlock()
	if err != nil {
		log.GetLogger().Errorf("faaSScheduler config json marshal failed, err:%s", err.Error())
		return nil, nil, err
	}
	schedulerArgs := []api.Arg{
		{
			Type: api.Value,
			Data: schedulerConf,
		},
	}
	return schedulerArgs, extraParams, nil
}

// SyncCreateInstanceByID -
func (s *SchedulerManager) SyncCreateInstanceByID(ctx context.Context, instanceID string) error {
	var cancel context.CancelFunc
	if ctx == nil {
		ctx, cancel = context.WithCancel(context.Background())
	}
	_, exist := s.recreateInstanceIDMap.LoadOrStore(instanceID, cancel)
	if exist {
		log.GetLogger().Warnf("scheduler instance[%s] is creating", instanceID)
		return nil
	}
	defer s.recreateInstanceIDMap.Delete(instanceID)
	log.GetLogger().Infof("start to sync create scheduler instance:%s", instanceID)
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	args, extraParams, err := genFunctionConfig(s.tenantID)
	if err != nil {
		return err
	}
	extraParams.DesignatedInstanceID = instanceID
	err = s.createOrRetry(ctx, args, *extraParams, config.GetFaaSControllerConfig().EnableRetry)
	if err != nil {
		return err
	}
	return nil
}

// GetInstanceCountFromEtcd get current instance count from etcd
func (s *SchedulerManager) GetInstanceCountFromEtcd() map[string]struct{} {
	resp, err := s.etcdClient.Client.Get(context.TODO(), types.FaaSSchedulerPrefixKey, clientv3.WithPrefix())
	if err != nil {
		log.GetLogger().Errorf("failed to search etcd key, prefixKey=%s, err=%s", types.FaaSSchedulerPrefixKey,
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
	log.GetLogger().Infof("get etcd scheduler instance count=%d, %+v", resp.Count, instanceIDs)
	return instanceIDs
}

func (s *SchedulerManager) clearInstanceAfterError(instanceID string) {
	if err := s.sdkClient.Kill(instanceID, types.KillSignalVal, []byte{}); err != nil {
		log.GetLogger().Errorf("failed to kill scheduler instance: %s", instanceID)
	}
}

// addInstance add an instance to cache
func (s *SchedulerManager) addInstance(instanceID string) {
	s.Lock()
	_, exists := s.instanceCache[instanceID]
	if exists {
		log.GetLogger().Warnf("the instance(id=%s) already exist for %s", instanceID, s.tenantID)
		s.Unlock()
		return
	}
	log.GetLogger().Infof("add instance(id=%s) to cache for %s", instanceID, s.tenantID)
	s.instanceCache[instanceID] = &types.InstanceSpecification{InstanceID: instanceID}
	s.Unlock()
	state.Update(instanceID, types.StateUpdate, types.FaasSchedulerInstanceState+s.tenantID)
}

// GetInstanceCache acquire instance cache
func (s *SchedulerManager) GetInstanceCache() map[string]*types.InstanceSpecification {
	return s.instanceCache
}

// SyncKillAllInstance kill all instances of system function, faaS scheduler
func (s *SchedulerManager) SyncKillAllInstance() {
	s.Lock()
	cache := s.instanceCache
	s.instanceCache = map[string]*types.InstanceSpecification{}
	s.Unlock()
	var wg sync.WaitGroup
	for instanceID := range cache {
		wg.Add(1)
		go func(instanceID string) {
			defer wg.Done()
			if err := s.sdkClient.Kill(instanceID, types.SyncKillSignalVal, []byte{}); err != nil {
				log.GetLogger().Errorf("failed to kill scheduler instance(id=%s), err:%s", instanceID, err.Error())
				return
			}
			log.GetLogger().Infof("success to kill scheduler instance(id=%s)", instanceID)
		}(instanceID)
	}
	wg.Wait()
}

// KillInstance kill a scheduler instance
func (s *SchedulerManager) KillInstance(instanceID string) error {
	log.GetLogger().Infof("start to kill instance %s", instanceID)
	return wait.ExponentialBackoffWithContext(
		context.Background(), createInstanceBackoff, func(context.Context) (bool, error) {
			var err error
			err = s.sdkClient.Kill(instanceID, types.KillSignalVal, []byte{})
			if err != nil && !strings.Contains(err.Error(), "instance not found") {
				log.GetLogger().Warnf("failed to kill instanceID: %s, err: %s", instanceID, err.Error())
				return false, nil
			}
			return true, nil
		})
}

// KillExceptInstance -
func (s *SchedulerManager) KillExceptInstance(count int) error {
	if len(s.instanceCache) < count {
		return nil
	}
	for instanceID := range s.instanceCache {
		if count <= 0 {
			return nil
		}
		if err := s.KillInstance(instanceID); err != nil {
			log.GetLogger().Errorf("kill frontend instance:%s, error:%s", instanceID, err.Error())
			return err
		}
		count--
	}
	return nil
}

// RecoverInstance recover a faaS scheduler instance when faults occur
func (s *SchedulerManager) RecoverInstance(info *types.InstanceSpecification) {
	err := s.KillInstance(info.InstanceID)
	if err != nil {
		log.GetLogger().Warnf("failed to kill instanceID: %s, err: %s", info.InstanceID, err.Error())
	}
}

// HandleInstanceUpdate handles function update
func (s *SchedulerManager) HandleInstanceUpdate(schedulerSpec *types.InstanceSpecification) {
	// 有多种情况会调进此函数
	// 1. 创建实例回调, 新建实例会多次进入此方法
	// 2. controller首次监听etcd, etcd当前存量实例会进入此方法

	// 理论上进入到这个方法的实例只有以下两种
	// 1. 版本正确的实例, 需要写入缓存, 并判断实例个数, 超出的需要删除, 未超出不做处理
	// 2. 版本不正确的实例有两种情况:
	// (1) 该实例存在于terminating缓存中, 实例可能在进行滚动更新时, 这时应该不进行处理
	// (2) 该实例不在terminating缓存中, controller重启时, 更新了实例版本, 首次监听到了老的版本实例, 这时应该直接将实例删除
	log.GetLogger().Infof("handling scheduler instance %s update", schedulerSpec.InstanceID)
	if schedulerSpec.InstanceSpecificationMeta.InstanceStatus.Code == int(commonconstant.KernelInstanceStatusExiting) {
		log.GetLogger().Infof("scheduler instance %s is exiting,no need to update", schedulerSpec.InstanceID)
		return
	}
	config.SchedulerConfigLock.RLock()
	signature := controllerutils.GetSchedulerConfigSignature(config.GetFaaSSchedulerConfig())
	config.SchedulerConfigLock.RUnlock()
	if isExpectInstance(&schedulerSpec.InstanceSpecificationMeta, signature) {
		s.Lock()
		currentNum := len(s.instanceCache)
		_, exist := s.instanceCache[schedulerSpec.InstanceID]
		if currentNum > s.count || (currentNum == s.count && !exist) {
			log.GetLogger().Infof("current scheduler num is %d, kill the new instance %s for %s with cache: %v",
				currentNum, schedulerSpec.InstanceID, s.tenantID, s.instanceCache)
			delete(s.instanceCache, schedulerSpec.InstanceID)
			s.Unlock()
			if err := s.sdkClient.Kill(schedulerSpec.InstanceID, types.KillSignalVal, []byte{}); err != nil {
				log.GetLogger().Errorf("failed to kill instance %s error:%s", schedulerSpec.InstanceID, err.Error())
			}
			return
		}
		// add instance to cache if not exist, otherwise update the instance
		if !exist {
			log.GetLogger().Infof("instance %s has been added to cache", schedulerSpec.InstanceID)
			s.instanceCache[schedulerSpec.InstanceID] = schedulerSpec
			state.Update(schedulerSpec.InstanceID, types.StateUpdate, types.FaasSchedulerInstanceState+s.tenantID)
			s.Unlock()
			return
		}
		s.instanceCache[schedulerSpec.InstanceID].InstanceSpecificationMeta = schedulerSpec.InstanceSpecificationMeta
		log.GetLogger().Infof("scheduler instance %s is updated, refresh instance cache",
			schedulerSpec.InstanceID)
		s.Unlock()
		return
	}
	s.RLock()
	_, exist := s.terminalCache[schedulerSpec.InstanceID]
	s.RUnlock()
	if !exist {
		if err := s.KillInstance(schedulerSpec.InstanceID); err != nil {
			log.GetLogger().Errorf("failed to kill instance %s error:%s", schedulerSpec.InstanceID, err.Error())
		}
	}
}

// HandleInstanceDelete handles function delete
func (s *SchedulerManager) HandleInstanceDelete(schedulerSpec *types.InstanceSpecification) {
	// Here will depend on the kernel recover capability, not need to maintain the number of instances
	log.GetLogger().Infof("handling scheduler instance %s delete", schedulerSpec.InstanceID)
	config.SchedulerConfigLock.RLock()
	signature := controllerutils.GetSchedulerConfigSignature(config.GetFaaSSchedulerConfig())
	config.SchedulerConfigLock.RUnlock()
	s.Lock()
	delete(s.instanceCache, schedulerSpec.InstanceID)
	s.Unlock()
	if isExpectInstance(&schedulerSpec.InstanceSpecificationMeta, signature) {
		s.RLock()
		if len(s.instanceCache) < s.count {
			log.GetLogger().Infof("current faasscheduler instance num is %d, need to recreate instance: %s",
				len(s.instanceCache), schedulerSpec.InstanceID)
			go func() {
				err := s.SyncCreateInstanceByID(nil, schedulerSpec.InstanceID)
				if err != nil {
					log.GetLogger().Errorf("failed to create instance: %s, err:%v", schedulerSpec.InstanceID,
						err)
				}
			}()
			s.RUnlock()
			return
		} else {
			state.Update(schedulerSpec.InstanceID, types.StateDelete, types.FaasSchedulerInstanceState+s.tenantID)
		}
		s.RUnlock()
	}
	cancel, exist := s.recreateInstanceIDMap.Load(schedulerSpec.InstanceID)
	if exist {
		if cancelFunc, ok := cancel.(context.CancelFunc); ok {
			cancelFunc()
			log.GetLogger().Infof("instance %s bring up has been canceled", schedulerSpec.InstanceID)
			return
		}
		log.GetLogger().Errorf("get cancel func failed from instanceIDMap, instanceID:%s",
			schedulerSpec.InstanceID)
	}
}

// RollingUpdate rolling update
func (s *SchedulerManager) RollingUpdate(ctx context.Context, event *types.ConfigChangeEvent) {
	// 1. 更新 预期实例个数
	// 2. 把不符合预期的instanceCache -> terminalCache
	// 3. 从terminalCache随机删除一个实例 同步
	// 4. 创建新实例 同步
	// 5. terminalCache为空时将实例数调谐为预期实例数(同步), instanceCache到达预期数量时清空terminalCache(异步)
	// !!!scheduler启动时需要等待hash环上的所有scheduler全部启动才能初始化完成，所以当正在运行的scheduler数和预期拉起的数量不符时，需要全量升级
	newSign := controllerutils.GetSchedulerConfigSignature(event.SchedulerCfg)
	s.Lock()
	runningNum := 0
	s.count = event.SchedulerCfg.SchedulerNum
	// exclusivity tenant only need 1 scheduler instance
	if s.tenantID != "" {
		s.count = 1
	}
	for _, ins := range s.instanceCache {
		if ins.InstanceSpecificationMeta.InstanceStatus.Code == int(commonconstant.KernelInstanceStatusRunning) {
			runningNum++
		}
		if !isExpectInstance(&ins.InstanceSpecificationMeta, newSign) {
			log.GetLogger().Infof("instance:%s is waiting for termination", ins.InstanceID)
			s.terminalCache[ins.InstanceID] = ins
			delete(s.instanceCache, ins.InstanceID)
		}
	}
	s.Unlock()
	if runningNum != s.count {
		log.GetLogger().Infof("running scheduler:%d, expected:%d, need full update", runningNum, s.count)
		err := s.fullUpdate(ctx)
		event.Error = err
		event.Done()
		return
	}
	for {
		select {
		case <-ctx.Done():
			event.Error = fmt.Errorf("rolling update has stopped")
			event.Done()
			return
		default:
		}
		s.RLock()
		if len(s.instanceCache) == s.count {
			s.RUnlock()
			log.GetLogger().Infof("instance count arrive at expectation:%d, delete all terminating instance",
				s.count)
			go s.killAllTerminalInstance()
			event.Done()
			return
		}
		if len(s.terminalCache) == 0 {
			s.RUnlock()
			log.GetLogger().Infof("no instance need to terminate, pull up missing instance count:%d",
				s.count-len(s.instanceCache))
			err := s.CreateExpectedInstanceCount(ctx)
			if err != nil {
				event.Error = err
			}
			event.Done()
			return
		}
		insID := ""
		for _, ins := range s.terminalCache {
			insID = ins.InstanceID
			break
		}
		s.RUnlock()
		var err error
		if err = s.sdkClient.Kill(insID, types.SyncKillSignalVal, []byte{}); err != nil {
			log.GetLogger().Errorf("failed to sync kill scheduler instance(id=%s), err:%v", insID, err)
		}
		s.Lock()
		delete(s.terminalCache, insID)
		s.Unlock()
		time.Sleep(constant.RecreateSleepTime)
		err = s.SyncCreateInstanceByID(ctx, insID)
		if err != nil {
			event.Error = err
			event.Done()
			return
		}
	}
}

func (s *SchedulerManager) killAllTerminalInstance() {
	s.Lock()
	insMap := s.terminalCache
	s.terminalCache = map[string]*types.InstanceSpecification{}
	s.Unlock()
	var wg sync.WaitGroup
	for _, ins := range insMap {
		insID := ins.InstanceID
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := s.KillInstance(insID)
			if err != nil {
				log.GetLogger().Errorf("Failed to kill instance %v,err: %v", insID, err)
			}
		}()
	}
	wg.Wait()
}

func (s *SchedulerManager) configChangeProcessor(ctx context.Context, cancel context.CancelFunc) {
	if ctx == nil || cancel == nil || s.ConfigChangeCh == nil {
		return
	}
	for {
		select {
		case cfgEvent, ok := <-s.ConfigChangeCh:
			if !ok {
				cancel()
				return
			}
			schedulerConfig, schedulerInsNum := s.ConfigDiff(cfgEvent)
			if schedulerConfig != nil || schedulerInsNum != -1 {
				log.GetLogger().Infof("scheduler config or instance num is changed," +
					" need to update scheduler instance")
				cancel()
				ctx, cancel = context.WithCancel(context.Background())
				config.UpdateSchedulerConfig(cfgEvent.SchedulerCfg)
				cfgEvent.Add(1)
				go s.RollingUpdate(ctx, cfgEvent)
			} else {
				log.GetLogger().Infof("scheduler config is same as current, no need to update")
			}
			cfgEvent.Done()
		}
	}
}

// ConfigDiff config diff
func (s *SchedulerManager) ConfigDiff(event *types.ConfigChangeEvent) (*types.SchedulerConfig, int) {
	newSign := controllerutils.GetSchedulerConfigSignature(event.SchedulerCfg)
	config.SchedulerConfigLock.RLock()
	schedulerOldCfg := config.GetFaaSSchedulerConfig()
	config.SchedulerConfigLock.RUnlock()
	if strings.Compare(newSign,
		controllerutils.GetSchedulerConfigSignature(schedulerOldCfg)) == 0 {
		if event.SchedulerCfg.SchedulerNum != schedulerOldCfg.SchedulerNum {
			config.SchedulerConfigLock.Lock()
			schedulerOldCfg.SchedulerNum = event.SchedulerCfg.SchedulerNum
			config.SchedulerConfigLock.Unlock()
			return nil, event.SchedulerCfg.SchedulerNum
		}
		return nil, -1
	}
	return event.SchedulerCfg, event.SchedulerCfg.SchedulerNum
}

func (s *SchedulerManager) fullUpdate(ctx context.Context) error {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.killAllTerminalInstance()
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.SyncKillAllInstance()
	}()
	wg.Wait()
	time.Sleep(constant.RecreateSleepTime)
	return s.CreateExpectedInstanceCount(ctx)
}
