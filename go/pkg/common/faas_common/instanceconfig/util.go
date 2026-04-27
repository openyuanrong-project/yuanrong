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

// Package instanceconfig -
package instanceconfig

import (
	"encoding/json"
	"fmt"
	"strings"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/types"
	wisecloudtypes "yuanrong.org/kernel/pkg/common/faas_common/wisecloudtool/types"
)

const (
	keySeparator                       = "/"
	insConfigTenantValueIndex          = 7
	insConfigFuncNameValueIndex        = 9
	insConfigVersionValueIndex         = 11
	validEtcdKeyLenForInsConfig        = 12
	insConfigLabelValueIndex           = 13
	validEtcdKeyLenForInsWithLabelConf = 14

	insConfigKeyIndex          = 1
	insConfigClusterKeyIndex   = 4
	insConfigClusterValueIndex = 5
	insConfigTenantKeyIndex    = 6
	insConfigFunctionKeyIndex  = 8
	insConfigLabelKeyIndex     = 12

	functionClusterKeyIdx = 5

	// InsConfigEtcdPrefix - 函数实例配置项元数据key前缀
	InsConfigEtcdPrefix = "/instances"
)

// GetLabelFromInstanceConfigEtcdKey -
func GetLabelFromInstanceConfigEtcdKey(etcdKey string) string {
	items := strings.Split(etcdKey, keySeparator)
	if len(items) != validEtcdKeyLenForInsWithLabelConf {
		return ""
	}
	return items[insConfigLabelValueIndex]
}

// ParseInstanceConfigFromEtcdEvent -
func ParseInstanceConfigFromEtcdEvent(etcdKey string, etcdValue []byte) (*Configuration, error) {
	items := strings.Split(etcdKey, keySeparator)
	if len(items) != validEtcdKeyLenForInsConfig && len(items) != validEtcdKeyLenForInsWithLabelConf {
		return nil, fmt.Errorf("etcdKey format error")
	}

	funcKey := fmt.Sprintf("%s/%s/%s", items[insConfigTenantValueIndex], items[insConfigFuncNameValueIndex],
		items[insConfigVersionValueIndex])

	label := ""
	if len(items) == validEtcdKeyLenForInsWithLabelConf {
		label = items[insConfigLabelValueIndex]
	}

	if len(etcdValue) == 0 {
		return nil, fmt.Errorf("etcdValue is empty")
	}
	insConfig := &Configuration{}
	err := json.Unmarshal(etcdValue, insConfig)
	if err != nil {
		return nil, fmt.Errorf("unmarshal etcdValue failed, err: %s", err.Error())
	}

	insConfig.FuncKey = funcKey
	insConfig.InstanceLabel = label
	return insConfig, nil
}

// Configuration -
type Configuration struct {
	FuncKey          string
	InstanceLabel    string
	InstanceMetaData types.InstanceMetaData         `json:"instanceMetaData" valid:",optional"`
	NuwaRuntimeInfo  wisecloudtypes.NuwaRuntimeInfo `json:"nuwaRuntimeInfo" valid:",optional"`
}

// DeepCopy return a Configuration Copy
func (i *Configuration) DeepCopy() *Configuration {
	return &(*i)
}

// GetWatcherFilter -
func GetWatcherFilter(clusterId string) func(event *etcd3.Event) bool {
	return func(event *etcd3.Event) bool {
		items := strings.Split(event.Key, keySeparator)
		if len(items) != validEtcdKeyLenForInsConfig && len(items) != validEtcdKeyLenForInsWithLabelConf {
			return true
		}
		if items[insConfigKeyIndex] != "instances" || items[insConfigClusterKeyIndex] != "cluster" ||
			items[insConfigTenantKeyIndex] != "tenant" || items[insConfigFunctionKeyIndex] != "function" {
			return true
		}
		if len(items) == validEtcdKeyLenForInsWithLabelConf && items[insConfigLabelKeyIndex] != "label" {
			return true
		}
		if clusterId != items[insConfigClusterValueIndex] {
			return true
		}
		return false
	}
}
