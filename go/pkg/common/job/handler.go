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

// Package job -
package job

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"yuanrong.org/kernel/pkg/common/constants"
	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/httputil/utils"
	"yuanrong.org/kernel/pkg/common/uuid"
)

// SubmitJobHandleReq -
func SubmitJobHandleReq(ctx *gin.Context) *SubmitRequest {
	traceID := ctx.Request.Header.Get(constant.HeaderTraceID)
	logger := log.GetLogger().With(zap.Any("traceID", traceID))
	var req SubmitRequest
	if err := ctx.ShouldBind(&req); err != nil {
		logger.Errorf("shouldBind SubmitJob request failed, err: %s", err)
		ctx.JSON(http.StatusBadRequest, fmt.Sprintf("shouldBind SubmitJob request failed, err: %v", err))
		return nil
	}
	err := req.CheckField()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return nil
	}
	req.EntrypointNumCpus = math.Ceil(req.EntrypointNumCpus * constants.CpuUnitConvert)
	req.EntrypointMemory =
		int(math.Ceil(float64(req.EntrypointMemory) / constants.MemoryUnitConvert / constants.MemoryUnitConvert))
	reqHeader := utils.ParseHeader(ctx)
	if tenantId, ok := reqHeader[constants.HeaderTenantId]; ok {
		req.AddCreateOptions(tenantIdKey, tenantId)
	}
	if labels, ok := reqHeader[constants.HeaderPoolLabel]; ok {
		req.Labels = labels
	}
	logger.Debugf("SubmitJob createApp start, req:%#v", req)
	return &req
}

// SubmitJobHandleRes -
// SubmitJob godoc
// @Summary      submit job
// @Description  submit a new job
// @Accept       json
// @Produce      json
// @Router       /api/jobs [POST]
// @Param        SubmitRequest body SubmitRequest true "提交job时定义的job信息。"
// @Success      200  {object}  map[string]string   "提交job成功，返回该job的submission_id"
// @Failure      400  {string}  string "用户请求错误，包含错误信息"
// @Failure      404  {string}  string "该job已经存在"
// @Failure      500  {string}  string "服务器处理错误，包含错误信息"
func SubmitJobHandleRes(ctx *gin.Context, resp Response) {
	if resp.Code != http.StatusOK || resp.Message != "" {
		ctx.JSON(resp.Code, resp.Message)
		return
	}
	var result map[string]string
	err := json.Unmarshal(resp.Data, &result)
	if err != nil {
		ctx.JSON(http.StatusBadRequest,
			fmt.Sprintf("unmarshal response data failed, data: %v", resp.Data))
		return
	}
	ctx.JSON(http.StatusOK, result)
	log.GetLogger().Debugf("SubmitJobHandleRes succeed, submission_id: %s", result)
}

// ListJobsHandleRes -
// ListJobs godoc
// @Summary      List Jobs
// @Description  list jobs with jobInfo
// @Accept       json
// @Produce      json
// @Router       /api/jobs [GET]
// @Success      200  {array}   constant.AppInfo "返回所有jobs的信息"
// @Failure      500  {string}  string "服务器处理错误，包含错误信息"
func ListJobsHandleRes(ctx *gin.Context, resp Response) {
	traceID := ctx.Request.Header.Get(constant.HeaderTraceID)
	logger := log.GetLogger().With(zap.Any("traceID", traceID))
	if resp.Code != http.StatusOK || resp.Message != "" {
		ctx.JSON(resp.Code, resp.Message)
		return
	}
	var result []*constant.AppInfo
	err := json.Unmarshal(resp.Data, &result)
	if err != nil {
		ctx.JSON(http.StatusBadRequest,
			fmt.Sprintf("unmarshal response data failed, data: %v", resp.Data))
		return
	}
	ctx.JSON(http.StatusOK, result)
	logger.Debugf("ListJobsHandleRes succeed")
}

// GetJobInfoHandleRes -
// GetJobInfo godoc
// @Summary      Get JobInfo
// @Description  get jobInfo by submission_id
// @Accept       json
// @Produce      json
// @Router       /api/jobs/{submissionId} [GET]
// @Param        submissionId	path string true "job的submission_id，以'app-'开头"
// @Success      200  {object}  constant.AppInfo "返回submission_id对应的job信息"
// @Failure      404  {string}  string "该job不存在"
// @Failure      500  {string}  string "服务器处理错误，包含错误信息"
func GetJobInfoHandleRes(ctx *gin.Context, resp Response) {
	submissionId := ctx.Param(PathParamSubmissionId)
	logger := log.GetLogger().With(zap.Any("SubmissionId", submissionId))
	if resp.Code != http.StatusOK || resp.Message != "" {
		ctx.JSON(resp.Code, resp.Message)
		return
	}
	var result *constant.AppInfo
	err := json.Unmarshal(resp.Data, &result)
	if err != nil {
		ctx.JSON(http.StatusBadRequest,
			fmt.Sprintf("unmarshal response data failed, data: %v", resp.Data))
		return
	}
	ctx.JSON(http.StatusOK, result)
	logger.Debugf("GetJobInfoHandleRes succeed")
}

