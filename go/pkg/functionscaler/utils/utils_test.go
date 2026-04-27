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

// Package utils -
package utils

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

func TestIsFaaSManager(t *testing.T) {
	convey.Convey("error funcKey", t, func() {
		is := IsFaaSManager("")
		convey.So(is, convey.ShouldBeFalse)
	})
	convey.Convey("right funcKey", t, func() {
		is := IsFaaSManager("0-system-faasmanager")
		convey.So(is, convey.ShouldBeTrue)
	})
}

func TestIsFaaSInstance(t *testing.T) {
	convey.Convey("TestIsFaaSInstance", t, func() {
		tt := []struct {
			FuncKey     string
			expectValue bool
		}{
			{
				FuncKey:     "0/0-system-faasscheduler/$latest",
				expectValue: false,
			},
			{
				FuncKey:     "0/0-system-faasmanager/$latest",
				expectValue: false,
			},
			{
				FuncKey:     "0/0-system-faasfrontend/$latest",
				expectValue: false,
			},
			{
				FuncKey:     "0/0-system-faascontroller/$latest",
				expectValue: false,
			},
			{
				FuncKey:     "12345678901234561234567890123456/0-system-faasExecutorGo1.x",
				expectValue: false,
			},
			{
				FuncKey:     "12345678901234561234567890123456",
				expectValue: false,
			},
			{
				FuncKey:     "",
				expectValue: false,
			},
			{
				FuncKey:     "12345678901234561234567890123456/0-system",
				expectValue: false,
			},
			{
				FuncKey:     "12345678901234561234567890123456/0-system-faasExecutorGo1.x/$latest",
				expectValue: true,
			},
		}

		for _, test := range tt {
			convey.So(IsFaaSInstance(test.FuncKey), convey.ShouldEqual, test.expectValue)
		}
	})
}

