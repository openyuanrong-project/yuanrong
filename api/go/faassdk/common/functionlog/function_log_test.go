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
	"os"
	"strings"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/runtime/faassdk/config"
	"yuanrong.org/kernel/runtime/faassdk/utils/urnutils"
)

func TestGetFunctionLogger(t *testing.T) {
	_, err := GetFunctionLogger(&config.Configuration{StartArgs: config.StartArgs{
		LogDir: "/home/sn/logs",
	}})
	assert.Nil(t, err)
}

func TestRefreshFileModTime(t *testing.T) {
	convey.Convey("Test RefreshFileModTime", t, func() {
		convey.Convey("Test 1", func() {
			f := &FunctionLogger{
				logAbsPath: "patch",
			}
			ch := make(chan struct{})
			close(ch)
			f.RefreshFileModTime(ch)

			p := gomonkey.NewPatches()
			p.ApplyFunc(time.NewTicker, func(d time.Duration) *time.Ticker {
				ch := make(chan time.Time, 1)
				ch <- time.Time{}
				return &time.Ticker{C: ch}
			})
			defer p.Reset()
			f.RefreshFileModTime(ch)
			f.refreshWithRetry()
		})
	})
	convey.Convey("Test RefreshFileModTime err", t, func() {
		convey.Convey("Test 2", func() {
			var stopCh chan struct{}
			f := &FunctionLogger{}
			f.RefreshFileModTime(stopCh)
		})
	})
}

func TestExtractLogTankService(t *testing.T) {
	l := &logTankService{}
	l.extractLogTankService("groupID;", "streamID;")
	l.extractLogTankService("ID", "ID")
	convey.Convey("Test extractLogTankService", t, func() {
		convey.Convey("extractLogTankService success", func() {
			convey.So(func() {
				l.extractLogTankService("groupID1;groupID2;groupID3", "streamID;")
			}, convey.ShouldNotPanic)
			convey.So(func() {
				l.extractLogTankService("groupID;", "streamID;")
			}, convey.ShouldNotPanic)
			convey.So(func() {
				l.extractLogTankService("ID", "ID")
			}, convey.ShouldNotPanic)
		})
	})
}

func Test_refreshWithRetry(t *testing.T) {
	convey.Convey("Test refreshWithRetry", t, func() {
		convey.Convey("Test 1", func() {
			var testMsg string
			patches := []*gomonkey.Patches{
				gomonkey.ApplyFunc(os.Chtimes, func(name string, atime time.Time, mtime time.Time) error {
					testMsg = "os Chtimes return nil"
					return nil
				}),
			}
			defer func() {
				for idx := range patches {
					patches[idx].Reset()
				}
			}()
			f := &FunctionLogger{}
			f.refreshWithRetry()
			convey.So(testMsg, convey.ShouldEqual, "os Chtimes return nil")
		})
	})
}

func TestSync(t *testing.T) {
	Sync(nil)
}

func TestGetDefaultInternalFunctionLog(t *testing.T) {
	convey.Convey("Test getDefaultInternalFunctionLog", t, func() {
		convey.Convey("getDefaultInternalFunctionLog success", func() {
			fl := &FunctionLog{}
			ifl := getDefaultInternalFunctionLog(fl)
			convey.So(ifl, convey.ShouldNotBeNil)
		})
	})
}

func TestNewLTSFunctionLogger(t *testing.T) {
	convey.Convey(
		"Test newLTSFunctionLogger", t, func() {
			convey.Convey(
				"newLTSFunctionLogger success", func() {
					cfg := &config.Configuration{StartArgs: config.StartArgs{
						LogDir: "",
					}}
					fl, err := newLTSFunctionLogger(cfg)
					convey.So(fl, convey.ShouldBeNil)
					convey.So(err, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestWriteStdLog(t *testing.T) {
	convey.Convey(
		"Test WriteStdLog", t, func() {
			convey.Convey(
				"WriteStdLog success", func() {
					convey.So(func() {
						functionLogger.WriteStdLog("logString", "timeStamp", false, "nanoTimestamp")
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestWriteDefaultLog(t *testing.T) {
	convey.Convey(
		"Test WriteDefaultLog", t, func() {
			convey.Convey(
				"WriteDefaultLog success", func() {
					convey.So(func() {
						fl := functionLogger.AcquireLog()
						functionLogger.WriteDefaultLog(fl)
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestNewZapLogger(t *testing.T) {
	convey.Convey(
		"Test newZapLogger", t, func() {
			convey.Convey(
				"newZapLogger success", func() {
					defer cleanFile("fullPath")
					zl, err := newZapLogger("fullPath", "messageKey")
					convey.So(zl, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestCreateFunctionInfoFile(t *testing.T) {
	convey.Convey(
		"Test createFunctionInfoFile", t, func() {
			convey.Convey(
				"createFunctionInfoFile success", func() {
					defer cleanFile("fullPath")
					err := createFunctionInfoFile(&urnutils.ComplexFuncName{})
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func cleanFile(fileName string) {
	files, _ := os.ReadDir("./")
	for _, file := range files {
		flag := strings.HasPrefix(file.Name(), fileName)
		if flag {
			os.Remove(file.Name())
		}
	}
}
