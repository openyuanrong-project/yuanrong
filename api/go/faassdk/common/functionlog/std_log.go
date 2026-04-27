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

// Package functionlog create a log module
package functionlog

import (
	"errors"

	"yuanrong.org/kernel/runtime/libruntime/common/faas/logger"
)

const (
	defaultSplitLimit = 90*1024 - 1
)

// STDLogger log reader of stdout and stderr
type STDLogger struct {
	// logCallback if it exists, collect log and send back
	logCallback func([]byte)
	splitLimit  int
}

// CreateSTDLogger new std log reader
func CreateSTDLogger() *STDLogger {
	return CreateSTDLoggerWithSplitLimit(defaultSplitLimit)
}

// CreateSTDLoggerWithSplitLimit -
func CreateSTDLoggerWithSplitLimit(splitLimit int) *STDLogger {
	stdLogger := &STDLogger{
		logCallback: nil,
		splitLimit:  splitLimit,
	}

	return stdLogger
}

// RegisterLogCallback callback function for std log collection
func (l *STDLogger) RegisterLogCallback(cb func([]byte)) {
	l.logCallback = cb
}

// Write covered std log write method
func (l *STDLogger) Write(p []byte) (int, error) {
	n := len(p)
	if n == 0 {
		return 0, nil
	}

	if l.logCallback == nil {
		logger.GetLogger().Errorf("logCallback is nil")
		return 0, errors.New("logCallback is nil")
	}

	if n < l.splitLimit {
		l.logCallback(p)
	} else {
		start := 0
		end := l.splitLimit
		for end < n {
			l.logCallback(p[start:end])
			start = end
			end += l.splitLimit
		}
		l.logCallback(p[start:])
	}

	logger.GetLogger().Debugf("successfully log %d bytes", n)

	return n, nil
}
