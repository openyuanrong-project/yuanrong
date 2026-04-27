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

// Package aliasroute alias routing
package aliasroute

import (
	"sync"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

const (
	aliasURN = "sn:cn:yrk:default:function:helloworld:myaliasv1"
)

// TestCase init
func GetFakeAliasEle() *AliasElement {
	fakeAliasEle := &AliasElement{
		AliasURN:           aliasURN,
		FunctionURN:        "sn:cn:yrk:default:function:helloworld",
		FunctionVersionURN: "sn:cn:yrk:default:function:helloworld:$latest",
		Name:               "myaliasv1",
		FunctionVersion:    "$latest",
		RevisionID:         "20210617023315921",
		Description:        "",
		RoutingConfigs: []*routingConfig{
			{
				FunctionVersionURN: "sn:cn:yrk:default:function:helloworld:$latest",
				Weight:             60,
			},
			{
				FunctionVersionURN: "sn:cn:yrk:default:function:helloworld:v1",
				Weight:             40,
			},
		},
	}
	return fakeAliasEle
}

func GetFakeRuleAliasEle() *AliasElement {
	fakeAliasEle := &AliasElement{
		AliasURN:           "sn:cn:yrk:default:function:helloworld:myaliasrulev1",
		FunctionURN:        "sn:cn:yrk:default:function:helloworld",
		FunctionVersionURN: "sn:cn:yrk:default:function:helloworld:$latest",
		Name:               "myaliasrulev1",
		FunctionVersion:    "$latest",
		RevisionID:         "20210617023315921",
		Description:        "",
		RoutingType:        "rule",
		RoutingRules: routingRules{
			RuleLogic:   "and",
			Rules:       []string{"userType:=:VIP", "age:<=:20", "devType:in:P40,P50,MATE40"},
			GrayVersion: "sn:cn:yrk:172120022620195843:function:0@default@test_func:3",
		},
	}
	return fakeAliasEle
}
func ClearAliasRoute() {
	aliases = &Aliases{
		AliasMap: &sync.Map{},
	}
}

func TestOptAlias(t *testing.T) {
	ClearAliasRoute()
	defer ClearAliasRoute()
	convey.Convey("AddAlias success", t, func() {
		fakeAliasEle := GetFakeAliasEle()
		aliases.AddAlias(fakeAliasEle)
		ele, ok := aliases.AliasMap.Load(fakeAliasEle.AliasURN)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(ele, convey.ShouldNotBeNil)
	})
	convey.Convey("update Alias success", t, func() {
		fakeAliasEle := GetFakeAliasEle()
		fakeAliasEle.RoutingConfigs = []*routingConfig{
			{
				FunctionVersionURN: "sn:cn:yrk:default:function:helloworld:$latest",
				Weight:             50,
			},
			{
				FunctionVersionURN: "sn:cn:yrk:default:function:helloworld:v1",
				Weight:             50,
			},
		}
		aliases.AddAlias(fakeAliasEle)
		ele, ok := aliases.AliasMap.Load(fakeAliasEle.AliasURN)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(ele.(*AliasElement).RoutingConfigs[0].Weight, convey.ShouldEqual, 50)
		convey.So(ele.(*AliasElement).RoutingConfigs[1].Weight, convey.ShouldEqual, 50)
	})
	convey.Convey("remove Alias success", t, func() {
		fakeAliasEle := GetFakeAliasEle()
		aliases.AddAlias(fakeAliasEle)
		aliases.RemoveAlias(fakeAliasEle.AliasURN)
		ele, ok := aliases.AliasMap.Load(fakeAliasEle.AliasURN)
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(ele, convey.ShouldBeNil)
	})
}

func TestGetFuncURNFromAlias(t *testing.T) {
	ClearAliasRoute()
	defer ClearAliasRoute()
	convey.Convey("alias does not exist", t, func() {
		urn := aliases.GetFuncURNFromAlias(aliasURN)
		convey.So(urn, convey.ShouldEqual, aliasURN)
	})

	convey.Convey("alias get error", t, func() {
		aliases.AliasMap.Store(aliasURN, "456")
		urn := aliases.GetFuncURNFromAlias(aliasURN)
		aliases.AliasMap.Delete(aliasURN)
		convey.So(urn, convey.ShouldEqual, "")
	})
	convey.Convey("alias get error", t, func() {
		aliases.AddAlias(GetFakeAliasEle())
		urn := aliases.GetFuncURNFromAlias(aliasURN)
		convey.So(urn, convey.ShouldNotEqual, aliasURN)
		convey.So(urn, convey.ShouldNotEqual, "")
		convey.So(urn, convey.ShouldNotContainSubstring, "myaliasv1")
	})

}

func TestFetchInfoFromAliasKey(t *testing.T) {
	path := "/sn/aliases/business/yrk/tenant/default/function/helloworld/myalias"
	aliasKey := FetchInfoFromAliasKey(path)

	assert.Equal(t, aliasKey.FunctionID, "helloworld")
	assert.Equal(t, aliasKey.AliasName, "myalias")

	path = "/sn/aliases/business/yrk/tenant/default/function/helloworld"
	aliasKey = FetchInfoFromAliasKey(path)
	assert.Empty(t, aliasKey)
}

func TestBuildURNFromAliasKey(t *testing.T) {
	path := "/sn/aliases/business/yrk/tenant/default/function/helloworld/myalias"
	urn := BuildURNFromAliasKey(path)
	assert.Contains(t, urn, "myalias")
}

func TestGetFuncVersionURNWithParamsMatch(t *testing.T) {
	ClearAliasRoute()
	defer ClearAliasRoute()
	fakeAliasEle := GetFakeRuleAliasEle()
	aliases.AddAlias(fakeAliasEle)
	params := map[string]string{}
	params["userType"] = "VIP"
	params["age"] = "10"
	params["devType"] = "P40"

	aliasUrn := "sn:cn:yrk:default:function:helloworld:myaliasrulev1"
	wantFuncVer := "sn:cn:yrk:172120022620195843:function:0@default@test_func:3"
	got := GetAliases().GetFuncVersionURNWithParams(aliasUrn, params)
	assert.Equal(t, wantFuncVer, got)
}

func TestGetFuncVersionURNWithParamsNotMatch(t *testing.T) {
	ClearAliasRoute()
	defer ClearAliasRoute()
	fakeAliasEle := GetFakeRuleAliasEle()
	aliases.AddAlias(fakeAliasEle)
	params := map[string]string{}
	params["userType"] = "VIP"
	params["age"] = "50"
	params["devType"] = "P40"

	aliasUrn := "sn:cn:yrk:default:function:helloworld:myaliasrulev1"
	wantFuncVer := "sn:cn:yrk:default:function:helloworld:$latest"
	got := GetAliases().GetFuncVersionURNWithParams(aliasUrn, params)
	assert.Equal(t, wantFuncVer, got)
}

func TestUpdateAliasesMap(t *testing.T) {
	convey.Convey(
		"Test UpdateAliasesMap", t, func() {
			convey.Convey(
				"UpdateAliasesMap success", func() {
					fakeAliasEle := GetFakeAliasEle()
					convey.So(func() {
						UpdateAliasesMap([]*AliasElement{fakeAliasEle})
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestGetFuncVersionURNWithParams(t *testing.T) {
	convey.Convey(
		"Test GetFuncVersionURNWithParams", t, func() {
			convey.Convey(
				"GetFuncVersionURNWithParams success", func() {
					params := map[string]string{}
					params["userType"] = "VIP"
					params["age"] = "50"
					params["devType"] = "P40"
					str := GetAliases().GetFuncVersionURNWithParams("aliasURN", params)
					convey.So(str, convey.ShouldEqual, "aliasURN")
				},
			)
		},
	)
}

func TestGetFuncVersionURNByRule(t *testing.T) {
	convey.Convey(
		"Test getFuncVersionURNByRule", t, func() {
			convey.Convey(
				"getFuncVersionURNByRule success", func() {
					params := map[string]string{}
					ele := GetFakeAliasEle()
					var s sync.RWMutex
					ele.aliasLock = &s
					str := ele.getFuncVersionURNByRule(params)
					convey.So(str, convey.ShouldEqual, ele.FunctionVersionURN)
					params["userType"] = "VIP"
					str = ele.getFuncVersionURNByRule(params)
					convey.So(str, convey.ShouldEqual, ele.FunctionVersionURN)
				},
			)
		},
	)
}

func TestMatchRule(t *testing.T) {
	convey.Convey(
		"Test matchRule", t, func() {
			convey.Convey(
				"matchRule success", func() {
					flag := matchRule(map[string]string{}, []Expression{}, "ruleLogic")
					convey.So(flag, convey.ShouldBeFalse)
				},
			)
		},
	)
}

func TestIsMatch(t *testing.T) {
	convey.Convey(
		"Test isMatch", t, func() {
			convey.Convey(
				"isMatch success when ruleLogic==or", func() {
					flag := isMatch([]bool{false, false}, "or")
					convey.So(flag, convey.ShouldBeFalse)
				},
			)
			convey.Convey(
				"isMatch success when default", func() {
					flag := isMatch([]bool{false, false}, "non")
					convey.So(flag, convey.ShouldBeFalse)
				},
			)
		},
	)
}
