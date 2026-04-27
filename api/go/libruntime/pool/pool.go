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

// Package pool for goroutine pool
package pool

import "github.com/panjf2000/ants/v2"

// DefaultFuncExecPoolSize -
const DefaultFuncExecPoolSize = 300000

// Pool implements a goroutine pool for execution of function calls
type Pool struct {
	p *ants.Pool
}

// NewPool implements a constructor for `Pool`
func NewPool(size int) *Pool {
	tmp := Pool{}
	p, _ := ants.NewPool(size)
	tmp.p = p
	return &tmp
}

// Submit implements a submitting to pool for function call task
func (p *Pool) Submit(task func()) error {
	return p.p.Submit(task)
}
