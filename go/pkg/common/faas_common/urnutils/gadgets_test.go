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

// Package urnutils contains URN element definitions and tools
package urnutils

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestComplexFuncName_GetSvcIDWithPrefix(t *testing.T) {
	tests := []struct {
		name   string
		fields ComplexFuncName
		want   string
	}{
		{
			name: "normal",
			fields: ComplexFuncName{
				prefix:    ServiceIDPrefix,
				ServiceID: "absserviceid",
				FuncName:  "absFuncName",
			},
			want: "0-absserviceid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ComplexFuncName{
				prefix:    tt.fields.prefix,
				ServiceID: tt.fields.ServiceID,
				FuncName:  tt.fields.FuncName,
			}
			if got := c.GetSvcIDWithPrefix(); got != tt.want {
				t.Errorf("GetSvcIDWithPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComplexFuncName_ParseFrom(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want *ComplexFuncName
	}{
		{
			name: "normal",
			args: args{
				name: "0-absserviceid-absFunc-Name",
			},
			want: &ComplexFuncName{
				prefix:    ServiceIDPrefix,
				ServiceID: "absserviceid",
				FuncName:  "absFunc-Name",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ComplexFuncName{}
			if got := c.ParseFrom(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComplexFuncName_String(t *testing.T) {
	tests := []struct {
		name   string
		fields ComplexFuncName
		want   string
	}{
		{
			name: "normal",
			fields: ComplexFuncName{
				prefix:    ServiceIDPrefix,
				ServiceID: "absserviceid",
				FuncName:  "absFunc-Name",
			},
			want: "0-absserviceid-absFunc-Name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ComplexFuncName{
				prefix:    tt.fields.prefix,
				ServiceID: tt.fields.ServiceID,
				FuncName:  tt.fields.FuncName,
			}
			if got := c.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewComplexFuncName(t *testing.T) {
	type args struct {
		svcID    string
		funcName string
	}
	tests := []struct {
		name string
		args args
		want *ComplexFuncName
	}{
		{
			name: "normal",
			args: args{
				svcID:    "absserviceid",
				funcName: "absFunc-Name",
			},
			want: &ComplexFuncName{
				prefix:    ServiceIDPrefix,
				ServiceID: "absserviceid",
				FuncName:  "absFunc-Name",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewComplexFuncName(tt.args.svcID, tt.args.funcName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewComplexFuncName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetSeparator(t *testing.T) {
	SetSeparator("@")
	assert.Equal(t, "@", separator)
}
