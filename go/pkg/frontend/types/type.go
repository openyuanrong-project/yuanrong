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

// Package types -
package types

import (
	"encoding/json"
	"time"

	"yuanrong.org/kernel/pkg/common/faas_common/alarm"
	"yuanrong.org/kernel/pkg/common/faas_common/crypto"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/localauth"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/config"
	"yuanrong.org/kernel/pkg/common/faas_common/redisclient"
	"yuanrong.org/kernel/pkg/common/faas_common/sts/raw"
	"yuanrong.org/kernel/pkg/common/faas_common/tls"
	"yuanrong.org/kernel/pkg/common/faas_common/types"
	wisecloudTypes "yuanrong.org/kernel/pkg/common/faas_common/wisecloudtool/types"
)

// FunctionRequestInfo function response info
type FunctionRequestInfo struct {
	URN        string `json:"Frn"`
	BusinessID string `json:"BusinessId"`
	TenantID   string `json:"TenantId"`
	Name       string `json:"FuncName"`
	Version    string `json:"FuncVersion"`
	TraceID    string `json:"TraceId"`
	Alias      string `json:"Alias"`
	AppID      string `json:"AppID"`
	StateKey   string `json:"StateKey"`
	NodeLabel  string `json:"NodeLabel"`
	FutureID   string `json:"-"`
}

// InvokeErrorResponse invoke error response
type InvokeErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ResourceSpecification contains resource specification of a requested instance
type ResourceSpecification struct {
	CPU            int64            `json:"cpu"`
	Memory         int64            `json:"memory"`
	CustomResource map[string]int64 `json:"customResource"`
}

// CallReq is the msg structure sent from the frontend to the executor
type CallReq struct {
	Header map[string]string `json:"header"`
	Path   string            `json:"path"`
	Method string            `json:"method"`
	Query  string            `json:"query"`
	Body   json.RawMessage   `json:"body"`
}

// CallResp is the msg structure returned by the executor to the frontend
type CallResp struct {
	Headers         map[string]string `json:"headers"`
	BillingDuration string            `json:"billingDuration"`
	InnerCode       string            `json:"innerCode"`
	InvokeSummary   string            `json:"invokeSummary"`
	LogResult       string            `json:"logResult"`
	UserFuncTime    float64           `json:"userFuncTime"`
	ExecutorTime    float64           `json:"executorTime"`
	Body            json.RawMessage   `json:"body"`
}

// InitResp -
type InitResp struct {
	ErrorCode string          `json:"errorCode"`
	Message   json.RawMessage `json:"message"`
}

// Config is the config used by faas frontend function
type Config struct {
	InstanceNum             int                        `json:"instanceNum"`
	CPU                     float64                    `json:"cpu" valid:"optional"`
	Memory                  float64                    `json:"memory" valid:"optional"`
	SLAQuota                int                        `json:"slaQuota" valid:"optional"`
	Runtime                 RuntimeConfig              `json:"runtime" valid:"optional"`
	LocalAuth               *localauth.AuthConfig      `json:"localAuth"`
	MetaEtcd                etcd3.EtcdConfig           `json:"metaEtcd" valid:"required"`
	DataSystemEtcd          etcd3.EtcdConfig           `json:"dataSystemEtcd" valid:"optional"`
	CAEMetaEtcd             etcd3.EtcdConfig           `json:"caeMetaEtcd" valid:"optional"`
	RouterEtcd              etcd3.EtcdConfig           `json:"routerEtcd" valid:"required"`
	RedisConfig             RedisConfig                `json:"redisConfig" valid:"optional"`
	HTTPConfig              *FrontendHTTP              `json:"http" valid:"optional"`
	HTTPSConfig             *tls.InternalHTTPSConfig   `json:"httpsConfig" valid:"optional"`
	DataSystemConfig        *types.DataSystemConfig    `json:"dataSystemConfig" valid:"optional"`
	StreamEnable            bool                       `json:"streamEnable"  valid:"optional"`
	StateDisable            bool                       `json:"stateDisable"  valid:"optional"`
	BusinessType            int                        `json:"businessType"`
	FunctionInvokeBackend   int                        `json:"functionInvokeBackend" valid:"optional"`
	SccConfig               crypto.SccConfig           `json:"sccConfig" valid:"optional"`
	Image                   string                     `json:"image" valid:"optional"`
	SchedulerKeyPrefixType  string                     `json:"schedulerKeyPrefixType" valid:"optional"`
	MemoryControlConfig     *types.MemoryControlConfig `json:"memoryControlConfig" valid:"optional"`
	MemoryEvaluatorConfig   *MemoryEvaluatorConfig     `json:"memoryEvaluatorConfig" valid:"optional"`
	DefaultTenantLimitQuota int                        `json:"defaultTenantLimitQuota" valid:"optional"`
	// frontend pool
	DynamicPoolEnable bool `json:"dynamicPoolEnable" valid:"optional"`
	// CaaS config
	AuthenticationEnable bool                `json:"authenticationEnable" valid:"optional"`
	RawStsConfig         raw.StsConfig       `json:"rawStsConfig,omitempty"`
	TrafficLimitParams   *TrafficLimitParams `json:"trafficLimitParams" valid:"optional"`
	NodeSelector         map[string]string   `json:"nodeSelector,omitempty"`
	AzID                 string              `json:"azID" valid:"optional"`
	ClusterID            string              `json:"clusterID" valid:"optional"`
	ClusterName          string              `json:"clusterName" valid:"optional"`
	AlarmConfig          alarm.Config        `json:"alarmConfig" valid:"optional"`
	Version              string              `json:"version" valid:"optional"`
	// FunctionGraph config
	FunctionNameSeparator   string           `json:"functionNameSeparator" valid:"optional"`
	AlarmServerAddress      string           `json:"alarmServerAddress" valid:"optional"`
	InvokeMaxRetryTimes     int              `json:"invokeMaxRetryTimes" valid:"optional"`
	EtcdLeaseConfig         *EtcdLeaseConfig `json:"etcdLeaseConfig" valid:"optional"`
	HeartbeatConfig         *HeartbeatConfig `json:"heartbeatConfig" valid:"optional"`
	E2EMaxDelayTime         int64            `json:"e2eMaxDelayTime" valid:"optional"`
	RetryConfig             *RetryConfig     `json:"retry" valid:"optional"`
	ShareKeys               ShareKeys        `json:"shareKeys" valid:"optional"`
	Affinity                string           `json:"affinity"`
	RPCClientConcurrentNum  int              `json:"rpcClientConcurrentNum" valid:"optional"`
	NodeAffinity            string           `json:"nodeAffinity" valid:"optional"`
	NodeAffinityPolicy      string           `json:"nodeAffinityPolicy" valid:"optional"`
	AuthConfig              AuthConfig       `json:"authConfig" valid:"optional"`
	VerifyFilePath          string           `json:"verifyFilePath" valid:"optional"`
	WiseCloudConfig         WiseCloudConfig  `json:"wiseCloudConfig" valid:"optional"`
	IamConfig               IamConfig        `json:"iamConfig" valid:"optional"`
	EnableEvent             bool             `json:"enableEvent" valid:"optional"`
	WatchedConfigFilePath   string           `json:"watchedConfigFilePath" valid:"optional"`
	AccessFaaSSchedulerType string           `json:"accessFaaSSchedulerType" valid:"optional"`
}

