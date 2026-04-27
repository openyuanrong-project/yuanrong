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
	"bytes"
	"container/list"
	"encoding/base64"
	"sync"
	"time"

	"yuanrong.org/kernel/runtime/faassdk/common/constants"
	"yuanrong.org/kernel/runtime/faassdk/utils/urnutils"
)

const (
	// user log will be deleted by cffagent for follow reasons:
	// 1.irregular log file name
	// 2.unchanged for 12 hours when log is not empty
	// 3.unchanged for 6 hours when log is empty
	timeModInterval = 5 * time.Hour
	timeModRetry    = 3
)

// LogHeaderFooterGenerator -
type LogHeaderFooterGenerator interface {
	GenerateLogHeader(*FunctionLog)
	GenerateLogFooter(*FunctionLog)
}

// FixSizeRecorder tries to maintain a number of logs with the total size just over limitSize. When we write some new
// logs to the recorder, old logs will be dropped.
type FixSizeRecorder struct {
	logs      *list.List
	guessSize func(funcLog *FunctionLog) int
	dropped   func(funcLog *FunctionLog, logTooLarge bool)
	size      int
	limitSize int
}

// NewFixSizeRecorder -
func NewFixSizeRecorder(
	limitSize int,
	guessSize func(funcLog *FunctionLog) int,
	dropped func(funcLog *FunctionLog, logTooLarge bool)) *FixSizeRecorder {
	return &FixSizeRecorder{
		guessSize: guessSize,
		dropped:   dropped,
		limitSize: limitSize,
	}
}

// Reset resets the recorder to its initial state
func (r *FixSizeRecorder) Reset() {
	if r.logs != nil {
		r.logs.Init()
	}
	r.size = 0
}

// Write -
func (r *FixSizeRecorder) Write(funcLog *FunctionLog) {
	if r.logs == nil {
		r.logs = list.New()
	}

	if r.size > r.limitSize {
		for {
			elem := r.logs.Front()
			if elem == nil {
				break
			}
			funcLog, ok := elem.Value.(*FunctionLog)
			if !ok {
				break
			}

			size := r.guessSize(funcLog)
			if r.size-size < r.limitSize {
				break
			}

			r.size -= size
			r.dropped(funcLog, false)
			r.logs.Remove(elem)
		}
	}

	size := r.guessSize(funcLog)
	if size > r.limitSize { // special case: the funcLog is too large
		r.Range(func(fl *FunctionLog) {
			r.dropped(fl, false)
		})
		r.logs = nil
		r.size = 0
		r.dropped(funcLog, true)
	} else {
		r.logs.PushBack(funcLog)
		r.size += size
	}
}

// RealTimeWrite -
func (r *FixSizeRecorder) RealTimeWrite(funcLog *FunctionLog) {
	if r.logs == nil {
		r.logs = list.New()
	}

	if r.size > r.limitSize {
		for {
			elem := r.logs.Front()
			if elem == nil {
				break
			}
			funcLog, ok := elem.Value.(*FunctionLog)
			if !ok {
				break
			}

			size := r.guessSize(funcLog)
			if r.size-size < r.limitSize {
				break
			}
			r.size -= size
			r.logs.Remove(elem)
		}
	}
	size := r.guessSize(funcLog)
	if size > r.limitSize {
		r.dropped(funcLog, true)
		return
	}
	r.logs.PushBack(funcLog.DeepCopy())
	r.size += size
	r.dropped(funcLog, false)
}

// Range -
func (r *FixSizeRecorder) Range(fn func(*FunctionLog)) {
	if r.logs == nil {
		return
	}

	for e := r.logs.Front(); e != nil; e = e.Next() {
		funcLog, ok := e.Value.(*FunctionLog)
		if !ok {
			continue
		}

		fn(funcLog)
	}
}

// WriteLimit is to save limitSize log
func (r *FixSizeRecorder) WriteLimit(funcLog *FunctionLog, limitSize int) {
	if r.logs == nil {
		r.logs = list.New()
	}

	if r.size > limitSize {
		for {
			elem := r.logs.Front()
			if elem == nil {
				break
			}
			funcLog, ok := elem.Value.(*FunctionLog)
			if !ok {
				break
			}

			size := r.guessSize(funcLog)
			if r.size-size < limitSize {
				break
			}
			r.size -= size
			r.logs.Remove(elem)
		}
	}

	size := r.guessSize(funcLog)
	r.logs.PushBack(funcLog)
	r.size += size
}

