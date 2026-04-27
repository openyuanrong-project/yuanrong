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

// Package urnutils contains URN element definitions and tools
package urnutils

import (
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
)

var (
	once     sync.Once
	serverIP = ""
)

const (
	funcNamePrefix        = "0@default@"
	shortFuncNameSplit    = 1
	standardFuncNameSplit = 3
)

// example of function URN: <ProductID>:<RegionID>:<BusinessID>:<TenantID>:<FunctionSign>:<FunctionName>:<FuncVersion>
// Indices of elements in FunctionURN
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

// An example of a function functionkey: <TenantID>/<FunctionName>/<FuncVersion>
const (
	// TenantIDIndexKey is the index of the tenant ID in a functionkey
	TenantIDIndexKey = iota
	// FunctionNameIndexKey is the index of the function name in a functionkey
	FunctionNameIndexKey
	// VersionIndexKey is the index of the version in a functionkey
	VersionIndexKey
)

const (
	// TenantMetadataTenantIndex is the index of the tenant ID in a tenantMetadataEtcdKey
	TenantMetadataTenantIndex = 6
)

const (
	urnLenWithoutVersion = URNLenWithVersion - 1
	// URNSep is a URN separator of functions
	URNSep = ":"
	// FunctionKeySep is a functionkey separator of functions
	FunctionKeySep = "/"
	// DefaultURNProductID is the default product ID of a URN
	DefaultURNProductID = "sn"
	// DefaultURNRegion is the default region of a URN
	DefaultURNRegion = "cn"
	// DefaultURNFuncSign is the default function sign of a URN
	DefaultURNFuncSign  = "function"
	defaultURNLayerSign = "layer"
	anonymization       = "****"
	anonymizeLen        = 3

	// BranchAliasPrefix is used to remove "!" from aliasing rules at the begining of "!"
	BranchAliasPrefix = 1
	// BranchAliasRule is an aliased rule that begins with an "!"
	BranchAliasRule        = "!"
	functionNameStartIndex = 2
	// ServiceNameIndex is index of service name in urn
	ServiceNameIndex = 1
	funcNameMinLen   = 3
	// defaultFunctionMaxLen is max length of function name
	defaultFunctionMaxLen = 128
)

//	An example of a worker-manager URN:
//
// /sn/workers/business/iot/tenant/j0f4413f7b4b4c33be576d432f7ee085/function/functest/version/$latest
// /cn-north-1a/cn-north-1a-#-ws-j0f4413f7b-functest-faaslatest-deployment-55b5f9dcb7-r2dsv
const (
	// URNIndexZero URN index 0
	URNIndexZero = iota
	// URNIndexOne URN index 1
	URNIndexOne
	// URNIndexTwo URN index 2
	URNIndexTwo
	// URNIndexThree URN index 3
	URNIndexThree
	// URNIndexFour URN index 4
	URNIndexFour
	// URNIndexFive URN index 5
	URNIndexFive
	// URNIndexSix URN index 6
	URNIndexSix
	// URNIndexSeven URN index 7
	URNIndexSeven
	// URNIndexEight URN index 8
	URNIndexEight
	// URNIndexNine URN index 9
	URNIndexNine
	// URNIndexTen URN index 10
	URNIndexTen
	// URNIndexEleven URN index 11
	URNIndexEleven
	// URNIndexTwelve URN index 12
	URNIndexTwelve
	// URNIndexThirteen URN index 13
	URNIndexThirteen
)
const (
	k8sLabelLen   = 63
	otherStrLen   = 4
	crHashMaxLen  = 10
	versionManLen = 30
)

const (
	// OwnerReadWrite -
	OwnerReadWrite = 416 // 640:rw- r-- ---
	// DefaultMode -
	DefaultMode = 420 // 644:rw- r-- r--
	// CertMode -
	CertMode = 384 // 600:rw- --- ---
)

