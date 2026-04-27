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

// Package faasfrontendmanager manages faasfrontend status and instance ID
package faasfrontendmanager

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
	"k8s.io/api/core/v1"
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

const (
	dsWorkerUnreadyKey = "is-ds-worker-unready"
)

var (
	once sync.Once
	// faasFrontendManager is the singaleton of Manager
	frontendManager *FrontendManager

	createInstanceBackoff = wait.Backoff{
		Steps:    constant.DefaultCreateRetryTime, // retry times (include first time)
		Duration: constant.DefaultCreateRetryDuration,
		Factor:   constant.DefaultCreateRetryFactor,
		Jitter:   constant.DefaultCreateRetryJitter,
	}
)

// FrontendManager manages faasfrontend status and instance ID
type FrontendManager struct {
	instanceCache         map[string]*types.InstanceSpecification
	terminalCache         map[string]*types.InstanceSpecification
	etcdClient            *etcd3.EtcdClient
	sdkClient             api.LibruntimeAPI
	ConfigChangeCh        chan *types.ConfigChangeEvent
	recreateInstanceIDCh  chan string
	recreateInstanceIDMap sync.Map
	stopCh                chan struct{}
	sync.RWMutex
	count     int
	isDynamic bool
}

// GetFrontendManager can only be called after NewFrontendManager
func GetFrontendManager() *FrontendManager {
	return frontendManager
}

// NewFaaSFrontendManager supply a singleton frontend manager
func NewFaaSFrontendManager(libruntimeAPI api.LibruntimeAPI,
	etcdClient *etcd3.EtcdClient, stopCh chan struct{}, size int, isDynamic bool) *FrontendManager {
	once.Do(func() {
		frontendManager = &FrontendManager{
			instanceCache:        make(map[string]*types.InstanceSpecification),
			terminalCache:        map[string]*types.InstanceSpecification{},
			etcdClient:           etcdClient,
			sdkClient:            libruntimeAPI,
			count:                size,
			stopCh:               stopCh,
			recreateInstanceIDCh: make(chan string, constant.DefaultChannelSize),
			ConfigChangeCh:       make(chan *types.ConfigChangeEvent, constant.DefaultChannelSize),
			isDynamic:            isDynamic,
		}
		go frontendManager.recreateInstance()
		frontendManager.initInstanceCache(etcdClient)
		ctx, cancelFunc := context.WithCancel(context.Background())
		go func() {
			err := frontendManager.CreateExpectedInstanceCount(ctx)
			if err != nil {
				log.GetLogger().Errorf("Failed to create expected frontend instance count, error: %v", err)
			}
		}()
		go frontendManager.configChangeProcessor(ctx, cancelFunc)
	})
	return frontendManager
}

func (ffm *FrontendManager) initInstanceCache(etcdClient *etcd3.EtcdClient) {
	response, err := etcdClient.Client.Get(context.Background(), types.FasSFrontendPrefixKey, clientv3.WithPrefix())
	if err != nil {
		log.GetLogger().Errorf("Failed to get frontend instance: %v", err)
		return
	}
	config.FrontendConfigLock.RLock()
	targetSign := controllerutils.GetFrontendConfigSignature(config.GetFaaSFrontendConfig())
	config.FrontendConfigLock.RUnlock()
	for _, kv := range response.Kvs {
		meta := &types.InstanceSpecificationMeta{}
		err := json.Unmarshal(kv.Value, meta)
		if err != nil {
			log.GetLogger().Errorf("Failed to unmarshal instance specification: %v", err)
			continue
		}
		if isExceptInstance(meta, targetSign) {
			funcCtx, cancelFunc := context.WithCancel(context.TODO())
			ffm.instanceCache[meta.InstanceID] = &types.InstanceSpecification{
				FuncCtx:                   funcCtx,
				CancelFunc:                cancelFunc,
				InstanceID:                meta.InstanceID,
				InstanceSpecificationMeta: *meta,
			}
			log.GetLogger().Infof("find expected frontend instance %s add to cache", meta.InstanceID)
		}
	}
}

