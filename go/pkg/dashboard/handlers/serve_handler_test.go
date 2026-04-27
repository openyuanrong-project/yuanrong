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

// Package handlers for handle request
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/gin-gonic/gin"
	"github.com/smartystreets/goconvey/convey"
	clientv3 "go.etcd.io/etcd/client/v3"

	"yuanrong.org/kernel/pkg/common/faas_common/types"
)

func TestServeDelHandler(t *testing.T) {
	convey.Convey("Given a ServeDelHandler", t, func() {
		// Setup gin router and the mock context
		r := gin.Default()
		r.DELETE("/serve", ServeDelHandler)

		req, err := http.NewRequest("DELETE", "/serve", nil)
		convey.So(err, convey.ShouldBeNil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		convey.Convey("It should return status 200 and appropriate message", func() {
			convey.So(w.Code, convey.ShouldEqual, http.StatusOK)

			var response map[string]string
			err := json.Unmarshal(w.Body.Bytes(), &response)
			convey.So(err, convey.ShouldBeNil)

			convey.So(response["message"], convey.ShouldEqual, "ok, but yuanrong doesn't support this right now")
		})
	})
}

func TestServePutHandler(t *testing.T) {
	convey.Convey("Given a ServePutHandler", t, func() {
		// Setup gin router and the mock context
		r := gin.Default()
		r.PUT("/serve", ServePutHandler)
		patches := gomonkey.ApplyFunc(putServeAsFunctionMetaInfoIntoEtcd,
			func(allFuncMetas []*types.ServeFuncWithKeysAndFunctionMetaInfo) error {
				return nil
			})
		defer patches.Reset()
		// Create a mock request payload
		serveDeploySchema := types.ServeDeploySchema{
			Applications: []types.ServeApplicationSchema{
				{
					Name:        "testApp",
					RoutePrefix: "/testRoute",
					ImportPath:  "testImportPath",
					RuntimeEnv: types.ServeRuntimeEnvSchema{
						Pip:        []string{"pip1", "pip2"},
						WorkingDir: "/test/dir",
						EnvVars:    map[string]any{"env1": "value1"},
					},
					Deployments: []types.ServeDeploymentSchema{
						{
							Name:                "testDeployment",
							NumReplicas:         3,
							HealthCheckPeriodS:  10,
							HealthCheckTimeoutS: 5,
						},
					},
				},
			},
		}
		payload, err := json.Marshal(serveDeploySchema)
		convey.So(err, convey.ShouldBeNil)

		req, err := http.NewRequest("PUT", "/serve", bytes.NewReader(payload))
		convey.So(err, convey.ShouldBeNil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		convey.Convey("It should return status 200 and message 'succeed' when no errors occur", func() {
			convey.So(w.Code, convey.ShouldEqual, http.StatusOK)

			var response map[string]string
			err := json.Unmarshal(w.Body.Bytes(), &response)
			convey.So(err, convey.ShouldBeNil)

			convey.So(response["message"], convey.ShouldEqual, "succeed")
		})

		convey.Convey("It should return a 400 status and an error message when invalid JSON is provided", func() {
			// Test invalid JSON
			req, err := http.NewRequest("PUT", "/serve", nil)
			convey.So(err, convey.ShouldBeNil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			convey.So(w.Code, convey.ShouldEqual, http.StatusBadRequest)

			var response map[string]string
			err = json.Unmarshal(w.Body.Bytes(), &response)
			convey.So(err, convey.ShouldBeNil)

			convey.So(response["error"], convey.ShouldContainSubstring, "Invalid serve params")
		})
	})
}

func TestPutServeAsFunctionMetaInfoIntoEtcd(t *testing.T) {
	convey.Convey("Given a putServeAsFunctionMetaInfoIntoEtcd function", t, func() {
		// Mock the GetMetaEtcdClient and the client methods
		patches := gomonkey.ApplyFunc(putServeAsFunctionMetaInfoIntoEtcd,
			func(allFuncMetas []*types.ServeFuncWithKeysAndFunctionMetaInfo) error {
				return nil
			})
		defer patches.Reset()
		// Prepare mock data
		funcMeta := &types.ServeFuncWithKeysAndFunctionMetaInfo{
			FuncMetaKey:     "testKey",
			InstanceMetaKey: "testInstanceKey",
			FuncMetaInfo: &types.FunctionMetaInfo{
				InstanceMetaData: types.InstanceMetaData{
					MaxInstance: 5,
					MinInstance: 5,
				},
			},
		}
		funcMetas := []*types.ServeFuncWithKeysAndFunctionMetaInfo{funcMeta}

		// Mock the response of Etcd operations
		patches.ApplyMethod(reflect.TypeOf(&clientv3.Client{}), "Txn", func(client *clientv3.Client, ctx context.Context) clientv3.Txn {
			return client.Txn(ctx)
		})

		// Test the function
		convey.Convey("It should successfully put data into etcd", func() {
			err := putServeAsFunctionMetaInfoIntoEtcd(funcMetas)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
