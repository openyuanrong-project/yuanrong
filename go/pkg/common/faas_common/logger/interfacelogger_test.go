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

package logger

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/config"
)

func TestInterfaceLogger(t *testing.T) {
	testDefaultCoreInfo := config.CoreInfo{
		FilePath:   os.Getenv("WORKSPACE"),
		Level:      "INFO",
		Tick:       0, // Unit: Second
		First:      0, // Unit: Number of logs
		Thereafter: 0, // Unit: Number of logs
		SingleSize: 100,
		Threshold:  10,
		Tracing:    false, // tracing log switch
		Disable:    false, // Disable file logger
	}
	coreInfoBytes, err := json.Marshal(testDefaultCoreInfo)
	assert.Empty(t, err)

	logConfig := os.Getenv("LOG_CONFIG")
	defer func() {
		os.Setenv("LOG_CONFIG", logConfig)
	}()
	err = os.Setenv("LOG_CONFIG", string(coreInfoBytes))
	assert.Empty(t, err)

	cfg := InterfaceEncoderConfig{ModuleName: "WorkerManager"}
	interfaceLog, err := NewInterfaceLogger("", "worker-manager-interface", cfg)
	interfaceLog.Write("123")
	assert.Empty(t, err)
	assert.NotEmpty(t, interfaceLog)
}

func TestCreateSink(t *testing.T) {
	convey.Convey("Test Create Sink Error", t, func() {
		coreInfo := config.CoreInfo{}
		w, err := CreateSink(coreInfo)
		convey.So(w, convey.ShouldBeNil)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("Test Create Sink Error 2", t, func() {
		patch := gomonkey.ApplyFunc(os.MkdirAll, func(path string, perm os.FileMode) error {
			return errors.New("err")
		})
		defer patch.Reset()
		coreInfo := config.CoreInfo{}
		w, err := CreateSink(coreInfo)
		convey.So(w, convey.ShouldBeNil)
		convey.So(err, convey.ShouldNotBeNil)
	})

}
