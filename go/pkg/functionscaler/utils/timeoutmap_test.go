/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
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

// Package utils -
package utils

import (
	"testing"
	"time"
)

func TestNewTimeoutMap(t *testing.T) {
	tm := NewTimeoutMap(time.Second)
	if tm == nil {
		t.Fatal("NewTimeoutMap returned nil")
	}
}

func TestSetAndGet(t *testing.T) {
	tm := NewTimeoutMap(time.Minute)

	// 测试正常设置和获取
	tm.Set("key1", "value1", 5*time.Second)

	val, ok := tm.Get("key1")
	if !ok {
		t.Error("Expected to get key1, but not found")
	}
	if val != "value1" {
		t.Errorf("Expected value 'value1', got '%v'", val)
	}
}

func TestGetExpired(t *testing.T) {
	tm := NewTimeoutMap(100 * time.Millisecond)

	// 设置一个立即过期的键
	tm.Set("key1", "value1", 1*time.Millisecond)

	// 等待过期
	time.Sleep(10 * time.Millisecond)

	// 应该获取不到
	val, ok := tm.Get("key1")
	if ok {
		t.Errorf("Expected key1 to be expired, but got value: %v", val)
	}
}

func TestDelete(t *testing.T) {
	tm := NewTimeoutMap(time.Minute)

	tm.Set("key1", "value1", 5*time.Second)

	// 删除前应该存在
	val, ok := tm.Get("key1")
	if !ok || val != "value1" {
		t.Error("Key should exist before deletion")
	}

	// 删除
	tm.Delete("key1")

	// 删除后应该不存在
	val, ok = tm.Get("key1")
	if ok {
		t.Errorf("Expected key1 to be deleted, but got value: %v", val)
	}
}

func TestCleanUp(t *testing.T) {
	tm := NewTimeoutMap(10 * time.Millisecond)

	tm.Set("key1", "value1", time.Millisecond)

	// 删除前应该存在
	val, ok := tm.Get("key1")
	if !ok || val != "value1" {
		t.Error("Key should exist before deletion")
	}
	time.Sleep(10 * time.Millisecond)

	// 删除后应该不存在
	val, ok = tm.Get("key1")
	if ok {
		t.Errorf("Expected key1 to be deleted, but got value: %v", val)
	}
}