func isExceptInstance(meta *types.InstanceSpecificationMeta, targetSign string) bool {
	if len(meta.Args) == 0 {
		log.GetLogger().Errorf("args is empty, %v", meta)
		return false
	}
	value := meta.Args[0]["value"]
	s, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		log.GetLogger().Errorf("Failed to decode args: %v", err)
		return false
	}
	cfg := &types.FrontendConfig{}
	err = json.Unmarshal(s, cfg)
	if err != nil && len(s) > commonconstant.LibruntimeHeaderSize {
		// args in libruntime create request with 16 bytes header.
		// except libruntime, if other modules use this field, should try yo delete the header
		err = json.Unmarshal(s[commonconstant.LibruntimeHeaderSize:], cfg)
		if err != nil {
			log.GetLogger().Errorf("Failed to unmarshal frontend config: %v, value: %s", err, s)
			return false
		}
	}
	oldSign := controllerutils.GetFrontendConfigSignature(cfg)
	if oldSign == "" {
		log.GetLogger().Errorf("old sign is empty, insID:%s", meta.InstanceID)
		return false
	}
	log.GetLogger().Infof("frontend(%s) sign: %s, expect sign: %s", meta.InstanceID, oldSign, targetSign)
	return strings.Compare(oldSign, targetSign) == 0
}

// GetInstanceCountFromEtcd get current instance count from etcd
func (ffm *FrontendManager) GetInstanceCountFromEtcd() map[string]struct{} {
	resp, err := ffm.etcdClient.Client.Get(context.TODO(), types.FasSFrontendPrefixKey, clientv3.WithPrefix())
	if err != nil {
		log.GetLogger().Errorf("failed to search etcd key, prefixKey=%s, err=%s", types.FasSFrontendPrefixKey, err.Error())
		return nil
	}
	instanceIDs := make(map[string]struct{}, resp.Count)
	for _, kv := range resp.Kvs {
		instanceID := controllerutils.ExtractInfoFromEtcdKey(string(kv.Key), commonconstant.InstanceIDIndexForInstance)
		if instanceID != "" {
			instanceIDs[instanceID] = struct{}{}
		}
	}
	log.GetLogger().Infof("get etcd frontend instance count=%d, %+v", resp.Count, instanceIDs)
	return instanceIDs
}

// CreateExpectedInstanceCount create expected frontend instance count
func (ffm *FrontendManager) CreateExpectedInstanceCount(ctx context.Context) error {
	// 此方法在启动的时候调用一次或者在配置更新的时候调用, 目的是将frontend实例数量补充至设置的实例数
	// 不需要删除多余实例, 多余实例会在 HandleInstanceUpdate 中删除
	// frontend的 instanceID不需要保持一致
	ffm.RLock()
	currentCount := len(ffm.instanceCache)
	expectedCount := ffm.count - currentCount
	ffm.RUnlock()
	return ffm.CreateMultiInstances(ctx, expectedCount)
}

// CreateMultiInstances create multi instances
func (ffm *FrontendManager) CreateMultiInstances(ctx context.Context, count int) error {
	if count <= 0 {
		log.GetLogger().Infof("no need to create frontend instance, kill %d instances instead.", -count)
		return ffm.KillExceptInstance(-count)
	}
	log.GetLogger().Infof("need to create %d faas frontend instances", count)

	args, params, err := genFunctionConfig()
	if err != nil {
		return err
	}

	var createErr error
	g := &sync.WaitGroup{}
	for i := 0; i < count; i++ {
		g.Add(1)
		go func() {
			defer g.Done()
			if err = ffm.createOrRetry(ctx, args, *params,
				config.GetFaaSControllerConfig().EnableRetry); err != nil {
				createErr = err
			}
		}()
	}
	g.Wait()
	if createErr != nil {
		return createErr
	}
	log.GetLogger().Infof("succeed to create %d faaS frontend instances", count)
	return nil
}

