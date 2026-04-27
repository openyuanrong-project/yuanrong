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

// Package queue -
package queue

import "errors"

const (
	// keyNotFoundIndex stands for index of a non exist key
	keyNotFoundIndex = -1
)

var (
	// ErrObjectNotFound is the error of object not found
	ErrObjectNotFound = errors.New("object not found")
	// ErrMethodUnsupported is the error of method unsupported
	ErrMethodUnsupported = errors.New("method unsupported")
)

// IdentityFunc will get ID from object in queue
type IdentityFunc func(interface{}) string

// Queue is interface of queue used in faas pattern
type Queue interface {
	Front() interface{}
	Back() interface{}
	PopFront() interface{}
	PopBack() interface{}
	PushBack(obj interface{}) error
	GetByID(objID string) interface{}
	DelByID(objID string) error
	UpdateObjByID(objID string, obj interface{}) error
	Len() int
	Range(f func(obj interface{}) bool) bool
	SortedRange(f func(obj interface{}) bool) bool
}
