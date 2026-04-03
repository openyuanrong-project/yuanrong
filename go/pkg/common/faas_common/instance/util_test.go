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

package instance

import (
	"testing"

	"github.com/stretchr/testify/assert"

	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
)

func TestGetInstanceIDFromEtcdKey(t *testing.T) {
	etcdKey := "/sn/instance/business/yrk/tenant/123/function/faasscheduler/version/$latest/defaultaz/requestID/abc"
	instanceID := GetInstanceIDFromEtcdKey(etcdKey)
	assert.Equal(t, "abc", instanceID)

	instanceIDNil := GetInstanceIDFromEtcdKey("")
	assert.Equal(t, "", instanceIDNil)
}

func TestGetInsSpecFromEtcdValue(t *testing.T) {
	etcdValue := []byte("{\"instanceID\":\"51f71580-3a07-4000-8000-004b56e7f471\",\"requestID\":\"7fb31" +
		"b50-7c5a-11ed-a991-fa163e3523c8\",\"runtimeID\":\"runtime-e06fe343-0000-4000-8000-00bbad15e23" +
		"8\",\"runtimeAddress\":\"127.0.0.1:33333\",\"functionAgentID\":\"function_agent_127.0.0" +
		".1-33333\",\"functionProxyID\":\"dggphis35893-8490\",\"function\":" +
		"\"default/0-system-hello/$latest\",\"resources\":{\"resources\":{\"Memory\":{\"name\":\"Memor" +
		"y\",\"scalar\":{\"value\":500}},\"CPU\":{\"name\":\"CPU\",\"scalar\":{\"value\":500}}}},\"sched" +
		"uleOption\":{\"affinity\":{\"instanceAffinity\":{}}},\"instanceStatus\":{\"code\":3,\"msg\":\"i" +
		"nstance is running\"}}")
	insSpecTrans := GetInsSpecFromEtcdValue("", etcdValue)
	insSpecExpected := &commonTypes.InstanceSpecification{
		InstanceID:      "51f71580-3a07-4000-8000-004b56e7f471",
		RequestID:       "7fb31b50-7c5a-11ed-a991-fa163e3523c8",
		RuntimeID:       "runtime-e06fe343-0000-4000-8000-00bbad15e238",
		RuntimeAddress:  "127.0.0.1:33333",
		FunctionAgentID: "function_agent_127.0.0.1-33333",
		FunctionProxyID: "dggphis35893-8490",
		Function:        "default/0-system-hello/$latest",
		RestartPolicy:   "",
		Resources: commonTypes.Resources{
			Resources: map[string]commonTypes.Resource{
				"CPU": commonTypes.Resource{
					Name:   "CPU",
					Scalar: commonTypes.ValueScalar{Value: 500},
				},
				"Memory": commonTypes.Resource{
					Name:   "Memory",
					Scalar: commonTypes.ValueScalar{Value: 500},
				},
			},
		},
		ActualUse: commonTypes.Resources{},
		ScheduleOption: commonTypes.ScheduleOption{
			Affinity: commonTypes.Affinity{
				InstanceAffinity: commonTypes.InstanceAffinity{},
			},
		},
		CreateOptions:  nil,
		Labels:         nil,
		StartTime:      "",
		InstanceStatus: commonTypes.InstanceStatus{Code: 3, Msg: "instance is running"},
		JobID:          "",
		SchedulerChain: nil,
	}
	assert.Equal(t, insSpecExpected, insSpecTrans)

	insSpecNil := GetInsSpecFromEtcdValue("", []byte(""))
	assert.Equal(t, true, insSpecNil != nil)
}