func createExtraParams(conf *types.FrontendConfig) (*commonTypes.ExtraParams, error) {
	extraParams := &commonTypes.ExtraParams{}
	extraParams.Resources = utils.GenerateResourcesMap(conf.CPU, conf.Memory)
	extraParams.CustomExtensions = utils.CreateCustomExtensions(extraParams.CustomExtensions,
		utils.MonopolyPolicyValue)
	extraParams.ScheduleAffinities = utils.CreatePodAffinity(constant.SystemFuncName, constant.FuncNameFaasfrontend,
		api.PreferredAntiAffinity)
	utils.AddNodeSelector(conf.NodeSelector, extraParams)
	createOpt, err := prepareCreateOptions(conf)
	extraParams.Label = []string{constant.FuncNameFaasfrontend}
	extraParams.CreateOpt = createOpt
	return extraParams, err
}

func genFunctionConfig() ([]api.Arg, *commonTypes.ExtraParams, error) {
	config.FrontendConfigLock.RLock()
	frontendConfig := config.GetFaaSFrontendConfig()
	extraParams, err := createExtraParams(frontendConfig)
	if err != nil {
		config.FrontendConfigLock.RUnlock()
		log.GetLogger().Errorf("failed to prepare faaSFrontend createExtraParams, err:%s", err.Error())
		return nil, nil, err
	}
	frontendConf, err := json.Marshal(frontendConfig)
	config.FrontendConfigLock.RUnlock()
	if err != nil {
		log.GetLogger().Errorf("faaSFrontend config json marshal failed, err:%s", err.Error())
		return nil, nil, err
	}
	args := []api.Arg{
		{
			Type: api.Value,
			Data: frontendConf,
		},
	}
	return args, extraParams, nil
}

func (ffm *FrontendManager) createOrRetry(ctx context.Context, args []api.Arg, extraParams commonTypes.ExtraParams,
	enableRetry bool) error {
	// 只有首次拉起/扩容时insID为空,需要在拉起失败时防止进入失败回调中的重试逻辑
	if extraParams.DesignatedInstanceID == "" {
		instanceID := uuid.New().String()
		ffm.recreateInstanceIDMap.Store(instanceID, nil)
		extraParams.DesignatedInstanceID = instanceID
	}
	defer ffm.recreateInstanceIDMap.Delete(extraParams.DesignatedInstanceID)
	if enableRetry {
		err := ffm.CreateWithRetry(ctx, args, &extraParams)
		if err != nil {
			return err
		}
	} else {
		instanceID := ffm.CreateInstance(ctx, types.FasSFrontendFunctionKey, args, &extraParams)
		if instanceID == "" {
			log.GetLogger().Errorf("failed to create frontend instance")
			return errors.New("failed to create frontend instance")
		}
	}
	return nil
}

// CreateWithRetry -
func (ffm *FrontendManager) CreateWithRetry(
	ctx context.Context, args []api.Arg, extraParams *commonTypes.ExtraParams,
) error {
	err := wait.ExponentialBackoffWithContext(
		ctx, createInstanceBackoff, func(context.Context) (done bool, err error) {
			instanceID := ffm.CreateInstance(ctx, types.FasSFrontendFunctionKey, args, extraParams)
			if instanceID == "" {
				log.GetLogger().Errorf("failed to create frontend instance")
				return false, nil
			}
			if instanceID == "cancelled" {
				return true, fmt.Errorf("create has been cancelled")
			}
			return true, nil
		})
	return err
}

// CreateInstance create an instance of system function, faaS frontend
func (ffm *FrontendManager) CreateInstance(ctx context.Context, function string, args []api.Arg,
	extraParams *commonTypes.ExtraParams) string {
	instanceID := extraParams.DesignatedInstanceID
	funcMeta := api.FunctionMeta{FuncID: function, Api: api.PosixApi, Name: &instanceID}
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
		_, createErr := ffm.sdkClient.CreateInstance(funcMeta, args, invokeOpts)
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
	defer timer.Stop()
	select {
	case err, ok := <-createCh:
		if !ok {
			log.GetLogger().Errorf("result channel of frontend instance request is closed")
			return ""
		}
		if !err.IsOk() {
			log.GetLogger().Errorf("failed to bring up frontend instance(id=%s),code:%d,err:%s",
				instanceID, err.Code, err.Error())
			ffm.clearInstanceAfterError(instanceID)
			return ""
		}
	case <-timer.C:
		log.GetLogger().Errorf("time out waiting for instance creation")
		ffm.clearInstanceAfterError(instanceID)
		return ""
	case <-ctx.Done():
		log.GetLogger().Errorf("create instance has been cancelled")
		ffm.clearInstanceAfterError(instanceID)
		return "cancelled"
	}
	log.GetLogger().Infof("succeed to create frontend instance(id=%s)", instanceID)
	ffm.addInstance(instanceID)
	return instanceID
}

