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
package utils

import (
	"reflect"
	"testing"

	"yuanrong.org/kernel/pkg/common/faas_common/types"
)

func TestGenerateErrorResponse(t *testing.T) {
	type args struct {
		errorCode    int
		errorMessage string
	}
	var a args
	a.errorCode = 0
	a.errorMessage = "0"
	resp := &types.CallHandlerResponse{
		Code:    0,
		Message: "0",
	}
	tests := []struct {
		name string
		args args
		want *types.CallHandlerResponse
	}{
		{"case1", a, resp},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateErrorResponse(tt.args.errorCode, tt.args.errorMessage); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateErrorResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateSuccessResponse(t *testing.T) {
	type args struct {
		code    int
		message string
	}
	var a args
	a.code = 0
	a.message = "0"
	resp := &types.CallHandlerResponse{
		Code:    0,
		Message: "0",
	}
	tests := []struct {
		name string
		args args
		want *types.CallHandlerResponse
	}{
		{"case1", a, resp},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateSuccessResponse(tt.args.code, tt.args.message); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateSuccessResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}
