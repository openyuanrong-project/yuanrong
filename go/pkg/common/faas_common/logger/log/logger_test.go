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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	uberZap "go.uber.org/zap"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/config"
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
