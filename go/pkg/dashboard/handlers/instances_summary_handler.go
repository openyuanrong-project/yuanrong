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

	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/resource"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/dashboard/getinfo"
)

// InstSummary for overviews page show instances summary
type InstSummary struct {
	Total   int `form:"total" json:"total"`
	Running int `form:"running" json:"running"`
	Exited  int `form:"exited" json:"exited"`
	Fatal   int `form:"fatal" json:"fatal"`
}

const (
	// RUNNING code
	RUNNING = 3
	// EXITED code
	EXITED = 8
	// FATAL code
	FATAL = 6
)

// InstancesSummaryHandler function for /instances/summary route
func InstancesSummaryHandler(ctx *gin.Context) {
	pbInfos, err := getinfo.GetInstances(listInstPath)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": errGetInstances,
			"msg":  "fail",
			"data": "",
		})
		log.GetLogger().Errorf("/instances/summary GetInst %d %v", errGetInstances, err)
		return
	}

	instSumm := PBToInstSummary(pbInfos)
	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "succeed",
		"data": instSumm,
	})
	log.GetLogger().Debugf("/instances/summary succeed, data:%#v", instSumm)
}

// PBToInstSummary function for switch pb struct to InstanceInfos struct
func PBToInstSummary(infos []*resource.InstanceInfo) InstSummary {
	var instSumm InstSummary
	instSumm.Total = len(infos)
	for _, info := range infos {
		switch info.InstanceStatus.Code {
		case RUNNING:
			instSumm.Running++
		case EXITED:
			instSumm.Exited++
		case FATAL:
			instSumm.Fatal++
		default:
		}
	}
	return instSumm
}
