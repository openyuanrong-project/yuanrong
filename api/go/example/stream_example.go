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

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"yuanrong.org/kernel/runtime/yr"
)

func InitYR() {
	flag.Parse()
	config := &yr.Config{
		Mode:           yr.ClusterMode,
		FunctionUrn:    "sn:cn:yrk:default:function:0-opc-opc:$latest",
		ServerAddr:     flag.Args()[0],
		DataSystemAddr: flag.Args()[1],
		InCluster:      true,
		LogLevel:       "DEBUG",
	}
	info, err := yr.Init(config)
	if err != nil {
		fmt.Println("Init failed, error: ", err)
	}
	fmt.Println(info)
}

func StreamExample() {
	producerConf := yr.ProducerConf{
		DelayFlushTime: 5,
		PageSize:       1024 * 1024,
		MaxStreamSize:  1024 * 1024 * 1024,
	}
	producer, err := yr.CreateProducer("teststream", producerConf)
	if err != nil {
		fmt.Println("create producer failed, err: ", err)
		return
	}

	consumer, err := yr.Subscribe("teststream", "substreamName", 0)
	if err != nil {
		fmt.Println("create consumer failed, err: ", err)
		return
	}

	data, err := msgpack.Marshal("test-message")
	if err != nil {
		fmt.Println("marshal failed, err: ", err)
		return
	}

	fmt.Println(producer.Send(data))

	subDatas, err := consumer.Receive(1, 30000)
	if err != nil {
		fmt.Println("receive failed, err: ", err)
	}

	fmt.Println(subDatas[0])
	var result string
	fmt.Println(msgpack.Unmarshal(subDatas[0].Data, &result))
	fmt.Println("result: ", result)
	fmt.Println(consumer.Ack(uint64(subDatas[0].Id)))

	fmt.Println(producer.Close())
	fmt.Println(consumer.Close())
	fmt.Println(yr.DeleteStream("teststream"))
}

func QueryStreamNumExample() {
	fmt.Println(yr.QueryGlobalProducersNum("testStream"))
	fmt.Println(yr.QueryGlobalConsumersNum("testStream"))
}

// compile command:
// CGO_LDFLAGS="-L$(path which contains libcpplibruntime.so)" go build -o main stream_example.go
func main() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	InitYR()
	StreamExample()
	QueryStreamNumExample()
	<-c
}
