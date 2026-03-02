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

// Package registry -
package registry

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"

	"yuanrong.org/kernel/pkg/common/faas_common/aliasroute"
	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/instanceconfig"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

func TestInitRegistry(t *testing.T) {
	config.GlobalConfig = types.Configuration{}
	patches := []*Patches{
		ApplyMethod(reflect.TypeOf(&etcd3.EtcdInitParam{}), "InitClient",
			func(_ *etcd3.EtcdInitParam) error {
				return nil
			}),
		ApplyMethod(reflect.TypeOf(&etcd3.EtcdWatcher{}), "StartList", func(_ *etcd3.EtcdWatcher) {
			return
		}),
		ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
			return &etcd3.EtcdClient{}
		}),
		ApplyFunc(etcd3.GetMetaEtcdClient, func() *etcd3.EtcdClient {
			return &etcd3.EtcdClient{}
		}),
		ApplyFunc(etcd3.GetCAEMetaEtcdClient, func() *etcd3.EtcdClient {
			return &etcd3.EtcdClient{}
		}),
		ApplyFunc((*FaasSchedulerRegistry).WaitForETCDList, func() {}),
	}
	defer func() {
		time.Sleep(1000 * time.Millisecond)
		for _, patch := range patches {
			time.Sleep(100 * time.Millisecond)
			patch.Reset()
		}
	}()
	stopCh := make(chan struct{})
	err := InitRegistry(stopCh)
	assert.Equal(t, true, err == nil)
	instance := GlobalRegistry.GetInstance("instance1")
	assert.Equal(t, true, instance == nil)
	function := GlobalRegistry.FetchSilentFuncSpec("function1")
	assert.Equal(t, true, function == nil)
	subChan := make(chan SubEvent, 1)
	GlobalRegistry.SubscribeFuncSpec(subChan)
	assert.Equal(t, 1, len(GlobalRegistry.FunctionRegistry.subscriberChans))
	GlobalRegistry.SubscribeInsSpec(subChan)
	assert.Equal(t, 1, len(GlobalRegistry.InstanceRegistry.subscriberChans))
	GlobalRegistry.SubscribeInsConfig(subChan)
	assert.Equal(t, 1, len(GlobalRegistry.InstanceConfigRegistry.subscriberChans))
	GlobalRegistry.SubscribeAliasSpec(subChan)
	assert.Equal(t, 1, len(GlobalRegistry.AliasRegistry.subscriberChans))
	GlobalRegistry.SubscribeSchedulerProxy(subChan)
	assert.Equal(t, 1, len(GlobalRegistry.FaaSSchedulerRegistry.subscriberChans))
}

func TestNewInstanceRegistry(t *testing.T) {
	stopCh := make(chan struct{})
	instanceRegistry := NewInstanceRegistry(stopCh)
	assert.Equal(t, true, instanceRegistry != nil)
}

func TestInstanceWatcherFilter(t *testing.T) {
	stopCh := make(chan struct{})
	ir := NewInstanceRegistry(stopCh)
	event := &etcd3.Event{
		Type:  etcd3.PUT,
		Key:   "/sn/instance/business/yrk/tenant",
		Value: []byte("value"),
		Rev:   1,
	}
	assert.Equal(t, true, ir.watcherFilter(event))

	event.Key = "/sn/instance/business/yrk/tenant/123/instance/faasscheduler/version/$latest/defaultaz/requestID/abc"
	assert.Equal(t, true, ir.watcherFilter(event))

	event.Key = "/sn/instance/business/yrk/tenant/123/function/faasscheduler/version/$latest/defaultaz/requestID/abc"
	assert.Equal(t, true, ir.watcherFilter(event))

	convey.Convey("EtcdList", t, func() {
		defer ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
			return &etcd3.EtcdClient{}
		}).Reset()
		defer ApplyMethod(reflect.TypeOf(&etcd3.EtcdClient{}), "Get", func(_ *etcd3.EtcdClient, ctx etcd3.EtcdCtxInfo,
			key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
			getRsp := &clientv3.GetResponse{
				Kvs: []*mvccpb.KeyValue{
					{
						Key:   []byte("/sn/instance/business/yrk/tenant/0/function/0-system-faasscheduler/version/$latest/defaultaz/8c1e66d5f21be4fc00/45a6e8e0-d99a-46ec-afb1-feb6f640f37d"),
						Value: []byte("{\"instanceID\":\"scheduler-faas-scheduler-6c776f9b98-jl2cq\",\"runtimeID\":\"scheduler-faas-scheduler-6c776f9b98-jl2cq\",\"runtimeAddress\":\"127.0.0.1:0\",\"functionProxyID\":\"dxxxxx9\",\"function\":\"0/0-system-faasscheduler/$latest\",\"instanceStatus\":{\"code\":3},\"jobID\":\"runtime-62865730\",\"tenantID\":\"0\",\"isSystemFunc\":true,\"extensions\":{\"source\":\"driver\"}}"),
					},
					{
						Key:   []byte("/sn/instance/business/yrk/tenant/12345678901234561234567890123456/function/0-system-faasExecutorGo1.x/version/$latest/defaultaz/f1d87744e81c386100/db547379-0000-4000-8000-00d8318e1989"),
						Value: []byte("{\"instanceID\":\"db547379-0000-4000-8000-00d8318e1989\",\"requestID\":\"f1d87744e81c386100\",\"runtimeID\":\"runtime-db547379-0000-4000-8000-00d8318e1989-6457919349af\",\"runtimeAddress\":\"127.0.0.1:0\",\"functionAgentID\":\"custom-function-agent-00d8318e1989-500m-500mi-4e001e94000000006\",\"functionProxyID\":\"dggphis191119\",\"function\":\"12345678901234561234567890123456/0-system-faasExecutorGo1.x/$latest\",\"resources\":{\"resources\":{\"Memory\":{\"name\":\"Memory\",\"scalar\":{\"value\":500}},\"CPU\":{\"name\":\"CPU\",\"scalar\":{\"value\":500}}}},\"scheduleOption\":{\"schedPolicyName\":\"monopoly\",\"affinity\":{\"instanceAffinity\":{},\"resource\":{},\"instance\":{\"scope\":\"NODE\"}},\"initCallTimeOut\":300,\"resourceSelector\":{\"resource.owner\":\"78000000-0000-4000-bc60-371c01ba313f\"},\"extension\":{\"schedule_policy\":\"monopoly\",\"DELEGATE_DIRECTORY_QUOTA\":\"512\"},\"range\":{},\"scheduleTimeoutMs\":\"5000\"},\"createOptions\":{\"init_call_timeout\":\"300\",\"INVOKE_TIMEOUT\":\"61\",\"FUNCTION_KEY_NOTE\":\"0aae5ae9ea9a42c3adc43dd368efb29b/0@default@hello/latest\",\"ConcurrentNum\":\"5\",\"DELEGATE_POD_LABELS\":\"\",\"lifecycle\":\"detached\",\"GRACEFUL_SHUTDOWN_TIME\":\"900\",\"DELEGATE_HOST_ALIASES\":\"{\\\"10.34.244.163\\\":[\\\"axxx.xxx.cn\\\"],\\\"127.0.0.1\\\":[\\\"wiseguard-rms.hwcloudtest.cn\\\"]}\",\"DELEGATE_CONTAINER\":\"\",\"DELEGATE_VOLUMES\":\"\",\"DELEGATE_AGENT_VOLUME_MOUNTS\":\"null\",\"INSTANCE_TYPE_NOTE\":\"reserved\",\"DELEGATE_VOLUME_MOUNTS\":\"\",\"DATA_AFFINITY_ENABLED\":\"false\",\"SCHEDULER_ID_NOTE\":\"scheduler-faas-scheduler-5f6cf4f58b-bk6ct-temporary\",\"RESOURCE_SPEC_NOTE\":\"\",\"call_timeout\":\"61\",\"DELEGATE_CONTAINER_ID\":\"252420214fc46dc2dc123ad6f0a9d2ff7797b027b737b5ff0b2383330c5b97e8\",\"DELEGATE_POD_INIT_LABELS\":\"{\\\"securityGroup\\\":\\\"0aae5ae9ea9a42c3adc43dd368efb29b\\\"}\",\"FUNCTION_SIGNATURE\":\"3854191954\",\"DELEGATE_ENCRYPT\":\"\",\"RecoverRetryTimes\":\"0\",\"DELEGATE_DIRECTORY_QUOTA\":\"512\",\"schedule_policy\":\"monopoly\",\"DELEGATE_DIRECTORY_INFO\":\"/tmp\",\"DELEGATE_ENV_VAR\":\"null\",\"DELEGATE_RUNTIME_MANAGER\":\"{\\\"env\\\":null}\",\"tenantId\":\"0aae5ae9ea9a42c3adc43dd368efb29b\"},\"labels\":[\"faas\"],\"instanceStatus\":{\"code\":3,\"msg\":\"running\"},\"jobID\":\"job-ime-74613546\",\"schedulerChain\":[\"custom-function-agent-00d8318e1989-500m-500mi-4e001e94000000006\"],\"parentID\":\"scheduler-faas-scheduler-5f6cf4f58b-bk6ct\",\"parentFunctionProxyAID\":\"dggphis191118-LocalSchedInstanceCtrlActor@7.227.25.28:22423\",\"storageType\":\"local\",\"scheduleTimes\":1,\"deployTimes\":1,\"args\":[{\"value\":\"\"}],\"version\":\"3\",\"dataSystemHost\":\"7.227.27.16\",\"detached\":true,\"gracefulShutdownTime\":\"900\",\"extensions\":{\"receivedTimestamp\":\"1765445562326\",\"pid\":\"116\",\"createTimestamp\":\"1765445562\",\"updateTimestamp\":\"1765445625\"},\"unitID\":\"custom-function-agent-00d8318e1989-500m-500mi-4e001e94000000006\"}"),
					},
				},
			}
			return getRsp, nil
		}).Reset()
		insList := ir.EtcdList()
		convey.So(len(insList), convey.ShouldEqual, 1)
	})
}