func TestConvertInstanceResource(t *testing.T) {
	convey.Convey("set resources", t, func() {
		b := []byte(`{"instanceID":"ad45c354-84bc-450c-8400-00000000caa9","requestID":"task-96bb75fc-f539-40f0-85c7-a884aa0ac3ef","functionAgentID":"function_agent_10.148.160.47-58866","functionProxyID":"10.41.5.115","function":"12345678901234561234567890123456/0-system-faasExecutorGo1.x/$latest","resources":{"resources":{"ephemeral-storage":{"name":"ephemeral-storage","scalar":{"value":6144}},"Memory":{"name":"Memory","scalar":{"value":1024}},"CPU":{"name":"CPU","scalar":{"value":1000}}}},"scheduleOption":{"schedPolicyName":"monopoly","affinity":{"instanceAffinity":{"affinity":{"yyrk2441-0-default-testcustom1024001-latest-920138984":"PreferredAntiAffinity"}},"resource":{},"instance":{}},"initCallTimeOut":65,"resourceSelector":{"resource.owner":"e6040c04-0000-4000-80ad-7ac099e8361d"},"nodeSelector":{"node-role":"system-parasitic"}},"createOptions":{"FUNCTION_KEY_Note":"244177614494719500/0@default@testcustom1024001/latest","DELEGATE_VOLUMES":"[{\"name\":\"bi-logs\",\"hostPath\":{\"path\":\"/mnt/daemonset/bi/244177614494719500\",\"type\":\"DirectoryOrCreate\"}},{\"name\":\"cgroup-memory\",\"hostPath\":{\"path\":\"/sys/fs/cgroup/memory/kubepods/burstable\"}},{\"name\":\"docker-socket\",\"hostPath\":{\"path\":\"/var/run/docker.sock\"}},{\"name\":\"docker-rootdir\",\"hostPath\":{\"path\":\"/var/lib/docker\"}}]","init_call_timeout":"65","GRACEFUL_SHUTDOWN_TIME":"0","DELEGATE_INIT_CONTAINERS":"null","DELEGATE_SIDECARS":"null","DELEGATE_INIT_VOLUME_MOUNTS":"[{\"name\":\"bi-log\",\"mountPath\":\"/opt/logs/caas/bi\",\"subPathExpr\":\"$(POD_NAME)/opt/logs/caas/bi\"}]","ConcurrentNum":"1","DELEGATE_POD_LABELS":"{\"funcName\":\"testcustom1024001\",\"instanceType\":\"reserved\",\"isPoolPod\":\"false\",\"securityGroup\":\"244177614494719500\",\"serviceID\":\"default\",\"standard\":\"1000-1024-fusion\",\"tenantID\":\"244177614494719500\",\"version\":\"latest\"}","DELEGATE_ENCRYPT":"{\"accessKey\":\"\",\"authToken\":\"\",\"encrypted_user_data\":\"\",\"environment\":\"\",\"secretKey\":\"\"}","DELEGATE_CONTAINER":"{\"image\":\"swr.image.com/xxx/custom_test_image:20240103180957\",\"env\":[{\"name\":\"INVOKE_TYPE\",\"value\":\"faas\"},{\"name\":\"x-system-tenantId\",\"value\":\"244177614494719500\"},{\"name\":\"x-system-functionName\",\"value\":\"testcustom1024001\"},{\"name\":\"x-system-functionVersion\",\"value\":\"latest\"},{\"name\":\"x-system-region\",\"value\":\"guiyang\"},{\"name\":\"x-system-clusterID\",\"value\":\"cluster001\"},{\"name\":\"x-system-NODE_IP\",\"valueFrom\":{\"fieldRef\":{\"apiVersion\":\"v1\",\"fieldPath\":\"status.hostIP\"}}},{\"name\":\"x-system-podName\",\"valueFrom\":{\"fieldRef\":{\"apiVersion\":\"v1\",\"fieldPath\":\"metadata.name\"}}},{\"name\":\"POD_IP\",\"valueFrom\":{\"fieldRef\":{\"apiVersion\":\"v1\",\"fieldPath\":\"status.podIP\"}}},{\"name\":\"HOST_IP\",\"valueFrom\":{\"fieldRef\":{\"apiVersion\":\"v1\",\"fieldPath\":\"status.hostIP\"}}},{\"name\":\"PodName\",\"valueFrom\":{\"fieldRef\":{\"apiVersion\":\"v1\",\"fieldPath\":\"metadata.name\"}}},{\"name\":\"POD_ID\",\"valueFrom\":{\"fieldRef\":{\"apiVersion\":\"v1\",\"fieldPath\":\"metadata.uid\"}}}],\"command\":null,\"args\":null,\"uid\":-1,\"gid\":-1,\"volumeMounts\":[{\"name\":\"stdlogs\",\"mountPath\":\"/opt/logs/caas/user\"},{\"name\":\"npulogs\",\"mountPath\":\"/opt/logs/caas/dataflow\"},{\"name\":\"data-system-logs\",\"mountPath\":\"/opt/logs/caas/dataSystem\"},{\"name\":\"custom-user-logs\",\"mountPath\":\"/opt/logs\"},{\"name\":\"bi-log\",\"mountPath\":\"/opt/logs/caas/bi\",\"subPathExpr\":\"$(x-system-podName)/opt/logs/caas/bi\"},{\"name\":\"datasystem-socket\",\"mountPath\":\"/home/uds\"}],\"runtime_graceful_shutdown\":{\"maxShutdownTimeout\":0}}","DELEGATE_VOLUME_MOUNTS":"[{\"name\":\"cgroup-memory\",\"mountPath\":\"/runtime/memory\",\"subPathExpr\":\"pod$(POD_ID)\"},{\"name\":\"docker-socket\",\"mountPath\":\"/var/run/docker.sock\"},{\"name\":\"docker-rootdir\",\"mountPath\":\"/var/lib/docker\"},{\"name\":\"sts-config\",\"mountPath\":\"/opt/certs/HMSClientCloudAccelerateService/HMSCaaSYuanRongWorker/HMSCaaSYuanRongWorker/apple/a\",\"subPath\":\"a\"},{\"name\":\"sts-config\",\"mountPath\":\"/opt/certs/HMSClientCloudAccelerateService/HMSCaaSYuanRongWorker/HMSCaaSYuanRongWorker/boy/b\",\"subPath\":\"b\"},{\"name\":\"sts-config\",\"mountPath\":\"/opt/certs/HMSClientCloudAccelerateService/HMSCaaSYuanRongWorker/HMSCaaSYuanRongWorker/cat/c\",\"subPath\":\"c\"},{\"name\":\"sts-config\",\"mountPath\":\"/opt/certs/HMSClientCloudAccelerateService/HMSCaaSYuanRongWorker/HMSCaaSYuanRongWorker/dog/d\",\"subPath\":\"d\"},{\"name\":\"sts-config\",\"mountPath\":\"/opt/certs/HMSClientCloudAccelerateService/HMSCaaSYuanRongWorker/HMSCaaSYuanRongWorker.ini\",\"subPath\":\"HMSCaaSYuanRongWorker.ini\"},{\"name\":\"sts-config\",\"mountPath\":\"/opt/certs/HMSClientCloudAccelerateService/HMSCaaSYuanRongWorker/HMSCaaSYuanRongWorker.sts.p12\",\"subPath\":\"HMSCaaSYuanRongWorker.sts.p12\"}]","DELEGATE_POD_SECCOMP_PROFILE":"{\"type\":\"RuntimeDefault\"}","DELEGATE_CONTAINER_ID":"82920897c041d93210033667fdc929695c24e26a592a874d6981a23626f58fef","DELEGATE_DOWNLOAD":"{\"appId\":\"\",\"bucketId\":\"\",\"objectId\":\"\",\"bucketUrl\":\"\",\"code_type\":\"\",\"code_url\":\"\",\"code_filename\":\"\",\"func_code\":{\"file\":\"\",\"link\":\"\"}}","SCHEDULER_ID_Note":"6a0ea93f-37e6-4101-8000-00000010049b"},"labels":["yyrk2441-0-default-testcustom1024001-latest-920138984"],"instanceStatus":{"code":2,"msg":"creating"},"parentID":"6a0ea93f-37e6-4101-8000-00000010049b","parentFunctionProxyAID":"10.41.5.115-LocalSchedInstanceCtrlActor@10.41.5.115:22423","storageType":"local","scheduleTimes":1,"deployTimes":1,"args":[{"value":"eyJmdW5jTWV0YURhdGEiOnsibGF5ZXJzIjpudWxsLCJuYW1lIjoiMEBkZWZhdWx0QHRlc3RjdXN0b20xMDI0MDAxIiwiZGVzY3JpcHRpb24iOiJ0ZXN0Y3VzdG9tMTAyNDAwMSIsImZ1bmN0aW9uVXJuIjoic246Y246eXJrOjI0NDE3NzYxNDQ5NDcxOTUwMDpmdW5jdGlvbjowQGRlZmF1bHRAdGVzdGN1c3RvbTEwMjQwMDEiLCJ0ZW5hbnRJZCI6IiIsInRhZ3MiOm51bGwsImZ1bmN0aW9uVXBkYXRlVGltZSI6IiIsImZ1bmN0aW9uVmVyc2lvblVybiI6InNuOmNuOnlyazoyNDQxNzc2MTQ0OTQ3MTk1MDA6ZnVuY3Rpb246MEBkZWZhdWx0QHRlc3RjdXN0b20xMDI0MDAxOmxhdGVzdCIsInJldmlzaW9uSWQiOiIiLCJjb2RlU2l6ZSI6MCwiY29kZVNoYTUxMiI6IiIsImhhbmRsZXIiOiIiLCJydW50aW1lIjoiY3VzdG9tIGltYWdlIiwidGltZW91dCI6NjAsInZlcnNpb24iOiJsYXRlc3QiLCJkZWFkTGV0dGVyQ29uZmlnIjoiIiwiYnVzaW5lc3NJZCI6InlyayIsImZ1bmN0aW9uVHlwZSI6IiIsImZ1bmNfaWQiOiIxNjIwMzE0OTMxNjI3Nzk1MjQ1IiwiZnVuY19uYW1lIjoidGVzdGN1c3RvbTEwMjQwMDEiLCJkb21haW5faWQiOiIiLCJwcm9qZWN0X25hbWUiOiIiLCJzZXJ2aWNlIjoiZGVmYXVsdCIsImRlcGVuZGVuY2llcyI6IiIsImVuYWJsZV9jbG91ZF9kZWJ1ZyI6IiIsImlzU3RhdGVmdWxGdW5jdGlvbiI6ZmFsc2UsImlzQnJpZGdlRnVuY3Rpb24iOmZhbHNlLCJpc1N0cmVhbUVuYWJsZSI6ZmFsc2UsInR5cGUiOiIiLCJlbmFibGVfYXV0aF9pbl9oZWFkZXIiOmZhbHNlLCJkbnNfZG9tYWluX2NmZyI6bnVsbCwidnBjVHJpZ2dlckltYWdlIjoiIiwic3RhdGVDb25maWciOnsibGlmZUN5Y2xlIjoiIn19LCJzM01ldGFEYXRhIjp7ImFwcElkIjoiIiwiYnVja2V0SWQiOiIiLCJvYmplY3RJZCI6IiIsImJ1Y2tldFVybCI6IiIsImNvZGVfdHlwZSI6IiIsImNvZGVfdXJsIjoiIiwiY29kZV9maWxlbmFtZSI6IiIsImZ1bmNfY29kZSI6eyJmaWxlIjoiIiwibGluayI6IiJ9fSwiZW52TWV0YURhdGEiOnsiZW52aXJvbm1lbnQiOiIiLCJlbmNyeXB0ZWRfdXNlcl9kYXRhIjoiIn0sInN0c01ldGFEYXRhIjp7ImVuYWJsZVN0cyI6ZmFsc2V9LCJyZXNvdXJjZU1ldGFEYXRhIjp7ImNwdSI6MTAwMCwibWVtb3J5IjoxMDI0LCJncHVfbWVtb3J5IjowLCJlbmFibGVfZHluYW1pY19tZW1vcnkiOmZhbHNlLCJjdXN0b21SZXNvdXJjZXMiOiIiLCJlbmFibGVfdG1wX2V4cGFuc2lvbiI6ZmFsc2UsImVwaGVtZXJhbF9zdG9yYWdlIjo2MTQ0LCJDdXN0b21SZXNvdXJjZXNTcGVjIjoiIn0sImluc3RhbmNlTWV0YURhdGEiOnsibWF4SW5zdGFuY2UiOjIsIm1pbkluc3RhbmNlIjowLCJjb25jdXJyZW50TnVtIjoxLCJpbnN0YW5jZVR5cGUiOiIiLCJpZGxlTW9kZSI6ZmFsc2V9LCJleHRlbmRlZE1ldGFEYXRhIjp7ImltYWdlX25hbWUiOiIiLCJyb2xlIjp7Inhyb2xlIjoiIiwiYXBwX3hyb2xlIjoiIn0sImZ1bmNfdnBjIjpudWxsLCJlbmRwb2ludF90ZW5hbnRfdnBjIjpudWxsLCJtb3VudF9jb25maWciOm51bGwsInN0cmF0ZWd5X2NvbmZpZyI6eyJjb25jdXJyZW5jeSI6MH0sImV4dGVuZF9jb25maWciOiIiLCJpbml0aWFsaXplciI6eyJpbml0aWFsaXplcl9oYW5kbGVyIjoiIiwiaW5pdGlhbGl6ZXJfdGltZW91dCI6NjB9LCJoZWFydGJlYXQiOnsiaGVhcnRiZWF0X2hhbmRsZXIiOiIifSwiZW50ZXJwcmlzZV9wcm9qZWN0X2lkIjoiIiwibG9nX3Rhbmtfc2VydmljZSI6eyJsb2dHcm91cElkIjoiIiwibG9nU3RyZWFtSWQiOiIifSwidHJhY2luZ19jb25maWciOnsidHJhY2luZ19hayI6IiIsInRyYWNpbmdfc2siOiIiLCJwcm9qZWN0X25hbWUiOiIifSwiY3VzdG9tX2NvbnRhaW5lcl9jb25maWciOnsiY29udHJvbF9wYXRoIjoiZXZlbnQiLCJpbWFnZSI6InN3ci5jbi1zb3V0aHdlc3QtMi5teWh1YXdlaWNsb3VkLmNvbS90MDE0LWNvbS5odWF3ZWkuYWdjL2N1c3RvbV90ZXN0X2ltYWdlOjEzLjAuNS4yMDAuMjAyNDAxMDMxODA5NTciLCJjb21tYW5kIjpudWxsLCJhcmdzIjpudWxsLCJ3b3JraW5nX2RpciI6IiIsInVpZCI6LTEsImdpZCI6LTF9LCJhc3luY19jb25maWdfbG9hZGVkIjpmYWxzZSwicmVzdG9yZV9ob29rIjp7fSwibmV0d29ya19jb250cm9sbGVyIjp7ImRpc2FibGVfcHVibGljX25ldHdvcmsiOmZhbHNlLCJ0cmlnZ2VyX2FjY2Vzc192cGNzIjpudWxsfSwidXNlcl9hZ2VuY3kiOnsiYWNjZXNzS2V5IjoiIiwic2VjcmV0S2V5IjoiIiwidG9rZW4iOiIifSwiY3VzdG9tX2ZpbGViZWF0X2NvbmZpZyI6eyJzaWRlY2FyQ29uZmlnSW5mbyI6bnVsbCwiY3B1IjowLCJtZW1vcnkiOjAsInZlcnNpb24iOiIiLCJpbWFnZUFkZHJlc3MiOiIifSwiY3VzdG9tX2hlYWx0aF9jaGVjayI6eyJ0aW1lb3V0U2Vjb25kcyI6MCwicGVyaW9kU2Vjb25kcyI6MCwiZmFpbHVyZVRocmVzaG9sZCI6MH0sImR5bmFtaWNfY29uZmlnIjp7ImVuYWJsZWQiOmZhbHNlLCJ1cGRhdGVfdGltZSI6IjIwMjQwMTIxMTUwMzQwNzgyIiwiY29uZmlnX2NvbnRlbnQiOltdfSwicnVudGltZV9ncmFjZWZ1bF9zaHV0ZG93biI6eyJtYXhTaHV0ZG93blRpbWVvdXQiOjB9LCJyYXNwX2NvbmZpZyI6eyJpbml0LWltYWdlIjoiIiwicmFzcC1pbWFnZSI6IiIsInJhc3Atc2VydmVyLWlwIjoiIiwicmFzcC1zZXJ2ZXItcG9ydCI6IiJ9fX0="},{"value":"eyJwb3J0Ijo4MDAwLCJpbml0Um91dGUiOiIiLCJjYWxsUm91dGUiOiJpbnZva2UifQ=="},{"value":"eyJzY2hlZHVsZXJGdW5jS2V5IjoiMTIzNDU2Nzg5MDEyMzQ1NjEyMzQ1Njc4OTAxMjM0NTYvMC1zeXN0ZW0tZmFhc3NjaGVkdWxlci8kbGF0ZXN0Iiwic2NoZWR1bGVySURMaXN0IjpbIjZhMGVhOTNmLTM3ZTYtNDEwMS04MDAwLTAwMDAwMDEwMDQ5YiIsIjcxZTQ3MTAxLTAwMDAtNDAwMC04MGQ5LTMxNDNhYzVhMzQ3NCJdfQ=="},{"value":"eyJjdXN0b21GYXVsdFJlcG9ydENvbmZpZyI6eyJjdXN0b21GYXVsdFJlcG9ydEVuYWJsZSI6dHJ1ZSwieHB1TWdtdENvbmZpZyI6eyJhY2Nlc3NLZXkiOiIyMjQ1MUQ4MzgwMUQzQTE5RjMzODg4QzE3QzY3ODcyQiIsInNlY3JldEtleSI6IkVOQyhrZXk9c2VydmljZWtlaywgdmFsdWU9NkI2RDczNzYzMDMwMDAwMTAxNkNGQ0VBRDc0REI1NjFGRTc2RDM2RkIxRkEyRTJCRThFOUM3OTNGRUQ5NjY1ODRFN0EwNTc1Q0IyNENCNUMzMTg3QkQ4M0ZEQkQxMzM2M0JGMDA5NUEwOEFGMTgxRDUxQ0UwMUNDRUYwMjA2QUY3M0RBMTAyQTBBRjc5ODFGNzRBNjE4RTE1MTIyMTFFNEE4MUI0MDdEMUE5NjQyN0JGMjI1NzMyQUUwQzQyRTgxOEU2NjMyM0FCOCkiLCJ0aW1lb3V0IjoxMH19LCJhbGFybUNvbmZpZyI6eyJlbmFibGVBbGFybSI6dHJ1ZSwiYWxhcm1Mb2dDb25maWciOnsiZmlsZXBhdGgiOiIvaG9tZS9zbnVzZXIvYWxhcm1zIiwibGV2ZWwiOiJJbmZvIiwidGljayI6MCwiZmlyc3QiOjAsInRoZXJlYWZ0ZXIiOjAsInRyYWNpbmciOmZhbHNlLCJkaXNhYmxlIjpmYWxzZSwic2luZ2xlc2l6ZSI6NTAwLCJ0aHJlc2hvbGQiOjN9LCJ4aWFuZ1l1bkZvdXJDb25maWciOnsic2l0ZSI6ImNuX2Rldl9kZWZhdWx0IiwidGVuYW50SUQiOiJUMDE0IiwiYXBwbGljYXRpb25JRCI6ImNvbS5odWF3ZWkuYWdjIiwic2VydmljZUlEIjoiY29tLmh1YXdlaS5obXNjbGllbnRjbG91ZGFjY2VsZXJhdGVzZXJ2aWNlIn0sIm1pbkluc1N0YXJ0SW50ZXJ2YWwiOjE1LCJtaW5JbnNDaGVja0ludGVydmFsIjoxNX0sInN0c1NlcnZlckNvbmZpZyI6eyJkb21haW4iOiIxMC4zNC4yNTUuMTc2OjgwODAiLCJwYXRoIjoiL29wdC9odWF3ZWkvY2VydHMvSE1TQ2xpZW50Q2xvdWRBY2NlbGVyYXRlU2VydmljZS9ITVNDYWFTWXVhblJvbmdXb3JrZXIvSE1TQ2FhU1l1YW5Sb25nV29ya2VyLmluaSJ9LCJjbHVzdGVyTmFtZSI6Imd5MV9hdXRvX2F6NSIsImRpc2tNb25pdG9yRW5hYmxlIjpmYWxzZX0="}],"version":"2","dataSystemHost":"10.41.5.115"}`)
		ins := &commonTypes.InstanceSpecification{}
		err := json.Unmarshal(b, ins)
		if err != nil {
			fmt.Println(err)
		}
		resourceSpecification := ConvertInstanceResource(ins.Resources)
		convey.So("cpu-1000-mem-1024-ephemeral-storage-6144", convey.ShouldEqual, resourceSpecification.String())
	})
	convey.Convey("set resources", t, func() {
		resources := commonTypes.Resources{
			Resources: map[string]commonTypes.Resource{
				constant.ResourceCPUName: commonTypes.Resource{
					Scalar: commonTypes.ValueScalar{
						Value: 300,
					},
				},
				constant.ResourceMemoryName: commonTypes.Resource{
					Scalar: commonTypes.ValueScalar{
						Value: 300,
					},
				},
			},
		}
		ConvertInstanceResource(resources)
	})
}

