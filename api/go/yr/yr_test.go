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
Package execution
This package provides methods to obtain the execution interface.
*/
package yr

import (
	"errors"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

var clientInfo *ClientInfo
var initErr error

func init() {
	yrConfig := &Config{
		FunctionUrn:    "sn:cn:yrk:default:function:0-opc-opc:$latest",
		Mode:           ClusterMode,
		AutoStart:      false,
		ServerAddr:     "127.0.0.1:12345",
		DataSystemAddr: "127.0.0.1:12346",
		InCluster:      true,
	}
	clientInfo, initErr = Init(yrConfig)
}

func TestInit(t *testing.T) {
	convey.Convey(
		"Test Init", t, func() {
			convey.Convey(
				"Init success", func() {
					convey.So(clientInfo.jobId, convey.ShouldNotBeEmpty)
					convey.So(initErr, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestInitWithFlags(t *testing.T) {
	convey.Convey(
		"Test InitWithFlags", t, func() {
			yrConfig := &Config{
				FunctionUrn:    "sn:cn:yrk:default:function:0-opc-opc:$latest",
				Mode:           ClusterMode,
				AutoStart:      false,
				ServerAddr:     "127.0.0.1:12345",
				DataSystemAddr: "127.0.0.1:12346",
				InCluster:      true,
			}
			yrFlagsConfig := &FlagsConfig{}
			convey.Convey(
				"InitWithFlags success", func() {
					clientInfo, err := InitWithFlags(yrConfig, yrFlagsConfig)
					convey.So(clientInfo.jobId, convey.ShouldNotBeEmpty)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestFinalize(t *testing.T) {
	convey.Convey(
		"Test Finalize", t, func() {
			convey.Convey(
				"Finalize killAllInstance success", func() {
					convey.So(func() {
						Finalize(true)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"Finalize notKillAllInstance success", func() {
					convey.So(func() {
						Finalize(false)
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestGet(t *testing.T) {
	convey.Convey(
		"Test Get", t, func() {
			convey.Convey(
				"Get success", func() {
					objRef, err := Put(250)
					if err != nil {
						t.Errorf("Put failed, error: %s", err)
						return
					}
					res, err := Get[int](objRef, 3000)
					convey.So(res, convey.ShouldBeZeroValue)
					convey.So(err, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestSubGet(t *testing.T) {
	convey.Convey(
		"Test SubGet", t, func() {
			convey.Convey(
				"SubGet success", func() {
					objRef, err := Put(250)
					if err != nil {
						t.Errorf("Put failed, error: %s", err)
						return
					}
					res, err := get[int](objRef, 3000)
					convey.So(res, convey.ShouldBeZeroValue)
					convey.So(err, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestBatchGet(t *testing.T) {
	convey.Convey(
		"Test BatchGet", t, func() {
			convey.Convey(
				"BatchGet success", func() {
					obj, err := Put(250)
					if err != nil {
						t.Errorf("Put failed, error: %s", err)
						return
					}
					obj1, err := Put(2560)
					if err != nil {
						t.Errorf("Put failed, error: %s", err)
						return
					}
					objs := []*ObjectRef{obj, obj1}

					res, err := BatchGet[int](objs, 3000, false)
					convey.So(len(res), convey.ShouldBeZeroValue)
					convey.So(err, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestPut(t *testing.T) {
	convey.Convey(
		"Test Put", t, func() {
			convey.Convey(
				"Put success", func() {
					objRef, err := Put(250)
					if err != nil {
						t.Errorf("Put failed, error: %s", err)
						return
					}
					convey.So(objRef.objId, convey.ShouldNotBeEmpty)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestWaitErr(t *testing.T) {
	convey.Convey(
		"Test waitErr", t, func() {
			convey.Convey(
				"waitErr success", func() {
					errs := map[string]error{"1": errors.New("wait failed")}
					err := waitErr(errs)
					convey.So(err, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestWaitNum(t *testing.T) {
	convey.Convey(
		"Test WaitNum", t, func() {
			convey.Convey(
				"WaitNum success", func() {
					objRef, err := Put(250)
					if err != nil {
						t.Errorf("Put failed, error: %s", err)
						return
					}

					readyObjRefs, unreadyObjRefs, err := WaitNum([]*ObjectRef{objRef}, 1, 3000)
					convey.So(len(readyObjRefs), convey.ShouldBeZeroValue)
					convey.So(len(unreadyObjRefs), convey.ShouldBeZeroValue)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestSetKV(t *testing.T) {
	convey.Convey(
		"Test SetKV", t, func() {
			convey.Convey(
				"SetKV success", func() {
					err := SetKV("key1", "value1")
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestGetKV(t *testing.T) {
	convey.Convey(
		"Test GetKV", t, func() {
			convey.Convey(
				"GetKV success", func() {
					str, err := GetKV("key1", 3000)
					convey.So(str, convey.ShouldBeEmpty)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestGetKVs(t *testing.T) {
	convey.Convey(
		"Test GetKVs", t, func() {
			convey.Convey(
				"GetKVs success", func() {
					values, err := GetKVs([]string{"key1"}, 3000, true)
					convey.So(values[0], convey.ShouldBeEmpty)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestDelKV(t *testing.T) {
	convey.Convey(
		"Test DelKV", t, func() {
			convey.Convey(
				"DelKV success", func() {
					err := DelKV("key1")
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestDelKVs(t *testing.T) {
	convey.Convey(
		"Test DelKVs", t, func() {
			convey.Convey(
				"DelKVs success", func() {
					failedKeys, err := DelKVs([]string{"key1", "key2"})
					convey.So(len(failedKeys), convey.ShouldBeZeroValue)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestCreateProducer(t *testing.T) {
	convey.Convey(
		"Test CreateProducer", t, func() {
			convey.Convey(
				"CreateProducer success", func() {
					producer, err := CreateProducer("streamName", ProducerConf{})
					convey.So(producer.producer, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestSubscribe(t *testing.T) {
	convey.Convey(
		"Test Subscribe", t, func() {
			convey.Convey(
				"Subscribe success", func() {
					consumer, err := Subscribe("streamName", "subscriptionName", 0)
					convey.So(consumer.consumer, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestDeleteStream(t *testing.T) {
	convey.Convey(
		"Test DeleteStream", t, func() {
			convey.Convey(
				"DeleteStream success", func() {
					err := DeleteStream("streamName")
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestQueryGlobalProducersNum(t *testing.T) {
	convey.Convey(
		"Test QueryGlobalProducersNum", t, func() {
			convey.Convey(
				"QueryGlobalProducersNum success", func() {
					producersNum, err := QueryGlobalProducersNum("streamName")
					convey.So(producersNum, convey.ShouldBeZeroValue)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestQueryGlobalConsumersNum(t *testing.T) {
	convey.Convey(
		"Test QueryGlobalConsumersNum", t, func() {
			convey.Convey(
				"QueryGlobalConsumersNum success", func() {
					consumersNum, err := QueryGlobalConsumersNum("streamName")
					convey.So(consumersNum, convey.ShouldBeZeroValue)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}
