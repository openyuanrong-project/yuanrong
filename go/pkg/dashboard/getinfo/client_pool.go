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

// Package getinfo for get function-master info
package getinfo

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/api"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/dashboard/flags"
)

// HttpClient is client pool
var HttpClient *http.Client

// PromClient is prometheus client pool
var PromClient prometheusv1.API

const (
	reqType             = "protobuf"
	connTimeout         = 30 * time.Minute // 连接超时时间
	maxIdleConns        = 10               // 最大空闲连接数
	maxIdleConnsPerHost = 5                // 每个主机最大空闲连接数
	maxConnsPerHost     = 10               // 每个主机最大连接数
	idleConnTimeout     = 30 * time.Second // 空闲连接的超时时间
	tlsHandshakeTimeout = 30 * time.Minute // 限制TLS握手的时间
)

func init() {
	initFunctionSystemClient()
	InitPromClient()
}

func newClientConfig(tlsConf *tls.Config) *http.Client {
	return &http.Client{
		Timeout: connTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        maxIdleConns,
			MaxIdleConnsPerHost: maxIdleConnsPerHost,
			MaxConnsPerHost:     maxConnsPerHost,
			IdleConnTimeout:     idleConnTimeout,
			TLSHandshakeTimeout: tlsHandshakeTimeout,
			TLSClientConfig:     tlsConf,
		}}
}

// NewTLSConfig -
func NewTLSConfig(certConfig flags.CertConfig) *tls.Config {
	if !certConfig.SslEnable {
		return nil
	}
	clientCert, err := tls.LoadX509KeyPair(certConfig.CertFile, certConfig.KeyFile)
	if err != nil {
		log.GetLogger().Errorf("failed to load client certificate, error: %s", err.Error())
		return nil
	}

	certPool := x509.NewCertPool()
	caCert, err := os.ReadFile(certConfig.CaFile)
	if err != nil {
		log.GetLogger().Errorf("failed to read ca certificate: %s", err.Error())
		return nil
	}
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		log.GetLogger().Errorf("failed to append ca certificate")
		return nil
	}
	return &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}
}

func initFunctionSystemClient() {
	HttpClient = newClientConfig(NewTLSConfig(flags.DashboardConfig.FunctionSystemConfig))
}

// InitPromClient -
func InitPromClient() {
	httpClient := newClientConfig(NewTLSConfig(flags.DashboardConfig.PrometheusConfig))
	apiClient, promClientErr := api.NewClient(api.Config{
		Address: flags.DashboardConfig.PrometheusAddr,
		Client:  httpClient,
	})
	if promClientErr != nil {
		log.GetLogger().Errorf("failed to connect prometheus, error: %s", promClientErr.Error())
	} else {
		PromClient = prometheusv1.NewAPI(apiClient)
	}
}

func requestFunctionMaster(path string) ([]byte, error) {
	req, err := http.NewRequest("GET", flags.DashboardConfig.FunctionMasterAddr+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Type", reqType)
	return handleRes(req)
}

func requestFrontend(method string, path string, reqBody io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, flags.DashboardConfig.FrontendAddr+path, reqBody)
	if err != nil {
		return nil, err
	}
	return handleRes(req)
}

func handleRes(req *http.Request) ([]byte, error) {
	resp, err := HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
