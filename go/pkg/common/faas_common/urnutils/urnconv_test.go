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

// Package urnutils -
package urnutils

import "testing"

func TestCrNameByKey(t *testing.T) {
	type args struct {
		funcKey string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"case1 succeed to get CrNameByKey", args{funcKey: "/sn/functions/business/yrk/tenant" +
			"/172120022624850603/function/0@default@testurpccustomoom002/version/latest"},
			"yyrk1721-0-default-testurpccustomoom002-latest-1257561201"},
		{"case2 long funcName", args{funcKey: "/sn/functions/business/yrk/tenant/default/" +
			"function/0-actordemo-test-actor-support-version-publish-delete-version/version/$latest"},
			"yyrkdefa-port-version-publish-delete-versio-$latest-3843651668"},
		{"case3 long version", args{funcKey: "/sn/functions/business/yrk/tenant/default/function" +
			"/0-actordemo-test-actor-support-version-publish-delete-version/version/123456789123456789123456789123456789123456789123456789123456"},
			"yyrkdefa-elete-versio-123456789123456789123456789123-503881880"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CrNameByKey(tt.args.funcKey); got != tt.want {
				t.Errorf("CrNameByKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
