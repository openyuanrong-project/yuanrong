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

package urnutils

import (
	"strings"
)

// Separator is the current system special
var Separator = "-"

const (
	// ServiceIDPrefix is the prefix of the function with serviceID.
	ServiceIDPrefix = "0"

	// DefaultSeparator is a character that separates functions and services.
	DefaultSeparator = "-"

	// ServicePrefix is the prefix of the function with serviceID.
	ServicePrefix = "0@"

	// TenantProductSplitStr separator between a tenant and a product
	TenantProductSplitStr = "@"

	minEleSize = 3
)

// ComplexFuncName contains service ID and raw function name
type ComplexFuncName struct {
	prefix    string
	ServiceID string
	FuncName  string
}

// NewComplexFuncName -
func NewComplexFuncName(svcID, funcName string) *ComplexFuncName {
	return &ComplexFuncName{
		prefix:    ServiceIDPrefix,
		ServiceID: svcID,
		FuncName:  funcName,
	}
}

// ParseFrom parse ComplexFuncName from string
func (c *ComplexFuncName) ParseFrom(name string) *ComplexFuncName {
	fields := strings.Split(name, Separator)
	if len(fields) < minEleSize || fields[0] != ServiceIDPrefix {
		c.prefix = ""
		c.ServiceID = ""
		c.FuncName = name
		return c
	}
	idx := 0
	c.prefix = fields[idx]
	idx++
	c.ServiceID = fields[idx]
	// $prefix$separator$ServiceID$separator$FuncName equals name
	c.FuncName = name[(len(c.prefix) + len(Separator) + len(c.ServiceID) + len(Separator)):]
	return c
}

// String -
func (c *ComplexFuncName) String() string {
	return strings.Join([]string{c.prefix, c.ServiceID, c.FuncName}, Separator)
}

// GetSvcIDWithPrefix get serviceID with prefix from function name
func (c *ComplexFuncName) GetSvcIDWithPrefix() string {
	return c.prefix + Separator + c.ServiceID
}

// SetSeparator -
func SetSeparator(separator string) {
	if separator != "" {
		Separator = separator
	}
}
