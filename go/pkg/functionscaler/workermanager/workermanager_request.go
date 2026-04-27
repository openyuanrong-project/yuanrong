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

package workermanager

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	"yuanrong.org/kernel/pkg/common/faas_common/statuscode"
	"yuanrong.org/kernel/pkg/common/faas_common/urnutils"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

var nodeID = os.Getenv("NODE_IP")

// ScaleUpParam scale up param
type ScaleUpParam struct {
	TraceID     string
	FunctionKey string
	NodeLabel   string
	Timeout     time.Duration
	CPU         int
	Memory      int
}

// GetWorkerSuccessResponse define
type GetWorkerSuccessResponse struct {
	Worker   *Worker           `json:"worker"`
	Instance *types.WmInstance `json:"instance"`
	Code     int               `json:"code"`
	Message  string            `json:"message"`
}

// Worker define a worker
type Worker struct {
	Instances       []*types.WmInstance `json:"instances"`
	FunctionName    string              `json:"functionname"`
	FunctionVersion string              `json:"functionversion"`
	Tenant          string              `json:"tenant"`
	Business        string              `json:"business"`
}

// DeployParam Deploy function param
type DeployParam struct {
	FuncName        string // deployed function name
	Applier         string // applier instance id
	DeployNode      string // deploy node id
	Business        string // business
	TenantID        string // tenant id
	Version         string // version
	OwnerIP         string // owner IP
	TraceID         string // trace id
	TriggerFlag     string // This is a trigger request flag.
	StateID         string
	CPU             int
	Memory          int
	OwnedWorkerView WorkersView // proxy owned worker view
}

// WorkersView proxy workers view
type WorkersView struct {
	// OwnedNum -1 num means scale up 1 worker by force
	// other num means proxy owned worker nums
	OwnedNum int `json:"ownedNum"`
	// CurrentWorkerNum -
	CurrentWorkerNum int `json:"currentWorkerNum"`
	// ScalingWorkersNum -
	ScalingWorkersNum int `json:"scalingWorkersNum"`
}

// DeleteParam Delete function param
type DeleteParam struct {
	InstanceID         string `json:"instance_id"`
	FuncName           string `json:"function_name"`
	FuncVersion        string `json:"function_version"`
	BusinessID         string `json:"business_id"`
	TenantID           string `json:"tenant_id"`
	Applier            string `json:"applier"`
	IsBrokenConnection bool   `json:"broken_connection_status"`
}

// DeleteWorkerResponse is response to delete worker
type DeleteWorkerResponse struct {
	Code     int    `json:"code"`
	Reserved bool   `json:"reserved"`
	Message  string `json:"message"`
}

