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
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/gin-gonic/gin"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/constants"
	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/job"
	"yuanrong.org/kernel/pkg/dashboard/flags"
	"yuanrong.org/kernel/pkg/dashboard/getinfo"
)

type TC struct {
	FrontendResStatus  int
	DashboardResStatus int
	FrontendResBody    string
	DashboardResBody   string
}

func addApps(submissionId string) []*constant.AppInfo {
	var result []*constant.AppInfo
	return append(result, buildAppInfo(submissionId))
}

func buildAppInfo(submissionId string) *constant.AppInfo {
	return &constant.AppInfo{
		Key:          submissionId,
		Type:         "SUBMISSION",
		SubmissionID: submissionId,
		RuntimeEnv: map[string]interface{}{
			"working_dir": "",
			"pip":         "",
			"envVars":     "",
		},
		DriverInfo: constant.DriverInfo{
			ID: submissionId,
		},
		Status: "RUNNING",
	}
}

func TestSubmitJobHandler(t *testing.T) {
	convey.Convey("test SubmitJobHandler", t, func() {
		submissionId := "app-frontend-job-submit1"
		gin.SetMode(gin.TestMode)
		rw := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rw)
		bodyBytes, _ := json.Marshal(job.SubmitRequest{
			Entrypoint:   "sleep 200",
			SubmissionId: submissionId,
			RuntimeEnv: &job.RuntimeEnv{
				WorkingDir: "file:///usr1/deploy/file.zip",
				Pip:        []string{"numpy==1.24", "scipy==1.25"},
				EnvVars: map[string]string{
					"SOURCE_REGION": "suzhou_std",
					"DEPLOY_REGION": "suzhou_std",
				},
			},
			Metadata: map[string]string{
				"autoscenes_ids": "auto_1-test",
				"task_type":      "task_1",
				"ttl":            "1250",
			},
			EntrypointResources: map[string]float64{
				"NPU": 0,
			},
			EntrypointNumCpus: 300,
			EntrypointNumGpus: 0,
			EntrypointMemory:  0,
		})
		reader := bytes.NewBuffer(bodyBytes)
		c.Request = &http.Request{
			Method: http.MethodPost,
			URL:    &url.URL{Path: job.PathGroupJobs},
			Header: http.Header{
				"Content-Type":            []string{"application/json"},
				constants.HeaderTenantID:  []string{"123456"},
				constants.HeaderPoolLabel: []string{"abc"},
			},
			Body: io.NopCloser(reader), // 使用 io.NopCloser 包装 reader，使其满足 io.ReadCloser 接口
		}
		convey.Convey("when job is exist", func() {
			defer gomonkey.ApplyFunc(getinfo.GetAppInfo, func(submissionID string) job.Response {
				return job.Response{
					Data:    nil,
					Code:    http.StatusOK,
					Message: "",
				}
			}).Reset()
			SubmitJobHandler(c)
			convey.So(rw.Code, convey.ShouldEqual, http.StatusBadRequest)
			convey.So(rw.Body.String(), convey.ShouldStartWith, "\"submit job has already exist, submissionId")
		})
		convey.Convey("when get job failed", func() {
			defer gomonkey.ApplyFunc(getinfo.GetAppInfo, func(string) job.Response {
				return job.Response{
					Data:    nil,
					Code:    http.StatusInternalServerError,
					Message: "get job failed",
				}
			}).Reset()
			SubmitJobHandler(c)
			convey.So(rw.Code, convey.ShouldEqual, http.StatusInternalServerError)
			convey.So(rw.Body.String(), convey.ShouldStartWith, "\"failed GetAppInfo, submissionId:")
		})
		convey.Convey("when not found job", func() {
			defer gomonkey.ApplyFunc(getinfo.GetAppInfo, func(string) job.Response {
				return job.Response{
					Data:    nil,
					Code:    http.StatusNotFound,
					Message: "not found job",
				}
			}).Reset()
			SubmitJobHandler(c)
			convey.So(rw.Code, convey.ShouldEqual, http.StatusBadRequest)
			convey.So(rw.Body.String(), convey.ShouldStartWith, "\"request to fronted failed, err:")
		})

		convey.Convey("when create app response status: "+strconv.Itoa(http.StatusOK), func() {
			defer gomonkey.ApplyFunc(getinfo.GetAppInfo, func(string) job.Response {
				return job.Response{
					Data:    nil,
					Code:    http.StatusNotFound,
					Message: "not found job",
				}
			}).Reset()
			defer gomonkey.ApplyFunc(getinfo.CreateApp, func([]byte) job.Response {
				return job.Response{
					Code:    http.StatusOK,
					Message: "",
					Data:    []byte(`{"submission_id":"app-123"}`),
				}
			}).Reset()
			SubmitJobHandler(c)
			convey.So(rw.Code, convey.ShouldEqual, http.StatusOK)
			convey.So(rw.Body.String(), convey.ShouldStartWith, `{"submission_id":"app-123"}`)
		})
		convey.Convey("when create app response status: "+strconv.Itoa(http.StatusInternalServerError), func() {
			defer gomonkey.ApplyFunc(getinfo.GetAppInfo, func(string) job.Response {
				return job.Response{
					Data:    nil,
					Code:    http.StatusNotFound,
					Message: "not found job",
				}
			}).Reset()
			defer gomonkey.ApplyFunc(getinfo.CreateApp, func([]byte) job.Response {
				return job.Response{
					Code:    http.StatusInternalServerError,
					Message: "failed submit job",
				}
			}).Reset()
			SubmitJobHandler(c)
			convey.So(rw.Code, convey.ShouldEqual, http.StatusInternalServerError)
			convey.So(rw.Body.String(), convey.ShouldStartWith, "\"failed submit job")
		})
	})
}

