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
	"encoding/json"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/alarm"
	"yuanrong.org/kernel/pkg/common/faas_common/crypto"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/localauth"
	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	"yuanrong.org/kernel/pkg/common/faas_common/sts/raw"
	"yuanrong.org/kernel/pkg/common/faas_common/tls"
	"yuanrong.org/kernel/pkg/common/faas_common/types"
	wisecloudTypes "yuanrong.org/kernel/pkg/common/faas_common/wisecloudtool/types"
)

// Configuration defines configuration faas scheduler needs
type Configuration struct {
	Scenario                     string                           `json:"scenario" valid:"optional"`
	InstanceOperationBackend     int                              `json:"instanceOperationBackend" valid:"optional"`
	CPU                          float64                          `json:"cpu" valid:"optional"`
	Memory                       float64                          `json:"memory" valid:"optional"`
	PredictGroupWindow           int64                            `json:"predictGroupWindow"`
	AutoScaleConfig              AutoScaleConfig                  `json:"autoScaleConfig" valid:"required"`
	ScaleRetryConfig             ScaleRetryConfig                 `json:"scaleRetryConfig" valid:"optional"`
	LeaseSpan                    int                              `json:"leaseSpan" valid:"required"`
	FunctionLimitRate            int                              `json:"functionLimitRate"`
	RouterETCDConfig             etcd3.EtcdConfig                 `json:"routerEtcd" valid:"required"`
	MetaETCDConfig               etcd3.EtcdConfig                 `json:"metaEtcd" valid:"required"`
	DataSystemEtcd               etcd3.EtcdConfig                 `json:"dataSystemEtcd" valid:"optional"`
	SchedulerNum                 int                              `json:"schedulerNum" valid:"optional"`
	DockerRootPath               string                           `json:"dockerRootPath"`
	ClusterID                    string                           `json:"clusterID" valid:"optional"`
	ClusterName                  string                           `json:"clusterName" valid:"optional"`
	DiskMonitorEnable            bool                             `json:"diskMonitorEnable"`
	RegionName                   string                           `json:"regionName" valid:"optional"`
	AlarmConfig                  alarm.Config                     `json:"alarmConfig" valid:"optional"`
	NodeSelector                 map[string]string                `json:"nodeSelector,omitempty"`
	EphemeralStorage             float64                          `json:"ephemeralStorage,omitempty"`
	NpuEphemeralStorage          float64                          `json:"npuEphemeralStorage,omitempty"`
	HostAliases                  []v1.HostAlias                   `json:"hostaliaseshostname"`
	FunctionConfig               []FunctionDefaultConfig          `json:"functionConfig"`
	HTTPSConfig                  *tls.InternalHTTPSConfig         `json:"httpsConfig" valid:"optional"`
	LocalAuth                    localauth.AuthConfig             `json:"localAuth"`
	RawStsConfig                 raw.StsConfig                    `json:"rawStsConfig,omitempty"`
	SystemAuthConfig             raw.Auth                         `json:"systemAuthConfig,omitempty" valid:"optional"`
	XpuNodeLabels                []XpuNodeLabel                   `json:"xpuNodeLabels,omitempty"`
	ServiceAccountJwt            wisecloudTypes.ServiceAccountJwt `json:"serviceAccountJwt,omitempty"`
	Version                      string                           `json:"version"`
	Image                        string                           `json:"image"`
	ConcurrentNum                int                              `json:"concurrentNum"`
	ModuleConfig                 *ModuleConfig                    `json:"moduleConfig" valid:"optional"`
	StateDisable                 bool                             `json:"stateDisable"`
	EnableNPUDriverMount         bool                             `json:"enableNPUDriverMount"`
	DisableReplicaScaler         bool                             `json:"disableReplicaScaler"`
	TenantInsNumLimitEnable      bool                             `json:"tenantInsNumLimitEnable"`
	EnableHealthCheck            bool                             `json:"enableHealthCheck"`
	EnableRollout                bool                             `json:"enableRollout"`
	SccConfig                    crypto.SccConfig                 `json:"sccConfig" valid:"optional"`
	NameSpace                    string                           `json:"nameSpace"`
	Affinity                     string                           `json:"affinity"`
	MicroServiceSchedulingPolicy string                           `json:"msSchedulingPolicy" valid:"optional"`
	AuthenticationEnable         bool                             `json:"authenticationEnable" valid:"optional"`
	NodeAffinity                 string                           `json:"nodeAffinity" valid:"optional"`
	NodeAffinityPolicy           string                           `json:"nodeAffinityPolicy" valid:"optional"`
	DeployMode                   string                           `json:"deployMode" valid:"optional"`
	MetricsAddr                  string                           `json:"metricsAddr" valid:"optional"`
	MetricsHTTPSEnable           bool                             `json:"metricsHttpsEnable" valid:"optional"`
	PprofAddr                    string                           `json:"pprofAddr" valid:"optional"`
	SchedulerDiscovery           *SchedulerDiscovery              `json:"schedulerDiscovery" valid:"optional"`
	EnableSessionRecover         bool                             `json:"enableSessionRecover" valid:"optional"`
	DataSystemConfig             DataSystemConfig                 `json:"dataSystemConfig" valid:"optional"`
	CustomContainerEnv           map[string]string                `json:"customContainerEnv" valid:"optional"`
	HttpServerPort               string                           `json:"httpServerPort" valid:",optional"`
}

