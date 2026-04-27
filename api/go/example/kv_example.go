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

func KVSetAndGetAndDelExample() {
	err := yr.SetKV("key", "value")
	if err != nil {
		fmt.Println("set kv failed, error: ", err)
	}

	fmt.Println(yr.GetKV("key", 30000))

	fmt.Println(yr.DelKV("key"))
}

func BatchKVSetAndGetAndDelExample() {
	err := yr.SetKV("key", "value")
	if err != nil {
		fmt.Println("set kv failed, error: ", err)
	}

	err = yr.SetKV("key1", "value1")
	if err != nil {
		fmt.Println("set kv failed, error: ", err)
	}

	fmt.Println(yr.GetKVs([]string{"key", "key1"}, 30000, false))

	fmt.Println(yr.DelKVs([]string{"key", "key1"}))

	fmt.Println(yr.GetKVs([]string{"key", "key1"}, 30000, false))
}

// compile command:
// CGO_LDFLAGS="-L$(path which contains libcpplibruntime.so)" go build -o main kv_example.go
func main() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	InitYR()
	KVSetAndGetAndDelExample()
	BatchKVSetAndGetAndDelExample()
	<-c
}