func TestGetFuncSpecFromEtcdValue(t *testing.T) {
	type args struct {
		etcdValue []byte
	}
	tests := []struct {
		name  string
		args  args
		IsNil bool
	}{
		{"test1",
			args{etcdValue: []byte("{\"funcMetaData\":{\"layers\":[],\"name\":\"0-system-hello\",\"description\"" +
				":\"\",\"functionUrn\":\"sn:cn:yrk:12345678901234561234567890123456:function:0-system-hello\",\"rever" +
				"sedConcurrency\":0,\"tags\":null,\"functionUpdateTime\":\"\",\"functionVersionUrn\":\"sn:cn:yrk:12345" +
				"678901234561234567890123456:function:0-system-hello:$latest\",\"codeSize\":1619264,\"codeSha256\":\"c" +
				"e9f7446a54331137c8386cedc38eec942f33bab0575c81d5f3b5633caff2596\",\"handler\":\"\",\"runtime\":\"go1" +
				".13\",\"timeout\":900,\"version\":\"$latest\",\"versionDescription\":\"$latest\",\"deadLetterConfi" +
				"g\":\"\",\"latestVersionUpdateTime\":\"\",\"publishTime\":\"\",\"businessId\":\"yrk\",\"tenantId\":\"1" +
				"2345678901234561234567890123456\",\"domain_id\":\"\",\"project_name\":\"\",\"revisionId\":\"202212150" +
				"92604748\",\"created\":\"2022-12-13 13:01:44.376 UTC\",\"statefulFlag\":false,\"hookHandler\":{\"cal" +
				"l\":\"main.CallHandler\",\"init\":\"main.InitHandler\"}},\"codeMetaData\":{\"storage_type\":\"s3\",\"a" +
				"ppId\":\"61022\",\"bucketId\":\"bucket-test-log1\",\"objectId\":\"hello-1671096364751\",\"bucketUr" +
				"l\":\"http://127.0.0.1:19002\",\"sha256\":\"\",\"code_type\":\"\",\"code_url\":\"\",\"code_filen" +
				"ame\":\"\",\"func_code\":{\"file\":\"\",\"link\":\"\"},\"code_path\":\"\"},\"envMetaData\":{\"envKe" +
				"y\":\"85c545e0e31241d681031542:8231fc7f6dd9f6411d03bb5cf751a398bcf1d3d4fa1098022228c75cdb7420116807" +
				"2edc1bb265f53bc8b4fee10e757693935bd8d412e292ac2349207c52311b9cef460a65c91a4103b9aed5dc920b49\",\"env" +
				"ironment\":\"9cce4db16c95d4a215999e26:010456a679a83eefda685c2eff8330c69285a1196e53afaca04fd0f0bef5b87" +
				"e369f603794c9942a1c38e9b8ef0e49286f8b2a06aebc007f90ebf11f97eeb16f2668eb66e551a23206896df0391f6e16536d" +
				"8141d6f4f94ce75ad4125e5c9fba83bd594cda705beb9b215846e580f2594930c7d61f9f2f2ce6c14de68d5a44369e7e51aea" +
				"3b8d60d44f7673bd143ac688b1e5530a9714083aac51d0d6a776ed9d72da2960972f37e48\",\"encrypted_user_data" +
				"\":\"\"},\"resourceMetaData\":{\"cpu\":500,\"memory\":500,\"customResources\":\"\"},\"extendedMetaDa" +
				"ta\":{\"image_name\":\"\",\"role\":{\"xrole\":\"\",\"app_xrole\":\"\"},\"mount_config\":{\"mount_use" +
				"r\":{\"user_id\":0,\"user_group_id\":0},\"func_mounts\":null},\"strategy_config\":{\"concurrenc" +
				"y\":0},\"extend_config\":\"\",\"initializer\":{\"initializer_handler\":\"\",\"initializer_timeou" +
				"t\":0},\"enterprise_project_id\":\"\",\"log_tank_service\":{\"logGroupId\":\"\",\"logStrea" +
				"mId\":\"\"},\"tracing_config\":{\"tracing_ak\":\"\",\"tracing_sk\":\"\",\"project_name\":\"\"},\"us" +
				"er_type\":\"\",\"instance_meta_data\":{\"maxInstance\":100,\"minInstance\":0,\"concurrentN" +
				"um\":100,\"cacheInstance\":0},\"extended_handler\":null,\"extended_timeout\":null}}")}, false},
		{"test2", args{etcdValue: []byte("")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFuncMetaInfoFromEtcdValue(tt.args.etcdValue)
			assert.Equal(t, tt.IsNil, got == nil)
		})
	}
}

func TestGetUserAgencyByFuncMeta(t *testing.T) {
	type args struct {
		funcMetaInfo *commonTypes.FunctionMetaInfo
	}
	tests := []struct {
		name string
		args args
		want commonTypes.UserAgency
	}{
		{
			name: "test funcMeta nil",
			args: args{funcMetaInfo: nil},
			want: commonTypes.UserAgency{},
		},
		{
			name: "test AppXRole not empty",
			args: args{funcMetaInfo: &commonTypes.FunctionMetaInfo{
				ExtendedMetaData: commonTypes.ExtendedMetaData{
					Role: commonTypes.Role{
						AppXRole: "",
					},
				},
			}},
			want: commonTypes.UserAgency{},
		},
		{
			name: "test AppXRole empty but XRole not empty",
			args: args{funcMetaInfo: &commonTypes.FunctionMetaInfo{
				ExtendedMetaData: commonTypes.ExtendedMetaData{
					Role: commonTypes.Role{
						XRole:    "test",
						AppXRole: "",
					},
				},
			}},
			want: commonTypes.UserAgency{},
		},
	}
	userAgencyRegistry := NewUserAgencyRegistry(make(chan struct{}))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, userAgencyRegistry.GetUserAgencyByFuncMeta(tt.args.funcMetaInfo),
				"GetUserAgencyByFuncMeta(%v)", tt.args.funcMetaInfo)
		})
	}
}

func TestAgencyWatcherFilter(t *testing.T) {
	userAgencyRegistry := NewUserAgencyRegistry(make(chan struct{}))
	event := &etcd3.Event{
		Type:  etcd3.PUT,
		Key:   "/sn/agency/business/yrk/tenant/123/domain/123/agency/123",
		Value: []byte("value"),
		Rev:   1,
	}
	assert.Equal(t, false, userAgencyRegistry.watcherFilter(event))

	event.Key = "/sn/functions/business/yrk/tenant/123/domain/123/agency/123"
	assert.Equal(t, true, userAgencyRegistry.watcherFilter(event))

	event.Key = "/sn/agency/busi/yrk/tenant/123/domain/123/agency/123"
	assert.Equal(t, true, userAgencyRegistry.watcherFilter(event))

	event.Key = "/sn/agency/business/yrk/ten/123/domain/123/agency/123"
	assert.Equal(t, true, userAgencyRegistry.watcherFilter(event))

	event.Key = "/sn/agency/business/yrk/tenant/123/dom/123/agency/123"
	assert.Equal(t, true, userAgencyRegistry.watcherFilter(event))

	event.Key = "/sn/agency/business/yrk/tenant/123/domain/123/agen/123"
	assert.Equal(t, true, userAgencyRegistry.watcherFilter(event))

	event.Key = "/sn/agency/business/yrk/tenant/123/instance/faasscheduler/version/$latest/defaultaz/abc"
	assert.Equal(t, true, userAgencyRegistry.watcherFilter(event))

	event.Key = "/sn/agency/business/yrk/tenant/123/function/faasscheduler/version/$latest/defaultaz/abc"
	assert.Equal(t, true, userAgencyRegistry.watcherFilter(event))
}

func TestAgencyWatcherHandler(t *testing.T) {
	agency := `{"accessKey": "ak","secretKey": "sk","token": "token","akSkExpireTime": "2021-06-11T23:00:00.101073Z","tokenExpireTime": "2021-06-12T13:23:00.483000Z"}`
	convey.Convey("agencyWatcherHandler", t, func() {
		stopCh := make(chan struct{})
		userAgencyRegistry := NewUserAgencyRegistry(stopCh)
		funcMetaInfo := &commonTypes.FunctionMetaInfo{
			FuncMetaData: commonTypes.FuncMetaData{
				TenantID: "123",
				DomainID: "123",
			},
			ExtendedMetaData: commonTypes.ExtendedMetaData{
				Role: commonTypes.Role{
					XRole: "123",
				},
			},
		}
		convey.Convey("agencyWatcherHandler put unmarshal failed", func() {
			defer ApplyFunc(json.Unmarshal, func(data []byte, v interface{}) error {
				return fmt.Errorf("unmarshal failed")
			}).Reset()
			event := &etcd3.Event{
				Type:  etcd3.PUT,
				Key:   "/sn/agency/business/yrk/tenant/123/domain/123/agency/123",
				Value: []byte(agency),
				Rev:   1,
			}
			userAgencyRegistry.watcherHandler(event)
			getAgency := userAgencyRegistry.GetUserAgencyByFuncMeta(funcMetaInfo)
			convey.So(getAgency.AccessKey, convey.ShouldBeZeroValue)
			convey.So(getAgency.SecretKey, convey.ShouldBeZeroValue)
			convey.So(getAgency.Token, convey.ShouldBeZeroValue)
			convey.So(getAgency.SecurityAk, convey.ShouldBeZeroValue)
			convey.So(getAgency.SecuritySk, convey.ShouldBeZeroValue)
			convey.So(getAgency.SecurityToken, convey.ShouldBeZeroValue)
		})
		convey.Convey("agencyWatcherHandler put", func() {
			event := &etcd3.Event{
				Type:  etcd3.PUT,
				Key:   "/sn/agency/business/yrk/tenant/123/domain/123/agency/123",
				Value: []byte(agency),
				Rev:   1,
			}
			userAgencyRegistry.watcherHandler(event)
			getAgency := userAgencyRegistry.GetUserAgencyByFuncMeta(funcMetaInfo)
			convey.So(getAgency.AccessKey, convey.ShouldNotBeZeroValue)
			convey.So(getAgency.SecretKey, convey.ShouldNotBeZeroValue)
			convey.So(getAgency.Token, convey.ShouldNotBeZeroValue)
			convey.So(getAgency.SecurityAk, convey.ShouldBeZeroValue)
			convey.So(getAgency.SecuritySk, convey.ShouldBeZeroValue)
			convey.So(getAgency.SecurityToken, convey.ShouldBeZeroValue)
		})
		convey.Convey("agencyWatcherHandler delete", func() {
			event := &etcd3.Event{
				Type:  etcd3.DELETE,
				Key:   "/sn/agency/business/yrk/tenant/123/domain/123/agency/123",
				Value: []byte(agency),
				Rev:   1,
			}
			userAgencyRegistry.watcherHandler(event)
			getAgency := userAgencyRegistry.GetUserAgencyByFuncMeta(funcMetaInfo)
			convey.So(getAgency.AccessKey, convey.ShouldBeZeroValue)
			convey.So(getAgency.SecretKey, convey.ShouldBeZeroValue)
			convey.So(getAgency.Token, convey.ShouldBeZeroValue)
			convey.So(getAgency.SecurityAk, convey.ShouldBeZeroValue)
			convey.So(getAgency.SecuritySk, convey.ShouldBeZeroValue)
			convey.So(getAgency.SecurityToken, convey.ShouldBeZeroValue)
		})

		convey.Convey("etcd3.SYNCED & unknow etcd opt", func() {
			event := &etcd3.Event{
				Type: etcd3.SYNCED,
			}
			userAgencyRegistry.watcherHandler(event)
			event.Type = 4
			userAgencyRegistry.watcherHandler(event)
		})
		close(stopCh)
	})
}

