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

package config

import (
	"testing"

	"yuanrong.org/kernel/runtime/faassdk/types"
)

func TestGetConfig(t *testing.T) {
	type args struct {
		funcSpec *types.FuncSpec
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"case1 succedd to get config", args{funcSpec: &types.FuncSpec{
			FuncMetaData: types.FuncMetaData{
				FunctionVersionURN: "sn:cn:yrk:12345678901234561234567890123456:function:0@yrservice@test-image-env",
			},
		}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetConfig(tt.args.funcSpec)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
