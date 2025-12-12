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

package logger

import (
	"errors"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/config"
)

func TestCreateSink(t *testing.T) {
	convey.Convey("Test Create Sink Error", t, func() {
		coreInfo := config.CoreInfo{}
		w, err := CreateSink(coreInfo)
		convey.So(w, convey.ShouldBeNil)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("Test Create Sink Error 2", t, func() {
		patch := gomonkey.ApplyFunc(os.MkdirAll, func(path string, perm os.FileMode) error {
			return errors.New("err")
		})
		defer patch.Reset()
		coreInfo := config.CoreInfo{}
		w, err := CreateSink(coreInfo)
		convey.So(w, convey.ShouldBeNil)
		convey.So(err, convey.ShouldNotBeNil)
	})

}