func TestListJobsHandler(t *testing.T) {
	convey.Convey("Test ListJobsHandler:", t, func() {
		dataBytes, err := json.Marshal(addApps("app-123"))
		convey.So(err, convey.ShouldBeNil)
		appInfoByte, err := json.Marshal(job.Response{
			Code: http.StatusOK,
			Data: dataBytes,
		})
		tt := []TC{
			{http.StatusOK, http.StatusOK, string(appInfoByte), string(dataBytes)},
			{http.StatusInternalServerError, http.StatusInternalServerError, `{"code": 500, "message": "failed get job"}`, `"failed get job"`},
		}
		for _, tc := range tt {
			convey.Convey("when frontend return status: "+strconv.Itoa(tc.FrontendResStatus), func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.FrontendResStatus)
					w.Write([]byte(tc.FrontendResBody))
				}))
				defer server.Close()
				flags.DashboardConfig.FrontendAddr = server.URL + "/app/v1"

				r := gin.Default()
				r.GET("/jobs", ListJobsHandler)
				req, err := http.NewRequest("GET", "/jobs", nil)
				convey.So(err, convey.ShouldBeNil)

				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				convey.So(w.Code, convey.ShouldEqual, tc.DashboardResStatus)
				convey.So(string(w.Body.Bytes()), convey.ShouldEqual, tc.DashboardResBody)
			})
		}
		convey.Convey("when request frontend failed "+strconv.Itoa(http.StatusInternalServerError), func() {
			r := gin.Default()
			r.GET("/jobs", ListJobsHandler)
			req, err := http.NewRequest("GET", "/jobs", nil)
			convey.So(err, convey.ShouldBeNil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			convey.So(w.Code, convey.ShouldEqual, http.StatusBadRequest)
			convey.So(w.Body.String(), convey.ShouldStartWith, "\"request to fronted failed")
		})
	})
}

func TestGetJobInfoHandler(t *testing.T) {
	convey.Convey("Test GetJobInfoHandler:", t, func() {
		dataBytes, err := json.Marshal(buildAppInfo("app-123"))
		convey.So(err, convey.ShouldBeNil)
		appInfoByte, err := json.Marshal(job.Response{
			Code: http.StatusOK,
			Data: dataBytes,
		})
		convey.So(err, convey.ShouldBeNil)
		tt := []TC{
			{http.StatusOK, http.StatusOK, string(appInfoByte), string(dataBytes)},
			{http.StatusNotFound, http.StatusNotFound, `{"code": 404, "message": "the job does not exist"}`, `"the job does not exist"`},
			{http.StatusInternalServerError, http.StatusInternalServerError, `{"code": 500, "message": "failed get job"}`, `"failed get job"`},
		}
		for _, tc := range tt {
			convey.Convey("when frontend return status: "+strconv.Itoa(tc.FrontendResStatus), func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.FrontendResStatus)
					w.Write([]byte(tc.FrontendResBody))
				}))
				defer server.Close()
				flags.DashboardConfig.FrontendAddr = server.URL + "/app/v1"

				r := gin.Default()
				r.GET("/jobs/:submission_id", GetJobInfoHandler)
				req, err := http.NewRequest("GET", "/jobs/123", nil)
				convey.So(err, convey.ShouldBeNil)

				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				convey.So(w.Code, convey.ShouldEqual, tc.DashboardResStatus)
				convey.So(w.Body.String(), convey.ShouldEqual, tc.DashboardResBody)
			})
		}
		convey.Convey("when request frontend failed ", func() {
			r := gin.Default()
			r.GET("/jobs/:submission_id", GetJobInfoHandler)
			req, err := http.NewRequest("GET", "/jobs/123", nil)
			convey.So(err, convey.ShouldBeNil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			convey.So(w.Code, convey.ShouldEqual, http.StatusBadRequest)
			convey.So(w.Body.String(), convey.ShouldStartWith, "\"request to fronted failed")
		})
	})
}