var (
	functionGraphFuncNameRegexp = regexp.MustCompile("^[a-zA-Z]([a-zA-Z0-9_-]*[a-zA-Z0-9])?$")
)

// FunctionURN contains elements of a product URN. It can expand to FunctionURN, LayerURN and WorkerURN
type FunctionURN struct {
	ProductID   string
	RegionID    string
	BusinessID  string
	TenantID    string
	TypeSign    string
	FuncName    string
	FuncVersion string
}

// String serializes elements of function URN struct to string
func (p *FunctionURN) String() string {
	urn := fmt.Sprintf("%s:%s:%s:%s:%s:%s", p.ProductID, p.RegionID,
		p.BusinessID, p.TenantID, p.TypeSign, p.FuncName)
	if p.FuncVersion != "" {
		return fmt.Sprintf("%s:%s", urn, p.FuncVersion)
	}
	return urn
}

// ParseFrom parses elements from a function URN
func (p *FunctionURN) ParseFrom(urn string) error {
	elements := strings.Split(urn, URNSep)
	urnLen := len(elements)
	if urnLen < urnLenWithoutVersion || urnLen > URNLenWithVersion {
		return fmt.Errorf("failed to parse urn from: %s, invalid length: %d", urn, urnLen)
	}
	p.ProductID = elements[ProductIDIndex]
	p.RegionID = elements[RegionIDIndex]
	p.BusinessID = elements[BusinessIDIndex]
	p.TenantID = elements[TenantIDIndex]
	p.TypeSign = elements[FunctionSignIndex]
	p.FuncName = elements[FunctionNameIndex]
	if urnLen == URNLenWithVersion {
		p.FuncVersion = elements[VersionIndex]
	}
	return nil
}

// StringWithoutVersion return string without version
func (p *FunctionURN) StringWithoutVersion() string {
	return fmt.Sprintf("%s:%s:%s:%s:%s:%s", p.ProductID, p.RegionID,
		p.BusinessID, p.TenantID, p.TypeSign, p.FuncName)
}

// GetFunctionInfo collects function information from a URN
func GetFunctionInfo(urn string) (FunctionURN, error) {
	var parsedURN FunctionURN
	if err := parsedURN.ParseFrom(urn); err != nil {
		log.GetLogger().Errorf("error while parsing an URN: %s", err.Error())
		return FunctionURN{}, fmt.Errorf("parsing an URN error: %s", err)
	}
	return parsedURN, nil
}

// GetFuncInfoWithVersion collects function information and distinguishes if the URN contains a version
func GetFuncInfoWithVersion(urn string) (FunctionURN, error) {
	parsedURN, err := GetFunctionInfo(urn)
	if err != nil {
		return parsedURN, err
	}
	if parsedURN.FuncVersion == "" {
		log.GetLogger().Errorf("incorrect URN length: %s", Anonymize(urn))
		return parsedURN, errors.New("incorrect URN length, no version")
	}
	return parsedURN, nil
}

// ParseAliasURN is used to remove "!" from the beginning of the alias
func ParseAliasURN(aliasURN string) string {
	elements := strings.Split(aliasURN, URNSep)
	if len(elements) == URNLenWithVersion {
		if strings.HasPrefix(elements[VersionIndex], BranchAliasRule) {
			elements[VersionIndex] = elements[VersionIndex][BranchAliasPrefix:]
		}
		return strings.Join(elements, ":")
	}
	return aliasURN
}

// GetAlias returns an alias
func (p *FunctionURN) GetAlias() string {
	if p.FuncVersion == constant.DefaultURNVersion {
		return ""
	}
	if _, err := strconv.Atoi(p.FuncVersion); err == nil {
		return ""
	}
	return p.FuncVersion
}

// GetAliasForFuncBranch returns an alias for function branch
func (p *FunctionURN) GetAliasForFuncBranch() string {
	if strings.HasPrefix(p.FuncVersion, BranchAliasRule) {
		// remove "!" from the beginning of the alias
		return p.FuncVersion[BranchAliasPrefix:]
	}
	return ""
}

