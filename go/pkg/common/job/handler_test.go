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

package job

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
)

func TestSubmitJobHandleReq(t *testing.T) {
	convey.Convey("test DeleteJobHandleRes", t, func() {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		bodyBytes, _ := json.Marshal(SubmitRequest{
			Entrypoint:   "",
			SubmissionId: "",
			RuntimeEnv: &RuntimeEnv{
				WorkingDir: "",
				Pip:        []string{""},
				EnvVars:    map[string]string{},
			},
			Metadata:            map[string]string{},
			EntrypointResources: map[string]float64{},
			EntrypointNumCpus:   0,
			EntrypointNumGpus:   0,
			EntrypointMemory:    0,
		})
		reader := bytes.NewBuffer(bodyBytes)
		c.Request = &http.Request{
			Method: "POST",
			URL:    &url.URL{Path: PathGroupJobs},
			Header: http.Header{
				"Content-Type":            []string{"application/json"},
				constants.HeaderTenantID:  []string{"123456"},
				constants.HeaderPoolLabel: []string{"abc"},
			},
			Body: io.NopCloser(reader), // 使用 io.NopCloser 包装 reader，使其满足 io.ReadCloser 接口
		}
		convey.Convey("when process success", func() {
			defer gomonkey.ApplyMethodFunc(&SubmitRequest{}, "CheckField", func() error {
				return nil
			}).Reset()
			result := SubmitJobHandleReq(c)
			convey.So(result.Labels, convey.ShouldEqual, "abc")
		})
		convey.Convey("when CheckField failed", func() {
			defer gomonkey.ApplyMethodFunc(&SubmitRequest{}, "CheckField", func() error {
				return errors.New("failed CheckField")
			}).Reset()
			result := SubmitJobHandleReq(c)
			convey.So(result, convey.ShouldBeNil)
			convey.So(w.Code, convey.ShouldEqual, http.StatusBadRequest)
			convey.So(w.Body.String(), convey.ShouldEqual, "\"failed CheckField\"")
		})
	})
}

func TestSubmitJobHandleRes(t *testing.T) {
	convey.Convey("test SubmitJobHandleRes", t, func() {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		resp := Response{
			Code:    http.StatusOK,
			Message: "",
			Data:    []byte("app-123"),
		}
		convey.Convey("when statusCode is "+strconv.Itoa(http.StatusNotFound), func() {
			resp.Code = http.StatusNotFound
			resp.Message = fmt.Sprintf("not found job")
			SubmitJobHandleRes(c, resp)
			convey.So(w.Code, convey.ShouldEqual, http.StatusNotFound)
			convey.So(w.Body.String(), convey.ShouldEqual, "\"not found job\"")
		})
		convey.Convey("when statusCode is "+strconv.Itoa(http.StatusInternalServerError), func() {
			resp.Code = http.StatusInternalServerError
			resp.Message = fmt.Sprintf("failed get job")
			SubmitJobHandleRes(c, resp)
			convey.So(w.Code, convey.ShouldEqual, http.StatusInternalServerError)
			convey.So(w.Body.String(), convey.ShouldEqual, "\"failed get job\"")
		})
		convey.Convey("when response data is nil", func() {
			resp.Data = nil
			SubmitJobHandleRes(c, resp)
			convey.So(w.Code, convey.ShouldEqual, http.StatusBadRequest)
			convey.So(w.Body.String(), convey.ShouldStartWith, "\"unmarshal response data failed, data:")
		})
		convey.Convey("when process success", func() {
			marshal, err := json.Marshal(map[string]string{
				"submission_id": "app-123",
			})
			resp.Data = marshal
			convey.So(err, convey.ShouldBeNil)
			SubmitJobHandleRes(c, resp)
			expectedResult, err := json.Marshal(map[string]string{
				"submission_id": "app-123",
			})
			if err != nil {
				t.Errorf("marshal expected result failed, err: %v", err)
			}
			convey.So(w.Code, convey.ShouldEqual, http.StatusOK)
			convey.So(w.Body.String(), convey.ShouldResemble, string(expectedResult))
		})
	})
}

