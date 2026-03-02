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

// Package localauth authenticates requests by local configmaps
package localauth

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"

	mockUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
)

func TestGetDecryptFromEnv(t *testing.T) {
	tests := []struct {
		name        string
		want        map[string]string
		wantErr     bool
		patchesFunc mockUtils.PatchesFunc
	}{
		{"case1 failed to unmarshal", make(map[string]string), false, func() mockUtils.PatchSlice {
			patches := mockUtils.InitPatchSlice()
			patches.Append(mockUtils.PatchSlice{
				gomonkey.ApplyFunc(json.Unmarshal, func(data []byte, v interface{}) error {
					return errors.New("failed to unmarshal json")
				}),
			})
			return patches
		}},
		{"case2 succeed to unmarshal", map[string]string{"test": "test"},
			false, func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					gomonkey.ApplyFunc(os.Getenv, func(key string) string {
						return `{"test":"test"}`
					}),
				})
				return patches
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := tt.patchesFunc()
			got, err := GetDecryptFromEnv()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDecryptFromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDecryptFromEnv() got = %v, want %v", got, tt.want)
			}
			patches.ResetAll()
		})
	}
}
