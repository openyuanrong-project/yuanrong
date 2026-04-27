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

// Package workermanager worker manager client
package workermanager

import (
	"net/http"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/localauth"
	commontls "yuanrong.org/kernel/pkg/common/faas_common/tls"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

func TestGetWorkerManagerClient(t *testing.T) {
	bakHttpsConfig := config.GlobalConfig.HTTPSConfig
	config.GlobalConfig.HTTPSConfig = &commontls.InternalHTTPSConfig{
		HTTPSEnable: true,
		TLSProtocol: "tls",
		TLSCiphers:  "ciphers",
	}
	client := GetWorkerManagerClient()
	if client == nil {
		t.Errorf("failed to get worker manager client")
	}
	config.GlobalConfig.HTTPSConfig = bakHttpsConfig
}

// TestFillInWorkerManagerRequestHeaders -
func TestFillInWorkerManagerRequestHeaders(t *testing.T) {
	convey.Convey("test: fill in workermanager request header", t, func() {
		patch := gomonkey.ApplyFunc(localauth.SignLocally, func(ak, sk, appID string, duration int) (string, string) {
			return "authorization", "timestamp"
		})
		config.GlobalConfig = types.Configuration{
			LocalAuth: localauth.AuthConfig{},
		}
		defer func() {
			patch.Reset()
			config.GlobalConfig = types.Configuration{}
		}()
		req := http.Request{
			Header: map[string][]string{},
		}
		FillInWorkerManagerRequestHeaders(&req)
		convey.So(req.Header.Get(constant.HeaderAuthTimestamp), convey.ShouldEqual, "timestamp")
		convey.So(req.Header.Get(constant.HeaderAuthorization), convey.ShouldEqual, "authorization")
		convey.So(req.Header.Get(constant.HeaderEventSourceID), convey.ShouldEqual, appID)
		convey.So(req.Header.Get(constant.HeaderCallType), convey.ShouldEqual, "active")
	})
}

// TestGetWorkerManagerBaseURL -
func TestGetWorkerManagerBaseURL(t *testing.T) {
	convey.Convey("test: get worker manager base url", t, func() {
		updateWmAddr("worker-manager:58866")
		tasks := []struct {
			caseName    string
			httpsEnable bool
			patch       func()
			recover     func()
			expect      string
		}{
			{
				caseName:    "http",
				httpsEnable: false,
				patch: func() {
					config.GlobalConfig = types.Configuration{HTTPSConfig: &commontls.InternalHTTPSConfig{HTTPSEnable: false}}
				},
				recover: func() {
					config.GlobalConfig = types.Configuration{}
				},
				expect: "http://worker-manager:58866",
			},
			{
				caseName:    "https",
				httpsEnable: true,
				patch: func() {
					config.GlobalConfig = types.Configuration{HTTPSConfig: &commontls.InternalHTTPSConfig{HTTPSEnable: true}}
				},
				recover: func() {
					config.GlobalConfig = types.Configuration{}
				},
				expect: "https://worker-manager:58866",
			},
		}
		// test
		for _, task := range tasks {
			task.patch()
			actual, _ := GetWorkerManagerBaseURL()
			convey.So(actual, convey.ShouldEqual, task.expect)
			task.recover()
		}
	})
}
