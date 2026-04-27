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

// Package handler for faas executor api
package handler

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
	"yuanrong.org/kernel/runtime/faassdk/config"
	"yuanrong.org/kernel/runtime/faassdk/utils"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

func TestGetUserCodePath(t *testing.T) {
	convey.Convey("GetUserCodePath", t, func() {
		convey.Convey("not found", func() {
			path, err := GetUserCodePath()
			convey.So(path, convey.ShouldEqual, "")
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("success", func() {
			os.Setenv(config.DelegateDownloadPath, "userCodePath")
			path, err := GetUserCodePath()
			convey.So(path, convey.ShouldEqual, "userCodePath")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDealEnv(t *testing.T) {
	environment := map[string]string{"test1": "abc"}
	environmentStr, _ := json.Marshal(environment)
	encryptedUserData := map[string]string{"test2": "def"}
	encryptedUserDataStr, _ := json.Marshal(encryptedUserData)
	env1 := map[string]string{"environment": string(environmentStr), "encrypted_user_data": string(encryptedUserDataStr)}
	env1Str, _ := json.Marshal(env1)

	tests := []struct {
		name        string
		want        error
		want1       map[string]string
		patchesFunc PatchesFunc
	}{
		{"case1 env is nil", nil, nil, func() PatchSlice {
			patches := InitPatchSlice()
			patches.Append(PatchSlice{
				ApplyFunc(os.Getenv, func(key string) string {
					return ""
				}),
			})
			return patches
		}},
		{"case2 env has environment and encrypted_user_data", nil,
			map[string]string{"test1": "abc", "test2": "def"},
			func() PatchSlice {
				patches := InitPatchSlice()
				patches.Append(PatchSlice{
					ApplyFunc(os.Getenv, func(key string) string {
						return string(env1Str)
					}),
					ApplyFunc(os.Environ, func() []string {
						return []string{"test1=abc"}
					}),
				})
				return patches
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := tt.patchesFunc()
			got, got1 := utils.DealEnv()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dealEnv() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("dealEnv() got1 = %v, want %v", got1, tt.want1)
			}
			patches.ResetAll()
		})
	}

	convey.Convey("failed to Unmarshal environment", t, func() {
		patches := InitPatchSlice()
		patches.Append(PatchSlice{
			ApplyFunc(json.Unmarshal, func(data []byte, v interface{}) error {
				return errors.New("json unmarshal error")
			}),
		})
		os.Setenv("ENV_DELEGATE_DECRYPT", "{}")
		err, _ := utils.DealEnv()
		convey.So(err, convey.ShouldBeError)
		patches.ResetAll()
	})
}
