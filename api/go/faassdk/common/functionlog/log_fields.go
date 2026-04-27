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
	"strconv"

	"go.uber.org/zap/zapcore"

	"yuanrong.org/kernel/runtime/libruntime/common/faas/logger"
)

const (
	traceIDPos = iota
	timePos
	errorTypePos
	stagePos
	statusPos
	finishPos
	livedataTraceIDPos
	endFields
)

type logFileds []zapcore.Field

func newLogFields() logFileds {
	return []zapcore.Field{
		{
			Key:  "requestId",
			Type: zapcore.StringType,
		},
		{
			Key:  "time",
			Type: zapcore.StringType,
		},
		{
			Key:  "errorType",
			Type: zapcore.StringType,
		},
		{
			Key:  "stage",
			Type: zapcore.StringType,
		},
		{
			Key:  "status",
			Type: zapcore.StringType,
		},
		{
			Key:  "finishLog",
			Type: zapcore.StringType,
		},
		{
			Key:  "livedataTraceId",
			Type: zapcore.StringType,
		},
	}
}

func (f logFileds) set(funcLog *internalFunctionLog) {
	if len(f) < endFields {
		logger.GetLogger().Warnf("invalid logFields")
		return
	}

	f[traceIDPos].String = funcLog.TraceID
	f[timePos].String = funcLog.NanoTime
	f[errorTypePos].String = funcLog.ErrorType
	f[stagePos].String = funcLog.Stage

	statusStr := "success"
	if funcLog.ErrorType != "" {
		statusStr = "fail"
	}
	f[statusPos].String = statusStr
	f[finishPos].String = strconv.FormatBool(funcLog.IsFinishedLog)
	f[livedataTraceIDPos].String = funcLog.LivedataID
}
