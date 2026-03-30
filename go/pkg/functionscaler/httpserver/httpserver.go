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

// Package httpserver -
package httpserver

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/localauth"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/statuscode"
	commonTls "yuanrong.org/kernel/pkg/common/faas_common/tls"
	"yuanrong.org/kernel/pkg/functionscaler"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
)

var isShutDown atomic.Bool = atomic.Bool{}

const (
	defaultReadBufferSize     = 1 * 1024
	defaultMaxRequestBodySize = 1 * 1024 * 1024
	defaultServerTimeout      = 900 * time.Second
	invokePath                = "/invoke"
)

// SetShutDownStatus -
func SetShutDownStatus() {
	isShutDown.Store(true)
}

// StartHTTPServer -
func StartHTTPServer(errChan chan<- error) (*fasthttp.Server, error) {
	fastServer := &fasthttp.Server{
		Handler:            route,
		TLSConfig:          getTLSConfig(),
		ReadBufferSize:     defaultReadBufferSize,
		ReadTimeout:        defaultServerTimeout,
		WriteTimeout:       defaultServerTimeout,
		MaxRequestBodySize: defaultMaxRequestBodySize,
	}
	if config.GlobalConfig.HTTPSConfig != nil && config.GlobalConfig.HTTPSConfig.HTTPSEnable {
		if err := commonTls.InitTLSConfig(*config.GlobalConfig.HTTPSConfig); err != nil {
			return nil, fmt.Errorf("init HTTPS config error: %s", err.Error())
		}
	}
	go func() {
		err := startServer(fastServer)
		if err != nil {
			log.GetLogger().Errorf("failed to start http server, err %s", err.Error())
		}
		errChan <- err
	}()
	return fastServer, nil
}

func getTLSConfig() *tls.Config {
	if config.GlobalConfig.HTTPSConfig == nil || !config.GlobalConfig.HTTPSConfig.HTTPSEnable {
		return nil
	}
	tlsConfig := commonTls.GetClientTLSConfig()
	if tlsConfig != nil {
		tlsConfig.NextProtos = []string{"http/1.1"}
	}
	return tlsConfig
}

func startServer(httpServer *fasthttp.Server) error {
	podIP := os.Getenv("POD_IP")
	if net.ParseIP(podIP) == nil {
		log.GetLogger().Errorf("failed to get pod ip, pod ip is %s", podIP)
		return errors.New("failed to get pod ip")
	}
	serverAddr := fmt.Sprintf("%s:%s", podIP, selfregister.GetFaaSSchedulerHttpPort())
	if config.GlobalConfig.HTTPSConfig != nil && config.GlobalConfig.HTTPSConfig.HTTPSEnable {
		log.GetLogger().Infof("start to listen the https request on addr: %s", serverAddr)
		if err := fastHTTPListenAndServeTLS(serverAddr, httpServer); err != nil {
			log.GetLogger().Errorf("failed to start the HTTPS server: %s", err.Error())
			return err
		}
		return nil
	}
	log.GetLogger().Infof("start to listen the http request on addr: %s", serverAddr)
	err := httpServer.ListenAndServe(serverAddr)
	if err != nil {
		log.GetLogger().Errorf("failed to start the HTTP server: %s", err.Error())
		return err
	}
	return nil
}

func fastHTTPListenAndServeTLS(addr string, server *fasthttp.Server) error {
	listener, err := net.Listen("tcp4", addr)
	if err != nil {
		return err
	}
	if server == nil || server.TLSConfig == nil {
		return errors.New("server or tls config is nil")
	}
	tlsListener := tls.NewListener(listener, server.TLSConfig)
	if err = server.Serve(tlsListener); err != nil {
		return err
	}
	return nil
}

func route(ctx *fasthttp.RequestCtx) {
	err := auth(ctx)
	if err != nil {
		ctx.SetStatusCode(http.StatusUnauthorized)
		log.GetLogger().Errorf("failed to check auth, error: %s", err.Error())
		return
	}
	path := string(ctx.Path())
	switch path {
	case invokePath:
		invokeHandler(ctx)
	default:
		ctx.SetStatusCode(http.StatusInternalServerError)
		log.GetLogger().Errorf("unsupported http request path %s", path)
	}
	return
}

func auth(ctx *fasthttp.RequestCtx) error {
	if !config.GlobalConfig.AuthenticationEnable {
		return nil
	}
	sign := string(ctx.Request.Header.Peek(constant.HeaderAuthorization))
	if strings.HasPrefix(sign, localauth.AuthPrefixHmacSha256) {
		return localauth.VerifySignWithHmacSha256(ctx, config.GlobalConfig.SystemAuthConfig.AccessKey,
			config.GlobalConfig.SystemAuthConfig.SecretKey)
	}
	timestamp := string(ctx.Request.Header.Peek(constant.HeaderAuthTimestamp))
	return localauth.AuthCheckLocally(config.GlobalConfig.LocalAuth.AKey, config.GlobalConfig.LocalAuth.SKey, sign,
		timestamp, config.GlobalConfig.LocalAuth.Duration)
}

func invokeHandler(ctx *fasthttp.RequestCtx) {
	traceID := string(ctx.Request.Header.Peek(constant.HeaderTraceID))
	logger := log.GetLogger().With(zap.Any("traceId", traceID))
	if isShutDown.Load() {
		ctx.SetStatusCode(http.StatusOK)
		ctx.Response.Header.Set(constant.HeaderInnerCode, strconv.Itoa(statuscode.ErrFinalized))
		logger.Errorf("scheduler is in shutdown pharse")
		return
	}
	reqBody := ctx.Request.Body()
	var args []api.Arg
	err := json.Unmarshal(reqBody, &args)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		logger.Errorf("unmarshl request body error, err %s", err.Error())
		return
	}
	if functionscaler.GetGlobalScheduler() == nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		logger.Errorf("scheduler is nil")
		return
	}
	respBody, err := functionscaler.GetGlobalScheduler().ProcessInstanceRequestLibruntime(args, traceID)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		logger.Errorf("marshl response body, err %s", err.Error())
		return
	}
	ctx.SetStatusCode(http.StatusOK)
	ctx.Response.SetBody(respBody)
}
