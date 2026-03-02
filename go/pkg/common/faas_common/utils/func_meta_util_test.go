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

// Package utils -
package utils

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/types"
)

func TestGetFuncMetaSignature(t *testing.T) {
	convey.Convey("success", t, func() {
		signature := GetFuncMetaSignature(&types.FunctionMetaInfo{}, true)
		convey.So(signature, convey.ShouldEqual, "1830885943")
	})
	convey.Convey("marshal error", t, func() {
		defer gomonkey.ApplyFunc(json.Marshal, func(v interface{}) ([]byte, error) {
			return nil, fmt.Errorf("marshal error")
		}).Reset()
		str := GetFuncMetaSignature(&types.FunctionMetaInfo{}, true)
		convey.So(str, convey.ShouldContainSubstring, "invalid function meta info")
	})
	convey.Convey("unmarshal error", t, func() {
		defer gomonkey.ApplyFunc(json.Unmarshal, func(data []byte, v interface{}) error {
			return fmt.Errorf("unmarshal error")
		}).Reset()
		str := GetFuncMetaSignature(&types.FunctionMetaInfo{}, true)
		convey.So(str, convey.ShouldContainSubstring, "invalid function meta info")
	})
}

func TestSetFuncMetaDynamicConfEnable(t *testing.T) {
	type args struct {
		metaInfo *types.FunctionMetaInfo
	}
	tests := []struct {
		name string
		args args
	}{
		{"case1", args{metaInfo: &types.FunctionMetaInfo{}}},
		{"case2", args{metaInfo: &types.FunctionMetaInfo{FuncMetaData: types.FuncMetaData{Version: constant.DefaultURNVersion},
			ExtendedMetaData: types.ExtendedMetaData{DynamicConfig: types.DynamicConfigEvent{UpdateTime: "1"}}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetFuncMetaDynamicConfEnable(tt.args.metaInfo)
		})
	}
}

func TestGetCustomResource(t *testing.T) {
	convey.Convey("success", t, func() {
		customResources := getCustomResourceSpec("{\"huawei.com/ascend-1980\":8}", "")
		convey.So(customResources, convey.ShouldEqual, "{\"instanceType\":\"376T\"}")
	})

	convey.Convey("success", t, func() {
		customResources := getCustomResourceSpec("{\"huawei.com/ascend-1980\": 8}", "{\"instanceType\": \"376T\"}")
		convey.So(customResources, convey.ShouldEqual, "{\"instanceType\":\"376T\"}")
	})

	convey.Convey("success", t, func() {
		customResources := getCustomResourceSpec("{\"huawei.com/ascend-1980\":8}", "{ \"instanceType\":  \"280T\"}")
		convey.So(customResources, convey.ShouldEqual, "{\"instanceType\":\"280T\"}")
	})
}
