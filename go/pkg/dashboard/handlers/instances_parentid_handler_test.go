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

	"github.com/agiledragon/gomonkey/v2"
	"github.com/gin-gonic/gin"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/dashboard/etcdcache"
)

func TestInstancesByParentIDHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	convey.Convey("Test InstancesByParentIDHandler:", t, func() {
		instance := &types.InstanceSpecification{
			InstanceID:     "id-123",
			InstanceStatus: types.InstanceStatus{},
			Extensions:     types.Extensions{},
			Resources: types.Resources{Resources: map[string]types.Resource{
				"CPU":    types.Resource{Scalar: types.ValueScalar{}},
				"Memory": types.Resource{Scalar: types.ValueScalar{}},
				"GPU":    types.Resource{Scalar: types.ValueScalar{}},
				"NPU":    types.Resource{Scalar: types.ValueScalar{}},
			}},
		}

		r := gin.Default()
		r.GET("/api/v1/instances", InstancesHandler)
		req, err := http.NewRequest("GET", "/api/v1/instances?parent_id="+instance.InstanceID, nil)
		convey.So(err, convey.ShouldBeNil)
		w := httptest.NewRecorder()

		convey.Convey("Test InstancesByParentID success", func() {
			patches := gomonkey.ApplyMethodReturn(&etcdcache.InstanceCache, "GetByParentID",
				map[string]*types.InstanceSpecification{instance.InstanceID: instance})
			defer patches.Reset()

			r.ServeHTTP(w, req)
			convey.So(w.Code, convey.ShouldEqual, http.StatusOK)
			convey.So(string(w.Body.Bytes()), convey.ShouldContainSubstring, instance.InstanceID)
		})
	})
}

func TestInstanceSpec2Instance(t *testing.T) {
	convey.Convey("Test instanceSpec2Instance:", t, func() {
		convey.Convey("Test instanceSpec2Instance success", func() {
			instance := instanceSpec2Instance(nil)
			convey.So(instance.ID, convey.ShouldBeEmpty)
		})
	})
}