// ScaleUpInstance send scale up req to worker manager
func ScaleUpInstance(scaleUpParam *ScaleUpParam) (*types.WmInstance, error) {
	anonymizedFuncKeyWithRes := urnutils.AnonymizeTenantKey(scaleUpParam.FunctionKey)
	ctx := context.TODO()
	ctx, cancel := context.WithTimeout(ctx, scaleUpParam.Timeout)
	defer cancel()
	request := makeScaleUpRequest(ctx, scaleUpParam)
	if request == nil {
		log.GetLogger().Errorf("traceID: %s | FuncKeyWithRes: %s | failed to make scale up request",
			scaleUpParam.TraceID, anonymizedFuncKeyWithRes)
		return nil, errors.New("make scale up request failed")
	}
	resp, err := GetWorkerManagerClient().Do(request)
	if err != nil {
		log.GetLogger().Errorf("traceID: %s | FuncKeyWithRes: %s | failed to send scale up request "+
			"to worker manager: %s", scaleUpParam.TraceID, anonymizedFuncKeyWithRes, err.Error())
		return nil, err
	}
	instance, err := handleScaleUpResponseFromWorkerManager(resp, scaleUpParam, anonymizedFuncKeyWithRes)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func handleScaleUpResponseFromWorkerManager(resp *http.Response,
	scaleUpParam *ScaleUpParam, anonymizedFuncKeyWithRes string) (*types.WmInstance, error) {
	respBody, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.GetLogger().Errorf("failed to get response body: %s, functionKey: %s, traceID: %s", err.Error(),
			urnutils.AnonymizeTenantKey(scaleUpParam.FunctionKey), scaleUpParam.TraceID)
		return nil, snerror.New(statuscode.ScaleUpRequestErrCode, statuscode.ScaleUpRequestErrMsg)
	}
	if resp.StatusCode != http.StatusOK {
		snErr := &snerror.BadResponse{}
		err = json.Unmarshal(respBody, snErr)
		if err != nil {
			log.GetLogger().Errorf("traceID: %s | FuncKeyWithRes: %s | failed to unmarshal scale up "+
				"response: %s", scaleUpParam.TraceID, anonymizedFuncKeyWithRes, err.Error())
			return nil, err
		}
		log.GetLogger().Errorf("traceID: %s | FuncKeyWithRes: %s | errCode: %d | errMessage: %s | failed to "+
			"scale up function instance", scaleUpParam.TraceID, anonymizedFuncKeyWithRes, snErr.Code, snErr.Message)
		return nil, snerror.New(snErr.Code, snErr.Message)
	}
	scaleUpResp := GetWorkerSuccessResponse{}
	if err = json.Unmarshal(respBody, &scaleUpResp); err != nil {
		log.GetLogger().Errorf("failed to unmarshal scaleUp response: %s, functionKey: %s, traceID: %s, body %s",
			err, urnutils.AnonymizeTenantKey(scaleUpParam.FunctionKey), scaleUpParam.TraceID, string(respBody))
		return nil, snerror.New(statuscode.ScaleUpRequestErrCode, statuscode.ScaleUpRequestErrMsg)
	}
	if scaleUpResp.Code == statuscode.InnerRuntimeInitTimeoutCode && scaleUpResp.Worker != nil {
		log.GetLogger().Infof("deploy functionKey %s instance: %v traceID: %s with code %d msg %s",
			urnutils.AnonymizeTenantKey(scaleUpParam.FunctionKey), scaleUpResp.Instance, scaleUpParam.TraceID,
			scaleUpResp.Code, scaleUpResp.Message)
		return scaleUpResp.Instance, nil
	}
	if scaleUpResp.Code != statusOKCode || scaleUpResp.Worker == nil {
		log.GetLogger().Warnf("deploy response code %d, functionKey: %s, traceID: %s", scaleUpResp.Code,
			urnutils.AnonymizeTenantKey(scaleUpParam.FunctionKey), scaleUpParam.TraceID)
		if scaleUpResp.Code == 0 {
			return nil, snerror.New(statuscode.ScaleUpRequestErrCode, scaleUpResp.Message)
		}
		return nil, snerror.New(scaleUpResp.Code, scaleUpResp.Message)
	}
	log.GetLogger().Infof("deploy functionKey %s instance: %v traceID: %s successfully",
		urnutils.AnonymizeTenantKey(scaleUpParam.FunctionKey), scaleUpResp.Instance, scaleUpParam.TraceID)
	return scaleUpResp.Instance, nil
}

// ScaleDownInstance sends request to workerManager scale down instance.
func ScaleDownInstance(instanceID, functionKey, traceID string) error {
	anonymizedFuncKey := urnutils.AnonymizeTenantKey(functionKey)
	log.GetLogger().Infof("traceID: %s | FuncKeyWithRes: %s | instanceID: %s | start to scale down instance",
		traceID, anonymizedFuncKey, instanceID)
	ctx := context.TODO()
	ctx, cancel := context.WithTimeout(ctx, scaleDownTimeout)
	defer cancel()
	request := makeScaleDownRequest(ctx, instanceID, functionKey)
	if request == nil {
		log.GetLogger().Errorf("traceID: %s | FuncKeyWithRes: %s | failed to make scale down request",
			traceID, anonymizedFuncKey)
		return errors.New("failed to make scale down request")
	}
	resp, err := GetWorkerManagerClient().Do(request)
	if err != nil {
		log.GetLogger().Errorf("traceID: %s | FuncKeyWithRes: %s | error: %s | sent http request error and "+
			"failed to send scale down request to worker manager", traceID, anonymizedFuncKey, err.Error())
		return err
	}
	return handleScaleDownResponseFromWorkerManager(instanceID, functionKey, traceID, resp)
}