// Valid check whether the self-verification function name complies with the specifications.
func (p *FunctionURN) Valid() error {
	serviceID, functionName, err := GetFunctionNameAndServiceName(p.FuncName)
	if err != nil {
		log.GetLogger().Errorf("failed to get serviceID and functionName")
		return err
	}
	if !(functionGraphFuncNameRegexp.MatchString(serviceID) ||
		functionGraphFuncNameRegexp.MatchString(functionName)) {
		errmsg := "failed to match reg%s"
		log.GetLogger().Errorf(errmsg, functionGraphFuncNameRegexp)
		return fmt.Errorf(errmsg, functionGraphFuncNameRegexp)
	}
	if len(serviceID) > defaultFunctionMaxLen || len(functionName) > defaultFunctionMaxLen {
		errmsg := "serviceID or functionName's len is out of range %d"
		log.GetLogger().Errorf(errmsg, defaultFunctionMaxLen)
		return fmt.Errorf(errmsg, defaultFunctionMaxLen)
	}
	return nil
}

// GetServiceNameFromFullName -
func GetServiceNameFromFullName(funcName string) string {
	if strings.HasPrefix(funcName, ServicePrefix) {
		splits := strings.Split(funcName, "@") // 0@default@funcName
		if len(splits) != funcNameMinLen {
			return ""
		}
		return splits[ServiceNameIndex]
	}
	return ""
}

// GetFunctionNameAndServiceName returns serviceName and FunctionName
func GetFunctionNameAndServiceName(funcName string) (string, string, error) {
	if strings.HasPrefix(funcName, ServiceIDPrefix) {
		split := strings.Split(funcName, separator)
		if len(split) < funcNameMinLen {
			log.GetLogger().Errorf("incorrect function name length: %s", len(split))
			return "", "", errors.New("parsing a function name error")
		}
		return split[ServiceNameIndex], strings.Join(split[functionNameStartIndex:], separator), nil
	}
	log.GetLogger().Errorf("incorrect function name: %s", funcName)
	return "", "", errors.New("parsing a function name error")
}

// Anonymize anonymize input str to xxx****xxx
func Anonymize(str string) string {
	if len(str) < anonymizeLen+1+anonymizeLen {
		return anonymization
	}
	return str[:anonymizeLen] + anonymization + str[len(str)-anonymizeLen:]
}

// AnonymizeTenantURN Anonymize tenant info in urn
func AnonymizeTenantURN(urn string) string {
	elements := strings.Split(urn, URNSep)
	urnLen := len(elements)
	if urnLen < urnLenWithoutVersion || urnLen > URNLenWithVersion {
		return urn
	}
	elements[TenantIDIndex] = Anonymize(elements[TenantIDIndex])
	return strings.Join(elements, URNSep)
}

// AnonymizeTenantKey Anonymize tenant info in functionkey
func AnonymizeTenantKey(functionKey string) string {
	elements := strings.Split(functionKey, FunctionKeySep)
	keyLen := len(elements)
	if TenantIDIndexKey >= keyLen {
		return functionKey
	}
	elements[TenantIDIndexKey] = Anonymize(elements[TenantIDIndexKey])
	return strings.Join(elements, FunctionKeySep)
}

// AnonymizeTenantURNSlice Anonymize tenant info in urn slice
func AnonymizeTenantURNSlice(urns []string) []string {
	var anonymizeUrns []string
	for i := 0; i < len(urns); i++ {
		anonymizeUrn := AnonymizeTenantURN(urns[i])
		anonymizeUrns = append(anonymizeUrns, anonymizeUrn)
	}
	return anonymizeUrns
}

