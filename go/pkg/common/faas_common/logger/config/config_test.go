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

// Package config is common logger client
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/utils"
	mockUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
)

func TestInitConfig(t *testing.T) {
	convey.Convey("TestInitConfig", t, func() {
		convey.Convey("test 1", func() {
			patches := gomonkey.ApplyFunc(GetCoreInfoFromEnv, func() (CoreInfo, error) {
				return defaultCoreInfo, nil
			})
			defer patches.Reset()
			coreInfo, err := GetCoreInfoFromEnv()
			fmt.Printf("log config:%+v\n", coreInfo)
			convey.So(err, convey.ShouldEqual, nil)
		})
	})
}

func TestInitConfigWithReadFileError(t *testing.T) {
	convey.Convey("TestInitConfigWithEmptyPath", t, func() {
		convey.Convey("test 1", func() {
			patches := [...]*gomonkey.Patches{
				gomonkey.ApplyFunc(ioutil.ReadFile,
					func(filename string) ([]byte, error) {
						return nil, errors.New("mock read file error")
					}),
			}
			defer func() {
				for _, p := range patches {
					p.Reset()
				}
			}()
			coreInfo, err := GetCoreInfoFromEnv()
			fmt.Printf("error:%s\n", err)
			fmt.Printf("log config:%+v\n", coreInfo)
			convey.So(err, convey.ShouldNotEqual, nil)
		})
	})
}

func TestInitConfigWithErrorJson(t *testing.T) {
	convey.Convey("TestInitConfigWithEmptyPath", t, func() {
		convey.Convey("test 1", func() {
			mockErrorJson := "{\n\"filepath\": \"/home/sn/mock\",\n\"level\": \"INFO\",\n\"maxsize\": " +
				"500,\n\"maxbackups\": 1,\n\"maxage\": 1,\n\"compress\": true\n"
			patches := [...]*gomonkey.Patches{
				gomonkey.ApplyFunc(ioutil.ReadFile,
					func(filename string) ([]byte, error) {
						return []byte(mockErrorJson), nil
					}),
			}
			defer func() {
				for _, p := range patches {
					p.Reset()
				}
			}()
			coreInfo, err := GetCoreInfoFromEnv()
			fmt.Printf("error:%s\n", err)
			fmt.Printf("log config:%+v\n", coreInfo)
			convey.So(err, convey.ShouldNotEqual, nil)
		})
	})
}

func TestInitConfigWithEmptyPath(t *testing.T) {
	convey.Convey("TestInitConfigWithEmptyPath", t, func() {
		convey.Convey("test 1", func() {
			mockCfgInfo := "{\n\"filepath\": \"\",\n\"level\": \"INFO\",\n\"maxsize\": " +
				"500,\n\"maxbackups\": 1,\n\"maxage\": 1,\n\"compress\": true\n}"
			patches := [...]*gomonkey.Patches{
				gomonkey.ApplyFunc(ioutil.ReadFile,
					func(filename string) ([]byte, error) {
						return []byte(mockCfgInfo), nil
					}),
			}
			defer func() {
				for _, p := range patches {
					p.Reset()
				}
			}()
			coreInfo, err := GetCoreInfoFromEnv()
			fmt.Printf("error:%s\n", err)
			fmt.Printf("log config:%+v\n", coreInfo)
			convey.So(err, convey.ShouldNotEqual, nil)
		})
	})
}

func TestInitConfigWithValidateError(t *testing.T) {
	convey.Convey("TestInitConfigWithEmptyPath", t, func() {
		convey.Convey("test 1", func() {
			mockErrorJson := "{\n\"filepath\": \"some_relative_path\",\n\"level\": \"INFO\",\n\"maxsize\": " +
				"500,\n\"maxbackups\": 1,\n\"maxage\": 1}"
			patches := [...]*gomonkey.Patches{
				gomonkey.ApplyFunc(ioutil.ReadFile,
					func(filename string) ([]byte, error) {
						return []byte(mockErrorJson), nil
					}),
			}
			defer func() {
				for _, p := range patches {
					p.Reset()
				}
			}()
			coreInfo, err := GetCoreInfoFromEnv()
			fmt.Printf("error:%s\n", err)
			fmt.Printf("log config:%+v\n", coreInfo)
			convey.So(err, convey.ShouldNotEqual, nil)
		})
	})
}

