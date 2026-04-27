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

package zap

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	uberZap "go.uber.org/zap"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/config"
)

// TestNewDvelopmentLog Test New Dvelopment Log
func TestNewDvelopmentLog(t *testing.T) {
	if _, err := NewDevelopmentLog(); err != nil {
		t.Errorf("NewDevelopmentLog() = %q, wants *logger", err)
	}
}

func TestNewConsoleLog(t *testing.T) {
	tests := []struct {
		name    string
		want    *uberZap.Logger
		wantErr bool
	}{
		{"case1", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewConsoleLog()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConsoleLog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestNewWithLevel(t *testing.T) {
	type args struct {
		coreInfo config.CoreInfo
		isAsync  bool
	}
	var a args
	tests := []struct {
		name    string
		args    args
		want    *uberZap.Logger
		wantErr bool
	}{
		{"case1", a, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWithLevel(tt.args.coreInfo, tt.args.isAsync)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWithLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewWithLevel() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoggerWithFormat_Infof(t *testing.T) {
	type fields struct {
		Logger *uberZap.Logger
	}
	type args struct {
		format string
		paras  []interface{}
	}
	coreInfo := config.CoreInfo{
		FilePath:   "tmp",
		Level:      "DEBUG",
		Tick:       0,
		First:      0,
		Thereafter: 0,
		Tracing:    false,
		Disable:    false,
	}
	logger, err := NewWithLevel(coreInfo, true)
	if err != nil {
		fmt.Println(err)
	}
	var f fields
	f.Logger = logger
	var a args
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"case1", f, a},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &LoggerWithFormat{
				Logger: tt.fields.Logger,
			}
			z.Infof(tt.args.format, tt.args.paras...)
		})
	}
}

func TestNewCoreWithDebugLevel(t *testing.T) {
	convey.Convey("TestNewCoreWithInfoLevel", t, func() {
		coreInfo := config.CoreInfo{
			FilePath:   "tmp",
			Level:      "DEBUG",
			Tick:       0,
			First:      0,
			Thereafter: 0,
			Tracing:    false,
			Disable:    false,
		}
		logger, err := NewWithLevel(coreInfo, true)
		convey.So(logger, convey.ShouldNotBeNil)
		convey.So(err, convey.ShouldBeNil)
		type fields struct {
			Logger *uberZap.Logger
		}
		z := &LoggerWithFormat{
			Logger: logger,
		}
		cnt := 0
		gomonkey.ApplyMethod(reflect.TypeOf(logger), "Debug",
			func(log *uberZap.Logger, msg string, fields ...uberZap.Field) {
				cnt += 1
			})
		z.Debugf("should print")
		convey.So(cnt, convey.ShouldEqual, 1)
	})
}

func TestNewCoreWithInfoLevel(t *testing.T) {
	convey.Convey("TestNewCoreWithInfoLevel", t, func() {
		coreInfo := config.CoreInfo{
			FilePath:   "tmp",
			Level:      "INFO",
			Tick:       0,
			First:      0,
			Thereafter: 0,
			Tracing:    false,
			Disable:    false,
		}
		logger, err := NewWithLevel(coreInfo, true)
		convey.So(logger, convey.ShouldNotBeNil)
		convey.So(err, convey.ShouldBeNil)
		type fields struct {
			Logger *uberZap.Logger
		}
		z := &LoggerWithFormat{
			Logger: logger,
		}
		cnt := 0
		gomonkey.ApplyMethod(reflect.TypeOf(logger), "Debug",
			func(log *uberZap.Logger, msg string, fields ...uberZap.Field) {
				cnt += 1
			})
		z.Debugf("should not print")
		convey.So(cnt, convey.ShouldEqual, 0)
	})
}
