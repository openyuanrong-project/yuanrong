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
	"io"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/logservice"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

func TestMockNewClient(t *testing.T) {
	Convey("Given a mock NewClient function", t, func() {
		patches := gomonkey.ApplyFunc(grpc.NewClient, func(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
			return nil, errors.New("NewClient failed")
		})
		defer patches.Reset()

		_, err := grpc.NewClient("127.0.0.1:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
		So(err, ShouldNotBeNil)
	})
}

func TestCollectorClient_Connect(t *testing.T) {
	Convey("Given a collectorClient instance", t, func() {
		client := &collectorClient{
			collectorClientInfo: collectorClientInfo{
				ID:      "collector-1",
				Address: "127.0.0.1:50051",
			},
		}

		Convey("When connecting to the collector successfully", func() {
			// 模拟 grpc.NewClient 成功
			patches := gomonkey.ApplyFunc(grpc.NewClient, func(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
				return &grpc.ClientConn{}, nil
			})
			defer patches.Reset()

			err := client.Connect()

			Convey("Then the connection should be established", func() {
				So(err, ShouldBeNil)
				So(client.grpcConn, ShouldNotBeNil)
				So(client.logClient, ShouldNotBeNil)
			})
		})

		Convey("When connecting to the collector fails", func() {
			// 模拟 grpc.NewClient 失败
			patches := gomonkey.ApplyFunc(grpc.NewClient, func(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
				return nil, errors.New("NewClient failed")
			})
			defer patches.Reset()
			err := client.Connect()

			Convey("Then the connection should fail", func() {
				So(err, ShouldNotBeNil)
				So(client.grpcConn, ShouldBeNil)
				So(client.logClient, ShouldBeNil)
			})
		})
	})
}

func TestCollectorClient_Healthcheck(t *testing.T) {
	Convey("Given a collectorClient instance", t, func() {
		client := &collectorClient{
			collectorClientInfo: collectorClientInfo{
				ID:      "collector-1",
				Address: "127.0.0.1:50051",
			},
			grpcConn: &grpc.ClientConn{},
		}

		Convey("When the connection state is Shutdown", func() {
			// 模拟 grpcConn.GetState 返回 Shutdown
			patches := gomonkey.ApplyMethodFunc(reflect.TypeOf(&grpc.ClientConn{}), "GetState",
				func() connectivity.State {
					log.GetLogger().Infof("calling get state...")
					return connectivity.Shutdown
				})
			defer patches.Reset()

			// 用于记录 shutdownCallback 是否被调用
			shutdownCalled := make(chan string, 1)
			shutdownCallback := func() {
				shutdownCalled <- "closed"
			}

			// 启动 Healthcheck
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				client.Healthcheck(shutdownCallback)
			}()
			// 等待 Healthcheck 检测到 Shutdown
			select {
			case req := <-shutdownCalled:
				So(req, ShouldEqual, "closed")
			case <-time.After(15 * time.Second):
				t.Fatalf("not receive any start log stream request in 15 seconds")
			}

			// 停止 Healthcheck
			wg.Wait()
		})
	})
}

func TestCollectorClient_CollectLog(t *testing.T) {
	Convey("Given a collectorClient instance", t, func() {
		client := &collectorClient{
			collectorClientInfo: collectorClientInfo{
				ID:      "collector-1",
				Address: "127.0.0.1:50051",
			},
			grpcConn:  &grpc.ClientConn{},
			logClient: &mockLogCollectorClient{},
		}

		Convey("When collecting logs successfully", func() {
			// 模拟 logClient.ReadLog 成功
			patches := gomonkey.ApplyMethodFunc(&mockLogCollectorClient{}, "ReadLog", func(ctx context.Context, req *logservice.ReadLogRequest, opts ...grpc.CallOption) (logservice.LogCollectorService_ReadLogClient, error) {
				return &fakeReadLogClient{}, nil
			})
			defer patches.Reset()

			outLog := make(chan *logservice.ReadLogResponse, 1)
			err := client.CollectLog(context.Background(), &logservice.ReadLogRequest{
				Item: &logservice.LogItem{
					Filename:    "target-1",
					CollectorID: "collector-1",
					Target:      logservice.LogTarget_USER_STD,
					RuntimeID:   "runtime-1",
				},
				StartLine: 0,
				EndLine:   0,
			}, outLog)

			Convey("Then the log collection should start successfully", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When collecting logs fails", func() {
			// 模拟 logClient.ReadLog 失败
			patches := gomonkey.ApplyMethodFunc(&mockLogCollectorClient{}, "ReadLog", func(ctx context.Context, req *logservice.ReadLogRequest, opts ...grpc.CallOption) (logservice.LogCollectorService_ReadLogClient, error) {
				return nil, errors.New("read log failed")
			})
			defer patches.Reset()

			outLog := make(chan *logservice.ReadLogResponse, 1)
			err := client.CollectLog(context.Background(), &logservice.ReadLogRequest{
				Item: &logservice.LogItem{
					Filename:    "target-1",
					CollectorID: "collector-1",
					Target:      logservice.LogTarget_USER_STD,
					RuntimeID:   "runtime-1",
				},
				StartLine: 0,
				EndLine:   0,
			}, outLog)

			Convey("Then the log collection should fail", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

// // 用于模拟 logservice.LogCollectorService_ReadLogClient
type fakeReadLogClient struct {
	grpc.ClientStream
}

func (c *fakeReadLogClient) Recv() (*logservice.ReadLogResponse, error) {
	return &logservice.ReadLogResponse{Content: []byte("log content")}, io.EOF
}

func (c *fakeReadLogClient) CloseSend() error {
	return nil
}
