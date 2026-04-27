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

// Package yr for actor
package yr

import "yuanrong.org/kernel/runtime/libruntime/api"

var holder *RuntimeHolder = nil

// RuntimeHolder holds a libRuntimeAPI
type RuntimeHolder struct {
	runtime api.LibruntimeAPI
}

// GetRuntimeHolder return RuntimeHolder
func GetRuntimeHolder() *RuntimeHolder {
	if holder != nil {
		return holder
	}

	holder = &RuntimeHolder{}
	return holder
}

// Init initializes the RuntimeHolder
func (holder *RuntimeHolder) Init(runtime api.LibruntimeAPI) {
	if holder != nil {
		holder.runtime = runtime
	}
}

// GetRuntime return libRuntimeAPI
func (holder *RuntimeHolder) GetRuntime() api.LibruntimeAPI {
	if holder != nil {
		return holder.runtime
	}
	return nil
}