func TestAddAffinityCPU(t *testing.T) {
	type args struct {
		crName            string
		schedulingOptions *types.SchedulingOptions
		resSpec           *resspeckey.ResourceSpecification
		affinityType      api.AffinityType
	}
	tests := []struct {
		name string
		args args
	}{
		{"case1", args{
			crName:            "yyrk1234-0-yrservice-test-image-env-call-latest-670698364",
			schedulingOptions: &types.SchedulingOptions{},
			resSpec:           &resspeckey.ResourceSpecification{CPU: 300, Memory: 128},
			affinityType:      api.PreferredAntiAffinity,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AddAffinityCPU(tt.args.crName, tt.args.schedulingOptions, tt.args.resSpec, tt.args.affinityType)
		})
	}
}

func TestGetNpuInstanceType(t *testing.T) {
	tests := []struct {
		name              string
		delegateContainer string
		envExist          bool
		envValue          string
	}{
		{
			name:              "376T",
			delegateContainer: "{\"image\":\"swr.image.com/ct:20231213175223\",\"env\":[{\"name\":\"INVOKE_TYPE\",\"value\":\"faas\"},{\"name\":\"X_SYSTEM_NODE_INSTANCE_TYPE\",\"value\":\"376T\"},{\"name\":\"x-system-tenantId\",\"value\":\"172120022620408014\"},{\"name\":\"x-system-functionName\",\"value\":\"portraitHDAlg_binning\"},{\"name\":\"x-system-functionVersion\",\"value\":\"28\"},{\"name\":\"x-system-region\",\"value\":\"guiyang\"},{\"name\":\"x-system-clusterID\",\"value\":\"cluster001\"},{\"name\":\"x-system-NODE_IP\",\"valueFrom\":{\"fieldRef\":{\"apiVersion\":\"v1\",\"fieldPath\":\"status.hostIP\"}}},{\"name\":\"x-system-podName\",\"valueFrom\":{\"fieldRef\":{\"apiVersion\":\"v1\",\"fieldPath\":\"metadata.name\"}}},{\"name\":\"POD_IP\",\"valueFrom\":{\"fieldRef\":{\"apiVersion\":\"v1\",\"fieldPath\":\"status.podIP\"}}},{\"name\":\"HOST_IP\",\"valueFrom\":{\"fieldRef\":{\"apiVersion\":\"v1\",\"fieldPath\":\"status.hostIP\"}}},{\"name\":\"PodName\",\"valueFrom\":{\"fieldRef\":{\"apiVersion\":\"v1\",\"fieldPath\":\"metadata.name\"}}},{\"name\":\"POD_ID\",\"valueFrom\":{\"fieldRef\":{\"apiVersion\":\"v1\",\"fieldPath\":\"metadata.uid\"}}},{\"name\":\"KMS_CREDENTIAL\",\"value\":\"ENC(key=servicekek, value=6B6D7376303000010160341E9C437627D9819DCFB412C3282AF8FC32BD9D4E7DAE64D40FB98A240EF2CB089777FB81E52D5513E1235F3E7EAC5677ECDC783196E36ED04D20ADD5847EB1783C041FEC169709EB56E085696F79751E9F113FF7671FDED44CB6F1FFC191853DA271E9C247553D7AC2C8C61317FF974C5C8C82B7EE86EEC56E9E11BE03712C89B7DA1190F06229506BFE3519CDBECC555AC45991384B88998754182DDDEF3203CC3E6C090CBA8ABCAC4504F83D34021DCA8188162DD0FC7A5453742CD531E2053161CF813754A796C4209AD526E27478B377EF5B6F97DCFE00C029CEAE034486FEBFF437B6C1DED3C5E54DB4750B5FF918BFB45E1747D246497D78055BAE0BB5570715545160E51AAE016F73D555F229123CD1F6F85C21D3B1C8024145CC28981465FFDDCF0A6439C7BF6AF08E6A43BDFA67650B562C5983563353B70CD5DA03B7E63C98341C1CCF92F922FC3848E95F8E8246204C723A52E6454446D421D438FE9487C1B4236AC1421D1AC6923D8C989316DF6F667298A686E4FC8A7F1ADB12576BDA00284BC5CA08C4998D58F6D8F4D7FE1007278338C3F27594E4BD6FE8CBB9419AD4F55016BA2FC3F814BDC5D09C5C6E3157E13B5887560B695A71B3C2C77A2D48CE5E6892A6E470554AA759F725BAF294526A7DB33B707A3EA1FD68A27193881BA150B492EE27A4834C7ECE2814DA6AD11E1C8E5DC82D1E2717BEFCCF724ED576C6B0FF8939745C1C7A26857DD31DAA393FFAB58C0E0F3052010BEB5F4E570D8F6C822D661D717036F5E16A556B180D1695D8C8E7F94CB2297C8566A2324E3C9BF92C20B805A381FACBB5AF8729AAEE8A46DCDAB7B66859D36F29742BB5869E7D1BE84BF01FF2D4AFDA543CCC07E5CB1893C04A1D1230D9ED7BA039BFE1448CCA046BDB641B92BA08861D22990BB88072AC4AB0A757C97F8E746B6AA2426D96517F0A602EACD5A494FBFAC40DBD316B8392C68C5A625943B5954FEDCA2C53D2B974CB5522FFDB1CD270E8099D6E0C4F63B6F236CD347966358173BBA033CA1BA75C6BF8FBE413F31A7E3FD8CD3FD494AD79B4708C586684669F8D338B08CF)\"},{\"name\":\"DATASYSTEM_OAUTH_CLIENTSECRET\",\"value\":\"ENC(key=servicekek, value=6B6D73763030000101C176169F064D5AE27A05363FCE387AEC060AA381D5CBCE4C5985171E4BB84BCD7C32FB354A654F0DB146D25A781C9B2A3670A1C6752289DDEACCA8399222697895C0C300D75C737A83408BEF15B785B4D9A93D811EB1A5E7EE124B4D)\"},{\"name\":\"CLOUDREDIS_PASSWORD\",\"value\":\"ENC(key=servicekek, value=6B6D73763030000101B7DC4A0E7AE98843924440986EF4FC665505A5C08F47E914A84944A3C08BB9DB1363663CED0E)\"},{\"name\":\"AGC_CLIENT_CONFIG\",\"value\":\"ENC(key=servicekek, value=6B6D73763030000101DE8BE9226C9C2BD1FB8D7A8AD87B8E7471929F57B298ED3B4D624ED94A0FFFA1008EAE12A2AA132B67BE6B4E24A764EAD8837C4E61E15153361190E2FC67DD47A1AFCF01D094822E1A18765F01D04BC916814166BAF9B3BD7B239C4D350E17F30DCFF2C1333A8D4798211E39E93963ADBB7BB003E2BF7DB4F6A4AA11D18602FE41A6780974B09514DD8668E66697801A387F01242CB5E46A2A4DFD13F4BE944DF2896EEAA4A705A379B590E442E59BBB96448DC9C853885BB2F87C6076317BAC50D5C38416A4631C8A3B71FBEDD401D9E5E72A3DD1C964CFBA10F2696C10886FB4291EDAF17702CDAEF4EBED3536DE239BA95176EC2A681F344318FAA78E084CFFF9402F2409F274EB375F2DAF4E37CA3A3C272DF2487A5CDA2E365A53679097C1D84498DF2831A5DB)\"},{\"name\":\"RANK_TABLE_FILE\",\"value\":\"/opt/config/ascend_config/ranktable_file.json\"}],\"command\":null,\"args\":null,\"uid\":-1,\"gid\":-1,\"volumeMounts\":[{\"name\":\"stdlogs\",\"mountPath\":\"/opt/logs/caas/user\"},{\"name\":\"data-system-logs\",\"mountPath\":\"/opt/logs/caas/dataSystem\"},{\"name\":\"custom-user-logs\",\"mountPath\":\"/opt/logs\"},{\"name\":\"bi-logs\",\"mountPath\":\"/opt/logs/caas/bi\",\"subPathExpr\":\"$(x-system-podName)/opt/logs/caas/bi\"},{\"name\":\"datasystem-socket\",\"mountPath\":\"/home/uds\"},{\"name\":\"dynamic-config\",\"mountPath\":\"/opt/dynamic-config\"},{\"name\":\"ascend-driver-path\",\"readOnly\":true,\"mountPath\":\"/usr/local/Ascend/driver\"},{\"name\":\"ascend-config\",\"mountPath\":\"/opt/config/ascend_config\"},{\"name\":\"ascend-npu-smi\",\"mountPath\":\"/usr/local/sbin/npu-smi\"},{\"name\":\"runtime-certs-volume\",\"mountPath\":\"/opt/certs/MediaCloudEnhanceService/MediaCloudEnhancePortraitHDService/\"},{\"name\":\"runtime-sts-config\",\"mountPath\":\"/opt/certs/MediaCloudEnhanceService/MediaCloudEnhancePortraitHDService/MediaCloudEnhancePortraitHDService/apple/a\",\"subPath\":\"a\"},{\"name\":\"runtime-sts-config\",\"mountPath\":\"/opt/certs/MediaCloudEnhanceService/MediaCloudEnhancePortraitHDService/MediaCloudEnhancePortraitHDService/boy/b\",\"subPath\":\"b\"},{\"name\":\"runtime-sts-config\",\"mountPath\":\"/opt/certs/MediaCloudEnhanceService/MediaCloudEnhancePortraitHDService/MediaCloudEnhancePortraitHDService/cat/c\",\"subPath\":\"c\"},{\"name\":\"runtime-sts-config\",\"mountPath\":\"/opt/certs/MediaCloudEnhanceService/MediaCloudEnhancePortraitHDService/MediaCloudEnhancePortraitHDService/dog/d\",\"subPath\":\"d\"},{\"name\":\"runtime-sts-config\",\"mountPath\":\"/opt/certs/MediaCloudEnhanceService/MediaCloudEnhancePortraitHDService/MediaCloudEnhancePortraitHDService.ini\",\"subPath\":\"MediaCloudEnhancePortraitHDService.ini\"},{\"name\":\"runtime-sts-config\",\"mountPath\":\"/opt/certs/MediaCloudEnhanceService/MediaCloudEnhancePortraitHDService/MediaCloudEnhancePortraitHDService.sts.p12\",\"subPath\":\"MediaCloudEnhancePortraitHDService.sts.p12\"}],\"runtime_graceful_shutdown\":{\"maxShutdownTimeout\":3600}}",
			envExist:          true,
			envValue:          "376T",
		}, {
			name:              "376T 1",
			delegateContainer: "{\"image\":\"swr.image.com/ct:20231213175223\",\"env\":[{\"name\":\"INVOKE_TYPE\",\"value\":\"faas\"},{\"name\":\"x-system-tenantId\",\"value\":\"172120022620408014\"},{\"name\":\"x-system-functionName\",\"value\":\"portraitHDAlg_binning\"},{\"name\":\"x-system-functionVersion\",\"value\":\"28\"},{\"name\":\"x-system-region\",\"value\":\"guiyang\"},{\"name\":\"x-system-clusterID\",\"value\":\"cluster001\"},{\"name\":\"x-system-NODE_IP\",\"valueFrom\":{\"fieldRef\":{\"apiVersion\":\"v1\",\"fieldPath\":\"status.hostIP\"}}},{\"name\":\"x-system-podName\",\"valueFrom\":{\"fieldRef\":{\"apiVersion\":\"v1\",\"fieldPath\":\"metadata.name\"}}},{\"name\":\"POD_IP\",\"valueFrom\":{\"fieldRef\":{\"apiVersion\":\"v1\",\"fieldPath\":\"status.podIP\"}}},{\"name\":\"HOST_IP\",\"valueFrom\":{\"fieldRef\":{\"apiVersion\":\"v1\",\"fieldPath\":\"status.hostIP\"}}},{\"name\":\"PodName\",\"valueFrom\":{\"fieldRef\":{\"apiVersion\":\"v1\",\"fieldPath\":\"metadata.name\"}}},{\"name\":\"POD_ID\",\"valueFrom\":{\"fieldRef\":{\"apiVersion\":\"v1\",\"fieldPath\":\"metadata.uid\"}}},{\"name\":\"KMS_CREDENTIAL\",\"value\":\"ENC(key=servicekek, value=6B6D7376303000010160341E9C437627D9819DCFB412C3282AF8FC32BD9D4E7DAE64D40FB98A240EF2CB089777FB81E52D5513E1235F3E7EAC5677ECDC783196E36ED04D20ADD5847EB1783C041FEC169709EB56E085696F79751E9F113FF7671FDED44CB6F1FFC191853DA271E9C247553D7AC2C8C61317FF974C5C8C82B7EE86EEC56E9E11BE03712C89B7DA1190F06229506BFE3519CDBECC555AC45991384B88998754182DDDEF3203CC3E6C090CBA8ABCAC4504F83D34021DCA8188162DD0FC7A5453742CD531E2053161CF813754A796C4209AD526E27478B377EF5B6F97DCFE00C029CEAE034486FEBFF437B6C1DED3C5E54DB4750B5FF918BFB45E1747D246497D78055BAE0BB5570715545160E51AAE016F73D555F229123CD1F6F85C21D3B1C8024145CC28981465FFDDCF0A6439C7BF6AF08E6A43BDFA67650B562C5983563353B70CD5DA03B7E63C98341C1CCF92F922FC3848E95F8E8246204C723A52E6454446D421D438FE9487C1B4236AC1421D1AC6923D8C989316DF6F667298A686E4FC8A7F1ADB12576BDA00284BC5CA08C4998D58F6D8F4D7FE1007278338C3F27594E4BD6FE8CBB9419AD4F55016BA2FC3F814BDC5D09C5C6E3157E13B5887560B695A71B3C2C77A2D48CE5E6892A6E470554AA759F725BAF294526A7DB33B707A3EA1FD68A27193881BA150B492EE27A4834C7ECE2814DA6AD11E1C8E5DC82D1E2717BEFCCF724ED576C6B0FF8939745C1C7A26857DD31DAA393FFAB58C0E0F3052010BEB5F4E570D8F6C822D661D717036F5E16A556B180D1695D8C8E7F94CB2297C8566A2324E3C9BF92C20B805A381FACBB5AF8729AAEE8A46DCDAB7B66859D36F29742BB5869E7D1BE84BF01FF2D4AFDA543CCC07E5CB1893C04A1D1230D9ED7BA039BFE1448CCA046BDB641B92BA08861D22990BB88072AC4AB0A757C97F8E746B6AA2426D96517F0A602EACD5A494FBFAC40DBD316B8392C68C5A625943B5954FEDCA2C53D2B974CB5522FFDB1CD270E8099D6E0C4F63B6F236CD347966358173BBA033CA1BA75C6BF8FBE413F31A7E3FD8CD3FD494AD79B4708C586684669F8D338B08CF)\"},{\"name\":\"DATASYSTEM_OAUTH_CLIENTSECRET\",\"value\":\"ENC(key=servicekek, value=6B6D73763030000101C176169F064D5AE27A05363FCE387AEC060AA381D5CBCE4C5985171E4BB84BCD7C32FB354A654F0DB146D25A781C9B2A3670A1C6752289DDEACCA8399222697895C0C300D75C737A83408BEF15B785B4D9A93D811EB1A5E7EE124B4D)\"},{\"name\":\"CLOUDREDIS_PASSWORD\",\"value\":\"ENC(key=servicekek, value=6B6D73763030000101B7DC4A0E7AE98843924440986EF4FC665505A5C08F47E914A84944A3C08BB9DB1363663CED0E)\"},{\"name\":\"AGC_CLIENT_CONFIG\",\"value\":\"ENC(key=servicekek, value=6B6D73763030000101DE8BE9226C9C2BD1FB8D7A8AD87B8E7471929F57B298ED3B4D624ED94A0FFFA1008EAE12A2AA132B67BE6B4E24A764EAD8837C4E61E15153361190E2FC67DD47A1AFCF01D094822E1A18765F01D04BC916814166BAF9B3BD7B239C4D350E17F30DCFF2C1333A8D4798211E39E93963ADBB7BB003E2BF7DB4F6A4AA11D18602FE41A6780974B09514DD8668E66697801A387F01242CB5E46A2A4DFD13F4BE944DF2896EEAA4A705A379B590E442E59BBB96448DC9C853885BB2F87C6076317BAC50D5C38416A4631C8A3B71FBEDD401D9E5E72A3DD1C964CFBA10F2696C10886FB4291EDAF17702CDAEF4EBED3536DE239BA95176EC2A681F344318FAA78E084CFFF9402F2409F274EB375F2DAF4E37CA3A3C272DF2487A5CDA2E365A53679097C1D84498DF2831A5DB)\"},{\"name\":\"RANK_TABLE_FILE\",\"value\":\"/opt/config/ascend_config/ranktable_file.json\"}],\"command\":null,\"args\":null,\"uid\":-1,\"gid\":-1,\"volumeMounts\":[{\"name\":\"stdlogs\",\"mountPath\":\"/opt/logs/caas/user\"},{\"name\":\"data-system-logs\",\"mountPath\":\"/opt/logs/caas/dataSystem\"},{\"name\":\"custom-user-logs\",\"mountPath\":\"/opt/logs\"},{\"name\":\"bi-logs\",\"mountPath\":\"/opt/logs/caas/bi\",\"subPathExpr\":\"$(x-system-podName)/opt/logs/caas/bi\"},{\"name\":\"datasystem-socket\",\"mountPath\":\"/home/uds\"},{\"name\":\"dynamic-config\",\"mountPath\":\"/opt/dynamic-config\"},{\"name\":\"ascend-driver-path\",\"readOnly\":true,\"mountPath\":\"/usr/local/Ascend/driver\"},{\"name\":\"ascend-config\",\"mountPath\":\"/opt/config/ascend_config\"},{\"name\":\"ascend-npu-smi\",\"mountPath\":\"/usr/local/sbin/npu-smi\"},{\"name\":\"runtime-certs-volume\",\"mountPath\":\"/opt/certs/MediaCloudEnhanceService/MediaCloudEnhancePortraitHDService/\"},{\"name\":\"runtime-sts-config\",\"mountPath\":\"/opt/certs/MediaCloudEnhanceService/MediaCloudEnhancePortraitHDService/MediaCloudEnhancePortraitHDService/apple/a\",\"subPath\":\"a\"},{\"name\":\"runtime-sts-config\",\"mountPath\":\"/opt/certs/MediaCloudEnhanceService/MediaCloudEnhancePortraitHDService/MediaCloudEnhancePortraitHDService/boy/b\",\"subPath\":\"b\"},{\"name\":\"runtime-sts-config\",\"mountPath\":\"/opt/certs/MediaCloudEnhanceService/MediaCloudEnhancePortraitHDService/MediaCloudEnhancePortraitHDService/cat/c\",\"subPath\":\"c\"},{\"name\":\"runtime-sts-config\",\"mountPath\":\"/opt/certs/MediaCloudEnhanceService/MediaCloudEnhancePortraitHDService/MediaCloudEnhancePortraitHDService/dog/d\",\"subPath\":\"d\"},{\"name\":\"runtime-sts-config\",\"mountPath\":\"/opt/certs/MediaCloudEnhanceService/MediaCloudEnhancePortraitHDService/MediaCloudEnhancePortraitHDService.ini\",\"subPath\":\"MediaCloudEnhancePortraitHDService.ini\"},{\"name\":\"runtime-sts-config\",\"mountPath\":\"/opt/certs/MediaCloudEnhanceService/MediaCloudEnhancePortraitHDService/MediaCloudEnhancePortraitHDService.sts.p12\",\"subPath\":\"MediaCloudEnhancePortraitHDService.sts.p12\"}],\"runtime_graceful_shutdown\":{\"maxShutdownTimeout\":3600}}",
			envExist:          false,
			envValue:          "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, v := GetNpuInstanceType(tt.delegateContainer)
			if ok != tt.envExist || v != tt.envValue {
				t.Errorf("actual ok: %v envValue: %v, expect ok: %v, envValue: %s", ok, v, tt.envExist, tt.envValue)
			}
		})
	}
}

