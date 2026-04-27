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
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestProductUrn_String(t *testing.T) {
	tests := []struct {
		name   string
		fields BaseURN
		want   string
	}{
		{
			"stringify with version",
			BaseURN{
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
			BaseURN{
				ProductID:  "absPrefix",
				RegionID:   "absZone",
				BusinessID: "absBusinessID",
				TenantID:   "absTenantID",
				TypeSign:   "absProductID",
				Name:       "absName",
			},
			"absPrefix:absZone:absBusinessID:absTenantID:absProductID:absName",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &BaseURN{
				ProductID:  tt.fields.ProductID,
				RegionID:   tt.fields.RegionID,
				BusinessID: tt.fields.BusinessID,
				TenantID:   tt.fields.TenantID,
				TypeSign:   tt.fields.TypeSign,
				Name:       tt.fields.Name,
				Version:    tt.fields.Version,
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
		fields BaseURN
		want   string
	}{
		{
			"stringify without version",
			BaseURN{
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
			p := &BaseURN{
				ProductID:  tt.fields.ProductID,
				RegionID:   tt.fields.RegionID,
				BusinessID: tt.fields.BusinessID,
				TenantID:   tt.fields.TenantID,
				TypeSign:   tt.fields.TypeSign,
				Name:       tt.fields.Name,
				Version:    tt.fields.Version,
			}
			if got := p.StringWithoutVersion(); got != tt.want {
				t.Errorf("StringWithoutVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFunctionInfo(t *testing.T) {
	convey.Convey("Test GetFunctionInfo", t, func() {
		convey.Convey("GetFunctionInfo when parsedURN error", func() {
			baseURN, err := GetFunctionInfo("urn")
			convey.So(baseURN, convey.ShouldEqual, BaseURN{})
			convey.So(err, convey.ShouldNotBeNil)
		})
		absURN := BaseURN{
			"absPrefix",
			"absZone",
			"absBusinessID",
			"absTenantID",
			"absProductID",
			"absName",
			"latest",
		}
		absURNStr := "absPrefix:absZone:absBusinessID:absTenantID:absProductID:absName:latest"
		convey.Convey("GetFunctionInfo success", func() {
			baseURN, err := GetFunctionInfo(absURNStr)
			convey.So(baseURN, convey.ShouldEqual, absURN)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestCombineFunctionKey(t *testing.T) {
	convey.Convey("Test CombineFunctionKey", t, func() {
		convey.Convey("CombineFunctionKey success", func() {
			str := CombineFunctionKey("tenantID1", "funcName1", "version1")
			convey.So(str, convey.ShouldEqual, "tenantID1/funcName1/version1")
		})
	})
}
