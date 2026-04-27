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

// Package config of plugin faas executor
package config

import (
	"errors"
	"os"
	"sync"

	"yuanrong.org/kernel/runtime/faassdk/common/constants"
	"yuanrong.org/kernel/runtime/faassdk/types"
	"yuanrong.org/kernel/runtime/faassdk/utils/urnutils"
	"yuanrong.org/kernel/runtime/libruntime/common/faas/logger"
)

// LogTankService -
type LogTankService struct {
	GroupID  string `json:"logGroupId" valid:",optional"`
	StreamID string `json:"logStreamId" valid:",optional"`
}

var (
	configSingleton = struct {
		sync.Once
		config *Configuration
	}{}
)

// DelegateDownloadPath contains downloaded user code packages
const DelegateDownloadPath string = "ENV_DELEGATE_DOWNLOAD"

// DelegateLayerDownloadPath contains downloaded layer packages
const DelegateLayerDownloadPath string = "DELEGATE_LAYER_DOWNLOAD "

// MaxPayloadSize 6 MB
const MaxPayloadSize int32 = 6 * 1024 * 1024

// MaxReturnSize 6MB
const MaxReturnSize int32 = 6 * 1024 * 1024

// DefaultRuntimeJsonFilePath -
const DefaultRuntimeJsonFilePath = "/home/snuser/config/runtime.json"

// Configuration represents total config for app
type Configuration struct {
	StartArgs             `yaml:"commandLine" valid:"optional"`
	RuntimeConfig         `yaml:"runtime" valid:"required"`
	UserLogTag            string `yaml:"userLogTag" valid:"optional"`
	FunctionNameSeparator string `yaml:"functionNameSeparator" valid:"optional"`
}

// StartArgs represents arguments of start
type StartArgs struct {
	LogDir string `yaml:"logDir" valid:"optional"`
}

// RuntimeConfig represents configuration of runtime configuration
type RuntimeConfig struct {
	URN                string         `yaml:"urn" valid:"optional"`
	RuntimeContainerID string         `json:"runtimeContainerID"`
	FuncLogLevel       string         `yaml:"loglevel" valid:"optional"`
	LogTankService     LogTankService `json:"logTankService"`
}

// GetConfig singleton pattern thread safe
func GetConfig(funcSpec *types.FuncSpec) (*Configuration, error) {
	configSingleton.Do(func() {
		runtimeContainerID := os.Getenv(constants.RuntimeContainerIDEnvKey)
		logger.GetLogger().Infof("runtime container id is: %s", runtimeContainerID)
		configSingleton.config = &Configuration{
			StartArgs: StartArgs{
				LogDir: "/home/snuser/log",
			},
			RuntimeConfig: RuntimeConfig{
				URN:                funcSpec.FuncMetaData.FunctionVersionURN,
				FuncLogLevel:       constants.FuncLogLevelInfo,
				RuntimeContainerID: runtimeContainerID,
				LogTankService: LogTankService{
					GroupID:  funcSpec.ExtendedMetaData.LogTankService.GroupID,
					StreamID: funcSpec.ExtendedMetaData.LogTankService.StreamID,
				}},
			UserLogTag:            "cff-log",
			FunctionNameSeparator: "@",
		}
		logger.GetLogger().Infof("config is: %v", configSingleton.config)
		urnutils.SetSeparator(configSingleton.config.FunctionNameSeparator)
		err := urnutils.LocalFuncURN.ParseFrom(configSingleton.config.RuntimeConfig.URN)
		if err != nil {
			logger.GetLogger().Errorf("failed to ParseFrom urn err is : %v", err.Error())
		}
	})

	if configSingleton.config == nil {
		return nil, errors.New("failed to get worker config")
	}
	return configSingleton.config, nil
}