func TestAppendInstanceTypeToInstanceResource(t *testing.T) {
	convey.Convey("test AppendInstanceTypeToInstanceResource", t, func() {
		convey.Convey("resSpec is nil", func() {
			AppendInstanceTypeToInstanceResource(nil, "")
		})
		convey.Convey("success", func() {
			resSpec := &resspeckey.ResourceSpecification{}
			AppendInstanceTypeToInstanceResource(resSpec, "NPU")
			convey.So(resSpec.CustomResourcesSpec["instanceType"], convey.ShouldEqual, "NPU")
		})
	})
}

func TestGetNpuTypeAndInstanceTypeFromStr(t *testing.T) {
	convey.Convey("Test GetNpuTypeAndInstanceTypeFromStr", t, func() {
		convey.Convey("Test GetNpuTypeAndInstanceTypeFromStr", func() {
			npuType, extraType := GetNpuTypeAndInstanceTypeFromStr(`{"huawei.com/ascend-1980":1}`,
				`{"instanceType":"280t"}`)
			convey.So(npuType, convey.ShouldEqual, "huawei.com/ascend-1980")
			convey.So(extraType, convey.ShouldEqual, "280t")
		})
	})
}

func TestAddNodeSelector(t *testing.T) {
	convey.Convey("Test AddNodeSelector", t, func() {
		convey.Convey("Test AddNodeSelector", func() {
			scheduleOption := &types.SchedulingOptions{Extension: make(map[string]string)}
			AddNodeSelector(map[string]string{"aaa": "123"}, scheduleOption, &resspeckey.ResourceSpecification{})
			convey.So(scheduleOption.Extension[utils.NodeSelectorKey], convey.ShouldEqual, `{"aaa":"123"}`)

			AddNodeSelector(map[string]string{"aaa": "123", "env": "123"},
				scheduleOption, &resspeckey.ResourceSpecification{})
			convey.So(scheduleOption.Extension[utils.NodeSelectorKey],
				convey.ShouldEqual, `{"aaa":"123","env":"123"}`)
		})
	})
}

