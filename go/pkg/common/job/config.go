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

import "yuanrong.org/kernel/pkg/common/constants"

// 处理job的对外接口
const (
	PathParamSubmissionId = "submissionId"
	PathGroupJobs         = "/api/jobs"
	PathGetJobs           = constants.DynamicRouterParamPrefix + PathParamSubmissionId
	PathDeleteJobs        = constants.DynamicRouterParamPrefix + PathParamSubmissionId
	PathStopJobs          = constants.DynamicRouterParamPrefix + PathParamSubmissionId + "/stop"
)

const (
	submissionIdPattern = "^[a-z0-9-]{1,64}$"
	jobIDPrefix         = "app-"
	tenantIdKey         = "tenantId"
)

// Response -
type Response struct {
	Code    int    `form:"code" json:"code"`
	Message string `form:"message" json:"message"`
	Data    []byte `form:"data" json:"data"`
}

// SubmitRequest is SubmitRequest struct
type SubmitRequest struct {
	Entrypoint          string             `form:"entrypoint" json:"entrypoint"`
	SubmissionId        string             `form:"submission_id" json:"submission_id"`
	RuntimeEnv          *RuntimeEnv        `form:"runtime_env" json:"runtime_env" valid:"optional"`
	Metadata            map[string]string  `form:"metadata" json:"metadata" valid:"optional"`
	Labels              string             `form:"labels" json:"labels" valid:"optional"`
	CreateOptions       map[string]string  `form:"createOptions" json:"createOptions" valid:"optional"`
	EntrypointResources map[string]float64 `form:"entrypoint_resources" json:"entrypoint_resources" valid:"optional"`
	EntrypointNumCpus   float64            `form:"entrypoint_num_cpus" json:"entrypoint_num_cpus" valid:"optional"`
	EntrypointNumGpus   float64            `form:"entrypoint_num_gpus" json:"entrypoint_num_gpus" valid:"optional"`
	EntrypointMemory    int                `form:"entrypoint_memory" json:"entrypoint_memory" valid:"optional"`
}

// RuntimeEnv args of invoking create_app
type RuntimeEnv struct {
	WorkingDir string            `form:"working_dir" json:"working_dir"  valid:"optional"`
	Pip        []string          `form:"pip" json:"pip"  valid:"optional" `
	EnvVars    map[string]string `form:"env_vars" json:"env_vars"  valid:"optional"`
}
