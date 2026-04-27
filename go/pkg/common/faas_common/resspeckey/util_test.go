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

package resspeckey

import (
	"encoding/json"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/common/faas_common/types"
)

func TestResSpecKey(t *testing.T) {
	resSpec := &ResourceSpecification{
		CPU:                 100,
		Memory:              100,
		CustomResources:     map[string]int64{"NPU": 1},
		CustomResourcesSpec: map[string]interface{}{"Type": "type1"},
		InvokeLabel:         "label1",
	}
	resKey := ConvertToResSpecKey(resSpec)
	resKeyString := resKey.String()
	assert.Equal(t, "cpu-100-mem-100-storage-0-cstRes-{\"NPU\":1}-cstResSpec-{\"Type\":\"type1\"}-invokeLabel-label1", resKeyString)
	resSpec1 := resKey.ToResSpec()
	assert.Equal(t, int64(100), resSpec1.CPU)
	assert.Equal(t, int64(100), resSpec1.Memory)
	assert.Equal(t, "label1", resSpec1.InvokeLabel)
}

func TestConvertResourceMetaData(t *testing.T) {
	convey.Convey("test ConvertResourceMetaData", t, func() {
		convey.Convey("Unmarshal error", func() {
			resMeta := types.ResourceMetaData{
				CustomResourcesSpec: "huawei.com/ascend-1980:D910B",
				CustomResources:     "",
			}
			resource := ConvertResourceMetaDataToResSpec(resMeta)
			convey.So(len(resource.CustomResources), convey.ShouldEqual, 0)
		})
		convey.Convey("Convert success", func() {
			customResources := map[string]int64{"huawei.com/ascend-1980": 10}
			data, _ := json.Marshal(customResources)
			resMeta := types.ResourceMetaData{
				CustomResourcesSpec: "CustomResourcesSpec",
				CustomResources:     string(data),
			}
			resource := ConvertResourceMetaDataToResSpec(resMeta)
			convey.So(resource.CustomResourcesSpec[ascendResourceD910BInstanceType],
				convey.ShouldEqual, "376T")
		})
	})
}
