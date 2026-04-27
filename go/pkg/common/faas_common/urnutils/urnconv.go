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

// Package urnutils -
package urnutils

import (
	"strings"
)

// FunctionInfo defines Function Info
type FunctionInfo struct {
	Business string
	Tenant   string
	FuncName string
	Version  string
}

// CrNameByKey return Cr Name By function key
func CrNameByKey(funcKey string) string {
	functionInfo := GetFunctionInfoByKey(funcKey)
	business, tenant, funcName, version := functionInfo.Business, functionInfo.Tenant,
		functionInfo.FuncName, functionInfo.Version

	return CrName(business, tenant, funcName, version)
}

// GetFunctionInfoByKey -
func GetFunctionInfoByKey(key string) FunctionInfo {
	var functionInfo FunctionInfo
	keyFields := strings.Split(key, "/")

	if len(keyFields) != URNIndexEleven && len(keyFields) != URNIndexThirteen {
		return functionInfo
	}

	functionInfo.Business = keyFields[URNIndexFour]
	functionInfo.Tenant = keyFields[URNIndexSix]
	functionInfo.FuncName = keyFields[URNIndexEight]
	functionInfo.Version = keyFields[URNIndexTen]

	return functionInfo
}
