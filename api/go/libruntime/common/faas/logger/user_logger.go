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

// Package logger for printing user runtime logger
package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	fileMode    = 0750
	logFileMode = 0644
	skipLevel   = 1
	loglevel    = zapcore.InfoLevel

	logFile        = "user-function.log"
	defaultLogPath = "/home/snuser/log"
)

var (
	userLogger *UserFunctionLogger
)

// SetupUserLogger to new a logger handler
func SetupUserLogger(level string) error {
	var l zapcore.Level

	if level == "" {
		l = loglevel
	} else {
		_ = l.Set(level)
	}
	userLogPath := os.Getenv("RUNTIME_LOG_DIR")
	if userLogPath == "" {
		userLogPath = defaultLogPath
	}
	if err := os.MkdirAll(userLogPath, fileMode); err != nil && !os.IsExist(err) {
		return err
	}
	fileName := filepath.Join(userLogPath, logFile)
	writeSyncer, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, logFileMode)
	if err != nil {
		return err
	}
	Config := zapcore.EncoderConfig{
		NameKey:      "userLogger",
		CallerKey:    "C",
		TimeKey:      "T",
		LevelKey:     "L",
		MessageKey:   "M",
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeLevel:  zapcore.CapitalLevelEncoder,
		LineEnding:   zapcore.DefaultLineEnding,
	}

	encoder := zapcore.NewConsoleEncoder(Config)
	core := zapcore.NewCore(encoder, writeSyncer, l)
	userLogger = &UserFunctionLogger{log: zap.New(core, zap.AddCaller(), zap.AddCallerSkip(skipLevel)).Sugar()}
	return nil
}

// UserFunctionLogger user log struct ,witch context use it
type UserFunctionLogger struct {
	log *zap.SugaredLogger
}

// GetUserLogger get user logger
func GetUserLogger() *UserFunctionLogger {
	return userLogger
}

// Infof to record info log
func (u *UserFunctionLogger) Infof(format string, params ...interface{}) {
	u.log.Infof(format, params...)
}

// Debugf to record info log
func (u *UserFunctionLogger) Debugf(format string, params ...interface{}) {
	u.log.Debugf(format, params...)
}

// Warnf to record info log
func (u *UserFunctionLogger) Warnf(format string, params ...interface{}) {
	u.log.Warnf(format, params...)
}

// Errorf to record info log
func (u *UserFunctionLogger) Errorf(format string, params ...interface{}) {
	u.log.Errorf(format, params...)
}
