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

package logcollector

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"yuanrong.org/kernel/pkg/collector/common"
	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/logservice"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

const (
	mb uint32 = 1024 * 1024
)

var (
	readLogChunkThreshold uint32 = 1024 * 1024
	redundantBytes        uint32 = 1024
)

type server struct {
	logservice.UnimplementedLogCollectorServiceServer
}

// ReadLog deal with grpc read log request
func (s *server) ReadLog(request *logservice.ReadLogRequest,
	stream logservice.LogCollectorService_ReadLogServer) error {
	filename, err := GetAbsoluteFilePath(request.Item)
	if err != nil {
		stream.Send(&logservice.ReadLogResponse{
			Code:    -1,
			Message: err.Error(),
		})
		return err
	}
	if _, err := os.Stat(filename); err != nil {
		stream.Send(&logservice.ReadLogResponse{
			Code:    -1,
			Message: err.Error(),
		})
		return err
	}
	return s.sendStreamResponse(request, stream, filename)
}

func appendToStringBuilder(builder *strings.Builder, line string) error {
	if _, err := builder.WriteString(line); err != nil {
		return err
	}
	if _, err := builder.WriteString("\n"); err != nil {
		return err
	}
	return nil
}

func (s *server) sendStreamResponse(request *logservice.ReadLogRequest,
	stream logservice.LogCollectorService_ReadLogServer, absoluteFilename string) error {
	file, err := os.OpenFile(absoluteFilename, os.O_RDONLY, 0)
	if err != nil {
		log.GetLogger().Errorf("failed to open file: %s, error: %v", absoluteFilename, err)
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	builder := &strings.Builder{}
	builder.Grow(int(readLogChunkThreshold + redundantBytes))
	var totalBytes uint32 = 0
	var lineCount uint32 = 0
	for scanner.Scan() {
		if lineCount < request.StartLine {
			lineCount++
			continue
		}
		if lineCount >= request.EndLine {
			break
		}
		lineCount++
		line := scanner.Text()
		if err := appendToStringBuilder(builder, line); err != nil {
			log.GetLogger().Errorf("failed to write line to buffer, error: %v", err)
			return err
		}
		totalBytes += uint32(len(line) + 1)
		if totalBytes >= readLogChunkThreshold {
			log.GetLogger().Infof("total bytes %d reaches threshold %d", totalBytes, readLogChunkThreshold)
			if err := s.send(stream, request.Item.Filename, totalBytes, builder); err != nil {
				return err
			}
			builder.Reset()
			builder.Grow(int(readLogChunkThreshold + redundantBytes))
			totalBytes = 0
		}
	}
	if err := scanner.Err(); err != nil {
		stream.Send(&logservice.ReadLogResponse{Code: -1, Message: err.Error()})
		return err
	}
	if builder.Len() > 0 {
		if err := s.send(stream, request.Item.Filename, totalBytes, builder); err != nil {
			return err
		}
	}
	return nil
}

func (s *server) send(stream logservice.LogCollectorService_ReadLogServer, filename string, totalBytes uint32,
	builder *strings.Builder) error {
	log.GetLogger().Infof("send read log response for %s, size: %f MB", filename, float64(totalBytes)/float64(mb))
	err := stream.Send(&logservice.ReadLogResponse{
		Code:    0,
		Content: []byte(builder.String()),
	})
	if err != nil {
		log.GetLogger().Errorf("failed to send read log response for %s, size: %f MB", filename,
			float64(totalBytes)/float64(mb))
		return err
	}
	return nil
}

// QueryLogStream -
func (s *server) QueryLogStream(ctx context.Context, request *logservice.QueryLogStreamRequest) (
	*logservice.QueryLogStreamResponse, error) {
	streamControlChans.Lock()
	defer streamControlChans.Unlock()
	streams := make([]string, 0, len(streamControlChans.hashmap))
	for streamName := range streamControlChans.hashmap {
		streams = append(streams, streamName)
	}
	return &logservice.QueryLogStreamResponse{Code: 0, Streams: streams}, nil
}

// StartReadLogService starts grpc server and then set ready channel
func StartReadLogService(ready chan<- bool) error {
	var creds credentials.TransportCredentials
	functionSystemConf := common.CollectorConfigs.FunctionSystemConfig
	if functionSystemConf.SslEnable {
		serverCert, err := tls.LoadX509KeyPair(functionSystemConf.CertFile, functionSystemConf.KeyFile)
		if err != nil {
			log.GetLogger().Errorf("failed to load collector server certificate: %s", err.Error())
			return err
		}

		certPool := x509.NewCertPool()
		caCert, err := os.ReadFile(functionSystemConf.CaFile)
		if err != nil {
			log.GetLogger().Errorf("failed to load collector ca certificate: %s", err.Error())
			return err
		}
		if ok := certPool.AppendCertsFromPEM(caCert); !ok {
			log.GetLogger().Errorf("failed to append collector ca certificate")
			return errors.New("failed to append collector ca certificate")
		}

		creds = credentials.NewTLS(&tls.Config{
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    certPool,
			Certificates: []tls.Certificate{serverCert},
		})
	}
	grpcServer := grpc.NewServer(grpc.Creds(creds))
	logservice.RegisterLogCollectorServiceServer(grpcServer, &server{})

	lis, err := net.Listen("tcp", common.CollectorConfigs.Address)
	if err != nil {
		log.GetLogger().Errorf("failed to listen to address %s, error: %v", common.CollectorConfigs.Address, err)
		return err
	}

	ready <- true
	log.GetLogger().Infof("start serve log service on address %s", common.CollectorConfigs.Address)
	if err = grpcServer.Serve(lis); err != nil {
		log.GetLogger().Errorf("failed to serve on address %s, error: %v", common.CollectorConfigs.Address, err)
		return err
	}
	log.GetLogger().Infof("stop serve log service on address %s", common.CollectorConfigs.Address)
	return nil
}