// AutoScaleConfig -
type AutoScaleConfig struct {
	SLAQuota      int `json:"slaQuota" valid:"required"`
	ScaleDownTime int `json:"scaleDownTime" valid:"required"`
	BurstScaleNum int `json:"burstScaleNum" valid:"required"`
}

// DataSystemConfig data system client put config
type DataSystemConfig struct {
	InitTimeoutMs   int      `json:"initTimeoutMs" validate:"required"`
	InitClusters    []string `json:"initClusters"`
	CurrentCluster  string   `json:"currentCluster" validate:"required"`
	UploadWriteMode int      `json:"uploadWriteMode" validate:"required"`
	UploadTTLSec    uint32   `json:"uploadTTLSec" validate:"required"`
}

// ScaleRetryConfig -
type ScaleRetryConfig struct {
	ReservedInstanceAlwaysRetry bool `json:"reservedInstanceAlwaysRetry" valid:"optional"`
}

// ModuleConfig config info
type ModuleConfig struct {
	ServicePort string `json:"servicePort" valid:",optional"`
}

// SchedulerDiscovery -
type SchedulerDiscovery struct {
	KeyPrefixType string `json:"keyPrefixType" valid:"optional"`
	RegisterMode  string `json:"registerMode" valid:"optional"`
}

// XpuNodeLabel Distinguish heterogeneous card types
type XpuNodeLabel struct {
	XpuType         string   `json:"xpuType,omitempty"`
	InstanceType    string   `json:"instanceType,omitempty"`
	NodeLabelKey    string   `json:"nodeLabelKey,omitempty"`
	NodeLabelValues []string `json:"NodeLabelValues,omitempty"`
}

// FunctionSpecification contains specification of a function
type FunctionSpecification struct {
	FuncCtx           context.Context        `json:"-"`
	CancelFunc        context.CancelFunc     `json:"-"`
	FuncKey           string                 `json:"-"`
	NameSpace         string                 `json:"-"`
	FuncMetaSignature string                 `json:"-"`
	FuncSecretName    string                 `json:"-"`
	MetaFromCR        bool                   `json:"-"`
	FuncMetaData      types.FuncMetaData     `json:"funcMetaData" valid:",optional"`
	S3MetaData        types.S3MetaData       `json:"s3MetaData" valid:",optional"`
	CodeMetaData      types.CodeMetaData     `json:"codeMetaData" valid:",optional"`
	EnvMetaData       types.EnvMetaData      `json:"envMetaData" valid:",optional"`
	StsMetaData       types.StsMetaData      `json:"stsMetaData" valid:",optional"`
	ResourceMetaData  types.ResourceMetaData `json:"resourceMetaData" valid:",optional"`
	InstanceMetaData  types.InstanceMetaData `json:"instanceMetaData" valid:",optional"` // new add
	ExtendedMetaData  types.ExtendedMetaData `json:"extendedMetaData" valid:",optional"`
	RootfsSpecMeta    types.RootfsSpecMeta   `json:"rootfs" valid:",optional"`
}

// PullTriggerRequestInfo include info of pullTrigger Option Create
type PullTriggerRequestInfo struct {
	PodName       string `json:"pod_name"`
	Image         string `json:"image"`
	DomainID      string `json:"domain_id,omitempty"`
	Namespace     string `json:"namespace,omitempty"`
	VpcName       string `json:"vpc_name,omitempty"`
	VpcID         string `json:"vpc_id,omitempty"`
	SubnetName    string `json:"subnet_name,omitempty"`
	SubnetID      string `json:"subnet_id,omitempty"`
	TenantCidr    string `json:"tenant_cidr,omitempty"`
	HostVMCidr    string `json:"host_vm_cidr,omitempty"`
	ContainerCidr string `json:"container_cidr"`
	Gateway       string `json:"gateway,omitempty"`
	Xrole         string `json:"xrole,omitempty"`
	AppXrole      string `json:"app_xrole,omitempty"`
}

