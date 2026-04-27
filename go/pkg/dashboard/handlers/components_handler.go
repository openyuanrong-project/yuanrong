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

// ComponentsAddition for response
type ComponentsAddition struct {
	Usage
	Nodes      []Component `form:"nodes" json:"nodes"`
	Components `form:"components" json:"components"`
}

// Components defines all Component
type Components map[string][]Component

// Component ComponentInfo
type Component struct {
	Hostname string `form:"hostname" json:"hostname"`
	Status   string `form:"status" json:"status"`
	Address  string `form:"address" json:"address"`
	Usage
}

// ComponentsHandler function for /components route
func ComponentsHandler(ctx *gin.Context) {
	resource, err := getinfo.GetResources()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": errGetResources,
			"msg":  "fail",
			"data": "",
		})
		log.GetLogger().Errorf("/components GetResources, %d, %s", errGetResources, err)
		return
	}
	compAdd, err := PBToCompAdd(resource)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": errPBToCompAdd,
			"msg":  "fail",
			"data": "",
		})
		log.GetLogger().Errorf("/components PBToComponents %d %s", errPBToCompAdd, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "succeed",
		"data": compAdd,
	})
	log.GetLogger().Debugf("/components succeed, data: %#v", compAdd)
}

// PBToCompAdd pb data switch to ComponentsAddition struct
func PBToCompAdd(resource *resource.ResourceUnit) (ComponentsAddition, error) {
	var compAdd ComponentsAddition
	var err error
	compAdd.Usage, err = PBToUsage(resource)
	if err != nil {
		return ComponentsAddition{}, err
	}
	comps := Components{}
	for _, f := range resource.Fragment {
		comp := Component{
			Hostname: f.Id,
			Status:   "alive",
		}
		if _, ok := f.NodeLabels["HOST_IP"]; ok {
			for address := range f.NodeLabels["HOST_IP"].Items {
				comp.Address = address
			}
		}
		comp.Usage, err = PBToUsage(f)
		if err != nil {
			return ComponentsAddition{}, err
		}
		comps[f.OwnerId] = append(comps[f.OwnerId], comp)
	}
	compAdd.Components = comps
	var nodes []Component
	for k, compUnit := range comps {
		node := Component{
			Hostname: k,
			Status:   "healthy",
			Address:  compUnit[0].Address,
		}
		node.AllocNPU = compUnit[0].AllocNPU
		for _, component := range compUnit {
			node.CapCPU += component.CapCPU
			node.CapMem += component.CapMem
			node.AllocCPU += component.AllocCPU
			node.AllocMem += component.AllocMem
		}
		nodes = append(nodes, node)
	}
	compAdd.Nodes = nodes
	return compAdd, nil
}
