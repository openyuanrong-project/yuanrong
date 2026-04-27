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

package logger

import (
	"errors"
	"io/fs"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/config"
)

type mockInfo struct {
	name  string
	isDir bool
	size  int64
}

func (m mockInfo) Name() string {
	return m.name
}

func (m mockInfo) IsDir() bool {
	return m.isDir
}

func (m mockInfo) Type() fs.FileMode {
	return 0
}

func (m mockInfo) Info() (fs.FileInfo, error) {
	return m, nil
}

func (m mockInfo) Size() int64 {
	return m.size
}

func (m mockInfo) Mode() fs.FileMode {
	return 0
}

func (m mockInfo) ModTime() time.Time {
	return time.Now()
}

func (m mockInfo) Sys() interface{} {
	return nil
}

func Test_initRollingLog(t *testing.T) {
	coreInfo := config.CoreInfo{
		FilePath: "./test-run.log",
	}
	defer gomonkey.ApplyFunc(os.ReadDir, func(string) ([]os.DirEntry, error) {
		return []os.DirEntry{
			mockInfo{name: "test-run.2006010215040507.log"},
			mockInfo{name: "test-run.2006010215040508.log"},
			mockInfo{name: "{funcName}@ABCabc@latest@pool22-300-128-fusion-85c55c66d7-zzj9x@{timeNow}#{logGroupID}#{logStreamID}#cff-log.log"},
		}, nil
	}).ApplyFunc(os.OpenFile, func(string, int, os.FileMode) (*os.File, error) {
		return nil, nil
	}).Reset()
	convey.Convey("init service log", t, func() {
		defer gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
			return mockInfo{name: strings.TrimPrefix(name, "./")}, nil
		}).Reset()
		log, err := initRollingLog(coreInfo, os.O_WRONLY|os.O_APPEND|os.O_CREATE, defaultPerm)
		convey.So(log, convey.ShouldNotBeNil)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("init user log", t, func() {
		defer gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
			return nil, &os.PathError{}
		}).Reset()
		coreInfo.FilePath = "{funcName}@ABCabc@latest@pool22-300-128-fusion-85c55c66d7-zzj9x@{timeNow}#{logGroupID}#{logStreamID}#cff-log.log"
		coreInfo.IsUserLog = true
		log, _ := initRollingLog(coreInfo, os.O_WRONLY|os.O_APPEND|os.O_CREATE, defaultPerm)
		convey.So(log, convey.ShouldNotBeNil)
		convey.So(GetLogName(coreInfo.FilePath), convey.ShouldNotBeEmpty)
	})
	convey.Convey("init wisecloud alarm log", t, func() {
		defer gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
			return nil, &os.PathError{}
		}).Reset()
		coreInfo.FilePath = "{funcName}@ABCabc@latest@pool22-300-128-fusion-85c55c66d7-zzj9x@{timeNow}#{logGroupID}#{logStreamID}#cff-log.log"
		coreInfo.IsWiseCloudAlarmLog = true
		log, _ := initRollingLog(coreInfo, os.O_WRONLY|os.O_APPEND|os.O_CREATE, defaultPerm)
		convey.So(log, convey.ShouldNotBeNil)
		convey.So(GetLogName(coreInfo.FilePath), convey.ShouldNotBeEmpty)
	})
}

func Test_rollingLog_Write(t *testing.T) {
	log := &rollingLog{}
	log.maxSize = 0
	log.isUserLog = true
	log.file = &os.File{}
	log.nameTemplate = "{funcName}@ABCabc@latest@pool22-300-128-fusion-85c55c66d7-zzj9x@{timeNow}#{logGroupID}#{logStreamID}#cff-log.log"
	convey.Convey("write rolling log", t, func() {
		convey.Convey("case1: failed to write rolling log", func() {
			defer gomonkey.ApplyMethod(reflect.TypeOf(log.file), "Write", func(f *os.File, b []byte) (n int, err error) {
				return len(b), nil
			}).ApplyMethod(reflect.TypeOf(log.file), "Stat", func(f *os.File) (info os.FileInfo, err error) {
				return mockInfo{size: 3}, nil
			}).ApplyMethod(reflect.TypeOf(log.file), "Sync", func(f *os.File) error {
				return nil
			}).Reset()
			n, err := log.Write([]byte("abc"))
			convey.So(n, convey.ShouldEqual, 3)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("case2: failed to write rolling log", func() {
			defer gomonkey.ApplyMethod(reflect.TypeOf(log.file), "Write", func(f *os.File, b []byte) (n int, err error) {
				return len(b), nil
			}).ApplyMethod(reflect.TypeOf(log.file), "Stat", func(f *os.File) (info os.FileInfo, err error) {
				return mockInfo{size: 3}, nil
			}).ApplyMethod(reflect.TypeOf(log.file), "Sync", func(f *os.File) error {
				return errors.New("test")
			}).Reset()
			n, err := log.Write([]byte("abc"))
			convey.So(n, convey.ShouldEqual, 3)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func Test_rollingLog_cleanRedundantSinks(t *testing.T) {
	log := &rollingLog{}
	log.maxBackups = 0
	tn := time.Now().String()
	os.Create("test_log_1#" + tn)
	os.Create("test_log_2#" + tn)
	log.sinks = []string{"test_log_1#" + tn, "test_log_2#" + tn}
	convey.Convey("rollingLog_cleanRedundantSinks", t, func() {
		log.cleanRedundantSinks()
		time.Sleep(50 * time.Millisecond)
		convey.So(isAvailable("test_log_1#"+tn), convey.ShouldEqual, false)
		convey.So(isAvailable("test_log_2#"+tn), convey.ShouldEqual, false)
	})
}
