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

// Package registry -
package registry

import (
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/system_function_controller/types"
)

const (
	defaultMapSize      = 16
	functionRegistryNum = 2
)

// FunctionRegistry defines the interface for kind of registry
type FunctionRegistry interface {
	AddSubscriberChan(subChan chan types.SubEvent)
	publishEvent(eventType types.EventType, schedulerSpec *types.InstanceSpecification)
	watcherFilter(event *etcd3.Event) bool
	watcherHandler(event *etcd3.Event)
	InitWatcher()
	RunWatcher()
	getFunctionSpec(instanceID string) *types.InstanceSpecification
}

var (
	// GlobalRegistry is the global registry
	GlobalRegistry *Registry
)

// Registry watches etcd and builds registry cache based on etcd watch
type Registry struct {
	functionRegistry map[types.EventKind]FunctionRegistry
}

// InitRegistry will initialize registry
func InitRegistry() {
	GlobalRegistry = &Registry{
		functionRegistry: make(map[types.EventKind]FunctionRegistry, functionRegistryNum),
	}
}

// AddFunctionRegistry add function registry
func (r *Registry) AddFunctionRegistry(registry FunctionRegistry, eventType types.EventKind) {
	if registry == nil || GlobalRegistry == nil {
		return
	}
	GlobalRegistry.functionRegistry[eventType] = registry
}

// ProcessETCDList ETCD List
func (r *Registry) ProcessETCDList() {
	for _, funcRegistry := range r.functionRegistry {
		funcRegistry.InitWatcher()
	}
}

// RegisterSubscriberChan function registry can subscriber channel
func (r *Registry) RegisterSubscriberChan(subChan chan types.SubEvent) {
	for _, funcRegistry := range r.functionRegistry {
		funcRegistry.AddSubscriberChan(subChan)
	}
	log.GetLogger().Infof("all registry subscriber channel")
}

// RunFunctionWatcher watch events and publish to subscribed channel
func (r *Registry) RunFunctionWatcher() {
	for _, funcRegistry := range r.functionRegistry {
		funcRegistry.RunWatcher()
	}
	log.GetLogger().Infof("all registry run watcher")
}