// AnonymizeTenantMetadataEtcdKey Anonymize tenant info in tenant metadata etcd key
// /sn/quota/cluster/cluster001/tenant/7e1ad6a6-cc5c-44fa-bd54-25873f72a86a/instancemetadata
func AnonymizeTenantMetadataEtcdKey(etcdKey string) string {
	elements := strings.Split(etcdKey, "/")
	if len(elements) <= TenantMetadataTenantIndex {
		return etcdKey
	}
	elements[TenantMetadataTenantIndex] = Anonymize(elements[TenantMetadataTenantIndex])
	return strings.Join(elements, "/")
}

// AnonymizeKeys - anonymize the input slice of string to slice of xxx****xxx
// data system key example: 638cf733-a625-4850-9f23-9ef49873f5a3;2ba6f9cd-c8d3-4655-a9d0-e67d7abcfb3f
func AnonymizeKeys(keys []string) []string {
	res := make([]string, len(keys))
	for i, str := range keys {
		res[i] = Anonymize(str)
	}
	return res
}

// BuildURNOrAliasURNTemp - build urn format
func BuildURNOrAliasURNTemp(business, tenant, function, versionOrAlias string) string {
	if business == "" || tenant == "" || function == "" || versionOrAlias == "" {
		return ""
	}
	return fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s", DefaultURNProductID, DefaultURNRegion,
		business, tenant, DefaultURNFuncSign, function, versionOrAlias)
}

// GetServerIP -
func GetServerIP() (string, error) {
	var err error
	once.Do(func() {
		addr, errMsg := GetHostAddr()
		if errMsg != nil {
			err = errMsg
			return
		}
		serverIP = addr[0]
	})
	return serverIP, err
}

// GetHostAddr -
func GetHostAddr() ([]string, error) {
	name, err := os.Hostname()
	if err != nil {
		log.GetLogger().Errorf("get hostname failed: %v", err)
		return nil, err
	}

	addrs, err := net.LookupHost(name)
	if err != nil || len(addrs) == 0 {
		log.GetLogger().Errorf("look up host by name failed")
		return nil, fmt.Errorf("look up host by name failed")
	}
	return addrs, nil
}

// CrNameByURN returns a CR name by URN
func CrNameByURN(urn string) string {
	if len(urn) == URNIndexZero {
		return ""
	}
	baseUrn, err := GetFunctionInfo(urn)
	if err != nil {
		return ""
	}
	return CrName(baseUrn.BusinessID, baseUrn.TenantID, baseUrn.FuncName, baseUrn.FuncVersion)
}

// CrName CR Name
// [y/z]brief-functionname-version-hash
func CrName(business, tenant, funcName, version string) string {
	hashStr := genFunctionCRStr(business, tenant, funcName, version)
	crHash := utils.FnvHash(hashStr)
	if len(crHash) > crHashMaxLen {
		crHash = crHash[:crHashMaxLen]
	}
	brief := acquireBrief(business, tenant)
	ver := VersionConvForBranch(version)
	// cannot contain (urnutils.separator, ususually @) or _. If contains, replace it with -.
	funcName = strings.ReplaceAll(funcName, "@", "-")
	funcName = strings.ReplaceAll(funcName, "_", "-")

	// otherStrLen is 4 contains three - and a z or y.
	shortFunctionNameLen := k8sLabelLen - len(brief) - len(ver) - len(crHash) - otherStrLen
	// funcName prefix is 0- means funcName has joint sn service id
	// k8s label max length is 63, so cr name need to delete sn service id
	// otherwise, cr name length more than 63 characters, error
	if strings.HasPrefix(funcName, ServiceIDPrefix) && len(funcName) > shortFunctionNameLen {
		funcName = acquireShorter(funcName, shortFunctionNameLen)
	}

	crName := brief + "-" + funcName + "-" + ver + "-" + crHash
	crNameLower := strings.ToLower(crName)
	if crName == crNameLower {
		return "y" + crNameLower
	}

	return "z" + crNameLower
}

