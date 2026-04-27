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

// Package client is define interface of client
package client

import (
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"

	fhttp "github.com/valyala/fasthttp"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	shttp "yuanrong.org/kernel/pkg/common/httputil/http"
	"yuanrong.org/kernel/pkg/common/httputil/http/client/fast"
)

const (
	// 默认最大重试次数
	defaultMaxRetryTimes = 3

	// MaxClientConcurrency is the max concurrency of fast http client
	MaxClientConcurrency = 1000

	// DialTimeOut -
	DialTimeOut = 10

	// TCPKeepAlivePeriod -
	TCPKeepAlivePeriod = 10
)

var tcpDialer = fhttp.TCPDialer{Concurrency: MaxClientConcurrency}

var globalTLSConf *tls.Config

// Client 客户端接口
type Client interface {
	PostMultipart(url string, params map[string]string,
		headers map[string]string, filePath string) (*shttp.SuccessResponse, error)
	Get(url string, headers map[string]string) (*shttp.SuccessResponse, error)
	PutMultipart(url string, params map[string]string,
		headers map[string]string, filePath string) (*shttp.SuccessResponse, error)
}

func adminDial(addr string) (net.Conn, error) {
	conn, err := tcpDialer.DialTimeout(addr, DialTimeOut*time.Second)
	if err != nil {
		log.GetLogger().Errorf("failed to dial %s, error: %s ", addr, err.Error())
		return nil, err
	}

	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		log.GetLogger().Errorf("failed to dial %s", addr)
		return nil, nil
	}
	err = tcpConn.SetKeepAlive(true)
	if err != nil {
		log.GetLogger().Errorf("failed to set connection keepalive %s, error: %s", addr, err.Error())
		return nil, err
	}

	err = tcpConn.SetKeepAlivePeriod(TCPKeepAlivePeriod * time.Second)
	if err != nil {
		log.GetLogger().Errorf("failed to set connection keepalive period %s, error: %s",
			addr, err.Error())
		return nil, err
	}

	return tcpConn, nil
}

// newClient 创建client
func newClient(tlsConf *tls.Config) Client {
	cli := &fast.FastClient{
		Client: &fhttp.Client{
			TLSConfig:                 tlsConf,
			MaxIdemponentCallAttempts: defaultMaxRetryTimes,
			ReadBufferSize:            http.DefaultMaxHeaderBytes,
			Dial:                      adminDial,
		}}
	return cli
}

var once sync.Once
var client Client

// GetInstance get client instance
func GetInstance() Client {
	once.Do(func() {
		client = newClient(globalTLSConf)
	})

	return client
}

// InitTlsConf init tls conf
func InitTlsConf(tlsConf *tls.Config) {
	globalTLSConf = tlsConf
}
