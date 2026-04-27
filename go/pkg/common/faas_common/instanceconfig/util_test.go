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

package instanceconfig

import (
	"encoding/json"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/instance"
	"yuanrong.org/kernel/pkg/common/faas_common/types"
	wisecloudtypes "yuanrong.org/kernel/pkg/common/faas_common/wisecloudtool/types"
)

func TestParseInstanceConfigFromEtcdEvent(t *testing.T) {
	convey.Convey("Test ParseInstanceConfigFromEtcdEvent", t, func() {
		testConfig := &Configuration{
			InstanceMetaData: types.InstanceMetaData{PoolID: "test"},
			NuwaRuntimeInfo:  wisecloudtypes.NuwaRuntimeInfo{WisecloudRuntimeId: "runtime1"},
		}
		testConfigData, _ := json.Marshal(testConfig)

		convey.Convey("should parse config with label successfully", func() {
			etcdKey := "/instances/business/yrk/cluster/cluster001/tenant/default/function/0@test111@yrfunc111/version/latest/label/aaa"
			config, err := ParseInstanceConfigFromEtcdEvent(etcdKey, testConfigData)
			convey.So(err, convey.ShouldBeNil)
			convey.So(config.FuncKey, convey.ShouldEqual, "default/0@test111@yrfunc111/latest")
			convey.So(config.InstanceLabel, convey.ShouldEqual, "aaa")
		})

		convey.Convey("should return error for invalid key format", func() {
			key := "/invalid/key"
			_, err := ParseInstanceConfigFromEtcdEvent(key, testConfigData)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestGetWatcherFilter(t *testing.T) {
	convey.Convey("Test GetWatcherFilter", t, func() {
		filter := GetWatcherFilter("cluster1")

		convey.Convey("should filter matching cluster key", func() {
			event := &etcd3.Event{
				Key: "/instances/business/yrk/cluster/cluster1/tenant/t1/function/f1/version/v1",
			}
			convey.So(filter(event), convey.ShouldBeFalse)
		})

		convey.Convey("should not filter invalid key structure", func() {
			event := &etcd3.Event{
				Key: "/invalid/key",
			}
			convey.So(filter(event), convey.ShouldBeTrue)
		})
	})
}

func TestGetLabelFromInstanceConfigEtcdKey(t *testing.T) {
	etcdKey := "/instances/business/yrk/cluster/cluster001/tenant/default/function/0@test111@yrfunc111/version/latest"
	label := GetLabelFromInstanceConfigEtcdKey(etcdKey)
	assert.Equal(t, "", label)

	etcdKey = "/instances/business/yrk/cluster/cluster001/tenant/default/function/0@test111@yrfunc111/version/latest/label/aaa"
	label = instance.GetInstanceIDFromEtcdKey(etcdKey)
	assert.Equal(t, "aaa", label)
}
