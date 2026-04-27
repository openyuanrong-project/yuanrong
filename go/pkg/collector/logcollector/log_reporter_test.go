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
	"testing"

	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/collector/common"
	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/logservice"
)

func TestGetRuntimeID(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"runtime-12345.out", "runtime-12345"},
		{"runtime-123e4567-e89b-12d3-a456-426614174000.err", "runtime-123e4567-e89b-12d3-a456-426614174000"},
		{"no-runtime.log", ""},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := getRuntimeID(tt.filename)
			if got != tt.want {
				t.Errorf("getRuntimeID(%v) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestParseLogFileName(t *testing.T) {
	tests := []struct {
		filePath string
		want     *logservice.LogItem
		wantOk   bool
	}{
		{
			filePath: "path/to/runtime-abc123-456.out",
			want: &logservice.LogItem{
				Filename:    "path/to/runtime-abc123-456.out",
				CollectorID: common.CollectorConfigs.CollectorID,
				Target:      logservice.LogTarget_USER_STD,
				RuntimeID:   "runtime-abc123-456",
			},
			wantOk: true,
		},
		{
			filePath: "/path/to/runtime-1c1.err",
			want: &logservice.LogItem{
				Filename:    "/path/to/runtime-1c1.err",
				CollectorID: common.CollectorConfigs.CollectorID,
				Target:      logservice.LogTarget_USER_STD,
				RuntimeID:   "runtime-1c1",
			},
			wantOk: true,
		},
		{
			filePath: "path/to/unknown.log",
			want:     nil,
			wantOk:   false,
		},
		{
			filePath: "/path/to/function-master.log",
			want:     nil,
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			got, ok := parseLogFileName(tt.filePath)
			assert.Equal(t, tt.wantOk, ok)
			if ok {
				assert.Equal(t, tt.want.Filename, got.Filename)
				assert.Equal(t, tt.want.CollectorID, got.CollectorID)
				assert.Equal(t, tt.want.Target, got.Target)
				assert.Equal(t, tt.want.RuntimeID, got.RuntimeID)
			}
		})
	}
}
