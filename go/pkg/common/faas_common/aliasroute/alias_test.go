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
	"fmt"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
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

func GetFakeWeightAliasEle() *AliasElement {
	fakeAliasEle := &AliasElement{
		AliasURN:           "sn:cn:yrk:default:function:0@default@aliasfunc:myaliasrulev1",
		FunctionURN:        "sn:cn:yrk:default:function:0@default@aliasfunc",
		FunctionVersionURN: "sn:cn:yrk:c53626012ba84727b938ca8bf03108ef:function:0@default@aliasfunc:latest",
		Name:               "myaliasrulev1",
		FunctionVersion:    "$latest",
		RevisionID:         "20210617023315921",
		Description:        "",
		RoutingType:        "weigh",
		RoutingConfigs: []*routingConfig{{
			FunctionVersionURN: "sn:cn:yrk:c53626012ba84727b938ca8bf03108ef:function:0@default@aliasfunc:latest",
			Weight:             80,
		}, {
			FunctionVersionURN: "sn:cn:yrk:c53626012ba84727b938ca8bf03108ef:function:0@default@aliasfunc:1",
			Weight:             0,
		}},
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

func TestMarshalTenantAliasList(t *testing.T) {
	ClearAliasRoute()
	defer ClearAliasRoute()
	fakeAliasEle := GetFakeRuleAliasEle()
	aliases.AddAlias(fakeAliasEle)
	params := map[string]string{}
	params["userType"] = "VIP"
	params["age"] = "10"
	params["devType"] = "P40"

	type args struct {
		tenantID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"case1", args{tenantID: "default"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := MarshalTenantAliasList(tt.args.tenantID)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalTenantAliasList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestCheckUrnWithParamsMatchRules(t *testing.T) {
	convey.Convey("CheckAliasRoutingChange", t, func() {
		aliases.AddAlias(GetFakeRuleAliasEle())

		aliasRuleURN := "sn:cn:yrk:default:function:helloworld:myaliasrulev1"
		urnWithParam_old := "sn:cn:yrk:172120022620195843:function:0@default@test_func:latest"
		convey.So(aliases.CheckAliasRoutingChange(aliasRuleURN, urnWithParam_old, make(map[string]string)),
			convey.ShouldEqual, true)

		convey.So(aliases.CheckAliasRoutingChange(aliasRuleURN, urnWithParam_old, make(map[string]string)),
			convey.ShouldEqual, true)

		aliases.AddAlias(GetFakeWeightAliasEle())
		aliasWeight := "sn:cn:yrk:default:function:0@default@aliasfunc:myaliasrulev1"
		aliasURN_old := "sn:cn:yrk:c53626012ba84727b938ca8bf03108ef:function:0@default@aliasfunc:1"
		convey.So(aliases.CheckAliasRoutingChange(aliasWeight, aliasURN_old, make(map[string]string)),
			convey.ShouldEqual, true)

		convey.So(aliases.CheckAliasRoutingChange(aliasWeight, "old alias urn needed update session",
			make(map[string]string)), convey.ShouldEqual, true)
	})
}

func TestAliasWeightLoadBalancer(t *testing.T) {
	convey.Convey("AliasWeightLoadBalancer", t, func() {
		fakeAliasEle := &AliasElement{
			AliasURN:           aliasURN,
			FunctionURN:        "sn:cn:yrk:default:function:helloworld",
			FunctionVersionURN: "sn:cn:yrk:default:function:helloworld:1",
			Name:               "myaliasv1",
			FunctionVersion:    "1",
			RevisionID:         "20210617023315921",
			Description:        "",
			RoutingConfigs: []*routingConfig{
				{
					FunctionVersionURN: "sn:cn:yrk:default:function:helloworld:2",
					Weight:             80,
				},
				{
					FunctionVersionURN: "sn:cn:yrk:default:function:helloworld:1",
					Weight:             20,
				},
			},
		}
		ClearAliasRoute()
		aliases.AddAlias(fakeAliasEle)

		aliasElementIf, _ := aliases.AliasMap.Load(fakeAliasEle.AliasURN)
		aliasElement := aliasElementIf.(*AliasElement)
		urnMap1 := []string{}
		urnMap2 := make([]string, 50)
		for i := 0; i < 50; i++ {
			urn := aliasElement.getFuncVersionURN()
			urnMap1 = append(urnMap1, urn)
		}
		var count int
		for index, urn := range urnMap1 {
			if urn == "sn:cn:yrk:default:function:helloworld:2" {
				count++
				urnMap2[index] = urn
			}
		}
		convey.So(count, convey.ShouldEqual, 40)

		for i := 0; i < 50; i++ {
			if urnMap2[i] != "" {
				newUrn := aliasElement.getFuncVersionURN()
				if newUrn != urnMap2[i] {
					fmt.Printf("index:%d oldUrn:%s, newUrn:%s \n", i, urnMap2[i], newUrn)
				}
				urnMap2[i] = newUrn
			}
		}
		count = 0
		for _, urn := range urnMap2 {
			if urn == "sn:cn:yrk:default:function:helloworld:2" {
				count++
			}
		}
		convey.So(count, convey.ShouldEqual, 32)
	})
}

func Test_ifAliasRoutingChanged(t *testing.T) {
	convey.Convey("ifAliasRoutingChanged", t, func() {
		convey.Convey("same type weight UpdateAllURN", func() {
			origin := &AliasElement{RoutingType: "weight"}
			newAlias := &AliasElement{RoutingType: "weight"}
			mapEvent := ifAliasRoutingChanged(origin, newAlias)
			convey.So(mapEvent[""], convey.ShouldEqual, UpdateAllURN)
		})

		convey.Convey("same type weight UpdateWeightGreyURN UpdateMainURN", func() {
			origin := &AliasElement{RoutingType: "weight", RoutingConfigs: []*routingConfig{
				{
					FunctionVersionURN: "function/latest",
					Weight:             100,
				},
			},
				FunctionVersion: "0",
			}
			newAlias := &AliasElement{RoutingType: "weight", RoutingConfigs: []*routingConfig{
				{
					FunctionVersionURN: "function/1",
					Weight:             80,
				},
				{
					FunctionVersionURN: "function/2",
					Weight:             20,
				},
				{
					FunctionVersionURN: "function/3",
					Weight:             0,
				},
			},
				FunctionVersion: "1",
			}
			mapEvent := ifAliasRoutingChanged(origin, newAlias)
			convey.So(mapEvent["function/3"], convey.ShouldEqual, UpdateWeightGreyURN)
			convey.So(mapEvent["function/latest"], convey.ShouldEqual, UpdateWeightGreyURN)
			convey.So(mapEvent["0"], convey.ShouldEqual, UpdateMainURN)
		})

		convey.Convey("same type rule", func() {
			origin := &AliasElement{RoutingType: routingTypeRule, RoutingRules: routingRules{
				RuleLogic:   "and",
				Rules:       nil,
				GrayVersion: "0",
			},
				FunctionVersionURN: "function/0",
			}
			newAlias := &AliasElement{RoutingType: routingTypeRule, RoutingRules: routingRules{
				RuleLogic:   "or",
				Rules:       nil,
				GrayVersion: "1",
			},
				FunctionVersionURN: "function/1",
			}
			mapEvent := ifAliasRoutingChanged(origin, newAlias)
			convey.So(mapEvent[origin.RoutingRules.GrayVersion], convey.ShouldEqual, UpdateAllURN)
			convey.So(mapEvent[origin.FunctionVersionURN], convey.ShouldEqual, UpdateMainURN)
		})

		convey.Convey("different type rule weight", func() {
			newAlias := &AliasElement{RoutingType: routingTypeRule, RoutingRules: routingRules{
				RuleLogic:   "and",
				Rules:       nil,
				GrayVersion: "0",
			},
				FunctionVersionURN: "function/0",
			}
			origin := &AliasElement{RoutingType: "weight", RoutingConfigs: []*routingConfig{
				{
					FunctionVersionURN: "function/latest",
					Weight:             100,
				},
			},
				FunctionVersion:    "0",
				FunctionVersionURN: "function/1",
			}
			mapEvent := ifAliasRoutingChanged(origin, newAlias)
			convey.So(mapEvent[""], convey.ShouldEqual, UpdateAllURN)
		})

		convey.Convey("different type weight rule", func() {
			origin := &AliasElement{RoutingType: routingTypeRule, RoutingRules: routingRules{
				RuleLogic:   "and",
				Rules:       nil,
				GrayVersion: "0",
			},
				FunctionVersionURN: "function/0",
			}
			newAlias := &AliasElement{RoutingType: "weight", RoutingConfigs: []*routingConfig{
				{
					FunctionVersionURN: "function/latest",
					Weight:             0,
				},
			},
				FunctionVersion:    "0",
				FunctionVersionURN: "function/1",
			}
			mapEvent := ifAliasRoutingChanged(origin, newAlias)
			convey.So(mapEvent[""], convey.ShouldEqual, UpdateAllURN)

			newAlias1 := &AliasElement{RoutingType: "weight", RoutingConfigs: []*routingConfig{
				{
					FunctionVersionURN: "function/latest",
					Weight:             0,
				},
			},
				FunctionVersion:    "0",
				FunctionVersionURN: "function/1",
			}
			mapEvent1 := ifAliasRoutingChanged(origin, newAlias1)
			convey.So(mapEvent1[""], convey.ShouldEqual, UpdateAllURN)

			newAlias2 := &AliasElement{RoutingType: "weight", RoutingConfigs: []*routingConfig{
				{
					FunctionVersionURN: "function/latest",
					Weight:             100,
				},
			},
				FunctionVersion:    "0",
				FunctionVersionURN: "function/1",
			}
			mapEvent2 := ifAliasRoutingChanged(origin, newAlias2)
			convey.So(mapEvent2[""], convey.ShouldEqual, NoneUpdate)
		})
	})
}

func TestResolveAliasedFunctionNameToURN(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(GetAliases().GetFuncVersionURNWithParams, func(aliasUrn string, params map[string]string) string {
		return "resolved_" + aliasUrn
	})

	testCases := []struct {
		name                  string
		functionNameWithAlias string
		tenantID              string
		params                map[string]string
		expectedURN           string
	}{
		{
			name:                  "Simple function name without alias",
			functionNameWithAlias: "myFunction",
			tenantID:              "tenant1",
			params:                nil,
			expectedURN:           "sn:cn:yrk:tenant1:function:0@default@myFunction:latest",
		},
		{
			name:                  "Function name with version number",
			functionNameWithAlias: "myFunction:2",
			tenantID:              "tenant1",
			params:                nil,
			expectedURN:           "sn:cn:yrk:tenant1:function:0@default@myFunction:2",
		},
		{
			name:                  "Function name with alias",
			functionNameWithAlias: "myFunction:prod",
			tenantID:              "tenant1",
			params:                map[string]string{"key": "value"},
			expectedURN:           "sn:cn:yrk:tenant1:function:0@default@myFunction:prod",
		},
		{
			name:                  "Invalid function name (too many splits)",
			functionNameWithAlias: "myFunction:prod:extra",
			tenantID:              "tenant1",
			params:                nil,
			expectedURN:           "",
		},
		{
			name:                  "Empty function name",
			functionNameWithAlias: "",
			tenantID:              "tenant1",
			params:                nil,
			expectedURN:           "sn:cn:yrk:tenant1:function:0@default@:latest",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ResolveAliasedFunctionNameToURN(tc.functionNameWithAlias, tc.tenantID, tc.params)
			assert.Equal(t, tc.expectedURN, result, "URN resolution should match expected output")
		})
	}
}
