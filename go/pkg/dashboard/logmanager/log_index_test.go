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

package logmanager

import (
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/logservice"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

func TestLogIndex(t *testing.T) {
	Convey("Given a LogIndex instance", t, func() {
		// 初始化 LogIndex
		logIndex := NewLogIndex(func(entry *LogEntry) string { return entry.CollectorID },
			func(query logDBQuery) string { return query.CollectorID })
		// 测试数据
		logItem := &logservice.LogItem{
			Filename:    "test.log",
			CollectorID: "collector-1",
			Target:      logservice.LogTarget_USER_STD,
			RuntimeID:   "runtime-1",
		}
		logEntry := NewLogEntry(logItem)

		Convey("When putting a LogEntry into the index", func() {
			logIndex.Put(logEntry)

			Convey("Then the LogEntry should be added to the correct index keys", func() {
				So(logIndex.index["collector-1"].Get(logEntry.ID()), ShouldEqual, logEntry)
			})
		})

		Convey("When removing a LogEntry from the index", func() {
			logIndex.Put(logEntry) // 先添加
			logIndex.Remove(logEntry)

			Convey("Then the LogEntry should be removed from the index keys", func() {
				So(logIndex.index["collector-1"].Get(logEntry.ID()), ShouldBeNil)
			})
		})

		Convey("When querying the index with a valid key", func() {
			logIndex.Put(logEntry)
			result := logIndex.Query("collector-1")

			Convey("Then the result should contain the LogEntry", func() {
				So(result.Get(logEntry.ID()), ShouldEqual, logEntry)
			})
		})

		Convey("When querying the index with an invalid key", func() {
			result := logIndex.Query("invalid-key")

			Convey("Then the result should be an empty LogEntries map", func() {
				So(result.Len(), ShouldBeZeroValue)
			})
		})

		Convey("When putting a LogEntry with nil LogItem", func() {
			nilLogEntry := NewLogEntry(nil)
			logIndex.Put(nilLogEntry)

			Convey("Then the LogEntry should not be added to the index", func() {
				So(logIndex.index[""], ShouldBeNil)
			})
		})

		Convey("When removing a LogEntry that does not exist", func() {
			lenBefore := len(logIndex.index)
			log.GetLogger().Infof("before, should be %d", lenBefore)
			nonExistentEntry := &LogEntry{
				LogItem: &logservice.LogItem{
					Filename:    "non-existent.log",
					CollectorID: "non-existent-collector",
					RuntimeID:   "non-existent-runtime",
				},
			}
			logIndex.Remove(nonExistentEntry)

			Convey("Then the index should remain unchanged", func() {
				So(len(logIndex.index), ShouldEqual, lenBefore)
			})
		})

		Convey("When putting a LogEntry concurrently", func() {
			var wg sync.WaitGroup
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					logIndex.Put(logEntry)
				}()
			}
			wg.Wait()

			Convey("Then the LogEntry should be added correctly without race conditions", func() {
				So(logIndex.index["collector-1"].Get(logEntry.ID()), ShouldEqual, logEntry)
			})
		})

		Convey("When removing a LogEntry concurrently", func() {
			logIndex.Put(logEntry) // 先添加
			var wg sync.WaitGroup
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					logIndex.Remove(logEntry)
				}()
			}
			wg.Wait()

			Convey("Then the LogEntry should be removed correctly without race conditions", func() {
				So(logIndex.index["collector-1"].Get(logEntry.ID()), ShouldBeNil)
			})
		})
	})
}
