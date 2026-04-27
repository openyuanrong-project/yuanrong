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

// Package zap zapper log
package zap

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"yuanrong.org/kernel/runtime/libruntime/common/logger"
	"yuanrong.org/kernel/runtime/libruntime/common/logger/async"
	"yuanrong.org/kernel/runtime/libruntime/common/logger/config"
)

const (
	skipLevel = 2
)

func init() {
	zap.RegisterEncoder("custom_console", logger.NewConsoleEncoder)
}

// NewDevelopmentLog returns a development logger based on uber zap and it output entry to stdout and stderr
func NewDevelopmentLog() (*zap.Logger, error) {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return cfg.Build()
}

// NewConsoleLog returns a console logger based on uber zap
func NewConsoleLog() (*zap.Logger, error) {
	outputPaths := []string{"stdout"}
	cfg := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
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
	return consoleLogger.WithOptions(zap.AddCaller(), zap.AddCallerSkip(skipLevel)), nil
}

// NewWithLevel returns a log based on zap with Level
func NewWithLevel(coreInfo config.CoreInfo, isAsync bool) (*zap.Logger, error) {
	core, err := newCore(coreInfo, isAsync)
	if err != nil {
		return nil, err
	}

	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(skipLevel)), nil
}

func newCore(coreInfo config.CoreInfo, isAsync bool) (zapcore.Core, error) {
	w, err := logger.CreateSink(coreInfo)
	if err != nil {
		return nil, err
	}

	var syncer zapcore.WriteSyncer
	if isAsync {
		syncer = async.NewAsyncWriteSyncer(w)
	} else {
		syncer = zapcore.AddSync(w)
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:      "T",
		LevelKey:     "L",
		NameKey:      "Logger",
		MessageKey:   "M",
		CallerKey:    "C",
		LineEnding:   zapcore.DefaultLineEnding,
		EncodeLevel:  zapcore.CapitalLevelEncoder,
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	fileEncoder := logger.NewCustomEncoder(&encoderConfig)

	if err := config.LogLevel.UnmarshalText([]byte(coreInfo.Level)); err != nil {
		config.LogLevel = zapcore.InfoLevel
	}

	priority := zap.LevelEnablerFunc(priorityHandler)

	if coreInfo.Tick == 0 || coreInfo.First == 0 || coreInfo.Thereafter == 0 {
		return zapcore.NewCore(fileEncoder, syncer, priority), nil
	}
	return zapcore.NewSamplerWithOptions(zapcore.NewCore(fileEncoder, syncer, priority),
		time.Duration(coreInfo.Tick)*time.Second, coreInfo.First, coreInfo.Thereafter), nil
}

func priorityHandler(lvl zapcore.Level) bool {
	return lvl >= config.LogLevel
}

// LoggerWithFormat zap logger
type LoggerWithFormat struct {
	*zap.Logger
}

// Infof stdout format and paras
func (z *LoggerWithFormat) Infof(format string, paras ...interface{}) {
	z.Logger.Info(fmt.Sprintf(format, paras...))
}

// Errorf stdout format and paras
func (z *LoggerWithFormat) Errorf(format string, paras ...interface{}) {
	z.Logger.Error(fmt.Sprintf(format, paras...))
}

// Warnf stdout format and paras
func (z *LoggerWithFormat) Warnf(format string, paras ...interface{}) {
	z.Logger.Warn(fmt.Sprintf(format, paras...))
}

// Debugf stdout format and paras
func (z *LoggerWithFormat) Debugf(format string, paras ...interface{}) {
	if config.LogLevel > zapcore.DebugLevel {
		return
	}
	z.Logger.Debug(fmt.Sprintf(format, paras...))
}

// Fatalf stdout format and paras
func (z *LoggerWithFormat) Fatalf(format string, paras ...interface{}) {
	z.Logger.Fatal(fmt.Sprintf(format, paras...))
}