func TestGetCreateTimeout(t *testing.T) {
	convey.Convey("test GetCreateTimeout", t, func() {
		convey.Convey("get create timeout", func() {
			timeout := GetCreateTimeout(&types.FunctionSpecification{})
			convey.So(timeout, convey.ShouldEqual, defaultCreateTimeout)
			timeout = GetCreateTimeout(&types.FunctionSpecification{
				ExtendedMetaData: commonTypes.ExtendedMetaData{
					Initializer: commonTypes.Initializer{
						Timeout: 50,
					},
				},
			})
			convey.So(timeout, convey.ShouldEqual, (50+constant.CommonExtraTimeout+constant.KernelScheduleTimeout)*time.Second)
			timeout = GetCreateTimeout(&types.FunctionSpecification{
				FuncMetaData: commonTypes.FuncMetaData{
					Runtime: types.CustomContainerRuntimeType,
				},
				ExtendedMetaData: commonTypes.ExtendedMetaData{
					Initializer: commonTypes.Initializer{
						Timeout: 50,
					},
				},
			})
			convey.So(timeout, convey.ShouldEqual, (50+constant.CustomImageExtraTimeout)*time.Second)
		})
	})
}

func TestGetLeaseInterval(t *testing.T) {
	convey.Convey("test GetLeaseInterval", t, func() {
		convey.Convey("get lease interval", func() {
			config.GlobalConfig.LeaseSpan = 50
			interval := GetLeaseInterval()
			convey.So(interval, convey.ShouldEqual, 500*time.Millisecond)
			config.GlobalConfig.LeaseSpan = 200
			interval = GetLeaseInterval()
			convey.So(interval, convey.ShouldEqual, 500*time.Millisecond)
			config.GlobalConfig.LeaseSpan = 600
			interval = GetLeaseInterval()
			convey.So(interval, convey.ShouldEqual, 600*time.Millisecond)
		})
	})
}

