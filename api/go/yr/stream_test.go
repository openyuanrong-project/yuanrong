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

// Package yr for actor
package yr

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/runtime/libruntime/api"
)

func TestProducer(t *testing.T) {
	convey.Convey(
		"Test Producer", t, func() {
			convey.Convey(
				"Init success", func() {
					yrConfig := &Config{
						Mode:           ClusterMode,
						FunctionUrn:    "sn:cn:yrk:default:function:0-opc-opc:$latest",
						ServerAddr:     "127.0.0.1:12345",
						DataSystemAddr: "127.0.0.1:12346",
						InCluster:      true,
						AutoStart:      false,
					}
					clientInfo, err := Init(yrConfig)
					convey.So(clientInfo.jobId, convey.ShouldNotBeEmpty)
					convey.So(err, convey.ShouldBeNil)
					convey.Convey(
						"CreateProducer success", func() {
							producerConf := ProducerConf{
								DelayFlushTime: 5,
								PageSize:       1024 * 1024,
								MaxStreamSize:  1024 * 1024 * 1024,
							}
							producer, err := CreateProducer("teststream", producerConf)
							convey.So(producer, convey.ShouldNotBeNil)
							convey.So(err, convey.ShouldBeNil)
						},
					)
				},
			)
		},
	)
}

func TestStream(t *testing.T) {
	producer, err := CreateProducer("streamName", ProducerConf{})
	if err != nil {
		t.Errorf("create producer failed, err: %s", err)
		return
	}
	consumer, err := Subscribe("streamName", "subStreamName", 0)
	if err != nil {
		t.Errorf("create consumer failed, err: %s", err)
		return
	}

	convey.Convey(
		"Test producer", t, func() {
			convey.Convey(
				"Send success", func() {
					err = producer.Send([]byte("value1"))
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)

	convey.Convey("Test consumer", t, func() {
		convey.Convey(
			"Receive success", func() {
				elements, err1 := consumer.Receive(1, 3000)
				convey.So(len(elements), convey.ShouldBeZeroValue)
				convey.So(err1, convey.ShouldBeNil)
			},
		)
		convey.Convey(
			"receiveArr success", func() {
				var value uint8 = 97
				data1 := api.Element{Ptr: &value, Size: 1, Id: 1}
				elements, err1 := consumer.receiveArr([]api.Element{data1})
				convey.So(len(elements), convey.ShouldEqual, 1)
				convey.So(err1, convey.ShouldBeNil)
			},
		)
		convey.Convey(
			"Ack success", func() {
				err = consumer.Ack(0)
				convey.So(err, convey.ShouldBeNil)
			},
		)
	},
	)

	convey.Convey("Test Close", t, func() {
		convey.Convey(
			"Producer Close success", func() {
				err = producer.Close()
				convey.So(err, convey.ShouldBeNil)
			},
		)
		convey.Convey(
			"Consumer Close success", func() {
				err = consumer.Close()
				convey.So(err, convey.ShouldBeNil)
			},
		)
	},
	)
}
