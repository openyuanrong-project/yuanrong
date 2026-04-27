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
	"context"
	"sync"
	"time"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/alarm"
	"yuanrong.org/kernel/pkg/common/faas_common/config"
	"yuanrong.org/kernel/pkg/common/faas_common/crypto"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/sts/raw"
	"yuanrong.org/kernel/pkg/common/faas_common/types"
	faasfrontendconf "yuanrong.org/kernel/pkg/frontend/types"
	faasschedulerconf "yuanrong.org/kernel/pkg/functionscaler/types"
)

// Config defines FaaS Controller configuration
type Config struct {
	ManagerInstanceNum   int              `json:"managerInstanceNum"`
	EnableRetry          bool             `json:"enableRetry"`
	NameSpace            string           `json:"nameSpace"`
	ClusterID            string           `json:"clusterID" valid:"optional"`
	ClusterName          string           `json:"clusterName" valid:"optional"`
	RegionName           string           `json:"regionName" valid:"optional"`
	AlarmConfig          alarm.Config     `json:"alarmConfig" valid:"optional"`
	RouterETCD           etcd3.EtcdConfig `json:"routerEtcd" valid:"required"`
	MetaETCD             etcd3.EtcdConfig `json:"metaEtcd"`
	TLSConfig            config.TLSConfig `json:"tlsConfig"`
	RawStsConfig         raw.StsConfig    `json:"rawStsConfig,omitempty"`
	SccConfig            crypto.SccConfig `json:"sccConfig"`
	SchedulerExclusivity []string         `json:"schedulerExclusivity" valid:"optional"`
}

// SchedulerConfig defines configuration faas scheduler needs
type SchedulerConfig struct {
	faasschedulerconf.Configuration
	SchedulerNum int `json:"schedulerNum"`
}

// FrontendConfig defines configuration faas frontend needs
type FrontendConfig struct {
	faasfrontendconf.Config
}

// ManagerConfig defines configuration faas manager needs
type ManagerConfig struct {
	CPU                float64           `json:"cpu"`
	Memory             float64           `json:"memory"`
	Image              string            `json:"image"`
	Version            string            `json:"version"`
	NodeSelector       map[string]string `json:"nodeSelector,omitempty"`
	ManagerInstanceNum int               `json:"managerInstanceNum"`
	LeaseRenewMinute   int               `json:"leaseRenewMinute" valid:"optional"`
	RouterEtcd         etcd3.EtcdConfig  `json:"routerEtcd" valid:"optional"`
	MetaEtcd           etcd3.EtcdConfig  `json:"metaEtcd" valid:"optional"`
	SccConfig          crypto.SccConfig  `json:"sccConfig" valid:"optional"`
	AlarmConfig        alarm.Config      `json:"alarmConfig" valid:"optional"`
	Affinity           string            `json:"affinity"`
	NodeAffinity       string            `json:"nodeAffinity" valid:"optional"`
	NodeAffinityPolicy string            `json:"nodeAffinityPolicy" valid:"optional"`
}

const (
	// CreatedTimeout max timeout of creating system function, kernel timeouts 120s
	CreatedTimeout = 15*time.Minute - 10*time.Second
	// KillSignalVal signal of kill
	KillSignalVal = 1
	// SyncKillSignalVal Synchronize signal of kill
	SyncKillSignalVal = 3
	// PreserveMetaKillSignalVal Preserve instance/app MeataData after sending the signal of kill
	PreserveMetaKillSignalVal = 5
	// FaaSSchedulerFunctionKey etcd key of faaS scheduler
	FaaSSchedulerFunctionKey = "0/0-system-faasscheduler/$latest"
	// FaaSSchedulerPrefixKey prefix etcd key of faaS scheduler
	FaaSSchedulerPrefixKey = "/sn/instance/business/yrk/tenant/0/function/" +
		"0-system-faasscheduler/version"
	// FasSFrontendFunctionKey etcd key of faaS frontend
	FasSFrontendFunctionKey = "0/0-system-faasfrontend/$latest"
	// FasSFrontendPrefixKey prefix etcd key of faaS frontend
	FasSFrontendPrefixKey = "/sn/instance/business/yrk/tenant/0/function/" +
		"0-system-faasfrontend/version/"
	// FasSManagerFunctionKey etcd key of faaS manager
	FasSManagerFunctionKey = "0/0-system-faasmanager/$latest"
	// FasSManagerPrefixKey prefix etcd key of faaS manager
	FasSManagerPrefixKey = "/sn/instance/business/yrk/tenant/0/function/" +
		"0-system-faasmanager/version/$latest/"
)

const (
	// SubEventTypeUpdate is update type of subscribe event
	SubEventTypeUpdate EventType = "update"
	// SubEventTypeDelete is delete type of subscribe event
	SubEventTypeDelete EventType = "delete"
	// SubEventTypeRecover is recover type of subscribe event
	SubEventTypeRecover EventType = "recover"
)

const (
	// EventKindInvalid is the wrong kind of function registry
	EventKindInvalid EventKind = iota
	// EventKindFrontend is the type of frontend registry
	EventKindFrontend
	// EventKindScheduler is the type of scheduler registry
	EventKindScheduler
	// EventKindManager is the type of function registry
	EventKindManager
)

