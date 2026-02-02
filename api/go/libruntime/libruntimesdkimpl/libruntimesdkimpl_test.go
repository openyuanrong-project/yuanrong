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

// Package libruntimesdkimpl implements
package libruntimesdkimpl

import (
	"os"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/runtime/libruntime/api"
)

func TestCreateClient(t *testing.T) {
	convey.Convey(
		"Test CreateClient", t, func() {
			libruntimeAPI := NewLibruntimeSDKImpl()
			convey.Convey(
				"create client should success", func() {
					var conf api.ConnectArguments
					conf.Host = "127.0.0.1"
					conf.Port = 11111
					conf.TimeoutMs = 500
					conf.Token = []byte{'1', '2', '3'}
					conf.ClientPublicKey = "client pub key"
					conf.ClientPrivateKey = []byte{'1', '2', '3'}
					conf.ServerPublicKey = "server pub key"
					conf.AccessKey = "access key"
					conf.SecretKey = []byte{'1', '2', '3'}
					conf.AuthclientID = "auth client id"
					conf.AuthclientSecret = []byte{'1', '2', '3'}
					conf.AuthURL = "auth url"
					conf.TenantID = "tenant id"
					conf.EnableCrossNodeConnection = true
					newClient, err := libruntimeAPI.CreateClient(conf)
					defer newClient.DestroyClient()
					convey.So(err, convey.ShouldBeNil)
					convey.So(newClient, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestLibruntimeSDKImpl(t *testing.T) {
	convey.Convey(
		"Test libruntimeSDKImpl", t, func() {
			libruntimeAPI := NewLibruntimeSDKImpl()
			convey.Convey(
				"CreateInstance success", func() {
					str, err := libruntimeAPI.CreateInstance(api.FunctionMeta{}, []api.Arg{}, api.InvokeOptions{})
					convey.So(str, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"InvokeByInstanceId success", func() {
					str, err := libruntimeAPI.InvokeByInstanceId(api.FunctionMeta{}, "instanceID",
						[]api.Arg{}, api.InvokeOptions{})
					convey.So(str, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"InvokeByFunctionName success", func() {
					str, err := libruntimeAPI.InvokeByFunctionName(api.FunctionMeta{}, []api.Arg{}, api.InvokeOptions{})
					convey.So(str, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"AcquireInstance success", func() {
					allocation, err := libruntimeAPI.AcquireInstance("state", api.FunctionMeta{}, api.InvokeOptions{})
					convey.So(allocation, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"ReleaseInstance success", func() {
					convey.So(func() {
						libruntimeAPI.ReleaseInstance(api.InstanceAllocation{}, "state", false, api.InvokeOptions{})
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"Kill success", func() {
					err := libruntimeAPI.Kill("instanceID", 0, []byte{})
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"CreateInstanceRaw success", func() {
					convey.So(func() {
						go libruntimeAPI.CreateInstanceRaw([]byte{})
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"InvokeByInstanceIdRaw success", func() {
					convey.So(func() {
						go libruntimeAPI.InvokeByInstanceIdRaw([]byte{})
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"KillRaw success", func() {
					convey.So(func() {
						go libruntimeAPI.KillRaw([]byte{})
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"SaveState success", func() {
					str, err := libruntimeAPI.SaveState([]byte{})
					convey.So(str, convey.ShouldBeEmpty)
					convey.So(err, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"LoadState success", func() {
					str, err := libruntimeAPI.LoadState("checkpointID")
					convey.So(str, convey.ShouldBeEmpty)
					convey.So(err, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"Finalize success", func() {
					convey.So(libruntimeAPI.Finalize, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"KVSet success", func() {
					err := libruntimeAPI.KVSet("key", []byte("value"), api.SetParam{})
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"KVSetWithoutKey success", func() {
					str, err := libruntimeAPI.KVSetWithoutKey([]byte("value"), api.SetParam{})
					convey.So(str, convey.ShouldBeEmpty)
					convey.So(err, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"KVGet success", func() {
					bytes, err := libruntimeAPI.KVGet("key", 300)
					convey.So(bytes, convey.ShouldBeEmpty)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"KVGetMulti success", func() {
					bytesArr, err := libruntimeAPI.KVGetMulti([]string{"key"}, 300)
					convey.So(bytesArr, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"KVDel success", func() {
					err := libruntimeAPI.KVDel("key")
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"KVDelMulti success", func() {
					strs, err := libruntimeAPI.KVDelMulti([]string{"key"})
					convey.So(strs, convey.ShouldBeEmpty)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"create producer", func() {
					producer, err := libruntimeAPI.CreateProducer("stream_001", api.ProducerConf{})
					convey.ShouldNotBeNil(producer)
					convey.ShouldBeNil(err)
				},
			)
			convey.Convey(
				"create consumer", func() {
					consumer, err := libruntimeAPI.Subscribe("stream_001", api.SubscriptionConfig{})
					convey.ShouldNotBeNil(consumer)
					convey.ShouldBeNil(err)
				},
			)
			convey.Convey(
				"DeleteStream success", func() {
					err := libruntimeAPI.DeleteStream("streamName")
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"QueryGlobalProducersNum success", func() {
					n, err := libruntimeAPI.QueryGlobalProducersNum("streamName")
					convey.So(n, convey.ShouldBeZeroValue)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"QueryGlobalConsumersNum success", func() {
					n, err := libruntimeAPI.QueryGlobalConsumersNum("streamName")
					convey.So(n, convey.ShouldBeZeroValue)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"SetTraceID success", func() {
					convey.So(func() {
						libruntimeAPI.SetTraceID("traceID")
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"SetTenantID success", func() {
					err := libruntimeAPI.SetTenantID("tenantID")
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"Put success", func() {
					err := libruntimeAPI.Put("objectID", []byte("data"), api.PutParam{}, "nestedObjectID")
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"PutRaw success", func() {
					err := libruntimeAPI.PutRaw("objectID", []byte("data"), api.PutParam{}, "nestedObjectID")
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"Get success", func() {
					bytesArr, err := libruntimeAPI.Get([]string{"objectID"}, 300)
					convey.So(bytesArr, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"GetRaw success", func() {
					bytesArr, err := libruntimeAPI.GetRaw([]string{"objectID"}, 300)
					convey.So(bytesArr, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"Wait success", func() {
					strs1, strs2, m := libruntimeAPI.Wait([]string{"objectID"}, 1, 300)
					convey.So(strs1, convey.ShouldBeEmpty)
					convey.So(strs2, convey.ShouldBeEmpty)
					convey.So(m, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"GIncreaseRef success", func() {
					strs, err := libruntimeAPI.GIncreaseRef([]string{"objectID"}, "remoteClientID")
					convey.So(strs, convey.ShouldBeEmpty)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"GIncreaseRefRaw success", func() {
					strs, err := libruntimeAPI.GIncreaseRefRaw([]string{"objectID"}, "remoteClientID")
					convey.So(strs, convey.ShouldBeEmpty)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"GDecreaseRef success", func() {
					strs, err := libruntimeAPI.GDecreaseRef([]string{"objectID"}, "remoteClientID")
					convey.So(strs, convey.ShouldBeEmpty)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"GDecreaseRefRaw success", func() {
					strs, err := libruntimeAPI.GDecreaseRefRaw([]string{"objectID"}, "remoteClientID")
					convey.So(strs, convey.ShouldBeEmpty)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"GetAsync success", func() {
					convey.So(func() {
						f := func(result []byte, err error) {}
						libruntimeAPI.GetAsync("objectID", f)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"GetEvent success", func() {
					convey.So(func() {
						f := func(result []byte, err error) {}
						libruntimeAPI.GetEvent("objectID", f)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"DeleteGetEventCallback success", func() {
					convey.So(func() {
						libruntimeAPI.DeleteGetEventCallback("objectID")
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"GetFormatLogger success", func() {
					fl := libruntimeAPI.GetFormatLogger()
					convey.So(fl, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestGetLogName(t *testing.T) {
	convey.Convey(
		"Test getLogName", t, func() {
			convey.Convey(
				"CreateInstance success when path is empty", func() {
					str := getLogName()
					convey.So(str, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"CreateInstance success", func() {
					os.Setenv("FUNCTION_LIB_PATH", "./go1.x")
					str := getLogName()
					convey.So(str, convey.ShouldNotBeNil)
				},
			)
		},
	)
}
