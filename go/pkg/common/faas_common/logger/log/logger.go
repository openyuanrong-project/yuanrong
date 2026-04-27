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

// Package log -
package log

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	uberZap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/config"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/zap"
)

const (
	skipLevel     = 1
	snuserLogPath = "/home/snuser/log"
)

type loggerWrapper struct {
	real api.FormatLogger
}

func (l *loggerWrapper) With(fields ...zapcore.Field) api.FormatLogger {
	return &loggerWrapper{
		real: l.real.With(fields...),
	}
}

func (l *loggerWrapper) Infof(format string, paras ...interface{}) {
	l.real.Infof(format, paras...)
}
func (l *loggerWrapper) Errorf(format string, paras ...interface{}) {
	l.real.Errorf(format, paras...)
}
func (l *loggerWrapper) Warnf(format string, paras ...interface{}) {
	l.real.Warnf(format, paras...)
}
func (l *loggerWrapper) Debugf(format string, paras ...interface{}) {
	l.real.Debugf(format, paras...)
}
func (l *loggerWrapper) Fatalf(format string, paras ...interface{}) {
	l.real.Fatalf(format, paras...)
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

var (
	once             sync.Once
	formatLogger     api.FormatLogger
	defaultLogger, _ = uberZap.NewProduction()
)

// InitRunLog init run log with log.json file
func InitRunLog(fileName string, isAsync bool) error {
	coreInfo, err := config.GetCoreInfoFromEnv()
	if err != nil {
		return err
	}
	if coreInfo.Disable {
		return nil
	}
	formatLogger, err = NewFormatLogger(fileName, isAsync, coreInfo)
	return err
}

// SetupLoggerLibruntime setup logger
func SetupLoggerLibruntime(runtimeLogger api.FormatLogger) {
	if runtimeLogger == nil {
		return
	}
	wrapLogger := &loggerWrapper{real: runtimeLogger}
	formatLogger = wrapLogger
}

// SetupLogger setup logger
func SetupLogger(runtimeLogger api.FormatLogger) {
	if runtimeLogger == nil {
		return
	}
	formatLogger = runtimeLogger
}

// GetLogger get logger directly
func GetLogger() api.FormatLogger {
	if formatLogger == nil {
		once.Do(func() {
			formatLogger = NewConsoleLogger()
		})
	}
	return formatLogger
}

// NewConsoleLogger returns a console logger
func NewConsoleLogger() api.FormatLogger {
	logger, err := newConsoleLog()
	if err != nil {
		fmt.Println("new console log error", err)
		logger = defaultLogger
	}
	return &zapLoggerWithFormat{
		Logger:  logger,
		SLogger: logger.Sugar(),
	}
}

// NewFormatLogger new formatLogger with log config info
func NewFormatLogger(fileName string, isAsync bool, coreInfo config.CoreInfo) (api.FormatLogger, error) {
	if strings.Compare(constant.MonitorFileName, fileName) == 0 {
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

// newConsoleLog returns a console logger based on uber zap
func newConsoleLog() (*uberZap.Logger, error) {
	outputPaths := []string{"stdout"}
	cfg := uberZap.Config{
		Level:             uberZap.NewAtomicLevelAt(uberZap.InfoLevel),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: true,
		Encoding:          "custom_console",
		OutputPaths:       outputPaths,
		ErrorOutputPaths:  outputPaths,
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:      "T",
			LevelKey:     "L",
			NameKey:      "Logger",
			MessageKey:   "M",
			CallerKey:    "C",
			LineEnding:   zapcore.DefaultLineEnding,
			EncodeLevel:  zapcore.CapitalLevelEncoder,
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}
	consoleLogger, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	return consoleLogger.WithOptions(uberZap.AddCaller(), uberZap.AddCallerSkip(skipLevel)), nil
}

// zapLoggerWithFormat define logger
type zapLoggerWithFormat struct {
	Logger  *uberZap.Logger
	SLogger *uberZap.SugaredLogger
}

// With add fields to log header
func (z *zapLoggerWithFormat) With(fields ...zapcore.Field) api.FormatLogger {
	logger := z.Logger.With(fields...)
	return &zapLoggerWithFormat{
		Logger:  logger,
		SLogger: logger.Sugar(),
	}
}

// Infof stdout format and paras
func (z *zapLoggerWithFormat) Infof(format string, paras ...interface{}) {
	z.SLogger.Infof(format, paras...)
}

// Errorf stdout format and paras
func (z *zapLoggerWithFormat) Errorf(format string, paras ...interface{}) {
	z.SLogger.Errorf(format, paras...)
}

// Warnf stdout format and paras
func (z *zapLoggerWithFormat) Warnf(format string, paras ...interface{}) {
	z.SLogger.Warnf(format, paras...)
}

// Debugf stdout format and paras
func (z *zapLoggerWithFormat) Debugf(format string, paras ...interface{}) {
	z.SLogger.Debugf(format, paras...)
}

// Fatalf stdout format and paras
func (z *zapLoggerWithFormat) Fatalf(format string, paras ...interface{}) {
	z.SLogger.Fatalf(format, paras...)
}

// Info stdout format and paras
func (z *zapLoggerWithFormat) Info(msg string, fields ...uberZap.Field) {
	z.Logger.Info(msg, fields...)
}

// Error stdout format and paras
func (z *zapLoggerWithFormat) Error(msg string, fields ...uberZap.Field) {
	z.Logger.Error(msg, fields...)
}

// Warn stdout format and paras
func (z *zapLoggerWithFormat) Warn(msg string, fields ...uberZap.Field) {
	z.Logger.Warn(msg, fields...)
}

// Debug stdout format and paras
func (z *zapLoggerWithFormat) Debug(msg string, fields ...uberZap.Field) {
	z.Logger.Debug(msg, fields...)
}

// Fatal stdout format and paras
func (z *zapLoggerWithFormat) Fatal(msg string, fields ...uberZap.Field) {
	z.Logger.Fatal(msg, fields...)
}

// Sync calls the underlying Core's Sync method, flushing any buffered log
// entries. Applications should take care to call Sync before exiting.
func (z *zapLoggerWithFormat) Sync() {
	err := z.Logger.Sync()
	if err != nil {
		return
	}
}
