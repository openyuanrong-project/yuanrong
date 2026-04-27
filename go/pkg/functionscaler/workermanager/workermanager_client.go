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

// Package workermanager worker manager client
package workermanager

import (
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/localauth"
	"yuanrong.org/kernel/pkg/common/faas_common/tls"
	"yuanrong.org/kernel/pkg/functionscaler/config"
)

const (
	workerManagerDeployURL                  = "/worker-manager/v1/functions/sn:cn:yrk:%s:function:%s:%s/worker"
	workerManagerDeleteURL                  = "/worker-manager/v1/functions/worker/delete"
	statusOKCode                            = 150200
	scaleDownTimeout                        = 30 * time.Second
	defaultRequestWorkerManagerImageTimeout = 25 * time.Minute
	appID                                   = "ondemand"
	dialTimeout                             = 3 * time.Second
	idleConnTimeout                         = 90 * time.Second
	connKeepAlive                           = 30 * time.Second
	maxIdleConns                            = 100
	httpScheme                              = "http://"
	httpsScheme                             = "https://"
)

var (
	once       sync.Once
	httpClient *http.Client
)

// GetWorkerManagerClient -
func GetWorkerManagerClient() *http.Client {
	once.Do(func() {
		tr := &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   dialTimeout,
				KeepAlive: connKeepAlive,
			}).DialContext,
			MaxIdleConns:        maxIdleConns,
			ForceAttemptHTTP2:   true,
			IdleConnTimeout:     idleConnTimeout,
			TLSHandshakeTimeout: dialTimeout,
		}
		if config.GlobalConfig.HTTPSConfig.HTTPSEnable {
			tr.TLSClientConfig = tls.GetClientTLSConfig()
		}

		httpClient = &http.Client{
			Timeout:   defaultRequestWorkerManagerImageTimeout,
			Transport: tr,
		}
	})
	return httpClient
}

// GetWorkerManagerBaseURL -
func GetWorkerManagerBaseURL() (string, error) {
	scheme := httpScheme
	if config.GlobalConfig.HTTPSConfig != nil && config.GlobalConfig.HTTPSConfig.HTTPSEnable {
		scheme = httpsScheme
	}
	addr := getWmAddr()
	if addr == "" {
		return "", errors.New("worker manager address is empty")
	}
	return scheme + addr, nil
}

// FillInWorkerManagerRequestHeaders contains authorization, source and so on
func FillInWorkerManagerRequestHeaders(request *http.Request) {
	authorization, timestamp := generateAuthorization()
	request.Header.Set(constant.HeaderAuthTimestamp, timestamp)
	request.Header.Set(constant.HeaderAuthorization, authorization)
	request.Header.Set(constant.HeaderEventSourceID, appID)
	request.Header.Set(constant.HeaderCallType, "active")
}

func generateAuthorization() (string, string) {
	var authorization, timestamp string
	authorization, timestamp = localauth.SignLocally(config.GlobalConfig.LocalAuth.AKey,
		config.GlobalConfig.LocalAuth.SKey, appID, config.GlobalConfig.LocalAuth.Duration)
	return authorization, timestamp
}
