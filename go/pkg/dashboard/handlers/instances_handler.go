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

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/resource"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/dashboard/getinfo"
)

// InstanceInfo is instance base info
type InstanceInfo struct {
	ID          string  `form:"id" json:"id"`
	Status      string  `form:"status" json:"status"`
	CreateTime  string  `form:"create_time" json:"create_time"`
	JobId       string  `form:"job_id" json:"job_id"`
	PID         string  `form:"pid" json:"pid"`
	IP          string  `form:"ip" json:"ip"`
	NodeID      string  `form:"node_id" json:"node_id"`
	AgentID     string  `form:"agent_id" json:"agent_id"`
	ParentID    string  `form:"parent_id" json:"parent_id"`
	RequiredCPU float64 `form:"required_cpu" json:"required_cpu"`
	RequiredMem float64 `form:"required_mem" json:"required_mem"`
	RequiredGPU float64 `form:"required_gpu" json:"required_gpu"`
	RequiredNPU float64 `form:"required_npu" json:"required_npu"`
	Restarted   int32   `form:"restarted" json:"restarted"`
	ExitDetail  string  `form:"exit_detail" json:"exit_detail"`
}

// InstanceInfos is all instance info
type InstanceInfos []InstanceInfo

const listInstPath = "/instance-manager/queryinstances"

var instanceStatusMap = map[constant.InstanceStatus]string{
	constant.KernelInstanceStatusExited:         "exited",
	constant.KernelInstanceStatusNew:            "new",
	constant.KernelInstanceStatusScheduling:     "scheduling",
	constant.KernelInstanceStatusCreating:       "creating",
	constant.KernelInstanceStatusRunning:        "running",
	constant.KernelInstanceStatusFailed:         "failed",
	constant.KernelInstanceStatusExiting:        "exiting",
	constant.KernelInstanceStatusFatal:          "fatal",
	constant.KernelInstanceStatusScheduleFailed: "schedule_failed",
	constant.KernelInstanceStatusEvicting:       "evicting",
	constant.KernelInstanceStatusEvicted:        "evicted",
	constant.KernelInstanceStatusSubHealth:      "sub_health",
}

// InstancesHandler function for /instances route
func InstancesHandler(ctx *gin.Context) {
	if ctx.Query("parent_id") != "" {
		InstancesByParentIDHandler(ctx)
		return
	}
	pbInfos, err := getinfo.GetInstances(listInstPath)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": errGetInstances,
			"msg":  "fail",
			"data": "",
		})
		log.GetLogger().Errorf("/instances GetInst, %d %v", errGetInstances, err)
		return
	}

	infos := PBToInstanceInfos(pbInfos)
	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "succeed",
		"data": infos,
	})
	log.GetLogger().Debugf("/instances succeed, infos:%#v", infos)
}

// PBToInstanceInfos function switch pb struct to InstanceInfos struct
func PBToInstanceInfos(pbInfos []*resource.InstanceInfo) InstanceInfos {
	infos := make(InstanceInfos, 0)
	for _, pbInfo := range pbInfos {
		info := InstanceInfo{
			ID:         pbInfo.InstanceID,
			Status:     getInstanceStatus(pbInfo.InstanceStatus.Code),
			JobId:      pbInfo.JobID,
			IP:         pbInfo.RuntimeAddress,
			NodeID:     pbInfo.FunctionProxyID,
			AgentID:    pbInfo.FunctionAgentID,
			ParentID:   pbInfo.ParentID,
			ExitDetail: pbInfo.InstanceStatus.Msg,
			Restarted:  pbInfo.DeployTimes - 1,
		}
		info.setFields(pbInfo)
		infos = append(infos, info)
	}
	return infos
}

func getInstanceStatus(code int32) string {
	if status, ok := instanceStatusMap[constant.InstanceStatus(code)]; ok {
		return status
	}
	return ""
}

func (info *InstanceInfo) setFields(pbInfo *resource.InstanceInfo) {
	if createTime, ok := pbInfo.Extensions["createTimestamp"]; ok {
		info.CreateTime = createTime
	}
	if pid, ok := pbInfo.Extensions["pid"]; ok {
		info.PID = pid
	}
	if pbInfo.Resources == nil || pbInfo.Resources.Resources == nil {
		return
	}
	resourcesMap := pbInfo.Resources.Resources
	info.RequiredCPU = getResourceValue("CPU", resourcesMap)
	info.RequiredMem = getResourceValue("Memory", resourcesMap)
	info.RequiredGPU = getResourceValue("GPU", resourcesMap)
	info.RequiredNPU = getResourceValue("NPU/.+/count", resourcesMap)
}

func getResourceValue(resourceName string, resourcesMap map[string]*resource.Resource) float64 {
	if _, ok := resourcesMap[resourceName]; ok {
		return resourcesMap[resourceName].Scalar.Value
	}
	return 0
}
