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

// Package logger log
package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"yuanrong.org/kernel/runtime/libruntime/common/logger/config"
)

const defaultPerm = 0666

// NewInterfaceLogger returns a new interface logger
func NewInterfaceLogger(logPath, fileName string, cfg InterfaceEncoderConfig) (*InterfaceLogger, error) {
	coreInfo, err := config.GetCoreInfoFromEnv()
	if err != nil {
		coreInfo = config.GetDefaultCoreInfo()
	}
	filePath := filepath.Join(coreInfo.FilePath, fileName+".log")

	coreInfo.FilePath = filePath
	cfg.EncodeCaller = zapcore.ShortCallerEncoder
	// skip level to print caller line of origin log
	const skipLevel = 5
	core, err := newCore(coreInfo, cfg)
	if err != nil {
		return nil, err
	}
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(skipLevel))

	return &InterfaceLogger{log: logger}, nil
}

// InterfaceLogger interface logger which implements by zap logger
type InterfaceLogger struct {
	log *zap.Logger
}

// Write writes message information
func (logger *InterfaceLogger) Write(msg string) {
	logger.log.Debug(msg)
}

func newCore(coreInfo config.CoreInfo, cfg InterfaceEncoderConfig) (zapcore.Core, error) {
	w, err := CreateSink(coreInfo)
	if err != nil {
		return nil, err
	}
	syncer := zapcore.AddSync(w)

	encoder := NewInterfaceEncoder(cfg, false)

	priority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		var customLevel zapcore.Level
		if err := customLevel.UnmarshalText([]byte(coreInfo.Level)); err != nil {
			customLevel = zapcore.InfoLevel
		}
		return lvl >= customLevel
	})

	return zapcore.NewCore(encoder, syncer, priority), nil
}

// CreateSink creates a new zap log sink
func CreateSink(coreInfo config.CoreInfo) (io.Writer, error) {
	// create directory if not already exist
	dir := filepath.Dir(coreInfo.FilePath)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Printf("failed to mkdir: %s", dir)
		return nil, err
	}
	w, err := initRollingLog(coreInfo, os.O_WRONLY|os.O_APPEND|os.O_CREATE, defaultPerm)
	if err != nil {
		fmt.Printf("failed to open log file: %s, err: %s\n", coreInfo.FilePath, err.Error())
		return nil, err
	}
	return w, nil
}
