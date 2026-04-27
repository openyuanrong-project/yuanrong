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

// Package aliasroute alias routing in busclient
package aliasroute

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/loadbalance"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/urnutils"
)

const (
	weightRatio     = 100 // max weight of a node
	routingTypeRule = "rule"
	// AliasKeySeparator is the separator in an alias key
	AliasKeySeparator = "/"
)

const (
	defaultVersion    = "latest"
	defaultBusinessID = "yrk"
)

// example of an aliasKey:
// /<ProductID>/<AliasSign>/<BusinessSign>/<BusinessID>/<TenantSign>/<TenantID>/<FunctionSign>/<FunctionID>/<AliasName>
const (
	ProductIDIndex = iota + 1
	AliasSignIndex
	BusinessSignIndex
	BusinessIDIndex
	TenantSignIndex
	TenantIDIndex
	FunctionSignIndex
	FunctionIDIndex
	AliasNameIndex
	aliasKeyLength
)

// Aliases map for stateless function alias
type Aliases struct {
	AliasMap *sync.Map // Key: aliasURN -- Value: *AliasElement
}

// aliases for alias routing
var (
	aliases = &Aliases{
		AliasMap: &sync.Map{},
	}
)

// GetAliases -
func GetAliases() *Aliases {
	return aliases
}

// AddAlias add alias to Aliases map from etcd
func (a *Aliases) AddAlias(alias *AliasElement) {
	existAliasIf, exist := a.AliasMap.Load(alias.AliasURN)
	var existAlias *AliasElement
	var ok bool
	if !exist {
		// new alias, initialize RR and Mutex
		existAlias = &AliasElement{
			AliasURN:           alias.AliasURN,
			FunctionURN:        alias.FunctionURN,
			FunctionVersionURN: alias.FunctionVersionURN,
			Name:               alias.Name,
			Description:        alias.Description,
			FunctionVersion:    alias.FunctionVersion,
			RevisionID:         alias.RevisionID,
			RoutingConfigs:     alias.RoutingConfigs,
			RoutingRules:       alias.RoutingRules,
			RoutingType:        alias.RoutingType,

			lb:        loadbalance.LBFactory(loadbalance.RoundRobinNginx),
			aliasLock: &sync.RWMutex{},
		}
		existAlias.resetRR()
		a.AliasMap.Store(alias.AliasURN, existAlias)
		return
	}
	existAlias, ok = existAliasIf.(*AliasElement)
	if ok {
		aliasUpdate(existAlias, alias)
		existAlias.resetRR()
	}
}

// RemoveAlias remove alias to aliases map
func (a *Aliases) RemoveAlias(aliasURN string) {
	a.AliasMap.Delete(aliasURN)
}

// GetFuncURNFromAlias  If the alias exists, the weighted route version is returned.
// If the alias does not exist, the original URN is returned.
func (a *Aliases) GetFuncURNFromAlias(urn string) string {
	existAliasIf, exist := a.AliasMap.Load(urn)
	if !exist {
		return urn
	}
	existAlias, ok := existAliasIf.(*AliasElement)
	if !ok {
		log.GetLogger().Warnf("Failed to convert the alias urn %s", urn)
		return ""
	}
	return existAlias.getFuncVersionURN()
}

// GetFuncVersionURNWithParams gets the routing version URN of stateless functionName with parmas for rules
func (a *Aliases) GetFuncVersionURNWithParams(aliasURN string, params map[string]string) string {
	existAliasIf, exist := a.AliasMap.Load(aliasURN)
	if !exist {
		return aliasURN
	}
	existAlias, ok := existAliasIf.(*AliasElement)
	if !ok {
		log.GetLogger().Warnf("Failed to convert the alias urn %s", aliasURN)
		return ""
	}
	return existAlias.GetFuncVersionURNWithParams(params)
}

// CheckAliasRoutingChange - return false means oldURN is not equal to newURN or alise is not exist
func (a *Aliases) CheckAliasRoutingChange(aliasURN, oldURN string, params map[string]string) bool {
	existAliasIf, exist := a.AliasMap.Load(aliasURN)
	if !exist {
		return true
	}
	existAlias, ok := existAliasIf.(*AliasElement)
	if ok && existAlias.RoutingType == routingTypeRule {
		return oldURN != existAlias.getFuncVersionURNByRule(params)
	}
	// routingType is weight
	for _, config := range existAlias.RoutingConfigs {
		if config.FunctionVersionURN == oldURN && config.Weight > 0.0 {
			return false
		}
	}
	return true
}

