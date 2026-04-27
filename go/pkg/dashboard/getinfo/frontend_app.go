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

// Package getinfo for get/post frontend
package getinfo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/job"
)

const (
	creatAppPath   = "/posix/instance/create"
	deleteAppPath  = "/delete/"
	getAppInfoPath = "/getappinfo/"
	listAppsPath   = "/list"
	stopAppPath    = "/posix/kill/"
)

// CreateApp function for create app
func CreateApp(reqBytes []byte) job.Response {
	return reqForJob("POST", creatAppPath, bytes.NewBuffer(reqBytes))
}

// DeleteApp function for delete app
func DeleteApp(submissionID string) job.Response {
	return reqForJob("DELETE", deleteAppPath+submissionID, nil)
}

// GetAppInfo function for get app info by submissionID
func GetAppInfo(submissionID string) job.Response {
	return reqForJob("GET", getAppInfoPath+submissionID, nil)
}

// ListApps function for list apps info
func ListApps() job.Response {
	return reqForJob("GET", listAppsPath, nil)
}

// StopApp function for stop app
func StopApp(submissionID string) job.Response {
	return reqForJob("POST", stopAppPath+submissionID, nil)
}

func reqForJob(method string, path string, reqBody io.Reader) job.Response {
	var result job.Response
	respBytes, err := requestFrontend(method, path, reqBody)
	if err != nil {
		log.GetLogger().Errorf("request to fronted failed, err: %v", err)
		return job.BuildJobResponse(nil, http.StatusBadRequest,
			fmt.Errorf("request to fronted failed, err: %v", err))
	}
	err = json.Unmarshal(respBytes, &result)
	if err != nil {
		log.GetLogger().Errorf("unmarshal response from fronted failed, err: %v", err)
		return job.BuildJobResponse(nil, http.StatusBadRequest,
			fmt.Errorf("unmarshal response from fronted failed, err: %v", err))
	}
	return result
}