func TestInstanceRegistryWatcherHandler(t *testing.T) {
	config.GlobalConfig.Scenario = types.ScenarioWiseCloud
	defer func() {
		config.GlobalConfig.Scenario = "faas"
	}()
	stopCh := make(chan struct{})
	GlobalRegistry = &Registry{
		FaaSSchedulerRegistry: NewFaasSchedulerRegistry(stopCh),
		FunctionRegistry:      NewFunctionRegistry(stopCh),
		InstanceRegistry:      NewInstanceRegistry(stopCh),
		FaaSManagerRegistry:   NewFaaSManagerRegistry(stopCh),
	}

	recvMsg := make(chan SubEvent, 1)
	ir := &InstanceRegistry{
		InstanceIDMap:         make(map[string]*types.Instance, utils.DefaultMapSize),
		functionInstanceIDMap: make(map[string]map[string]*commonTypes.InstanceSpecification, utils.DefaultMapSize),
		listDoneCh:            make(chan struct{}, 1),
	}
	ir.addSubscriberChan(recvMsg)

	event := &etcd3.Event{
		Type:      etcd3.PUT,
		Key:       "/sn/instance/business/yrk/tenant/12345678901234561234567890123456/function/0-system-faasExecutorJava8/version/$latest/defaultaz/task-b23aa1c4-2084-42b8-99b2-8907fa5ae6f4/f71875b1-3c20-4827-8600-0000000005d5",
		Value:     []byte("123"),
		PrevValue: []byte("123"),
		Rev:       1,
	}

	convey.Convey("etcd put value error", t, func() {
		event.Type = etcd3.PUT
		ir.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg, convey.ShouldResemble, SubEvent{})
	})
	convey.Convey("etcd put value success", t, func() {
		selfregister.SelfInstanceID = ""
		instanceSpecByte, _ := json.Marshal(&commonTypes.InstanceSpecification{Labels: []string{""},
			CreateOptions: map[string]string{types.FunctionKeyNote: "test-function"}})
		event.Value = instanceSpecByte
		event.Type = etcd3.PUT
		ir.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		}
		convey.So(msg.EventType, convey.ShouldEqual, SubEventTypeUpdate)
		convey.So(msg.EventMsg, convey.ShouldResemble, &commonTypes.InstanceSpecification{
			InstanceID:    "f71875b1-3c20-4827-8600-0000000005d5",
			Labels:        []string{""},
			CreateOptions: map[string]string{types.FunctionKeyNote: "test-function"},
		})
	})
	convey.Convey("etcd delete value error", t, func() {
		event.Type = etcd3.DELETE
		ir.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg, convey.ShouldResemble, SubEvent{})
	})
	convey.Convey("etcd delete value success", t, func() {
		instanceSpecByte, _ := json.Marshal(&commonTypes.InstanceSpecification{Labels: []string{""},
			CreateOptions: map[string]string{types.FunctionKeyNote: "test-function"}})
		event.PrevValue = instanceSpecByte
		event.Type = etcd3.DELETE
		ir.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		}
		convey.So(msg.EventType, convey.ShouldEqual, SubEventTypeDelete)
		convey.So(msg.EventMsg, convey.ShouldResemble, &commonTypes.InstanceSpecification{
			InstanceID:    "f71875b1-3c20-4827-8600-0000000005d5",
			Labels:        []string{""},
			CreateOptions: map[string]string{types.FunctionKeyNote: "test-function"},
		})
	})
	convey.Convey("etcd put invalid funcKey", t, func() {
		event.Type = etcd3.PUT
		event.Key = "/sn/instance/business/yrk/tenant//function"
		ir.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg, convey.ShouldResemble, SubEvent{})
	})
	convey.Convey("etcd put invalid instanceID", t, func() {
		event.Type = etcd3.PUT
		event.Key = "/sn/instance/business/yrk/tenant/12345678901234561234567890123456/function/0-system-faasExecutorJava8/version/$latest/defaultaz//"
		ir.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg, convey.ShouldResemble, SubEvent{})
	})
	convey.Convey("etcd SYNCED", t, func() {
		event.Type = etcd3.SYNCED
		ir.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg.EventType, convey.ShouldEqual, SubEventTypeSynced)
	})
	convey.Convey("etcd put invalid instanceID", t, func() {
		event.Type = etcd3.PUT
		event.Key = "/sn/instance/business/yrk/tenant/12345678901234561234567890123456/function/0-system-faasExecutorJava8/version/$latest/defaultaz//"
		ir.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg, convey.ShouldResemble, SubEvent{})
	})
	close(stopCh)
}

