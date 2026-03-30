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

// Package sts provide methods for obtaining sensitive information
package sts

import (
	"fmt"

	"yuanrong.org/kernel/pkg/common/faas_common/sts/raw"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

const faasExecutorStsCertPath = "/opt/certs/%s/%s/%s.ini"

// GetStsServerConfig -
func GetStsServerConfig(funcSpec *types.FunctionSpecification) raw.ServerConfig {
	if !funcSpec.StsMetaData.EnableSts {
		return raw.ServerConfig{}
	}
	domain := config.GlobalConfig.RawStsConfig.ServerConfig.Domain
	if config.GlobalConfig.RawStsConfig.StsConfig.StsDomainForRuntime != "" {
		domain = config.GlobalConfig.RawStsConfig.StsDomainForRuntime
	}
	faasExecutorStsServerConfig := raw.ServerConfig{
		Domain: domain,
		Path: fmt.Sprintf(faasExecutorStsCertPath, funcSpec.StsMetaData.ServiceName,
			funcSpec.StsMetaData.MicroService, funcSpec.StsMetaData.MicroService),
	}
	return faasExecutorStsServerConfig
}
