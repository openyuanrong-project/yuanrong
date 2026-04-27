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

// Package getinfo for get function-master info
package getinfo

import (
	"google.golang.org/protobuf/proto"

	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/message"
	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/resource"
)

// GetInstances function for get instances
func GetInstances(instPath string) ([]*resource.InstanceInfo, error) {
	body, err := requestFunctionMaster(instPath)
	if err != nil {
		return nil, err
	}
	var inst message.QueryInstancesInfoResponse
	err = proto.Unmarshal(body, &inst)
	if err != nil {
		return nil, err
	}
	return inst.InstanceInfos, nil
}