func (ffm *FrontendManager) clearInstanceAfterError(instanceID string) {
	var err error
	err = ffm.KillInstance(instanceID)
	if err != nil {
		log.GetLogger().Errorf("failed to kill frontend instance: %s", instanceID)
	}
}

func (ffm *FrontendManager) addInstance(instanceID string) {
	ffm.Lock()
	defer ffm.Unlock()
	_, exist := ffm.instanceCache[instanceID]
	if exist {
		log.GetLogger().Warnf("the frontend instance(id=%s) already exist", instanceID)
		return
	}
	log.GetLogger().Infof("add instance(id=%s) to local cache", instanceID)
	ffm.instanceCache[instanceID] = &types.InstanceSpecification{InstanceID: instanceID}
	state.Update(instanceID, types.StateUpdate, types.FaasFrontendInstanceState)
}

// GetInstanceCache supply local instance cache
func (ffm *FrontendManager) GetInstanceCache() map[string]*types.InstanceSpecification {
	return ffm.instanceCache
}

// SyncKillAllInstance kill all instances of system function, faaS frontend
func (ffm *FrontendManager) SyncKillAllInstance() {
	var wg sync.WaitGroup
	ffm.Lock()
	defer ffm.Unlock()
	var deletedInstance []string
	for instanceID := range ffm.instanceCache {
		wg.Add(1)
		go func(instanceID string) {
			defer wg.Done()
			if err := ffm.sdkClient.Kill(instanceID, types.SyncKillSignalVal, []byte{}); err != nil {
				log.GetLogger().Errorf("failed to kill frontend instance(id=%s), err:%s", instanceID, err.Error())
				return
			}
			deletedInstance = append(deletedInstance, instanceID)
			log.GetLogger().Infof("success to kill frontend instance(id=%s)", instanceID)
		}(instanceID)
	}
	wg.Wait()
	for _, instanceID := range deletedInstance {
		log.GetLogger().Infof("delete frontend instance(id=%s) from local cache", instanceID)
		delete(ffm.instanceCache, instanceID)
		state.Update(instanceID, types.StateDelete, types.FaasFrontendInstanceState)
		ffm.count--
	}
}

// KillInstance kill an instance of system function, faaS frontend
func (ffm *FrontendManager) KillInstance(instanceID string) error {
	log.GetLogger().Infof("start to kill instance %s", instanceID)
	return wait.ExponentialBackoffWithContext(
		context.Background(), createInstanceBackoff, func(context.Context) (bool, error) {
			var err error
			err = ffm.sdkClient.Kill(instanceID, types.KillSignalVal, []byte{})
			if err != nil && !strings.Contains(err.Error(), "instance not found") {
				log.GetLogger().Warnf("failed to kill instanceID: %s, err: %s", instanceID, err.Error())
				return false, nil
			}
			return true, nil
		})
}

// KillExceptInstance -
func (ffm *FrontendManager) KillExceptInstance(count int) error {
	if len(ffm.instanceCache) < count {
		return nil
	}
	if ffm.isDynamic {
		// will be scaled down by function-master, no need to process it
		return nil
	}
	for instanceID := range ffm.instanceCache {
		if count <= 0 {
			return nil
		}
		if err := ffm.KillInstance(instanceID); err != nil {
			log.GetLogger().Errorf("kill frontend instance:%s, error:%s", instanceID, err.Error())
			return err
		}
		count--
	}
	return nil
}

