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

// Package rpcclient -
package rpcclient

import (
	"context"
	"errors"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/metadata"

	rtapi "yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb" // production: package api
	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/common"
	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/core"
	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/runtime"
	"yuanrong.org/kernel/pkg/common/faas_common/kernelrpc/connection"
	"yuanrong.org/kernel/pkg/common/faas_common/kernelrpc/utils"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	"yuanrong.org/kernel/pkg/common/faas_common/statuscode"
)

const (
	maxMsgSize          = 1024 * 1024 * 10
	maxWindowSize       = 1024 * 1024 * 10
	maxBufferSize       = 1024 * 1024 * 10
	dialBaseDelay       = 300 * time.Millisecond
	dialMultiplier      = 1.2
	dialJitter          = 0.1
	runtimeDialMaxDelay = 100 * time.Second
)

var (
	// ErrUnsupportedMethod -
	ErrUnsupportedMethod     = snerror.New(statuscode.InternalErrorCode, "unsupported method")
	dialTimeout              = 5 * time.Second
	dialRetryTime            = 10
	dialRetryInterval        = 3 * time.Second
	setupStreamRetryInterval = 3 * time.Second
	streamMessagePool        = sync.Pool{}
	invokeRequestPool        = sync.Pool{}
)

// StreamClientParams -
type StreamClientParams struct {
	SendReqConcurrentNum int
	SendRspConcurrentNum int
	RecvConcurrentNum    int
}

// BasicSteamClient is basic implementation of KernelClient which only sends POSIX calls as a runtime
type BasicSteamClient struct {
	clientConn *grpc.ClientConn
	streamConn connection.Connection
	peerAddr   string
}

// CreateBasicStreamClient creates BasicSteamClient
func CreateBasicStreamClient(peerAddr string, params StreamClientParams) (KernelClient, error) {
	conn, err := dialConnection(peerAddr)
	if err != nil {
		log.GetLogger().Errorf("failed to dial connection to %s error %s", peerAddr, err.Error())
		return nil, err
	}
	stream, err := createStream(conn, nil)
	if err != nil {
		log.GetLogger().Errorf("failed to create stream to %s error %s", peerAddr, err.Error())
		return nil, err
	}
	client := &BasicSteamClient{
		peerAddr:   peerAddr,
		clientConn: conn,
	}
	streamConn := connection.CreateStreamConnection(stream,
		connection.StreamParams{
			PeerAddr:             peerAddr,
			SendReqConcurrentNum: params.SendReqConcurrentNum,
			SendRspConcurrentNum: params.SendRspConcurrentNum,
			RecvConcurrentNum:    params.RecvConcurrentNum,
		},
		client.checkClientConnHealth, client.repairStream)
	client.streamConn = streamConn
	return client, nil
}

// Create -
func (k *BasicSteamClient) Create(funcKey string, args []*rtapi.Arg, createParams CreateParams,
	callback KernelClientCallback) (string, snerror.SNError) {
	return "", ErrUnsupportedMethod
}

// Invoke -
func (k *BasicSteamClient) Invoke(funcKey string, instanceID string, args []*rtapi.Arg, invokeParams InvokeParams,
	callback KernelClientCallback) (string, snerror.SNError) {
	CalibrateTransportParams(&invokeParams.TransportParams)
	message := acquireStreamMessageInvokeRequest()
	defer releaseStreamMessageInvokeRequest(message)
	invokeReq := message.GetInvokeReq()
	invokeReq.Function = funcKey
	invokeReq.Args = pb2Arg(args)
	invokeReq.InstanceID = instanceID
	if len(invokeParams.RequestID) != 0 {
		invokeReq.RequestID = invokeParams.RequestID
	} else {
		invokeReq.RequestID = utils.GenTaskID()
	}
	if len(invokeParams.TraceID) != 0 {
		invokeReq.TraceID = invokeParams.TraceID
	} else {
		invokeReq.TraceID = utils.GenTaskID()
	}
	sendOption := connection.SendOption{
		Timeout: invokeParams.Timeout,
	}
	sendCallback := func(notifyReq *runtime.NotifyRequest) {
		var (
			notifyMsg []byte
			notifyErr snerror.SNError
		)
		if notifyReq.Code != common.ErrorCode_ERR_NONE {
			notifyErr = snerror.New(int(notifyReq.Code), notifyReq.Message)
		} else {
			notifyMsg = []byte(notifyReq.Message)
		}
		callback(notifyMsg, notifyErr)
	}
	msg, err := k.streamConn.Send(message, sendOption, sendCallback)
	if err != nil {
		return "", snerror.New(statuscode.InternalErrorCode, err.Error())
	}
	sendRsp, ok := msg.GetBody().(*api.StreamingMessage_InvokeRsp)
	if !ok {
		return "", snerror.New(statuscode.InternalErrorCode, "invoke response type error")
	}
	if sendRsp.InvokeRsp.Code != common.ErrorCode_ERR_NONE {
		return "", snerror.New(int(sendRsp.InvokeRsp.Code), sendRsp.InvokeRsp.Message)
	}
	return sendRsp.InvokeRsp.Message, nil
}

