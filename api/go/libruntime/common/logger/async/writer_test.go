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

// Package async makes io.Writer write async
package async

import (
	"bytes"
	"io"
	"testing"
	"time"

	"go.uber.org/zap/buffer"

	"github.com/smartystreets/goconvey/convey"
)

func TestWriter(t *testing.T) {
	convey.Convey(
		"Test Writer", t, func() {
			opt := WithCachedLimit(0)
			var ioW io.Writer = &bytes.Buffer{}
			w := NewAsyncWriteSyncer(ioW, opt)
			convey.Convey(
				"WithCachedLimit success", func() {
					convey.So(opt, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"NewAsyncWriteSyncer success", func() {
					convey.So(w, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"Write success", func() {
					w.cachedLimit = -1
					n, err := w.Write([]byte("data"))
					convey.So(n, convey.ShouldEqual, 4)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"Write success when default", func() {
					w.lines = make(chan *buffer.Buffer, 0)
					n, err := w.Write([]byte("data"))
					convey.So(n, convey.ShouldEqual, 4)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"Sync success", func() {
					err := w.Sync()
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"logConsumer success", func() {
					lp := linePool.Get()
					lp.Write([]byte("data"))
					w.lines <- lp
					convey.So(func() {
						go w.logConsumer()
						time.Sleep(500 * time.Millisecond)
						w.sync <- struct{}{}
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"write success", func() {
					lp := linePool.Get()
					for lp.Len() < diskFlushSize {
						lp.Write([]byte("data"))
					}
					convey.So(func() {
						w.write(lp)
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}
