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

// Package selfregister contains service route logic
package selfregister

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/rollout"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

const (
	// CurrentVersionEnvKey envKey for current version
	CurrentVersionEnvKey = "CURRENT_VERSION"
	defaultClusterID     = "defaultCluster"
	defaultNodeID        = "defaultNode"
	validRolloutKeyLen   = 7
)

var (
	// IsRollingOut -
	IsRollingOut bool
	// IsRolloutObject -
	IsRolloutObject bool
	// RolloutSubjectID -
	RolloutSubjectID string
	// RolloutRegisterKey -
	RolloutRegisterKey string
	rolloutRegister    *etcd3.EtcdRegister
	rolloutLocker      *etcd3.EtcdLocker
)

// RegisterRolloutToEtcd -
func RegisterRolloutToEtcd(stopCh <-chan struct{}) error {
	log.GetLogger().Infof("start to register rollout key in etcd, rollout enable %t", config.GlobalConfig.EnableRollout)
	if !config.GlobalConfig.EnableRollout || IsRolloutObject {
		log.GetLogger().Infof("skip register rollout key in etcd, isRolloutObject %t", IsRolloutObject)
		return nil
	}
	key, err := getSelfRolloutKey()
	if err != nil {
		return err
	}
	rolloutRegister = &etcd3.EtcdRegister{
		EtcdClient:  etcd3.GetRouterEtcdClient(),
		InstanceKey: key,
		StopCh:      stopCh,
	}
	if err = rolloutRegister.Register(); err != nil {
		log.GetLogger().Errorf("failed to register to etcd, register failed error %s", err.Error())
		return err
	}
	log.GetLogger().Infof("succeed to register rollout key %s in etcd", key)
	return nil
}

// ContendRolloutInEtcd -
func ContendRolloutInEtcd(stopCh <-chan struct{}) error {
	log.GetLogger().Infof("start to contend for rollout key in etcd, rollout enable %t",
		config.GlobalConfig.EnableRollout)
	rolloutLocker = &etcd3.EtcdLocker{
		EtcdClient:     etcd3.GetRouterEtcdClient(),
		LeaseTTL:       defaultLeaseTTL,
		StopCh:         stopCh,
		LockCallback:   putInsSpecForRolloutKey,
		UnlockCallback: delInsSpecForRolloutKey,
		FailCallback:   unsetRolloutRegister,
	}
	var err error
	for i := 0; i < maxContendTime; i++ {
		err = rolloutLocker.TryLockWithPrefix(constant.SchedulerRolloutPrefix, contendFilterForRollout)
		if err != nil {
			log.GetLogger().Errorf("failed to contend for rollout key, lock failed error %s", err.Error())
			time.Sleep(contendWaitInterval)
			continue
		}
		break
	}
	if err != nil {
		log.GetLogger().Errorf("contend retry time reaches max value %d, lock failed error %s", maxContendTime,
			err.Error())
		return err
	}
	log.GetLogger().Infof("succeed to contend for rollout key in etcd, rollout key is %s",
		rolloutLocker.GetLockedKey())
	return nil
}

// ReplaceRolloutSubject -
func ReplaceRolloutSubject(stopCh <-chan struct{}) {
	if rollout.GetGlobalRolloutHandler().GetCurrentRatio() != 100 { // 100 is finish rollout
		log.GetLogger().Infof("current ratio %d is still not 100, skip ReplaceRolloutSubject",
			rollout.GetGlobalRolloutHandler().GetCurrentRatio())
		return
	}
	log.GetLogger().Infof("start to replace rollout subject %s of instance name %s", RolloutSubjectID,
		SelfInstanceName)
	if rolloutLocker == nil || selfLocker == nil {
		log.GetLogger().Errorf("failed to replace rollout subject, rolloutLocker or selfLocker is nil")
		return
	}
	lockedRolloutKey := rolloutLocker.GetLockedKey()
	if err := rolloutLocker.Unlock(); err != nil {
		log.GetLogger().Errorf("failed to unlock rollout key %s , unlock error %s", lockedRolloutKey, err.Error())
	}
	ctx := etcd3.CreateEtcdCtxInfoWithTimeout(context.Background(), etcd3.DurationContextTimeout)
	if err := rolloutLocker.EtcdClient.Delete(ctx, lockedRolloutKey); err != nil {
		log.GetLogger().Errorf("failed to delete rollout key %s, delete error %s", lockedRolloutKey, err.Error())
	}
	if err := selfLocker.TryLock(RolloutRegisterKey); err != nil {
		log.GetLogger().Warnf("failed to lock rollout register key %s, lock error %s, try lock another key",
			RolloutRegisterKey, err.Error())
		err = contendInstanceInEtcd(stopCh)
		if err != nil {
			log.GetLogger().Errorf("failed to lock register key %s, lock error %s", RolloutRegisterKey, err.Error())
		}
		return
	}
	if err := processLockedInstanceKey(selfLocker.GetLockedKey()); err != nil {
		log.GetLogger().Errorf("failed to process lock key %s, process error %s", RolloutRegisterKey, err.Error())
		return
	}
	IsRollingOut = false
	IsRolloutObject = false
	if err := RegisterRolloutToEtcd(stopCh); err != nil {
		log.GetLogger().Errorf("failed to replace to etcd for rollout, register failed error %s", err.Error())
	}
	log.GetLogger().Infof("succeed to replace rollout subject %s of instance name %s", RolloutSubjectID,
		SelfInstanceName)
}