// IamConfig -
type IamConfig struct {
	Addr string `json:"addr"`
}

// WiseCloudConfig -
type WiseCloudConfig struct {
	ServiceAccountJwt wisecloudTypes.ServiceAccountJwt `json:"serviceAccountJwt" valid:"optional"`
}

// RetryConfig define retry config
type RetryConfig struct {
	InstanceExceptionRetry bool `json:"instanceExceptionRetry" valid:"optional"`
}

// RedisConfig redis config
type RedisConfig struct {
	ClusterID   string                  `json:"clusterID,omitempty" valid:",optional"`
	ServerAddr  string                  `json:"serverAddr,omitempty" valid:",optional"`
	ServerMode  string                  `json:"serverMode,omitempty" valid:",optional"`
	Password    string                  `json:"password,omitempty" valid:",optional"`
	EnableTLS   bool                    `json:"enableTLS,omitempty" valid:",optional"`
	TimeoutConf redisclient.TimeoutConf `json:"timeoutConf,omitempty" valid:",optional"`
}

// MemoryEvaluatorConfig memory evaluator config
type MemoryEvaluatorConfig struct {
	RequestMemoryEvaluator float64 `json:"requestMemoryEvaluator" valid:",optional"`
}

// ShareKeys -
type ShareKeys struct {
	AccessKey string `json:"accessKey" valid:"optional"`
}

// RuntimeConfig config info
type RuntimeConfig struct {
	Port             string `json:"port" valid:",optional"`
	AvailableZoneKey string `json:"azkey,omitempty" valid:",optional"`

	// SDK
	LogConfig        config.CoreInfo  `json:"logConfig" valid:"optional"`
	SystemAuthConfig SystemAuthConfig `json:"systemAuthConfig" valid:"optional"`
	EnableSigaction  bool             `json:"enableSigaction" valid:"optional"`
}

// FrontendHTTP Used to configure the ResponseTimeout
type FrontendHTTP struct {
	RespTimeOut               int64 `json:"resptimeout"  valid:",optional"`
	WorkerInstanceReadTimeOut int64 `json:"workerInstanceReadTimeOut"  valid:",optional"`
	// MaxRequestBodySize unit is M
	MaxRequestBodySize int `json:"maxRequestBodySize" valid:"required"`
	// MaxStreamRequestBodySize unit is M
	MaxStreamRequestBodySize int `json:"maxStreamRequestBodySize" valid:"optional"`
	// ServerReadTimeout unit is S
	ServerReadTimeout int `json:"serverReadTimeout" valid:"optional"`
	// ServerWriteTimeout unit is S
	ServerWriteTimeout int `json:"serverWriteTimeout" valid:"optional"`
	// ClientIdleTimeout unit is S
	ClientIdleTimeout int `json:"clientIdleTimeout" valid:"optional"`
	// MaxDataSystemMultiDataBodySize unit is M
	MaxDataSystemMultiDataBodySize int    `json:"maxDataSystemMultiDataBodySize" valid:"optional"`
	ServerListenPort               int    `json:"serverListenPort" valid:"optional"`
	ServerListenIP                 string `json:"serverListenIP" valid:"optional"`
}

