/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
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

// Package utils -
package utils

import (
	"net"
	"strconv"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

// IsNeedRaspSideCar -
func IsNeedRaspSideCar(funcSpec *types.FunctionSpecification) bool {
	if funcSpec.ExtendedMetaData.RaspConfig.RaspImage == "" || funcSpec.ExtendedMetaData.RaspConfig.InitImage == "" {
		return false
	}
	if net.ParseIP(funcSpec.ExtendedMetaData.RaspConfig.RaspServerIP) == nil {
		log.GetLogger().Warnf("failed to parse rasp ip: %s ", funcSpec.ExtendedMetaData.RaspConfig.RaspServerIP)
		return false
	}
	if !isValidPort(funcSpec.ExtendedMetaData.RaspConfig.RaspServerPort) {
		log.GetLogger().Warnf("failed to parse rasp "+
			"port: %s ", funcSpec.ExtendedMetaData.RaspConfig.RaspServerPort)
		return false
	}
	return true
}

func isValidPort(port string) bool {
	p, err := strconv.Atoi(port)
	if err != nil {
		return false
	}
	return p > 0 && p <= 65535 // port should between 0 and 65535
}
