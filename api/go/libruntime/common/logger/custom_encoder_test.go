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

// Package logger for log
package logger

import (
	"math"
	"os"
	"testing"
	"time"
	"unsafe"

	"github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap/zapcore"

	"yuanrong.org/kernel/runtime/libruntime/common/constants"
)

func TestNewConsoleEncoder(t *testing.T) {
	convey.Convey(
		"Test NewConsoleEncoder", t, func() {
			convey.Convey(
				"NewConsoleEncoder success", func() {
					enc, err := NewConsoleEncoder(zapcore.EncoderConfig{})
					convey.So(enc, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestNewCustomEncoder(t *testing.T) {
	convey.Convey(
		"Test NewCustomEncoder", t, func() {
			convey.Convey(
				"NewCustomEncoder success", func() {
					enc := NewCustomEncoder(&zapcore.EncoderConfig{})
					convey.So(enc, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestPutCustomEncoder(t *testing.T) {
	convey.Convey(
		"Test putCustomEncoder", t, func() {
			convey.Convey(
				"putCustomEncoder success", func() {
					enc := &customEncoder{
						buf: _customBufferPool.Get(),
					}
					convey.So(func() {
						putCustomEncoder(enc)
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestClone(t *testing.T) {
	convey.Convey(
		"Test Clone", t, func() {
			convey.Convey(
				"Clone success", func() {
					ce := &customEncoder{
						buf: _customBufferPool.Get(),
					}
					ce.buf.Write([]byte{0})
					clone := ce.Clone()
					convey.So(clone, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestEncodeEntry(t *testing.T) {
	convey.Convey(
		"Test EncodeEntry", t, func() {
			convey.Convey(
				"EncodeEntry success", func() {
					os.Setenv(constants.PodNameEnvKey, "pod_name1")
					enc := &customEncoder{
						EncoderConfig: &zapcore.EncoderConfig{},
						buf:           _customBufferPool.Get(),
						podName:       getPodName(),
						traceID:       "traceID1",
					}
					enc.buf.Write([]byte{0})
					// enc.EncodeCaller = func(caller zapcore.EntryCaller, aEnc zapcore.PrimitiveArrayEncoder){
					// 	return
					// }
					field := zapcore.Field{
						Key:    "key",
						Type:   zapcore.StringType,
						String: "value",
					}
					ent := zapcore.Entry{
						Level:      zapcore.InfoLevel,
						Time:       time.Now(),
						LoggerName: "loggerName1",
						Message:    "msg1",
						Caller:     zapcore.EntryCaller{Defined: false},
						Stack:      "Stack1",
					}
					convey.So(func() {
						enc.EncodeEntry(ent, []zapcore.Field{field})
					}, convey.ShouldPanic)
				},
			)
		},
	)
}

func TestEncoderAdd(t *testing.T) {
	convey.Convey(
		"Test EncoderAdd", t, func() {
			enc := &customEncoder{
				buf: _customBufferPool.Get(),
			}
			convey.Convey(
				"AddArray success", func() {
					err := enc.AddArray("key", nil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"AddObject success", func() {
					err := enc.AddObject("key", nil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"AddBinary success", func() {
					convey.So(func() {
						enc.AddBinary("key", []byte(""))
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddByteString success", func() {
					convey.So(func() {
						enc.AddByteString("traceID", []byte(""))
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddBool success", func() {
					convey.So(func() {
						enc.AddBool("key", true)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddComplex128 success", func() {
					convey.So(func() {
						enc.AddComplex128("key", 0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddComplex64 success", func() {
					convey.So(func() {
						enc.AddComplex64("key", 0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddDuration success", func() {
					convey.So(func() {
						enc.AddDuration("key", time.Duration(1))
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddFloat64 success", func() {
					convey.So(func() {
						enc.AddFloat64("key", 1.0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddFloat32 success", func() {
					convey.So(func() {
						enc.AddFloat32("key", 1.0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddInt success", func() {
					convey.So(func() {
						enc.AddInt("key", 1)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddInt64 success", func() {
					convey.So(func() {
						enc.AddInt64("key", 1)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddInt32 success", func() {
					convey.So(func() {
						enc.AddInt32("key", 1)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddInt16 success", func() {
					convey.So(func() {
						enc.AddInt16("key", 1)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddInt8 success", func() {
					convey.So(func() {
						enc.AddInt8("key", 1)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddString success", func() {
					convey.So(func() {
						enc.AddString("key", " ")
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddTime success", func() {
					convey.So(func() {
						enc.AddTime("key", time.Now())
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddUint success", func() {
					convey.So(func() {
						enc.AddUint("key", 0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddUint64 success", func() {
					convey.So(func() {
						enc.AddUint64("key", 0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddUint32 success", func() {
					convey.So(func() {
						enc.AddUint32("key", 0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddUint16 success", func() {
					convey.So(func() {
						enc.AddUint16("key", 0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddUint8 success", func() {
					convey.So(func() {
						enc.AddUint8("key", 0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddUintptr success", func() {
					convey.So(func() {
						s := "hello"
						addr := unsafe.Pointer(&s)
						ptr := uintptr(addr)
						enc.AddUintptr("key", ptr)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddReflected success", func() {
					err := enc.AddReflected("key", struct{}{})
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestOpenNamespace(t *testing.T) {
	convey.Convey(
		"Test OpenNamespace", t, func() {
			convey.Convey(
				"OpenNamespace success", func() {
					enc := NewCustomEncoder(&zapcore.EncoderConfig{})
					convey.So(func() {
						enc.OpenNamespace("key")
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestEncoderAppend(t *testing.T) {
	convey.Convey(
		"Test EncoderAppend", t, func() {
			enc := &customEncoder{
				buf: _customBufferPool.Get(),
			}
			convey.Convey(
				"AppendBool success", func() {
					convey.So(func() {
						enc.AppendBool(true)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AppendByteString success", func() {
					convey.So(func() {
						enc.AppendByteString([]byte("value"))
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AppendComplex128 success", func() {
					convey.So(func() {
						enc.AppendComplex128(0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AppendComplex64 success", func() {
					convey.So(func() {
						enc.AppendComplex64(0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AppendFloat32 success", func() {
					convey.So(func() {
						enc.AppendFloat32(1.0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"appendFloat success", func() {
					convey.So(func() {
						enc.appendFloat(math.NaN(), 8)
					}, convey.ShouldNotPanic)
					convey.So(func() {
						enc.appendFloat(math.Inf(1), 8)
					}, convey.ShouldNotPanic)
					convey.So(func() {
						enc.appendFloat(math.Inf(-1), 8)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AppendInt success", func() {
					convey.So(func() {
						enc.AppendInt(0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AppendInt32 success", func() {
					convey.So(func() {
						enc.AppendInt32(0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AppendInt16 success", func() {
					convey.So(func() {
						enc.AppendInt16(0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AppendInt8 success", func() {
					convey.So(func() {
						enc.AppendInt8(0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AppendUint success", func() {
					convey.So(func() {
						enc.AppendUint(0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AppendUint32 success", func() {
					convey.So(func() {
						enc.AppendUint32(0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AppendUint16 success", func() {
					convey.So(func() {
						enc.AppendUint16(0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AppendUint8 success", func() {
					convey.So(func() {
						enc.AppendUint8(0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AppendUintptr success", func() {
					convey.So(func() {
						s := "hello"
						addr := unsafe.Pointer(&s)
						ptr := uintptr(addr)
						enc.AppendUintptr(ptr)
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}