// PullTriggerDeleteInfo include info of pullTrigger Option delete
type PullTriggerDeleteInfo struct {
	PodName string `json:"pod_name,omitempty"`
}

// NetworkConfig is a config in createOption which describes how to setup network in the environment where instance
// is running
type NetworkConfig struct {
	RouteConfig    RouteConfig    `json:"routeConfig"`
	TunnelConfig   TunnelConfig   `json:"tunnelConfig"`
	FirewallConfig FirewallConfig `json:"firewallConfig"`
	ProberConfig   ProberConfig   `json:"proberConfig"`
}

// RouteConfig is the config describes how to setup route
type RouteConfig struct {
	Gateway string `json:"gateway"`
	Cidr    string `json:"cidr"`
}

// TunnelConfig is the config describes how to setup tunnel
type TunnelConfig struct {
	TunnelName string `json:"tunnelName"`
	RemoteIP   string `json:"remoteIP"`
	Mode       string `json:"mode"`
}

// FirewallConfig is the config describes how to setup firewall
type FirewallConfig struct {
	Chain     string `json:"chain"`
	Table     string `json:"table"`
	Operation string `json:"operation"`
	Target    string `json:"target"`
	Args      string `json:"args"`
}

// ProberConfig is a config in createOption which describes how to perform certain prober action for instance
type ProberConfig struct {
	Protocol         string `json:"protocol"`
	SubMetaDigest    string `json:"subMetaDigest,omitempty"`
	Address          string `json:"address"`
	Gateway          string `json:"gateway,omitempty"`
	Interval         int    `json:"interval"`
	Timeout          int    `json:"timeout"`
	FailureThreshold int    `json:"failureThreshold"`
}

// DelegateNetworkConfig configures network in kernel
type DelegateNetworkConfig struct {
	PatInstances []PatInstance `json:"patInstances"`
}

// PatInstance pat instance message
type PatInstance struct {
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
}

// DelegateContainerConfig configures custom image in kernel
type DelegateContainerConfig struct {
	Image                  string                       `json:"image"`
	ImagePullPolicy        v1.PullPolicy                `json:"imagePullPolicy"`
	Env                    []v1.EnvVar                  `json:"env"`
	Command                []string                     `json:"command"`
	Args                   []string                     `json:"args"`
	UID                    int                          `json:"uid"`
	GID                    int                          `json:"gid"`
	VolumeMounts           []v1.VolumeMount             `json:"volumeMounts"`
	CustomGracefulShutdown types.CustomGracefulShutdown `json:"runtime_graceful_shutdown"`
	Lifecycle              v1.Lifecycle                 `json:"lifecycle"`
	ServiceAccountName     string                       `json:"serviceAccountName"`
}

// DelegateContainerSideCarConfig configures custom image sidecar in kernel
type DelegateContainerSideCarConfig struct {
	Name                 string                  `json:"name"`
	Image                string                  `json:"image"`
	Env                  []v1.EnvVar             `json:"env"`
	ResourceRequirements v1.ResourceRequirements `json:"resourceRequirements"`
	VolumeMounts         []v1.VolumeMount        `json:"volumeMounts"`
	Lifecycle            v1.Lifecycle            `json:"lifecycle"`
	LivenessProbe        v1.Probe                `json:"livenessProbe"`
	ReadinessProbe       v1.Probe                `json:"readinessProbe"`
}

// DelegateInitContainerConfig configures custom image init container in kernel
type DelegateInitContainerConfig struct {
	Name                 string                  `json:"name"`
	Image                string                  `json:"image"`
	Env                  []v1.EnvVar             `json:"env"`
	ResourceRequirements v1.ResourceRequirements `json:"resourceRequirements"`
	VolumeMounts         []v1.VolumeMount        `json:"volumeMounts"`
	Command              []string                `json:"command"`
	Args                 []string                `json:"args"`
}

// InstanceType defines instance type
type InstanceType string

