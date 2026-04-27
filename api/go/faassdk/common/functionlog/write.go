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

package functionlog

import (
	"strings"
	"time"

	"go.uber.org/zap/zapcore"

	"yuanrong.org/kernel/runtime/faassdk/common/constants"
	"yuanrong.org/kernel/runtime/faassdk/config"
	"yuanrong.org/kernel/runtime/libruntime/common/faas/logger"
)

// Sync ensure all log is written in file before faas-executor exit
func Sync(conf *config.Configuration) {
	f, err := GetFunctionLogger(conf)
	if err != nil {
		logger.GetLogger().Warnf("failed to get functionLogger: %s", err.Error())
		return
	}

	f.logger.Sync()
}

func (f *FunctionLogger) write(funcLog *internalFunctionLog) {
	if f.logger == nil {
		logger.GetLogger().Errorf("invalid logger")
		return
	}
	defer f.ReleaseLog(funcLog.FunctionLog)

	var level zapcore.Level
	if err := level.UnmarshalText([]byte(funcLog.Level)); err != nil {
		level = zapcore.InfoLevel
	}
	message := getLTSLoggerMessage(funcLog)
	ent := zapcore.Entry{
		LoggerName: "",
		Time:       time.Now(),
		Level:      level,
		Message:    message,
	}

	logFieldsIf := f.fieldsPool.Get()
	defer f.fieldsPool.Put(logFieldsIf)

	logFileds, ok := logFieldsIf.(logFileds)
	if !ok {
		logger.GetLogger().Errorf("failed to assert logFields")
		return
	}
	logFileds.set(funcLog)

	err := f.logger.Core().Write(ent, logFileds)
	if err != nil {
		logger.GetLogger().Errorf("failed to write, %s", err.Error())
		return
	}
}

func getLTSLoggerMessage(funcLog *internalFunctionLog) string {
	if !funcLog.IsStdLog {
		if funcLog.Index > 0 {
			var b strings.Builder
			b.WriteString(funcLog.Time)
			b.WriteString(" ")
			b.WriteString(funcLog.TraceID)
			b.WriteString(" ")
			b.WriteString(funcLog.Level)
			b.WriteString(" ")
			b.Write(funcLog.Message.Bytes())
			return b.String()
		}
		if funcLog.Index == constants.DefaultFuncLogIndex {
			var b strings.Builder
			b.WriteString(funcLog.Time)
			b.WriteString(" ")
			b.Write(funcLog.Message.Bytes())
			return b.String()
		}
	}
	return funcLog.Message.String()
}
