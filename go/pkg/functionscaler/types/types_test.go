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

package types

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
)

func TestResourceSpecification_DeepCopy(t *testing.T) {
	tests := []struct {
		name         string
		resourceSpec *resspeckey.ResourceSpecification
	}{
		{
			name: "deep copy success",
			resourceSpec: &resspeckey.ResourceSpecification{
				CPU:             1,
				Memory:          2,
				InvokeLabel:     "xxx",
				CustomResources: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			specCopy := tt.resourceSpec.DeepCopy()
			if &specCopy == &tt.resourceSpec {
				t.Errorf("deepCopyFailed is %d, expectStr is %d", &specCopy, &tt.resourceSpec)
			}
			if &specCopy.InvokeLabel == &tt.resourceSpec.InvokeLabel {
				t.Errorf("deepCopy map failed is %d, expectStr is %d", &specCopy, &tt.resourceSpec)
			}
		})
	}
}

func TestIntOrString_UnmarshalJSONJSON(t *testing.T) {
	tests := []struct {
		name        string
		b           []byte
		targetCPU   int64
		targetLabel string
		targetErr   error
	}{
		{
			name:        "Unmarshal Json Normal",
			b:           []byte("{\"CPU\": 128, \"label\": \"aaaaa\"}"),
			targetCPU:   128,
			targetLabel: "aaaaa",
			targetErr:   nil,
		},
		{
			name:        "Unmarshal Json error label",
			b:           []byte("{\"CPU\": 128, \"label\": 123}"),
			targetCPU:   128,
			targetLabel: "",
			targetErr:   nil,
		},
		{
			name:        "Unmarshal Json type error",
			b:           []byte("{\"CPU\": 128, \"label\": []}"),
			targetCPU:   0,
			targetLabel: "",
			targetErr:   fmt.Errorf("expected int or string, but got []"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resMap := map[string]IntOrString{}
			err := json.Unmarshal(tt.b, &resMap)
			if err != nil {
				assert.Equal(t, tt.targetErr.Error(), err.Error())
				return
			}
			assert.Equal(t, tt.targetCPU, resMap["CPU"].IntVal)
			assert.Equal(t, tt.targetLabel, resMap["label"].StrVal)
		})
	}
}
