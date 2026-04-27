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
	"math"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"

	"yuanrong.org/kernel/runtime/libruntime/common/constants"
)

const (
	float64bitSize          = 64
	float32bitSize          = 32
	headerSeparator         = ' '
	elementSeparator        = "  "
	customDefaultLineEnding = "\n"
	logMsgMaxLen            = 1024 * 2
	fieldSeparator          = " | "
)

var (
	_customBufferPool = buffer.NewPool()

	_customPool = sync.Pool{New: func() interface{} {
		return &customEncoder{}
	}}

	replComp = regexp.MustCompile(`\s+`)

	clusterName = os.Getenv("CLUSTER_ID")
)

// customEncoder represents the encoder for zap logger
// project's interface log
type customEncoder struct {
	*zapcore.EncoderConfig
	buf     *buffer.Buffer
	podName string
	traceID string
}

func getPodName() string {
	podName := os.Getenv(constants.HostNameEnvKey)
	if os.Getenv(constants.PodNameEnvKey) != "" {
		podName = os.Getenv(constants.PodNameEnvKey)
	}
	return podName
}

// NewConsoleEncoder new custom console encoder to zap log module
func NewConsoleEncoder(cfg zapcore.EncoderConfig) (zapcore.Encoder, error) {
	return &customEncoder{
		EncoderConfig: &cfg,
		buf:           _customBufferPool.Get(),
		podName:       getPodName(),
	}, nil
}

// NewCustomEncoder new custom encoder to zap log module
func NewCustomEncoder(cfg *zapcore.EncoderConfig) zapcore.Encoder {
	return &customEncoder{
		EncoderConfig: cfg,
		buf:           _customBufferPool.Get(),
		podName:       getPodName(),
	}
}

// Clone return zap core Encoder
func (enc *customEncoder) Clone() zapcore.Encoder {
	clone := enc.clone()
	if enc.buf.Len() > 0 {
		_, _ = clone.buf.Write(enc.buf.Bytes())
	}
	return clone
}

// EncodeEntry -
func (enc *customEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	final := enc.clone()
	// add time
	final.AppendString(ent.Time.UTC().Format("2006-01-02 15:04:05.000"))
	final.buf.AppendString(fieldSeparator)

	final.EncodeLevel(ent.Level, final)
	final.buf.AppendString(fieldSeparator)

	// add caller
	if ent.Caller.Defined {
		final.EncodeCaller(ent.Caller, final)
		final.buf.AppendString(fieldSeparator)
	}
	// add podName
	if enc.podName != "" {
		final.buf.AppendString(enc.podName)
	}
	final.buf.AppendString(fieldSeparator)
	// add clusterName
	if clusterName != "" {
		final.buf.AppendString(clusterName)
	}
	final.buf.AppendString(fieldSeparator)
	if enc.traceID != "" {
		final.buf.AppendString(enc.traceID)
	}
	final.buf.AppendString(fieldSeparator)
	if enc.buf.Len() > 0 {
		final.buf.Write(enc.buf.Bytes())
	}
	// add msg
	if len(ent.Message) > logMsgMaxLen {
		final.AppendString(ent.Message[0:logMsgMaxLen])
	} else {
		final.AppendString(ent.Message)
	}
	if ent.Stack != "" && final.StacktraceKey != "" {
		final.buf.AppendString(elementSeparator)
		final.AddString(final.StacktraceKey, ent.Stack)
	}
	for _, field := range fields {
		field.AddTo(final)
	}
	final.buf.AppendString(customDefaultLineEnding)
	ret := final.buf
	putCustomEncoder(final)
	return ret, nil
}

func putCustomEncoder(enc *customEncoder) {
	enc.EncoderConfig = nil
	enc.buf = nil
	_customPool.Put(enc)
}

func getCustomEncoder() *customEncoder {
	return _customPool.Get().(*customEncoder)
}

func (enc *customEncoder) clone() *customEncoder {
	clone := getCustomEncoder()
	clone.buf = _customBufferPool.Get()
	clone.EncoderConfig = enc.EncoderConfig
	clone.podName = enc.podName
	clone.traceID = enc.traceID
	return clone
}

