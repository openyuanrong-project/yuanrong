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
	"strings"
	"testing"
)

func TestConvertFunctionUrn2Id(t *testing.T) {
	// convert correct value
	functionUrn := "sn:cn:yrk:default:function:0-opc-opc:$latest"
	if functionId, err := ConvertFunctionUrn2Id(functionUrn); functionId != "default/0-opc-opc/$latest" || err != nil {
		t.Errorf("Convert %s failed.", functionUrn)
	}

	// convert invalid value
	functionUrn = "sn:cn:yrk:default^function:0-opc-opc"
	if functionId, err := ConvertFunctionUrn2Id(functionUrn); len(functionId) > 0 || err == nil {
		t.Errorf("Convert %s failed.", functionUrn)
	}
}

func TestConfigManagerInit(t *testing.T) {
	yrConfig := &Config{
		Mode:                ClusterMode,
		FunctionUrn:         "sn:cn:yrk:default:function:0-opc-opc:$latest",
		ServerAddr:          "127.0.0.1:31220",
		DataSystemAddr:      "127.0.0.1:31501",
		DataSystemAgentAddr: "",
		InCluster:           true,
		LogLevel:            "INFO",
	}
	configManager := GetConfigManager()
	res := configManager.Init(yrConfig, &FlagsConfig{})
	if res != nil {
		t.Errorf("configManager init failed")
	}

	if configManager.Config.FunctionSystemAddress != "127.0.0.1:31220" {
		t.Errorf("FunctionSystemAddress init failed")
	}

	if configManager.Config.DataSystemAddress != "127.0.0.1:31501" {
		t.Errorf("DataSystemIPAddr init failed")
	}

	if !configManager.Config.IsDriver {
		t.Errorf("IsDriver init failed")
	}

	if !strings.Contains(configManager.Config.JobID, "job-") || len(configManager.Config.JobID) != 12 {
		t.Errorf("JobID init failed")
	}

	if configManager.Config.RuntimeID != "driver" {
		t.Errorf("RuntimeID init failed")
	}

	if configManager.Config.LogLevel != "INFO" {
		t.Errorf("LogLevel init failed")
	}

	if !configManager.Config.InCluster {
		t.Errorf("InCluster init failed")
	}

	if configManager.FunctionId != "default/0-opc-opc/$latest" {
		t.Errorf("FunctionId init failed")
	}

	if configManager.Mode != ClusterMode {
		t.Errorf("Mode init failed")
	}
}
