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
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"yuanrong.org/kernel/runtime/faassdk/common/constants"
	"yuanrong.org/kernel/runtime/faassdk/config"
	runtimeLogger "yuanrong.org/kernel/runtime/libruntime/common/logger"
	runtimeLoggerConfig "yuanrong.org/kernel/runtime/libruntime/common/logger/config"
)

const logDir = "/home/sn/log"

func TestMain(m *testing.M) {
	p := gomonkey.NewPatches()
	p.ApplyFunc(runtimeLoggerConfig.GetCoreInfoFromEnv, func() (runtimeLoggerConfig.CoreInfo, error) { return runtimeLoggerConfig.GetDefaultCoreInfo(), nil })
	p.ApplyFunc(os.OpenFile, func(name string, flag int, perm os.FileMode) (*os.File, error) {
		return &os.File{}, nil
	})
	p.ApplyFunc(runtimeLogger.CreateSink, func(coreInfo runtimeLoggerConfig.CoreInfo) (io.Writer, error) {
		return io.Discard, nil
	})

	GetFunctionLogger(&config.Configuration{StartArgs: config.StartArgs{
		LogDir: logDir,
	}})

	p.Reset()
	m.Run()
	createLoggerOnce = sync.Once{}
	os.RemoveAll(logDir)
}

func TestLogRecorder_Fusion(t *testing.T) {
	functionLogger, err := GetFunctionLogger(&config.Configuration{
		StartArgs: config.StartArgs{LogDir: logDir},
	})
	if err != nil {
		t.Fatalf("failed to get function functionLogger, capability:fusion, %s", err.Error())
	}
	functionLogger.SetLogLevel(constants.FuncLogLevelWarn)
	invokeID := "1"
	traceID := "traceID"
	logRecorder := functionLogger.NewLogRecorder(invokeID, traceID, constants.InvokeStage, WithoutLineBreak(), WithLogOptionTail())
	if functionLogger.GetLogRecorder(invokeID) == nil {
		t.Error("failed to get log recorder by invokeID")
	}
	logRecorder.Write(&FunctionLog{
		Level:   "INFO",
		Message: bytes.NewBufferString("message"),
	})

	functionLogger.WriteStdLog("std logs", "my time", false, "time")
	logRecorder.logs.Reset()
	logRecorder.generator = &fakeGenerator{}
	logRecorder.getFooter()
	logRecorder.header = &FunctionLog{}
	logRecorder.footer = &FunctionLog{}
	logRecorder.Finish()
}

func TestGetLTSLoggerMessage(m *testing.T) {
	functionLog := &internalFunctionLog{FunctionLog: &FunctionLog{}}
	functionLog.Index = -1
	functionLog.Time = ""
	functionLog.TraceID = ""
	functionLog.Message = new(bytes.Buffer)
	msg := getLTSLoggerMessage(functionLog)
	assert.Equal(m, len(msg), 0)

	functionLog = &internalFunctionLog{FunctionLog: &FunctionLog{}}
	functionLog.Index = 1
	functionLog.Time = "test time"
	functionLog.TraceID = "test 1"
	functionLog.Message = bytes.NewBufferString("time test message")
	msg = getLTSLoggerMessage(functionLog)
	assert.NotEqual(m, len(msg), 0)

	functionLog = &internalFunctionLog{FunctionLog: &FunctionLog{}}
	functionLog.Index = constants.DefaultFuncLogIndex
	functionLog.Time = "test time"
	functionLog.TraceID = "test 2"
	functionLog.Message = bytes.NewBufferString("time test message")
	str := functionLog.Time + " " + "time test message"
	msg = getLTSLoggerMessage(functionLog)
	assert.Equal(m, str, msg)
}

func TestFormatLog(m *testing.T) {
	r := &LogRecorder{}
	fl := &FunctionLog{
		Time:    "time",
		Level:   "WARNING",
		Message: new(bytes.Buffer),
	}
	t := r.formatLog(fl)
	assert.Equal(m, "time  WARN ", string(t))
	fl.IsStdLog = true
	t = r.formatLog(fl)
	assert.Equal(m, "", string(t))
}