// RecoverInstance recover a faaS frontend instance when faults occur
func (ffm *FrontendManager) RecoverInstance(info *types.InstanceSpecification) {
	err := ffm.KillInstance(info.InstanceID)
	if err != nil {
		log.GetLogger().Warnf("failed to kill instanceID: %s, err: %s", info.InstanceID, err.Error())
	}
}

func (ffm *FrontendManager) recreateInstance() {
	for {
		select {
		case <-ffm.stopCh:
			return
		case instanceID, ok := <-ffm.recreateInstanceIDCh:
			if !ok {
				log.GetLogger().Warnf("recreateInstanceIDCh is closed")
				return
			}
			ffm.RLock()
			if _, exist := ffm.instanceCache[instanceID]; exist || len(ffm.instanceCache) >= ffm.count {
				log.GetLogger().Infof("current frontend num is %d, no need  to recreate instance:%s",
					len(ffm.instanceCache), instanceID)
				ffm.RUnlock()
				break
			}
			ffm.RUnlock()
			ctx, cancel := context.WithCancel(context.Background())
			_, loaded := ffm.recreateInstanceIDMap.LoadOrStore(instanceID, cancel)
			if loaded {
				log.GetLogger().Warnf("instance[%s] is recreating", instanceID)
				break
			}
			args, extraParams, err := genFunctionConfig()
			if err != nil {
				log.GetLogger().Errorf("failed to prepare createExtraParams, err:%s", err.Error())
				break
			}
			extraParams.DesignatedInstanceID = instanceID
			go func() {
				time.Sleep(constant.RecreateSleepTime)
				log.GetLogger().Infof("start to recover faaSFrontend instance: %s", instanceID)
				if err = ffm.createOrRetry(ctx, args, *extraParams,
					config.GetFaaSControllerConfig().EnableRetry); err != nil {
					log.GetLogger().Errorf("failed to recreate instance: %s", instanceID)
				}
			}()
		}
	}
}

// SyncCreateInstance -
func (ffm *FrontendManager) SyncCreateInstance(ctx context.Context) error {
	log.GetLogger().Infof("start to sync create frontend instance")
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	args, extraParams, err := genFunctionConfig()
	if err != nil {
		return err
	}
	err = ffm.createOrRetry(ctx, args, *extraParams, config.GetFaaSControllerConfig().EnableRetry)
	if err != nil {
		return err
	}
	return nil
}

// HandleInstanceUpdate handle the etcd PUT event
func (ffm *FrontendManager) HandleInstanceUpdate(instanceSpec *types.InstanceSpecification) {
	log.GetLogger().Infof("handling frontend instance %s update", instanceSpec.InstanceID)
	if instanceSpec.InstanceSpecificationMeta.InstanceStatus.Code == int(commonconstant.KernelInstanceStatusExiting) {
		log.GetLogger().Infof("frontend instance %s is exiting,no need to update", instanceSpec.InstanceID)
		return
	}
	config.FrontendConfigLock.RLock()
	signature := controllerutils.GetFrontendConfigSignature(config.GetFaaSFrontendConfig())
	config.FrontendConfigLock.RUnlock()

	if isExceptInstance(&instanceSpec.InstanceSpecificationMeta, signature) {
		ffm.Lock()
		currentNum := len(ffm.instanceCache)
		_, exist := ffm.instanceCache[instanceSpec.InstanceID]
		if currentNum > ffm.count || (currentNum == ffm.count && !exist) {
			log.GetLogger().Infof("current frontend num is %s, kill the new instance %s",
				currentNum, instanceSpec.InstanceID)
			delete(ffm.instanceCache, instanceSpec.InstanceID)
			ffm.Unlock()
			if err := ffm.sdkClient.Kill(instanceSpec.InstanceID, types.KillSignalVal, []byte{}); err != nil {
				log.GetLogger().Errorf("failed to kill instance %s error:%s", instanceSpec.InstanceID,
					err.Error())
			}
			return
		}
		// add instance to cache if not exist, otherwise update the instance
		if !exist {
			ffm.instanceCache[instanceSpec.InstanceID] = instanceSpec
			log.GetLogger().Infof("add frontend instance %s to cache", instanceSpec.InstanceID)
			ffm.Unlock()
			state.Update(instanceSpec.InstanceID, types.StateUpdate, types.FaasFrontendInstanceState)
			return
		}
		ffm.instanceCache[instanceSpec.InstanceID].InstanceSpecificationMeta = instanceSpec.InstanceSpecificationMeta
		log.GetLogger().Infof("frontend instance %s is updated, refresh instance cache", instanceSpec.InstanceID)
		ffm.Unlock()
		return
	}
	ffm.RLock()
	_, exist := ffm.terminalCache[instanceSpec.InstanceID]
	ffm.RUnlock()
	if !exist {
		log.GetLogger().Infof("frontend instance %s is not expected, start to delete", instanceSpec.InstanceID)
		if err := ffm.KillInstance(instanceSpec.InstanceID); err != nil {
			log.GetLogger().Errorf("failed to kill instance %s error:%s", instanceSpec.InstanceID, err.Error())
		}
	}
}