func TestBuildInstanceFromInsSpec(t *testing.T) {
	convey.Convey("test BuildInstanceFromInsSpec", t, func() {
		convey.Convey("get instance IP and port", func() {
			instance := BuildInstanceFromInsSpec(&commonTypes.InstanceSpecification{
				RuntimeAddress: "1.2.3.4:1234",
			}, nil)
			convey.So(instance.InstanceIP, convey.ShouldEqual, "1.2.3.4")
			convey.So(instance.InstancePort, convey.ShouldEqual, "1234")
		})
		convey.Convey("get instance type", func() {
			instance := BuildInstanceFromInsSpec(&commonTypes.InstanceSpecification{}, nil)
			convey.So(instance.InstanceType, convey.ShouldEqual, types.InstanceTypeUnknown)
			instance = BuildInstanceFromInsSpec(&commonTypes.InstanceSpecification{
				CreateOptions: map[string]string{
					types.InstanceTypeNote: string(types.InstanceTypeReserved),
				},
			}, nil)
			convey.So(instance.InstanceType, convey.ShouldEqual, types.InstanceTypeReserved)
		})
		convey.Convey("get scheduler id", func() {
			instance := BuildInstanceFromInsSpec(&commonTypes.InstanceSpecification{
				CreateOptions: map[string]string{
					types.SchedulerIDNote: "scheduler1-permanent",
				},
			}, nil)
			convey.So(instance.Permanent, convey.ShouldEqual, true)
			convey.So(instance.CreateSchedulerID, convey.ShouldEqual, "scheduler1")
			instance = BuildInstanceFromInsSpec(&commonTypes.InstanceSpecification{
				CreateOptions: map[string]string{
					types.SchedulerIDNote: "scheduler1-temporary",
				},
			}, nil)
			convey.So(instance.Permanent, convey.ShouldEqual, false)
			convey.So(instance.CreateSchedulerID, convey.ShouldEqual, "scheduler1")
		})
		convey.Convey("get concurrentNum", func() {
			instance := BuildInstanceFromInsSpec(&commonTypes.InstanceSpecification{
				CreateOptions: map[string]string{
					types.ConcurrentNumKey: "",
				},
			}, nil)
			convey.So(instance.ConcurrentNum, convey.ShouldEqual, 0)
			instance = BuildInstanceFromInsSpec(&commonTypes.InstanceSpecification{
				CreateOptions: map[string]string{
					types.ConcurrentNumKey: "wrong data",
				},
			}, nil)
			convey.So(instance.ConcurrentNum, convey.ShouldEqual, defaultConcurrentNum)
			instance = BuildInstanceFromInsSpec(&commonTypes.InstanceSpecification{
				CreateOptions: map[string]string{
					types.ConcurrentNumKey: "50",
				},
			}, nil)
			convey.So(instance.ConcurrentNum, convey.ShouldEqual, 50)
		})
		convey.Convey("get metricLabelValue ", func() {
			instance := BuildInstanceFromInsSpec(&commonTypes.InstanceSpecification{
				Extensions: commonTypes.Extensions{},
			}, nil)
			convey.So(instance.MetricLabelValues, convey.ShouldBeEmpty)
			instance = BuildInstanceFromInsSpec(&commonTypes.InstanceSpecification{
				Extensions: commonTypes.Extensions{
					PodName:           "aaa",
					PodNamespace:      "bbb",
					PodDeploymentName: "ccc",
				},
			}, &types.FunctionSpecification{
				FuncMetaData: commonTypes.FuncMetaData{
					Name:       "111",
					TenantID:   "222",
					BusinessID: "333",
					FuncName:   "444",
				},
			})
			convey.So(len(instance.MetricLabelValues), convey.ShouldEqual, 8)
		})
	})
}

