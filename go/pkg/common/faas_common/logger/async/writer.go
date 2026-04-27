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
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"go.uber.org/zap/buffer"
)

const (
	diskBufferSize     = 1024 * 1024
	diskFlushSize      = diskBufferSize >> 1
	diskFlushTime      = 500 * time.Millisecond
	defaultChannelSize = 200000
	softLimitFactor    = 0.8 // must be smaller than 1
)

var (
	linePool = buffer.NewPool()
)

// Opt -
type Opt func(*Writer)

// WithCachedLimit -
func WithCachedLimit(limit int) Opt {
	return func(w *Writer) {
		w.cachedLimit = limit
		w.cachedSoftLimit = int(float64(limit) * softLimitFactor)
		w.cachedLow = w.cachedSoftLimit >> 1
	}
}

// NewAsyncWriteSyncer wrappers io.Writer to async zapcore.WriteSyncer
func NewAsyncWriteSyncer(w io.Writer, opts ...Opt) *Writer {
	writer := &Writer{
		w:        w,
		diskBuf:  bytes.NewBuffer(make([]byte, 0, diskBufferSize)),
		lines:    make(chan *buffer.Buffer, defaultChannelSize),
		sync:     make(chan struct{}),
		syncDone: make(chan struct{}),
	}
	for _, opt := range opts {
		opt(writer)
	}
	go writer.logConsumer()
	return writer
}

// Writer -
type Writer struct {
	diskBuf  *bytes.Buffer
	lines    chan *buffer.Buffer
	w        io.Writer
	sync     chan struct{}
	syncDone chan struct{}

	cachedLimit     int
	cachedSoftLimit int
	cachedLow       int
	cached          int64 // atomic
}

// Write sends data to channel non-blocking
func (w *Writer) Write(data []byte) (int, error) {
	// note: data will be put back to zap's inner pool after Write, so we couldn't send it to channel directly
	lp := linePool.Get()
	lp.Write(data)
	select {
	case w.lines <- lp:
		if w.cachedLimit != 0 && atomic.AddInt64(&w.cached, int64(len(data))) > int64(w.cachedLimit) {
			w.doSync()
		}
	default:
		fmt.Println("failed to push log to channel, skip")
		lp.Free()
	}
	return len(data), nil
}

// Sync implements zapcore.WriteSyncer. Current do nothing.
func (w *Writer) Sync() error {
	w.doSync()
	return nil
}

func (w *Writer) doSync() {
	w.sync <- struct{}{}
	<-w.syncDone
}

func (w *Writer) logConsumer() {
	ticker := time.NewTicker(diskFlushTime)
loop:
	for {
		select {
		case line := <-w.lines:
			w.write(line)
			if w.cachedLimit != 0 && atomic.LoadInt64(&w.cached) > int64(w.cachedSoftLimit) {
				w.flushLines(len(w.lines), w.cachedLow)
			}
		case <-ticker.C:
			if w.diskBuf.Len() == 0 {
				continue
			}
			if _, err := w.w.Write(w.diskBuf.Bytes()); err != nil {
				fmt.Println("failed to write", err.Error())
			}
			w.diskBuf.Reset()
		case _, ok := <-w.sync:
			if !ok {
				close(w.syncDone)
				break loop
			}
			nLines := len(w.lines)
			if nLines == 0 && w.diskBuf.Len() == 0 {
				w.syncDone <- struct{}{}
				continue
			}
			w.flushLines(nLines, -1)
			if _, err := w.w.Write(w.diskBuf.Bytes()); err != nil {
				fmt.Println("failed to write", err.Error())
			}
			w.diskBuf.Reset()
			w.syncDone <- struct{}{}
		}
	}
	ticker.Stop()
}

func (w *Writer) flushLines(nLines int, upTo int) {
	nBytes := 0
	for i := 0; i < nLines; i++ {
		line := <-w.lines
		nBytes += line.Len()
		w.write(line)
		if upTo >= 0 && nBytes > upTo {
			break
		}
	}
}

func (w *Writer) write(line *buffer.Buffer) {
	w.diskBuf.Write(line.Bytes())
	if w.cachedLimit != 0 {
		atomic.AddInt64(&w.cached, -int64(line.Len()))
	}
	line.Free()
	if w.diskBuf.Len() < diskFlushSize {
		return
	}
	if _, err := w.w.Write(w.diskBuf.Bytes()); err != nil {
		fmt.Println("failed to write", err.Error())
	}
	w.diskBuf.Reset()
}
