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
	"context"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"

	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/logservice"
	"yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/dashboard/etcdcache"
)

type mockLogCollectorClient struct {
}

func newMockLogCollectorClient() *mockLogCollectorClient {
	m := &mockLogCollectorClient{}
	return m
}

func (m *mockLogCollectorClient) StartLogStream(ctx context.Context,
	request *logservice.StartLogStreamRequest, opts ...grpc.CallOption) (*logservice.StartLogStreamResponse, error) {
	return nil, nil
}

func (m *mockLogCollectorClient) QueryLogStream(ctx context.Context,
	request *logservice.QueryLogStreamRequest, opts ...grpc.CallOption) (*logservice.QueryLogStreamResponse, error) {
	return nil, nil
}

func (m *mockLogCollectorClient) StopLogStream(ctx context.Context,
	request *logservice.StopLogStreamRequest, opts ...grpc.CallOption) (*logservice.StopLogStreamResponse, error) {
	return nil, nil
}

func (m *mockLogCollectorClient) Reset() {
}

func (m *mockLogCollectorClient) ReadLog(ctx context.Context, in *logservice.ReadLogRequest,
	opts ...grpc.CallOption) (logservice.LogCollectorService_ReadLogClient, error) {
	//m.ReadLogRequestCh <- in
	return nil, nil
}

func TestManager(t *testing.T) {
	Convey("Given a manager instance", t, func() {
		// 初始化 manager
		m := managerSingleton

		// 测试数据
		fakeLogCollectorClient := newMockLogCollectorClient()
		collectorInfo := collectorClientInfo{
			ID:      "collector-1",
			Address: "127.0.0.1:50051",
		}
		logItem := &logservice.LogItem{
			Filename:    "test.log",
			CollectorID: "collector-1",
			Target:      logservice.LogTarget_USER_STD,
			RuntimeID:   "runtime-1",
		}
		instance := &types.InstanceSpecification{
			InstanceID: "instance-1",
			JobID:      "job-1",
			RuntimeID:  "runtime-1",
		}

		Convey("When registering a log collector", func() {
			// 模拟 collectorClient.Connect 成功
			patches := gomonkey.ApplyMethodFunc(reflect.TypeOf(&grpc.ClientConn{}), "Connect", func() error {
				return nil
			})
			defer patches.Reset()

			err := m.RegisterLogCollector(collectorInfo)
			Convey("Then the collector should be registered successfully", func() {
				So(err, ShouldBeNil)
				c := m.GetCollector(collectorInfo.ID)
				So(c, ShouldNotBeNil)
				So(c.ID, ShouldEqual, collectorInfo.ID)
			})
		})

		Convey("When unregistering a log collector", func() {
			// 先注册一个 collector
			m.Collectors[collectorInfo.ID] = collectorClient{
				collectorClientInfo: collectorInfo,
			}

			// 模拟 grpcConn.Close 成功
			patches := gomonkey.ApplyMethodFunc(reflect.TypeOf(&grpc.ClientConn{}), "Close", func() error {
				return nil
			})
			defer patches.Reset()

			m.UnregisterLogCollector(collectorInfo.ID)

			Convey("Then the collector should be unregistered successfully", func() {
				So(m.GetCollector(collectorInfo.ID), ShouldBeNil)
			})
		})

		Convey("When getting a collector by ID", func() {
			// 先注册一个 collector
			m.Collectors[collectorInfo.ID] = collectorClient{
				collectorClientInfo: collectorInfo,
				logClient:           fakeLogCollectorClient,
			}

			collector := m.GetCollector(collectorInfo.ID)

			Convey("Then the collector should be returned", func() {
				So(collector, ShouldNotBeNil)
				So(collector.ID, ShouldEqual, collectorInfo.ID)
			})
		})

		Convey("When reporting a log item with empty RuntimeID", func() {
			logItem := &logservice.LogItem{
				Filename:    "test.log",
				CollectorID: "collector-1",
				Target:      logservice.LogTarget_USER_STD,
				RuntimeID:   "", // 空 RuntimeID
			}

			m.ReportLogItem(logItem)

			Convey("Then the log item should be added to the LogDB", func() {
				entry := m.LogDB.GetLogEntries().Get(logItem.Filename + "//" + logItem.CollectorID + "//" + logItem.RuntimeID)
				So(entry, ShouldNotBeNil)
				So(entry.LogItem, ShouldResemble, logItem)
			})
		})

		Convey("When reporting a log item with non-empty RuntimeID", func() {
			// 模拟 etcdcache.InstanceCache.GetByRuntimeID 返回实例
			patches := gomonkey.ApplyMethodReturn(&etcdcache.InstanceCache, "GetByRuntimeID", instance)
			defer patches.Reset()

			m.ReportLogItem(logItem)

			Convey("Then the log item should be added to the LogDB with JobID and InstanceID", func() {
				entry := m.LogDB.GetLogEntries().Get(logItem.Filename + "//" + logItem.CollectorID + "//" + logItem.RuntimeID)
				So(entry, ShouldNotBeNil)
				So(entry.JobID, ShouldEqual, instance.JobID)
				So(entry.InstanceID, ShouldEqual, instance.InstanceID)
			})
		})
	})
}
