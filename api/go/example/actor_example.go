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

type Counter struct {
	Count int
}

func NewCounter(init int) *Counter {
	return &Counter{Count: init}
}

func (c *Counter) Add(x int) int {
	c.Count += x
	return c.Count
}

func (c *Counter) AddValue(x int) int {
	instance := yr.Instance(NewValue).Invoke(30)
	defer instance.Terminate()
	objs := instance.Function((*Value).Show).Invoke()
	value, err := yr.Get[int](objs[0], 30000)
	if err != nil {
		return 0
	}
	return value + x
}

func (c *Counter) CallAddException(x int) (int, error) {
	instance := yr.Instance(NewCounter).Invoke(x)
	defer instance.Terminate()
	objs := instance.Function((*Counter).AddException).Invoke(200)
	value, err := yr.Get[int](objs[0], 30000)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (c *Counter) AddException(x int) (int, error) {
	return x + 1, errors.New("error from go counter")
}

type Value struct {
	value int
}

func NewValue(init int) *Value {
	return &Value{value: init}
}

func (v *Value) Show() int {
	fmt.Println("Value is ", v.value)
	return v.value
}

func ActorExample() {
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

	instance := yr.Instance(NewCounter).Invoke(28)
	objs := instance.Function((*Counter).Add).Invoke(100)
	fmt.Println(yr.Get[int](objs[0], 30000))
	fmt.Println(instance.Terminate())

	// task call task
	instance1 := yr.Instance(NewCounter).Invoke(30)
	objs1 := instance1.Function((*Counter).AddValue).Invoke(120)
	fmt.Println(yr.Get[int](objs1[0], 30000))
	fmt.Println(instance1.Terminate())
}

// compile command:
// CGO_LDFLAGS="-L$(path which contains libcpplibruntime.so)" go build -buildmode=plugin -o yrlib.so actor_example.go
// CGO_LDFLAGS="-L$(path which contains libcpplibruntime.so)" go build -o main actor_example.go
func main() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	ActorExample()
	<-c
}
