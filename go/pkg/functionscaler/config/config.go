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

	"github.com/asaskevich/govalidator/v11"

	"yuanrong.org/kernel/pkg/common/faas_common/alarm"
	"yuanrong.org/kernel/pkg/common/faas_common/crypto"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/localauth"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/sts"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/common/faas_common/wisecloudtool/serviceaccount"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

const (
	// MetaEtcdPwdKey -
	MetaEtcdPwdKey = "metaEtcdPwd"
	// DockerRootPathEnv -
	DockerRootPathEnv               = "DOCKER_ROOT_DIR"
	defaultDockerRootPath           = "/var/lib/docker"
	defaultFaasschedulerSTScertPath = "/opt/certs/HMSClientCloudAccelerateService/" +
		"HMSCaaSYuanRongWorkerManager/HMSCaaSYuanRongWorkerManager.ini"
	defaultPredictGroupWindow      = 15 * 60 * 1000
	defaultDataSystemTimeoutMs     = 60000
	defaultUploadDataSystemTimeout = 10
)

var (
	// GlobalConfig is the global config
	GlobalConfig     types.Configuration
	configEnvKey     = "SCHEDULER_CONFIG"
	dockerRootPrefix = []byte("Docker Root Dir: ")
)

// InitModuleConfig initializes config for module
func InitModuleConfig() error {
	config, err := loadConfigFromEnv()
	if err != nil {
		return err
	}
	GlobalConfig = *config
	log.GetLogger().Infof("succeed to init module config")
	return nil
}

