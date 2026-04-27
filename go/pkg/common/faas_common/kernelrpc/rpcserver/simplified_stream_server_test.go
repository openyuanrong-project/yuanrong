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

// Package rpcserver -
package rpcserver

import (
	"errors"
	"net"
	"testing"
	"time"

	gomonkey "github.com/agiledragon/gomonkey/v2"
	ants "github.com/panjf2000/ants/v2"
	"github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/common"
	"yuanrong.org/kernel/pkg/common/faas_common/kernelrpc/rpcclient"
)

func TestCreateSimplifiedStreamServer(t *testing.T) {
	convey.Convey("test CreateSimplifiedStreamServer", t, func() {
		patch := gomonkey.ApplyFunc(ants.NewPoolWithFunc, func(size int, pf func(interface{}), options ...ants.Option) (
			*ants.PoolWithFunc, error) {
			return nil, errors.New("some error")
		})
		server, err := CreateSimplifiedStreamServer("0.0.0.0:0", 100)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(server, convey.ShouldBeNil)
		patch.Reset()
		server, err = CreateSimplifiedStreamServer("0.0.0.0:0", 100)
		convey.So(err, convey.ShouldBeNil)
		convey.So(server, convey.ShouldNotBeNil)
	})
}

func TestSimplifiedStreamServerServeAndClose(t *testing.T) {
	convey.Convey("test SimplifiedStreamServer serve", t, func() {
		patch := gomonkey.ApplyFunc(net.Listen, func(network, address string) (net.Listener, error) {
			return nil, errors.New("some error")
		})
		server, _ := CreateSimplifiedStreamServer("0.0.0.0:0", 100)
		err := server.Serve()
		convey.So(err, convey.ShouldNotBeNil)
		patch.Reset()
		patch = gomonkey.ApplyFunc((*grpc.Server).Serve, func(_ *grpc.Server, lis net.Listener) error {
			return errors.New("some error")
		})
		server, _ = CreateSimplifiedStreamServer("0.0.0.0:0", 100)
		err = server.Serve()
		convey.So(err, convey.ShouldNotBeNil)
		patch.Reset()
		server, _ = CreateSimplifiedStreamServer("0.0.0.0:0", 100)
		go func() {
			time.Sleep(100 * time.Millisecond)
			server.Stop()
			server.Stop()
		}()
		err = server.Serve()
		convey.So(err, convey.ShouldBeNil)
		server.Stop()
	})
}

func TestSimplifiedStreamServerHandleInvoke(t *testing.T) {
	convey.Convey("test SimplifiedStreamServer handleInvoke", t, func() {
		server, _ := CreateSimplifiedStreamServer("0.0.0.0:5678", 100)
		go server.Serve()
		client, _ := rpcclient.CreateBasicStreamClient("0.0.0.0:5678", rpcclient.StreamClientParams{})
		args := []*api.Arg{{
			Type: api.Value,
			Data: []byte("123"),
		}}
		convey.Convey("invokeHandler is nil", func() {
			_, err := client.Invoke("testFunc", "testIns", args, rpcclient.InvokeParams{}, nil)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Code(), convey.ShouldEqual, common.ErrorCode_ERR_USER_FUNCTION_EXCEPTION)
		})
		convey.Convey("invokeHandler return error", func() {
			invokeHandler := func(args []*api.Arg, traceID string) (string, error) {
				return "", errors.New("some error")
			}
			server.RegisterInvokeHandler(invokeHandler)
			msg, err := client.Invoke("testFunc", "testIns", args, rpcclient.InvokeParams{}, nil)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Code(), convey.ShouldEqual, common.ErrorCode_ERR_USER_FUNCTION_EXCEPTION)
			convey.So(msg, convey.ShouldBeEmpty)
		})
		convey.Convey("invokeHandler return ok", func() {
			invokeHandler := func(args []*api.Arg, traceID string) (string, error) {
				return "abc", nil
			}
			server.RegisterInvokeHandler(invokeHandler)
			msg, err := client.Invoke("testFunc", "testIns", args, rpcclient.InvokeParams{}, nil)
			convey.So(err, convey.ShouldBeNil)
			convey.So(msg, convey.ShouldEqual, "abc")
		})
		grpcServer, _ := server.(*SimplifiedStreamServer)
		grpcServer.grpcServer.Stop()
	})
}
