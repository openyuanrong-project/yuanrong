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
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"yuanrong.org/kernel/runtime/libruntime/api"
)

const defaultClassName = "main"

func getFuncInfos(s string) (string, string, string, error) {
	var methodName, parameters string
	items := strings.Split(s, ".")
	if len(items) > 0 {
		s = strings.TrimPrefix(s, items[0]+".")
	}
	if s == "" {
		return "", "", "", fmt.Errorf("failed to parse stack")
	}
	sByte := []byte(s)
	l := len(sByte)
	if sByte[l-1] != ')' {
		// no param
		return defaultClassName, s, "", nil
	}
	for j := l - 1; j >= 0; j-- {
		if sByte[j] == '(' {
			parameters = string(sByte[j:l])
			break
		}
	}
	methodName = strings.TrimSuffix(s, parameters)
	return defaultClassName, methodName, parameters, nil
}

func getFileInfos(s string) (string, string, int64, error) {
	var fileName, offset string
	var linNumber int64
	fileInfoSep := "[: ]"
	fileInfos := regexp.MustCompile(fileInfoSep).Split(strings.TrimSpace(s), -1)
	// for example: [a.go 22 0x03], lengh should larger then 2
	if len(fileInfos) >= 2 {
		fileName = fileInfos[0]
		line, err := strconv.Atoi(fileInfos[1])
		if err != nil {
			return "", "", 0, err
		} else {
			linNumber = int64(line)
		}
		if len(fileInfos) > 2 {
			offset = fileInfos[2]
		}
	}
	return fileName, offset, linNumber, nil
}

func parseStack(funcInfo, fileInfo string) (api.StackTrace, error) {
	className, methodName, parameters, err := getFuncInfos(funcInfo)
	if err != nil {
		fmt.Println(err.Error(), funcInfo)
		return api.StackTrace{}, err
	}
	fineName, offset, linNumber, err := getFileInfos(fileInfo)
	if err != nil {
		fmt.Println(err.Error(), funcInfo)
		return api.StackTrace{}, err
	}
	stack := api.StackTrace{
		ClassName:     className,
		MethodName:    methodName,
		FileName:      fineName,
		LineNumber:    linNumber,
		ExtensionInfo: make(map[string]string, 10),
	}
	if parameters != "" {
		stack.ExtensionInfo["parameters"] = parameters
	}
	if offset != "" {
		stack.ExtensionInfo["offset"] = offset
	}
	return stack, nil
}

func getStackTraceInfos() []api.StackTrace {
	if !GetConfigManager().EnableCallStack {
		return nil
	}
	stackNum := GetConfigManager().CallStackLayerNum
	stack := make([]byte, 1024*1024)
	n := runtime.Stack(stack, false)
	info := string(stack[:n])
	items := strings.Split(info, "\n")
	var stacks []api.StackTrace
	for i := 0; i < len(items); i++ {
		if (strings.HasPrefix(items[i], "plugin") || strings.HasPrefix(items[i], "main")) && i+1 < len(items) {
			stack, err := parseStack(items[i], items[i+1])
			if err != nil {
				continue
			}
			stacks = append(stacks, stack)
			if len(stacks) == stackNum {
				break
			}
			i++
		}
	}
	return stacks
}

// GetStackTrace return api.StackTrace
func GetStackTrace(instance *reflect.Value, functionName string) api.StackTrace {
	if !GetConfigManager().EnableCallStack {
		return api.StackTrace{}
	}
	return api.StackTrace{
		ClassName:  defaultClassName,
		MethodName: getMethodName(instance, functionName),
	}
}

func getMethodName(instance *reflect.Value, funcName string) string {
	if instance == nil {
		return funcName
	}
	// <*main.Item Value>
	s := instance.String()
	infos := strings.Split(s, ".")
	// for example: [<*main Item Value>], lengh should larger then 2
	if len(infos) < 2 {
		return funcName
	}
	items := strings.Split(infos[1], " ")
	return fmt.Sprintf("(*%s).%s", items[0], funcName)
}

func getCallStackTracesMgr() (int, bool) {
	isEnable := os.Getenv("ENABLE_DIS_CONV_CALL_STACK") != "false"
	if !isEnable {
		return 0, false
	}
	callStackLayer := os.Getenv("MAX_CALL_STACK_LAYER_NUM")
	layerNum, err := strconv.Atoi(callStackLayer)
	if err != nil {
		layerNum = defaultStackTracesNum
	}
	return layerNum, true
}
