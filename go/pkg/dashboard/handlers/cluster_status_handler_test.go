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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/proto"

	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/message"
	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/resource"
	"yuanrong.org/kernel/pkg/dashboard/flags"
)

func TestClusterStatusHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	convey.Convey("Test ClusterStatusHandler:", t, func() {
		resources1 := resource.Resources{
			Resources: map[string]*resource.Resource{
				"CPU": {
					Name: "CPU",
					Scalar: &resource.Value_Scalar{
						Value: 1.57,
					},
				},
				"Memory": {
					Name: "Memory",
					Scalar: &resource.Value_Scalar{
						Value: 300,
					},
				},
				"GPU": {
					Name: "GPU",
					Scalar: &resource.Value_Scalar{
						Value: 0,
					},
				},
				"NPU": {
					Name: "NPU",
					Scalar: &resource.Value_Scalar{
						Value: 0,
					},
				},
			},
		}
		resources2 := resource.Resources{
			Resources: map[string]*resource.Resource{
				"CPU": {
					Name: "CPU",
					Scalar: &resource.Value_Scalar{
						Value: 0.23,
					},
				},
				"Memory": {
					Name: "Memory",
					Scalar: &resource.Value_Scalar{
						Value: 300,
					},
				},
				"memory": {
					Name: "memory",
					Scalar: &resource.Value_Scalar{
						Value: 300,
					},
				},
			},
		}
		instances := []*resource.InstanceInfo{
			&resource.InstanceInfo{Resources: &resources1},
			&resource.InstanceInfo{Resources: &resources2},
			&resource.InstanceInfo{},
		}

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

		r := gin.Default()
		r.GET("/api/cluster_status", ClusterStatusHandler)
		req, err := http.NewRequest("GET", "/api/cluster_status?format=1", nil)
		convey.So(err, convey.ShouldBeNil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		newResources := map[string]float64{
			"CPU":    0.0,
			"GPU":    0.0,
			"memory": 0.0,
			"NPU":    0.0,
		}

		convey.Convey("It should return status 200 and appropriate message", func() {
			convey.So(w.Code, convey.ShouldEqual, http.StatusOK)
			var res map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &res)
			convey.So(err, convey.ShouldBeNil)
			clusterStatus := res["data"].(map[string]interface{})["clusterStatus"].(string)
			demands := strings.Split(clusterStatus, "Resources")[1]
			infos := strings.Split(strings.Split(demands, "Demands")[1], "\n")
			for _, info := range infos {
				if !strings.Contains(info, "pending tasks/actors") {
					continue
				}
				n := strings.LastIndexByte(info, ':')
				var subInfo map[string]float64
				newInfo := strings.ReplaceAll(info, `'`, `"`)
				err = json.Unmarshal([]byte(newInfo[:n]), &subInfo)
				convey.So(err, convey.ShouldBeNil)
				replicas, err := strconv.Atoi(strings.Split(info[n+2:], "+")[0])
				convey.So(err, convey.ShouldBeNil)
				for k, v := range subInfo {
					newResources[k] += float64(replicas) * v
				}
			}
			convey.So(newResources["CPU"], convey.ShouldNotEqual, 0)
			convey.So(newResources["GPU"], convey.ShouldEqual, 0)
		})
	})
}

func TestClusterStatusHandlerError(t *testing.T) {
	convey.Convey("Test ClusterStatusHandler error:", t, func() {
		r := gin.Default()
		r.GET("/api/cluster_status", ClusterStatusHandler)
		req, err := http.NewRequest("GET", "/api/cluster_status?format=1", nil)
		convey.So(err, convey.ShouldBeNil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		convey.Convey("test ClusterStatusHandler when function master error", func() {
			convey.So(w.Code, convey.ShouldEqual, http.StatusInternalServerError)
			convey.So(w.Body.String(), convey.ShouldContainSubstring, "Failed to get formatted cluster status.")
		})
	})
}