func TestCheckInstanceSessionValid(t *testing.T) {
	convey.Convey("test CheckInstanceSessionValid", t, func() {
		convey.Convey("CheckInstanceSessionValid", func() {
			res := CheckInstanceSessionValid(commonTypes.InstanceSessionConfig{
				SessionID:  "_123&0",
				SessionTTL: 10,
				Concurrency: 1,
			})
			convey.So(res, convey.ShouldEqual, true)
			res = CheckInstanceSessionValid(commonTypes.InstanceSessionConfig{
				SessionTTL: 10,
				Concurrency: 1,
			})
			convey.So(res, convey.ShouldEqual, false)
			res = CheckInstanceSessionValid(commonTypes.InstanceSessionConfig{
				SessionID:  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				SessionTTL: 10,
				Concurrency: 1,
			})
			convey.So(res, convey.ShouldEqual, false)
			res = CheckInstanceSessionValid(commonTypes.InstanceSessionConfig{
				SessionID:  "aaa",
				SessionTTL: 0,
				Concurrency: 1,
			})
			convey.So(res, convey.ShouldEqual, true)
			res = CheckInstanceSessionValid(commonTypes.InstanceSessionConfig{
				SessionID:   "aaa",
				SessionTTL:  1,
				Concurrency: -1,
			})
			convey.So(res, convey.ShouldEqual, true)
			res = CheckInstanceSessionValid(commonTypes.InstanceSessionConfig{
				SessionID:   "aaa",
				SessionTTL:  1,
				Concurrency: 0,
			})
			convey.So(res, convey.ShouldEqual, false)
			res = CheckInstanceSessionValid(commonTypes.InstanceSessionConfig{
				SessionID:   "aaa",
				SessionTTL:  1,
				Concurrency: -2,
			})
			convey.So(res, convey.ShouldEqual, false)
		})
	})
}

