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

// Package common for tools
package common

import (
	"encoding/json"
	"os"

	"yuanrong.org/kernel/runtime/libruntime/common/faas/logger"
)

func loadStsConfig(configPath string) {
	if configPath == "" {
		return
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		logger.GetLogger().Warnf("read config failed, err %s", err.Error())
		return
	}
	c := &GlobalConfig{}
	err = json.Unmarshal(data, c)
	if err != nil {
		logger.GetLogger().Warnf("unmarshal config failed, err %s", err.Error())
		return
	}
	if !c.RawStsConfig.StsEnable {
		return
	}
}

func loadSensitiveConfig(c *GlobalConfig) {
	return
}
