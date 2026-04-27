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

package registry

import (
	"reflect"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

func TestGetInstanceIDFromFGEtcdKey(t *testing.T) {
	type args struct {
		etcdKey string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"test1",
			args{etcdKey: "/sn/workers/business/yrk/tenant/6d5b16f6ef0e4b7d938d5035356aa378/function/0@default@app1/version/latest/defaultaz" +
				"/defaultaz-#-custom-600-512-cbf49869-a2e7-46e0-b9bc-c11533f38db5"},
			"defaultaz-#-custom-600-512-cbf49869-a2e7-46e0-b9bc-c11533f38db5"},
		{"test2", args{etcdKey: ""}, ""},
		{"test3",
			args{etcdKey: "/sn/workers/business/yrk/tenant/6d5b16f6ef0e4b7d938d5035356aa378/function/0@default@app1/version/latest/defaultaz" +
				"/defaultaz"}, "defaultaz"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetInstanceIDFromFGEtcdKey(tt.args.etcdKey)
			assert.Equal(t, tt.want, got)
		})
	}

}

func TestInstanceWatcherFGFilter(t *testing.T) {
	stopCh := make(chan struct{})
	instanceRegistry := NewInstanceRegistry(stopCh)
	event := &etcd3.Event{
		Type: etcd3.PUT,
		Key: "/sn/workers/business/yrk/tenant/6d5b16f6ef0e4b7d938d5035356aa378/function/0@default@app1/version/latest/defaultaz" +
			"/defaultaz-#-custom-600-512-cbf49869-a2e7-46e0-b9bc-c11533f38db5",
		Value: []byte("value"),
		Rev:   1,
	}
	assert.Equal(t, false, instanceRegistry.watcherFGFilter(event))

	event.Key = "/sn/workers/business/yrk/tenant/123/instance/faasscheduler/version/$latest/defaultaz/requestID/abc"
	assert.Equal(t, true, instanceRegistry.watcherFGFilter(event))
}

func TestGetInsSpecFromEtcdKV(t *testing.T) {
	etcdKey := "/sn/workers/business/yrk/tenant/6d5b16f6ef0e4b7d938d5035356aa378/function/0@default@app1/" +
		"version/latest/defaultaz/defaultaz-#-custom-600-512-cbf49869-a2e7-46e0-b9bc-c11533f38db5"
	etcdValue := []byte("{\"ip\":\"192.168.0.97\",\"port\":\"8080\",\"cluster\":\"cluster001\",\"status\"" +
		":\"ready\",\"p2pPort\":\"22668\",\"nodeIP\":\"127.0.0.1\",\"nodePort\":\"22423\"," +
		"\"applier\":\"worker-manager\",\"ownerIP\":\"127.0.0.1\",\"cpu\":600,\"memory\":512,\"businessType\":\"CAE\"," +
		"\"hasInitializer\":true,\"creationTime\":1719788553,\"resource\":{\"worker\":{\"cpuLimit\":1000,\"cpuRequest\":100," +
		"\"memoryLimit\":200,\"memoryRequest\":100}," +
		"\"runtime\":{\"cpuLimit\":400,\"cpuRequest\":60,\"memoryLimit\":256,\"memoryRequest\":256}}}")
	defer ApplyMethod(reflect.TypeOf(GlobalRegistry), "GetFuncSpec",
		func(_ *Registry, funcKey string) *types.FunctionSpecification {
			return &types.FunctionSpecification{
				ResourceMetaData: commonTypes.ResourceMetaData{
					EphemeralStorage: 512,
				},
			}
		}).Reset()
	insSpecTrans := GetInsSpecFromEtcdKVForFG(etcdKey, etcdValue)
	insSpecExpected := &commonTypes.InstanceSpecification{
		InstanceID:      "",
		RequestID:       "",
		RuntimeID:       "",
		RuntimeAddress:  "127.0.0.1:22423",
		FunctionAgentID: "",
		FunctionProxyID: "192.168.0.97:8080",
		Function:        "6d5b16f6ef0e4b7d938d5035356aa378/0@default@app1/latest",
		RestartPolicy:   "",
		Resources: commonTypes.Resources{
			Resources: map[string]commonTypes.Resource{
				constant.ResourceCPUName: {
					Name: constant.ResourceCPUName,
					Scalar: commonTypes.ValueScalar{
						Value: float64(400),
					},
				},
				constant.ResourceMemoryName: {
					Name: constant.ResourceMemoryName,
					Scalar: commonTypes.ValueScalar{
						Value: float64(256),
					},
				},
				constant.ResourceEphemeralStorage: {
					Name: constant.ResourceEphemeralStorage,
					Scalar: commonTypes.ValueScalar{
						Value: float64(512),
					},
				},
			},
		},
		ActualUse: commonTypes.Resources{},
		ScheduleOption: commonTypes.ScheduleOption{
			Affinity: commonTypes.Affinity{
				InstanceAffinity: commonTypes.InstanceAffinity{},
			},
		},
		CreateOptions: map[string]string{
			types.InstanceTypeNote: "reserved",
			types.FunctionKeyNote:  "6d5b16f6ef0e4b7d938d5035356aa378/0@default@app1/latest",
			types.SchedulerIDNote:  "worker-manager",
		},
		Labels:    nil,
		StartTime: "1719788553",
		InstanceStatus: commonTypes.InstanceStatus{
			Code: 3,
			Msg:  "",
		},
		JobID:          "",
		SchedulerChain: nil,
		ParentID:       "worker-manager",
	}
	assert.Equal(t, insSpecExpected, insSpecTrans)
}