func TestGetInvokeLabelFromResKey(t *testing.T) {
	convey.Convey("test GetInvokeLabelFromResKey", t, func() {
		convey.Convey("get GetInvokeLabelFromResKey", func() {
			res := resspeckey.ResourceSpecification{
				CPU:              500,
				Memory:           1000,
				InvokeLabel:      "aaaaa",
				EphemeralStorage: 0,
			}
			label := GetInvokeLabelFromResKey(res.String())
			convey.So(label, convey.ShouldEqual, "aaaaa")
		})
	})
}

func TestIntMax(t *testing.T) {
	assert.Equal(t, IntMax(3, 1), 3)
	assert.Equal(t, IntMax(1, 3), 3)
}

func TestParseFromCrKey(t *testing.T) {
	_, _, err := ParseFromCrKey("invalidCrKey")
	assert.NotNil(t, err)

	namespace, crName, err := ParseFromCrKey("default:crName")
	assert.Nil(t, err)
	assert.Equal(t, "default", namespace)
	assert.Equal(t, "crName", crName)

	namespace, crName, err = ParseFromCrKey("919afa3d-5dc8-4f13-b2e7-eee79d51e2cd:crName")
	assert.Nil(t, err)
	assert.Equal(t, "919afa3d-5dc8-4f13-b2e7-eee79d51e2cd", namespace)
	assert.Equal(t, "crName", crName)
}
