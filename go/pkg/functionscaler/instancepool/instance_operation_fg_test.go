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

package instancepool

import (
	"context"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/workermanager"
)

func TestCreateInstanceForFG(t *testing.T) {
	funcSpec := &types.FunctionSpecification{
		FuncKey: "testFunction",
		FuncCtx: context.TODO(),
		FuncMetaData: commonTypes.FuncMetaData{
			Handler: "myHandler",
		},
		ResourceMetaData: commonTypes.ResourceMetaData{
			CPU:    500,
			Memory: 500,
		},
		InstanceMetaData: commonTypes.InstanceMetaData{
			MaxInstance:   100,
			ConcurrentNum: 100,
		},
		ExtendedMetaData: commonTypes.ExtendedMetaData{
			Initializer: commonTypes.Initializer{
				Handler: "myInitializer",
			},
			VpcConfig: &commonTypes.VpcConfig{},
		},
	}
	resKey := resspeckey.ConvertToResSpecKey(resspeckey.ConvertResourceMetaDataToResSpec(funcSpec.ResourceMetaData))
	instanceBuilder := func(instanceID string) *types.Instance {
		return &types.Instance{
			InstanceID: instanceID,
			ResKey:     resKey,
		}
	}
	request := createInstanceRequest{
		funcSpec:        funcSpec,
		resKey:          resKey,
		instanceBuilder: instanceBuilder,
	}
	convey.Convey("Test createInstanceForFG", t, func() {
		convey.Convey("create success", func() {
			defer gomonkey.ApplyFunc(workermanager.ScaleUpInstance,
				func(scaleUpParam *workermanager.ScaleUpParam) (*types.WmInstance, error) {
					return &types.WmInstance{}, nil
				}).Reset()
			instance, err := createInstanceForFG(request)
			convey.So(instance, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("create failed", func() {
			defer gomonkey.ApplyFunc(workermanager.ScaleUpInstance,
				func(scaleUpParam *workermanager.ScaleUpParam) (*types.WmInstance, error) {
					return nil, errors.New("create failed")
				}).Reset()
			instance, err := createInstanceForFG(request)
			convey.So(instance, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestDeleteInstanceForFG(t *testing.T) {
	convey.Convey("Test deleteInstanceForFG", t, func() {
		convey.Convey("delete success", func() {
			defer gomonkey.ApplyFunc(workermanager.ScaleDownInstance,
				func(instanceID, functionKey, traceID string) error {
					return nil
				}).Reset()
			err := deleteInstanceForFG(&types.FunctionSpecification{}, faasManagerInfo{}, &types.Instance{})
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("delete failed", func() {
			defer gomonkey.ApplyFunc(workermanager.ScaleDownInstance,
				func(instanceID, functionKey, traceID string) error {
					return errors.New("delete failed")
				}).Reset()
			err := deleteInstanceForFG(&types.FunctionSpecification{}, faasManagerInfo{}, &types.Instance{})
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestDeleteInstanceByIDForFG(t *testing.T) {
	convey.Convey("Test deleteInstanceByIDForFG", t, func() {
		convey.Convey("delete by id success", func() {
			defer gomonkey.ApplyFunc(workermanager.ScaleDownInstance,
				func(instanceID, functionKey, traceID string) error {
					return nil
				}).Reset()
			err := deleteInstanceByIDForFG("instance", "testFunc")
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("delete by id failed", func() {
			defer gomonkey.ApplyFunc(workermanager.ScaleDownInstance,
				func(instanceID, functionKey, traceID string) error {
					return errors.New("delete failed")
				}).Reset()
			err := deleteInstanceByIDForFG("instance", "testFunc")
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}
