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

// Package api for libruntime
package api

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ProducerConf represents configuration information for a producer.
type ProducerConf struct {
	DelayFlushTime int64  // Trigger Flush after the maximum Delay duration after Send.
	PageSize       int64  // Size of the buffer page corresponding to the producer.
	MaxStreamSize  uint64 // Maximum size of the shared memory that can be used by a stream on a worker.
	TraceId        string // trace id of producer config
}

// SubscriptionType subscribe type
type SubscriptionType int

const (
	// Stream Single consumer consumption
	Stream SubscriptionType = iota
	// RoundRobin Multi-consumer load-balanced consumption round robin
	RoundRobin
	// KeyPartitions Multi-consumer load-balanced consumption by key partition
	KeyPartitions
)

// SubscriptionConfig represents configuration information for a consumer.
type SubscriptionConfig struct {
	SubscriptionName string
	SubscriptionType SubscriptionType
	TraceId          string
}

// Element data to be sent
type Element struct {
	Ptr  *uint8
	Size uint64
	Id   uint64
}

// StreamProducer producer interface
type StreamProducer interface {
	Send(element Element) error
	SendWithTimeout(element Element, timeoutMs int64) error
	Close() error
}

// StreamConsumer consumer interface
type StreamConsumer interface {
	ReceiveExpectNum(expectNum uint32, timeoutMs uint32) ([]Element, error)
	Receive(timeoutMs uint32) ([]Element, error)
	Ack(elementId uint64) error
	Close() error
}

// ApiType api type, for example, actor
type ApiType int32

const (
	// ActorApi actor api
	ActorApi ApiType = 0
	// FaaSApi faas api
	FaaSApi ApiType = 1
	// PosixApi posix api
	PosixApi ApiType = 2
	// ServeApi posix api
	ServeApi ApiType = 3
)

// LanguageType language type, for example, cpp
type LanguageType int32

const (
	// Cpp cpp language
	Cpp LanguageType = iota
	// Python python language
	Python
	// Java java language
	Java
	// Golang go language
	Golang
	// NodeJS node js language
	NodeJS
	// CSharp c# language
	CSharp
	// Php php language
	Php
)

// FunctionMeta function meta
type FunctionMeta struct {
	AppName   string
	FuncName  string
	FuncID    string
	Sig       string
	PoolLabel string
	Name      *string
	Namespace *string
	Api       ApiType
	Language  LanguageType
}

// StackTracesInfo stack info
type StackTracesInfo struct {
	Code        int
	MCode       int
	Message     string
	StackTraces []StackTrace
}

// StackTrace stack trace
type StackTrace struct {
	ClassName     string
	MethodName    string
	FileName      string
	LineNumber    int64
	ExtensionInfo map[string]string
}

// ArgType value or objectRef
type ArgType int32

const (
	// Value value of object
	Value ArgType = 0
	// ObjectRef ref of object
	ObjectRef ArgType = 1
)

// Arg function arg
type Arg struct {
	Type            ArgType
	Data            []byte
	ObjectID        string
	TenantID        string
	NestedObjectIDs []string
}

// OperatorType operator type
type OperatorType int32

const (
	// LabelOpIn in
	LabelOpIn OperatorType = 0
	// LabelOpNotIn not in
	LabelOpNotIn OperatorType = 1
	// LabelOpExists exists
	LabelOpExists OperatorType = 2
	// LabelOpNotExists not exists
	LabelOpNotExists OperatorType = 3
)

// LabelOperator affinity label operator
type LabelOperator struct {
	Type        OperatorType
	LabelKey    string
	LabelValues []string
}

// AffinityKindType affinity type
type AffinityKindType int32

const (
	// AffinityKindResource resource
	AffinityKindResource AffinityKindType = 0
	// AffinityKindInstance instance
	AffinityKindInstance AffinityKindType = 1
)

// AffinityType affinity type
type AffinityType int32

