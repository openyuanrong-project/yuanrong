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
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/vmihailenco/msgpack"

	"yuanrong.org/kernel/runtime/libruntime/api"
)

// IsFunction return if type is Func
func IsFunction(fn any) bool {
	return reflect.TypeOf(fn).Kind() == reflect.Func
}

// GetFunctionName return function name
func GetFunctionName(fn any) (string, error) {
	functionName := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	functionNameSlice := strings.Split(functionName, ".")
	if len(functionNameSlice) >= 1 {
		return functionNameSlice[len(functionNameSlice)-1], nil
	}

	return "", fmt.Errorf("get function name %s failed", functionName)
}

// PackInvokeArgs pack func args and return []Arg
func PackInvokeArgs(args ...any) ([]api.Arg, error) {
	res := make([]api.Arg, 0, len(args))
	for _, arg := range args {
		packedArg, err := PackInvokeArg(arg)
		if err != nil {
			return res, err
		}
		res = append(res, packedArg)
	}
	return res, nil
}

// PackInvokeArg pack func arg and return Arg
func PackInvokeArg(arg any) (api.Arg, error) {
	res := api.Arg{}
	switch reflect.TypeOf(arg).String() {
	case "yr.ObjectRef":
		return res, nil
	case "*yr.ObjectRef":
		return res, nil
	case "[]yr.ObjectRef":
		return res, nil
	case "[]*yr.ObjectRef":
		return res, nil
	default:
		{
			data, err := msgpack.Marshal(arg)
			if err != nil {
				return res, err
			}
			res.Data = data
			return res, nil
		}
	}
}
