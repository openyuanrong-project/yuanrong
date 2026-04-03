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

package urnutils

import (
	"net"
	"os"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	mockUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
)

func TestProductUrn_ParseFrom(t *testing.T) {
	absURN := FunctionURN{
		"absPrefix",
		"absZone",
		"absBusinessID",
		"absTenantID",
		"absProductID",
		"absName",
		"latest",
	}
	absURNStr := "absPrefix:absZone:absBusinessID:absTenantID:absProductID:absName:latest"
	type args struct {
		urn string
	}
	tests := []struct {
		name   string
		fields FunctionURN
		args   args
		want   FunctionURN
	}{
		{
			name: "normal test",
			args: args{
				absURNStr,
			},
			want: absURN,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &FunctionURN{}
			if _ = p.ParseFrom(tt.args.urn); !reflect.DeepEqual(*p, tt.want) {
				t.Errorf("ParseFrom() p = %v, want %v", *p, tt.want)
			}
		})
	}
}

func TestProductUrn_String(t *testing.T) {
	tests := []struct {
		name   string
		fields FunctionURN
		want   string
	}{
		{
			"stringify with version",
			FunctionURN{
				"absPrefix",
				"absZone",
				"absBusinessID",
				"absTenantID",
				"absProductID",
				"absName",
				"latest",
			},
			"absPrefix:absZone:absBusinessID:absTenantID:absProductID:absName:latest",
		},
		{
			"stringify without version",
			FunctionURN{
				ProductID:  "absPrefix",
				RegionID:   "absZone",
				BusinessID: "absBusinessID",
				TenantID:   "absTenantID",
				TypeSign:   "absProductID",
				FuncName:   "absName",
			},
			"absPrefix:absZone:absBusinessID:absTenantID:absProductID:absName",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &FunctionURN{
				ProductID:   tt.fields.ProductID,
				RegionID:    tt.fields.RegionID,
				BusinessID:  tt.fields.BusinessID,
				TenantID:    tt.fields.TenantID,
				TypeSign:    tt.fields.TypeSign,
				FuncName:    tt.fields.FuncName,
				FuncVersion: tt.fields.FuncVersion,
			}
			if got := p.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProductUrn_StringWithoutVersion(t *testing.T) {
	tests := []struct {
		name   string
		fields FunctionURN
		want   string
	}{
		{
			"stringify without version",
			FunctionURN{
				"absPrefix",
				"absZone",
				"absBusinessID",
				"absTenantID",
				"absProductID",
				"absName",
				"latest",
			},
			"absPrefix:absZone:absBusinessID:absTenantID:absProductID:absName",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &FunctionURN{
				ProductID:   tt.fields.ProductID,
				RegionID:    tt.fields.RegionID,
				BusinessID:  tt.fields.BusinessID,
				TenantID:    tt.fields.TenantID,
				TypeSign:    tt.fields.TypeSign,
				FuncName:    tt.fields.FuncName,
				FuncVersion: tt.fields.FuncVersion,
			}
			if got := p.StringWithoutVersion(); got != tt.want {
				t.Errorf("StringWithoutVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnonymize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"0", anonymization},
		{"123", anonymization},
		{"123456", anonymization},
		{"1234567", "123****567"},
		{"12345678901234546", "123****546"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, Anonymize(tt.input))
	}
}

func TestAnonymizeTenantURN(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"absPrefix:absZone:absBusinessID:absTenantID:absProductID:absName", "absPrefix:absZone:absBusinessID:abs****tID:absProductID:absName"},
		{"absPrefix:absZone:absBusinessID:absTenantID:absProductID:absName:latest", "absPrefix:absZone:absBusinessID:abs****tID:absProductID:absName:latest"},
		{"a:b:c", "a:b:c"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, AnonymizeTenantURN(tt.input))
	}
}

func TestBaseURN_Valid(t *testing.T) {
	separator = "@"
	urn := FunctionURN{
		ProductID:   "",
		RegionID:    "",
		BusinessID:  "",
		TenantID:    "",
		TypeSign:    "",
		FuncName:    "0@a_-9AA@AA",
		FuncVersion: "",
	}
	success := urn.Valid()
	assert.Equal(t, nil, success)

	urn.FuncName = "0@a_-9AA@tttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttt"
	success = urn.Valid()
	assert.Equal(t, nil, success)

	urn.FuncName = "0@a_-9AA@ttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttt"
	err := urn.Valid()
	assert.NotEqual(t, nil, err)

	urn.FuncName = "@func"
	err = urn.Valid()
	assert.NotEqual(t, nil, err)

	urn.FuncName = "0@func"
	err = urn.Valid()
	assert.NotEqual(t, nil, err)

	urn.FuncName = "0@^@^"
	err = urn.Valid()
	assert.NotEqual(t, nil, err)

	separator = "-"
}

// TestGetServiceNameFromFullName 是主测试函数
func TestGetServiceNameFromFullName(t *testing.T) {
	// 使用一个 map 来存储所有测试用例，方便管理和扩展
	tests := map[string]struct {
		input    string
		expected string
	}{
		"standard format with multiple @": {
			input:    "0@my-service@my-function",
			expected: "my-service",
		},
		"minimal valid format": {
			input:    "0@core-service",
			expected: "",
		},
		"service name with special characters": {
			input:    "0@default@streamtest",
			expected: "default",
		},
		"multiple separators, should pick by index": {
			input:    "@a@b@c@d",
			expected: "",
		},
		"wrong prefix": {
			input:    "invalid-prefix@my-service",
			expected: "",
		},
		"empty input": {
			input:    "",
			expected: "",
		},
		"input starts with @ but not the prefix": {
			input:    "@not-a-service@my-func",
			expected: "",
		},
	}

	// 遍历并执行每个测试用例
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// 调用被测函数
			got := GetServiceNameFromFullName(tt.input)

			// 断言结果是否符合预期
			if got != tt.expected {
				t.Errorf("GetServiceNameFromFullName(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestBaseURN_GetAlias(t *testing.T) {
	urn := FunctionURN{
		ProductID:   "",
		RegionID:    "",
		BusinessID:  "",
		TenantID:    "",
		TypeSign:    "",
		FuncName:    "0@a_-9AA@AA",
		FuncVersion: constant.DefaultURNVersion,
	}

	alias := urn.GetAlias()
	assert.Equal(t, "", alias)

	urn.FuncVersion = "old"
	alias = urn.GetAlias()
	assert.Equal(t, "old", alias)
}

func TestGetFuncInfoWithVersion(t *testing.T) {
	urn := "urn"
	_, err := GetFuncInfoWithVersion(urn)
	assert.NotEqual(t, nil, err)

	urn = "absPrefix:absZone:absBusinessID:absTenantID:absProductID:absName"
	_, err = GetFuncInfoWithVersion(urn)
	assert.NotEqual(t, nil, err)

	urn = "absPrefix:absZone:absBusinessID:absTenantID:absProductID:absName:latest"
	parsedURN, err := GetFuncInfoWithVersion(urn)
	assert.Equal(t, "absName", parsedURN.FuncName)
}

func TestAnonymizeTenantKey(t *testing.T) {
	inputKey := ""
	outputKey := AnonymizeTenantKey(inputKey)
	assert.Equal(t, "****", outputKey)

	inputKey = "input/key"
	outputKey = AnonymizeTenantKey(inputKey)
	assert.Equal(t, "****/key", outputKey)
}

func TestParseAliasURN(t *testing.T) {
	urn := ""
	alias := ParseAliasURN(urn)
	assert.Equal(t, urn, alias)

	urn = "absPrefix:absZone:absBusinessID:absTenantID:absProductID:absName:!latest"
	alias = ParseAliasURN(urn)
	assert.Equal(t, "absPrefix:absZone:absBusinessID:absTenantID:absProductID:absName:latest", alias)
}

func TestAnonymizeTenantURNSlice(t *testing.T) {
	inUrn := []string{"in", "in/urn"}
	outUrn := AnonymizeTenantURNSlice(inUrn)
	assert.Equal(t, "in", outUrn[0])
	assert.Equal(t, "in/urn", outUrn[1])
}

func TestBaseURN_GetAliasForFuncBranch(t *testing.T) {
	urn := FunctionURN{
		ProductID:   "",
		RegionID:    "",
		BusinessID:  "",
		TenantID:    "",
		TypeSign:    "",
		FuncName:    "0@a_-9AA@AA",
		FuncVersion: "!latest",
	}

	alias := urn.GetAliasForFuncBranch()
	assert.Equal(t, "latest", alias)

	urn.FuncVersion = "latest"
	alias = urn.GetAliasForFuncBranch()
	assert.Equal(t, "", alias)
}

func TestAnonymizeKeys(t *testing.T) {
	type args struct {
		keys []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"case", args{keys: []string{"123", "1234567"}}, []string{"****", "123****567"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, AnonymizeKeys(tt.args.keys), "AnonymizeKeys(%v)", tt.args.keys)
		})
	}
}

func TestBuildURNOrAliasURNTemp(t *testing.T) {
	type args struct {
		business       string
		tenant         string
		function       string
		versionOrAlias string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{}, ""},
		{"empty", args{"1", "2", "3", "4"}, "sn:cn:1:2:function:3:4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, BuildURNOrAliasURNTemp(tt.args.business, tt.args.tenant, tt.args.function, tt.args.versionOrAlias), "BuildURNOrAliasURNTemp(%v, %v, %v, %v)", tt.args.business, tt.args.tenant, tt.args.function, tt.args.versionOrAlias)
		})
	}
}

func TestCrNameByUrn(t *testing.T) {
	type args struct {
		args string
	}
	var a args
	a.args = "sn:cn:yrk:default:function:0@yrservice@test_func:v1"
	var b args
	b.args = ""
	tests := []struct {
		name string
		args args
		want string
	}{
		{"case1", a, "yyrkdefa-0-yrservice-test-func-v1-1258487427"},
		{"case2", b, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CrNameByURN(tt.args.args); got != tt.want {
				t.Errorf("CrNameByURN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetServerIP(t *testing.T) {
	tests := []struct {
		name        string
		want        string
		wantErr     bool
		patchesFunc mockUtils.PatchesFunc
	}{
		{"case1 succeed to get ip", "127.0.0.1", false, func() mockUtils.PatchSlice {
			patches := mockUtils.InitPatchSlice()
			patches.Append(mockUtils.PatchSlice{
				gomonkey.ApplyFunc(os.Hostname, func() (name string, err error) { return "127.0.0.1", nil })})
			patches.Append(mockUtils.PatchSlice{
				gomonkey.ApplyFunc(net.LookupHost,
					func(host string) (addrs []string, err error) { return []string{"127.0.0.1", "0"}, nil })})
			return patches
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := tt.patchesFunc()
			got, err := GetServerIP()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetServerIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetServerIP() got = %v, want %v", got, tt.want)
			}
			patches.ResetAll()
		})
	}
}

func TestCheckAliasUrnTenant(t *testing.T) {
	type args struct {
		tenantID string
		aliasUrn string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"case1", args{tenantID: "default",
			aliasUrn: "sn:cn:yrk:default:function:helloworld:myaliasv1"}, true},
		{"case2 error", args{tenantID: "default",
			aliasUrn: "sn:cn:yrk:default:function:helloworld"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckAliasUrnTenant(tt.args.tenantID, tt.args.aliasUrn); got != tt.want {
				t.Errorf("CheckAliasUrnTenant() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTenantFormFuncKey(t *testing.T) {
	type args struct {
		funcKey string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"case1", args{funcKey: "default/0-system-faasscheduler/$latest"},
			"default"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetTenantFromFuncKey(tt.args.funcKey); got != tt.want {
				t.Errorf("GetTenantFromFuncKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetShortFuncName(t *testing.T) {
	funcName := "testFunc1111111111111111111111111111111111111111111111111111111"
	shortFuncName := GetShortFuncName(funcName)
	assert.Equal(t, "testFunc1111111111111111111111111111111111111111111111111111111", shortFuncName)

	funcName = "testFunc1111111111111111111111111111111111111111111111111111111111111111111111111111111"
	shortFuncName = GetShortFuncName(funcName)
	assert.Equal(t, "X11111111111111111111111111111111111111111111111111111111111111", shortFuncName)
}

func TestGetFuncNameFromFuncKey(t *testing.T) {
	funcKey := "12345/test_func/latest/1"
	funcName := GetFuncNameFromFuncKey(funcKey)
	assert.Equal(t, "", funcName)

	funcKey = "12345/test_func/latest"
	funcName = GetFuncNameFromFuncKey(funcKey)
	assert.Equal(t, "12345/test_func", funcName)
}

func TestAnonymizeTenantMetadataEtcdKey(t *testing.T) {
	etcdKey := "/sn/quota/cluster/cluster001/tenant/7e1ad6a6-cc5c-44fa-bd54-25873f72a86a"
	AnonymizedKey := AnonymizeTenantMetadataEtcdKey(etcdKey)
	assert.Equal(t, "/sn/quota/cluster/cluster001/tenant/7e1****86a", AnonymizedKey)

	etcdKey = "/sn/quota/cluster/cluster001/tenant/7e1ad6a6-cc5c-44fa-bd54-25873f72a86a/instancemetadata"
	AnonymizedKey = GetFuncNameFromFuncKey(etcdKey)
	assert.Equal(t, "", AnonymizedKey)
}
