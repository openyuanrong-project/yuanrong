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

import (
	"github.com/vmihailenco/msgpack"

	"yuanrong.org/kernel/runtime/libruntime/config"
)

// NamedInstance provides function to construct *InstanceFunctionHandler
type NamedInstance struct {
	instanceId string
}

// NewNamedInstance return *NamedInstance
func NewNamedInstance(instanceId string) *NamedInstance {
	instance := &NamedInstance{
		instanceId: instanceId,
	}
	return instance
}

// EncodeMsgpack serialize
func (instance NamedInstance) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeMulti(instance.instanceId)
}

// DecodeMsgpack deserialize
func (instance *NamedInstance) DecodeMsgpack(dec *msgpack.Decoder) error {
	return dec.DecodeMulti(&instance.instanceId)
}

// Function return *InstanceFunctionHandler
func (instance *NamedInstance) Function(fn any) *InstanceFunctionHandler {
	return NewInstanceFunctionHandler(fn, instance.instanceId)
}

// Terminate kill instance
func (instance *NamedInstance) Terminate() error {
	return GetRuntimeHolder().GetRuntime().Kill(instance.instanceId, config.KillInstance, make([]byte, 0))
}
