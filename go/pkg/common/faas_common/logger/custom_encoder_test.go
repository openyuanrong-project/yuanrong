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

package logger

import (
	"math"
	"os"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
)

func TestNewCustomEncoder(t *testing.T) {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:      "T",
		LevelKey:     "L",
		NameKey:      "Logger",
		MessageKey:   "M",
		CallerKey:    "C",
		LineEnding:   zapcore.DefaultLineEnding,
		EncodeLevel:  zapcore.CapitalLevelEncoder,
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	}
	encoder := NewCustomEncoder(&encoderConfig)
	clone := encoder.Clone()
	assert.NotEmpty(t, clone)
	encoder.AddBool("3", true)
	err := encoder.AddArray("4", nil)
	assert.Empty(t, err)
	err = encoder.AddObject("4", nil)
	assert.Empty(t, err)
	encoder.AddBinary("4", []byte{})
	encoder.AddComplex128("4", complex(1, 2))
	encoder.AddComplex64("4", complex(1, 2))
	encoder.AddDuration("4", time.Second)
	encoder.AddByteString("4", []byte{})
	encoder.AddFloat64("2", 3.14)
	encoder.AddFloat32("2", 3.14)
	encoder.AddInt("1", 1)
	encoder.AddInt8("1", 1)
	encoder.AddInt16("1", 1)
	encoder.AddInt32("1", 1)
	encoder.AddInt64("1", 1)
	encoder.AddString("5", "12")
	encoder.AddString("5", "1 2")
	encoder.AddTime("6", time.Time{})
	encoder.AddUint("1", uint(1))
	encoder.AddUint8("1", uint8(10))
	encoder.AddUint16("1", uint16(100))
	encoder.AddUint32("1", uint32(1000))
	encoder.AddUint64("1", uint64(1000))
	b := make([]int, 1)
	encoder.AddUintptr("12", uintptr(unsafe.Pointer(&b[0])))
	encoder.OpenNamespace("3")
	err = encoder.AddReflected("3", 1)
	assert.Empty(t, err)

}

func Test_customEncoder_Append(t *testing.T) {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:      "T",
		LevelKey:     "L",
		NameKey:      "Logger",
		MessageKey:   "M",
		CallerKey:    "C",
		LineEnding:   zapcore.DefaultLineEnding,
		EncodeLevel:  zapcore.CapitalLevelEncoder,
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	}
	encoder := &customEncoder{
		EncoderConfig: &encoderConfig,
		buf:           _customBufferPool.Get(),
		podName:       os.Getenv(constant.HostNameEnvKey),
	}
	encoder.AppendInt16(1)
	encoder.AppendUint32(2)
	encoder.AppendByteString([]byte("abc"))
	encoder.AppendFloat32(3)
	encoder.appendFloat(math.Inf(1), 10)
	encoder.appendFloat(math.Inf(-1), 10)
	encoder.AppendComplex64(1 + 2i)
	encoder.AppendUintptr(uintptr(1))
	encoder.AppendUint8(1)
	encoder.AppendInt32(2)
	encoder.AppendUint16(3)
	encoder.AppendUint(4)
	encoder.AppendInt8(0)
	encoder.AppendInt(5)
	encoder.AppendInt32(7)
	assert.NotEmpty(t, encoder.buf.Len())
}