const (
	// PreferredAffinity prefer
	PreferredAffinity AffinityType = 0
	// PreferredAntiAffinity prefer anti
	PreferredAntiAffinity AffinityType = 1
	// RequiredAffinity required
	RequiredAffinity AffinityType = 2
	// RequiredAntiAffinity required anti
	RequiredAntiAffinity AffinityType = 3
)

// Affinity -
type Affinity struct {
	Kind                     AffinityKindType
	Affinity                 AffinityType
	PreferredPriority        bool
	PreferredAntiOtherLabels bool
	LabelOps                 []LabelOperator
}

// InstanceSessionConfig contains session config for instance
type InstanceSessionConfig struct {
	SessionID   string `json:"sessionID"`
	SessionTTL  int    `json:"sessionTTL"`
	Concurrency int    `json:"concurrency"`
}

// InvokeOptions invoke option
type InvokeOptions struct {
	Cpu                  int
	Memory               int
	InvokeLabels         map[string]string
	CustomResources      map[string]float64
	CustomExtensions     map[string]string
	CreateOpt            map[string]string
	InstanceSession      *InstanceSessionConfig
	Labels               []string
	Affinities           map[string]string // deprecated
	ScheduleAffinities   []Affinity
	ScheduleTimeoutMs    int64
	RetryTimes           int
	RecoverRetryTimes    int
	Priority             int
	CodePaths            []string
	SchedulerFunctionID  string
	SchedulerInstanceIDs []string
	TraceID              string
	Timeout              int
	AcquireTimeout       int
	TrafficLimited       bool
}

// InstanceAllocation holds the allocation of instance
type InstanceAllocation struct {
	FuncKey       string
	FuncSig       string
	InstanceID    string
	LeaseID       string
	LeaseInterval int64
}

// WriteModeEnum kv write mode
type WriteModeEnum int

// constants of WriteModeEnum
const (
	NoneL2Cache WriteModeEnum = iota
	WriteThroughL2Cache
	WriteBackL2Cache
	NoneL2CacheEvict
)

// CacheTypeEnum kv cache type
type CacheTypeEnum int

// constants of CacheTypeEnum
const (
	MEMORY CacheTypeEnum = iota
	DISK
)

// SetParam structure is used to transfer parameters during the set operation.
type SetParam struct {
	WriteMode WriteModeEnum
	TTLSecond uint32
	Existence int32
	CacheType CacheTypeEnum
}

// MSetParam structure is used to transfer parameters during the mSetTx operation.
type MSetParam struct {
	WriteMode WriteModeEnum
	TTLSecond uint32
	Existence int32
	CacheType CacheTypeEnum
}

// ConsistencyTypeEnum is the new type defined.
// Use const Param and Causal together to simulate the enumeration type of C++.
type ConsistencyTypeEnum int

// The constant Pram and Causal is of the ConsistencyTypeEnum type.
// They is used as two enumeration constants of ConsistencyTypeEnum to simulates the enumeration type of C++.
const (
	PRAM ConsistencyTypeEnum = iota
	CAUSAL
)

// PutParam structure is used to transfer parameters during the Put operation.
// It takes parameters including WriteMode and ConsistencyType.
// The WriteMode is a WriteModeEnum "enumeration" type,
// The ConsistencyType is a ConsistencyTypeEnum "enumeration" type,
type PutParam struct {
	WriteMode       WriteModeEnum
	ConsistencyType ConsistencyTypeEnum
	CacheType       CacheTypeEnum
}

// HealthType -
type HealthType int32

// HealthCheckRequest -
type HealthCheckRequest struct {
}

// HealthCheckResponse -
type HealthCheckResponse struct {
	Code HealthType
}

// constants of HealthType
const (
	Healthy HealthType = iota
	HealthCheckFailed
	SubHealth
)

// Credential -
type Credential struct {
	AccessKey string
	SecretKey []byte
	DataKey   []byte
}

// ErrorInfo error info
type ErrorInfo struct {
	Code            int
	Err             error
	StackTracesInfo StackTracesInfo
}

// IsOk -
func (e ErrorInfo) IsOk() bool {
	return e.Code == Ok
}

