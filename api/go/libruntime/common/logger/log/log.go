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

// Package log is common logger client
package log

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/asaskevich/govalidator/v11"
	uberZap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"yuanrong.org/kernel/runtime/libruntime/api"
	"yuanrong.org/kernel/runtime/libruntime/common/constants"
	"yuanrong.org/kernel/runtime/libruntime/common/logger/config"
	"yuanrong.org/kernel/runtime/libruntime/common/logger/zap"
)

const snuserLogPath = "/home/snuser/log"

var (
	once             sync.Once
	onceRuntimeLog   sync.Once
	formatLogger     api.FormatLogger
	runtimeLogger    api.FormatLogger
	defaultLogger, _ = uberZap.NewProduction()
)

// InitRunLog init run log with log.json file
func InitRunLog(fileName string, isAsync bool) (api.FormatLogger, error) {
	var err error
	onceRuntimeLog.Do(
		func() {
			coreInfo := config.GetDefaultCoreInfo()
			formatLoggerImpl, newErr := newFormatLogger(fileName, isAsync, coreInfo)
			if newErr != nil {
				err = newErr
			}
			formatLogger = formatLoggerImpl
			runtimeLogger = &loggerWrapper{real: formatLogger}
		},
	)
	return formatLogger, err
}

// InitRunLogWithConfig init run log with config
func InitRunLogWithConfig(fileName string, isAsync bool, coreInfo config.CoreInfo) (api.FormatLogger, error) {
	if _, err := govalidator.ValidateStruct(coreInfo); err != nil {
		return nil, err
	}
	return newFormatLogger(fileName, isAsync, coreInfo)
}

// zapLoggerWithFormat define logger
type zapLoggerWithFormat struct {
	Logger  *uberZap.Logger
	SLogger *uberZap.SugaredLogger
}

// newFormatLogger new formatLogger with log config info
func newFormatLogger(fileName string, isAsync bool, coreInfo config.CoreInfo) (api.FormatLogger, error) {
	if strings.Compare(constants.MonitorFileName, fileName) == 0 {
		coreInfo.FilePath = snuserLogPath
	}
	coreInfo.FilePath = filepath.Join(coreInfo.FilePath, fileName+"-run.log")
	logger, err := zap.NewWithLevel(coreInfo, isAsync)
	if err != nil {
		return nil, err
	}

	return &zapLoggerWithFormat{
		Logger:  logger,
		SLogger: logger.Sugar(),
	}, nil
}

// NewConsoleLogger returns a console logger
func NewConsoleLogger() api.FormatLogger {
	logger, err := zap.NewConsoleLog()
	if err != nil {
		fmt.Println("new console log error", err)
		logger = defaultLogger
	}
	return &zapLoggerWithFormat{
		Logger:  logger,
		SLogger: logger.Sugar(),
	}
}

// GetLogger get logger directly
func GetLogger() api.FormatLogger {
	if runtimeLogger == nil {
		once.Do(
			func() {
				formatLogger = NewConsoleLogger()
				runtimeLogger = &loggerWrapper{real: formatLogger}
			},
		)
	}
	return runtimeLogger
}

// With add fields to log header
func (z *zapLoggerWithFormat) With(fields ...zapcore.Field) api.FormatLogger {
	logger := z.Logger.With(fields...)
	return &zapLoggerWithFormat{
		Logger:  logger,
		SLogger: logger.Sugar(),
	}
}

// Infof stdout format and params
func (z *zapLoggerWithFormat) Infof(format string, params ...interface{}) {
	z.SLogger.Infof(format, params...)
}

// Errorf stdout format and params
func (z *zapLoggerWithFormat) Errorf(format string, params ...interface{}) {
	z.SLogger.Errorf(format, params...)
}

// Warnf stdout format and params
func (z *zapLoggerWithFormat) Warnf(format string, params ...interface{}) {
	z.SLogger.Warnf(format, params...)
}

// Debugf stdout format and params
func (z *zapLoggerWithFormat) Debugf(format string, params ...interface{}) {
	if config.LogLevel > zapcore.DebugLevel {
		return
	}
	z.SLogger.Debugf(format, params...)
}

// Fatalf stdout format and params
func (z *zapLoggerWithFormat) Fatalf(format string, params ...interface{}) {
	z.SLogger.Fatalf(format, params...)
}

// Info stdout format and params
func (z *zapLoggerWithFormat) Info(msg string, fields ...uberZap.Field) {
	z.Logger.Info(msg, fields...)
}

// Error stdout format and params
func (z *zapLoggerWithFormat) Error(msg string, fields ...uberZap.Field) {
	z.Logger.Error(msg, fields...)
}

// Warn stdout format and params
func (z *zapLoggerWithFormat) Warn(msg string, fields ...uberZap.Field) {
	z.Logger.Warn(msg, fields...)
}

// Debug stdout format and params
func (z *zapLoggerWithFormat) Debug(msg string, fields ...uberZap.Field) {
	if config.LogLevel > zapcore.DebugLevel {
		return
	}
	z.Logger.Debug(msg, fields...)
}

// Fatal stdout format and params
func (z *zapLoggerWithFormat) Fatal(msg string, fields ...uberZap.Field) {
	z.Logger.Fatal(msg, fields...)
}

// Sync calls the underlying Core's Sync method, flushing any buffered log
// entries. Applications should take care to call Sync before exiting.
func (z *zapLoggerWithFormat) Sync() {
	z.Logger.Sync()
}

type loggerWrapper struct {
	real api.FormatLogger
}

func (l *loggerWrapper) With(fields ...zapcore.Field) api.FormatLogger {
	return &loggerWrapper{
		real: l.real.With(fields...),
	}
}

func (l *loggerWrapper) Infof(format string, params ...interface{}) {
	l.real.Infof(format, params...)
}

func (l *loggerWrapper) Errorf(format string, params ...interface{}) {
	l.real.Errorf(format, params...)
}

func (l *loggerWrapper) Warnf(format string, params ...interface{}) {
	l.real.Warnf(format, params...)
}

func (l *loggerWrapper) Debugf(format string, params ...interface{}) {
	l.real.Debugf(format, params...)
}

func (l *loggerWrapper) Fatalf(format string, params ...interface{}) {
	l.real.Fatalf(format, params...)
}

func (l *loggerWrapper) Info(msg string, fields ...uberZap.Field) {
	l.real.Info(msg, fields...)
}

func (l *loggerWrapper) Error(msg string, fields ...uberZap.Field) {
	l.real.Error(msg, fields...)
}

func (l *loggerWrapper) Warn(msg string, fields ...uberZap.Field) {
	l.real.Warn(msg, fields...)
}

func (l *loggerWrapper) Debug(msg string, fields ...uberZap.Field) {
	l.real.Debug(msg, fields...)
}

func (l *loggerWrapper) Fatal(msg string, fields ...uberZap.Field) {
	l.real.Fatal(msg, fields...)
}

func (l *loggerWrapper) Sync() {
	l.real.Sync()
}
