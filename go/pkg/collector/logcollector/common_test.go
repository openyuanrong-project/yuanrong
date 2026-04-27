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
	"net"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"yuanrong.org/kernel/pkg/collector/common"
	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/logservice"
)

type constTestImpl struct{}

func (c *constTestImpl) GetRetryInterval() time.Duration {
	return 100 * time.Microsecond
}

func (c *constTestImpl) GetMaxRetryTimes() int {
	return 3
}

func (c *constTestImpl) GetMaxTimeForUnfinishedLine() time.Duration {
	return 1 * time.Millisecond
}

func TestGetAbsoluteFilePath(t *testing.T) {
	{
		item := &logservice.LogItem{
			Target:   logservice.LogTarget_USER_STD,
			Filename: "example.log",
		}
		expectedPath := filepath.Join(common.CollectorConfigs.UserLogPath, item.Filename)
		path, err := GetAbsoluteFilePath(item)
		assert.NoError(t, err)
		assert.Equal(t, expectedPath, path)
	}

	{
		item := &logservice.LogItem{
			Target:   logservice.LogTarget_LIB_RUNTIME,
			Filename: "example.log",
		}
		path, err := GetAbsoluteFilePath(item)
		assert.Error(t, err)
		assert.Empty(t, path)
	}
}

func TestGetLogServiceClient(t *testing.T) {
	logServiceClientOnce = sync.Once{}
	LogServiceClient = nil

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer lis.Close()

	go func() {
		grpc.NewServer().Serve(lis)
	}()

	common.CollectorConfigs.ManagerAddress = lis.Addr().String()

	client := GetLogServiceClient()

	if client == nil {
		t.Errorf("Expected a valid log manager grpc client, but got nil")
	}
}
