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
	"strings"

	"github.com/gin-gonic/gin"

	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/resource"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/dashboard/getinfo"
)

// Usage for check cpu and mem usage
type Usage struct {
	CapCPU   float64 `form:"cap_cpu" json:"cap_cpu"`
	CapMem   float64 `form:"cap_mem" json:"cap_mem"`
	AllocCPU float64 `form:"alloc_cpu" json:"alloc_cpu"`
	AllocMem float64 `form:"alloc_mem" json:"alloc_mem"`
	AllocNPU float64 `form:"alloc_npu" json:"alloc_npu"`
}

// RsrcSummary for check summary
type RsrcSummary struct {
	Usage
	ProxyNum int `form:"proxy_num" json:"proxy_num"`
}

// ResourcesSummaryHandler function for /logical-resources/summary route
func ResourcesSummaryHandler(ctx *gin.Context) {
	resource, err := getinfo.GetResources()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": errGetResources,
			"msg":  "fail",
			"data": "",
		})
		log.GetLogger().Errorf("/logical-resources/summary GetResources %d %v", errGetResources, err)
		return
	}

	rsrcSummary, err := PBToRsrcSummary(resource)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": errPBToRsrcSummary,
			"msg":  "fail",
			"data": "",
		})
		log.GetLogger().Warnf("/logical-resources/summary PBToSummary %d %v", errPBToRsrcSummary, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "succeed",
		"data": rsrcSummary,
	})
	log.GetLogger().Debugf("/logical-resources/summary succeed, data:%#v", rsrcSummary)
}

// PBToUsage function for pb data switch to Usage struct
func PBToUsage(resource *resource.ResourceUnit) (Usage, error) {
	var usage Usage
	capCPU, ok := resource.Capacity.Resources["CPU"]
	if !ok {
		return Usage{}, errors.New(`no resource.Capacity.Resources["CPU"]`)
	}
	if capCPU != nil && capCPU.Scalar != nil {
		usage.CapCPU = capCPU.Scalar.Value
	}
	capMem, ok := resource.Capacity.Resources["Memory"]
	if !ok {
		return Usage{}, errors.New(`no resource.Capacity.Resources["Memory"]`)
	}
	if capMem != nil && capMem.Scalar != nil {
		usage.CapMem = capMem.Scalar.Value
	}
	allocResources := resource.Allocatable.Resources
	allocCPU, ok := allocResources["CPU"]
	if !ok {
		return Usage{}, errors.New(`no resource.Allocatable.Resources["CPU"]`)
	}
	if allocCPU != nil && allocCPU.Scalar != nil {
		usage.AllocCPU = allocCPU.Scalar.Value
	}
	allocMem, ok := allocResources["Memory"]
	if !ok {
		return Usage{}, errors.New(`no resource.Allocatable.Resources["Memory"]`)
	}
	if allocMem != nil && allocMem.Scalar != nil {
		usage.AllocMem = allocMem.Scalar.Value
	}
	for resourceName, resourceData := range allocResources {
		if strings.Contains(resourceName, "NPU") {
			usage.AllocNPU = resourceData.Scalar.Value
		}
	}
	return usage, nil
}

// PBToRsrcSummary function for pb data switch to Summary struct
func PBToRsrcSummary(resource *resource.ResourceUnit) (RsrcSummary, error) {
	var summary RsrcSummary
	var err error
	summary.Usage, err = PBToUsage(resource)
	if err != nil {
		return RsrcSummary{}, err
	}
	summary.ProxyNum = len(resource.NodeLabels["NODE_ID"].Items)
	return summary, nil
}
