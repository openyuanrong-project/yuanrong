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
	"io"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"

	"github.com/smartystreets/goconvey/convey"
	api "yuanrong.org/kernel/pkg/common/faas_common/grpc/pb"
	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/core"
	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/runtime"
)

type fakeStream struct {
	sendDelay time.Duration
	sendErrCh chan error
	recvErrCh chan error
	sendCh    chan *api.StreamingMessage
	recvCh    chan *api.StreamingMessage
}

func createFakeStream(sendDelay time.Duration) *fakeStream {
	return &fakeStream{
		sendDelay: sendDelay,
		sendErrCh: make(chan error, 1),
		recvErrCh: make(chan error, 1),
		sendCh:    make(chan *api.StreamingMessage, 1),
		recvCh:    make(chan *api.StreamingMessage, 1),
	}
}

func (f *fakeStream) Send(msg *api.StreamingMessage) error {
	if f.sendDelay != 0 {
		<-time.After(f.sendDelay)
	}
	select {
	case err := <-f.sendErrCh:
		return err
	default:
		f.sendCh <- msg
		return nil
	}
}

func (f *fakeStream) Recv() (*api.StreamingMessage, error) {
	select {
	case err := <-f.recvErrCh:
		return nil, err
	case msg := <-f.recvCh:
		return msg, nil
	}
}

