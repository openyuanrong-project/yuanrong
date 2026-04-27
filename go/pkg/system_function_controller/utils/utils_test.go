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

package utils

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/system_function_controller/types"
)

func TestCheckNeedRecover(t *testing.T) {
	convey.Convey("checkNeedRecover", t, func() {
		var newInstanceMeta types.InstanceSpecificationMeta
		newInstanceMeta.InstanceStatus.Code = 6
		convey.So(CheckNeedRecover(newInstanceMeta), convey.ShouldBeTrue)
		newInstanceMeta.InstanceStatus.Code = 7
		convey.So(CheckNeedRecover(newInstanceMeta), convey.ShouldBeTrue)
		newInstanceMeta.InstanceStatus.Code = 10
		convey.So(CheckNeedRecover(newInstanceMeta), convey.ShouldBeTrue)
		newInstanceMeta.InstanceStatus.Code = 5
		convey.So(CheckNeedRecover(newInstanceMeta), convey.ShouldBeFalse)
	})
}

func TestExtractInfoFromEtcdKey(t *testing.T) {
	convey.Convey("extractInfoFromEtcdKey", t, func() {
		convey.So(ExtractInfoFromEtcdKey("/a/b/c", 10), convey.ShouldBeEmpty)
	})
}

func TestGetMinSleepTime(t *testing.T) {
	convey.Convey("getMinSleepTime", t, func() {
		time1 := time.Duration(9) * time.Second
		time2 := time.Duration(11) * time.Second
		maxTime := time.Duration(10) * time.Second
		convey.So(GetMinSleepTime(time1, maxTime), convey.ShouldEqual, time1)
		convey.So(GetMinSleepTime(time2, maxTime), convey.ShouldEqual, maxTime)
	})
}

func TestGetSchedulerConfigSignature(t *testing.T) {
	convey.Convey("getSchedulerConfigSignature", t, func() {
		outputs := []gomonkey.OutputCell{
			{Values: gomonkey.Params{nil, errors.New("err")}},
			{Values: gomonkey.Params{[]byte{'{', '}'}, nil}},
			{Values: gomonkey.Params{nil, errors.New("err")}}}
		patch := gomonkey.ApplyFuncSeq(json.Marshal, outputs)
		defer patch.Reset()
		convey.So(GetSchedulerConfigSignature(nil), convey.ShouldBeEmpty)
		faaSSchedulerConfig := &types.SchedulerConfig{}
		convey.So(GetSchedulerConfigSignature(faaSSchedulerConfig), convey.ShouldBeEmpty)
	})
}

func TestGetFrontendConfigSignature(t *testing.T) {
	convey.Convey("getFrontendConfigSignature", t, func() {
		outputs := []gomonkey.OutputCell{
			{Values: gomonkey.Params{nil, errors.New("err")}},
			{Values: gomonkey.Params{[]byte{'{', '}'}, nil}},
			{Values: gomonkey.Params{nil, errors.New("err")}}}
		patch := gomonkey.ApplyFuncSeq(json.Marshal, outputs)
		defer patch.Reset()
		convey.So(GetFrontendConfigSignature(nil), convey.ShouldBeEmpty)
		frontendConfig := &types.FrontendConfig{}
		convey.So(GetFrontendConfigSignature(frontendConfig), convey.ShouldBeEmpty)
	})
}

func TestGetManagerConfigSignature(t *testing.T) {
	convey.Convey("getManagerConfigSignature", t, func() {
		outputs := []gomonkey.OutputCell{
			{Values: gomonkey.Params{nil, errors.New("err")}},
			{Values: gomonkey.Params{[]byte{'{', '}'}, nil}},
			{Values: gomonkey.Params{nil, errors.New("err")}}}
		patch := gomonkey.ApplyFuncSeq(json.Marshal, outputs)
		defer patch.Reset()
		convey.So(GetManagerConfigSignature(nil), convey.ShouldBeEmpty)
		frontendConfig := &types.ManagerConfig{}
		convey.So(GetManagerConfigSignature(frontendConfig), convey.ShouldBeEmpty)
	})
}

