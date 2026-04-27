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

package service

import (
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"

	"yuanrong.org/kernel/pkg/common/faas_common/k8sclient"
	"yuanrong.org/kernel/pkg/system_function_controller/config"
	"yuanrong.org/kernel/pkg/system_function_controller/types"
)

func TestCreateFrontendService(t *testing.T) {
	convey.Convey("CreateFrontendService", t, func() {
		defer gomonkey.ApplyMethod(reflect.TypeOf(&k8sclient.KubeClient{}), "CreateK8sService",
			func(_ *k8sclient.KubeClient, service *v1.Service) error {
				return nil
			}).Reset()
		defer gomonkey.ApplyFunc(config.GetFaaSControllerConfig, func() types.Config {
			return types.Config{NameSpace: "default"}
		}).Reset()
		err := CreateFrontendService()
		convey.So(err, convey.ShouldBeNil)
	})
}
