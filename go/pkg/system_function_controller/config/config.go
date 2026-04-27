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

// Package config -
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/asaskevich/govalidator/v11"

	"yuanrong.org/kernel/pkg/common/faas_common/alarm"
	"yuanrong.org/kernel/pkg/common/faas_common/crypto"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/localauth"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/sts"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/system_function_controller/state"
	"yuanrong.org/kernel/pkg/system_function_controller/types"
)

const (
	// MetaEtcdPwdKey -
	MetaEtcdPwdKey = "metaEtcdPwd"
)

var (
	faaSControllerConfig *types.Config
	faaSSchedulerConfig  *types.SchedulerConfig
	faaSFrontendConfig   *types.FrontendConfig
	faaSManagerConfig    *types.ManagerConfig

	// SchedulerConfigLock scheduler config rw lock
	SchedulerConfigLock sync.RWMutex
	// FrontendConfigLock  frontend config rw lock
	FrontendConfigLock sync.RWMutex
	// ManagerConfigLock manager config rw lock
	ManagerConfigLock sync.RWMutex
)

// RecoverConfig will recover config
func RecoverConfig() error {
	stateConf := state.GetState()
	faaSControllerConfig = &types.Config{}
	err := utils.DeepCopyObj(stateConf.FaaSControllerConfig, faaSControllerConfig)
	if err != nil {
		return err
	}
	if err = setFaaSConfigurations(); err != nil {
		return err
	}
	log.GetLogger().Infof("configuration recovered")
	return nil
}

// InitConfig will initialize global config
func InitConfig(configData []byte) error {
	faaSControllerConfig = &types.Config{}
	err := json.Unmarshal(configData, faaSControllerConfig)
	if err != nil {
		log.GetLogger().Errorf("json unmarshal faaS controller config error: %v", err)
		return err
	}

	if err = setFaaSConfigurations(); err != nil {
		return err
	}
	if _, err = govalidator.ValidateStruct(faaSControllerConfig); err != nil {
		return err
	}
	if faaSControllerConfig.RawStsConfig.StsEnable {
		if err := sts.InitStsSDK(faaSControllerConfig.RawStsConfig.ServerConfig); err != nil {
			log.GetLogger().Errorf("failed to init sts sdk, err: %s", err.Error())
			return err
		}
		if err = os.Setenv(sts.EnvSTSEnable, "true"); err != nil {
			log.GetLogger().Errorf("failed to set env of %s, err: %s", sts.EnvSTSEnable, err.Error())
			return err
		}
	}
	if faaSControllerConfig.SccConfig.Enable && crypto.InitializeSCC(faaSControllerConfig.SccConfig) != nil {
		return fmt.Errorf("failed to initialize scc")
	}
	return nil
}

func setFaaSConfigurations() error {
	if faaSControllerConfig == nil {
		return fmt.Errorf("faaSController config is nil")
	}
	if faaSControllerConfig.RouterETCD.UseSecret {
		etcd3.SetETCDTLSConfig(&faaSControllerConfig.RouterETCD)
	} else {
		faaSControllerConfig.RouterETCD.CaFile = faaSControllerConfig.TLSConfig.CaContent
		faaSControllerConfig.RouterETCD.CertFile = faaSControllerConfig.TLSConfig.CertContent
		faaSControllerConfig.RouterETCD.KeyFile = faaSControllerConfig.TLSConfig.KeyContent
	}
	if faaSControllerConfig.MetaETCD.UseSecret {
		etcd3.SetETCDTLSConfig(&faaSControllerConfig.MetaETCD)
	}
	etcdConfig, err := DecryptEtcdConfig(faaSControllerConfig.MetaETCD)
	if err != nil {
		return err
	}
	faaSControllerConfig.MetaETCD = *etcdConfig

	err = setAlarmEnv(faaSControllerConfig)
	if err != nil {
		return err
	}
	return nil
}

func setAlarmEnv(faaSControllerConfig *types.Config) error {
	if faaSControllerConfig == nil || !faaSControllerConfig.AlarmConfig.EnableAlarm {
		log.GetLogger().Infof("enable alarm is false")
		return nil
	}
	utils.SetClusterNameEnv(faaSControllerConfig.ClusterName)
	alarm.SetAlarmEnv(faaSControllerConfig.AlarmConfig.AlarmLogConfig)
	alarm.SetXiangYunFourConfigEnv(faaSControllerConfig.AlarmConfig.XiangYunFourConfig)
	err := alarm.SetPodIP()
	if err != nil {
		return err
	}
	return nil
}

// GetFaaSControllerConfig will get faas controller config
func GetFaaSControllerConfig() types.Config {
	return *faaSControllerConfig
}

// GetFaaSSchedulerConfig will get faas scheduler config
func GetFaaSSchedulerConfig() *types.SchedulerConfig {
	return faaSSchedulerConfig
}

// GetFaaSFrontendConfig will get faas frontend config
func GetFaaSFrontendConfig() *types.FrontendConfig {
	return faaSFrontendConfig
}

// GetFaaSManagerConfig will get faas manager config
func GetFaaSManagerConfig() *types.ManagerConfig {
	return faaSManagerConfig
}

// DecryptEtcdConfig decrypt etcd secret
func DecryptEtcdConfig(config etcd3.EtcdConfig) (*etcd3.EtcdConfig, error) {
	decryptEnvMap, err := localauth.GetDecryptFromEnv()
	if err != nil {
		log.GetLogger().Errorf("get decrypt from env error: %v", err)
		return nil, err
	}
	if decryptEnvMap[MetaEtcdPwdKey] != "" {
		config.Password = decryptEnvMap[MetaEtcdPwdKey]
	}
	return &config, nil
}

// InitEtcd - init router etcd and meta etcd
func InitEtcd(stopCh <-chan struct{}) error {
	if faaSControllerConfig == nil {
		return fmt.Errorf("config is not initialized")
	}
	if err := etcd3.InitRouterEtcdClient(faaSControllerConfig.RouterETCD,
		faaSControllerConfig.AlarmConfig, stopCh); err != nil {
		return fmt.Errorf("faaSController failed to init route etcd: %s", err.Error())
	}

	if err := etcd3.InitMetaEtcdClient(faaSControllerConfig.MetaETCD,
		faaSControllerConfig.AlarmConfig, stopCh); err != nil {
		return fmt.Errorf("faaSController failed to init metadata etcd: %s", err.Error())
	}
	return nil
}

// UpdateSchedulerConfig update scheduler config
func UpdateSchedulerConfig(cfg *types.SchedulerConfig) {
	SchedulerConfigLock.Lock()
	faaSSchedulerConfig = cfg
	SchedulerConfigLock.Unlock()
}

// UpdateFrontendConfig update frontend config
func UpdateFrontendConfig(cfg *types.FrontendConfig) {
	FrontendConfigLock.Lock()
	faaSFrontendConfig = cfg
	FrontendConfigLock.Unlock()
}

// UpdateManagerConfig update manager config
func UpdateManagerConfig(cfg *types.ManagerConfig) {
	ManagerConfigLock.Lock()
	faaSManagerConfig = cfg
	ManagerConfigLock.Unlock()
}
