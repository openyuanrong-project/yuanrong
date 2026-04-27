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

// Package connection -
package connection

import (
	"errors"
	"reflect"
	"sync"
	"time"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb" // production: package api
	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/runtime"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/uuid"
)

var (
	defaultChannelSize = 300000
)

var (
	// ErrStreamConnectionBroken is the error of stream connection broken
	ErrStreamConnectionBroken = errors.New("stream connection is broken")
	// ErrStreamConnectionClosed is the error of stream connection closed
	ErrStreamConnectionClosed = errors.New("stream connection is closed")
	// ErrRequestIDAlreadyExist is the error of requestID already exist
	ErrRequestIDAlreadyExist = errors.New("requestID already exist")
)

// StreamParams -
type StreamParams struct {
	PeerAddr             string
	SendReqConcurrentNum int
	SendRspConcurrentNum int
	RecvConcurrentNum    int
}

// Stream is the common interface of stream for both server and client
type Stream interface {
	Send(*api.StreamingMessage) error
	Recv() (*api.StreamingMessage, error)
}

// RepairStreamFunc repairs stream
type RepairStreamFunc func() Stream

// HealthCheckFunc checks stream health
type HealthCheckFunc func() bool

type sendAckPack struct {
	rsp *api.StreamingMessage
	err error
}

type sendCbPack struct {
	t  time.Time
	cb SendCallback
}

// StreamConnection is an implementation of Connection with stream
type StreamConnection struct {
	stream        Stream
	sendAckRecord map[string]chan sendAckPack
	sendCbRecord  map[string]sendCbPack
	peerAddr      string
	closed        bool
	repairing     bool
	repairFunc    RepairStreamFunc
	healthFunc    HealthCheckFunc
	sendReqCh     chan *api.StreamingMessage
	sendRspCh     chan *api.StreamingMessage
	recvCh        chan *api.StreamingMessage
	repairCh      chan struct{}
	closeCh       chan struct{}
	*sync.RWMutex
	*sync.Cond
}

// CreateStreamConnection creates a StreamConnection
func CreateStreamConnection(stream Stream, params StreamParams, healthFunc HealthCheckFunc,
	repairFunc RepairStreamFunc) Connection {
	calibrateParams(&params)
	mutex := new(sync.RWMutex)
	sc := &StreamConnection{
		stream:        stream,
		sendAckRecord: make(map[string]chan sendAckPack, constant.DefaultMapSize),
		sendCbRecord:  make(map[string]sendCbPack, constant.DefaultMapSize),
		peerAddr:      params.PeerAddr,
		healthFunc:    healthFunc,
		repairFunc:    repairFunc,
		sendReqCh:     make(chan *api.StreamingMessage, defaultChannelSize),
		sendRspCh:     make(chan *api.StreamingMessage, defaultChannelSize),
		recvCh:        make(chan *api.StreamingMessage, defaultChannelSize),
		repairCh:      make(chan struct{}, 1),
		closeCh:       make(chan struct{}),
		RWMutex:       mutex,
		Cond:          sync.NewCond(mutex),
	}
	startLoopProcess(func() { sc.sendLoop(sc.sendReqCh) }, params.SendReqConcurrentNum)
	startLoopProcess(func() { sc.sendLoop(sc.sendRspCh) }, params.SendRspConcurrentNum)
	startLoopProcess(sc.recvLoop, params.RecvConcurrentNum)
	if repairFunc != nil {
		startLoopProcess(sc.repairLoop, 1)
	}
	return sc
}

// Send sends stream message
func (sc *StreamConnection) Send(message *api.StreamingMessage, option SendOption, callback SendCallback) (
	*api.StreamingMessage, error) {
	select {
	case <-sc.closeCh:
		return nil, ErrStreamConnectionClosed
	default:
	}
	if sc.healthFunc != nil && !sc.healthFunc() {
		return nil, ErrStreamConnectionBroken
	}
	if len(message.MessageID) == 0 {
		message.MessageID = uuid.New().String()
	}
	sc.Lock()
	ackCh := make(chan sendAckPack, 1)
	sc.sendAckRecord[message.MessageID] = ackCh
	// message with requestID is an async message which needs a callback
	requestID := getRequestID(message.GetBody())
	if len(requestID) != 0 && callback != nil {
		if _, exist := sc.sendCbRecord[requestID]; exist {
			sc.Unlock()
			return nil, ErrRequestIDAlreadyExist
		}
		sc.sendCbRecord[requestID] = sendCbPack{t: time.Now(), cb: callback}
	}
	sc.Unlock()
	defer func() {
		sc.Lock()
		delete(sc.sendAckRecord, message.MessageID)
		sc.Unlock()
	}()
	select {
	case sc.sendReqCh <- message:
	default:
		log.GetLogger().Warnf("send channel reach limit %d for connection of %s", defaultChannelSize, sc.peerAddr)
		sc.Lock()
		delete(sc.sendCbRecord, requestID)
		sc.Unlock()
		return nil, errors.New("stream send is blocked")
	}
	timer := time.NewTimer(option.Timeout)
	select {
	case <-timer.C:
		// send failed, no need to record callback
		sc.Lock()
		delete(sc.sendCbRecord, requestID)
		sc.Unlock()
		return nil, errors.New("send timeout")
	case ackPack, ok := <-ackCh:
		// consider to add retry here
		if !ok {
			return nil, errors.New("send response channel closed")
		}
		return ackPack.rsp, ackPack.err
	}
}

