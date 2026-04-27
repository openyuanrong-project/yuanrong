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

// Package raw define the sts structure
package raw

// StsConfig -
type StsConfig struct {
	StsEnable           bool             `json:"stsEnable,omitempty"`
	SensitiveConfigs    SensitiveConfigs `json:"sensitiveConfigs,omitempty"`
	ServerConfig        ServerConfig     `json:"serverConfig,omitempty"`
	MgmtServerConfig    MgmtServerConfig `json:"mgmtServerConfig"`
	StsDomainForRuntime string           `json:"stsDomainForRuntime"`
}

// SensitiveConfigs -
type SensitiveConfigs struct {
	ShareKeys map[string]string `json:"shareKeys"`
	Auth      Auth              `json:"auth"`
}

// ServerConfig -
type ServerConfig struct {
	Domain string `json:"domain,omitempty" validate:"max=255"`
	Path   string `json:"path,omitempty" validate:"max=255"`
}

// MgmtServerConfig -
type MgmtServerConfig struct {
	Domain string `json:"domain,omitempty"`
}

// Auth -
type Auth struct {
	EnableIam string `json:"enableIam"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	DataKey   string `json:"dataKey"`
}
