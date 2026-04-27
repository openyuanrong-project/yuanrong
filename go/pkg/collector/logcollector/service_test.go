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
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"yuanrong.org/kernel/pkg/collector/common"
	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/logservice"
)

type MockLogCollectorService_ReadLogServer struct {
	grpc.ServerStream
	SentResponses []*logservice.ReadLogResponse
}

func (m *MockLogCollectorService_ReadLogServer) Send(response *logservice.ReadLogResponse) error {
	m.SentResponses = append(m.SentResponses, response)
	return nil
}

func TestFailedReadLog(t *testing.T) {
	s := &server{}
	stream := &MockLogCollectorService_ReadLogServer{}

	testCases := []struct {
		request      *logservice.ReadLogRequest
		expectedCode int32
		errorMsg     string
	}{
		{
			request:      &logservice.ReadLogRequest{Item: &logservice.LogItem{Filename: "", Target: logservice.LogTarget_LIB_RUNTIME}},
			expectedCode: -1,
			errorMsg:     fmt.Sprintf("undefined log item target"),
		},
		{
			request:      &logservice.ReadLogRequest{Item: &logservice.LogItem{Filename: "123.txt", Target: logservice.LogTarget_USER_STD}},
			expectedCode: -1,
			errorMsg:     fmt.Sprintf("no such file or directory"),
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			stream.SentResponses = nil
			err := s.ReadLog(tc.request, stream)
			assert.Contains(t, err.Error(), tc.errorMsg, "unexpected error message")

			if len(stream.SentResponses) > 0 {
				actualResponse := stream.SentResponses[0]
				assert.Equal(t, actualResponse.Code, tc.expectedCode, "unexpected error code")
			}
		})
	}
}

func TestSuccessReadLog(t *testing.T) {
	s := &server{}
	stream := &MockLogCollectorService_ReadLogServer{}
	stream.SentResponses = nil
	tempDir := t.TempDir()
	file := filepath.Join(tempDir, "runtime-111.err")
	f, _ := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	for i := 0; i < 50; i++ {
		content := fmt.Sprintf("timestamp: %d\n", i)
		f.WriteString(content)
	}

	request := &logservice.ReadLogRequest{
		Item:      &logservice.LogItem{Filename: file, Target: logservice.LogTarget_USER_STD},
		StartLine: 1,
		EndLine:   5,
	}
	s.ReadLog(request, stream)
	assert.Equal(t, len(stream.SentResponses), 1, "wrong count of responses")
	response := stream.SentResponses[0]
	assert.Equal(t, response.Code, int32(0), "wrong code")
	assert.Equal(t, response.Content, []byte("timestamp: 1\ntimestamp: 2\ntimestamp: 3\ntimestamp: 4\n"), "wrong content")
}

func TestMultipleSuccessReadLog(t *testing.T) {
	readLogChunkThreshold = 30
	redundantBytes = 20
	s := &server{}
	stream := &MockLogCollectorService_ReadLogServer{}
	stream.SentResponses = nil
	tempDir := t.TempDir()
	file := filepath.Join(tempDir, "runtime-112.err")
	f, _ := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	for i := 0; i < 50; i++ {
		content := fmt.Sprintf("0123456789\n")
		f.WriteString(content)
	}

	request := &logservice.ReadLogRequest{
		Item:      &logservice.LogItem{Filename: file, Target: logservice.LogTarget_USER_STD},
		StartLine: 10,
		EndLine:   40,
	}
	s.ReadLog(request, stream)
	assert.Equal(t, 10, len(stream.SentResponses), "wrong count of responses")
	for _, response := range stream.SentResponses {
		assert.Equal(t, int32(0), response.Code, "wrong code")
		assert.Equal(t, []byte("0123456789\n0123456789\n0123456789\n"), response.Content, "wrong content")
	}
}

func TestFailedStartReadLogService(t *testing.T) {
	common.CollectorConfigs.Address = "xxx"
	ready := make(chan bool)
	go func() {
		<-ready
	}()
	err := StartReadLogService(ready)
	assert.NotEqual(t, nil, err)
}