func TestMarshalAll(t *testing.T) {
	convey.Convey("Test MarshalAll", t, func() {
		convey.Convey("Test no log", func() {
			r := &LogRecorder{
				f:                      &FunctionLogger{},
				separatedWithLineBreak: true,
			}

			limitSize := maxTailLogSizeKB
			droppedNum := 0
			r.logs = NewFixSizeRecorder(
				limitSize,
				r.guessSize,
				func(funcLog *FunctionLog, logTooLarge bool) {
					droppedNum++
					r.handleDropped(funcLog, logTooLarge)
				},
			)

			r.MarshalAll()

			var buffer bytes.Buffer
			r.logs.Range(func(piece *FunctionLog) {
				buffer.Write(r.formatLog(piece))
				if r.separatedWithLineBreak {
					buffer.WriteString("\n")
				}
			})

			assert.Equal(t, 0, droppedNum)
			assert.Equal(t, 0, len(buffer.Bytes()))
			assert.Equal(t, 0, r.idx)

			r.Finish()

			assert.Equal(t, 1, r.idx)
		})
		convey.Convey("Test too many logs", func() {
			r := &LogRecorder{
				f:                      &FunctionLogger{},
				separatedWithLineBreak: true,
			}

			limitSize := maxTailLogSizeKB
			droppedNum := 0
			r.logs = NewFixSizeRecorder(
				limitSize,
				r.guessSize,
				func(funcLog *FunctionLog, logTooLarge bool) {
					droppedNum++
					r.handleDropped(funcLog, logTooLarge)
				},
			)

			writeNum := 150
			for idx := 0; idx < writeNum; idx++ {
				r.Write(&FunctionLog{
					Message: bytes.NewBufferString("Function log test"),
				})
			}
			r.MarshalAll()

			var buffer bytes.Buffer
			r.logs.Range(func(piece *FunctionLog) {
				buffer.Write(r.formatLog(piece))
				if r.separatedWithLineBreak {
					buffer.WriteString("\n")
				}
			})

			assert.True(t, droppedNum > 0)
			assert.True(t, len(buffer.Bytes()) > limitSize)
			assert.Equal(t, droppedNum+1, r.idx)

			r.Finish()

			assert.Equal(t, writeNum+1, r.idx)
		})
		convey.Convey("Test logs", func() {
			r := &LogRecorder{
				f:                      &FunctionLogger{},
				separatedWithLineBreak: true,
			}

			limitSize := maxTailLogSizeKB
			droppedNum := 0
			r.logs = NewFixSizeRecorder(
				limitSize,
				r.guessSize,
				func(funcLog *FunctionLog, logTooLarge bool) {
					droppedNum++
					r.handleDropped(funcLog, logTooLarge)
				},
			)

			writeNum := 20
			for idx := 0; idx < writeNum; idx++ {
				r.Write(&FunctionLog{
					Message: bytes.NewBufferString("Function log test"),
				})
			}
			r.MarshalAll()

			var buffer bytes.Buffer
			r.logs.Range(func(piece *FunctionLog) {
				buffer.Write(r.formatLog(piece))
				if r.separatedWithLineBreak {
					buffer.WriteString("\n")
				}
			})

			assert.Equal(t, 0, droppedNum)
			assert.True(t, len(buffer.Bytes()) < limitSize)
			assert.Equal(t, 0, r.idx)

			r.Finish()

			assert.Equal(t, writeNum+1, r.idx)
		})
	})
}

