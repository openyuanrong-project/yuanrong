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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/proto"

	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/message"
	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/resource"
	"yuanrong.org/kernel/pkg/dashboard/flags"
)

func TestInstancesSummaryHandler(t *testing.T) {
	convey.Convey("Test InstancesSummaryHandler:", t, func() {
		r := gin.Default()
		r.GET("/instances/summary", InstancesSummaryHandler)
		req, err := http.NewRequest("GET", "/instances/summary", nil)
		convey.So(err, convey.ShouldBeNil)

		w := httptest.NewRecorder()

		convey.Convey("Test InstancesSummary when function master error", func() {
			r.ServeHTTP(w, req)
			convey.So(w.Code, convey.ShouldEqual, http.StatusInternalServerError)
			convey.So(string(w.Body.Bytes()), convey.ShouldContainSubstring, "2002")
		})

		resources := resource.Resources{
			Resources: map[string]*resource.Resource{
				"CPU": {
					Name: "CPU",
					Scalar: &resource.Value_Scalar{
						Value: 500,
					},
				},
				"Memory": {
					Name: "Memory",
					Scalar: &resource.Value_Scalar{
						Value: 300,
					},
				},
			},
		}

		instance1 := resource.InstanceInfo{
			InstanceID:      "app-dfb5ed67-9342-4b50-82ff-8fa8f055a9f4",
			InstanceStatus:  &resource.InstanceStatus{Code: 3},
			FunctionAgentID: "function-agent-127.0.0.1-31630",
			FunctionProxyID: "phish232340-1936114",
			ParentID:        "driver-faas-frontend-dggphis232340-1936114",
			Resources:       &resources,
		}
		instance2 := instance1
		instance2.InstanceStatus = &resource.InstanceStatus{Code: 8}
		instance3 := instance1
		instance3.InstanceStatus = &resource.InstanceStatus{Code: 6}
		instances := []*resource.InstanceInfo{&instance1, &instance2, &instance3}

		instancesInfo := message.QueryInstancesInfoResponse{
			InstanceInfos: instances,
		}
		//message.QueryInstancesInfoResponse
		instancesServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			instancesInfoBytes, _ := proto.Marshal(&instancesInfo)
			w.WriteHeader(http.StatusOK)
			w.Write(instancesInfoBytes)
		}))
		defer instancesServer.Close()
		flags.DashboardConfig.FunctionMasterAddr = instancesServer.URL

		convey.Convey("Test InstancesSummary success", func() {
			r.ServeHTTP(w, req)
			convey.So(w.Code, convey.ShouldEqual, http.StatusOK)
			convey.So(string(w.Body.Bytes()), convey.ShouldContainSubstring, `"msg":"succeed"`)
		})
	})
}
