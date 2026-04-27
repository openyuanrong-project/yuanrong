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
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/common/faas_common/localauth"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	"yuanrong.org/kernel/pkg/common/faas_common/statuscode"
	"yuanrong.org/kernel/pkg/common/faas_common/tls"
	mockUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

const mockFuncKey = "7e1ad6a6-cc5c-44fa-bd54-25873f72a86a/0@default@testyrhttpwebsocket/latest"

func TestScaleUpInstance(t *testing.T) {
	type args struct {
		scaleUpParam *ScaleUpParam
	}
	config.GlobalConfig = types.Configuration{
		LocalAuth: localauth.AuthConfig{
			AKey:     "000",
			SKey:     "111",
			Duration: 5,
		},
		HTTPSConfig: &tls.InternalHTTPSConfig{
			HTTPSEnable: false,
			TLSProtocol: "TLSv1.2",
			TLSCiphers:  "TLS_ECDHE_RSA",
		},
	}
	mockArg := args{
		scaleUpParam: &ScaleUpParam{
			TraceID:     "testTraceID",
			FunctionKey: mockFuncKey,
			Timeout:     1 * time.Second,
			CPU:         400,
			Memory:      256,
		},
	}
	mockRespBody := `{"code":150200,"instance":{"instanceID":"pool30-xxx"},"worker":{"instances":[],
		"functionname":"","functionversion":"","tenant":"","business":""}}`

	tests := []struct {
		name        string
		args        args
		patchesFunc mockUtils.PatchesFunc
		want        *types.WmInstance
		wantErr     error
	}{
		{
			name: "case 1 scale up instance successfully",
			args: mockArg,
			patchesFunc: func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					ApplyFunc((*http.Client).Do, func(_ *http.Client, req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body: ioutil.NopCloser(bytes.NewReader(
								[]byte(mockRespBody))),
						}, nil
					}),
					ApplyFunc(FillInWorkerManagerRequestHeaders, func(request *http.Request) { return }),
				})
				return patches
			},
			want: &types.WmInstance{
				InstanceID: "pool30-xxx",
			},
			wantErr: nil,
		},
		{
			name: "case 2 scale up instance failed",
			args: mockArg,
			patchesFunc: func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					ApplyFunc((*http.Client).Do, func(_ *http.Client, req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 500,
							Body: ioutil.NopCloser(bytes.NewReader(
								[]byte(`{"code": 5000,"message":"mockMessage"}`))),
						}, nil
					}),
					ApplyFunc(FillInWorkerManagerRequestHeaders, func(request *http.Request) { return }),
				})
				return patches
			},
			want:    nil,
			wantErr: snerror.New(5000, "mockMessage"),
		},
		{
			name: "case 3 scale up instance instance failed when do request error",
			args: mockArg,
			patchesFunc: func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					ApplyFunc((*http.Client).Do, func(_ *http.Client, req *http.Request) (*http.Response, error) {
						return nil, errors.New("do request mock error")
					}),
					ApplyFunc(FillInWorkerManagerRequestHeaders, func(request *http.Request) { return }),
				})
				return patches
			},
			want:    nil,
			wantErr: errors.New("do request mock error"),
		},
		{
			name: "case 4 scale up instance failed when unmarshal scale up bad response error",
			args: mockArg,
			patchesFunc: func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					ApplyFunc((*http.Client).Do, func(_ *http.Client, req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 500,
							Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{`))),
						}, nil
					}),
					ApplyFunc(FillInWorkerManagerRequestHeaders, func(request *http.Request) { return }),
				})
				return patches
			},
			want:    nil,
			wantErr: errors.New("unexpected end of JSON input"),
		},
		{
			name: "case 5 scale up instance instance successfully when unmarshal scale up response error",
			args: mockArg,
			patchesFunc: func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					ApplyFunc((*http.Client).Do, func(_ *http.Client, req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body: ioutil.NopCloser(bytes.NewReader(
								[]byte("}"))),
						}, nil
					}),
					ApplyFunc(FillInWorkerManagerRequestHeaders, func(request *http.Request) { return }),
				})
				return patches
			},
			want:    nil,
			wantErr: snerror.New(statuscode.ScaleUpRequestErrCode, statuscode.ScaleUpRequestErrMsg),
		},
		{
			name: "case 6 response code in GetWorkerSuccessResponse is not 200",
			args: mockArg,
			patchesFunc: func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					ApplyFunc((*http.Client).Do, func(_ *http.Client, req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
						}, nil
					}),
					ApplyFunc(FillInWorkerManagerRequestHeaders, func(request *http.Request) { return }),
				})
				return patches
			},
			want:    nil,
			wantErr: snerror.New(0, ""),
		},
		{
			name: "case 7 scale up instance successfully when failed to make scale up request",
			args: mockArg,
			patchesFunc: func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					ApplyFunc(json.Marshal, func(v any) ([]byte, error) {
						return nil, errors.New("json marshal error")
					}),
				})
				return patches
			},
			want:    nil,
			wantErr: errors.New("make scale up request failed"),
		},
		{
			name: "case 8 scale up instance instance failed when read body error",
			args: mockArg,
			patchesFunc: func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					ApplyFunc((*http.Client).Do, func(_ *http.Client, req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body: ioutil.NopCloser(bytes.NewReader(
								[]byte(mockRespBody))),
						}, nil
					}),
					ApplyFunc(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
						return nil, errors.New("read body mock error")
					}),
					ApplyFunc(FillInWorkerManagerRequestHeaders, func(request *http.Request) { return }),
				})
				return patches
			},
			want:    nil,
			wantErr: snerror.New(statuscode.ScaleUpRequestErrCode, statuscode.ScaleUpRequestErrMsg),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := tt.patchesFunc()
			wmInstance, err := ScaleUpInstance(tt.args.scaleUpParam)
			if err != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.Equal(t, tt.want.InstanceID, wmInstance.InstanceID)
			}
			patches.ResetAll()
		})
	}
}

func TestScaleDownInstance(t *testing.T) {
	config.GlobalConfig = types.Configuration{
		LocalAuth: localauth.AuthConfig{
			AKey:     "000",
			SKey:     "111",
			Duration: 5,
		},
		HTTPSConfig: &tls.InternalHTTPSConfig{
			HTTPSEnable: false,
			TLSProtocol: "TLSv1.2",
			TLSCiphers:  "TLS_ECDHE_RSA",
		},
	}
	type args struct {
		instanceID  string
		functionKey string
		traceID     string
	}
	tests := []struct {
		name        string
		args        args
		patchesFunc mockUtils.PatchesFunc
		wantErr     error
	}{
		{
			name: "case 1 scale down function instance",
			args: args{
				instanceID:  "default-#-pool30",
				functionKey: mockFuncKey,
				traceID:     "mockTraceID",
			},
			patchesFunc: func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					ApplyFunc((*http.Client).Do, func(_ *http.Client, req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"code":200,"message":"mock message"}`))),
						}, nil
					}),
					ApplyFunc(FillInWorkerManagerRequestHeaders, func(request *http.Request) { return }),
				})
				return patches
			},
			wantErr: nil,
		},
		{
			name: "case 2 scale down function instance when do request err",
			args: args{
				instanceID:  "default-#-pool30",
				functionKey: mockFuncKey,
				traceID:     "mockTraceID",
			},
			patchesFunc: func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					ApplyFunc((*http.Client).Do, func(_ *http.Client, req *http.Request) (*http.Response, error) {
						return nil, errors.New("do request mock error")
					}),
					ApplyFunc(FillInWorkerManagerRequestHeaders, func(request *http.Request) { return }),
				})
				return patches
			},
			wantErr: errors.New("do request mock error"),
		},
		{
			name: "case 3 scale down function instance when ioutil.ReadAll error",
			args: args{
				instanceID:  "default-#-pool30",
				functionKey: mockFuncKey,
				traceID:     "mockTraceID",
			},
			patchesFunc: func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					ApplyFunc((*http.Client).Do, func(_ *http.Client, req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"code":200,"message":"mock message"}`))),
						}, nil
					}),
					ApplyFunc(FillInWorkerManagerRequestHeaders, func(request *http.Request) { return }),
					ApplyFunc(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
						return nil, errors.New("read body mock error")
					}),
				})
				return patches
			},
			wantErr: errors.New("read body mock error"),
		},
		{
			name: "case 4 scale down  function instance when returned bad response body",
			args: args{
				instanceID:  "default-#-pool30",
				functionKey: mockFuncKey,
				traceID:     "mockTraceID",
			},
			patchesFunc: func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					ApplyFunc((*http.Client).Do, func(_ *http.Client, req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 500,
							Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"code":500,"message":"mock message"}`))),
						}, nil
					}),
					ApplyFunc(FillInWorkerManagerRequestHeaders, func(request *http.Request) { return }),
				})
				return patches
			},
			wantErr: snerror.New(500, "mock message"),
		},
		{
			name: "case 5 scale down  function instance when failed to unmarshal bad response body",
			args: args{
				instanceID:  "default-#-pool30",
				functionKey: mockFuncKey,
				traceID:     "mockTraceID",
			},
			patchesFunc: func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					ApplyFunc((*http.Client).Do, func(_ *http.Client, req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 500,
							Body: ioutil.NopCloser(bytes.NewReader(
								[]byte(`{`),
							)),
						}, nil
					}),
					ApplyFunc(FillInWorkerManagerRequestHeaders, func(request *http.Request) { return }),
				})
				return patches
			},
			wantErr: errors.New("unexpected end of JSON input"),
		},
		{
			name: "case 6 scale down function instance when failed to unmarshal DeleteWorkerResponse",
			args: args{
				instanceID:  "default-#-pool30",
				functionKey: mockFuncKey,
				traceID:     "mockTraceID",
			},
			patchesFunc: func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					ApplyFunc((*http.Client).Do, func(_ *http.Client, req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{`))),
						}, nil
					}),
					ApplyFunc(FillInWorkerManagerRequestHeaders, func(request *http.Request) { return }),
				})
				return patches
			},
			wantErr: errors.New("unexpected end of JSON input"),
		},
		{
			name: "case 7 scale down function instance when status code in DeleteWorkerResponse is not 200",
			args: args{
				instanceID:  "default-#-pool30",
				functionKey: mockFuncKey,
				traceID:     "mockTraceID",
			},
			patchesFunc: func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					ApplyFunc((*http.Client).Do, func(_ *http.Client, req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"code":500}`))),
						}, nil
					}),
					ApplyFunc(FillInWorkerManagerRequestHeaders, func(request *http.Request) { return }),
				})
				return patches
			},
			wantErr: errors.New("worker manager returned error"),
		},
		{
			name: "case 8 scale down function instance failed because json marshal failed",
			args: args{
				instanceID:  "default-#-pool30",
				functionKey: mockFuncKey,
				traceID:     "mockTraceID",
			},
			patchesFunc: func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					ApplyFunc(json.Marshal, func(v any) ([]byte, error) {
						return nil, errors.New("json marshal error")
					}),
				})
				return patches
			},
			wantErr: errors.New("failed to make scale down request"),
		},
		{
			name: "case 9 scale down function instance failed because new http request failed",
			args: args{
				instanceID:  "default-#-pool30",
				functionKey: mockFuncKey,
				traceID:     "mockTraceID",
			},
			patchesFunc: func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					ApplyFunc(http.NewRequestWithContext, func(
						ctx context.Context, method string, url string, body io.Reader) (*http.Request, error) {
						return nil, errors.New("http.NewRequestWithContext")
					}),
				})
				return patches
			},
			wantErr: errors.New("failed to make scale down request"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := tt.patchesFunc()
			err := ScaleDownInstance(tt.args.instanceID, tt.args.functionKey, tt.args.traceID)
			if err != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.Equal(t, tt.wantErr, err)
			}
			patches.ResetAll()
		})
	}
}

func TestNeedTryError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "SNError with GettingPodErrorCode",
			err:  snerror.New(statuscode.GettingPodErrorCode, ""),
			want: true,
		},
		{
			name: "SNError with CancelGeneralizePod",
			err:  snerror.New(statuscode.ReachMaxOnDemandInstancesPerTenant, ""),
			want: true,
		},
		{
			name: "Non-SNError",
			err:  errors.New("non-SNError"),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NeedTryError(tt.err); got != tt.want {
				t.Errorf("NeedTryError() = %v, want %v", got, tt.want)
			}
		})
	}
}
