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

// Package logcollector collects and publish logs
package logcollector

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"yuanrong.org/kernel/pkg/collector/common"
	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/logservice"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

const (
	realRetryInterval = 100 * time.Millisecond
	realMaxRetryTimes = 30

	maxTimeForUnfinishedLine time.Duration = 1 * time.Second
)

type constantInterface interface {
	GetRetryInterval() time.Duration
	GetMaxRetryTimes() int
	GetMaxTimeForUnfinishedLine() time.Duration
}

type constImpl struct{}

func (c *constImpl) GetRetryInterval() time.Duration {
	return realRetryInterval
}

func (c *constImpl) GetMaxRetryTimes() int {
	return realMaxRetryTimes
}

func (c *constImpl) GetMaxTimeForUnfinishedLine() time.Duration {
	return maxTimeForUnfinishedLine
}

var constant constantInterface = &constImpl{}

var (
	// LogServiceClient is grpc log service interface
	LogServiceClient     logservice.LogManagerServiceClient
	logServiceClientOnce sync.Once
)

// streamControlChans maps streamName to stream done channel for each file
var streamControlChans = struct {
	sync.Mutex
	hashmap map[string]map[string]chan bool
}{
	hashmap: make(map[string]map[string]chan bool),
}

// GetAbsoluteFilePath returns absolute path based on the target type
func GetAbsoluteFilePath(item *logservice.LogItem) (string, error) {
	switch item.Target {
	case logservice.LogTarget_USER_STD:
		if filepath.IsAbs(item.Filename) {
			return item.Filename, nil
		}
		return filepath.Join(common.CollectorConfigs.UserLogPath, item.Filename), nil
	default:
		break
	}
	log.GetLogger().Warnf("undefined log item target: %v", item.Target)
	return "", fmt.Errorf("undefined log item target: %v", item.Target)
}

// GetLogServiceClient returns log service client
func GetLogServiceClient() logservice.LogManagerServiceClient {
	logServiceClientOnce.Do(func() {
		conn := common.GetConnection()
		if conn == nil {
			log.GetLogger().Errorf("failed to get connection to log manager")
			return
		}
		LogServiceClient = logservice.NewLogManagerServiceClient(conn)
	})

	return LogServiceClient
}
