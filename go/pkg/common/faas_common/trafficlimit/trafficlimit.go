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

// Package trafficlimit -
package trafficlimit

import (
	"math"
	"sync"

	"golang.org/x/time/rate"
)

const (
	// DefaultFunctionLimitRate  default function limit rate for traffic limitation
	DefaultFunctionLimitRate = 5000
	// TrafficRedundantRate  limit redundancy rate for traffic limitation
	TrafficRedundantRate = 1.1
	// DefaultAccessorInitCopies Initial number of copies
	DefaultAccessorInitCopies = 3
)

var functionLimitRate int

// LimiterContainer function key and function's Limiter
type LimiterContainer struct {
	funcLimiterMap *sync.Map
}

// FunctionLimiter -
type FunctionLimiter struct {
	Quota   int
	Limiter *rate.Limiter
}

var (
	// FunctionBuf -
	funcLimiterContainer = &LimiterContainer{
		funcLimiterMap: &sync.Map{},
	}
)

// RateLimiter rate limiter struct
type RateLimiter struct {
	*rate.Limiter
}

// Take return if a function request is allowed
func (r *RateLimiter) take() bool {
	return r.Limiter.Allow()
}

// SetFunctionLimitRate -
func SetFunctionLimitRate(limit int) {
	if limit <= 0 {
		limit = DefaultFunctionLimitRate
	}
	functionLimitRate = limit
}

// FuncTrafficLimit is the main function of function traffic limitation
func FuncTrafficLimit(funcKey string) bool {
	return funcLimiterContainer.funcTakeOneToken(funcKey)
}

func (t *LimiterContainer) funcTakeOneToken(funcKey string) bool {
	funcLimiter := t.getFunctionLimiter(funcKey)
	if funcLimiter.Limiter == nil {
		return true
	}
	return funcLimiter.Limiter.Allow()
}

// getFunctionInfo  to generator the function limiter
func (t *LimiterContainer) getFunctionLimiter(functionKey string) FunctionLimiter {
	funcLimiter, ok := t.funcLimiterMap.Load(functionKey)
	if !ok {
		if functionLimitRate <= 0 {
			functionLimitRate = DefaultFunctionLimitRate
		}
		limiter := FunctionLimiter{Limiter: t.getLimiter(functionLimitRate), Quota: DefaultFunctionLimitRate}
		t.funcLimiterMap.Store(functionKey, limiter)
		return limiter
	}
	return funcLimiter.(FunctionLimiter)
}

func (t *LimiterContainer) getLimiter(quota int) *rate.Limiter {
	limitRate := float64(quota) / DefaultAccessorInitCopies
	limitBucketSize := int(math.Ceil(float64(quota)) /
		DefaultAccessorInitCopies * TrafficRedundantRate)
	tenantLimiter := rate.NewLimiter(rate.Limit(limitRate), limitBucketSize)
	return tenantLimiter
}
