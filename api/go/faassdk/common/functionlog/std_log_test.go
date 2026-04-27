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

package functionlog

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

type logCallback struct {
	res [][]byte
}

func (l *logCallback) callback(b []byte) {
	l.res = append(l.res, b)
}

func (l *logCallback) reset() {
	l.res = nil
}

func Test_stdLogReader(t *testing.T) {
	convey.Convey("Test stdLogReader", t, func() {
		convey.Convey("Test basic", func() {
			logger := CreateSTDLogger()
			cb := logCallback{}
			logger.RegisterLogCallback(cb.callback)
			log := []byte("123456789")
			count, err := logger.Write(log)
			assert.NoError(t, err)
			assert.Equal(t, len(log), count)
			assert.Equal(t, 1, len(cb.res))
			assert.Equal(t, log, cb.res[0])
		})
		convey.Convey("Test no content", func() {
			logger := CreateSTDLogger()
			cb := logCallback{}
			logger.RegisterLogCallback(cb.callback)
			count, err := logger.Write([]byte(""))
			assert.NoError(t, err)
			assert.Equal(t, 0, count)
			assert.Equal(t, 0, len(cb.res))
		})
		convey.Convey("Test no callback", func() {
			logger := CreateSTDLogger()
			_, err := logger.Write([]byte("1234567\n89"))
			assert.Error(t, err)
		})
		convey.Convey("Test large lines", func() {
			logger := CreateSTDLoggerWithSplitLimit(5)
			cb := logCallback{}
			logger.RegisterLogCallback(cb.callback)
			log := []byte("12345")
			count, err := logger.Write(log)
			assert.NoError(t, err)
			assert.Equal(t, len(log), count)
			assert.Equal(t, 1, len(cb.res))
			assert.Equal(t, log, cb.res[0])
			cb.reset()

			log = []byte("123456")
			count, err = logger.Write(log)
			assert.NoError(t, err)
			assert.Equal(t, len(log), count)
			assert.Equal(t, 2, len(cb.res))
			assert.Equal(t, []byte("12345"), cb.res[0])
			assert.Equal(t, []byte("6"), cb.res[1])
			cb.reset()

			log = []byte("1234567890\n")
			count, err = logger.Write(log)
			assert.NoError(t, err)
			assert.Equal(t, len(log), count)
			assert.Equal(t, 3, len(cb.res))
			assert.Equal(t, []byte("12345"), cb.res[0])
			assert.Equal(t, []byte("67890"), cb.res[1])
			assert.Equal(t, []byte("\n"), cb.res[2])
			cb.reset()
		})
	})
}