func TestListJobsHandleRes(t *testing.T) {
	convey.Convey("test ListJobsHandleRes", t, func() {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
		}
		dataBytes, err := json.Marshal([]*constant.AppInfo{
			{
				Type:         "SUBMISSION",
				Entrypoint:   "python script.py",
				SubmissionID: "app-123",
			},
		})
		if err != nil {
			t.Errorf("marshal expected result failed, err: %v", err)
		}
		resp := Response{
			Code:    http.StatusOK,
			Message: "",
			Data:    dataBytes,
		}
		convey.Convey("when statusCode is "+strconv.Itoa(http.StatusInternalServerError), func() {
			resp.Code = http.StatusInternalServerError
			resp.Message = fmt.Sprintf("failed get job")
			ListJobsHandleRes(c, resp)
			convey.So(w.Code, convey.ShouldEqual, http.StatusInternalServerError)
			convey.So(w.Body.String(), convey.ShouldEqual, "\"failed get job\"")
		})
		convey.Convey("when unmarshal response data failed", func() {
			resp.Data = []byte(",aa,")
			ListJobsHandleRes(c, resp)
			convey.So(w.Code, convey.ShouldEqual, http.StatusBadRequest)
			convey.So(w.Body.String(), convey.ShouldStartWith, "\"unmarshal response data failed")
		})
		convey.Convey("when response data is nil", func() {
			resp.Data = []byte("[]")
			ListJobsHandleRes(c, resp)
			convey.So(w.Code, convey.ShouldEqual, http.StatusOK)
			convey.So(w.Body.String(), convey.ShouldEqual, "[]")
		})
		convey.Convey("when process success", func() {
			ListJobsHandleRes(c, resp)
			expectedResult, err := json.Marshal([]*constant.AppInfo{
				{
					Type:         "SUBMISSION",
					Entrypoint:   "python script.py",
					SubmissionID: "app-123",
				},
			})
			if err != nil {
				t.Errorf("marshal expected result failed, err: %v", err)
			}
			convey.So(w.Code, convey.ShouldEqual, http.StatusOK)
			convey.So(w.Body.String(), convey.ShouldResemble, string(expectedResult))
		})
	})
}

func TestGetJobInfoHandleRes(t *testing.T) {
	convey.Convey("test GetJobInfoHandleRes", t, func() {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		dataBytes, err := json.Marshal(&constant.AppInfo{
			Type:         "SUBMISSION",
			Entrypoint:   "python script.py",
			SubmissionID: "app-123",
		})
		if err != nil {
			t.Errorf("marshal expected result failed, err: %v", err)
		}
		resp := Response{
			Code:    http.StatusOK,
			Message: "",
			Data:    dataBytes,
		}
		convey.Convey("when statusCode is "+strconv.Itoa(http.StatusNotFound), func() {
			resp.Code = http.StatusNotFound
			resp.Message = fmt.Sprintf("not found job")
			GetJobInfoHandleRes(c, resp)
			convey.So(w.Code, convey.ShouldEqual, http.StatusNotFound)
			convey.So(w.Body.String(), convey.ShouldEqual, "\"not found job\"")
		})
		convey.Convey("when statusCode is "+strconv.Itoa(http.StatusInternalServerError), func() {
			resp.Code = http.StatusInternalServerError
			resp.Message = fmt.Sprintf("failed get job")
			GetJobInfoHandleRes(c, resp)
			convey.So(w.Code, convey.ShouldEqual, http.StatusInternalServerError)
			convey.So(w.Body.String(), convey.ShouldEqual, "\"failed get job\"")
		})
		convey.Convey("when unmarshal response data failed", func() {
			resp.Data = []byte(",aa,")
			GetJobInfoHandleRes(c, resp)
			convey.So(w.Code, convey.ShouldEqual, http.StatusBadRequest)
			convey.So(w.Body.String(), convey.ShouldStartWith, "\"unmarshal response data failed")
		})
		convey.Convey("when response data is nil", func() {
			resp.Data = nil
			GetJobInfoHandleRes(c, resp)
			convey.So(w.Code, convey.ShouldEqual, http.StatusBadRequest)
			convey.So(w.Body.String(), convey.ShouldStartWith, "\"unmarshal response data failed")
		})
		convey.Convey("when process success", func() {
			GetJobInfoHandleRes(c, resp)
			expectedResult, err := json.Marshal(&constant.AppInfo{
				Type:         "SUBMISSION",
				Entrypoint:   "python script.py",
				SubmissionID: "app-123",
			})
			if err != nil {
				t.Errorf("marshal expected result failed, err: %v", err)
			}
			convey.So(w.Code, convey.ShouldEqual, http.StatusOK)
			convey.So(w.Body.String(), convey.ShouldResemble, string(expectedResult))
		})
	})
}

