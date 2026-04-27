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
	"reflect"

	"github.com/vmihailenco/msgpack"

	"yuanrong.org/kernel/runtime/libruntime/common/uuid"
)

// ObjectRef Reference to a data object
type ObjectRef struct {
	objId   string
	refType reflect.Type
}

// EncodeMsgpack serialize
func (ref ObjectRef) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeMulti(ref.objId)
}

// DecodeMsgpack deserialize
func (ref *ObjectRef) DecodeMsgpack(dec *msgpack.Decoder) error {
	return dec.DecodeMulti(&ref.objId)
}

// NewObjectRef return a *ObjectRef
func NewObjectRef() *ObjectRef {
	objId := "yr-api-obj-" + uuid.New().String()
	o := &ObjectRef{objId: objId}
	return o
}
