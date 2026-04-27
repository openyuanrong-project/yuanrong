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
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/logservice"
)

func TestLogEntry(t *testing.T) {
	Convey("Given a LogEntry instance", t, func() {
		// 正常场景
		Convey("When creating a new LogEntry with valid LogItem", func() {
			logItem := &logservice.LogItem{
				Filename:    "test.log",
				CollectorID: "collector-1",
				Target:      logservice.LogTarget_USER_STD,
				RuntimeID:   "runtime-1",
			}
			logEntry := NewLogEntry(logItem)

			Convey("Then the LogEntry should be initialized correctly", func() {
				So(logEntry.LogItem, ShouldEqual, logItem)
				So(logEntry.JobID, ShouldBeEmpty)
				So(logEntry.InstanceID, ShouldBeEmpty)
			})

			Convey("Then the ID method should return the correct key", func() {
				expectedID := "test.log//collector-1//runtime-1"
				So(logEntry.ID(), ShouldEqual, expectedID)
			})
		})
	})
}

func TestLogEntriesIntersection(t *testing.T) {
	Convey("Given multiple LogEntries maps", t, func() {
		// 正常场景
		Convey("When computing the intersection of multiple LogEntries", func() {
			entry1 := &LogEntry{
				LogItem: &logservice.LogItem{
					Filename:    "file1.log",
					CollectorID: "collector-1",
					RuntimeID:   "runtime-1",
				},
			}
			entry2 := &LogEntry{
				LogItem: &logservice.LogItem{
					Filename:    "file2.log",
					CollectorID: "collector-2",
					RuntimeID:   "runtime-2",
				},
			}
			entry3 := &LogEntry{
				LogItem: &logservice.LogItem{
					Filename:    "file1.log",
					CollectorID: "collector-1",
					RuntimeID:   "runtime-1",
				},
			}

			entries1 := NewLogEntries()
			entries1.Put(entry1, putOptionIfExistsReplace)
			entries1.Put(entry2, putOptionIfExistsReplace)
			entries2 := NewLogEntries()
			entries2.Put(entry1, putOptionIfExistsReplace)
			entries2.Put(entry3, putOptionIfExistsReplace)

			intersection := logEntriesIntersection(entries1, entries2)

			Convey("Then the intersection should contain only common Entries", func() {
				So(intersection.Len(), ShouldEqual, 1)
				So(intersection.Get(entry1.ID()), ShouldEqual, entry1)
			})
		})

		// 异常场景
		Convey("When computing the intersection with no input maps", func() {
			intersection := logEntriesIntersection()

			Convey("Then the intersection should be an empty map", func() {
				So(intersection.Len(), ShouldEqual, 0)
			})
		})

		Convey("When computing the intersection with one input map", func() {
			entry := &LogEntry{
				LogItem: &logservice.LogItem{
					Filename:    "file1.log",
					CollectorID: "collector-1",
					RuntimeID:   "runtime-1",
				},
			}
			entries := NewLogEntries()
			entries.Put(entry, putOptionIfExistsReplace)

			intersection := logEntriesIntersection(entries)

			Convey("Then the intersection should be the same as the input map", func() {
				So(intersection.Len(), ShouldEqual, 1)
				So(intersection.Get(entry.ID()), ShouldEqual, entry)
			})
		})

		Convey("When computing the intersection with no common Entries", func() {
			entry1 := &LogEntry{
				LogItem: &logservice.LogItem{
					Filename:    "file1.log",
					CollectorID: "collector-1",
					RuntimeID:   "runtime-1",
				},
			}
			entry2 := &LogEntry{
				LogItem: &logservice.LogItem{
					Filename:    "file2.log",
					CollectorID: "collector-2",
					RuntimeID:   "runtime-2",
				},
			}

			entries1 := NewLogEntries()
			entries1.Put(entry1, putOptionIfExistsReplace)
			entries2 := NewLogEntries()
			entries2.Put(entry2, putOptionIfExistsReplace)

			intersection := logEntriesIntersection(entries1, entries2)

			Convey("Then the intersection should be an empty map", func() {
				So(intersection.Len(), ShouldEqual, 0)
			})
		})
	})
}