// HandleInstanceDelete handle the etcd DELETE event
func (ffm *FrontendManager) HandleInstanceDelete(instanceSpec *types.InstanceSpecification) {
	log.GetLogger().Infof("handling frontend instance %s delete", instanceSpec.InstanceID)
	config.FrontendConfigLock.RLock()
	signature := controllerutils.GetFrontendConfigSignature(config.GetFaaSFrontendConfig())
	config.FrontendConfigLock.RUnlock()
	ffm.Lock()
	delete(ffm.instanceCache, instanceSpec.InstanceID)
	ffm.Unlock()
	state.Update(instanceSpec.InstanceID, types.StateDelete, types.FaasFrontendInstanceState)
	if isExceptInstance(&instanceSpec.InstanceSpecificationMeta, signature) {
		ffm.RLock()
		if len(ffm.instanceCache) < ffm.count {
			log.GetLogger().Infof("current faaSFrontend instance num is %d, need to recreate instance: %s",
				len(ffm.instanceCache), instanceSpec.InstanceID)
			ffm.RUnlock()
			ffm.recreateInstanceIDCh <- instanceSpec.InstanceID
			return
		}
		ffm.RUnlock()
	}
	cancel, exist := ffm.recreateInstanceIDMap.Load(instanceSpec.InstanceID)
	if exist {
		if cancelFunc, ok := cancel.(context.CancelFunc); ok {
			cancelFunc()
			log.GetLogger().Infof("instance %s bring up has been canceled", instanceSpec.InstanceID)
			return
		}
		log.GetLogger().Errorf("get cancel func failed from instanceIDMap, instanceID:%s",
			instanceSpec.InstanceID)
	}
}