const (
	// InstanceTypeOnDemand is the type of onDemand instance
	InstanceTypeOnDemand InstanceType = "onDemand"
	// InstanceTypeScaled is the type of scaled instance
	InstanceTypeScaled InstanceType = "scaled"
	// InstanceTypeReserved is the type of reserved instance
	InstanceTypeReserved InstanceType = "reserved"
	// InstanceTypeState is the type of state instance
	InstanceTypeState InstanceType = "state"
	// InstanceTypeUnknown is the type of unknown instance
	InstanceTypeUnknown InstanceType = "unknown"
)

const (
	base = 0.00001
)

// ValueType -
type ValueType int32

// ValueScalar -
type ValueScalar struct {
	Value float64 `json:"value"`
	Limit float64 `json:"limit"`
}

// ValueRanges -
type ValueRanges struct {
	Range []ValueRange `protobuf:"bytes,1,rep,name=range,proto3" json:"range,omitempty"`
}

// ValueSet -
type ValueSet struct {
	Items string `json:"items"`
}

// ValueRange -
type ValueRange struct {
	Begin uint64 `json:"begin"`
	End   uint64 `json:"end"`
}

// DiskInfo -
type DiskInfo struct {
	Volume    Volume `json:"volume"`
	Type      string `json:"type"`
	DevPath   string `json:"devPath"`
	MountPath string `json:"mountPath"`
}

// Volume -
type Volume struct {
	Mode          int32  `json:"mode"`
	SourceType    int32  `json:"sourceType"`
	HostPaths     string `json:"hostPaths"`
	ContainerPath string `json:"containerPath"`
	ConfigMapPath string `json:"configMapPath"`
	EmptyDir      string `json:"emptyDir"`
	ElaraPath     string `json:"elaraPath"`
}

// Affinity -
type Affinity struct {
	NodeAffinity         NodeAffinity     `json:"nodeAffinity"`
	InstanceAffinity     InstanceAffinity `json:"instanceAffinity"`
	InstanceAntiAffinity InstanceAffinity `json:"instanceAntiAffinity"`
}

// NodeAffinity -
type NodeAffinity struct {
	Affinity map[string]string `json:"affinity"`
}

// InstanceAffinity -
type InstanceAffinity struct {
	Affinity map[string]string `json:"affinity"`
}

// InstanceSpecification contains specification of a instance in etcd
type InstanceSpecification struct {
	InstanceID      string
	InstanceName    string
	RequestID       string               `json:"requestID" valid:",optional"`
	RuntimeID       string               `json:"runtimeID" valid:",optional"`
	RuntimeAddress  string               `json:"runtimeAddress" valid:",optional"`
	FunctionAgentID string               `json:"functionAgentID" valid:",optional"`
	FunctionProxyID string               `json:"functionProxyID" valid:",optional"`
	Function        string               `json:"function"`
	RestartPolicy   string               `json:"restartPolicy" valid:",optional"`
	Resources       types.Resources      `json:"resources"`
	ActualUse       types.Resources      `json:"actualUse" valid:",optional"`
	ScheduleOption  types.ScheduleOption `json:"scheduleOption"`
	CreateOptions   map[string]string    `json:"createOptions"`
	Labels          []string             `json:"labels"`
	StartTime       string               `json:"startTime"`
	InstanceStatus  types.InstanceStatus `json:"instanceStatus"`
	JobID           string               `json:"jobID"`
	SchedulerChain  []string             `json:"schedulerChain" valid:",optional"`
	ParentID        string               `json:"parentID"`
	ConcurrentNum   int
}

// InstanceThreadMetrics contains metrics of a specified instance thread collected by function accessor
type InstanceThreadMetrics struct {
	InsThdID      string
	ProcNumPS     float32
	ProcReqNum    int    `json:"procReqNum"`
	AvgProcTime   int    `json:"avgProcTime"` // millisecond
	MaxProcTime   int    `json:"maxProcTime"`
	IsAbnormal    bool   `json:"isAbnormal"`
	ReacquireData []byte `json:"reacquireData"`
	FunctionKey   string `json:"functionKey"`
}

// Instance defines a instance
type Instance struct {
	InstanceStatus    types.InstanceStatus
	InstanceType      InstanceType
	MetricLabelValues []string
	ResKey            resspeckey.ResSpecKey
	InstanceID        string
	InstanceName      string
	FuncKey           string
	FuncSig           string
	ConcurrentNum     int
	CreateSchedulerID string
	InstanceIP        string
	InstancePort      string
	NodeIP            string
	NodePort          string
	Permanent         bool
	ParentID          string
	PodID             string
	PodDeploymentName string
}