func TestGetInstanceSpecFromEtcdValue(t *testing.T) {
	convey.Convey("getInstanceSpecFromEtcdValue", t, func() {
		spec := GetInstanceSpecFromEtcdValue([]byte("{\"instanceID\":\"bfc2c7a6-0f26-42b1-9f36-72c0e47b8daf\",\"requestID\":\"113f09375917969b00\",\"functionAgentID\":\"function-agent-72c0e47b8daf-500m-500mi-faasscheduler-dc310000ef\",\"functionProxyID\":\"dggphis190720\",\"function\":\"0/0-system-faasscheduler/$latest\",\"resources\":{\"resources\":{\"Memory\":{\"name\":\"Memory\",\"scalar\":{\"value\":500}},\"CPU\":{\"name\":\"CPU\",\"scalar\":{\"value\":500}}}},\"scheduleOption\":{\"schedPolicyName\":\"monopoly\",\"affinity\":{\"instanceAffinity\":{},\"resource\":{},\"instance\":{\"topologyKey\":\"node\"}},\"resourceSelector\":{\"resource.owner\":\"08d513ba-14f1-4cf0-b400-00000000009a\"},\"extension\":{\"schedule_policy\":\"monopoly\",\"DELEGATE_DIRECTORY_QUOTA\":\"512\"},\"range\":{}},\"createOptions\":{\"resource.owner\":\"system\",\"tenantId\":\"\",\"lifecycle\":\"detached\",\"DELEGATE_POD_LABELS\":\"{\\\"systemFuncName\\\":\\\"faasscheduler\\\"}\",\"RecoverRetryTimes\":\"0\",\"DELEGATE_RUNTIME_MANAGER\":\"{\\\"image\\\":\\\"\\\"}\",\"DELEGATE_DIRECTORY_QUOTA\":\"512\",\"DELEGATE_ENCRYPT\":\"{\\\"metaEtcdPwd\\\":\\\"\\\"}\",\"schedule_policy\":\"monopoly\",\"ConcurrentNum\":\"32\",\"DATA_AFFINITY_ENABLED\":\"false\",\"DELEGATE_NODE_AFFINITY\":\"{\\\"preferredDuringSchedulingIgnoredDuringExecution\\\":[{\\\"preference\\\":{\\\"matchExpressions\\\":[{\\\"key\\\":\\\"node-type\\\",\\\"operator\\\":\\\"In\\\",\\\"values\\\":[\\\"system\\\"]}]},\\\"weight\\\":1}]}\"},\"labels\":[\"faasscheduler\"],\"instanceStatus\":{\"code\":2,\"msg\":\"creating\"},\"jobID\":\"job-12345678\",\"schedulerChain\":[\"function-agent-72c0e47b8daf-500m-500mi-faasscheduler-dc310000ef\"],\"parentID\":\"0-system-faascontroller-0\",\"parentFunctionProxyAID\":\"dggphis190721-LocalSchedInstanceCtrlActor@10.28.83.232:22772\",\"storageType\":\"local\",\"scheduleTimes\":1,\"deployTimes\":1,\"args\":[{\"value\":\"eyJzY2VuYXJpbyI6IiIsImNwdSI6MTAwMCwibWVtb3J5Ijo0MDAwLCJwcmVkaWN0R3JvdXBXaW5kb3ciOjAsInNsYVF1b3RhIjoxMDAwLCJzY2FsZURvd25UaW1lIjo2MDAwMCwiYnVyc3RTY2FsZU51bSI6MTAwMCwibGVhc2VTcGFuIjoxMDAwLCJmdW5jdGlvbkxpbWl0UmF0ZSI6NTAwMCwicm91dGVyRXRjZCI6eyJzZXJ2ZXJzIjpbImRzLWNvcmUtZXRjZC5kZWZhdWx0LnN2Yy5jbHVzdGVyLmxvY2FsOjIzNzkiXSwidXNlciI6IiIsInBhc3N3b3JkIjoiIiwic3NsRW5hYmxlIjpmYWxzZSwiYXV0aFR5cGUiOiJOb2F1dGgiLCJ1c2VTZWNyZXQiOmZhbHNlLCJzZWNyZXROYW1lIjoiZXRjZC1jbGllbnQtc2VjcmV0IiwiQ2FGaWxlIjoiIiwiQ2VydEZpbGUiOiIiLCJLZXlGaWxlIjoiIiwiUGFzc3BocmFzZUZpbGUiOiIifSwibWV0YUV0Y2QiOnsic2VydmVycyI6WyJkcy1jb3JlLWV0Y2QuZGVmYXVsdC5zdmMuY2x1c3Rlci5sb2NhbDoyMzc5Il0sInVzZXIiOiIiLCJwYXNzd29yZCI6IiIsInNzbEVuYWJsZSI6ZmFsc2UsImF1dGhUeXBlIjoiTm9hdXRoIiwidXNlU2VjcmV0IjpmYWxzZSwic2VjcmV0TmFtZSI6ImV0Y2QtY2xpZW50LXNlY3JldCIsIkNhRmlsZSI6IiIsIkNlcnRGaWxlIjoiIiwiS2V5RmlsZSI6IiIsIlBhc3NwaHJhc2VGaWxlIjoiIn0sImRvY2tlclJvb3RQYXRoIjoiL3Zhci9saWIvZG9ja2VyIiwicmF3U3RzQ29uZmlnIjp7InNlbnNpdGl2ZUNvbmZpZ3MiOnsic2hhcmVLZXlzIjpudWxsfSwic2VydmVyQ29uZmlnIjp7InBhdGgiOiIvb3B0L2h1YXdlaS9jZXJ0cy9ITVNDbGllbnRDbG91ZEFjY2VsZXJhdGVTZXJ2aWNlL0hNU0NhYVNZdWFuUm9uZ1dvcmtlci9ITVNDYWFzWXVhblJvbmdXb3JrZXIuaW5pIn0sIm1nbXRTZXJ2ZXJDb25maWciOnt9fSwiY2x1c3RlcklEIjoiY2x1c3RlcjAwMSIsImNsdXN0ZXJOYW1lIjoiZHN3ZWJfY2NldHVyYm9fYmo0X2F1dG9fYXoxIiwiZGlza01vbml0b3JFbmFibGUiOmZhbHNlLCJyZWdpb25OYW1lIjoiYmVpamluZzQiLCJhbGFybUNvbmZpZyI6eyJlbmFibGVBbGFybSI6ZmFsc2UsImFsYXJtTG9nQ29uZmlnIjp7ImZpbGVwYXRoIjoiL29wdC9odWF3ZWkvbG9ncy9hbGFybXMiLCJsZXZlbCI6IkluZm8iLCJ0aWNrIjowLCJmaXJzdCI6MCwidGhlcmVhZnRlciI6MCwidHJhY2luZyI6ZmFsc2UsImRpc2FibGUiOmZhbHNlLCJzaW5nbGVzaXplIjo1MDAsInRocmVzaG9sZCI6M30sInhpYW5nWXVuRm91ckNvbmZpZyI6eyJzaXRlIjoiY25fZGV2X2RlZmF1bHQiLCJ0ZW5hbnRJRCI6IlQwMTQiLCJhcHBsaWNhdGlvbklEIjoiY29tLmh1YXdlaS5jbG91ZF9lbmhhbmNlX2RldmljZSIsInNlcnZpY2VJRCI6ImNvbS5odWF3ZWkuaG1zY29yZWNhbWVyYWNsb3VkZW5oYW5jZXNlcnZpY2UifSwibWluSW5zU3RhcnRJbnRlcnZhbCI6MTUsIm1pbkluc0NoZWNrSW50ZXJ2YWwiOjE1fSwiZXBoZW1lcmFsU3RvcmFnZSI6NTEyLCJob3N0YWxpYXNlc2hvc3RuYW1lIjpudWxsLCJmdW5jdGlvbkNvbmZpZyI6bnVsbCwibG9jYWxBdXRoIjp7ImFLZXkiOiIiLCJzS2V5IjoiIiwiZHVyYXRpb24iOjB9LCJjb25jdXJyZW50TnVtIjowLCJ2ZXJzaW9uIjoiIiwiaW1hZ2UiOiIiLCJuYW1lU3BhY2UiOiJkZWZhdWx0Iiwic2NjQ29uZmlnIjp7ImVuYWJsZSI6ZmFsc2UsInNlY3JldE5hbWUiOiJzY2Mta3Mtc2VjcmV0IiwiYWxnb3JpdGhtIjoiQUVTMjU2X0dDTSJ9LCJub2RlQWZmaW5pdHkiOiJ7XCJwcmVmZXJyZWREdXJpbmdTY2hlZHVsaW5nSWdub3JlZER1cmluZ0V4ZWN1dGlvblwiOlt7XCJwcmVmZXJlbmNlXCI6e1wibWF0Y2hFeHByZXNzaW9uc1wiOlt7XCJrZXlcIjpcIm5vZGUtdHlwZVwiLFwib3BlcmF0b3JcIjpcIkluXCIsXCJ2YWx1ZXNcIjpbXCJzeXN0ZW1cIl19XX0sXCJ3ZWlnaHRcIjoxfV19Iiwic2NoZWR1bGVyTnVtIjoyfQ==\"}],\"version\":\"2\",\"dataSystemHost\":\"127.0.0.1\",\"detached\":true,\"gracefulShutdownTime\":\"600\",\"tenantID\":\"0\",\"isSystemFunc\":true}"))
		convey.So(spec, convey.ShouldNotBeNil)
	})
}