func (enc *customEncoder) writeField(k string, writeVal func()) *customEncoder {
	enc.buf.AppendString("(" + k + ":")
	writeVal()
	enc.buf.AppendString(")")
	return enc
}

// AddArray Add Array
func (enc *customEncoder) AddArray(k string, marshaler zapcore.ArrayMarshaler) error {
	return nil
}

// AddObject Add Object
func (enc *customEncoder) AddObject(k string, marshaler zapcore.ObjectMarshaler) error {
	return nil
}

// AddBinary Add Binary
func (enc *customEncoder) AddBinary(k string, v []byte) {
	enc.AddString(k, string(v))
}

// AddByteString Add Byte String
func (enc *customEncoder) AddByteString(k string, v []byte) {
	enc.AddString(k, string(v))
}

// AddBool Add Bool
func (enc *customEncoder) AddBool(k string, v bool) {
	enc.writeField(k, func() {
		enc.AppendBool(v)
	})
}

// AddComplex128 Add Complex128
func (enc *customEncoder) AddComplex128(k string, val complex128) {}

// AddComplex64 Add Complex64
func (enc *customEncoder) AddComplex64(k string, v complex64) {}

// AddDuration Add Duration
func (enc *customEncoder) AddDuration(k string, val time.Duration) {
	enc.writeField(k, func() {
		enc.AppendString(val.String())
	})
}

// AddFloat64 Add Float64
func (enc *customEncoder) AddFloat64(k string, val float64) {
	enc.writeField(k, func() {
		enc.AppendFloat64(val)
	})
}

// AddFloat32 Add Float32
func (enc *customEncoder) AddFloat32(k string, v float32) {
	enc.writeField(k, func() {
		enc.AppendFloat64(float64(v))
	})
}

// AddInt Add Int
func (enc *customEncoder) AddInt(k string, v int) {
	enc.writeField(k, func() {
		enc.AppendInt64(int64(v))
	})
}

// AddInt64 Add Int64
func (enc *customEncoder) AddInt64(k string, val int64) {
	enc.writeField(k, func() {
		enc.AppendInt64(val)
	})
}

// AddInt32 Add Int32
func (enc *customEncoder) AddInt32(k string, v int32) {
	enc.writeField(k, func() {
		enc.AppendInt64(int64(v))
	})
}

// AddInt16 Add Int16
func (enc *customEncoder) AddInt16(k string, v int16) {
	enc.writeField(k, func() {
		enc.AppendInt64(int64(v))
	})
}

// AddInt8 Add Int8
func (enc *customEncoder) AddInt8(k string, v int8) {
	enc.writeField(k, func() {
		enc.AppendInt64(int64(v))
	})
}

// AddString Append String
func (enc *customEncoder) AddString(k, v string) {
	if k == "traceID" || k == "traceId" {
		enc.traceID = v
		return
	}
	enc.writeField(k, func() {
		v = replComp.ReplaceAllString(v, " ")
		if strings.Contains(v, " ") {
			enc.buf.AppendString("(" + v + ")")
			return
		}
		enc.AppendString(v)
	})
}

// AddTime Add Time
func (enc *customEncoder) AddTime(k string, v time.Time) {
	enc.writeField(k, func() {
		enc.AppendString(v.UTC().Format("2006-01-02 15:04:05.000"))
	})
}

// AddUint Add Uint
func (enc *customEncoder) AddUint(k string, v uint) {
	enc.writeField(k, func() {
		enc.AppendUint64(uint64(v))
	})
}

// AddUint64 Add Uint64
func (enc *customEncoder) AddUint64(k string, v uint64) {
	enc.writeField(k, func() {
		enc.AppendUint64(v)
	})
}

// AddUint32 Add Uint32
func (enc *customEncoder) AddUint32(k string, v uint32) {
	enc.writeField(k, func() {
		enc.AppendUint64(uint64(v))
	})
}

// AddUint16 Add Uint16
func (enc *customEncoder) AddUint16(k string, v uint16) {
	enc.writeField(k, func() {
		enc.AppendUint64(uint64(v))
	})
}

