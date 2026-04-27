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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/job"
	"yuanrong.org/kernel/pkg/dashboard/getinfo"
)

// SubmitJobHandler -
func SubmitJobHandler(ctx *gin.Context) {
	traceID := ctx.Request.Header.Get(constant.HeaderTraceID)
	logger := log.GetLogger().With(zap.Any("traceID", traceID))
	req := job.SubmitJobHandleReq(ctx)
	if req == nil {
		return
	}
	if req.SubmissionId == "" {
		req.NewSubmissionID()
	} else {
		resp := getinfo.GetAppInfo(req.SubmissionId)
		if resp.Message != "" {
			if resp.Code != http.StatusNotFound {
				logger.Errorf("failed GetAppInfo, submissionId: %s, code: %d, err: %s",
					req.SubmissionId, errFrontend, resp.Message)
				ctx.JSON(http.StatusInternalServerError,
					fmt.Sprintf("failed GetAppInfo, submissionId: %s, code: %d, err: %s",
						req.SubmissionId, errFrontend, resp.Message))
				return
			}
		}
		if resp.Code == http.StatusOK {
			logger.Errorf("submit job has already exist, submissionId: %s, code: %d",
				req.SubmissionId, errFrontend)
			ctx.JSON(http.StatusBadRequest,
				fmt.Sprintf("submit job has already exist, submissionId: %s, code: %d",
					req.SubmissionId, errFrontend))
			return
		}
	}
	logger.Debugf("start to CreateApp, req:%#v", req)
	reqBytes, err := json.Marshal(req)
	if err != nil {
		logger.Errorf("marshal CreateApp request failed, err:%s", err.Error())
		ctx.JSON(http.StatusBadRequest, fmt.Sprintf("marshal CreateApp request failed, err:%v", err))
		return
	}
	job.SubmitJobHandleRes(ctx, getinfo.CreateApp(reqBytes))
}

// ListJobsHandler -
func ListJobsHandler(ctx *gin.Context) {
	job.ListJobsHandleRes(ctx, getinfo.ListApps())
}

// GetJobInfoHandler -
func GetJobInfoHandler(ctx *gin.Context) {
	submissionId := ctx.Param(job.PathParamSubmissionId)
	log.GetLogger().Debugf("start to GetJobInfoHandler, submissionId: %s", submissionId)
	job.GetJobInfoHandleRes(ctx, getinfo.GetAppInfo(submissionId))
}

// DeleteJobHandler -
func DeleteJobHandler(ctx *gin.Context) {
	submissionId := ctx.Param(job.PathParamSubmissionId)
	log.GetLogger().Debugf("start to DeleteJobHandler, submissionId: %s", submissionId)
	job.DeleteJobHandleRes(ctx, getinfo.DeleteApp(submissionId))
}

// StopJobHandler -
func StopJobHandler(ctx *gin.Context) {
	submissionId := ctx.Param(job.PathParamSubmissionId)
	log.GetLogger().Debugf("start to StopJobHandler, submissionId: %s", submissionId)
	job.StopJobHandleRes(ctx, getinfo.StopApp(submissionId))
}