func TestInstanceRegistryWatcherFGHandler(t *testing.T) {
	config.GlobalConfig.Scenario = types.ScenarioWiseCloud
	defer func() {
		config.GlobalConfig.Scenario = "faas"
	}()
	stopCh := make(chan struct{})
	GlobalRegistry = &Registry{
		FaaSSchedulerRegistry: NewFaasSchedulerRegistry(stopCh),
		FunctionRegistry:      NewFunctionRegistry(stopCh),
		InstanceRegistry:      NewInstanceRegistry(stopCh),
		FaaSManagerRegistry:   NewFaaSManagerRegistry(stopCh),
	}

	recvMsg := make(chan SubEvent, 1)
	ir := &InstanceRegistry{
		InstanceIDMap:         make(map[string]*types.Instance, utils.DefaultMapSize),
		functionInstanceIDMap: make(map[string]map[string]*commonTypes.InstanceSpecification, utils.DefaultMapSize),
		fgListDoneCh:          make(chan struct{}, 1),
	}
	ir.addSubscriberChan(recvMsg)

	event := &etcd3.Event{
		Type: etcd3.PUT,
		Key: "/sn/workers/business/yrk/tenant/c53626012ba84727b938ca8bf03108ef/function/0@default@zscaetest/" +
			"version/latest/defaultaz/defaultaz-#-custom-1000-1024-935e9454-93fa-43f1-b5e4-7cd82737dd62",
		Value:     []byte("{\"ip\":\"192.168.0.154\",\"port\":\"8080\",\"cluster\":\"cluster001\",\"status\":\"ready\",\"p2pPort\":\"22668\",\"nodeIP\":\"127.0.0.1\",\"nodePort\":\"22423\",\"applier\":\"worker-manager\",\"ownerIP\":\"127.0.0.1\",\"businessType\":\"CAE\",\"hasInitializer\":true,\"creationTime\":1724393756,\"podUID\":\"c00e66f0-a4b1-46db-8d9c-61d7ab8c2405\",\"containerIDs\":{\"worker\":\"58dd3e59f60f7533cfe5604a076470d51b1e9b5f3e87a70c0954937fddfa7280\",\"runtime\":\"b616e60c55adab60271507c8df3aefec5923ee02fc851689294d04325ef522d1\"},\"resource\":{\"worker\":{\"cpuLimit\":1000,\"cpuRequest\":100,\"memoryLimit\":3686,\"memoryRequest\":100},\"runtime\":{\"cpuLimit\":1000,\"cpuRequest\":250,\"memoryLimit\":1024,\"memoryRequest\":1024}}}"),
		PrevValue: []byte("{\"ip\":\"192.168.0.154\",\"port\":\"8080\",\"cluster\":\"cluster001\",\"status\":\"ready\",\"p2pPort\":\"22668\",\"nodeIP\":\"127.0.0.1\",\"nodePort\":\"22423\",\"applier\":\"worker-manager\",\"ownerIP\":\"127.0.0.1\",\"businessType\":\"CAE\",\"hasInitializer\":true,\"creationTime\":1724393756,\"podUID\":\"c00e66f0-a4b1-46db-8d9c-61d7ab8c2405\",\"containerIDs\":{\"worker\":\"58dd3e59f60f7533cfe5604a076470d51b1e9b5f3e87a70c0954937fddfa7280\",\"runtime\":\"b616e60c55adab60271507c8df3aefec5923ee02fc851689294d04325ef522d1\"},\"resource\":{\"worker\":{\"cpuLimit\":1000,\"cpuRequest\":100,\"memoryLimit\":3686,\"memoryRequest\":100},\"runtime\":{\"cpuLimit\":1000,\"cpuRequest\":250,\"memoryLimit\":1024,\"memoryRequest\":1024}}}"),
		Rev:       1,
	}

	convey.Convey("etcd put value success", t, func() {
		event.Type = etcd3.PUT
		ir.watcherFGHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		}
		convey.So(msg.EventType, convey.ShouldEqual, SubEventTypeUpdate)
		convey.So(msg.EventMsg, convey.ShouldResemble, &commonTypes.InstanceSpecification{
			InstanceID:      "defaultaz-#-custom-1000-1024-935e9454-93fa-43f1-b5e4-7cd82737dd62",
			DataSystemHost:  "",
			RequestID:       "",
			RuntimeID:       "",
			RuntimeAddress:  "127.0.0.1:22423",
			FunctionAgentID: "",
			FunctionProxyID: "192.168.0.154:8080",
			Function:        "c53626012ba84727b938ca8bf03108ef/0@default@zscaetest/latest",
			RestartPolicy:   "",
			Resources: commonTypes.Resources{
				Resources: map[string]commonTypes.Resource{
					constant.ResourceCPUName: {
						Name: constant.ResourceCPUName,
						Scalar: commonTypes.ValueScalar{
							Value: float64(1000),
						},
					},
					constant.ResourceMemoryName: {
						Name: constant.ResourceMemoryName,
						Scalar: commonTypes.ValueScalar{
							Value: float64(1024),
						},
					},
					constant.ResourceEphemeralStorage: {
						Name: constant.ResourceEphemeralStorage,
						Scalar: commonTypes.ValueScalar{
							Value: float64(0),
						},
					},
				},
			},
			ActualUse: commonTypes.Resources{
				Resources: map[string]commonTypes.Resource(nil),
			},
			ScheduleOption: commonTypes.ScheduleOption{
				SchedPolicyName: "",
				Priority:        0,
				Affinity: commonTypes.Affinity{
					NodeAffinity: commonTypes.NodeAffinity{
						Affinity: map[string]string(nil),
					},
					InstanceAffinity: commonTypes.InstanceAffinity{
						Affinity: map[string]string(nil),
					},
					InstanceAntiAffinity: commonTypes.InstanceAffinity{
						Affinity: map[string]string(nil),
					},
				},
			},
			CreateOptions: map[string]string{
				"FUNCTION_KEY_NOTE":  "c53626012ba84727b938ca8bf03108ef/0@default@zscaetest/latest",
				"INSTANCE_TYPE_NOTE": "reserved",
				"SCHEDULER_ID_NOTE":  "worker-manager",
			},
			Labels:    []string(nil),
			StartTime: "1724393756",
			InstanceStatus: commonTypes.InstanceStatus{
				Code: 3,
				Msg:  "",
			},
			JobID:          "",
			SchedulerChain: []string(nil),
			ParentID:       "worker-manager",
		})
	})
	convey.Convey("etcd delete value failed", t, func() {
		event.Type = etcd3.DELETE
		event.Key = "/sn/workers/business/yrk/tenant/c53626012ba84727b938ca8bf03108ef/function//" +
			"version/latest/defaultaz/defaultaz-#-TEST"
		ir.watcherFGHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(string(msg.EventType), convey.ShouldEqual, "")
		convey.So(msg.EventMsg, convey.ShouldResemble, nil)
	})
	convey.Convey("etcd delete value success", t, func() {
		event.Key = "/sn/workers/business/yrk/tenant/c53626012ba84727b938ca8bf03108ef/function/0@default@zscaetest/" +
			"version/latest/defaultaz/defaultaz-#-custom-1000-1024-935e9454-93fa-43f1-b5e4-7cd82737dd62"
		event.Type = etcd3.DELETE
		ir.watcherFGHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg.EventType, convey.ShouldEqual, SubEventTypeDelete)
		convey.So(msg.EventMsg, convey.ShouldResemble, &commonTypes.InstanceSpecification{
			InstanceID:      "defaultaz-#-custom-1000-1024-935e9454-93fa-43f1-b5e4-7cd82737dd62",
			DataSystemHost:  "",
			RequestID:       "",
			RuntimeID:       "",
			RuntimeAddress:  "127.0.0.1:22423",
			FunctionAgentID: "",
			FunctionProxyID: "192.168.0.154:8080",
			Function:        "c53626012ba84727b938ca8bf03108ef/0@default@zscaetest/latest",
			RestartPolicy:   "",
			Resources: commonTypes.Resources{
				Resources: map[string]commonTypes.Resource{
					constant.ResourceCPUName: {
						Name: constant.ResourceCPUName,
						Scalar: commonTypes.ValueScalar{
							Value: float64(1000),
						},
					},
					constant.ResourceMemoryName: {
						Name: constant.ResourceMemoryName,
						Scalar: commonTypes.ValueScalar{
							Value: float64(1024),
						},
					},
					constant.ResourceEphemeralStorage: {
						Name: constant.ResourceEphemeralStorage,
						Scalar: commonTypes.ValueScalar{
							Value: float64(0),
						},
					},
				},
			},
			ActualUse: commonTypes.Resources{
				Resources: map[string]commonTypes.Resource(nil),
			},
			ScheduleOption: commonTypes.ScheduleOption{
				SchedPolicyName: "",
				Priority:        0,
				Affinity: commonTypes.Affinity{
					NodeAffinity: commonTypes.NodeAffinity{
						Affinity: map[string]string(nil),
					},
					InstanceAffinity: commonTypes.InstanceAffinity{
						Affinity: map[string]string(nil),
					},
					InstanceAntiAffinity: commonTypes.InstanceAffinity{
						Affinity: map[string]string(nil),
					},
				},
			},
			CreateOptions: map[string]string{
				"FUNCTION_KEY_NOTE":  "c53626012ba84727b938ca8bf03108ef/0@default@zscaetest/latest",
				"INSTANCE_TYPE_NOTE": "reserved",
				"SCHEDULER_ID_NOTE":  "worker-manager",
			},
			Labels:    []string(nil),
			StartTime: "1724393756",
			InstanceStatus: commonTypes.InstanceStatus{
				Code: 3,
				Msg:  "",
			},
			JobID:          "",
			SchedulerChain: []string(nil),
			ParentID:       "worker-manager",
		})
	})
	convey.Convey("etcd put invalid instanceID", t, func() {
		event.Type = etcd3.PUT
		event.Key = "/sn/workers/business/yrk/tenant/c53626012ba84727b938ca8bf03108ef/function/0@default@zscaetest/" +
			"version/latest/defaultaz//"
		ir.watcherFGHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg, convey.ShouldResemble, SubEvent{})
	})
	convey.Convey("etcd synced test", t, func() {
		event.Type = etcd3.SYNCED
		ir.watcherFGHandler(event)
		msg := SubEvent{}
		convey.So(msg, convey.ShouldResemble, SubEvent{})
	})
	close(stopCh)
}

func TestFunctionRegistryWatcherHandler(t *testing.T) {
	fr := &FunctionRegistry{
		userAgencyRegistry: &UserAgencyRegistry{},
		funcSpecs:          make(map[string]*types.FunctionSpecification),
		listDoneCh:         make(chan struct{}, 1),
	}
	recvMsg := make(chan SubEvent, 1)
	fr.addSubscriberChan(recvMsg)

	stopCh := make(chan struct{})
	GlobalRegistry = &Registry{
		FaaSSchedulerRegistry: NewFaasSchedulerRegistry(stopCh),
		FunctionRegistry:      NewFunctionRegistry(stopCh),
		InstanceRegistry:      NewInstanceRegistry(stopCh),
		FaaSManagerRegistry:   NewFaaSManagerRegistry(stopCh),
	}

	event := &etcd3.Event{
		Type:      etcd3.PUT,
		Key:       "/sn/functions/business/yrk/tenant/7e1ad6a6-cc5c-44fa-bd54-25873f72a86a/function/0@base@testresourcejava11128/version/latest",
		Value:     []byte("123"),
		PrevValue: []byte("123"),
		Rev:       1,
	}

	convey.Convey("etcd put value error", t, func() {
		event.Type = etcd3.PUT
		fr.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg, convey.ShouldResemble, SubEvent{})
	})
	convey.Convey("etcd put value success", t, func() {
		instanceSpecByte, _ := json.Marshal(&commonTypes.FunctionMetaInfo{})
		event.Value = instanceSpecByte
		event.Type = etcd3.PUT
		fr.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg.EventType, convey.ShouldEqual, SubEventTypeUpdate)
		convey.So(msg.EventMsg, convey.ShouldNotResemble, &commonTypes.FunctionMetaInfo{})
	})
	convey.Convey("etcd put exist value success", t, func() {
		instanceSpecByte, _ := json.Marshal(&commonTypes.FunctionMetaInfo{})
		event.Value = instanceSpecByte
		event.Type = etcd3.PUT
		fr.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg.EventType, convey.ShouldEqual, SubEventTypeUpdate)
		convey.So(msg.EventMsg, convey.ShouldNotResemble, &commonTypes.FunctionMetaInfo{})
	})

	convey.Convey("etcd delete value success", t, func() {
		event.Type = etcd3.DELETE
		fr.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg.EventType, convey.ShouldEqual, SubEventTypeDelete)
		convey.So(msg.EventMsg, convey.ShouldNotResemble, &commonTypes.FunctionMetaInfo{})
	})
	convey.Convey("etcd delete doesn't exist value error", t, func() {
		event.Type = etcd3.DELETE
		fr.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg, convey.ShouldResemble, SubEvent{})
		convey.So(fr.funcSpecs, convey.ShouldHaveLength, 0)
	})

	convey.Convey("etcd put invalid funcKey", t, func() {
		event.Type = etcd3.PUT
		event.Key = "/sn/functions/busi"
		fr.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg, convey.ShouldResemble, SubEvent{})
	})
	convey.Convey("etcd SYNCED", t, func() {
		event.Type = etcd3.SYNCED
		fr.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg.EventType, convey.ShouldEqual, SubEventTypeSynced)
	})
	convey.Convey("getFuncSpec", t, func() {
		res := fr.getFuncSpec("")
		convey.So(res, convey.ShouldBeNil)
	})
	convey.Convey("fetchSilentFuncSpec", t, func() {
		defer ApplyFunc(etcd3.GetValueFromEtcdWithRetry,
			func(key string, etcdClient *etcd3.EtcdClient) ([]byte, error) {
				funcSpec := types.FunctionSpecification{}
				spec, _ := json.Marshal(funcSpec)
				return spec, nil
			}).Reset()
		defer ApplyPrivateMethod(reflect.TypeOf(&FunctionRegistry{}), "buildFuncSpec",
			func(_ *FunctionRegistry, etcdKey string, etcdValue []byte,
				funcKey string) *types.FunctionSpecification {
				return &types.FunctionSpecification{
					FuncKey: "1234/test-func/latest",
				}
			}).Reset()
		res := fr.fetchSilentFuncSpec("1234/test-func/latest")
		convey.So(res.FuncKey, convey.ShouldEqual, "1234/test-func/latest")
	})
	convey.Convey("EtcdList", t, func() {
		client, _ := clientv3.New(clientv3.Config{Endpoints: []string{"127.0.0.1"}})
		etcdClient := &etcd3.EtcdClient{
			Client: client,
		}
		defer ApplyFunc(etcd3.GetMetaEtcdClient, func() *etcd3.EtcdClient {
			return etcdClient
		}).Reset()
		defer ApplyMethodFunc(client, "Get", func(ctx context.Context,
			key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
			getRsp := &clientv3.GetResponse{
				Kvs: []*mvccpb.KeyValue{
					{
						Key:   []byte("/sn/functions/business/yrk/tenant/0/function/0-system-faasscheduler/version/$latest"),
						Value: []byte("{}"),
					},
				},
			}
			return getRsp, nil
		}).Reset()
		funcList := fr.EtcdList()
		convey.So(len(funcList), convey.ShouldEqual, 0)
	})
}