type (
	// EventType defines registry event type
	EventType string
	// EventKind defines registry event kind
	EventKind uint8
)

// SubEvent contains event published to subscribers
type SubEvent struct {
	EventType
	EventKind
	EventMsg interface{}
}

// SystemFunctionCreator is the creator interface of system function
type SystemFunctionCreator interface {
	CreateExpectedInstanceCount(ctx context.Context) error
	CreateMultiInstances(ctx context.Context, count int) error
	CreateWithRetry(ctx context.Context, args []*api.Arg, extraParams *types.ExtraParams) error
	CreateInstance(ctx context.Context, function string, args []*api.Arg, extraParams *types.ExtraParams) string
	RollingUpdate(ctx context.Context, event *ConfigChangeEvent)
}

// SystemFunctionGetter is the getter interface of system function
type SystemFunctionGetter interface {
	GetInstanceCountFromEtcd() map[string]struct{}
	GetInstanceCache() map[string]*InstanceSpecification
}

// SystemFunctionKiller is the kill interface of system function
type SystemFunctionKiller interface {
	SyncKillAllInstance()
	KillInstance(instanceID string) error
}

// SystemFunctionRestorer is the recover interface of system function
type SystemFunctionRestorer interface {
	RecoverInstance(info *InstanceSpecification)
}

// SystemFunctionHandler is the handle interface of system function
type SystemFunctionHandler interface {
	HandleInstanceUpdate(instanceSpec *InstanceSpecification)
	HandleInstanceDelete(instanceSpec *InstanceSpecification)
}

// SystemFunction is group by system function interfaces
type SystemFunction interface {
	SystemFunctionCreator
	SystemFunctionGetter
	SystemFunctionKiller
	SystemFunctionRestorer
	SystemFunctionHandler
}

// InstanceSpecification contains specification of an instance
type InstanceSpecification struct {
	FuncCtx                   context.Context
	CancelFunc                context.CancelFunc
	InstanceID                string
	InstanceSpecificationMeta InstanceSpecificationMeta
}

// InstanceSpecificationMeta contains specification meta of a faas scheduler
type InstanceSpecificationMeta struct {
	InstanceID      string              `json:"instanceID"`
	RequestID       string              `json:"requestID"`
	RuntimeID       string              `json:"runtimeID"`
	RuntimeAddress  string              `json:"runtimeAddress"`
	FunctionAgentID string              `json:"functionAgentID"`
	FunctionProxyID string              `json:"functionProxyID"`
	Function        string              `json:"function"`
	RestartPolicy   string              `json:"restartPolicy"`
	Resources       Resources           `json:"resources" valid:",optional"`
	ScheduleOption  ScheduleOption      `json:"scheduleOption"`
	CreateOptions   map[string]string   `json:"createOptions"`
	StartTime       string              `json:"startTime"`
	InstanceStatus  InstanceStatus      `json:"instanceStatus"`
	Labels          []string            `json:"labels"`
	JobID           string              `json:"jobID"`
	SchedulerChain  []string            `json:"schedulerChain"`
	Args            []map[string]string `json:"args"`
}

// Resources contains resource specification of a scheduler instance in etcd
type Resources struct {
	Resources map[string]Resource `json:"resources"`
}

// Resource is system function resource
type Resource struct {
	name   string `json:"name"`
	Scalar Scalar `json:"scalar"`
}

// Scalar is system function scalar
type Scalar struct {
	Value int `json:"value"`
	Limit int `json:"limit"`
}

// ScheduleOption is system function scheduleOption
type ScheduleOption struct {
	SchedPolicyName string   `json:"schedPolicyName"`
	Priority        int      `json:"priority"`
	Affinity        Affinity `json:"affinity"`
}

// Affinity is system function Affinity
type Affinity struct {
	InstanceAffinity     InstanceAffinity `json:"instanceAffinity"`
	InstanceAntiAffinity InstanceAffinity `json:"instanceAntiAffinity"`
	NodeAffinity         NodeAffinity     `json:"nodeAffinity"`
}

// NodeAffinity is system function nodeAffinity
type NodeAffinity struct {
	Affinity map[string]string `json:"affinity"`
}

// InstanceAffinity is system function instanceAffinity
type InstanceAffinity struct {
	Affinity map[string]string `json:"affinity"`
}

// InstanceStatus is system function InstanceStatus
type InstanceStatus struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// ConfigChangeEvent config change event
type ConfigChangeEvent struct {
	FrontendCfg  *FrontendConfig
	SchedulerCfg *SchedulerConfig
	ManagerCfg   *ManagerConfig
	Error        error
	sync.WaitGroup
	TraceID string
}

const (
	// StateUpdate -
	StateUpdate = "update"
	// StateDelete -
	StateDelete = "delete"

	// FaasFrontendInstanceState -
	FaasFrontendInstanceState = "FaasFrontend"
	// FaasSchedulerInstanceState -
	FaasSchedulerInstanceState = "FaasScheduler"
	// FaasManagerInstanceState -
	FaasManagerInstanceState = "FaasManager"
)
