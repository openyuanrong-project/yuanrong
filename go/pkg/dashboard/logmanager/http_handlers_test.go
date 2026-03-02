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
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/logservice"
	"yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/dashboard/etcdcache"
)

var logManagerAddress = "127.0.0.1:55556"
var logCollectAddress = "127.0.0.1:55557"
var logManagerServer *grpc.Server
var logCollectorServer *grpc.Server

// MockLogCollectorServiceServer -
type MockLogCollectorServiceServer struct {
	logservice.UnimplementedLogCollectorServiceServer
}

func (MockLogCollectorServiceServer) ReadLog(req *logservice.ReadLogRequest, s logservice.LogCollectorService_ReadLogServer) error {
	if req.Item.Filename == "123" {
		resp := &logservice.ReadLogResponse{Content: []byte("123")}
		err := s.Send(resp)
		if err != nil {
			return err
		}
		return nil
	}
	return status.Error(codes.Unknown, "file not found")
}

func startLogManager(t *testing.T) {
	lis, err := net.Listen("tcp", logManagerAddress)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
	assert.Nil(t, err)
	logManagerServer = grpc.NewServer()
	logservice.RegisterLogManagerServiceServer(logManagerServer, &Server{})

	err = logManagerServer.Serve(lis)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
	assert.Nil(t, err)
}

func startLogCollector(t *testing.T) {
	lis, err := net.Listen("tcp", logCollectAddress)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
	assert.Nil(t, err)
	logCollectorServer = grpc.NewServer()
	logservice.RegisterLogCollectorServiceServer(logCollectorServer, &MockLogCollectorServiceServer{})

	err = logCollectorServer.Serve(lis)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
	assert.Nil(t, err)
}

func register(t *testing.T, collectorID string) {
	conn, err := grpc.Dial(logManagerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
	assert.Nil(t, err)
	defer func(conn *grpc.ClientConn) {
		_ = conn.Close()
	}(conn)
	logClient := logservice.NewLogManagerServiceClient(conn)
	req := logservice.RegisterRequest{CollectorID: collectorID, Address: logCollectAddress}
	_, err = logClient.Register(context.Background(), &req)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
	assert.Nil(t, err)
}

func reportLog(t *testing.T, fileName, runtimeID, collectorID string) *logservice.ReportLogResponse {
	conn, err := grpc.Dial(logManagerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
	assert.Nil(t, err)
	defer func(conn *grpc.ClientConn) {
		_ = conn.Close()
	}(conn)
	logClient := logservice.NewLogManagerServiceClient(conn)
	logItem := &logservice.LogItem{
		Filename:    fileName,
		CollectorID: collectorID,
		RuntimeID:   runtimeID,
		Target:      logservice.LogTarget_USER_STD,
	}
	req := logservice.ReportLogRequest{
		Items: make([]*logservice.LogItem, 0),
	}
	req.Items = append(req.Items, logItem)
	resp, err := logClient.ReportLog(context.Background(), &req)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
	assert.Nil(t, err)
	return resp
}

func closeServer() {
	if logManagerServer != nil {
		logManagerServer.Stop()
	}
	if logCollectorServer != nil {
		logCollectorServer.Stop()
	}

}

func TestReadLogHandler(t *testing.T) {

	convey.Convey("ReadLog succeed", t, func() {
		go startLogManager(t)
		go startLogCollector(t)
		defer closeServer()
		time.Sleep(100 * time.Millisecond)
		register(t, "123")
		etcdcache.InstanceCache.Put(&types.InstanceSpecification{
			InstanceID: "123",
			JobID:      "123",
			RuntimeID:  "123"})
		defer etcdcache.InstanceCache.Remove("123")
		reportLog(t, "123", "123", "123")
		defer managerSingleton.LogDB.Remove(&LogEntry{LogItem: &logservice.LogItem{
			Filename:    "123",
			RuntimeID:   "123",
			CollectorID: "123",
		}})

		rw := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rw)
		ctx.Request, _ = http.NewRequest("GET", "/?filename=123", nil)
		ReadLogHandler(ctx)
		convey.So(rw.Body.String(), convey.ShouldEqual, "123")
	})

	convey.Convey("ReadLog failed with duplicate log file", t, func() {
		go startLogManager(t)
		go startLogCollector(t)
		defer closeServer()
		time.Sleep(100 * time.Millisecond)
		register(t, "123")

		reportLog(t, "123", "123", "123")
		defer managerSingleton.LogDB.Remove(&LogEntry{LogItem: &logservice.LogItem{
			Filename:    "123",
			RuntimeID:   "123",
			CollectorID: "123",
		}})

		reportLog(t, "123", "234", "123")
		defer managerSingleton.LogDB.Remove(&LogEntry{LogItem: &logservice.LogItem{
			Filename:    "123",
			RuntimeID:   "234",
			CollectorID: "123",
		}})

		rw := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rw)
		ctx.Request, _ = http.NewRequest("GET", "/?filename=123", nil)
		ReadLogHandler(ctx)
		convey.So(rw.Code, convey.ShouldEqual, http.StatusInternalServerError)
		convey.So(rw.Body.String(), convey.ShouldContainSubstring, "duplicate")
	})

	convey.Convey("ReadLog failed with no collector", t, func() {
		go startLogManager(t)
		defer closeServer()
		time.Sleep(100 * time.Millisecond)
		reportLog(t, "123", "123", "234")
		defer managerSingleton.LogDB.Remove(&LogEntry{LogItem: &logservice.LogItem{
			Filename:    "123",
			RuntimeID:   "123",
			CollectorID: "234",
		}})
		rw := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rw)
		ctx.Request, _ = http.NewRequest("GET", "/?filename=123", nil)
		ReadLogHandler(ctx)
		convey.So(rw.Code, convey.ShouldEqual, http.StatusInternalServerError)
		convey.So(rw.Body.String(), convey.ShouldContainSubstring, "can not find collector")
	})

	convey.Convey("ReadLog failed with connect collector failed", t, func() {
		go startLogManager(t)
		defer closeServer()
		time.Sleep(100 * time.Millisecond)
		register(t, "123")
		reportLog(t, "123", "123", "123")
		defer managerSingleton.LogDB.Remove(&LogEntry{LogItem: &logservice.LogItem{
			Filename:    "123",
			RuntimeID:   "123",
			CollectorID: "123",
		}})
		rw := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rw)
		ctx.Request, _ = http.NewRequest("GET", "/?filename=123", nil)
		ReadLogHandler(ctx)
		convey.So(rw.Code, convey.ShouldEqual, http.StatusInternalServerError)
		convey.So(rw.Body.String(), convey.ShouldContainSubstring, "connection error")
	})

}

func TestListLogsHandler(t *testing.T) {
	convey.Convey("ReadLog succeed with empty logs", t, func() {
		rw := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rw)
		ctx.Request, _ = http.NewRequest("GET", "/?instance_id=123", nil)
		ListLogsHandler(ctx)
		convey.So(rw.Code, convey.ShouldEqual, http.StatusOK)
	})
}
