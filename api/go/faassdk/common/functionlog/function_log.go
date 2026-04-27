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

// Package functionlog gets function services from URNs
package functionlog

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"yuanrong.org/kernel/runtime/faassdk/common/constants"
	"yuanrong.org/kernel/runtime/faassdk/config"
	"yuanrong.org/kernel/runtime/faassdk/utils/urnutils"
	logger2 "yuanrong.org/kernel/runtime/libruntime/common/faas/logger"
	"yuanrong.org/kernel/runtime/libruntime/common/logger"
	"yuanrong.org/kernel/runtime/libruntime/common/logger/async"
	runtimeLoggerConfig "yuanrong.org/kernel/runtime/libruntime/common/logger/config"
	"yuanrong.org/kernel/runtime/libruntime/common/utils"
)

const (
	innerLogSeparator = ","
	funcLogSeparator  = ";"

	// Max length of function log in "tail" mode
	maxTailLogSizeKB = 2 * 1024
	// Max length of normal instance initializer log
	maxInitLogSizeMB = 10 * 1024 * 1024
	// InitializerInvokeID -
	InitializerInvokeID = "initializer"
	// LoadInvokeID -
	LoadInvokeID = "load"
	// RestoreHookInvokeID -
	RestoreHookInvokeID = "restore"

	// LogTimeFormat -
	LogTimeFormat = "2006-01-02T15:04:05Z"
	// NanoLogLayout  -
	NanoLogLayout = "2006-01-02T15:04:05.999999999Z07:00"

	// RuntimeLogFormat  -
	RuntimeLogFormat = "2006-01-02 15:04:05.999999999Z07:00"

	userLogTimeFormat = "2006-01-02 15:04:05.999"

	logTankServiceLength = 2
	originalLocation     = 0
	alternativeLocation  = 1
	userFuncLogChanCap   = 5000
	funcInfoMapCap       = 10
	userLogKey           = "log"
	fileMode             = 0640
	cacheLimit           = 10 * 1 << 20 // 10 mb
	uint64Width          = 8

	dateIndex    = 0
	timeIndex    = 1
	traceIDIndex = 2
	minLength    = 3
	traceIDLen   = 36
)

var (
	functionLogger   *FunctionLogger
	createLoggerErr  error
	createLoggerOnce sync.Once
)

// FunctionLogger process function's log
type FunctionLogger struct {
	logger     *zap.Logger
	fieldsPool *sync.Pool
	logPool    *sync.Pool
	// key is invokeID, value is logRecorder
	logRecorders sync.Map
	logLevel     string
	logAbsPath   string
}

// NewFunctionLogger -
func NewFunctionLogger(logger *zap.Logger, logLevel string, logPath string) *FunctionLogger {
	return &FunctionLogger{
		logger: logger,
		fieldsPool: &sync.Pool{New: func() interface{} {
			return newLogFields()
		}},
		logPool: &sync.Pool{New: func() interface{} {
			return NewFunctionLog()
		}},
		logLevel:   logLevel,
		logAbsPath: logPath,
	}
}

type logTankService struct {
	originalGroupID     string
	originalStreamID    string
	alternativeGroupID  string
	alternativeStreamID string
}

// GetFunctionLogger -
func GetFunctionLogger(cfg *config.Configuration) (*FunctionLogger, error) {
	createLoggerOnce.Do(func() {
		functionLogger, createLoggerErr = newLTSFunctionLogger(cfg)
		if createLoggerErr != nil {
			return
		}
		if functionLogger == nil {
			createLoggerErr = errors.New("failed to new FunctionLogger")
			return
		}
	})
	return functionLogger, createLoggerErr
}

// SetLogLevel -
func (f *FunctionLogger) SetLogLevel(level string) {
	logger2.GetLogger().Infof("set function log level: %s", level)
	f.logLevel = level
}

// NewLogRecorder new "invokeID:LogRecorder" pair in FunctionLogger
func (f *FunctionLogger) NewLogRecorder(invokeID, traceID, stage string, opts ...LogRecorderOption) *LogRecorder {
	r := &LogRecorder{
		f:                      f,
		invokeID:               invokeID,
		traceID:                traceID,
		logLevel:               f.logLevel,
		stage:                  stage,
		separatedWithLineBreak: true,
	}
	r.logs = NewFixSizeRecorder(maxTailLogSizeKB, r.guessSize, r.handleDropped)
	for _, opt := range opts {
		opt(r)
	}

	f.logRecorders.Store(invokeID, r)
	return r
}

// GetLogRecorder get LogRecorder of invokeID in FunctionLogger
func (f *FunctionLogger) GetLogRecorder(invokeID string) *LogRecorder {
	value, ok := f.logRecorders.Load(invokeID)
	if !ok {
		return nil
	}
	if recorder, ok := value.(*LogRecorder); ok {
		return recorder
	}
	return nil
}

