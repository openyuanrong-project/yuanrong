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

/*
Package clibruntime
This package encapsulates all cgo invocations.
*/
package clibruntime

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"unsafe"

	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/runtime/libruntime/api"
	"yuanrong.org/kernel/runtime/libruntime/common/uuid"
	"yuanrong.org/kernel/runtime/libruntime/config"
)

func GetConnectArguments() api.ConnectArguments {
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
	return conf
}

func TestInit(t *testing.T) {
	conf := config.Config{}
	id := uuid.New()
	conf.JobID = fmt.Sprintf("job-%s", strings.ReplaceAll(id.String(), "-", "")[:8])
	err := Init(conf)
	if err != nil {
		t.Errorf("test Init failed")
	}
}

func TestCheckNil(t *testing.T) {
	convey.Convey(
		"Test kvClientImplCheckNil", t, func() {
			convey.Convey(
				"kvClientCheckNil success", func() {
					var ptr *KvClientImpl
					err := kvClientCheckNil(ptr)
					convey.So(err, convey.ShouldNotBeNil)
					ptr = &KvClientImpl{}
					err = kvClientCheckNil(ptr)
					convey.So(err, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestKvClientImpl_CreateClient(t *testing.T) {
	convey.Convey(
		"Test CreateClient", t, func() {
			convey.Convey(
				"create client should success", func() {
					conf := GetConnectArguments()
					newClient, err := CreateClient(conf)
					defer newClient.DestroyClient()
					convey.So(err, convey.ShouldBeNil)
					convey.So(newClient, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestKvClientImpl_SetTraceID(t *testing.T) {
	convey.Convey(
		"Test SetTraceID", t, func() {
			conf := GetConnectArguments()
			client, _ := CreateClient(conf)
			defer client.DestroyClient()
			convey.Convey(
				"set traceID should success", func() {
					convey.So(func() { client.SetTraceID("traceId") }, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestKvClientImpl_GenerateKey(t *testing.T) {
	convey.Convey(
		"Test GenerateKey", t, func() {
			client, _ := CreateClient(GetConnectArguments())
			defer client.DestroyClient()
			convey.Convey(
				"generate key should return a key", func() {
					key := client.GenerateKey()
					convey.So(key, convey.ShouldBeEmpty)
				},
			)
		},
	)
}

func TestKvClientImpl_Set(t *testing.T) {
	convey.Convey(
		"Test Set", t, func() {
			client, _ := CreateClient(GetConnectArguments())
			defer client.DestroyClient()
			value := "value"
			var param api.SetParam
			convey.Convey(
				"set key value", func() {
					rightKey := "rightKey"
					convey.Convey(
						"mock a right key", func() {
							status := client.KVSet(rightKey, []byte(value), param)
							convey.So(status.IsOk(), convey.ShouldBeTrue)
						},
					)
				},
			)
		},
	)
}

func TestGetCredential(t *testing.T) {
	convey.Convey(
		"Test GoFunctionExecutionPoolSubmit", t, func() {
			credential := GetCredential()
			convey.So(credential.AccessKey, convey.ShouldBeEmpty)
			convey.So(credential.SecretKey, convey.ShouldBeEmpty)
			convey.So(credential.DataKey, convey.ShouldBeEmpty)
		},
	)
}

func TestKvClientImpl_SetValue(t *testing.T) {
	convey.Convey(
		"Test SetValue", t, func() {
			client, _ := CreateClient(GetConnectArguments())
			defer client.DestroyClient()

			var param api.SetParam
			convey.Convey(
				"set value return key", func() {
					rightValue := "value"
					convey.Convey(
						"set a not empty value", func() {
							key, status := client.KVSetWithoutKey([]byte(rightValue), param)
							convey.So(key, convey.ShouldBeEmpty)
							convey.So(status.IsOk(), convey.ShouldBeTrue)
						},
					)

					emptyValue := ""
					convey.Convey(
						"set emptyValue return empty key", func() {
							key, _ := client.KVSetWithoutKey([]byte(emptyValue), param)
							convey.So(key, convey.ShouldBeEmpty)
						},
					)
				},
			)
		},
	)
}

func TestKvClientImpl_Get(t *testing.T) {
	convey.Convey(
		"Test Get", t, func() {
			client, _ := CreateClient(GetConnectArguments())
			defer client.DestroyClient()

			convey.Convey(
				"use a key to get value", func() {
					rightKey := "rightKey"
					convey.Convey(
						"get a rightKey", func() {
							_, status := client.KVGet(rightKey, 1)
							convey.So(status.IsOk(), convey.ShouldBeTrue)
						},
					)
				},
			)
		},
	)
}

func TestKvClientImpl_GetArray(t *testing.T) {
	convey.Convey(
		"Test GetArray", t, func() {
			client, _ := CreateClient(GetConnectArguments())
			defer client.DestroyClient()

			convey.Convey(
				"use a keys to get values", func() {

					rightKeys := []string{"rightKey"}
					convey.Convey(
						"get values with rightKey", func() {
							values, status := client.KVGetMulti(rightKeys, 1)
							convey.So(status.IsOk(), convey.ShouldBeTrue)
							convey.So(len(values), convey.ShouldEqual, len(rightKeys))
						},
					)
				},
			)
		},
	)
}

func TestKvClientImpl_QuerySize(t *testing.T) {
	convey.Convey(
		"Test QuerySize", t, func() {
			client, _ := CreateClient(GetConnectArguments())
			defer client.DestroyClient()

			convey.Convey(
				"use keys to query value sizes", func() {

					queryKeys := []string{"queryKeys"}
					convey.Convey(
						"get values with queryKeys", func() {
							values, status := client.KVQuerySize(queryKeys)
							convey.So(status.IsOk(), convey.ShouldBeTrue)
							convey.So(len(values), convey.ShouldEqual, len(queryKeys))
						},
					)
				},
			)
		},
	)
}

func TestKvClientImpl_Del(t *testing.T) {
	convey.Convey(
		"Test Del", t, func() {
			client, _ := CreateClient(GetConnectArguments())
			defer client.DestroyClient()

			convey.Convey(
				"del key value", func() {
					rightKey := "rightKey"
					convey.Convey(
						"del a rightKey", func() {
							status := client.KVDel(rightKey)
							convey.So(status.IsOk(), convey.ShouldBeTrue)
						},
					)
				},
			)
		},
	)
}

func TestKvClientImpl_complex(t *testing.T) {
	convey.Convey("concurrency", t, func() {
		convey.So(func() {
			client, _ := CreateClient(GetConnectArguments())

			wg := sync.WaitGroup{}

			wg.Add(4)
			go func() {
				for i := 0; i < 10; i++ {
					client.KVDel("1")
				}
				wg.Done()
			}()
			go func() {
				for i := 0; i < 10; i++ {
					client.KVGet("1", 10)
				}
				wg.Done()
			}()

			go func() {
				for i := 0; i < 10; i++ {
					client.GenerateKey()
				}
				wg.Done()
			}()
			go func() {
				client.DestroyClient()
				client.DestroyClient()
				wg.Done()
			}()
			wg.Wait()
		}, convey.ShouldNotPanic)
	})
}

func TestKvClientImpl_DelArray(t *testing.T) {
	convey.Convey(
		"Test del array", t, func() {
			client, _ := CreateClient(GetConnectArguments())
			defer client.DestroyClient()

			convey.Convey(
				"del keys", func() {
					rightKeys := []string{"key1", "key2"}
					convey.Convey(
						"delete right keys", func() {
							values, status := client.KVDelMulti(rightKeys)
							convey.So(status.IsOk(), convey.ShouldBeTrue)
							convey.So(len(values), convey.ShouldEqual, 0)
						},
					)
				},
			)
		},
	)
}

func TestKvClientImpl_DestroyClient(t *testing.T) {
	convey.Convey(
		"Test destroy client", t, func() {
			client, _ := CreateClient(GetConnectArguments())
			client.DestroyClient()

			convey.Convey(
				"after destroy use client should be safe", func() {
					status := client.KVSet("", []byte{}, api.SetParam{})
					convey.So(status.Code, convey.ShouldEqual, api.DsClientNilError)
					key, status := client.KVSetWithoutKey([]byte{}, api.SetParam{})
					convey.So(status.Code, convey.ShouldEqual, api.DsClientNilError)
					convey.So(key, convey.ShouldBeEmpty)
				},
			)

			convey.Convey(
				"repeat destroy client should not panic", func() {
					client.DestroyClient()
				},
			)

		},
	)
}

func TestCreateStreamProducer(t *testing.T) {
	convey.Convey(
		"Test create stream producer", t, func() {
			convey.Convey(
				"create stream producer success", func() {
					producer, err := CreateStreamProducer("stream_001", api.ProducerConf{})
					convey.So(err, convey.ShouldBeNil)
					convey.So(producer, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestProducerSendAndFlush(t *testing.T) {
	convey.Convey(
		"Test producer send", t, func() {
			producer, err := CreateStreamProducer("stream_001", api.ProducerConf{})
			convey.So(err, convey.ShouldBeNil)
			convey.So(producer, convey.ShouldNotBeNil)
			convey.Convey(
				"producer send", func() {
					ele := api.Element{
						Ptr:  nil,
						Size: 0,
						Id:   0,
					}
					err = producer.Send(ele)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"producer sendWithTimeout", func() {
					ele := api.Element{
						Ptr:  nil,
						Size: 0,
						Id:   0,
					}
					err = producer.SendWithTimeout(ele, 1)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"producer close", func() {
					err = producer.Close()
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestStreamConsumerImpl(t *testing.T) {
	convey.Convey(
		"Test StreamConsumerImpl", t, func() {
			consumer, err := CreateStreamConsumer("stream_001", api.SubscriptionConfig{})
			convey.Convey(
				"CreateStreamConsumer success", func() {
					convey.So(consumer, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"Receive success", func() {
					eles, err := consumer.Receive(3000)
					convey.So(eles, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"ReceiveExpectNum success", func() {
					eles, err := consumer.ReceiveExpectNum(1, 3000)
					convey.So(eles, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"Ack success", func() {
					err = consumer.Ack(1)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"Close success", func() {
					err = consumer.Close()
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestSetTenantID(t *testing.T) {
	convey.Convey(
		"Test SetTraceID", t, func() {
			SetTenantID("tenantId")
			convey.SkipSo()
		},
	)
}

func TestKillInstance(t *testing.T) {
	convey.Convey(
		"Test KillInstance successfully", t, func() {
			err := Kill("instanceid", 1, []byte{})
			convey.So(err, convey.ShouldBeNil)
		},
	)
	convey.Convey(
		"Test KillInstance failed", t, func() {
			err := Kill("instanceid", 128, []byte{})
			convey.So(err.Error(), convey.ShouldEqual, "kill instance: failed to kill")
		},
	)
}

func TestCScheduleAffinities(t *testing.T) {
	convey.Convey(
		"Test cScheduleAffinities", t, func() {
			convey.Convey(
				"Test convert []api.Affinity to *C.CAffinity", func() {
					goLength := 2
					affinities := make([]api.Affinity, goLength)
					affinities[0].Affinity = api.PreferredAffinity
					cAffinities, length := cScheduleAffinities(affinities)
					convey.So(length, convey.ShouldEqual, 2)
					cSchedAffsSlice := unsafe.Slice(cAffinities, 2)
					convey.So(int(cSchedAffsSlice[0].size_labelOps), convey.ShouldEqual, 0)
					freeCScheduleAffinities(cAffinities, length)
				},
			)
		},
	)
}

func TestCheckIfRef(t *testing.T) {
	convey.Convey(
		"Test checkIfRef", t, func() {
			convey.Convey(
				"checkIfRef success when t==api.ObjectRef", func() {
					cChar := checkIfRef(api.ObjectRef)
					convey.So(cChar, convey.ShouldEqual, 1)
				},
			)
		},
	)
}

func TestApiTypeToCApiType(t *testing.T) {
	convey.Convey(
		"Test apiTypeToCApiType", t, func() {
			convey.Convey(
				"apiTypeToCApiType success when apiType ==api.FaaSApi", func() {
					cApiType := apiTypeToCApiType(api.FaaSApi)
					convey.So(cApiType, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"apiTypeToCApiType success when apiType ==api.PosixApi", func() {
					cApiType := apiTypeToCApiType(api.PosixApi)
					convey.So(cApiType, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"apiTypeToCApiType success when default", func() {
					cApiType := apiTypeToCApiType(api.ApiType(9))
					convey.So(cApiType, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestReceiveRequestLoop(t *testing.T) {
	convey.Convey(
		"Test ReceiveRequestLoop", t, func() {
			ReceiveRequestLoop()
			convey.So(ReceiveRequestLoop, convey.ShouldNotPanic)
		},
	)
}

func TestCreateInstance(t *testing.T) {
	convey.Convey(
		"Test CreateInstance", t, func() {
			str, err := CreateInstance(api.FunctionMeta{}, []api.Arg{}, api.InvokeOptions{})
			convey.So(str, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		},
	)
}

func TestInvokeByInstanceId(t *testing.T) {
	convey.Convey(
		"Test InvokeByInstanceId", t, func() {
			str, err := InvokeByInstanceId(api.FunctionMeta{}, "", []api.Arg{}, api.InvokeOptions{})
			convey.So(str, convey.ShouldBeEmpty)
			convey.So(err, convey.ShouldBeNil)
		},
	)
}

func TestInvokeByFunctionName(t *testing.T) {
	convey.Convey(
		"Test InvokeByFunctionName", t, func() {
			str, err := InvokeByFunctionName(api.FunctionMeta{}, []api.Arg{}, api.InvokeOptions{})
			convey.So(str, convey.ShouldBeEmpty)
			convey.So(err, convey.ShouldBeNil)
		},
	)
}

func TestAcquireInstance(t *testing.T) {
	convey.Convey(
		"Test AcquireInstance", t, func() {
			allocation, err := AcquireInstance("", api.FunctionMeta{}, api.InvokeOptions{})
			convey.So(allocation, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		},
	)
}

func TestReleaseInstance(t *testing.T) {
	convey.Convey(
		"Test ReleaseInstance", t, func() {
			convey.So(func() {
				ReleaseInstance(api.InstanceAllocation{}, "", false, api.InvokeOptions{})
			}, convey.ShouldNotPanic)
		},
	)
}

func TestGetAsync(t *testing.T) {
	convey.Convey(
		"Test GetAsync", t, func() {
			convey.So(func() {
				GetAsync("", nil)
			}, convey.ShouldNotPanic)
		},
	)
}

func TestGetEvent(t *testing.T) {
	convey.Convey(
		"Test GetEvent", t, func() {
			convey.So(func() {
				GetEvent("", nil)
			}, convey.ShouldNotPanic)
		},
	)
}

func TestDeleteGetEventCallback(t *testing.T) {
	convey.Convey(
		"Test DeleteGetEventCallback", t, func() {
			convey.So(func() {
				DeleteGetEventCallback("")
			}, convey.ShouldNotPanic)
		},
	)
}

func TestDeleteStream(t *testing.T) {
	convey.Convey(
		"Test DeleteStream", t, func() {
			err := DeleteStream("streamName")
			convey.So(err, convey.ShouldBeNil)
		},
	)
}

func TestQueryGlobalProducersNum(t *testing.T) {
	convey.Convey(
		"Test QueryGlobalProducersNum", t, func() {
			n, err := QueryGlobalProducersNum("streamName")
			convey.So(n, convey.ShouldBeZeroValue)
			convey.So(err, convey.ShouldBeNil)
		},
	)
}

func TestQueryGlobalConsumersNum(t *testing.T) {
	convey.Convey(
		"Test QueryGlobalConsumersNum", t, func() {
			n, err := QueryGlobalConsumersNum("streamName")
			convey.So(n, convey.ShouldBeZeroValue)
			convey.So(err, convey.ShouldBeNil)
		},
	)
}

func TestGetGetAsyncCallback(t *testing.T) {
	convey.Convey(
		"Test getGetAsyncCallback", t, func() {
			convey.Convey(
				"getGetAsyncCallback success", func() {
					cb, err := getGetAsyncCallback("objectID")
					convey.So(cb, convey.ShouldBeNil)
					convey.So(err, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"getGetAsyncCallback success when ok==true", func() {
					cb, err := getGetAsyncCallback("")
					convey.So(cb, convey.ShouldBeNil)
					convey.So(err, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestGetRawCallback(t *testing.T) {
	convey.Convey(
		"Test getRawCallback", t, func() {
			convey.Convey(
				"getGetAsyncCallback success", func() {
					cb, flag := getRawCallback("key")
					convey.So(cb, convey.ShouldBeNil)
					convey.So(flag, convey.ShouldBeFalse)
				},
			)
		},
	)
}

func TestCreateInstanceRaw(t *testing.T) {
	convey.Convey(
		"Test InstanceRaw", t, func() {
			convey.Convey(
				"InstanceRaw success", func() {
					convey.So(func() {
						go CreateInstanceRaw([]byte{0})
						go InvokeByInstanceIdRaw([]byte{0})
						go KillRaw([]byte{0})
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestFinalize(t *testing.T) {
	convey.Convey(
		"Test Finalize", t, func() {
			convey.So(Finalize, convey.ShouldNotPanic)
		},
	)
}

func TestKVSet(t *testing.T) {
	convey.Convey(
		"Test KVSet", t, func() {
			err := KVSet("key", []byte{0}, api.SetParam{})
			convey.So(err, convey.ShouldBeNil)
		},
	)
}

func TestKVGet(t *testing.T) {
	convey.Convey(
		"Test KVGet", t, func() {
			bytes, err := KVGet("key", 3000)
			convey.So(bytes, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		},
	)
}

func TestKVGetMulti(t *testing.T) {
	convey.Convey(
		"Test KVGetMulti", t, func() {
			bytesArr, err := KVGetMulti([]string{"key"}, 3000)
			convey.So(bytesArr, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		},
	)
}

func TestKVDel(t *testing.T) {
	convey.Convey(
		"Test KVDel", t, func() {
			err := KVDel("key")
			convey.So(err, convey.ShouldBeNil)
		},
	)
}

func TestKVDelMulti(t *testing.T) {
	convey.Convey(
		"Test KVDelMulti", t, func() {
			strs, err := KVDelMulti([]string{"key"})
			convey.So(strs, convey.ShouldBeNil)
			convey.So(err, convey.ShouldBeNil)
		},
	)
}

func TestPut(t *testing.T) {
	convey.Convey(
		"Test Put", t, func() {
			err := Put("objectID ", []byte("value"), api.PutParam{}, "nestedObjectIDs")
			convey.So(err, convey.ShouldBeNil)
		},
	)
}

func TestPutRaw(t *testing.T) {
	convey.Convey(
		"Test PutRaw", t, func() {
			err := PutRaw("objectID ", []byte("value"), api.PutParam{}, "nestedObjectIDs")
			convey.So(err, convey.ShouldBeNil)
		},
	)
}

func TestGet(t *testing.T) {
	convey.Convey(
		"Test Get", t, func() {
			bytesArr, err := Get([]string{"key"}, 3000)
			convey.So(bytesArr, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		},
	)
}

func TestGetRaw(t *testing.T) {
	convey.Convey(
		"Test GetRaw", t, func() {
			bytesArr, err := GetRaw([]string{"key"}, 3000)
			convey.So(bytesArr, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		},
	)
}

func TestWait(t *testing.T) {
	convey.Convey(
		"Test Wait", t, func() {
			readyObjectIds, unReadyObjectIds, errs := Wait([]string{"objectIDs"}, 1, 3000)
			convey.So(readyObjectIds, convey.ShouldBeEmpty)
			convey.So(unReadyObjectIds, convey.ShouldBeEmpty)
			convey.So(errs, convey.ShouldBeNil)
		},
	)
}

func TestGIncreaseRef(t *testing.T) {
	convey.Convey(
		"Test GIncreaseRef", t, func() {
			strs, err := GIncreaseRef([]string{"objectID"}, "remoteClientID")
			convey.So(strs, convey.ShouldBeEmpty)
			convey.So(err, convey.ShouldBeNil)
		},
	)
}

func TestGIncreaseRefRaw(t *testing.T) {
	convey.Convey(
		"Test GIncreaseRefRaw", t, func() {
			strs, err := GIncreaseRefRaw([]string{"objectID"}, "remoteClientID")
			convey.So(strs, convey.ShouldBeEmpty)
			convey.So(err, convey.ShouldBeNil)
		},
	)
}

func TestGDecreaseRef(t *testing.T) {
	convey.Convey(
		"Test GDecreaseRef", t, func() {
			strs, err := GDecreaseRef([]string{"objectID"}, "remoteClientID")
			convey.So(strs, convey.ShouldBeEmpty)
			convey.So(err, convey.ShouldBeNil)
		},
	)
}

func TestGDecreaseRefRaw(t *testing.T) {
	convey.Convey(
		"Test GDecreaseRefRaw", t, func() {
			strs, err := GDecreaseRefRaw([]string{"objectID"}, "remoteClientID")
			convey.So(strs, convey.ShouldBeEmpty)
			convey.So(err, convey.ShouldBeNil)
		},
	)
}

func TestAllocReturnObject(t *testing.T) {
	convey.Convey(
		"Test AllocReturnObject", t, func() {
			var size uint = 8
			err := AllocReturnObject(&config.DataObject{}, 8, []string{"nestedIds"}, &size)
			convey.So(err, convey.ShouldBeNil)
		},
	)
}

func TestSetReturnObject(t *testing.T) {
	convey.Convey(
		"Test SetReturnObject", t, func() {
			convey.So(func() {
				SetReturnObject(&config.DataObject{}, 8)
			}, convey.ShouldNotPanic)
		},
	)
}

func TestWriterLatch(t *testing.T) {
	convey.Convey(
		"Test WriterLatch", t, func() {
			err := WriterLatch(&config.DataObject{})
			convey.So(err, convey.ShouldBeNil)
		},
	)
}

func TestMemoryCopy(t *testing.T) {
	convey.Convey(
		"Test MemoryCopy", t, func() {
			err := MemoryCopy(&config.DataObject{}, []byte{0})
			convey.So(err, convey.ShouldBeNil)
		},
	)
}

func TestSeal(t *testing.T) {
	convey.Convey(
		"Test Seal", t, func() {
			err := Seal(&config.DataObject{})
			convey.So(err, convey.ShouldBeNil)
		},
	)
}

func TestWriterUnlatch(t *testing.T) {
	convey.Convey(
		"Test WriterUnlatch", t, func() {
			err := WriterUnlatch(&config.DataObject{})
			convey.So(err, convey.ShouldBeNil)
		},
	)
}

func TestCWriteMode(t *testing.T) {
	convey.Convey(
		"Test cWriteMode", t, func() {
			convey.Convey(
				"WriteThroughL2Cache success", func() {
					cWriteMode := cWriteMode(api.WriteThroughL2Cache)
					convey.So(cWriteMode, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"WriteBackL2Cache success", func() {
					cWriteMode := cWriteMode(api.WriteBackL2Cache)
					convey.So(cWriteMode, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"NoneL2CacheEvict success", func() {
					cWriteMode := cWriteMode(api.NoneL2CacheEvict)
					convey.So(cWriteMode, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"default success", func() {
					cWriteMode := cWriteMode(api.WriteModeEnum(9))
					convey.So(cWriteMode, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestCExistenceOpt(t *testing.T) {
	convey.Convey(
		"Test cExistenceOpt", t, func() {
			convey.Convey(
				"cExistenceOpt success", func() {
					cExistenceOpt := cExistenceOpt(1)
					convey.So(cExistenceOpt, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestCStringOptional(t *testing.T) {
	convey.Convey(
		"Test cStringOptional", t, func() {
			convey.Convey(
				"cStringOptional success", func() {
					str := "str"
					charPtr, char := cStringOptional(&str)
					convey.So(charPtr, convey.ShouldNotBeNil)
					convey.So(char, convey.ShouldEqual, 1)
				},
			)
		},
	)
}

func TestCArgs(t *testing.T) {
	convey.Convey(
		"Test cArgs", t, func() {
			convey.Convey(
				"cArgs success", func() {
					cInvokeArg, cArgsLen := cArgs([]api.Arg{api.Arg{}, api.Arg{}})
					convey.So(cInvokeArg, convey.ShouldNotBeNil)
					convey.So(cArgsLen, convey.ShouldNotBeZeroValue)
					convey.So(func() {
						freeCArgs(cInvokeArg, cArgsLen)
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestCLabelOperators(t *testing.T) {
	convey.Convey(
		"Test cLabelOperators", t, func() {
			convey.Convey(
				"cLabelOperators success", func() {
					cLabelOperator, cLen := cLabelOperators([]api.LabelOperator{api.LabelOperator{}})
					convey.So(cLabelOperator, convey.ShouldNotBeNil)
					convey.So(cLen, convey.ShouldNotBeZeroValue)
					convey.So(func() {
						freeCLabelOperators(cLabelOperator, cLen)
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestCLabelOpType(t *testing.T) {
	convey.Convey(
		"Test cLabelOpType", t, func() {
			convey.Convey(
				"LabelOpNotIn success", func() {
					cLabelOpType := cLabelOpType(api.LabelOpNotIn)
					convey.So(cLabelOpType, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"LabelOpExists success", func() {
					cLabelOpType := cLabelOpType(api.LabelOpExists)
					convey.So(cLabelOpType, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"LabelOpNotExists success", func() {
					cLabelOpType := cLabelOpType(api.LabelOpNotExists)
					convey.So(cLabelOpType, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"default success", func() {
					cLabelOpType := cLabelOpType(api.OperatorType(9))
					convey.So(cLabelOpType, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestCAffinityKind(t *testing.T) {
	convey.Convey(
		"Test cAffinityKind", t, func() {
			convey.Convey(
				"AffinityKindInstance success", func() {
					cAffinityKind := cAffinityKind(api.AffinityKindInstance)
					convey.So(cAffinityKind, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"default success", func() {
					cAffinityKind := cAffinityKind(api.AffinityKindType(9))
					convey.So(cAffinityKind, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestCAffinityType(t *testing.T) {
	convey.Convey(
		"Test cAffinityType", t, func() {
			convey.Convey(
				"PreferredAntiAffinity success", func() {
					cAffinityType := cAffinityType(api.PreferredAntiAffinity)
					convey.So(cAffinityType, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"RequiredAffinity success", func() {
					cAffinityType := cAffinityType(api.RequiredAffinity)
					convey.So(cAffinityType, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"RequiredAntiAffinity success", func() {
					cAffinityType := cAffinityType(api.RequiredAntiAffinity)
					convey.So(cAffinityType, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"default success", func() {
					cAffinityType := cAffinityType(api.AffinityType(9))
					convey.So(cAffinityType, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestCCustomResources(t *testing.T) {
	convey.Convey(
		"Test cCustomResources", t, func() {
			convey.Convey(
				"cCustomResources success", func() {
					cCustomResource, cLen := cCustomResources(map[string]float64{"k": 1.5})
					convey.So(cCustomResource, convey.ShouldNotBeNil)
					convey.So(cLen, convey.ShouldNotBeZeroValue)
				},
			)
		},
	)
}

func TestCCustomExtensions(t *testing.T) {
	convey.Convey(
		"Test cCustomExtensions", t, func() {
			convey.Convey(
				"cCustomExtensions success", func() {
					cCustomExtension, cLen := cCustomExtensions(map[string]string{"k": "v"})
					convey.So(cCustomExtension, convey.ShouldNotBeNil)
					convey.So(cLen, convey.ShouldNotBeZeroValue)
				},
			)
		},
	)
}

func TestCCreateOpt(t *testing.T) {
	convey.Convey(
		"Test cCreateOpt", t, func() {
			convey.Convey(
				"cCreateOpt success", func() {
					cCreateOpt, cLen := cCreateOpt(map[string]string{"k": "v"})
					convey.So(cCreateOpt, convey.ShouldNotBeNil)
					convey.So(cLen, convey.ShouldNotBeZeroValue)
				},
			)
		},
	)
}

func TestByteSliceToCBinaryData(t *testing.T) {
	convey.Convey(
		"Test ByteSliceToCBinaryData", t, func() {
			convey.Convey(
				"ByteSliceToCBinaryData success", func() {
					ptr, len := ByteSliceToCBinaryData([]byte{0})
					convey.So(ptr, convey.ShouldNotBeNil)
					convey.So(len, convey.ShouldNotBeZeroValue)
				},
			)
		},
	)
}

func TestStringToCBinaryDataNoCopy(t *testing.T) {
	convey.Convey(
		"Test StringToCBinaryDataNoCopy", t, func() {
			convey.Convey(
				"StringToCBinaryDataNoCopy success", func() {
					ptr, len := StringToCBinaryDataNoCopy("data")
					convey.So(ptr, convey.ShouldNotBeNil)
					convey.So(len, convey.ShouldNotBeZeroValue)
				},
			)
			convey.Convey(
				"StringToCBinaryDataNoCopy success when len(data)==0", func() {
					ptr, len := StringToCBinaryDataNoCopy("")
					convey.So(ptr, convey.ShouldEqual, unsafe.Pointer(nil))
					convey.So(len, convey.ShouldBeZeroValue)
				},
			)
		},
	)
}

func TestErrtoCerr(t *testing.T) {
	convey.Convey(
		"Test errtoCerr", t, func() {
			convey.Convey(
				"errtoCerr success", func() {
					cErrorInfo := errtoCerr(errors.New("errtoCerr"))
					convey.So(cErrorInfo, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"errtoCerr success when e==nil", func() {
					cErrorInfo := errtoCerr(nil)
					convey.So(cErrorInfo, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestGoHealthCheck(t *testing.T) {
	convey.Convey(
		"Test GoHealthCheck", t, func() {
			convey.Convey(
				"GoHealthCheck success", func() {
					cHealthCheckCode := GoHealthCheck()
					convey.So(cHealthCheckCode, convey.ShouldBeZeroValue)
				},
			)
		},
	)
}

func TestGoHasHealthCheck(t *testing.T) {
	convey.Convey(
		"Test GoHasHealthCheck", t, func() {
			convey.Convey(
				"GoHasHealthCheck success", func() {
					cChar := GoHasHealthCheck()
					convey.So(cChar, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestGoFunctionExecutionPoolSubmit(t *testing.T) {
	convey.Convey(
		"Test GoFunctionExecutionPoolSubmit", t, func() {
			convey.Convey(
				"GoFunctionExecutionPoolSubmit success", func() {
					convey.So(func() {
						GoFunctionExecutionPoolSubmit(unsafe.Pointer(nil))
					}, convey.ShouldPanic)
				},
			)
		},
	)
}
