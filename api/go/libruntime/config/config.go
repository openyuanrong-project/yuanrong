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

// Package config for init
package config

import (
	"unsafe"

	"yuanrong.org/kernel/runtime/libruntime/api"
)

// InvokeType invoke type
type InvokeType int32

const (
	// CreateInstance create actor
	CreateInstance InvokeType = 0
	// InvokeInstance invoke actor
	InvokeInstance InvokeType = 1
	// CreateInstanceStateless create task
	CreateInstanceStateless InvokeType = 2
	// InvokeInstanceStateless invoke task
	InvokeInstanceStateless InvokeType = 3
)

const (
	// DefaultSignal default signal
	DefaultSignal int = 0
	// KillInstance kill single signal
	KillInstance int = 1
	// KillAllInstances kill all signal
	KillAllInstances int = 2
	// Exit exit signal
	Exit int = 64
	// Cancel cancel signal
	Cancel int = 65
)

// DataBuffer save the data pointer
type DataBuffer struct {
	Ptr             unsafe.Pointer
	Size            int
	SharedPtrBuffer unsafe.Pointer
}

// DataObject data obj
type DataObject struct {
	ID              string
	Buffer          DataBuffer
	NestedObjectIds []string
	CSharedPtr      unsafe.Pointer
}

// LoadFunctionCallback load function callback
type LoadFunctionCallback func([]string) error

// FunctionExecuteCallback function execute callback
type FunctionExecuteCallback func(api.FunctionMeta, InvokeType, []api.Arg, []DataObject) error

// CheckpointCallback checkpoint callback
type CheckpointCallback func(string) ([]byte, error)

// RecoverCallback recover callback
type RecoverCallback func([]byte) error

// ShutdownCallback shutdown callback
type ShutdownCallback func(uint64) error

// SignalCallback signal callback
type SignalCallback func(int, []byte) error

// HealthCheckCallback health check callback
type HealthCheckCallback func() (api.HealthType, error)

// HookIntfs hook struct
type HookIntfs struct {
	LoadFunctionCb      LoadFunctionCallback
	FunctionExecutionCb FunctionExecuteCallback
	CheckpointCb        CheckpointCallback
	RecoverCb           RecoverCallback
	ShutdownCb          ShutdownCallback
	SignalCb            SignalCallback
	HealthCheckCb       HealthCheckCallback
}

// Pool execution pool abstract interface
type Pool interface {
	Submit(task func()) error
}

// Config init config
type Config struct {
	FunctionSystemAddress        string
	FunctionSystemRtServerIPAddr string
	FunctionSystemRtServerPort   int
	GrpcAddress                  string
	DataSystemAddress            string
	DataSystemIPAddr             string
	DataSystemPort               int
	IamAddress                   string
	IsDriver                     bool
	JobID                        string
	RuntimeID                    string
	InstanceID                   string
	FunctionName                 string
	LogLevel                     string
	LogDir                       string
	LogFileSizeMax               uint32
	LogFileNumMax                uint32
	LogFlushInterval             int
	Hooks                        HookIntfs
	FunctionExectionPool         Pool
	RecycleTime                  int
	MaxTaskInstanceNum           int
	MaxConcurrencyCreateNum      int
	EnableSigaction              bool
	EnableMetrics                bool
	ThreadPoolSize               uint32
	LoadPaths                    []string
	EnableMTLS                   bool
	PrivateKeyPath               string
	CertificateFilePath          string
	VerifyFilePath               string
	PrivateKeyPaaswd             string
	HttpIocThreadsNum            uint32
	ServerName                   string
	InCluster                    bool
	Namespace                    string
	Api                          api.ApiType
	FunctionId                   string
	SystemAuthAccessKey          string
	SystemAuthSecretKey          string
	SystemAuthDataKey            string
	// EnableCallStack whether to enable distribute call stack
	EnableCallStack              bool
	CallStackLayerNum            int
	EncryptPrivateKeyPasswd      string
	PrimaryKeyStoreFile          string
	StandbyKeyStoreFile          string
	EnableDsEncrypt              bool
	RuntimePublicKeyContextPath  string
	RuntimePrivateKeyContextPath string
	DsPublicKeyContextPath       string
	EnableEvent                  bool
}
