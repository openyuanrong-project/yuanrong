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

	"github.com/gin-gonic/gin"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/dashboard/etcdcache"
)

// InstancesByInstanceIDHandler function for /api/v1/instances/:instance-id route
func InstancesByInstanceIDHandler(ctx *gin.Context) {
	instanceId := ctx.Param("instance-id")
	instanceSpec := etcdcache.InstanceCache.Get(instanceId)
	instance := instanceSpec2Instance(instanceSpec)
	ctx.JSON(http.StatusOK, instance)
	log.GetLogger().Debugf("/instances/%s succeed, instanceInfo: %#v", instanceId, instance)
}
