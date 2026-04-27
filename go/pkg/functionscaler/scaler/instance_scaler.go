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

// Package scaler -
package scaler

import (
	"time"

	"yuanrong.org/kernel/pkg/common/faas_common/instanceconfig"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

// ScaleUpCallback executes some logic after scale up
type ScaleUpCallback func(int)

// ScaleDownCallback executes some logic after scale down
type ScaleDownCallback func(int)

// CheckReqNumFunc returns current request number of instance thread
type CheckReqNumFunc func() int

// ScaleUpHandler handles instance scale up
type ScaleUpHandler func(int, ScaleUpCallback)

// ScaleDownHandler handles instance scale down
type ScaleDownHandler func(int, ScaleDownCallback)

// InstanceScaler scales instance to meet certain need
type InstanceScaler interface {
	SetEnable(enable bool)
	TriggerScale()
	CheckScaling() bool
	UpdateCreateMetrics(coldStartTime time.Duration)
	HandleInsThdUpdate(inUseInsThdDiff, totalInsThdDiff int)
	HandleFuncSpecUpdate(funcSpec *types.FunctionSpecification)
	HandleInsConfigUpdate(insConfig *instanceconfig.Configuration)
	HandleCreateError(createError error)
	GetExpectInstanceNumber() int
	Destroy()
}
