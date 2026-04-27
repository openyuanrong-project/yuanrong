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

// Package handlers for handle request
package etcdcache

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/types"
)

func TestGetJobIDByParentID(t *testing.T) {
	convey.Convey("Test GetJobIDByParentID:", t, func() {
		convey.Convey("Test InstancesByInstanceID when instances is empty", func() {
			instancesMap := InstanceCache.GetByParentID("app-123")
			convey.So(len(instancesMap), convey.ShouldEqual, 0)
		})
		convey.Convey("Test InstancesByInstanceID success", func() {
			instance := &types.InstanceSpecification{InstanceID: "id-123", ParentID: "app-123", JobID: "job-123"}
			InstanceCache.id2Instance[instance.InstanceID] = instance
			InstanceCache.jobID2Instances[instance.JobID] = map[string]*types.InstanceSpecification{
				instance.InstanceID: instance}
			instancesMap := InstanceCache.GetByParentID("app-123")
			convey.So(instancesMap[instance.InstanceID].InstanceID, convey.ShouldEqual, instance.InstanceID)
			instancesMap = InstanceCache.GetByJobID(instance.JobID)
			convey.So(instancesMap[instance.InstanceID].InstanceID, convey.ShouldEqual, instance.InstanceID)
			newInstance := InstanceCache.Get(instance.InstanceID)
			convey.So(newInstance.InstanceID, convey.ShouldEqual, instance.InstanceID)
			str := InstanceCache.String()
			convey.So(str, convey.ShouldContainSubstring, instance.InstanceID)
		})
	})
}
