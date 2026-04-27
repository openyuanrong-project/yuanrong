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

func PlusOne(x int) int {
	fmt.Println("Hello PlusOne")
	return x + 1
}

func PlusOne2(x int) int {
	fmt.Println("Hello PlusOne2")
	function := yr.Function(PlusOne).Options(yr.NewInvokeOptions())
	ref := function.Invoke(300)
	fmt.Println(yr.Get[int](ref[0], 3000))
	return x + 1
}

func TaskExample() {
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

	function := yr.Function(PlusOne).Options(yr.NewInvokeOptions())
	ref := function.Invoke(298)
	fmt.Println(yr.Get[int](ref[0], 3000))

	// task call task
	function1 := yr.Function(PlusOne2).Options(yr.NewInvokeOptions())
	ref1 := function1.Invoke(298)
	fmt.Println(yr.Get[int](ref1[0], 3000))
}

// compile command:
// CGO_LDFLAGS="-L$(path which contains libcpplibruntime.so)" go build -buildmode=plugin -o yrlib.so task_example.go
// CGO_LDFLAGS="-L$(path which contains libcpplibruntime.so)" go build -o main task_example.go
func main() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	TaskExample()
	<-c
}
