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
)

// ComponentsByComponentIDHandler function for /components/:component-id route
func ComponentsByComponentIDHandler(ctx *gin.Context) {
	componentId := ctx.Param("component-id")
	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "succeed",
		"data": "Components by ComponentID:" + componentId + " json",
	})
	log.GetLogger().Debugf("/components/%s succeed", componentId)
}
