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

// Package logger log
package logger

import (
	"errors"
	"math"
	"os"
	"sync"
	"time"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"

	"yuanrong.org/kernel/runtime/libruntime/common/constants"
)

var (
	_bufferPool = buffer.NewPool()

	_interfacePool = sync.Pool{New: interfaceFunc}
)

// InterfaceEncoderConfig holds interface log encoder config
type InterfaceEncoderConfig struct {
	ModuleName   string
	HTTPMethod   string
	ModuleFrom   string
	TenantID     string
	FuncName     string
	FuncVer      string
	EncodeCaller zapcore.CallerEncoder
}

// interfaceEncoder represents the encoder for interface log
// project's interface log
type interfaceEncoder struct {
	*InterfaceEncoderConfig
	buf     *buffer.Buffer
	podName string
	spaced  bool
}

func interfaceFunc() interface{} {
	return &interfaceEncoder{}
}

func getInterfaceEncoder() *interfaceEncoder {
	return _interfacePool.Get().(*interfaceEncoder)
}

func putInterfaceEncoder(enc *interfaceEncoder) {
	enc.InterfaceEncoderConfig = nil
	enc.spaced = false
	enc.buf = nil
	_interfacePool.Put(enc)
}

// NewInterfaceEncoder create a new interface log encoder
func NewInterfaceEncoder(cfg InterfaceEncoderConfig, spaced bool) zapcore.Encoder {
	return newInterfaceEncoder(cfg, spaced)
}

func newInterfaceEncoder(cfg InterfaceEncoderConfig, spaced bool) *interfaceEncoder {
	return &interfaceEncoder{
		InterfaceEncoderConfig: &cfg,
		buf:                    _bufferPool.Get(),
		spaced:                 spaced,
		podName:                os.Getenv(constants.HostNameEnvKey),
	}
}

// Clone return zap core Encoder
func (enc *interfaceEncoder) Clone() zapcore.Encoder {
	return enc.clone()
}

func (enc *interfaceEncoder) clone() *interfaceEncoder {
	clone := getInterfaceEncoder()
	clone.InterfaceEncoderConfig = enc.InterfaceEncoderConfig
	clone.spaced = enc.spaced
	clone.buf = _bufferPool.Get()
	return clone
}

// EncodeEntry Encode Entry
func (enc *interfaceEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	final := enc.clone()
	final.buf.AppendByte('[')
	// add time
	final.AppendString(ent.Time.UTC().Format("2006-01-02 15:04:05.000"))
	// add level
	// Level of interfaceLog is eternally INFO
	final.buf.AppendByte(headerSeparator)
	final.AppendString("INFO")
	// add caller
	if ent.Caller.Defined {
		final.buf.AppendByte(headerSeparator)
		final.EncodeCaller(ent.Caller, final)
	}
	final.buf.AppendByte(']')
	final.buf.AppendByte(headerSeparator)
	final.buf.AppendByte('[')
	// add podName
	if enc.podName != "" {
		final.buf.AppendString(enc.podName)
	}
	final.buf.AppendByte(']')
	final.buf.AppendByte(headerSeparator)
	if enc.buf.Len() > 0 {
		final.buf.Write(enc.buf.Bytes())
	}
	// add msg
	final.AppendString(ent.Message)
	for _, field := range fields {
		field.AddTo(final)
	}
	final.buf.AppendString(customDefaultLineEnding)
	ret := final.buf
	putInterfaceEncoder(final)
	return ret, nil
}

// AddString Append String
func (enc *interfaceEncoder) AddString(key, val string) {
	enc.buf.AppendString(val)
}

// AppendString Append String
func (enc *interfaceEncoder) AppendString(val string) {
	enc.buf.AppendString(val)
}

// AddDuration Add Duration
func (enc *interfaceEncoder) AddDuration(key string, val time.Duration) {
	enc.AppendDuration(val)
}

func (enc *interfaceEncoder) addElementSeparator() {
	last := enc.buf.Len() - 1
	if last < 0 {
		return
	}
	switch enc.buf.Bytes()[last] {
	case headerSeparator:
		return
	default:
		enc.buf.AppendByte(headerSeparator)
		if enc.spaced {
			enc.buf.AppendByte(' ')
		}
	}
}

// AppendTime Append Time
func (enc *interfaceEncoder) AppendTime(val time.Time) {
	cur := enc.buf.Len()
	interfaceTimeEncode(val, enc)
	if cur == enc.buf.Len() {
		// User-supplied EncodeTime is a no-op. Fall back to nanos since epoch to keep
		// output JSON valid.
		enc.AppendInt64(val.UnixNano())
	}
}

// AddArray Add Array
func (enc *interfaceEncoder) AddArray(key string, marshaler zapcore.ArrayMarshaler) error {
	return errors.New("unsupported method")
}