func TestStreamSend(t *testing.T) {
	convey.Convey("test steam send", t, func() {
		healthReturn := true
		healthFunc := func() bool {
			return healthReturn
		}
		repairCount := 0
		repairFunc := func() Stream {
			repairCount++
			return &fakeStream{}
		}
		var callbackRes *runtime.NotifyRequest
		callbackFunc := func(message *runtime.NotifyRequest) {
			callbackRes = message
		}
		convey.Convey("unhealthy stream", func() {
			healthReturn = false
			repairCount = 0
			stream := createFakeStream(0)
			sc := CreateStreamConnection(stream, StreamParams{}, healthFunc, repairFunc)
			_, err := sc.Send(&api.StreamingMessage{
				Body: &api.StreamingMessage_InvokeReq{
					InvokeReq: &core.InvokeRequest{
						RequestID: "reqID-123",
					},
				},
			}, SendOption{Timeout: 200 * time.Millisecond}, nil)
			convey.So(err, convey.ShouldEqual, ErrStreamConnectionBroken)
		})
		convey.Convey("requestID already exist", func() {
			healthReturn = true
			repairCount = 0
			stream := createFakeStream(1 * time.Second)
			sc := CreateStreamConnection(stream, StreamParams{}, healthFunc, repairFunc)
			go sc.Send(&api.StreamingMessage{
				Body: &api.StreamingMessage_InvokeReq{
					InvokeReq: &core.InvokeRequest{
						RequestID: "reqID-123",
					},
				},
			}, SendOption{Timeout: 200 * time.Millisecond}, callbackFunc)
			time.Sleep(100 * time.Millisecond)
			_, err := sc.Send(&api.StreamingMessage{
				Body: &api.StreamingMessage_InvokeReq{
					InvokeReq: &core.InvokeRequest{
						RequestID: "reqID-123",
					},
				},
			}, SendOption{Timeout: 200 * time.Millisecond}, callbackFunc)
			convey.So(err, convey.ShouldEqual, ErrRequestIDAlreadyExist)
		})
		convey.Convey("stream send blocked", func() {
			healthReturn = true
			repairCount = 0
			patch := gomonkey.ApplyGlobalVar(&defaultChannelSize, 1)
			stream := createFakeStream(1 * time.Second)
			sc := CreateStreamConnection(stream, StreamParams{}, healthFunc, repairFunc)
			sc.(*StreamConnection).sendReqCh <- &api.StreamingMessage{}
			sc.(*StreamConnection).sendReqCh <- &api.StreamingMessage{}
			_, err := sc.Send(&api.StreamingMessage{
				Body: &api.StreamingMessage_InvokeReq{
					InvokeReq: &core.InvokeRequest{
						RequestID: "reqID-789",
					},
				},
			}, SendOption{Timeout: 200 * time.Millisecond}, nil)
			convey.So(err.Error(), convey.ShouldEqual, "stream send is blocked")
			patch.Reset()
		})
		convey.Convey("stream send error and repair", func() {
			healthReturn = true
			repairCount = 0
			stream := createFakeStream(0)
			var newStream *fakeStream
			repairFunc = func() Stream {
				repairCount++
				newStream = createFakeStream(0)
				return newStream
			}
			sc := CreateStreamConnection(stream, StreamParams{}, healthFunc, repairFunc)
			stream.sendErrCh <- io.EOF
			_, err := sc.Send(&api.StreamingMessage{
				Body: &api.StreamingMessage_InvokeReq{
					InvokeReq: &core.InvokeRequest{
						RequestID: "reqID-123",
					},
				},
			}, SendOption{Timeout: 200 * time.Millisecond}, nil)
			time.Sleep(100 * time.Millisecond)
			convey.So(err, convey.ShouldEqual, io.EOF)
			convey.So(repairCount, convey.ShouldEqual, 1)
			stream.recvErrCh <- io.EOF
			time.Sleep(100 * time.Millisecond)
			go func() {
				time.Sleep(100 * time.Millisecond)
				msg := <-newStream.sendCh
				newStream.recvCh <- &api.StreamingMessage{
					MessageID: msg.GetMessageID(),
					Body: &api.StreamingMessage_InvokeRsp{
						InvokeRsp: &core.InvokeResponse{},
					},
				}
			}()
			_, err = sc.Send(&api.StreamingMessage{
				Body: &api.StreamingMessage_InvokeReq{
					InvokeReq: &core.InvokeRequest{
						RequestID: "reqID-123",
					},
				},
			}, SendOption{Timeout: 200 * time.Millisecond}, nil)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("stream send error and no repair", func() {
			healthReturn = true
			repairCount = 0
			stream := createFakeStream(0)
			sc := CreateStreamConnection(stream, StreamParams{}, healthFunc, nil)
			stream.sendErrCh <- io.EOF
			_, err := sc.Send(&api.StreamingMessage{
				Body: &api.StreamingMessage_InvokeReq{
					InvokeReq: &core.InvokeRequest{
						RequestID: "reqID-123",
					},
				},
			}, SendOption{Timeout: 200 * time.Millisecond}, nil)
			convey.So(err, convey.ShouldEqual, io.EOF)
			convey.So(repairCount, convey.ShouldEqual, 0)
			_, err = sc.Send(&api.StreamingMessage{
				Body: &api.StreamingMessage_InvokeReq{
					InvokeReq: &core.InvokeRequest{
						RequestID: "reqID-123",
					},
				},
			}, SendOption{Timeout: 200 * time.Millisecond}, nil)
			convey.So(err, convey.ShouldEqual, ErrStreamConnectionClosed)
			convey.So(repairCount, convey.ShouldEqual, 0)
		})
		convey.Convey("stream send timeout", func() {
			healthReturn = true
			repairCount = 0
			stream := createFakeStream(0)
			sc := CreateStreamConnection(stream, StreamParams{}, healthFunc, repairFunc)
			_, err := sc.Send(&api.StreamingMessage{
				MessageID: "msgID-123",
				Body: &api.StreamingMessage_InvokeReq{
					InvokeReq: &core.InvokeRequest{
						RequestID: "reqID-123",
					},
				},
			}, SendOption{Timeout: 100 * time.Millisecond}, nil)
			convey.So(err.Error(), convey.ShouldEqual, "send timeout")
			convey.So(repairCount, convey.ShouldEqual, 0)
		})
		convey.Convey("stream send expect no response", func() {
			healthReturn = true
			repairCount = 0
			stream := createFakeStream(0)
			sc := CreateStreamConnection(stream, StreamParams{}, healthFunc, repairFunc)
			_, err := sc.Send(&api.StreamingMessage{
				MessageID: "msgID-123",
				Body: &api.StreamingMessage_InvokeRsp{
					InvokeRsp: &core.InvokeResponse{},
				},
			}, SendOption{Timeout: 100 * time.Millisecond}, nil)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("stream send expect response", func() {
			healthReturn = true
			repairCount = 0
			stream := createFakeStream(0)
			sc := CreateStreamConnection(stream, StreamParams{}, healthFunc, repairFunc)
			go func() {
				time.Sleep(100 * time.Millisecond)
				msg := <-stream.sendCh
				stream.recvCh <- &api.StreamingMessage{
					MessageID: msg.GetMessageID(),
					Body: &api.StreamingMessage_InvokeRsp{
						InvokeRsp: &core.InvokeResponse{},
					},
				}
				stream.recvCh <- &api.StreamingMessage{
					Body: &api.StreamingMessage_NotifyReq{
						NotifyReq: &runtime.NotifyRequest{
							RequestID: "reqID-123",
						},
					},
				}
			}()
			_, err := sc.Send(&api.StreamingMessage{
				Body: &api.StreamingMessage_InvokeReq{
					InvokeReq: &core.InvokeRequest{
						RequestID: "reqID-123",
					},
				},
			}, SendOption{Timeout: 200 * time.Minute}, callbackFunc)
			time.Sleep(100 * time.Millisecond)
			convey.So(err, convey.ShouldBeNil)
			convey.So(repairCount, convey.ShouldEqual, 0)
			convey.So(callbackRes, convey.ShouldNotBeNil)
		})
	})
}

func TestStreamRecv(t *testing.T) {
	convey.Convey("test steam receive", t, func() {
		healthReturn := true
		healthFunc := func() bool {
			return healthReturn
		}
		repairCount := 0
		repairFunc := func() Stream {
			repairCount++
			return createFakeStream(0)
		}
		convey.Convey("stream receive error and no repair", func() {
			healthReturn = true
			repairCount = 0
			stream := createFakeStream(0)
			sc := CreateStreamConnection(stream, StreamParams{}, healthFunc, nil)
			stream.recvErrCh <- errors.New("some error")
			time.Sleep(100 * time.Millisecond)
			_, err := sc.Recv()
			convey.So(err, convey.ShouldEqual, ErrStreamConnectionClosed)
		})
		convey.Convey("stream receive error and repair", func() {
			healthReturn = true
			repairCount = 0
			stream := createFakeStream(0)
			var newStream *fakeStream
			repairFunc = func() Stream {
				repairCount++
				newStream = createFakeStream(0)
				return newStream
			}
			sc := CreateStreamConnection(stream, StreamParams{}, healthFunc, repairFunc)
			stream.recvErrCh <- io.EOF
			time.Sleep(100 * time.Millisecond)
			newStream.recvCh <- &api.StreamingMessage{
				Body: &api.StreamingMessage_CallReq{
					CallReq: &runtime.CallRequest{
						RequestID: "reqID-123",
					},
				},
			}
			time.Sleep(100 * time.Millisecond)
			_, err := sc.Recv()
			convey.So(err, convey.ShouldBeNil)
			convey.So(repairCount, convey.ShouldEqual, 1)
		})
		convey.Convey("receive call message", func() {
			healthReturn = true
			repairCount = 0
			stream := createFakeStream(0)
			sc := CreateStreamConnection(stream, StreamParams{}, healthFunc, repairFunc)
			stream.recvCh <- &api.StreamingMessage{
				Body: &api.StreamingMessage_CallReq{
					CallReq: &runtime.CallRequest{
						RequestID: "reqID-123",
					},
				},
			}
			msg, err := sc.Recv()
			convey.So(err, convey.ShouldBeNil)
			callReq := msg.GetCallReq()
			convey.So(callReq, convey.ShouldNotBeNil)
			convey.So(callReq.GetRequestID(), convey.ShouldEqual, "reqID-123")
		})
		convey.Convey("receive heartbeat message", func() {
			healthReturn = true
			repairCount = 0
			stream := createFakeStream(0)
			CreateStreamConnection(stream, StreamParams{}, healthFunc, repairFunc)
			stream.recvCh <- &api.StreamingMessage{
				Body: &api.StreamingMessage_HeartbeatReq{},
			}
			msg := <-stream.sendCh
			heartbeatRsp := msg.GetHeartbeatRsp()
			convey.So(heartbeatRsp, convey.ShouldNotBeNil)
		})
		convey.Convey("receive unexpected id", func() {
			healthReturn = true
			repairCount = 0
			stream := createFakeStream(0)
			CreateStreamConnection(stream, StreamParams{}, healthFunc, repairFunc)
			stream.recvCh <- &api.StreamingMessage{
				MessageID: "msgID-123",
				Body: &api.StreamingMessage_NotifyRsp{
					NotifyRsp: &runtime.NotifyResponse{},
				},
			}
			stream.recvCh <- &api.StreamingMessage{
				MessageID: "msgID-123",
				Body: &api.StreamingMessage_NotifyReq{
					NotifyReq: &runtime.NotifyRequest{
						RequestID: "reqID-123",
					},
				},
			}
			time.Sleep(100 * time.Millisecond)
			convey.So(len(stream.sendCh), convey.ShouldEqual, 1)
		})
		convey.Convey("receive unsupported message", func() {
			healthReturn = true
			repairCount = 0
			stream := createFakeStream(0)
			CreateStreamConnection(stream, StreamParams{}, healthFunc, repairFunc)
			stream.recvCh <- &api.StreamingMessage{
				Body: &api.StreamingMessage_SignalRsp{},
			}
			convey.So(len(stream.sendCh), convey.ShouldEqual, 0)
		})
	})
}

func TestStreamClose(t *testing.T) {
	convey.Convey("test steam close", t, func() {
		repairFunc := func() Stream {
			return createFakeStream(0)
		}
		stream := createFakeStream(0)
		sc := CreateStreamConnection(stream, StreamParams{}, nil, repairFunc)
		closeCh := sc.CheckClose()
		sc.Close()
		sc.Close()
		convey.So(sc.(*StreamConnection).closed, convey.ShouldEqual, true)
		select {
		case <-closeCh:
		default:
			t.Errorf("closeCh is not closed")
		}
		_, err := sc.Send(nil, SendOption{}, nil)
		convey.So(err, convey.ShouldEqual, ErrStreamConnectionClosed)
		stream.recvErrCh <- io.EOF
		_, err = sc.Recv()
		convey.So(err, convey.ShouldEqual, ErrStreamConnectionClosed)
	})
}