// LogRecorder record user log of function
type LogRecorder struct {
	f         *FunctionLogger
	header    *FunctionLog
	footer    *FunctionLog
	logs      *FixSizeRecorder
	logsMux   sync.Mutex
	generator LogHeaderFooterGenerator
	idx       int
	base      []byte
	syncCh    chan struct{}

	invokeID               string
	traceID                string
	stage                  string
	logLevel               string
	livedataID             string
	logOption              string
	separatedWithLineBreak bool
}

type internalFunctionLog struct {
	*FunctionLog
	FunctionName string
	TraceID      string
	LivedataID   string
	Stage        string
	Index        int
}

// FunctionLog is format of function logs which are flushed
type FunctionLog struct {
	Message       *bytes.Buffer
	Time          string
	NanoTime      string
	Level         string
	ErrorType     string
	IsFinishedLog bool
	IsStdLog      bool
}

// NewFunctionLog -
func NewFunctionLog() *FunctionLog {
	return &FunctionLog{Message: new(bytes.Buffer)}
}

// Reset -
func (l *FunctionLog) Reset() {
	l.Time = ""
	l.Level = ""
	l.Message.Reset()
	l.ErrorType = ""
	l.IsFinishedLog = false
	l.IsStdLog = false
}

// DeepCopy -
func (l *FunctionLog) DeepCopy() *FunctionLog {
	x := *l // copy
	oldBuf := l.Message.Bytes()
	newBuf := make([]byte, len(oldBuf))
	copy(newBuf, oldBuf)
	x.Message = bytes.NewBuffer(newBuf)
	return &x
}

// NewLogRecorder -
func NewLogRecorder() *LogRecorder {
	r := &LogRecorder{}
	r.logs = NewFixSizeRecorder(maxTailLogSizeKB, r.guessSize, r.handleDropped)
	return r
}

func (r *LogRecorder) handleDropped(funcLog *FunctionLog, logTooLarge bool) {
	if r.idx == 0 {
		if header := r.getHeader(); header != nil {
			// Header can still be used in 'MarshalAll'. Pass in a copy to prevent releasing the old header
			// to the sync.Pool.
			r.f.write(r.generateInternalFunctionLog(header.DeepCopy(), 0))
		}
		r.idx++
	}

	// when new drop happens, we no longer need the previous base
	if r.base != nil {
		r.base = nil
	}

	if logTooLarge {
		res := r.formatLog(funcLog)
		if len(res) > maxTailLogSizeKB {
			res = res[len(res)-maxTailLogSizeKB:]
		}
		r.base = make([]byte, len(res))
		copy(r.base, res)
	}

	r.f.write(r.generateInternalFunctionLog(funcLog, r.idx))
	r.idx++
}

// guessSize returns the guessed size of bytes when calling 'formatLog' for the same 'functionLog'. The guessed size
// may not be accurate.
func (r *LogRecorder) guessSize(functionLog *FunctionLog) int {
	if functionLog.IsStdLog {
		return functionLog.Message.Len()
	}
	// 1 is the ' ' in between
	return len(functionLog.Time) + 1 + len(r.traceID) + 1 + len(r.logLevel) + 1 + functionLog.Message.Len()
}

func (r *LogRecorder) formatLog(functionLog *FunctionLog) []byte {
	// user std log without additional information
	if functionLog.IsStdLog {
		return functionLog.Message.Bytes()
	}
	level := functionLog.Level
	if level == "WARNING" {
		level = constants.FuncLogLevelWarn
	}

	var b bytes.Buffer
	b.Grow(len(functionLog.Time) + 1 + len(r.traceID) + 1 + len(level) + 1 + functionLog.Message.Len())
	b.WriteString(functionLog.Time)
	b.WriteString(" ")
	b.WriteString(r.traceID)
	b.WriteString(" ")
	b.WriteString(level)
	b.WriteString(" ")
	b.Write(functionLog.Message.Bytes())

	return b.Bytes()
}

// StartSync -
func (r *LogRecorder) StartSync() {
	r.syncCh = make(chan struct{})
}

