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

func TestInstancesByInstanceIDHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	convey.Convey("Test InstancesByInstanceIDHandler:", t, func() {
		patches := gomonkey.ApplyMethodReturn(&etcdcache.InstanceCache, "Get", &types.InstanceSpecification{
			InstanceID:     "id-123",
			InstanceStatus: types.InstanceStatus{},
			Extensions:     types.Extensions{},
			Resources:      types.Resources{Resources: map[string]types.Resource{}},
		})
		defer patches.Reset()

		r := gin.Default()
		r.GET("/instances/:instance-id", InstancesByInstanceIDHandler)
		req, err := http.NewRequest("GET", "/instances/id-123", nil)
		convey.So(err, convey.ShouldBeNil)
		w := httptest.NewRecorder()

		convey.Convey("Test InstancesByInstanceID success", func() {
			r.ServeHTTP(w, req)
			convey.So(w.Code, convey.ShouldEqual, http.StatusOK)
			convey.So(string(w.Body.Bytes()), convey.ShouldContainSubstring, "id-123")
		})
	})
}
