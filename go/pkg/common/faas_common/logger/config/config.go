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

// Package config is common logger client
package config

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/asaskevich/govalidator/v11"
	"go.uber.org/zap/zapcore"

	"yuanrong.org/kernel/pkg/common/faas_common/utils"
)

const (
	configPath   = "/home/sn/config/log.json"
	fileMode     = 0750
	logConfigKey = "LOG_CONFIG"
)

var (
	defaultCoreInfo CoreInfo
	// LogLevel -
	LogLevel zapcore.Level = zapcore.InfoLevel
)

func init() {
	defaultFilePath := os.Getenv("GLOG_log_dir")
	if defaultFilePath == "" {
		defaultFilePath = "/home/snuser/log"
	}
	defaultLevel := "INFO"
	// defaultCoreInfo default logger config
	defaultCoreInfo = CoreInfo{
		FilePath:   defaultFilePath,
		Level:      defaultLevel,
		Tick:       0, // Unit: Second
		First:      0, // Unit: Number of logs
		Thereafter: 0, // Unit: Number of logs
		SingleSize: 100,
		Threshold:  10,
		Tracing:    false, // tracing log switch
		Disable:    false, // Disable file logger
	}
}

// CoreInfo contains the core info
type CoreInfo struct {
	FilePath            string `json:"filepath" valid:",optional"`
	Level               string `json:"level" valid:",optional"`
	Tick                int    `json:"tick" valid:"range(0|86400),optional"`
	First               int    `json:"first" valid:"range(0|20000),optional"`
	Thereafter          int    `json:"thereafter" valid:"range(0|1000),optional"`
	Tracing             bool   `json:"tracing" valid:",optional"`
	Disable             bool   `json:"disable" valid:",optional"`
	SingleSize          int64  `json:"singlesize" valid:",optional"`
	Threshold           int    `json:"threshold" valid:",optional"`
	IsUserLog           bool   `json:"-"`
	IsWiseCloudAlarmLog bool   `json:"isWiseCloudAlarmLog" valid:",optional"`
}

// GetDefaultCoreInfo get defaultCoreInfo
func GetDefaultCoreInfo() CoreInfo {
	return defaultCoreInfo
}

// GetCoreInfoFromEnv extracts the logger config and ensures that the log file is available
func GetCoreInfoFromEnv() (CoreInfo, error) {
	coreInfo, err := ExtractCoreInfoFromEnv(logConfigKey)
	if err != nil {
		return defaultCoreInfo, err
	}
	if err = utils.ValidateFilePath(coreInfo.FilePath); err != nil {
		return defaultCoreInfo, err
	}
	if err = os.MkdirAll(coreInfo.FilePath, fileMode); err != nil && !os.IsExist(err) {
		return defaultCoreInfo, err
	}

	return coreInfo, nil
}

// ExtractCoreInfoFromEnv extracts the logger config from ENV
func ExtractCoreInfoFromEnv(env string) (CoreInfo, error) {
	var coreInfo CoreInfo
	conf := os.Getenv(env)
	if conf == "" {
		return defaultCoreInfo, errors.New(env + " is empty")
	}
	err := json.Unmarshal([]byte(conf), &coreInfo)
	if err != nil {
		return defaultCoreInfo, err
	}

	// if the file path is empty, return error
	// if the log file is not writable, zap will create a new file with the configured file path and file name
	if coreInfo.FilePath == "" {
		return defaultCoreInfo, errors.New("the log file path is empty")
	}
	if _, err = govalidator.ValidateStruct(coreInfo); err != nil {
		return defaultCoreInfo, err
	}

	return coreInfo, nil
}
