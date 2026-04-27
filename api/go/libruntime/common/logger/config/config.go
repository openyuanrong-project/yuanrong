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
	"sync"

	"github.com/asaskevich/govalidator/v11"
	"go.uber.org/zap/zapcore"

	"yuanrong.org/kernel/runtime/libruntime/common/utils"
)

const (
	configPath      = "/home/sn/config/log.json"
	fileMode        = 0750
	defaultLogLevel = "INFO"
	// LogConfigKey environment variable key of log config
	LogConfigKey = "LOG_CONFIG"
)

var (
	defaultCoreInfo CoreInfo
	onceSetDefault  sync.Once
	// LogLevel -
	LogLevel zapcore.Level = zapcore.InfoLevel
	// LogLevelFromFlag -
	LogLevelFromFlag string
)

func initDefaultCoreInfo() {
	defaultFilePath := os.Getenv("GLOG_log_dir")
	if defaultFilePath == "" {
		defaultFilePath = "/home/snuser/log"
	}
	logLevel := os.Getenv("YR_LOG_LEVEL")
	if logLevel == "" {
		logLevel = defaultLogLevel
	}
	// defaultCoreInfo default logger config
	defaultCoreInfo = CoreInfo{
		FilePath:   defaultFilePath,
		Level:      logLevel,
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
	FilePath   string `json:"filepath" valid:",required"`
	Level      string `json:"level" valid:",required"`
	Tick       int    `json:"tick" valid:"range(0|86400),optional"`
	First      int    `json:"first" valid:"range(0|20000),optional"`
	Thereafter int    `json:"thereafter" valid:"range(0|1000),optional"`
	Tracing    bool   `json:"tracing" valid:",optional"`
	Disable    bool   `json:"disable" valid:",optional"`
	SingleSize int64  `json:"singlesize" valid:",optional"`
	Threshold  int    `json:"threshold" valid:",optional"`
	IsUserLog  bool   `json:"-"`
}

// GetDefaultCoreInfo get defaultCoreInfo
func GetDefaultCoreInfo() CoreInfo {
	onceSetDefault.Do(func() {
		initDefaultCoreInfo()
	})
	return defaultCoreInfo
}

// GetCoreInfoFromEnv extracts the logger config and ensures that the log file is available
func GetCoreInfoFromEnv() (CoreInfo, error) {
	coreInfo, err := ExtractCoreInfoFromEnv(LogConfigKey)
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
	config := os.Getenv(env)
	if config == "" {
		return defaultCoreInfo, errors.New(env + " is empty")
	}
	err := json.Unmarshal([]byte(config), &coreInfo)
	if err != nil {
		return defaultCoreInfo, err
	}

	// if the file path is empty, return error
	// if the log file is not writable, zap will create a new file with the configured file path and file name
	if coreInfo.FilePath == "" {
		return defaultCoreInfo, errors.New("log file path is empty")
	}
	if _, err = govalidator.ValidateStruct(coreInfo); err != nil {
		return defaultCoreInfo, err
	}

	return coreInfo, nil
}
