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

// Package lease -
package lease

import (
	"time"

	"yuanrong.org/kernel/pkg/functionscaler/types"
)

// InstanceLeaseManager manages leases of a specific function
type InstanceLeaseManager interface {
	CreateInstanceLease(insAlloc *types.InstanceAllocation, interval time.Duration, callback func()) (
		types.InstanceLease, error)
	HandleInstanceDelete(instance *types.Instance)
	CleanAllLeases()
}