// Recv receives stream message
func (sc *StreamConnection) Recv() (*api.StreamingMessage, error) {
	select {
	case <-sc.closeCh:
		return nil, ErrStreamConnectionClosed
	case msg, ok := <-sc.recvCh:
		if !ok {
			return nil, errors.New("recv channel is closed")
		}
		return msg, nil
	}
}

// Close closes stream
func (sc *StreamConnection) Close() {
	sc.Lock()
	if sc.closed {
		sc.Unlock()
		return
	}
	sc.closed = true
	sc.Unlock()
	close(sc.closeCh)
}

// CheckClose checks if stream is closed
func (sc *StreamConnection) CheckClose() chan struct{} {
	return sc.closeCh
}

func (sc *StreamConnection) sendLoop(sendCh chan *api.StreamingMessage) {
	for {
		select {
		case <-sc.closeCh:
			log.GetLogger().Debugf("stop send loop for connection of %s", sc.peerAddr)
			return
		case msg, ok := <-sendCh:
			if !ok {
				log.GetLogger().Warnf("close stream, send channel closed for connection of %s", sc.peerAddr)
				return
			}
			if !sc.waitForStreamFix() {
				log.GetLogger().Warnf("cannot fix stream, stop send loop for connection of %s", sc.peerAddr)
				return
			}
			err := sc.stream.Send(msg)
			sc.RLock()
			ackCh, exist := sc.sendAckRecord[msg.GetMessageID()]
			sc.RUnlock()
			if err != nil {
				if exist && ackCh != nil {
					ackCh <- sendAckPack{
						rsp: nil,
						err: err,
					}
				} else {
					log.GetLogger().Warnf("response channel for sending message %s doesn't exist for connection %s",
						msg.MessageID, sc.peerAddr)
				}
				sc.repairStream()
				continue
			}
			if !expectResponse(msg) {
				if exist && ackCh != nil {
					ackCh <- sendAckPack{
						rsp: nil,
						err: nil,
					}
				} else {
					log.GetLogger().Warnf("response channel for sending message %s doesn't exist for connection %s",
						msg.MessageID, sc.peerAddr)
				}
			}
		}
	}
}

