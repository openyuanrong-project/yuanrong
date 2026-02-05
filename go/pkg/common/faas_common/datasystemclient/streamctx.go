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

package datasystemclient

import (
	"bufio"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/valyala/fasthttp"
)

// StreamCtx -
type StreamCtx interface {
	GetRequestHeader(key string) string
	SetResponseHeader(key, value string)
	Stream(writer func(w io.Writer) bool)
	FlushResult(w io.Writer, result []byte) error
	Done() <-chan struct{}
}

// GinCtxAdapter -
type GinCtxAdapter struct {
	*gin.Context
}

// GetRequestHeader -
func (gt *GinCtxAdapter) GetRequestHeader(key string) string {
	return gt.Request.Header.Get(key)
}

// SetResponseHeader -
func (gt *GinCtxAdapter) SetResponseHeader(key, value string) {
	gt.Writer.Header().Set(key, value)
}

// Stream -
func (gt *GinCtxAdapter) Stream(writer func(w io.Writer) bool) {
	gt.Context.Stream(writer)
}

// FlushResult -
func (gt *GinCtxAdapter) FlushResult(w io.Writer, result []byte) error {
	_, err := w.Write(result)
	if err != nil {
		return err
	}
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	return nil
}

// Done -
func (gt *GinCtxAdapter) Done() <-chan struct{} {
	return gt.Request.Context().Done()
}

// FastHttpCtxAdapter -
type FastHttpCtxAdapter struct {
	*fasthttp.RequestCtx
}

// GetRequestHeader -
func (ft *FastHttpCtxAdapter) GetRequestHeader(key string) string {
	return string(ft.Request.Header.Peek(key))
}

// SetResponseHeader -
func (ft *FastHttpCtxAdapter) SetResponseHeader(key, value string) {
	ft.Response.Header.Set(key, value)
}

// Stream -
func (ft *FastHttpCtxAdapter) Stream(writer func(w io.Writer) bool) {
	ft.SetBodyStreamWriter(func(w *bufio.Writer) {
		writer(w)
	})
}

func (ft *FastHttpCtxAdapter) FlushResult(w io.Writer, result []byte) error {
	_, err := w.Write(result)
	if err != nil {
		return err
	}
	if f, ok := w.(*bufio.Writer); ok {
		// When the client is disconnected, return and close consumer.
		if err = f.Flush(); err != nil {
			return err
		}
	}
	return nil
}

// Done -
func (ft *FastHttpCtxAdapter) Done() <-chan struct{} {
	return ft.RequestCtx.Done()
}
