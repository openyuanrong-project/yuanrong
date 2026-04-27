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

// Package common for config type
package common

// StsConfig -
type StsConfig struct {
	StsEnable        bool             `json:"stsEnable,omitempty"`
	ServerConfig     ServerConfig     `json:"serverConfig,omitempty"`
	SensitiveConfigs SensitiveConfigs `json:"sensitiveConfigs,omitempty"`
}

// SensitiveConfigs -
type SensitiveConfigs struct {
	Auth map[string]string `json:"auth"`
}

// ServerConfig -
type ServerConfig struct {
	Domain string `json:"domain,omitempty" validate:"max=255"`
	Path   string `json:"path,omitempty" validate:"max=255"`
}

// GlobalConfig -
type GlobalConfig struct {
	RawStsConfig StsConfig `json:"rawStsConfig"`
}