func TestGetDefaultCoreInfo(t *testing.T) {
	tests := []struct {
		name string
		want CoreInfo
	}{
		{
			name: "test001",
			want: CoreInfo{
				FilePath:   "/home/snuser/log",
				Level:      "INFO",
				Tick:       0, // Unit: Second
				First:      0, // Unit: Number of logs
				Thereafter: 0, // Unit: Number of logs
				SingleSize: 100,
				Threshold:  10,
				Tracing:    false, // tracing log switch
				Disable:    false, // Disable file logger
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDefaultCoreInfo(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDefaultCoreInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractCoreInfoFromEnv(t *testing.T) {
	normalInfo, _ := json.Marshal(defaultCoreInfo)
	abnormal1 := mockUtils.PatchSlice{}
	abnormalInfo1, _ := json.Marshal(abnormal1)
	abnormal2 := CoreInfo{
		FilePath:   "",
		Level:      "INFO",
		Tick:       10,    // Unit: Second
		First:      10,    // Unit: Number of logs
		Thereafter: 5,     // Unit: Number of logs
		Tracing:    false, // tracing log switch
		Disable:    false, // Disable file logger
	}
	abnormalInfo2, _ := json.Marshal(abnormal2)
	type args struct {
		env string
	}
	tests := []struct {
		name        string
		args        args
		want        CoreInfo
		wantErr     bool
		patchesFunc mockUtils.PatchesFunc
	}{
		{
			name:    "case1",
			args:    args{logConfigKey},
			want:    defaultCoreInfo,
			wantErr: false,
			patchesFunc: func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					gomonkey.ApplyFunc(os.Getenv,
						func(key string) string {
							return string(normalInfo)
						}),
				})
				return patches
			},
		},
		{
			name:    "case2",
			args:    args{logConfigKey},
			want:    defaultCoreInfo,
			wantErr: true,
			patchesFunc: func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					gomonkey.ApplyFunc(os.Getenv,
						func(key string) string {
							return string(abnormalInfo1)
						}),
				})
				return patches
			},
		},
		{
			name:    "case3",
			args:    args{logConfigKey},
			want:    defaultCoreInfo,
			wantErr: true,
			patchesFunc: func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					gomonkey.ApplyFunc(os.Getenv,
						func(key string) string {
							return string(abnormalInfo2)
						}),
				})
				return patches
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := tt.patchesFunc()
			got, err := ExtractCoreInfoFromEnv(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractCoreInfoFromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractCoreInfoFromEnv() got = %v, want %v", got, tt.want)
			}
			patches.ResetAll()
		})
	}
}

func TestGetCoreInfoFromEnv(t *testing.T) {
	convey.Convey("GetCoreInfoFromEnv", t, func() {
		convey.Convey("ValidateFilePath error", func() {
			defer gomonkey.ApplyFunc(ExtractCoreInfoFromEnv, func(env string) (CoreInfo, error) {
				return CoreInfo{FilePath: "../test"}, nil
			}).Reset()
			_, err := GetCoreInfoFromEnv()
			convey.So(err, convey.ShouldBeError)
		})

		convey.Convey("MkdirAll error", func() {
			patches := []*gomonkey.Patches{
				gomonkey.ApplyFunc(ExtractCoreInfoFromEnv, func(env string) (CoreInfo, error) {
					return CoreInfo{FilePath: "/home/test"}, nil
				}),
				gomonkey.ApplyFunc(utils.ValidateFilePath, func(path string) error {
					return nil
				}),
				gomonkey.ApplyFunc(os.MkdirAll, func(path string, perm os.FileMode) error {
					return errors.New("create dir error")
				}),
			}
			defer func() {
				for _, patch := range patches {
					patch.Reset()
				}
			}()
			_, err := GetCoreInfoFromEnv()
			convey.So(err, convey.ShouldBeError)
		})

		convey.Convey("success", func() {
			patches := []*gomonkey.Patches{
				gomonkey.ApplyFunc(ExtractCoreInfoFromEnv, func(env string) (CoreInfo, error) {
					return CoreInfo{FilePath: "/home/test"}, nil
				}),
				gomonkey.ApplyFunc(utils.ValidateFilePath, func(path string) error {
					return nil
				}),
				gomonkey.ApplyFunc(os.MkdirAll, func(path string, perm os.FileMode) error {
					return nil
				}),
			}
			defer func() {
				for _, patch := range patches {
					patch.Reset()
				}
			}()
			env, err := GetCoreInfoFromEnv()
			convey.So(err, convey.ShouldBeNil)
			convey.So(env.FilePath, convey.ShouldEqual, "/home/test")
		})
	})
}