func TestMarshalAll2(t *testing.T) {
	convey.Convey("Test logTooLarge", t, func() {
		r := &LogRecorder{
			f:                      &FunctionLogger{},
			separatedWithLineBreak: true,
		}

		limitSize := maxTailLogSizeKB
		droppedNum := 0
		r.logs = NewFixSizeRecorder(
			limitSize,
			r.guessSize,
			func(funcLog *FunctionLog, logTooLarge bool) {
				droppedNum++
				r.handleDropped(funcLog, logTooLarge)
			},
		)

		msg := strings.Repeat("a", maxTailLogSizeKB+1)
		r.Write(&FunctionLog{
			Message: bytes.NewBufferString(msg),
		})
		r.Write(&FunctionLog{
			Message: bytes.NewBufferString("Function log test"),
		})
		r.MarshalAll()

		assert.NotNil(t, r.base)

		var buffer bytes.Buffer
		r.logs.Range(func(piece *FunctionLog) {
			buffer.Write(r.formatLog(piece))
			if r.separatedWithLineBreak {
				buffer.WriteString("\n")
			}
		})

		assert.Equal(t, 1, droppedNum)
		assert.True(t, len(buffer.Bytes()) != 0)
		assert.True(t, len(buffer.Bytes()) < limitSize)
		assert.Equal(t, 2, r.idx)

		r.Finish()

		assert.Equal(t, 3, r.idx)
	})
	convey.Convey("Test logTooLarge then drop", t, func() {
		r := &LogRecorder{
			f:                      &FunctionLogger{},
			separatedWithLineBreak: true,
		}

		limitSize := maxTailLogSizeKB
		droppedNum := 0
		r.logs = NewFixSizeRecorder(
			limitSize,
			r.guessSize,
			func(funcLog *FunctionLog, logTooLarge bool) {
				droppedNum++
				r.handleDropped(funcLog, logTooLarge)
			},
		)

		msg := strings.Repeat("a", maxTailLogSizeKB+1)
		r.Write(&FunctionLog{
			Message: bytes.NewBufferString(msg),
		})

		writeNum := 150
		for idx := 0; idx < writeNum; idx++ {
			r.Write(&FunctionLog{
				Message: bytes.NewBufferString("Function log test"),
			})
		}
		r.MarshalAll()

		assert.Nil(t, r.base)

		var buffer bytes.Buffer
		r.logs.Range(func(piece *FunctionLog) {
			buffer.Write(r.formatLog(piece))
			if r.separatedWithLineBreak {
				buffer.WriteString("\n")
			}
		})

		assert.True(t, droppedNum > 0)
		assert.True(t, len(buffer.Bytes()) > limitSize)
		assert.Equal(t, droppedNum+1, r.idx)

		r.Finish()

		assert.Equal(t, writeNum+2, r.idx)

		r.Finish()
	})
	convey.Convey("Test drop then logTooLarge", t, func() {
		r := &LogRecorder{
			f:                      &FunctionLogger{},
			separatedWithLineBreak: true,
		}

		limitSize := maxTailLogSizeKB
		droppedNum := 0
		r.logs = NewFixSizeRecorder(
			limitSize,
			r.guessSize,
			func(funcLog *FunctionLog, logTooLarge bool) {
				droppedNum++
				r.handleDropped(funcLog, logTooLarge)
			},
		)

		writeNum := 150
		for idx := 0; idx < writeNum; idx++ {
			r.Write(&FunctionLog{
				Message: bytes.NewBufferString("Function log test"),
			})
		}
		msg := strings.Repeat("a", maxTailLogSizeKB+1)
		r.Write(&FunctionLog{
			Message: bytes.NewBufferString(msg),
		})
		r.MarshalAll()

		assert.NotNil(t, r.base)

		var buffer bytes.Buffer
		r.logs.Range(func(piece *FunctionLog) {
			buffer.Write(r.formatLog(piece))
			if r.separatedWithLineBreak {
				buffer.WriteString("\n")
			}
		})

		assert.True(t, droppedNum > 0)
		assert.Equal(t, 0, len(buffer.Bytes()))
		assert.Equal(t, droppedNum+1, r.idx)

		r.Finish()

		assert.Equal(t, droppedNum+1, r.idx)
	})
}

type fakeGenerator struct {
	header *FunctionLog
}

func (f *fakeGenerator) GenerateLogHeader(header *FunctionLog) { f.header = header }
func (f *fakeGenerator) GenerateLogFooter(*FunctionLog)        {}

func TestLogPool(t *testing.T) {
	generator := &fakeGenerator{}
	r := &LogRecorder{
		f:                      NewFunctionLogger(zap.NewNop(), "", ""),
		separatedWithLineBreak: true,
		generator:              generator,
	}
	r.handleDropped(r.f.AcquireLog(), false)
	assert.True(t, r.header == generator.header)
	assert.True(t, r.f.AcquireLog() != r.header) // header should not be released
}

func TestFunctionLogDeepCopy(t *testing.T) {
	l := NewFunctionLog()
	l.Message.WriteString("abc")
	l.ErrorType = "xxx"
	l.IsFinishedLog = true
	l.IsStdLog = true
	l.Level = "info"

	x := l.DeepCopy()
	assert.True(t, l != x)
	assert.Equal(t, l, x)
}