// Copy -
func (i *Instance) Copy() *Instance {
	if i == nil {
		return nil
	}
	return &Instance{
		InstanceStatus:    i.InstanceStatus,
		InstanceType:      i.InstanceType,
		MetricLabelValues: append([]string{}, i.MetricLabelValues...),
		ResKey:            i.ResKey,
		InstanceID:        i.InstanceID,
		InstanceName:      i.InstanceName,
		FuncKey:           i.FuncKey,
		FuncSig:           i.FuncSig,
		ConcurrentNum:     i.ConcurrentNum,
		CreateSchedulerID: i.CreateSchedulerID,
		InstanceIP:        i.InstanceIP,
		InstancePort:      i.InstancePort,
		NodeIP:            i.NodeIP,
		NodePort:          i.NodePort,
		Permanent:         i.Permanent,
		ParentID:          i.ParentID,
		PodID:             i.PodID,
		PodDeploymentName: i.PodDeploymentName,
	}
}

// WmInstance defines instance from workerManger
type WmInstance struct {
	IP             string                `json:"ip"`
	Port           string                `json:"port"`
	InstanceID     string                `json:"instanceID"`
	DeployedIP     string                `json:"deployed_ip"`
	DeployedNode   string                `json:"deployed_node"`
	TenantID       string                `json:"tenant_id"`
	Version        string                `json:"version"`
	OwnerIP        string                `json:"owner_ip"`
	IsReserved     bool                  `json:"isReserved"`
	IsDirectFunc   bool                  `json:"isDirectFunc"`
	HasInitializer bool                  `json:"hasInitializer"`
	Resource       types.PodResourceInfo `json:"resource,omitempty"`
}

// InstanceLease defines lease operations
type InstanceLease interface {
	Extend() error
	Release() error
	GetInterval() time.Duration
}

// SessionInfo -
type SessionInfo struct {
	SessionID  string
	SessionCtx context.Context
}

// InstanceAllocation defines a instance thread
type InstanceAllocation struct {
	Instance     *Instance
	Lease        InstanceLease
	SessionInfo  SessionInfo
	AllocationID string
}

// InstanceBuilder will create a instance
type InstanceBuilder func(string) *Instance

// InstanceAcquireRequest contains specifications for acquiring an instance
type InstanceAcquireRequest struct {
	FuncSpec            *FunctionSpecification
	ResSpec             *resspeckey.ResourceSpecification
	InstanceSession     types.InstanceSessionConfig
	InstanceName        string
	DesignateInstanceID string
	DesignateThreadID   string
	PoolLabel           string
	PoolID              string
	TraceID             string
	StateID             string
	CallerPodName       string
	TrafficLimited      bool

	SkipWaitPending bool
}

// InstanceCreateRequest contains specifications for creating an instance
type InstanceCreateRequest struct {
	FuncSpec     *FunctionSpecification
	ResSpec      *resspeckey.ResourceSpecification
	InstanceName string
	TraceID      string
	CreateEvent  []byte
}

// InstanceDeleteRequest contains specifications for deleting an instance
type InstanceDeleteRequest struct {
	FuncSpec     *FunctionSpecification
	ResSpec      *resspeckey.ResourceSpecification
	InstanceID   string
	InstanceName string
	TraceID      string
}

// FuncMetaArg defines funcMeta args
type FuncMetaArg struct {
	CodeID           string           `json:"codeID"`
	Kind             string           `json:"kind"`
	InvokeType       int              `json:"invokeType"`
	ObjectDescriptor ObjectDescriptor `json:"objectDescriptor"`
	Config           ConfigDescriptor `json:"config"`
}

// ObjectDescriptor defines object descriptor
type ObjectDescriptor struct {
	ModuleName     string `json:"moduleName"`
	ClassName      string `json:"className"`
	FunctionName   string `json:"functionName"`
	TargetLanguage string `json:"targetLanguage"`
	SrcLanguage    string `json:"srcLanguage"`
}

// ConfigDescriptor defines config descriptor
type ConfigDescriptor struct {
	RecycleTime int               `json:"RecycleTime"`
	FunctionID  map[string]string `json:"functionID"`
	JobID       string            `json:"jobID"`
	LogLevel    int               `json:"logLevel"`
}

// ExecutorInitResponse -
type ExecutorInitResponse struct {
	ErrorCode string          `json:"errorCode"`
	Message   json.RawMessage `json:"message"`
}