// TrafficLimitParams parameters of traffic limitation
type TrafficLimitParams struct {
	InstanceLimitRate  float64 `json:"instanceLimitRate" valid:",optional"`
	InstanceBucketSize int     `json:"instanceBucketSize" valid:",optional"`
	FuncLimitRate      float64 `json:"funcLimitRate" valid:",optional"`
	FuncBucketSize     int     `json:"funcBucketSize" valid:",optional"`
}

// StreamContext -
type StreamContext struct {
	StreamName string
	TimeoutMs  uint32
	ExpectNum  int32
}

// InvokeProcessContext -
type InvokeProcessContext struct {
	// func basic info
	TraceID                string
	RequestID              string
	FuncKey                string
	ShouldRetry            bool
	TrafficLimited         bool
	StartTime              time.Time
	RequestTraceInfo       *RequestTraceInfo
	IsHTTPUploadStream     bool
	StreamInfo             *StreamInvokeInfo
	AcquireTimeout         int64
	InvokeTimeout          int64
	InvokeWithoutScheduler bool

	// request info
	ReqHeader map[string]string
	ReqPath   string
	ReqMethod string
	ReqQuery  string
	ReqBody   []byte
	// response info
	StatusCode int
	RespHeader map[string]string
	RespBody   []byte

	// 响应透传
	NeedReadRespHeader bool

	// stream
	StreamCtx *StreamContext
	types.ResponseWriter
}

// CreateInvokeProcessContext -
func CreateInvokeProcessContext() *InvokeProcessContext {
	return &InvokeProcessContext{
		ReqHeader:  make(map[string]string),
		RespHeader: make(map[string]string),
		StartTime:  time.Now(),
	}
}

// AuthConfig -
type AuthConfig struct {
	LocalAuthConfig LocalAuthConfig `json:"localAuthConfig"`
}

// PolicyConfig -
type PolicyConfig struct {
	Allow string `json:"allow"`
	Deny  string `json:"deny"`
}

// LocalAuthConfig -
type LocalAuthConfig struct {
	LocalAuthCryptoPath string `json:"localAuthCryptoPath"`
}

// SystemAuthConfig -
type SystemAuthConfig struct {
	Enable    bool   `json:"enable" validate:"optional"`
	AccessKey string `json:"accessKey" validate:"optional"`
	SecretKey string `json:"secretKey" validate:"optional"`
	DataKey   string `json:"dataKey" validate:"optional"`
}

// APIGTriggerResponse extern interface of web response
type APIGTriggerResponse struct {
	Body            string              `json:"body"`
	Headers         map[string][]string `json:"headers"`
	StatusCode      int                 `json:"statusCode"`
	IsBase64Encoded bool                `json:"isBase64Encoded"`
}

// APIGTriggerEvent extern interface of web request
type APIGTriggerEvent struct {
	IsBase64Encoded       bool                   `json:"isBase64Encoded"`
	HTTPMethod            string                 `json:"httpMethod"`
	Path                  string                 `json:"path"`
	Body                  string                 `json:"body"`
	PathParameters        map[string]string      `json:"pathParameters"`
	RequestContext        APIGRequestContext     `json:"requestContext"`
	Headers               map[string]interface{} `json:"headers"`
	QueryStringParameters map[string]interface{} `json:"queryStringParameters"`
	UserData              string                 `json:"user_data"`
}

// APIGRequestContext -
type APIGRequestContext struct {
	APIID     string `json:"apiId"`
	RequestID string `json:"requestId"`
	Stage     string `json:"stage"`
	SourceIP  string `json:"sourceIp"`
}

// EtcdLeaseConfig etcd lease config
type EtcdLeaseConfig struct {
	LeaseTTL int64 `yaml:"leaseTTL" valid:"optional"`
	RenewTTL int64 `yaml:"renewTTL" valid:"optional"`
}

// HeartbeatConfig heartbeat config
type HeartbeatConfig struct {
	HeartbeatTimeout          int `json:"heartbeatTimeout"  valid:",optional"`
	HeartbeatInterval         int `json:"heartbeatInterval" valid:"optional"`
	HeartbeatTimeoutThreshold int `json:"heartbeatTimeoutThreshold" valid:"optional"`
}

// RequestTraceInfo -
type RequestTraceInfo struct {
	URN          string
	BusinessID   string
	TenantID     string
	FuncName     string
	Version      string
	AnonymizeURN string
	TryCount     int
	InnerCode    int
	AllBusCost   time.Duration
	LastBusCost  time.Duration
	Deadline     time.Time
	CallInstance string
	CallNode     string
	TotalCost    time.Duration
	FrontendCost time.Duration
	BusCost      time.Duration
	WorkerCost   time.Duration
}
