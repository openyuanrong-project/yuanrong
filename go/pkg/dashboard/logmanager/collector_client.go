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

package logmanager

import (
	"context"
	"io"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/logservice"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/dashboard/flags"
	"yuanrong.org/kernel/pkg/dashboard/getinfo"
)

type collectorClientInfo struct {
	ID      string
	Address string
}

type collectorClient struct {
	collectorClientInfo
	grpcConn  *grpc.ClientConn
	logClient logservice.LogCollectorServiceClient
}

// Connect will connect the collector and store the connection
func (c *collectorClient) Connect() error {
	creds := insecure.NewCredentials()
	if flags.DashboardConfig.FunctionSystemConfig.SslEnable {
		tlsConf := getinfo.NewTLSConfig(flags.DashboardConfig.FunctionSystemConfig)
		creds = credentials.NewTLS(tlsConf)
	}
	// connect it
	conn, err := grpc.NewClient(c.Address, grpc.WithTransportCredentials(creds))
	if err != nil {
		return err
	}
	c.grpcConn = conn
	c.logClient = logservice.NewLogCollectorServiceClient(c.grpcConn)
	return nil
}

// Healthcheck should be used in a goroutine
func (c *collectorClient) Healthcheck(shutdownCallback func()) {
	if c.grpcConn == nil {
		return
	}
	for {
		state := c.grpcConn.GetState()
		// TransientFailure means recoverable failure (maybe), but Shutdown means the connection has been closed
		if state == connectivity.Shutdown {
			log.GetLogger().Errorf("connection to collector %s at %s lost or shutting down...", c.ID, c.Address)
			shutdownCallback()
			return
		}
		time.Sleep(time.Second) // Wait before checking again
	}
}

// CollectLog will collect the
func (c *collectorClient) CollectLog(ctx context.Context, readLogReq *logservice.ReadLogRequest,
	outLog chan<- *logservice.ReadLogResponse) error {
	stream, err := c.logClient.ReadLog(ctx, readLogReq)
	if err != nil {
		log.GetLogger().Warnf("failed to read log %s from %s, err: %s", readLogReq.Item.Filename,
			readLogReq.Item.CollectorID, err.Error())
		close(outLog)
		return err
	}
	return redirectLog(readLogReq, stream, outLog)
}

func redirectLog(readLogReq *logservice.ReadLogRequest, stream logservice.LogCollectorService_ReadLogClient,
	outLog chan<- *logservice.ReadLogResponse) error {
	for {
		response, err := stream.Recv()
		if err == nil {
			outLog <- response
			continue
		}
		if err == io.EOF {
			log.GetLogger().Infof("read log stream stopped for %s from %s", readLogReq.Item.Filename,
				readLogReq.Item.CollectorID)
			close(outLog)
			return nil
		}
		log.GetLogger().Warnf("failed to receive log %s from %s, err: %s", readLogReq.Item.Filename,
			readLogReq.Item.CollectorID, err.Error())
		close(outLog)
		return err
	}
}