// WriteStdLog write std logs shown on all requests
func (f *FunctionLogger) WriteStdLog(logString, timeStamp string, isReserved bool, nanoTimestamp string) {
	l := f.AcquireLog()
	l.Time = timeStamp
	l.NanoTime = nanoTimestamp
	l.IsStdLog = true
	l.Level = constants.FuncLogLevelInfo
	l.Message.WriteString(logString)
	var logged bool
	f.logRecorders.Range(func(_, value interface{}) bool {
		logged = true
		if recorder, ok := value.(*LogRecorder); ok {
			recorder.Write(l)
		}
		return true
	})
	if !logged {
		f.WriteDefaultLog(l)
	}
}

// AcquireLog retrieves a FunctionLog struct from the cached pool
func (f *FunctionLogger) AcquireLog() *FunctionLog {
	functionLogIf := f.logPool.Get()
	functionLog, ok := functionLogIf.(*FunctionLog)
	if !ok {
		logger2.GetLogger().Errorf("failed to assert FunctionLog")
		return nil
	}
	functionLog.Reset()
	return functionLog
}

// ReleaseLog puts a FunctionLog struct to the cached pool
func (f *FunctionLogger) ReleaseLog(functionLog *FunctionLog) {
	f.logPool.Put(functionLog)
}

// WriteDefaultLog -
func (f *FunctionLogger) WriteDefaultLog(functionLog *FunctionLog) {
	f.write(getDefaultInternalFunctionLog(functionLog))
}

func getDefaultInternalFunctionLog(functionLog *FunctionLog) *internalFunctionLog {
	return &internalFunctionLog{
		FunctionLog:  functionLog,
		FunctionName: urnutils.LocalFuncURN.Name,
		TraceID:      "default",
		Stage:        constants.InitializeStage,
		Index:        constants.DefaultFuncLogIndex,
	}
}

func (f *FunctionLogger) deleteLogRecorder(invokeID string) {
	f.logRecorders.Delete(invokeID)
}

// RefreshFileModTime timer of refreshing file modified time in case of the user log file being deleted
func (f *FunctionLogger) RefreshFileModTime(stopCh chan struct{}) {
	if stopCh == nil {
		logger2.GetLogger().Warnf("empty stop channel")
		return
	}
	ticker := time.NewTicker(timeModInterval)
	logger2.GetLogger().Infof("start to regularly modify the file modification time")
	for {
		select {
		case <-ticker.C:
			f.refreshWithRetry()
		case <-stopCh:
			logger2.GetLogger().Warnf("received the runtime exit signal and stopped refreshing the file " +
				"modification time")
			ticker.Stop()
			return
		}
	}
}

func (f *FunctionLogger) refreshWithRetry() {
	newTime := time.Now()
	for i := 0; i < timeModRetry; i++ {
		path := logger.GetLogName(f.logAbsPath)
		if path == "" {
			path = f.logAbsPath
		}
		if err := os.Chtimes(path, newTime, newTime); err != nil {
			logger2.GetLogger().Warnf("failed to change the modification time of the user log file: %s",
				err.Error())
			continue
		}
		logger2.GetLogger().Infof("succeeded to change the modification time of the user log file")
		break
	}
}

func newZapLogger(fullPath, messageKey string) (*zap.Logger, error) {
	coreInfo, err := runtimeLoggerConfig.GetCoreInfoFromEnv()
	if err != nil {
		logger2.GetLogger().Errorf("failed to get core info: %s", err.Error())
		coreInfo = runtimeLoggerConfig.GetDefaultCoreInfo()
	}
	coreInfo.FilePath = fullPath
	if messageKey == userLogKey {
		coreInfo.IsUserLog = true
	}
	sink, err := logger.CreateSink(coreInfo)
	if err != nil {
		logger2.GetLogger().Errorf("failed to create sink: %s", err.Error())
		return nil, err
	}

	ws := async.NewAsyncWriteSyncer(sink, async.WithCachedLimit(cacheLimit))

	encoderConfig := zapcore.EncoderConfig{
		LevelKey:     "Level",
		NameKey:      "Logger",
		MessageKey:   messageKey,
		CallerKey:    "CallerKey",
		LineEnding:   zapcore.DefaultLineEnding,
		EncodeLevel:  zapcore.CapitalLevelEncoder,
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	rollingFileEncoder := zapcore.NewJSONEncoder(encoderConfig)

	priority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.DebugLevel
	})

	return zap.New(zapcore.NewCore(rollingFileEncoder, ws, priority)), nil
}

