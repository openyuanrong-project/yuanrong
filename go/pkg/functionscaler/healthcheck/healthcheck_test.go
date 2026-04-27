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

package healthcheck

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/common/faas_common/tls"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

func TestStartHealthCheck(t *testing.T) {
	config.GlobalConfig.EnableHealthCheck = true
	convey.Convey("TestStartHealthCheck", t, func() {
		os.Setenv("POD_IP", "127.0.0.1")
		config.InitModuleConfig()
		config.GlobalConfig.HTTPSConfig = &tls.InternalHTTPSConfig{
			HTTPSEnable: false,
		}
		errChan := make(chan error, 2)
		err := StartHealthCheck(errChan)
		convey.So(err, convey.ShouldBeNil)
		time.Sleep(1 * time.Second)
	})
}

func TestStartServer_WithHTTPS(t *testing.T) {
	config.GlobalConfig.EnableHealthCheck = true
	os.Setenv("POD_IP", "127.0.0.1")
	config.InitModuleConfig()
	config.GlobalConfig.HTTPSConfig = &tls.InternalHTTPSConfig{
		HTTPSEnable: true,
	}
	httpServer := &http.Server{
		Addr: ":443",
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(tls.InitTLSConfig, func(httpsConfig tls.InternalHTTPSConfig) error {
		return nil
	})

	mockedTLSConfig := &tls.InternalHTTPSConfig{}
	patches.ApplyFunc(tls.GetClientTLSConfig, func() *tls.InternalHTTPSConfig {
		return mockedTLSConfig
	})

	mockedError := errors.New("mocked ListenAndServeTLS error")
	patches.ApplyMethod(reflect.TypeOf(httpServer), "ListenAndServeTLS",
		func(_ *http.Server, certFile, keyFile string) error {
			return mockedError
		})

	err := startServer(httpServer)

	assert.Equal(t, mockedError, err)
}

func TestCheck(t *testing.T) {
	req, err := http.NewRequest("GET", "/check", nil)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	convey.Convey("Test check", t, func() {
		convey.Convey("normal case", func() {
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(check)
			handler.ServeHTTP(rr, req)
			expectedBody := ""
			convey.So(rr.Code, convey.ShouldEqual, http.StatusOK)
			convey.So(rr.Body.String(), convey.ShouldEqual, expectedBody)
		})
		convey.Convey("rollout case", func() {
			discoveryConfig := config.GlobalConfig.SchedulerDiscovery
			rolloutConfig := config.GlobalConfig.EnableRollout
			isRolloutObject := selfregister.IsRolloutObject
			defer func() {
				config.GlobalConfig.SchedulerDiscovery = discoveryConfig
				config.GlobalConfig.EnableRollout = rolloutConfig
				selfregister.IsRolloutObject = isRolloutObject
			}()
			config.GlobalConfig.SchedulerDiscovery = &types.SchedulerDiscovery{
				RegisterMode: types.RegisterTypeContend,
			}
			handler := http.HandlerFunc(check)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			convey.So(rr.Code, convey.ShouldEqual, http.StatusInternalServerError)
			config.GlobalConfig.EnableRollout = true
			selfregister.IsRolloutObject = true
			rr = httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			convey.So(rr.Code, convey.ShouldEqual, http.StatusOK)
		})
	})

}

func TestStartHealthCheckFail(t *testing.T) {
	convey.Convey("TestStartHealthCheckFail", t, func() {
		config.InitModuleConfig()
		config.GlobalConfig.HTTPSConfig = &tls.InternalHTTPSConfig{
			HTTPSEnable: false,
		}

		err := StartHealthCheck(nil)
		convey.So(err, convey.ShouldNotBeNil)

		os.Clearenv()
		errChan := make(chan error, 1)
		err = StartHealthCheck(errChan)
		convey.So(err, convey.ShouldNotBeNil)

		os.Setenv("POD_IP", "123")
		err = StartHealthCheck(errChan)
		convey.So(err, convey.ShouldNotBeNil)

		os.Setenv("POD_IP", "1.1.1.1")
		err = StartHealthCheck(errChan)
		convey.So(err, convey.ShouldBeNil)
		err = <-errChan
		convey.So(err, convey.ShouldNotBeNil)

		config.GlobalConfig.HTTPSConfig.HTTPSEnable = true
		err = StartHealthCheck(errChan)
		convey.So(err, convey.ShouldBeNil)
		err = <-errChan
		convey.So(err, convey.ShouldNotBeNil)
	})
}
