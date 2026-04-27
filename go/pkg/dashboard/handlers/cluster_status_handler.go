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
	"math"
	"net/http"

	"github.com/gin-gonic/gin"

	"yuanrong.org/kernel/pkg/common/constants"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/dashboard/getinfo"
)

const pendingInstPath = "/global-scheduler/scheduling_queue"

var resourcesName = []string{"CPU", "NPU", "GPU", "Memory"} // enum

var decimal = 100.0

// ClusterStatusHandler function for /api/cluster_status route
func ClusterStatusHandler(ctx *gin.Context) {
	format := ctx.Query("format")
	if format != "1" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"result": false,
			"msg":    "Please check request params.",
			"data":   "Resources:\nDemands:",
		})
		log.GetLogger().Errorf("/api/cluster_status failed, format=%s", format)
		return
	}
	pbInfos, err := getinfo.GetInstances(pendingInstPath)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"result": false,
			"msg":    "Failed to get formatted cluster status. Error: " + err.Error(),
			"data":   "Resources:\nDemands:",
		})
		log.GetLogger().Errorf("/api/cluster_status failed, %d %v", errGetInstances, err)
		return
	}

	resourcesMap := initResourcesMap()
	for _, pbInfo := range pbInfos {
		if pbInfo.Resources == nil {
			continue
		}
		for name, rsrc := range pbInfo.Resources.Resources {
			if rsrc.Scalar.Value == 0 || !validateName(name) {
				continue
			}
			resourcesMap[name][rsrc.Scalar.Value]++
		}
	}
	resStr := formatResourcesMap(resourcesMap)

	ctx.JSON(http.StatusOK, gin.H{
		"result": true,
		"msg":    "Got formatted cluster status.",
		"data": map[string]string{
			"clusterStatus": "Resources:\nDemands:" + resStr,
		},
	})
	log.GetLogger().Debugf("/api/cluster_status succeed")
}

func initResourcesMap() map[string]map[float64]int {
	resourcesMap := make(map[string]map[float64]int)
	for _, name := range resourcesName {
		resourcesMap[name] = make(map[float64]int)
	}
	return resourcesMap
}

func validateName(newName string) bool {
	for _, name := range resourcesName {
		if newName == name {
			return true
		}
	}
	return false
}

func formatResourcesMap(resourcesMap map[string]map[float64]int) string {
	var resStr string
	for name, m := range resourcesMap {
		if name == "Memory" {
			name = "memory"
		}
		for value, count := range m {
			if name == "memory" {
				value = value * constants.MemoryUnitConvert * constants.MemoryUnitConvert
			}
			if name == "CPU" {
				value /= constants.CpuUnitConvert
			}

			value = math.Ceil(value*decimal) / decimal
			resStr += fmt.Sprintf("\n{'%s':%.2f}: %d+ pending tasks/actors", name, value, count)
		}
	}
	return resStr
}
