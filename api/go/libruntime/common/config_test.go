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

// Package common for tools
package common

import (
	"testing"
	"github.com/smartystreets/goconvey/convey"
)

func TestGetConfig(t *testing.T) {
	convey.Convey(
		"Test Get config", t, func() {
			convey.Convey(
				"Test get config success", func() {
					conf := GetConfig()
					convey.So(conf.LogPath, convey.ShouldBeEmpty)
				},
			)
		},
	)
}