func TestFixSizeRecorder_WriteLimit(t *testing.T) {
	r := &LogRecorder{}
	case2Logs := list.New()
	case2Logs.PushBack(&FunctionLog{
		Level:   "INFO",
		Message: bytes.NewBufferString("abc"),
	})
	type fields struct {
		logs      *list.List
		guessSize func(funcLog *FunctionLog) int
		dropped   func(funcLog *FunctionLog, logTooLarge bool)
		size      int
		limitSize int
	}
	type args struct {
		funcLog   *FunctionLog
		limitSize int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"case1 write less than limitSize", fields{logs: nil, guessSize: r.guessSize, limitSize: 20},
			args{funcLog: &FunctionLog{
				Level:   "INFO",
				Message: bytes.NewBufferString("abc"),
			}, limitSize: 10}},
		{"case2 write more than limitSize", fields{logs: case2Logs, guessSize: r.guessSize, size: 20,
			limitSize: 20}, args{funcLog: &FunctionLog{
			Level:   "INFO",
			Message: bytes.NewBufferString("abc"),
		}, limitSize: 10}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &FixSizeRecorder{
				logs:      tt.fields.logs,
				guessSize: tt.fields.guessSize,
				dropped:   tt.fields.dropped,
				size:      tt.fields.size,
				limitSize: tt.fields.limitSize,
			}
			r.WriteLimit(tt.args.funcLog, tt.args.limitSize)
			if r.size > 20 {
				t.Errorf("write limit error, size is: %d", r.size)
			}
		})

	}
}

func TestLogRecorder_WriteLimit(t *testing.T) {
	functionLog := &FunctionLog{
		Level:   "INFO",
		Message: bytes.NewBufferString(strings.Repeat("a", 90*1024)),
	}
	r := NewLogRecorder()
	for i := 0; i < 256; i++ {
		r.WriteLimit(functionLog)
	}
	if r.logs.size > maxInitLogSizeMB+90*1024*2 {
		t.Errorf("write limit error, size is: %d", r.logs.size)
	}
}

func TestReset(t *testing.T) {
	convey.Convey("Test Reset", t, func() {
		convey.Convey("Reset success", func() {
			r := &FixSizeRecorder{}
			convey.So(r.Reset, convey.ShouldNotPanic)
		})
	})
}

func TestRealTimeWrite(t *testing.T) {
	convey.Convey("Test RealTimeWrite", t, func() {
		convey.Convey("RealTimeWrite success", func() {
			functionLog := &FunctionLog{
				Level:   "INFO",
				Message: bytes.NewBufferString("a"),
			}
			r := NewLogRecorder()
			r.logs.size = 2049
			convey.So(func() {
				r.RealTimeWrite(functionLog)
			}, convey.ShouldPanic)
		})
	})
}

func TestStartSync(t *testing.T) {
	convey.Convey(
		"Test StartSync", t, func() {
			convey.Convey(
				"StartSync success", func() {
					r := &LogRecorder{}
					convey.So(r.StartSync, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestFinishSync(t *testing.T) {
	convey.Convey(
		"Test FinishSync", t, func() {
			convey.Convey(
				"FinishSync success", func() {
					r := &LogRecorder{}
					r.StartSync()
					convey.So(r.FinishSync, convey.ShouldNotPanic)
					convey.So(r.FinishSync, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestGetLogRecorderField(t *testing.T) {
	convey.Convey(
		"Test GetLogRecorderField", t, func() {
			r := &LogRecorder{}
			var str string
			convey.Convey(
				"LogOption success", func() {
					str = r.LogOption()
					convey.So(str, convey.ShouldBeEmpty)
				},
			)
			convey.Convey(
				"InvokeID success", func() {
					str = r.InvokeID()
					convey.So(str, convey.ShouldBeEmpty)
				},
			)
			convey.Convey(
				"TraceID success", func() {
					str = r.TraceID()
					convey.So(str, convey.ShouldBeEmpty)
				},
			)
			convey.Convey(
				"Stage success", func() {
					str = r.Stage()
					convey.So(str, convey.ShouldBeEmpty)
				},
			)
			convey.Convey(
				"Generator success", func() {
					generator := r.Generator()
					convey.So(generator, convey.ShouldBeEmpty)
				},
			)
		},
	)
}