// IsError -
func (e ErrorInfo) IsError() bool {
	return !e.IsOk()
}

func (e ErrorInfo) Error() string {
	if len(e.StackTracesInfo.Message) > 0 {
		errMsg := e.StackTracesInfo.Message
		stackTraces := e.getStackTracesInfo()
		return fmt.Sprintf("%s\n%s", errMsg, stackTraces)
	}
	return e.Err.Error()
}

// AddStack add stack to ErrorInfo
func AddStack(err error, stack StackTrace) error {
	// if stack is empty, it is not appended to the stack.
	if stack.ClassName == "" {
		return err
	}
	errInfo := TurnErrInfo(err)
	// if errInfo.StackTracesInfo.StackTraces is not empty, this stack has been collected
	if len(errInfo.StackTracesInfo.StackTraces) == 0 {
		errInfo.StackTracesInfo.StackTraces = []StackTrace{stack}
	}
	return errInfo
}

// NewErrorInfoWithStackInfo add StackInfo to ErrorInfo
func NewErrorInfoWithStackInfo(err error, stacks []StackTrace) error {
	if len(stacks) == 0 {
		return err
	}
	errInfo := TurnErrInfo(err)
	errInfo.StackTracesInfo.StackTraces = stacks
	return errInfo
}

// TurnErrInfo turn err To ErrorInfo
func TurnErrInfo(err error) ErrorInfo {
	errInfo, ok := err.(ErrorInfo)
	if ok {
		return errInfo
	}
	errInfo = ErrorInfo{
		Err: err,
		StackTracesInfo: StackTracesInfo{
			Message: err.Error(),
		},
	}
	return errInfo
}

// NewErrInfo new ErrorInfo
func NewErrInfo(code int, message string, stackTracesInfo StackTracesInfo) ErrorInfo {
	return ErrorInfo{
		Code:            code,
		Err:             fmt.Errorf(message),
		StackTracesInfo: stackTracesInfo,
	}
}

func (e ErrorInfo) getStackTracesInfo() string {
	var info string
	for _, v := range e.StackTracesInfo.StackTraces {
		var funcInfo, fileInfo, parameters, offset string
		if v.ExtensionInfo != nil {
			parameters = v.ExtensionInfo["parameters"]
			offset = v.ExtensionInfo["offset"]
		}
		funcInfo = fmt.Sprintf("%s.%s%s\n", v.ClassName, v.MethodName, parameters)
		if v.LineNumber == 0 {
			fileInfo = fmt.Sprintf("   %s\n", v.FileName)
		} else {
			fileInfo = fmt.Sprintf("   %s:%d %s\n", v.FileName, v.LineNumber, offset)
		}
		info = info + funcInfo + fileInfo
	}
	return info
}

// GetAsyncCallback define the get async callback function type.
type GetAsyncCallback func(result []byte, err error)

// GetEventCallback define the get event callback function type.
type GetEventCallback func(result []byte, err error)

// FormatLogger format logger interface
type FormatLogger interface {
	With(fields ...zapcore.Field) FormatLogger

	Infof(format string, paras ...interface{})
	Errorf(format string, paras ...interface{})
	Warnf(format string, paras ...interface{})
	Debugf(format string, paras ...interface{})
	Fatalf(format string, paras ...interface{})

	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)

	Sync()
}

const (
	// Ok indicates the operation was successful.
	Ok = 0

	// InvalidParam indicates that an invalid parameter was provided.
	InvalidParam = 2

	// DsClientNilError indicates that dsclient is destructed.
	DsClientNilError = 11001
)

// ConnectArguments -
type ConnectArguments struct {
	Host                      string
	Port                      int
	TimeoutMs                 int
	Token                     []byte
	ClientPublicKey           string
	ClientPrivateKey          []byte
	ServerPublicKey           string
	AccessKey                 string
	SecretKey                 []byte
	AuthclientID              string
	AuthclientSecret          []byte
	AuthURL                   string
	TenantID                  string
	EnableCrossNodeConnection bool
}