// SaveState -
func (k *BasicSteamClient) SaveState(state []byte) (string, snerror.SNError) {
	return "", ErrUnsupportedMethod
}

// LoadState -
func (k *BasicSteamClient) LoadState(checkpointID string) ([]byte, snerror.SNError) {
	return nil, ErrUnsupportedMethod
}

// Kill -
func (k *BasicSteamClient) Kill(instanceID string, signal int32, payload []byte) snerror.SNError {
	return ErrUnsupportedMethod
}

// Exit -
func (k *BasicSteamClient) Exit() {
}

func (k *BasicSteamClient) checkClientConnHealth() bool {
	return checkClientConnHealth(k.clientConn)
}

func (k *BasicSteamClient) repairStream() connection.Stream {
	if !k.checkClientConnHealth() {
		conn, err := dialConnection(k.peerAddr)
		if err != nil {
			log.GetLogger().Errorf("failed to repair stream, dial connection to %s error %s", k.peerAddr, err.Error())
			return nil
		}
		k.clientConn = conn
	}
	stream, err := createStream(k.clientConn, nil)
	if err != nil {
		log.GetLogger().Errorf("failed to repair stream, create stream to %s error %s", k.peerAddr, err.Error())
		return nil
	}
	return stream
}

func dialConnection(addr string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), dialTimeout)
	defer cancel()
	dialFunc := func() (*grpc.ClientConn, error) {
		return grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock(),
			grpc.WithInitialWindowSize(maxWindowSize),
			grpc.WithInitialConnWindowSize(maxWindowSize),
			grpc.WithWriteBufferSize(maxBufferSize),
			grpc.WithReadBufferSize(maxBufferSize),
			grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(maxMsgSize), grpc.MaxCallRecvMsgSize(maxMsgSize)),
			grpc.WithConnectParams(grpc.ConnectParams{
				Backoff: backoff.Config{BaseDelay: dialBaseDelay, Multiplier: dialMultiplier, Jitter: dialJitter,
					MaxDelay: runtimeDialMaxDelay}, MinConnectTimeout: dialBaseDelay,
			}))
	}
	var (
		conn *grpc.ClientConn
		err  error
	)
	for i := 0; i < dialRetryTime; i++ {
		conn, err = dialFunc()
		if err == nil {
			return conn, err
		}
		log.GetLogger().Warnf("failed to dial connection to %s error %s", addr, err.Error())
		time.Sleep(time.Duration(i+1) * dialRetryInterval)
	}
	log.GetLogger().Errorf("failed to dial connection to %s after %d retries error %s", addr, dialRetryTime,
		err.Error())
	return nil, err
}

func createStream(conn *grpc.ClientConn, mdMap map[string]string) (api.RuntimeRPC_MessageStreamClient, error) {
	if !checkClientConnHealth(conn) {
		log.GetLogger().Errorf("grpc connection is nil, failed to create stream rpcclient")
		return nil, errors.New("conn is unhealthy")
	}
	client := api.NewRuntimeRPCClient(conn)
	md := metadata.New(mdMap)
	var (
		stream api.RuntimeRPC_MessageStreamClient
		err    error
	)
	var retryTimes int
	for i := 0; i < dialRetryTime; i++ {
		stream, err = client.MessageStream(metadata.NewOutgoingContext(context.Background(), md))
		if err == nil {
			log.GetLogger().Infof("succeed to get stream from function proxy")
			break
		}
		log.GetLogger().Errorf("failed to get stream from function proxy for %d times, err: %s",
			retryTimes, err.Error())
		time.Sleep(setupStreamRetryInterval)
	}
	if err != nil {
		log.GetLogger().Errorf("failed to create stream rpcclient to %s when setup message stream error %s", conn.Target(),
			err.Error())
		return nil, err
	}
	return stream, nil
}

func checkClientConnHealth(conn *grpc.ClientConn) bool {
	if conn == nil {
		return false
	}
	return conn.GetState() == connectivity.Idle || conn.GetState() == connectivity.Ready
}

func acquireStreamMessageInvokeRequest() *api.StreamingMessage {
	var (
		streamMsg *api.StreamingMessage
		invokeReq *api.StreamingMessage_InvokeReq
		ok        bool
	)
	streamMsg, ok = streamMessagePool.Get().(*api.StreamingMessage)
	if !ok {
		streamMsg = &api.StreamingMessage{}
	}
	invokeReq, ok = invokeRequestPool.Get().(*api.StreamingMessage_InvokeReq)
	if !ok {
		invokeReq = &api.StreamingMessage_InvokeReq{InvokeReq: &core.InvokeRequest{}}
	}
	streamMsg.Body = invokeReq
	return streamMsg
}

func releaseStreamMessageInvokeRequest(streamMsg *api.StreamingMessage) {
	invokeReq, ok := streamMsg.GetBody().(*api.StreamingMessage_InvokeReq)
	if !ok {
		return
	}
	invokeReq.InvokeReq.Reset()
	invokeRequestPool.Put(invokeReq)
	streamMsg.Reset()
	streamMessagePool.Put(streamMsg)
}
