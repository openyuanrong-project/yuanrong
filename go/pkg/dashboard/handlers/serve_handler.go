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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	clientv3 "go.etcd.io/etcd/client/v3"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/dashboard/models"
)

const (
	// from python/ray/serve/_private/common.py DeploymentStatus
	serveDeploymentStatusHealthy   = "HEALTHY"   // for normal
	serveDeploymentStatusUnhealthy = "UNHEALTHY" // for abnormal cases
	serveDeploymentStatusUpScaling = "UPSCALING" // for scaling

	// from python/ray/serve/_private/common.py ApplicationStatus
	serveAppStatusDeploying = "DEPLOYING" // for scaling
	serveAppStatusRunning   = "RUNNING"   // for normal

	defaultEtcdRequestTimeout = 30 * time.Second
)

// ServeApplicationStatus -
type ServeApplicationStatus struct {
	Deployments map[string]ServeDeploymentStatus `json:"deployments"`
	Name        string                           `json:"name,omitempty"`
	Status      string                           `json:"status"`
	Message     string                           `json:"message,omitempty"`
}

// ServeDeploymentStatus -
type ServeDeploymentStatus struct {
	Name    string `json:"name,omitempty"`
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}

func setServeGetErrorRsp(ctx *gin.Context, err error, msgPrefix string) {
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"message": msgPrefix + err.Error(),
		})
	}
}

// ServeGetHandler -
// RayService use this to check if the service is ready or not (some level a healthy probe)
// RayService needs ( len(Apps) > 0 and all apps.STATUS == "RUNNING" )
// Ray returns each app's status by
// ==> python/ray/serve/_private/controller.py:ServeController.get_serve_instance_details
// ==> python/ray/serve/_private/application_state.py:ApplicationStateManager.list_app_statuses
// ==> python/ray/serve/_private/application_state.py:ApplicationStateManager._determine_app_status
// ==> python/ray/serve/_private/deployment_state.py:DeploymentState.check_curr_status
// it is determined by every deployment, it any deployment is either not running or not have enough replicas, the app is
// not running.
// So, in yuanrong, we have some conditions
// 1. get all serve func by fetch /sn/functions, and fill the responses' `Applications` part
// 2. get all serve instances by fetch /sn/instances, and check each instances' status ( RUNNING and others )
func ServeGetHandler(ctx *gin.Context) {
	// get without request options
	serveFuncMetaInfos, err := getAllServeFunctions()
	if err != nil {
		setServeGetErrorRsp(ctx, err, "")
		return
	}
	serveInstances, err := getAllServeRunningInstances()
	if err != nil {
		setServeGetErrorRsp(ctx, err, "")
		return
	}
	serveApps, err := convertServeDeploymentFaasFunctionsToServeDetails(serveFuncMetaInfos, serveInstances)
	if err != nil {
		setServeGetErrorRsp(ctx, err, "")
		return
	}
	ctx.JSON(http.StatusOK, serveApps)
	return
}

func getAllServeRunningInstances() ([]*types.InstanceSpecification, error) {
	resp, err := etcd3.GetRouterEtcdClient().Get(
		etcd3.CreateEtcdCtxInfoWithTimeout(context.Background(), defaultEtcdRequestTimeout),
		"/sn/instance/business/yrk/tenant/default/function/0-system-serveExecutor",
		clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("invalid serve params: %s", err.Error())
	}
	var serveInstances []*types.InstanceSpecification
	for _, kv := range resp.Kvs {
		instSpec := &types.InstanceSpecification{}
		err := json.Unmarshal(kv.Value, instSpec)
		if err != nil {
			log.GetLogger().Warnf("failed to marshal instance spec: %s", err.Error())
		}
		if instSpec.InstanceStatus.Code == int32(constant.KernelInstanceStatusRunning) {
			serveInstances = append(serveInstances, instSpec)
		}
	}
	return serveInstances, nil
}