// GetAliasRoutingType -
func (a *Aliases) GetAliasRoutingType(aliasURN string) string {
	existAliasIf, exist := a.AliasMap.Load(aliasURN)
	if !exist {
		return ""
	}
	if existAlias, ok := existAliasIf.(*AliasElement); ok {
		return existAlias.RoutingType
	}
	return ""
}

// change means the following 3 conditions
func isAliasWeightTypeChange(originAlias, srcAlias *AliasElement) map[string]int {
	changedURNMap := make(map[string]int)
	if originAlias.RoutingType == routingTypeRule || srcAlias.RoutingType == routingTypeRule {
		return map[string]int{"": NoneUpdate}
	}
	// 1、delete weight alias
	if len(srcAlias.RoutingConfigs) == 0 {
		return map[string]int{"": UpdateAllURN}
	}
	srcAliasMap := make(map[string]float64, len(srcAlias.RoutingConfigs))
	for _, config := range srcAlias.RoutingConfigs {
		if config.Weight <= 0 {
			// 2、weight decrease to 0
			changedURNMap[config.FunctionVersionURN] = UpdateWeightGreyURN
		}
		srcAliasMap[config.FunctionVersionURN] = config.Weight
	}

	for _, config := range originAlias.RoutingConfigs {
		// 3、grey functionURN in originAlias is not in srcAlias
		// old device still follow old urn, new device follow the weight
		if _, ok := srcAliasMap[config.FunctionVersionURN]; !ok {
			changedURNMap[config.FunctionVersionURN] = UpdateWeightGreyURN
		}
	}
	if originAlias.FunctionVersion != srcAlias.FunctionVersion {
		changedURNMap[originAlias.FunctionVersion] = UpdateMainURN
	}
	return changedURNMap
}

type routingRules struct {
	RuleLogic   string   `json:"ruleLogic"`
	Rules       []string `json:"rules"`
	GrayVersion string   `json:"grayVersion"`
}

// AliasElement struct stores an alias configs of stateless function
type AliasElement struct {
	aliasLock          *sync.RWMutex
	lb                 loadbalance.LoadBalance
	AliasURN           string           `json:"aliasUrn"`
	FunctionURN        string           `json:"functionUrn"`
	FunctionVersionURN string           `json:"functionVersionUrn"`
	Name               string           `json:"name"`
	FunctionVersion    string           `json:"functionVersion"`
	RevisionID         string           `json:"revisionId"`
	Description        string           `json:"description"`
	RoutingType        string           `json:"routingType"`
	RoutingRules       routingRules     `json:"routingRules"`
	RoutingConfigs     []*routingConfig `json:"routingconfig"`
}

type routingConfig struct {
	FunctionVersionURN string  `json:"functionVersionUrn"`
	Weight             float64 `json:"weight"`
}

func (a *AliasElement) getFuncVersionURN() string {
	a.aliasLock.RLock()
	defer a.aliasLock.RUnlock()
	funcVersion := a.lb.Next("", true)
	if funcVersion == nil {
		return ""
	}
	res, ok := funcVersion.(string)
	if !ok {
		return ""
	}
	return res
}

func (a *AliasElement) resetRR() {
	a.aliasLock.Lock()
	defer a.aliasLock.Unlock()
	a.lb.RemoveAll()
	for _, v := range a.RoutingConfigs {
		a.lb.Add(v.FunctionVersionURN, int(v.Weight*weightRatio))
	}
}

func (a *AliasElement) getFuncVersionURNByRule(params map[string]string) string {
	a.aliasLock.RLock()
	defer a.aliasLock.RUnlock()
	if len(params) == 0 {
		log.GetLogger().Warnf("params is empty, use default func version")
		return a.FunctionVersionURN
	}
	if len(a.RoutingRules.Rules) == 0 {
		log.GetLogger().Warnf("rule len is 0, use default func version")
		return a.FunctionVersionURN
	}

	matchRules, err := parseRules(a.RoutingRules)
	if err != nil {
		log.GetLogger().Warnf("parse rule error, use default func version: %s", err.Error())
		return a.FunctionVersionURN
	}

	// To obtain the final matching result by matching each rule and considering the "AND" or "OR"relationship of the rules
	matched := matchRule(params, matchRules, a.RoutingRules.RuleLogic)
	// got to default version if not matched
	if matched {
		return a.RoutingRules.GrayVersion
	}
	return a.FunctionVersionURN
}

