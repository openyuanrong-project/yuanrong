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

// Package constants implements vars of all
package constants

import "time"

const (
	// DefaultMapSize is the default map size
	DefaultMapSize = 16
	// MonitorFileName monitor file name
	MonitorFileName = "monitor-disk"

	// HostNameEnvKey defines the hostname env key
	HostNameEnvKey = "HOSTNAME"

	// PodNameEnvKey define pod name env key
	PodNameEnvKey = "POD_NAME"

	// MaxMsgSize grpc client max message size(bit)
	MaxMsgSize = 1024 * 1024 * 10
	// MaxWindowSize grpc flow control window size(bit)
	MaxWindowSize = 1024 * 1024 * 10
	// MaxBufferSize grpc read/write buffer size(bit)
	MaxBufferSize = 1024 * 1024 * 10
	// DialBaseDelay -
	DialBaseDelay = 300 * time.Millisecond
	// DialMultiplier -
	DialMultiplier = 1.2
	// DialJitter -
	DialJitter = 0.1
	// DialMaxDelay -
	DialMaxDelay = 15 * time.Second
	// RuntimeDialMaxDelay -
	RuntimeDialMaxDelay = 100 * time.Second
)