func getAllServeFunctions() ([]*types.FunctionMetaInfo, error) {
	// Step 1. get all serve functions
	resp, err := etcd3.GetMetaEtcdClient().Get(
		etcd3.CreateEtcdCtxInfoWithTimeout(context.Background(), defaultEtcdRequestTimeout),
		"/sn/functions", clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("invalid serve params: %s", err.Error())
	}
	var allServeFunc []*types.FunctionMetaInfo
	for _, kv := range resp.Kvs {
		functionInfo := types.FunctionMetaInfo{}
		err := json.Unmarshal(kv.Value, &functionInfo)
		if err != nil {
			return nil, fmt.Errorf("invalid serve params: %s", err.Error())
		}
		// yes, this is a serve func
		if len(functionInfo.ExtendedMetaData.ServeDeploySchema.Applications) > 0 {
			allServeFunc = append(allServeFunc, &functionInfo)
		}
	}
	return allServeFunc, nil
}

func getInstancesOfServeFunc(allRunningInstances []*types.InstanceSpecification,
	serveFunc *types.FunctionMetaInfo) []*types.InstanceSpecification {
	var matchInstances []*types.InstanceSpecification
	var faasKey string
	var keyNoteExists bool
	for _, inst := range allRunningInstances {
		if faasKey, keyNoteExists = inst.CreateOptions["FUNCTION_KEY_NOTE"]; !keyNoteExists {
			continue
		}
		meta := types.ServeFunctionKey{}
		_ = meta.FromFaasFunctionKey(faasKey)
		if meta.ToFaasFunctionVersionUrn() == serveFunc.FuncMetaData.FunctionVersionURN {
			matchInstances = append(matchInstances, inst)
		}
	}
	return matchInstances
}

func getServeDeploymentStatus(serveInstances []*types.InstanceSpecification,
	serveFunc *types.FunctionMetaInfo) (string, string) {
	if len(serveFunc.ExtendedMetaData.ServeDeploySchema.Applications) == 0 {
		log.GetLogger().Errorf("there is no application in %v", serveFunc.ExtendedMetaData.ServeDeploySchema)
		return serveDeploymentStatusUnhealthy, "no application info found in func meta info"
	}
	if len(serveFunc.ExtendedMetaData.ServeDeploySchema.Applications[0].Deployments) == 0 {
		log.GetLogger().Errorf("there is no deployments in %v", serveFunc.ExtendedMetaData.ServeDeploySchema)
		return serveDeploymentStatusUnhealthy, "no deployment info found in func meta info"
	}
	expectedReplicas := serveFunc.ExtendedMetaData.ServeDeploySchema.Applications[0].Deployments[0].NumReplicas
	if expectedReplicas == int64(len(serveInstances)) {
		return serveDeploymentStatusHealthy, "healthy"
	}
	return serveDeploymentStatusUpScaling, fmt.Sprintf("now: %d expect: %d", len(serveInstances), expectedReplicas)
}

