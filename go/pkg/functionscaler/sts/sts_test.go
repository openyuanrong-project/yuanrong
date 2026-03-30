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

// Package sts provide methods for obtaining sensitive information
package sts

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

func Test_getStsServerConfig(t *testing.T) {
	convey.Convey("test GetStsServerConfig", t, func() {
		convey.Convey("baseline", func() {
			funcSpec := &types.FunctionSpecification{
				StsMetaData: commonTypes.StsMetaData{
					EnableSts:    true,
					ServiceName:  "a",
					MicroService: "b",
				},
			}
			config.GlobalConfig.RawStsConfig.ServerConfig.Domain = "12345"
			config.GlobalConfig.RawStsConfig.StsDomainForRuntime = ""
			serverConfig := GetStsServerConfig(funcSpec)
			convey.So(serverConfig.Domain, convey.ShouldEqual, "12345")
			convey.So(serverConfig.Path, convey.ShouldEqual, "/opt/certs/a/b/b.ini")
			config.GlobalConfig.RawStsConfig.StsDomainForRuntime = "67890"
			convey.So(serverConfig.Domain, convey.ShouldEqual, "67890")
			convey.So(serverConfig.Path, convey.ShouldEqual, "/opt/certs/a/b/b.ini")
		})
	})
}