func makeScaleUpRequest(ctx context.Context, scaleUpParam *ScaleUpParam) *http.Request {
	tenantID, functionName, functionVersion := utils.ParseFuncKey(scaleUpParam.FunctionKey)
	baseURL, err := GetWorkerManagerBaseURL()
	if err != nil {
		log.GetLogger().Errorf("failed to get worker manager baseURL: %s", err.Error())
		return nil
	}
	requestURI := baseURL + fmt.Sprintf(workerManagerDeployURL, tenantID, functionName, functionVersion)
	param := DeployParam{
		FuncName:   functionName,
		Applier:    selfregister.SelfInstanceID,
		DeployNode: nodeID,
		Business:   "yrk",
		TenantID:   tenantID,
		Version:    functionVersion,
		OwnerIP:    nodeID,
		CPU:        scaleUpParam.CPU,
		Memory:     scaleUpParam.Memory,
		OwnedWorkerView: WorkersView{
			OwnedNum: -1,
		},
	}
	data, err := json.Marshal(param)
	if err != nil {
		log.GetLogger().Errorf("failed to marshal scaleUp request: %s", err.Error())
		return nil
	}
	request, err := http.NewRequestWithContext(ctx, "POST", requestURI, bytes.NewBuffer(data))
	if err != nil {
		log.GetLogger().Errorf("action failed when make scaleUp request, err %s", err.Error())
		return nil
	}
	FillInWorkerManagerRequestHeaders(request)
	log.GetLogger().Debugf("succeeded to sign the authorization of function: %s, traceID: %s",
		urnutils.AnonymizeTenantKey(scaleUpParam.FunctionKey), scaleUpParam.TraceID)
	request.Header.Set(constant.HeaderTraceID, scaleUpParam.TraceID)
	request.Header.Set(constant.HeaderForceDeploy, strconv.FormatBool(false))
	return request
}

func makeScaleDownRequest(ctx context.Context, instanceID, functionKey string) *http.Request {
	tenantID, functionName, functionVersion := utils.ParseFuncKey(functionKey)
	baseURL, err := GetWorkerManagerBaseURL()
	if err != nil {
		log.GetLogger().Errorf("failed to get worker manager baseURL: %s", err.Error())
		return nil
	}
	requestURI := baseURL + workerManagerDeleteURL
	param := DeleteParam{
		InstanceID:  instanceID,
		FuncName:    functionName,
		FuncVersion: functionVersion,
		BusinessID:  "yrk",
		TenantID:    tenantID,
		Applier:     selfregister.SelfInstanceID,
	}
	data, err := json.Marshal(param)
	if err != nil {
		log.GetLogger().Errorf("marshal error when make scale down request")
		return nil
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, requestURI, bytes.NewBuffer(data))
	if err != nil {
		log.GetLogger().Errorf("error when make scale down request: %s", err.Error())
		return nil
	}
	FillInWorkerManagerRequestHeaders(request)
	return request
}

func handleScaleDownResponseFromWorkerManager(instanceID, functionKey, traceID string, resp *http.Response) error {
	anonymizedFuncKey := urnutils.AnonymizeTenantKey(functionKey)
	respBody, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.GetLogger().Errorf("traceID: %s | FuncKeyWithRes: %s | error: %s | failed to get response body",
			traceID, anonymizedFuncKey, err.Error())
		return err
	}
	if resp.StatusCode != http.StatusOK {
		snErr := &snerror.BadResponse{}
		err = json.Unmarshal(respBody, snErr)
		if err != nil {
			log.GetLogger().Errorf("traceID: %s | FuncKeyWithRes: %s | error: %s | failed to unmarshal scale "+
				"down bad response", traceID, anonymizedFuncKey, err.Error())
			return err
		}
		log.GetLogger().Errorf("traceID: %s | FuncKeyWithRes: %s | responseCode: %d | responseMessage: %s | "+
			"worker manager returned bad response", traceID, anonymizedFuncKey, snErr.Code, snErr.Message)
		return snerror.New(snErr.Code, snErr.Message)
	}
	scaleDownResp := DeleteWorkerResponse{}
	if err := json.Unmarshal(respBody, &scaleDownResp); err != nil {
		log.GetLogger().Errorf("traceID: %s | FuncKeyWithRes: %s | failed to unmarshal scale down response: "+
			"DeleteWorkerResponse", traceID, anonymizedFuncKey)
		return err
	}
	if scaleDownResp.Code != http.StatusOK {
		log.GetLogger().Warnf("traceID: %s | FuncKeyWithRes: %s | responseCode: %d | responseMessage: %s | "+
			"worker manager returned error", traceID, anonymizedFuncKey, scaleDownResp.Code, scaleDownResp.Message)
		return errors.New("worker manager returned error")
	}
	log.GetLogger().Infof("traceID: %s | FuncKeyWithRes: %s | instanceID: %s | succeed to scale down instance",
		traceID, anonymizedFuncKey, instanceID)
	return nil
}

// NeedTryError no response error and wait for retrying scaling
// or local worker turned idle until request queued timeout
func NeedTryError(err error) bool {
	if snErr, ok := err.(snerror.SNError); ok {
		if snErr.Code() == statuscode.GettingPodErrorCode ||
			snErr.Code() == statuscode.CancelGeneralizePod ||
			snErr.Code() == statuscode.ReachMaxInstancesCode ||
			snErr.Code() == statuscode.ReachMaxOnDemandInstancesPerTenant {
			return true
		}
	}
	return false
}
