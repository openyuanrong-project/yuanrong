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

// Package yr for actor
package yr

import (
	"fmt"
	"strings"

	"yuanrong.org/kernel/runtime/libruntime/api"
	"yuanrong.org/kernel/runtime/libruntime/common/uuid"
	"yuanrong.org/kernel/runtime/libruntime/config"
)

const defaultStackTracesNum = 10

// ConfigManager config manager
type ConfigManager struct {
	config.Config
	Mode Mode
}

var manager *ConfigManager = nil

// ClientInfo client info
type ClientInfo struct {
	jobId   string
	version string
}

// GetClientInfo return ClientInfo
func (m *ConfigManager) GetClientInfo() *ClientInfo {
	return &ClientInfo{m.JobID, ""}
}

// GetConfigManager return the ConfigManager
func GetConfigManager() *ConfigManager {
	if manager != nil {
		return manager
	}

	manager = &ConfigManager{}
	manager.JobID = ""
	manager.RuntimeID = "driver"
	manager.IsDriver = true
	manager.LoadPaths = make([]string, 0)
	manager.InCluster = true
	manager.FunctionSystemAddress = ""
	manager.DataSystemIPAddr = ""
	manager.DataSystemPort = 0
	manager.LogDir = ""
	manager.LogLevel = "DEBUG"
	manager.ThreadPoolSize = 10
	manager.EnableMTLS = false
	manager.ServerName = ""
	manager.Namespace = ""
	manager.EnableMetrics = false
	manager.MaxConcurrencyCreateNum = 100
	manager.HttpIocThreadsNum = 200
	manager.RecycleTime = 2
	manager.MaxTaskInstanceNum = -1
	manager.LogFileNumMax = 20
	manager.LogFileSizeMax = 400
	manager.EnableEvent = false
	return manager
}

// Init initializes the ConfigManager
func (m *ConfigManager) Init(yrConfig *Config, yrFlagsConfig *FlagsConfig) error {
	m.IsDriver = !yrConfig.AutoStart

	m.FunctionSystemAddress = yrConfig.ServerAddr
	m.DataSystemAddress = yrConfig.DataSystemAddr
	m.LogLevel = yrConfig.LogLevel
	m.GrpcAddress = yrFlagsConfig.Address
	m.InCluster = yrConfig.InCluster
	m.LogDir = yrFlagsConfig.LogPath
	m.Api = api.ActorApi
	m.Hooks = config.HookIntfs{
		LoadFunctionCb:      yrFlagsConfig.Hooks.LoadFunctionCb,
		FunctionExecutionCb: yrFlagsConfig.Hooks.FunctionExecutionCb,
		CheckpointCb:        yrFlagsConfig.Hooks.CheckpointCb,
		RecoverCb:           yrFlagsConfig.Hooks.RecoverCb,
		ShutdownCb:          yrFlagsConfig.Hooks.ShutdownCb,
		SignalCb:            yrFlagsConfig.Hooks.SignalCb,
	}
	m.FunctionExectionPool = yrFlagsConfig.FunctionExectionPool
	if yrConfig.Mode != Invalid {
		m.Mode = yrConfig.Mode
	}

	if yrConfig.Mode == ClusterMode && (m.IsDriver || len(yrConfig.FunctionUrn) != 0) {
		var err error
		m.FunctionId, err = ConvertFunctionUrn2Id(yrConfig.FunctionUrn)
		if err != nil {
			return err
		}
	}

	if !m.IsDriver && len(yrFlagsConfig.RuntimeID) != 0 {
		m.RuntimeID = yrFlagsConfig.RuntimeID
	}

	if m.IsDriver {
		id := uuid.New()
		m.JobID = fmt.Sprintf("job-%s", strings.ReplaceAll(id.String(), "-", "")[:8])
	} else {
		m.JobID = yrFlagsConfig.JobID
	}
	callStackLayerNum, enableCallStack := getCallStackTracesMgr()
	m.EnableCallStack = enableCallStack
	m.CallStackLayerNum = callStackLayerNum
	return nil
}

// ConvertFunctionUrn2Id convert a functionUrn to id
func ConvertFunctionUrn2Id(functionUrn string) (string, error) {
	functionUrns := strings.Split(functionUrn, ":")
	if len(functionUrns) != 7 {
		return "", fmt.Errorf("functionUrn is not valid")
	}

	return fmt.Sprintf("%s/%s/%s", functionUrns[3], functionUrns[5], functionUrns[6]), nil
}
