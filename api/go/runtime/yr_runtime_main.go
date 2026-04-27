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

// Package main for start the program.
package main

import (
	"fmt"

	"yuanrong.org/kernel/runtime/faassdk"
	"yuanrong.org/kernel/runtime/libruntime/common"
	"yuanrong.org/kernel/runtime/libruntime/execution"
	"yuanrong.org/kernel/runtime/posixsdk"
	"yuanrong.org/kernel/runtime/yr"
)

func start() {
	execType, err := posixsdk.Load()
	if err != nil {
		fmt.Printf("Load function execution Intfs error:%s\n", err.Error())
	}

	switch execType {
	case execution.ExecutionTypeActor:
		err = yr.InitRuntime()
		if err != nil {
			fmt.Print("init runtime failed: " + err.Error())
			return
		}
		yr.Run()
	case execution.ExecutionTypeFaaS:
		err = faassdk.InitRuntime()
		if err != nil {
			fmt.Print("init runtime failed: " + err.Error())
			return
		}
		faassdk.Run()
	case execution.ExecutionTypePosix:
		conf := common.GetConfig()
		intfs, err := posixsdk.NewPosixFuncExecutionIntfs()
		if err != nil {
			fmt.Printf("failed to new posix intfs, error %s\n", err.Error())
			return
		}
		err = posixsdk.InitRuntime(conf, intfs)
		if err != nil {
			fmt.Print("init runtime failed: " + err.Error())
			return
		}
		posixsdk.Run()
	default:
		err = yr.InitRuntime()
		if err != nil {
			fmt.Print("init runtime failed: " + err.Error())
			return
		}
		yr.Run()
	}
}

func main() {
	start()
}