func TestDeleteJobHandleRes(t *testing.T) {
	convey.Convey("test DeleteJobHandleRes", t, func() {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		resp := Response{
			Code:    http.StatusOK,
			Message: "",
			Data:    []byte("SUCCEEDED"),
		}
		convey.Convey("when statusCode is "+strconv.Itoa(http.StatusForbidden), func() {
			resp.Code = http.StatusForbidden
			DeleteJobHandleRes(c, resp)
			convey.So(w.Code, convey.ShouldEqual, http.StatusOK)
			convey.So(w.Body.String(), convey.ShouldEqual, "false")
		})
		convey.Convey("when statusCode is "+strconv.Itoa(http.StatusBadRequest), func() {
			resp.Code = http.StatusBadRequest
			resp.Message = fmt.Sprintf("failed delete job")
			DeleteJobHandleRes(c, resp)
			convey.So(w.Code, convey.ShouldEqual, http.StatusBadRequest)
			convey.So(w.Body.String(), convey.ShouldEqual, "\"failed delete job\"")
		})
		convey.Convey("when response data is nil", func() {
			resp.Data = nil
			DeleteJobHandleRes(c, resp)
			convey.So(w.Code, convey.ShouldEqual, http.StatusOK)
			convey.So(w.Body.String(), convey.ShouldEqual, "true")
		})
		convey.Convey("when process success", func() {
			DeleteJobHandleRes(c, resp)
			convey.So(w.Code, convey.ShouldEqual, http.StatusOK)
			convey.So(w.Body.String(), convey.ShouldEqual, "true")
		})
	})
}

func TestStopJobHandleRes(t *testing.T) {
	convey.Convey("test StopJobHandleRes", t, func() {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		resp := Response{
			Code:    http.StatusOK,
			Message: "",
			Data:    []byte(`SUCCEEDED`),
		}
		convey.Convey("when statusCode is "+strconv.Itoa(http.StatusForbidden), func() {
			resp.Code = http.StatusForbidden
			StopJobHandleRes(c, resp)
			convey.So(w.Code, convey.ShouldEqual, http.StatusOK)
			convey.So(w.Body.String(), convey.ShouldEqual, "false")
		})
		convey.Convey("when statusCode is "+strconv.Itoa(http.StatusBadRequest), func() {
			resp.Code = http.StatusBadRequest
			resp.Message = fmt.Sprintf("failed stop job")
			StopJobHandleRes(c, resp)
			convey.So(w.Code, convey.ShouldEqual, http.StatusBadRequest)
			convey.So(w.Body.String(), convey.ShouldEqual, "\"failed stop job\"")
		})
		convey.Convey("when response data is nil", func() {
			resp.Data = nil
			StopJobHandleRes(c, resp)
			convey.So(w.Code, convey.ShouldEqual, http.StatusOK)
			convey.So(w.Body.String(), convey.ShouldEqual, "true")
		})
		convey.Convey("when process success", func() {
			StopJobHandleRes(c, resp)
			convey.So(w.Code, convey.ShouldEqual, http.StatusOK)
			convey.So(w.Body.String(), convey.ShouldEqual, "true")
		})
	})
}