func TestFaaSManagerRegistryWatcherHandler(t *testing.T) {
	fr := &FaaSManagerRegistry{}
	recvMsg := make(chan SubEvent, 1)
	fr.addSubscriberChan(recvMsg)

	stopCh := make(chan struct{})
	GlobalRegistry = &Registry{
		FaaSSchedulerRegistry: NewFaasSchedulerRegistry(stopCh),
		FunctionRegistry:      NewFunctionRegistry(stopCh),
		InstanceRegistry:      NewInstanceRegistry(stopCh),
		FaaSManagerRegistry:   NewFaaSManagerRegistry(stopCh),
	}

	event := &etcd3.Event{
		Type:      etcd3.PUT,
		Key:       "/sn/instance/business/yrk/tenant/12345678901234561234567890123456/function/0-system-faasExecutorJava8/version/$latest/defaultaz/task-b23aa1c4-2084-42b8-99b2-8907fa5ae6f4/f71875b1-3c20-4827-8600-0000000005d5",
		Value:     []byte("123"),
		PrevValue: []byte("123"),
		Rev:       1,
	}
	convey.Convey("etcd put value error", t, func() {
		event.Type = etcd3.PUT
		fr.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg, convey.ShouldResemble, SubEvent{})
	})
	convey.Convey("etcd put value success", t, func() {
		instanceSpecByte, _ := json.Marshal(&commonTypes.InstanceSpecification{Labels: []string{""}})
		event.Value = instanceSpecByte
		event.Type = etcd3.PUT
		fr.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg.EventType, convey.ShouldEqual, SubEventTypeUpdate)
		convey.So(msg.EventMsg, convey.ShouldResemble, &commonTypes.InstanceSpecification{
			InstanceID: "f71875b1-3c20-4827-8600-0000000005d5",
			Labels:     []string{""},
		})
	})
	convey.Convey("etcd delete value error", t, func() {
		event.Type = etcd3.DELETE
		fr.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg, convey.ShouldResemble, SubEvent{})
	})
	convey.Convey("etcd delete value success", t, func() {
		instanceSpecByte, _ := json.Marshal(&commonTypes.InstanceSpecification{Labels: []string{""}})
		event.PrevValue = instanceSpecByte
		event.Type = etcd3.DELETE
		fr.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg.EventType, convey.ShouldEqual, SubEventTypeDelete)
		convey.So(msg.EventMsg, convey.ShouldResemble, &commonTypes.InstanceSpecification{
			InstanceID: "f71875b1-3c20-4827-8600-0000000005d5",
			Labels:     []string{""},
		})
	})
	convey.Convey("etcd put invalid funcKey", t, func() {
		event.Type = etcd3.PUT
		event.Key = "/sn/instance/business/yrk/tenant//function"
		fr.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg, convey.ShouldResemble, SubEvent{})
	})
	convey.Convey("etcd SYNCED", t, func() {
		event.Type = etcd3.SYNCED
		fr.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg, convey.ShouldResemble, SubEvent{})
	})
}

func TestStartRegistry(t *testing.T) {
	stopCh := make(chan struct{})
	GlobalRegistry = &Registry{
		FaaSSchedulerRegistry: NewFaasSchedulerRegistry(stopCh),
		FunctionRegistry:      NewFunctionRegistry(stopCh),
		InstanceRegistry:      NewInstanceRegistry(stopCh),
		FaaSManagerRegistry:   NewFaaSManagerRegistry(stopCh),
	}
	ew := &etcd3.EtcdWatcher{}
	GlobalRegistry.FaaSSchedulerRegistry.schedulerInstanceWatcher = ew
	GlobalRegistry.FaaSSchedulerRegistry.schedulerHashWatcher = ew
	GlobalRegistry.FunctionRegistry.watcher = ew
	GlobalRegistry.FunctionRegistry.userAgencyRegistry = &UserAgencyRegistry{watcher: ew}
	GlobalRegistry.InstanceRegistry.watcher = ew
	GlobalRegistry.InstanceRegistry.fgWatcher = ew
	GlobalRegistry.FaaSManagerRegistry.watcher = ew
	config.GlobalConfig = types.Configuration{}

	watch := false
	defer ApplyMethod(reflect.TypeOf(ew), "StartWatch", func(_ *etcd3.EtcdWatcher) {
		watch = true
	}).Reset()

	convey.Convey("start success", t, func() {
		StartRegistry()
		time.Sleep(100 * time.Millisecond)
		convey.So(watch, convey.ShouldBeTrue)
	})
	convey.Convey("start success", t, func() {
		GlobalRegistry.FunctionRegistry = nil
		GlobalRegistry.InstanceRegistry = nil
		GlobalRegistry.FaaSManagerRegistry = nil
		watch = false
		StartRegistry()
		time.Sleep(100 * time.Millisecond)
		convey.So(watch, convey.ShouldBeTrue)
	})
	close(stopCh)
}

