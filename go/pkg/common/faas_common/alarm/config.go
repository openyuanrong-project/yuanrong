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

// Package alarm alarm log by filebeat
package alarm

import (
	"yuanrong.org/kernel/pkg/common/faas_common/logger/config"
	"yuanrong.org/kernel/pkg/common/faas_common/types"
)

// Config -
type Config struct {
	EnableAlarm         bool                     `json:"enableAlarm"`
	AlarmLogConfig      config.CoreInfo          `json:"alarmLogConfig" valid:"optional"`
	XiangYunFourConfig  types.XiangYunFourConfig `json:"xiangYunFourConfig" valid:"optional"`
	MinInsStartInterval int                      `json:"minInsStartInterval"`
	MinInsCheckInterval int                      `json:"minInsCheckInterval"`
}
