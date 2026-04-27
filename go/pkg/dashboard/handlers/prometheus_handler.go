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
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/dashboard/flags"
	"yuanrong.org/kernel/pkg/dashboard/getinfo"
)

// PrometheusHandler -
func PrometheusHandler(ctx *gin.Context) {
	// reconnect
	if getinfo.PromClient == nil {
		getinfo.InitPromClient()
	}

	if getinfo.PromClient == nil {
		log.GetLogger().Errorf("failed to connect prometheus, prometheus address: %v", flags.DashboardConfig.PrometheusAddr)
		ctx.JSON(http.StatusBadRequest,
			fmt.Sprintf("failed to connect prometheus, prometheus address: %v", flags.DashboardConfig.PrometheusAddr))
		return
	}

	query := ctx.Query("query")
	res, warnings, err := getinfo.PromClient.Query(ctx, query, time.Time{})
	if err != nil {
		log.GetLogger().Errorf("prometheus query failed, error: %v", err.Error())
		ctx.JSON(http.StatusBadRequest, fmt.Sprintf("prometheus query failed, error: %v", err.Error()))
		return
	}

	if len(warnings) > 0 {
		log.GetLogger().Errorf("prometheus query return warnings: %v", warnings)
		ctx.JSON(http.StatusBadRequest, fmt.Sprintf("prometheus query return warnings: %v", warnings))
		return
	}

	ctx.JSON(http.StatusOK, res)
	log.GetLogger().Debugf("/prometheus/query succeed, infos:%#v", res)
}