func TestDeleteJobHandler(t *testing.T) {
	convey.Convey("Test DeleteJobHandler:", t, func() {
		tt := []TC{
			{http.StatusOK, http.StatusOK, `{"code": 200}`, "true"},
			{http.StatusForbidden, http.StatusOK, `{"code": 403}`, "false"},
			{http.StatusNotFound, http.StatusNotFound, `{"code": 404, "message": "the job does not exist"}`, `"the job does not exist"`},
			{http.StatusInternalServerError, http.StatusInternalServerError, `{"code": 500, "message": "failed delete job"}`, `"failed delete job"`},
		}
		for _, tc := range tt {
			convey.Convey("when frontend return status: "+strconv.Itoa(tc.FrontendResStatus), func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.FrontendResStatus)
					w.Write([]byte(tc.FrontendResBody))
				}))
				defer server.Close()
				flags.DashboardConfig.FrontendAddr = server.URL + "/app/v1"

				r := gin.Default()
				r.DELETE("/jobs/:submission_id", DeleteJobHandler)
				req, err := http.NewRequest("DELETE", "/jobs/123", nil)
				convey.So(err, convey.ShouldBeNil)

				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				convey.So(w.Code, convey.ShouldEqual, tc.DashboardResStatus)
				convey.So(string(w.Body.Bytes()), convey.ShouldEqual, tc.DashboardResBody)
			})
		}
		convey.Convey("when request frontend failed", func() {
			r := gin.Default()
			r.DELETE("/jobs/:submission_id", DeleteJobHandler)
			req, err := http.NewRequest("DELETE", "/jobs/123", nil)
			convey.So(err, convey.ShouldBeNil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			convey.So(w.Code, convey.ShouldEqual, http.StatusBadRequest)
			convey.So(w.Body.String(), convey.ShouldStartWith, "\"request to fronted failed")
		})
	})
}

func TestStopJobHandler(t *testing.T) {
	convey.Convey("Test StopJobHandler", t, func() {
		tt := []TC{
			{http.StatusOK, http.StatusOK, `{"code": 200}`, "true"},
			{http.StatusForbidden, http.StatusOK, `{"code": 403}`, "false"},
			{http.StatusNotFound, http.StatusNotFound, `{"code": 404, "message": "the job does not exist"}`, `"the job does not exist"`},
			{http.StatusInternalServerError, http.StatusInternalServerError, `{"code": 500, "message": "failed stop job"}`, `"failed stop job"`},
		}
		for _, tc := range tt {
			convey.Convey("when frontend return status: "+strconv.Itoa(tc.FrontendResStatus), func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.FrontendResStatus)
					w.Write([]byte(tc.FrontendResBody))
				}))
				defer server.Close()
				flags.DashboardConfig.FrontendAddr = server.URL + "/app/v1"

				r := gin.Default()
				r.POST("/jobs/:submission_id/stop", StopJobHandler)
				req, err := http.NewRequest("POST", "/jobs/123/stop", nil)
				convey.So(err, convey.ShouldBeNil)

				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				convey.So(w.Code, convey.ShouldEqual, tc.DashboardResStatus)
				convey.So(string(w.Body.Bytes()), convey.ShouldEqual, tc.DashboardResBody)
			})
		}
		convey.Convey("when request frontend failed", func() {
			r := gin.Default()
			r.POST("/jobs/:submission_id/stop", StopJobHandler)
			req, err := http.NewRequest("POST", "/jobs/123/stop", nil)
			convey.So(err, convey.ShouldBeNil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			convey.So(w.Code, convey.ShouldEqual, http.StatusBadRequest)
			convey.So(w.Body.String(), convey.ShouldStartWith, "\"request to fronted failed")
		})
	})
}