func TestMiscellaneous(t *testing.T) {
	/*convey.Convey("ProcessETCDList", t, func() {
		count := 0
		defer ApplyMethod(reflect.TypeOf(&etcd3.EtcdWatcher{}), "StartList", func(_ *etcd3.EtcdWatcher) {
			count++
		}).Reset()
		stopCh := make(chan struct{})
		GlobalRegistry = &Registry{
			FaaSSchedulerRegistry:     NewFaasSchedulerRegistry(stopCh),
			FunctionRegistry:          NewFunctionRegistry(stopCh),
			InstanceRegistry:          NewInstanceRegistry(stopCh),
			FaaSManagerRegistry:       NewFaaSManagerRegistry(stopCh),
			InstanceConfigRegistry:    NewInstanceConfigRegistry(stopCh),
			AliasRegistry:             NewAliasRegistry(stopCh),
			FunctionAvailableRegistry: NewFunctionAvailableRegistry(stopCh),
			FaaSFrontendRegistry:      NewFaaSFrontendRegistry(stopCh),
			TenantQuotaRegistry:       NewTenantQuotaRegistry(stopCh),
		}
		ProcessETCDList()
		convey.So(count, convey.ShouldEqual, 8)
	})*/
	convey.Convey("FunctionRegistry watcherFilterForConfig", t, func() {
		stopCh := make(chan struct{})
		fr := NewFunctionRegistry(stopCh)
		convey.Convey("success", func() {
			res := fr.watcherFilter(&etcd3.Event{
				Key: "/sn/functions/business/yrk/tenant/7e1ad6a6-cc5c-44fa-bd54-25873f72a86a/function/0@base@testresourcejava11128/version/latest",
			})
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("failed1", func() {
			res := fr.watcherFilter(&etcd3.Event{
				Key: "/sn/functions/business/yrk/tenant/7e1ad6a6-cc5c-44fa-bd5",
			})
			convey.So(res, convey.ShouldBeTrue)
		})
		convey.Convey("failed2", func() {
			res := fr.watcherFilter(&etcd3.Event{
				Key: "/sn/instance/business/yrk/tenant/7e1ad6a6-cc5c-44fa-bd54-25873f72a86a/function/0@base@testresourcejava11128/version/latest",
			})
			convey.So(res, convey.ShouldBeTrue)
		})
	})
	convey.Convey("FaaSManagerRegistry watcherFilterForConfig", t, func() {
		stopCh := make(chan struct{})
		fm := NewFaaSManagerRegistry(stopCh)
		convey.Convey("success", func() {
			res := fm.watcherFilter(&etcd3.Event{
				Key: "/sn/instance/business/yrk/tenant/1234567890123456/function/0-system-faasmanager/version/$latest/defaultaz/task-29eea890-fd17/a16e7302-0000-4000-80de-84e02e5d6717",
			})
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("failed1", func() {
			res := fm.watcherFilter(&etcd3.Event{
				Key: "/sn/instance/business/yrk/302-0000-4000-80de-84e02e5d6717",
			})
			convey.So(res, convey.ShouldBeTrue)
		})
		convey.Convey("failed2", func() {
			res := fm.watcherFilter(&etcd3.Event{
				Key: "/sn/function/business/yrk/tenant/1234567890123456/function/0-system-faasmanager/version/$latest/defaultaz/task-29eea890-fd17/a16e7302-0000-4000-80de-84e02e5d6717",
			})
			convey.So(res, convey.ShouldBeTrue)
		})
	})

	convey.Convey("*FaaSSchedulerRegistry watcherFilterForConfig", t, func() {
		stopCh := make(chan struct{})
		fsr := NewFaasSchedulerRegistry(stopCh)
		convey.Convey("instance success", func() {
			res := fsr.schedulerInstanceFilter(&etcd3.Event{
				Key: "/sn/instance/business/yrk/tenant/1234567890123456/function/0-system-faasscheduler/version/$latest/defaultaz/task-29eea890-fd17/a16e7302-0000-4000-80de-84e02e5d6717",
			})
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("instance failed1", func() {
			res := fsr.schedulerInstanceFilter(&etcd3.Event{
				Key: "/sn/instance/business/yrk/302-0000-4000-80de-84e02e5d6717",
			})
			convey.So(res, convey.ShouldBeTrue)
		})
		convey.Convey("module success", func() {
			res := fsr.schedulerHashFilter(&etcd3.Event{
				Key: "/sn/faas-scheduler/instances/cluster001/7.218.100.25/faas-scheduler-59ddbc4b75-8xdjf",
			})
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("module failed1", func() {
			res := fsr.schedulerHashFilter(&etcd3.Event{
				Key: "/sn/instance/business/yrk/302-0000-4000-80de-84e02e5d6717",
			})
			convey.So(res, convey.ShouldBeTrue)
		})
	})

	convey.Convey("InstancesInfoRegistry watcherFilterForConfig", t, func() {
		convey.Convey("success", func() {
			config.GlobalConfig = types.Configuration{ClusterID: "cluster001"}
			res := instanceconfig.GetWatcherFilter("cluster001")(&etcd3.Event{
				Key: "/instances/business/yrk/cluster/cluster001/tenant/12345678901234561234567890123456/function/0@yrservice@test-faas-scheduler-reserved-exist/version/$latest",
			})
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("failed1", func() {
			config.GlobalConfig = types.Configuration{ClusterID: "cluster001"}
			res := instanceconfig.GetWatcherFilter("cluster001")(&etcd3.Event{
				Key: "/instances/business/yrk/cluster/cluster001/tenant/12345678901234561234567890123456",
			})
			convey.So(res, convey.ShouldBeTrue)
		})
	})
}

func TestInstancesInfoRegistryWatcherHandler(t *testing.T) {
	stopCh := make(chan struct{})
	GlobalRegistry = &Registry{
		FaaSSchedulerRegistry:  NewFaasSchedulerRegistry(stopCh),
		FunctionRegistry:       NewFunctionRegistry(stopCh),
		InstanceRegistry:       NewInstanceRegistry(stopCh),
		FaaSManagerRegistry:    NewFaaSManagerRegistry(stopCh),
		InstanceConfigRegistry: NewInstanceConfigRegistry(stopCh),
	}

	recvMsg := make(chan SubEvent, 1)
	ifr := &InstanceConfigRegistry{
		listDoneCh: make(chan struct{}, 1),
	}
	ifr.addSubscriberChan(recvMsg)

	event := &etcd3.Event{
		Type:      etcd3.PUT,
		Key:       "/instances/business/yrk/cluster/cluster001/tenant/12345678901234561234567890123456/function/0@yrservice@test-faas-scheduler-reserved-exist/version/$latest",
		Value:     []byte("123"),
		PrevValue: []byte("123"),
		Rev:       1,
	}

	Patches := ApplyMethod(reflect.TypeOf(&selfregister.SchedulerProxy{}), "IsFuncOwner",
		func(_ *selfregister.SchedulerProxy, funcKey string) bool {
			return true
		})

	convey.Convey("etcd put value error", t, func() {
		event.Type = etcd3.PUT
		ifr.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg, convey.ShouldResemble, SubEvent{})
	})
	convey.Convey("etcd put value success", t, func() {
		instanceSpecByte, _ := json.Marshal(&instanceconfig.Configuration{})
		event.Value = instanceSpecByte
		event.Type = etcd3.PUT
		ifr.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg.EventType, convey.ShouldEqual, SubEventTypeUpdate)
		convey.So(msg.EventMsg, convey.ShouldResemble, &instanceconfig.Configuration{
			FuncKey: "12345678901234561234567890123456/0@yrservice@test-faas-scheduler-reserved-exist/$latest",
		})
	})
	convey.Convey("etcd delete value success", t, func() {
		instanceSpecByte, _ := json.Marshal(&instanceconfig.Configuration{})
		event.PrevValue = instanceSpecByte
		event.Type = etcd3.DELETE
		ifr.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}

		convey.So(msg.EventType, convey.ShouldEqual, SubEventTypeDelete)
		convey.So(msg.EventMsg, convey.ShouldResemble, &instanceconfig.Configuration{
			FuncKey: "12345678901234561234567890123456/0@yrservice@test-faas-scheduler-reserved-exist/$latest",
		})
	})
	convey.Convey("IsFuncOwner does not allow", t, func() {
		Patches.Reset()
		event.Type = etcd3.PUT
		ifr.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg.EventMsg.(*instanceconfig.Configuration).InstanceMetaData.MinInstance, convey.ShouldEqual, 0)
	})
	convey.Convey("etcd put invalid funcKey", t, func() {
		event.Type = etcd3.PUT
		event.Key = "/instances/business/yrk/cluster/cluster001/tenant/12345678901234561234567890123456"
		ifr.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg, convey.ShouldResemble, SubEvent{})
	})
	convey.Convey("etcd SYNCED", t, func() {
		event.Type = etcd3.SYNCED
		ifr.watcherHandler(event)
		msg := SubEvent{}
		select {
		case msg = <-recvMsg:
		default:
		}
		convey.So(msg, convey.ShouldResemble, SubEvent{SubEventTypeSynced, &instanceconfig.Configuration{}})
	})
}

func TestWaitForHash(t *testing.T) {
	sp := &selfregister.SchedulerProxy{}
	sp.FaaSSchedulers.Store("instance121", "scheduler1")
	sp.FaaSSchedulers.Store("instance122", "scheduler1")
	sp.FaaSSchedulers.Store("instance123", "scheduler1")
	cnt := 0 * selfregister.GetHashLenInternal
	defer ApplyFunc(time.Sleep, func(d time.Duration) {
		cnt += d
	}).Reset()
	convey.Convey("WaitForHash", t, func() {
		sp.WaitForHash(0)
		convey.So(cnt, convey.ShouldEqual, 0)
	})
	convey.Convey("WaitForHash", t, func() {
		go sp.WaitForHash(4)
		time.Sleep(selfregister.GetHashLenInternal * 2)
		sp.FaaSSchedulers.Store("instance124", "scheduler1")
		convey.So(cnt, convey.ShouldBeGreaterThan, 0*selfregister.GetHashLenInternal)
	})
}

func TestAliasRegistry_RunWatcher(t *testing.T) {
	convey.Convey("AliasRegistry", t, func() {
		GlobalRegistry = nil
		stopCh := make(chan struct{})
		patches := []*Patches{
			ApplyMethod(reflect.TypeOf(&etcd3.EtcdWatcher{}), "StartList", func(ew *etcd3.EtcdWatcher) {
				ew.ResultChan <- &etcd3.Event{
					Type:      etcd3.SYNCED,
					Key:       "",
					Value:     nil,
					PrevValue: nil,
					Rev:       0,
					ETCDType:  "",
				}
			}),
			ApplyMethod(reflect.TypeOf(&etcd3.EtcdWatcher{}), "EtcdHistory", func(ew *etcd3.EtcdWatcher,
				revision int64) {
			}),
			ApplyFunc((*FunctionRegistry).WaitForETCDList, func() {}),
			ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient { return &etcd3.EtcdClient{} }),
			ApplyFunc(etcd3.GetMetaEtcdClient, func() *etcd3.EtcdClient { return &etcd3.EtcdClient{} }),
			ApplyMethod(reflect.TypeOf(&etcd3.EtcdWatcher{}), "StartWatch", func(ew *etcd3.EtcdWatcher) {}),
		}
		defer func() {
			for _, p := range patches {
				p.Reset()
			}
		}()
		os.Setenv(constant.EnableAgentCRDRegistry, "")
		GlobalRegistry = &Registry{
			FaaSSchedulerRegistry:     NewFaasSchedulerRegistry(stopCh),
			FunctionRegistry:          NewFunctionRegistry(stopCh),
			InstanceRegistry:          NewInstanceRegistry(stopCh),
			FaaSManagerRegistry:       NewFaaSManagerRegistry(stopCh),
			InstanceConfigRegistry:    NewInstanceConfigRegistry(stopCh),
			AliasRegistry:             NewAliasRegistry(stopCh),
			FunctionAvailableRegistry: NewFunctionAvailableRegistry(stopCh),
			FaaSFrontendRegistry:      NewFaaSFrontendRegistry(stopCh),
			TenantQuotaRegistry:       NewTenantQuotaRegistry(stopCh),
			RolloutRegistry:           NewRolloutRegistry(stopCh),
			AgentRegistry:             NewAgentRegistry(stopCh),
		}
		ProcessETCDList()
		convey.Convey("update", func() {
			defer ApplyMethod(reflect.TypeOf(&etcd3.EtcdWatcher{}), "StartWatch",
				func(ew *etcd3.EtcdWatcher) {
					alias := &aliasroute.AliasElement{AliasURN: "123456"}
					bytes, _ := json.Marshal(alias)
					ew.ResultChan <- &etcd3.Event{
						Type:      etcd3.PUT,
						Key:       "/sn/aliases/xxx/xxx/tenant/1234567890/function/functionName/requestID",
						Value:     bytes,
						PrevValue: nil,
						Rev:       0,
					}
				}).Reset()
			GlobalRegistry.AliasRegistry.RunWatcher()
			ch := make(chan SubEvent, 1)
			GlobalRegistry.AliasRegistry.addSubscriberChan(ch)
			envet := <-ch
			convey.So(envet.EventType, convey.ShouldEqual, SubEventTypeUpdate)
			convey.So(envet.EventMsg.(string), convey.ShouldEqual, "123456")
		})

		convey.Convey("delete", func() {
			defer ApplyMethod(reflect.TypeOf(&etcd3.EtcdWatcher{}), "StartWatch",
				func(ew *etcd3.EtcdWatcher) {
					alias := &aliasroute.AliasElement{AliasURN: "123456"}
					bytes, _ := json.Marshal(alias)
					ew.ResultChan <- &etcd3.Event{
						Type:      etcd3.DELETE,
						Key:       "/sn/aliases/xxx/xxx/tenant/1234567890/function/functionName/requestID",
						Value:     bytes,
						PrevValue: nil,
						Rev:       0,
					}
				}).Reset()
			GlobalRegistry.AliasRegistry.RunWatcher()
			ch := make(chan SubEvent, 1)
			GlobalRegistry.AliasRegistry.addSubscriberChan(ch)
			envet := <-ch
			convey.So(envet.EventType, convey.ShouldEqual, SubEventTypeDelete)
			convey.So(envet.EventMsg.(string), convey.ShouldEqual,
				"sn:cn:xxx:1234567890:function:functionName:requestID")
		})

		convey.Convey("error", func() {
			defer ApplyMethod(reflect.TypeOf(&etcd3.EtcdWatcher{}), "StartWatch",
				func(ew *etcd3.EtcdWatcher) {
					alias := &aliasroute.AliasElement{AliasURN: "123456"}
					bytes, _ := json.Marshal(alias)
					ew.ResultChan <- &etcd3.Event{
						Type:      etcd3.ERROR,
						Key:       "/sn/aliases/xxx/xxx/tenant/1234567890/function/functionName/requestID",
						Value:     bytes,
						PrevValue: nil,
						Rev:       0,
					}
					ew.ResultChan <- &etcd3.Event{
						Type:      4,
						Key:       "/sn/aliases/xxx/xxx/tenant/1234567890/function/functionName/requestID",
						Value:     bytes,
						PrevValue: nil,
						Rev:       0,
					}
				}).Reset()
			GlobalRegistry.AliasRegistry.RunWatcher()
			ch := make(chan SubEvent, 1)
			GlobalRegistry.AliasRegistry.addSubscriberChan(ch)
			convey.So(len(ch), convey.ShouldEqual, 0)
		})
	})
}

func TestGetSchedulerInfo(t *testing.T) {
	convey.Convey("GetAllSchedulerInfo", t, func() {
		schedulerRegistry := NewFaasSchedulerRegistry(make(chan struct{}))
		selfregister.GlobalSchedulerProxy.Add(&commonTypes.InstanceInfo{InstanceName: "scheduler1",
			InstanceID: "scheduler1-id"}, "")
		info := schedulerRegistry.GetAllSchedulerInfo()
		convey.So(info.SchedulerIDList[0], convey.ShouldEqual, "scheduler1")
		convey.So(info.SchedulerInstanceList[0].InstanceID, convey.ShouldEqual, "scheduler1-id")
		convey.So(info.SchedulerInstanceList[0].InstanceName, convey.ShouldEqual, "scheduler1")
	})
}

func TestFaaSFrontendRegistry_RunWatcher(t *testing.T) {
	convey.Convey("FaaSFrontendRegistry", t, func() {
		GlobalRegistry = nil
		stopCh := make(chan struct{})
		patches := []*Patches{
			ApplyMethod(reflect.TypeOf(&etcd3.EtcdWatcher{}), "StartList", func(ew *etcd3.EtcdWatcher) {
				ew.ResultChan <- &etcd3.Event{
					Type:      etcd3.SYNCED,
					Key:       "",
					Value:     nil,
					PrevValue: nil,
					Rev:       0,
					ETCDType:  "",
				}
			}),
			ApplyFunc((*FunctionRegistry).WaitForETCDList, func() {}),
			ApplyMethod(reflect.TypeOf(&etcd3.EtcdWatcher{}), "EtcdHistory", func(ew *etcd3.EtcdWatcher,
				revision int64) {
			}),
			ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient { return &etcd3.EtcdClient{} }),
			ApplyFunc(etcd3.GetMetaEtcdClient, func() *etcd3.EtcdClient { return &etcd3.EtcdClient{} }),
			ApplyMethod(reflect.TypeOf(&etcd3.EtcdWatcher{}), "StartWatch", func(ew *etcd3.EtcdWatcher) {}),
		}
		defer func() {
			for _, p := range patches {
				p.Reset()
			}
		}()
		GlobalRegistry = &Registry{
			FaaSSchedulerRegistry:     NewFaasSchedulerRegistry(stopCh),
			FunctionRegistry:          NewFunctionRegistry(stopCh),
			InstanceRegistry:          NewInstanceRegistry(stopCh),
			FaaSManagerRegistry:       NewFaaSManagerRegistry(stopCh),
			InstanceConfigRegistry:    NewInstanceConfigRegistry(stopCh),
			AliasRegistry:             NewAliasRegistry(stopCh),
			FunctionAvailableRegistry: NewFunctionAvailableRegistry(stopCh),
			FaaSFrontendRegistry:      NewFaaSFrontendRegistry(stopCh),
			TenantQuotaRegistry:       NewTenantQuotaRegistry(stopCh),
			RolloutRegistry:           NewRolloutRegistry(stopCh),
		}
		ProcessETCDList()
		convey.Convey("update", func() {
			errEv := &etcd3.Event{
				Type:      etcd3.PUT,
				Key:       "/sn/frontend/instances/cluster001/127.0.0.1/frontend-768df8f66b-gvz4z",
				Value:     []byte(`aaa`),
				PrevValue: nil,
				Rev:       0,
			}
			GlobalRegistry.FaaSFrontendRegistry.watcherHandler(errEv)
			assert.Equal(t, len(GlobalRegistry.FaaSFrontendRegistry.ClusterFrontends), 0)

			errEv1 := &etcd3.Event{
				Type:      etcd3.PUT,
				Key:       "/sn/frontend/instances/cluster001/127.0.0.1/frontend-768df8f66b-gvz4z/latest",
				Value:     []byte(`active`),
				PrevValue: nil,
				Rev:       0,
			}
			GlobalRegistry.FaaSFrontendRegistry.watcherHandler(errEv1)
			GlobalRegistry.FaaSFrontendRegistry.watcherFilter(errEv1)
			assert.Equal(t, len(GlobalRegistry.FaaSFrontendRegistry.ClusterFrontends), 0)

			ev := &etcd3.Event{
				Type:      etcd3.PUT,
				Key:       "/sn/frontend/instances/cluster001/127.0.0.1/frontend-768df8f66b-gvz4z",
				Value:     []byte(`active`),
				PrevValue: nil,
				Rev:       0,
			}
			GlobalRegistry.FaaSFrontendRegistry.watcherHandler(ev)
			ev1 := &etcd3.Event{
				Type:      etcd3.PUT,
				Key:       "/sn/frontend/instances/cluster001/127.0.0.1/frontend-768df8f66b-gvz4z",
				Value:     []byte(`active`),
				PrevValue: nil,
				Rev:       0,
			}
			GlobalRegistry.FaaSFrontendRegistry.watcherHandler(ev1)
			assert.Equal(t, len(GlobalRegistry.FaaSFrontendRegistry.ClusterFrontends["cluster001"]), 2)
			assert.Equal(t, GlobalRegistry.FaaSFrontendRegistry.GetFrontends("cluster001"),
				[]string{"127.0.0.1", "127.0.0.1"})

			ev2 := &etcd3.Event{
				Type:      etcd3.PUT,
				Key:       "/sn/frontend/instances/cluster001/127.0.0.1/frontend-768df8f66b-gvz4z",
				Value:     []byte(``),
				PrevValue: nil,
				Rev:       0,
			}
			GlobalRegistry.FaaSFrontendRegistry.watcherHandler(ev2)
			assert.Equal(t, len(GlobalRegistry.FaaSFrontendRegistry.ClusterFrontends["cluster001"]), 1)
			assert.Equal(t, GlobalRegistry.FaaSFrontendRegistry.GetFrontends("cluster001"), []string{"127.0.0.1"})
			ev3 := &etcd3.Event{
				Type:      etcd3.PUT,
				Key:       "/sn/frontend/instances/cluster001/127.0.0.1/frontend-768df8f66b-gvz4z",
				Value:     []byte(``),
				PrevValue: nil,
				Rev:       0,
			}
			GlobalRegistry.FaaSFrontendRegistry.watcherHandler(ev3)
			assert.Equal(t, len(GlobalRegistry.FaaSFrontendRegistry.ClusterFrontends), 0)
			assert.Equal(t, GlobalRegistry.FaaSFrontendRegistry.GetFrontends("cluster001"), []string(nil))
		})

		convey.Convey("delete", func() {
			ev := &etcd3.Event{
				Key:   "/sn/frontend/instances/cluster001/127.0.0.1/frontend-768df8f66b-gvz4z",
				Value: []byte(`active`),
			}
			GlobalRegistry.FaaSFrontendRegistry.updateFrontendInstances(ev)
			assert.Equal(t, len(GlobalRegistry.FaaSFrontendRegistry.ClusterFrontends), 1)

			deleteErrEv := &etcd3.Event{
				Type:      etcd3.DELETE,
				Key:       "/sn/frontend/instances/cluster001/127.0.0.1/frontend-768df8f66b-gvz4z",
				Value:     []byte(`active`),
				PrevValue: nil,
				Rev:       0,
			}
			GlobalRegistry.FaaSFrontendRegistry.watcherHandler(deleteErrEv)
			assert.Equal(t, len(GlobalRegistry.FaaSFrontendRegistry.ClusterFrontends), 1)

			deleteErrEv1 := &etcd3.Event{
				Key:   "/sn/frontend/instances/cluster001/frontend-768df8f66b-gvz4z",
				Value: []byte(`active`),
			}
			GlobalRegistry.FaaSFrontendRegistry.deleteFrontendInstances(deleteErrEv1)
			assert.Equal(t, len(GlobalRegistry.FaaSFrontendRegistry.ClusterFrontends), 1)

			deleteEv := &etcd3.Event{
				Type:      etcd3.DELETE,
				Key:       "/sn/frontend/instances/cluster001/127.0.0.1/frontend-768df8f66b-gvz4z",
				Value:     []byte(`active`),
				PrevValue: nil,
				Rev:       0,
			}
			GlobalRegistry.FaaSFrontendRegistry.watcherHandler(deleteEv)
			assert.Equal(t, len(GlobalRegistry.FaaSFrontendRegistry.ClusterFrontends), 0)
		})
	})
}

func TestFunctionAvailableRegistry_RunWatcher(t *testing.T) {
	convey.Convey("FunctionAvailableRegistry", t, func() {
		GlobalRegistry = nil
		stopCh := make(chan struct{})
		patches := []*Patches{
			ApplyMethod(reflect.TypeOf(&etcd3.EtcdWatcher{}), "StartList", func(ew *etcd3.EtcdWatcher) {
				ew.ResultChan <- &etcd3.Event{
					Type:      etcd3.SYNCED,
					Key:       "",
					Value:     nil,
					PrevValue: nil,
					Rev:       0,
					ETCDType:  "",
				}
			}),
			ApplyFunc((*FunctionRegistry).WaitForETCDList, func() {}),
			ApplyMethod(reflect.TypeOf(&etcd3.EtcdWatcher{}), "EtcdHistory", func(ew *etcd3.EtcdWatcher,
				revision int64) {
			}),
			ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient { return &etcd3.EtcdClient{} }),
			ApplyFunc(etcd3.GetMetaEtcdClient, func() *etcd3.EtcdClient { return &etcd3.EtcdClient{} }),
			ApplyMethod(reflect.TypeOf(&etcd3.EtcdWatcher{}), "StartWatch", func(ew *etcd3.EtcdWatcher) {}),
		}
		defer func() {
			for _, p := range patches {
				p.Reset()
			}
		}()
		GlobalRegistry = &Registry{
			FaaSSchedulerRegistry:     NewFaasSchedulerRegistry(stopCh),
			FunctionRegistry:          NewFunctionRegistry(stopCh),
			InstanceRegistry:          NewInstanceRegistry(stopCh),
			FaaSManagerRegistry:       NewFaaSManagerRegistry(stopCh),
			InstanceConfigRegistry:    NewInstanceConfigRegistry(stopCh),
			AliasRegistry:             NewAliasRegistry(stopCh),
			FunctionAvailableRegistry: NewFunctionAvailableRegistry(stopCh),
			FaaSFrontendRegistry:      NewFaaSFrontendRegistry(stopCh),
			TenantQuotaRegistry:       NewTenantQuotaRegistry(stopCh),
			RolloutRegistry:           NewRolloutRegistry(stopCh),
			AgentRegistry:             NewAgentRegistry(stopCh),
		}
		ProcessETCDList()
		convey.Convey("update and delete", func() {
			errEv := &etcd3.Event{
				Type:      etcd3.PUT,
				Key:       "/sn/function/available/clusters/sn:cn:yrk:580943580943580943:function:0@debugservice@hello-world/latest",
				Value:     []byte(`["cluster1", "cluster2"]`),
				PrevValue: nil,
				Rev:       0,
			}
			GlobalRegistry.FunctionAvailableRegistry.watcherHandler(errEv)
			assert.Equal(t, len(GlobalRegistry.FunctionAvailableRegistry.FuncAvailableClusters), 0)

			errEv1 := &etcd3.Event{
				Type:      etcd3.PUT,
				Key:       "/sn/function/available/clusters/sn:cn:yrk:580943580943580943:function:0@debugservice@hello-world",
				Value:     []byte("aaa"),
				PrevValue: nil,
				Rev:       0,
			}
			GlobalRegistry.FunctionAvailableRegistry.watcherHandler(errEv1)
			assert.Equal(t, len(GlobalRegistry.FunctionAvailableRegistry.FuncAvailableClusters), 0)

			ev := &etcd3.Event{
				Type:      etcd3.PUT,
				Key:       "/sn/function/available/clusters/sn:cn:yrk:580943580943580943:function:0@debugservice@hello-world",
				Value:     []byte(`["cluster1", "cluster2"]`),
				PrevValue: nil,
				Rev:       0,
			}
			GlobalRegistry.FunctionAvailableRegistry.watcherHandler(ev)
			clusters := []string{"cluster1", "cluster2"}
			assert.Equal(t,
				GlobalRegistry.FunctionAvailableRegistry.GeClusters("sn:cn:yrk:580943580943580943:function:0@debugservice@hello-world"),
				clusters)

			deleteEv := &etcd3.Event{
				Type:      etcd3.DELETE,
				Key:       "/sn/function/available/clusters/sn:cn:yrk:580943580943580943:function:0@debugservice@hello-world",
				Value:     []byte(`["cluster1", "cluster2"]`),
				PrevValue: nil,
				Rev:       0,
			}
			GlobalRegistry.FunctionAvailableRegistry.watcherHandler(deleteEv)
			assert.Equal(t, len(GlobalRegistry.FunctionAvailableRegistry.FuncAvailableClusters), 0)
		})
	})
}

func Test_handleDefaultQuotaEvent(t *testing.T) {
	tenantMetaValueForTest := `{"tenantInstanceMetaData": {"maxOnDemandInstance": 1000,"maxReversedInstance": 1000}}`
	type args struct {
		event *etcd3.Event
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "case_01 invalid event",
			args: args{
				event: &etcd3.Event{},
			},
			wantErr: true,
		}, {
			name: "case_02 unmarshal failed",
			args: args{
				event: &etcd3.Event{
					Type:      etcd3.PUT,
					Key:       "/sn/quota/cluster/cluster001/default/instancemetadata",
					Value:     []byte(""),
					PrevValue: nil,
					Rev:       0,
				},
			},
			wantErr: true,
		}, {
			name: "case_03 update event",
			args: args{
				event: &etcd3.Event{
					Type:      etcd3.PUT,
					Key:       "/sn/quota/cluster/cluster001/default/instancemetadata",
					Value:     []byte(tenantMetaValueForTest),
					PrevValue: nil,
					Rev:       0,
				},
			},
			wantErr: false,
		}, {
			name: "case_04 delete event",
			args: args{
				event: &etcd3.Event{
					Type:      etcd3.DELETE,
					Key:       "/sn/quota/cluster/cluster001/default/instancemetadata",
					Value:     nil,
					PrevValue: []byte(tenantMetaValueForTest),
					Rev:       0,
				},
			},
			wantErr: false,
		}, {
			name: "case_05 unkown event",
			args: args{
				event: &etcd3.Event{
					Type:      etcd3.ERROR,
					Key:       "/sn/quota/cluster/cluster001/default/instancemetadata",
					Value:     []byte(tenantMetaValueForTest),
					PrevValue: nil,
					Rev:       0,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := handleDefaultQuotaEvent(tt.args.event); (err != nil) != tt.wantErr {
				t.Errorf("handleDefaultQuotaEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_handleTenantQuotaEvent(t *testing.T) {
	tenantMetaValueForTest := `{"tenantInstanceMetaData": {"maxOnDemandInstance": 1000,"maxReversedInstance": 1000}}`
	type args struct {
		event    *etcd3.Event
		tenantID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "case_01 invalid event",
			args: args{
				event:    &etcd3.Event{},
				tenantID: "test",
			},
			wantErr: true,
		}, {
			name: "case_02 unmarshal failed",
			args: args{
				event: &etcd3.Event{
					Type:      etcd3.PUT,
					Key:       "/sn/quota/cluster/cluster001/tenant/7e1ad6a6-cc5c-44fa-bd54-25873f72a86a/instancemetadata",
					Value:     []byte(""),
					PrevValue: nil,
					Rev:       0,
				},
				tenantID: "test",
			},
			wantErr: true,
		}, {
			name: "case_04 update event",
			args: args{
				event: &etcd3.Event{
					Type:      etcd3.PUT,
					Key:       "/sn/quota/cluster/cluster001/tenant/7e1ad6a6-cc5c-44fa-bd54-25873f72a86a/instancemetadata",
					Value:     []byte(tenantMetaValueForTest),
					PrevValue: nil,
					Rev:       0,
				},
				tenantID: "test",
			},
			wantErr: false,
		}, {
			name: "case_05 delete event",
			args: args{
				event: &etcd3.Event{
					Type:      etcd3.DELETE,
					Key:       "/sn/quota/cluster/cluster001/tenant/7e1ad6a6-cc5c-44fa-bd54-25873f72a86a/instancemetadata",
					Value:     nil,
					PrevValue: []byte(tenantMetaValueForTest),
					Rev:       0,
				},
				tenantID: "test",
			},
			wantErr: false,
		}, {
			name: "case_06 unkown event",
			args: args{
				event: &etcd3.Event{
					Type:      etcd3.ERROR,
					Key:       "/sn/quota/cluster/cluster001/tenant/7e1ad6a6-cc5c-44fa-bd54-25873f72a86a/instancemetadata",
					Value:     []byte(tenantMetaValueForTest),
					PrevValue: nil,
					Rev:       0,
				},
				tenantID: "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := handleTenantQuotaEvent(tt.args.event, tt.args.tenantID); (err != nil) != tt.wantErr {
				t.Errorf("handleTenantQuotaEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTenantInstanceRegistry_RunWatcher(t *testing.T) {
	convey.Convey("TenantQuotaRegistry", t, func() {
		GlobalRegistry = nil
		stopCh := make(chan struct{})
		patches := []*Patches{
			ApplyMethod(reflect.TypeOf(&etcd3.EtcdWatcher{}), "StartList", func(ew *etcd3.EtcdWatcher) {
				ew.ResultChan <- &etcd3.Event{
					Type:      etcd3.SYNCED,
					Key:       "",
					Value:     nil,
					PrevValue: nil,
					Rev:       0,
					ETCDType:  "",
				}
			}),
			ApplyFunc((*FunctionRegistry).WaitForETCDList, func() {}),
			ApplyMethod(reflect.TypeOf(&etcd3.EtcdWatcher{}), "EtcdHistory", func(ew *etcd3.EtcdWatcher,
				revision int64) {
			}),
			ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient { return &etcd3.EtcdClient{} }),
			ApplyFunc(etcd3.GetMetaEtcdClient, func() *etcd3.EtcdClient { return &etcd3.EtcdClient{} }),
			ApplyMethod(reflect.TypeOf(&etcd3.EtcdWatcher{}), "StartWatch", func(ew *etcd3.EtcdWatcher) {}),
		}
		defer func() {
			for _, p := range patches {
				p.Reset()
			}
		}()
		GlobalRegistry = &Registry{
			FaaSSchedulerRegistry:     NewFaasSchedulerRegistry(stopCh),
			FunctionRegistry:          NewFunctionRegistry(stopCh),
			InstanceRegistry:          NewInstanceRegistry(stopCh),
			FaaSManagerRegistry:       NewFaaSManagerRegistry(stopCh),
			InstanceConfigRegistry:    NewInstanceConfigRegistry(stopCh),
			AliasRegistry:             NewAliasRegistry(stopCh),
			FunctionAvailableRegistry: NewFunctionAvailableRegistry(stopCh),
			FaaSFrontendRegistry:      NewFaaSFrontendRegistry(stopCh),
			TenantQuotaRegistry:       NewTenantQuotaRegistry(stopCh),
			RolloutRegistry:           NewRolloutRegistry(stopCh),
		}
		ProcessETCDList()
		convey.Convey("update and delete", func() {
			errEv := &etcd3.Event{
				Type:      etcd3.PUT,
				Key:       "/sn/quota/cluster/cluster001/tenant",
				Value:     []byte(`["cluster1", "cluster2"]`),
				PrevValue: nil,
				Rev:       0,
			}
			GlobalRegistry.TenantQuotaRegistry.watcherHandler(errEv)
			assert.Equal(t, GlobalRegistry.TenantQuotaRegistry.watcherFilter(errEv), true)

			ev := &etcd3.Event{
				Type:      etcd3.PUT,
				Key:       "/sn/quota/cluster/cluster001/tenant/7e1ad6a6-cc5c-44fa-bd54-25873f72a86a/instancemetadata",
				Value:     []byte(""),
				PrevValue: nil,
				Rev:       0,
			}
			GlobalRegistry.TenantQuotaRegistry.watcherHandler(ev)
			assert.Equal(t, GlobalRegistry.TenantQuotaRegistry.watcherFilter(ev), false)

			deleteEv := &etcd3.Event{
				Type:      etcd3.DELETE,
				Key:       "/sn/quota/cluster/cluster001/default/instancemetadata",
				Value:     nil,
				PrevValue: []byte(`{"tenantInstanceMetaData": {"maxOnDemandInstance": 1000,"maxReversedInstance": 1000}}`),
				Rev:       0,
			}
			GlobalRegistry.TenantQuotaRegistry.watcherHandler(deleteEv)
			assert.Equal(t, GlobalRegistry.TenantQuotaRegistry.watcherFilter(deleteEv), false)
		})
	})
}