func genFunctionCRStr(business string, tenant string, funcName string, version string) string {
	return business + "-" + tenant + "-" + funcName + "-" + version
}

func acquireBrief(business, tenant string) string {
	if len(business) > URNIndexFour {
		business = business[:URNIndexFour]
	}
	product, tenant := splitTenant(tenant)
	if len(tenant) > URNIndexFour {
		tenant = tenant[:URNIndexFour]
	}

	if len(product) > URNIndexFour {
		product = product[:URNIndexFour]
	}

	return business + tenant + product
}

func splitTenant(tenant string) (string, string) {
	var product string
	t := strings.Split(tenant, TenantProductSplitStr)
	l := len(t)
	if l == URNIndexOne {
		return product, tenant
	}
	if l == URNIndexTwo {
		tenant = t[URNIndexZero]
		product = t[URNIndexOne]
		return product, tenant
	}
	return "", product
}

// VersionConvForBranch return version Conv for branch
func VersionConvForBranch(v string) string {
	// cannot contain _. If the version cr contains _, replace it with -.
	version := strings.ReplaceAll(v, "_", "-")
	if len(version) > versionManLen {
		version = version[:versionManLen]
	}
	return version
}

// if funcName contains sn service id, this method can acquire
// first 4 character of sn id and real function name with split _
// return shorter serviceID and shorter funcName
func acquireShorter(funcName string, functionNameLen int) string {
	shorterFuncName := []rune(funcName)
	return string(shorterFuncName[len(shorterFuncName)-functionNameLen : len(shorterFuncName)-1])
}

// GetTenantFromFuncKey -
func GetTenantFromFuncKey(funcKey string) string {
	elements := strings.Split(funcKey, FunctionKeySep)
	keyLen := len(elements)
	if keyLen != URNIndexThree {
		return ""
	}
	return elements[TenantIDIndexKey]
}

// GetFuncNameFromFuncKey -
func GetFuncNameFromFuncKey(funcKey string) string {
	elements := strings.Split(funcKey, FunctionKeySep)
	keyLen := len(elements)
	if keyLen != URNIndexThree {
		return ""
	}
	return elements[TenantIDIndexKey] + FunctionKeySep + elements[FunctionNameIndexKey]
}

// GetTenantFromAliasUrn -
func GetTenantFromAliasUrn(aliasUrn string) string {
	elements := strings.Split(aliasUrn, URNSep)
	keyLen := len(elements)
	if keyLen != URNIndexSeven {
		return ""
	}
	return elements[URNIndexThree]
}

// CheckAliasUrnTenant -
func CheckAliasUrnTenant(tenantID string, aliasUrn string) bool {
	if GetTenantFromAliasUrn(aliasUrn) != "" &&
		GetTenantFromAliasUrn(aliasUrn) == tenantID {
		return true
	}
	return false
}

// CombineFunctionKey will generate funcKey from three IDs
func CombineFunctionKey(tenantID, funcName, version string) string {
	return fmt.Sprintf("%s/%s/%s", tenantID, funcName, version)
}

// GetShortFuncName -
func GetShortFuncName(funcName string) string {
	if len(funcName) > k8sLabelLen {
		// labels must begin and end with an alphanumeric character, so set first character always X
		funcName = "X" + funcName[len(funcName)-k8sLabelLen+1:]
	}
	return funcName
}

// BuildStandardFunctionName - 将不带版本、别名的方法名拼接成0@default@开头的完整方法名
func BuildStandardFunctionName(functionName string) string {
	splits := strings.Split(functionName, "@")
	if len(splits) != shortFuncNameSplit && len(splits) != standardFuncNameSplit {
		return ""
	}
	standardFunctionName := functionName
	if len(splits) == shortFuncNameSplit {
		standardFunctionName = funcNamePrefix + standardFunctionName
	}
	return standardFunctionName
}
