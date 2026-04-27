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

// Package loadbalance provides load balancing algorithm
package loadbalance

import "time"

const (
	// RoundRobinNginx represents type of Round Robin Nginx
	RoundRobinNginx LBType = iota
	// RoundRobinLVS represents type of Round Robin LVS
	RoundRobinLVS
	// ConsistentHashGeneric represents type of Generic Consistent Hash
	ConsistentHashGeneric
	// SimpleHashGeneric represents type of simple Generic Consistent Hash
	SimpleHashGeneric
)

// Request -
type Request struct {
	Name      string
	TraceID   string
	Timestamp time.Time
}

// LBType is the type of load loadbalance algorithm
type LBType int

const defaultCHGenericConcurrency = 100

// LoadBalance is the interface of loadbalance algorithm
type LoadBalance interface {
	Next(name string, move bool) interface{} // move parameter controls whether the hash loop moves
	Add(node interface{}, weight int)
	Remove(node interface{})
	RemoveAll()
	Reset()
	DeleteBalancer(name string)
}

// LBFactory is the factory of loadbalance algorithm
func LBFactory(t LBType) LoadBalance {
	switch t {
	case RoundRobinNginx:
		return &WNGINX{}
	case ConsistentHashGeneric:
		return NewCHGeneric()
	case SimpleHashGeneric:
		return NewSimpleCHGeneric()
	default:
		return NewCHGeneric()
	}
}
