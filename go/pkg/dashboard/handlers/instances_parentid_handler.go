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
	"yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/dashboard/etcdcache"
)

// InstancesByParentIDHandler function for /api/v1/instances?parent_id= route
func InstancesByParentIDHandler(ctx *gin.Context) {
	parentId := ctx.Query("parent_id")
	instancesMap := etcdcache.InstanceCache.GetByParentID(parentId)
	instances := make([]*InstanceInfo, 0, len(instancesMap))
	for _, instanceSpec := range instancesMap {
		instances = append(instances, instanceSpec2Instance(instanceSpec))
	}
	ctx.JSON(http.StatusOK, instances)
	log.GetLogger().Debugf("/instances?parent_id=%s succeed, instances:%s", parentId, instances)
}

func instanceSpec2Instance(instanceSpec *types.InstanceSpecification) *InstanceInfo {
	if instanceSpec == nil {
		return &InstanceInfo{}
	}
	instance := &InstanceInfo{
		ID:          instanceSpec.InstanceID,
		Status:      getInstanceStatus(instanceSpec.InstanceStatus.Code),
		CreateTime:  instanceSpec.Extensions.CreateTimestamp,
		JobId:       instanceSpec.JobID,
		PID:         instanceSpec.Extensions.PID,
		IP:          instanceSpec.RuntimeAddress,
		NodeID:      instanceSpec.FunctionProxyID,
		AgentID:     instanceSpec.FunctionAgentID,
		ParentID:    instanceSpec.ParentID,
		RequiredCPU: 0,
		RequiredMem: 0,
		RequiredGPU: 0,
		RequiredNPU: 0,
		Restarted:   instanceSpec.DeployTimes - 1,
		ExitDetail:  instanceSpec.InstanceStatus.Msg,
	}
	instance.setResources(instanceSpec.Resources.Resources)
	return instance
}

func (info *InstanceInfo) setResources(resourcesMap map[string]types.Resource) {
	info.RequiredCPU = getTypesResourceValue("CPU", resourcesMap)
	info.RequiredMem = getTypesResourceValue("Memory", resourcesMap)
	info.RequiredGPU = getTypesResourceValue("GPU", resourcesMap)
	info.RequiredNPU = getTypesResourceValue("NPU/.+/count", resourcesMap)
}

func getTypesResourceValue(resourceName string, resourcesMap map[string]types.Resource) float64 {
	if _, ok := resourcesMap[resourceName]; ok {
		return resourcesMap[resourceName].Scalar.Value
	}
	return 0
}
