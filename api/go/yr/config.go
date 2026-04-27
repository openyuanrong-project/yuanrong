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
	"yuanrong.org/kernel/runtime/libruntime/api"
	"yuanrong.org/kernel/runtime/libruntime/config"
)

// Mode defines the function running mode
type Mode int

const (
	// LocalMode Single-machine multithreading
	LocalMode = iota
	// ClusterMode Multi-machine multiprocessing
	ClusterMode
	// Invalid invalid
	Invalid
)

// Config Used for initializing the YR system
type Config struct {
	Mode                Mode
	FunctionUrn         string
	ServerAddr          string
	DataSystemAddr      string
	DataSystemAgentAddr string
	InCluster           bool
	LogLevel            string
	AutoStart           bool
}

// FlagsConfig from the command line
type FlagsConfig struct {
	RuntimeID            string
	InstanceID           string
	LogLevel             string
	Address              string
	LogPath              string
	JobID                string
	Hooks                config.HookIntfs
	FunctionExectionPool config.Pool
}

// InvokeOptions function invoke option
type InvokeOptions struct {
	Cpu                 int
	Memory              int
	InvokeLabels        map[string]string
	CustomResources     map[string]float64
	CustomExtensions    map[string]string
	RetryTime           uint
	RecoverRetryTimes   int
	CreateNotifyTimeout int
	Labels              []string
	Affinity            map[string]string
	ScheduleAffinities  []api.Affinity
}

// NewInvokeOptions return *InvokeOptions
func NewInvokeOptions() *InvokeOptions {
	return &InvokeOptions{
		Cpu:                 500,
		Memory:              500,
		InvokeLabels:        make(map[string]string),
		CustomResources:     make(map[string]float64),
		CustomExtensions:    make(map[string]string),
		RetryTime:           0,
		CreateNotifyTimeout: -1,
		Labels:              make([]string, 0),
		Affinity:            make(map[string]string),
		ScheduleAffinities:  make([]api.Affinity, 0),
	}
}