// AddUint8 Add Uint8
func (enc *customEncoder) AddUint8(k string, v uint8) {
	enc.writeField(k, func() {
		enc.AppendUint64(uint64(v))
	})
}

// AddUintptr Add Uint ptr
func (enc *customEncoder) AddUintptr(k string, v uintptr) {
	enc.writeField(k, func() {
		enc.AppendUint64(uint64(v))
	})
}

// AddReflected uses reflection to serialize arbitrary objects, so it's slow
// and allocation-heavy.
func (enc *customEncoder) AddReflected(k string, v interface{}) error {
	return nil
}

// OpenNamespace opens an isolated namespace where all subsequent fields will
// be added. Applications can use namespaces to prevent key collisions when
// injecting loggers into sub-components or third-party libraries.
func (enc *customEncoder) OpenNamespace(k string) {}

// AppendBool Append Bool
func (enc *customEncoder) AppendBool(v bool) { enc.buf.AppendBool(v) }

// AppendByteString Append Byte String
func (enc *customEncoder) AppendByteString(v []byte) { enc.AppendString(string(v)) }

// AppendComplex128 Append Complex128
func (enc *customEncoder) AppendComplex128(v complex128) {}

// AppendComplex64 Append Complex64
func (enc *customEncoder) AppendComplex64(v complex64) {}

// AppendFloat64 Append Float64
func (enc *customEncoder) AppendFloat64(v float64) { enc.appendFloat(v, float64bitSize) }

// AppendFloat32 Append Float32
func (enc *customEncoder) AppendFloat32(v float32) { enc.appendFloat(float64(v), float32bitSize) }

func (enc *customEncoder) appendFloat(v float64, bitSize int) {
	switch {
	// If the condition is not met, a string is returned to prevent blankness.
	// IsNaN reports whether f is an IEEE 754 ``not-a-number'' value.
	case math.IsNaN(v):
		enc.buf.AppendString(`"NaN"`)
	// IsInf reports whether f is an infinity, according to sign
	case math.IsInf(v, 1):
		// IsInf reports whether f is positive infinity
		enc.buf.AppendString(`"+Inf"`)
	// IsInf reports whether f is negative infinity
	case math.IsInf(v, -1):
		enc.buf.AppendString(`"-Inf"`)
	default:
		enc.buf.AppendFloat(v, bitSize)
	}
}

// AppendInt Append Int
func (enc *customEncoder) AppendInt(v int) { enc.buf.AppendInt(int64(v)) }

// AppendInt64 Append Int64
func (enc *customEncoder) AppendInt64(v int64) { enc.buf.AppendInt(v) }

// AppendInt32 Append Int32
func (enc *customEncoder) AppendInt32(v int32) { enc.buf.AppendInt(int64(v)) }

// AppendInt16 Append Int16
func (enc *customEncoder) AppendInt16(v int16) { enc.buf.AppendInt(int64(v)) }

// AppendInt8 Append Int8
func (enc *customEncoder) AppendInt8(v int8) { enc.buf.AppendInt(int64(v)) }

// AppendString Append String
func (enc *customEncoder) AppendString(val string) { enc.buf.AppendString(val) }

// AppendUint Append Uint
func (enc *customEncoder) AppendUint(v uint) { enc.buf.AppendUint(uint64(v)) }

// AppendUint64 Append Uint64
func (enc *customEncoder) AppendUint64(v uint64) { enc.buf.AppendUint(v) }

// AppendUint32 Append Uint32
func (enc *customEncoder) AppendUint32(v uint32) { enc.buf.AppendUint(uint64(v)) }

// AppendUint16 Append Uint16
func (enc *customEncoder) AppendUint16(v uint16) { enc.buf.AppendUint(uint64(v)) }

// AppendUint8 Append Uint8
func (enc *customEncoder) AppendUint8(v uint8) { enc.buf.AppendUint(uint64(v)) }

// AppendUintptr Append Uint ptr
func (enc *customEncoder) AppendUintptr(v uintptr) { enc.buf.AppendUint(uint64(v)) }
