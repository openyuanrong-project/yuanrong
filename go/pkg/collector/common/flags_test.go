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

package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandLineArguments(t *testing.T) {
	testCases := []struct {
		args                   []string
		expectedID             string
		expectedIP             string
		expectedPort           string
		expectedManagerAddress string
		expectedDatasystemPort int
		expectedLogRoot        string
		expectedUserLogPath    string
	}{
		{
			args:                   []string{"--collect_id", "12345", "--ip", "192.168.1.1", "--port", "8080", "--manager_address", "10.10.10.10:5678", "--log_root", "/var/log", "--user_log_path", "/tmp"},
			expectedID:             "12345",
			expectedIP:             "192.168.1.1",
			expectedPort:           "8080",
			expectedManagerAddress: "10.10.10.10:5678",
			expectedDatasystemPort: 9090,
			expectedLogRoot:        "/var/log",
			expectedUserLogPath:    "/tmp",
		},
		{
			args:                   []string{"--collect_id", "abcde12345", "--ip", "192.168.1.1", "--port", "8080", "--manager_address", "10.10.10.10:5678", "--log_root", "/var/log"},
			expectedID:             "abcde12345",
			expectedIP:             "192.168.1.1",
			expectedPort:           "8080",
			expectedManagerAddress: "10.10.10.10:5678",
			expectedDatasystemPort: 9090,
			expectedLogRoot:        "/var/log",
			expectedUserLogPath:    "/var/log",
		},
	}

	for _, tc := range testCases {
		CollectorConfigs.UserLogPath = ""
		rootCmd.SetArgs(tc.args)

		if err := rootCmd.Execute(); err != nil {
			t.Errorf("Failed to execute command: %v", err)
		}
		assert.Equal(t, validateCmdArgs(), nil, "failed to validate cmd args")

		assert.Equal(t, tc.expectedID, CollectorConfigs.CollectorID, "Collector ID mismatch")
		assert.Equal(t, tc.expectedIP, CollectorConfigs.IP, "IP mismatch")
		assert.Equal(t, tc.expectedPort, CollectorConfigs.Port, "Port mismatch")
		assert.Equal(t, tc.expectedManagerAddress, CollectorConfigs.ManagerAddress, "Manager Address mismatch")
		assert.Equal(t, tc.expectedLogRoot, CollectorConfigs.LogRoot, "Log Root mismatch")
		assert.Equal(t, tc.expectedUserLogPath, CollectorConfigs.UserLogPath, "User Log Path mismatch")
	}
}

func TestInvalidCommandLineArguments(t *testing.T) {
	testCases := []struct {
		args     []string
		errorMsg string
	}{
		{
			args:     []string{"--collect_id", ""},
			errorMsg: "please check characters and length. The valid length range is",
		},
		{
			args:     []string{"--collect_id", "[]"},
			errorMsg: "please check characters and length. The valid length range is",
		},
		{
			args:     []string{"--collect_id", "01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789"},
			errorMsg: "please check characters and length. The valid length range is",
		},
		{
			args:     []string{"--collect_id", "1", "--ip", "255.255.255.256"},
			errorMsg: "collector IP 255.255.255.256 is invalid",
		},
		{
			args:     []string{"--collect_id", "1", "--ip", "10.10.10.10", "--port", "65537"},
			errorMsg: "collector port 65537 is invalid",
		},
		{
			args:     []string{"--collect_id", "1", "--ip", "10.10.10.10", "--port", "8080", "--manager_address", "abcde"},
			errorMsg: "manager address abcde is invalid",
		},
		{
			args:     []string{"--collect_id", "1", "--ip", "10.10.10.10", "--port", "8080", "--manager_address", "abc:def"},
			errorMsg: "manager address abc:def has invalid IP",
		},
		{
			args:     []string{"--collect_id", "1", "--ip", "10.10.10.10", "--port", "8080", "--manager_address", "10.10.10.10:-3"},
			errorMsg: "manager address 10.10.10.10:-3 has invalid port",
		},
		{
			args:     []string{"--collect_id", "1", "--ip", "10.10.10.10", "--port", "8080", "--manager_address", "10.10.10.10:90", "--log_root", "./here"},
			errorMsg: "./here should be absolute path",
		},
		{
			args:     []string{"--collect_id", "1", "--ip", "10.10.10.10", "--port", "8080", "--manager_address", "10.10.10.10:90", "--log_root", "/var/log", "--user_log_path", "./here"},
			errorMsg: "./here should be absolute path",
		},
		{
			args:     []string{"--collect_id", "1", "--ip", "10.10.10.10", "--port", "8080", "--manager_address", "10.10.10.10:90", "--log_root", "/var/log", "--user_log_path", "/here"},
			errorMsg: "no such directory: /here",
		},
	}

	for _, tc := range testCases {
		rootCmd.SetArgs(tc.args)

		if err := rootCmd.Execute(); err != nil {
			t.Errorf("Failed to execute command: %v", err)
		}

		assert.Contains(t, validateCmdArgs().Error(), tc.errorMsg, "unexpected error message")
	}
}
