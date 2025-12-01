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
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"yuanrong.org/kernel/runtime/yr"
)

func Functiond(x int) (int, error) {
	if x == 0 {
		return 0, errors.New("invalid param")
	}
	return x - 1, nil
}

func Functionc(x int) (int, error) {
	function := yr.Function(Functiond).Options(yr.NewInvokeOptions())
	ref := function.Invoke(x)
	res, err := yr.Get[int](ref[0], 30000)
	if err != nil {
		fmt.Println("IN FUNCTIONC")
		return 0, err
	}
	return res, nil
}

func Functionb(x int) (int, error) {
	function := yr.Function(Functionc).Options(yr.NewInvokeOptions())
	ref := function.Invoke(x)
	res, err := yr.Get[int](ref[0], 30000)
	if err != nil {
		fmt.Println("IN FUNCTIONB")
		return 0, err
	}
	return res, nil
}

func Functiona() {
	fmt.Println("TaskExceptionExample")
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

	function := yr.Function(Functionb).Options(yr.NewInvokeOptions())
	ref := function.Invoke(0)
	res, err := yr.Get[int](ref[0], 30000)
	if err != nil {
		fmt.Printf(err.Error())
	} else {
		fmt.Printf("FunctionA result is %d\n", res)
	}
}

// compile command:
// CGO_LDFLAGS="-L$(path which contains libcpplibruntime.so)" go build -buildmode=plugin -o yrlib.so task_example.go
// CGO_LDFLAGS="-L$(path which contains libcpplibruntime.so)" go build -o main task_example.go
func main() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	Functiona()
	<-c
}
