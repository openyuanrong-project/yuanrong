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

// Package urnutils
package urnutils

import (
	"fmt"
	"strings"

	"yuanrong.org/kernel/runtime/libruntime/common/faas/logger"
)

// An example of a function URN: <ProductID>:<RegionID>:<BusinessID>:<TenantID>:<FunctionSign>:<FunctionName>:<Version>
// Indices of elements in BaseURN
const (
	// ProductIDIndex is the index of the product ID in a URN
	ProductIDIndex = iota
	// RegionIDIndex is the index of the region ID in a URN
	RegionIDIndex
	// BusinessIDIndex is the index of the business ID in a URN
	BusinessIDIndex
	// TenantIDIndex is the index of the tenant ID in a URN
	TenantIDIndex
	// FunctionSignIndex is the index of the product ID in a URN
	FunctionSignIndex
	// FunctionNameIndex is the index of the product name in a URN
	FunctionNameIndex
	// VersionIndex is the index of the version in a URN
	VersionIndex
	// URNLenWithVersion is the normal URN length with a version
	URNLenWithVersion
)

const (
	urnLenWithoutVersion = URNLenWithVersion - 1
	// URNSep is a URN separator of functions
	URNSep = ":"
	// DefaultURNProductID is the default product ID of a URN
	DefaultURNProductID = "sn"
	// DefaultURNRegion is the default region of a URN
	DefaultURNRegion = "cn"
	// DefaultURNFuncSign is the default function sign of a URN
	DefaultURNFuncSign = "function"
)

var (
	// LocalFuncURN is URN of local function
	LocalFuncURN = &BaseURN{}
)

// BaseURN contains elements of a product URN. It can expand to FunctionURN, LayerURN and WorkerURN
type BaseURN struct {
	ProductID  string
	RegionID   string
	BusinessID string
	TenantID   string
	TypeSign   string
	Name       string
	Version    string
}

// String serializes elements of function URN struct to string
func (p *BaseURN) String() string {
	urn := fmt.Sprintf("%s:%s:%s:%s:%s:%s", p.ProductID, p.RegionID,
		p.BusinessID, p.TenantID, p.TypeSign, p.Name)
	if p.Version != "" {
		return fmt.Sprintf("%s:%s", urn, p.Version)
	}
	return urn
}

// ParseFrom parses elements from a function URN
func (p *BaseURN) ParseFrom(urn string) error {
	elements := strings.Split(urn, URNSep)
	urnLen := len(elements)
	if urnLen < urnLenWithoutVersion || urnLen > URNLenWithVersion {
		return fmt.Errorf("failed to parse urn from %s, invalid length", urn)
	}
	p.ProductID = elements[ProductIDIndex]
	p.RegionID = elements[RegionIDIndex]
	p.BusinessID = elements[BusinessIDIndex]
	p.TenantID = elements[TenantIDIndex]
	p.TypeSign = elements[FunctionSignIndex]
	p.Name = elements[FunctionNameIndex]
	if urnLen == URNLenWithVersion {
		p.Version = elements[VersionIndex]
	}
	return nil
}

// StringWithoutVersion return string without version
func (p *BaseURN) StringWithoutVersion() string {
	return fmt.Sprintf("%s:%s:%s:%s:%s:%s", p.ProductID, p.RegionID,
		p.BusinessID, p.TenantID, p.TypeSign, p.Name)
}

// GetFunctionInfo collects function information from a URN
func GetFunctionInfo(urn string) (BaseURN, error) {
	var parsedURN BaseURN
	if err := parsedURN.ParseFrom(urn); err != nil {
		logger.GetLogger().Errorf("error while parsing an URN: %s", err.Error())
		return BaseURN{}, fmt.Errorf("parsing an URN error: %s", err)
	}
	return parsedURN, nil
}

// CombineFunctionKey will generate funcKey from three IDs
func CombineFunctionKey(tenantID, funcName, version string) string {
	return fmt.Sprintf("%s/%s/%s", tenantID, funcName, version)
}
