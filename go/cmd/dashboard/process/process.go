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

// Package process for run dashboard server
package process

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/logservice"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/dashboard/etcdcache"
	"yuanrong.org/kernel/pkg/dashboard/flags"
	"yuanrong.org/kernel/pkg/dashboard/logmanager"
	"yuanrong.org/kernel/pkg/dashboard/routers"
)

// StartGrpcServices of dashboard
func StartGrpcServices() {
	// start grpc service
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", flags.DashboardConfig.Ip, flags.DashboardConfig.GrpcPort))
	if err != nil {
		log.GetLogger().Fatalf("failed to listen: %v", err)
	}
	var creds credentials.TransportCredentials
	functionSystemConf := flags.DashboardConfig.FunctionSystemConfig
	if functionSystemConf.SslEnable {
		serverCert, err := tls.LoadX509KeyPair(functionSystemConf.CertFile, functionSystemConf.KeyFile)
		if err != nil {
			log.GetLogger().Fatalf("failed to load log manager server certificate: %s", err.Error())
		}

		certPool := x509.NewCertPool()
		caCert, err := os.ReadFile(functionSystemConf.CaFile)
		if err != nil {
			log.GetLogger().Fatalf("failed to load log manager ca certificate: %s", err.Error())
		}
		if ok := certPool.AppendCertsFromPEM(caCert); !ok {
			log.GetLogger().Fatalf("failed to append log manager ca certificate")
		}

		creds = credentials.NewTLS(&tls.Config{
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    certPool,
			Certificates: []tls.Certificate{serverCert},
		})
	}
	grpcServer := grpc.NewServer(grpc.Creds(creds))
	logservice.RegisterLogManagerServiceServer(grpcServer, &logmanager.Server{})

	if err := grpcServer.Serve(lis); err != nil {
		log.GetLogger().Fatalf("failed to serve grpc: %s", err.Error())
	}
}

// StartDashboard - function for run dashboard server
func StartDashboard() {
	gin.SetMode(gin.ReleaseMode)
	stopCh := make(chan struct{})

	// init the etcd config first
	err := flags.InitEtcdClient()
	if err != nil {
		log.GetLogger().Fatalf("failed to init etcd, err: %s", err)
	}

	// register self
	err = flags.RegisterSelfToEtcd(stopCh)
	if err != nil {
		log.GetLogger().Fatalf("failed to register self to etcd, err: %s", err)
	}

	// start watcher
	etcdcache.StartWatchInstance(stopCh)

	// start grpc at background
	go StartGrpcServices()

	// start http, and use http as main thread
	r := routers.SetRouter()
	srv := &http.Server{
		Addr:    flags.DashboardConfig.ServerAddr,
		Handler: r,
	}
	sslConfig := flags.DashboardConfig.SslConfig
	log.GetLogger().Debugf("%s is running...",
		flags.AddHTTPPrefix(flags.DashboardConfig.ServerAddr, sslConfig.SslEnable))
	if sslConfig.SslEnable {
		err = srv.ListenAndServeTLS(sslConfig.CertFile, sslConfig.KeyFile)
	} else {
		err = srv.ListenAndServe()
	}
	if err != nil {
		log.GetLogger().Fatalf("srv.ListenAndServe: %s", err.Error())
	}
}