// FinishSync -
func (r *LogRecorder) FinishSync() {
	select {
	case <-r.syncCh: // If chan is closed, the case is executed immediately.
		return
	default: // If chan is not closed, the case is executed.
		close(r.syncCh)
	}
}

// MarshalAll -
func (r *LogRecorder) MarshalAll() string {
	if r.syncCh != nil {
		<-r.syncCh
	}

	r.logsMux.Lock()

	var buffer bytes.Buffer
	if r.base != nil {
		buffer.Write(r.base)
		if r.separatedWithLineBreak {
			buffer.WriteString("\n")
		}
	}
	r.logs.Range(func(piece *FunctionLog) {
		buffer.Write(r.formatLog(piece))
		if r.separatedWithLineBreak {
			buffer.WriteString("\n")
		}
	})

	res := buffer.String()
	if len(res) > maxTailLogSizeKB {
		res = res[len(res)-maxTailLogSizeKB:]
	}

	buffer.Reset()
	if header := r.getHeader(); header != nil {
		buffer.Write(header.Message.Bytes())
	}
	buffer.WriteString("\n")
	buffer.WriteString(res)

	if footer := r.getFooter(); footer != nil {
		buffer.Write(footer.Message.Bytes())
	}

	encoded := base64.StdEncoding.EncodeToString(buffer.Bytes())
	r.logsMux.Unlock()
	return encoded
}

// Finish finalizes the logRecorder. This MUST be called.
func (r *LogRecorder) Finish() {
	if r.syncCh != nil {
		<-r.syncCh
	}

	r.logsMux.Lock()

	if r.idx == 0 {
		if header := r.getHeader(); header != nil {
			r.f.write(r.generateInternalFunctionLog(header, 0))
		}
		r.idx++
	}
	r.logs.Range(func(piece *FunctionLog) {
		r.f.write(r.generateInternalFunctionLog(piece, r.idx))
		r.idx++
	})
	if footer := r.getFooter(); footer != nil {
		r.f.write(r.generateInternalFunctionLog(footer, -1))
	}

	r.f.deleteLogRecorder(r.invokeID)

	r.logsMux.Unlock()
}

func (r *LogRecorder) generateInternalFunctionLog(functionLog *FunctionLog, idx int) *internalFunctionLog {
	return &internalFunctionLog{
		FunctionLog:  functionLog,
		FunctionName: urnutils.LocalFuncURN.Name,
		TraceID:      r.traceID,
		LivedataID:   r.livedataID,
		Stage:        r.stage,
		Index:        idx,
	}
}

func (r *LogRecorder) getHeader() *FunctionLog {
	if r.header != nil {
		return r.header
	}

	if r.generator != nil {
		r.header = r.f.AcquireLog()
		r.generator.GenerateLogHeader(r.header)
	}
	return r.header
}

func (r *LogRecorder) getFooter() *FunctionLog {
	if r.footer != nil {
		return r.footer
	}

	if r.generator != nil {
		r.footer = r.f.AcquireLog()
		r.generator.GenerateLogFooter(r.footer)
	}
	return r.footer
}

// Write log to logRecorder. This can be called from a different goroutine
func (r *LogRecorder) Write(funcLog *FunctionLog) {
	r.logsMux.Lock()
	r.logs.Write(funcLog)
	r.logsMux.Unlock()
}

// WriteLimit write limited log to logRecorder. not reserved to disk
func (r *LogRecorder) WriteLimit(funcLog *FunctionLog) {
	r.logsMux.Lock()
	r.logs.WriteLimit(funcLog, maxInitLogSizeMB)
	r.logsMux.Unlock()
}

// RealTimeWrite log to logRecorder. This can be called from a different goroutine
func (r *LogRecorder) RealTimeWrite(funcLog *FunctionLog) {
	r.logsMux.Lock()
	r.logs.RealTimeWrite(funcLog)
	r.logsMux.Unlock()
}

// LogOption -
func (r *LogRecorder) LogOption() string {
	return r.logOption
}

// InvokeID -
func (r *LogRecorder) InvokeID() string {
	return r.invokeID
}

// TraceID -
func (r *LogRecorder) TraceID() string {
	return r.traceID
}

// Stage -
func (r *LogRecorder) Stage() string {
	return r.stage
}

// Generator -
func (r *LogRecorder) Generator() LogHeaderFooterGenerator {
	return r.generator
}