// GetFuncVersionURNWithParams -
func (a *AliasElement) GetFuncVersionURNWithParams(params map[string]string) string {
	if a.RoutingType == routingTypeRule {
		return a.getFuncVersionURNByRule(params)
	}
	// default to go weight
	return a.getFuncVersionURN()
}

func aliasUpdate(destAlias, srcAlias *AliasElement) {
	destAlias.AliasURN = srcAlias.AliasURN
	destAlias.FunctionURN = srcAlias.FunctionURN
	destAlias.FunctionVersionURN = srcAlias.FunctionVersionURN
	destAlias.Name = srcAlias.Name
	destAlias.FunctionVersion = srcAlias.FunctionVersion
	destAlias.RevisionID = srcAlias.RevisionID
	destAlias.Description = srcAlias.Description
	destAlias.RoutingConfigs = srcAlias.RoutingConfigs
	destAlias.RoutingRules = srcAlias.RoutingRules
	destAlias.RoutingType = srcAlias.RoutingType
}

func ifAliasRoutingChanged(originAlias, srcAlias *AliasElement) map[string]int {
	changedURNMap := make(map[string]int)
	if originAlias.RoutingType == srcAlias.RoutingType {
		if originAlias.RoutingType == routingTypeRule {
			if !reflect.DeepEqual(originAlias.RoutingRules, srcAlias.RoutingRules) {
				changedURNMap[originAlias.RoutingRules.GrayVersion] = UpdateAllURN
			}
			if originAlias.FunctionVersionURN != srcAlias.FunctionVersionURN {
				changedURNMap[originAlias.FunctionVersionURN] = UpdateMainURN
			}
			return changedURNMap
		}
		return isAliasWeightTypeChange(originAlias, srcAlias)
	}
	// routingTypeWeight change to routingTypeRule
	if srcAlias.RoutingType == routingTypeRule {
		return map[string]int{"": UpdateAllURN}
	}
	// routingTypeRule change to routingTypeWeight
	for _, config := range srcAlias.RoutingConfigs {
		if config.Weight <= 0 {
			return map[string]int{"": UpdateAllURN}
		}
	}
	if len(srcAlias.RoutingConfigs) == 0 {
		return map[string]int{"": UpdateAllURN}
	}
	return map[string]int{"": NoneUpdate}
}

// AliasKey contains the elements of an alias key
type AliasKey struct {
	ProductID    string
	AliasSign    string
	BusinessSign string
	BusinessID   string
	TenantSign   string
	TenantID     string
	FunctionSign string
	FunctionID   string
	AliasName    string
}

// ParseFrom parses elements from an alias key
func (a *AliasKey) ParseFrom(aliasKeyStr string) error {
	elements := strings.Split(aliasKeyStr, AliasKeySeparator)
	urnLen := len(elements)
	if urnLen != aliasKeyLength {
		return fmt.Errorf("failed to parse an alias key %s, incorrect length", aliasKeyStr)
	}
	a.ProductID = elements[ProductIDIndex]
	a.AliasSign = elements[AliasSignIndex]
	a.BusinessSign = elements[BusinessSignIndex]
	a.BusinessID = elements[BusinessIDIndex]
	a.TenantSign = elements[TenantSignIndex]
	a.TenantID = elements[TenantIDIndex]
	a.FunctionSign = elements[FunctionSignIndex]
	a.FunctionID = elements[FunctionIDIndex]
	a.AliasName = elements[AliasNameIndex]
	return nil
}

// FetchInfoFromAliasKey collects alias information from an alias key
func FetchInfoFromAliasKey(aliasKeyStr string) *AliasKey {
	var aliasKey AliasKey
	if err := aliasKey.ParseFrom(aliasKeyStr); err != nil {
		log.GetLogger().Errorf("error while parsing an URN: %s", err.Error())
		return &AliasKey{}
	}
	return &aliasKey
}

