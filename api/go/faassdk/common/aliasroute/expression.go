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
	"strconv"
	"strings"

	"yuanrong.org/kernel/runtime/libruntime/common/faas/logger"
)

const (
	expressionSize = 3
)

// Expression rule expression struct
type Expression struct {
	leftVal  string
	operator string
	rightVal string
}

func compareIntegerStrings(a, b string) (int, error) {
	numA, err := strconv.Atoi(a)
	if err != nil {
		return 0, err
	}

	numB, err := strconv.Atoi(b)
	if err != nil {
		return 0, err
	}

	if numA < numB {
		return -1, nil
	} else if numA > numB {
		return 1, nil
	} else {
		return 0, nil
	}
}

// Execute the rule expression
func (exp *Expression) Execute(params map[string]string) bool {
	logger.GetLogger().Debugf("params %v, exp.leftVal %v,exp.rightVal %v", params, exp.leftVal, exp.rightVal)
	val, exist := params[exp.leftVal]
	if !exist {
		logger.GetLogger().Warnf("cannot find val for %s in params", exp.leftVal)
		return false
	}

	switch exp.operator {
	case "=":
		return strings.TrimSpace(val) == strings.TrimSpace(exp.rightVal)
	case "!=":
		return strings.TrimSpace(val) != strings.TrimSpace(exp.rightVal)
	case ">":
		ret, err := compareIntegerStrings(val, exp.rightVal)
		return err == nil && ret == 1
	case "<":
		ret, err := compareIntegerStrings(val, exp.rightVal)
		return err == nil && ret == -1
	case ">=":
		ret, err := compareIntegerStrings(val, exp.rightVal)
		return err == nil && (ret == 1 || ret == 0)
	case "<=":
		ret, err := compareIntegerStrings(val, exp.rightVal)
		return err == nil && (ret == -1 || ret == 0)
	case "in":
		return matchStr(val, exp.rightVal)
	default:
		logger.GetLogger().Warnf("unknown operator(%s), return false", val, exp.operator)
		return false
	}
}

func convertStrsToNum(strs []string) ([]int, error) {
	var ret []int
	for _, str := range strs {
		num, err := strconv.Atoi(strings.TrimSpace(str))
		if err != nil {
			return nil, err
		}
		ret = append(ret, num)
	}
	return ret, nil
}

func matchStr(str string, targetStr string) bool {
	tars := strings.Split(targetStr, ",")
	for _, tar := range tars {
		// The rvalue of the 'in' operator ignores ""
		if tar != "" && strings.TrimSpace(str) == strings.TrimSpace(tar) {
			return true
		}
	}
	return false
}
