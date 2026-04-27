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

// Package healthcheck -
package healthcheck

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/functionmanager/config"
)

func TestStartHealthCheck(t *testing.T) {
	t.Run("not enable healthy check", func(t *testing.T) {
		os.Setenv("POD_IP", "127.0.0.1")
		config.InitConfig([]byte("{\"enableHealthCheck\":false}"))
		errChan := make(chan error, 2)
		err := StartHealthCheck(errChan)
		assert.NoError(t, err)
	})
	t.Run("empty PodIP", func(t *testing.T) {
		os.Setenv("POD_IP", "")
		config.InitConfig([]byte("{\"enableHealthCheck\":true}"))
		errChan := make(chan error, 2)
		err := StartHealthCheck(errChan)
		assert.EqualError(t, err, "failed to get pod ip")
	})
	t.Run("empty PodIP", func(t *testing.T) {
		os.Setenv("POD_IP", "")
		config.InitConfig([]byte("{\"enableHealthCheck\":true}"))
		err := StartHealthCheck(nil)
		assert.EqualError(t, err, "errChan is nil")
	})
	t.Run("success to start healthy", func(t *testing.T) {
		os.Setenv("POD_IP", "127.0.0.1")
		config.InitConfig([]byte("{\"enableHealthCheck\":false}"))
		errChan := make(chan error, 2)
		err := StartHealthCheck(errChan)
		assert.NoError(t, err)

		req, err := http.NewRequest("GET", "/healthcheck", nil)
		if err != nil {
			t.Fatalf("err %v", err)
		}
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(check)
		handler.ServeHTTP(rr, req)
		expectedBody := ""
		assert.Equal(t, rr.Code, http.StatusOK)
		assert.Equal(t, rr.Body.String(), expectedBody)
	})
	t.Run("failed ot start healthy", func(t *testing.T) {
		os.Setenv("POD_IP", "127.0.0.1")
		config.InitConfig([]byte("{\"enableHealthCheck\":true}"))
		errChan := make(chan error, 2)
		httpServer := &http.Server{}
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyMethod(reflect.TypeOf(httpServer), "ListenAndServe", func(_ *http.Server) error {
			return errors.New("mocked ListenAndServe error")
		})
		err := StartHealthCheck(errChan)
		assert.NoError(t, err)
		err = <-errChan
		assert.EqualError(t, err, "mocked ListenAndServe error")
	})
}
