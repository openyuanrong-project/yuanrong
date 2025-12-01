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
		FunctionUrn:    "sn:cn:yrk:12345678901234561234567890123456:function:0-opc-opc:$latest",
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

func PutAndGetExample() {
	obj, err := yr.Put(250)
	if err != nil {
		fmt.Println("Put failed, error: ", err)
	}

	fmt.Println(yr.Get[int](obj, 30000))
}

func PutAndWaitExample() {
	obj, err := yr.Put(250)
	if err != nil {
		fmt.Println("Put failed, error: ", err)
	}

	fmt.Println(yr.Wait(obj, 30000))
}

func PutAndBatchGetExample() {
	obj, err := yr.Put(250)
	if err != nil {
		fmt.Println("Put failed, error: ", err)
	}

	obj1, err := yr.Put(2560)
	if err != nil {
		fmt.Println("Put failed, error: ", err)
	}

	objs := []*yr.ObjectRef{obj, obj1}
	fmt.Println(yr.BatchGet[int](objs, 30000, false))
}

func PutAndBatchWaitExample() {
	obj, err := yr.Put(250)
	if err != nil {
		fmt.Println("Put failed, error: ", err)
	}

	obj1, err := yr.Put(2560)
	if err != nil {
		fmt.Println("Put failed, error: ", err)
	}

	objs := []*yr.ObjectRef{obj, obj1}
	fmt.Println(objs)
	fmt.Println(yr.WaitNum(objs, 3, 2))
}

// compile command:
// CGO_LDFLAGS="-L$(path which contains libcpplibruntime.so)" go build -o main put_get_wait_example.go
func main() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	InitYR()
	PutAndGetExample()
	PutAndWaitExample()
	PutAndBatchGetExample()
	PutAndBatchWaitExample()
	<-c
}