func loadConfigFromEnv() (*types.Configuration, error) {
	configJSON := os.Getenv(configEnvKey)
	config := &types.Configuration{}
	err := json.Unmarshal([]byte(configJSON), config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// InitConfig will initialize global config
func InitConfig(configData []byte) error {
	GlobalConfig = types.Configuration{}
	err := json.Unmarshal(configData, &GlobalConfig)
	if err != nil {
		return err
	}
	return loadFunctionConfig(&GlobalConfig)
}

func loadFunctionConfig(GlobalConfig *types.Configuration) error {
	setETCDConfig(GlobalConfig)
	decryptEnvMap, err := localauth.GetDecryptFromEnv()
	if err != nil {
		log.GetLogger().Errorf("get decrypt from env error: %v", err)
		return err
	}
	setDecryptPwd(decryptEnvMap, GlobalConfig)

	if _, err = govalidator.ValidateStruct(GlobalConfig); err != nil {
		return err
	}
	err = setAlarmEnv()
	if err != nil {
		return err
	}

	if GlobalConfig.DockerRootPath != "" {
		if err = os.Setenv(DockerRootPathEnv, GlobalConfig.DockerRootPath); err != nil {
			log.GetLogger().Warnf("cannot set env DOCKER_ROOT_DIR")
		}
	} else {
		if err = os.Setenv(DockerRootPathEnv, defaultDockerRootPath); err != nil {
			log.GetLogger().Warnf("cannot set default env DOCKER_ROOT_DIR")
		}
	}
	if GlobalConfig.RawStsConfig.StsEnable {
		if err := sts.InitStsSDK(GlobalConfig.RawStsConfig.ServerConfig); err != nil {
			log.GetLogger().Errorf("failed to init sts sdk, err: %s", err.Error())
			return err
		}
		if err = os.Setenv(sts.EnvSTSEnable, "true"); err != nil {
			log.GetLogger().Errorf("failed to set env of %s, err: %s", sts.EnvSTSEnable, err.Error())
			return err
		}
		GlobalConfig.SystemAuthConfig = sts.DecryptSystemAuthConfig(GlobalConfig.RawStsConfig.SensitiveConfigs.Auth)
	}
	if len(GlobalConfig.Scenario) == 0 {
		GlobalConfig.Scenario = types.ScenarioWiseCloud
	}
	if GlobalConfig.PredictGroupWindow == 0 {
		GlobalConfig.PredictGroupWindow = defaultPredictGroupWindow
	}
	if GlobalConfig.SccConfig.Enable && crypto.InitializeSCC(GlobalConfig.SccConfig) != nil {
		return fmt.Errorf("failed to initialize scc")
	}
	err = setServiceAccountJwt(GlobalConfig)
	if err != nil {
		return fmt.Errorf("failed to set serviceaccount jwt config %s", err.Error())
	}
	if GlobalConfig.DataSystemConfig.InitTimeoutMs <= 0 {
		GlobalConfig.DataSystemConfig.InitTimeoutMs = defaultDataSystemTimeoutMs
	}
	if GlobalConfig.DataSystemConfig.UploadTTLSec <= 0 {
		GlobalConfig.DataSystemConfig.UploadTTLSec = defaultUploadDataSystemTimeout
	}
	return nil
}

func setServiceAccountJwt(cfg *types.Configuration) error {
	if cfg.RawStsConfig.StsEnable && len(cfg.ServiceAccountJwt.ServiceAccountKeyStr) > 0 {
		if cfg.ServiceAccountJwt.ServiceAccount, err :=
			serviceaccount.ParseServiceAccount(cfg.ServiceAccountJwt.ServiceAccountKeyStr); err != nil {
			return err
		}
	}
	if cfg.ServiceAccountJwt.TlsConfig != nil &&
		len(cfg.ServiceAccountJwt.TlsConfig.TlsCipherSuitesStr) > 0 {
		var err error
		cfg.ServiceAccountJwt.TlsConfig.TlsCipherSuites, err =
			serviceaccount.ParseTlsCipherSuites(cfg.ServiceAccountJwt.TlsConfig.TlsCipherSuitesStr)
		if err != nil {
			return err
		}
	}
	return nil
}

func setETCDConfig(GlobalConfig *types.Configuration) {
	if GlobalConfig == nil {
		return
	}
	if GlobalConfig.RouterETCDConfig.UseSecret {
		etcd3.SetETCDTLSConfig(&GlobalConfig.RouterETCDConfig)
	}
	if GlobalConfig.MetaETCDConfig.UseSecret {
		etcd3.SetETCDTLSConfig(&GlobalConfig.MetaETCDConfig)
	}
}

func setDecryptPwd(decryptEnvMap map[string]string, config *types.Configuration) {
	_, ok := decryptEnvMap[MetaEtcdPwdKey]
	if !ok {
		return
	}
	if decryptEnvMap[MetaEtcdPwdKey] != "" {
		config.MetaETCDConfig.Password = decryptEnvMap[MetaEtcdPwdKey]
		decryptEnvMap[MetaEtcdPwdKey] = ""
	}
}

// InitEtcd - init router etcd and meta etcd
func InitEtcd(stopCh <-chan struct{}) error {
	if &GlobalConfig == nil {
		return fmt.Errorf("config is not initialized")
	}
	if err := etcd3.InitRouterEtcdClient(GlobalConfig.RouterETCDConfig, GlobalConfig.AlarmConfig, stopCh); err != nil {
		return fmt.Errorf("faaSScheduler failed to init route etcd: %s", err.Error())
	}

	if err := etcd3.InitMetaEtcdClient(GlobalConfig.MetaETCDConfig, GlobalConfig.AlarmConfig, stopCh); err != nil {
		return fmt.Errorf("faaSScheduler failed to init metadata etcd: %s", err.Error())
	}
	if len(GlobalConfig.DataSystemEtcd.Servers) != 0 {
		if err := etcd3.InitDataSystemEtcdClient(GlobalConfig.DataSystemEtcd, GlobalConfig.AlarmConfig, stopCh); err != nil {
			return fmt.Errorf("faaSFrontend failed to init dataSystemEtcd etcd: %s", err.Error())
		}
		log.GetLogger().Infof("init dataSystemEtcd success")
	}

	return nil
}

// ClearSensitiveInfo -
func ClearSensitiveInfo() {
	if &GlobalConfig == nil {
		return
	}
	utils.ClearStringMemory(GlobalConfig.MetaETCDConfig.Password)
	utils.ClearStringMemory(GlobalConfig.RouterETCDConfig.Password)
}

func setAlarmEnv() error {
	if &GlobalConfig == nil || !GlobalConfig.AlarmConfig.EnableAlarm {
		log.GetLogger().Infof("enable alarm is false")
		return nil
	}
	utils.SetClusterNameEnv(GlobalConfig.ClusterName)
	alarm.SetAlarmEnv(GlobalConfig.AlarmConfig.AlarmLogConfig)
	alarm.SetXiangYunFourConfigEnv(GlobalConfig.AlarmConfig.XiangYunFourConfig)
	err := alarm.SetPodIP()
	if err != nil {
		return err
	}
	return nil
}