func (sc *StreamConnection) recvLoop() {
	for {
		select {
		case <-sc.closeCh:
			log.GetLogger().Debugf("close stream, stop recv loop for connection of %s", sc.peerAddr)
			return
		default:
			if !sc.waitForStreamFix() {
				log.GetLogger().Warnf("cannot fix stream, stop recv loop for connection of %s", sc.peerAddr)
				return
			}
			msg, err := sc.stream.Recv()
			if err != nil {
				log.GetLogger().Errorf("receive error %s for connection of %s", err.Error(), sc.peerAddr)
				sc.repairStream()
				continue
			}
			switch msg.GetBody().(type) {
			case *api.StreamingMessage_CreateRsp, *api.StreamingMessage_InvokeRsp, *api.StreamingMessage_ExitRsp,
				*api.StreamingMessage_SaveRsp, *api.StreamingMessage_LoadRsp, *api.StreamingMessage_KillRsp,
				*api.StreamingMessage_NotifyRsp:
				sc.Lock()
				askCh, exist := sc.sendAckRecord[msg.GetMessageID()]
				if exist {
					delete(sc.sendAckRecord, msg.GetMessageID())
				} else {
					log.GetLogger().Warnf("receive unexpected response messageID %s for connection %s",
						msg.GetMessageID(), sc.peerAddr)
				}
				sc.Unlock()
				if exist {
					askCh <- sendAckPack{
						rsp: msg,
						err: nil,
					}
					continue
				}

			case *api.StreamingMessage_CallReq, *api.StreamingMessage_CheckpointReq, *api.StreamingMessage_RecoverReq,
				*api.StreamingMessage_ShutdownReq, *api.StreamingMessage_SignalReq, *api.StreamingMessage_InvokeReq:
				// StreamingMessage_InvokeReq is used in simplified server mode
				select {
				case sc.recvCh <- msg:
				default:
					log.GetLogger().Warnf("receive channel reaches limit %d for connection %s", defaultChannelSize,
						sc.peerAddr)
				}
			case *api.StreamingMessage_NotifyReq:
				notifyReq := msg.GetNotifyReq()
				requestID := notifyReq.GetRequestID()
				sc.Lock()
				cbPack, exist := sc.sendCbRecord[requestID]
				if exist {
					delete(sc.sendCbRecord, requestID)
				} else {
					log.GetLogger().Warnf("receive unexpected notify requestID %s for connection %s", requestID,
						sc.peerAddr)
				}
				sc.Unlock()
				if exist {
					go cbPack.cb(notifyReq)
				}
				select {
				case sc.sendRspCh <- &api.StreamingMessage{
					MessageID: msg.GetMessageID(),
					Body: &api.StreamingMessage_NotifyRsp{
						NotifyRsp: &runtime.NotifyResponse{},
					},
				}:
				default:
					log.GetLogger().Warnf("sendRsp channel reaches limit %d for connection %s", defaultChannelSize,
						sc.peerAddr)
				}
			case *api.StreamingMessage_HeartbeatReq:
				select {
				case sc.sendRspCh <- &api.StreamingMessage{
					MessageID: msg.GetMessageID(),
					Body: &api.StreamingMessage_HeartbeatRsp{
						HeartbeatRsp: &runtime.HeartbeatResponse{},
					},
				}:
				default:
					log.GetLogger().Warnf("sendRsp channel reaches limit %d for connection %s", defaultChannelSize,
						sc.peerAddr)
				}
			default:
				log.GetLogger().Warnf("receive unknown type message %s", reflect.TypeOf(msg.GetBody()).String())
			}

		}
	}
}

func (sc *StreamConnection) repairLoop() {
	for {
		select {
		case <-sc.closeCh:
			log.GetLogger().Debugf("stop recv loop for connection of %s", sc.peerAddr)
			return
		case _, ok := <-sc.repairCh:
			if !ok {
				log.GetLogger().Warnf("repair channel closed for connection of %s", sc.peerAddr)
				return
			}
			stream := sc.repairFunc()
			if stream == nil {
				log.GetLogger().Warnf("failed to fix stream during fix loop")
				continue
			}
			sc.Lock()
			sc.repairing = false
			sc.stream = stream
			sc.Unlock()
			sc.Broadcast()
		}
	}
}

func (sc *StreamConnection) waitForStreamFix() bool {
	sc.L.Lock()
	if sc.repairing || (sc.healthFunc != nil && !sc.healthFunc()) {
		sc.Wait()
	}
	if sc.closed {
		sc.L.Unlock()
		return false
	}
	sc.L.Unlock()
	return true
}

func (sc *StreamConnection) repairStream() {
	sc.Lock()
	// close stream if there is no way to fix it
	if sc.repairFunc == nil {
		sc.Unlock()
		sc.Close()
		return
	}
	if sc.repairing {
		sc.Unlock()
		return
	}
	sc.repairing = true
	sc.Unlock()
	select {
	case sc.repairCh <- struct{}{}:
	default:
	}
}

func calibrateParams(params *StreamParams) {
	if params.SendReqConcurrentNum < 1 {
		params.SendReqConcurrentNum = 1
	}
	if params.SendRspConcurrentNum < 1 {
		params.SendRspConcurrentNum = 1
	}
	if params.RecvConcurrentNum < 1 {
		params.RecvConcurrentNum = 1
	}
}

func startLoopProcess(loop func(), num int) {
	for i := 0; i < num; i++ {
		go loop()
	}
}

// only async request contains requestID (create and invoke)
func getRequestID(req interface{}) string {
	switch req.(type) {
	case *api.StreamingMessage_CreateReq:
		return req.(*api.StreamingMessage_CreateReq).CreateReq.GetRequestID()
	case *api.StreamingMessage_InvokeReq:
		return req.(*api.StreamingMessage_InvokeReq).InvokeReq.GetRequestID()
	default:
		return ""
	}
}

// server may send some message which doesn't expect response
func expectResponse(msg *api.StreamingMessage) bool {
	switch msg.GetBody().(type) {
	case *api.StreamingMessage_CreateReq, *api.StreamingMessage_InvokeReq, *api.StreamingMessage_ExitReq,
		*api.StreamingMessage_SaveReq, *api.StreamingMessage_LoadReq, *api.StreamingMessage_KillReq,
		*api.StreamingMessage_NotifyReq:
		return true
	default:
		return false
	}
}