// DeleteJobHandleRes -
// DeleteJob godoc
// @Summary      Delete Job
// @Description  delete job by submission_id
// @Accept       json
// @Produce      json
// @Router       /api/jobs/{submissionId} [DELETE]
// @Param        submissionId	path string true "job的submission_id，以'app-'开头"
// @Success      200  {boolean} bool "返回true则说明可以删除对应的job，返回false则说明无法删除job"
// @Failure      403  {string}  string "禁止删除job，包含错误信息和job运行状态"
// @Failure      404  {string}  string "该job不存在"
// @Failure      500  {string}  string "服务器处理错误，包含错误信息"
func DeleteJobHandleRes(ctx *gin.Context, resp Response) {
	submissionId := ctx.Param(PathParamSubmissionId)
	logger := log.GetLogger().With(zap.Any("SubmissionId", submissionId))
	if resp.Code == http.StatusForbidden {
		log.GetLogger().Errorf("forbidden to delete, status: %s", resp.Data)
		ctx.JSON(http.StatusOK, false)
		return
	}
	if resp.Code != http.StatusOK || resp.Message != "" {
		ctx.JSON(resp.Code, resp.Message)
		return
	}
	ctx.JSON(http.StatusOK, true)
	logger.Debugf("DeleteJobHandleRes succeed")
}

// StopJobHandleRes -
// StopJob godoc
// @Summary      Stop Job
// @Description  stop job by submission_id
// @Accept       json
// @Produce      json
// @Router       /api/jobs/{submissionId}/stop [POST]
// @Param        submissionId	path string true "job的submission_id，以'app-'开头"
// @Success      200  {boolean} bool   "返回true表示可以停止运行对应的job，返回false表示job当前状态不能被停止"
// @Failure      403  {string}  string "禁止删除job，包含错误信息和job运行状态"
// @Failure      404  {string}  string "该job不存在"
// @Failure      500  {string}  string "服务器处理错误，包含错误信息"
func StopJobHandleRes(ctx *gin.Context, resp Response) {
	submissionId := ctx.Param(PathParamSubmissionId)
	logger := log.GetLogger().With(zap.Any("SubmissionId", submissionId))
	if resp.Code == http.StatusForbidden {
		log.GetLogger().Errorf("forbidden to stop job, status: %s", resp.Data)
		ctx.JSON(http.StatusOK, false)
		return
	}
	if resp.Code != http.StatusOK || resp.Message != "" {
		ctx.JSON(resp.Code, resp.Message)
		return
	}
	ctx.JSON(http.StatusOK, true)
	logger.Debugf("StopJobHandleRes succeed")
}

// CheckField -
func (req *SubmitRequest) CheckField() error {
	if req.Entrypoint == "" {
		log.GetLogger().Errorf("entrypoint should not be empty")
		return fmt.Errorf("entrypoint should not be empty")
	}
	if req.RuntimeEnv == nil || req.RuntimeEnv.WorkingDir == "" {
		log.GetLogger().Errorf("runtime_env.working_dir should not be empty")
		return fmt.Errorf("runtime_env.working_dir should not be empty")
	}
	if err := req.ValidateResources(); err != nil {
		log.GetLogger().Errorf("validateResources error: %s", err.Error())
		return err
	}
	if err := req.CheckSubmissionId(); err != nil {
		log.GetLogger().Errorf("chechk submission_id: %s, error: %s", req.SubmissionId, err.Error())
		return err
	}
	return nil
}

// ValidateResources -
func (req *SubmitRequest) ValidateResources() error {
	if req.EntrypointNumCpus < 0 {
		return errors.New("entrypoint_num_cpus should not be less than 0")
	}
	if req.EntrypointNumGpus < 0 {
		return errors.New("entrypoint_num_gpus should not be less than 0")
	}
	if req.EntrypointMemory < 0 {
		return errors.New("entrypoint_memory should not be less than 0")
	}
	return nil
}

// CheckSubmissionId -
func (req *SubmitRequest) CheckSubmissionId() error {
	if req.SubmissionId == "" {
		return nil
	}
	if strings.Contains(req.SubmissionId, "driver") {
		return errors.New("submission_id should not contain 'driver'")
	}
	if !strings.HasPrefix(req.SubmissionId, jobIDPrefix) {
		req.SubmissionId = jobIDPrefix + req.SubmissionId
	}
	isMatch, err := regexp.MatchString(submissionIdPattern, req.SubmissionId)
	if err != nil || !isMatch {
		return fmt.Errorf("regular expression validation error, submissionId: %s, pattern: %s, err: %v",
			req.SubmissionId, submissionIdPattern, err)
	}
	return nil
}

// NewSubmissionID -
func (req *SubmitRequest) NewSubmissionID() {
	if req.SubmissionId == "" {
		req.SubmissionId = jobIDPrefix + uuid.New().String()
	}
}

// AddCreateOptions -
func (req *SubmitRequest) AddCreateOptions(key, value string) {
	if req.CreateOptions == nil {
		req.CreateOptions = map[string]string{}
	}
	if key != "" {
		req.CreateOptions[key] = value
	}
}

// BuildJobResponse -
func BuildJobResponse(data any, code int, err error) Response {
	dataBytes, jsonErr := json.Marshal(data)
	if jsonErr != nil {
		return Response{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("marshal job response failed, err: %v", jsonErr),
		}
	}
	var resp Response
	resp.Code = code
	if data != nil {
		resp.Data = dataBytes
	}
	if err != nil {
		resp.Message = err.Error()
	}
	return resp
}
