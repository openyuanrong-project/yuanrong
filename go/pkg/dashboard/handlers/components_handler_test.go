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
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agiledragon/gomonkey"
	"github.com/gin-gonic/gin"
	"github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/proto"

	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/message"
	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/resource"
	"yuanrong.org/kernel/pkg/dashboard/flags"
)

func TestComponentsHandler(t *testing.T) {
	convey.Convey("Test ComponentsHandler:", t, func() {
		r := gin.Default()
		r.GET("/components", ComponentsHandler)
		req, err := http.NewRequest("GET", "/components", nil)
		convey.So(err, convey.ShouldBeNil)

		w := httptest.NewRecorder()

		convey.Convey("Test Components when function master error", func() {
			r.ServeHTTP(w, req)
			convey.So(w.Code, convey.ShouldEqual, http.StatusInternalServerError)
			convey.So(string(w.Body.Bytes()), convey.ShouldContainSubstring, "2001")
		})

		resources := &resource.Resources{
			Resources: map[string]*resource.Resource{
				"Memory": {
					Name:   "Memory",
					Scalar: &resource.Value_Scalar{Value: 38912},
				},
				"CPU": {
					Name:   "CPU",
					Scalar: &resource.Value_Scalar{Value: 10000},
				},
			},
		}
		nodeLabels := map[string]*resource.Value_Counter{
			"NODE_ID": {
				Items: map[string]uint64{"dggphis232340-1936114": 1},
			},
		}
		fragment1 := &resource.ResourceUnit{
			Id:          "function-agent-127.0.0.1-31630",
			Capacity:    resources,
			Allocatable: resources,
			NodeLabels:  nodeLabels,
			OwnerId:     "dggphis232340-1936114",
		}
		resourceInfo := message.ResourceInfo{
			RequestID: "145aa8bc-d616-4000-8000-000000734df7",
			Resource: &resource.ResourceUnit{
				Id:          "InnerDomainScheduler",
				Capacity:    resources,
				Allocatable: resources,
				Fragment: map[string]*resource.ResourceUnit{
					"function-agent-127.0.0.1-31630": fragment1,
				},
				NodeLabels:   nodeLabels,
				Revision:     19,
				ViewInitTime: "542b3f0f-0000-4000-8000-0089a640b071",
			},
		}
		resourcesServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resourceInfoBytes, _ := proto.Marshal(&resourceInfo)
			w.WriteHeader(http.StatusOK)
			w.Write(resourceInfoBytes)
		}))
		defer resourcesServer.Close()
		flags.DashboardConfig.FunctionMasterAddr = resourcesServer.URL

		convey.Convey("Test Components success", func() {
			r.ServeHTTP(w, req)
			convey.So(w.Code, convey.ShouldEqual, http.StatusOK)
			convey.So(string(w.Body.Bytes()), convey.ShouldContainSubstring, `"msg":"succeed"`)
		})
		convey.Convey("Test Components when PBToUsage error", func() {
			patches := gomonkey.ApplyFunc(PBToUsage, func(resource *resource.ResourceUnit) (Usage, error) {
				return Usage{}, errors.New("PBToUsage error")
			})
			defer patches.Reset()
			r.ServeHTTP(w, req)
			convey.So(w.Code, convey.ShouldEqual, http.StatusInternalServerError)
			convey.So(string(w.Body.Bytes()), convey.ShouldContainSubstring, "1001")
		})
	})
}