// AddObject Add Object
func (enc *interfaceEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error {
	return errors.New("unsupported method")
}

// AddBinary Add Binary
func (enc *interfaceEncoder) AddBinary(key string, value []byte) {}

// AddByteString Add Byte String
func (enc *interfaceEncoder) AddByteString(key string, val []byte) {
	enc.AppendByteString(val)
}

// AddBool Add Bool
func (enc *interfaceEncoder) AddBool(key string, value bool) {}

// AddComplex64 Add Complex64
func (enc *interfaceEncoder) AddComplex64(k string, v complex64) { enc.AddComplex128(k, complex128(v)) }

// AddFloat32 Add Float32
func (enc *interfaceEncoder) AddFloat32(k string, v float32) { enc.AddFloat64(k, float64(v)) }

// AddInt Add Int
func (enc *interfaceEncoder) AddInt(k string, v int) { enc.AddInt64(k, int64(v)) }

// AddInt32 Add Int32
func (enc *interfaceEncoder) AddInt32(k string, v int32) { enc.AddInt64(k, int64(v)) }

// AddInt16 Add Int16
func (enc *interfaceEncoder) AddInt16(k string, v int16) { enc.AddInt64(k, int64(v)) }

// AddInt8 Add Int8
func (enc *interfaceEncoder) AddInt8(k string, v int8) { enc.AddInt64(k, int64(v)) }

// AddUint Add Uint
func (enc *interfaceEncoder) AddUint(k string, v uint) { enc.AddUint64(k, uint64(v)) }

// AddUint32 Add Uint32
func (enc *interfaceEncoder) AddUint32(k string, v uint32) { enc.AddUint64(k, uint64(v)) }

// AddUint16 Add Uint16
func (enc *interfaceEncoder) AddUint16(k string, v uint16) { enc.AddUint64(k, uint64(v)) }

// AddUint8 Add Uint8
func (enc *interfaceEncoder) AddUint8(k string, v uint8) { enc.AddUint64(k, uint64(v)) }

// AddUintptr Add Uint ptr
func (enc *interfaceEncoder) AddUintptr(k string, v uintptr) { enc.AddUint64(k, uint64(v)) }

// AddComplex128 Add Complex128
func (enc *interfaceEncoder) AddComplex128(key string, val complex128) {
	enc.AppendComplex128(val)
}

// AddFloat64 Add Float64
func (enc *interfaceEncoder) AddFloat64(key string, val float64) {
	enc.AppendFloat64(val)
}

// AddInt64 Add Int64
func (enc *interfaceEncoder) AddInt64(key string, val int64) {
	enc.AppendInt64(val)
}

// AddTime Add Time
func (enc *interfaceEncoder) AddTime(key string, value time.Time) {
	enc.AppendTime(value)
}

// AddUint64 Add Uint64
func (enc *interfaceEncoder) AddUint64(key string, value uint64) {}

// AddReflected uses reflection to serialize arbitrary objects, so it's slow
// and allocation-heavy.
func (enc *interfaceEncoder) AddReflected(key string, value interface{}) error {
	return nil
}

// OpenNamespace opens an isolated namespace where all subsequent fields will
// be added. Applications can use namespaces to prevent key collisions when
// injecting loggers into sub-components or third-party libraries.
func (enc *interfaceEncoder) OpenNamespace(key string) {}

// AppendComplex128 Append Complex128
func (enc *interfaceEncoder) AppendComplex128(val complex128) {}

// AppendInt64 Append Int64
func (enc *interfaceEncoder) AppendInt64(val int64) {
	enc.addElementSeparator()
	enc.buf.AppendInt(val)
}

// AppendBool Append Bool
func (enc *interfaceEncoder) AppendBool(val bool) {
	enc.addElementSeparator()
	enc.buf.AppendBool(val)
}

func (enc *interfaceEncoder) appendFloat(val float64, bitSize int) {
	enc.addElementSeparator()
	switch {
	case math.IsNaN(val):
		enc.buf.AppendString(`"NaN"`)
	case math.IsInf(val, 1):
		enc.buf.AppendString(`"+Inf"`)
	case math.IsInf(val, -1):
		enc.buf.AppendString(`"-Inf"`)
	default:
		enc.buf.AppendFloat(val, bitSize)
	}
}

// AppendUint64 Append Uint64
func (enc *interfaceEncoder) AppendUint64(val uint64) {
	enc.addElementSeparator()
	enc.buf.AppendUint(val)
}

// AppendByteString Append Byte String
func (enc *interfaceEncoder) AppendByteString(val []byte) {}

// AppendDuration Append Duration
func (enc *interfaceEncoder) AppendDuration(val time.Duration) {}

// AppendComplex64 Append Complex64
func (enc *interfaceEncoder) AppendComplex64(v complex64) { enc.AppendComplex128(complex128(v)) }

// AppendFloat64 Append Float64
func (enc *interfaceEncoder) AppendFloat64(v float64) { enc.appendFloat(v, float64bitSize) }

// AppendFloat32 Append Float32
func (enc *interfaceEncoder) AppendFloat32(v float32) { enc.appendFloat(float64(v), float32bitSize) }

// AppendInt Append Int
func (enc *interfaceEncoder) AppendInt(v int) { enc.AppendInt64(int64(v)) }

// AppendInt32 Append Int32
func (enc *interfaceEncoder) AppendInt32(v int32) { enc.AppendInt64(int64(v)) }

// AppendInt16 Append Int16
func (enc *interfaceEncoder) AppendInt16(v int16) { enc.AppendInt64(int64(v)) }

// AppendInt8 Append Int8
func (enc *interfaceEncoder) AppendInt8(v int8) { enc.AppendInt64(int64(v)) }

// AppendUint Append Uint
func (enc *interfaceEncoder) AppendUint(v uint) { enc.AppendUint64(uint64(v)) }

// AppendUint32 Append Uint32
func (enc *interfaceEncoder) AppendUint32(v uint32) { enc.AppendUint64(uint64(v)) }

// AppendUint16 Append Uint16
func (enc *interfaceEncoder) AppendUint16(v uint16) { enc.AppendUint64(uint64(v)) }

// AppendUint8 Append Uint8
func (enc *interfaceEncoder) AppendUint8(v uint8) { enc.AppendUint64(uint64(v)) }

// AppendUintptr Append Uint ptr
func (enc *interfaceEncoder) AppendUintptr(v uintptr) { enc.AppendUint64(uint64(v)) }

func interfaceTimeEncode(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	t = t.UTC()
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}