// InstancePoolState instance pool queue state save
type InstancePoolState struct {
	// Key: stateID - val: InstanceID
	StateInstance map[string]*Instance `json:"StateInstance" valid:"optional"`
}

// InstancePoolStateInput -
type InstancePoolStateInput struct {
	InstanceType       InstanceType
	ResKey             resspeckey.ResSpecKey
	ConcurrentNum      int
	InstanceStatusCode int32
	FuncKey            string
	FuncSig            string
	InstanceID         string
	StateID            string
	InstanceIP         string
	InstancePort       string
	NodeIP             string
	NodePort           string
}

// RolloutInstanceSpecification -
type RolloutInstanceSpecification struct {
	RegisterKey    string `json:"registerKey"`
	InstanceID     string `json:"instanceID"`
	RuntimeAddress string `json:"runtimeAddress"`
}

const (
	// StateUpdate -
	StateUpdate = "update"
	// StateDelete -
	StateDelete = "delete"
	// InstanceLifeCycleConsistentWithState -
	InstanceLifeCycleConsistentWithState = "ConsistentWithInstance"
)

// PodRequest define pod standard
type PodRequest struct {
	FunSvcID  string `json:"funSvcID"`
	NameSpace string `json:"nameSpace"`
}

// CustomUserArgs -
type CustomUserArgs struct {
	AlarmConfig       alarm.Config         `json:"alarmConfig" valid:"optional"`
	StsServerConfig   raw.ServerConfig     `json:"stsServerConfig"`
	ClusterName       string               `json:"clusterName"`
	DiskMonitorEnable bool                 `json:"diskMonitorEnable"`
	LocalAuth         localauth.AuthConfig `json:"localAuth"`
}

// StateArgs value of the 'stateInstanceID' Map
type StateArgs struct {
	IsStateValid bool
	StateID      string
}

// FunctionDefaultConfig -
type FunctionDefaultConfig struct {
	ConfigName string         `json:"configName"`
	Mount      v1.VolumeMount `json:"mount"`
}

// PredictResult -
type PredictResult struct {
	DataSetTimeWindow []int64              `json:"dataSetTimeWindow"`
	QPSResult         map[string][]float64 `json:"QpsResult"`
	IsValid           bool
}

// PredictQPSGroups -
type PredictQPSGroups struct {
	FuncKey   string
	QPSGroups []float64
}

// TenantInstanceMetaData define tenant instance quota
type TenantInstanceMetaData struct {
	MaxOnDemandInstance int64 `json:"maxOnDemandInstance" valid:",optional"`
	MaxReversedInstance int64 `json:"maxReversedInstance" valid:",optional"`
}

// TenantMetaInfo define tenant meta info
type TenantMetaInfo struct {
	TenantInstanceMetaData TenantInstanceMetaData `json:"tenantInstanceMetaData" valid:",optional"`
}

// TenantInsInfo define tenant instance info
type TenantInsInfo struct {
	OnDemandInsNum int64 `json:"onDemandInsNum" valid:",optional"`
	ReversedInsNum int64 `json:"reversedInsNum" valid:",optional"`
}

// SchedulingOptions define tenant instance Scheduling Options
type SchedulingOptions struct {
	resourcesMap     map[string]float64
	Priority         int32
	Resources        map[string]float64
	Extension        map[string]string
	Affinity         []api.Affinity
	ScheduleAffinity []byte
}

// TLSConfig tls config
type TLSConfig struct {
	HttpsInsecureSkipVerify bool     `json:"httpsInsecureSkipVerify"`
	TlsCipherSuitesStr      []string `json:"tlsCipherSuites"`
	TlsCipherSuites         []uint16 `json:"-"`
}

// IntOrString is a type that can hold an int32 or a string
type IntOrString struct {
	Type   Type
	IntVal int64
	StrVal string
}

// Type represents the stored type of IntOrString.
type Type int64

const (
	// Int The IntOrString holds an int.
	Int Type = iota
	// String The IntOrString holds a string.
	String
)

// UnmarshalJSON Overwrite JSON.Unmarshal
func (i *IntOrString) UnmarshalJSON(data []byte) error {
	var intValue int64
	if err := json.Unmarshal(data, &intValue); err == nil {
		i.Type = Int
		i.IntVal = intValue
		return nil
	}
	var stringValue string
	if err := json.Unmarshal(data, &stringValue); err == nil {
		i.Type = String
		i.StrVal = stringValue
		return nil
	}
	return fmt.Errorf("expected int or string, but got %s", data)
}
