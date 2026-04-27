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
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/logservice"
)

func TestServer(t *testing.T) {
	Convey("Given a Server instance", t, func() {
		server := &Server{}

		// 测试数据
		registerReq := &logservice.RegisterRequest{
			CollectorID: "collector-1",
			Address:     "127.0.0.1:50051",
		}
		reportLogReq := &logservice.ReportLogRequest{
			Items: []*logservice.LogItem{
				{
					Filename:    "test.log",
					CollectorID: "collector-1",
					Target:      logservice.LogTarget_USER_STD,
					RuntimeID:   "runtime-1",
				},
			},
		}

		Convey("When handling a Register RPC request", func() {
			// 模拟 managerSingleton.RegisterLogCollector 成功
			patches := gomonkey.ApplyMethodReturn(managerSingleton, "RegisterLogCollector", nil)
			defer patches.Reset()

			resp, err := server.Register(context.Background(), registerReq)

			Convey("Then the request should be handled successfully", func() {
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
			})
		})

		Convey("When handling a Register RPC request with an error", func() {
			// 模拟 managerSingleton.RegisterLogCollector 返回错误
			patches := gomonkey.ApplyMethodReturn(managerSingleton, "RegisterLogCollector", errors.New("register failed"))
			defer patches.Reset()

			resp, err := server.Register(context.Background(), registerReq)

			Convey("Then the request should return an error", func() {
				So(err, ShouldNotBeNil)
				So(resp, ShouldBeNil)
			})
		})

		Convey("When handling a ReportLog RPC request", func() {
			// 模拟 managerSingleton.ReportLogItem
			patches := gomonkey.ApplyMethodFunc(managerSingleton, "ReportLogItem", func(item *logservice.LogItem) {})
			defer patches.Reset()

			resp, err := server.ReportLog(context.Background(), reportLogReq)

			Convey("Then the request should be handled successfully", func() {
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Code, ShouldEqual, 0)
				So(resp.Message, ShouldEqual, "Log reported successfully")
			})
		})

		Convey("When handling a ReportLog RPC request with empty items", func() {
			// 模拟 managerSingleton.ReportLogItem
			patches := gomonkey.ApplyMethodFunc(managerSingleton, "ReportLogItem", func(item *logservice.LogItem) {})
			defer patches.Reset()

			emptyReq := &logservice.ReportLogRequest{
				Items: []*logservice.LogItem{},
			}
			resp, err := server.ReportLog(context.Background(), emptyReq)

			Convey("Then the request should be handled successfully", func() {
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Code, ShouldEqual, 0)
				So(resp.Message, ShouldEqual, "Log reported successfully")
			})
		})
	})
}
