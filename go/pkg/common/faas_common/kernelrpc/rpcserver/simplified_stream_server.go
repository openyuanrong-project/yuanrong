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
	"net"
	"reflect"
	"sync"

	"github.com/panjf2000/ants/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb"
	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/common"
	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/core"
	"yuanrong.org/kernel/pkg/common/faas_common/kernelrpc/connection"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

var (
	streamMessagePool  = sync.Pool{}
	invokeResponsePool = sync.Pool{}
)

type requestPack struct {
	msg  *api.StreamingMessage
	conn connection.Connection
}

// SimplifiedStreamServer is a simplified stream server which can respond to invokeRequest
type SimplifiedStreamServer struct {
	api.UnimplementedRuntimeRPCServer
	grpcServer    *grpc.Server
	taskPool      *ants.PoolWithFunc
	streamConnMap map[string]connection.Connection
	invokeHandler KernelInvokeHandler
	listenAddr    string
	stopped       bool
	stopCh        chan struct{}
	sync.Mutex
}

// CreateSimplifiedStreamServer creates SimplifiedStreamServer
func CreateSimplifiedStreamServer(listenAddr string, concurrentNum int) (KernelServer, error) {
	server := &SimplifiedStreamServer{
		grpcServer:    grpc.NewServer(),
		listenAddr:    listenAddr,
		streamConnMap: make(map[string]connection.Connection, constant.DefaultMapSize),
		stopCh:        make(chan struct{}),
	}
	taskPool, err := ants.NewPoolWithFunc(concurrentNum, func(arg interface{}) {
		reqPack, ok := arg.(requestPack)
		if !ok {
			return
		}
		server.handleRequest(reqPack.msg, reqPack.conn)
	})
	if err != nil {
		log.GetLogger().Errorf("failed to create task pool error %s", err.Error())
		return nil, err
	}
	server.taskPool = taskPool
	api.RegisterRuntimeRPCServer(server.grpcServer, server)
	return server, nil
}

// MessageStream handles stream from grpc server
func (s *SimplifiedStreamServer) MessageStream(stream api.RuntimeRPC_MessageStreamServer) error {
	peerObj, _ := peer.FromContext(stream.Context())
	peerAddr := peerObj.Addr.String()
	streamConn := connection.CreateStreamConnection(stream, connection.StreamParams{PeerAddr: peerAddr}, nil, nil)
	closeCh := streamConn.CheckClose()
	s.Lock()
	s.streamConnMap[peerAddr] = streamConn
	s.Unlock()
	log.GetLogger().Infof("create streamConn success,peer:%s", peerAddr)
	defer func() {
		s.Lock()
		delete(s.streamConnMap, peerAddr)
		s.Unlock()
	}()
	for {
		select {
		case <-s.stopCh:
			log.GetLogger().Warnf("server stops, closing stream connection to %s", peerAddr)
			streamConn.Close()
			return nil
		case <-closeCh:
			log.GetLogger().Warnf("stream connection to %s is closed", peerAddr)
			return nil
		default:
			msg, err := streamConn.Recv()
			if err != nil {
				log.GetLogger().Errorf("failed to receive stream message from %s error %s", peerAddr, err.Error())
				continue
			}
			if err = s.taskPool.Invoke(requestPack{msg: msg, conn: streamConn}); err != nil {
				log.GetLogger().Errorf("failed to invoke task pool error %s", err.Error())
			}
		}
	}
}

// Serve starts serving on listenAddr
func (s *SimplifiedStreamServer) Serve() error {
	lis, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		log.GetLogger().Errorf("failed to listen to address %s error %s\n", s.listenAddr, err.Error())
		return err
	}
	if err = s.grpcServer.Serve(lis); err != nil {
		log.GetLogger().Errorf("failed to serve on address %s error %s", s.listenAddr, err.Error())
		return err
	}
	log.GetLogger().Infof("stop serve on address %s", s.listenAddr)
	return nil
}

// Stop stops server
func (s *SimplifiedStreamServer) Stop() {
	s.Lock()
	if s.stopped {
		s.Unlock()
		return
	}
	s.stopped = true
	s.Unlock()
	s.grpcServer.GracefulStop()
	s.Lock()
	for _, stream := range s.streamConnMap {
		stream.Close()
	}
	s.Unlock()
}

func (s *SimplifiedStreamServer) handleRequest(msg *api.StreamingMessage, conn connection.Connection) {
	switch msg.GetBody().(type) {
	case *api.StreamingMessage_InvokeReq:
		invokeReq := msg.GetInvokeReq()
		message := acquireStreamMessageInvokeResponse()
		message.MessageID = msg.GetMessageID()
		InvokeRsp := message.GetInvokeRsp()
		defer func() {
			if _, err := conn.Send(message, connection.SendOption{Timeout: defaultSendTimeout}, nil); err != nil {
				log.GetLogger().Errorf("failed to send invoke response error %s", err.Error())
			}
			releaseStreamMessageInvokeResponse(message)
		}()
		if s.invokeHandler == nil {
			log.GetLogger().Errorf("invoke handler is nil")
			InvokeRsp.Code = common.ErrorCode_ERR_USER_FUNCTION_EXCEPTION
			InvokeRsp.Message = "invoke handler is nil"
			return
		}
		rsp, err := s.invokeHandler(pb2Arg(invokeReq.GetArgs()), invokeReq.TraceID)
		if err != nil {
			InvokeRsp.Code = common.ErrorCode_ERR_USER_FUNCTION_EXCEPTION
			InvokeRsp.Message = err.Error()
		} else {
			InvokeRsp.Code = common.ErrorCode_ERR_NONE
			InvokeRsp.Message = rsp
		}
	default:
		log.GetLogger().Warnf("receive unknown type message %s", reflect.TypeOf(msg.GetBody()).String())
	}
}

// RegisterInvokeHandler registers invokeHandler
func (s *SimplifiedStreamServer) RegisterInvokeHandler(handler KernelInvokeHandler) {
	s.invokeHandler = handler
}

func acquireStreamMessageInvokeResponse() *api.StreamingMessage {
	var (
		streamMsg *api.StreamingMessage
		invokeRsp *api.StreamingMessage_InvokeRsp
		ok        bool
	)
	streamMsg, ok = streamMessagePool.Get().(*api.StreamingMessage)
	if !ok {
		streamMsg = &api.StreamingMessage{}
	}
	invokeRsp, ok = invokeResponsePool.Get().(*api.StreamingMessage_InvokeRsp)
	if !ok {
		invokeRsp = &api.StreamingMessage_InvokeRsp{InvokeRsp: &core.InvokeResponse{}}
	}
	streamMsg.Body = invokeRsp
	return streamMsg
}

func releaseStreamMessageInvokeResponse(streamMsg *api.StreamingMessage) {
	invokeRsp, ok := streamMsg.GetBody().(*api.StreamingMessage_InvokeRsp)
	if !ok {
		return
	}
	invokeRsp.InvokeRsp.Reset()
	invokeResponsePool.Put(invokeRsp)
	streamMsg.Reset()
	streamMessagePool.Put(streamMsg)
}
