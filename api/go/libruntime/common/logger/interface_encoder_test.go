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
	"testing"
	"time"
	"unsafe"

	"github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

func TestInterfaceFunc(t *testing.T) {
	convey.Convey(
		"Test interfaceFunc", t, func() {
			convey.Convey(
				"interfaceFunc success", func() {
					enc := interfaceFunc()
					convey.So(enc, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestGetPutInterfaceEncoder(t *testing.T) {
	convey.Convey(
		"Test GetPutInterfaceEncoder", t, func() {
			enc := getInterfaceEncoder()
			convey.Convey(
				"getInterfaceEncoder success", func() {
					convey.So(enc, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"putInterfaceEncoder success", func() {
					convey.So(func() {
						putInterfaceEncoder(enc)
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestNewInterfaceEncoder(t *testing.T) {
	convey.Convey(
		"Test NewInterfaceEncoder", t, func() {
			convey.Convey(
				"NewInterfaceEncoder success", func() {
					enc := NewInterfaceEncoder(InterfaceEncoderConfig{}, false)
					convey.So(enc, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestInterfaceEncoderClone(t *testing.T) {
	convey.Convey(
		"Test InterfaceEncoderClone", t, func() {
			convey.Convey(
				"Clone success", func() {
					enc := NewInterfaceEncoder(InterfaceEncoderConfig{}, false)
					clone := enc.Clone()
					convey.So(clone, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestInterfaceEncodeEntry(t *testing.T) {
	convey.Convey(
		"Test InterfaceEncodeEntry", t, func() {
			convey.Convey(
				"EncodeEntry success", func() {
					enc := newInterfaceEncoder(InterfaceEncoderConfig{}, false)
					enc.buf.Write([]byte{0})
					enc.EncodeCaller = func(caller zapcore.EntryCaller, aEnc zapcore.PrimitiveArrayEncoder) {
						return
					}
					field := zapcore.Field{
						Key:    "key",
						Type:   zapcore.StringType,
						String: "value",
					}
					ent := zapcore.Entry{Caller: zapcore.EntryCaller{Defined: true}}
					buf, err := enc.EncodeEntry(ent, []zapcore.Field{field})
					convey.So(buf, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestInterfaceEncoderAdd(t *testing.T) {
	convey.Convey(
		"Test InterfaceEncoderAdd", t, func() {
			enc := NewInterfaceEncoder(InterfaceEncoderConfig{}, false)
			convey.Convey(
				"AddString success", func() {
					convey.So(func() {
						enc.AddString("key", "val")
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddDuration success", func() {
					convey.So(func() {
						enc.AddDuration("key", time.Second)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"addElementSeparator success", func() {
					ent := &interfaceEncoder{
						buf: &buffer.Buffer{},
					}
					ent.buf.Write([]byte("hello"))
					ent.spaced = true
					convey.So(ent.addElementSeparator, convey.ShouldNotPanic)
					ent.buf.Write([]byte{' '})
					convey.So(ent.addElementSeparator, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddArray success", func() {
					var f zapcore.ArrayMarshalerFunc = func(zapcore.ArrayEncoder) error {
						return nil
					}
					convey.So(func() {
						enc.AddArray("key", f)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddObject success", func() {
					var f zapcore.ObjectMarshalerFunc = func(zapcore.ObjectEncoder) error {
						return nil
					}
					convey.So(func() {
						enc.AddObject("key", f)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddBinary success", func() {
					convey.So(func() {
						enc.AddBinary("key", []byte("val"))
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddByteString success", func() {
					convey.So(func() {
						enc.AddByteString("key", []byte("val"))
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
				"AddComplex64 success", func() {
					convey.So(func() {
						enc.AddComplex64("key", 0)
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
						enc.AddInt("key", 0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddInt32 success", func() {
					convey.So(func() {
						enc.AddInt32("key", 0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddInt16 success", func() {
					convey.So(func() {
						enc.AddInt16("key", 0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddInt8 success", func() {
					convey.So(func() {
						enc.AddInt8("key", 0)
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
				"AddComplex128 success", func() {
					convey.So(func() {
						enc.AddComplex128("key", 0)
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
				"AddInt64 success", func() {
					convey.So(func() {
						enc.AddInt64("key", 0)
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
				"AddUint64 success", func() {
					convey.So(func() {
						enc.AddUint64("key", 0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AddReflected success", func() {
					convey.So(func() {
						enc.AddReflected("key", struct{}{})
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestInterfaceEncoderAppend(t *testing.T) {
	convey.Convey(
		"Test InterfaceEncoderAppend", t, func() {
			enc := newInterfaceEncoder(InterfaceEncoderConfig{}, false)
			convey.Convey(
				"OpenNamespace success", func() {
					convey.So(func() {
						enc.OpenNamespace("key")
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AppendBool success", func() {
					convey.So(func() {
						enc.AppendBool(false)
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
				"AppendUint64 success", func() {
					convey.So(func() {
						enc.AppendUint64(0)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"AppendFloat32 success", func() {
					convey.So(func() {
						enc.AppendFloat32(0)
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
				"AppendComplex64 success", func() {
					convey.So(func() {
						enc.AppendComplex64(0)
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
