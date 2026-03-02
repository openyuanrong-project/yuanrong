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

// Package instancepool -
package instancepool

import (
	"encoding/json"

	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

const (
	// DefaultInstanceLabel is empty
	DefaultInstanceLabel = ""
	// DefaultDataVolumeName default dataVolume name
	DefaultDataVolumeName = "data-volume"

	// RaspDefaultInitialDelaySeconds -
	RaspDefaultInitialDelaySeconds = 10
	// RaspDefaultTimeoutSeconds -
	RaspDefaultTimeoutSeconds = 5
	// RaspDefaultPeriodSeconds -
	RaspDefaultPeriodSeconds = 20
	// RaspDefaultSuccessThreshold -
	RaspDefaultSuccessThreshold = 1
	// RaspDefaultFailureThreshold -
	RaspDefaultFailureThreshold = 3
	// RaspDefaultCPU -
	RaspDefaultCPU = 300
	// RaspDefaultMemory -
	RaspDefaultMemory = 500
	// RaspInitDefaultCPU -
	RaspInitDefaultCPU = 100
	// RaspInitDefaultMemory -
	RaspInitDefaultMemory = 100
)

func sideCarAdd(funcSpec *types.FunctionSpecification) ([]byte, error) {
	var sideCars []types.DelegateContainerSideCarConfig

	if utils.IsNeedRaspSideCar(funcSpec) {
		sideCars = append(sideCars, makeRaspContainer(funcSpec))
	}

	configData, err := json.Marshal(sideCars)
	if err != nil {
		return nil, err
	}
	return configData, nil
}

func initContainerAdd(funcSpec *types.FunctionSpecification) ([]byte, error) {
	var initContainers []types.DelegateInitContainerConfig
	if utils.IsNeedRaspSideCar(funcSpec) {
		initContainers = []types.DelegateInitContainerConfig{makeRaspInitContainer(funcSpec)}
	}
	if funcSpec.ExtendedMetaData.UserOtelConfig.Enable {
		initContainers = []types.DelegateInitContainerConfig{makeOtelInitContainer(funcSpec)}
	}
	configData, err := json.Marshal(initContainers)
	if err != nil {
		return nil, err
	}
	return configData, nil
}