// BuildURNFromAliasKey builds a URN from a alias key
func BuildURNFromAliasKey(aliasKeyStr string) string {
	aliasKey := FetchInfoFromAliasKey(aliasKeyStr)
	productURN := &urnutils.FunctionURN{
		ProductID:   urnutils.DefaultURNProductID,
		RegionID:    urnutils.DefaultURNRegion,
		BusinessID:  aliasKey.BusinessID,
		TenantID:    aliasKey.TenantID,
		TypeSign:    urnutils.DefaultURNFuncSign,
		FuncName:    aliasKey.FunctionID,
		FuncVersion: aliasKey.AliasName,
	}
	return productURN.String()
}

func parseRules(routingRules routingRules) ([]Expression, error) {
	rules := routingRules.Rules
	var expressions []Expression
	const expressionSize = 3
	for _, value := range rules {
		partition := strings.Split(value, ":")
		if len(partition) != expressionSize {
			return nil, fmt.Errorf("rules (%s) fields size not equal %v", value, expressionSize)
		}
		expression := Expression{
			leftVal:  partition[0],
			operator: partition[1],
			rightVal: partition[2],
		}
		expressions = append(expressions, expression)
	}
	return expressions, nil
}

func matchRule(params map[string]string, expressions []Expression, ruleLogic string) bool {
	var matchResultList []bool

	for _, exp := range expressions {
		matchResultList = append(matchResultList, exp.Execute(params))
	}
	if len(matchResultList) > 0 {
		return isMatch(matchResultList, ruleLogic)
	}
	return false
}

func isMatch(matchResultList []bool, ruleLogic string) bool {
	matchResult := matchResultList[0]
	if len(matchResultList) > 1 {
		switch ruleLogic {
		case "or":
			for _, value := range matchResultList {
				matchResult = matchResult || value
			}
		case "and":
			for _, value := range matchResultList {
				matchResult = matchResult && value
			}
		default:
			log.GetLogger().Warnf("unknow rulelogic: %s, return false", ruleLogic)
			return false
		}
	}
	return matchResult
}

// MarshalTenantAliasList marshal alias map to list with specific tenant id
func MarshalTenantAliasList(tenantID string) ([]byte, error) {
	var aliasList []*AliasElement
	GetAliases().AliasMap.Range(func(key, value interface{}) bool {
		aliasElement, _ := value.(*AliasElement)
		if urnutils.CheckAliasUrnTenant(tenantID, aliasElement.AliasURN) {
			aliasList = append(aliasList, aliasElement)
			return true
		}
		return true
	})
	aliasData, err := json.Marshal(aliasList)
	if err != nil {
		return nil, errors.New("marshal alias list error")
	}
	return aliasData, nil
}

// ProcessDelete -
func ProcessDelete(event *etcd3.Event) string {
	aliasURN := BuildURNFromAliasKey(event.Key)
	GetAliases().RemoveAlias(aliasURN)
	return aliasURN
}

// ProcessUpdate -
func ProcessUpdate(event *etcd3.Event) (string, error) {
	alias := &AliasElement{}
	err := json.Unmarshal(event.Value, alias)
	if err != nil {
		log.GetLogger().Errorf("failed to unmarshal alias event, err: %s", err.Error())
		return "", err
	}
	GetAliases().AddAlias(alias)
	return alias.AliasURN, nil
}

// ResolveAliasedFunctionNameToURN - {functionName}:{alias|version} 解析别名路由
func ResolveAliasedFunctionNameToURN(functionNameWithAlias string, tenantID string, params map[string]string) string {
	splits := strings.Split(functionNameWithAlias, ":")
	if len(splits) > 2 || len(splits) == 0 { // {functionName}:{alias|version}
		return ""
	}

	if len(splits) == 1 {
		return urnutils.BuildURNOrAliasURNTemp(defaultBusinessID, tenantID,
			urnutils.BuildStandardFunctionName(functionNameWithAlias), defaultVersion)
	}

	functionName := urnutils.BuildStandardFunctionName(splits[0])
	versionOrAlias := splits[1]
	_, err := strconv.Atoi(versionOrAlias)
	if err != nil && versionOrAlias != defaultVersion {
		aliasUrn := urnutils.BuildURNOrAliasURNTemp(defaultBusinessID, tenantID, functionName, versionOrAlias)
		return GetAliases().GetFuncVersionURNWithParams(aliasUrn, params)
	}
	return urnutils.BuildURNOrAliasURNTemp(defaultBusinessID, tenantID, functionName, versionOrAlias)
}
