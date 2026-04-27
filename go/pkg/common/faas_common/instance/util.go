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

// Package instance
package instance

import (
	"encoding/json"
	"fmt"
	"strings"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/types"
)

const (
	keySeparator = "/"

	instanceIDValueIndex       = 13
	validEtcdKeyLenForInstance = 14
)

// GetInstanceIDFromEtcdKey gets instance id from etcd key of instance
func GetInstanceIDFromEtcdKey(etcdKey string) string {
	items := strings.Split(etcdKey, keySeparator)
	if len(items) != validEtcdKeyLenForInstance {
		return ""
	}
	return fmt.Sprintf("%s", items[instanceIDValueIndex])
}

// GetInsSpecFromEtcdValue gets InstanceSpecification from etcd value of instance
func GetInsSpecFromEtcdValue(etcdKey string, etcdValue []byte) *types.InstanceSpecification {
	insSpec := &types.InstanceSpecification{}
	if len(etcdValue) != 0 {
		err := json.Unmarshal(etcdValue, insSpec)
		if err != nil {
			log.GetLogger().Errorf("failed to unmarshal etcd value to instance specification %s", err.Error())
			return nil
		}
	} else {
		log.GetLogger().Warnf("etcd value is empty when get instance specification from key %s", etcdKey)
	}
	return insSpec
}
