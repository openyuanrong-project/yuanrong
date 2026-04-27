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

// Package engine defines interfaces for a storage kv engine
package engine

import (
	"context"
	"errors"
)

var (
	// ErrKeyNotFound means key does not exist
	ErrKeyNotFound = errors.New("key not found")
	// ErrTransaction means a transaction request has failed. User should retry.
	ErrTransaction = errors.New("failed to execute a transaction")
)

// Engine defines a storage kv engine
type Engine interface {
	// Get retrieves the value for a key. It returns ErrKeyNotFound if key does not exist.
	Get(ctx context.Context, key string) (val string, err error)

	// Count retries the number of key value pairs from a prefix key.
	Count(ctx context.Context, prefix string) (int64, error)

	// FirstInRange retries the first key value pairs sort by keys from a prefix. It returns ErrKeyNotFound if the
	// range is empty.
	FirstInRange(ctx context.Context, prefix string) (key, val string, err error)

	// LastInRange is similar to FirstInRange by searches backwards. It returns ErrKeyNotFound if the range is empty.
	LastInRange(ctx context.Context, prefix string) (key, val string, err error)

	// PrepareStream creates a stream that users can combine it with "Filter" and "Execute" to fuzzy search from a
	// range of key value pairs.
	PrepareStream(ctx context.Context, prefix string, decode DecodeHandleFunc, by SortBy) PrepareStmt

	// Put writes a key value pairs.
	Put(ctx context.Context, key string, value string) error

	// Delete removes a key value pair.
	Delete(ctx context.Context, key string) error

	// BeginTx starts a new transaction.
	BeginTx(ctx context.Context) Transaction

	// Close cleans up any resources holds by the engine.
	Close() error
}

// FilterFunc filters a range stream
type FilterFunc func(interface{}) bool

// PrepareStmt allows fuzzy searching of a range of key value pairs.
type PrepareStmt interface {
	Filter(filter FilterFunc) PrepareStmt
	Execute() (Stream, error)
}

// Stream defines interface to range a filtered key value pairs.
type Stream interface {
	// Next returns the next value from a stream. It returns io.EOF when stream ends
	Next() (interface{}, error)
}

// DecodeHandleFunc decodes key value pairs to a user defined type.
type DecodeHandleFunc func(key, value string) (interface{}, error)

// SortOrder is the sorting order of a stream.
type SortOrder int

const (
	// Ascend starts from small to large.
	Ascend SortOrder = iota

	// Descend starts from large to small.
	Descend
)

// SortTarget is the sorting target of a stream.
type SortTarget int

const (
	// SortName tells a stream to sort by name.
	SortName SortTarget = iota

	// SortCreate tells a stream to sort by create time.
	SortCreate

	// SortModify tells a stream to sort by update time.
	SortModify
)

// SortBy is the sorting order and target of a stream.
type SortBy struct {
	Order  SortOrder
	Target SortTarget
}

// Transaction ensures consistent view and change of the data.
type Transaction interface {
	Get(key string) (val string, err error)
	GetPrefix(key string) (keys, values []string, err error)
	Put(key string, value string)
	Del(key string)
	DelPrefix(key string)
	Commit() error
	Cancel()
}
