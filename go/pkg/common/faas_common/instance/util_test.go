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
		"8\",\"runtimeAddress\":\"x.x.x.x:33333\",\"functionAgentID\":\"function_agent_10.244.16" +
		"2.129-33333\",\"functionProxyID\":\"dggphis35893-8490\",\"function\":\"12345678901234561234567" +
		"890123456/0-system-hello/$latest\",\"resources\":{\"resources\":{\"Memory\":{\"name\":\"Memor" +
		"y\",\"scalar\":{\"value\":500}},\"CPU\":{\"name\":\"CPU\",\"scalar\":{\"value\":500}}}},\"sched" +
		"uleOption\":{\"affinity\":{\"instanceAffinity\":{}}},\"instanceStatus\":{\"code\":3,\"msg\":\"i" +
		"nstance is running\"}}")
	insSpecTrans := GetInsSpecFromEtcdValue("", etcdValue)
	insSpecExpected := &commonTypes.InstanceSpecification{
		InstanceID:      "51f71580-3a07-4000-8000-004b56e7f471",
		RequestID:       "7fb31b50-7c5a-11ed-a991-fa163e3523c8",
		RuntimeID:       "runtime-e06fe343-0000-4000-8000-00bbad15e238",
		RuntimeAddress:  "x.x.x.x:33333",
		FunctionAgentID: "function_agent_x.x.x.x-33333",
		FunctionProxyID: "dggphis35893-8490",
		Function:        "12345678901234561234567890123456/0-system-hello/$latest",
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