// prepareCreateOptions for create faasfrontend createOpt
func prepareCreateOptions(conf *types.FrontendConfig) (map[string]string, error) {
	podLabels := map[string]string{
		constant.SystemFuncName: constant.FuncNameFaasfrontend,
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

	createOptions := map[string]string{
		commonconstant.DelegatePodLabels:          string(labels),
		commonconstant.DelegateRuntimeManagerTag:  string(delegateRuntime),
		commonconstant.InstanceLifeCycle:          commonconstant.InstanceLifeCycleDetached,
		commonconstant.DelegateNodeAffinity:       conf.NodeAffinity,
		commonconstant.DelegateNodeAffinityPolicy: conf.NodeAffinityPolicy,
	}

	if config.GetFaaSControllerConfig().RawStsConfig.StsEnable {
		secretVolumeMounts, err := sts.GenerateSecretVolumeMounts(sts.FaasfrontendName, utils.NewVolumeBuilder())
		if err != nil {
			return nil, fmt.Errorf("secretVolumeMounts json marshal failed, err:%s", err.Error())
		}
		createOptions[commonconstant.DelegateVolumeMountKey] = string(secretVolumeMounts)
	}

	if conf.Affinity != "" {
		createOptions[commonconstant.DelegateAffinity] = conf.Affinity
	}

	if conf.NodeSelector != nil && len(conf.NodeSelector) != 0 {
		var podTolerations []v1.Toleration
		for k, v := range config.GetFaaSFrontendConfig().NodeSelector {
			podTolerations = append(podTolerations, v1.Toleration{
				Key:      k,
				Operator: v1.TolerationOpEqual,
				Value:    v,
				Effect:   v1.TaintEffectNoSchedule,
			})
		}
		podTolerations = append(podTolerations, v1.Toleration{
			Key:      dsWorkerUnreadyKey,
			Operator: v1.TolerationOpEqual,
			Value:    "true",
			Effect:   v1.TaintEffectPreferNoSchedule,
		})
		tolerations, err := json.Marshal(podTolerations)
		if err != nil {
			return nil, fmt.Errorf("pod tolerations json marshal failed, err:%s", err.Error())
		}
		createOptions[commonconstant.DelegateTolerations] = string(tolerations)
	}

	err = prepareVolumesAndMounts(conf, createOptions)
	if err != nil {
		return nil, err
	}

	return createOptions, nil
}

func prepareVolumesAndMounts(conf *types.FrontendConfig, createOptions map[string]string) error {
	if createOptions == nil {
		return fmt.Errorf("createOptions is nil")
	}
	var delegateVolumes string
	var delegateVolumesMounts string
	var delegateInitVolumesMounts string
	var err error
	builder := utils.NewVolumeBuilder()
	if conf.RawStsConfig.StsEnable {
		delegateVolumesMountsData, err := sts.GenerateSecretVolumeMounts(sts.FaasfrontendName, builder)
		if err != nil {
			return fmt.Errorf("secretVolumeMounts json marshal failed, err:%s", err.Error())
		}
		delegateVolumesMounts = string(delegateVolumesMountsData)
	}
	if conf.HTTPSConfig != nil && conf.HTTPSConfig.HTTPSEnable {
		delegateVolumes, delegateVolumesMounts, err = sts.GenerateHTTPSAndLocalSecretVolumeMounts(*conf.HTTPSConfig, builder)
		if err != nil {
			log.GetLogger().Errorf("failed to generate https volumes and mounts")
			return fmt.Errorf("httpsVolumeMounts json marshal failed, err:%s", err.Error())
		}
	}

	delegateInitVolumesMounts = delegateVolumesMounts

	if delegateVolumes != "" {
		createOptions[commonconstant.DelegateVolumesKey] = delegateVolumes
	}
	if delegateVolumesMounts != "" {
		createOptions[commonconstant.DelegateVolumeMountKey] = delegateVolumesMounts
	}
	if delegateInitVolumesMounts != "" {
		createOptions[commonconstant.DelegateInitVolumeMountKey] = delegateInitVolumesMounts
	}
	log.GetLogger().Debugf("delegateVolumes: %s, delegateVolumesMounts: %s, delegateInitVolumesMounts: %s",
		delegateVolumes, delegateVolumesMounts, delegateInitVolumesMounts)
	return nil
}

// RollingUpdate rolling update frontend
func (ffm *FrontendManager) RollingUpdate(ctx context.Context, event *types.ConfigChangeEvent) {
	// 1. 更新 预期实例个数
	// 2. 把不符合预期的instanceCache -> terminalCache
	// 3. 从terminalCache随机删除一个实例 同步
	// 4. 创建新实例 同步
	// 5. terminalCache为空时将实例数调谐为预期实例数(同步), instanceCache到达预期数量时清空terminalCache(异步)
	newSign := controllerutils.GetFrontendConfigSignature(event.FrontendCfg)
	ffm.Lock()
	ffm.count = event.FrontendCfg.InstanceNum
	for _, ins := range ffm.instanceCache {
		if !isExceptInstance(&ins.InstanceSpecificationMeta, newSign) {
			ffm.terminalCache[ins.InstanceID] = ins
			delete(ffm.instanceCache, ins.InstanceID)
		}
	}
	ffm.Unlock()
	for {
		select {
		case <-ctx.Done():
			event.Error = fmt.Errorf("rolling update has stopped")
			event.Done()
			return
		default:
		}
		ffm.RLock()
		if len(ffm.instanceCache) == ffm.count {
			ffm.RUnlock()
			log.GetLogger().Infof("frontend instance count arrive at expectation:%d,"+
				" delete all terminating instance", ffm.count)
			go ffm.killAllTerminalInstance()
			event.Done()
			return
		}
		if len(ffm.terminalCache) == 0 {
			ffm.RUnlock()
			log.GetLogger().Infof("no frontend instance need to terminate, pull up missing instance count:%d",
				ffm.count-len(ffm.instanceCache))
			err := ffm.CreateExpectedInstanceCount(ctx)
			if err != nil {
				event.Error = err
			}
			event.Done()
			return
		}
		insID := ""
		for _, ins := range ffm.terminalCache {
			insID = ins.InstanceID
			break
		}
		ffm.RUnlock()
		log.GetLogger().Infof("start to terminate instance:%s", insID)
		var err error
		if err = ffm.sdkClient.Kill(insID, types.SyncKillSignalVal, []byte{}); err != nil {
			log.GetLogger().Errorf("failed to kill frontend instance(id=%s), err:%v", insID, err)
		}
		ffm.Lock()
		delete(ffm.terminalCache, insID)
		ffm.Unlock()
		time.Sleep(constant.RecreateSleepTime)
		err = ffm.SyncCreateInstance(ctx)
		if err != nil {
			event.Error = err
			event.Done()
			return
		}
	}
}

func (ffm *FrontendManager) killAllTerminalInstance() {
	ffm.Lock()
	insMap := ffm.terminalCache
	ffm.terminalCache = map[string]*types.InstanceSpecification{}
	ffm.Unlock()
	for _, ins := range insMap {
		insID := ins.InstanceID
		go func() {
			err := ffm.KillInstance(insID)
			if err != nil {
				log.GetLogger().Errorf("Failed to kill instance %v, err: %v", insID, err)
			}
		}()
	}
}

func (ffm *FrontendManager) configChangeProcessor(ctx context.Context, cancel context.CancelFunc) {
	if ctx == nil || cancel == nil || ffm.ConfigChangeCh == nil {
		return
	}
	for {
		select {
		case cfgEvent, ok := <-ffm.ConfigChangeCh:
			if !ok {
				cancel()
				return
			}
			frontendConfig, frontendInsNum := ffm.ConfigDiff(cfgEvent)
			if frontendConfig != nil || frontendInsNum != -1 {
				log.GetLogger().Infof("frontend config or instance num is changed," +
					" need to update frontend instance")
				cancel()
				ctx, cancel = context.WithCancel(context.Background())
				config.UpdateFrontendConfig(cfgEvent.FrontendCfg)
				cfgEvent.Add(1)
				go ffm.RollingUpdate(ctx, cfgEvent)
			} else {
				log.GetLogger().Infof("frontend config is same as current, no need to update")
			}
			cfgEvent.Done()
		}
	}
}

// ConfigDiff config diff
func (ffm *FrontendManager) ConfigDiff(event *types.ConfigChangeEvent) (*types.FrontendConfig, int) {
	newSign := controllerutils.GetFrontendConfigSignature(event.FrontendCfg)
	config.FrontendConfigLock.RLock()
	frontendOldCfg := config.GetFaaSFrontendConfig()
	config.FrontendConfigLock.RUnlock()
	if strings.Compare(newSign,
		controllerutils.GetFrontendConfigSignature(frontendOldCfg)) == 0 {
		if event.FrontendCfg.InstanceNum != frontendOldCfg.InstanceNum {
			config.FrontendConfigLock.Lock()
			frontendOldCfg.InstanceNum = event.FrontendCfg.InstanceNum
			config.FrontendConfigLock.Unlock()
			return nil, event.FrontendCfg.InstanceNum
		}
		return nil, -1
	}
	return event.FrontendCfg, event.FrontendCfg.InstanceNum
}
