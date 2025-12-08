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

	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/logservice"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

// Server 实现了 LogManagerService 服务
type Server struct {
	logservice.UnimplementedLogManagerServiceServer
}

// Register 处理 Register RPC 请求
func (s *Server) Register(_ context.Context, req *logservice.RegisterRequest) (*logservice.RegisterResponse, error) {
	log.GetLogger().Infof("receive register request: %v", req)
	err := managerSingleton.RegisterLogCollector(collectorClientInfo{
		ID:      req.CollectorID,
		Address: req.Address,
	})
	if err != nil {
		return nil, err
	}
	return &logservice.RegisterResponse{
		Code: 0,
	}, nil
}

// ReportLog 处理 ReportLog RPC 请求
func (s *Server) ReportLog(_ context.Context, req *logservice.ReportLogRequest) (*logservice.ReportLogResponse, error) {
	log.GetLogger().Infof("receive report request: %v", req)
	for _, item := range req.GetItems() {
		managerSingleton.ReportLogItem(item)
	}
	return &logservice.ReportLogResponse{
		Code:    0,
		Message: "Log reported successfully",
	}, nil
}

// RemoveLog 处理 RemoveLog RPC 请求
func (s *Server) RemoveLog(_ context.Context, req *logservice.ReportLogRequest) (*logservice.ReportLogResponse, error) {
	log.GetLogger().Infof("receive remove log request: %v", req)
	for _, item := range req.GetItems() {
		managerSingleton.RemoveLogItem(item)
	}
	return &logservice.ReportLogResponse{
		Code:    0,
		Message: "Log removed successfully",
	}, nil
}
