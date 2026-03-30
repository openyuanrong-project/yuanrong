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

// Package sts used for init sts
package sts

import (
	"os"
	"time"

	"yuanrong.org/kernel/pkg/common/faas_common/alarm"
	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/config"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/sts/raw"
)

// EnvSTSEnable flag
const EnvSTSEnable = "STS_ENABLE"
const fileMode = 0640

// InitStsSDK - Configure sts go sdk
func InitStsSDK(serverCfg raw.ServerConfig) error {
	initStsSdkLog()
	var err error
	if err != nil {
		reportStsAlarm(err.Error())
	}
	return err
}

func reportStsAlarm(errMsg string) {
	alarmDetail := &alarm.Detail{
		SourceTag: os.Getenv(constant.PodNameEnvKey) + "|" + os.Getenv(constant.PodIPEnvKey) +
			"|" + os.Getenv(constant.ClusterName),
		OpType:         alarm.GenerateAlarmLog,
		Details:        "Init sts err, " + errMsg,
		StartTimestamp: int(time.Now().Unix()),
		EndTimestamp:   0,
	}
	alarmInfo := &alarm.LogAlarmInfo{
		AlarmID:    alarm.InitStsSdkErr00001,
		AlarmName:  "InitStsSdkErr",
		AlarmLevel: alarm.Level3,
	}

	alarm.ReportOrClearAlarm(alarmInfo, alarmDetail)
}

func initStsSdkLog() {
	coreInfo, err := config.GetCoreInfoFromEnv()
	if err != nil {
		coreInfo = config.GetDefaultCoreInfo()
	}
	stsSdkLogFilePath := coreInfo.FilePath + "/sts.sdk.log"
	file, err := os.OpenFile(stsSdkLogFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
	if err != nil {
		log.GetLogger().Errorf("failed to open stsSdkLogFile")
		return
	}
	defer file.Close()
	return
}

// DecryptSystemAuthConfig -
func DecryptSystemAuthConfig(auth raw.Auth) raw.Auth {
	systemAuth := raw.Auth{}
	return systemAuth
}