// DeploymentStatus: UPDATING, HEALTHY, UNHEALTHY, UPSCALING, DOWNSCALING
// ApplicationStatus: NOT_STARTED, DEPLOYING, DEPLOY_FAILED, RUNNING, UNHEALTHY, DELETING
func convertServeDeploymentFaasFunctionsToServeDetails(allServeFunc []*types.FunctionMetaInfo,
	allRunningInstances []*types.InstanceSpecification) (models.ServeDetails,
	error) {
	serveDetails := models.ServeDetails{
		Applications: make(map[string]*models.ServeApplicationDetails),
	}
	for _, serveFunc := range allServeFunc {
		theCorrespondingServeDeploySchema := serveFunc.ExtendedMetaData.ServeDeploySchema
		if len(theCorrespondingServeDeploySchema.Applications) < 1 {
			log.GetLogger().Errorf("failed to validate app info from serve deploy schema, contains no app")
			continue
		}
		theCorrespondingServeAppSchema := theCorrespondingServeDeploySchema.Applications[0]
		if len(theCorrespondingServeAppSchema.Deployments) < 1 {
			log.GetLogger().Errorf("failed to validate app info from serve deploy schema, contains no deployment")
			continue
		}
		theCorrespondingServeDeploymentSchema := theCorrespondingServeAppSchema.Deployments[0]
		serveInstances := getInstancesOfServeFunc(allRunningInstances, serveFunc)

		status, statusMsg := getServeDeploymentStatus(serveInstances, serveFunc)
		dpDetails := models.ServeDeploymentDetails{
			ServeDeploymentStatus: models.ServeDeploymentStatus{
				Name:    theCorrespondingServeDeploymentSchema.Name,
				Status:  status,
				Message: statusMsg,
			},
			RoutePrefix: theCorrespondingServeAppSchema.RoutePrefix,
		}

		// no app, add one
		if _, ok := serveDetails.Applications[theCorrespondingServeAppSchema.Name]; !ok {
			serveDetails.Applications[theCorrespondingServeAppSchema.Name] = &models.ServeApplicationDetails{
				// `deployments` seems not really useful, just make it always ready and healthy
				Deployments: map[string]models.ServeDeploymentDetails{
					theCorrespondingServeDeploymentSchema.Name: dpDetails,
				},
				RoutePrefix: theCorrespondingServeDeploySchema.Applications[0].RoutePrefix,
				ServeApplicationStatus: models.ServeApplicationStatus{
					Name:   theCorrespondingServeAppSchema.Name,
					Status: serveAppStatusRunning,
				},
			}
		} else {
			// there is an existing app, add deployment into that one
			serveDetails.Applications[theCorrespondingServeAppSchema.Name].
				Deployments[theCorrespondingServeDeploymentSchema.Name] = dpDetails
		}
	}

	modifyAppStatusByDeploymentStatus(&serveDetails)
	return serveDetails, nil
}

func modifyAppStatusByDeploymentStatus(serveDetails *models.ServeDetails) {
	for i, app := range serveDetails.Applications {
		isRunning := true
		for _, dp := range app.Deployments {
			if dp.Status != serveDeploymentStatusHealthy {
				isRunning = false
			}
		}
		if !isRunning {
			serveDetails.Applications[i].Status = serveAppStatusDeploying
		}
	}
}

// ServeDelHandler -
func ServeDelHandler(ctx *gin.Context) {
	// get without request options
	ctx.JSON(http.StatusOK, gin.H{
		"message": "ok, but yuanrong doesn't support this right now",
	})
}

// ServePutHandler function for /serve routes
func ServePutHandler(ctx *gin.Context) {
	var serveDeploySchema types.ServeDeploySchema
	err := ctx.ShouldBindJSON(&serveDeploySchema)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid serve params: %s", err.Error()),
		})
		return
	}

	log.GetLogger().Infof("allIncomingDeploySchema: %#v\n", serveDeploySchema)
	if err = serveDeploySchema.Validate(); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid serve params: %s", err.Error()),
		})
		return
	}
	allServeFuncMetas := serveDeploySchema.ToFaaSFuncMetas()
	err = putServeAsFunctionMetaInfoIntoEtcd(allServeFuncMetas)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Failed to publish serve functions: %s", err.Error()),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "succeed",
	})
	return
}

type instanceConfiguration struct {
	InstanceMetaData types.InstanceMetaData `json:"instanceMetaData" valid:",optional"`
}

// use transaction to avoid partially success
func putServeAsFunctionMetaInfoIntoEtcd(allFuncMetas []*types.ServeFuncWithKeysAndFunctionMetaInfo) error {
	txn := etcd3.GetMetaEtcdClient().Client.Txn(context.Background())
	var ops []clientv3.Op
	for _, value := range allFuncMetas {
		funcMetaValue, err := json.Marshal(value.FuncMetaInfo)
		if err != nil {
			return err
		}
		instanceMetaValue, err := json.Marshal(instanceConfiguration{
			InstanceMetaData: value.FuncMetaInfo.InstanceMetaData,
		})
		if err != nil {
			return err
		}
		ops = append(ops, clientv3.OpPut(value.FuncMetaKey, string(funcMetaValue)))
		ops = append(ops, clientv3.OpPut(value.InstanceMetaKey, string(instanceMetaValue)))
	}
	commit, err := txn.Then(ops...).Commit()
	if err != nil {
		return err
	}
	if !commit.Succeeded {
		return fmt.Errorf("failed to put function meta into etcd")
	}
	return nil
}
