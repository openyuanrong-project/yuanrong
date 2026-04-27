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

package aliasroute

import (
	"fmt"
	"strings"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

type ExpressionTestSuite struct {
	alias AliasElement
}

func (suite *ExpressionTestSuite) SetupTest() {

}

func (suite *ExpressionTestSuite) TearDownTest() {

}

func (suite *ExpressionTestSuite) TestEquel() {

}

func genExpression(str string) (Expression, error) {
	partition := strings.Split(str, ":")
	if len(partition) != expressionSize {
		return Expression{}, fmt.Errorf("express(#{str}) string format is error")
	}
	return Expression{
		leftVal:  partition[0],
		operator: partition[1],
		rightVal: partition[2],
	}, nil
}

func ExecuteExp(t *testing.T, expStr string, params map[string]string) bool {
	exp, err := genExpression(expStr)
	if err != nil {
		t.Error("gen expression fail: ", expStr)
		return false
	}
	return exp.Execute(params)
}

func TestExpEq(t *testing.T) {
	params := map[string]string{}
	params["id"] = "123"

	got := ExecuteExp(t, "id:=:123", params)
	assert.True(t, got)

	got = ExecuteExp(t, "id:=:444", params)
	assert.False(t, got)
}

func TestExpNotEq(t *testing.T) {
	params := map[string]string{}
	params["id"] = "123"

	got := ExecuteExp(t, "id:!=:200", params)
	assert.True(t, got)

	got = ExecuteExp(t, "id:!=:123", params)
	assert.False(t, got)
}

func TestExpLt(t *testing.T) {
	params := map[string]string{}
	params["id"] = "123"
	params["type"] = "p40"

	got := ExecuteExp(t, "id:<:200", params)
	assert.True(t, got)

	got = ExecuteExp(t, "id:<:100", params)
	assert.False(t, got)

	got = ExecuteExp(t, "type:<:100", params)
	assert.False(t, got)
}

func TestExpLtEq(t *testing.T) {
	params := map[string]string{}
	params["id"] = "123"
	params["type"] = "p40"

	got := ExecuteExp(t, "id:<=:200", params)
	assert.True(t, got)

	got = ExecuteExp(t, "id:<=:100", params)
	assert.False(t, got)

	got = ExecuteExp(t, "id:<=:123", params)
	assert.True(t, got)

	got = ExecuteExp(t, "type:<=:100", params)
	assert.False(t, got)
}

func TestExpGt(t *testing.T) {
	params := map[string]string{}
	params["id"] = "123"
	params["type"] = "p40"

	got := ExecuteExp(t, "id:>:200", params)
	assert.False(t, got)

	got = ExecuteExp(t, "id:>:100", params)
	assert.True(t, got)

	got = ExecuteExp(t, "type:>:100", params)
	assert.False(t, got)
}

func TestExpGtEq(t *testing.T) {
	params := map[string]string{}
	params["id"] = "123"
	params["type"] = "p40"

	got := ExecuteExp(t, "id:>=:200", params)
	assert.False(t, got)

	got = ExecuteExp(t, "id:>=:100", params)
	assert.True(t, got)

	got = ExecuteExp(t, "id:>=:123", params)
	assert.True(t, got)

	got = ExecuteExp(t, "type:>=:1", params)
	assert.False(t, got)
}

func TestExpIn(t *testing.T) {
	params := map[string]string{}
	params["type"] = "p40"

	got := ExecuteExp(t, "type:in:p40,mate40", params)
	assert.True(t, got)

	got = ExecuteExp(t, "type:in:mate40, p40", params)
	assert.True(t, got)

	got = ExecuteExp(t, "type:in:mate40, p40 , p30", params)
	assert.True(t, got)

	got = ExecuteExp(t, "type:in:mate40,p30", params)
	assert.False(t, got)

	got = ExecuteExp(t, "type:in:", params)
	assert.False(t, got)
}

func TestExpExcept(t *testing.T) {
	params := map[string]string{}
	params["id"] = "123"
	params["type"] = "p40"

	got := ExecuteExp(t, "age:<:30", params)
	assert.False(t, got)

	got = ExecuteExp(t, "id:<:", params)
	assert.False(t, got)

	got = ExecuteExp(t, "id:<:abc", params)
	assert.False(t, got)

	got = ExecuteExp(t, "id:||:123", params)
	assert.False(t, got)
}

func TestConvertStrsToNum(t *testing.T) {
	convey.Convey(
		"Test convertStrsToNum", t, func() {
			convey.Convey(
				"convertStrsToNum success when err", func() {
					n, err := convertStrsToNum([]string{"a"})
					convey.So(n, convey.ShouldBeNil)
					convey.So(err, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"convertStrsToNum success", func() {
					n, err := convertStrsToNum([]string{"1", "2"})
					convey.So(n, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}