func getSelfRolloutKey() (string, error) {
	clusterID := config.GlobalConfig.ClusterID
	if len(clusterID) == 0 {
		clusterID = defaultClusterID
	}
	instanceID := SelfInstanceID
	if len(instanceID) == 0 {
		return "", errors.New("self instanceID is empty")
	}
	return fmt.Sprintf("%s/%s/%s/%s", constant.SchedulerRolloutPrefix, clusterID, defaultNodeID, instanceID), nil
}

func contendFilterForRollout(key, value []byte) bool {
	items := strings.Split(string(key), "/")
	if len(items) != validRolloutKeyLen {
		return true
	}
	return false
}

func putInsSpecForRolloutKey(locker *etcd3.EtcdLocker) error {
	lockedKey := locker.GetLockedKey()
	log.GetLogger().Infof("start to put insSpec for rollout key %s", lockedKey)
	if len(lockedKey) == 0 {
		log.GetLogger().Errorf("failed to get locked key")
		return errors.New("locked key is empty")
	}
	instanceID := SelfInstanceID
	if len(instanceID) == 0 {
		log.GetLogger().Errorf("failed to get self instance key")
		return errors.New("self instance key is empty")
	}
	selfSpec := getSelfInstanceSpec()
	if selfSpec == nil {
		log.GetLogger().Errorf("failed to get insSpec of this scheduler %s", instanceID)
		return errors.New("insSpec not found")
	}
	if err := processRolloutRequest(lockedKey); err != nil {
		log.GetLogger().Errorf("failed to process rollout request error %s", err.Error())
		return err
	}
	rolloutInsSpec := types.RolloutInstanceSpecification{
		RegisterKey:    RolloutRegisterKey,
		InstanceID:     selfSpec.InstanceID,
		RuntimeAddress: selfSpec.RuntimeAddress,
	}
	rolloutInsSpecData, err := json.Marshal(rolloutInsSpec)
	if err != nil {
		log.GetLogger().Errorf("failed to marshal insSpec error %s", err.Error())
		return err
	}
	if err = processEtcdPut(locker.EtcdClient, lockedKey, string(rolloutInsSpecData)); err != nil {
		log.GetLogger().Errorf("failed to put insSpec for rollout key into etcd error %s", err.Error())
		return err
	}
	log.GetLogger().Infof("succeed to put insSpec for rollout key %s", lockedKey)
	return nil
}

func delInsSpecForRolloutKey(locker *etcd3.EtcdLocker) error {
	lockedKey := locker.GetLockedKey()
	log.GetLogger().Infof("start to clean insSpec for rollout key %s", lockedKey)
	if len(lockedKey) == 0 {
		log.GetLogger().Errorf("failed to get locked key")
		return errors.New("locked key is empty")
	}
	if exist, err := isKeyExist(locker.EtcdClient, lockedKey); err != nil || !exist {
		return fmt.Errorf("key not exist or get error %v, no need clean it", err)
	}
	if err := processEtcdPut(locker.EtcdClient, lockedKey, ""); err != nil {
		log.GetLogger().Errorf("failed to clean insSpec for rollout key in etcd error %s", err.Error())
		return err
	}
	log.GetLogger().Infof("succeed to clean insSpec for rollout key %s", lockedKey)
	return nil
}

func unsetRolloutRegister() {
	log.GetLogger().Warnf("locker of rollout key %s failed, unset register status", RolloutRegisterKey)
	IsRolloutObject = false
	RolloutRegisterKey = ""
}

func processRolloutRequest(rolloutKey string) error {
	log.GetLogger().Infof("start to process rollout request, rollout key %s", rolloutKey)
	items := strings.Split(rolloutKey, "/")
	if len(items) != validRolloutKeyLen {
		log.GetLogger().Errorf("failed to parse rollout scheduler from key %s", rolloutKey)
		return errors.New("invalid rollout key")
	}
	RolloutSubjectID = items[validRolloutKeyLen-1]
	log.GetLogger().Infof("set RolloutSubjectID to %s", RolloutSubjectID)
	rsp, err := rollout.GetGlobalRolloutHandler().SendRolloutRequest(SelfInstanceID, RolloutSubjectID)
	if err != nil {
		log.GetLogger().Errorf("failed to send rollout request to instance %s error %s", RolloutSubjectID, err.Error())
		return err
	}
	RolloutRegisterKey = rsp.RegisterKey
	log.GetLogger().Infof("succeed to set RolloutRegisterKey to %s", RolloutRegisterKey)
	registerInfo, err := utils.GetSchedulerInfoFromEtcdKey(RolloutRegisterKey)
	if err != nil {
		log.GetLogger().Errorf("failed to get register info from key %s error %s", rsp.RegisterKey, err.Error())
		return err
	}
	SetSelfInstanceName(registerInfo.InstanceName)
	IsRollingOut = true
	IsRolloutObject = true
	log.GetLogger().Infof("succeed to process rollout request, rollout key %s", rolloutKey)
	return nil
}
