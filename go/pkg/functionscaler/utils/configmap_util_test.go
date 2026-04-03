/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
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

package utils

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

func TestIsNeedRaspSideCar(t *testing.T) {
	convey.Convey("IsNeedRaspSideCar", t, func() {
		isNeed := IsNeedRaspSideCar(&types.FunctionSpecification{
			ExtendedMetaData: commonTypes.ExtendedMetaData{
				RaspConfig: commonTypes.RaspConfig{
					RaspImage:      "someImage",
					InitImage:      "someImage",
					RaspServerIP:   "1.2.3.4",
					RaspServerPort: "1234",
				},
			},
		})
		convey.So(isNeed, convey.ShouldEqual, true)
		isNeed = IsNeedRaspSideCar(&types.FunctionSpecification{
			ExtendedMetaData: commonTypes.ExtendedMetaData{
				RaspConfig: commonTypes.RaspConfig{
					InitImage:      "someImage",
					RaspServerIP:   "1.2.3.4",
					RaspServerPort: "1234",
				},
			},
		})
		convey.So(isNeed, convey.ShouldEqual, false)
		isNeed = IsNeedRaspSideCar(&types.FunctionSpecification{
			ExtendedMetaData: commonTypes.ExtendedMetaData{
				RaspConfig: commonTypes.RaspConfig{
					RaspImage:      "someImage",
					InitImage:      "someImage",
					RaspServerPort: "1234",
				},
			},
		})
		convey.So(isNeed, convey.ShouldEqual, false)
		isNeed = IsNeedRaspSideCar(&types.FunctionSpecification{
			ExtendedMetaData: commonTypes.ExtendedMetaData{
				RaspConfig: commonTypes.RaspConfig{
					RaspImage:    "someImage",
					InitImage:    "someImage",
					RaspServerIP: "1.2.3.4",
				},
			},
		})
		convey.So(isNeed, convey.ShouldEqual, false)
	})
}