func newLTSFunctionLogger(cfg *config.Configuration) (*FunctionLogger, error) {
	path := os.Getenv("function-log")
	if path == "" {
		path = cfg.StartArgs.LogDir
	}
	if err := utils.ValidateFilePath(path); err != nil {
		logger2.GetLogger().Errorf("failed to valid log path, err: %s", err.Error())
		return nil, err
	}

	path = filepath.Join(path, urnutils.LocalFuncURN.TenantID)
	c := &urnutils.ComplexFuncName{}
	logger2.GetLogger().Infof("POD_NAME is: %s", os.Getenv("POD_NAME"))

	c.ParseFrom(urnutils.LocalFuncURN.Name)
	l := &logTankService{}
	l.extractLogTankService(cfg.RuntimeConfig.LogTankService.GroupID, cfg.RuntimeConfig.LogTankService.StreamID)

	if err := createFunctionInfoFile(c); err != nil {
		logger2.GetLogger().Errorf("failed to create function info file, %s", err.Error())
		return nil, err
	}

	// functionName@serviceID@version@podName@time#logGroupId#logStreamId
	functionLogName := fmt.Sprintf("%s@%s@%s@%s@%s#%s#%s#%s", c.FuncName, c.ServiceID,
		urnutils.LocalFuncURN.Version, os.Getenv("POD_NAME"), time.Now().Format("20060102150405"),
		l.originalGroupID, l.originalStreamID, cfg.UserLogTag)

	fullPath, err := getAbsFilePath(path, functionLogName)
	if err != nil {
		return nil, err
	}

	rollingLogger, err := newZapLogger(fullPath, userLogKey)
	if err != nil {
		logger2.GetLogger().Errorf("failed to new zapLogger, %s", err.Error())
		return nil, err
	}

	rollingLogger = createLTSLogger(rollingLogger, c, l)

	return NewFunctionLogger(rollingLogger, cfg.RuntimeConfig.FuncLogLevel, fullPath), nil
}

func createLTSLogger(rollingLogger *zap.Logger, c *urnutils.ComplexFuncName, l *logTankService) *zap.Logger {
	return rollingLogger.With(zapcore.Field{
		Key:    "projectId",
		Type:   zapcore.StringType,
		String: urnutils.LocalFuncURN.TenantID,
	}, zapcore.Field{
		Key:    "podName",
		Type:   zapcore.StringType,
		String: os.Getenv("POD_NAME"),
	}, zapcore.Field{
		Key:    "package",
		Type:   zapcore.StringType,
		String: c.ServiceID,
	}, zapcore.Field{
		Key:    "function",
		Type:   zapcore.StringType,
		String: c.FuncName,
	}, zapcore.Field{
		Key:    "version",
		Type:   zapcore.StringType,
		String: urnutils.LocalFuncURN.Version,
	}, zapcore.Field{
		Key:    "stream",
		Type:   zapcore.StringType,
		String: "stdout",
	}, zapcore.Field{
		Key:    "instanceId",
		Type:   zapcore.StringType,
		String: os.Getenv("POD_ID"),
	}, zapcore.Field{
		Key:    "newLogGroupId",
		Type:   zapcore.StringType,
		String: l.alternativeGroupID,
	}, zapcore.Field{
		Key:    "newLogStreamId",
		Type:   zapcore.StringType,
		String: l.alternativeStreamID,
	})
}

func (l *logTankService) extractLogTankService(groupID, streamID string) {
	if strings.Contains(groupID, funcLogSeparator) && strings.Contains(streamID, funcLogSeparator) {
		splitGroupID := strings.Split(groupID, funcLogSeparator)
		splitStreamID := strings.Split(streamID, funcLogSeparator)
		if len(splitGroupID) != logTankServiceLength || len(splitStreamID) != logTankServiceLength {
			l.originalGroupID = groupID
			l.originalStreamID = streamID
			return
		}
		l.originalGroupID = splitGroupID[originalLocation]
		l.alternativeGroupID = splitGroupID[alternativeLocation]
		l.originalStreamID = splitStreamID[originalLocation]
		l.alternativeStreamID = splitStreamID[alternativeLocation]
		return
	}
	l.originalGroupID = groupID
	l.originalStreamID = streamID
}

func createFunctionInfoFile(complexFuncName *urnutils.ComplexFuncName) error {
	coreInfo, err := runtimeLoggerConfig.GetCoreInfoFromEnv()
	if err != nil {
		logger2.GetLogger().Errorf("failed to get core info: %s", err.Error())
		coreInfo = runtimeLoggerConfig.GetDefaultCoreInfo()
	}
	// urn:fss:projectID:function:package:function-name:version.info
	infoFileName := strings.Join([]string{"urn:fss", urnutils.LocalFuncURN.TenantID,
		"function", complexFuncName.ServiceID, complexFuncName.FuncName,
		urnutils.LocalFuncURN.Version + ".info"}, ":")
	f, err := os.OpenFile(filepath.Join(coreInfo.FilePath, infoFileName), os.O_RDONLY|os.O_CREATE, fileMode)
	if err != nil {
		return err
	}
	defer f.Close()

	return nil
}

func getAbsFilePath(path, fileName string) (string, error) {
	logPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	fullPath := filepath.Join(logPath, fileName+".log")
	return fullPath, nil
}
