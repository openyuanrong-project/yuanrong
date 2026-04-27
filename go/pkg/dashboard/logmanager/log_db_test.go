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

func TestGeneralLogDBImpl(t *testing.T) {
	Convey("Given a generalLogDBImpl instance", t, func() {
		db := newGeneralLogDBImpl()

		// 测试数据
		logItem := &logservice.LogItem{
			Filename:    "test.log",
			CollectorID: "collector-1",
			Target:      logservice.LogTarget_USER_STD,
			RuntimeID:   "runtime-1",
		}
		logEntry := &LogEntry{
			LogItem:    logItem,
			JobID:      "job-1",
			InstanceID: "instance-1",
		}

		Convey("When putting a LogEntry into the DB", func() {
			db.Put(logEntry, putOptionIfExistsReplace)
			Convey("Then the LogEntry should be added to the Entries and indexes", func() {

				So(db.Entries.Get(logEntry.ID()).Equal(logEntry), ShouldBeTrue)
				So(db.Query(logDBQuery{InstanceID: "instance-1"}).Get(logEntry.ID()), ShouldEqual, logEntry)
				So(db.Query(logDBQuery{SubmissionID: "job-1"}).Get(logEntry.ID()), ShouldEqual, logEntry)
				So(db.Query(logDBQuery{Filename: "test.log"}).Get(logEntry.ID()), ShouldEqual, logEntry)
			})
		})

		Convey("When removing a LogEntry from the DB", func() {
			db.Put(logEntry, putOptionIfExistsReplace) // 先添加
			db.Remove(logEntry)

			Convey("Then the LogEntry should be removed from the Entries and indexes", func() {
				So(db.Entries.Get(logEntry.ID()), ShouldBeNil)
				So(db.Query(logDBQuery{InstanceID: "instance-1"}).Get(logEntry.ID()), ShouldNotEqual, logEntry)
				So(db.Query(logDBQuery{SubmissionID: "job-1"}).Get(logEntry.ID()), ShouldNotEqual, logEntry)
				So(db.Query(logDBQuery{Filename: "test.log"}).Get(logEntry.ID()), ShouldNotEqual, logEntry)
			})
		})

		Convey("When getting all LogEntries", func() {
			db.Put(logEntry, putOptionIfExistsReplace)
			entries := db.GetLogEntries()

			Convey("Then the returned Entries should match the DB's Entries", func() {
				So(entries.Get(logEntry.ID()), ShouldEqual, logEntry)
			})
		})

		Convey("When serializing && deserializing the DB", func() {
			db.Put(logEntry, putOptionIfExistsReplace)
			log.GetLogger().Infof("Entries: %#v", db.Entries)

			data, err := db.Serialize()
			log.GetLogger().Infof("Entries: %s", data)
			So(err, ShouldBeNil)

			newDB := newGeneralLogDBImpl()
			err = newDB.Deserialize(data)
			So(err, ShouldBeNil)

			So(newDB.Query(logDBQuery{InstanceID: "instance-1"}).Get(logEntry.ID()).Equal(logEntry), ShouldBeTrue)
			So(newDB.Query(logDBQuery{SubmissionID: "job-1"}).Get(logEntry.ID()).Equal(logEntry), ShouldBeTrue)
			So(newDB.Query(logDBQuery{Filename: "test.log"}).Get(logEntry.ID()).Equal(logEntry), ShouldBeTrue)
		})

		Convey("When querying the DB with a valid query", func() {
			db.Put(logEntry, putOptionIfExistsReplace)
			query := logDBQuery{
				Filename:     "test.log",
				InstanceID:   "instance-1",
				SubmissionID: "job-1",
			}
			result := db.Query(query)
			Convey("Then the result should contain the matching LogEntry", func() {
				So(result.Get(logEntry.ID()).Equal(logEntry), ShouldBeTrue)
			})
		})

		Convey("When querying the DB with an invalid query", func() {
			db.Put(logEntry, putOptionIfExistsReplace)
			query := logDBQuery{
				Filename:     "non-existent.log",
				InstanceID:   "non-existent-instance",
				SubmissionID: "non-existent-job",
			}
			result := db.Query(query)

			log.GetLogger().Infof("result: %v", result)
			Convey("Then the result should be an empty LogEntries map", func() {
				So(result.Len(), ShouldEqual, 0)
			})
		})

		Convey("When putting a LogEntry concurrently", func() {
			var wg sync.WaitGroup
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					db.Put(logEntry, putOptionIfExistsReplace)
				}()
			}
			wg.Wait()

			Convey("Then the LogEntry should be added correctly without race conditions", func() {
				So(db.Entries.Get(logEntry.ID()).Equal(logEntry), ShouldBeTrue)

				So(db.Query(logDBQuery{InstanceID: "instance-1"}).Get(logEntry.ID()).Equal(logEntry), ShouldBeTrue)
				So(db.Query(logDBQuery{SubmissionID: "job-1"}).Get(logEntry.ID()).Equal(logEntry), ShouldBeTrue)
				So(db.Query(logDBQuery{Filename: "test.log"}).Get(logEntry.ID()).Equal(logEntry), ShouldBeTrue)
			})
		})

		Convey("When removing a LogEntry concurrently", func() {
			db.Put(logEntry, putOptionIfExistsReplace) // 先添加
			var wg sync.WaitGroup
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					db.Remove(logEntry)
				}()
			}
			wg.Wait()

			Convey("Then the LogEntry should be removed correctly without race conditions", func() {
				So(db.Entries.Get(logEntry.ID()), ShouldBeNil)
				So(db.Query(logDBQuery{InstanceID: "instance-1"}).Get(logEntry.ID()), ShouldNotEqual, logEntry)
				So(db.Query(logDBQuery{SubmissionID: "job-1"}).Get(logEntry.ID()), ShouldNotEqual, logEntry)
				So(db.Query(logDBQuery{Filename: "test.log"}).Get(logEntry.ID()), ShouldNotEqual, logEntry)
			})
		})
	})
}