func TestSubmitRequest_CheckField(t *testing.T) {
	convey.Convey("test (req *SubmitRequest) CheckField", t, func() {
		req := &SubmitRequest{
			Entrypoint:   "python script.py",
			SubmissionId: "",
			RuntimeEnv: &RuntimeEnv{
				WorkingDir: "file:///home/disk/tk/file.zip",
				Pip:        []string{"numpy==1.24", "scipy==1.11.0"},
				EnvVars: map[string]string{
					"SOURCE_REGION": "suzhou_std",
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
			EntrypointNumCpus: 0,
			EntrypointNumGpus: 0,
			EntrypointMemory:  0,
		}
		convey.Convey("when req.Entrypoint is empty", func() {
			req.Entrypoint = ""
			err := req.CheckField()
			convey.So(err, convey.ShouldBeError, errors.New("entrypoint should not be empty"))
		})
		convey.Convey("when req.RuntimeEnv is empty", func() {
			req.RuntimeEnv = nil
			err := req.CheckField()
			convey.So(err, convey.ShouldBeError, errors.New("runtime_env.working_dir should not be empty"))
		})
		convey.Convey("when req.RuntimeEnv.WorkingDir is empty", func() {
			req.RuntimeEnv.WorkingDir = ""
			err := req.CheckField()
			convey.So(err, convey.ShouldBeError, errors.New("runtime_env.working_dir should not be empty"))
		})
		convey.Convey("when ValidateResources failed", func() {
			defer gomonkey.ApplyMethodFunc(&SubmitRequest{}, "ValidateResources", func() error {
				return errors.New("failed ValidateResources")
			}).Reset()
			defer gomonkey.ApplyMethodFunc(&SubmitRequest{}, "CheckSubmissionId", func() error {
				return nil
			}).Reset()
			err := req.CheckField()
			convey.So(err, convey.ShouldBeError, errors.New("failed ValidateResources"))
		})
		convey.Convey("when CheckSubmissionId failed", func() {
			defer gomonkey.ApplyMethodFunc(&SubmitRequest{}, "ValidateResources", func() error {
				return nil
			}).Reset()
			defer gomonkey.ApplyMethodFunc(&SubmitRequest{}, "CheckSubmissionId", func() error {
				return errors.New("failed CheckSubmissionId")
			}).Reset()
			err := req.CheckField()
			convey.So(err, convey.ShouldBeError, errors.New("failed CheckSubmissionId"))
		})
		convey.Convey("when process success", func() {
			defer gomonkey.ApplyMethodFunc(&SubmitRequest{}, "ValidateResources", func() error {
				return nil
			}).Reset()
			defer gomonkey.ApplyMethodFunc(&SubmitRequest{}, "CheckSubmissionId", func() error {
				return nil
			}).Reset()
			err := req.CheckField()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestSubmitRequest_ValidateResources(t *testing.T) {
	convey.Convey("test (req *SubmitRequest) ValidateResources()", t, func() {
		req := &SubmitRequest{
			Entrypoint:   "python script.py",
			SubmissionId: "",
			EntrypointResources: map[string]float64{
				"NPU": 0,
			},
			EntrypointNumCpus: 0,
			EntrypointNumGpus: 0,
			EntrypointMemory:  0,
		}
		convey.Convey("when req.EntrypointNumCpus < 0", func() {
			req.EntrypointNumCpus = -0.1
			err := req.ValidateResources()
			convey.So(err.Error(), convey.ShouldEqual, "entrypoint_num_cpus should not be less than 0")
		})
		convey.Convey("when req.EntrypointNumGpus < 0", func() {
			req.EntrypointNumGpus = -0.1
			err := req.ValidateResources()
			convey.So(err.Error(), convey.ShouldEqual, "entrypoint_num_gpus should not be less than 0")
		})
		convey.Convey("when req.EntrypointMemory < 0", func() {
			req.EntrypointMemory = -1
			err := req.ValidateResources()
			convey.So(err.Error(), convey.ShouldEqual, "entrypoint_memory should not be less than 0")
		})
		convey.Convey("when process success", func() {
			err := req.ValidateResources()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestSubmitRequest_CheckSubmissionId(t *testing.T) {
	convey.Convey("test (req *SubmitRequest) CheckSubmissionId()", t, func() {
		req := &SubmitRequest{
			Entrypoint:   "python script.py",
			SubmissionId: "123",
		}
		convey.Convey("when req.SubmissionId is empty", func() {
			req.SubmissionId = ""
			err := req.CheckSubmissionId()
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("when req.SubmissionId start with driver", func() {
			req.SubmissionId = "driver-123"
			err := req.CheckSubmissionId()
			convey.So(err.Error(), convey.ShouldEqual, "submission_id should not contain 'driver'")
		})
		convey.Convey("when req.SubmissionId doesn't start with 'app-'", func() {
			err := req.CheckSubmissionId()
			convey.So(err, convey.ShouldBeNil)
			convey.So(req.SubmissionId, convey.ShouldStartWith, jobIDPrefix)
		})
		convey.Convey("when req.SubmissionId length is 60 without 'app-'", func() {
			req.SubmissionId = "023456781234567822345678323456784234567852345678623456787234"
			err := req.CheckSubmissionId()
			convey.So(err, convey.ShouldBeNil)
			convey.So(req.SubmissionId, convey.ShouldStartWith, jobIDPrefix)
		})
		convey.Convey("when req.SubmissionId length is more than 60 without 'app-'", func() {
			req.SubmissionId = "0234567812345678223456783234567842345678523456786234567872345"
			err := req.CheckSubmissionId()
			convey.So(err.Error(), convey.ShouldStartWith, "regular expression validation error,")
			convey.So(req.SubmissionId, convey.ShouldStartWith, jobIDPrefix)
		})
		convey.Convey("when req.SubmissionId length is 64 with 'app-'", func() {
			req.SubmissionId = "app-023456781234567822345678323456784234567852345678623456787234"
			err := req.CheckSubmissionId()
			convey.So(err, convey.ShouldBeNil)
			convey.So(req.SubmissionId, convey.ShouldStartWith, jobIDPrefix)
		})
		convey.Convey("when req.SubmissionId length is more than 64 with 'app-'", func() {
			req.SubmissionId = "app-0234567812345678223456783234567842345678523456786234567872345"
			err := req.CheckSubmissionId()
			convey.So(err.Error(), convey.ShouldStartWith, "regular expression validation error,")
			convey.So(req.SubmissionId, convey.ShouldStartWith, jobIDPrefix)
		})
		convey.Convey("when process success", func() {
			err := req.CheckSubmissionId()
			convey.So(err, convey.ShouldBeNil)
			convey.So(req.SubmissionId, convey.ShouldStartWith, jobIDPrefix)
		})
	})
}

func TestSubmitRequest_NewSubmissionID(t *testing.T) {
	convey.Convey("test (req *SubmitRequest) NewSubmissionID()", t, func() {
		req := &SubmitRequest{
			Entrypoint:   "python script.py",
			SubmissionId: "",
		}
		convey.Convey("when req.SubmissionId is empty", func() {
			req.NewSubmissionID()
			convey.So(req.SubmissionId, convey.ShouldNotBeEmpty)
			convey.So(req.SubmissionId, convey.ShouldStartWith, jobIDPrefix)
		})
	})
}

func TestSubmitRequest_AddCreateOptions(t *testing.T) {
	convey.Convey("test (req *SubmitRequest) AddCreateOptions()", t, func() {
		req := &SubmitRequest{
			Entrypoint:   "python script.py",
			SubmissionId: "123",
		}
		convey.Convey("when req.CreateOptions is empty", func() {
			req.AddCreateOptions("key", "value")
			convey.So(len(req.CreateOptions), convey.ShouldEqual, 1)
		})
		convey.Convey("when key is empty", func() {
			req.AddCreateOptions("", "value")
			convey.So(len(req.CreateOptions), convey.ShouldEqual, 0)
		})
		convey.Convey("when key is not empty", func() {
			req.AddCreateOptions("key", "value")
			convey.So(len(req.CreateOptions), convey.ShouldEqual, 1)
		})
	})
}

func TestBuildJobResponse(t *testing.T) {
	convey.Convey("test BuildJobResponse", t, func() {
		convey.Convey("when process success", func() {
			expectedResult := Response{
				Code:    0,
				Message: "",
				Data:    []byte("test"),
			}
			result := BuildJobResponse("test", 0, nil)
			convey.So(result.Code, convey.ShouldEqual, expectedResult.Code)
			convey.So(result.Message, convey.ShouldEqual, expectedResult.Message)
			convey.So(string(result.Data), convey.ShouldEqual, "\""+string(expectedResult.Data)+"\"")
		})
		convey.Convey("when data is nil", func() {
			expectedResult := Response{
				Code:    http.StatusOK,
				Message: "",
				Data:    nil,
			}
			result := BuildJobResponse(nil, http.StatusOK, nil)
			convey.So(result, convey.ShouldResemble, expectedResult)
		})
		convey.Convey("when response status is "+strconv.Itoa(http.StatusBadRequest), func() {
			expectedResult := Response{
				Code:    http.StatusBadRequest,
				Message: "error request",
				Data:    nil,
			}
			result := BuildJobResponse(nil, http.StatusBadRequest, errors.New("error request"))
			convey.So(result, convey.ShouldResemble, expectedResult)
		})
		convey.Convey("when data marshal failed", func() {
			expectedResult := Response{
				Code:    http.StatusInternalServerError,
				Message: "marshal job response failed, err:",
			}
			result := BuildJobResponse(func() {}, http.StatusOK, nil)
			convey.So(result.Code, convey.ShouldEqual, expectedResult.Code)
			convey.So(result.Message, convey.ShouldStartWith, expectedResult.Message)
			convey.So(result.Data, convey.ShouldBeNil)
		})
	})
}
