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

package log

import (
	"errors"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	uberZap "go.uber.org/zap"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/config"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/zap"
)

func TestSetupLoggerRuntime(t *testing.T) {
	SetupLoggerLibruntime(nil)
	assert.Equal(t, formatLogger, nil)
}

func TestInitLogger(t *testing.T) {
	errCtrl := ""
	patch := gomonkey.ApplyFunc(config.GetCoreInfoFromEnv, func() (config.CoreInfo, error) {
		if errCtrl == "returnError" {
			return config.CoreInfo{}, errors.New("some error")
		}
		return config.CoreInfo{}, nil
	})
	defer patch.Reset()
	SetupLogger(nil)
	SetupLogger(NewConsoleLogger())
	assert.NotNil(t, formatLogger)
	errCtrl = "returnError"
	err := InitRunLog("test", false)
	assert.NotNil(t, err)
	errCtrl = ""
	err = InitRunLog("test", false)
	assert.Nil(t, err)
}

func TestGetLogger(t *testing.T) {
	convey.Convey("log", t, func() {
		logger := GetLogger()
		logger.With(uberZap.Any("name", "test-log"))
		logger.Info("info log")
		logger.Infof("info log")
		logger.Debug("debug log")
		logger.Debugf("debug log")
		logger.Warn("warn log")
		logger.Warnf("warn log")
		logger.Error("error log")
		logger.Errorf("error log")
	})
}

func TestFormatLogger(t *testing.T) {
	convey.Convey("new log error", t, func() {
		patch := gomonkey.ApplyFunc(zap.NewWithLevel, func(coreInfo config.CoreInfo, isAsync bool) (*uberZap.Logger, error) {
			return nil, errors.New("1")
		})
		defer patch.Reset()
		_, err := NewFormatLogger(constant.MonitorFileName, true, config.CoreInfo{})
		assert.NotNil(t, err)
	})
	convey.Convey("new log success", t, func() {
		filePath := os.Getenv("WORKSPACE")
		logger, err := NewFormatLogger("test", true, config.CoreInfo{
			FilePath: filePath,
		})
		assert.Nil(t, err)
		logger.With(uberZap.Any("name", "test-log"))
		logger.Info("info log")
		logger.Infof("info log")
		logger.Debug("debug log")
		logger.Debugf("debug log")
		//logger.Fatal("fatal log")
		//logger.Fatalf("fatal log")
		logger.Warn("warn log")
		logger.Warnf("warn log")
		logger.Error("error log")
		logger.Errorf("error log")
		logger.Sync()
	})
}
